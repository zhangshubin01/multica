# Multica 长期记忆接入研究

> 2026-09-30 · 版本 v2.4（补候选）· 依据三类：本仓库与本机实例的实测、上游 issue/PR 一手记录、全网调研（各 harness 官方
> 文档、记忆系统实现、论文与基准）。
> **怎么读**：只要结论与选择 → §0；要动手 → §4 + §5；要复核任何一条论断的出处 → 附录 B；假设的核实状态 → 附录 C。
> **证据分级**：未特别标注 = 本仓库实测或一手文档；推断处写明「推断」；未验证处标 ❓。
> **v2.1 补了什么**：v2 的 §3 只留了 7 行候选，把底稿里研究过的 MemOS、supermemory、LangMem、MemoryOS、Zep、
> letta-code、memobase 的正文信息**丢了**（其中 MemOS 在底稿里是第一顺位备选）。v2.1 把候选补回，并新增
> **§3.2 完整筛查清单**——把本轮逐条核实过的 **150+ 个候选**（14 个平台型 + 18 个编码 agent 专用 + 10 个研究型
> + 中文生态 / 云厂商 + MCP registry 41 个记忆 server + 榜单 82 项）全部落表，标注结论，**杜绝再次丢失**。

> **v2.2 清理了什么（本轮）**：按两条规则重过一遍候选——① **最后提交早于 2026-08-17**（一个月）的库；② **不能自部署**的库
> （云厂商托管、闭源引擎、无服务端形态）。命中即从 §3.1 / §3.2 各表**移出**，但**不静默删除**：20+ 项连同**最后提交日期**、
> 命中规则、以及原本的价值（token-savior 的 `hooks/` 样板、LycheeMem 的两个设计…）全部留痕在**新增的 §3.2.1**。
> 日期为 2026-09-17 用 GitHub commits API 逐个读出的**最后提交**（`updated_at` 随 star 刷新，不作数）。
> 清理后仍列在表里的：平台型 9 · 编码 agent 专用 13 · 研究型 7 · 论文实现 3 · 中文生态 6 · MCP registry 2。
> **本轮检索的边界（必读）**：通用搜索引擎（DuckDuckGo / Bing）已被反爬拦下或返回无关结果，**本轮没有用过网页搜索**；
> 所有新发现来自 **GitHub API + 官方 MCP registry API + 各项目官方文档 / 仓库原文**。因此"全网"有边界：
> **已收录进公开榜单、被打过 topic、或发布在 MCP registry 的项目覆盖较全；只在博客 / 推文 / 公众号里流传的项目会漏。**

> **v2.3 核实了什么（本轮）**：把**附录 C 的 19 条未验证假设逐条做了一轮核实**——手段是「只读代码 + 本机实机 + GitHub API」，
> 不做任何仓库改动。结果：**15 条闭环**（其中 6 条推翻了正文原有结论）、**1 条部分核实**（U3：TAM 链路的**每一跳都实测了**，只差"真实 Multica 任务"这一步）、
> **3 条仍需授权或生产实例**（U12/U16/U18）。
> 实机跑过的：`ReMe` 0.4.1.12、`memu-cli` 0.11.0b3、`MemPalace` 3.10.0、`hindsight` 容器、**TAM 三件套**（U3/U17）、
> claude 2.1.274 的插件 hooks 隔离实验。**被推翻/需要收窄的结论有 6 处**，都已在正文就地标注「v2.3 修订」：
> ① claude 的用户级插件 hooks **会**触发（§4.1 D、§8.2-⑩、§8.6、§8.8）；② Reasonix 的每轮注入只能走 `AGENTS.md`（§4.1 D）；
> ③ 架构 E「覆盖全部 provider」**证伪**（§4.1 E）；④ memU 的「不需要额外 LLM」只对蒸馏成立（§3.1.3）；
> ⑤ ReMe 的 LongMemEval 89.4% 跑在**自家修订过的 ground truth** 上（§3.1.3/附录 C U11）；
> ⑥ 每轮索引块 **≈900 tokens**，不是假设的 ≤300（§3.1.3、§8.6、§8.9）。
> 逐条状态与依据见**已改写的附录 C**（「未验证假设 → 假设与核实结果」）。

> **v2.4 补了什么（本轮）**：v2.2 那句「本轮没有用过网页搜索」在**这一轮被部分改写**——找到并实测了**可用的聚合检索通道**
> （Hacker News Algolia API、Lobsters `search.json`、Marginalia），不再只靠 GitHub API。由此**新捞到 8 个此前筛查清单里没有的候选**，
> 全部落表在**新增的 §3.1.4**；其中一个（**OKF Agent Memory**）**已在本机实机验证过**（含"按项目物理隔离"与中文检索的实测缺陷）。
> **结论不变**（平台后端仍推 ReMe），但 §3.1.4 给出了**明确的翻转条件**与一个"当架构 A 契约层实现来试"的落地路径。
> Firecrawl 的搜索档这次**不可用**（vault 里没有 `FIRECRAWL_API_KEY`，免 key 档报 `Alexandria requires a Firecrawl API key with access enabled`）。
> **v2.4 还补了 §4.5「架构 A+B 接入手册（ReMe）」**——把"怎么把 ReMe 配到 Multica 上"写成可照做的五步，
> 含实测的启动前置（`LLM_API_KEY` 必填）、MCP 端点形状（`POST /mcp`，serverInfo `ReMe 3.4.7`）、
> `path_prefix` 项目隔离、以及四个"会让你误判记忆没生效"的坑。

> **已给出的具体条件**（20 人 / 70 个 Android 项目 / 自托管 GitLab / Lark 国际版 / 70+ agents / 编码时使用）→ 直接看
> **§8 场景适配**：它把上面的结论落成一套可执行的方案，并给出四个被改掉的默认（真相层放 GitLab、作用域改 (workspace, project)、
> 不给 70 个 agent 挂 MCP、**注入点改到每一轮**）。

## 目录

