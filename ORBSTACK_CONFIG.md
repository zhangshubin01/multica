# OrbStack 配置清单 —— 换一台电脑复现这套 Multica 自托管栈

> 盘点于 **2026-09-20 16:53 CST**，源机 macOS 26.6.2 / arm64 / OrbStack 2.2.3。
> 本文只回答一件事：**OrbStack 这一层要配成什么样**，才能让同一套 compose 栈在新电脑上跑出一模一样的行为。
> 数据怎么搬（dump / 卷 / `.env` / daemon 身份）看 `MIGRATION_TO_NEW_HOST.md`；上团队服务器的差距看 `TEAM_DEPLOY_PLAN.md`。
>
> 标注规则同 `MIGRATION_TO_NEW_HOST.md`：**[实测]** = 写本文时在本机真跑过并贴出结果；**[推论]** = 由配置/代码得出、未在新机验证；**[注意]** = 实测踩过的坑。

---

## 0. 一句话

**OrbStack 层要复现的只有 4 样东西**，其余全部由 OrbStack 自己生成：

```
1. OrbStack 本体          2.2.3（arm64）—— 装图形 App，不是 brew cask
2. 引擎级配置             ~/.orbstack/config/docker.json（镜像加速；**必带，否则拉不动镜像**）
3. 客户端接线             docker context = orbstack + ~/.docker/config.json 的 currentContext
4. 宿主回连能力           host.docker.internal（OrbStack 自带，**不用配 extra_hosts**）
```

**不要搬**：`data.img.raw` 整盘（460 GB 稀疏 / 实占 87 GB，而且里面装着**别的 5 个栈**）。
**必须在新机重新取值**：Docker 网段（`RATE_LIMIT_TRUSTED_PROXIES` 要它）、LAN IP（三处 `.env`）。

---

## 1. 源机 OrbStack 现状（逐项实测）

| 项 | 值 | 怎么查 |
|---|---|---|
| OrbStack 版本 | `2.2.3 (2020300)`，commit `c83556b0ef8f1ba9a33abbb194622b6b7a1c0307` | `orb version` |
| 安装位置 | `/Applications/OrbStack.app`，**非 brew cask [实测]**：`brew` 可用（7.0.1）但 `brew list --cask` 为空，且 `brew info --cask orbstack` 输出 **`Not installed`**（cask 版本号 `2.2.3,20963`，与本机 `2.2.3 (2020300)` 的 build 号也对不上） | `ls -d /Applications/OrbStack.app` |
| CLI 软链 | `/usr/local/bin/orb`、`/usr/local/bin/orbctl` → App 内 `Contents/MacOS/bin/` | `ls -l /usr/local/bin/orb*` |
| 宿主系统 | macOS 26.6.2 (build 25G83)，**arm64**（Apple Silicon） | `sw_vers` / `uname -m` |
| Docker Engine | `29.4.0`，存储驱动 `overlayfs`，cgroup v2（cgroupfs） | `docker info` |
| 内核 | `7.0.14-orbstack-00380-ga7e0a2dc9535` | `docker info` |
| 架构（引擎视角） | `aarch64` | `docker info --format '{{.Architecture}}'` |
| 资源：CPU | **14**（= 宿主全部 14 核，不设上限） | `docker info --format '{{.NCPU}}'` |
| 资源：内存 | 引擎可见 **15.7 GB**（`MemTotal=16818978816` B；宿主 48 GB） | `docker info --format '{{.MemTotal}}'` |
| 当前 context | `orbstack` | `docker context show` |
| socket | `unix:///Users/shubinzhang/.orbstack/run/docker.sock` | `docker context ls` |
| `DOCKER_HOST` | **未设置**（靠 context 走 socket） | `echo $DOCKER_HOST` |
| Rosetta | **已装且运行中**（可跑 amd64 镜像） | `pgrep oahd` / `arch -x86_64 true` |
| 特权助手 | `/Library/LaunchDaemons/dev.orbstack.OrbStack.privhelper.plist`（root） | `ls /Library/LaunchDaemons \| grep orb` |
| 数据盘 | `~/Library/Group Containers/HUAQ24HBR6.dev.orbstack/data/data.img.raw` | 见 §7 |

