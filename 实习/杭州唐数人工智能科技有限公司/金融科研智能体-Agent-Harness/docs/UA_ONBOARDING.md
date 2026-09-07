# TFA Agent Harness 项目上手指南

> 本指南基于项目知识图谱自动生成（understand-onboard），图谱生成于 `2026-08-20`，对应提交 `0b6ab75fe`。
> 注意：当前工作区有 5 个未纳入图谱的新文件（`docs/ideas/`、`docs/specs/`、`tasks/`），本指南暂未覆盖它们的内容。

## 1. 项目概览

- **名称**：`tfa-monorepo`（TFA Agent Harness）
- **一句话定位**：一个可自行扩展的编程智能体（coding agent）框架，类似开源的 Claude Code / Codex CLI
- **语言**：TypeScript 主导（1119 个文件），另有 JS、Markdown、JSON、Vue、SQL、Shell、YAML 等共 15 种
- **框架**：Vitest、Vue、Vite、GitHub Actions
- **规模**：约 1414 个文件；知识图谱含 3988 个节点（1185 文件 / 2294 函数 / 276 类）、8626 条边
- **仓库结构**：npm workspaces monorepo，核心产物发布为 `@earendil-works/tfa-coding-agent`、`tfa-agent-core`、`tfa-ai`、`tfa-tui`、`tfa-telemetry`

## 2. 架构分层（10 层）

| 层 | 职责 | 关键文件 | 节点数 |
|---|---|---|---|
| 入口与 CLI 层 | CLI 命令解析、agent 驱动、工具注册、提示词模板、可执行示例 | `packages/coding-agent/src/cli.ts`、`main.ts`、`config.ts`、`src/index.ts` | 613 |
| 核心 Agent 运行时 | agent 循环、会话/消息编排、记忆与上下文、工具调用调度（不依赖 UI） | `packages/agent/src/index.ts`、`coding-agent/src/core/agent-session-runtime.ts`、`core/session-manager.ts` | 80 |
| 模型与 Provider 层 | 多 provider LLM API、模型目录、认证、流式请求 | `packages/ai/src/index.ts`、`types.ts`、`models.ts`、`providers/all.ts`、`oauth.ts` | 316 |
| UI/TUI 组件层 | 终端 UI 组件、键盘交互、原生渲染 | `packages/tui/src/index.ts`、`coding-agent/src/modes/interactive/interactive-mode.ts`、`theme/theme.ts` | 91 |
| Shell 执行与集成层 | Web 前端（Vue）与 shell 会话接入 | `packages/shell/`（`web/src/*.vue`、`src/server/*`） | 50 |
| 服务与客户端层 | HTTP/WebSocket 服务端与客户端 SDK | `packages/server/src/index.ts`、`packages/client/src/index.ts` | 54 |
| 协议与数据层 | 跨进程消息协议、SQLite 会话存储 | `packages/protocol/src/index.ts`、`packages/protocol/src/cbor/*`、`session-backends/sqlite-node`、`001_initial.sql` | 60 |
| 评估与遥测层 | 评测基准、埋点与遥测上报 | `packages/evals/`、`packages/telemetry/` | 24 |
| 基础设施与 CI/CD 层 | 构建/发布脚本、GitHub Actions、husky 钩子、根级配置 | `scripts/*`、`.github/workflows/*`、`tsconfig.base.json`、`biome.json` | 50 |
| 文档与配置层 | 文档、`.tfa` 扩展配置、分析产物 | `README.md`、`CONTRIBUTING.md`、`.tfa/`、`graphify-out/` | 80 |

依赖方向：CLI 入口 → Agent 运行时 → AI Provider → TUI → 协议/服务/客户端 → Shell/基础设施 → 文档与评测。

## 3. 关键概念

- **Agent 会话循环**：`AgentSession` 驱动一次交互的完整生命周期（输入 → 思考 → 工具调用 → 输出），由 `session-manager.ts` 负责会话文件的读写、迁移与列表。
- **事件总线**：`core/event-bus.ts` 提供轻量订阅/发布，让模块解耦；`system-prompt.ts` 按配置与工具动态生成 LLM 系统提示词。
- **工具系统（Tool）**：`core/tools/index.ts` 汇总 bash/edit/find/grep/ls/read/write 等内置工具；`bash.ts` 是典型实现（spawn 子进程、超时、abort 清理）。
- **统一 Provider 抽象**：`tfa-ai` 的 `types.ts` 是全仓库 fan-in 最高（157）的类型契约；`models.ts` 统一管理多 provider 凭据、认证与推理调用。
- **扩展系统**：`core/extensions/index.ts` 定义扩展 API；`examples/extensions/*` 提供从「自定义工具」到「TUI 游戏」「自定义 provider」的完整样例。
- **消息流 vs LLM 消息**：见 `packages/agent/README.md`，理解 agent 内部 `AgentMessage` 与厂商 LLM Message 的转换。
- **会话持久化**：`session-backends/sqlite-node` 提供 SQLite 存储；`packages/protocol/src/cbor/` 实现协议编解码。
- **差分渲染 TUI**：`packages/tui` 只渲染变化的行，支撑 CLI 交互模式。

## 4. 引导式学习路径（12 步）

