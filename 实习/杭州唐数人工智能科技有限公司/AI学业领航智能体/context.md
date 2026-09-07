# 学业导航系统 — 当前阶段交接文档

> 记录时间：2026-08-13
> 适用范围：本文档描述项目在当前阶段的代码实现现状，供下一步开发接手或评审使用。
> 说明：本文档已按 `academic-navigation-demo` 前端实际呈现状态全面更新（上一版记录的"后端只有认证、前端全 Mock"已过时）。

---

## 一、项目概览

**项目名称**：学业导航系统 / AI 学业领航系统（Academic Navigation，浙师大）
**定位**：面向高校的学业健康管理平台，集成多角色视图、AI 智能体对话、预警干预、发展引导等功能。
**技术架构**：前后端分离。前端 Vue 3 + TypeScript，后端 Spring Boot + MyBatis-Plus + MySQL + Flyway。

### 目录结构

```
academic-navigation-demo/      # 前端（Vue 3 + Vite）
academic-navigation-backend/   # 后端（Spring Boot，当前主开发目录）
docs/                          # 数据接入说明等
tools/sample-data/             # 样例数据集
outputs/                       # 运行日志、PID、db-export
```

---

## 二、前端现状（按实际呈现）

### 技术栈

- Vue 3.5 + TypeScript + Vite 8
- Element Plus 2.14（全量注册）+ @element-plus/icons-vue
- ECharts 6 + vue-echarts 8（图表）
- Pinia 3（状态管理）、Vue Router 4（路由）
- Axios（请求封装，`baseURL: '/api'`）
- Tailwind CSS 4（`@tailwindcss/vite` 插件）+ 组件内 Scoped CSS 混用

### 路由结构（5 个角色，各自独立 Layout）

| 路径前缀 | 角色 | Layout | 默认落地页 |
|---|---|---|---|
| `/student` | 学生 | StudentLayout.vue | `/student/agents`（AI 领航助手） |
| `/teacher` | 班主任 | TeacherLayout.vue | `/teacher/dashboard`（班级驾驶舱） |
| `/course-teacher` | 任课教师 | CourseTeacherLayout.vue | `/course-teacher/dashboard`（我的课程） |
| `/department` | 系主任 | DepartmentLayout.vue | `/department/overview`（专业态势） |
| `/dean` | 院领导 | DeanLayout.vue | `/dean/dashboard`（全院仪表盘） |

**路由守卫（已实现）**：`router.beforeEach` 全局守卫，未登录访问业务页一律重定向 `/login`；刷新后通过 `auth.restore()` 调 `/auth/me` 恢复登录态。

### 页面清单（views/，共 14 个业务视图 + 3 个共享视图 + 1 个登录页）

**学生端**
- `Dashboard.vue` — 学业总览（GPA、排名、学分、预警数、健康分、GPA 趋势图、AI 周报）
- `Profile.vue` — 学业画像（五维雷达图 + 趋势）
- `Alerts.vue` — 预警通知列表（分页）
- `DevelopmentPath.vue` — 发展引导（考研/留学/考公考编/就业路径，含职位库、院校库、目标分析）
- `QuickAccess.vue` — 快捷入口

**班主任端**
- `Dashboard.vue` — 班级驾驶舱（KPI、预警分布、维度均值、预警列表）
- `GpaProgress.vue` — GPA 进退情况（多学生趋势对比）
- `CourseStats.vue` — 课程成绩统计（通过率、挂科名单）
- `AlertManagement.vue` — 预警管理（筛选 + 确认 + 干预记录表单）
- `StudentDetail.vue` — 学生详情（`/teacher/student/:id`）

**任课教师端**
- `Dashboard.vue` — 我的课程（课程列表、成绩分布、预警学生）
- `Alerts.vue` — 学情预警（课程维度，分页）

**系主任端**
- `Overview.vue` — 专业态势（KPI、各年级 GPA 趋势、重点关注学生）
- `Courses.vue` — 课程分析（热力图 + 课程健康度表）
- `Alerts.vue` — 预警管理（按年级/等级筛选，分页）

**院领导端**
- `Dashboard.vue` — 全院仪表盘（KPI、专业排名、预警分布）
- `Alerts.vue` — 全院预警汇总（按专业/年级分组 + 趋势）

**共享页面**
- `AgentPlaza.vue` — AI 智能体广场（按角色过滤、搜索、分类 Tab）
- `AgentChat.vue` — 智能体对话页（支持 Dify / E-Agent / Mock 三种模式）
- `QuickAccess.vue` — 快捷入口（静态配置，无后端依赖）