---

## 2. Docker 客户端接线（新机照抄）

OrbStack 装了三个 context，但**只有 `orbstack` 是能用的**：

```bash
$ docker context ls
NAME            DOCKER ENDPOINT                                   ERROR
default         unix:///var/run/docker.sock                       # ← 不存在！见下
desktop-linux   unix:///Users/<user>/.docker/run/docker.sock      # ← Docker Desktop 的
orbstack *      unix:///Users/<user>/.orbstack/run/docker.sock    # ← 用这个
```

**[实测]** `/var/run/docker.sock` **不存在**（`ls: No such file or directory`）→ `default` context 是坏的。
所以新机上**不要在 shell 里设 `DOCKER_HOST`**，也**不要**依赖裸 `docker` 落到 `default`——保持
`currentContext: orbstack` 即可。查一下：

```bash
docker context show                       # 期望输出 orbstack
python3 -c "import json;print(json.load(open('$HOME/.docker/config.json'))['currentContext'])"
```

源机 `~/.docker/config.json` 的键（**值里没有敏感内容，可以整份拷**）：`auths`（**空**，镜像都是公开的）、
`credsStore: osxkeychain`、`currentContext: orbstack`、`plugins`、`features`。
`credsStore` 依赖 `docker-credential-osxkeychain`；源机 `/usr/local/bin/` 下还有
`docker-credential-desktop`、`docker-credential-ecr-login`（都是别的工具装的，本栈**不需要**）。

---

## 3. 引擎级配置：⚠️ 两个 `daemon.json` 的坑（本文最重要的一条）

这台机器上有**两份**名字都叫 daemon.json 的文件，**OrbStack 只读其中一份**：

| 文件 | 内容（实测） | 谁读它 |
|---|---|---|
| `~/.orbstack/config/docker.json` | `registry-mirrors: [docker.m.daocloud.io, docker.1ms.run]` | ✅ **OrbStack 引擎真正读这份** |
| `~/.docker/daemon.json` | `registry-mirrors: [docker.1ms.run, docker.xuanyuan.me]` + `builder.gc.defaultKeepStorage: 20GB` | ❌ Docker Desktop 的，**OrbStack 完全忽略** |

**证据 [实测]**：把两份的 `registry-mirrors` 摆在一起看引擎实际生效值 ——

```
~/.docker/daemon.json 声明   : [docker.1ms.run, docker.xuanyuan.me]
docker info 实际生效         : [docker.m.daocloud.io, docker.1ms.run]   ← 与 ~/.orbstack/config/docker.json 逐字一致
```

**[注意] 迁移含义**：新机如果只拷了 `~/.docker/daemon.json`（这是大部分"迁移 Docker 配置"教程教的那个文件），
镜像加速**一行都不会生效**；国内新机拉 `ghcr.io` / npm 基础镜像会直接超时，而你会以为已经配好了。
**要带的是 `~/.orbstack/config/docker.json`。**

源机这份文件的完整内容（可以直接照抄到新机同路径）：

```json
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://docker.1ms.run"
  ]
}
```

**[推论]** `~/.docker/daemon.json` 里的 `builder.gc` 在 OrbStack 下同样无效；build cache 的回收
走 OrbStack 自己的 GC（`docker builder prune` 手动清即可，新机首次构建前不必配）。

---

## 4. OrbStack 自身的设置（两个小文件）

**`~/.orbstack/vmconfig.json` [实测]**：

```json
{
	"setup.use_admin": false
}
```

`use_admin: false` = 容器内默认用户**不是** root 等价体。这套栈不依赖它，但如果你以后要
`docker exec -u root`，新机保持同值即可（默认也是 `false`，**不用手工建这个文件**）。

**`~/.orbstack/vmstate.json` [实测]**：

```json
{
	"version": 2, "majorVersion": 0, "minorVersion": 4, "arch": "arm64",
	"setupState": {
		"pathUpdateRequested": false,
		"editedShellProfiles": ["zsh"],
		"sshEdited": true
	}
}
```

