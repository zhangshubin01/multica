# Multica 自托管 —— 20 人 Android 团队 / 60 个项目 部署方案

> 现状是一台 MacBook 上的 OrbStack 单机 dev 栈（只绑 127.0.0.1）。本文是从"能跑"到"20 人能共用"的差距清单与目标架构。
> 生成于 2026-09-14，基于仓库 v0.4.43-21 与官方文档（`SELF_HOSTING.md`、`SELF_HOSTING_ADVANCED.md`、`apps/docs/content/docs/**`、https://multica.ai/docs）。
>
> **时间基准**：本文跨 2026-09-14 → 09-17 多天写成，条目上的日期是"写这条时"记的。实测锚点（2026-09-17
> 17:50 CST / 09:50 UTC）：postgres/redis/edge 创建于 `2026-09-15T04:16Z`（=09-15 12:16 CST），
> backend/frontend 重建于 `2026-09-17T06:06Z`（=09-17 14:06 CST），`.env` 最后修改 `2026-09-15 12:31 CST`。
> 标 09-15 的条目大概率落在 09-15 当天；09-16 / 09-17 两天的产出（WS 探针、P0-2 隔离实验、GitHub App 调研、
> §12 文档通读）此前被顺手记成 09-15，**以本节与 §12 的日期为准**。会话结束后不再补记历史日期。

## 1. 现状 vs 目标

| 维度 | 现状 | 20 人共用所需 |
|---|---|---|
| 可达性 | `127.0.0.1:3100/8080`，仅本机 | 常开主机 + 域名 + TLS 反代（含 `/ws` 升级） |
| 认证邮件 | 未配置，验证码只打在容器日志 | 必须接 Resend 或公司 SMTP |
| 注册 | `ALLOW_SIGNUP=true`，无域名白名单 | 邀请制 / 限制公司邮箱域 |
| 版本 | 源码 dev 构建（`multica dev`），`MULTICA_IMAGE_TAG=latest` | 固定版本（发布镜像 tag 或固定 commit） |
| 主机 | 笔记本，合盖即离线 | 常开服务器（Linux，4C/8–16G 起） |
| 备份 | 无定时备份（仅手动 dump 过一次） | 每日 pg_dump + 附件卷 + `.env`，异地保留 |
| Redis | **已开启（2026-09-14，见 §3）** | 建议开 |
| 指标 | `METRICS_ADDR` 空 | 至少开指标 + 日志留存 |
| 对象存储 | 附件在 Docker 卷 `multica_backend_uploads` | **决定（2026-09-15）：暂不切 S3**，继续用卷并纳入备份；多副本/迁移时再上 S3 |
| GitHub | App 未配置 | 60 个项目要挂仓库/PR 卡片 → 建议配 GitHub App |

## 2. 目标架构（单机 compose 起步）

```
浏览器 ──TLS──> Caddy (443)
                 ├── /ws, /ws/*  ──> 127.0.0.1:8080   (flush_interval -1)
                 └── 其余         ──> 127.0.0.1:3000
                                        │
                              frontend ─┴─> backend ──> postgres (pgvector 17)
                                                └─────> redis 7 (可选/推荐)

成员各自机器：multica CLI + agent daemon（运行时 = 各人的"工位"，与服务器规模无关）
```

- **服务器只承载 API/WS/DB**，Agent 执行发生在成员本机，所以不必为 20 人的 agent 算力买大机器；但别用笔记本当服务器。
- 数据量参考：当前整库 17 MB、5 个 issue、2 个 project —— 60 个项目短期内单机 PG 毫无压力。

## 3. Redis：这个规模下建议开

官方定性是"可选"（`/docs/developers/architecture`：单实例开发环境用进程内实现），但你的场景有两条把它推成"推荐/必需"：

1. **公开访问必须开认证限流**，限流依赖 `REDIS_URL`；而且 20 人共用一个出口时代理后的真实客户端 IP 必须通过 `RATE_LIMIT_TRUSTED_PROXIES` 传进来，否则所有人共用一个 5 次/分钟的桶，必然误伤。
2. **将来多副本**（滚动升级、横向扩容）必需：realtime 中继（`REALTIME_RELAY_MODE`，`REDIS_URL` 非空即启用，无 off 档）、频道 WebSocket 租约、token 缓存。

```bash
# 栈内加 redis（独立容器亦可），.env 里：
REDIS_URL=redis://<redis-host>:6379/0
RATE_LIMIT_TRUSTED_PROXIES=<反代所在网段，如 172.16.0.0/12>
# 共享该 Redis 的实例必须 maxmemory-policy noeviction（租约键被驱逐会断连）
```

### 3.1 当前实现状态（2026-09-14 已开启）

- 容器：compose 服务 `redis`（容器名 `multica-redis-1`，`redis:7-alpine`，`redis-server --maxmemory-policy noeviction`，
  发布到 `127.0.0.1:6379`，healthcheck `redis-cli ping`）。定义在未跟踪的 `docker-compose.lan.yml` 里，
  因此和 backend/frontend/postgres/edge 同属 `multica` 项目（OrbStack 同一个组、随项目命令起停）；
  与它并列的 backend 那段还补了限流相关变量（见下）。
- `.env`：`REDIS_URL=redis://host.docker.internal:6379/0`。**故意不用服务名 `redis:6379`** —— 走已发布的宿主
  回环端口，这样某次只用两个 `-f` 起栈（服务列表里没有 redis）时后端也连得上，不会因少给 override 而失联。
- `docker-compose.lan.yml`（未跟踪）里给 backend 补：`RATE_LIMIT_TRUSTED_PROXIES=192.168.148.0/24`、
  `RATE_LIMIT_AUTH=30`、`RATE_LIMIT_AUTH_VERIFY=120`（默认 5/20 是按"每 IP"设计的，团队共用一个出口会误伤）。
- 生效证据：启动日志 `realtime: Redis relay enabled ... mode=sharded`；`auth rate limiting disabled` 告警消失；
  Redis 中已有 `mul:runtime:hb:*`、`mul:claim:runtime:*`、`mul:auth:pat:*` 等键；`/auth/send-code` 后出现
  `mul:ratelimit:auth:send-code:<ip>`（count 递增、TTL 60s）。

**实测发现（重要）**：限流在这个拓扑下**分不出客户端** —— 浏览器 → Docker 发布端口 → `edge` → `frontend` → `backend`，
后端看到的对端始终是前端容器 IP（`192.168.148.4`），即使注入 `X-Forwarded-For` 也传不到后端。
所以认证限流实际是"全局刹车"，键里永远是那一个 IP。要让限流按真实客户端 IP 计数，
必须把反代放在**宿主机或云上**（客户端 IP 不被 Docker NAT 折叠），而不是反代容器里。

回退：`docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml stop redis`
+ 删 `.env` 的 `REDIS_URL` 行 + 重建 backend（`.env` 里那两处 LAN 改动保持不变）。

## 4. 必须改的配置清单

| # | 变量 / 动作 | 值 / 说明 |
|---|---|---|
| 1 | Caddy 反代 + 自动 TLS | `/ws` 与 `/ws/*` → 8080 且 `flush_interval -1`；其余 → 3000 |
| 2 | `FRONTEND_ORIGIN` / `CORS_ALLOWED_ORIGINS` | `https://multica.<公司域>` —— 不设的话公网下 WS 升级 403，实时功能静默失效 |
| 3 | `MULTICA_APP_URL` | 同上公开地址 |
| 4 | `MULTICA_PUBLIC_URL` | 公开 API 地址，autopilot webhook / 插件回调用 |
| 5 | `COOKIE_DOMAIN` | 跨子域部署时设 |
| 6 | `RESEND_API_KEY` + `RESEND_FROM_EMAIL` 或 `SMTP_*` | 认证必需 |
| 7 | `ALLOW_SIGNUP=false` + `ALLOWED_EMAIL_DOMAINS=<公司域>` | 收敛注册，成员走邀请 |
| 8 | `REDIS_URL` + `RATE_LIMIT_TRUSTED_PROXIES` | ✅ 已开（2026-09-14），见 §3.1 |
| 9 | `METRICS_ADDR=:9090` | 只绑内网/加鉴权，接 Prometheus 可选 |
| 10 | `MULTICA_IMAGE_TAG=v0.4.43` | 固定版本，别用 `latest` |
| 11 | `GITHUB_APP_ID` / `GITHUB_APP_PRIVATE_KEY` / `GITHUB_WEBHOOK_SECRET` / `GITHUB_APP_SLUG` | PR 卡片、CI 快照、"从 GitHub 选择仓库"；需公网 webhook |
| 12 | 备份 cron | 见 §6 |
| 13 | 迁移时**保留原 `JWT_SECRET`** | 否则所有会话/凭据失效 |

### 4.1 邮件怎么配（认证必需）

