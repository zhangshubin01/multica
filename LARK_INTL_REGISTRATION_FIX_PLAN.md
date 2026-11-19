# Lark 国际版注册「链接已失效」根因与修复方案

状态：已实施（代码与单测），部署与实测见 §9
日期：2026-11-19
影响文件：`server/internal/integrations/lark/registration.go`（+ 其单测）

---

## 1. 现象

国际版 Lark 用户（本团队）在「添加消息渠道 → Lark」时，扫码或点击「或点这里直接在 Lark 中打开」链接，Lark 侧均先打开一个登录/授权页，随后报 **「链接已失效」**。后端轮询一直停在 `authorization_pending`，直到超时。

（同一用户更早的「暂不支持访问」是另一回事：新版 Lark App 拦截了跨云扫码，属于客户端行为，与本方案无关。）

## 2. 根因（已多源确证）

**Lark 官方废弃了 `/page/launcher` 这个设备流授权页面，国际版现在会立刻拒绝它**，报错即「链接已失效」（Lark 错误码 `10074`）。

Multica 当前的实现是：

1. 按 region 选 begin 端点：`lark` → `accounts.larksuite.com`，`feishu` → `accounts.feishu.cn`；
2. `request_user_info` 只请求 `open_id`；
3. **直接采用 begin 响应里的 `verification_uri_complete` 作为二维码/链接地址**。

而 `accounts.larksuite.com` 的 begin 响应至今仍返回旧的 `https://open.larksuite.com/page/launcher?user_code=…`（本机实测可复现）。于是每次拿到的链接都是一出即废的旧地址。

### 证据

