# Multica 自托管 · 迁移到另一台机器（操作手册）

> 基于 **2026-09-15** 本机现状盘点：macOS 26.6.2 + OrbStack（Docker 29.4.0），compose 项目 `multica`，
> **源码构建栈**（`multica dev`，commit `7d9613bd2`），内网 Caddy 入口 + Redis + Lark 邮件 / 飞书机器人。
> 本文只回答「怎么把这套搬到别的机器」。上团队服务器还差什么（域名、TLS、备份、GitHub App）见 `TEAM_DEPLOY_PLAN.md`。
>
> 标注规则：**[实测]** = 本文写完后在本机真跑过并贴出结果；**[推论]** = 由表结构 / 代码得出、未在目标机验证；
> **[注意]** = 本机实测踩过的坑。
>
> **2026-09-20 复核**：**宿主/引擎那一层**另有专文 `ORBSTACK_CONFIG.md`（OrbStack 版本、两个 `daemon.json` 的坑、
> 动态网段、`host.docker.internal`、数据盘构成、新机落地与验收清单），本文只管**数据与配置的搬运**。
> 复核中发现两处前提已过期，已在正文就地修正：**附件卷不再是空的（6.55 MB）**、**git HEAD 漂到 `8c4f4328`**。

---

## 0. 一句话结论

要搬的是 **3 份数据 + 2 份本地文件 + 1 份身份**，其余全部重新生成：

```
多主机迁移 = Postgres dump  +  附件卷 tar  +  根 .env（含全部密钥）
          + docker-compose.lan.yml + Caddyfile.lan（都是未跟踪文件，git clone 拿不到）
          + （可选）~/.multica/{config.json, daemon.id} —— 保住 agent 与 runtime 的绑定
```

代码从 GitHub 取、镜像在目标机重建、Redis 与 Caddy 数据**不搬**、daemon 重新上线。

---

## 1. 迁移对象清单（体积均为本机实测）

| # | 资产 | 本机位置 | 搬运方式 | 体积 | 不搬 / 搬丢的后果 |
|---|---|---|---|---|---|
| 1 | **数据库** | 卷 `multica_pgdata`（容器内 `/var/lib/postgresql/data`） | `pg_dump -Fc` → `pg_restore` | 卷 75.85 MB，dump **804 KB** | 全部工作区/agent/issue/会话/集成安装全丢 |
| 2 | **附件** | 卷 `multica_backend_uploads` | `tar` 打包 | **2026-09-20 实测 6.55 MB（已有内容）**；09-15 写本文时为空 | **现在搬丢就是真丢附件**——§5 的 tar/还原路径不能再当"空卷演练"，必须真跑一次 |
| 3 | **配置与密钥** | 根 `.env`（`.gitignore` 第 21 行 `.env*` 忽略，**永不随 clone 来**） | `scp` 原样拷贝 | 37 KB / 651 行 | `JWT_SECRET` 丢 = 所有人被登出；`MULTICA_LARK_SECRET_KEY` 丢 = 飞书机器人解不开、只能重装；`POSTGRES_PASSWORD` 丢 = 旧 dump 里的连接信息对不上 |
| 4 | **内网入口** | `docker-compose.lan.yml`、`Caddyfile.lan`（未跟踪） | 原样拷贝 | 4.6 KB + 903 B | 少了 `/ws` 直连后端那一层，实时更新挂在 Next.js rewrite 上（WebSocket 升级不会被转发） |
| 5 | **代码** | `git` 仓库，origin `https://github.com/multica-ai/multica.git`，分支 `main` | `git clone` + `git checkout 7d9613bd2` | ~20 MB | 版本漂移。**2026-09-20 实测 HEAD 已是 `8c4f4328`（09-19）**，且 `multica-backend:dev` 是当天重建的 → 要么就用当前 HEAD，要么明确指定要哪个 commit |
| 6 | **镜像** | 本地 `multica-backend:dev`(141 MB) / `multica-web:dev`(373 MB) | 目标机重建（推荐）/ `docker save`（**仅同 CPU 架构**）/ 改用官方 tag | 共 ~1.5 GB 解压 | 见 §11 镜像三选一 |
| 7 | Redis 数据 | 卷 `multica_redis_data` | **不搬** | 3 KB | 只丢限流计数、心跳/租约与缓存，daemon 重连即恢复 |
| 8 | Caddy 数据 | 卷 `multica_edge_data` / `multica_edge_config` | **不搬** | <1 KB | 当前是纯 HTTP，无证书可丢 |
| 9 | daemon 身份 | `~/.multica/config.json` + `~/.multica/daemon.id` | 见 §8（拷贝 = 保住 agent 绑定） | 236 B | 不搬 → 新机注册出新 runtime，4 个 agent 要重新选 runtime |
| 10 | daemon 会话历史 | `~/.multica/pi-sessions/`、`reasonix-state/` | 可搬可不搬 | 数 MB | 只影响本机会话记录，不影响服务端 |
| 11 | `TEAM_DEPLOY_PLAN.md`、`reasonix.toml`、`--help` | 仓库根（未跟踪） | 可选 | 36 KB | 不影响运行（`--help` 疑似误建的 5 KB 文件） |

