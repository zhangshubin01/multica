# Multica 最佳配置总结

> 综合仓库内全部文档得出：官方自托管指南（`SELF_HOSTING*.md`、`.env.example`、`CLI_AND_DAEMON.md`、`DESKTOP_CLIENT_CONFIG.md`）、官方文档站（`apps/docs/content/docs/*.mdx`）、以及本仓库团队实测文档（`TEAM_DEPLOY_PLAN.md`、`TEAM_USAGE_STANDARDS.md`、`ORBSTACK_CONFIG.md`、`MIGRATION_TO_NEW_HOST.md`）。
> 标注 **[实测]** 的条目是本机真跑过的结论；**[官方]** 来自上游文档；**[团队约定]** 是团队规范。结论聚焦 20 人 Android 团队、单机自托管场景，同样适用于更小规模。

---

## 0. 结论速览

1. **部署形态**：单机 Docker Compose + same-origin（单域）Caddy 反代，是上游 sanctioned 且本机验证过的最简正确形态。
2. **`.env` 必填**：`JWT_SECRET`、`POSTGRES_PASSWORD`、`FRONTEND_ORIGIN`、`CORS_ALLOWED_ORIGINS`、`MULTICA_APP_URL`、SMTP 全套 —— 缺任何一个都出问题，且三个"公开地址"必须**一致**。
3. **开关裁决**：开 Redis（noeviction）、开限流可信代理、`DO_NOT_TRACK=1` 关遥测、`ALLOW_SIGNUP=false` 收紧注册、`DISABLE_WORKSPACE_CREATION` **保持关闭**（无管理员旁路，锁自己）。
4. **客户端**：CLI 一切走 `multica setup self-host` + `daemon autostart enable`；桌面端只认 `~/.multica/desktop.json` 的 `apiUrl`，写域名不写 IP。
5. **智能体**：Access 默认收紧（Only me）、Instructions 八项必写、单 agent 并发 2、单 daemon 并发 4、MCP 少挂 skills 优先、`custom_env` 只放低权限 key。
6. **备份三件套**：Postgres dump（`-Fc`）+ 附件卷 + 根目录 `.env`（含全部密钥）；恢复顺序先灌库再起栈。

---

## 1. 部署形态与架构

| 项 | 推荐 | 来源 |
|---|---|---|
| 拓扑 | 单机 Docker Compose（`docker-compose.selfhost.yml`），postgres + backend + frontend + edge（Caddy） | `SELF_HOSTING.md` |
| 域名 | **same-origin 单域**（如 `https://multica.example.com`）；分域只在"同父域且父域下所有主机同一受信方"时可用，否则安全前提不成立 | `SELF_HOSTING_ADVANCED.md` → Reverse Proxy |
| LAN 入口 | 独立 `docker-compose.lan.yml`（edge 暴露 80），内网先用 IP，上域名+TLS 后切换 | `TEAM_DEPLOY_PLAN.md` §9 |
| 版本 | `MULTICA_IMAGE_TAG` 钉住具体版本，禁用 `latest` **[实测]** | `TEAM_DEPLOY_PLAN.md` §10.1 |
| 起停 | **三个 `-f` 一起给**（selfhost + selfhost.build + lan），少给会静默丢 7 个 env + logging + healthcheck；改 `.env` 后必须 `up -d` 重建，`restart` 不重读 `.env` **[实测]** | `TEAM_DEPLOY_PLAN.md` §10.5 |

---

## 2. 服务端 `.env` 最佳配置

### 2.1 必填项（不设则不安全或功能失效）

