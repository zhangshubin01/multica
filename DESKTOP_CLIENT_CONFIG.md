# Multica 桌面客户端（macOS / Windows）连接自建服务端

> 目标实例：内网入口 `http://192.168.20.131`（部署现状见 `TEAM_DEPLOY_PLAN.md` §9）
> 2026-09-15，基于仓库 v0.4.43 与本机/内网实测。

打包的客户端（mac DMG/ZIP、win NSIS）**默认连官方云** `https://api.multica.ai`；客户端没有 UI 切换服务端的入口，指向自建实例只有一条路：**在家目录放一个配置文件并重启客户端**。Windows 与 macOS 只有路径和写文件的坑不同，内容完全一致。

## 1. 配置内容

```json
{
  "schemaVersion": 1,
  "apiUrl": "http://192.168.20.131",
  "wsUrl": "ws://192.168.20.131/ws",
  "appUrl": "http://192.168.20.131"
}
```

- `schemaVersion` 必须为 `1`；`apiUrl` 必填（后端公开地址，`http`/`https`）。
- `wsUrl`、`appUrl` **可以省略**，客户端自动推导：ws 协议 + 追加 `/ws`；域名以 `api.` 开头且 ≥3 段才剥前缀，`192.168.20.131` 这类 IP 原样保留。写全只是为了排错时一眼看清。
- 这里只能写明文 `http://`（局域网 Caddy 只有 80、无 TLS）。上了域名 + TLS 后换成 `https://` 即可，ws 会跟着变 `wss://`。

> 注意不是 CLI 的 `~/.multica/config.json`（那是 `server_url`）；Desktop 只读下面这两个路径，两者互不影响。

## 2. macOS

路径 `~/.multica/desktop.json`（即 `/Users/<你>/.multica/desktop.json`）：

```bash
mkdir -p "$HOME/.multica" && cat > "$HOME/.multica/desktop.json" <<'JSON'
{ "schemaVersion": 1, "apiUrl": "http://192.168.20.131", "wsUrl": "ws://192.168.20.131/ws", "appUrl": "http://192.168.20.131" }
JSON
```

重启才生效（**⌘Q 完全退出**，关窗口不算）：

```bash
osascript -e 'quit app "Multica"'; sleep 2; open -a Multica
```

用 `cat` 而不是图形编辑器，是为了避免编辑器写入 BOM——带 BOM 的 JSON 解析失败，客户端直接弹配置错误。

## 3. Windows

路径 `%USERPROFILE%\.multica\desktop.json`（即 `C:\Users\<你>\.multica\desktop.json`）：

```powershell
$dir = "$env:USERPROFILE\.multica"
New-Item -ItemType Directory -Force $dir | Out-Null
[System.IO.File]::WriteAllText("$dir\desktop.json", @'
{ "schemaVersion": 1, "apiUrl": "http://192.168.20.131", "wsUrl": "ws://192.168.20.131/ws", "appUrl": "http://192.168.20.131" }
'@)
```

**必须用 `WriteAllText`**：`>` 重定向和 `Out-File` 会写成 UTF-16/BOM，客户端读不了。另外**记事本会偷偷补 `.txt`**（得到 `desktop.json.txt`），客户端当作「文件不存在」→ 静默连云端、**不报任何错**。存完自查：

```powershell
Get-ChildItem "$env:USERPROFILE\.multica" -Filter "desktop.json*" | Select-Object Name
# Name 必须正好是 desktop.json
```

客户端无托盘图标，关掉窗口即退出，重开即可（残留进程用任务管理器结束 `Multica.exe`）。

## 4. 验证（两端一样）

- 文件：`cat ~/.multica/desktop.json`（win 用 `Get-Content "$env:USERPROFILE\.multica\desktop.json"`）
- 目录：`ls ~/.multica/profiles/` 应出现 `desktop-192.168.20.131/`
- 客户端：**设置 → 桌面应用 → 守护进程**，`serverUrl` 显示 `http://192.168.20.131`

首次连接需要重新登录一次（内网走邮箱验证码；Google 登录会打开系统浏览器，回调靠 `multica://` 深链跳回客户端）。