### 状态管理（Pinia）

- `stores/auth.ts`：登录态。`login()` 调 `POST /api/auth/login`，持久化 token+user 到 localStorage；`logout()` 调 `/api/auth/logout` 并清本地；`restore()` 调 `/api/auth/me` 恢复会话；按 role 通过 `routeMap` 跳转。
- `stores/alerts.ts`：预警未读数（`GET /api/alerts/unread-count`），供布局角标刷新。

### 请求封装

`utils/request.ts`：Axios 实例，`baseURL: '/api'`、超时 10s；请求拦截器自动附加 `Authorization: Bearer {token}`；响应拦截器处理 `code !== 200` 统一报错、401 清 token 跳登录、403 无权限提示。

### API 层（src/api/index.ts）

已封装约 30 个接口，按角色/模块分组：

| 模块 | 接口 |
|---|---|
| 学生端 | `/student/dashboard`、`/student/profile`、`/student/gpa-trend`、`/student/alerts` |
| 班主任端 | `/teacher/dashboard`、`/teacher/gpa-progress`、`/teacher/course-stats`、`/teacher/students`（分页）、`/teacher/students/{id}` |
| 任课教师端 | `/course-teacher/courses`、`/course-teacher/alerts` |
| 预警模块 | `/alerts`（筛选分页）、`/alerts/{id}`、`/alerts/unread-count`、`/alerts/{id}/confirm`、`/alerts/{id}/intervention`、`/alerts/stats` |
| 系主任端 | `/department/overview`、`/department/courses/heatmap`、`/department/courses`、`/department/plan-analysis`、`/department/alerts` |
| 院领导端 | `/dean/dashboard`、`/dean/alerts`、`/dean/alerts/trend` |
| 发展引导 | `/development/paths`、`/development/jobs`、`/development/schools`、`/development/analyze` |
| 认证 | `/auth/login`、`/auth/logout`、`/auth/me` |

### AI 智能体（src/mock/agents.ts，14 个）

按角色分组（`roles` 字段控制可见性）：

| 分组 | 智能体 |
|---|---|
| 学生（8 个） | AI 校策知询（政策权益）、AI 人生规划（核心服务）、AI 学业顾问（核心服务）、AI 就业助手（就业发展）、AI 考研助手（升学规划）、AI 留学助手（升学规划）、AI 考编助手（考公考编）、AI 考公助手（考公考编） |
| 班主任 + 任课教师共享（2 个） | 学情分析（班级学情）、育心导学（教育引导） |
| 系主任（2 个） | 学情分析（学情管理）、育心导学（教育引导） |
| 院领导（2 个） | AI 课程健康度诊断（学情管理）、AI 跨专业对比分析（质量监控） |

**对话实现（AgentChat.vue）**：每个智能体配置 `externalApi`（type: `dify` | `eagent` + baseUrl + assistantId）。实际对话按优先级走：
1. Dify 流式 API（SSE，`/v1/chat-messages`，支持多轮 `conversation_id` 与 `<think>` 思考块过滤）；
2. E-Agent API（`/api/v2/assistant/chat/completions`）；
3. 未配置外部 API 或调用失败时，回退 `getMockResponse()` 本地兜底回复。

智能体**列表与配置为前端静态数据**，不经过后端接口。

### Mock 数据现状

| 文件 | 状态 |
|---|---|
| `mock/agents.ts` | ✅ 仍在使用：智能体配置 + Mock 对话兜底 |
旧版 mock 数据文件和 GPA 沙盘预测算法已清理；当前发展路径数据统一来自后端 `/development/*`。

---

## 三、后端现状

### 技术栈

- Spring Boot 3.2.5（Java 21）+ Spring Security（JWT 无状态认证）
- MyBatis-Plus 3.5.7 + MySQL + Flyway（迁移已到 V7）
- SpringDoc（Swagger UI）、Lombok、JJWT 0.12

### 已实现的业务模块（9 个 Controller）

| 模块 | Controller | 说明 |
|---|---|---|
| 认证 | AuthController | 登录 / 登出 / 当前用户，JWT + 黑名单 |
| 学生 | StudentController | Dashboard / 画像 / GPA 趋势 / 预警分页 |
| 班主任 | TeacherController | 班级驾驶舱 / GPA 进度 / 课程统计 / 学生列表与详情 |
| 任课教师 | CourseTeacherController | 我的课程 / 课程预警 |
| 系主任 | DepartmentController | 专业态势 / 课程热力图 / 培养方案分析 / 预警 |
| 院领导 | DeanController | 全院仪表盘 / 预警汇总 / 趋势 |
| 预警 | AlertController | 预警列表 / 详情 / 未读数 / 确认 / 干预回填 / 统计 |
| 发展引导 | DevelopmentController | 路径 / 职位库 / 院校库 / 目标差距分析 |
| 校方数据同步 | SchoolSyncController | 预留的校方数据同步入口 |