| 变量 | 推荐值 | 原因 | 来源 |
|---|---|---|---|
| `JWT_SECRET` | `openssl rand -hex 32` | 签名 auth/daemon/附件 JWT；compose 缺它拒启，production 拒绝弱默认值；**迁移时必须原样保留**，否则全员登出 | `SELF_HOSTING_ADVANCED.md`；`TEAM_DEPLOY_PLAN.md` §10.1 |
| `POSTGRES_PASSWORD` | 随机强密码 | compose 初始化 PG 用；首次初始化后不可改角色 | `SELF_HOSTING_ADVANCED.md` |
| `FRONTEND_ORIGIN` | 公开地址，如 `https://multica.example.com` | CORS/WS Origin 白名单默认只放行 localhost；还决定 cookie 是否带 `Secure`（HTTP 阶段必须写 `http://…` 才能登录，换 HTTPS 必须同步改）**[实测]** | `SELF_HOSTING_ADVANCED.md`；`TEAM_DEPLOY_PLAN.md` §4 |
| `CORS_ALLOWED_ORIGINS` | 与 `FRONTEND_ORIGIN` 一致 | 漏改 = WS 升级 403、实时功能静默失效 | `TEAM_DEPLOY_PLAN.md` §4 |
| `MULTICA_APP_URL` | 公开地址 | 经 `/api/config` 的 `daemon_server_url` 下发给同事的安装命令；漏改 = 大家拿到指向 localhost 的命令，后果最隐蔽 | `TEAM_DEPLOY_PLAN.md` §4、§12.7 |
| SMTP 全套 | 见 §2.4 | 无邮件则验证码只打印日志，不能算生产登录 | `SELF_HOSTING_ADVANCED.md` → Email |

### 2.2 数据库

| 变量 | 推荐值 | 说明 | 来源 |
|---|---|---|---|
| PG 版本 | PostgreSQL 17（`pgcrypto`+`pg_trgm` 必须；`pg_bigm` 可选、显著改善 CJK 搜索） | 无需 pgvector/pg_cron（历史残留） | `SELF_HOSTING_ADVANCED.md` |
| `DATABASE_MAX_CONNS`/`MIN_CONNS` | 保持默认 25/5 | `pod 数 × MAX_CONNS` 须远低于 PG `max_connections`；前置 PgBouncer 可上调 | `.env.example` → Database |
| `DATABASE_SEARCH_WORK_MEM_MB` | 默认 64 | 搜索事务 work_mem 上限，1-64 可降 | `.env.example` |
| `DATABASE_REPLICA_URL` | 可选；必须带 `target_session_attrs=read-only` | 副本失败自动回退主库；副本读无陈旧上界 | `.env.example` |
| `MULTICA_DATABASE_STARTUP_TIMEOUT`/`CONNECT_TIMEOUT` | 3m / 5s（默认） | 迁移+启动重试预算 | `.env.example` |

### 2.3 Server 端口

- `PORT` 是**该改的那个**：裸跑时是监听端口，compose 里是宿主发布端口；容器内部恒为 8080，改值无需重编译。别名优先级 `BACKEND_PORT` > `API_PORT` > `SERVER_PORT` > `PORT`。
- `FRONTEND_PORT=3100`（宿主侧，容器内恒 3000）**[实测]**。
- `METRICS_ADDR` 默认关；开启只绑 `127.0.0.1:9090`，绝不进公网入口。
- compose 端口只绑 127.0.0.1，不要改成 0.0.0.0（Docker 绕过宿主防火墙），公网一律走反代 TLS。

### 2.4 邮件 SMTP（认证必需，Lark 已实测跑通）

| 变量 | 推荐值 | 来源 |
|---|---|---|
| `SMTP_HOST` | `smtp.larksuite.com`（国际版）/ `smtp.feishu.cn`（中国版）/ 内网中继。**一旦非空就只走 SMTP、不再把验证码打日志**，主机与密码必须同时配好 | `.env.example`；`TEAM_DEPLOY_PLAN.md` §4.1 |
| `SMTP_PORT` | 465（自动 implicit TLS）或 587+STARTTLS | `.env.example` |
| `SMTP_TLS` | 465 配 `implicit`；其余默认 starttls | `TEAM_DEPLOY_PLAN.md` §4.1 **[实测]** |
| `SMTP_USERNAME` | = 认证账号（SaaS 中继必须；留空=匿名，供应商默认拒绝） | `.env.example` |
| `SMTP_FROM_EMAIL` | 与认证账号一致（`noreply@你的域`）；不设回落 `noreply@multica.ai` 会被 SPF/DKIM 拒 | `.env.example` |
| `SMTP_PASSWORD` | 应用专用密码（不是登录密码）；535=密码类型错、550=From 与账号不一致 | `TEAM_DEPLOY_PLAN.md` §4.1 |
| `SMTP_TLS_INSECURE` | 仅私有 CA/自签才 `true`；Lark 证书可信，**不需要** | 同上 |
| `SMTP_EHLO_NAME` | 严格中继（Google Workspace）拒默认 EHLO 时设真实 FQDN | `.env.example` |

### 2.5 注册与建工作区控制