- [0. 速览](#0-速览) — 结论、该选哪条路、五条设计约束
- [1. 现状](#1-现状) — Multica 有什么、没有什么、上游什么立场
- [2. 选型依据一 宿主能力](#2-选型依据一-宿主能力) — 注入点在哪一层
- [3. 选型依据二 记忆后端](#3-选型依据二-记忆后端) — 候选对比 + **全网检索新候选与 OKF 实测（§3.1.4）** + 完整筛查清单 + **两条规则的剔除记录（§3.2.1）** + 已排除与澄清 + 腾讯 TAM 专题
- [4. 五种接入架构](#4-五种接入架构) — 对比、推荐、C 的改动清单、**E 的接入手册（TAM）**、**A+B 的接入手册（ReMe，§4.5）**
- [5. 落地计划](#5-落地计划) · [6. 评测方案](#6-评测方案) · [7. 风险登记](#7-风险登记)
- [8. 场景适配](#8-场景适配20-人--70-个-android-项目--自托管-gitlab--lark-国际版--70-agents--编码时使用) — 20 人 / 70 个 Android 项目 / 自托管 GitLab / Lark 国际版 / 70+ agents / **编码时使用**
- [附录 A 来源](#附录-a-来源) · [附录 B 证据索引](#附录-b-证据索引) · [附录 C 假设与核实结果](#附录-c-假设与核实结果原未验证假设v23-已逐条核实)

---

## 0. 速览

### 0.1 三句话结论

> **先回答"还有没有别的记忆库没研究"**：有，而且不止一批。① **本报告自己丢过东西**——底稿研究过 16 个后端，
> v2 的 §3 只剩 7 行，MemOS（底稿里的**第一顺位备选**）、supermemory、LangMem、MemoryOS、Zep、letta-code、
> memobase 的正文信息都在重构时丢了（v2.1 已补回）；② **有两个 ★ 远超我们短名单的高星项目被完全漏掉**——
> `MemPalace/mempalace`（★59k，MIT，原文存储 + 语义检索）与 `vectorize-io/hindsight`（★23.8k，MIT，**正好是
> 上游 issue #677 的原主**）；③ **本轮共核实 150+ 个候选**，其中真正值得动手的是 **ReMe、memU、honcho、
> Neo4j Agent Memory** 四个，其余绝大多数要么只做 MCP（模型自决）、要么不可自托管、要么已停更/是营销仓库。
> **v2.2 已按「停更 >1 个月 / 不可自部署」两条规则清理这份清单：20+ 项出局、全部留痕（§3.2.1）；
> 四个候选与 §3.1 / §3.2 剩下条目的最后提交都在 2026-08-17 之后。**

1. **现状**：Multica 没有记忆层——数据库里没有任何 memory 表（实测）。跨任务的"记忆"是 **CLI 会话续接**，
   只覆盖「同一条 issue × 同一 agent × 同一 runtime」，换机器、换 runtime 即断。
2. **缺口不在存储，在选择**：真正决定成败的第一性问题是「**谁保证每一轮都把记忆放进上下文**」——
   宿主生命周期 hook / MCP 工具（模型自决）/ 平台注入 / 传输层代理，四者的可靠性差一个数量级（§2.1）。
   记忆引擎的品牌是次要的。
3. **建议**：长期正解是 **A + C**——契约层（0 代码）今天可用，平台侧索引注入（150–250 行）1–2 周拿到
   Multica 自己拥有的记忆数据。若要在**今天**就拿到"自动召回 + 自动回写"的可用形态，走 **E（腾讯 TAM，MIT，0 代码）**，
   代价是模型流量多一跳，且记忆的数据与契约都不在你手里。
   **已知具体条件（20 人 / 70 个 Android 项目 / 自托管 GitLab / Lark 国际版 / 70+ agents）的落地版本见 §8。**

### 0.2 该选哪条路

| 你的目标 / 处境 | 选 | 代码量 | 工期 | 主要代价 |
|---|---|---|---|---|
| 只想让 agent 少重复探索同一件事 | **A 契约层**（skill + git 知识库） | 0 | 今天 | 靠人写、靠 agent 主动读；不会自动积累 |
| 先验证"记忆到底有没有用" | **B**（官方 memory MCP + pgvector 自建可选） | 0 | 半天 | 模型可能根本不调用，且"没调用"是静默的 |
| 想直接挑一个**现成后端**练手（本轮新结论） | **ReMe**（§3.1.3；备选 memU / honcho） | 0 | 半天 | **v2.3：ReMe 与 memU 已实机跑过**，各有要先处理的坑（中文分词 / per-task workdir）；honcho 仍未验证（附录 C U5–U7） |
| 今天就要自动召回 + 自动回写，愿意多运维一个服务 | **E 腾讯 TAM**（传输层代理） | 0（只改 `custom_env`） | 1 天 | 多一跳、多一个中间人、两组 LLM 账 |
| 要把记忆变成平台自己的契约与数据（长期正解） | **C 平台侧索引注入** | 150–250 行 | 1–2 周 | 要改 server + daemon，需接受改代码逻辑 |
| C 覆盖不到的 provider 也要自动召回 | **D per-provider hook** | 每种 ~50 行 | 视 provider | 每种都要跟上游 hook API 的变更 |
| 需要"事实随时间失效、可回溯当时为真" | 后端换 **Graphiti**（架构不变，§3.1） | — | — | 多一个图库 |

**反建议**（有据可依的"别做"，详见 §0.3 反例）：

- 不要一上来就上图记忆 / 时序图谱——结构化表示有明确适用边界；
- 不要在缺少审查门时让 agent 直接写**团队可见**的记忆——投毒只需一次普通对话；
- 不要用"把整段历史塞进上下文"当记忆。

### 0.3 五条设计约束（业界最佳实践的提炼）

1. **作用域即加载策略**：谁常驻上下文由**作用域**决定，不由模型判断。宽→窄分层（Claude Code 四级 CLAUDE.md），
   常驻内容设**硬预算**（官方口径：CLAUDE.md <200 行；auto memory 前 200 行 / 25KB）。
2. **写入两步走**：LLM 抽候选 → 检索 top-s 相似记忆后由 LLM 选 **ADD/UPDATE/DELETE/NOOP**；
   旧事实**失效而非删除**，保留血缘（valid_at / invalid_at + provenance）。
3. **读取默认为"不召回"**：recency + importance + relevance 归一化加权，recency 按"上次被检索"指数衰减（0.995）。
4. **维护要能回滚**：记忆版本化（git 化）、每次写入产 before/after 审计 diff、后台整理、TTL / 冷清理，
   **人类可读、可改、可删**。
5. **写入内容是不可信输入**：MINJA 仅靠普通对话即可投毒，且**来源与内容检测都可被洗白**；
   密钥必须在应用侧剥离，不能指望模型"通常拒绝"。

**反例**（同一批来源）：

- **长上下文当记忆**：Mem0 相对 full-context 基线，LLM-as-Judge 相对提升 26%、p95 延迟低 91%、token 省 90%+。
- **图记忆一定更好**：Mem0g 单跳略降、多跳无增益，只在**时间推理**上最好。
- **召回更多 ≠ 更有用**：语义相关但不答问题的干扰项，在长上下文下危害最大（与约束 3 同源）。
- **引用注意口径**：Mem0 **论文**里的自动 DELETE 冲突记忆，与当前**产品文档**的 "ADD-only 累加" 存在落差。

---

## 1. 现状

### 1.1 已有的四个机制

| # | 机制 | 覆盖面 | 失效条件 |
|---|---|---|---|
| 1 | **会话续接** | 同一 (agent, issue) 的下一次任务 `--resume` 上一次会话 | 换 runtime、workdir 被 GC、会话"被污染"、Codex rollout 缺失；续不上会**显式告知**（`PriorSessionResumeUnavailable`） |
| 2 | **每轮重发的上下文** | `workspace.context`（`## Workspace Context`）、项目描述、`project_resource`（`.multica/project/resources.json`）、`agent.instructions`、交办说明、绑定的 skills | 不会累积——只有**人**写进去才会变 |
| 3 | **provider 原生记忆** | 只有 Hermes：per-agent store `<profile>/hermes-state/<agent>/<profile>`，**90 天** GC（注释原话：回收它们等于"可见的失忆"）；Codex 会话 store 14 天 | 其它 provider 没有；Codex 的 auto-memory 被**主动关闭**（要 `MULTICA_CODEX_MEMORY=1` 才开） |
| 4 | **聊天记录** | `chat_session` + `chat_message` 真入库，可回读（`multica chat history`）；web / 飞书 / 企微 / 钉钉都可，Slack 例外 | 不是"记忆"，是转录 |

### 1.2 数据库里没有的东西

实测 `information_schema.tables`：无 `memory` / `workspace_memory` / `knowledge` / `wiki` / `artifact` 任何相关表。
仓库里唯一一处 `memory_scope` 出现在 Hermes 记忆的注释里，指向 **MUL-5969**——即"平台级 memory scope"是**上游的待做项**。

### 1.3 上游的立场与既有尝试

| 记录 | 结果 | 对本次的意义 |
|---|---|---|
| [#838](https://github.com/multica-ai/multica/issues/838) Workspace Memory System（社区提案，含完整 DDL / CLI / 注入设计） | **open**，3 条评论全是社区支持，无维护者承诺 | 提案形状值得采用：**只注入索引（name + description），全文按需取** |
| [#1211](https://github.com/multica-ai/multica/issues/1211) "团队知识该不该像 skills 一样复利" | open，14 条评论，含维护者长回复 | 上游立场的一手来源（下引） |
| [#1417](https://github.com/multica-ai/multica/pull/1417) 给 workspace memory 加 project 作用域 | **closed（未合并）**，理由是 PR 与 main 冲突、非否定想法 | 曾有社区实现过 `workspace_memory` 表形态，可参考 |
| #677 / #2086 / #2100 / #1696 / #4535 / #3147 / #3202 / #6693 / #6658 | 均已 closed：Hindsight 记忆、memory_artifact 基座、知识库 UI、Memory v1、共享记忆召回、Codex 记忆清理与 auto-memory 禁用、Hermes 记忆、隔离 Cursor/Claude 记忆 | "提了又关"很多——**难点在形态与信任模型，不在实现**。⚠️ **#677 的关闭理由本轮已查清**（附录 C U10）：它是一条 **PR**（不是 issue），2026-04-10 开、2026-08-26 由 collaborator 因"与 main 冲突 + 积压清理"关闭，原话 *"This is not a judgment on the idea itself… please rebase onto current `main` and reopen"* —— **Hindsight 这条路没有被否决** |

> ⚠️ **一个闭环**：#677 里的 **Hindsight** 不是别人，正是本轮才被发现的 `vectorize-io/hindsight`（★23,813，MIT，
> 有第三方独立复现的 LongMemEval 成绩，**§3.1.2**）。也就是说：**上游已经试过这个项目并关掉了 issue**——
> "为什么没继续"是本轮唯一可以直接向上游求证的问题，比再选一个后端更有价值。

**维护者原话**（#1211，COLLABORATOR）：

> "the gap you're pointing at is real, but we're being deliberately restrained… every new first-class module
> (agent memory vs. project memory vs. workspace memory vs. a standalone knowledge artifact vs. extending Skills)
> carries a non-trivial cost"

现有假设是 **① Multica CLI 当统一数据层 ② 团队工作流用 Skills 承载**；关于外部知识库：

> "**if we do anything here, we'd own the contract, not the store.** MCP is part of how we're thinking about that."

关于捕获回写：

> "**we want to understand the review/trust model before letting agents write to anything team-visible.**"

社区已验证的替代做法（同上 thread）：**独立 git 仓库当知识库 + agent 用自己的 git 身份提 PR**——
"这一套跑得很好，而且不需要 MCP"。

---

## 2. 选型依据一 宿主能力

Multica 自己是**宿主**（它拉起各 CLI 跑任务），所以"自动召回"要按 provider 分包实现，或绕开 provider 在更外层做。

### 2.1 三种集成模式（决定可靠性与成本）

| 模式 | 可靠性 | token 成本 | 可审计性 | 用在哪 |
|---|---|---|---|---|
| **生命周期 hook 自动注入** | 最高：不依赖模型想起用工具 | 中：按事件与预算精确控制 | 高：注入内容 / 时机 / 失败态可日志化 | 要"开箱即用的长期记忆" |
| **MCP 工具（模型自决）** | 最低：可能不调用 / 调错 / 不该调时调用 | 低（不调不花）但不可预测 | 中：调用有记录，**"没调用"是静默的** | 模型自主探索，或宿主不给 hook |
| **平台侧注入 system prompt** | 高：平台保证每轮存在 | 最高：每轮常驻 | 中：内容可见但注入逻辑不在使用者手里 | 少量、稳定、全局的规则 |

（第四种是**传输层代理**：它在协议层再包一层，等价于"平台注入"，但对 client 透明——见 §4.4。）

OpenViking 的插件标准把这条讲透了：**"Hooks 自动化生命周期行为；MCP 暴露模型选择调用的工具。两者共用同一份配置和身份，但职责分离"**，
并要求只做 MCP 的集成**声明自己的能力缺口**（无自动注入、无完整会话记录）。

**两条硬要求**：

- **压缩后必须重注入**：只在会话开始注入一次的设计，会在第一次 auto-compact 后失效
  （Claude Code 官方原话 "Context that hooks added earlier → Summarized with the rest of the conversation"）。
  Claude Code 用 `SessionStart` + `matcher: "compact"`；Codex 用 `SessionStart(source=compact)`；OpenCode 用 `experimental.session.compacting`。
- **注入是概率，不是保证**：CLAUDE.md 的定位是 "context, not enforced configuration"；
  凡"绝不能发生"的约束（例如不要把虚拟 URI 交给 shell）必须下沉到 `PreToolUse` 层拦截，或**明确文档化做不到**。

### 2.2 各 provider 的注入点

| Provider | 每轮注入 | 压缩后重注入 | 自动捕获 | 宿主需要做什么 |
|---|---|---|---|---|
| **claude** | ✅ `UserPromptSubmit` → `additionalContext` | ✅ `SessionStart(matcher:"compact")` | ✅ `Stop`/`SessionEnd` + `transcript_path` | 传 `--settings`（含 hooks）或 `--plugin-dir`；**别加 `--bare`** |
| **codex** | ✅ `UserPromptSubmit.additionalContext` | ✅ `SessionStart(source=compact)` | ✅ `SessionEnd` | `~/.codex/hooks.json` 或 `--profile`/`-c`；**非托管 hook 首次要过 trust** |
| **opencode** | ✅ `experimental.chat.system.transform` | ✅ `experimental.session.compacting` | ✅ `event` / `tool.execute.after` | 装插件（API 带 `experimental.` 前缀，不保证稳定） |
| **openclaw** | ✅ `before_prompt_build` → `prependContext` | ✅ context engine 四阶段 | ✅ | context engine 插件槽位 |
| **hermes** | ✅ `pre_llm_call` → `{"context":…}` | ✅ | ✅ `post_llm_call` | **shell hooks 即可，不必写 Python**（需 allow-list + 首次同意） |
| **dsh** | ✅ `agent/pre-step` + `systemPrompt.context()` | ✅ | ✅ `session/event` | 装插件；`dsh-openviking` 就是现成范例 |
| **cursor** | ❌ 没有"每轮 prompt 前注入"通道（`beforeSubmitPrompt` 只能拦截） | ❌ `preCompact` 只读 | ✅ `afterAgentResponse` | 只能"会话级一次性注入 + 工具后追加 + `stop.followup_message`（默认限 5 次）" |
| **reasonix** | ⚠️ `UserPromptSubmit` **有**事件，但 **stdout 被丢弃**（只当闸门，不能注入） | ✅ `SessionStart`（可注入，10k/20k 上限） | ✅ | **v2.3 已核实**（附录 C U2）：11 个 hook 事件；每轮注入只能走**改写 `{workDir}/AGENTS.md`**——Multica 已经写这个文件（外加 `.reasonix/skills/`、`reasonix.toml`），且**没有**占用 `.reasonix/settings.json` |

### 2.3 对 Multica 的四条硬约束

1. **不采纳插件 hooks**：Multica 只采纳用户级 Claude Code 插件的 `skills` 与 `mcpServers`（manifest 只有这两个字段），
   于是"装一个带 hooks 的记忆插件就自动召回"在 daemon 任务里**是否生效尚未验证**——daemon 还会传自己的 `--settings`（用于技能开关）。
   ✅ **v2.3 已实测（附录 C U1）：生效**——claude CLI 自己会加载用户级插件的 hooks，daemon 的 `--settings` 只覆盖 skillOverrides/permissions，
   探针确认 hook 真的执行（`Registered 75 hooks from 31 plugins`）；但粒度是 **per-host / per-user**，按 agent 隔离要靠 `CLAUDE_CONFIG_DIR`。
2. **参数有白名单**：daemon 对每个 provider 的启动参数有硬编码拦截表，
   想给某个 provider 注入 hook 配置，要先确认参数没被拦掉。
3. **workdir 每任务重建**（`local_directory` 资源除外），而"运行时本地 MCP 继承"只对 6 个 provider 生效——
   写 `~/.claude.json` 这类"全局注入"覆盖不全，且会波及用户机器上的其它项目。
4. **交互式工具被禁用**：Multica 对 claude（及 codebuddy）传 `--disallowedTools AskUserQuestion`。
   任何依赖"弹表单让用户选择"的第三方组件，在 daemon 任务里都走不通——这一条直接决定了 §4.4 的接法。

---

## 3. 选型依据二 记忆后端

### 3.1 候选对比

先看两条路的"骨架"（自建表 / 官方 MCP），再看现成产品。

| 方案 | 形态 / 存储 | 依赖 | 隔离 | 接口 | license | 部署成本 | 判断 |
|---|---|---|---|---|---|---|---|
| 官方 `@modelcontextprotocol/server-memory` | 单进程 MCP，本地 JSONL | 无（纯字符串匹配） | ❌ 无分桶 | MCP stdio | MIT/Apache 过渡 | 极低（`npx` 一行，实测常驻 ≈253MB） | **只用于验证"记忆有没有用"** |
| **mem0 OSS 自托管** ★65k | FastAPI + dashboard，向量可落 pgvector | LLM 抽取 + embedding（可 Ollama/DeepSeek） | ✅ user/agent/run + metadata | **REST** + SDK（MCP 是 Platform 专属） | Apache-2.0 | 中（+2 容器） | 唯一"REST + per-user key + 自托管 + 可换 pgvector"的成熟服务 |
| **自建 Postgres + pgvector** | 复用现有库 | embedding 自选（本地 Ollama） | ✅ 任意分桶 + 复用现成鉴权 | 自己定（SQL / Go / MCP 包装） | 自有 | **最低（+1 migration）** | 零新进程、租户模型天然正确；抽取 / 去重 / 失效要自己写 |
| **Graphiti** ★31k | 图库（FalkorDB/Neo4j）+ MCP server | LLM + embedding（可 Ollama） | ✅ `group_id` | **MCP HTTP `/mcp/`** | Apache-2.0 | 中高（+图库） | 唯一原生支持"事实随时间失效、可回溯当时为真" |
| OpenViking | 服务 + Studio + CLI | **需 embedding + VLM** | ✅ 账号/用户 + 可选 ACL | REST / CLI / SDK + **MCP** + 各家插件 | **AGPL-3.0** | 中高 | 能力最全（含自动召回范例），代价是 AGPL + VLM + 体量 |
| **TAM（腾讯）** ★26.9k | Memory Core + Panel/Knowledge + **LLM 代理**，三件套均为公开 Docker Hub 镜像 | 两组 LLM（抽取 + 转发）+ Docker | ✅ team/agent/task/session + private/team/restricted(ACL)/agent | **不靠接口**：改 base URL 即接入；另有 HTTP v3 API / Panel | **MIT** | 中（+3 容器） | **唯一"零代码 + 覆盖全部 provider"的现成解**；代价是成为模型流量中间人（§3.4、§4.4） |
| basic-memory | Markdown 文件 + 索引 | 本地 embedding（fastembed） | 弱 | MCP | **AGPL-3.0** | 低 | "人要能直接读改记忆"时合适 |

**活跃度**（2026-09-17 逐个读最后提交）：官方 memory MCP `2026-09-03` · mem0 OSS `2026-09-16` · Graphiti `2026-09-11` ·
OpenViking `2026-09-17` · TAM `活跃`（v2.0.1-beta.1）· basic-memory `2026-09-16` —— **全部在窗口内**；
「自建 Postgres + pgvector」不是库，没有日期可言。

#### 3.1.1 底稿丢失的候选（v2.1 补回，v2.2 过筛）

v2 的重构把底稿 B 里研究过的 7 个后端压缩成了"排除"一句话，其中 **MemOS 是底稿标注的第一顺位备选**。
逐条补回（★ 为 2026-09-17 读数）：

| 方案 | ★ | license | 形态 / 存储 | 为什么当初被看上 / 为什么降级 |
|---|---|---|---|---|
| **MemOS**（`MemTensor/MemOS`） | 11.5k | Apache-2.0 | 服务端（API + **Neo4j + Qdrant**）+ TS 插件 + 本地 SQLite 插件 | 底稿第一顺位：**hook 框架最"正统"**——服务端自带 `plugins/hooks.py`，DSH / OpenClaw / Hermes 插件已就绪（README：每次 agent run 的第一次模型调用前召回、结束后写回）。降级原因：向量后端只有 `qdrant.py` / `milvus.py`，**Postgres 只能当图后端** → 用不上我们现有的 PG，且要多运维两个库 |
| **supermemory** | 29.8k | **MIT** | 云优先 + 本地单二进制（`supermemory local`，数据全落 `./.supermemory`）+ 全套 coding harness 插件 | 底稿评语"本地部署最轻"：自带本地 embedding（`Xenova/bge-base-en-v1.5`），可全离线。降级原因：**自托管版本没有多租户**，团队/合规隔离属 Enterprise |
| **LangMem + LangGraph store** | 1.7k | MIT | 库（非服务） | "Core memory API that works with any storage system" + agent 可调用的记忆管理工具。要自己写服务与注入 → 只剩"设计参考"价值 |
| **letta-code**（`letta-ai/letta-code`） | 3.4k | Apache-2.0 | npm CLI harness + App Server + TS SDK；MemFS 用 git 追踪全部上下文 | `letta-ai/letta`（★24.8k）**仓内已无代码**（只剩 README/PRIVACY/TERMS，V1 进了 `archive` 分支）。"必须换运行时"仍成立，但形态从"起 server"变成"装 CLI" |

**v2.2 过筛**：底稿这 7 个里 **4 个通过两条规则**（上表）；**MemoryOS（最后提交 2026-07-07）、Zep CE（CE 官方下线）、
memobase（2026-01-11）已剔除**——日期与理由见 §3.2.1，正文不再把它们当候选。留存四项的活跃度：
MemOS `09-16` · supermemory `09-17` · LangMem `09-09`（⚠️ 其 **PyPI 包停在 0.0.30 / 2025-10-27**）· letta-code `09-17`。

#### 3.1.2 两个高星盲区（★ 远超短名单，但此前从未被研究）

| 项目 | ★ / fork | license | 它是什么 | 与我们的关系 |
|---|---|---|---|---|
| **`MemPalace/mempalace`** | **59,114** / 7,558 | **MIT**（LICENSE 已读） | "Local-first AI memory. **Verbatim storage, pluggable backend, 96.6% R@5 raw on LongMemEval — zero API calls.**" **不摘要、不抽取**：原文存入 wings（人/项目）/ rooms（话题）/ drawers（原文），检索用语义。默认后端 ChromaDB（`mempalace/backends/base.py` 可插拔）。有 CLI（`uv tool install mempalace`）+ MCP + 三个 skill | ★ 是本报告任何候选的 2 倍，MIT、零 API 调用。**代价**：靠 hooks 自动保存（README 自承 "Claude Code sessions expire in 30 days without auto-save hooks wired"）；社区移植生态巨大且混杂（Go/Rust/.NET 插件、`mempalace-cn`，以及 README 专门警告的 **impostor site**）。本地默认后端不是 PG → 要接我们的话属"架构 D"。**v2.3 实机核实**（附录 C U9）：`uv tool install mempalace` → 3.10.0，`init --yes --no-llm` + `mine --direct` 写入 3 条 → 中英各查一次都是 **rank-1 命中**；**零 key**（ChromaDB + 本地 ONNX MiniLM），但"zero API calls"≠零网络（首次要下 ONNX 模型）；⚠️ **中文只走向量通道**（bm25 0.0）；官方点名的 impostor 域名是 **`mempalace.tech`**（`.net` 与其他 `.com` 变体同样非法） |
| **`vectorize-io/hindsight`** | **23,813** / 1,768 | **MIT** | 宣称 LongMemEval SOTA，且**基准由 Virginia Tech Sanghani Center 与 Washington Post 独立复现**（其余厂商均为自报）——这是本批唯一有第三方复现的。有 Docker server、Python embedded 模式、MCP、2 行 LLM wrapper、coding agent 集成、Hindsight Cloud。论文 arXiv:2512.12818 | ⚠️ **闭环**：Multica 上游 **#677「Hindsight 记忆」已 closed**（§1.3 表里就有）——**上游试过的正是这个项目**，而本报告此前从未把它当一个后端来研究。**v2.3 实机核实**（附录 C U9）：`ghcr.io/vectorize-io/hindsight` 容器**起得来**（v0.10.0，API `:8888` + `/mcp` + Control Plane `:9999`），但 **retain/recall 必须有可达的 LLM 端点**——`POST /v1/…/memories` 实测 500 `Fact extraction failed`；且 **LongMemEval/LoComo 的仓内副本已删**（`hindsight-dev/benchmarks/README.md` 明说迁去外部 AMB 仓库），所以"第三方独立复现"本机无法验证。**#677 的关闭原因已查清：不是否决**（§1.3、U10） |

另外两个高星但**建议排除**的：`mindverse/Second-Me`（★15,686）定位是"Train your AI self"个人分身，**不是记忆库**；
`EverMind-AI/EverOS`（★13,016，Apache-2.0）虽然形态漂亮，但官方文档**明说无任何内置鉴权**
（*"EverOS ships no built-in authentication"*），存储用 Markdown+SQLite+LanceDB，两头都对不上我们。

#### 3.1.3 本轮真正值得动手的四个（结论）

| 候选 | 一句话理由 | 上手前必须验证 |
|---|---|---|
| **1. ReMe**（`agentscope-ai/ReMe`，★3.5k，Apache-2.0） | **形态与我们对齐度最高**：Markdown 为真源、索引可重建、原生 MCP + HTTP + CLI、默认只绑 `127.0.0.1`、向量可关（可纯离线 + 中文 BM25）。`reme start` 一条命令起。国内团队（阿里 AgentScope），是 MemoryScope 的继任者 | 中文 BM25 分词质量；多 agent 共享同一 workspace 的并发写；`auto_memory`/`auto_dream` 演化流程的 token 成本。**v2.3 已实测**：起得来（0.4.1.12，33 个 HTTP 端点，MCP streamable HTTP 32 工具），并发写 25+25 条无丢写；⚠️ 但默认分词器是 **regex 逐字**（中文等于单字 OR，`签名` 会召回只含"名"字的条目）→ 上生产前应换 jieba 并做 A/B；⚠️ `reme-ai` 装不全（要 `[as]` 额外依赖）、**HTTP 服务无 LLM key 起不来**（"无凭证可跑"只对 SDK 成立）；⚠️ 写入到可检索有 4–20s 的 watcher 延迟。LongMemEval 89.4% 跑在**自家修订过 ground truth** 的 clean-S 上（附录 C U6/U11） |
| **2. memU**（`NevaMind-AI/memU`，★14.4k，Apache-2.0） | 唯一同时满足三件事：**直接复用现有 PG + pgvector**、**不需要额外 LLM**（蒸馏由宿主 agent 自己做）、**改写宿主指令文件逼模型先 `retrieve`**；还会挖宿主自己的 session 日志 | **v2.3 修订**：包名是 **`memu-cli`**（PyPI 正式版 `0.10.0`/2026-08-05，预发布 `0.11.0b3`，分类器 `4 - Beta`）——报告此前写的"正式版停在 `v2.0.0-beta.0`/`v1.5.1`"是 GitHub tag 线，与发布线对不上；⚠️ **embedding 是硬依赖**（README："Embedding key required"）→ "不需要额外 LLM"只对蒸馏成立；无 MCP 属实；`AGENTS.md` 注入是幂等可回滚的 managed block（与 Multica 的标记块共存），**但 Multica 的 workdir 是 per-task 的**，打 `./AGENTS.md` 的适配器写的块随任务目录消失 → **结构性不兼容，不是被覆盖**（附录 C U5/U7） |
| **3. honcho**（`plastic-labs/honcho`，★7.2k，**AGPL-3.0**） | harness 覆盖第一：**claude / codex / opencode / cursor / dsh 逐个有 hook 级插件**（6 个插件仓库已逐个核实存在且仍在更新，`dsh-honcho` 建于 2026-08-31）；PG+pgvector+Redis，真多租户 JWT/workspace 鉴权，后台 deriver | **AGPL-3.0**（本批唯一许可风险命中）；多一个 deriver + Redis；`dsh-honcho` 成熟度最低 |
| **4. Neo4j Agent Memory**（`neo4j-labs/agent-memory`） | 如果报告要保留"图记忆 + 关系/时间推理"这条线，它比 Graphiti 更贴"agent 记忆"：POLE+O 模型、实体去重、**16 工具 MCP**、官方托管（NAMS）与自托管双形态 | 官方标注 **Experimental + Community Supported**（生产风险自负）；自运维 Neo4j |

> **v2.3：四个里 ReMe 与 memU 已实机跑过**（2026-09-17，附录 C U5–U7）；**honcho 与 Neo4j Agent Memory 仍未实机**。

**四个候选的最后提交（2026-09-17 读数）**：ReMe `2026-09-16` · memU `2026-09-10` · honcho `2026-09-16` ·
Neo4j Agent Memory `2026-09-16` —— 均在一周内，且四个都能自托管。

**综合本轮全部信息后的唯一推荐（在 §8 那套条件下）：ReMe。** 判据与它自己的 README 原文（已直读核实）：

| 判据 | ReMe 的事实 |
|---|---|
| 注入不能靠 hook、也不能靠模型自觉（§8.6 的三条死路） | 它同时是**服务（HTTP）+ CLI + MCP + Python** 四种接口 —— 架构 C 的每轮注入由 daemon 直接 HTTP 取索引，**不依赖 hooks、不依赖模型配合** |
| 与被选定的"真相层"同构（§8.1） | **"Memory as File, File as Memory"**：Markdown + frontmatter + wikilinks 是真源，索引（BM25/向量）可重建 → 记忆条目可 git diff、可走 GitLab MR 审查门 |
| 内网自托管、少依赖 | **默认只绑 `127.0.0.1:2333`**；**BM25 默认、embedding 默认关闭**（*"the default setup does not start an embedding model or require an embedding API key"*）→ 可完全离线；需要时再开向量并用 RRF 融合 |
| 70 个 agent 的规模（§8.4） | *"One memory workspace works across agents"* —— 一个 workspace 被多 agent 共享（原生集成 / SKILL.md / CLI / HTTP / MCP）→ **一个进程服务全部 70 个 agent**，不是 70 份安装 |
| 许可与可读性 | Apache-2.0；自带 Studio（`127.0.0.1:2333`）能人读人改 |

**配套接法**：GitLab 建 `multica-memory`（真源 + 审查门）→ ReMe 把它当 workspace 索引（一项目一子目录，天然就是 project 硬过滤）
→ 架构 C 的**每轮块**向 ReMe 查"本项目 + 本轮相关"的索引（**v2.3 实测：30 条记忆的索引块 ≈896 tokens，只注入 name+desc ≈570；原假设的 ≤300 偏低，见附录 C U16**）→ 全文走 ReMe CLI/skill 按需取。

**换选条件**（不是更好，是不同取舍）：要"今天 0 代码就有自动召回"→ 用 **TAM 代理**（不是库）；
要"接口面最全（REST + MCP + Control Plane）"→ **Hindsight**（上游 #677 已查清：不是否决，rebased 后有戏，U10）；
要"单二进制、零 Python/零模型/零 Docker"→ **engram**；要"直接复用现有 PG+pgvector"→ **memU**（2.0 仍 beta）。
**MemPalace（★59k）与 honcho（★7.2k）在这套条件下不推荐**：前者的自动保存依赖 hooks、后者最大的卖点（6 个 harness 插件）
恰好在 Multica 里用不上（§8.2-⑨⑩），且 honcho 是 AGPL-3.0。

#### 3.1.4 全网检索的新候选与 OKF 实机结论（v2.4）

**检索通道（本轮的"全网"边界，先读）**：通用搜索引擎仍不可信——**Bing** 能拿 200 但结果基本是垃圾（Westpac / VirtualDJ /
福特澳洲…，中文查询直接"没有与此相关的结果"）；**DuckDuckGo lite 连不上**，Mojeek 出 captcha，Brave / Qwant / Startpage
与 5 个公共 SearXNG 实例**全部超时**，Reddit JSON 超时，百度要安全验证。**Firecrawl 这轮不可用**（vault 无 key，见页首 v2.4 块）。
真正好用并据此捞出下列候选的是四个**能直连的聚合通道**：**Hacker News Algolia API**
（`hn.algolia.com/api/v1/search_by_date?query=…&tags=story`）、**Lobsters**（`lobste.rs/search.json?q=…`）、
**Marginalia**（`search.marginalia.nu/search?query=…`）、以及 GitHub API（`gh` 已认证）。若想让下一轮拿到**真·通用搜索**，
给 vault 加 `FIRECRAWL_API_KEY` 即可（SKILL 已就位：`agent_state/skills/firecrawl/SKILL.md`）。

**新候选（★ / 最后推送均为 2026-09-30 读数，走 GitHub API）**：

| 仓库 | ★ | license | 最后推送 | 语言 | 一句话 | 结论 |
|---|---|---|---|---|---|---|
| **`okf-memory/okf-agent-memory`** | 739 | **MIT** | 2026-09-27 | Go | **Git 原生项目记忆**：`knowledge/` OKF v0.2 bundle + 内存 BM25(<300µs) + 内嵌 MCP(stdio) + **零依赖**（无 DB / 无外部进程） | 🟢 **本轮最大发现**：架构 A 契约层的现成实现，**按项目物理隔离**（见下） |
| `mindmuxai/brain.md` | 554 | Apache-2.0 | 2026-09-11 | JS | 仓库内 `brain/` Markdown 记忆 + 零依赖 CLI（`brain init`，v0.3.0） | 🟡 同类的第二选择：项目内 Markdown，但没有 `code_refs` / 校验器 |
| `ctxrs/ctx` | 1,145 | Apache-2.0 | 2026-09-30 | Rust | **检索本机已有的 agent 会话历史**（跨 claude/codex/cursor 的 transcript 搜索） | 🟡 **不是记忆库**——但它解决"70 个 agent 的历史在哪"这个相邻问题，值得单列 |
| `Avinash-jetwani/jevmem` | 99 | MIT | 2026-09-30 | TS | Claude Code 的自动项目记忆（已进 Anthropic 插件目录） | 🟡 单 harness（claude），覆盖不了 70 个 agent |
| `hyperb1iss/sibyl` | 63 | Apache-2.0 | 2026-09-30 | Python | 知识图谱 + 持久记忆运行时 | 🟡 与 Graphiti / Neo4j Agent Memory 重叠，无增量 |
| `salishforge/memforge` | 14 | MIT | 2026-09-28 | TS | PG + MCP + "睡眠周期"巩固 | 🟡 太新、★ 太低，仅留痕 |
| `samvallad33/vestige` | 639 | **AGPL-3.0** | 2026-09-30 | Rust | "queryless memory layer for multi agent coding fleets"——定位与我们几乎重合 | 🔴 **许可出局**（AGPL，同 honcho / basic-memory 的理由） |
| `Fr4nZ82/mwe-mcp` | 6 | **AGPL-3.0** | 2026-09-19 | Rust | Markdown wiki + per-fragment ACL，MCP | 🔴 **许可出局**（AGPL） |

另有一批**已停更或只有资讯价值、不入表**：Recalld、IngotDB、SelMem、nxm-memory、two-tier-memory（07-07）、PondDB（03-19）、
Fraise、Slowave、Neotoma，以及 HuggingFace blog（09-04）提到的 **Funes**。

**OKF Agent Memory 本机实测（v0.4.4，`okf-darwin-arm64`，校验和通过）**——这是本轮唯一"到手就能跑"的新候选：

- **形态**：单二进制 CLI。命令面 `validate / search / show / create / update / relate / init / bootstrap / agents /
  mcp(stdio) / hub(push pull sync serve --port 8080，零知识 CAS 同步) / version`。`bootstrap` 会在仓库里生成
  `knowledge/`、`.agents/skills/okf-memory/`、`AGENTS.md`、`Makefile`。
- **✅ 实测通过**：`create` 写概念（自动维护每目录的 `index.md` 与追加式 `log.md`）；`search "构建缓存"` 命中；
  `search --for-path app/build.gradle.kts` 经 **`code_refs`** 命中（**这正是"改这个文件前该看哪条记忆"的原生机制**）；
  `search --filter "type=Decision"` 命中；`validate` 报 orphan / broken-link / stale 并给 "Conformant"；
  **两个项目物理隔离**——projA 的记忆在 projB 里搜不到（projB 只搜得到自己的 `flavor-channel`）。
- **⚠️ 中文检索硬伤（实测，源码 `pkg/okf/search.go`）**：`tokenize()` 只在遇到非 ASCII 字母/数字时才切分 →
  **整段中文几乎总是一个 token**，且 TF 用 `HasPrefix`、DF 用 `strings.Contains`（作者自己的 TODO 承认这个不对称）。
  实测：`构建缓存` ✓、`构建` ✓、**`缓存` ✗**（后缀 / 中缀查不到）。**缓解**：把关键词写进 `tags`（tags 是独立 token）。
- **⚠️ 成熟度风险**：仓库 **2026-09-05 创建**（截至 09-30 仅 3.5 周），v0.4.4，主力贡献者只有 `sknr`（144 commits，其余为 bot）。
  739 ★ 是**建仓三周半**的读数 → 典型的发布曲线高点，`§3.2.1` 的规则①（停更一个月）**下次复核必须重跑**。
- **⚠️ 没有 HTTP 服务**（只有 CLI + stdio MCP）→ 要在 daemon 侧做"每轮注入"，得多起一个子进程（或走 MCP），
  这与 ReMe「一个进程服务 70 个 agent」（§3.1.3）的成本模型不同。

**结论与取舍（本轮回答"multica 适合哪个长期记忆开源库"）**：

1. **平台后端仍推 ReMe**，理由不变（§3.1.3）：服务化 + HTTP + 多 agent 共享 + 默认离线 + Apache-2.0 + 成熟度，
   且 v2.3 已实测其 `path_prefix` 过滤与并发写。
2. **OKF Agent Memory 是"按项目隔离"这个诉求下比 ReMe 更贴形的新挑战者**，但它**不是后端，是架构 A（契约层 / 项目真源）的实现**：
   Git 原生 → 记忆就是 `knowledge/*.md`，随仓库走、可 diff、可走 GitLab MR 审查门；`code_refs` / `verified.by` / `log.md`
   与你们的"审查门"天然对齐；**它是唯一"把项目隔离做在文件系统层"（而非 query filter 层）的候选**——隔离是物理的，不是元数据约定。
   建议的落地方式：**单二进制丢进某个试点项目、当架构 A 试**，不进平台；不满意就 `rm -rf .agents knowledge`，可一键移除、风险可控。
3. **翻转条件（满足其一就能顶替 ReMe 当平台后端）**：① 在你们的**真实中文语料**上召回达标（把关键词同时写进 `tags` 后仍达标）；② 上游补上 HTTP 服务，
   使 daemon 能像打 ReMe 那样一次 HTTP 取索引；③ 上游修掉 `HasPrefix`/`Contains` 的不对称分词。
4. **`vestige` 的定位最像我们要的东西，但 AGPL-3.0 直接出局**；`brain.md` 可作 OKF 的备胎；`ctx` 只解决会话历史检索，不解决记忆。

### 3.2 完整筛查清单（本轮逐条核实过的全部候选，v2.2 已过筛）

**为什么要这张表**：v2 丢过候选（§3.1.1 就是证据）。下表把"看过但没进正文"的东西**全部留痕**，
标注结论，供后人"看过、为何不选"。★ / fork 均为 **2026-09-17** 读数，license 取自**仓库根 LICENSE 原文**（不采信 README badge）。

**v2.2 的状态**：下表是**两条规则清理之后**的样子——停更 >1 个月、或不可自部署的已移出（见 §3.2.1）；每张小表后面跟一行
**活跃度**（最后提交日期，逐个走 GitHub commits API 读出）。⚠️ 表内仍有几项标 🔴 却**留在表里**（semantica、utopia、EverOS、
MindMemOS、VoiceMem、AgentRecall-X…），它们**不是**被这两条规则否掉的，而是"仓库活着但不能用"（营销仓库 / 功能没实现 /
无鉴权 / 无 LICENSE / 场景不适用）——与 §3.2.1 的"已出局"分开看。

**结论标记**：🟢 进入候选 · 🟡 备选 / 仅作参考 · 🔴 排除 · ⚪ 只用于评测或方法论

#### A. 平台型 / 服务端记忆系统（14 个 → 过筛 9 个）

| 项目 | ★ | license | 形态 / 存储 | 召回模式 | 结论 |
|---|---|---|---|---|---|
| `NevaMind-AI/memU` | 14,410 | Apache-2.0 | CLI sidecar / SQLite 或 **PG+pgvector** | 指令文件注入 → 模型执行 retrieve | 🟢 见 §3.1.3 |
| `plastic-labs/honcho` | 7,212 | **AGPL-3.0** | API+deriver+PG+Redis | **宿主 hook（6 种 harness）** | 🟢 见 §3.1.3 |
| `MemTensor/MemOS` | 11,500 | Apache-2.0 | 服务端 + TS 插件 / Neo4j+Qdrant | 宿主 hook + REST | 🟢 见 §3.1.1 |
| `supermemoryai/supermemory` | 29,830 | MIT | 云 / 本地单二进制 `.supermemory` | 插件 hook + MCP 工具 | 🟡 自托管无多租户 |
| `MemMachine/MemMachine` | 3,221 | Apache-2.0 | app + PG(pgvector) + Neo4j | REST/SDK + MCP | 🟡 要 Neo4j |
| `doobidoo/mcp-memory-service` | 1,945 | Apache-2.0 | 1 进程 / SQLite-vec（无 PG） | **宿主 hook（SessionStart 注入）最完整** | 🟡 **hook 注入的最佳参考实现**，但存储不合 |
| `semantica-agi/semantica` | 13,065 | MIT | Polyglot（RDF/LPG/向量） | MCP/REST 自决 | 🔴 根目录 `GROWTH.md` = 星标漏斗 playbook 与渠道清单 |
| `deeplethe/utopia` | 7,959 | Apache-2.0 | 1 二进制 + PG+pgvector | MCP 工具 | 🔴 **记忆功能尚未实现**（路线图未勾选），只有 rc、无 tests |
| `EverMind-AI/EverOS` | 13,016 | Apache-2.0 | Markdown+SQLite+LanceDB | REST + 另仓插件 | 🔴 **官方明说无任何内置鉴权**；用不上 PG |

**活跃度**（最后提交，2026-09-17 读数）：`memU` 09-10 · `honcho` 09-16 · `MemOS` 09-16 · `supermemory` 09-17 ·
`MemMachine` 09-17 · `mcp-memory-service` 09-16 · `semantica` 09-17 · `utopia` 09-16 · `EverOS` 09-08 —— 全部在窗口内。

#### B. 编码 agent 专用记忆工具（18 个 → 过筛 13 个）

**先记住两个坑**：① GitHub 搜索结果的 `updated_at` **不是推送时间**（随 star 刷新）——`byterover` 就被它误导成"昨天还在更新"，真实 `pushed_at` 是 6 月 25 日；下表时间一律走 commits API / `pushed_at`。
② README 说"支持 N 个 client"基本等价于"MCP 是通用的"；**真正逐宿主落了插件的只有 8 个**：deja-vu、memorix、engram、memsearch、openwolf、iai、token-savior、byterover（仅 skill，无 hook）——**后两个已按 v2.2 剔除（§3.2.1）；token-savior 的 `hooks/` 仍是本批最好的注入样板**。

| 项目 | ★ | license | 形态 / 存储 | 召回模式 | 结论 |
|---|---|---|---|---|---|
| `vshulcz/deja-vu` | 822 | MIT | Go 单二进制 + 逐宿主真插件（`claude-plugin/`、`codex-plugin/`、`kimi.plugin.json`、`dsh-deja`） | **反建磁盘会话**（不需 agent 配合写）；`install --auto` 自动 / `--all` 克制 | 🟢 唯一"不改造 harness 也有价值"+ 覆盖 DSH/Hermes/omp/Kimi/opencode/Crush/Grok |
| `AVIDS2/memorix` | 779 | Apache-2.0 | TS monorepo + CLI/HTTP MCP/dashboard / SQLite FTS5（+可选 LanceDB） | 默认模型自决，可选 `memorix hooks` 捕获 | 🟢 **唯一把"多 agent 共享一个 store"做成正式特性**（identity/team/lock/orchestrate/Postgres relay）；需评估"只取记忆子集" |
| `Gentleman-Programming/engram` | 6,665 | MIT | **单 Go 二进制 + SQLite FTS5，无 embedding/无 Python/Node/Docker** | hook 只做"提醒"，记忆靠模型主动调 | 🟢 工程与依赖最优、有 `docs/TEAM-USAGE.md` + Git Sync；**前提是先实测 v2.0.0-rc 系列**（正式版停在 7 月的 v1.20.0，Homebrew 还在旧线） |
| `zilliztech/memsearch` | 2,614 | MIT | Markdown 真源 + Milvus 影子索引；5 个宿主插件（含 **dsh 原生**） | 浅注入（2 个日记 ≤1800B）+ 深拉取 skill | 🟡 basic-memory 的直系竞品且是 MIT；代价 558MB 本地模型 |
| `riponcm/projectmem` | 824 | MIT | 纯文件（`events.jsonl` + 派生 Markdown），无 DB 无 embedding | MCP 工具 + git hooks + `pjm wrap` 预注入 | 🟡 判断层设计好（precheck 警告"这做法失败过"、supersedes 版本链），但无身份/共享库语义 |
| `cortexkit/magic-context` | 2,132 | MIT | Bun 宿主插件+CLI+桌面 / 单一 SQLite | ✅ **每轮自动注入**（最彻底） | 🟡 只覆盖 OpenCode/Pi/OMP 三个宿主 |
| `cytostack/openwolf` | 2,336 | **AGPL-3.0** | npm + hooks/插件/CLI，`.wolf/` 纯文件 | 读前引导 + 每 25 个 tool batch 重述规则 | 🟡 hook 工程值得抄；作者自认"省 token"未被证明 |
| `ReflexioAI/claude-smart` | 779 | Apache-2.0 | 宿主插件 + Python 后端（需常驻） | ✅ 每次 user prompt 检索注入 | 🟡 仅 3 家宿主，已缓速（release 停 07-15） |
| `Dataojitori/nocturne_memory` | 1,356 | MIT | FastAPI+React+MCP / SQLite 或 PG 图模型 | ❌ 模型自决，无 hook | 🟡 定位是"AI 人格/第二大脑"，对编码 agent 错配 |
| `CodeAbra/iai-personal-memory-engine` | 886 | MIT | — | — | 🟡 观察（见 D 报告） |
| `tigerless-labs/agent-memory` | 919 | MIT | — | — | ⚪ **观察不作依赖**：建仓 16 天、星速 ≈57★/天、0 release、0 PyPI、版本 0.1.0、LICENSE 09-08 才补 |
| `okf-memory/okf-agent-memory` | 687 | MIT | — | 无 hook 无 embedding | ⚪ 同上：建仓 12 天、12 天里从 0 到 0.4.0-rc **全是 rc** |
| `Goldentrii/AgentRecall-X` | 371 | MIT | TS monorepo + MCP + skill | 🟡 靠 `/arstart`·`/arsave` 仪式（作者实测 pull 工具零自然调用） | 🔴 ★/fork = 6.4（本批最低），定位是"行为纠正账本" |

**活跃度**（最后提交，2026-09-17 读数）：`deja-vu` 09-17 · `memorix` 09-14 · `engram` 09-17 · `memsearch` 09-12 ·
`projectmem` 09-14 · `magic-context` 09-16 · `openwolf` 09-15 · `claude-smart` 08-28 · `nocturne_memory` 08-27 ·
`iai` 09-16 · `tigerless-labs/agent-memory` 09-09 · `okf-agent-memory` 09-17 · `AgentRecall-X` 09-13 —— 全部在窗口内。

#### C. 研究型 / 图 / 时序 / 榜单（D3：10 个 → 过筛 7 个）

| 项目 | ★ | license | 一句话 | 结论 |
|---|---|---|---|---|
| `kayba-ai/agentic-context-engine`（ACE） | 2,573 | Apache-2.0 | **程序性记忆**：Reflector→SkillManager→JSON Skillbook | 🟡 与 `agent_task_queue.error` 天然对口（"从失败中学习"），纯库可只借设计 |
| `topoteretes/cognee` | 30,757 | Apache-2.0（**LICENSE 纯 Apache，无附加条款**） | SDK + REST:8000 + UI:3000 + MCP:8001 | 🔴 结论**要改写**：收费点精确在「**用单个 Postgres 承载生产级图记忆**」+ Cognee Cloud——而这一条恰好是我们要用的功能。"单 Postgres 全栈"是 demo |
| `letta-ai/letta` / `letta-code` | 24,770 / 3,363 | Apache-2.0 | 主线已是 CLI harness；MemFS 用 git 追踪全部上下文 | 🔴 结论**要改写**：`letta` 仓**已无代码**（该仓本身已按 §3.2.1 规则2 出局，此行留作**引用纠正**）；"必须换运行时"仍成立，可用的是 `letta-code`，部署形态改为"装 CLI" |
| `mindscale-noah/MindMemOS` | 985 | ⚠️ **README 称 MIT，仓内无 LICENSE 文件** | schema learning + dreaming + memory↔skill | 🔴 许可未落地 + Qdrant/Neo4j/Kafka 三件套 |
| `xzf-thu/VoiceMem` | 1,792 | Apache-2.0 | 语音 agent 记忆（左脑信息/右脑情绪） | 🔴 对文本 CLI 平台不适用 |
| `IAAR-Shanghai/Awesome-AI-Memory` | 1,218 | Apache-2.0 | 榜单（794 篇论文 + **82 个开源系统**） | ⚪ **名单价值 > 可用性价值**：含 **5 个 404 死链**；82 项里只有 10 个与本报告重合 → **70 个我们从未提过** |
| `NirDiamant/Agent_Memory_Techniques` | 1,067 | Apache-2.0 | 30 个可跑 notebook（教学） | ⚪ 框架层只覆盖 Graphiti / Mem0 / Letta / Zep——**与我们的短名单独立吻合**，算一次交叉验证；无 hook 注入、无安全 |

**活跃度**（最后提交，2026-09-17 读数）：`ACE` 09-11 · `cognee` 09-09 · `letta` / `letta-code` 09-10 / 09-17 ·
`MindMemOS` 08-29 · `VoiceMem` 09-14 · `Awesome-AI-Memory` 09-15 · `Agent_Memory_Techniques` 09-15 —— 全部在窗口内。

> **被剔除、但设计留下**：`LycheeMem`（2026-07-17 停更）**作为方案**出局，而它的两个设计——**零 LLM 融合**（cosine 去重聚类）
> 与**动作锚定的检索状态**——仍可直接搬到 pgvector，不需要引入图库。这是 §3.1.1 那条教训的落实：剔除只剔候选，不丢信息。

**论文系的开源实现**（"仅论文"= abs 页与全文中均无代码链接，且 GitHub 无官方实现）：

| 论文 | 实现 | ★ | license | 备注 |
|---|---|---|---|---|
| HippoRAG（NeurIPS'24） | `OSU-NLP-Group/HippoRAG` | 4,010 | MIT | 官方，仍在维护 |
| LightRAG（EMNLP2025） | `HKUDS/LightRAG` | 39,714 | MIT | 官方，最热的 graph-RAG 实现 |
| GraphRAG（微软） | `microsoft/graphrag` | 36,004 | MIT | 官方 |
| SAGE（arXiv 2409.00872） | ❌ **仅论文** | — | — | 已核实无官方实现 |

**活跃度**（最后提交，2026-09-17 读数）：`HippoRAG` 09-03 · `LightRAG` 09-17 · `graphrag` 09-16。
A-mem、MemoryBank、MemInsight 已按规则1 剔除（§3.2.1）；**论文引用本身不受影响**（附录 A）。

**榜单里 ★≥13k 且我们从未提过的其他名字**（完整 82 项见 `D3-research-and-graph.md`，**v2.2 已过筛**）：
`SuanmoSuanyangTechnology/MemoryBear`（★6,276，最后提交 09-11）· `agentscope-ai/ReMe`（★3,482，已升为 POC 首选）·
`Mirix-AI/MIRIX`（★3,443，08-20）· `MemMachine/MemMachine`（★3,221，09-17）·
`AMAP-ML/SkillClaw`（★2,616，**08-17，恰好卡在窗口边界**）· `ModernRelay/omnigraph`（★1,132，09-16）……
以及 ★<1k 的 40 余个个人项目。
**v2.2 从这批里剔除 3 个停更的**：`volcengine/MineContext`（2026-05-07）、`puppyone-ai/puppyone`（2026-06-17）、
`kingjulio8238/Memary`（2024-10-18）→ 见 §3.2.1。
留下的 5 个本轮补核了一次"**可自部署**"（看仓库根）：MemoryBear 有 `identity-service` / `gateway-service` / `mem-knowledge` 多服务；
MIRIX 有 `Dockerfile` + `docker-compose.yml`；SkillClaw 有 `evolve_server` + `requirements-server.txt`；omnigraph 是 Rust
workspace + `deploy/` + `Dockerfile` —— **仍未逐个通读 README，故"是否空壳"仍未判定**。**v2.3 已用 `gh` 把榜单 85 个仓库整体过了一遍**（附录 C U8）：5 个已死链，
活跃且本报告从未提过、★≥300 的只有 4 个（`marciopuga/cog` 376 · `MaxFreedomPollard/Compartment` 581 · `samvallad33/vestige` 624 · `Tura-AI/tura` 637），
其余活跃项 ★<300 —— 即**漏网风险已被量化**。

#### D. 中文生态 / 云厂商（D4）

| 项目 | ★ | license | 结论 |
|---|---|---|---|
| **`agentscope-ai/ReMe`** | 3,482 | Apache-2.0 | 🟢 **唯一同时满足"国内团队 / Apache-2.0 / 本地优先文件式记忆 / 原生 MCP+HTTP+CLI / 默认只绑 127.0.0.1 / 向量可关"** → POC 首选（§3.1.3） |
| `neo4j-labs/agent-memory` | — | Apache-2.0 | 🟢 见 §3.1.3；标注 Experimental + Community Supported |
| `MemTensor/memmy-agent` | 1.9k | MIT | 🟡 与 MemOS 同门；主体已转向 TS 插件 |
| `fuyuxiang/echo-agent` | 1,048 | MIT | 🟡 "agent 自带记忆"，**无独立包/服务，只能 vendor**；四层认知记忆 + 遗忘曲线 + 矛盾检测可作算法参考 |
| `shiwenwen/hope-agent` | 1,604 | — | 🟡 同上；`ha-memory` 的 Dreaming 离线整理流水线可参考 |
| `zhayujie/CowAgent` | 47k | — | 🟡 同上（原 chatgpt-on-wechat）；IM 渠道 + 三层记忆 + Deep Dream 的产品形态参考 |

**活跃度**（最后提交，2026-09-17 读数）：`ReMe` 09-16 · `neo4j-labs/agent-memory` 09-16 · `memmy-agent` 09-15 ·
`echo-agent` 09-01 · `hope-agent` 09-16 · `CowAgent` 09-17 —— **六个全在窗口内**。

**云厂商托管记忆：全部纯托管，内网自托管直接出局（规则2）** —— 下面四项**已从候选剔除**（§3.2.1）：
**AWS AgentCore Memory**（官方定位 *"fully managed service"*）、**Google Agent Platform Memory Bank**（原 Vertex AI
Agent Engine Memory Bank，旧 URL 已重定向）、**Microsoft Foundry Agent Service memory**（仍是 preview）、
**OpenAI ChatGPT memory 与 Responses API 会话状态**（**不通过 API 暴露**；`store` 是**会话状态，不是抽取式长期记忆**）。

同线但**可自托管**的形态另算：**Zep** 云不可托管、自托管只剩 **Graphiti**（已收进 §3.1）；**Mem0 Platform** 不可，但
**mem0 OSS 可自托管**；**Redis Agent Memory** 是托管产品，其开源 `redis/agent-memory-server` 已被官方降级到 `V0/`
（README：*"no longer actively maintained and is not the supported production path"*，详见 §3.3-②）；**Supermemory local**
（MIT，`supermemory local`）与 **Neo4j Agent Memory**（自托管形态）✅；**LangMem / LlamaIndex memory** 是**库**，随你部署 ✅。

#### E. 官方 MCP registry 里的记忆 server（41 个 → 可自托管 2 个）

用官方 registry API（`registry.modelcontextprotocol.io/v0/servers?search=memory`）拉全量：**41 个记忆类 server，
绝大多数是托管 SaaS / 按次付费的 x402 微服务**，不符合"自托管 + 多租户"的要求。自托管或形态特殊、值得记一笔的：

| server | 仓库 | 形态 | 为什么记它 |
|---|---|---|---|
| `ai.rcll/fleet-memory` | `Holetron-lab/fleet-memory` | 团队共享，L0–L3 分层 | **读路径不用 LLM**（团队共享记忆里少见） |
| `io.github.0Mattias/bettermemory` | `0Mattias/bettermemory` | 本地 | **写入前先与文件系统 / git 校验**，反投毒思路 |

**v2.2 规则2 处理**：其余条目——`com.letta/memory-mcp`（Letta 本体仓内已无代码，只剩托管出口）、`ai.korely/memory`（纯托管、
双时序事实）、`ai.smartmemory/smartmemory-mcp` 与 `com.cortex-mem/memory`（托管 / 自托管混合）——作为**托管 SaaS**出局。
上表两行的活跃度：`fleet-memory` 最后提交 **2026-08-28**（仓库根有 compose 栈，能自己跑）、`bettermemory` **2026-09-17**
（纯本地 + 写入前与文件系统 / git 校验）。

**结论**：MCP 记忆 server 已是红海，但**同时满足"自托管 + 多租户 + 复用我们现有 PG"的极少**——
这也解释了为什么 §3.1 的短名单仍然只有那几条路。

#### 3.2.1 按两条规则剔除的候选（v2.2 留痕）

**规则（2026-09-17 执行）**：① **停更** —— 最后提交早于 **2026-08-17**（一个月）；② **不可自部署** —— 云厂商托管、闭源引擎、
仓库里根本没有服务端。**为什么单独留成一张表**：v2 曾把底稿候选静默删掉（§3.1.1 就是证据），所以本轮剔除**一律留痕**——
名字、最后提交、命中哪条、原来为什么值得看，都在下面；§3.1 / §3.2 的表格里只留通过两条规则的。

| 项目 | 最后提交 | 命中 | 原结论 / 备注 |
|---|---|---|---|
| `zhongwanjun/MemoryBank-SiliconFriend` | 2023-05-24 | 规则1（3 年） | AAAI'24 官方实现；**§0.3 引用的艾宾浩斯遗忘曲线出处仍保留**（论文引用与候选资格是两件事） |
| `kingjulio8238/Memary` | 2024-10-18 | 规则1（近 2 年） | 榜单 ★2.6k 项 |
| `amazon-science/MemInsight` | 2025-09-15 | 规则1（1 年） | 官方实现；⚠️ 同名仿冒 `eng-rana-s/MemInsight`（`main.py` 只有 1 字节）仍别用错（§3.3-③） |
| `agiresearch/A-mem` | 2025-12-12 | 规则1（9 个月） | 论文官方实现；代码完整、被大量论文当 baseline，引用见附录 A |
| `memodb-io/memobase` | 2026-01-11 | 规则1（8 个月） | 底稿 B 的用户画像型长期记忆（Apache-2.0） |
| `0xK3vin/MegaMemory` | 2026-05-03 | 规则1 | 提交与 release 同一天停 |
| `volcengine/MineContext` | 2026-05-07 | 规则1 | 榜单 ★5.5k 项 |
| `shaneholloman/mcp-knowledge-graph` | 2026-05-29 | 规则1 | 单文件 `index.ts`；功能是 basic-memory 的真子集 |
| `alash3al/stash` | 2026-06-14 | 规则1 | Go + PG+pgvector，但开源版无多租户 |
| `puppyone-ai/puppyone` | 2026-06-17 | 规则1 | 榜单 ★1.3k 项 |
| `campfirein/byterover-cli` | 2026-06-25 | 规则1 + 规则2 | 同时是 **ELv2 非开源**；`updated_at` 骗人的典型（§3.2 B 坑①） |
| `BAI-LAB/MemoryOS` | 2026-07-07 | 规则1 | 论文代码（EMNLP 2025 Oral）；**分层记忆 OS 的设计仍值得读** |
| `memvid/memvid` | 2026-07-14 | 规则1 | 单文件 `.mv2`，与多租户平台错配 |
| `LycheeMem/LycheeMem` | 2026-07-17 | 规则1 | 原 🟡「形态参考」——**两个设计已摘进 §3.2 C 的注**，不随仓库一起丢 |
| `aiming-lab/SimpleMem` | 2026-07-24 | 规则1 | 压缩优先 + 多模态；MCP 只做纯文本 |
| `FlowElement-xinliuyuansu/m_flow` | 2026-08-03 | 规则1 | 依赖栈重（图库 + 人脸识别，仅 Linux 全 Docker） |
| `Prismer-AI/PrismerCloud` | 2026-08-06 | 规则1 + 规则2 | 仓库里**根本没有服务端**（hook 打向云） |
| `Mibayy/token-savior` | 2026-08-10 | 规则1 | ⚠️ **`hooks/` 仍是本批最好的注入样板**——剔除只针对"拿它当方案"；作者自己撤回了 benchmark |
| `OpenMemory` | 已归档 | 规则1 | 仓库被作者归档，不再维护 |
| `moorcheh-ai/memanto` | 2026-09-17（活跃） | **规则2** | 自托管仍要**云 key + 引擎闭源**：活跃也没用 |
| `getzep/zep`（Zep CE） | 2026-09-11（活跃） | **规则2** | CE 官方停止支持、代码进 `legacy/`；**自托管只剩 Graphiti**（已收进 §3.1） |
| `letta-ai/letta` | 2026-09-10（活跃） | **规则2** | 仓内**已无代码**（仅 README/PRIVACY/TERMS）→ 没有可部署产物；代码在 `letta-code`（保留） |
| `MemoriLabs/Memori` | 2026-09-03（活跃） | **规则2**（半） | OSS 收缩成 SDK + BYODB，**服务端 MCP / on-prem 只对企业开放** → 作为"现成服务端"出局 |
| AWS AgentCore Memory · Google Agent Platform Memory Bank · Microsoft Foundry Agent Service memory · OpenAI ChatGPT memory / Responses 会话状态 | — | **规则2** | 四项全部纯托管（§3.2 D 的结论句） |
| MCP registry 里的托管项：`com.letta/memory-mcp` · `ai.korely/memory` · `ai.smartmemory/smartmemory-mcp` · `com.cortex-mem/memory` | — | **规则2** | 见 §3.2 E |

**恰好卡在窗口边界、这次不剔除但下次要看的**（留个提醒，免得下次误判）：
- `AMAP-ML/SkillClaw` —— 最后提交 **2026-08-17**，正好一个月；它有服务端（`evolve_server` + `requirements-server.txt`），**下次清理若仍不动就出局**。
- `ReflexioAI/claude-smart` —— 提交 08-28，但 **release 停在 2026-07-15**（代码在动、发版停了）。
- `LangMem + LangGraph store` —— 提交 09-09，但 **PyPI 包停在 0.0.30 / 2025-10-27**（仓库活、发版停）→ 用它之前先看包版本。
- `Dataojitori/nocturne_memory`（08-27）、`mindscale-noah/MindMemOS`（08-29）—— 都在窗口内，保留原来的 🟡 / 🔴 理由。

### 3.3 已排除的候选与三条必须澄清的误导

**本节与 §3.2.1 的分工**：**§3.2.1 = 按「停更 / 不可自部署」两条规则出局的**（含底稿 B 的 OpenMemory、Zep CE、
memobase、MemoryOS，以及 v2.1 曾排除的 PrismerCloud、memanto、memvid 等）；**本节 = 另有理由的排除**，与新鲜度、可部署性无关。

**仍在排除名单、但理由不是这两条的**（因此仍留在 §3.2 各表内并标 🔴）：

- **cognee 全 Postgres**：结论要照旧保留——LICENSE 其实是**纯 Apache-2.0**（无附加条款），收费点精确在"**用单个 Postgres
  承载生产级图记忆**"这个用法本身 + Cognee Cloud；而这一条恰好就是我们要用的功能，"单 Postgres 全栈"只是 demo。
- **Letta**：`letta-ai/letta` 仓内已无代码（规则2 出局，见 §3.2.1），但 **`letta-code` 保留**——"必须换运行时"仍成立。
- **utopia**（记忆功能尚未实现）、**EverOS**（官方明说无任何内置鉴权）、**semantica**（`GROWTH.md` = 星标漏斗 playbook）、
  **MindMemOS**（仓内无 LICENSE 文件）、**VoiceMem**（语音栈）、**AgentRecall-X**（★/fork 6.4）——**这些仓库都是活的**，
  否掉的理由与两条规则无关。

**三条必须澄清的误导**（都来自本轮一手核实，写进报告以免后人再踩）：

1. **TiDB 的 `agent-memory` topic 与 Dolt 的 `agent-memory-server` topic 都只是生态标签**——
   官方**没有任何** agent 记忆组件（TiDB 只有向量/全文/混合检索 + 一个"自己写"的示例；Dolt 只有一句 README 话术 + 一个通用数据库 MCP）。
   **不要把 topic 当成"官方记忆能力"。**
2. **"`redis/agent-memory-server` 开源可自托管"已过时**：Redis 现在的产品是托管的 **Redis Agent Memory**（Redis Iris 的一部分），
   仓库里的开源服务器被官方降级到 `V0/`，README 原文 **"no longer actively maintained and is not the supported production path"**。
   把它写成"Redis 的官方自托管方案"是**错误**表述。
3. **同一个项目别引错仓库**：Zep 论文的引擎是 **Graphiti**（不是 `getzep/zep`）；Letta 的代码在 **`letta-code`**（不是 `letta`）；
   MemInsight 要用 **`amazon-science/`** 那个。

### 3.4 腾讯 TAM 专题

**它是什么**：`TencentCloud/TencentDB-Agent-Memory`（README 的 clone 地址误写成 `Tencent/…`），
★26.9k / fork 2.5k（2026-09-17 读数），**MIT**，2026-04 创建，当前 v2.0.1-beta.1；
⚠️ **默认分支是 `feat/server_team`，不是 main**。

四个记忆资产：Chat Memory（L0 对话 → L1 原子 → L2 场景 → L3 画像）、Skill（带版本与触发边界，私有 → 团队共享）、
Wiki（文档 + 链接图）、CodeGraph（符号 / 调用 / 影响面）。可见性分 private / team / restricted(ACL) / agent 四档。
自报 PersonaMem 48% → 76%。

**最大特点：注入发生在传输层。** README 原话 **"One Proxy, unchanged protocol, zero-code integration — point the
Agent's base URL to the Proxy"**——不需要插件、hook，**也不需要 MCP Server**。
（仓库里唯一的 MCP 是 `MemoryKnowledge/src/mcp/server.ts`，只把 Wiki/CodeGraph 的 HTTP API 包成 stdio；记忆本体没有官方 MCP。）

**三件套**（公开 Docker Hub，多架构 amd64/arm64，无需登录）：

| 组件 | 镜像 | 端口 | 作用 |
|---|---|---|---|
| memory-core | `agentmemory/memory-core` | 8420 | 内核 gateway：记忆读写、鉴权、skill / RAG 数据面 |
| memory-hub | `agentmemory/memory-hub` | 8125 / 8424 | 管理面板（Panel）+ Knowledge（Wiki / CodeGraph） |
| proxy | `agentmemory/memory-proxy` | 8096 | LLM 转发代理，agent 的 API 入口（Anthropic + OpenAI 双协议） |

`deploy/global-images/start-all.sh` 一键拉起；数据落 named volume（默认 sqlite；MongoDB 是试验特性且**切换不迁移**）。
需要**两组独立 LLM**：`MEMORY_LLM_*`（抽取 / embed / 总结）与 `PROXY_UPSTREAM_*`（转发）——两组都能指向 DeepSeek。
`MEMORY_PROMPT_MODE=code`（默认）才是工程场景的提示词族；chat 模式下纯工程对话可能抽出 0 条记忆。

> **v2.3 实机进度**（附录 C U3/U17）：**三件套全部在本机拉起过**（memory-core 8420 / memory-hub 8125+8424 / memory-proxy 8096，
> 独立网络 `tamlab-net` 与独立卷），**proxy 那一跳已跑通**——扮演 Claude Code 打 `POST /claude-code/default/v1/messages`、
> 上游指向本机 mock，**200 转发 + 注入 + 会话初始化全部发生**；注入块与隔离、可关性都实测了（见 §4.4、§8.3.1）。
> 仍未测的只有"**真实 Multica 任务里的跨任务召回**"（要动某个测试 agent 的 `custom_env`）。
> 实验用容器、卷、网络**已全部删除**；三个镜像（共 ~4.5 GB）刻意留着便于复跑。

**社区桥接**（❓未实测，仅备查）：`kuaizhongqiang/TencentAgentMemoryBridge`（包成 MCP Server + OpenClaw 插件）、
`baodq97/tencentdb-agent-memory`（Claude Code 插件）、`RayYeung1989/dsh-memory-tencentdb`、`kuro1981/memory-tencentdb-nix`。

---

## 4. 五种接入架构

### 4.1 对比

| 架构 | 改代码？ | 覆盖 provider | 自动召回 | 成本 | 可审计 |
|---|---|---|---|---|---|
| **A 契约层**：skill + `workspace.context` + 一个 git 知识库仓库（作为 `project_resource` 的 `github_repo`，agent 用 `multica repo checkout` 取） | ❌ 0 | 全部 | ⚠️ 靠人写 / agent 主动读 | 极低 | 高（git diff / PR） |
| **B MCP 记忆后端**：`multica workspace mcp add` → `multica agent mcp add` | ❌ 0 | 21 个 MCP provider | ❌ 模型自决 | 低 | 中（后端侧日志） |
| **C 平台侧召回注入**：claim 响应带"记忆索引" → daemon 渲染进 brief | ✅ 150–250 行 | **全部**（含 cursor / reasonix） | ✅ | 中（每轮 token 预算可调） | 高（注入即日志） |
| **D per-provider hook 适配**：给 claude/codex/opencode/openclaw/hermes/dsh 各挂薄脚本 | ✅ 每种 ~50 行 + 参数白名单适配 | 6 个（cursor 只能降级） | ✅ | 中高（每种都要维护，且随上游 hook API 变） | 高 |
| **E 传输层代理**：agent 的 base URL 指向 TAM proxy，由它注入 / 回流 | ❌ 0（只改该 agent 的 `custom_env`） | ❓ 理论 8 个（Cursor 适配尚在 TAM roadmap 上） | ✅ 代理保证每轮 | 中（+3 容器 + 两组 LLM 账）；**多一个流量中间人** | 中高（Panel 可看注入内容，注入逻辑不在你手里） |

> **v2.3 实测修订**（附录 C U1/U2/U4）：
> - **D 的"6 个"里 claude 已实测可用**——用户级插件的 hooks 在 daemon 同款调用参数下**会触发**（探针落盘 + debug 显示
>   `Registered 75 hooks from 31 plugins`），**零代码**；粒度是 **per-host / per-user**，要按 agent 隔离得给该 agent 的
>   `custom_env` 设 `CLAUDE_CONFIG_DIR`（它不在 daemon 黑名单里）。
> - **Reasonix**：有 11 个 hook 事件，但 `UserPromptSubmit` 的 stdout **被丢弃**（只当闸门）⇒ **每轮注入必须走 `AGENTS.md`**，
>   "同一套薄脚本挂所有 provider 做每轮注入"在它身上会**静默失效**。
> - **E 的"覆盖全部 provider"证伪**：`custom_env` 的黑名单只拦 `MULTICA_*` + 14 个具名键，所以端点变量的可设性是逐 provider 看的——
>   **claude / codebuddy / opencode / dsh 可直接设**（`ANTHROPIC_BASE_URL` / `DEEPSEEK_BASE_URL`）、**codex 只能走 `custom_args`**
>   （`CODEX_HOME` 被拦，但 `-c model_providers.…base_url=` 不被拦）、**reasonix 只能改配置文件**（它没有任何 `*_BASE_URL` 环境变量）、
>   **cursor 未证实**（本机 `cursor-agent` 是悬空软链、仓库里也没有端点变量）。

### 4.2 推荐组合与判定

**A + C 是最优组合**：A 今天就能用，C 拿到"Multica 自己拥有的记忆"（这正是上游注释里的
"Multica-owned, user-visible, project- or issue-scoped memory store"，也是 #838 的形状）。
**B** 用于验证与"模型主动查"的场景；**D** 只在 C 无法覆盖的 provider 上补；**E** 是"今天就要可用形态、
且愿意多运维一个服务"时的最短路径——它用 0 代码换到了自动召回 + 自动回写，但**记忆的契约与数据都不在你手里**，
与上游 "own the contract, not the store" 的取向相反。

选择路径见 §0.2 的判定表。

### 4.3 架构 C 的逐文件改动清单

已核对到文件与行：

| 位置 | 改动 |
|---|---|
| `server/migrations/xxx_agent_memory.up.sql` | 新表：`workspace_id / project_id / agent_id / scope / name / description / content / source_task_id / importance / valid_at / invalid_at / created_by`（+ 索引） |
| `server/pkg/db/queries/memory.sql` | sqlc 查询（list index / get / upsert / invalidate） |
| `server/internal/handler/agent.go:380-385` 附近 | ClaimTask 响应加 `memory_index`（对照现有 `WorkspaceContext` 字段） |
| `server/internal/handler/daemon.go:2817-2828` 附近 | 填充：按 (workspace, project, agent) 取索引，按重要性 / 新鲜度截断到预算 |
| `server/internal/daemon/types.go:86-90` | 镜像字段 |
| `server/internal/daemon/execenv/execenv.go:199-203` 附近 | `TaskContextForEnv` 加字段 |
| `server/internal/daemon/execenv/runtime_config_sections.go:187` / `:998-999` | 新增 `writeMemoryIndex()` 并挂进 section 列表（**只注入索引**，全文用 `multica memory get` / MCP 工具按需取） |
| `server/internal/daemon/prompt.go:63-72`（`perTurnContextBlocks`）+ `execenv/runtime_config_sections.go` 新增 `BuildMemoryIndexBlock` | **每轮用户消息**里注入"与本轮相关的记忆索引"——**编码时使用的关键**（每轮重发、压缩冲不掉、不破坏 prompt-cache 前缀）。照 `BuildConnectedAppsBlock` / `BuildTaskInitiatorBlock` 的现有写法即可，见 §8.6 |
| `server/cmd/multica/cmd_memory.go`（新） | `memory list/get/add/forget`，让 agent 能读全文、人能改能删 |

> **只注入索引**是这个设计能被接受的前提——常驻内容是每轮成本，必须小（约束 1）。
> **注（v2.1）**：上表最后一行是本轮新增的——如果记忆主要在**编码时**用，只在 claim 时注入一次不够（§8.6）。

### 4.4 架构 E 接入手册（TAM）

**前提结论：能在 Multica 的任务里用，但必须走 header 预选，不能走表单。**

1. **冲突**：proxy 的会话初始化默认是**交互式表单**——在 Claude Code 上就是用 Anthropic 原生 `AskUserQuestion`
   发 `tool_use`，让用户选 team → agent → task。而 Multica 恰恰传了 `--disallowedTools AskUserQuestion`（§2.3 约束 4）。
   表单路径在 daemon 任务里无法完成。
2. **出路**：`start-proxy.sh` 生成的配置里 **`sessionInit.headerAutoSelect.enabled: true`**，header 齐全即跳过表单。
   判定逻辑是 `canRegister = !!agentId && !hadMismatch`——**team + agent 两个 header 就够，task 是可选的**
   （未知 / 过期的 `x-task-id` 只把召回面放宽，不拦注册）。
   ⚠️ 上游 `agents/README.md` 仍写"`x-task-id` 当前必填"，与代码不一致，**以代码为准**。
   ✅ **v2.3 已实测**（附录 C U3）：带 `x-team-id` + `x-agent-id` + 客户端 session id → **直接转发**；
   **不带身份 header → 返回流式的 `msg_cc_session_init_…` 表单**——即 §7 R6 那条失败模式被**在实验里复现**了。

   ⚠️ **v2.3 新增的前置条件（会直接影响接入，只给两个 header 不够）**：进入注入路径还要求**客户端 session id header**
   （Claude Code 发的是 **`x-claude-code-session-id`**，见 `MemoryProxy/src/session/session-key.ts:9-19`）。
   缺它 → `conversationId = null` → **`injectedSkipped = true`** → **静默跳过 sessionInit 与全部注入**（不报错）。
   **好消息**：真实的 Claude Code 自带这个 header，所以 Multica 的 claude 运行时**天然满足**；
   但用手工 `curl` 验证时漏掉它，就会误判成"注入没生效"。
   另：鉴权 key 读的是 `x-api-key` 或 `Authorization: Bearer`（即 `ANTHROPIC_AUTH_TOKEN`），**不是** `x-tdai-user-key`。
   📌 **一处对旧版本的修正**：proxy ↔ **memory-hub** 那一跳在当前版本**已不存在**——`MemoryProxy/src/session/registrar.ts`
   开头写明"本地构建 SessionInfo，**不再调用 TMC**"，会话状态落在 proxy 本地（内存 + Redis/SQLite）；
   hub 现在只提供 Panel 与 Knowledge。
3. **Multica 侧怎么把 header 带上**：Claude Code 支持 `ANTHROPIC_CUSTOM_HEADERS`（格式 `Header-Name: value`，
   已在 claude CLI 2.1.274 二进制中核实），所以只要给这个 agent 的 `custom_env` 加
   `ANTHROPIC_CUSTOM_HEADERS="x-team-id: <tid>\nx-agent-id: <aid>"`。
   Multica 的 `custom_env` **不做名字白名单**，且 DB 里已有同类先例——两个 Android agent 就是靠它把
   `ANTHROPIC_BASE_URL` 指向 `127.0.0.1:8787`（headroom）。
4. **完整链路**（复用现有 headroom，不必绕过）：TAM 的 `upstream.agents.<agent-name>` 支持按 client 名单独覆盖上游 URL + key，于是：

   ```text
   claude(daemon 拉起) --ANTHROPIC_BASE_URL=http://127.0.0.1:8096/claude-code/default--> TAM proxy :8096
        --upstream.agents.claude-code.url=http://127.0.0.1:8787 + apiKey--> headroom :8787 --> api.deepseek.com/anthropic
   ```

5. **只改该 agent 的 `custom_env`**：`ANTHROPIC_BASE_URL`（**要带 `/claude-code/default` 路径**）+
   `ANTHROPIC_AUTH_TOKEN`（Panel 里签的**业务用户** `sk-mem-…` user_key，不要用 admin key）+
   `ANTHROPIC_CUSTOM_HEADERS`。headroom 侧的 `model` / `DEFAULT_*_MODEL` 保持原样。

**两个必须先知道的默认值陷阱**：

- `start-proxy.sh` 生成的最小配置里 **auth / sessionInit / tdai 三个开关默认全 0**（只做纯转发），
  要 `PROXY_FULL_STACK=1` 才一并打开（且开 sessionInit 会自动补 auth）。别拿默认配置去测"记忆没生效"。
- 默认内部凭据是 `MEMORY_CORE_GATEWAY_API_KEY=local`、admin 用户名与登录 key 都是 `admin`——
  **端口一旦非回环暴露，等于把 system_admin 送人**，自托管必须先换。

**排查路径**：任务卡在首轮时，先看 proxy 日志的 sessionInit 阶段，而不是 Multica 的报错——header 预选失败会回落到
表单，而 Multica 这个 client 弹不出来（见 §7 R6）。

### 4.5 架构 A+B 接入手册（ReMe）

ReMe 有**三种**接法，先按"要不要改 daemon 代码"分清楚——**它们不是互斥的，A+B 是同一套东西的两半**：

| 接法 | 改代码？ | 自动召回 | 形状 |
|---|---|---|---|
| **B：MCP（今天可用）** | ❌ | ❌ 模型自决 | `workspace mcp add` + `agent mcp add` → 工具出现在 agent 手里 |
| **A：契约层（今天可用）** | ❌ | ⚠️ 靠 skill 规定时机 | ReMe 的 `workspace_dir` 指向 GitLab 记忆仓库 + 一个 skill 教"何时读、何时写" |
| **C：每轮注入** | ✅ §4.3 | ✅ | daemon 每轮向 ReMe 要索引 → 这才是"编码时自动召回"（§8.6） |

**A+B 合起来**才是完整形态：**GitLab 是真源（可 diff、可过 MR 审查门），ReMe 是它的索引，MCP/skill 是取用口**。

#### 第 1 步：起 ReMe（宿主上，绑回环）

ReMe 的 HTTP 服务与 MCP 服务**同端口**（`service: {backend: http, mcp_enabled: true, mcp_path: /mcp}`，
`config/default.yaml`），默认 `127.0.0.1:2333`：

```bash
reme start workspace_dir=/srv/multica-memory enable_logo=false
```

⚠️ **必须先给 LLM 凭证，否则起不来**（v2.4 实测：不给就是 `openai.OpenAIError: Missing credentials`，
`as_llm` 组件在启动时就要构造客户端）。四个变量（`config/default.yaml:840-856`）：
`LLM_BACKEND`（`openai`/`anthropic`/`dashscope`/**`deepseek`**/`gemini`/`moonshot`/`ollama`/`xai`）·
`LLM_MODEL_NAME`（默认 `qwen3-plus` 系，**用 DeepSeek 要显式改**）· `LLM_API_KEY` · `LLM_BASE_URL`。
注意这套 LLM **只服务抽取 / 总结 / 演化**，检索本身是 BM25，**没有 LLM 也能查、但不能起服务**。

验证（三条，全部实测过）：

```bash
curl -s localhost:2333/health_check                       # 服务活着
curl -s -X POST localhost:2333/mcp -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"p","version":"0"}}}'
# → 返回 serverInfo {"name":"ReMe","version":"3.4.7"}
curl -s -X POST localhost:2333/search -H 'Content-Type: application/json' \
  -d '{"query":"proguard","limit":3,"search_filter":{"path_prefix":"digest/wiki/android-build"}}'
```

#### 第 2 步：workspace 布局 = 项目隔离的落点

`path_prefix` 是**路径前缀**，所以"一项目一子目录"就是项目硬过滤。ReMe 只**监听两个目录**
（`index_update_loop.watch_dirs: [daily_dir, digest_dir]`），所以项目目录必须开在这两棵树下：

```text
/srv/multica-memory/            # = GitLab 记忆仓库的 checkout
├── daily/2026-09-30/<project>/*.md
└── digest/wiki/<project>/*.md
```

检索一律带 `search_filter.path_prefix = "digest/wiki/<project>"`（v2.4 实测：不加过滤 3 条命中，加前缀后收窄到该项目的 1 条）。
⚠️ 写入到可检索有 **4–20s** 的 watcher 延迟（v2.3 U5），"刚写完就查不到"是正常的。

#### 第 3 步：把 `/mcp` 注册成工作区 MCP server，再绑给 agent（0 代码）

payload 就是 MCP 的通用 server entry（服务端 `mcpTransportOf` 认 `local/stdio/remote/http/streamable-http/sse`）：

```bash
echo '{"type":"http","url":"http://127.0.0.1:2333/mcp"}' | multica workspace mcp add reme --server-config-stdin
multica workspace mcp list            # 取 server id
multica agent mcp add <agent-id> <server-id>
```

⚠️ **对 claude 而言这是"唯一 MCP 集合"**：daemon 会写 `--mcp-config` 并加 `--strict-mcp-config`
（`pkg/agent/claude.go:57-69`、`:1097`），所以绑上之后**运行时本地的 MCP 会与它合并**（`mergeRuntimeAndAgentMcpConfig`），
但不会有别的 MCP 偷偷进来。另外 `agent copy` **不复制 mcp_config / custom_env**（`cmd_agent_copy.go:74-78`），扩编时要脚本化补绑。

#### 第 4 步：一个 skill 规定"何时读、何时写"

MCP 是"模型自决"，所以要靠 skill 把时机写死：开工查 `search`（带 `path_prefix`）、
发现"下次还会再讲一遍"的事实才写、写完按项目路径落 `digest/wiki/<project>/`。
**别给 70 个 agent 各绑一个 MCP**——token 账见 §8.4；少数确实要"模型主动查全文"的 agent 才绑，其余走 CLI。

#### 第 5 步：每轮自动召回 —— 这一步要改代码

到此为止拿到的是**任务级、模型自决**的召回。要做到 §8.6 的"编码时、每一轮、压缩后仍在"，
只有 §4.3 的架构 C（`perTurnContextBlocks()` 里加一个 `BuildMemoryIndexBlock`，向 ReMe 查本轮相关索引）。
**那 150–250 行 daemon 改动需要你确认后才动。**

#### 四个实测坑（都会让你误判"记忆没生效"）

1. **没有 LLM key 起不来**（见第 1 步）——"无凭证可跑"只对 SDK 成立（U5）。
2. **MCP 端点默认要 session**：`service.mcp_stateless_http` 默认 `false`，`tools/list` 不带 `mcp-session-id`
   会返回 `-32600 Missing session ID`。客户端不支持会话时（或走裸 curl 验证时）设 `service.mcp_stateless_http=true`。
3. **macOS 系统代理会劫持 localhost**（v2.4 实测）：开了 Clash 一类系统代理后，Python 的
   `urllib.request.getproxies()` 返回 `{'http':'http://127.0.0.1:7897'}`（**不含例外列表**），于是 httpx 把
   `127.0.0.1:2333` 也送进代理 → **502 Bad Gateway**，而 Go 的 daemon 与 Node 的 claude **读的是环境变量、不受影响**。
   给 ReMe 自己的 CLI / SDK 加 `NO_PROXY=127.0.0.1,localhost` 即可。
4. **默认只绑 `127.0.0.1`**：daemon 与 agent 都在宿主机上，所以可达；一旦把 ReMe 挪进容器/别的机器，
   要先改 `service.host` 并自己补鉴权（ReMe 默认无鉴权）。

---

## 5. 落地计划

| 阶段 | 做什么 | 验收标准 |
|---|---|---|
| **P0 · 今天**（0 代码） | ① 定"什么算项目事实"的规范（要"下次还会再讲一遍"才算）；② 写一个 skill 明确**何时读、何时写、写什么**；③ 可选：建知识库 git 仓库并作为 `project_resource` 挂到项目 | 下一次同类 issue，agent 不再重复探索同一件事；记忆条目可 git diff |
| **P1 · 1 天** | 起一个记忆后端（先官方 memory MCP 验证"记忆有没有用"；**要产品化就试 ReMe**，见 §3.1.3），`workspace mcp add` + 绑 1 个测试 agent | 跨两次任务能召回；同时记录 **token / 延迟 / 命中** 三列 |
| **P2 · 1–2 周** | 架构 C：表 + claim 注入 + daemon 渲染 + `multica memory` CLI；带**总预算**与**开关**（默认关，可 A/B） | A/B：一次通过率提升，护栏（p95 延迟、每任务 token）不破 |
| **P3 · 之后** | ① 捕获回写（插件 `event` 触发，或任务结束钩子）走**审查门**（PR / 人工确认）再落库；② 写入审计 diff + 失效而非删除；③ 记忆体检与 TTL | "写入-读取比"可观测；污染事件有回滚演练 |

**P1 的备选路径（架构 E）**：也可以换成"拉起 TAM 三件套 + 给 1 个 claude agent 加 3 个 `custom_env`"，
一次拿到自动召回 + 自动回写（P1 + P3 一步到位），代价是模型流量多一跳、多一个要运维的服务与两组 LLM 账。

**三条路径的取舍**：先验证"记忆到底有没有用"→ 官方 memory MCP（最便宜）；想直接拿到可用形态 → TAM；
想把记忆变成平台自己的契约与数据 → P2 的架构 C。

---

## 6. 评测方案

**离线**（钉住版本，随代码回归）

- **LoCoMo** 极长期对话记忆（arXiv:2402.17753）
- **LongMemEval** 五项能力分项 + **必须含弃答（abstention）**（arXiv:2410.10813）
- **tau2-bench** 任务成功率（arXiv:2506.07982）
- 诊断：MemoryAgentBench（arXiv:2507.05257）/ MemBench（arXiv:2506.21605）；安全：MemSecBench（arXiv:2607.27080）
- **每次必须同时报 准确率 / token / 延迟 三列**

**线上**（才说明净收益）

- A/B：记忆开 / 关，或"索引注入 vs 索引 + 按需取"；分流单位优先 workspace（避免串味）
- 主指标：任务一次通过率（或 issue 关闭率）；护栏：p95 延迟、每任务 token
- 诊断：**召回命中率**（该召回时是否召回）与**写入-读取比**（只写不用的记忆 = 纯成本）

---

## 7. 风险登记

| # | 风险 | 影响 | 缓解 / 现状 |
|---|---|---|---|
| R1 | **记忆投毒** | 团队可见记忆被污染，且可被其它 agent 继承 | 写入内容当不可信输入；不要只依赖内容检测或血缘（都可洗白）；密钥不入记忆（应用侧剥离 + 路径穿越防护）；落库前过审查门（P3） |
| R2 | **写入放大的 token 账单** | 凡带 LLM 抽取的后端，成本随写入线性增长 | 抽取异步、可节流、有上限；用便宜模型做抽取（TAM 的 memory 组 / proxy 组分离就是为此） |
| R3 | **许可** | 做成对外产品时才成为问题 | **AGPL-3.0 命中三个**：OpenViking、basic-memory、**honcho**（外加 `openwolf`）；**ELv2 一个**：byterover（非 OSI 开源、不可 fork；**已按 v2.2 规则剔除**，见 §3.2.1）；mem0、Graphiti、**MemOS**、**memU**、**ReMe**、Neo4j Agent Memory = Apache-2.0；**supermemory、MemPalace、Hindsight、TAM = MIT**（最宽松，且 TAM 镜像是公开 Docker Hub） |
| R4 | **引用口径** | 决策依据被高估 | OpenViking 的 benchmark 是自建评测、非第三方复现；TAM 的 PersonaMem 是自报；Mem0 论文与产品文档在"冲突删除"上不一致 |
| R5 | **架构 E：成为模型流量中间人** | 可用性、延迟、"谁能看到 prompt"都多一个环 | 只给单个 agent 开（先拿测试 agent 验证）；保留一键回退（删 `custom_env` 即恢复直连） |
| R6 | **架构 E：header 预选失败回落表单** | 任务**卡在首轮**，且 Multica 侧看不出原因 | team/agent 必须在内核里真实存在；监控 proxy 日志的 sessionInit 阶段；配错时删 header 回退。**v2.3 已实测复现**（U3）：不带身份 header 确实返回流式表单。⚠️ **还有一条更隐蔽的**：缺**客户端 session id header**（`x-claude-code-session-id`）时**不报错**，只是 `injectedSkipped = true` 静默跳过全部注入——真实的 Claude Code 自带它，但手工验证时极易踩 |
| R7 | **架构 E：默认凭据 `local` / `admin`** | 端口暴露即等于交出 system_admin | 上线前换随机长串；只在回环或受控网络监听 |
| R8 | **架构 E：抽取 LLM 读到全部对话** | 数据出域 / 脱敏边界问题 | 抽取用的 LLM 与转发用的 LLM 分开配置；必要时指向本地 Ollama |
| R9 | **架构 E：provider 覆盖只是理论** | 以为覆盖 8 个，实际只有 claude 系 | Cursor 适配尚在 TAM roadmap（v2.0.1 计划），reasonix 无适配；逐个 provider 实测后再宣称覆盖 |
| R10 | **架构 E：上游自身早期** | 升级可能破坏接法 | v2.0.1-beta.1，默认分支停在 `feat/server_team`——**别把默认分支当发布分支用**，镜像固定 tag |

---

## 8. 场景适配：20 人 / 70 个 Android 项目 / 自托管 GitLab / Lark 国际版 / 70+ agents / 编码时使用

> 这一节把前面的结论落到一套**具体条件**上。所有"已核实"条目都给了仓库文件，未核实处标 ❓。

### 8.1 结论先行：主干不变，四个参数变了

主体结论仍是 **A（契约层）+ C（平台侧索引注入）**（§0.2），但在这套条件下：

1. **记忆的"真相层"应该放在 GitLab，而不是新起一个记忆服务。** 理由不是偏好——是**自托管 GitLab 已经被 Multica 当一等公民接进来了**（8.2），于是"记忆写入走 MR 审查门"这条 P3 设计**不需要任何新代码**就能落地，而且天然满足上游"own the contract, not the store"的取向。
2. **不要给 70 个 agent 各挂一个记忆 MCP。** 工作区 MCP 必须**逐 agent 显式绑定**（已核实，8.2-⑤），而 MCP 的工具 schema 是**每轮常驻**成本。70 个 agent × 每任务一份工具清单，是纯浪费——正确做法是**平台侧注入索引**（架构 C），需要取全文的少数 agent 用 CLI 按需取。
3. **作用域要从"agent"改成"(workspace, project)"。** agent 数 ≈ 项目数（70），按 agent 分桶等于造 70 个孤岛；而真正能复利的记忆在**项目维度**（同一团队 70 个 Android App 的共性）与**workspace 维度**（构建/GitLab/Lark 的流程约定）。
4. **"主要编码时使用"⇒ 注入点必须在每一轮，而不是任务开始那一次。** 编码任务是几十到几百轮，中间会 auto-compact；
   而 Multica **已经有一条每轮的用户消息注入通道**（`perTurnContextBlocks`，现在装的是 Task Initiator / Connected Apps 等）。
   所以架构 C 要做**两个注入点**：claim 时的 brief（任务级）+ **每轮的索引块**（编码时召回）。详见 **§8.6**——
   那里也说明了为什么"装个记忆插件就自动召回"在 Multica 里**不成立**（插件没有轮次触发器，`--settings` 又被 daemon 独占）。

### 8.2 已核实的硬事实（决定了上面四条）

| # | 事实 | 证据 |
|---|---|---|
| ① | **自托管 GitLab 是内置能力**：`vcs_connection` 表 `provider IN ('forgejo','gitea','gitlab')` + `instance_url`，凭据用 `MULTICA_VCS_SECRET_KEY` 加密存储；MR 与 CI 管线状态被镜像进 `vcs_pull_request` / `vcs_commit_status`，并按 issue 标识**自动关联 issue、合并后把 issue 移到 Done** | `server/internal/integrations/vcs/gitlab.go` · `server/migrations/216_vcs_integration.up.sql` · [vcs-integration 文档](https://github.com/multica-ai/multica/blob/main/apps/docs/content/docs/vcs-integration.mdx) |
| ② | **该 token 只要 `read_api`，不用于克隆/推送**——"Agents check out repositories in the runtime and **push branches and open PRs using the runtime host's own Git credentials**… Repository checkout works with any Git URL" | 同上文档 §"Agents opening pull requests" |
| ③ | **`project_resource` 的 `github_repo` 类型校验的是通用 git URL**（`isValidGitRepoURL`：http(s) 或 ssh）→ **自托管 GitLab 仓库地址今天就能当项目资源挂上去**，不需要改代码 | `server/internal/handler/project_resource.go:79-106` |
| ④ | **Lark 国际版是一等公民**：`registrationDefaultLarkDomain = https://accounts.larksuite.com`，`region=lark` 映射到 `open.larksuite.com`，前端是两个 CTA（Bind to Feishu / Bind to Lark） | `server/internal/integrations/lark/registration.go:51-52` · `http_client.go:45-46` · `server/internal/handler/lark.go:301-315` |
| ⑤ | **工作区 MCP 必须逐 agent 绑定**：`agent_mcp_server(agent_id, server_id, enabled)`，claim 时只折叠"绑定到该 agent 且启用"的；CLI 是 `multica agent mcp add <agent-id> <server-id>`（**无双 agent / 批量开关**），API 是 `/api/agents/{id}/mcp-servers`（**可脚本化**） | `server/migrations/315_workspace_mcp_server.up.sql:26-35` · `server/internal/handler/workspace_mcp.go:25-57` · `server/cmd/multica/cmd_agent_mcp.go` |
| ⑥ | **Lark 安装是 (workspace, agent) 粒度**：`channel_installation` 的 `agent_id NOT NULL` → **每个 agent 一个 Lark 应用**；70 个 agent 要么装 70 个应用，要么只给需要聊天入口的那几个装 | `channel_installation` 表结构（本机库实测）· `server/internal/handler/lark.go:271-296` |
| ⑦ | **用量可按 (workspace, runtime, agent, project, provider, model) 精确读**：`task_usage_hourly` / `task_usage_daily` 有 `input/output/cache_read/cache_write tokens` 与 `task_count` | `server/migrations/101_task_usage_hourly_schema.up.sql` · `073_task_usage_daily_rollup.up.sql` |
| ⑧ | **"每一轮的用户消息"是一条独立的注入通道**，已在用：`perTurnContextBlocks()` 每轮拼 `## Shared working directory`、未解决合并冲突、会话续接提示、**Task Initiator**、**Connected Apps**。注释说明这些**故意不放 runtime brief**，因为会破坏 resume 的 prompt-cache 前缀稳定性 | `server/internal/daemon/prompt.go:63-72` · `execenv/runtime_config_sections.go:158-215` |
| ⑨ | **插件 hook 没有"提示词 / 轮次"触发器**——只有 `ui / manual / event / agent / schedule`。`agent` 是"把 hook 包成 MCP 工具"（模型自决），没有任何一条能把内容注入到某一轮的 prompt | `server/migrations/362_plugin_hook_engine.up.sql:23` · `399_plugin_hook_schedule.up.sql:33-35` · `server/internal/daemon/plugin_hook_mcp.go` |
| ⑩ | **`--settings` 被 daemon 独占**（`blockedArgs["--settings"]`，daemon 用它做技能开关），所以**"自己塞一份带 hooks 的 settings"这条路在 claude 上走不通**——要挂 per-provider hook 得改 daemon 代码。⚠️ **v2.3 补充**：这条只说明"自己塞 `--settings`"不可行——**claude CLI 自己加载的用户级插件 hooks 不受此限制**，2026-09-17 隔离实测确认其触发（附录 C U1） | `server/pkg/agent/claude.go:771-783` |
| ⑪ | **传输层代理是"每轮都在"的注入点**（一手文档）：L2 Agent Profile 与 L3 Team/Global 记忆**直接注入 system prompt**，`<session_context>` 会话初始化后**每轮追加**，L0 短期记忆**每轮对话结束时主动写回**；L0/L1 则做成工具**按需查询**——原文理由正是"避免破坏上游 KV cache" | [TAM MemoryProxy README_CN](https://github.com/TencentCloud/TencentDB-Agent-Memory/blob/feat/server_team/MemoryProxy/README_CN.md) §记忆层与注入策略 |

> 本机这套自托管实例的实测规模是 **29 agents / 2 workspaces / 5 projects / 156 tasks**（2026-09-17 查库），
> `.env` 里 `MULTICA_VCS_INTEGRATION_ENABLED` 与 `MULTICA_LARK_SECRET_KEY` 都已配置 —— 但它是**开发/验证实例**，
> 与本节描述的 70-agent 生产环境不是同一套；下面的容量与成本结论按生产参数推。

**③ 的一个前置项（必须先解决，否则整条路走不通）**：②说明**克隆与推送用的是 daemon 宿主的 git 凭证**。
70 个私有 Android 仓库 ⇒ **每台 daemon 宿主需要一个覆盖这 70 个仓库的凭证**（GitLab **group deploy key** / group access token / SSH key），
并且要能推分支、开 MR。这是本节所有方案（包括记忆仓库本身）的共同前置，**建议第一步就测通"一个 agent 在一个私有项目上 checkout → 推分支 → 开 MR"**。

### 8.3 作用域怎么定（70 agents × 70 projects × 20 人）

| 层 | 放什么 | 占比建议 | 为什么 |
|---|---|---|---|
| **workspace** | Android 构建约定、GitLab 流程（分支命名 / MR 模板）、Lark 里"找谁"、Multica 用法本身的坑 | ~10% | 70 个 App 共享的部分，写一次全用 |
| **project**（推荐主战场） | 每 App 的模块结构、签名与 flavor 矩阵、内网 maven、已知 flaky 测试、发布节奏 | ~80% | **真正的复利在这层**：同一团队的 App 彼此的坑高度相似，但配置不相同 |
| **agent / personal** | 个人的偏好与临时脉络 | ~10% | 20 人 × 70 项目，个人记忆的边际价值最低 |

**两条必须钉住的规则**：

- **project 是硬过滤，不能靠向量相似度**。70 个 Android 项目"长得很像"（同样的 gradle 版本、相似的权限清单、相近的包名），
  语义检索最容易犯的错就是把 **A 项目的 flavor / 签名配置喂给 B 项目**。检索必须 `WHERE workspace_id=? AND project_id=?`，
  跨 project 的召回要显式开（且只取 workspace 层）。
- **不要按 agent 分桶**。70 个项目 → 70 个 agent，若按 agent 隔离，A 项目的 agent 永远学不到 B 项目的经验，
  而这恰好是这套条件里最大的收益来源。分桶键 = `(workspace, project)`，`created_by` 记人（用 Lark 身份）。
- **冷启动要"人灌"而不是"等它学"**：先把 70 个项目里人已经知道的事实（README、构建说明、签名方式、发布流程）
  批量灌进去，比让 agent 在任务里慢慢积累快一个数量级。审查人用 GitLab `CODEOWNERS` 指定
  （20 人 → 人均 ≈3.5 个项目，正好一人管几个）。

#### 8.3.1 用"项目专属 agent"来做隔离，行不行？

**行，而且这是走 TAM 路线时最现实的一条——但有三处必须先钉住。**

**为什么行**：`agent` 在 TAM 里本来就是**一等隔离维度**，不是我们硬凑的。MemoryCore 文档原话：
*"Agent 由 `team_id + agent_id` 唯一确定"*，v3 数据面**强制要求** `team_id` / `agent_id` / `user_id`
（`session_id` 可选），"新 Adapter … 始终提供 Team、Agent、User 隔离维度"。既然 agent 数 ≈ 项目数（70），
**agent 维度就等价于项目维度**，正好落在 §8.3 的三层里的"项目层"。

**缺口只有一个，但它恰好是关键那一层：L3 —— ✅ v2.3 已钉住（附录 C U17）。**

| 层 | Proxy 文档的称呼 | MemoryCore 的称呼 | 项目专属 agent 能不能隔离 |
|---|---|---|---|
| L0 | 短期对话记忆（每轮写回） | L0 对话 | ✅ 会话/任务级 |
| L1 | 会话级关键记忆（工具按需召回） | L1 原子记忆 | ✅ 会话级 |
| L2 | **Agent Profile**（注入 system prompt） | L2 场景记忆 | ✅ **按 agent 天然隔离 = 按项目隔离**（实测） |
| L3 | **Team / Global 记忆**（注入 system prompt） | L3 核心画像 | ✅ **也是 `(team, agent)` 级**（实测），且注入**可整体关掉** |

**v2.3 实测结论**（附录 C U17）：两份文档叫法不一致，只是因为**说的不是同一件事**——Proxy 的"L3 Team/Global"是**内容语义**
（`MEMORY_PROMPT_MODE=code` 下 L3 的生成提示词是 `TEAM_MEMORY_SYSTEM_PROMPT = "# Team Operating Doctrine Architect"`），
MemoryCore 的"L3 核心画像"是**内容形态**；**存储归属由 `profile-scope.ts` 决定：`team:${teamId}|agent:${agentId}`**，
且源码注释明说 L2/L3 **有意忽略** user/session/task 维度。实机（同一 team 两个 agent）：alpha 读不到 beta 的 L3、beta 也读不到 alpha 的；
L2（真实 LLM 抽取出的 scenario 文件）同样互不可见；换个 user 去读 alpha 的 L3 **仍然可见**（L3 不按 user 隔离）。
`injectL2L3`（`MemoryProxy/src/config.ts`）**默认全关**，`PROXY_FULL_STACK=1` 才会打开 ⇒ **"可关"也成立**。

⚠️ **但串味的真实形态换了**：缺失的维度会被**静默兜底成 `agent:default`** ——
"忘配 `x-agent-id`"不是"拿到 team 级 L3"，而是**所有 agent 的记忆全落进同一个 `agent:default` 桶**（§7 R6 因此更值得防）。

也就是说：**L2、L3 都按 `(team, agent)` 隔离，项目专属 agent 撑得住全部四层**。

- 于是**不必**再把 L3 当"受治理的 workspace 层"：它天然按 (team, agent) 隔离；要 workspace 级复利，仍走 §8.3 的 10%（那是架构 C 的 workspace 层）。
- 若某个 agent 不该看到 L3 画像，**直接关掉 `injectL2L3`**（总开关）即可。

**三处必须先钉住的事**

1. **Multica 不强制 agent ↔ project 绑定。** `agent` 表只有 `workspace_id`，**全库没有任何 agent→project 的外键**；
   `issue` 的受派人又是泛型的（`assignee_type + assignee_id`）——**任何 agent 都能被派到任何项目的 issue**。
   所以"项目专属"是**约定，不是约束**：需要写进 agent 的 instructions（"你只服务 X 项目，遇到其它项目要拒绝并说明"），
   并在 squad / autopilot / quick-create 这些会自动派活的地方确认路由规则。
2. **`multica agent copy` 不会复制 `custom_env`。** 源码原话：`--custom-env` *"never copied from the source"*
   （`mcp_config` 同理）；它复制 name / description / instructions / avatar / custom_args / max_concurrent_tokens / runtime_config / skills。
   而 TAM 的绑定恰恰靠 `custom_env` 里的 `x-team-id` / `x-agent-id`（§4.4）——
   于是**复制出来的 agent 会丢掉记忆身份**；更糟的是缺 header 时代理会**回落到交互式表单**，而 Multica 弹不出来 →
   **任务卡在首轮**（§7 R6）。70 个项目扩编时这是最容易踩的坑。⚠️ **v2.3 补充**（U17）：即使表单那步被绕过，
   缺 `x-agent-id` 也会在 MemoryCore 侧**静默兜底成 `agent:default`** —— 于是 70 个项目的记忆**全落进同一个桶**，
   而不是报错（这比"卡住"更隐蔽）。
3. **扩容成本要算进运维**：每新增一个项目就要（a）建 agent、（b）注入两个 header、（c）如果要 Lark 入口再装一个应用（8.2-⑥）。
   全都可以脚本化（`PATCH /api/agents/{id}` + `agent_mcp`/`channel` 接口），但**必须有脚本**，否则一定会漏配。

**和"架构 C 的 `(workspace, project)` 硬过滤"比一比**

| | 项目专属 agent（TAM 路线） | `(workspace, project)` 表（架构 C） |
|---|---|---|
| 改代码 | **0** | 150–250 行 |
| 隔离强度 | 依赖约定 + 每个 agent 的配置都正确 | **数据库层硬过滤**，配错也绕不过去 |
| 跨项目复利 | 无原生维度（L3 也是 `(team, agent)` 级，见 8.3.1） | workspace 层天然有 |
| 新项目扩容 | 每项目 2–3 步手工/脚本 | 无（`project_id` 自动） |
| 失败模式 | 漏配/复制丢失 → **任务卡首轮**，且 Multica 侧看不出原因 | 缺字段即"没有记忆"，静默降级 |

**结论**：**走 TAM 就用 agent 当隔离键——这是对的**（比在 TAM 里硬找"项目"维度现实得多）；
但**不要**把这个技巧带到架构 C 上：那里用 `project_id` 硬过滤更稳，还能顺带拿到跨项目复利。
两者可以并存（先用 agent 隔离跑起来，C 落地后把隔离键换成 project）。

### 8.4 别给 70 个 agent 各挂一个记忆 MCP（token 账）

```
额外 token / 月 ≈ 70 agents × 任务数/天/agent × 30 × ( 索引注入 tokens + MCP 工具 schema tokens? )
```

- **索引注入**（架构 C）：每任务一次性、有硬预算、可日志化（§0.3 约束 1 的"只注入索引"）。
- **MCP 工具 schema**：**每轮常驻**，且"模型自决"不保证真的调用（§2.1）——成本确定，收益不确定。
- 结论：**默认走 C**；只有确实需要"模型主动查全文"的少数 agent 才绑 MCP；其余用 `multica memory get`（CLI，0 常驻成本）。
- **先量基线再上功能**：用 ⑦ 的两张表取 1–2 周的 `input_tokens` / `cache_read_tokens` / `task_count`，
  按 (agent, project) 分组，这就是 §6 线上 A/B 的护栏数据源——**不用新建任何埋点**。

### 8.5 Android 特有的两条红线

1. **密钥绝不入记忆**（比通用场景更尖锐）：keystore 口令、签名文件路径、Play Console 的 service account JSON、
   内网 maven 凭据、CI 变量。落库 / 落 git 之前**必须过 secret 扫描**——走 GitLab MR 路线天然有 pre-commit + MR 两道门，
   这是选 GitLab 当真相层的一个附加好处；同时按 §7 R1 在应用侧剥离 + 防路径穿越。
   （顺带：`server/pkg/redact` 已经在做日志脱敏，但**不要**把它当作记忆的防线。）
2. **flavor / 环境矩阵是收益最高的一类记忆**：minSdk、AGP 与 Gradle 版本约束、flavor×环境组合、各渠道包差异、
   已知 flaky 测试与真机差异——判据仍是 §5 P0 那条："下次还会再讲一遍"的才算。

### 8.6 「编码时使用」改变的那一条：注入点必须在**每一轮**，不能只在任务开始

前几节默认"每次任务注入一次索引"。**如果记忆主要在编码时用，这条默认就不成立了**——一个编码任务不是一问一答，
而是几十到几百轮的读文件 / 改文件 / 跑测试，中间还会 auto-compact。§2.1 里那句"压缩后必须重注入"从脚注变成了**主线要求**：
只在 claim 时注入一次，第一次压缩后记忆就没了。

**已核实的三个候选注入点**（事实见 8.2-⑧⑨⑩⑪）：

| 注入点 | 每轮？ | 覆盖 provider | 要不要改代码 | 判断 |
|---|---|---|---|---|
| **A. claim 时的 runtime brief** | ❌ 只一次 | 全部 | 要（§4.3） | 放**任务级**上下文（这个 issue 属于哪个项目、记忆索引），**不足以支撑编码时召回** |
| **B. 每轮用户消息 `perTurnContextBlocks()`** | ✅ 每轮 | 全部 | 要，但很小（照 `BuildConnectedAppsBlock` 写一个 `BuildMemoryIndexBlock` 挂进 `prompt.go:63` 的列表） | ⭐ **编码时使用的正解**：与压缩无关（每轮重发）、不破坏 prompt-cache 前缀、注入即可日志化 |
| **C. 传输层代理（TAM）** | ✅ 每轮 | claude 系起步 | **0 代码** | 唯一"零代码 + 每轮都在"的路；代价是流量多一跳、数据与契约不在你手里，且**没有按项目隔离的原生维度**（见下） |
| D. per-provider hook | ✅ | 6 个 | **claude 侧不用改代码**（用户级插件 hooks 实测会触发，U1）；Reasonix 走 `AGENTS.md`（U2）；其余 provider 仍要适配或没有端点 | **v2.3 修订：claude 上已可用**（粒度 per-host/per-user，按 agent 隔离需 `CLAUDE_CONFIG_DIR`）；其他 provider 仍排除或要写适配 |

**结论**：编码时使用的推荐 = **B（默认）**，需要"今天就自动生效、愿意多一跳"时用 **C**。
注意 ⑨ 与 ⑩ 堵死的是**Multica 自己那套插件引擎**（没有轮次触发器）与"自塞 `--settings`"这条路；
**但 claude 自己的用户级插件 hooks 不受影响**——2026-09-17 隔离实测确认会触发（附录 C U1），Reasonix 也有 11 个 hook 事件 +
`AGENTS.md` 通道（U2）。所以"装个插件就自动召回"在 **claude / reasonix 上成立，但要按 agent 隔离**（`CLAUDE_CONFIG_DIR`）；
要**覆盖全部 provider**，仍然只有 B 或 C。

**每轮注入的内容形状（与"任务级注入"很不一样）**

| | 每轮注入（编码时） | 任务级注入 |
|---|---|---|
| 粒度 | **索引级**（名字 + 一行 + `applies_to`），**硬预算**（原建议 ≤300 tokens/轮；**v2.3 实测 30 条库约 570–900 tokens**，预算要按库大小缩放，见 U16） | 可以稍大 |
| 排序依据 | **本轮文本**（评论文本 / 交办说明 / 聊天消息，daemon 每轮都拿得到）+ issue 上下文；**不需要等工具调用** | 项目 + 重要性 |
| 详细内容 | 一律**不注入**，走 pull（skill / `multica memory get`）——与 TAM 自己的分法一致（L2/L3 注入、L0/L1 工具化，理由就是"别破坏 KV cache"） | 可带摘要 |
| 条目字段 | 必须带 **`applies_to`**（路径 / 模块 glob，如 `app/build.gradle.kts`、`**/proguard-rules.pro`）与"何时有用" | 可省 |

**为什么 `applies_to` 在 Android 上特别重要**：编码时的语境是"我正在改**这个文件**"，不是"我在做这个项目"。
70 个 App 的 gradle / manifest / proguard 文件同名、内容相近，只有绑定到路径与模块，
才能在**改到那个文件的那一刻**召回"这个坑踩过"——纯语义相似度会召回到另一个 App 的同类文件。

**内容的价值排序（编码时）**：程序性（**怎么做**：这个项目的 build/flavor/签名流程）> 语义（**是什么**：模块职责、依赖方向）
> 情景（上次发生了什么）。前两类才是编码时高频命中的；情景类记忆更适合放在任务开始时注入。

**Android 编码时的高频召回点**（建议作为 §6 评测的取样集）：Gradle 版本目录（`libs.versions.toml`）、AGP/Gradle/compileSdk 约束、
flavor × 环境的构建命令、签名与 keystore 位置（**只记位置与流程，绝不记口令**）、ProGuard/R8 规则、
Room/Retrofit/Hilt 之类的既有约定、flaky 测试与真机差异、CI 上跑单测的正确命令。

### 8.7 在这套条件下的落地顺序

| 阶段 | 做什么（相对 §5 的变化已标出） |
|---|---|
| **P0 · 今天** | ① **先测通私有 GitLab 的 checkout → 推分支 → 开 MR**（8.2 前置项，不通就别往下走）；② 在 GitLab 建 `multica-memory`（一个仓库 70 个目录，或每项目一仓）+ `CODEOWNERS`；③ 写清"何时读、何时写、写什么"的 skill |
| **P1 · 1 天** | 起 **ReMe**（§3.1.3）在 **1 个 Android 项目**上做 POC；**把 70 个 agent 的绑定脚本化但不绑 MCP** |
| **P2 · 1–2 周** | 架构 C **两个注入点一起做**：① claim 时的 `memory_index`；② **每轮用户消息的 `BuildMemoryIndexBlock`**（编码时使用的关键，§8.6）+ `multica memory get` CLI；**project 硬过滤 + `applies_to`**；A/B 用 ⑦ 的表读护栏 |
| **P3 · 之后** | 写入走 **GitLab MR + CODEOWNERS 审查门**（0 新代码，Multica 已会关联 issue）；Lark 里推审查提醒（复用现有集成） |

### 8.8 这套条件下被否掉的选项

| 选项 | 为什么否 |
|---|---|
| 给 70 个 agent 各挂记忆 MCP | 每轮常驻的工具 schema × 70 agent，成本确定、收益不确定（8.4） |
| **"装个记忆插件就自动召回"** | **v2.3 修订为"部分成立"**：Multica 自己的插件引擎确实没有轮次触发器、`--settings` 也被 daemon 独占（8.2-⑨⑩）；但 **claude 的用户级插件 hooks 实测会触发**（附录 C U1）⇒ 在 claude 上今天就能零代码自动召回，**代价是 per-host/per-user 粒度**（70 个 agent 要隔离得逐个 agent 设 `CLAUDE_CONFIG_DIR`）；要覆盖全部 provider 仍须改代码或走代理（8.6） |
| **只在任务开始时注入一次** | 编码任务会 auto-compact，第一次压缩后记忆即失效（8.6） |
| 把记忆全文每轮注入 | 每一轮都是成本，且会挤掉代码本身；只注入索引 + 按需取全文（8.6） |
| 每人一套私有记忆 | 20 人 × 70 项目，个人记忆边际价值最低；复利在项目层（8.3） |
| 云厂商托管记忆（§3.2 D 全部） | Lark 国际版与 GitLab 都在内网，托管方案依旧出局；且 70 个项目的记忆是团队资产 |
| 每个 agent 一个独立记忆库 | 等于 70 个孤岛，跨项目经验无法复用（8.3） |
| 等 agent 自己攒出跨项目知识 | 冷启动太慢；人灌 + 审查门更快也更安全（8.3） |

### 8.9 本节未验证的部分

- **70 个 agent 的实测规模**：本机是 29/2/5 的开发实例，生产规模**没有实测**（附录 C U12）。
- **GitLab MR 审查门与记忆仓库的端到端**：Multica 已支持 MR 关联/关闭 issue，但**"记忆仓库的 MR 走同一套"没有实机验证**（U13）。
- **ReMe 在 Android/Java/Kotlin 语料上的检索质量**：未测（U6 已记）。
- **Lark 国际版的落地细节**（70 个应用 vs 少量应用、`accounts.larksuite.com` 的租户配置）**未实测**（U14）。
- **每轮注入（`BuildMemoryIndexBlock`）的成本与收益**：通道已在代码层确认（8.2-⑧），**A/B 没有做**。v2.3 取到了两端常量：
  一块 30 条规模的索引实测 **≈570–900 tokens**（U16），本机 claude 任务 `task_usage` 实测 **avg input 69k / median 5k /
  cache_read 416k tokens 每任务** —— 而 A/B 的度量通道**确实存在**（`task_usage` 按 task 记 input/cache_read/output），
  只差实现后开跑（要先改代码，需确认）。
- **TAM 的每轮注入在 Multica 任务里的实际表现**：✅ **注入本身已实测**（附录 C U3）——上游收到的 system prompt 里
  **`<tdai_profile_memory>` 整块 ≈350 tokens/请求**（`<l3_core_memory>` 全文 ≈250 + `<l2_scene_index>` 只给 path+summary ≈35），
  两轮之间内容稳定（`cacheStrategy: session_init`）、**按 agent 隔离无串味**、`injectL2L3: false` 能干净关掉。
  与"每轮 ≈570–900 tokens 的索引块"比**更省**（L2 只给索引），代价是 L3 全文注入且注入逻辑不在自己手里。
  **仍未测的是"真实 Multica 任务里的跨任务召回"**（要把测试 agent 的 base URL/header 指过去）。
- **"编码时召回"的效果本身**：本轮给的是设计（`applies_to` + 每轮索引注入），**没有做过 A/B**（U18）。

---

## 附录 A 来源

**调研底稿**（**七份**，均未纳入交付物）：`<scratchpad>/session-2026-09-15-19-19-10-af3b2e06/research/`
A = 各 harness 注入点 · B = 后端选型（16 个）· C = 最佳实践与评测 ·
**D1 = 平台型记忆系统（14 个）· D2 = 编码 agent 专用记忆工具（18 个）· D3 = 研究型/图/时序与榜单 · D4 = 中文生态与云厂商**。

**Multica 上游**：#838 · #1211（含维护者立场）· #1417（closed）· #1767 · #3437 · #6658 · MUL-5969

**宿主文档**

- Claude Code：[hooks](https://docs.claude.com/en/docs/claude-code/hooks) · [memory](https://docs.claude.com/en/docs/claude-code/memory) · [context-window](https://docs.claude.com/en/docs/claude-code/context-window)
- Codex：[hooks](https://developers.openai.com/codex/hooks) · [memories](https://developers.openai.com/codex/memories)
- [Cursor hooks](https://cursor.com/docs/hooks) · [OpenCode plugins](https://opencode.ai/docs/plugins/) · [OpenClaw hooks](https://docs.openclaw.ai/plugins/hooks/prompt-and-session) · [Hermes hooks](https://hermes-agent.nousresearch.com/docs/user-guide/features/hooks) · [DSH lifecycle](https://github.com/deepseek-ai/deepseek-harness/blob/master/docs/agent-lifecycle.md)

**记忆系统**

- [OpenViking 插件标准](https://docs.openviking.ai/en/agent-integrations/18-plugin-development) · [Claude memory tool](https://platform.claude.com/docs/en/agents-and-tools/tool-use/memory-tool) · [Letta MemFS](https://docs.letta.com/concepts/memfs) · [Graphiti](https://github.com/getzep/graphiti) · [mem0](https://github.com/mem0ai/mem0) · [LangGraph memory](https://docs.langchain.com/oss/python/concepts/memory)
- 腾讯 TAM：[仓库](https://github.com/TencentCloud/TencentDB-Agent-Memory)（分支 `feat/server_team`）· [MemoryProxy 说明](https://github.com/TencentCloud/TencentDB-Agent-Memory/blob/feat/server_team/MemoryProxy/README_CN.md) · [客户端接入与 header 预选](https://github.com/TencentCloud/TencentDB-Agent-Memory/blob/feat/server_team/agents/README.md) · [三件套部署](https://github.com/TencentCloud/TencentDB-Agent-Memory/blob/feat/server_team/deploy/global-images/README.md) · [ROADMAP](https://github.com/TencentCloud/TencentDB-Agent-Memory/blob/feat/server_team/ROADMAP_CN.md)

**本轮新增的来源**（★/fork 一律 2026-09-17；license 取自仓库根 LICENSE 原文）

- 候选：[ReMe](https://github.com/agentscope-ai/ReMe) · [memU](https://github.com/NevaMind-AI/memU) · [honcho](https://github.com/plastic-labs/honcho)（+ 6 个插件仓 `plastic-labs/{claude,codex,opencode,cursor,openclaw,dsh}-honcho`）· [Neo4j Agent Memory](https://github.com/neo4j-labs/agent-memory) · [MemOS](https://github.com/MemTensor/MemOS) · [supermemory](https://github.com/supermemoryai/supermemory) · [ACE](https://github.com/kayba-ai/agentic-context-engine) · [engram](https://github.com/Gentleman-Programming/engram) · [memorix](https://github.com/AVIDS2/memorix) · [deja-vu](https://github.com/vshulcz/deja-vu)（`LycheeMem`、`token-savior`、`byterover`、`SimpleMem`、`A-mem`、`m_flow` 等**已按 v2.2 剔除**，其链接与最后提交见 §3.2.1）
- 高星盲区：[MemPalace](https://github.com/MemPalace/mempalace) · [Hindsight](https://github.com/vectorize-io/hindsight)
- 榜单与交叉校验：[Awesome-AI-Memory](https://github.com/IAAR-Shanghai/Awesome-AI-Memory)（82 个开源系统，含 5 个死链）· [Agent_Memory_Techniques](https://github.com/NirDiamant/Agent_Memory_Techniques)
- MCP 记忆 server 全景：官方 registry API `https://registry.modelcontextprotocol.io/v0/servers?search=memory`（41 项）·
  [fleet-memory](https://github.com/Holetron-lab/fleet-memory) · [bettermemory](https://github.com/0Mattias/bettermemory)
- 云厂商官方文档（用于"内网自托管出局"那一句）：[AWS AgentCore Memory](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/memory.html) ·
  [Google Agent Platform Memory Bank](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank) ·
  [Microsoft Foundry Agent Service memory](https://learn.microsoft.com/en-us/azure/ai-foundry/agents/concepts/what-is-memory) ·
  [OpenAI conversation state](https://platform.openai.com/docs/guides/conversation-state) · [Redis Agent Memory](https://redis.io/agent-memory/) ·
  [Redis agent-memory-server（已降级 V0）](https://github.com/redis/agent-memory-server) · [Supermemory self-hosting](https://supermemory.ai/docs/self-hosting/overview) ·
  [LangMem](https://github.com/langchain-ai/langmem) · [LlamaIndex memory](https://developers.llamaindex.ai/python/framework/module_guides/deploying/agents/memory/)

**检索工具的一手经验**（本轮通用搜索引擎不可用，改用以下三条，可复现）

- 官方 MCP registry API：`https://registry.modelcontextprotocol.io/v0/servers?search=<q>&limit=100`（JSON；search 长词组如 `long-term memory` 会超时，`memory` 可用）
- GitHub API（`search_repositories` / `get_file_contents` / `list_commits` / `list_releases`）+ `raw.githubusercontent.com` 直读 LICENSE / README（无需代理；`git clone` 才需要 `HTTPS_PROXY`）
- ⚠️ 并行调研时会**打到 GitHub 限流**（`rate limit exceeded, retry after 12s`）——多路并行时应错开 API 使用；**v2.2 实测：`api.github.com` 直连与经代理都报
`rate limit exceeded`（未认证 IP 共享额度），所以「最后提交」一律用 GitHub MCP 的 `list_commits`（`perPage=1`）读。**

**论文**

- 记忆系统：Mem0 arXiv:2504.19413 · Zep arXiv:2501.13956 · Generative Agents arXiv:2304.03442 · CoALA arXiv:2309.02427 · Voyager arXiv:2305.16291 · MemoryBank arXiv:2305.10250
- 本轮新增：Hindsight arXiv:**2512.12818** · SAGE arXiv:**2409.00872**（无开源实现）· A-MEM arXiv:2502.12110 · HippoRAG NeurIPS'24 arXiv:2405.14831（v2: 2502.14802）· LightRAG arXiv:2410.05779 · GraphRAG arXiv:2404.16130 · MemInsight arXiv:2503.21760 · MemOS arXiv:2507.03724 / 2505.22101
- 上下文利用：Lost-in-the-Middle arXiv:2307.03172 · [Chroma Context Rot](https://research.trychroma.com/context-rot)
- 安全：MINJA arXiv:2503.03704 · AgentPoison arXiv:2407.12784
- 评测：LoCoMo arXiv:2402.17753 · LongMemEval arXiv:2410.10813 · tau2-bench arXiv:2506.07982

---

## 附录 B 证据索引

正文里每条论断的出处，便于复核（行号以 2026-09-17 的工作区为准）：

| 位置 | 证明了什么 |
|---|---|
| `agent_task_queue.session_id / work_dir`（migration 020） | 会话续接的持久化字段 |
| `server/internal/handler/daemon.go:2190` / `:2817-2828` | 续接失败时置 `PriorSessionResumeUnavailable`（显式告警） |
| `server/internal/daemon/execenv/runtime_config_sections.go:187` · `:998-999` | `## Workspace Context` 的渲染与 section 挂载点 |
| `server/internal/handler/agent.go:380-385` | claim 响应的 `workspace_context` |
| `server/internal/daemon/types.go:86-90` | daemon 侧的字段镜像 |
| `server/internal/daemon/execenv/execenv.go:199-203` | `TaskContextForEnv.WorkspaceContext` |
| `server/internal/daemon/execenv/context.go:268-298` | `project_resource` → `.multica/project/resources.json` |
| `server/internal/daemon/execenv/hermes_memory.go` + `daemon/config.go:79-81` | 仅 Hermes 有原生记忆；TTL Codex session 14d / Hermes memory 90d |
| `server/internal/daemon/execenv/codex_memory.go` | Codex auto-memory 默认关闭，`MULTICA_CODEX_MEMORY=1` 才开 |
| `server/internal/daemon/execenv/hermes_memory.go:58` | 唯一的 `memory_scope` 注释 → MUL-5969 |
| `server/internal/daemon/claude_plugins.go` + `runtime_mcp.go:359/515` | 只采纳插件的 `skills` / `mcpServers`，不采纳 hooks |
| `server/internal/daemon/runtime_mcp.go:262-305` | 运行时本地 MCP 继承只支持 6 个 provider（`default` 直接返回未支持） |
| `server/pkg/agent/claude.go:710-713` | `claudeBlockedArgs` 参数白名单 |
| `server/pkg/agent/claude.go:741` | `--disallowedTools AskUserQuestion` |
| `server/pkg/agent/claude.go:891-933` | `buildEnv` = `os.Environ()` + extra，**无名字白名单** |
| 本机 Postgres（2026-09-17 实测） | 无任何 memory 类表；`agent.custom_env` 已有把 `ANTHROPIC_BASE_URL` 指向 `127.0.0.1:8787` 的先例 |
| claude CLI `2.1.274` 二进制 | 支持 `ANTHROPIC_CUSTOM_HEADERS`（格式 `Header-Name: value`） |
| `http://127.0.0.1:8787/health` | headroom-proxy 0.37.0 在跑，上游 `https://api.deepseek.com/anthropic` |
| TAM `MemoryProxy/src/session/preset.ts` | `canRegister = !!agentId && !hadMismatch`（task 可选；未知 task 不拦注册） |
| TAM `deploy/global-images/start-proxy.sh` | 生成的 config 含 `headerAutoSelect.enabled: true`；auth/sessionInit/tdai 默认全 0 |
| TAM `deploy/dockerhub/README.md` + `deploy/global-images/README.md` | 三镜像名与端口、公开可拉、两组 LLM 参数、默认凭据 `local`/`admin` |
| `MemPalace/mempalace` 的 LICENSE + README（2026-09-17 直读） | ★59,114 / MIT / 默认分支 `develop` / "verbatim storage, pluggable backend, 96.6% R@5 on LongMemEval" / 默认 ChromaDB / 有 impostor-site 警告 |
| `vectorize-io/hindsight` 的 LICENSE + README | ★23,813 / MIT / 基准由 Virginia Tech Sanghani Center 与 Washington Post 独立复现 / arXiv:2512.12818 |
| Multica PR **#677**（`gh pr view 677`） | 上游给 Hindsight 提过记忆集成：**是 PR、2026-04-10 开、2026-08-26 因"与 main 冲突 + 积压清理"关闭**，原话 *"not a judgment on the idea itself"* —— §3.1.2/§1.3 的闭环**解除** |
| `agentscope-ai/ReMe`、`NevaMind-AI/memU`、`plastic-labs/honcho`、`neo4j-labs/agent-memory` 的仓库与 LICENSE 原文 | §3.1.3 四个候选的 ★ / license / 形态 / 存储 |
| `org:plastic-labs` 下的 6 个插件仓（`{claude,codex,opencode,cursor,openclaw,dsh}-honcho`） | honcho 的 harness 覆盖是**真插件**（各自独立仓库、仍在更新，`dsh-honcho` 建于 2026-08-31） |
| `MemTensor/MemOS` 的 `src/memos/vec_dbs/` 与 `graph_dbs/` | 向量后端只有 Qdrant/Milvus，PG 只能当图后端 → 用不上我们现有 PG |
| 官方 MCP registry `…/v0/servers?search=memory`（41 项） | 记忆类 MCP server 绝大多数是托管 SaaS |
| `redis/agent-memory-server` README + `V0/` 目录 | 开源服务器已被官方降级、不再维护（"is not the supported production path"） |
| `IAAR-Shanghai/Awesome-AI-Memory` README（1.0 MB） | 82 个开源系统 + 5 个 404 死链；70 个不在本报告早期版本里 |
| `MemoriLabs/Memori` 的 LICENSE + `docs/memori-byodb/databases/postgres.mdx` + HEAD 提交 | Apache-2.0；PG 是推荐生产库；OSS 无服务端、HEAD 是 README 营销改动 |
| `Cyber`（byterover-cli）的 LICENSE + `pushed_at` | ELv2 非开源；真实停更 2026-06-25（`updated_at` 会骗人） |
| `server/internal/integrations/vcs/gitlab.go` + `server/migrations/216_vcs_integration.up.sql` | 自托管 GitLab 是一等公民：`vcs_connection(provider, instance_url)`、MR/CI 镜像、凭据加密（§8.2-①） |
| `apps/docs/.../vcs-integration.mdx` §"Agents opening pull requests" | token 只要 `read_api`；**clone/push 用 daemon 宿主的 git 凭证**（§8.2-②） |
| `server/internal/handler/project_resource.go:79-106` | `github_repo` 校验的是通用 git URL → GitLab 地址可直接当项目资源（§8.2-③） |
| `server/internal/integrations/lark/registration.go:51-52` · `http_client.go:45-46` | `accounts.larksuite.com` / `open.larksuite.com` 是内置区域，Lark 国际版无需改代码（§8.2-④） |
| `server/migrations/315_workspace_mcp_server.up.sql` + `workspace_mcp.go:25-57` + `cmd_agent_mcp.go` | 工作区 MCP 必须逐 agent 绑定、无批量开关、API 可脚本化（§8.2-⑤、§8.4） |
| `channel_installation`（本机库实测：`agent_id NOT NULL`） | Lark 安装是 (workspace, agent) 粒度（§8.2-⑥） |
| `server/migrations/101_task_usage_hourly_schema.up.sql` · `073_task_usage_daily_rollup.up.sql` | 用量可按 (workspace, runtime, agent, project, provider, model) 读，现成的 A/B 护栏（§8.2-⑦、§8.4） |
| 本机库 `count(*)`（2026-09-17） | 29 agents / 2 workspaces / 5 projects / 156 tasks —— 开发实例，非 70-agent 生产规模（§8.2 注） |
| `server/internal/daemon/prompt.go:63-72`（`perTurnContextBlocks`）+ `runtime_config_sections.go:158-215` | **存在"每轮用户消息"注入通道**，已承载 Shared working directory / 合并冲突 / 会话续接 / Task Initiator / Connected Apps；注释说明故意不放 brief 以保 prompt-cache 前缀（§8.2-⑧、§8.6） |
| `server/migrations/362_plugin_hook_engine.up.sql:23` · `399_plugin_hook_schedule.up.sql:33-35` | 插件 hook 触发只有 `ui/manual/event/agent/schedule`——**没有轮次/提示词触发器**（§8.2-⑨） |
| `server/pkg/agent/claude.go:771-783` | `--settings` 在 `blockedArgs` 里被 daemon 独占（技能开关）→ 自塞 hook 配置不可行（§8.2-⑩） |
| TAM `MemoryProxy/README_CN.md` §记忆层与注入策略 | L0 每轮写回 / L1 工具化 / **L2 与 L3 直接注入 system prompt**、`<session_context>` **每轮追加**；工具化的理由是"避免破坏上游 KV cache"（§8.2-⑪、§8.6） |
| `server/internal/daemon/plugin_hook_mcp.go` | 插件 hook 以**本地 MCP 工具**形式呈现给模型（模型自决），不是注入 |
| TAM `MemoryCore/README_CN.md`（§核心能力 · §存储与隔离） | *"Agent 由 `team_id + agent_id` 唯一确定"*；v3 数据面**强制** team/agent/user 三个隔离维度 ⇒ **agent 是一等隔离维度**，可当项目键（§8.3.1） |
| `server/cmd/multica/cmd_agent_copy.go:74-78` | `agent copy` 的 `--custom-env` / `--mcp-config` 均注明 *"never copied from the source"* ⇒ **复制 agent 会丢 TAM 绑定 header**（§8.3.1-2） |
| `agent` 表结构（本机库实测：只有 `workspace_id`，无 project 列）+ `001_init.up.sql:62` `issue.assignee_id` 泛型 | **没有 agent→project 绑定**，"项目专属 agent"是约定不是约束（§8.3.1-1） |
| `agentscope-ai/ReMe` 的 README（2026-09-17 直读，26.7 KB） | Apache-2.0；"Memory as File, File as Memory"（Markdown+frontmatter+wikilinks 为真源）；**BM25 默认、embedding 默认关**；HTTP `127.0.0.1:2333` + CLI + MCP(streamable HTTP) + Python + Studio；"One memory workspace works across agents"（§3.1.3 的推荐依据） |
| GitHub `list_commits`（`perPage=1`，2026-09-17 逐个读）对 §3.2.1 里每个仓库的最后提交 | 规则①的判定依据（另：`updated_at` 随 star 刷新、不代表推送时间——byterover 就是被它骗过的例子） |
| 仓库根目录清单（`get_file_contents`，2026-09-17）：MemoryBear / MIRIX / SkillClaw / omnigraph / memmy-agent | 规则②的判定依据：`identity-service`+`gateway-service`+`mem-knowledge`；`Dockerfile`+`docker-compose.yml`；`evolve_server`+`requirements-server.txt`；Rust workspace + `deploy/`+`Dockerfile`；`Memory/`+`Migrations/` —— 都是"能自己跑"的形态 |
| `claude` 2.1.274 隔离实验（2026-09-17，临时 plugin dir + 隔离 HOME/`CLAUDE_CONFIG_DIR`，未动 `~/.claude`） | **用户级 Claude Code 插件的 hooks 在 daemon 同款参数下会触发**（marker 落盘；debug：`Registered 75 hooks from 31 plugins`）⇒ 架构 D 在 claude 上零代码可用（附录 C U1） |
| `reasonix` 二进制内嵌串 + 上游 `esengine/DeepSeek-Reasonix` 的 `internal/hook/hook.go`、`control/turn_orchestrator.go` | Reasonix 有 **11 个 hook 事件**；`UserPromptSubmit` 的 stdout **被丢弃**（不可注入）⇒ 每轮注入须走 `AGENTS.md`（附录 C U2） |
| 本机实跑 `agentscope-ai/ReMe` 0.4.1.12（`127.0.0.1:2333`） | 33 个 HTTP 端点 + MCP streamable HTTP（32 工具）；默认分词器 `regex` 逐字（中文单字 OR）；并发写 25+25 无丢写；30 条索引块 ≈896 tokens（附录 C U5/U6/U16） |
| `memu-cli` 0.10.0 / 0.11.0b3（PyPI 2026-08-05 / 08-21）本机实跑 | memU 的发布线在 **PyPI**（不是 GitHub tag）；**embedding 是硬依赖**；`AGENTS.md` managed block 幂等可回滚，但 Multica 的 per-task workdir 使打 `./AGENTS.md` 的适配器写的块随任务目录消失（附录 C U5/U7） |
| HF 数据集卡 `agentscope-ai/ReMe_longmemeval_clean_s_v2` | ReMe 的 LongMemEval 89.4% 跑在**自家修订过 ground truth** 的 clean-S 上（*"replacing the answer and supporting-session ground truth"*）（附录 C U6/U11） |
| `MemPalace` 3.10.0 本机实跑；`ghcr.io/vectorize-io/hindsight` 容器 | MemPalace：零 key 读写通过、中文只走向量通道、impostor 是 `mempalace.tech`；Hindsight：容器能起但 retain/recall 必须有可达 LLM 端点（附录 C U9） |
| `git ls-remote` 两个团队 GitLab 私有仓库（宿主机 keychain 凭证，`http://192.168.5.254/…`） | daemon 宿主的 clone 权限**可用**（exit 0）；团队仓库里已有 `agent/<name>/<taskKey>` 分支（Multica 的分支命名）⇒ clone/push 链路已在跑（附录 C U15） |
| `channel_installation` 的 `UNIQUE INDEX (channel_type, config->>'app_id')`（migration 124:100-101） | **一个 Lark 应用只能绑一个 (workspace, agent)** ⇒ 70 个 agent 要 70 个应用（附录 C U14） |
| `server/internal/handler/vcs_webhook.go:210-231` | MR 靠**标题 / 分支 / 正文 closing 关键字里的 issue 标识符**自动关联；接入 provider 即 opt-in、与仓库用途无关 ⇒ 记忆仓库的 MR 走同一套（附录 C U13） |
| `gh` 批量核对 Awesome-AI-Memory 的 85 个仓库（2026-09-17） | 5 个死链；**活跃且本报告未提过的 ★≥300 只有 4 个**（`marciopuga/cog` 376 · `MaxFreedomPollard/Compartment` 581 · `samvallad33/vestige` 624 · `Tura-AI/tura` 637）（附录 C U8） |
| 本机拉起 TAM 三件套（`tamlab-memory-core` / `-hub` / `-proxy`，独立 net + 卷，上游指向 mock，2026-09-17） | **注入实测**：system prompt 里 `<tdai_profile_memory>` = 1,394 / 1,407 chars（≈348 / 352 tokens），拆 L3 全文 ≈250 + L2 索引 ≈35；两轮稳定（`cacheStrategy: session_init`）、按 agent 无串味、`injectL2L3:false` 后 8,332 → 5,505 chars 干净关闭（附录 C U3） |
| TAM `MemoryProxy/src/session/session-key.ts:9-19` + 实机对照 | **缺客户端 session id header（`x-claude-code-session-id`）→ `injectedSkipped = true`，静默跳过注入**；身份 header 只决定"是否回落到表单"（附录 C U3、§7 R6） |
| TAM `MemoryProxy/src/session/registrar.ts`（"本地构建 SessionInfo，不再调用 TMC"） | proxy ↔ memory-hub 那一跳**在当前版本已不存在**（hub 只提供 Panel / Knowledge）——旧说法"sessionInit 的 control plane 在 hub:8125"已过时（附录 C U3） |
| TAM `MemoryCore/src/core/profile/profile-scope.ts:34-41` + 两 agent 实机对照 | **L2 与 L3 的存储归属都是 `(team, agent)` 复合键**（源码里拼成 `team:…` 加 `agent:…`），L3 不按 user 隔离；缺 `x-agent-id` 会被静默兜底成 `agent:default`（附录 C U17、§8.3.1） |
| 聚合检索通道实测（2026-09-30）：HN Algolia API、`lobste.rs/search.json`、`search.marginalia.nu` 均可用；Bing 结果不可信、DDG/Mojeek/Brave/Qwant/Startpage/5 个 SearXNG/Reddit 全部不可用；Firecrawl 因 vault 无 key 不可用 | §3.1.4 的"全网边界"与它列出的 8 个新候选的来源 |
| `gh api repos/{okf-memory/okf-agent-memory, mindmuxai/brain.md, ctxrs/ctx, Avinash-jetwani/jevmem, hyperb1iss/sibyl, salishforge/memforge, samvallad33/vestige, Fr4nZ82/mwe-mcp}`（2026-09-30） | §3.1.4 表里的 ★ / license / `pushed_at` / 语言；`okf-agent-memory` 的 `created_at = 2026-09-05`、contributors 以 `sknr` 为主 |
| 本机实跑 OKF `okf-darwin-arm64` v0.4.4（scratchpad `okfprobe/projA` 与 `projB`，2026-09-30） | `search` / `--filter type=Decision` / `--for-path …（code_refs）` / `validate` 全通；**中文 `构建缓存`✓ `构建`✓ `缓存`✗**（分词前缀不对称，源码 `pkg/okf/search.go`）；**projA 的记忆在 projB 搜不到**（物理隔离）（§3.1.4） |
| 本机实跑 ReMe 0.4.1.12（`127.0.0.1:2333`，2026-09-30，§4.5 的每一条都可复现） | `config/default.yaml` 的 `service.mcp_path=/mcp` + `as_llm.credential` 读 `LLM_API_KEY`/`LLM_BASE_URL`；无 key 启动即 `openai.OpenAIError: Missing credentials`；`POST /mcp` initialize → `serverInfo {"name":"ReMe","version":"3.4.7"}`，不带 session 的 `tools/list` → `-32600 Missing session ID`；`POST /search` 加 `search_filter.path_prefix` 把 3 条命中收窄到 1 条 |
| `scutil --proxy` + `python -c "import urllib.request;print(urllib.request.getproxies())"`（2026-09-30） | macOS 系统代理开着（HTTP/HTTPS/SOCKS → `127.0.0.1:7897`），`getproxies()` **不含例外列表** ⇒ httpx 连 `127.0.0.1:2333` 也被送进代理并得到 502（§4.5 坑 3） |
| `server/internal/handler/workspace_mcp_api.go:47-71` + `pkg/agent/claude.go:57-69`/`:1097` + `internal/daemon/runtime_mcp.go:26-79` | MCP entry 的 transport 判定（`local/stdio/remote/http/streamable-http/sse` → stdio/http/sse）；claude 走 `--mcp-config` + `--strict-mcp-config`；agent 绑定的 MCP 与运行时本地 MCP 在 daemon 内合并（§4.5 第 3 步） |

---

## 附录 C 假设与核实结果（原「未验证假设」；v2.3 已逐条核实）

v2.2 及以前，这里是 **19 条未验证假设**。v2.3（2026-09-17）用「只读代码 + 本机实机 + GitHub API」把每一条都过了一遍，
**全程不改仓库代码、不动生产配置**。状态标记：

- ✅ **已证实**（原假设成立，并补上了实机证据）
- ⚠️ **已修正**（假设成立但细节不对，或结论需要收窄 —— 正文已就地标注「v2.3 修订」）
- ❌ **已证伪**（原结论被推翻）
- ⏳ **仍未知**（卡在"要先改代码"或"要有生产实例"，本轮无法在不改代码的前提下闭环）

合计：**15 条闭环**（其中 **6 条修正/证伪**，见 §0 头部）、**1 条部分核实**（U3：只跑通了链路的一端）、
**3 条仍未知**（U12 / U16 的一半 / U18 —— 都卡在"要先改代码"或"要有生产实例"）。

| # | 原假设 | 状态 | 核实结果（一句话） | 依据 |
|---|---|---|---|---|
| U1 | claude 任务里用户级插件 hooks 是否触发 | ✅ **触发** | daemon 同款参数下探针落盘，`Registered 75 hooks from 31 plugins` ⇒ 架构 D 在 claude 上**零代码可用**（粒度 per-host/per-user） | 本机 claude 2.1.274 隔离实验 |
| U2 | Reasonix 的注入点（无一手文档） | ⚠️ **已修正** | 11 个 hook 事件，但 `UserPromptSubmit` 的 stdout **被丢弃** ⇒ 每轮注入**只能走 `AGENTS.md`** | 二进制内嵌串 + 上游源码 |
| U3 | TAM 在 Multica 任务里的端到端链路 | ⚠️ **部分核实**（链路的**每一跳都实测了**，只差真实 Multica 任务） | 三件套全部拉起、proxy 跳已跑通；**注入确实发生**（`<tdai_profile_memory>` ≈350 tokens/请求，两轮稳定、按 agent 隔离、可干净关掉）；**只差"真实 Multica 任务里的跨任务召回"** | 本机拉起三件套 + mock 上游 |
| U4 | "传输层代理天然覆盖全部 provider" | ❌ **证伪** | `custom_env` 只拦 `MULTICA_*` + 14 个具名键，端点变量的可设性**逐 provider** 不同（reasonix 根本没有、codex 只能走 `custom_args`、cursor 未证实） | `daemon.go` 黑名单 + 各 runtime |
| U5 | §3.1.3 四个候选"一个都没实机跑过" | ⚠️ **已修正** | **ReMe 与 memU 本轮实机跑通**（起服务、写、查）；honcho / Neo4j Agent Memory 仍未实机 | 本机实跑 |
| U6 | ReMe 的能力与自报数字 | ⚠️ **已修正** | 33 个 HTTP 端点 / MCP 32 工具，并发写无丢写；但默认**分词器 regex 逐字**（中文单字 OR）、无 key 起不来、写入→可检索有 4–20s 延迟；89.4% 跑在**自家修订 ground truth** 上 | 本机实跑 + HF 数据集卡 |
| U7 | memU 2.0 的成熟度 | ⚠️ **已修正** | 包名是 **`memu-cli`**（PyPI 0.10.0 正式 / 0.11.0b3 预发布）；**embedding 是硬依赖**；`AGENTS.md` 注入是幂等 managed block，但 **Multica 的 workdir 是 per-task** ⇒ 结构性不兼容 | 本机实跑 |
| U8 | Awesome-AI-Memory 里"是否还有漏网之鱼" | ✅ **已量化** | 85 个仓库批量核过：**5 个死链**；活跃、本报告未提过、★≥300 的只有 **4 个** | `gh api` 批量 |
| U9 | MemPalace 与 Hindsight 的实际可用性 | ✅ **已核实** | MemPalace：零 key 可读写、中英 rank-1 命中、**中文只走向量通道**、impostor 域名 `mempalace.tech`；Hindsight：容器能起，但 retain/recall **必须有可达 LLM 端点**，且 benchmark 已迁出仓库 | 本机实跑 + 容器 |
| U10 | 上游 #677 关闭 Hindsight 的真实原因 | ✅ **已证实** | **不是否决**：是一条 PR，2026-08-26 因"与 main 冲突 + 积压清理"被关，原话 *"not a judgment on the idea itself… rebase and reopen"* | `gh pr view 677` |
| U11 | MemTensor 工商主体 / benchmark / DSH 插件 | ⚠️ **已核实（多为否定）** | benchmark harness 可跑但 ground truth 是自家修订版；**第三方独立复现查不到**；MemTensor 只有显示名 `OpenMem`、无备案号/信用代码；两个 DSH 插件**都真装成**（隔离 `DSH_HOME`），运行期未测 | 仓内原文 + npm 实装 |
| U12 | 生产规模（70/70/20）下的表现 | ⏳ **仍未知** | 需生产实例；本机是 29 agents / 2 workspaces / 5 projects / 156 tasks 的开发库 | 查库 |
| U13 | 记忆写入走 GitLab MR 审查门的端到端 | ✅ **机制已确认**（端到端仍需跑一次） | `vcs_webhook.go:210-231`：MR 靠标题/分支/正文的 closing 关键字关联 issue，接入 provider 即 opt-in、**与仓库用途无关** | 读代码 |
| U14 | 70 个 agent 要不要 70 个 Lark 应用 | ✅ **已证实（要）** | `UNIQUE INDEX (channel_type, config->>'app_id')` ⇒ **一个应用只能绑一个 (workspace, agent)** | migration 124 |
| U15 | 70 个私有仓库在 daemon 宿主上的 git 凭证 | ✅ **已证实可用** | 宿主 keychain 里已有 `192.168.5.254` 的凭证，两个私有仓库 `ls-remote` **exit 0**；团队仓库里已存在 `agent/<name>/<taskKey>` 分支 ⇒ clone/push 链路**已在跑** | 本机 `git ls-remote` |
| U16 | 每轮注入的真实成本 | ⏳ **取到两端常量，A/B 未做** | 通道（`perTurnContextBlocks`）与度量（`task_usage` 按 task 记 input/cache_read/output）都在；30 条索引块实测 **≈570–900 tokens**，本机 claude 任务 avg input 69k / cache_read 416k | 实跑 + 查库 |
| U17 | TAM 的 L3 是 team 级还是 agent 级 | ✅ **已证实** | L3 **和** L2 的存储归属都是 **`(team, agent)` 级** ⇒ 项目专属 agent 隔得住；`injectL2L3` **可整体关掉**。真正的串味路径是"漏配 `x-agent-id` → 静默落进 `agent:default` 桶" | 源码 + 实机双向验证 |
| U18 | "编码时召回"的净收益 | ⏳ **仍未知** | §8.6 目前是设计；A/B 要先把 `BuildMemoryIndexBlock` 写出来（**需改代码，待确认**） | — |
| U19 | 被剔除的 20+ 项有没有误判 | ✅ **已复核** | 23 个仓库无改名/迁移（`gh` 跟随重定向后 `full_name` 未变）、无归档误判；`pushed_at` 差异**全部**来自 dependabot / competition 分支，不影响判决 | `gh api` 逐个 |

下面按条给依据与**对报告的改动**（没有改动的不重复列）。

---

### U1 ✅ claude 的用户级插件 hooks **会触发** —— 架构 D 在 claude 上零代码可用

- **做法**：造一个临时 plugin dir + 隔离 `HOME`/`CLAUDE_CONFIG_DIR` 的探针，用 **daemon 同款调用参数**跑 claude 2.1.274，
  让 hook 写一个 marker 文件；**没有改动 `~/.claude`、没有改动任何仓库代码**。
- **结果**：marker 落盘；debug 输出 `Registered 75 hooks from 31 plugins`。daemon **不传** `--plugin-dir`、也**不注入**
  `HOME` / `CLAUDE_CONFIG_DIR` ⇒ claude 自己加载用户级插件；daemon 的 `--settings` 里只有 `skillOverrides` 与 `permissions`，
  **不覆盖 hooks**。
- **隔离手段**：给某个 agent 的 `custom_env` 设 `CLAUDE_CONFIG_DIR`——它**不在** daemon 的环境变量黑名单里（见 U4）。
- **副作用（如实记录）**：探针 A/D 用了真实 `HOME`，因此执行了**用户已启用插件**的 hooks（未改任何配置，未写用户仓库）；
  一次真跑 claude 花费 `total_cost_usd 0.085205`（走宿主 8787 网关、`deepseek-v4-pro`）；临时目录 `/tmp/u01-exp.byerOA` 已确认清理。
- **对报告的改动**：§2.3 硬约束①、§4.1 架构 D、§8.2-⑩、§8.6 判定表与结论段、§8.8 那一行改「部分成立」、附录 B 新增证据行。

### U2 ⚠️ Reasonix：11 个 hook 事件，但**每轮注入只能走 `AGENTS.md`**

- **做法**：从 `reasonix` 二进制里抽内嵌字符串 + 读上游 `esengine/DeepSeek-Reasonix` 的 `internal/hook/hook.go`、
  `control/turn_orchestrator.go`（源码级，**未做端到端**）。
- **结果**：共 **11 个 hook 事件**。`UserPromptSubmit` **有**事件，但它的 stdout 被**丢弃**（只当"放行/拦截"的闸门）——
  所以"同一套薄脚本挂所有 provider 做每轮注入"在它身上会**静默失效**。可注入的是 `SessionStart`（上限 10k/20k）。
  **每轮注入的正解 = 改写 `{workDir}/AGENTS.md`**。Multica 已经写 `AGENTS.md`（外加 `.reasonix/skills/`、`reasonix.toml`），
  且**没有占用** `.reasonix/settings.json` ⇒ 挂全局 hooks 在任务里**可以生效**（此句为源码级推断，未端到端）。
- **对报告的改动**：§2.2 的 reasonix 行由「❓ 未验证」改为已核实；§4.1 架构 D 修订块。

### U3 ⚠️ TAM 端到端链路 —— **每一跳都实测了，只差"真实 Multica 任务"这一步**

- **做法**：三件套在本机以**独立网络 / 独立卷**拉起（`tamlab-memory-core`:8420、`tamlab-memory-hub`:8125+8424、
  `tamlab-proxy`:8096，等价 `PROXY_FULL_STACK=1` + header 预选打开），自建 HTTP 客户端**扮演 Claude Code**
  打 `POST /claude-code/default/v1/messages`，上游指向本机 mock 以**确定性**观察注入。
- **结果**：**200 转发 + 会话初始化 + 注入全部发生**。
  - **header 预选生效**：带 `x-team-id` + `x-agent-id` + 客户端 session id → 直接转发；
    **不带身份 header → 返回流式的 `msg_cc_session_init_…` 表单** —— §7 R6 的失败模式**被实验复现**。
  - **注入量与形状**：上游收到的 system prompt 里 **`<tdai_profile_memory>` 整块 = 1,394 / 1,407 chars（≈348 / 352 tokens）**，
    拆成 `<l3_core_memory>`（L3 画像**全文**，998 / 1,027 chars ≈ 250 / 257 tok）+ `<l2_scene_index>`
    （**只给 path + summary，正文用工具拉**，145 / 129 chars ≈ 36 / 32 tok）。同一 session 连打两轮，system prompt 均为 8,356 chars 且都含该块
    （`cacheStrategy: session_init`）——**内容稳定，不破坏 KV cache 前缀**。
  - **按 agent 隔离**（交叉检验）：alpha 的块含 `project-alpha` / `alphaSign`、**不含** beta 的任何事实，beta 反之。
  - **可关且关得干净**：`tdai.memory.injectL2L3: false` 后重启 proxy，同一请求 system prompt 由 8,332 → **5,505 chars**，
    `<tdai_profile_memory>` / `<l3_core_memory>` / `<memory-tools-guide>` **全部消失、无记忆内容泄漏**
    （`<skill_tools>` / `<tdai_memory_tools>` / `<session_context>` 仍在，属别的注入器）。
  - **memory-core 侧**的真实 LLM 抽取链路也跑通了（写 L0 后 20s 内产出 L1/L2；L2 是按项目聚合的 Markdown）。
- **⚠️ 两条对报告的修正（都会影响接入）**：
  1. **只给 `x-team-id` + `x-agent-id` 不足以进入注入路径**——还要**客户端 session id header**
     （Claude Code 是 `x-claude-code-session-id`，见 `session/session-key.ts:9-19`）：缺它 → `conversationId = null`
     → **`injectedSkipped = true`，静默跳过 sessionInit 与全部注入，不报错**。
     真实 Claude Code 自带该 header，所以 Multica 的 claude 运行时天然满足；但手工 `curl` 验证时极易误判成"没生效"。
     另：鉴权 key 读 `x-api-key` / `Authorization: Bearer`（= `ANTHROPIC_AUTH_TOKEN`），**不是** `x-tdai-user-key`。
  2. **proxy ↔ memory-hub 那一跳在当前版本已不存在**：`MemoryProxy/src/session/registrar.ts` 写明"本地构建 SessionInfo，
     **不再调用 TMC**"，会话状态落在 proxy 本地；hub 只提供 Panel 与 Knowledge。
     ⇒ 报告此前"sessionInit 的 control plane 在 hub 上"的说法**对当前版本已过时**（这也是本轮一开始被 409 MB 镜像卡住的原因——那个依赖已经没有了）。
- **仍未验证**：① **真实 Multica 任务里的跨任务召回**（要给测试 agent 的 `custom_env` 加 base URL / token / header，
  本轮**没有动任何 Multica 的 agent 配置**）；② L0 每轮写回与 L1 召回的实际效果；③ 70 agent 规模。
- **环境收尾**：三个实验容器、两个卷、`tamlab-net` 网络**全部删除**，mock 进程已停，8420 / 8125 / 8424 / 8096 / 8791 端口均已释放；
  `multica-*`（5 个）与 `postgres-mcp-hardened` 容器未动，`multica-backend` `/health` 仍 200。
  三个 TAM 镜像（共 ~4.5 GB）**刻意保留**以便复跑（拉取很慢）。
- **对报告的改动**：§3.4 实机进度块、§4.4 接入手册（session id 前置条件 + hub 跳修正）、§7 R6、§8.9、附录 B。
- **横向对比（本轮唯一可直接比的一个数）**：TAM 注入 ≈ **350 tokens/请求**，比架构 B/C 设想的"每轮 570–900 tokens 索引块"**更省**
  （因为 L2 只给 path+summary）；代价是 L3 **全文**注入（上游自带"≤2000 字符"约束）且注入逻辑不在自己手里。

### U4 ❌ "传输层代理天然覆盖全部 provider" —— **证伪**

- **做法**：读 `server/internal/daemon/daemon.go` 的环境变量黑名单（`daemon.go:10174` 附近）+ 逐个 runtime 的端点变量。
- **结果**：黑名单 = `MULTICA_*` **前缀** + 14 个具名键（`HOME` / `PATH` / `USER` / `SHELL` / `TERM` / `TMPDIR` / `TMP` / `TEMP` /
  `CODEX_HOME` / `REASONIX_STATE_HOME` / `CURSOR_DATA_DIR` / `CursorMcpAuthSourceEnv` / `OPENCLAW_CONFIG_PATH` / `OPENCLAW_INCLUDE_ROOTS`）。
  于是端点可设性**逐 provider** 不同：
  - **可直接设**：claude / codebuddy（`ANTHROPIC_BASE_URL`）、opencode（config `baseURL` + env）、dsh（`DEEPSEEK_BASE_URL`）
  - **只能走 `custom_args`**：codex（`CODEX_HOME` 被拦，但 `-c model_providers.<id>.base_url=…` 不被拦）
  - **只能改配置文件**：reasonix（**没有任何 `*_BASE_URL` 环境变量**）
  - **未证实**：cursor（本机 `cursor-agent` 是悬空软链、仓库里也没有端点变量）；openclaw / hermes / 另外 16 个 runtime 未逐个查
- **顺带证实**：`ANTHROPIC_CUSTOM_HEADERS` 存在于 claude 2.1.274 二进制**且真的用于建客户端**（所以"注入请求头"这条路对 claude 可行）。
- **对报告的改动**：§4.1 架构 E 的"覆盖列"由理论改为**逐 provider 结论**。

### U5 / U6 / U7 ⚠️ 四个候选里 **ReMe 与 memU 已实机跑通**

- **U5**：报告的"四个候选一个都没跑过"**不再成立**——本轮在 `127.0.0.1:2333` 起了 ReMe，另用一次性容器验了 memU。
  **honcho 与 Neo4j Agent Memory 仍未实机**（这条限制保留）。
- **U6（ReMe 0.4.1.12）**：33 个 HTTP 端点、MCP streamable HTTP 32 工具；**并发写 25+25 条无丢写**（同进程；
  跨进程同路径无锁未测）。三个必须先处理的坑：① 默认分词器是 **regex 逐字** ⇒ 中文等于**单字 OR**，
  `签名` 会召回只含"名"的条目（上生产前应换 jieba 并做 A/B）；② `reme-ai` **装不全**（HTTP 服务要 `[as]` 额外依赖），
  且 **HTTP 服务没有 LLM key 起不来**（"无凭证可跑"只对 SDK 成立）；③ 写入到可检索有 **4–20s** 的 watcher 延迟。
  **自报 89.4% 的水分**：它跑在 HF 上自家发布的 `ReMe_longmemeval_clean_s_v2` 上，数据集卡原文写明做了
  *"replacing the answer and supporting-session ground truth"* —— 即**自己在改过 ground truth 的 clean-S 子集上跑**，
  44,324 tokens/题、3.69 tool calls/题。**这不是"未复现"，是"复现口径与别人不同"**。
  顺带推翻了一个预算假设：**30 条记忆的索引块实测 ≈896 tokens**（只注入 name+desc ≈570），**不是 ≤300**。
- **U7（memU）**：包名是 **`memu-cli`**（PyPI 正式 `0.10.0` / 2026-08-05，预发布 `0.11.0b3` / 08-21，分类器 `4 - Beta`；
  注意 `pip install memu` 会装到**另一个叫 pyboy 的 GUI**）。SQLite 与 `Postgres + pgvector` 都实测建表 + `memu commit` /
  `retrieve` 中文命中。**两个结构性发现**：① **embedding 是硬依赖**（README："Embedding key required"）⇒ 正文"不需要额外 LLM"
  **只对蒸馏成立**；② `AGENTS.md` 注入是**幂等可回滚的 managed block**（与 Multica 自己的标记块能共存），
  但 **Multica 的 workdir 是 per-task 的** ⇒ 打 `./AGENTS.md` 的适配器写的块随任务目录消失。**这不是"被覆盖"，是"无处可写"**。
  坑：缺 `psycopg2` 驱动时 `postgresql+psycopg://` DSN 会被当成**文件路径静默降级**。
- **环境收尾**：ReMe 进程已 kill、端口已释放；一次性容器 `memu-probe-pg` 已删除。
- **对报告的改动**：§0 选型表、§3.1.3 四个候选、附录 B。

### U8 ✅ 榜单漏网风险已量化

- **做法**：`gh api` 批量核 Awesome-AI-Memory 的 **85 个仓库**（存活 + ★ + 最后提交 + 是否在报告里）。
- **结果**：**5 个已死链**（`elizaOS/agentmemory`、`kaderosio/agent-brain`、`tcconnally/mimir`、`tcconnally/perseus`、
  `WeberG619/neveronce`）；**活跃 + 报告未提过 + ★≥300 的只有 4 个**
  （`marciopuga/cog` 376 · `MaxFreedomPollard/Compartment` 581 · `samvallad33/vestige` 624 · `Tura-AI/tura` 637），其余 ★<300。
- **对报告的改动**：§3.2C 那条"仍未判定"改为已量化；附录 B 新增证据行。**"是否空壳"仍未逐个读 README**（保留为已知边界）。

### U9 ✅ MemPalace 与 Hindsight 的实机结论

- **MemPalace 3.10.0**：`uv tool install mempalace` → `init --yes --no-llm` → `mine --direct` 写 3 条 → 中英各查一次
  **都是 rank-1 命中**；**零 key**（ChromaDB + 本地 ONNX MiniLM）。三点修正：① "zero API calls" **≠ 零网络**（首次要下 ONNX 模型）；
  ② **中文只走向量通道**（同一查询 BM25 得分为 0）；③ 官方点名的 impostor 域名是 **`mempalace.tech`**（`.net` 与其他 `.com` 变体同样非法）。
  自动记忆**必须**配 hooks（与 §8.2 的结论一致）。
- **Hindsight**：`ghcr.io/vectorize-io/hindsight` 容器**能起**（v0.10.0：API `:8888` + `/mcp` + Control Plane `:9999`），
  但 **retain / recall 必须有可达的 LLM 端点**——`POST …/memories` 实测 500 `Fact extraction failed`。
  且 **LongMemEval / LoComo 的仓内副本已被删除**（`hindsight-dev/benchmarks/README.md` 明说迁往外部 AMB 仓库）
  ⇒ **"第三方独立复现"在本机无法验证**，"唯二有第三方复现"这句只能保留为**官方自述**。
  三个容器与卷已删。
- **对报告的改动**：§3.1.2 两行。

### U10 ✅ #677 是 PR，**不是否决**

- `gh pr view 677 -R multica-ai/multica`：作者 **benfrank241**，2026-04-10 开，**2026-08-26 由 collaborator Bohan-J 关闭**，
  理由是"与 main 冲突 + 积压清理"，原话 *"This is not a judgment on the idea itself… rebase onto current `main` and reopen"*。
- **对报告的改动**：§1.3 表内注释、§3.1.2 的"闭环"改为**解除**、§0 的换选条件、附录 B。

### U11 ⚠️ 主体 / benchmark / DSH 插件 —— 多为**否定**结论

- **benchmark**：ReMe 的 harness 在仓内（`benchmark/longmemeval/`），**能跑**，但要装 `lme` / `lme-judge` 插件 + HF 数据集 +
  **两份 LLM key**，500 题约 **22M tokens**；ground truth 仍是自家修订版（见 U6）。MemOS 的 benchmark 外包给
  `MemTensor/OmniMemEval`（★60，也是自家 harness）。**第三方独立复现：查不到。**
- **MemTensor 工商主体**：GitHub org 显示名 `OpenMem`、**未认证**，company / blog 字段全为 `null`；
  只能确认官方材料署名"记忆张量（上海）科技有限公司"、自述由上海算法创新研究院孵化，**没有备案号、没有统一信用代码**。
  引用时请按"未核实"处理。
- **DSH 插件**：两个都**真装成**（隔离 `DSH_HOME`）——`@agentscope-ai/reme-dsh-plugin` 0.1.0、
  `@memtensor/memos-cloud-dsh-plugin` 0.1.1；**运行期未测**（ReMe 插件要求 ReMe 在跑、MemOS 云插件要 key）。
- **对报告的改动**：§3.2 D 的中文生态表、附录 B。

### U12 ⏳ 生产规模 —— 仍未知

本轮查库：本机是 **29 agents / 2 workspaces / 5 projects / 156 tasks** 的**开发实例**，不是 70/70/20 的生产形态。
容量与成本结论仍属"按参数推"。**卡点：需要生产实例**。

### U13 ✅ 机制确认，端到端仍需跑一次

`server/internal/handler/vcs_webhook.go:210-231`：MR 靠**标题 / 分支名 / 正文里的 closing 关键字（issue 标识符）**自动关联 issue；
**接入 provider 即 opt-in，与该仓库的用途无关** ⇒ 记忆仓库的 MR 会走**同一套**规则。
剩下的只是"用 1 个测试项目真跑一遍"，**需用户授权**（会往团队 GitLab 写东西）。

### U14 ✅ 一个 Lark 应用只能绑一个 (workspace, agent)

`server/migrations/124_*.up.sql:100-101` 的 `UNIQUE INDEX (channel_type, config->>'app_id')` + 邻近注释：
**同一个 Lark 应用不可能同时绑给两个 agent** ⇒ **70 个 agent 要 70 个应用**（或改成"少量应用 + 显式路由"这类上游目前没有的形态）。

### U15 ✅ daemon 宿主的 git 凭证**可用**

- 团队 GitLab 地址来自 `vcs_connection`：`http://192.168.5.254`；宿主 keychain 里已有该地址的凭证（acct=`zhangshubin`）。
- `GIT_TERMINAL_PROMPT=0 git ls-remote` 两个私有仓库（`zhangshubin/myskills.git`、`wangxinchao/efectivo-amigo.git`）**都 exit 0**。
- 团队仓库里**已经存在** `agent/<name>/<taskKey>` 形式的分支（正是 Multica 在 `repocache/cache.go:865` 的命名规则）
  ⇒ **clone → 推分支这条链路已经在生产里跑着了**。
- 注意：宿主 shell 有 `HTTP_PROXY` / `HTTPS_PROXY`，测试时已 unset（内网地址不该走代理）。

### U16 ⏳ 取到了两端常量，A/B 未做

- **通道**：`perTurnContextBlocks()` 每轮拼块（8.2-⑧）——**已存在**，加一个 `BuildMemoryIndexBlock` 即可。
- **度量**：`task_usage` **按 task 记** input / cache_read / output tokens —— **已存在**，A/B 有现成的表可读。
- **两端常量的量级**：一块 30 条规模的索引实测 **≈570–900 tokens**（U6）；本机 claude 任务实测
  **avg input 69,087 / median 4,999 / avg cache_read 416,161 / avg output 3,456 tokens**（70 个任务），reasonix 任务 avg input 77,646（29 个任务）。
  ⇒ "每轮多几百 token"相对"每轮本来就几万 input"是**小项**，但**乘 70 agent × 轮数**才是要算的账。
- **卡点**：A/B 需要先实现 `BuildMemoryIndexBlock`（**改代码逻辑，需用户确认**）。

### U17 ✅ TAM 的 L3 是 `(team, agent)` 级 —— 既能隔离，也能关

**先说"两份文档为什么不一致"：它们说的不是同一件事。** Proxy 说的 "L3 Team/Global 记忆" 是**内容语义**——
`MEMORY_PROMPT_MODE=code`（默认）下 L3 的生成提示词是 `TEAM_MEMORY_SYSTEM_PROMPT = "# Team Operating Doctrine Architect"`
（`prompts/persona-generation.ts:144,326`）；MemoryCore 说的 "L3 核心画像" 是**内容形态**。**存储归属是第三件事，由 `profile-scope.ts` 决定**：

```ts
// MemoryCore/src/core/profile/profile-scope.ts:34-41
const teamId  = ctx.teamId  || ctx.userId || "default";
const agentId = ctx.agentId || "default";
// L2/L3 are team+agent-level memories. L0/L1 keep user/session/task-level
// isolation, but profiles intentionally ignore userId/sessionId/taskId ...
return `team:${teamId}|agent:${agentId}`;
```

`v2-router.ts:126` 也直接写着「L2 `scenario/*` 与 L3 `core/*` **原本就是 team+agent 级 profile 聚合**」；
API 契约里 `(team_id, agent_id)` 是全接口统一的复合唯一键。

**实机验证**（同一 team、两个 agent，直接读写）：
`core/read as alpha` 只看到 `ALPHA-ONLY-FACT`、`as beta` 只看到 `BETA-ONLY-FACT`，各 1 条、互不可见；
**换一个 user 去读 alpha 的 L3，仍然可见**（⇒ L3 **不**按 user 隔离）。L2 同样隔离且走的是真实 LLM 抽取
（alpha 的 `scenario/ls` 只有 `project-alpha-构建与测试配置.md`，beta 只有 `project-beta-构建与发布配置.md`）。

**能不能关**：能。`MemoryProxy/src/config.ts:108-125` 里 `tdai.memory.injectL2L3` 是总开关、**默认全关**，
`start-proxy.sh` 的 `PROXY_FULL_STACK=1` 才会打开。

⚠️ **额外发现一条真实串味路径**：缺失维度会被**静默兜底成 `agent:default`**。
所以"漏配 `x-agent-id`"**不是**"退化成 team 级 L3"，而是**所有 agent 的记忆全落进同一个 `agent:default` 桶**——
这才是串味的实际形态（v3 会拦，只有绕过 v3 的写入路径可能触发）。§8.3.1 与 §7 R6 的坑要按这个版本理解。

- **对报告的改动**：§8.3.1 的"唯一缺口 L3"整段改写（缺口已闭合，串味形态改为 `agent:default`）、§8.3.1 对比表、
  §8.9 的那条未验证项、§4.1 架构 E 的行。**仍未验证**：注入那一跳（见 U3）。

### U18 ⏳ "编码时召回"的净收益 —— 仍未知

取样集与判据已在 §8.6 定好（`applies_to` + 高频召回点），但**没有做过 A/B**，因此 §8.6 至今是**设计**而不是**结论**。
卡点同 U16：要先实现再开跑。

### U19 ✅ 被剔除的 20+ 项没有误判

- 对 §3.2.1 里 23 个仓库逐个 `gh api`：**没有改名 / 迁移**（`gh` 会跟随重定向，重定向后 `full_name` 均未变）、
  **没有把"归档"误当成"停更"**。
- `pushed_at` 与"默认分支最后提交"的差异**逐个查了分支**：只来自 **dependabot** 或 **比赛分支**
  （`m_flow` 仅 dependabot 09-01、`LycheeMem` 的 `competition/aml-2026` 08-06、`Acontext` 仅 dependabot、`puppyone` 07-30）
  ⇒ **不影响判决**。
- 交叉核对包仓库：`memobase` 的 PyPI 最后发布 `0.0.27` / **2025-11-08**，与主仓停更结论一致。

---

**这一轮核实留下的、仍未闭环的四件事**（按建议顺序）：

1. **U16 / U18**：实现 `BuildMemoryIndexBlock` + 一轮 A/B（**需要先改代码，请先给确认**）。这是唯一能把 §8.6 从"设计"变成"结论"的动作。
2. **U3 的最后一步**：给某个**测试 agent** 的 `custom_env` 加上 base URL / token / header（改配置，不是改代码），跑两次任务看跨任务召回——
   链路的每一跳本轮都已单独验证过，只差"真实 Multica 任务"这一个组合。**需用户确认动 agent 配置。**
3. **U13**：在一个测试项目上真跑一次"记忆仓库 MR → 自动关联 issue → 合并关闭"（需授权写团队 GitLab）。
4. **U12**：在生产实例上复核 70/70/20 的容量与成本。