机制（`server/internal/service/email.go`）：**`SMTP_HOST` 一旦非空就优先走 SMTP**，否则用 Resend；
两者都空 → 验证码只打在容器日志里（当前状态）。compose 已透传全部相关变量，
所以配置 = 改 `.env`（第 234–259 行那块）+ 重建 backend，**不需要改上游 compose 文件**。

方案 A：Resend（云部署最省事）

```
RESEND_API_KEY=re_xxxxxxxx
RESEND_FROM_EMAIL=noreply@yourcompany.com   # 必须是 Resend 里已验证的发信域
```
前提：在 resend.com 建 key 并验证发信域（按提示加 SPF/DKIM DNS 记录）；未验证的域只能发给自己的测试地址。

方案 B：公司/自建 SMTP（内网团队更常见）

```
SMTP_HOST=smtp.exmail.qq.com            # 换成你们邮件服务商的
SMTP_PORT=465
SMTP_USERNAME=notify@yourcompany.com
SMTP_PASSWORD=<客户端专用密码/授权码>     # 通常不是登录密码
SMTP_FROM_EMAIL=notify@yourcompany.com  # 不设会回退到 RESEND_FROM_EMAIL
SMTP_TLS=implicit                       # 465 会自动 implicit；587 用默认 starttls
# SMTP_TLS_INSECURE=true                # 仅自签/私有 CA
# SMTP_EHLO_NAME=multica.yourcompany.com # 严格中继（如 Google Workspace）拒默认 EHLO 时设真实 FQDN
```

常见坑：465 = 隐式 TLS（`implicit`/`smtps`/`ssl` 或留空自动），587 = STARTTLS（默认）；
Google Workspace 之类需要"应用专用密码"或把出口 IP 加进 SMTP 中继白名单，并且常需 `SMTP_EHLO_NAME`；
注意服务商的发信额度与频率限制。

飞书 / Lark 邮箱（2026-09-14 在本机实测过连通性）

本项目采用**国际版 Lark + `zhangshubin@guidefuture.com`**。已核实该域名的邮件确实托管在 Lark 国际版：
`dig mx guidefuture.com` → `mx1/mx2/mx3.larksuite.com`，SPF 为 `include:spf.onlarksuite.com`。

```
SMTP_HOST=smtp.larksuite.com             # 中国版飞书邮箱用 smtp.feishu.cn
SMTP_PORT=465
SMTP_TLS=implicit                        # 465 自动 implicit；587 留空走 STARTTLS
SMTP_USERNAME=zhangshubin@guidefuture.com
SMTP_PASSWORD=<Lark 的"应用专用密码">
SMTP_FROM_EMAIL=zhangshubin@guidefuture.com   # 必须与认证账号一致，留空会回退到 noreply@multica.ai
```

实测：`smtp.larksuite.com` 与 `smtp.feishu.cn` 的 465/587 均开放，TLS 主机名校验通过
（`smtp.larksuite.com` 的证书 SAN 覆盖该域名，公开可信 CA），**不需要** `SMTP_TLS_INSECURE`。
管理侧需要在 Lark 里允许该邮箱使用第三方客户端 / IMAP-SMTP；组织若启用 SSO 或二次验证，
必须用"应用专用密码"而不是账号登录密码（这是 535 认证失败最常见的原因）。

⚠️ 代码事实（`server/internal/service/email.go:349`）：**`SMTP_HOST` 一旦非空就只走 SMTP，不再把验证码打到
日志**——所以主机和密码必须同时配好再重建，否则验证码既发不出去也不打印，全员登录不上。

验证：重建 backend 后，`docker logs multica-backend-1 | grep -i "no email backend"` 应**不再出现**；
在登录页发一次验证码，邮箱应收到 6 位码（日志里也有发送结果）。

✅ **2026-09-14 本机已跑通**（源码构建栈，`.env` 第 256–263 行）：启动日志出现
`EmailService: SMTP relay smtp.larksuite.com:465 (implicit-tls) from=zhangshubin@guidefuture.com`，
无 `no email backend` 告警；`POST /auth/send-code`（经内网入口 `http://192.168.20.9/auth/send-code`）
返回 200、耗时约 2.3s（真实 SMTP 往返），`verification_code` 表落库，IMAP 侧确认邮件已进收件箱。
注意两点：① Lark 会用邮箱账号的显示名覆盖 `From` 的显示名，收到的发件人显示为该账号昵称，
信封地址仍是 `SMTP_FROM_EMAIL`，不影响投递；② 同一邮箱 60 秒内只能发一次验证码（代码里按 email 限流）。
排查顺序：535 → 密码类型；550 → `From` 与认证账号不一致；EOF / EHLO 被拒 → 配 `SMTP_EHLO_NAME`。

注册收敛 ✅ **2026-09-14 已生效**：`ALLOW_SIGNUP=false` + `ALLOWED_EMAIL_DOMAINS=guidefuture.com`。
判定顺序（`server/internal/handler/auth.go:244`）：显式邮箱白名单 → 域名白名单 → `ALLOW_SIGNUP` 开关 →
"设了白名单但没匹配上"兜底；**老用户永远放行**（新用户才受这套限制）。匹配是域名整段 `EqualFold`，
子域（`mail.guidefuture.com`）不算。实测：`someone@gmail.com` → 403 `user registration is disabled`；
`<新地址>@guidefuture.com` → 200；管理员本人（老用户）→ 200。`/api/config` 的 `allow_signup` 已为 `false`，
登录页会据此隐藏注册入口；改这个开关只需重启 backend，不用重建前端。
工作区邀请邮件走同一个后端，配一次两边都通。

## 5. 20 人 / 60 个项目的建模

- 一个 workspace（现有 `土卫九`）+ 60 个 project；成员按业务线用 squad 分组。
- 每个成员在自己机器装 CLI + daemon 并注册 runtime；runtime 数量与项目数无关，按"人/机器"增长。
- 用 label / 自定义属性做跨项目筛选与看板分组。

容量参考：API 连接池默认 `DATABASE_MAX_CONNS=25`/副本，20 人远未触顶；WS 并发 20 人 × 1–2 tab 对单副本无压力。

## 6. 备份（迁移前必做）

```bash
# 不要管道给 gzip：失败时退出码仍是 0
docker compose -f docker-compose.selfhost.yml exec -T postgres \
  pg_dump -U multica multica > /backup/multica-$(date +%F).sql && gzip /backup/multica-$(date +%F).sql
```
- 覆盖对象：Postgres 卷 `multica_pgdata`、附件卷 `multica_backend_uploads`、根 `.env`（含 `JWT_SECRET` 等）。
- migration 只向前：**每次升级前**先 dump。

## 7. 从当前 OrbStack 迁到团队的步骤

1. 准备常开主机；用 compose（或 `deploy/helm/multica`）部署固定版本。
2. 从本机 dump 数据库并恢复；同步附件卷；沿用原密钥类 env。
3. 配 Caddy + DNS + 邮件 + 注册收敛 + Redis + 指标。
4. 邀请 20 人，各自装 CLI/daemon 注册 runtime。
5. 切 DNS；本机 OrbStack 栈留作个人 dev 环境。

## 8. 风险清单

- **不要把端口改成 `0.0.0.0` 裸奔**：compose 注释明确警告 Docker 会绕过宿主防火墙（UFW/iptables），弱凭据下等于直接暴露；一律走反代。
- 开放注册 + 公网 = 任何人都能建号；必须先关再开放访问。
- 无邮件后端 = 20 人无法登录（现在只有日志里的 DEV 验证码）。
- `latest` + 源码 dev 构建 = 版本漂移；团队环境固定 tag。
- 笔记本当服务器：合盖/换网即全体掉线，WS 重连风暴。
- 附件只存在容器卷里：不纳入备份就会随主机故障一起丢。

## 9. 内网试跑（已启用，2026-09-14）

目标：先让同局域网的同事用浏览器访问，验证连通性 / 实时更新 / 多人协作是否可用，暂不做公网与 TLS。

**已落地的改动**（两个新增文件都未纳入版本控制）：

| 文件 | 作用 |
|---|---|
| `docker-compose.lan.yml` | 新增 `edge` 服务（caddy:2-alpine），只发布 `80:80`，加入同一个 compose 项目 `multica` |
| `Caddyfile.lan` | `/ws`、`/ws/*` → `host.docker.internal:8080`（`flush_interval -1`）；其余 → `host.docker.internal:3100` |
| `.env` | `CORS_ALLOWED_ORIGINS=http://localhost:3100,http://192.168.20.9,http://192.168.4.36`；`MULTICA_APP_URL=http://192.168.20.9` |

后端/前端仍只绑 `127.0.0.1`，本机 daemon（`server_url=http://localhost:3100`）不受影响；对外只多开 80。