| 变量 | 推荐 | 说明 | 来源 |
|---|---|---|---|
| `ALLOW_SIGNUP` | 生产 `false` | 新账号仅白名单/邀请；已建账号不受影响 | `.env.example` |
| `ALLOWED_EMAIL_DOMAINS` | 团队域（如 `guidefuture.com`，已实测生效） | 域名整段匹配，**子域不算**；老用户永远放行 | `TEAM_DEPLOY_PLAN.md` §4.1 **[实测]** |
| `DISABLE_WORKSPACE_CREATION` | **保持关闭（空）** | `true` 对**所有调用者一律 403、无 role/owner 例外**，管理员自己也被锁。官方文档推荐序列即"建好共享工作区后置 true"，但上游没有平台管理员旁路，团队实测后回退 **[实测]** | `SELF_HOSTING_ADVANCED.md` → Signup Controls；`TEAM_DEPLOY_PLAN.md` §10.4 |
| `WORKSPACE_CREATION_ALLOWED_EMAILS` | 想限制时用它替代上面：列少数组长邮箱，其余人 403 | 评估顺序：`DISABLE=true` 全禁 → 列表空全放行 → 命中放行 → 否则 403；UI 不隐藏入口 | `SELF_HOSTING_ADVANCED.md` |
| `MULTICA_DEV_VERIFICATION_CODE` | **公网实例禁止设置** | 固定验证码=任意邮箱可登录；`APP_ENV=production` 下本就被忽略 | `.env.example` |

### 2.6 存储 / Cookie / LLM / 遥测

| 变量 | 推荐 | 来源 |
|---|---|---|
| 附件存储 | 默认本地卷（`LOCAL_UPLOAD_DIR`）；上传上限 **100 MiB**，反代需放行 ≥101 MiB 请求体。S3 可选（`S3_BUCKET`+`S3_REGION`，`AWS_ENDPOINT_URL` 兼容 MinIO/R2）；私桶内网端点用 `ATTACHMENT_DOWNLOAD_MODE=proxy` | `.env.example`；`SELF_HOSTING_ADVANCED.md` → File Storage |
| `COOKIE_DOMAIN` | **留空**（单域/LAN）；分域同父域才设最窄父域（如 `.agent.example.com`）；**禁止 IP 字面量**；改后需清 cookie 重登 | `SELF_HOSTING_ADVANCED.md` → Cookies |
| `MULTICA_LLM_*` | 内容不出网 → 保持双空（默认全关，功能优雅降级）；要标题/追问按钮则配 `MULTICA_LLM_BASE_URL` + `MULTICA_LLM_DEFAULT_MODEL`（实测 DeepSeek 可用） | `.env.example`；`TEAM_DEPLOY_PLAN.md` §10.1 **[实测]** |
| `DO_NOT_TRACK` | **`1`** | 匿名遥测默认开启，每天向 `telemetry.multica.ai` 发部署快照；`ANALYTICS_DISABLED` 关不掉它 | `SELF_HOSTING.md`；`TEAM_DEPLOY_PLAN.md` §12.2 **[实测]** |
| `APP_ENV` | 显式写 `production` | 当前靠 compose 兜底，换跑法会意外落进 dev | `TEAM_DEPLOY_PLAN.md` §10.4 |

### 2.7 Redis 与限流

| 变量 | 推荐 | 说明 | 来源 |
|---|---|---|---|
| `REDIS_URL` | 开：`redis://redis:6379/0`（Linux）/ 团队实测用 `redis://host.docker.internal:6379/0` | 支撑 auth 限流、实时 fan-out、频道租约；多副本生产**必须**配 | `.env.example`；`TEAM_DEPLOY_PLAN.md` §3 **[实测]** |
| Redis 策略 | `--maxmemory-policy noeviction` **必配** | 租约键被驱逐会断连 | `TEAM_DEPLOY_PLAN.md` §3 |
| `RATE_LIMIT_TRUSTED_PROXIES` | Docker 网段（如 `192.168.148.0/24`，换机必查 `docker network inspect`） | 不设则全团队共享反代 IP 一个桶（5 次/分钟）**[实测]**；与 `MULTICA_TRUSTED_PROXIES`（webhook 限流）是两套 | `TEAM_DEPLOY_PLAN.md` §10.2 |
| `RATE_LIMIT_AUTH`/`RATE_LIMIT_AUTH_VERIFY` | 上游默认 5/20；团队临时放宽 30/120，**待 `/auth/*` 直连后端落地后收回** | 限流键在前端容器 IP 上，经前端代理时全团队一桶 | `TEAM_DEPLOY_PLAN.md` §10.2、§11 P0-4 |
| `RATE_LIMIT_INVITATION_ACTOR_10M` 等 | 团队放宽到 100/200/10（默认 10/50/6） | 默认 10 = 管理员 10 分钟最多 10 封邀请，拉 20 人必然 429。⚠️ 该组变量 **compose 未透传**，需先补 override 再写 `.env` **[实测]** | `TEAM_DEPLOY_PLAN.md` §10.4#15 |

