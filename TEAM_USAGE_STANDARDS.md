# Multica 使用规范 —— 20 人 Android 团队 / 70 个维护项目

> **适用对象**：workspace `土卫九`（slug `phoebe-1yu1`，issue 前缀 `PHOE`）全体成员，以及 70 个在维护的 Android 项目。
> **事实基线**：Multica 仓库 v0.4.43-21（本次核对 HEAD `4d4cae775`，2026-09-17）+ 本团队自托管实例（内网入口见 §2.2）。
> **本文只管"人怎么用"**；服务器怎么部署见 `TEAM_DEPLOY_PLAN.md`，怎么换机器见 `MIGRATION_TO_NEW_HOST.md`。本文不要求任何人改服务器配置。
> **标记约定**：**[产品约束]** = Multica 的既定行为，违反会报错或产生副作用；**[团队约定]** = 本团队自定的纪律，产品不强制，但违反要走评审；**[待办]** = 已知缺口，别当它是正常状态。

### 相关文档

| 文档 | 管什么 |
|---|---|
| `TEAM_DEPLOY_PLAN.md` | 服务器部署、配置审计、待办优先级（§11） |
| `MIGRATION_TO_NEW_HOST.md` | 把整套栈搬到另一台机器（含成员 daemon 的两种上线方式） |
| `DESKTOP_CLIENT_CONFIG.md` | 桌面客户端连自建实例（`desktop.json`） |
| `LONG_TERM_MEMORY.md` | 长期记忆/知识库方案，其 §8 是"70 项目 / 70 agent / GitLab"条件下的落地版本 |

### 现状快照（2026-09-17 查库，会变，只作参照）

| 项 | 现状 |
|---|---|
| workspace | 2 个：`土卫九`（团队，前缀 `PHOE`，已用到 `PHOE-19`）+ `泰国个人项目`（个人，前缀 `TAIG`）。**团队协作只在 `土卫九`** |
| 成员 | 4 人（20 人正在陆续进场） |
| agent | `土卫九` 21 个（其中 **9 个是 Agent Builder 生成的系统 agent**，名字形如 `.multica-agent-builder-<uuid>`；另有 `测试1111` 这类临时 agent）；`泰国个人项目` 8 个 |
| 项目 / issue / 运行 | `土卫九` 4 个 project、18 个 issue、109 次运行 |
| runtime | `土卫九` 12 个（11 在线），**全部 private** |
| 代码托管 | **自托管 GitLab `http://192.168.5.254`，已于 2026-09-11 连接**；GitHub App 未装（0 安装） |
| IM | 飞书/Lark 3 个安装（其中 `土卫九` 2 个），均 active |
| 其它 | 2 个 squad、1 个 autopilot、2 个工作区 MCP 服务、2 个 skill、0 个保存视图 |

---

## 0. 十二条铁律（先读这个）

| # | 铁律 |
|---|---|
| 1 | **团队工作只用一个 workspace**（`土卫九`）。70 个项目一律建成 project；个人/试验可以另开 workspace，但**不要**在其中做团队协作（workspace 之间数据完全隔离）。 |
| 2 | **workspace 级仓库清单只放公共库**（公共 UI/SDK/规范仓库，建议 ≤5 个）。70 个业务仓库一律挂到各自 project —— workspace 清单**永远**进每个 run 的上下文。 |
| 3 | **每个 issue 必须落 project、带 label、有 assignee**；`backlog` 是停车场，停在里面的 issue 不会触发任何 agent。 |
| 4 | **仓库工作目录一律用 worktree（Parallel）**；Android 仓库必须 gitignore 构建产物 —— 未跟踪文件超过 **2000 个 / 200 MiB**，run 直接失败。 |
| 5 | **agent 的 Instructions 必须写清构建命令、测试命令、边界与停止条件**；改配置只影响之后的 run，不影响正在跑的 run。 |
| 6 | **密钥不进 `custom_env`、MCP 配置、skill、知识库仓库**（前两者明文存服务端）。keystore 口令、Play Console service account JSON、内网 maven 凭据、CI 变量，都只在受控位置。 |
| 7 | **agent 的 Access 默认 Only me**；团队共享的显式放开，且 **owner/admin 也无权绕过**别人的 Access。 |
| 8 | **agent 推分支、开 MR 用的是它所在机器的 git 凭据** —— 那台机器的 GitLab 权限，就是 agent 的权限。 |
| 9 | **MR 标题或分支名必须带 `PHOE-<编号>`**；只有带 `Closes PHOE-123` 的合并才会把 issue 自动置为 Done。 |
| 10 | **评论里 @ 必须从建议列表选**（手打 `@名字` 无效）；只讨论不触发用 `/note` 开头；**不要用 `@all` 当广播**。 |
| 11 | **run 绑定 runtime、永不换机器**；改 assignee / 改状态都不会停掉已开始的 run，要停就去 issue 右侧执行日志里停。 |
| 12 | **服务器换机或换 IP 后，每台机器都要把 daemon 重新指到新入口**（§14）；旧地址 `192.168.20.9` **已失效**。 |

---

## 1. 术语与模型（避免混用）

| 词 | 含义 | 中文界面口径 |
|---|---|---|
| workspace | 顶层隔离边界：成员、issue、project、agent、skill、运行历史都归属它 | 工作区 |
| project | 组织 issue 的容器，提供描述与资源（仓库），**不是权限单位** | 项目 |
| issue | 工作单元，编号 `PHOE-123` | 界面译作"**任务**"，不要与 run 混用 |
| run | 一次 agent 执行（库里 `agent_task_queue` 一行），有独立状态、日志、用量 | 运行 |
| agent | 可复用的"身份 + 能力 + 执行配置"，**不是常驻进程**，有活才跑 | 智能体 |
| runtime | 一个执行环境 = **一台机器 + 一个 AI 编码工具** | 运行时 |
| daemon | 某台机器上的 Multica 后台进程，负责领活（claim）与执行 | 守护进程 |
| squad | 1 个 leader agent + 若干成员，用于把活分给"谁合适" | 小队 |
| channel | 飞书/Lark 等 IM 接入，一个 Bot 绑一个 agent | 渠道 |

链路：**issue / chat / autopilot 触发 → 建 run → 某个 runtime 领走 → 在那台机器上调起 AI 编码工具 → 结果（评论、状态、分支/MR）回写 issue。**

决定下面所有规范的结构性事实：

- **一个 issue 至多属于一个 project**；issue 从 A 项目移到 B 项目**不产生副本**。
- **project 没有 slug、没有 key、没有权限模型**，URL 用 UUID，可见性由 workspace 成员身份决定。
- **agent 没有"绑定仓库"这个字段**：代码上下文来自 ① workspace 仓库清单（永远进上下文）+ ② project resources。
- **run 与 issue 状态相互独立**：run `completed` ≠ issue 完成；状态由 agent 或人显式写入。
- **服务端只存元数据与文本**（issue/评论/run 上下文/回写内容/附件），代码与工具凭据留在各人机器上。

---

## 2. 账号、准入与角色

### 2.1 加入流程 **[产品约束]**