三个信息：**数据盘格式版本 `0.4`**、**arch `arm64`**、安装时 OrbStack **改过 `~/.zshrc`**（`orb` 命令的 PATH/补全）。
新机装完 OrbStack 后开一个新终端确认 `orb version` 能跑；跑不了就是 shell profile 没被改（用 bash 或改过 `~/.zshrc` 的用户会碰到）。

**自启 [实测]**：`launchctl list | grep -i orb` 为空，`~/Library/LaunchAgents` 里没有 OrbStack，
只有 root 的 `/Library/LaunchDaemons/dev.orbstack.OrbStack.privhelper.plist`（特权网络助手）。
系统登录项是 `FigmaAgent, Lark` —— **OrbStack 不在里面**。它是标准 GUI App，**开机会话登录后需要它自己启动**；
新机上把它拖进「登录项」再开机验证一次（`docker info` 能通 = 起来了）。**[注意]** 这台 MacBook 的
合盖休眠会让整栈离线，这一点与 OrbStack 配置无关，是笔记本形态决定的。

**Linux 机器**：`orbctl list` **为空** —— 本栈**没有**用到 OrbStack 的 Linux machine 功能，新机不用建。

---

## 5. 日志轮转：OrbStack 有一个「侥幸的默认值」

**[实测]** OrbStack 引擎给容器的 `json-file` 默认就是 `max-size=20m, max-file=5`。
证据不只是我们自己的容器（我们在 `docker-compose.lan.yml` 里显式写了同一组值），
`docker ps` 里**任何别的栈、从没配过 logging 的容器**也是这个值：

```
loving_goldwasser     json-file {"max-file":"5","max-size":"20m"}
heuristic_ritchie     json-file {"max-file":"5","max-size":"20m"}
intelligent_shamir    json-file {"max-file":"5","max-size":"20m"}
zen_hamilton          json-file {"max-file":"5","max-size":"20m"}
```

**为什么这条对迁移很重要**：普通 Linux Docker 的 `json-file` **默认不限大小**，日志会一路涨到撑爆磁盘。
源机"从没管过日志也没出事"，靠的正是 OrbStack 这个默认值 —— **换成 Linux 服务器就没了**。
这也正是 `docker-compose.lan.yml` 里那段 `x-logging` 锚点存在的原因：

```yaml
x-logging: &default-logging
  driver: json-file
  options:
    max-size: "20m"
    max-file: "5"
```

→ **`docker-compose.lan.yml` 必须跟着走**（它是未跟踪文件，`git clone` 拿不到）。

---

## 6. 宿主回连：`host.docker.internal`

**[实测]** OrbStack 解析到的是 `0.250.250.254`（**不是** 网关地址，是 OrbStack 特有的魔法地址）：

```
$ docker run --rm alpine sh -c 'getent hosts host.docker.internal'
0.250.250.254     host.docker.internal  host.docker.internal

$ docker run --rm alpine wget -qO- --timeout=3 http://host.docker.internal:8080/healthz
{"status":"ok","checks":{"db":"ok","migrations":"ok"}}
```

栈里有两处**硬依赖**它，新机（macOS + OrbStack）**零改动可用**：

| 位置 | 值 |
|---|---|
| `.env` | `REDIS_URL=redis://host.docker.internal:6379/0` |
| `Caddyfile.lan` | `reverse_proxy host.docker.internal:8080`（`/ws` 与 `/ws/*`）与 `host.docker.internal:3100`（其余） |

**[注意]** 这是**有意为之**的设计（走已发布的宿主回环端口，即使某次只给两个 `-f` 起栈也连得上 redis），
代价是拓扑绑死在"容器能回连宿主"上。**Linux 目标机不成立** —— 要改成服务名（`backend:8080` /
`frontend:3000` / `redis://redis:6379/0`），详见 `MIGRATION_TO_NEW_HOST.md` §11.1–11.2。

---

## 7. 数据盘：在哪、多大、搬不搬