1. [`NousResearch/hermes-agent#100048`](https://github.com/NousResearch/hermes-agent/issues/100048)：标题《[Bug]: Lark scan-to-create uses expired /page/launcher flow》，现象逐字一致（`链接已失效` / Link expired，code `10074`，`authorization_pending` 轮询到超时），并附根因分析与「官方 `larksuite/cli` 的对照实验」。
2. [`eddiearc/larkin#142`](https://github.com/eddiearc/larkin/pull/142)：《fix(setup): use official /page/cli for International Lark QR》—— 明确 `/page/launcher` 被 Lark 拒 `HTTP 400 code 10074（链接已失效）`，修复是改走 `/page/cli`。
3. 官方工具 [`larksuite/cli`](https://github.com/larksuite/cli)（`internal/auth/app_registration.go`）—— 当前正确实现，与上面两份材料一致。

## 3. 现状 vs 官方正确流程

| 环节 | Multica 现状（`registration.go` `Begin`） | 官方正确做法（`larksuite/cli`） |
|---|---|---|
| begin 端点 | 按 region：lark 走 `accounts.larksuite.com` | **永远从 `accounts.feishu.cn` 起手**，国际版也如此 |
| `request_user_info` | `open_id` | `open_id tenant_brand` |
| 授权 URL | 采用返回的 `verification_uri_complete`（=`/page/launcher`） | **自行拼 `{目标域 open 主机}/page/cli?user_code=…`**，加 `from=cli` |
| 轮询跨云 | 已实现 `tenant_brand` 驱动的 feishu↔lark 切换 | 同（起手 feishu，见 `tenant_brand=lark` 切到 larksuite） |

关键点：Multica 的 `Poll` 已经写好了 `tenant_brand` 跨云切换（`registration.go` `Poll`，对称处理 feishu↔lark），只是 `Begin` 三个点没接上，导致这套机制从未真正生效——国际用户拿到的仍是废链接。

## 4. 修复方案（3 处，都在 `registration.go`）

### 4.1 永远 feishu 起手

删掉 `Begin` 里的按 region 选端点的逻辑：

```go
// 现状（删除）：
domain := c.cfg.Domain
if region == RegionLark {
    domain = c.cfg.LarkDomain
}

// 改为：
// accounts.larksuite.com 的 begin 仍返回旧 /page/launcher 地址，
// 国际版现在直接拒掉；官方流程永远从 feishu 起手，再按 tenant_brand 切。
domain := c.cfg.Domain
```

### 4.2 请求 tenant_brand

```go
// 现状：
"request_user_info": []string{"open_id"},

// 改为：
"request_user_info": []string{"open_id tenant_brand"},
```

（`Poll` 已读 `user_info.tenant_brand` 做跨云切换；不请求这个字段，切换信号就永远拿不到。）

### 4.3 自行拼 /page/cli 授权 URL

不再用 `verification_uri_complete` + `decorateQRCodeURL`，改为按目标 region 的 open 主机拼 `/page/cli`：

```go
// 现状：
if resp.VerificationURIComplete == "" { …invalid_response… }
qr, err := decorateQRCodeURL(resp.VerificationURIComplete, c.cfg.Source, namePreset)

// 改为：
if resp.UserCode == "" { …invalid_response… }
qr, err := buildVerificationURL(region.OpenPlatformBaseURL(), resp.UserCode)
```

新增：

```go
// buildVerificationURL 在目标 region 的 open 主机上拼当前官方 /page/cli 授权地址。
// /page/launcher 已被 Lark 国际版废弃；官方 larksuite/cli 用 /page/cli?user_code=…&from=cli。
func buildVerificationURL(openHost, userCode string) (string, error) {
    u, err := url.Parse(strings.TrimRight(openHost, "/") + "/page/cli")
    if err != nil {
        return "", err
    }
    q := u.Query()
    q.Set("user_code", userCode)
    q.Set("from", "cli")
    u.RawQuery = q.Encode()
    return u.String(), nil
}
```

`region.OpenPlatformBaseURL()` 已在 `types.go` 就绪：`lark` → `https://open.larksuite.com`，否则 `https://open.feishu.cn`。

### 4.4 顺带清理（可做可不做）

- `decorateQRCodeURL`、`RegistrationConfig.Source`、`namePreset` 不再被使用，可退休（`name` 预填 bot 名的能力随 launcher 一起废弃，代价是首次需在 Lark 控制台手动改 bot 名）。
- 单测 `registration_test.go` 里钉住旧行为的断言要改：`request_user_info` 期望值、以及把 mock 的 `verification_uri_complete` 换成 `user_code` + 断言 `/page/cli` URL。

## 5. 评审：这是最佳、最简修复吗？

**结论：是。** 3 处改动各自必要、合起来充分，恰好复用 Multica 已有但从未接通的跨云切换机制，没有更简的正确修法。已逐个排除的替代方案：

| 替代方案 | 为什么否决 |
|---|---|
| **只把 URL 换成 `/page/cli`**，保留 larksuite 起手 + `open_id` | 不充分。`accounts.larksuite.com` 的 begin 端仍返回 legacy launcher 语义；官方与两份材料都明确「起手必须在 feishu」「request `open_id tenant_brand`」。只换 URL 属撞运气。 |
| **前端硬编码拼 URL** | 不可行。`user_code` 由 begin 返回，只有后端持有；且 begin 端点/表单仍需改，绕不开后端。 |
| **BYO（手动填 App ID/Secret）** | 可用 workaround（hermes issue 提及），但 Multica 的 Lark 连接流没有 BYO UI，新增 UI+后端改动更大，不是最简。 |
| **升级官方 `oapi-sdk-go` 的 scene/registration** | 无效。核对过官方 SDK 的 `buildQRCodeURL` 仍拼 `/page/launcher`（`v3_main` 分支），同样会挂；真正修复的是另一条线 `larksuite/cli`。 |
| **保留 region 分支、只修 lark 一条** | 官方就是「永远 feishu 起手」，region 分支本身是错误设计；统一掉反而更简单。 |

## 6. 代价与风险

1. **`name` 预填丢失**：`/page/cli` 官方不带 `name`，新装的 bot 用默认名，首次需手动改名。纯装饰性。
2. **依赖 Lark 当前未公开文档化的 `/page/cli` 行为**：这是官方 CLI 现用流程，但不在公开 API 文档里，Lark 未来可能再变。已按官方源码对齐，风险最低。
3. **需后端重建镜像 + 部署**：前端不动（前端已传 `region=lark`，二维码 URL 是后端返回的）。

## 7. 验证计划

1. 改完跑 `registration_test.go`（更新断言后全绿）。
2. 重建后端镜像、`up -d` 部署（前端无需重建）。
3. 本机 API 实测：`install/begin?region=lark` 返回的 `qr_code_url` 应为 `https://open.larksuite.com/page/cli?user_code=…&from=cli`（不再是 `/page/launcher`）。
4. 实机：用户新开对话框 → 扫码或点链接 → 在 Lark 授权 → 后端落库 `channel_installation.region=lark` + WS 连接 + 首条入站消息，三处核验。

## 8. 回滚

- 代码：`git revert` 本次提交。
- 部署：重建后端镜像回退即可（前端未改）。

## 9. 实施与实测记录（2026-11-19）

- 代码：`registration.go` 3 处改动 + `decorateQRCodeURL`/`Source`/`namePreset` 退休 + `registration_test.go`/`registration_service_test.go` 断言更新；`lark` 包全量测试通过，`internal/handler` Lark 测试通过，`go build ./...` 通过。
- 部署：重建 `multica-backend:dev`（新镜像 manifest `sha256:b2da37b6…`），backend 容器 healthy。
- API 实测：`POST /api/workspaces/…/lark/install/begin?region=lark`（Bearer JWT）返回：

  ```json
  {"session_id":"U0hCtlhGtu6XotuikBddYNaq3e_pCfWf","qr_code_url":"https://open.larksuite.com/page/cli?from=cli&user_code=YRXL-47YA","expires_in_seconds":3600,"poll_interval_seconds":5}
  ```

  `qr_code_url` 已是 `open.larksuite.com/page/cli`（不再是 `/page/launcher`）。
- 待办：用户实机扫码授权，核验 `channel_installation.region=lark` + WS 连接 + 首条入站消息。