1. 用**公司邮箱**（`@guidefuture.com`）注册 —— 实例是 `ALLOW_SIGNUP=false` + 域名白名单 `guidefuture.com`，**子域不算**（`mail.guidefuture.com` 不通过）。
2. 登录方式是**邮箱验证码**（6 位、10 分钟有效，走公司 Lark SMTP）；**同一邮箱 60 秒内只能发一次**。
3. 由 owner/admin 在 Settings → Members 发邀请：**一次只能邀请一个邮箱**、有效期 **7 天**、邀请链接只出现在邮件里、**`owner` 不能被直接邀请**（须先进来再由 owner 提升）。
4. **[产品约束]** 邀请限流三道闸：邀请人 **100 次/10 分钟**、整个 workspace **200 次/24 小时**、同一邮箱 **10 次/24 小时**（2026-09-15 已从默认 10/50/6 调高并透传进容器）→ **20 人可以一次拉完，不需要分批**。

### 2.2 从哪个地址进 **[待办，当前最大的坑]**

| 地址 | 状态 |
|---|---|
| `http://192.168.20.131` | ✅ 当前主用（en0，DHCP） |
| `http://192.168.4.36` | ✅ 另一块网卡 en8，另一网段同事可用（已在 Origin 白名单里） |
| `http://192.168.20.9` | ❌ **已失效**（换网后地址变了） |

- 只有白名单里的地址能用：用别的 IP/主机名打开会 **HTTP 正常但 WebSocket 403**，表现为"能登录、界面不实时更新"。
- ⚠️ **[待办]** `/api/config` 目前仍返回**旧地址** `http://192.168.20.9`（`.env` 的 `MULTICA_APP_URL` 没跟着改）→ **网页上给出的 daemon 安装命令是错的**。在运维改掉之前，一律**手工指定**入口地址：

```bash
multica setup self-host --server-url http://192.168.20.131 --app-url http://192.168.20.131
```

- 地址由 DHCP 分配，**还会再变**（长期方案是固定 IP / DHCP 保留 + 内网域名）。地址一变，成员要做的事见 §14。
- 桌面客户端（mac/win 打包客户端）走另一份配置，见 §13。

### 2.3 角色分工 **[团队约定]**

| 角色 | 人数 | 职责 |
|---|---|---|
| `owner` | **2 人**（技术负责人 + 备份；系统强制"至少保留 1 名 owner"） | 删工作区、授予/收回 owner、最终裁决 |
| `admin` | 每条业务线 1 人（建议 ≤4） | 邀请成员、改工作区设置与 issue 前缀、GitLab/VCS 连接、自定义状态、自定义 label 字典、runtime profile、agent 上线复核 |
| `member` | 其余 | 建/改 project、建/改/删 issue、评论、建 label、建 squad、建 autopilot、跑自己有权限的 agent |

**[产品约束]** 角色只决定"工作区设置与团队管理"，日常协作对所有成员开放（任何成员都能建 project、删 issue）；**角色不授予运行权** —— 能不能跑某个 agent 由该 agent 自己的 Access 决定。

### 2.4 个人访问令牌（PAT） **[产品约束]**

- 前缀 `mul_`，代表**你的账号**、可访问你有权限的一切（**没有 scope 细分**）；`multica login` 默认建 90 天，剩余不足 7 天自动续期。
- `multica auth logout` / 网页登出**都不会在服务端吊销**，只会删本机配置/浏览器 cookie。要真吊销：Settings → API Token。

> **[团队约定]** ① PAT 等同密码，禁止写进仓库、脚本、issue、聊天；② 共享机器上不要留活跃 PAT，用完立即吊销；③ 在 run 里用的是任务级临时凭据（`mat_`，daemon 注入 `MULTICA_TOKEN`，最长 24h），不要自己另造 token。

---

## 3. 每人工位：CLI + daemon + runtime

### 3.1 装什么 **[产品约束]**

| 组件 | 要求 |
|---|---|
| Multica CLI | 含 daemon；`multica version` 验证（本机现为 0.4.43） |
| **≥1 个被支持的 AI 编码 CLI** | 支持 26 个命令名（`claude`、`codex`、`cursor-agent`、`omp`、`reasonix`、`kimi`、`opencode`…）。**一个都没有 daemon 不启动**；部分工具有最低版本门槛（Claude Code ≥2.0.0、Codex ≥0.100.0…） |
| 工具自身登录 | 必须在**同一系统账号、同一 PATH** 下已登录；工具凭据只留在本机，Multica 从不接收 |
| **GitLab 访问凭据** | 见 §3.3 —— agent 要用它 checkout 私有仓库、推分支、开 MR |

**[团队约定]** 团队统一以 **claude / codex / cursor-agent / omp** 为主力（本实例现有 runtime 覆盖这些），减少"同一条指令在不同工具上行为不同"的排障成本。

### 3.2 runtime 命名与可见性 **[团队约定]**

- **装完先改名**。默认名形如 `Claude (bogon)`、`Claude (macdeMacBook-Pro.local)` —— 一旦多台机器或主机名解析失败（`bogon`），你在界面上分不清哪台是哪台。
- 命名格式：**`<姓名拼音>-<机型>-<工具>`**，例 `zhangsan-mbp16-claude`、`lisi-t14-codex`。`multica runtime rename <id> "<新名>"`（`--machine` 可批量改同机全部 runtime）。
- **[产品约束]** runtime 默认 **private**：只有 runtime owner 能在其上建 agent，**workspace owner/admin 也不能破例**。要共享（例如一台常开机当公共编译机），**由 runtime owner 自己**改成 public；共享的只是路由，**不共享**工具登录凭据。
- 装了新工具或改了 profile：`multica daemon restart` 重新探测。

### 3.3 Git 凭据：每台 daemon 机器的硬要求 **[团队约定，源自自托管 GitLab 的工作方式]**

**[产品约束]** Multica 的 VCS 集成只做一件事：**只读**地把 MR 状态、CI 状态镜像到 issue 上（GitLab token 只需 `read_api`）。**checkout、推分支、开 MR 全部由 agent 在 runtime 上用宿主机自己的 git 凭据完成**。

于是每条规则都成立：

1. 每台要跑 agent 的机器，必须有能访问**该项目私有仓库**的凭据 —— **GitLab group deploy key / group access token（可写）/ 本机 SSH key** 三选一，推荐 group 级（70 个仓库一次配好）。
2. **凭据必须能推分支、开 MR**，否则 agent 只能改本地。
3. 这台机器的 git 权限 = agent 的权限 → **[团队约定]** 生产分支开**受保护分支 + MR 门**，不要让 agent 用能直推 `main` 的凭据；`CODEOWNERS` 指定每个 App 的审查人（20 人 ÷ 70 项目 ≈ 人均管 3–4 个）。
4. **[团队约定]** 上线前先跑通一次"**私有项目 checkout → 推分支 → 开 MR**"（这是后面所有方案的前置项）。

### 3.4 并发与机器纪律 **[产品约束 + 团队约定]**

| 层级 | 默认 | 本团队约定 |
|---|---|---|
| 单 daemon | 20 个 run 并发（环境变量 `MULTICA_DAEMON_MAX_CONCURRENT_TASKS`） | **4**（写进本机启动环境/profile） |
| 单 agent | 6 个 run 并发（agent 设置里可改） | **2**（大项目 1）；跨项目领域 agent 3–4。**不要设 20** —— Android 一次构建常吃 3–4 GB 内存，并发一高就互相 OOM |

