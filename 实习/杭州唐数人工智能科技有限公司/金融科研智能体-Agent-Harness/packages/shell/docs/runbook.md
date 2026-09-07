# tfa-shell 运行手册(Windows)

tfa-shell 是以 tfa 引擎为 SDK 的金融科技 agent 壳:CLI 交互式工具 + Web 前端页面。

## 一、构建

```powershell
# 全仓构建(含 shell 包;离线模式跳过联网拉取模型目录)
npm run build:offline

# 单独构建 shell 包
npm run build -w tfa-shell

# 构建 Web 前端
cd packages/shell/web
npm install
npm run build
```

## 二、CLI 使用

```powershell
# 交互式终端(需要真实终端窗口;faux 离线模式用下面这条)
$env:TFA_SHELL_FAUX='1'
node packages/shell/dist/cli.js

# 单发模式
node packages/shell/dist/cli.js -p "帮我分析一下行情"

# 列出可用模型(无需 API key)
node packages/shell/dist/cli.js --list-models

# 指定模型/厂商(需先在 ~/.tfa/agent/auth.json 配置 API key)
node packages/shell/dist/cli.js --model claude-opus-4-5 --provider anthropic
```

常用参数(与 tfa 一致):`--model`、`--provider`、`--tools`、`--session`、`-c/--continue`、`-r/--resume`、`--offline`。

## 三、Web 服务

```powershell
# 生产模式:后端服务 + 静态托管前端构建产物(默认 127.0.0.1:8787)
$env:TFA_SHELL_FAUX='1'
node packages/shell/dist/cli.js web

# 指定端口/工作目录
node packages/shell/dist/cli.js web --port 9000
$env:TFA_SHELL_CWD = 'D:\projects\research'   # 会话默认工作目录
$env:TFA_SHELL_PORT = '9000'

# 开发模式:前端热更新(另开终端)
cd packages/shell/web
npm run dev    # http://127.0.0.1:5173,代理 /api 到 8787
```

### Web 页面功能

- 左侧会话列表:新建、切换、删除会话
- 主区聊天:流式回复、思考过程折叠、工具调用卡片(工具名/参数/状态/结果)
- 顶部:模型切换下拉(标记未认证模型)、连接状态徽标
- 输入框:Enter 发送,Shift+Enter 换行;生成中可点击"停止生成"

### REST API 摘要

| 方法 | 路径 | 说明 |
| -- | -- | -- |
| GET | /api/health | 健康检查 |
| GET | /api/sessions | 会话列表 |
| POST | /api/sessions | 创建会话 {name?, cwd?, model?, thinkingLevel?} |
| DELETE | /api/sessions/:id | 删除会话 |
| GET | /api/sessions/:id | 会话快照 |
| POST | /api/sessions/:id/prompt | 发送消息 {text} |
| POST | /api/sessions/:id/steer | 生成中插话 {text} |
| POST | /api/sessions/:id/abort | 中止生成 |
| POST | /api/sessions/:id/model | 切换模型 {model:{provider,id}} |
| POST | /api/sessions/:id/thinking | 设置思考级别 {thinkingLevel} |
| GET | /api/sessions/:id/events | SSE 事件流(snapshot/progress/error) |
| GET | /api/models | 模型列表 |

## 四、离线模式(TFA_SHELL_FAUX)

`TFA_SHELL_FAUX=1` 时注册 faux provider:不访问任何真实 LLM,回复回显最后一条用户消息。
用于无 API key 时的全链路验证(CLI、REST、SSE、Web 页面)。

```powershell
$env:TFA_SHELL_FAUX='1'
node packages/shell/dist/cli.js -p "你好"
```

## 五、加入金融工具(三步法)

1. **实现数据源**:在 `packages/shell/src/fintech/datasources/` 新建文件,实现 `MarketDataSource` 接口(见该目录 README)
2. **注册工具**:把数据源包装成 `ToolDefinition`(`defineTool`),注册到 `FintechToolRegistry`
3. **注入会话**:`createShellRuntimeFactory({ cwd, customTools: registry.toCustomTools() })`(CLI)或 `new ShellSessionService({ cwd, customTools: registry.toCustomTools() })`(Web)

注册后工具自动出现在 CLI 可用工具与 Web 的工具卡片中。

参考示例:`packages/shell/src/fintech/echo-market.ts`(占位工具,用于验证链路)。

## 六、测试

```powershell
npm test -w tfa-shell                 # 单元测试(18 个,全部离线)
cd packages/shell/web && npm run typecheck   # 前端类型检查
cd packages/shell/web && npm run build       # 前端构建
```

## 七、常见问题

- **`No models available`**:未配置 API key。运行 `tfa-shell` 前先执行 `npx tfa login` 或参考 `packages/coding-agent/docs/providers.md` 配置 `~/.tfa/agent/auth.json`
- **Web 页面空白**:确认先执行 `npm run build`(web 目录),后端从 `packages/shell/web/dist` 托管前端
- **端口占用**:换端口 `tfa-shell web --port 9000`,并同步修改 `packages/shell/web/vite.config.ts` 的代理目标(开发模式)