**项目名固定**：`docker-compose.selfhost.yml` 顶层有 `name: multica` **[实测]**，所以目标机的目录叫什么、放哪里都不影响
卷名（`multica_pgdata`）与容器名（`multica-backend-1`）——两边的命令可以逐字照抄。

**当前数据规模 [实测]**：库 18 MB / `schema_migrations` 507 行 / 1 个 workspace（土卫九）/ 4 个 agent /
2 个用户 / 5 个 issue / 26 条任务队列 / 附件 0 / 飞书安装 1。装了 `pgcrypto 1.3`、`pg_trgm 1.6`（镜像虽叫
pgvector，但**没有** `vector` 扩展）。

---

## 2. 目标机前置条件

| 项 | 要求 | 说明 |
|---|---|---|
| CPU 架构 | **arm64 或 amd64 二选一，且与迁移方式匹配** | 本机镜像是 `linux/arm64` **[实测]**，搬到 x86_64 服务器上不能直接 `docker load` |
| 操作系统 | macOS + OrbStack / Docker Desktop，或 Linux + Docker Engine | `host.docker.internal` 只有前两者自带，Linux 见 §11 |
| Docker | Engine 24+ **且必须带 `docker compose` v2 插件** | 自托管 compose 用顶层 `name:` 语法，legacy `docker-compose` v1 直接解析失败（`Makefile` 有专门的早退提示） |
| 网络 | 容器能出网：`ghcr.io`（若用官方镜像）、`registry.npmjs.org` + `proxy.golang.org`（若源码构建） | 本机 OrbStack 容器**直连**这些站点，宿主才需要 `HTTPS_PROXY` **[实测]**；国内服务器请给 Docker daemon 配代理或镜像站，否则 `make selfhost-build` 会卡在依赖下载 |
| 磁盘 | ≥ 20 GB 空闲 | 镜像 ~1.5 GB + 构建缓存（本机 build cache 20.13 GB）+ 数据 80 MB |
| 端口 | `80`（内网入口）、`3100`、`8080`、`6379` 可用 | 前者绑 `0.0.0.0`，后三者只绑 `127.0.0.1` |
| 工具 | `git`、`openssl`、`tar` | |

```bash
# 目标机自检
docker compose version            # 必须 v2.x，不能是 v1
docker info --format '{{.Architecture}}'   # 本机实测输出 aarch64；与源机比对
lsof -nP -iTCP:80,3100,8080,6379 -sTCP:LISTEN || echo "端口空闲"
df -h / | tail -1
```

---

## 3. 源机：冻结 + 打包（停机窗口约 5 分钟）

### 3.1 冻结写入

```bash
cd /Users/shubinzhang/Documents/GitHub/multica
COMPOSE="docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml"

# 停掉所有写入方（edge/frontend/backend），Postgres 与 Redis 继续跑，用于取一致快照
$COMPOSE stop edge frontend backend
$COMPOSE ps            # 确认只有 postgres / redis 是 Up
```

> **三个 `-f` 必须一起给**：`docker-compose.lan.yml` 是本机唯一一份未跟踪 override，redis、edge 以及
> backend 的 `RATE_LIMIT_*` / `FF_PLUGINS_V1` / healthcheck 都在它里面。少给一个，下次 recreate 会**静默丢掉**这些配置。

### 3.2 打数据库 dump

```bash
mkdir -p ~/multica-migrate && cd ~/multica-migrate

# ✅ 用 -Fc（自定义格式，pg_restore 可并行、可选择性恢复）
docker exec multica-postgres-1 pg_dump -U multica -d multica -Fc > multica.dump
ls -lh multica.dump                       # 本机实测 804 KB

# 记下版本，便于目标机核对（507）
docker exec multica-postgres-1 psql -U multica -d multica -tAc \
  "select count(*), max(version) from schema_migrations"
```