- **[产品约束]** 超限不丢任务，只排队；同 agent 在同 issue 上强制串行；**同一本地目录（in_place 模式）串行**。
- **[团队约定]** 离机（合盖/休眠/断网）前确认没有在跑的 run：daemon 心跳 15s，异常退出后 runtime 约 **3 分钟**判离线，**在跑的 run 会失败**。长期休假请 `multica daemon stop`。
- **[产品约束]** 离线超过 **7 天**且没有任何 agent 绑定的 runtime 会被自动清理（回来会重建）。
- **[团队约定]** **不要设置 `MULTICA_DAEMON_PORT`**（会被误判成受管任务而拒绝登录）；多环境用命名 profile（`~/.multica/profiles/<name>/`）。

### 3.5 离职 / 移出成员 **[产品约束，必须按顺序处理]**

移除成员会**立即**：停用其拥有的 runtime → 归档绑定在这些 runtime 上的 agent → 取消未完成的 run → 清空其在 agent Access 里的授权（**重新加入也不恢复**）。

> **[团队约定] 离职回收顺序**：① 转移其 agent 的 owner 与绑定 runtime → ② 处理活跃 run（收尾或取消）→ ③ 再移除成员 → ④ Settings → API Token 吊销其 PAT。

---

## 4. Agent 规范

### 4.1 建多少、怎么分层 **[团队约定]**

70 个项目、70 个 agent 是基线（**约每项目 1 个**），再加少量跨项目角色：

| 层 | 命名 | Access | 用途 | 参考现状 |
|---|---|---|---|---|
| **项目 agent（基线）** | `<App 代号>-<职能>`，如 `member-dev`、`th2-android` | Entire workspace，或 Specific people（该 App 的 3–4 人） | 该项目的日常开发/修复 | 实例里 `MyNewProject 开发工程师`、`TH2 Android工程师` |
| **跨项目领域 agent** | `android-<能力>` | Entire workspace | 定时/横切任务 | 实例里的 `Android Pre-Push 代码安全检查`、`Doris崩溃ANR分析助手` 是好例子 |
| **攻坚 squad** | `<专项>-squad` | 按 leader 判定 | 跨模块专项：性能、崩溃治理、大版本升级 | 实例已有 2 个 squad |

**硬规矩**

- **命名必须语义化**：禁止 `测试1111`、`ceshi 11`、`测试2222` 这类名字（实例里已存在，需要清理）。中文职能名可以（可读性好），但要带 App 前缀。
- **系统 agent 不纳入治理口径**：`.multica-agent-builder-<uuid>` 是 Agent Builder 生成的，别去动它，也别算进"我们有几个 agent"。
- **不用了就归档**（归档会取消其全部未完成 run）；**测试 agent 用完当天归档**。
- **一个 agent 只有一个 owner**；owner 离职/换人前先转移归属。

### 4.2 Access：默认收紧 **[产品约束]**

| Access（界面） | 谁能运行 | 底层 |
|---|---|---|
| Only me（默认） | 仅 agent owner | `permission_mode=private` |
| Specific people | owner + 指定成员 | `public_to` + 成员白名单 |
| Entire workspace | 该 workspace 全部成员 | `public_to` + workspace 白名单 |

- **只有 agent owner 能改 Access**（workspace admin 也不行）；owner/admin 能看能管 agent，但**不能绕过 Access 去运行**别人的 Only me agent。
- 分配给别人、评论 `@`、squad 唤醒、channel 触发**都受同一套 Access 约束**；agent 之间互相触发按**链条顶端的人类**判定。

> **[团队约定]** 项目 agent 用 **Entire workspace** 或 **Specific people=该项目 3–4 人**；个人试验 agent 保持 Only me 并在 30 天内归档。

### 4.3 Instructions 必写八项 **[团队约定]**

1. **仓库与分支**：GitLab 仓库 URL、默认工作分支、禁止直推的分支（`main`/`release/*`）。
2. **模块结构**：有哪些 Gradle module、哪个是 app 入口。
3. **构建/测试命令**（照抄可用）：`./gradlew :app:assembleDebug`、`./gradlew testDebugUnitTest`、`./gradlew :app:lintDebug`。
4. **代码风格与静态检查**：ktlint/detekt/Android Lint 必须通过。
5. **硬约束**：不改 `release/`、不动签名配置、不改 `local.properties`、不升级 AGP/Gradle 主版本（除非 issue 明确要求）。
6. **产出与回写约定**：分支命名、MR 标题带 `PHOE-<编号>`、用 `Closes PHOE-<编号>` 收尾、评论里贴关键命令输出。
7. **停止条件**：验收标准未满足就继续；需要产品决策时**停下来提问**，不要自行扩大改动面。
8. **不要碰的东西**：生产配置、上架账号、付费服务开关、签名与 keystore。

> **[产品约束]** 描述字段（description）**只是展示**，永远不进执行 prompt；instructions 才是每次 run 都会给的。

### 4.4 配置纪律 **[产品约束]**

- **`custom_env` 与 MCP 配置明文存在服务端数据库**，只有 owner/admin 能解锁；每次读取/修改都留审计记录（无查询界面，属取证用）。→ 只放**低权限、可轮换**的 key。
- **改配置只影响后续 run**：skill / MCP / env / model 都在 runtime **领走 run 的那一刻**准备，不影响正在跑的 run。
- **复制 agent** 会带上 instructions、skills、custom args、并发上限、Access，但**从不复制** `custom_env`、MCP 配置、runtime 配置。
- **Skill 不做审查/签名/沙箱**，原样交给 agent；只从评审过的来源导入。
- **运行环境事实（安全相关）**：agent 默认以 daemon 用户的完整权限执行，**没有文件系统沙箱**，非交互模式下工具审批被自动应答（Claude Code `bypassPermissions`、Codex `danger-full-access`），并继承该用户真实 `HOME` 与 `XDG_*`。
- **清理**：不用了先**归档**，不要删除；归档 ≠ offline。

### 4.5 能力挂载要算 token 账 **[团队约定，源自长期记忆方案的结论]**

- **工作区 MCP 必须逐 agent 显式绑定**（`agent_mcp_server`），没有批量"全工作区生效"。而 **MCP 的工具清单是每轮常驻成本**。
  → **不要给 70 个 agent 各挂一堆 MCP**：每任务一份工具 schema × 70 agent，成本确定、收益不确定。
- 共享能力优先做成 **skill**（按 run 下发、可版本化）；需要取大量全文的场景用 **CLI 按需取**，而不是常驻 MCP。
- **[团队约定]** 团队级 MCP 服务（如代码检索、知识库）只给**确实需要的少数 agent** 挂；挂了要在项目群里说明理由。

### 4.6 新 agent 上线复核 **[团队约定]**

任何要设为"Entire workspace"的 agent，上线前由 1 名 admin 复核并留一条评论记录：

- [ ] 名字符合 §4.1 命名规范，description ≤255 字符
- [ ] Instructions 覆盖 §4.3 八项，构建/测试命令**在本机实跑过一次**
- [ ] Access 与用途匹配；owner 明确
- [ ] 并发上限已按 §3.4 收窄（不是默认 6 / 不是 20）
- [ ] `custom_env` / MCP 无高价值凭据
- [ ] 已挂载需要的团队 skill；没有为它单独堆 MCP
- [ ] 绑定 runtime 是**本人机器**或经同意的公共编译机