**起停命令**（三个 `-f` 必须一起用，否则会丢掉源码构建的 dev 镜像）：

```bash
cd /Users/shubinzhang/Documents/GitHub/multica
# 起/更新
docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml up -d
# 只停内网入口
docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml stop edge
# 改完 Caddyfile.lan 热加载
docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml exec edge caddy reload --config /etc/caddy/Caddyfile
```

**已验证**（2026-09-14）：`http://192.168.20.9/` → 200；`/api/workspaces` → 401（代理链通）；
`/health` → 200；`/ws?workspace_slug=phoebe-1yu1` 带合法 Origin → **101 Switching Protocols**；
带非法 Origin → **403**（白名单生效）。本机 `localhost:3100` 与 6 个 runtime 均不受影响。macOS 防火墙当前是关闭状态，不会拦。

**试跑期的注意点**

1. **登录** ✅ 已解决：邮件走 Lark SMTP，同事自己收验证码即可（§4.1）。`APP_ENV=production` 下固定
   `MULTICA_DEV_VERIFICATION_CODE` 无效（`isDevVerificationCode()` 在 production 直接返回 false），
   排查时可临时从库里取码：`docker exec multica-postgres-1 psql -U multica -d multica -tAc "select code from verification_code where email='…' and used=false and expires_at>now()"`。
2. ✅ 已收敛：`ALLOW_SIGNUP=false` + `ALLOWED_EMAIL_DOMAINS=guidefuture.com`（§4.1）。
3. 本机 IP 变更（换网/DHCP）→ 需同步改 `CORS_ALLOWED_ORIGINS` 与 `MULTICA_APP_URL` 并重启 backend，否则 WS 会 403。
4. 机器休眠/合盖即全体掉线；试跑期间请保持不休眠。
5. 同事必须通过 `http://192.168.20.9/`（白名单里的地址）访问，用别的 IP/主机名会被 403。
6. 回退：`stop edge` + 还原 `.env` 里的 `CORS_ALLOWED_ORIGINS` / `MULTICA_APP_URL` + 删除 `docker-compose.lan.yml` 与 `Caddyfile.lan`。

## 10. 配置审计（2026-09-14，逐项实测）

### 10.1 已实测通过、不用再动

| 项目 | 证据 |
|---|---|
| 浏览器完整登录链路 | 经内网入口 `/auth/verify-code` → 200 + `multica_auth`/`multica_csrf` cookie（HttpOnly、SameSite=Strict）→ `/api/workspaces`、`/api/me` 均 200 |
| cookie 不加 `Secure` 的原因 | `isSecureCookie()` 按 `FRONTEND_ORIGIN` 的 scheme 推导（`server/internal/auth/cookie.go:118`），当前是 `http://localhost:3100` → 内网 HTTP 才能登录。**换 HTTPS 时必须同步改它** |
| 同事的 daemon 安装地址 | `/api/config` → `daemon_server_url` / `daemon_app_url` = `http://192.168.20.9`（`MULTICA_APP_URL` 生效） |
| 密钥强度 | `JWT_SECRET` 64 hex / `POSTGRES_PASSWORD` 48 位，达标 |
| 附件落盘 | `LOCAL_UPLOAD_DIR` 未透传 → 代码默认 `./data/uploads`，与卷挂载 `/app/data/uploads` 一致 |
| Redis 的职责边界 | 只用于限流与 realtime 中继；`CHANNEL_WS_LEASE_BACKEND` 默认 `postgres`，因此**不需要** `CHANNEL_WS_LEASE_NAMESPACE` |
| 无外发遥测 | `POSTHOG_API_KEY` 未配 → 分析客户端 no-op（`.env` 里的 `ANALYTICS_DISABLED` 其实传不进容器，但结论相同） |

### 10.2 建议尽快改（都是入口/运维配置，风险低）

> 进度：**4、5 已于 2026-09-15 完成**；1、2、3（Caddy 入口三处）与 6（第二套可信代理）待做。

1. **`/readyz` 在内网入口 404**：前端 rewrite 白名单只有 `/health`、`/api/*`、`/auth/*`、`/v1/*`、`/uploads/*`、`/ws`、`/docs`
   （`apps/web/next.config.ts:51`），没有 `/readyz`。后端直连返回 `{"status":"ok","checks":{"db":"ok","migrations":"ok"}}`，
   而官方文档要求「用 `/readyz` 验证，不要用 `/health`」→ 入口要放它过去。
2. **`/auth/*` 经前端导致限流按前端容器 IP 计数**：实测伪造 `X-Forwarded-For: 203.0.113.77` 后，
   限流键仍是 `mul:ratelimit:auth:send-code:192.168.148.4`；而让反代**直连后端**时键变成 `…:203.0.113.77`。
   即 `RATE_LIMIT_TRUSTED_PROXIES` 现在其实没起作用（Next.js rewrite 不转发 XFF），全团队共用一个桶，
   所以才把阈值放宽到 30/120。把 `/auth/*` 直连后端后，可信代理解析（`middleware/ratelimit.go:91`，从右往左取
   最右的不可信跳，防伪造）才生效，阈值可以收回合理值。
3. **`/api/daemon/ws` 直连后端：由"必要"降级为"可选（对齐上游拓扑）"**。上游本来就是**两个入口**——helm chart 给
   `api.*` 一个 ingress 直连 `backend:8080`、给 web 一个 ingress 连 `frontend:3000`（`deploy/helm/multica/templates/ingress.yaml`），
   文档让 daemon `--server-url http://api.multica.dev.lan`；而我们单入口（`daemon_server_url=http://192.168.20.9`）把 daemon 的
   WS/心跳/claim 全压在前端上，所以当初判定要改。**2026-09-15 实测推翻了"前端代理不可靠"这个前提**：
   - **25 分钟原始 WS 探针**（复刻 daemon 的读纪律：60s 滚动读超时、收 ping 回 pong）：`backend:8080` 与 `frontend:3100` 两条路径
     各收到 27 个服务端 ping，间隔**全部精确 54.0s**，**零 deadline 触发**。前端这条跳没有丢帧、没有增抖。
   - **复用长连接的 HTTP 探针**（正是 daemon 的用法，10s 超时）对比 `localhost:3100` / `localhost:8080` / 内网入口：
     19 分钟各 230 次请求，**零超时零错误**，最慢 70ms。新建连接版本 20 分钟各 599 次，p50 前端 14ms vs 后端 7ms
     （即这跳大约 +7ms，对 15s 心跳/3min claim 无意义）。
   - 日志里那 463 次 `read tcp …127.0.0.1:3100: i/o timeout` 和 421 次 `close 1006` **全部属于 daemon v0.4.42**。
     今天 09:44 二进制自更新到 **v0.4.43** 后，当前 run（09:44→14:38，5 小时）**只剩 12 次失败**，时间戳全部落在我重启容器的那几刻
     （12:11–12:16、12:31；其中 `dial tcp [::1]:3100 connection refused` 就是前端容器当时不在）。→ 那是旧版客户端的 bug，
     **与经不经过前端无关**。
   结论：现在做它**没有可测收益**，理由只剩"少一跳、少一个进程参与、与上游拓扑一致"。归档到 P3，或与 P2 域名/TLS 阶段一起设计入口。
   顺带（仍然有效）：`docker logs multica-backend-1 | grep "daemon/ws"` 里 `status=0` 的那条就是 daemon 的长连接被接管。
4. ~~**容器日志没有轮转**：三个 compose 文件都没有 `logging` 段，json-file 无上限。~~ ✅ **2026-09-15 已加固**
   （写进未跟踪的 `docker-compose.lan.yml`）。**先纠正原判断**：本机其实一直在轮转——OrbStack 守护进程的内置默认值就是
   `json-file, max-size=20m, max-file=5`（`~/.orbstack/config/docker.json` 里没有这一项，是 daemon 硬编码；
   五个容器的 `HostConfig.LogConfig.Config` 实测都是 `map[max-file:5 max-size:20m]`）。所以问题不在本机，而在
   **换到普通 Linux Docker 主机的团队服务器时** json-file 默认不限制。现在 override 用 `x-logging` 锚点把同样的
   `20m×5` 显式声明给全部五个服务，数值与现状一致 → 换机器不会突然变行为（`compose config` 已验证五个服务都带上）。
   注意：本次 `up -d backend` **连带重建了 postgres**（它是 backend 的 `depends_on` 依赖，`logging` 变更改了它的配置哈希，
   31 秒后 healthy，数据在卷里无影响）；frontend / redis / edge 未动，它们的值本来就相同，会在下一次不带服务名的
   `up -d` 时顺带重建一次以承接声明（正常部署副作用）。
