<p align="center">
  <a href="https://pi.dev">
    <img alt="tfa logo" src="https://pi.dev/logo-auto.svg" width="128">
  </a>
</p>
<p align="center">
  <a href="https://discord.com/invite/3cU7Bz4UPx"><img alt="Discord" src="https://img.shields.io/badge/discord-community-5865F2?style=flat-square&logo=discord&logoColor=white" /></a>
  <a href="https://www.npmjs.com/package/@earendil-works/pi-coding-agent"><img alt="npm" src="https://img.shields.io/npm/v/@earendil-works/pi-coding-agent?style=flat-square" /></a>
</p>

> 新贡献者提交的 issue 和 PR 默认会被自动关闭。维护者会每天审阅这些自动关闭的 issue。参见 [CONTRIBUTING.md](CONTRIBUTING.md)。

# TFA Agent Harness

这里是 TFA 智能体框架（TFA Agent Harness）项目的大本营，包含我们可自行扩展的编程智能体。

* **[@earendil-works/tfa-coding-agent](packages/coding-agent)**：交互式编程智能体 CLI
* **[@earendil-works/tfa-agent-core](packages/agent)**：支持工具调用与状态管理的智能体运行时
* **[@earendil-works/tfa-ai](packages/ai)**：统一的 multi-provider LLM API（OpenAI、Anthropic、Google 等）

想进一步了解 TFA：

* [访问 pi.dev](https://pi.dev)，这是带有演示的项目官网
* [阅读文档](https://pi.dev/docs/latest)，也可以直接让智能体自己解释

## 全部包（Packages）

| 包 | 说明 |
|---------|-------------|
| **[@earendil-works/tfa-telemetry](packages/telemetry)** | 厂商中立的遥测契约与类型化 schema 工具 |
| **[@earendil-works/tfa-ai](packages/ai)** | 统一的 multi-provider LLM API（OpenAI、Anthropic、Google 等） |
| **[@earendil-works/tfa-agent-core](packages/agent)** | 支持工具调用与状态管理的智能体运行时 |
| **[@earendil-works/tfa-coding-agent](packages/coding-agent)** | 交互式编程智能体 CLI |
| **[@earendil-works/tfa-tui](packages/tui)** | 基于差分渲染的终端 UI 库 |

Slack/聊天自动化与工作流请参见 [earendil-works/pi-chat](https://github.com/earendil-works/pi-chat)。

## 权限与容器化

TFA 没有内置权限系统来限制文件系统、进程、网络或凭据的访问。默认情况下，它以启动它的用户和进程所拥有的权限运行。

如果需要更强的边界，请对 TFA 进行容器化或沙箱化。参见 [packages/coding-agent/docs/containerization.md](packages/coding-agent/docs/containerization.md) 中的三种模式：

- **Gondolin 扩展**：把 `tfa` 和 provider 认证保留在宿主机上，同时将内置工具和 `!` 命令路由到本地 Linux 微型虚拟机中。
- **纯 Docker**：将整个 `tfa` 进程运行在本地容器中，实现简单的隔离。
- **OpenShell**：将整个 `tfa` 进程运行在受策略控制的沙箱中。

## 贡献

贡献指南参见 [CONTRIBUTING.md](CONTRIBUTING.md)，项目专属规则（适用于人类和智能体）参见 [AGENTS.md](AGENTS.md)。TFA 的长期规划也可以在这里找到：[RFCs](https://rfc.earendil.com/keyword/pi/)。

## 开发

```bash
npm install --ignore-scripts  # 安装所有依赖但不执行生命周期脚本
npm run build         # 刷新模型数据，然后构建所有包
npm run build:offline # 使用已有模型数据重新构建，无需网络
npm run check         # 代码检查、格式化和类型检查
./test.sh            # 运行测试（无 API key 时跳过依赖 LLM 的测试）
./tfa-test.sh         # 从源码运行 tfa（可在任意目录执行）
```

## 从发布源码构建独立二进制

GitHub 发布包含一个带版本的源码归档，并有该发布的 `SHA256SUMS` 文件做校验。解压后，运行与官方独立二进制相同的构建脚本：

```bash
VERSION="<release-version>"
tar -xzf "tfa-${VERSION}-source.tar.gz"
cd "tfa-${VERSION}"
./scripts/build-binaries.sh --offline-model-data --platform linux-x64 --out "$PWD/out"
```

源码归档包含该发布所使用的已生成 provider 模型数据。`--offline-model-data` 会使用这份快照构建，而不是从线上 provider 目录刷新数据。该脚本仍会安装依赖、构建 monorepo、编译 Bun 可执行文件，并暂存其运行时资源。如果维护者单独提供依赖，可以传入 `--skip-install --skip-deps`。

## 供应链加固

我们把 npm 依赖变更视为需要评审的代码变更。

- 直接外部依赖固定到精确版本。内部 workspace 包保持版本区间。
- `.npmrc` 设置了 `save-exact=true` 和 `min-release-age=2`，避免在 npm 解析时用到当天发布的依赖。
- `package-lock.json` 是依赖的权威基准。pre-commit 会阻止意外的 lockfile 提交，除非设置 `TFA_ALLOW_LOCKFILE_CHANGE=1`。
- `npm run check` 会校验固定的直接依赖、原生 TypeScript 导入兼容性，以及生成的 coding-agent shrinkwrap。
- 发布的 CLI 包包含 `packages/coding-agent/npm-shrinkwrap.json`（由根 lockfile 生成），为 npm 用户固定传递依赖。
- 发布冒烟测试使用 `npm run release:local` 在打 tag 之前在仓库外构建、打包，并创建隔离的 npm 与 Bun 安装。
- 本地发布安装、文档说明的 npm 安装以及 `tfa update --self` 在支持的地方都使用 `--ignore-scripts`。
- CI 使用 `npm ci --ignore-scripts` 安装，并有一个定时 GitHub workflow 运行 `npm audit --omit=dev` 和 `npm audit signatures --omit=dev`。
- shrinkwrap 生成对依赖的生命周期脚本有显式白名单；新增带生命周期脚本的依赖在评审通过前无法通过检查。

## 分享你的开源编程智能体会话

如果你在开源工作中使用了 TFA 或其他编程智能体，请分享你的会话。

公开的开源会话数据有助于用真实世界的任务、工具使用、失败与修复来改进编程智能体，而不是只靠玩具基准。

完整说明参见 [X 上的这篇文章](https://x.com/badlogicgames/status/2037811643774652911)。

要发布会话，请使用 [`badlogic/tfa-share-hf`](https://github.com/badlogic/pi-share-hf)。阅读其 README.md 获取配置说明。你只需要一个 Hugging Face 账号、Hugging Face CLI 和 `tfa-share-hf`。

你也可以观看[这个视频](https://x.com/badlogicgames/status/2041151967695634619)，里面演示了我是如何发布我的 `tfa-mono` 会话的。

我定期在这里发布自己的 `tfa-mono` 工作会话：

- [badlogicgames/tfa-mono on Hugging Face](https://huggingface.co/datasets/badlogicgames/pi-mono)

## 许可证

MIT

<p align="center">
  <a href="https://pi.dev">pi.dev</a> 域名由
  <br /><br />
  <a href="https://exe.dev"><img src="packages/coding-agent/docs/images/exy.png" alt="Exy mascot" width="48" /><br />exe.dev</a>
</p>