---

## 5. 项目建模：70 个维护项目的盘子

### 5.1 粒度与命名 **[团队约定]**

- **粒度：一个可独立构建/发版的仓库 = 一个 project**（70 个在维护项目 → ≈70 个 project）。同一 App 多仓库用"父 issue + 子任务"或"1 个 project 挂多仓库"解决。
- **命名格式：`[A] <App 名> · <仓库名>`**，例 `[A] 会员中心 · app-member`。前缀 `[A]` 标识 Android 团队，便于将来与其他端共处一个 workspace。
  project **没有 slug/key**，名字就是唯一标识 → **改名要在团队频道公告**。
- **[待办] 现有 4 个项目的名字不符合规范**（`Dome项目`、`Dome1`、`Myskills`、`ZM1A`），接入真实项目时统一改名。
- **台账**：admin 维护唯一台账（表格或 `PROJECTS.yaml`）：`project ↔ 仓库 URL ↔ 默认分支 ↔ 发版分支 ↔ 负责人 ↔ 审查人(CODEOWNERS)`。CLI 侧用 `multica project --help` 查子命令，`--output json` 便于脚本核对。

### 5.2 资源挂载（**最容易踩的坑**） **[产品约束]**

| 层级 | 作用 | 本团队规则 |
|---|---|---|
| **workspace 仓库清单**（Settings → Repositories） | **永远**进入每个 run 的上下文 | **只放公共库**（公共 UI/SDK/规范仓库，建议 ≤5 个）。**禁止把 70 个业务仓库塞进来** |
| **project resources** | 指定"这批活用什么代码、默认哪个 ref" | 每个 project 挂 **1–3 个** GitLab 仓库；`ref` 钉住默认/发版分支（`default_branch_hint` 只是提示，不强制切分支） |
| `local_directory` | 指向某台机器上的既有目录 | 每个 project **每台 daemon 至多 1 个**；只用于无法/不便 checkout 的场景 |

- **[产品约束]** 仓库类型只校验"是不是合法 git URL"（http(s) / ssh）→ **自托管 GitLab 地址今天就能直接挂**，不需要改代码。
- run 会把项目名、描述、资源清单写进上下文，并在工作目录落 `.multica/project/resources.json` → **[团队约定]** 把它加进仓库 `.gitignore`。
- **[待办]** 内网 GitLab 是 `http://192.168.5.254`，与 Multica 入口**不同网段**：成员机器必须能访问它，否则 agent 无法 checkout。

### 5.3 工作目录模式 **[产品约束]**

| 模式 | 行为 | 用在哪 |
|---|---|---|
| `github_repo` 类型资源（可填任意 git URL，含 GitLab，推荐） | runtime 负责 checkout，**默认 worktree**，同一仓库可并发 | **所有 Android 项目的默认选择** |
| `local_directory` + `in_place`（Direct） | 直接改你的工作副本，**同目录串行**，后来的 run 进入 `waiting_local_directory`；不自动切分支/stash/commit/push | 只用于"必须用本机未提交改动"的调试 |
| `local_directory` + `worktree`（Parallel） | 每个 run 一个 git worktree，并发不等；要求目录是 git 仓库且 runtime 声明能力 | 需要本地目录又想并发时**必须**显式用它 |

worktree 模式的三条事实（Android 团队必读）：

1. **agent 从"你看到的状态"开工**：未提交改动与未跟踪文件会重放进 worktree；**未跟踪内容超过 2000 个文件 / 200 MiB 时 run 直接失败** —— 这就是必须 gitignore 构建产物的原因（§6.2）。
2. **产物是一个分支，不是 MR**：每条 issue 一个分支 `agent/<agent>/<issue>`（chat 是 `agent/<agent>/chat-<session>`）；**Multica 从不替你 merge/push**，推分支与开 MR 由 agent 用宿主 git 凭据完成。
3. **后续 run 接着这条分支做**，只把你之后在本地的改动重放上去；你与 agent 改到同一行时由 agent 先解冲突，未解决就不交付。

### 5.4 project 字段纪律 **[团队约定]**

| 字段 | 规则 |
|---|---|
| status | `planned`（立项）/ `in_progress`（**在维护中 = 本值**）/ `paused`（暂停维护）/ `completed` / `cancelled` |
| priority | 用内置 5 档；`urgent` 仅限线上事故 |
| lead | 填 1 名负责人（**不是权限**，不会自动分派或自动运行） |
| 描述 | 会进入该项目下每次 run 的上下文 → 写架构要点、构建方式、flavor 矩阵、发布节奏、负责人、CODEOWNERS |
| 进度 | **系统自动算** `(done ∪ cancelled) ÷ 全部 issue`，不要手填 |
| 归档 | project **没有 archived 字段**：停维护 = `paused`/`cancelled` + label `state:archived` |

> **[产品约束]** 删除 project **不删 issue**（issue 变成无项目留存），但描述/状态/资源永久删除；**只有 owner/admin 能删**。**[团队约定]** 生产项目只归档，不删除。

---

## 6. Android 工程规范

### 6.1 构建环境 **[团队约定]**

| 项 | 规则 |
|---|---|
| JDK / Gradle / AGP | 与 CI 固定同一版本，写进 agent Instructions；不自升级主版本 |
| SDK 路径 | `local.properties`（`sdk.dir`）**不入库**；用本机 `ANDROID_HOME` 或 `~/.gradle/gradle.properties` |
| 依赖下载 | 走内网 maven 镜像或本机代理；代理变量写在**本机 shell profile**（daemon 继承用户环境），**不要**写进 `custom_env`（明文存服务端） |
| 内存 | `org.gradle.jvmargs=-Xmx3g`、`--max-workers=2`；与 §3.4 的"单机 2 个 Android 构建"配套 |
| 凭据 | keystore、signing 口令、内网 maven 凭据、Play Console service account **只在本机受控位置**，见 §8 |

### 6.2 `.gitignore` 底线 **[产品约束的间接要求]**

worktree 模式下未跟踪文件超限会让 run **直接失败**。所有 Android 仓库必须（至少）包含：

```gitignore
# Gradle
.gradle/
build/
**/build/
!**/src/**/build/
# Android Studio / IDE
.idea/
*.iml
local.properties
captures/
.externalNativeBuild/
.cxx/
# 产物
*.apk
*.aab
*.ap_
*.dex
*.jks
*.keystore
# Multica
.multica/
```

**[团队约定]** 项目接入前自查：`git status --porcelain | wc -l` 与未跟踪文件总量，别把 200 MiB 的 `build/` 带进 worktree。

### 6.3 agent 能跑的测试与不能碰的活 **[团队约定]**

| 可以交给 agent | 必须人来 |
|---|---|
| 单测（`testDebugUnitTest`）、Lint/ktlint/detekt、编译、依赖升级 MR、崩溃堆栈定位、资源/文案改动、小范围重构、CI 失败归因、发版说明草稿 | 签名与打包发版、上架、生产配置与开关、支付/账号类变更、数据库迁移上线、超过 ~20 文件或跨 3 个 module 的架构重写（**先建 design issue**） |
| instrumentation / 模拟器测试 | 默认**不交给 agent**：需要设备/模拟器、耗时长 |