5. ~~**backend 没有 healthcheck**~~（postgres / redis 都有）✅ **2026-09-15 已加**：探针
   `wget -q -O /dev/null http://127.0.0.1:8080/readyz`，`interval 30s / timeout 5s / retries 3 / start_period 40s`，
   实测重建后 **10 秒内 healthy**（`FailingStreak: 0`，探针 exit 0）。两个实现细节：镜像里的 `wget` 是 **BusyBox** 版
   （`wget --version` 不认），**没有 `--spider`**，所以靠丢弃响应体+退出码；非 200 确实会让它失败
   （实测 `/no-such-route` → `HTTP/1.1 404`、`exit=1`），所以这不是"永远返回 0"的假探针。
   另注意 Docker 的 healthcheck **不会**因此自动重启容器（`restart: unless-stopped` 不响应 unhealthy），
   它的价值是让 `docker ps` 说真话、以及将来给监控/`depends_on: service_healthy` 用。
6. **两套可信代理别混**：`MULTICA_TRUSTED_PROXIES` 管 autopilot webhook 限流，`RATE_LIMIT_TRUSTED_PROXIES`
   管 `/auth/*`（`.env.example:100-113`）。做第 2、3 项时两个一起设成 Docker 网段。

### 10.3 上团队环境（域名 + TLS）前

7. `FRONTEND_ORIGIN` 改 `https://…`（决定 cookie `Secure`），同时改 `MULTICA_APP_URL`、`MULTICA_PUBLIC_URL`
   （autopilot webhook 的绝对地址）、`CORS_ALLOWED_ORIGINS`；分域时再加 `COOKIE_DOMAIN`。
   `MULTICA_APP_URL` 漏改 = 同事拿到的 daemon 安装命令指向 localhost。
8. **GitHub App 在纯内网做不了**：webhook 要公网可达。放到域名 + TLS 之后（`GITHUB_APP_*` + `GITHUB_WEBHOOK_SECRET`）。
9. 备份要覆盖三样：pgdata、附件卷 `multica_backend_uploads`、根 `.env`（含 `JWT_SECRET`）；
   目前只验证了 dump 能跑，**没做恢复演练** —— 上线前补一次。
10. 磁盘：本机可用 91GB；Docker 占 46GB 镜像（30GB 可回收）+ 17GB build cache。上常开服务器前清一遍。

### 10.4 可选 / 按需

11. `METRICS_ADDR`（compose 已透传）：开启后 HTTP 请求指标才开始累积，且后端只绑容器内回环，要用还得再映射端口。
12. `REALTIME_METRICS_TOKEN`：`/health/realtime` 在反代后一律 404（无 token），要监控实时链路才配。
13. **服务端 LLM 层 ✅ 已接 DeepSeek（2026-09-14）**：`MULTICA_LLM_BASE_URL=https://api.deepseek.com`、
    `MULTICA_LLM_DEFAULT_MODEL=deepseek-flash`、key 已写入 `.env`（key 从环境里的 `DEEPSEEK_API_KEY` 注入，
    没有经过对话）。重启后启动日志 `llm retry policy ... enabled=true`。
    实测（应用内真实调用，非仅 curl）：`POST /api/chat/sessions/{sid}/quick-actions/regenerate` → 202，
    4 秒内 `chat_message.quick_actions` 写入三条中文建议（“介绍你的能力/查看项目现状/说明任务所需信息”），
    后端无任何失败日志；另用应用同款 system prompt 直连 DeepSeek 预演标题生成 → `Android CI 构建偶发失败排查`，2.1s（应用侧超时 20s）。
    两个注意点：① 它只服务**两件事**（新会话自动标题、网页端聊天回复下方的追问按钮），
    而 **channel（Lark/Slack）会话没有按钮界面**（`service/chat_quick_actions.go:217`），所以你们主要用 Lark 的话
    只会看到自动标题生效；② 一旦配上，第一条用户消息原文与最多 6 条对话尾部会发往 DeepSeek，
    若要内容不出内网，应把它指向局域网内的 OpenAI 兼容端点（Ollama/vLLM）。
    另：`deepseek-flash` 是推理模型，标题生成会消耗 reasoning token，但实测 2s 内返回，远低于 20s 上限。
14. `DISABLE_WORKSPACE_CREATION`：团队运营策略——先建好 `土卫九`，再置 `true`，之后只能靠邀请加入。
    ❌ **2026-09-15 试过又回退了**：置 `true` 后**管理员自己也被锁**——该判定对**所有调用者**一律 403，
    **没有 role/owner 例外**（`handler/handler.go:72` 明说 "There is no role/owner exception because the repo has no
    platform-admin concept"；全仓库唯一判定点 `handler/workspace.go:213`；`multica workspace create` CLI 也打同一个端点）。
    实测（`true` 时）：`/api/config` → `workspace_creation_disabled: true`；**owner 身份** POST `/api/workspaces`
    直连后端与经内网入口都 `403 workspace creation is disabled for this instance`；侧边栏工作区下拉里
    "创建工作区"整项消失（`app-sidebar.tsx:679`）；新人无工作区时看到「工作区创建已禁用 / 请联系管理员邀请你加入工作区」。
    团队规模下"偶尔自己建一个工作区"的需求真实存在，而**上游没有管理员旁路**，所以**最终决定：保持关闭**（`.env` 该行为空）。
    如果以后还想要这个限制：**临时窗口流程已验证可用** —— 该行改回空 → `up -d backend`（约 8 秒）→ 建工作区 →
    改回 `true` → `up -d backend`；两个方向都实测生效（关掉后 `/api/config` 不再返回该字段，恢复后立刻又 403）。
    判定开关有没有生效有个**零副作用**的探针：POST `/api/workspaces` 带空 body `{}` —— 闸门开着是 403，
    关着则请求走到校验层返回 `400 name and slug are required`（不会真的建出工作区）。
    另一个不需要这个开关的替代做法：多一个管理员直接在 Settings → 成员 里邀成 owner/admin，与建工作区无关。
15. **邀请限流**（`handler/invitation_rate_limiter.go:30` 默认值；`server/cmd/server/router.go:442-446` 读环境变量）：
    三道独立闸门，**窗口固定、只有数量可配**：
    | 变量 | 默认 | 作用域 |
    |---|---|---|
    | `RATE_LIMIT_INVITATION_ACTOR_10M` | 10 | 每位邀请人 / 10 分钟滑动窗口 |
    | `RATE_LIMIT_INVITATION_WORKSPACE_24H` | 50 | 整个工作区所有管理员 / 24 小时 |
    | `RATE_LIMIT_INVITATION_RECIPIENT_24H` | 6 | 同一收件邮箱跨工作区 / 24 小时（按 email 的 SHA-256 摘要计数） |

    设 `0` = 关闭该道闸。取值解析用 `envNonNegativeInt`（`main.go:145`）：**空 = 用默认，非法/负数 = 告警并用默认**
    （所以写错值不会静默变成无限，会退回保守默认 —— 启动日志里能 grep 到 `invalid env var, using default`）。
    有 Redis 时装配的是 `NewRedisInvitationRateLimiters`（`router.go:508`），窗口跨副本共享；无 Redis 才是进程内存。
    ⚠️ 三个变量 **compose 没有透传**（`.env.example:488-490` 有文档、`docker-compose.selfhost.yml` 里没有），
    要调必须先补模板 —— 这是 §10.5 那条通用教训的又一例。

    ✅ **2026-09-15 已调高并实测到位**：`.env` 里 `ACTOR=100 / WORKSPACE=200 / RECIPIENT=10`，
    透传补在 `docker-compose.lan.yml` 的 `backend.environment`（保持"本地加固都放这个未跟踪覆盖"的一致做法）。
    实测：`compose config` 渲染出 `100/200/10`；容器内 `printenv` 三个值都在；无 `invalid env var` 告警；重建后 healthy。
    **为什么是这三个数**：默认 actor=10 意味着"一位管理员 10 分钟内最多发 10 封" —— 一次拉 20 人必然在中途 429。
    actor 是最先被撞到的闸，所以给到 100（一次坐满拉完还能再来一轮）；workspace 50→200 留增长与重发余量；
    recipient 6→10（它不是拉新人的瓶颈，只是容忍"撤销后重发"这类操作）。
    ⚠️ **验证的边界**：邀请限流只在 `CreateInvitation` 的**第 137 行**（`invitation.go`）触发，
    而"已是成员/角色非法/邮箱为空"都在它**之前**就 400/409 返回了，所以**没有零副作用的办法真正打到限流器**。
    要端到端证明只能发真实邀请（会建 pending 行 + 发邮件）。如需严格验证，用这个可控流程：
    把 actor 临时设成 `1` → `up -d backend` → 邀第 1 个新地址（应成功）→ 邀第 2 个（应 429）→ 撤销邀请 → 改回并重启。
    日常判断依据：**如果拉 20 人时没出现 429，就说明这三道闸已经生效在一个够用的值上**。
