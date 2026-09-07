# school-api-v1

school-api 后端重构版本，一个 AI API 转发与配额计费平台。管理员接入上游平台账号并配置模型，用户通过 API Key 调用 OpenAI / Anthropic 兼容接口，平台负责转发、用量记录、配额预扣与余额结算。

## 技术栈

- Go 1.24+（go.mod 声明 1.24.0，Docker 构建同样使用 golang:1.24-alpine）
- Gin（HTTP）
- Ent（ORM 与代码生成）
- PostgreSQL 15（业务数据）
- Redis（登录限流等）
- JWT（管理员认证）+ API Key（用户认证）
- koanf（配置加载）

## 功能概览

**转发面**（用户 API Key 认证）

- `POST /v1/chat/completions`：OpenAI 兼容转发，支持 SSE 流式
- `POST /v1/messages`：Anthropic 兼容转发
- `GET /v1/models`：当前分组可用模型列表
- 转发前按 `max_tokens` 预扣配额与余额，成功后按实际 usage 结算，失败自动退还
- 多账号加权随机选择与失败切换（5xx / 429 重试）

**用户面**（API Key 认证，`/api/v1`）

- 充值订单（mock 支付渠道）、余额流水、配额流水、配额申请
- 通知与公告（含按用户隔离的已读状态）
- 个人用量统计

**管理面**（JWT 认证，`/api/v1/admin`，全程审计日志）

- 用户 / API Token / 部门管理
- 平台 / 上游账号（API Key AES-256-GCM 加密存储）/ AI 模型 / 分组绑定
- 配额申请审批、余额调整、充值订单查看
- 告警规则与告警记录、Dashboard 与按模型/用户/日期用量统计
- 审计日志、订单对账与退款

## 快速开始

### 前置要求

- Go 1.24+（本地运行）
- Docker / Docker Compose（容器方式）

### 本地开发

```bash
cp config.yaml.example config.yaml   # 按需修改
go run ./cmd/server
curl http://localhost:8080/health
```

数据库迁移使用独立命令（见下文「数据库迁移」），迁移完成后服务即可正常启动。

### Docker 开发环境

```bash
cp .env.example .env
# 编辑 .env，至少设置 APP_JWT__SECRET 与 APP_ENCRYPTION__KEY

docker compose up --build
curl http://localhost:8080/health
```

Compose 会依次启动 PostgreSQL、Redis、迁移任务（`migrate`）与后端服务。PostgreSQL 与 Redis 分别映射到宿主机 `15432` / `16379`。

## 配置说明

配置来源优先级（低 → 高）：默认值 < YAML 配置文件 < 环境变量（前缀 `APP_`，双下划线 `__` 表示层级）：

```bash
APP_SERVER__PORT=8080
APP_DATABASE__URL=postgres://user:pass@localhost:5432/school_api_v1
APP_REDIS__URL=redis://localhost:6379/0
APP_JWT__SECRET=your-jwt-secret
APP_JWT__EXPIRE_HOURS=24
APP_ENCRYPTION__KEY=base64url编码的32字节密钥
APP_RELAY__TIMEOUT=120s
APP_RELAY__MAX_RETRIES=2
APP_BILLING__DEFAULT_MAX_TOKENS=8192
```

显式设置 `APP_CONFIG_PATH` 但文件不存在时程序拒绝启动；不设置时可纯环境变量启动。

关键配置项：

| 配置键 | 说明 |
|--------|------|
| `server.env` | `development` / `test` / `production`，生产环境启用可信代理限制 |
| `jwt.secret` | 管理员 JWT 签名密钥，生产必须设置为强随机值 |
| `encryption.key` | 上游账号 API Key 的 AES-256-GCM 加密密钥，base64url 编码的 32 字节，生产必须设置且妥善保管 |
| `relay.timeout` / `relay.max_retries` | 上游转发超时与失败切换重试次数 |
| `billing.default_max_tokens` | 请求未声明 `max_tokens` 时的预扣估算值 |

生成加密密钥：

```bash
openssl rand -base64 32
```

## 数据库迁移

迁移文件位于 `migrations/`，由 `cmd/migrate` 执行（同样读取 `APP_DATABASE__URL`）：

```bash
go run ./cmd/migrate up        # 应用全部迁移
go run ./cmd/migrate down      # 回滚一个版本
go run ./cmd/migrate down 3    # 回滚三个版本
go run ./cmd/migrate version   # 查看当前版本
```

Docker 部署时由 compose 中的 `migrate` 服务自动执行 `up`。

## Mock 数据

`cmd/mockgen` 可在开发库生成一批管理员、分组、部门、用户与 Token（拒绝生产环境）：

```bash
go run ./cmd/mockgen            # 生成
go run ./cmd/mockgen -cleanup   # 清理
```

## 常用命令

```bash
make build             # 构建 server / migrate / admin 到 bin/
make test              # 单元测试
make test-integration  # Docker 集成测试（含数据库）
make generate          # 重新生成 Ent 代码（固定 GOTOOLCHAIN=go1.24.0）
make migrate-up        # 应用迁移
```

## 测试

- 单元测试：`go test ./...`，repository 层使用 SQLite 内存库，无需外部依赖。
- 集成测试：`make test-integration`，通过 `docker-compose.test.yml` 启动真实 PostgreSQL 与 Redis，覆盖认证、用户、计费、运营支撑等端到端链路（build tag `integration`）。

## 项目结构

```text
school-api-v1/
├── cmd/
│   ├── server/         # 服务入口
│   ├── migrate/        # 数据库迁移命令
│   ├── admin/          # 管理员工具
│   └── mockgen/        # mock 数据生成器
├── internal/
│   ├── auth/           # JWT 与 API Key 认证
│   ├── config/         # 配置加载与校验
│   ├── domain/         # 领域错误与公共类型
│   ├── encryption/     # AES-256-GCM 加密
│   ├── handler/        # HTTP handler
│   ├── relay/          # 上游转发器与 SSE 工具
│   ├── repository/     # 数据访问（Ent）
│   ├── server/         # HTTP 服务器、路由、中间件
│   ├── service/        # 业务逻辑
│   ├── mockdata/       # mock 数据生成
│   └── testutil/       # 测试工具
├── ent/                # Ent schema 与生成代码
├── migrations/         # 数据库迁移（up/down 成对）
├── test/integration/   # 集成测试（build tag integration）
└── learning/           # 各阶段构建笔记
```

## 文档

- [SPEC.md](docs/SPEC.md)：需求规格
- [SPEC-capability-map.md](docs/SPEC-capability-map.md)：能力映射
- [DEPLOY.md](docs/DEPLOY.md)：部署手册
- [plan.md](tasks/plan.md) / [todo.md](tasks/todo.md)：重构计划与任务清单
