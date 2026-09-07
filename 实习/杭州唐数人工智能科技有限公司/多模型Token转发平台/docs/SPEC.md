# Spec: school-api-v1 后端重构

## 假设清单

在写本 spec 之前，我做出以下假设，如有错误请立即纠正：

1. 本项目是后端 API 服务，不直接服务前端页面（前端后续单独重构）。
2. 认证采用双轨制：对外 AI 转发使用 API Key，管理后台使用 JWT。
3. 数据库使用 PostgreSQL，缓存/限流/会话使用 Redis。
4. 目标用户规模为 5000–10000 人，长期高并发性能上限优先。
5. 旧项目 `school-api` 的数据不需要迁移（原数据为编造数据），但新 Schema 核心字段与枚举值仍尽量与旧系统保持命名一致，减少团队认知成本。
6. 生产部署优先使用 Docker，本地开发使用 Docker Compose。

## Objective

从零重建 `school-api` 后端，使其成为生产级、可长期演进、新增功能不易与现有代码冲突的 API 中转与管理系统。

主要用户与场景：

- **普通用户**：通过 `Authorization: Bearer sk-xxx` 调用 `/v1/chat/completions` 或 `/v1/messages` 使用 AI 模型。
- **管理员**：通过管理后台管理用户、Token、模型/上游账号、配额、充值订单、退款、公告、告警等。

成功标准：

- 所有功能模块均有单元测试 + 回归测试，大功能合并后有集成测试。
- AI 转发支持 OpenAI 与 Anthropic 兼容格式，支持 SSE 流式。
- 配额、余额、计费链路在高并发下正确、可追踪。
- 代码结构清晰，新增功能有明确归属位置，不会随意侵入其他模块。
- 可通过 `docker compose up` 一键启动完整开发/测试环境。

## Tech Stack

| 层级 | 技术 |
|---|---|
| 语言 | Go 1.23+ |
| Web 框架 | Gin |
| ORM | Ent |
| 数据库 | PostgreSQL 15+ |
| 缓存/会话/限流 | Redis 7+ |
| 配置加载 | koanf |
| 日志 | slog（结构化） |
| 依赖注入 | Google Wire |
| 测试 | Go 标准 testing + httptest + testcontainers（集成测试） |
| 迁移 | golang-migrate |
| 部署 | Docker / Docker Compose |

## Commands

```bash
# 开发启动（Docker Compose）
docker compose up --build

# 运行所有测试
go test ./...

# 运行集成测试（使用 docker-compose.test.yml）
docker compose -f docker-compose.test.yml up --abort-on-container-exit

# 数据库迁移
make migrate-up

# 生成 Ent 代码
go generate ./ent/...

# 构建生产二进制
make build
```

## Project Structure

```
school-api-v1/
├── cmd/
│   └── server/
│       ├── main.go
│       ├── wire.go
│       └── wire_gen.go
├── docs/
│   ├── SPEC.md
│   └── SPEC-capability-map.md
├── ent/
│   ├── schema/
│   ├── generate.go
│   └── ...（Ent 生成代码）
├── internal/
│   ├── config/          # 配置加载与校验
│   ├── domain/          # 领域常量、错误定义、仓库接口
│   ├── handler/         # HTTP handler
│   ├── middleware/      # Gin 中间件
│   ├── pkg/             # 可复用工具包（如 openai/anthropic 协议处理）
│   ├── repository/      # 仓库实现（Ent 封装）
│   ├── server/          # Gin 引擎、路由注册、优雅关闭
│   ├── service/         # 业务逻辑
│   └── testutil/        # 测试辅助函数、fixtures、stubs
├── migrations/          # 数据库迁移文件
├── scripts/             # 工具脚本
├── tasks/               # 计划与任务清单
│   ├── plan.md
│   └── todo.md
├── docker-compose.yml
├── docker-compose.test.yml
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

## Code Style

### 命名约定

- 包名：全小写，无下划线（`repository`、`service`）。
- 接口名：以 `Repository`、`Service` 等后缀结尾，或根据行为命名（`Tokenizer`、`Quoter`）。
- 结构体名：名词（`UserService`、`TokenRepository`）。
- 函数名：动词开头（`CreateUser`、`DeductQuota`）。
- 错误变量：以 `Err` 开头（`ErrUserNotFound`）。

### 分层规则

- `handler` 只负责 HTTP 解析/封装、调用 service、返回响应，不直接访问数据库。
- `service` 负责业务逻辑、事务编排、跨领域协调。
- `repository` 负责数据访问，对外隐藏 Ent 细节。
- `domain` 定义仓库接口，由 `repository` 实现，便于测试时 mock。

### 示例

```go
// internal/domain/user.go
type UserRepository interface {
    Create(ctx context.Context, user *User) (*ent.User, error)
    GetByUsername(ctx context.Context, username string) (*ent.User, error)
    GetByID(ctx context.Context, id int) (*ent.User, error)
}