**[注意]** 不要写成 `pg_dump ... | gzip > x.sql.gz`：管道失败时退出码仍是 0，会得到一个静默损坏的备份。要压缩就用 `-Fc` 后单独 `gzip`。

### 3.3 打附件卷

```bash
docker run --rm -v multica_backend_uploads:/data -v "$PWD":/backup alpine \
  tar czf /backup/multica-uploads.tgz -C /data .
ls -lh multica-uploads.tgz                # 本机实测：卷为空，包内只有 ./ 一条
```

### 3.4 打包配置与入口文件

```bash
cd /Users/shubinzhang/Documents/GitHub/multica
mkdir -p ~/multica-migrate/conf
cp .env Caddyfile.lan docker-compose.lan.yml ~/multica-migrate/conf/
# 可选：把 daemon 身份也带上（见 §8），它是 600 权限的密钥文件
cp ~/.multica/config.json ~/.multica/daemon.id ~/multica-migrate/conf/ 2>/dev/null || true

# 强烈建议：把密钥单独再存一份到密码管理器/离线介质
#   .env 里必须原样保留的键（值不要写进任何文档或聊天）：
#   JWT_SECRET  POSTGRES_PASSWORD  DATABASE_URL  MULTICA_VCS_SECRET_KEY
#   MULTICA_LARK_SECRET_KEY  SMTP_PASSWORD  MULTICA_LLM_API_KEY
#   （若曾启用）MULTICA_SLACK_SECRET_KEY / MULTICA_WECOM_SECRET_KEY /
#              MULTICA_DINGTALK_SECRET_KEY / MULTICA_PLUGIN_SECRET_KEY
#              CLOUDFRONT_PRIVATE_KEY_SECRET / GITHUB_APP_PRIVATE_KEY

# 记录一下源机 git 状态，迁移后要与之对齐
git log -1 --format='%H %ad %s'
git status --short
```

### 3.5 打包传输

```bash
cd ~/multica-migrate && tar czf ../multica-migrate.tgz .
scp ../multica-migrate.tgz <user>@<new-host>:~/
# 校验（两端一致）
shasum -a 256 ../multica-migrate.tgz
```

**[注意]** 这个包里有 `.env`（全部生产密钥，含 SMTP 密码与 LLM API Key）。只能走你信任的通道（内网 `scp` / 加密介质），
不要进网盘、不要进 git、不要进聊天记录。

### 3.6 源机收尾

```bash
$COMPOSE stop            # 停掉 postgres / redis；卷与数据原样保留
```

**不要执行 `docker compose down -v`** —— `-v` 会删掉 `multica_pgdata` 与 `multica_backend_uploads`，这是本手册里唯一的不可逆操作。

---

## 4. 目标机：部署代码与配置

### 4.1 取代码（固定到同一 commit）

```bash
git clone https://github.com/multica-ai/multica.git ~/multica
cd ~/multica && git checkout 7d9613bd2c446d1d281d5ff28ab571a397324ddd
```

### 4.2 落配置

```bash
tar xzf ~/multica-migrate.tgz -C ~/multica-migrate
cd ~/multica
cp ~/multica-migrate/conf/.env ~/multica-migrate/conf/Caddyfile.lan \
   ~/multica-migrate/conf/docker-compose.lan.yml .
chmod 600 .env
```

`make selfhost-build` 只在 `.env` **不存在**时才生成随机密钥；这里 `.env` 已存在，所以它不会覆盖——但**先放好 `.env` 再构建**，
顺序别反，否则会先生成一套新密钥（新 `JWT_SECRET`、新 `POSTGRES_PASSWORD`），后面还得手工改回来。

### 4.3 按新机器改 4 处配置（§7 有查找命令）

| 文件 | 键 | 源机现值 | 目标机应为 |
|---|---|---|---|
| `.env` | `CORS_ALLOWED_ORIGINS` | `http://localhost:3100,http://192.168.20.9,http://192.168.4.36` | 保留 `http://localhost:3100`，其余换成新机的 LAN 地址 |
| `.env` | `MULTICA_APP_URL` | `http://192.168.20.9` | 新机 LAN 地址；**这个值决定 `/api/config` 发给同事的 daemon 安装地址** |
| `docker-compose.lan.yml` | `RATE_LIMIT_TRUSTED_PROXIES` | `192.168.148.0/24` | 新机 `multica_default` 网段（查找命令见 §7.3） |
| `Caddyfile.lan` | `host.docker.internal:8080/3100` | — | Linux 目标机改成 `backend:8080` / `frontend:3000`（§11.2） |