```
~/Library/Group Containers/HUAQ24HBR6.dev.orbstack/
├── data/data.img.raw      460 GB（稀疏文件，带 @ 扩展属性）
├── Library/{Application Support, Preferences, Caches}    Preferences 为空
└── .com.apple.containermanagerd.metadata.plist
```

| 项 | 实测值 |
|---|---|
| 目录实际占用 | **87 GB** (`du -sh`) |
| 镜像文件表观大小 | 460 GB（`ls -lh`；稀疏，实际没占这么多） |
| 里面装了什么 | **不止 multica** —— 见下 |

**[注意] 这是最关键的一条：OrbStack 的数据盘是全机器共享的，不是 multica 专用的。** **[实测]** 除本栈外还有：

```
网络: clawith_network, deepthink_default, langfuse_default, lanhu-mcp_default, openviking-shared
卷  : clawith-agent_{pgdata,redisdata,agentdata}(9.2GB), langfuse_*(点击/日志/minio ≈4GB),
      gradle_cache_global(12.6GB), global_android_sdk(2.8GB), global_jdk_cache(334MB), ...
```

→ **`rsync` / `cp` 整个 `data.img.raw` = 连带搬走这 5 个无关栈的几十 GB 数据和缓存**，
既慢又容易把新机污染成"半台旧机"。**不建议。**

**OrbStack 层真正要搬的 multica 资产只有 5 个卷 [实测]**：

| 卷 | 大小 | 搬不搬 |
|---|---|---|
| `multica_pgdata` | **82.17 MB** | ✅ 或用 `pg_dump -Fc`（`.dump` 只有 ~800 KB，**推荐**） |
| `multica_backend_uploads` | **6.55 MB** | ✅ **已经有内容了**（见下方提醒） |
| `multica_redis_data` | 3.1 KB | ❌ 不搬，只丢限流计数与租约 |
| `multica_edge_config` / `multica_edge_data` | 354 B / 143 B | ❌ 不搬，纯 HTTP 无证书 |

> ⚠️ **`MIGRATION_TO_NEW_HOST.md` 的前提已过期**：那份文档 §1 写「附件卷当前**为空**（0 B）」，
> 2026-09-20 实测已是 **6.55 MB**。所以「附件卷 tar/还原路径未实测」这个待办**现在必须真跑一次**，
> 不能等"以后有附件再说"——已经晚了。同时注意 git HEAD 也从 `7d9613bd2` 漂到了 `8c4f4328`（09-19）。

---

## 8. 网络与网段：**每次换机都要重新取值**

**[实测]** OrbStack 的网段是**动态分配**的，不是固定值：

| 网络 | 当前网段 |
|---|---|
| `multica_default`（本栈） | **192.168.148.0/24** |
| `bridge`（内置） | 192.168.215.0/24 |
| 引擎声明的默认地址池 | 192.168.97.0/24、192.168.107.0/24 |

新机上**大概率不是** `192.168.148.0/24`（本机有 5+ 个栈在抢地址池）。所以：

```bash
# 新机上必做，取值后再填进 docker-compose.lan.yml
docker network inspect multica_default -f '{{(index .IPAM.Config 0).Subnet}}'
```

这个值喂给 `RATE_LIMIT_TRUSTED_PROXIES`（当前 `192.168.148.0/24`）。**为什么必须对**：backend 看到的对端是
frontend/edge 的**容器 IP**，只有把 Docker 网段登记为可信代理，限流才会按 `X-Forwarded-For` 里的真实客户端
IP 计数；否则**全团队共用一个桶**，认证类默认 5 次/分钟必然误伤。

---

## 9. 端口发布（新机照此核对）

**[实测]** `docker ps` 的完整映射：

```
multica-edge-1      0.0.0.0:80->80/tcp, [::]:80->80/tcp      ← 内网入口（唯一对外）
                    （443/tcp、443/udp、2019/tcp 是 EXPOSE，未发布到宿主）
multica-frontend-1  127.0.0.1:3100->3000/tcp                 ← 仅回环
multica-backend-1   127.0.0.1:8080->8080/tcp                 ← 仅回环
multica-redis-1     127.0.0.1:6379->6379/tcp                 ← 仅回环
multica-postgres-1  5432/tcp                                 ← 只暴露给容器网络，宿主根本没有
```

