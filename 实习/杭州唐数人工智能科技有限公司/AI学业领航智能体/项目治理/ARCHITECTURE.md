# Academic Navigation - 架构参考手册

> 本文件记录当前运行链的架构事实和约束。代码清单、接口行号、数据库行数和部署地址以实时检查为准。

## 一、项目目标与边界

- 项目目标：将高校学生的成绩、学分、画像、体测/考勤和预警汇聚成可行动的分角色学业导航。
- 主要用户：学生、班主任、任课教师、专业负责人、院领导。
- 当前包含：五角色门户、学业分析、预警和干预、发展路径、职位/院校缓存、本地 LLM 指标解读、校方数据源适配骨架。
- 当前不包含：已上线的 CAS、已提供的校方生产 API、后端 Agent 会话中心、生产部署编排和实时消息总线。
- 部署形态：当前为本地前后端分离；前端 Vite 5173 代理到后端 8080，后端连接 MySQL 和 Redis。生产拓扑尚未批准。
- 数据敏感等级：内部敏感。学生成绩、心理、体测、考勤和身份数据按最小权限处理，演示数据必须脱敏。

## 二、统一词表

| 词 | 统一含义 | 禁止或易混淆用法 |
|---|---|---|
| 学生 | `t_user.role=student`，关联 `t_student` 的最终用户 | 不把前端 mock 学生当成生产身份 |
| 班主任 | `t_user.role=teacher`，负责班级范围 | 不等同于任课教师 |
| 任课教师 | `t_user.role=course_teacher`，只看所授课程班级 | 不因访问 `/teacher/**` 就获得班级全量权限 |
| 专业负责人 | `t_user.role=department`，管理所属专业数据 | 不扩大为全院权限 |
| 院领导 | `t_user.role=dean`，可见全院汇总，并可执行校方同步 | 不把前端隐藏按钮当成授权 |
| 当前学期 | API 的 `term` 参数，格式由 `AcademicTerm` 统一解析 | 不在页面自行拼接不同格式的学期字符串 |
| 预警等级 | `yellow`、`orange`、`red`；状态另用 `pending/processing/resolved` | 不用心理测评结果直接替代预警等级 |

## 三、技术栈基线

| 层/领域 | 选型 | 选择原因 | 禁止随意替换为 |
|---|---|---|---|
| 前端 | Vue 3.5、TypeScript 6、Vite 8 | 已有五端路由和组件体系 | 不在单页引入另一套框架 |
| UI/图表 | Element Plus 2.14、ECharts 6、vue-echarts 8 | 现有门户和统计图表依赖 | 不重复引入同类 UI/图表库 |
| 状态/请求 | Pinia 3、Axios 1.17 | 登录态、预警状态和统一错误拦截 | 不在页面直接复制鉴权逻辑 |
| 后端 | Java 21、Spring Boot 3.2.5、Spring Security | 当前可运行的服务基线 | 不以根目录旧模块的 Spring Boot 4 配置作为事实 |
| ORM/迁移 | MyBatis-Plus 3.5.7、Flyway、MySQL 8 | 领域实体和可追踪迁移 | 不手工改已应用迁移 |
| 缓存/令牌 | Redis、JWT、BCrypt | Token 黑名单、缓存和无状态认证 | 不把 JWT 密钥写入源码或前端 |
| AI/外部数据 | 本地 Ollama 解读；Mock/School REST 数据源；Dify/E-Agent 前端外部通道 | 保留演示可用性并隔离外部依赖 | 不把密钥打包进浏览器 |
| 运行监控 | Spring Actuator、Micrometer Prometheus | 健康、指标和运行观测 | 不用日志打印替代审计和指标 |

## 四、模块与依赖边界

```text
Vue 页面 -> src/api/index.ts -> /api -> Spring Controller
Controller -> Domain Service -> MyBatis Mapper -> MySQL
Domain Service -> AcademicDataSource -> Mock 或 School REST
AcademicComputeService -> 原始数据表 -> GPA/画像/预警派生表
AgentChat -> Dify/E-Agent（当前前端外部通道）
AiController -> LocalLlmClient -> Ollama（后端指标解读通道）
```