### 数据库迁移（V1–V7）

| 版本 | 内容 |
|---|---|
| V1 | `t_user` 基础账号表 + 4 个初始账号 |
| V2 | `course_teacher` 角色账号补充 |
| V3 | 核心业务表：学院/专业/教师/学生/班级/课程/开课/成绩/GPA 历史/画像/体测/预警/干预 + 种子数据 |
| V4 | M3 模块：职位缓存、院校缓存、培养方案表 + 多年级多专业演示数据 |
| V5 | 补齐系主任/院长多学院账号 |
| V6 | 数据集扩充（多学院、多专业、多班级、学生成绩、预警记录） |
| V7 | 补齐旧五维画像缺失维度（心理/体质/思想品德） |
| V8 | 统一五维画像维度为学业成绩/实践能力/综合素质/人文素养/身心健康，并按固定顺序返回 |

**测试账号**：密码统一 `123456`（V1–V5 种子，含学生/班主任/任课教师/系主任/院长各角色，V5 起按专业/学院细分了多个系主任与院长账号）。

### 数据源抽象

- `AcademicDataSource` 接口（`getGpa` / `getRank` / `getCredits`）
- `MockAcademicDataSource`（`@Primary`，当前激活，返回演示数据）
- `SchoolApiDataSource` + `SchoolApiClient` + `SchoolApiSyncService/Runner`：校方 API 接入预留（需校方提供接口文档后启用）

### 已知运行注意

- `TokenBlacklistService`：优先 Redis，Redis 不可用时降级内存黑名单并打 WARN（当前环境 Redis 未启动，属预期降级，非错误）。

---

## 四、前后端对接现状

| 范围 | 状态 |
|---|---|
| 认证（登录/登出/me） | ✅ 已对接真实后端 |
| 全部 14 个业务视图 | ✅ 已对接真实后端 API（见上方 API 层清单） |
| 智能体广场/对话 | 🟡 前端静态配置 + Dify/E-Agent 直连 + Mock 兜底，未走后端 `/api/agents` |
| 快捷访问 | ✅ 纯前端静态配置 |
| Mock data.ts / jobs.ts / schools.ts | ⚠️ 已弃用（无视图引用） |

---

## 五、当前阶段主要缺口

| 缺口 | 说明 |
|---|---|
| 智能体对话需真实密钥 | Dify/E-Agent 的 baseUrl/assistantId 为演示配置，需接入真实服务与密钥后可用；无配置时回退 Mock |
| 智能体列表未走后端 | `agents.ts` 为前端静态配置，集中管理/灰度/统计（usageCount）仍缺失 |
| Redis 未启用 | JWT 黑名单降级为内存实现，重启失效；生产需启动 Redis |
| 校方数据源未启用 | 当前为 `MockAcademicDataSource` + 本地种子库，校方 API 接入待接口文档 |
| GPA 沙盘页面已移除 | 无运行时页面或预测算法；如需恢复，应基于当前数据契约重新实现 |
| 历史后端与旧版 mock 文件 | 已从工作区清理；当前运行链只保留 `academic-navigation-backend` 与 `academic-navigation-demo` |

---

## 六、本地启动方式

**后端**
1. MySQL 启动并创建库（配置见 `academic-navigation-backend/src/main/resources/application-dev.yml`）
2. Flyway 自动执行 V1–V7 迁移并写入种子数据
3. `mvn spring-boot:run`（默认 8080 端口）
4. Swagger UI：`http://localhost:8080/swagger-ui.html`

**前端**
1. `cd academic-navigation-demo && npm install`
2. `npm run dev`（Vite 5173 端口，`/api` 代理到 `localhost:8080`）
3. 登录使用测试账号（如学生 `2022001` / `123456`；任课教师 `C20180099`）

---

## 七、相关文档索引

- `项目进度说明.md` — 功能架构与页面级说明（已同步更新）
- `docs/数据接入说明.md` — 表结构与校方数据接入方案
- `后端架构设计.md` / `后端设置规划.md` — 后端设计文档
- 招标相关：`学业领航智能体系统-详细功能及技术参数（招标版）.docx`、`build_bid_doc.py`