### 6.4 分支与 MR 规范 **[产品约束 + 团队约定]**

- **建立关联**：把 `PHOE-123` 放进**分支名或 MR 标题**（忽略大小写，只匹配本 workspace 前缀）；只写在 body 里必须用 `Closes / Fixes / Resolves PHOE-123`，`Related to PHOE-123` 这类裸提**不算关联**。一个 MR 可关联多个 issue。
- **自动置 Done 的三个条件缺一不可**：① 至少一个**已合并**的关联 MR 用了紧贴编号的 close intent；② 该 issue 没有其他仍 Open/Draft 的关联 MR；③ issue 当前不是 done/cancelled。→ 只把编号写进分支名只是建立了关联，**不会**自动完成。
- 命名：**人工分支** `feat/PHOE-123-简短描述`、`fix/PHOE-123-...`；**agent 分支**由系统给 `agent/<agent>/<issue>`，不要手动创建/改写同名分支（Multica 靠自己的记录续跑）。
- **[团队约定]** 所有 MR 至少 1 名人类 reviewer（CODEOWNERS 指定）；**agent 自己写的 MR 不得由同一个 agent 自审自合**；目标分支用受保护分支。
- **[产品约束] VCS 集成是只读的**：Multica 只镜像 MR / CI 状态，**从不** push commit、写状态检查。面板上的 CI 与可合并性来自快照，GitLab 暂时不可用时会保留最后快照并标 stale。

---

## 7. issue、触发与日常协作

### 7.1 issue 规范 **[团队约定]**

- **编号**：`PHOE-<数字>`（全 workspace 共用一套自增编号，当前已到 `PHOE-19`）。**[产品约束] 不要改 issue 前缀** —— 一改，历史 issue 全部按新前缀重新显示，旧编号在应用内**不再可解析**，历史 MR/分支/聊天里的旧编号也不会跟着变。
- **标题**：`[模块] 动词 + 现象/目标`，例 `[会员] 修复支付回调重复发货`。
- **描述**：背景 / 复现步骤 / 期望结果 / 验收标准 / 涉及仓库与分支（模板见附录 B）。
- **状态**：**只用 7 个内置状态**（`backlog` `todo` `in_progress` `in_review` `blocked` `done` `cancelled`）。
  → 自定义状态是 **workspace 级共享**且**不继承**内置状态的关键行为（不继承 backlog 停车、不继承 in_review 收尾、不继承 blocked 失败、不继承 in_progress 回滚），看板还会因此多出列。**当前实例里两个 workspace 都只有 7 个内置状态**，团队决定：**不新增**，需要区分时用 label。
- **优先级**：用内置 `priority` 字段，**不要**用 label 表达优先级。
- **截止日期**：`YYYY-MM-DD`。
- **父子与批次**：父 issue 保整体目标，子 issue 独立推进（父子状态互不联动）；分批推进用子任务的 **stage**，一个 stage 全部完成会通知父 issue，父 issue 的 agent 会被唤醒决定下一步。

### 7.2 label 字典 **[团队约定]**

label 是 workspace 级共享资源（≤32 字符、同类型内大小写不敏感唯一）。当前实例只有 `Bug` 一个 issue label，**接入真实项目时按下表统一**：

| 类别 | 取值 |
|---|---|
| 类型 | `type:bug` `type:feature` `type:chore` `type:refactor` `type:incident` |
| 领域 | `area:build` `area:ui` `area:network` `area:payment` `area:perf` `area:crash` |
| 状态 | `state:blocked` `state:needs-design` `state:archived` |
| 协作 | `agent-ready`（信息足够，可直接派给 agent）、`human-only`（禁止派给 agent） |

**[团队约定]** 新增 label 前先查字典，禁止近义重复（`Bug` vs `type:bug` 只能留一个）。

### 7.3 什么时候用什么触发方式 **[产品约束 + 团队约定]**

agent **不会自己开始工作**，每一次 run 都由显式动作触发：

| 触发 | 何时用 | 关键规则 |
|---|---|---|
| **指派 issue**（主线） | 有明确验收标准的开发/修复任务 | **[产品约束]** issue 不在 `backlog` 才立即起 run；可勾 **Don't start yet** 只落 assignee。改 assignee 会为新 assignee 起新 run；**取消指派不起 run** |
| **评论 @agent** | 追加讨论、评审反馈、追问 | **[产品约束]** 不改 assignee、不改状态；同一 agent 已有等待中的 run 时新评论**合并**进该 run |
| **Web chat** | 探索、查代码、一次性问答 | **[产品约束]** 完全私有（admin 也看不到）、不会自动建 issue；每条消息一次 run |
| **Channel（飞书/Lark）** | 群里快速起 issue、随手问 | **[产品约束]** 群里**必须显式 @Bot**，未 @ 的消息既不触发也不进上下文；**一个 Bot 绑一个 agent**（70 个 agent 不等于要装 70 个应用 —— 只给需要聊天入口的装）；首次使用要先绑定账号 |
| **Autopilot** | 定时/事件驱动的例行工作 | **[产品约束]** `create_issue` 先建 issue 再跑；`run_only` 直接跑、runtime 离线则跳过 |
| **CLI** | 脚本、批量 | `multica issue assign/comment add/rerun/status`、`multica autopilot trigger` |

**[团队约定] 送进 agent 之前先自检**：验收标准可判定吗？范围不超过 20 文件 / 3 module 吗？需要的人为凭据（内网 maven、设备）agent 拿得到吗？有一条是否，先补 issue 或打 `human-only`。

### 7.4 backlog 与"启动"语义 **[产品约束]**

- `backlog` = 停车场：**只有停在 backlog 的 issue 不触发 agent**；把它移到其他状态**本身就是触发器**（会弹确认框，可选 Don't start yet）。
- 已经 `done`/`cancelled` 的 issue 被指派给 agent 时**仍会立即起 run**。

### 7.5 评论与通知 **[产品约束]**

- `@` 必须从建议列表里选，**手打纯文本无效**；编辑器下方有**触发预览**，可临时取消勾选本次触发（不改 Access）。
- **只讨论不触发**：评论以 `/note` 开头（成员之间的备注）。`@all` 也会关掉本条评论对 assignee 的自动触发，但它**会通知全体成员**（不含 agent）—— 它是广播，不是静默留言。
- **回复的自动路由**：回复某 agent 的评论 → 交给它；回复在它已参与的讨论里 → 保持它；都不匹配的顶级评论 → 落到 issue 的 agent assignee。**显式 @ 了任何人，assignee 兜底就失效。**
- inbox 是个人通知中心（不是活动日志）：订阅来源 = 创建者 / assignee / 评论者 / 被 @ 的人；**自己的动作不会通知自己**；同一 issue 多条通知会合并；可在设置里按 6 组关闭（assignments / status / comments / mentions / updates / agent activity；reactions 与 autopilot 类始终投递）。
- **agent 没有 inbox**：`@` agent 是直接建 run。

### 7.6 run 生命周期：成员必知 **[产品约束]**