16. **Lark 机器人集成 ✅ 已启用并已安装（2026-09-14 18:05）**：`channel_installation` 1 行，`channel_type=feishu`、
    `region=lark`（国际版）、`status=active`、绑在 workspace `土卫九` 的 agent「MyNewProject 开发工程师」上，
    WS 长连接租约存活。已有 1 个用户绑定、1 个会话绑定、7 条入站去重、7 次任务投递，日志里每分钟都有
    `channel router: dispatch outcome channel_type=feishu outcome=ingested` 与 typing 反应 → 从 Lark 侧对话是通的。
    它走**长连接出站**（`connector=ws-long-conn`）而不是 webhook，所以**不需要公网地址，局域网阶段就能一直用**。
    ⚠️ `MULTICA_LARK_SECRET_KEY` 是加密各安装 AppSecret 的密钥，**必须随 `JWT_SECRET` 一起备份**，丢了所有
    已装机器人都解不开、只能重装。企业微信/钉钉/Slack/Telegram/Composio 保持关闭正常。
17. **查集成状态别看错表**（2026-09-14 踩过）：实际生效的安装/绑定在**通用表** `channel_installation` /
    `channel_user_binding` / `channel_chat_session_binding` / `channel_inbound_message_dedup` / `channel_task_delivery`；
    而 `lark_installation`、`lark_user_binding`、`lark_inbound_audit` 这一套是空的（历史路径）。只查 `lark_*` 会误判成
    "没有安装"；启动日志里的 `lark backfill: union_id pass ... attempted=0` 也只表示"本次没有需要回填的行"，≠ 没有安装。
18. `APP_ENV` 建议显式写 `production`：现在为空，靠 compose 的 `${APP_ENV:-production}` 兜底，换个跑法（直接跑二进制、
    或别的 compose）就可能意外落进 dev 模式。
19. **插件整个功能默认是关的**：`Settings → Plugins` 这个入口受功能开关 `plugins_v1` 控制（`featureflags/keys.go:25`，
    默认 false），而本次部署实测 `/api/config` 返回 `"plugins_v1": false` → 侧边栏**不显示 Plugins 入口**（`settings-page.tsx:206`），
    后端也同步关掉整套插件 API（`handler/plugin.go:21` `requirePluginsV1` → 503 "Plugin management is not enabled"）。
    开关来自 `FF_<KEY>` 环境变量或 `MULTICA_FEATURE_FLAGS_FILE` 指向的 YAML（`pkg/featureflag/config.go`），
    两者 compose **都没有透传**，`.env` 里 `MULTICA_FEATURE_FLAGS_FILE=` 也是空的。
    要开就在 `docker-compose.lan.yml` 的 `backend.environment` 里加一行 `FF_PLUGINS_V1: ${FF_PLUGINS_V1:-false}` 并 `up -d backend`。
    ✅ **2026-09-14 已开**（该 override 里加的是 `${FF_PLUGINS_V1:-true}`），`/api/config` 已返回 `"plugins_v1": true`，
    真浏览器登录态下 `Settings → Connections & extensions → Plugins` 入口出现、`?tab=plugins` 页面正常渲染
    （"Publish a Plugin" + "Installed"），`/api/workspaces/{id}/plugins` 与 `/plugins/packages` 均 200（从前是 503）。
    注意生效路径：compose 用的是显式 `environment:` 列表（没有 `env_file`），`${FF_PLUGINS_V1:-…}` 这一行**必须先存在**，
    之后 `.env` 或 shell 里的 `FF_PLUGINS_V1=false` 才能代入覆盖（已实测：`FF_PLUGINS_V1=false … compose config` → `"false"`）。
    没被列进模板的变量（如 `MULTICA_FEATURE_FLAGS_FILE`）无论写在哪都进不了容器。
20. **插件 surface（`MULTICA_PLUGIN_*`，Settings → Plugins）——不写插件就不用配**。`.env` 四个键目前全空，compose 已透传，
    留空是安全的：`secret`/surface/hook 三类能力 fail closed（`MULTICA_PLUGIN_SECRET_KEY` 缺失时
    `GetPluginSurfaceLaunch` 直接 503），`mpi_` Public API token 不受影响。
    只有"自己写插件、且插件带 UI 面板"时才需要 `MULTICA_PLUGIN_SURFACE_ORIGIN`：
    它是**给插件 UI 专用的第二个浏览器 origin**（`MULTICA_PLUGIN_SURFACE_ORIGIN`，如 `https://plugins.example.com`），
    必须与 app/API origin **Host 不同**（`plugin_surface.go:72` `pluginSurfaceOriginIsDedicated` 比对 `PublicURL`/`AppURL`/
    `AttachmentFrameAncestors`），反代要保留 `Host`；后端在该 host 上**只**服务 `/plugin-surfaces/*`，
    同一路径在普通 API origin 上被 `PluginSurfaceHostBoundary`（`plugin_surface.go:90`）判 404。
    两层保险：surface 文档带 no-store + CSP（`connect-src` 由安装 scope 决定，无 `net:` 就是 `'none'`），
    且请求带任何 `Cookie`/`Authorization` 就被 `ServePluginSurface` 判 400（`plugin_surface.go:225`）。
    ⚠️ 因此"同主机换个端口"**不成立**：cookie 不受端口约束，浏览器照样会把 `multica_auth` 发到 `:8081`，surface 必 400。
    局域网阶段要真做只能说客户端 hosts 加个独立名字（如 `192.168.20.9 multica-plugins.lan`）+ 一个 Caddy site 块，
    否则等域名 + TLS 时再一并做（本机不配 hosts 的话，这条现在无法给出可用形态）。

### 10.5 一条通用教训

**写进 `.env` 不等于进了容器。** 后端会读、但 `docker-compose.selfhost.yml` 没有 `VAR: ${VAR:-}` 透传的键有：
`ANALYTICS_DISABLED`、`ANALYTICS_ENVIRONMENT`、`POSTHOG_API_KEY`、`POSTHOG_HOST`、`CLOUDFRONT_PRIVATE_KEY_SECRET`、
`LOCAL_UPLOAD_DIR`、`MULTICA_DAEMON_ID`、`MULTICA_WORKSPACE_ID`、`MULTICA_LARK_WS_PROXY_URL`（当前都无害，
但改配置前先 `grep VAR docker-compose*.yml` 确认一次）。同理，`docker compose restart` 不会重新读 `.env`，必须 `up -d`。

---

## 11. 待办优先级（2026-09-15，按 收益 / 风险 / 影响 排序）

**排序口径**（不是按"技术上多优雅"，而是按"不做会怎样"）：
`影响面`（多少人的日常被影响）× `不做的后果严重度` × `何时会爆` ÷ `实施成本与实施风险`。
"何时会爆"权重最高：**现在就会咬人 > 上量后咬人 > 只在扩容/迁移时咬人**。

### P0 —— 上量前必须做（20 人正在进场，这几条不做就会出事）

| # | 事项 | 收益 | 不做的风险 | 影响面 | 成本/实施风险 | 依赖 |
|---|---|---|---|---|---|---|
| P0-1 | **备份自动化 + 恢复演练**（§10.3#9） | 唯一的"全丢"防线 | **极高**：数据只在这台笔记本上，目前**没有定时备份**（只手动 dump 过一次）；要覆盖 pgdata + 附件卷 `multica_backend_uploads` + 根 `.env`。`.env` 丢了 → `JWT_SECRET` 变 → **全员被登出**；`MULTICA_LARK_SECRET_KEY` 丢 → 已装机器人全废、只能重装 | 全员 + 全部 60 个项目的历史 | 低（脚本 + 定时任务，源侧只读）；**恢复演练要另找一台机器实跑一次**，不能只看 dump 成功 | 无 |
| P0-2 | **把"三个 `-f` 一起用"固化**（今天已真实踩过一次）｜**必须在 P0-4 之前做** | 消除静默降级 | **中高**：只带两个 `-f`（=`make selfhost-build` 的行为；`SELF_HOSTING.md:462/494` 教的甚至只有**一个**）会**无声丢掉** backend 的 7 个 env（插件开关 + 6 个限流）+ logging + healthcheck —— 且**没有任何提示**（实测：多出来的服务不会被删，只打印一句 orphan warning，内网入口照旧 200）。**P0-4 做完后这条后果更重**：`/auth/*` 路由挂在 edge 上仍在生效，但 `RATE_LIMIT_TRUSTED_PROXIES` 被丢 → 限流退回"全团队一个桶 + 上游默认 5/分" | 全员，且无感 | 极低（`stack-check.sh` + `lan-up.sh`，10 分钟）；**不要**改成 `.env` 里设 `COMPOSE_FILE`（会劫持 `make db-down` 等裸调用） | 无 |
| P0-3 | ~~**邀请限流：compose 透传 `RATE_LIMIT_INVITATION_*` 并调阈值**（§10.4#15）~~ ✅ **2026-09-15 已完成** | 20 人能一次拉完 | — | 管理员 + 每位新同事的入场体验 | 已完成：actor 10→100、workspace 50→200、recipient 6→10，透传补进 `docker-compose.lan.yml` | 无 |
| P0-4 | **Caddy 入口修正 ②**：`/auth/*` 直连后端（§10.2#2） | 限流从"全团队共用一个桶"变回"每人一份"；登录不再经过前端容器 | 中高：不做 → 键恒为前端容器 IP（实测 `mul:ratelimit:auth:send-code:192.168.148.9`），**任何人**把 `/auth/send-code` 打到 30 次/分，接下来 1 分钟内**全团队**都收 429；阈值放宽到 30/120 只是掩盖 | 全员（登录是必经路径） | 低（`Caddyfile.lan` 一处 + `caddy reload` 热加载，可回退）；**做完可把 `RATE_LIMIT_AUTH` 从 30 收回上游默认 5** | 无 |