---

## 3. 反向代理与入口

| 项 | 推荐 | 来源 |
|---|---|---|
| 拓扑 | same-origin Caddy：`@ws path /ws /ws/*` → `127.0.0.1:8080`（`flush_interval -1`），其余 → 前端 | `SELF_HOSTING_ADVANCED.md` → Caddy **[实测]** |
| `/ws` 匹配 | 写 `/ws /ws/*` **不能写 `/ws*`**（会误吞 `/ws-foo` 类 workspace slug） | 同上 |
| `flush_interval -1` | 必须；缺了表现为"评论要刷新才出现" | 同上 |
| Nginx 等价 | `/ws` 块加 `Upgrade`/`Connection` 头 + `proxy_read_timeout 86400`；**Next.js rewrite 不转发 Upgrade 握手**，必须直连后端 | `SELF_HOSTING_ADVANCED.md` → Nginx |
| LAN 直接 IP | `FRONTEND_ORIGIN`+`CORS_ALLOWED_ORIGINS` 写 `http://<LAN-IP>:3000` 并重启；一份清单同时管 HTTP CORS 与 WS Origin | `SELF_HOSTING_ADVANCED.md` |
| `/auth/*` 直连后端 | 团队 P0：否则限流键恒为前端容器 IP，任何人打到 30 次/分全团队 1 分钟 429 **[实测]** | `TEAM_DEPLOY_PLAN.md` §11 P0-4 |
| `/health` vs `/readyz` | 内网入口 `/readyz` 404 是上游有意（rewrite 白名单只有 `/health`）；探活用 `/health` | `TEAM_DEPLOY_PLAN.md` §12.4 |

---

## 4. CLI / daemon / 桌面端

### 4.1 CLI 与 daemon

| 项 | 推荐 | 来源 |
|---|---|---|
| 安装 | macOS/Linux 首选 Homebrew（`brew install multica-ai/tap/multica`）；无 brew 用 GitHub Releases 单二进制 | `CLI_INSTALL.md` |
| 自托管接入 | 一条命令：`multica setup self-host`（可带 `--port/--frontend-port/--server-url/--app-url`）；手动等价物 `multica config set server_url/app_url` | `CLI_AND_DAEMON.md` |
| 生产 TLS | `server_url https://api.example.com` + `app_url https://app.example.com`（same-origin 时两个同域） | 同上 |
| 无头登录 | `multica login --token <mul_...>`；用 `--token=` 空值交互输入避免 token 进 shell history | 同上 |
| 自启动 | `multica daemon autostart enable` 是**唯一**注册途径（`daemon start` 自己不注册）；不注册 = 重启后 runtime 离线、排队任务无人认领 | 同上 |
| 环境变量 | 自启只快照 `PATH`；其余用 `multica config set` 持久化，直接 export 会静默丢失 | 同上 |
| Profiles | 一台机器多 daemon（生产/staging）用 `--profile`；状态目录 `~/.multica/profiles/<name>/` | 同上 |
| 关键默认 | 心跳 15s、poll 30s、daemon 全局并发 20、workspace 根 `~/multica_workspaces`、GC 默认开 | 同上 |
| 支持的 agent | 至少要检测到一个受支持 CLI（claude/codex/opencode/cursor-agent 等 30+）daemon 才工作；装完 `daemon stop && daemon start` 重扫 | 同上 |
| Local Environments | 仓库根 `.multica/environments.toml`：`ready` 指向 `/health` 探针、`stop` 写清理命令、`url` 用 `$MULTICA_PORT` 别写死 | 同上 |

### 4.2 桌面客户端

