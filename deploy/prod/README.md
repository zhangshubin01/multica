# 生产自托管部署（Mac mini）

目标：新主机上 **clone → 拷一个密钥文件 → 跑一条命令**，不再逐项配 `.env`。

## 仓库里已经带上的东西

| 文件 | 作用 |
|---|---|
| `docker-compose.lan.yml` | 本地覆盖层：redis 发布、限流（auth/invite）、功能开关（`FF_PLUGINS_V1` / `FF_DOCS_V1` 默认开）、edge 与 backend 健康检查、端口绑定 |
| `docker-compose.selfhost.build.yml` | 从源码构建，镜像固定为 `multica-backend:dev` / `multica-web:dev` |
| `deploy/prod/env.template` | 栈会读取的**全部** 107 个键。密钥位是 `@SECRET:NAME@`，主机相关值是 `@URL@` / `@WS@` / `@CIDR@`（可嵌在组合值里，如 `@URL@,http://localhost:3100`），`DATABASE_URL` 由 `POSTGRES_*` 现场拼 |
| `deploy/prod/bootstrap.sh` | 渲染 `.env` → 构建 → 起栈 → 自检；带 `--dry-run` / `--render-only` / `--extract-secrets` |

密钥值不在仓库里，也永远不应进去：这个 fork 是公开仓库。

## 三步上线

**1. 在现役主机上导出密钥（一次）**

```bash
./deploy/prod/bootstrap.sh --extract-secrets
# → ~/.multica-secrets/secrets.env   25 个键，mode 600，仓库外
```

只导出确实有值的键；源码主机上本来就空着的集成（Resend、Google、AWS/CloudFront、DingTalk、Slack、企微、Telegram、GitHub App 私钥等）会以 `NAME=` 空值写入并带一行注释说明。空值是有意义的取值——表示那套集成不启用。

**2. 拷到新主机**

```bash
scp ~/.multica-secrets/secrets.env mini:~/.multica-secrets/secrets.env
git clone -b Gif_android <repo> ~/multica && cd ~/multica
```

**3. 在 Mac mini 上部署**

```bash
./deploy/prod/bootstrap.sh
```

它会：探测本机 LAN 地址 → 按该地址填 `MULTICA_APP_URL` / `MULTICA_PUBLIC_URL` / `FRONTEND_ORIGIN` / `CORS_ALLOWED_ORIGINS` / `NEXT_PUBLIC_API_URL` / `NEXT_PUBLIC_WS_URL` / `MULTICA_DAEMON_SERVER_URL` / `*_REDIRECT_URI` / `*_CALLBACK_BASE_URL`、`MULTICA_TRUSTED_PROXIES`、`RATE_LIMIT_TRUSTED_PROXIES`（新网段 `/24`）→ 用密钥文件填 24 个密钥位 → 写 `.env`（600，已有则先备份，需 `--force`）→ `docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml -f docker-compose.lan.yml up -d --build` → 等 `/healthz` → 打印迁移版本、开关、前端与 edge 探针。

其它用法：

```bash
./deploy/prod/bootstrap.sh --dry-run                 # 只打印渲染结果，不写文件
./deploy/prod/bootstrap.sh --render-only             # 只写 .env，不碰 docker
./deploy/prod/bootstrap.sh --host 192.168.20.180 --app-url https://multica.example.com
./deploy/prod/bootstrap.sh --allow-missing-secrets   # 全新空装：缺的密钥留空，POSTGRES_PASSWORD/JWT_SECRET 现场生成
```

## 多套环境：生产与测试的口径

模板给的是**新主机默认值**：入口 `:80`（由 `docker-compose.lan.yml` 的 `edge` 提供）、`FRONTEND_PORT=3100`。现役生产主机有一层刻意的覆盖，两边不要互相纠偏：

| 项 | 生产（现役主机） | 测试 / 新主机 |
|---|---|---|
| 入口 | `http://<host>:8090`（宿主 Caddy → `127.0.0.1:3000`），`:80` 的 edge 作为第二条等价入口 | `http://<test-ip>`（`:80`，edge） |
| `FRONTEND_PORT` | `3000` | `3100`（模板默认） |
| 渲染方式 | 手工维护的 `.env`（历史较久，含 SMTP/Lark 等实际值） | `./deploy/prod/bootstrap.sh --host <test-ip>` |