## 5. 行为与坑

| 事项 | 说明 |
|---|---|
| 读取时机 | **只在启动时读一次**，改完必须重启 |
| 文件不存在 | **静默**用云默认 `api.multica.ai`，不提示——"配了没生效"最常见的原因 |
| 文件存在但无效 | 报配置错误并**阻塞启动**，不会回退云端 |
| 守护进程 | 按 `apiUrl` 的 host 建独立 profile `~/.multica/profiles/desktop-<host>/`，其 `server_url` 钉成同一个地址；原有的 `desktop-api.multica.ai` 保留，云/内网互不干扰 |
| 不需要 CORS | 客户端渲染层是 `file://` + `webSecurity: false` |
| 联网依赖 | 自动更新走 GitHub Releases；内置 CLI 随安装包携带，纯内网也能跑守护进程（杀软隔离内置 CLI 会导致守护进程起不来） |

## 6. 回滚到官方云

```bash
rm ~/.multica/desktop.json        # win: Remove-Item "$env:USERPROFILE\.multica\desktop.json"
```

重启客户端即可。`profiles/desktop-192.168.20.131/` 会留下（含该实例的 PAT 与日志），不影响使用，要清干净可手删。

## 7. 故障速查

| 现象 | 处理 |
|---|---|
| 仍是云地址、且无报错 | 文件位置或文件名不对（win 常见 `.txt`）；按 §4 的命令核对后重启 |
| 启动即报配置错误 | JSON 语法 / `schemaVersion` / `apiUrl` 有问题，或文件是 UTF-16/BOM；按 §2、§3 重写 |
| 登录了但实时更新不动 | WS 不通。`curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" http://192.168.20.131/ws`：返回后端 400 JSON = 链路通；返回前端 404 = `Caddyfile.lan` 的 `@ws` 块没热加载 |

## 附录

**本部署已实测（2026-09-15，主机 IP 从 `192.168.20.9` 变为 `192.168.20.131` 之后）**：

| 探测 | 结果 |
|---|---|
| `http://192.168.20.131/health` | 200（`edge` 监听 `0.0.0.0:80`，任何本机网卡地址都能进） |
| `http://192.168.20.9/health` | 已不可达 |
| `http://192.168.20.131/ws` 升级 | 已转给后端 8080 |
| `http://192.168.20.131/api/config` | ⚠️ 200，但 `daemon_server_url` 仍是**旧地址** `http://192.168.20.9` —— `.env` 的 `MULTICA_APP_URL` 没跟着改，CLI/守护进程按这个地址装会连不上 |

> 主机地址是 DHCP 分配的，**还会再变**。长期方案是给这台机器固定 IP（或 DHCP 保留）+ 一个内网域名，客户端配置里写域名而不是 IP；否则每次变都要改一遍 `.env`、重发客户端配置、并且让同事重装 CLI。
> 本机另一块网卡 `en8` 是 `192.168.4.36`，若同事在那个网段，用这个地址。

**代码位置**（要改默认值或做定制包时看）：

| 关注点 | 位置 |
|---|---|
| 云默认三地址 | `apps/desktop/src/shared/runtime-config.ts:16-21` |
| `wsUrl`/`appUrl` 推导 | `apps/desktop/src/shared/runtime-config.ts:85-111` |
| 读取规则（dev 用 env、生产读 desktop.json、缺失回落） | `apps/desktop/src/main/runtime-config-loader.ts:13-45` |
| 启动装载、传给渲染层 | `apps/desktop/src/main/index.ts:619-634`、`:718` |
| daemon profile 命名 / 注入 `server_url` | `apps/desktop/src/main/daemon-profile.ts:26-45`、`daemon-manager.ts:732` |
| 官方说明（含 Win 编码/文件名坑） | `apps/docs/content/docs/desktop-app.zh.mdx` |

> 想让所有安装包**默认**就指向 `192.168.20.131`，需改 `DEFAULT_RUNTIME_CONFIG` 并自行重新打包——属于改仓库代码逻辑，动手前先确认。