| 状态 | 含义 | 你要注意 |
|---|---|---|
| `queued` | 等 runtime 领走 | 只要 runtime 还心跳就无限等；**一直 queued 通常是 runtime 离线 / 工具没登录 / 并发满** |
| `dispatched` | 已领走、正在拉起工具 | 超 **5 分钟**判失败 |
| `waiting_local_directory` | 同一本地目录被别的 run 占着（仅 in_place 模式） | 改 worktree 可彻底消除 |
| `running` | 执行中 | **没有时长上限**，靠 runtime 心跳维持 |
| `completed` / `failed` / `cancelled` | 结束 | `completed` ≠ issue 完成 |

- **停止**：改 assignee / 改状态都**不会**停掉已开始的 run；要停在 issue 右侧**执行日志**里停对应 run。
- **重试**：只有瞬时故障自动重试（runtime 离线、daemon 重启回收、平台超时、Codex 无输出停滞、skill bundle 下载失败 → 默认重试 1 次；工具网络中断 → 最多 3 次）。**agent 自身报错（凭据过期、配额、模型不可用、上下文超限）不自动重试**，先修因再手动重试。
- **失败回滚**：失败且该 issue 无其他活跃 run、无待重试时，`in_progress` 会打回 `todo`。
- **手动重试** 用当时跑该 run 的 agent（即使 issue 已换 assignee），并尽量保留本地目录里的产出。

### 7.7 Autopilot 规范 **[团队约定]**

只做"例行、可自动判定"的事，命名 `ap-<用途>`：

| 例子 | 模式 | 频率建议 |
|---|---|---|
| `ap-ci-failure-triage` | create_issue | 每班次 |
| `ap-deps-patch-bump` | create_issue | 每周 |
| `ap-mr-awaiting-review`（催评审） | run_only | 每个工作日 |
| `ap-release-notes` | create_issue | 每次发版窗口 |

**[产品约束]** autopilot 默认保存即启用；过去 7 天内 ≥50 次运行且失败率 ≥90% 会自动暂停并通知创建者；`run_only` 不自动重试。

### 7.8 成本纪律 **[团队约定]**

- 用量**可查**：`multica runtime usage <runtime-id>`、`multica issue usage <issue-id>`、runtime 详情页趋势图（底层是 `task_usage_hourly`，**小时级汇总，有延迟**）。
- **[产品约束]** 自托管**没有** token 费用硬闸门（配额与席位门禁只在 Cloud 侧），靠纪律：
  1. 单 agent 并发 2、单 daemon 4（§3.4）；
  2. 任务必须带验收标准，禁止"让 agent 自己看看"这类无边界任务；
  3. autopilot 频率按 §7.7，不叠加；
  4. MCP 不滥用（每轮常驻 schema，§4.5）；
  5. 每周例会看一次成本趋势。

---

## 8. 安全与合规红线（违反视为事故）

1. **agent 以你的 OS 用户身份、无沙箱执行**：能读写你能读写的文件、使用你的凭据、无限制访问网络，并继承真实 `HOME`。→ 跑 daemon 的机器上不要放生产密钥、签名 keystore、Play Console 凭据、云管理员凭据。
2. **agent 用宿主 git 凭据推分支开 MR**：那台机器的 GitLab 权限就是 agent 的权限。→ 用 **group deploy key / 可写 group token**，配合**受保护分支 + MR 门 + CODEOWNERS**；不要让 agent 拥有直推 `main` 的能力。
3. **`custom_env` 与 MCP 配置明文存在服务端数据库**：只放低权限、可轮换的凭据。
4. **PAT（`mul_`）等同账号密码且无 scope**：不写进仓库/脚本/日志；登出**不吊销**，必须去 Settings → API Token。
5. **密钥绝不入"知识/记忆"载体**：keystore 口令、签名文件路径、Play Console service account JSON、内网 maven 凭据、CI 变量 —— **skill、知识库仓库、issue 评论、chat 里都不许出现**；落 git 之前过 pre-commit + secret 扫描，MR 审查是第二道门。
6. **数据边界**：服务端保存 issue、评论、agent 配置、run 上下文与记录、agent 回写内容、附件（**可能包含 agent 读进去的代码片段**）。自托管 ≠ 不出内网：**服务端 LLM 层会把新会话的"首条用户消息原文"发给所配置的端点（当前指向外部 DeepSeek）**做自动标题。→ 涉密信息不要写进 web chat 首条消息，也不要通过 chat 交给 agent。
7. **渠道消息也在服务端**：飞书/Lark 群里的对话进入 Multica 会话并可被检索；断开 Bot 不删除既有对话与审计记录。→ 群里不贴密钥与个人敏感信息。
8. **附件上限 100 MB**；上传公司代码、日志、崩溃 dump 前确认合规。
9. **Skill / MCP / 插件来源必须评审**：不做审查/签名/沙箱，原样交给 agent。
10. **服务器运维不在成员职责内**：备份、升级、配置变更找运维。**不要自己起/停这套栈** —— 少给一个 `-f` 就会**静默丢配置**（详见 `TEAM_DEPLOY_PLAN.md` §11 P0-2），统一走运维提供的脚本/流程。
11. **离职**：按 §3.5 顺序回收；PAT 必须吊销。

---

## 9. 命名与命名空间速查

| 对象 | 规则 | 例 |
|---|---|---|
| workspace | 全局唯一 slug，创建后不可改 | `土卫九` / `phoebe-1yu1` |
| issue 前缀 | `^[A-Z0-9]{1,10}$`，**定了就不改** | `PHOE` |
| project | `[A] <App 名> · <仓库>`（无 slug/key） | `[A] 会员中心 · app-member` |
| agent（项目） | `<App 代号>-<职能>`，workspace 内唯一 | `member-dev`、`th2-android` |
| agent（跨项目） | `android-<能力>` | `android-pre-push-scan` |
| runtime | `<姓名拼音>-<机型>-<工具>`（**必改默认名**） | `zhangsan-mbp16-claude` |
| squad | `<专项>-squad` | `perf-squad` |
| skill | `<领域>-<用途>` | `android-build-rules` |
| autopilot | `ap-<用途>` | `ap-ci-failure-triage` |
| 人工分支 | `<type>/PHOE-<编号>-<描述>` | `fix/PHOE-421-pay-callback` |
| agent 分支 | 系统生成，不要仿造 | `agent/android-dev/PHOE-421` |
| label | `type:` / `area:` / `state:` + 小写词，≤32 字符 | `type:bug` |
| saved view | `<团队>-<用途>` ≤80 字符 | `Android-待评审` |

---

## 10. 度量与巡检

**每周（admin 轮值，10 分钟）**
- 运行积压：多少 `queued` / `waiting_local_directory`，卡在谁那里；
- 失败归因：`failure_reason` Top 3（`provider_auth_or_access` = 工具没登录；`runtime_offline` = 有人机器掉了；`context_overflow` = 任务切得不够小）；
- runtime 在线率：谁长期离线；
- 成本趋势：runtime 用量页（注意小时级汇总延迟）；
- 未关联 MR 的 issue：`in_review` 却没有 MR 卡片的，要求补编号。