`.env` 里这些**不用改**：`POSTGRES_DB/USER/PASSWORD`、`DATABASE_URL`、`JWT_SECRET`、`MULTICA_*_SECRET_KEY`、
`SMTP_*`、`MULTICA_LLM_*`、`ALLOW_SIGNUP=false`、`ALLOWED_EMAIL_DOMAINS=guidefuture.com`、`FRONTEND_PORT=3100`、
`PORT=8080`、`FRONTEND_ORIGIN=http://localhost:${FRONTEND_PORT}`（compose 会把它插值成 `http://localhost:3100` **[实测]**）。

---

## 5. 起栈 + 恢复数据（顺序不能反）

```bash
cd ~/multica
COMPOSE="docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml"

# ① 先只起 Postgres —— 它第一次启动会按 POSTGRES_DB 建一个**空**库
$COMPOSE up -d postgres
until docker exec multica-postgres-1 pg_isready -U multica -d multica >/dev/null 2>&1; do sleep 1; done

# ② 把 dump 灌进这个空库
docker exec -i multica-postgres-1 pg_restore -U multica -d multica --no-owner --no-acl \
  < ~/multica-migrate/multica.dump
# [实测] 804 KB 的 dump 恢复耗时 0.86 s，零报错

# ③ 还原附件卷（命名卷不存在时 docker run 会自动创建，不必先起 backend）
docker run --rm -v multica_backend_uploads:/data -v ~/multica-migrate:/backup alpine \
  sh -c 'tar xzf /backup/multica-uploads.tgz -C /data'

# ④ 起其余服务（backend 启动时会自己跑 migration）
$COMPOSE up -d --build
```

**[注意] 顺序**：如果一上来就 `up -d`（全栈），backend 会先把 507 条 migration 跑完、把表都建好，
之后 `pg_restore` 就会满屏 `relation already exists`。要么按上面①②④的顺序，要么给 `pg_restore` 加 `--clean --if-exists`（不推荐，慢且噪音大）。

**[注意]** 若目标机不需要内网入口（比如只自己用），可以只用两个 `-f`（`selfhost` + `selfhost.build`）——
但**不要**在源机这么做：源机的 `.env` 里 `REDIS_URL=redis://host.docker.internal:6379/0` 指向的是 lan override 起的 redis 容器。

### 5.1 首次构建的耗时与网络

`--build` 会在容器内跑 `go mod download` 和三段 `go build`，以及 `pnpm install --frozen-lockfile` + Next.js 构建。
国内机器先确认 Docker 能出网（`docker run --rm alpine wget -q -O- https://proxy.golang.org`），否则先给 daemon 配代理/镜像站。

---

## 6. 验证清单（逐条对齐源机）

```bash
COMPOSE="docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml"
cd ~/multica

# 1) 五容器都在，且 postgres/backend/redis 是 healthy
$COMPOSE ps

# 2) 后端健康（含 DB 与迁移检查）
curl -s 127.0.0.1:8080/healthz      # 期望 {"status":"ok","checks":{"db":"ok","migrations":"ok"}}

# 3) 版本与源机一致
docker exec multica-backend-1 ./multica version   # 期望 multica dev (commit: 7d9613bd2, ...)

# 4) 数据一条不少（与 §3.2 记录的 507 对齐）
docker exec multica-postgres-1 psql -U multica -d multica -tAc \
  "select (select count(*) from schema_migrations) mig,
          (select count(*) from workspace) ws,
          (select count(*) from agent) agents,
          (select count(*) from \"user\") users,
          (select count(*) from issue) issues,
          (select count(*) from channel_installation) chan"
# 源机实测：507 | 1 | 4 | 2 | 5 | 1

# 5) 扩展与源机一致
docker exec multica-postgres-1 psql -U multica -d multica -tAc \
  "select extname, extversion from pg_extension order by 1"
# 源机实测：pg_trgm 1.6 / pgcrypto 1.3 / plpgsql 1.0

# 6) 公开配置正确（这里能看出 §7 的 MULTICA_APP_URL 有没有生效）
curl -s 127.0.0.1:3100/api/config
# 期望 "daemon_server_url":"http://<新机IP>"，且 "plugins_v1":true、"allow_signup":false

# 7) 内网入口三连
curl -s -o /dev/null -w '%{http_code}\n' http://<新机IP>/            # 200
curl -s -o /dev/null -w '%{http_code}\n' http://<新机IP>/health      # 200
curl -s -o /dev/null -w '%{http_code}\n' http://<新机IP>/api/workspaces   # 401（未登录，说明代理链通）

# 8) WebSocket 升级（需要 seed 里的 workspace slug）
docker exec multica-postgres-1 psql -U multica -d multica -tAc "select slug from workspace"
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     -H "Origin: http://<新机IP>" \
     "http://<新机IP>/ws?workspace_slug=<slug>" | head -3
# 期望 101 Switching Protocols；换成白名单外的 Origin 应得 403

# 9) 邮件（发一次验证码）
curl -s -X POST http://<新机IP>/auth/send-code -H 'Content-Type: application/json' \
     -d '{"email":"<你的邮箱>@guidefuture.com"}' -o /dev/null -w '%{http_code}\n'   # 200
docker logs --tail=50 multica-backend-1 | grep -i "SMTP relay"
# 期望出现 SMTP relay smtp.larksuite.com:465 (implicit-tls) from=...
# 且**不应**出现 "no email backend"
```

