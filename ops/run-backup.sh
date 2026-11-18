#!/bin/bash
#
# ============================================================================
# Multica 栈每日备份脚本（三件套）
# ============================================================================
#
# 备份对象（缺一不可，恢复时才不会"库在附件没在"）：
#   1) Postgres 数据库 —— pg_dump 自定义格式（-Fc）
#      为什么用 -Fc：pg_restore 可以并行恢复、可以只恢复单个对象；纯 SQL 文本则不能。
#      为什么在容器里跑：宿主机不必装与容器同版本的 pg_dump，版本不匹配会直接失败。
#      为什么不能 `pg_dump | gzip > x.gz`：管道上游失败时 gzip 的退出码仍是 0，
#      会得到一份静默损坏、还自以为完整的备份（本机迁移手册实测教训）。
#      -Fc 单事务快照，本身就是一致性的，不需要为每日备份停栈。
#   2) 附件卷（本地上传/头像）—— 用一次性 alpine 容器只读挂载后 tar。
#      为什么不直接 tar Docker 卷目录：卷的真实路径在 OrbStack VM 里，宿主不可见。
#      已知边界：tar 期间若正好有人上传附件，该文件可能被打进半成品状态；
#      凌晨 3 点窗口基本无写入，可接受。要绝对一致需先 `docker compose ... stop backend`。
#   3) 仓库根 .env —— 它是唯一不在版本控制里的东西，含全部密钥
#      （JWT_SECRET / POSTGRES_PASSWORD / MULTICA_LARK_SECRET_KEY / SMTP_PASSWORD …）。
#      .env 丢了 = 全员被登出 + 飞书机器人全废，只能重装。备份文件与目录一律 600/700。
#
# 落位与轮转：
#   - 所有产物先写进 BACKUP_DIR 下的 .tmp.* 临时目录，打包完再 `mv`（同文件系统 rename
#     是原子的）。任何一步失败，trap 删临时目录，不会留下半成品冒充备份。
#   - 轮转按 mtime 排序（见末尾 ls -1t），只保留最近 KEEP 份。
#     不是纯覆盖：昨天备份里才发现的问题（误删、坏数据），覆盖式方案就没得救了。
#
# 部署（新机器 / Mac mini）—— 本文件即唯一来源，无需改文件：
#   一步跑通：`ops/run-backup.sh`；参数全部走环境变量覆盖，默认值见下方参数块。
#     BACKUP_DIR=~/multica-backups KEEP=5 ops/run-backup.sh
#   REPO（只为取 .env）默认按脚本位置推断：脚本在 <clone>/ops/ 下就自动用该 clone；
#   脚本被拷到别处（如 ~/multica-backups/run-backup.sh）时回落到
#   $HOME/Documents/GitHub/multica，因此拷贝部署必须显式设 REPO。
#   容器名/卷名随 compose 项目名变化（独立项目名会带前缀），用
#   `docker ps --format '{{.Names}}' | grep postgres` 与 `docker volume ls | grep uploads`
#   核对后覆盖 PG_CONTAINER / UPLOADS_VOLUME。
#
# 定时（macOS）：launchd 的 StartCalendarInterval，Label=com.multica.backup。
#   选 launchd 而不是 crontab 的原因：笔记本凌晨 3 点大概率合盖睡眠，cron 错过不补；
#   launchd 会在唤醒后补跑错过的日历触发。Linux 等价物是 systemd timer：
#   OnCalendar=*-*-* 03:00:00 + Persistent=true（后者就是"唤醒/开机后补跑"）。
#   plist 放 ~/Library/LaunchAgents/com.multica.backup.plist，下面的 __REPO__
#   换成本机 clone 绝对路径（launchd 不展开 $HOME 与命令替换）：
#     <?xml version="1.0" encoding="UTF-8"?>
#     <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
#     <plist version="1.0"><dict>
#       <key>Label</key><string>com.multica.backup</string>
#       <key>ProgramArguments</key>
#       <array><string>/bin/bash</string><string>__REPO__/ops/run-backup.sh</string></array>
#       <key>StartCalendarInterval</key><dict><key>Hour</key><integer>3</integer><key>Minute</key><integer>0</integer></dict>
#       <key>StandardOutPath</key><string>__HOME__/multica-backups/backup.log</string>
#       <key>StandardErrorPath</key><string>__HOME__/multica-backups/backup.log</string>
#       <key>RunAtLoad</key><false/>
#     </dict></plist>
#   加载/卸载：
#     launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.multica.backup.plist
#     launchctl bootout  gui/$(id -u) ~/Library/LaunchAgents/com.multica.backup.plist
#
# 验证与恢复：
#   验证 dump 可读（不真恢复）：
#     docker exec -i <PG_CONTAINER> pg_restore --list < <(tar -xzOf <包名> "$(tar -tzf <包名> | grep '^db-')") | head
#   完整恢复流程见仓库根 MIGRATION_TO_NEW_HOST.md §5（先只起 postgres 灌空库，
#   再起 backend 跑迁移 —— 顺序反了会满屏 relation already exists）。
#   红线：永远不要 `docker compose down -v`（-v 会删掉 pgdata 与附件卷）。
# ============================================================================
#
# ── 部署参数（换机器部署时只需要核对这一个块；每一项都可用环境变量覆盖）──────
# 取值方法：
#   容器名：docker ps --format '{{.Names}}' | grep postgres
#   卷  名：docker volume ls | grep uploads
#   独立 compose 项目名会让两者都带 <项目名> 前缀，注意同步改。
BACKUP_DIR="${BACKUP_DIR:-$HOME/multica-backups}"  # 备份目录；生产机建议放独立磁盘/另做异地副本
KEEP="${KEEP:-5}"                                  # 保留份数；库越大保留越少，反之亦然
PG_CONTAINER="${PG_CONTAINER:-multica-postgres-1}" # postgres 容器名
UPLOADS_VOLUME="${UPLOADS_VOLUME:-multica_backend_uploads}"  # 附件卷名
DB_NAME="${DB_NAME:-multica}"                      # 库名（与 POSTGRES_DB 一致）
DB_USER="${DB_USER:-multica}"                      # 库用户（与 POSTGRES_USER 一致）
# REPO 未显式给出时：脚本在 <clone>/ops/ 下就用该 clone（Mac mini 直接 clone 后即用），
# 否则回落到本机既有路径（拷贝式部署继续可用）。
if [ -z "${REPO:-}" ]; then
  SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
  if [ -f "$SCRIPT_DIR/../.env" ]; then
    REPO=$(cd "$SCRIPT_DIR/.." && pwd)
  else
    REPO="$HOME/Documents/GitHub/multica"
  fi