**每月（owner + admin）**
- **agent 盘点**：30 天没跑过的归档；Access 与 owner 复核；清理测试 agent 与失效系统 agent 之外的噪声；
- label / 状态字典清理：合并近义 label；
- project 状态与新项目接入对齐（台账 ↔ 实际）；
- 规范修订：把本月踩到的坑补进 §11。

---

## 11. 已知限制与坑（成员会感知到的表现）

| 现象 | 真实原因 | 你该做什么 |
|---|---|---|
| 按网页给的命令装 CLI/daemon，连不上 | **[待办]** `/api/config` 仍返回**旧地址** `http://192.168.20.9`（`MULTICA_APP_URL` 未更新） | 手工写 `--server-url http://192.168.20.131 --app-url http://192.168.20.131`（或另一网段 `192.168.4.36`）；让运维改配置 |
| 能登录，但评论/状态**不实时更新** | WebSocket 被拒：你用的地址不在白名单 Origin 里 | 用 §2.2 的地址；换网络/换 IP 后找运维同步白名单与 `.env` |
| 点"发送验证码"没反应/收不到 | 同一邮箱 60 秒限一次；或服务端 SMTP 配置问题（配错时验证码既不发信也不打日志） | 等 60 秒重试；仍不行找 owner/admin |
| 登录时说"太频繁" | **[待办]** 认证限流仍按前端口径统计 → 全团队共用一个桶（`/auth/*` 直连后端还没做） | 等一会儿；频繁出现找运维（`TEAM_DEPLOY_PLAN.md` §11 P0-4） |
| 指派了但任务不动 | runtime 离线 / 工具未登录 / agent 并发满 / daemon 全局并发满 | `multica runtime list`、`multica daemon status`；工具要在本机终端手动跑通一次 |
| 任务显示在等目录锁（`waiting_local_directory`） | `in_place`（Direct）模式同目录串行 | 改用 worktree（Parallel）；必要时 `multica daemon restart` 释放内存锁 |
| run 失败提示工具鉴权/不可用 | daemon 在线 ≠ 工具可用（未登录/没配额） | 在本机终端登录该工具后重试 |
| agent 改完了但推不动 / 开不了 MR | 宿主 git 凭据没有该仓库权限，或不能写 | 按 §3.3 给该机器配 GitLab group deploy key / 可写 group token |
| worktree run 起不来 | 未跟踪文件超过 2000 个 / 200 MiB（通常是 `build/`） | 补 `.gitignore` 并清理构建产物 |
| Usage 页面是 0 | 读的是小时级汇总，有延迟 | 稍等或换 `multica runtime usage` |
| 附件上传失败 | 超过 100 MB / 反代请求体限制 / 存储目录不可写 | 压缩后重传，仍失败找运维 |
| 群里的 Bot 没反应 | 消息没 @Bot（群内必须显式 @）或账号未绑定；或该 agent 没绑 Bot | 首次使用先完成账号绑定；确认这个 agent 确实装了 Lark 应用 |
| 界面上分不清哪台机器 | runtime 还是默认名 `Claude (bogon)` / 主机名解析失败 | 按 §3.2 改名 |
| 我改的/agent 改的代码"消失"了 | worktree 模式下产物在 `agent/<agent>/<issue>` 分支上，不在你的工作副本 | 用 `git branch` 找 `agent/` 前缀分支；分支名也可在 run 详情面板复制 |
| 归档 agent 后任务全没了 | **[产品约束]** 归档会取消该 agent 全部未完成 run | 归档前先处理在跑的 run |
| 移出成员后 agent 全挂了 | **[产品约束]** 移除成员会停 runtime、归档其 agent、取消 run | 按 §3.5 顺序回收 |
| 打字 `@某人` 没触发 agent | 手打纯文本不是 mention | 从建议列表里选；或用 CLI `issue comment add` |
| daemon 日志里有历史 `close 1006` / `i/o timeout` | **旧版客户端（v0.4.42）自身的 bug**，与"daemon 是否经前端代理"无关；v0.4.43 后已基本消失 | 先 `multica update` 升级 CLI；仍频繁失败再找运维 |
| 保存视图/看板越用越卡 | issue 列表接口默认 20/上限 50、表格端点 50/上限 100 且有 8 秒超时；`open_only`、project 列表不分页 | 用筛选与保存视图收敛范围，别用"全部 issue"当工作台 |

---

## 12. 落地清单（照做即可）

**A. 新成员接入（约 30 分钟）**
- [ ] 用公司邮箱在入口地址注册并登录（owner/admin 可直接发出邀请，**不必分批**）
- [ ] 装 CLI + 至少一个 AI 编码 CLI，并完成该工具自身登录
- [ ] `multica setup self-host --server-url http://192.168.20.131 --app-url http://192.168.20.131` → `multica daemon start`
- [ ] Runtimes 页确认自己的 runtime 在线，**按 `<姓名拼音>-<机型>-<工具>` 改名**
- [ ] 配好 GitLab 访问凭据（§3.3），跑通一次私有仓库 checkout
- [ ] 读完 §0、§8；加入团队 Lark 并完成 Bot 账号绑定；（可选）按 §13 配桌面客户端

**B. 新项目接入（每个 project 5 分钟）**
- [ ] 台账登记：GitLab 仓库 URL / 默认分支 / 发版分支 / 负责人 / CODEOWNERS 审查人
- [ ] 建 project，命名 `[A] <App> · <仓库>`，status = `in_progress`，lead = 负责人，描述写架构、构建方式、flavor 矩阵
- [ ] 挂 project resources：GitLab 仓库 + `ref`；需要本地目录时用 `worktree` 模式
- [ ] 核对 `.gitignore`（§6.2）并清理未跟踪产物
- [ ] 加 label：`type:*` + 对应 `area:*`；确认该仓库**没有**被加进 workspace 仓库清单
- [ ] 建项目 agent（§4.6 复核），Instructions 用附录 C 模板

**C. 新 agent 上线**：见 §4.6 复核表（并发收窄、Access、凭据、skill、runtime）。

**D. 离职回收**：转移 agent 归属与 runtime → 处理活跃 run → 移除成员 → 吊销 PAT。

**E. 服务器换机 / 换 IP 当天**：见 §14。

**F. 每周 / 每月**：按 §10 执行并留档。

---

## 13. 桌面客户端接入（macOS / Windows）

打包客户端默认连官方云 `https://api.multica.ai`，**没有 UI 切换服务端**；指向自建实例只有一条路：在家目录放 `desktop.json` 后**完全退出重启**。

```json
{ "schemaVersion": 1, "apiUrl": "http://192.168.20.131",
  "wsUrl": "ws://192.168.20.131/ws", "appUrl": "http://192.168.20.131" }
```

| 平台 | 路径 | 写文件的坑 |
|---|---|---|
| macOS | `~/.multica/desktop.json` | 用 `cat` 写，别用图形编辑器（会带 BOM） |
| Windows | `%USERPROFILE%\.multica\desktop.json` | 必须 `WriteAllText`（`>`/`Out-File` 会写成 UTF-16）；**记事本会补 `.txt`** → 客户端静默连回云端且不报错 |

- **[产品约束]** 只在启动时读一次，**必须 ⌘Q / 完全退出重启**。
- 验证：设置 → 桌面应用 → 守护进程，`serverUrl` 应显示配置的地址；`~/.multica/profiles/` 下会出现 `desktop-<host>/`。
- 首次连接要重新登录一次（内网走邮箱验证码）。
- 回滚到云：删掉 `desktop.json` 重启。
- 细节与故障速查见 `DESKTOP_CLIENT_CONFIG.md`。