| 项 | 推荐 | 来源 |
|---|---|---|
| 配置入口 | 唯一：`~/.multica/desktop.json`（Windows `%USERPROFILE%\.multica\desktop.json`），客户端无 UI 切换入口 | `DESKTOP_CLIENT_CONFIG.md` |
| 内容 | `{"schemaVersion":1,"apiUrl":"http://<内网地址>","wsUrl":"ws://<内网地址>/ws","appUrl":"http://<内网地址>"}`；上 TLS 后换 `https://` | 同上 |
| 地址 | **写内网域名而不是 IP**（主机 DHCP 会漂移）；`.env` 的 `MULTICA_APP_URL` 必须同步 | 同上 |
| 写入方法 | Windows 必须 `[System.IO.File]::WriteAllText`（`>`/`Out-File` 写 UTF-16、记事本补 `.txt` 都会失效） | 同上 |
| 生效 | 只在启动时读一次，改完 ⌘Q 完全退出重开 | 同上 |
| 验证三处 | ① `desktop.json` 内容 ② `~/.multica/profiles/` 出现 `desktop-<host>/` ③ 设置里 `serverUrl` 显示目标地址 | 同上 |
| 回滚 | 删文件 + 重启即回官方云 | 同上 |

---

## 5. 产品层：智能体 / 运行时 / 项目

### 5.1 智能体配置（团队场景）

| 项 | 推荐 | 来源 |
|---|---|---|
| 数量与分层 | 约每项目 1 个（70 项目 ≈ 70），命名 `<App 代号>-<职能>`；另配少量 `android-<能力>` 跨项目 agent 与 `<专项>-squad` | `TEAM_USAGE_STANDARDS.md` §4.1 |
| Access | 默认 **Only me**；项目 agent 用 Entire workspace 或 Specific people（该 App 3-4 人）；owner/admin 也不能绕过别人的 Access 运行私有 agent | `agents.mdx`；`TEAM_USAGE_STANDARDS.md` §4.2 |
| Instructions | **八项必写**：①仓库与分支 ②模块结构 ③构建/测试命令 ④代码风格与静态检查 ⑤硬约束（不碰 release/签名/主版本升级）⑥产出与回写（分支命名、MR 带 `PHOE-<n>`、`Closes` 收尾）⑦停止条件 ⑧不要碰的东西。**description 只做展示、不进 prompt** | `TEAM_USAGE_STANDARDS.md` §4.3 |
| 并发 | 单 agent 默认 6 → 团队约定 **2**（大项目 1）；单 daemon 默认 20 → 团队约定 **4**。Android 一次构建吃 3-4 GB 内存，高并发互相 OOM | `agents-create.mdx`；`TEAM_USAGE_STANDARDS.md` §3.4 |
| 模型 | 先选 runtime（机器+工具）再选该工具支持的模型；模型名不跨工具共享；留空用工具默认。DeepSeek Harness 用 `provider/model` 全 id | `providers.mdx` |
| MCP | 逐 agent 显式绑定，**别给 70 个 agent 各堆一套**（工具 schema 每轮常驻）；共享能力优先做 skill（按 run 下发、可版本化）。优先级：agent 自有 > 工作区共享 > 机器继承 | `agents-create.mdx`；`TEAM_USAGE_STANDARDS.md` §4.5 |
| `custom_env` | **明文存服务端库**，只放低权限可轮换 key；`PATH`/`HOME`/`MULTICA_*` 不可覆盖；Custom arguments 禁止放密钥（argv 对 `ps` 可见） | `agents-create.mdx` |
| Token | run 由 daemon 注入 `mat_` 临时 token（绑定用户+工作区+agent+run，最长 24h）；用户 `mul_` PAT 绝不交给 agent | `auth-tokens.mdx` |

### 5.2 运行时

| 项 | 推荐 | 来源 |
|---|---|---|
| 安装 | 本机装好工具并**单独登录跑通**；低于最低版本不注册（Claude Code 2.0.0、Codex 0.100.0、OpenCode 1.1.54）；装完 `daemon restart` | `install-agent-runtime.mdx` |
| 命名 | `<姓名拼音>-<机型>-<工具>`；runtime 默认 private | `TEAM_USAGE_STANDARDS.md` §3.2 |
| 看门狗 | `MULTICA_AGENT_IDLE_WATCHDOG=2h`；`MULTICA_AGENT_TIMEOUT=0`（默认无绝对时限）；掉线 3 分钟标 offline，进行中 run 有 3h reconnect grace | `environment-variables.mdx` |
| 安全 | **run 以 daemon OS 用户全权限执行，无文件系统沙箱**；隔离放 daemon 外面：专用 Unix 用户 → 容器 → 虚拟机；用 per-purpose deploy key 不用个人 SSH key | `security-model.mdx` |