// internal/service/user.go
type UserService struct {
    userRepo domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) *UserService {
    return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*ent.User, error) {
    // 业务校验、密码哈希、调用 repo
}

// internal/handler/user.go
func (h *UserHandler) Create(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respond.Error(c, domain.ErrInvalidRequest)
        return
    }
    user, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        respond.Error(c, err)
        return
    }
    respond.OK(c, user)
}
```

### 错误处理

- 业务错误使用 `domain.AppError`，包含 HTTP 状态码和业务码。
- 系统错误直接返回，由中间件统一转换为 500。
- 不吞异常、不返回伪造默认值。

## Testing Strategy

### 测试分层

1. **单元测试**：与源码同包，命名 `*_test.go`，覆盖 service、repository、handler、middleware。
2. **集成测试**：命名 `*_integration_test.go`，覆盖数据库 + Redis + HTTP 完整链路。
3. **端到端测试**：位于 `internal/integration/`，覆盖"注册 → 充值 → 创建 Token → 调用 AI → 查看统计"主链路。

### 测试要求

- 每个新增函数/方法都应有对应单元测试。
- 涉及数据库的测试使用 testcontainers 或事务回滚，保证隔离。
- 每个 Task 完成后必须能通过 `go test ./...`。
- 集成测试在 `docker-compose.test.yml` 环境中可一键运行。

### Mock 与 fixtures

- `internal/testutil` 提供通用 fixtures（如创建测试用户、Token、模型、账号）。
- 仓库接口定义在 `domain` 包，便于 service 层单元测试时 mock。

## Boundaries

### Always（默认执行）

- 每次提交/结束任务前运行 `go test ./...`。
- 遵循目录结构与命名约定。
- 所有输入在 handler 层做校验。
- 关键业务操作（扣费、充值、退款）写入流水表，保证可追溯。
- 密码、API Key 等敏感信息必须哈希或加密存储。

### Ask First（先询问）

- 修改公共接口、路由路径、请求/响应结构。
- 新增核心依赖。
- 修改数据库 Schema（Ent schema 之外）。
- 改变认证/权限逻辑。
- 引入全局状态或单例。
- 超出本计划范围的新功能。

### Never（不做）

- 在代码中硬编码密码、密钥、令牌。
- 删除或跳过失败的测试以让测试通过。
- 直接在生产环境执行未验证的迁移或脚本。
- 为展示技巧引入无必要的设计模式或抽象。

## Success Criteria

- [ ] 后端所有 Phase 完成，每个 Task 都有单元测试 + 回归测试。
- [ ] 每个大 Phase 合并后都有集成测试并通过。
- [ ] OpenAI `/v1/chat/completions` 与 Anthropic `/v1/messages` 均可通过统一 `/v1` 入口正常转发，支持 SSE。
- [ ] 配额扣减、充值、余额流水链路正确，并发下无超扣。
- [ ] 管理后台 JWT 认证与 API Key 认证双轨运行正常。
- [ ] `docker compose up` 可一键启动完整开发环境。
- [ ] 代码结构清晰，新增功能模块有明确位置，不侵入其他模块。

## Open Questions

1. Go module 名称是否确认使用 `github.com/school-api/school-api-v1`？
2. 是否需要保留旧项目中的 WebSocket 实时通知，还是先使用普通 HTTP 轮询？
3. mock 支付是否只需要一个虚拟渠道用于测试，还是需要模拟支付宝/微信等特定渠道？
4. 是否需要集成第三方 OAuth 登录（微信、钉钉等），还是仅保留用户名密码登录？

## 附录：Phase 8 增量需求（2026-09-04）

Phase 1–7 完成后对照原项目 `school-api` 逐端点核对，确认以下功能缺口并纳入 Phase 8。命名沿用 v1 现有约定（`/api/v1` 用户面、`/api/v1/admin` 管理面、`{code,message,data}` 信封）。

### A. 用户端账户体系（Task 30，最高优先级）

原项目的用户门户基于账号密码登录，v1 用户只有 API Key、无法进入网页端。

- `POST /api/v1/user/login`：用户名 + 密码 → 用户 JWT。安全要求对齐原项目：错误统一为"账号或密码错误"（防用户枚举）、用户不存在时执行等成本哈希校验（防时序侧信道）、登录限流（Redis，与管理端一致）。
- 用户 JWT 与管理端 JWT 使用同一 TokenManager 但建议加 audience 声明区分（`aud=admin`/`aud=user`），两个鉴权中间件各自校验 audience，防止令牌串用。
- `POST /api/v1/user/change-password`：校验旧密码，更新 bcrypt 哈希。
- `POST /api/v1/user/token/reveal`：密码确认后返回指定 user_token 的明文。前提：`user_tokens` 新增加密明文列（AES-256-GCM，复用 `encryption.key`），创建 Token 时同步写入；存量无密文返回 409 `TOKEN_REVEAL_UNAVAILABLE`。用户 Token 的 `plaintext_token` 仍为一次性展示，reveal 是补救通道。
- users 表 password 字段已存在（管理员创建用户时写入），本任务为其补上消费方。

### B. 通知增强（Task 31）

- `GET /api/v1/notifications/unread-count`：当前用户未读数（广播按 notification_reads 判定 + 个人按 is_read）。
- `POST /api/v1/notifications/read-all`：全部标记已读（广播批量插入 reads，容忍唯一冲突；个人批量更新）。

### C. 系统设置（Task 32）

settings 表已存在但无消费接口，补齐：

- `feature_mode`：`quota_only` / `recharge_only` / `both`，管理端 `GET/PUT`，影响用户端是否显示充值/配额申请入口（后端仅在创建订单/申请时校验模式，前端展示以此为准）。
- 默认配额：新 user_token 的默认 quota_limit，管理端 `GET/PUT`，在创建 Token 时生效。
- 充值配置：单笔限额 min/max、快捷金额列表，管理端 `GET/PUT` + 用户端 `GET /api/v1/recharge/config`（含支付渠道列表）。

### D. 使用指南（Task 33）

`guide_sections` 新表：`audience`（user/admin）、`title`、`content_md`、`sort_order`、`is_enabled`。管理端 CRUD，用户端 `GET /api/v1/guide/sections`（只返回 enabled，按 sort_order 排序）。

### E. 统计扩展（Task 34）

基于 call_logs / balance_records / recharge_orders 聚合（只读，不改写 Phase 6 已有 dashboard/stats/usage）：

- 管理端：用户排行 `stats/rankings`、调用趋势 `stats/trend`（按小时/按天）、模型分布 `stats/model-dist`、部门分布 `stats/dept-dist`、财务汇总 `finance/summary`（充值/消费/退款）。
- 用户端：`stats/trend`、`stats/model-stats`。

### F. 导出与批量导入（Task 35）

- XLSX 导出（引入 excelize）：管理端使用统计/汇总/计费流水/充值订单导出，用户端余额流水导出。
- 用户批量导入：管理端 `POST /users/bulk-import`（JSON 数组）与 `POST /users/bulk-import-file`（XLSX 上传），逐行校验、返回逐行结果（成功/失败原因），不整批失败。

### G. 退款申请流程（Task 36）

当前只有管理员直接退款。改为用户发起、管理员审批：

- `refund_requests` 表：order_id、user_id、amount、reason、status(pending/approved/rejected)、reviewed_by/at。
- 用户端：`POST /api/v1/refund-requests`（校验订单归属与可退余额）、`DELETE /api/v1/refund-requests/:id`（仅 pending 可撤销）、`GET /api/v1/refund-requests`。
- 管理端：列表、`POST /refund-requests/:id/approve`（审批通过走现有 RefundOrder 事务，幂等）、`POST /refund-requests/:id/reject`。同一订单同时最多一条 pending 申请。
- 管理员直接退款接口保留（原有路径），不与申请流程互斥但需在文档注明适用场景。

### H. 转发面补齐（Task 37）

- `POST /v1/responses`：OpenAI Responses API 兼容，内部转换为 chat/completions 走现有转发链路（含计费），响应再转回 Responses 格式。仅保证 Codex 类客户端可连通，不保证全部 Responses 参数语义。

### 明确延后项及理由

| 功能 | 理由 |
|------|------|
| WebSocket 实时推送 | SPEC 原有 Open Question #2 未决；前端先轮询实现，推送可作为 Phase 9 增强。Phase 8 接口设计需为推送留事件模型余地（通知/审批结果已落表，推送只需订阅） |
| 模型健康定时探测 | 需引入后台任务基础设施（scheduler、探测限流），当前 evaluate 也是手动触发；等告警定时化时一并设计 |
| 系统日志面板 | 生产环境应走 slog stdout + 容器日志采集（DEPLOY.md 已定），再落一套 DB 日志表造成双写与性能负担 |
| 配额告警独立面板 | v1 通用告警规则已覆盖 balance_low / quota_low 指标，独立面板只是另一种展示，由前端复用现有 alerts 接口实现 |

### Phase 8 约束

- 所有新接口遵循既有分层与约定（AppError、分页、micro-currency、审计中间件自动覆盖管理端写操作）。
- 每个 Task 完成后 `go test ./...` 通过；Task 30、36 必须有集成测试（登录安全属性、退款申请并发/幂等）。
- 完成后同步更新 `README.md`、`docs/DEPLOY.md`（如新增配置键）、`learning/frontend-handoff.md`。