---

## 14. 服务器换机 / 换 IP 时的成员动作

地址一变（换机、换网、DHCP 续到新 IP），**所有人和所有机器都要重指**。完整迁移见 `MIGRATION_TO_NEW_HOST.md`，成员侧只需：

1. **本机 daemon**：把 `server_url` 指向**入口地址（80 端口）**，不是服务器的 `:3100`（那两个端口只绑服务器回环）。

```bash
multica config set server_url http://<新入口>
multica daemon restart && multica daemon status
```

2. **两种上线方式**（由运维决定用哪种）：
   - **携带身份（推荐，保住 agent 绑定）**：从旧机拷 `~/.multica/config.json` + `~/.multica/daemon.id`（`server_url` 改新地址）→ runtime 行**原地复用**，agent 的 runtime 绑定不断。
     ⚠️ **同一 `daemon.id` 不能两台机器同时在线**（会互相抢任务）：旧机必须先 `multica daemon stop`。
   - **干净重登**：直接 `multica setup self-host --server-url <新入口> --app-url <新入口>` → 注册出**新 runtime**，**每个 agent 需要在界面上重新选一次 runtime**（含绑了 Lark 的那个）。
3. **新机器至少装一个被识别的 CLI**，否则 daemon 起来了但 `Agents:` 为空，任务会报 "no runtime"。
4. **桌面客户端**：改 `desktop.json` 里的三个地址并重启（§13）。
5. **确认入口白名单**已包含新地址，否则"能登录、不实时更新"。
6. 迁移期间：旧机数据在新机验收通过前**不要回收**（两边会各自分叉）。

---

## 附录 A：成员常用 CLI 速查

```bash
# 身份与状态
multica auth status                 multica daemon status
multica runtime list                multica daemon logs --follow
multica runtime rename <id> "zhangsan-mbp16-claude"

# issue
multica issue list --status in_progress --project <id> --output json
multica issue get PHOE-123
multica issue create --title "[会员] …" --description-stdin < issue.md
multica issue assign PHOE-123 --to "member-dev"            # 起 run
multica issue assign PHOE-123 --to-id <uuid> --no-start     # 只落 assignee
multica issue status PHOE-123 in_progress --no-start
multica issue comment add PHOE-123 --content-stdin < review.md
multica issue runs PHOE-123 --active --output json
multica issue run-messages <task-id> --since 42 --output json   # 增量看 transcript
multica issue cancel-task <task-id> --issue PHOE-123
multica issue rerun PHOE-123
multica issue pull-requests PHOE-123
multica issue usage PHOE-123

# 仓库 / 工具
multica repo list | add | remove | checkout <git-url> [--ref <branch>]
multica runtime usage <id>
multica skill list | import --url <url> | create
multica agent skills set <agent-id> --skills <a,b>       # 覆盖；add 为追加
multica autopilot trigger <id>
multica workspace member invite <email> --role member
multica config set workspaces_root <path>                # 工作目录根，默认 ~/multica_workspaces
multica config set server_url http://<入口>              # 换地址后重指
```

约定：脚本里一律 `--output json`；长文本走 `--*-stdin` / `--*-file`；退出码 `0/1/2/3/4/5` = 成功/通用/网络/鉴权/未找到/校验。

**附录 B：issue 描述模板**

```markdown
## 背景
（为什么做、影响谁）

## 现象 / 目标
（可观察的现象，或达成后的样子）

## 复现步骤
1. …
2. …

## 期望结果 / 验收标准
- [ ] …
- [ ] 单测通过；Lint/ktlint 无新增告警

## 涉及仓库与分支
- GitLab：`http://192.168.5.254/…/app-member.git`  默认分支 `develop`  发版分支 `release/2.8`

## 边界
- 不要动：签名配置、release/*、AGP 版本
- 需要人决策的点：…
```

**附录 C：agent Instructions 模板（Android）**

```markdown
你是 Android 团队 <项目> 的开发 agent。

## 工作方式
1. 先读 issue 描述与评论，确认验收标准；不清晰就在评论里提问并停止。
2. 在 worktree 里工作（分支 agent/<你>/<issue>），改动前先跑通基线构建。
3. 完成后：跑 `./gradlew :app:assembleDebug testDebugUnitTest :app:lintDebug`，
   把关键输出贴回 issue 评论；用 `multica issue status PHOE-<n> in_review --no-start` 收尾。
4. 需要交付代码时，用本机 git 凭据推分支并开 MR（标题 `[PHOE-<n>] <简述>`，正文含 `Closes PHOE-<n>`）。

## 硬约束
- 只改本 issue 范围内的代码；超过 20 文件或跨 3 个 module 时先停下来提问。
- 禁止：升级 AGP/Gradle 主版本、改 `release/*`、动签名与 `local.properties`、改生产配置。
- Kotlin 代码遵循项目 ktlint/detekt 规则；不引入新依赖除非 issue 明确要求。

## 交付
- MR 标题：`[PHOE-<n>] <简述>`，正文含 `Closes PHOE-<n>`。
- 评论里说明：改了什么、怎么验证的、还有什么风险。
```

**附录 D：规范条款的事实来源**

- **产品行为**（权限、触发、状态、并发、MR 关联、worktree、限流、PAT、渠道等）：仓库 `apps/docs/content/docs/**`、`server/internal/**`、`server/migrations/**` 核对。
- **实例约束**（登录取码、域名白名单、Origin 白名单、邀请限流、GitLab 连接、已知缺陷）：本机实测 + `TEAM_DEPLOY_PLAN.md` §4/§9/§10/§11、`MIGRATION_TO_NEW_HOST.md`、`DESKTOP_CLIENT_CONFIG.md`。
- **70 项目 / 70 agent / GitLab 条件下的工程约定**：`LONG_TERM_MEMORY.md` §8（含 §8.5 Android 两条红线）。
- 出现冲突时：以 `TEAM_DEPLOY_PLAN.md` 的实测记录为准。

---

## 变更记录

| 版本 | 日期 | 变更 |
|---|---|---|
| v1 | 2026-09-15 | 首版（按"20 人 / 60 个项目"编写） |
| v2 | 2026-09-17 | ① 项目口径 60 → **70**，agent 基线改为**每项目 1 个（≈70）**；② 代码托管由"待配 GitHub App"改为**自托管 GitLab（已连接）**，新增宿主 git 凭据要求与受保护分支+MR 门纪律；③ 入口地址改为当前 `192.168.20.131`（`192.168.20.9` 已失效），并新增 **[待办]** `/api/config` 仍吐旧地址导致安装命令不可用；④ 邀请限流已放宽到 100/200/10，删除"分两批邀请"的旧约定；⑤ 修正旧判断：daemon 的 `close 1006` 是 v0.4.42 客户端 bug、与前端代理无关；⑥ 新增 §13 桌面客户端接入、§14 换机/换 IP 时成员动作、§4.5 MCP token 账纪律；⑦ 补入实例现状快照与实测数字，明确"系统 agent / 测试 agent 的清理"与现有 4 个项目名待改名 |