| 模块 | 职责 | 可依赖 | 不得依赖/不得承担 |
|---|---|---|---|
| `auth`/`common/security` | 登录、JWT、黑名单、角色和统一错误 | user、Redis | 不在前端决定最终权限 |
| `student` | 学生档案、成绩、GPA、画像、健康报告 | mapper、compute、data source | 不读取其他角色的越权数据 |
| `teacher`/`course` | 班级和课程教师分析、学生详情 | student、alert、course mapper | 不绕过数据范围条件 |
| `department`/`dean` | 专业和全院聚合 | student、course、alert | 不把聚合接口当作明细越权入口 |
| `alert` | 预警查询、统计、确认、干预 | student、security、mapper | 不把确认和干预做成不可审计更新 |
| `development` | 路径分析、职位/院校缓存和目标分析 | student、缓存表 | 不承诺外部岗位/院校信息绝对准确 |
| `datasource`/`compute` | 外部数据适配、同步、派生计算和规则 | mapper、配置 | 业务服务不直接调用校方 API |
| `llm` | 本地模型指标解读 | 配置、HTTP client | 不保存或输出未脱敏敏感数据 |
| `academic-navigation-demo` | 页面、路由、API 封装和展示状态 | 后端公开契约 | 不臆造字段或把 mock 当生产数据源 |

## 五、关键架构约束

### 数据与事务

- 原始表与派生表分离；`AcademicComputeService` 负责 GPA 历史、五维画像和预警重算。
- Flyway 迁移按版本顺序执行；已应用迁移只新增后续版本，不直接修改历史文件。
- 同步写入必须可重试、幂等并可记录来源；跨外部系统动作失败时保留对账信息。
- 预警等级依据当前规则和业务数据计算；心理测评不得单独形成诊断或替代人工核查。

### 身份与安全

- JWT 无状态认证；Token 登出后进入 Redis 黑名单；密码使用 BCrypt strength 12。
- 服务端 SecurityConfig 做 URL 级角色校验，领域 Service 再按用户身份生成数据范围。
- 成绩、心理、体测和考勤只按最小角色范围返回；默认 fail-secure。
- JWT secret、数据库密码、校方 token、LLM 凭据只能来自环境或本地未提交配置。

### 部署与运行

- `application.yml` 提供非敏感默认项，`application-dev.yml`/环境变量覆盖数据库、JWT、校方和 LLM 参数。
- 健康端点：`/actuator/health`；指标端点：Actuator 暴露的 metrics/prometheus。
- 当前仓库没有已批准的 Docker、Nginx、CAS、RabbitMQ 或 WebSocket 生产方案；相关内容只能作为后续设计。

### API 与前端

- 前端统一通过 `src/api/index.ts` 调用 `/api`，Axios 拦截器处理 401/403/业务错误。
- 所有页面应覆盖加载、空数据、错误、无权限和重复提交；term 由后端统一解析。
- 后端 DTO 是契约来源之一，前端 `src/types/` 必须与其同步；禁止页面直接依赖 mapper/entity。

## 六、红线

| 红线 | 原则 | 反例 |
|---|---|---|
| 权限 | 服务端和 Service 双重校验 | 仅隐藏前端按钮 |
| 数据 | 破坏性操作需确认、审计、可恢复或批准 | 直接删除学生/预警数据 |
| 配置 | 敏感配置不入库、不入日志、不进前端包 | 提交 `application-dev.yml` 密码或 API key |
| 迁移 | 新增版本、可回滚/可恢复 | 修改已应用 V26 文件 |
| 外部调用 | 明确超时、重试、降级和来源 | 外部超时导致页面白屏 |
| AI | 外部模型输出需标注不确定性，危机信号转人工流程 | 承诺录取、就业或医学结论 |

## 七、架构决策记录

| 日期 | 决策 | 选择 | 否决方案 | 原因/证据 | 影响 |
|---|---|---|---|---|---|
| 2026-08-20 | 运行模块 | 以 `academic-navigation-backend` 为后端事实，历史 `academic-navigation` 模块的源代码与配置已清理 | 两套模块并列作为生产入口 | Vite 代理目标和完整 controller/迁移均位于 backend | 计划、验收和修复默认针对 backend |
| 2026-08-19 | 数据源边界 | 业务依赖 `AcademicDataSource`，默认 Mock，可切换 School REST | 业务服务直接访问校方 API | 当前代码已有接口、条件装配和同步 controller | 需补校方字段映射和生产凭据 |
| 2026-08-19 | 认证与数据范围 | JWT + Redis 黑名单 + URL/Service 双层校验 | 仅 session 或仅前端路由守卫 | `SecurityConfig`、`JwtAuthFilter`、`SecurityUtil` 已实现 | 需继续做真实越权回归 |
| 2026-08-19 | AI 通道 | 后端本地 LLM 解读与前端 Dify/E-Agent 通道分开记录 | 将所有 AI 调用混成一个后端服务 | 当前 `AiController` 与 `AgentChat.vue` 的实际代码路径不同 | 是否统一后端代理列为开放决策 |