### 5.3 项目与工作区

| 项 | 推荐 | 来源 |
|---|---|---|
| 资源挂载 | workspace 仓库清单**永远进每个 run 上下文**：只放公共库（≤5 个），禁止塞 70 个业务仓库；每个 project 挂 1-3 个仓库，`ref` 钉住 | `TEAM_USAGE_STANDARDS.md` §5.2 |
| 工作目录 | 默认 worktree 模式；`local_directory + in_place` 同目录串行仅调试用 | 同上 |
| Git 凭据 | 每台 daemon 机器必须有 GitLab 凭据（group deploy key / 可写 group token / SSH key 三选一），agent checkout/推分支/开 MR 全用宿主 git 凭据 | `TEAM_USAGE_STANDARDS.md` §3.3 |
| 分支保护 | 生产分支开受保护分支 + MR 门 + CODEOWNERS；agent 不得有直推 main 凭据；agent 写的 MR 不自审自合 | 同上 §6.4 |
| issue 纪律 | 每个 issue 落 project、带 label、有 assignee；`backlog` 是停车场（不触发 agent）；MR 标题/分支带 `PHOE-<n>` | 同上 §7 |

---

## 6. 集成配置

| 平台 | 要点 | 来源 |
|---|---|---|
| 通用 | 一 Bot 只绑一个 agent；群聊只有显式 @ 触发；每个平台自托管配 base64-32 字节加密 key（`openssl rand -base64 32`），丢失必须重连 | `channels.mdx` |
| **Lark/飞书** | `MULTICA_LARK_SECRET_KEY` 是加密各安装 AppSecret 的密钥，**必须随 `JWT_SECRET` 一起备份**，丢了已装机器人全废；走 WS 长连接**不需要公网**；订阅方式必须是长连接 + 订阅 `im.message.receive_v1` | `lark-bot-integration.mdx`；`TEAM_DEPLOY_PLAN.md` §10.4 **[实测]** |
| Slack | Socket Mode 免公网；两个 token（`xoxb-` + `xapp-`）；附件加 `files:read` scope 后**必须重装 app** | `slack-bot-integration.mdx` |
| 钉钉 | Stream Mode；社区维护无官方 SLA；加 `qyapi_chat_manage` 权限才显示 Bot 名 | `dingtalk-bot-integration.mdx` |
| 微信（个人版） | 扫码必须是操作者本人微信，消息按该用户权限执行；仅私聊；无需公网回调 | `weixin-bot-integration.mdx` |
| Telegram | 服务器需出站到 `api.telegram.org`（走 `HTTPS_PROXY`）；保留 Group Privacy 开启；getUpdates 前清既有 webhook | `telegram-bot-integration.mdx` |
| GitHub App | 六个 env 全配才能新连接；权限全部 **Read-only**；**Webhook secret ≠ OAuth Client secret**，不一致就是 401；纯内网做不了（webhook 需公网） | `github-integration.mdx`；`TEAM_DEPLOY_PLAN.md` §10.3 |
| 自托管 Git | 仅自托管 Multica 可用：`MULTICA_VCS_INTEGRATION_ENABLED=true` + `MULTICA_VCS_SECRET_KEY`；token 只读（GitLab `read_api`）；私有 CA 只能进后端信任链（`SSL_CERT_DIR`），无跳过校验选项；agent 开 PR 用宿主 git 凭据，无需 provider 配置 | `vcs-integration.mdx` |

---

## 7. 宿主层（OrbStack）与运维