宿主侧监听这些端口的是 **OrbStack 进程**（`lsof` 里显示 `OrbStack *:80` 等），不是 Docker 之外的第三个东西：

```
OrbStack  *:80
OrbStack  127.0.0.1:3100
OrbStack  127.0.0.1:6379
OrbStack  127.0.0.1:8080
```

**[注意]** 新机上 `80` 若被占用（macOS 上常见的是别的 Web 服务），入口起不来；而 **OrbStack 不会帮你换端口**。
先自检：

```bash
lsof -nP -iTCP:80,3100,8080,6379 -sTCP:LISTEN || echo "端口空闲"
```

`FRONTEND_PORT=3100` 是我们刻意选的（不是上游默认的 3000）——`FRONTEND_PORT` 只改**宿主侧**端口，
容器内恒为 3000。新机上把这个值原样带过去即可。

---

## 10. 新电脑：OrbStack 层落地步骤

### 10.1 装 OrbStack

1. 从 https://orbstack.dev 下 **arm64** 版，拖进 `/Applications` 打开一次，等它把引擎跑起来
   （源机就是这种方式装的，**不是** `brew install --cask orbstack`）。
2. 开一个新终端确认 `orb version` 与 `docker context show`（应为 `orbstack`）。
3. **打开"登录时启动"**（源机的登录项里没有它，别把这个遗漏照抄过去）。
4. 按需装 Rosetta（源机已装；只有你需要跑 `linux/amd64` 镜像时才必需）。

### 10.2 落引擎级配置

```bash
mkdir -p ~/.orbstack/config
cat > ~/.orbstack/config/docker.json <<'JSON'
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://docker.1ms.run"
  ]
}
JSON
# 生效：重启 OrbStack（菜单栏 Quit 再打开），然后核对
docker info | grep -A3 'Registry Mirrors'
```

**[注意]** 改完**重启 OrbStack**（菜单栏 Quit 再打开）让它生效 —— 引擎配置在引擎启动时读取。
**[推论，未实测]** 这一点没能在源机验证（源机这份文件早已就位、无需重启即可见生效），
所以新机上先写文件、重启、再用 `docker info | grep -A3 'Registry Mirrors'` 确认，不要只看文件写对了就往下走。

### 10.3 网络自检（在拉任何镜像之前）

```bash
docker run --rm alpine sh -c 'wget -qO- --timeout=5 https://registry-1.docker.io/v2/ >/dev/null && echo "registry OK"'
docker run --rm alpine sh -c 'wget -qO- --timeout=5 https://ghcr.io/ >/dev/null && echo "ghcr OK"'
```

**[实测] 源机**：容器**直连**这些站点（宿主才需要 `HTTPS_PROXY`）。**[注意]** 新机若在需要代理的网络里，
配的是 **Docker 引擎的代理**（OrbStack 设置里），**不是**宿主 shell 的 `HTTPS_PROXY`——这一点与
`git`/`brew` 的配法不同，容易混。

### 10.4 然后才回到 `MIGRATION_TO_NEW_HOST.md`

数据搬运（dump → 先起 postgres → `pg_restore` → 起全栈）与验收清单在那边，本文不重复。
新机上仍需改的 **4 处**（其中 2 处属于本文范围）：

| # | 文件 | 键 | 怎么取新值 |
|---|---|---|---|
| 1 | `.env` | `MULTICA_APP_URL` | 新机 LAN IP（`ipconfig getifaddr en0`）；**它决定 `/api/config` 发给同事的 daemon 安装地址** |
| 2 | `.env` | `CORS_ALLOWED_ORIGINS` | 同上 |
| 3 | `docker-compose.lan.yml` | `RATE_LIMIT_TRUSTED_PROXIES` | §8 的 `docker network inspect` |
| 4 | `Caddyfile.lan` | `host.docker.internal` | **仅 Linux 目标机**才改服务名 |

---

## 11. 新机 OrbStack 层验收清单