### P1 —— 现在就能做、收益明确但只影响内部

| # | 事项 | 收益 | 不做的风险 | 影响面 | 成本 |
|---|---|---|---|---|---|
| P1-1 | `MULTICA_TRUSTED_PROXIES` 一并设成 Docker 网段（§10.2#6） | autopilot webhook 限流按真实来源计数 | 低-中：只有用到 webhook 触发 autopilot 才暴露 | 用 autopilot webhook 的人 | 极低 |
| P1-2 | `/readyz` 放行进内网入口（§10.2#1）**优先级已下调** | 可从外部/监控系统探活 | **低**：新加的 backend healthcheck 是**容器内** `wget 127.0.0.1:8080/readyz`，不经过入口，所以探活已经能用；这条只为"从外面看" | 运维视角 | 低 |
| P1-3 | 给 healthcheck 接一个"有人看"的东西（告警/每日巡检） | unhealthy 能被告知 | 低-中：现在 health 只让 `docker ps` 说真话，进程卡死仍无人知 | 运维 | 低 |
| P1-4 | 清 Docker 磁盘（46GB 镜像其中 ~30GB 可回收 + 17GB build cache） | 腾出空间 | 低：本机可用 91GB，暂时够；但会持续涨 | 本机 | 极低（`docker builder prune` / `image prune`） |
| P1-5 | `APP_ENV` 显式写 `production`（§10.4#18） | 换跑法不会意外落进 dev | 低：compose 已用 `${APP_ENV:-production}` 兜底 | 无（防御性） | 极低 |

### P2 —— 依赖基础设施，是"真正上团队"的总闸门

| # | 事项 | 说明 |
|---|---|---|
| P2-1 | **常开服务器 + 域名 + TLS** | §10.3 全部条目的前置。笔记本合盖即离线，这是团队部署的真正起点 |
| P2-2 | 随之切换 `FRONTEND_ORIGIN` / `MULTICA_APP_URL` / `MULTICA_PUBLIC_URL` / `CORS_ALLOWED_ORIGINS` → `https://…`，分域时加 `COOKIE_DOMAIN`（§10.3#7） | `MULTICA_APP_URL` 漏改 = 同事拿到的 daemon 安装命令指向 localhost |
| P2-3 | **GitHub App**（§10.3#8） | 60 个项目要挂仓库/PR 卡片才需要；**必须公网 webhook**，纯内网做不了。对 Android 团队的实际价值高，但要等 P2-1 |
| P2-4 | 上服务器前先清镜像/build cache、搬数据、按 P0-1 的演练流程恢复一次 | 迁移当天才第一次恢复 = 没有备份 |
| P2-5 | **在新服务器上给这个栈一个独立 compose 项目名**（实测覆盖文件可以改 `name:`） | 新环境零成本，做完"少带 `-f`"就从**静默降级**变成**当场端口冲突报错**（另一份栈起不来，也不会动到真正在跑的容器）。现在做要迁卷名（`multica_pgdata` 等换前缀 = 换库），风险不成比例，所以留到这里 |

### P3 —— 按需，可以一直不做

- **`/api/daemon/ws`（乃至 `/api/daemon/*`）由 Caddy 直连后端**：**2026-09-15 从 P0-4 降级**——实测前端这条跳对 WS 完全无损害
  （25 分钟探针、27 个 ping 全部精确 54.0s、零超时；长连接 HTTP 19 分钟零超时），旧日志里的周期性 `i/o timeout` 是 daemon v0.4.42
  自身的问题，v0.4.43 已消失（详见 §10.2#3）。现在只有"少一跳、与上游两入口拓扑一致"这点收益。若做，注意：本机 daemon 走
  `localhost:3100` **根本不经过 edge**，改 Caddyfile 对它无效，要改 `~/.multica/config.json` 的 `server_url`；
  同事的 daemon 走 `MULTICA_APP_URL=http://192.168.20.9`（`/api/config` 的 `daemon_server_url`），改 edge 才对他们生效。
- **渠道**：钉钉 / Telegram / Composio（**都要先改 compose**，目前三个变量连模板都没有）；Slack / 企业微信（只改 `.env`，compose 已透传）。用哪个才做哪个。
- **插件 surface**（`MULTICA_PLUGIN_SECRET_KEY` + `SURFACE_ORIGIN` + `API_URL`）：只有**自己写插件且带 UI 面板**才需要。
- **指标**：`METRICS_ADDR` + `REALTIME_METRICS_TOKEN`（要接监控才配）。
- **`ALLOWED_EMAILS`**：要精确到人时再加（现在域名白名单够）。
- **对象存储 S3**：**已决定暂缓**（2026-09-15，卷目前 0 文件；一旦开始传图，切换就要写迁移脚本）。
- **`DISABLE_WORKSPACE_CREATION`**：**已决定保持关闭**（上游无管理员旁路，会把管理员一起锁）。
- **Google 登录 / PostHog 埋点 / billing**：不需要或刻意保持关闭。

### 11.1 执行顺序（把上面四档串成一条线，2026-09-15 复核）

排法三条原则：**① 先做"保护其他改动"的**（否则后面做的会被一次误操作抹掉）；**② 再做"现在就会咬人"的**；
**③ 最后做"只在扩容/迁移时咬人"的**。每条都写清 收益 / 影响 / 不做的风险 / 做的风险。

**第 1 组 · 今天就能做完（分钟级，互不依赖，都不改应用行为）**

1. **固化"三个 `-f` 一起用"**（P0-2）——**必须先于第 2 项做**，理由见下
   ｜收益：后面所有加固都不会被无声抹掉 ｜影响：全员、无感 ｜不做：**中高**——一次按文档执行（`SELF_HOSTING.md:462/494`
   教的甚至只有**一个** `-f`；`make selfhost` / `make selfhost-build` 是两个）就把插件开关、认证限流、healthcheck 全丢，**且毫无提示** ｜做：极低。
   **实测过的降级机制**（2026-09-15，隔离环境复现）：少带 `-f` 时 Compose **不会删掉**多余服务，只打印一句
   `Found orphan containers …`，于是 `edge`/`redis` 照旧运行、内网入口照旧 200 —— 事故因此完全不外显。
   精确 diff（`docker compose config` 只比键名）：**backend 丢 7 个 env**（`FF_PLUGINS_V1`、`RATE_LIMIT_AUTH`、
   `RATE_LIMIT_AUTH_VERIFY`、`RATE_LIMIT_INVITATION_{ACTOR_10M,WORKSPACE_24H,RECIPIENT_24H}`、`RATE_LIMIT_TRUSTED_PROXIES`）
   **+ logging + healthcheck**，frontend/postgres 丢 logging，edge/redis 变 orphan。这 7 个变量在 base 模板里
   **一个都没透传**（`grep RATE_LIMIT\|FF_ docker-compose.selfhost.yml` 为空），所以**只有 compose 文件能给它们**，
   `.env` 里写多少都没用。
   ⚠️ **第 2 项做完后这条的后果会变严重**：`Caddyfile.lan` 挂在 edge 上、edge 作为 orphan 仍在跑 → `/auth/*` 直连后端的路由
   **依然生效**，但 `RATE_LIMIT_TRUSTED_PROXIES` 一起被丢 → 限流键退回 **edge 容器 IP**（又变回全团队一个桶），
   同时阈值退回上游默认 **5**（放宽后的 30 也没了）→ 全团队每分钟只剩 5 次登录机会。所以护栏先行。
   **方案（全部只动未跟踪文件）**：① `stack-check.sh` —— 一条命令回答"现在带几个 `-f` + 7 个关键 env 在不在"，降级 `exit 1`
   （现在的检查方式 `docker inspect …config_files` 太长，没人会敲）；② `lan-up.sh` —— 正常起停/更新都走它，内部三个 `-f`，
   `up` 完自动跑 `stack-check`；③ 定时跑 `stack-check`（launchd 每小时 / 或本 harness 的定时任务），把"哪天有人才发现"
   变成"一小时内发现"。
   ❌ **不要用 `COMPOSE_FILE` 写进根 `.env`**（实测它确实生效：`docker compose config` 会认它、显式 `-f` 会完全覆盖它）。
   这个仓库里有一堆**裸 `docker compose`** 调用：`Makefile:237/240`（`db-up`/`db-down`）、`scripts/ensure-postgres.sh`、
   `scripts/drop-database.sh`。一旦设上，`make db-down` 的爆炸半径会从"只删 postgres 一个容器"扩到"删掉整个 5 容器栈"
   （dev 的 `docker-compose.yml` 与自托管栈**共用项目名 `multica`**，本来就已经相互影响，这不是把它扩大的理由）。
   ⭐ **结构性做法，留到第 9 项（新服务器）再做**：给这个栈一个**独立 compose 项目名**（实测覆盖文件可以改 `name:`）。
   届时按文档只带一个 `-f` 起的那份栈会**因端口被占用当场报错**（`Bind for 127.0.0.1:8080 failed: port is already allocated`），
   "静默降级"就变成"当场失败"。现在做要迁移卷名（`multica_pgdata` 等换前缀 = 等于换库），风险不成比例。