- **生产为什么不跟模板迁到 `:80`**：四个对外 URL 已被 GitLab webhook（`gitlab-hooks-sync.mjs` 跟随 `MULTICA_PUBLIC_URL`）、各机器的 daemon `app_url`、书签钉住；`MULTICA_PLUGIN_SURFACE_ORIGIN` 又是独立 host + `:8090`。切一次 URL 的收益只是“数字与模板一致”，成本是全链路改址。edge 走 compose 服务名（`frontend:3000` / `backend:8080`），所以 `FRONTEND_PORT` 是 3000 还是 3100 不影响 `:80` 入口。
- **插件表面要单独填**：`MULTICA_PLUGIN_SURFACE_ORIGIN` 留空 = 表面关闭（插件能装，页面打不开）。它必须是**与 app/API 不同的 host**，并另起一个反代站点把该 host 转给 `backend:8080`。
- **两套栈同机时要先改端口**：`edge` 现在写死 `80:80`，同机起第二套要参数化成 `${EDGE_PORT:-80}:80`，并用 `-p multica-test` 隔离 compose 项目与卷。
- 真要统一到 `:80` 的一次性清单：`.env` 四个 URL 去掉 `:8090` → 重建 backend（环境变量只在创建时读）→ GitLab hook 5 分钟内自动跟随 → 宿主 Caddy 的 `:8090` 改成 302 到 `:80`（插件表面那个站点原地保留，靠 Host 区分）→ 通知各人改书签与 `multica config` 的 app_url。回滚 = 还原 `.env` 备份并重建 backend。

## 必须人工处理的三件事（无法随仓库走）

1. **密钥文件**——`secrets.env` 只能线下传递（scp / AirDrop / 1Password）。
2. **数据库数据**——全新主机是从空库开始并自动跑迁移；要沿用现有数据，先 `pg_dump -Fc` 再在目标机 `pg_restore`，口径与回滚见仓库根 `MIGRATION_TO_NEW_HOST.md`。
3. **宿主 daemon 身份**——`multica daemon` 的 PAT、以及每台机器上各 runtime 的登录（claude / codex / cursor …）属于机器身份，重新在新主机注册一次：`multica daemon start`。

## 校验清单

脚本跑完会打印这几项，也可以手跑：

```bash
curl -s --noproxy '*' 127.0.0.1:8080/healthz
docker exec multica-postgres-1 psql -U multica -d multica -tAc 'select max(version) from schema_migrations'
curl -s --noproxy '*' 127.0.0.1:8080/api/config | grep -o '"docs_v1":[a-z]*'
curl -s --noproxy '*' -o /dev/null -w '%{http_code}\n' 127.0.0.1:3100/
curl -s --noproxy '*' -o /dev/null -w '%{http_code}\n' 127.0.0.1/
docker logs multica-backend-1 2>&1 | grep -iE "error|panic|fatal" | grep -v 41050
```

预期：`status: ok` + `db/migrations: ok`；迁移到最新版本；`docs_v1:true`；两个 200；日志只有飞书 41050 这类已知噪声。

## 细节与坑

- **换 IP 不用重建前端镜像**：`NEXT_PUBLIC_API_URL` 在运行期由 `apps/web/config/runtime-urls.ts` 读取，不是构建期烘进 bundle 的（`Dockerfile.web` 只烘了 `REMOTE_API_URL=http://backend:8080` 和 `NEXT_PUBLIC_APP_VERSION`）。但 `.env` 改过之后 **backend 容器必须重建**（容器只在创建时读环境变量），`docker compose … up -d` 会自己做。
- **`RATE_LIMIT_TRUSTED_PROXIES` 必须覆盖**：`docker-compose.lan.yml` 里的默认值绑死在作者当前网段，脚本按新主机地址重算 `/24`。
- **镜像名**：源码构建固定为 `multica-*:dev`；`deploy/prod/env.template` 里的 `MULTICA_IMAGE_TAG` / `MULTICA_BACKEND_IMAGE` / `MULTICA_WEB_IMAGE` 只在改用官方 GHCR 镜像时生效。
- **Docs（`docs_v1`）默认开**，关掉用 `FF_DOCS_V1=false docker compose … up -d backend`；无需改文件。
- **`.env` 不入库**：`fork` 是公开仓库，真实密钥一旦进历史就撤不回来。要"配置跟着仓库走"，走 `env.template` + `bootstrap.sh` 这条路。
