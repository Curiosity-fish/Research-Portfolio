# Capability Map: school-api-v1 后端重构

## 模块划分

| 模块 id | 职责 | 依赖 |
|---|---|---|
| foundation | 项目骨架、配置、日志、Docker 环境、健康检查 | — |
| auth | 管理员 JWT 认证、API Key 认证、权限上下文 | foundation |
| user-management | 用户 CRUD、API Token 管理、部门层级 | auth |
| ai-relay | 统一 `/v1` 入口、模型/平台/账号/分组管理、OpenAI/Anthropic 转发、SSE、失败切换 | auth, user-management |
| quota-billing | 配额扣减与流水、配额申请审批、充值订单、mock 支付、余额流水、退款 | user-management, ai-relay |
| operations | 通知公告、统计聚合、告警规则、审计日志、对账 | ai-relay, quota-billing |
| integration | 端到端集成测试、部署文档、README | operations |

## 构建顺序

```
foundation
    ↓
auth
    ↓
user-management
    ↓
ai-relay
    ↓
quota-billing
    ↓
operations
    ↓
integration
```

## 说明

- 模块边界按"数据所有权 + 可独立验收"划分。每个模块有自己的 Ent schema、service、handler、repository 和测试。
- `ai-relay` 依赖 `user-management` 是因为转发前需要验证 API Key 并加载用户/分组信息。
- `quota-billing` 依赖 `ai-relay` 是因为配额扣减发生在转发完成后的 usage 结算阶段。
- `operations` 依赖 `quota-billing` 和 `ai-relay` 是因为统计、告警、审计需要调用日志和余额/配额流水数据。
- 每个模块完成后都进行单元测试 + 回归测试；模块合并到大 Phase 后进行集成测试。