浏览器侧再手工走一遍：内网入口 → 填邮箱 → 收码 → 登录 → 侧边栏能看到 `土卫九` 与 4 个 agent、`Settings → Plugins` 入口在。

---

## 7. 需要按新机器改的配置：怎么查值

### 7.1 找新机 LAN 地址

```bash
# macOS
ipconfig getifaddr en0
# Linux
hostname -I | awk '{print $1}'
```

### 7.2 改 `.env`

```bash
cd ~/multica
NEWIP=<新机IP>
# 备份后改（值里不要留尾斜杠）
sed -i.bak "s#^CORS_ALLOWED_ORIGINS=.*#CORS_ALLOWED_ORIGINS=http://localhost:3100,http://${NEWIP}#" .env
sed -i.bak "s#^MULTICA_APP_URL=.*#MULTICA_APP_URL=http://${NEWIP}#" .env
docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml up -d backend
```

**[注意]** **改 `.env` 后必须 `up -d`，`restart` 不重新读 `.env`**（compose 用的是显式 `environment:` 列表，没有 `env_file`）。

### 7.3 找 Docker 网段（填 `RATE_LIMIT_TRUSTED_PROXIES`）

```bash
docker network inspect multica_default -f '{{(index .IPAM.Config 0).Subnet}}'
# 本机实测：192.168.148.0/24 —— 这正是 docker-compose.lan.yml 里那个默认值
```

为什么必须对：`backend` 看到的对端是 **frontend 容器的 IP**，若把 Docker 网段登记为可信代理，限流就会按
`X-Forwarded-For` 里的真实客户端 IP 计数；否则整个团队共用一个桶（认证类默认 5 次/分钟）必然误伤。
把这一行改成**目标机**的网段即可（`RATE_LIMIT_AUTH=30` / `RATE_LIMIT_AUTH_VERIFY=120` 是放宽后的阈值，与机器无关，保持不动）。

### 7.4 改 `Caddyfile.lan`（仅 Linux 目标机，见 §11.2）

改完热加载，不用重建容器：

```bash
$COMPOSE exec edge caddy reload --config /etc/caddy/Caddyfile
```

---

## 8. 成员 / 本机 daemon 的上线方式

服务器只承载 API/WS/DB；**agent 任务在各自的机器上执行**，所以每台要用的机器都要装 CLI + daemon。

### 8.1 装 CLI（目标机）

```bash
brew install multica-ai/tap/multica          # 首选（macOS/Linux）
multica version                              # 源机：0.4.43 (commit: 2ae2dbbb8)
```

无 Homebrew 时按 `CLI_INSTALL.md` 的 Option B 从 GitHub Releases 取对应架构的二进制。

### 8.2 方式 A（推荐，保住 agent 绑定）：携带身份

```bash
# 目标机
mkdir -p ~/.multica && chmod 700 ~/.multica
scp <源机>:~/.multica/config.json <源机>:~/.multica/daemon.id ~/.multica/
chmod 600 ~/.multica/config.json
multica daemon start && multica daemon status
```

**为什么可行 [推论，依据表结构与迁移注释]**：
- `daemon.id` 是**磁盘上的持久 UUID**（`server/internal/daemon/config.go:485-513`，回退顺序 `MULTICA_DAEMON_ID` > override > 文件）；
- `agent_runtime` 的唯一键是 `(workspace_id, daemon_id, provider)`（`migrations/004`、`120`、`121`），
  注册是 upsert —— 带上同一个 `daemon_id`，6 个 runtime 行会被**原地复用**并重新转为 online；
- **[实测]** 当前 4 个 agent 的 `runtime_id` 全部指向 `daemon_id = 01a088f5-4675…` 的 runtime 行，
  而这个前缀正是 `~/.multica/daemon.id`。所以沿用同一身份，绑定不会断。