1. **项目概览** — 读 `README.md`（定位、包清单、权限模型、开发命令）。
2. **CLI 入口与启动链路** — `cli.ts` → `main.ts` → `config.ts`（用户敲 `tfa` 后发生了什么）。
3. **SDK 公共 API** — `coding-agent/src/index.ts`（barrel）、根 `package.json`、`docs/sdk.md`。
4. **核心 Agent 运行时** — `packages/agent/src/index.ts`、`agent-session-runtime.ts`、`session-manager.ts`、`agent/README.md`。
5. **事件总线与系统提示词** — `core/event-bus.ts`、`core/system-prompt.ts`。
6. **内置工具系统** — `core/tools/index.ts`、`bash.ts`、`model-registry.ts`。
7. **统一 AI Provider 层** — `packages/ai/src/index.ts`、`types.ts`、`models.ts`、`ai/README.md`。
8. **Provider 实现与认证** — `compat.ts`、`providers/all.ts`、`anthropic-messages.ts`、`oauth.ts`、`model-catalog.ts`。
9. **TUI 终端界面** — `packages/tui/src/index.ts`、`interactive-mode.ts`、`theme.ts`。
10. **会话持久化与 SQLite** — `session-backends/sqlite-node` 与迁移 SQL。
11. **协议、服务与客户端** — `protocol`、`server`、`client`、`shell`。
12. **扩展生态与 CI/CD** — `examples/extensions/*`、`.github/workflows/*`、发布脚本。

## 5. 关键文件地图

- **入口**：`packages/coding-agent/src/cli.ts`（CLI 解析）、`main.ts`（主循环）、`src/index.ts`（SDK barrel）、`src/rpc-entry.ts`（RPC 子进程入口）。
- **运行时**：`packages/agent/src/index.ts`、`coding-agent/src/core/agent-session-runtime.ts`、`core/session-manager.ts`、`core/event-bus.ts`、`core/system-prompt.ts`。
- **模型**：`packages/ai/src/types.ts`（类型契约）、`models.ts`（多 provider 管理）、`compat.ts`（统一 stream/complete）、`providers/*`（各厂商适配）、`oauth.ts`（认证）。
- **UI**：`packages/tui/src/index.ts`、`interactive-mode.ts`、`modes/interactive/components/*`（选择器/对话框/动画）。
- **协议/数据**：`packages/protocol/src/cbor/{codec,framing,decoder,encoder}.ts`、`session-backends/sqlite-node`、`001_initial.sql`（11 张表）。
- **Web Shell**：`packages/shell/web/src/App.vue`、`components/*.vue`、`src/server/*`（HTTP 服务 + 项目存储）。

## 6. 复杂度热点（新成员谨慎处理）

- `packages/coding-agent/src/index.ts` — 巨型 barrel，公共 API 面广，改动影响面大。
- `packages/coding-agent/src/modes/interactive/interactive-mode.ts` — 交互主模块，集中全部 TUI 交互逻辑。
- `packages/coding-agent/src/modes/interactive/theme/theme.ts` — 主题系统（schema 校验、颜色转换、终端背景检测）。
- `packages/ai/src/models.ts` / `compat.ts` / `providers/all.ts` — provider 兼容层与模型目录，改错会波及所有上游。
- `packages/coding-agent/src/core/session-manager.ts` — 会话文件格式与迁移，破坏兼容会丢用户数据。
- `packages/protocol/src/cbor/*` — 协议编解码，跨进程契约，修改需服务端/客户端同步。
- `examples/extensions/*`（subagent、gondolin、sandbox、custom-provider-anthropic 等）— 复杂扩展示例，是学习参考但别照抄到核心代码。

## 7. 常用命令

```bash
npm install --ignore-scripts   # 安装依赖（不跑生命周期脚本）
npm run build                  # 刷新模型数据并构建全部包
npm run build:offline          # 离线重建（用已有模型数据）
npm run check                  # lint + format + 类型检查
./test.sh                      # 运行测试（无 API key 时跳过 LLM 相关测试）
./tfa-test.sh                  # 从源码直接运行 tfa
```

## 8. 常见任务入口（How to）

- **新增一个模型 provider**：`packages/ai/src/providers/` 新增适配 + 在 `compat.ts`/`model-catalog.ts` 注册。
- **新增内置工具**：`packages/coding-agent/src/core/tools/` 新增定义与工厂，参考 `bash.ts`。
- **写一个扩展**：复制 `packages/coding-agent/examples/extensions/*` 任一示例，按 `docs/sdk.md` 的扩展 API 编写。
- **改 TUI**：`packages/tui/src`（组件库）+ `coding-agent/src/modes/interactive/`（交互逻辑与组件）。
- **改会话存储**：`packages/session-backends/sqlite-node` + `packages/protocol`（如需改协议）。

## 9. 交互式图谱与仪表盘

- 知识图谱数据：`.ua/knowledge-graph.json`
- 仪表盘：`http://127.0.0.1:5174/?token=e45b91c2c5b9454975c45135cbd10839`
  - **学习**：按上面的 12 步导览浏览
  - **概览**：分层架构视图
  - **深入**：搜索文件/函数/类，查看依赖与被依赖关系