2. **② `/auth/*` 直连后端**（P0-4），同一处顺手把 `RATE_LIMIT_AUTH` 从 30 收回上游默认 5、`RATE_LIMIT_AUTH_VERIFY` 收回 20
   ｜收益：限流从"全团队共用一个桶"变回"每人一份"；登录/退出不再经过前端容器 ｜影响：全员（登录是必经路径） ｜不做：**任何人**把 `/auth/send-code` 打到 30 次/分，全团队接下来 1 分钟都收 429 ｜做：低（`Caddyfile.lan` 一处 + `caddy reload` 热加载，秒级回退；唯一失败模式是登不上）。
3. **`MULTICA_TRUSTED_PROXIES` 设成 Docker 网段**（P1-1，`192.168.148.0/24`）
   ｜收益：与第 2 项同源（可信代理解析），一次改完 ｜影响：只有用 autopilot webhook 的人 ｜不做：低-中，只在 webhook 触发时暴露 ｜做：极低。
4. **清 Docker 磁盘**（P1-4，`docker builder prune` + `image prune`，约 30GB 镜像 + 17GB cache）
   ｜收益：给第 5 项的备份腾出空间 ｜影响：本机 ｜不做：低——但会在"要恢复时"变成没有空间 ｜做：极低。
5. **`APP_ENV` 显式写 `production`**（P1-5）｜收益：换跑法不会意外落进 dev ｜影响：无（纯防御） ｜不做：低 ｜做：极低。

**第 2 组 · 本周（这一组都是"不做会全丢"或"坏了不会被告知"）**

6. **备份自动化：脚本 + 定时任务**（P0-1 前半），覆盖 pgdata + 附件卷 `multica_backend_uploads` + 根 `.env`
   ｜收益：唯一的"全丢"防线 ｜影响：全员 + 60 个项目的全部历史 ｜不做：**极高**——数据只在这台笔记本上，`.env` 丢了等于 `JWT_SECRET` 变、全员被登出，`MULTICA_LARK_SECRET_KEY` 丢了机器人全废 ｜做：低（源侧只读）。
7. **恢复演练**（P0-1 后半）：在另一台机器/另一个卷上按流程真恢复一次，不能只看 dump 成功
   ｜收益：把"有备份"这句话变成事实 ｜影响：同第 6 项 ｜不做：高——迁移当天才第一次恢复 = 其实没有备份 ｜做：中（要资源 + 一段时间窗口）。
8. **给 healthcheck 接一个"有人看"的巡检/告警**（P1-3）｜收益：unhealthy 能被知道 ｜影响：运维 ｜不做：低-中，进程卡死时 `docker ps` 仍说 Up 且无人知 ｜做：低。

**第 3 组 · 上量前（互相有依赖，是"真正上团队"的总闸门）**

9. **常开服务器 + 域名 + TLS**（P2-1）→ 10. **配置切 `https://`**（P2-2：`FRONTEND_ORIGIN` / `MULTICA_APP_URL` / `MULTICA_PUBLIC_URL` / `CORS_ALLOWED_ORIGINS`，分域加 `COOKIE_DOMAIN`）→ 11. **迁移演练**（P2-4，按第 7 项流程搬一次）→ 12. **GitHub App**（P2-3，要公网 webhook，只能排在这里）。
   这一组里 **10 漏改 `MULTICA_APP_URL` = 同事拿到的 daemon 安装命令指向 localhost**，是整组里最容易漏且后果最隐蔽的一条。

**第 4 组 · 按需 / 条件触发**

13. **`/readyz` 放行进内网入口**（P1-2）：此前判定是"不改"（上游有意只代理 `/health`），只有你想从外部/监控探活时才加；**可以并入第 2 项同一次 `Caddyfile.lan` 编辑**，那时成本≈0。
14. **③ daemon 直连后端**（已从 P0 降级，见 P3 首条）：并入第 9 项设计入口时一起做，那时本来就要重写 Caddyfile + 分域。
15. 渠道扩展 / 插件 surface / 指标 / `ALLOWED_EMAILS` / S3：用哪个做哪个（P3）。

## 12. 上游自托管文档通读（2026-09-17，`SELF_HOSTING.md` + `SELF_HOSTING_ADVANCED.md`）

逐条与我们的现场配置对照，只记**会影响决策**的部分。

### 12.1 我们的入口写法是上游 sanctioned 的 ✅

`Caddyfile.lan` 与 `SELF_HOSTING_ADVANCED.md:398-413` 的 **"单域名 Caddy（推荐）"** 布局**逐字一致**：`@…_ws path /ws /ws/*`
→ `reverse_proxy …:8080` + `flush_interval -1`，其余请求 → 前端。文档专门解释了这两个写法为什么不能省（`:436-443`）：
`path /ws /ws/*` 而非 `/ws*`（Caddy 的 `*` 无路径段边界，会把合法的 `ws` 之外 slug 如 `/ws-foo` 一起吃掉）；`flush_interval -1`
（否则帧卡在 Caddy 默认 flush 窗口里，表现为"评论要刷新才出现"）。**结论：这块不用改，第 9 项换域名时沿用即可。**

### 12.2 ⚠️ 匿名遥测默认开启，且正在成功上报 —— **需用户决策**

- `SELF_HOSTING.md:15-23`：自托管 API **每个 UTC 日**向**编译进二进制、配置无法改址**的
  `https://telemetry.multica.ai/v1/telemetry/events` 发一份**部署级匿名快照**。内容只有版本号、
  工作区/成员/agent/daemon 的**分桶计数**、24h 内 run 的起止失败计数；明确不含姓名邮箱域名、IP、业务 ID、
  主机/设备信息、仓库信息、模型/插件信息、prompt 或输出、评论/聊天、路径、token/成本、凭据、报错栈、日志。
  也**不用于**计费、授权、认证或安全判定。
- 关闭方式：API server 上设 **`DO_NOT_TRACK=1`（或 `true`）**。`ANALYTICS_DISABLED` 只管 PostHog，**关不掉这个**。
- **实测（本机）**：容器内 `DO_NOT_TRACK=` 为空 → 开启；`instance_telemetry_state` 有 1 行，
  `last_successful_day=2026-09-17`、`attempt_count=0`、`instance_id=d3ed5303-…` → **确实在成功上报，今天已报过**。
  PostHog 侧确认为关（`.env` 里 `POSTHOG_API_KEY=` 为空，且 posthog 变量**根本没进容器**）。
- **改法成本极低**：`DO_NOT_TRACK: ${DO_NOT_TRACK:-}` **base 模板已透传**（`docker-compose.selfhost.yml:114`）
  → 只在 `.env` 加 `DO_NOT_TRACK=1` + 重建/重启 backend 即可，**不用碰覆盖文件**（与 GitHub App 四个变量同类）。
- **另一个必须记住的坑**（与 P0-1 恢复演练直接相关）：文档要求**克隆生产库做演练前先清 `instance_telemetry_state`**
  （或设 `DO_NOT_TRACK=true`），否则两份部署**共用同一个 `instance_id`**，同一天的快照会**互相静默去重**。
  → 第 7 项恢复演练的检查清单里要加这一条。

### 12.3 官方 LAN 指引 vs 我们的现状

- 官方（`ADVANCED:554-568`）说 LAN 访问要把 `FRONTEND_ORIGIN` / `CORS_ALLOWED_ORIGINS` 设成**你访问用的那个 origin**。
  我们 `FRONTEND_ORIGIN` 仍是 `http://localhost:3100`（**`.env` 里的裸 `localhost`**），只有 `CORS_ALLOWED_ORIGINS`
  列了 `localhost:3100,192.168.20.9,192.168.4.36` —— 而 `.9` 已经不在任何网卡上（见 12.7）。