> ⚠️ 前提：**源机上的 daemon 必须先停**（`multica daemon stop`）。同一个 `daemon_id` 两台机器同时在线会互相抢任务。

### 8.3 方式 B（干净重登）：新身份

```bash
# 在目标机（服务器本机用 localhost:3100；其他成员机用服务器入口 http://<服务器IP>）
multica setup self-host --server-url http://<服务器IP> --app-url http://<服务器IP> \
                        --callback-host <本机IP>
# 也可以先只改配置文件指向：
multica config set server_url http://<服务器IP>
multica daemon start
```

代价 **[推论]**：新 `daemon.id` → 新 runtime 行 → 旧行随旧机离线，**4 个 agent 需要在界面上重新选一次 runtime**
（provider 名相同，如 `Reasonix`/`Oh-My-Pi` 挑新机器上线的那个）。Lark 机器人绑在「MyNewProject 开发工程师」上，
它依赖 `omp` runtime，重绑时别漏。

**[注意]** 成员机的 `--server-url` 应该指向**内网入口**（`http://<服务器IP>`，80 端口），不是 `:3100` ——
`3100`/`8080` 只绑服务器回环。这也正是 `/api/config` 里 `daemon_server_url` 由 `MULTICA_APP_URL` 推导的原因。

**[注意]** 目标机至少要有一个被识别的 agent CLI（`claude`/`codex`/`kimi`/`opencode`/`omp`/`reasonix`…），
否则 daemon 起来了但 `Agents:` 为空，任务会报「no runtime」。启动日志里每个 runtime 一行 `agent version detected`。

---

## 9. 两台机器并行 / 切流量

| 场景 | 做法 | 注意 |
|---|---|---|
| 只是搬家，一次性切 | 源机停 → 目标机起 → 验证 → 源机长期保留停机 | 最简单，推荐 |
| 想并行跑一段时间 | 目标机用**同一份 dump**起新的独立实例，但 **`JWT_SECRET` / `POSTGRES_PASSWORD` 必须与源机一致**才能让同一批账号有效 | 两套库会各自分叉；**app 只能有一个**：daemon 的 `workspace_id` 与 token 绑定的是具体实例 |
| 用同一台 daemon 连两边 | **不行** —— `daemon.id` + `token` 决定它认哪台服务器 | 要并行就用两份 `~/.multica/profiles/<name>/` 配置 |
| 内网入口 IP 变更 | 跟 §7 三处（CORS / MULTICA_APP_URL / RATE_LIMIT_TRUSTED_PROXIES） | 漏改 `CORS_ALLOWED_ORIGINS` → WS 会 **403**（本机已验证过白名单生效） |

---

## 10. 回滚

```bash
# 目标机：停掉新栈（保留卷，别加 -v）
$COMPOSE stop

# 源机：原样起回
cd /Users/shubinzhang/Documents/GitHub/multica
$COMPOSE up -d
curl -s 127.0.0.1:8080/healthz
```

源机的卷从未被删，`up -d` 即回到迁移前状态。**建议在目标机验收通过前不要回收源机数据**：
DB dump 会随两侧继续使用而分叉，一旦新机开始产生数据，回滚就不是无损的了。

---

## 11. 平台差异：macOS + OrbStack  ↔  Linux 服务器

### 11.1 `host.docker.internal`（本机能用，Linux 不一定）

本机 OrbStack 把它解析到宿主 **[实测]**，所以 `.env` 的 `REDIS_URL=redis://host.docker.internal:6379/0`
与 `Caddyfile.lan` 里的 `host.docker.internal:8080/3100` 都工作正常。

这是**有意为之**的设计：走已发布的宿主回环端口，即使某次只给两个 `-f` 起栈（服务列表里没有 redis）后端也连得上。
**代价是拓扑绑死在「容器能回连宿主」这个前提上**。

Linux Docker 默认**不注册** `host.docker.internal`（要 `extra_hosts: ["host.docker.internal:host-gateway"]`），
而且即便加上，`host-gateway` 指向的是网桥地址（如 `172.17.0.1`），而 `3100`/`8080`/`6379` 只绑在 **127.0.0.1** ——
从容器连 `172.17.0.1:3100` 会被拒。**所以在 Linux 上这套要改成走服务名。**

### 11.2 Linux 推荐改法（本机已实测服务名可用）

三个容器本来就在同一个 `multica_default` 网络里，用服务名更短、更快、且不依赖宿主：