| 项 | 要点 | 来源 |
|---|---|---|
| 引擎配置位置 | OrbStack 只读 `~/.orbstack/config/docker.json`，`~/.docker/daemon.json` **被完全忽略**——迁移时只拷后者 = 镜像加速一行都不生效 | `ORBSTACK_CONFIG.md` §3 **[实测]** |
| 镜像加速 | `registry-mirrors: ["https://docker.m.daocloud.io", ...]`；改完**必须重启 OrbStack**，`docker info` 确认 | 同上 |
| 日志轮转 | OrbStack 内置 `json-file 20m×5`（侥幸默认）；Linux 默认不限 → `x-logging` 锚点必须跟走 | 同上 §5 |
| `host.docker.internal` | OrbStack 解析到 `0.250.250.254`；**Linux 不成立**，改容器服务名 | `MIGRATION_TO_NEW_HOST.md` |
| Docker 网段 | 动态分配，换机必查 `docker network inspect multica_default -f '{{(index .IPAM.Config 0).Subnet}}'` 后回填 `RATE_LIMIT_TRUSTED_PROXIES` | `ORBSTACK_CONFIG.md` §8 |
| 备份对象 | **三样**：Postgres dump（`pg_dump -Fc`，不要管道 gzip——失败退出码仍是 0）+ 附件卷 `multica_backend_uploads` + 根目录 `.env`（含全部密钥，`.gitignore` 排除，永不随 clone 来） | `MIGRATION_TO_NEW_HOST.md`；`TEAM_DEPLOY_PLAN.md` §6 |
| 恢复顺序 | 先只 `up -d postgres`（建空库）→ `pg_restore --no-owner --no-acl` → 还原附件卷 → 最后 `up -d --build` 起全栈；反过来会满屏 `relation already exists` | `MIGRATION_TO_NEW_HOST.md` §5 **[实测]** |
| daemon 身份 | 迁机推荐拷 `~/.multica/config.json` + `daemon.id`（runtime 原地复用、agent 绑定不断；**源机 daemon 必须先停**，同 id 双机在线互抢任务） | 同上 §8 |
| 红线 | **永远不要 `docker compose down -v`**（`-v` 删 pgdata + uploads，全书唯一不可逆操作）；克隆生产库做演练先清 `instance_telemetry_state` 或 `DO_NOT_TRACK=1` | `SELF_HOSTING.md`；`MIGRATION_TO_NEW_HOST.md` |

---

## 8. 与上游默认值的分歧裁决（团队实测）

| 项 | 上游默认 | 本机裁决 | 理由 |
|---|---|---|---|
| `DISABLE_WORKSPACE_CREATION` | 官方推荐 bootstrap 后置 `true` | **保持关闭**，用 `WORKSPACE_CREATION_ALLOWED_EMAILS` 替代 | 无 role/owner 旁路，管理员自己也被锁（`handler/handler.go:72` 明说无 platform-admin concept）**[实测]** |
| `RATE_LIMIT_AUTH*` | 5 / 20 | 临时放宽 30 / 120，P0-4 落地后收回 | 经前端代理时限流键恒为前端容器 IP，全团队一桶 |
| 邀请限流三变量 | 10 / 50 / 6 | 放宽 100 / 200 / 10 | 默认 actor=10 拉 20 人必 429；且 compose 未透传 |
| daemon 并发 | 20 | 团队约定 4；单 agent 2 | Android 构建内存占用 |
| `REDIS_URL` | `redis://redis:6379/0` | 实测用 `redis://host.docker.internal:6379/0` | 少给 `-f` 起栈时后端仍连得上；Linux 换回服务名 |

---

## 附录：最小生产基线（可直抄，密钥全部占位）

```dotenv
# 必填（密钥用命令生成，绝不写进文档）
JWT_SECRET=<openssl rand -hex 32>
POSTGRES_PASSWORD=<随机强密码>
FRONTEND_ORIGIN=https://multica.example.com
CORS_ALLOWED_ORIGINS=https://multica.example.com
MULTICA_APP_URL=https://multica.example.com

# 邮件（认证必需）
SMTP_HOST=smtp.larksuite.com
SMTP_PORT=465
SMTP_TLS=implicit
SMTP_USERNAME=noreply@你的域
SMTP_PASSWORD=<应用专用密码>
SMTP_FROM_EMAIL=noreply@你的域

# 安全收紧
ALLOW_SIGNUP=false
ALLOWED_EMAIL_DOMAINS=你的域
APP_ENV=production
DO_NOT_TRACK=1

# Redis + 限流
REDIS_URL=redis://redis:6379/0
RATE_LIMIT_TRUSTED_PROXIES=<Docker 网段，docker network inspect 取>
MULTICA_TRUSTED_PROXIES=<同上>

# 版本钉住
MULTICA_IMAGE_TAG=v0.x.y
```

保持留空即可：`MULTICA_LLM_*`（内容不出网）、Google OAuth、S3、`COOKIE_DOMAIN`（单域）、`MULTICA_DEV_VERIFICATION_CODE`（生产禁止）、`DISABLE_WORKSPACE_CREATION`（无管理员旁路）。