fi
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

# launchd 给的环境 PATH 只有 /usr/bin:/bin:…，docker 在 /usr/local/bin（OrbStack）
# 或 /opt/homebrew/bin（Homebrew，Apple Silicon），必须显式补上。
export PATH=/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin

STAMP=$(date +%Y%m%d-%H%M%S)
LOG="$BACKUP_DIR/backup.log"

log() { echo "[$(date '+%F %T')] $*" >>"$LOG"; }

# 日志轮转：backup.log 每天只涨几十行，但 launchd 的 stdout/stderr 也往这里写；
# 超过 1MB 就压成 .log.1，避免几年后变成意外的大文件。
if [ -f "$LOG" ] && [ "$(stat -f %z "$LOG" 2>/dev/null || stat -c %s "$LOG")" -gt 1048576 ]; then
  mv "$LOG" "$LOG.1"
fi

log "=== backup start ($STAMP) ==="

# 前置检查：postgres 容器不在跑（比如栈被 stop）就直接放弃本次。
# 这里不打成"失败退出"是刻意的：定时任务场景下留一条 ABORT 日志比天天报错更可读。
# .env 缺失则是配置错误（REPO 指错机器/指错路径），必须报失败，否则会"备份成功但没有密钥"。
if ! docker inspect "$PG_CONTAINER" --format '{{.State.Running}}' 2>/dev/null | grep -q true; then
  log "ABORT: $PG_CONTAINER not running — no backup taken"
  exit 0
fi
if [ ! -f "$REPO/.env" ]; then
  log "ABORT: $REPO/.env not found — set REPO to the deploy clone"
  exit 1
fi

# 临时目录建在 BACKUP_DIR 同文件系统内，保证最后 mv 是原子 rename 而不是跨盘复制。
TMP=$(mktemp -d "$BACKUP_DIR/.tmp.XXXXXX")
trap 'rm -rf "$TMP"' EXIT

# 1) 数据库
docker exec "$PG_CONTAINER" pg_dump -Fc -U "$DB_USER" "$DB_NAME" >"$TMP/db-$STAMP.dump"
log "pg_dump ok ($(du -h "$TMP/db-$STAMP.dump" | cut -f1))"

# 2) 附件卷（只读挂载；alpine 镜像体积小、纯打包工具用途）
docker run --rm -v "$UPLOADS_VOLUME:/data:ro" -v "$TMP":/backup alpine \
  tar czf "/backup/uploads-$STAMP.tar.gz" -C /data .
log "uploads tar ok"

# 3) .env（复制后立刻收紧权限，防止被同机其他进程读）
cp "$REPO/.env" "$TMP/env-$STAMP"
chmod 600 "$TMP/env-$STAMP"

# 总包：三件套打成一个带日期文件，恢复时一个文件就是完整现场。
tar -czf "$TMP/backup-$STAMP.tar.gz" -C "$TMP" \
  "db-$STAMP.dump" "uploads-$STAMP.tar.gz" "env-$STAMP"
chmod 600 "$TMP/backup-$STAMP.tar.gz"
mv "$TMP/backup-$STAMP.tar.gz" "$BACKUP_DIR/backup-$STAMP.tar.gz"
log "package ok: backup-$STAMP.tar.gz"

# 轮转：ls -1t 按 mtime 排序删除最旧的，只留 KEEP 份。
# 用 mtime 而不是文件名排序：手工补拷进来的旧备份也能被正确清掉。
ls -1t "$BACKUP_DIR"/backup-*.tar.gz 2>/dev/null | tail -n +$((KEEP + 1)) | while read -r f; do
  rm -f "$f"
  log "pruned $(basename "$f")"
done

log "=== backup done ($STAMP) ==="