```caddyfile
# Caddyfile.lan（Linux）
:80 {
	@ws path /ws /ws/*
	handle @ws {
		reverse_proxy backend:8080 {
			flush_interval -1
		}
	}
	reverse_proxy frontend:3000
}
```

```bash
# .env（Linux）
REDIS_URL=redis://redis:6379/0     # 替代 redis://host.docker.internal:6379/0
```

**[实测]** 在现栈里验证过这三种容器间访问都通（无需任何改动即可复现）：

| 验证 | 结果 |
|---|---|
| `docker exec multica-edge-1 wget -q -O- http://backend:8080/healthz` | `{"status":"ok","checks":{"db":"ok","migrations":"ok"}}` |
| `docker exec multica-edge-1 wget -S -O /dev/null http://frontend:3000/` | `HTTP/1.1 200 OK` |
| `docker exec multica-backend-1 nc -z -w2 redis 6379` | 通 |

`frontend` 侧的 `REMOTE_API_URL=http://backend:8080` 本来就是服务名，没有这个问题。

### 11.3 日志轮转（换到普通 Linux 主机会突然变行为）

上游 compose 三个文件都没有 `logging` 段。本机靠 OrbStack 守护进程的**内置默认值**（`json-file, max-size=20m, max-file=5`）
侥幸有轮转，而**普通 Linux Docker 的 json-file 默认不限制**，日志会一路涨到填满磁盘。
本机已在未跟踪的 `docker-compose.lan.yml` 里用 `x-logging` 锚点给五个服务显式声明了同一组数值 —— **这份文件必须一起搬过去**。

### 11.4 其他平台差异

| 项 | macOS + OrbStack | Linux |
|---|---|---|
| 宿主是否需要代理 | 需要（本机 `HTTPS_PROXY=http://127.0.0.1:7897`） | 视机房而定 |
| 容器能否直连 npm/goproxy | **能**（OrbStack 直连）**[实测]** | 通常需要镜像站或 daemon 代理 |
| 合盖/休眠 | 掉线，全员重连 | 不适用 |
| 镜像架构 | `linux/arm64`（Apple Silicon） | 多为 `amd64` → 见 §11.5 镜像三选一 |
| 防火墙 | macOS 防火墙当前关闭 | 记得只放行 80，别放 3100/8080/6379 |

### 11.5 镜像三选一

| 方案 | 命令 | 适用 |
|---|---|---|
| **A. 目标机重建（推荐）** | `docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml up -d --build` | 架构随意，但要能出网；首次约十几分钟 |
| B. 传镜像 | 源机 `docker save multica-backend:dev multica-web:dev \| gzip > img.tgz`；目标机 `gunzip -c img.tgz \| docker load` | **仅同架构**（两机都是 Apple Silicon，或都是 amd64）；省去重建 |
| C. 换官方发布镜像 | 去掉 `-f docker-compose.selfhost.build.yml`，`.env` 里把 `MULTICA_IMAGE_TAG` 从 `latest` 改成固定版本，再 `docker compose -f docker-compose.selfhost.yml -f docker-compose.lan.yml up -d` | 想上团队环境、要固定版本；**注意 `latest` 会漂移** |

选 C 时务必先确认该 tag 已发布且有多架构 manifest（本机的 `ghcr.io/multica-ai/multica-backend:latest` 是 3 天前拉的 `linux/arm64`）。
选 C 之后，`.env` 里的 `MULTICA_BACKEND_IMAGE` / `MULTICA_WEB_IMAGE`（默认即 `ghcr.io/multica-ai/multica-*`）保持原值即可。

---

## 12. 已实测的坑（本机沉淀，迁移时逐条对照）