- 官方还强调（`ADVANCED:415`、`:588`）：后端**独立校验 WS 的 `Origin`**，默认白名单**只有 `localhost`**，不在名单里
  → 升级被 `403` 拒、浏览器 console 循环 `disconnected, reconnecting in 3s`，而 **HTTP 照常工作**（因为对页面而言是同源）。
- **但文档没写、我们实测到的更重要的一条**：`checkOrigin`（`realtime/hub.go:158`）有**同源放行**分支
  （Origin 的 host == 请求的 `Host` 即放行；另有"X-Forwarded-Host 匹配 + 来源是可信代理"分支）。因为浏览器访问的是
  edge 的 `http://192.168.20.131`，请求到达后端时 `Host` 就是 `.131`，与 Origin 同 host → **直接放行，不看白名单**。
  实测 `.131` / `.4.36` / `.9` / `localhost:3100` 全 101，`evil.example` → **403**，无 Origin → 101。
  → **两条结论**：① 本机 IP 变了**不会**打断实时（我此前"白名单过期会断实时"的担心被推翻）；
  ② 一旦第 9 项换成**域名**访问，`Host` 变成域名、Origin 也是域名，**仍同源**，所以即便白名单没跟上实时也不会断；
  真正会断的是"域名 + 另一个来源跨域调 API"这类场景 —— 那时才必须把域名写进 `CORS_ALLOWED_ORIGINS`。
- 官方给的另一个解法（`ADVANCED:570-590`）：**没有反代时**可以 build 期把 `NEXT_PUBLIC_WS_URL` 烧进 web 镜像
  （必须配 `selfhost.build.yml` 重建，**只写 `environment:` 对预构建镜像无效**）。我们已有反代，**不适用**。
- 关于 `NEXT_PUBLIC_WS_URL` 的默认值：留空时浏览器就是打**页面自己的 origin 的 `/ws`**（`ADVANCED:522-536`），
  这正是我们现在的形态 —— 所以**不需要任何前端镜像改动**。

### 12.4 运维事实（读文档即可确认，无需动作）

- **进程内 rollup 调度器**：每个 backend 副本每 30s tick、争抢 5 分钟窗口的计划，唯一键 `(job_name, scope_kind, scope_id, plan_time)`
  保证**只有一个副本赢** → 多副本安全。文档给的巡检 SQL 我们实测过：最近 2 小时 `rollup_task_usage_hourly` **23 条 SUCCESS**
  （≈每 5 分钟一条），`rows_affected=0`，**无需运维**。Redis 已开（sharded relay）→ 上述多副本前提成立。
- **Health Check 三兄弟**（`ADVANCED:595-610`）：`/health`（liveness，只证进程活着）、`/readyz`（**含 db + migrations**，
  未就绪时 5xx）、`/healthz`（**= `/readyz`**，同一个响应）。文档明确：探活用 `/health`，判断"能不能服务"用 `/readyz`。
  **实测**：内网入口 `http://<ip>/health` → 200，`/readyz` 与 `/healthz` → **404**（这正是 P1-2 的"有意只代理 `/health`"）；
  直连 `127.0.0.1:8080/readyz` → 200。**结论不变**：backend 自身的 healthcheck 是容器内 `wget 127.0.0.1:8080/readyz`，
  不经过入口，所以**探活已经有效**；`/readyz` 放行只为"从外面看"。
- **Metrics / pprof**：`METRICS_ADDR` 默认**空**（我们也是空）；**pprof 固定在容器内回环 `127.0.0.1:6060`**，
  要抓得 `docker compose … exec backend` 进去（文档示例 `ADVANCED:656-658`），**不要**为了抓 profile 把它暴露到宿主。
- **端口默认值**：`FRONTEND_PORT` 默认 **3000**（`ADVANCED:159`，容器内恒为 3000，只有宿主侧可变），`BACKEND_PORT` 默认 **8080**
  （走别名链 `BACKEND_PORT → API_PORT → SERVER_PORT → PORT → 8080`）。**我们前端宿主端口是 3100**（因为 3000 另有用途），
  这是刻意的，不用跟文档示例对齐；但**看文档/抄文档时注意别人的 3000 不等于我们的 3100**。
- **端口变量的优先级陷阱**（`ADVANCED:165-172`）：Compose 里**调用环境压过 `.env`**（`PORT=9100 docker compose … up -d`），
  而 `make` **反过来**（它 `include` env 文件，文件值压过环境变量，命令行赋值压过两者）。
  → **改 env 文件是唯一处处一致的做法** —— 这条支持我们"只用 `.env` + 固定三个 `-f`"的路线。

### 12.5 文档同时也是 P0-2 的又一处证据

**升级章节（`SELF_HOSTING.md:472-473` 与 `:504-505`、`ADVANCED:665-666`）教的都是只带一个 `-f`**：
```
docker compose -f docker-compose.selfhost.yml pull
docker compose -f docker-compose.selfhost.yml up -d
```
—— 与 `SELF_HOSTING.md:462/494` 的安装段落同样**只带一个** `-f`。只有"GHCR tag 还没发布"的**兜底**分支才写两个 `-f`
（`:477`/`:670`）。**所以"按官方文档升级一次"= 触发 §11.1 第 1 项描述的那次静默降级**（丢 7 个 env + healthcheck + logging）。
文档里唯一沾到我们三个 `-f` 写法的地方是 `ADVANCED:583/590` 的源码构建兜底。→ **`stack-check.sh` + `lan-up.sh` 的护栏是必须的，不是可选的。**

### 12.6 对第 1 组待办的修订

- **第 5 项 `APP_ENV=production` 已完成 / 无需动作**：`.env` 里是空值，但 base 模板写死 `${APP_ENV:-production}`
  （`docker-compose.selfhost.yml:103`，`:-` 对空值也生效）→ 实测容器内 `APP_ENV=production`。任何 compose 跑法都拿到 production。
- **新增一条决策项**：`DO_NOT_TRACK`（见 12.2）—— 加一行 `.env` 即可关闭，是否关闭由用户定。
- **第 7 项（恢复演练）检查清单加一条**：演练库起 backend 前先 `DELETE FROM instance_telemetry_state;`（或设 `DO_NOT_TRACK=true`），
  避免两份部署共用 `instance_id` 互相去重。
- **新增一条现场问题（见 12.7）**：主机 IP 从 `.9` 变成 **`.131`**，而 `.env` 里三处地址仍指向 `.9`。

### 12.7 ⚠️ 现场问题：主机内网 IP 已漂移，`.env` 还在用旧地址（**建议提到第 1 组**）

- **实测（2026-09-17 17:50 CST）**：`en0=192.168.20.131`、`en8=192.168.4.36`；**`.9` 已不在任何网卡上**。
- 但 `.env` 里三处地址仍是 `.9`：`MULTICA_APP_URL=http://192.168.20.9`（容器内实测同一个值）、
  `CORS_ALLOWED_ORIGINS=…,192.168.20.9,…`，以及 `FRONTEND_ORIGIN` 的裸 `localhost`。
- **后果（都已确认，不是推测）**：
  1. **同事手上的 `.9` 打不开**；
  2. 新同事拿到的 **daemon 安装命令指向一个不存在的地址**（命令来自 `/api/config` 的 `daemon_server_url`，
     值就是 `MULTICA_APP_URL`）→ 他的 daemon 永远连不上、主机永远"离线"；
  3. **实时更新不受影响**（同源放行，见 12.3）—— 这一条**不是**问题，只是容易被误判成问题。
- **已排除"容器降级"**：5 个容器全部带**三个 `-f`** 标签，7 个关键 env 全在（`printenv` 实测 `FF_PLUGINS_V1=true`、
  `RATE_LIMIT_TRUSTED_PROXIES=192.168.148.0/24` 等在位），frontend/backend 是 09-16 正确重建的（`Up 4 hours`）。
- **根因性质**：DHCP 租约变化。`en0` 的地址**会再变** → 单纯把 `.9` 改成 `.131` 只是把下一次故障延后。
  真正的解法是**给这台机器固定内网地址**（DHCP 静态绑定 / 手动配置），否则每换一次地址都要重做一次本节的改动。
- **改法**：`.env` 里 `MULTICA_APP_URL` 与 `CORS_ALLOWED_ORIGINS` 改成实际地址；`FRONTEND_ORIGIN` 一并显式写成
  `http://<lan-ip>:3100`（官方 LAN 指引就是这么要求的，见 12.3）；然后**重建** backend + frontend
  （`restart` 不重读 `.env`）。这与第 9 项的 `https://` 切换是同一批变量，届时一起改即可。