```bash
# ① 版本与接线
orb version                              # 期望 2.2.3 或更新
docker context show                      # 期望 orbstack
docker info --format '{{.ServerVersion}} {{.Architecture}} {{.NCPU}}'   # 期望 29.x aarch64 14

# ② 引擎配置真的生效了（本文最容易假通过的一条）
docker info 2>/dev/null | grep -A3 'Registry Mirrors'
#   期望出现 docker.m.daocloud.io —— 若为空或出现 xuanyuan.me，说明你改错了文件（见 §3）

# ③ 宿主回连
docker run --rm alpine getent hosts host.docker.internal
#   期望 0.250.250.254（OrbStack）—— Linux 上会是别的地址，且需要 extra_hosts

# ④ 日志轮转默认值
docker inspect <任意容器> --format '{{json .HostConfig.LogConfig.Config}}'
#   期望 {"max-file":"5","max-size":"20m"}（OrbStack 自带；Linux 上为空）

# ⑤ 网段（唯一必须回填到配置里的值）
docker network inspect multica_default -f '{{(index .IPAM.Config 0).Subnet}}'

# ⑥ 端口
lsof -nP -iTCP:80,3100,8080,6379 -sTCP:LISTEN
```

再补两条与 OrbStack 无关但同批要验的：`docker ps` 五个容器齐全、`curl -s 127.0.0.1:8080/healthz`
返回 `{"status":"ok","checks":{"db":"ok","migrations":"ok"}}`。

---

## 12. OrbStack 专属的坑（逐条对照）

| # | 坑 | 结论 |
|---|---|---|
| 1 | 只拷 `~/.docker/daemon.json` 当"迁移 Docker 配置" | **OrbStack 不读它**。要带 `~/.orbstack/config/docker.json`（§3，已实测对照） |
| 2 | 改了 `docker.json` 不重启 OrbStack | 只在引擎启动时读；必须重启 |
| 3 | 依赖 `default` context | `/var/run/docker.sock` **不存在**，`default` 是坏的；用 `orbstack`（§2） |
| 4 | 把整台机器的 Docker 当 multica 专用 | 本机还有 5 个别的栈；搬卷要**点名**，别整盘（§7） |
| 5 | 以为日志轮转是"我们配的" | 是 OrbStack 引擎默认 20m×5；Linux 上消失 → `docker-compose.lan.yml` 必带（§5） |
| 6 | 在 Linux 上照抄 `host.docker.internal` | OrbStack 解析到 `0.250.250.254`，Linux 默认不注册且 `host-gateway` 也连不通（端口只绑 127.0.0.1）；改服务名（§6） |
| 7 | 以为 Docker 网段固定 | 动态分配，本机 5+ 个栈在抢；每次换机重查（§8） |
| 8 | 容器需要代理时去配宿主 `HTTPS_PROXY` | 要配**引擎**的代理；源机容器是直连的（§10.3） |
| 9 | 登录项里没有 OrbStack 就以为它会自启 | 源机登录项只有 `FigmaAgent, Lark`；新机要手工加（§4） |
| 10 | `use_admin: false` | 容器内默认非 root 等价；本栈不受影响，别去改它 |

---

## 13. 还没验证的

- 本文全部数据取自**源机**。**新机上的 OrbStack 层本身没有跑过**（没有第二台机器做过验证）——
  §11 的清单就是为这次验证准备的。
- **架构相关**：源机是 `arm64`。新机若是 Intel Mac（`amd64`），`docker save` / `load` 镜像**不通用**，
  必须在新机重建（见 `MIGRATION_TO_NEW_HOST.md` §11.5 的三选一）。
- **整盘搬迁 `data.img.raw`** 未做也未推荐（§7）；若确实要试，`rsync --sparse` + 迁移时**两侧都不能运行** OrbStack，
  且需自行承担跨用户/跨架构的兼容性风险。
- OrbStack 的**自动更新**：源机是 2.2.3。大版本升级可能改变引擎默认值（比如 §5 的日志轮转、
  §8 的地址池），**升级后请重跑 §11**。