| # | 现象 / 坑 | 结论 |
|---|---|---|
| 1 | `pg_dump \| gzip > x.sql.gz` 失败仍返回 0 | 用 `-Fc`，或分开压缩 |
| 2 | 先起全栈再 `pg_restore` | 满屏 `already exists`；先只起 postgres 再灌 |
| 3 | `docker compose restart` 不重读 `.env` | 改完配置一律 `up -d`（可只给服务名） |
| 4 | 只给两个 `-f` 起栈 | 静默丢掉 redis / edge / 限流 / 插件开关 / backend healthcheck |
| 5 | `down -v` | 删除 `pgdata` 与 `uploads`，**不可逆**，不要用 |
| 6 | 写进 `.env` ≠ 进容器 | 未被 compose `VAR: ${VAR:-}` 透传的键（如 `MULTICA_WORKSPACE_ID`、`LOCAL_UPLOAD_DIR`、`ANALYTICS_*`）永远进不去；改配置前 `grep 变量名 docker-compose*.yml` |
| 7 | `MULTICA_LARK_SECRET_KEY` / `JWT_SECRET` 丢失 | 前者 = 已装飞书机器人全部失效、只能重装；后者 = 所有人被登出 |
| 8 | 邮件配置了两端不一致 | `SMTP_HOST` 一旦非空就**只走 SMTP、不再把验证码打到日志**；主机或密码写错 = 谁都登不上（`server/internal/service/email.go:349`） |
| 9 | 查飞书集成状态看不出来 | 要看通用表 `channel_installation` / `channel_user_binding` / `channel_task_delivery`；`lark_*` 那套是空的历史路径 |
| 10 | `APP_ENV` 为空 | 靠 compose 兜底成 `production`；目标机若换跑法（直接跑二进制）会意外落进 dev 模式，建议显式写 `production` |
| 11 | `FRONTEND_ORIGIN` 的 `Secure` cookie | 它决定 cookie 是否带 `Secure`（`server/internal/auth/cookie.go:118`）：`http://…` 才允许内网 HTTP 登录；**换 HTTPS 时必须同步改** |
| 12 | `/readyz` 在内网入口 404 | 前端 rewrite 白名单没有它（上游有意为之），深度探针只在服务器本机 `127.0.0.1:8080` 跑 |
| 13 | 没有 Redis 时 | 限流退化为进程内实现、realtime 走进程内中继；单实例可跑，多副本必须 Redis |
| 14 | 邀请限流 | 默认 actor 10 次/10 分钟，**一次性拉 20 人会被卡**，要分两批（`RATE_LIMIT_INVITATION_*` 未被 compose 透传，要调得先加进 compose） |

---

## 13. 尚未验证 / 上线前要补的

- **恢复演练**已做过一次（本手册的 dump→`pg_restore`→行数校验，§5 的 [实测]），但**目标机上的整套验收（§6）尚未在真实新机跑过**。
- **附件卷 tar/还原**路径未做实测（当前卷为空，包内无内容可验证）；等有附件后应先演练一次。
- **方式 A（携带 `daemon.id`）未实测**：本书依据表唯一键与 `daemon.id` 持久化代码得出的 [推论]，
  第一次迁移时建议先在目标机验证 `agent_runtime.daemon_id` 与 `agent.runtime_id` 是否原样复用，再停源机 daemon。
- **定时备份仍缺失**：迁移到常开机器后必须补 —— 至少要覆盖 `pgdata`、`multica_backend_uploads`、根 `.env`（含 `JWT_SECRET`、`MULTICA_LARK_SECRET_KEY`）三样，并做一次异地保留。
- **HTTPS / 域名**：本手册全程 HTTP。上公网前照 `TEAM_DEPLOY_PLAN.md` §4 / §10.3 改 `FRONTEND_ORIGIN`、`MULTICA_PUBLIC_URL`、`CORS_ALLOWED_ORIGINS`（cookie 的 `Secure` 会随之变化），并关掉 80 之外的暴露面。

---

## 附：完整命令速查（目标机视角）

```bash
# 0. 解包 + 取代码
tar xzf ~/multica-migrate.tgz -C ~/multica-migrate
git clone https://github.com/multica-ai/multica.git ~/multica
cd ~/multica && git checkout 7d9613bd2c446d1d281d5ff28ab571a397324ddd
cp ~/multica-migrate/conf/{.env,Caddyfile.lan,docker-compose.lan.yml} . && chmod 600 .env
# 改 4 处（§4.3 / §7）：CORS_ALLOWED_ORIGINS、MULTICA_APP_URL、
#                      RATE_LIMIT_TRUSTED_PROXIES、Caddyfile.lan（Linux）

# 1. 起库 → 灌数据 → 起全栈
COMPOSE="docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml"
$COMPOSE up -d postgres
until docker exec multica-postgres-1 pg_isready -U multica -d multica >/dev/null 2>&1; do sleep 1; done
docker exec -i multica-postgres-1 pg_restore -U multica -d multica --no-owner --no-acl < ~/multica-migrate/multica.dump
$COMPOSE up -d --build
docker run --rm -v multica_backend_uploads:/data -v ~/multica-migrate:/backup alpine \
  sh -c 'tar xzf /backup/multica-uploads.tgz -C /data'

# 2. 验收（§6）
$COMPOSE ps && curl -s 127.0.0.1:8080/healthz && docker exec multica-backend-1 ./multica version

# 3. daemon（§8）
multica daemon start && multica daemon status
```
