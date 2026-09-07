# TFA 多租户科研 Agent 平台 — 后端设计

> 状态：v0.1 草案（本会话编写，待用户 review）
> 日期：2026-08-21
> 关联：`docs/specs/tfa-multi-tenant-platform-spec.md`（需求）· `tasks/plan.md`（计划与决策）· `tasks/todo.md`（任务清单）
> 范围：Phase 0 已落地部分的设计回顾 + Phase 1~3 后端（T4~T8b、T10、T11）的完整设计

## 1. 目标与范围

把现有 TFA shell（单用户本地）改造成多用户 Web 平台。本文档定义**后端**的：

- 分层架构与文件规划（与现有 `packages/shell/src/server` 风格一致）
- 权限模型（ACL 核心规则 + 存储层强制点）
- 全部 REST API 契约（按任务组织，供前端并行开发）
- 存储布局（技能/数据集落盘位置）
- 执行层设计（Docker 每会话 + 会话桥接）
- 错误契约、校验、测试计划

非后端范围：前端页面（仅给 API 契约）、部署（T13）、二期功能。

## 2. 现状盘点（Phase 0 已落地）

| 模块 | 文件 | 状态 |
|---|---|---|
| 平台库 schema + 迁移 | `src/server/db/migrations/001_platform.sql`、`002_auth_sessions.sql` | ✅ 8 表 + auth_sessions，幂等迁移 |
| Repository 层 | `src/server/db/repos.ts`（`PlatformRepos`） | ✅ 全部实体 CRUD |
| 认证 | `src/server/auth.ts`（scrypt、token 哈希、`AuthError`、`authenticateRequest`） | ✅ |
| 用户/组/项目/资源/会话服务 | `src/server/session-service.ts`（`ShellSessionService`） | 部分：认证 + 用户管理 + 会话管理已落地 |
| 项目存储 | `src/server/project-store.ts`（`ProjectStore`，DB 化 + 旧数据导入） | ✅ 但 `ProjectInfo` 无 scope/ownerId |
| HTTP 层 | `src/server/http-server.ts`（Hono，REST + SSE + 静态托管） | ✅ 认证/用户/会话/项目路由已挂 |
| 测试 | `test/auth.test.ts`、`test/platform-db.test.ts` 等 | ✅ shell 全量 39/39 |

**设计约束（写死，不改）**：

- 栈：Hono + node:sqlite（`DatabaseSync`），无 ORM；`PlatformRepos` 同步 CRUD
- 错误统一 `{ error: { code, message } }`，`app.onError` 映射：AuthError→401/403、TfaServerError→400/404/409、其余→500
- 认证中间件：`authenticateRequest(c, service)` 解析 `Authorization: Bearer <token>`
- 时间戳毫秒 INTEGER，id 一律 TEXT(uuid)
- 服务层为业务规则所在，路由层薄；现有 `ShellSessionService` 承载会话 + 认证，新模块逐步拆分

## 3. 总体架构

```
                    ┌────────────────────────── 控制面（单进程，Hono） ──────────────────────────┐
  Vue3 前端  ──────▶ │  http-server.ts 路由层（薄：解析/鉴权/转发）                                   │
  (Web/SSE)  ──────▶ │  ├─ auth.ts             认证中间件 + 角色助手                                  │
                    │  ├─ permissions.ts       权限助手（新增）：requireRole / isGroupMember / ...    │
                    │  ├─ groups.ts            课程组服务（T4 新增）                                  │
                    │  ├─ projects.ts          项目服务/ACL（T5 新增）                                │
                    │  ├─ skills.ts            技能库服务（T6 新增）                                  │
                    │  ├─ datasets.ts          数据集服务（T7 新增）                                  │
                    │  ├─ session-service.ts   会话管理（既有）+ 桥接接入（T8b）                      │
                    │  └─ executor/            SessionExecutor 抽象（T8a 新增）                       │
                    │       ├─ types.ts / docker.ts / local.ts                                      │
                    │  └─ db/                  PlatformRepos + 迁移（既有，按需扩展）                  │
                    └───────────────┬───────────────────────────────────────────────────┘
                                    │ 每会话一个容器（执行面）
                             ┌──────▼──────┐     只读挂载      ┌──────────┐
                             │ Docker 容器  │ ◀─────────────── │ 数据集存储 │
                             │ tfa CLI      │     --skill 注入  └──────────┘
                             │ (rpc-entry)  │ ◀─────────────── │ 技能存储   │
                             └──────┬──────┘     项目目录读写   └──────────┘
                                    │ RPC/SSE（T8b 桥接）
```

依赖方向：路由 → 服务 → repos/executor → db。控制面/执行面分离：agent 只跑在容器里，控制面不直接持有 AgentSession（T8b 完成后；宿主直连仅作为过渡）。

## 4. 分层与文件规划（按任务）

| 任务 | 新增/修改文件 | 说明 |
|---|---|---|
| T4 课程组 | 新增 `src/server/groups.ts`；改 `http-server.ts`、`db/repos.ts`（组查询补 memberCount 等）、`db/types.ts`；测试 `test/groups.test.ts` | 服务层含 ACL |
| T5 项目作用域 | 改 `project-store.ts`（`ProjectInfo` 加 `ownerId`/`scope`）、`session-service.ts`、`http-server.ts`；测试 `test/projects-scope.test.ts` | 存储 + 可见性查询 |
| T6 技能库 | 新增 `src/server/skills.ts`（含 git clone 助手）；改 `http-server.ts`、`repos.ts`；测试 `test/skills.test.ts` | 本地裸仓库 fixture 模拟 git |
| T7 数据集库 | 新增 `src/server/datasets.ts`（含上传/配额）；改 `http-server.ts`、`repos.ts`；测试 `test/datasets.test.ts` | multipart 上传 |
| T8a 执行器 | 新增 `src/server/executor/{types,docker,local}.ts`；测试 `test/executor.test.ts` | docker 命令构造可单测（mock spawn） |
| T8b 桥接 | 改 `session-service.ts`（接入 executor）、`http-server.ts`（SSE 透传）；新增 `ProxyRuntime`；测试 `test/session-bridge.test.ts` | 先宿主直连再容器化 |
| T10 管理端 | 新增 `src/server/admin.ts`（组/资源治理聚合）；改 `http-server.ts`；测试 `test/admin.test.ts` | 用户管理已有 |
| T11 会话监控 | 改 `http-server.ts`、`executor/`；测试 `test/sessions-admin.test.ts` | 列表 + 终止 |

权限助手统一放 `src/server/permissions.ts`，避免每个 service 重复判断。


## 5. 权限模型（ACL 核心）

### 5.1 可见性规则（存储层强制，不只 UI 过滤）

**技能/数据集可见集合**（对用户 U）：

```
visible = status = 'active' AND (
    visibility = 'public'
    OR (visibility = 'private' AND owner_id = U)
    OR (visibility = 'group' AND group_id IN (U 所在且未归档的组))
)
```

SQL 落在 `PlatformRepos` 新增方法（如 `skillListVisible(userId)`、`datasetListVisible(userId)`），所有 API 只能走可见性查询，杜绝「查出全部再过滤」。

**项目可见集合**（对用户 U）：

```
visible = owner_id = U
       OR (scope_type = 'group' AND group_id IN (U 所在的组))   // 含归档组:历史项目对成员保留可见(只读)
       OR U.role = 'admin'
```

**组管理权限**：

| 操作 | 组 owner | admin | 组 member | 其他 |
|---|---|---|---|---|
| 看组/组成员 | ✅ | ✅ | ✅ | ❌ |
| 改名 | ✅ | ✅ | ❌ | ❌ |
| 归档 | ✅ | ✅ | ❌ | ❌ |
| 加/移成员、任免组长 | ✅（本组） | ✅（任意组） | ❌ | ❌ |
| 建组 | ✅（老师） | ✅ | ❌（学生 403） | — |

### 5.2 归档（archived）语义

- 归档组：不可加新成员、不可新建挂组项目
- 组内资源与项目**保留**；现有成员仍可见组库，但**只读**（编辑/移除组内资源仅资源 owner 或 admin 可做）
- 组 owner 变更是唯一改 `course_groups.owner_id` 的路径（任免组长时同步 `group_members.role`）

### 5.3 权限助手（`src/server/permissions.ts`）

```ts
requireRole(user, roles: UserRole[]): void            // 不满足抛 AuthError("forbidden")
isGroupMember(repos, userId, groupId): boolean        // 含 owner，且组未归档（归档仍算成员以保留只读）
canManageGroup(repos, user, group): boolean           // group.ownerId === user.id || user.role === 'admin'
canAccessProject(repos, user, project): boolean       // 5.1 项目规则
canEditResource(user, resource): boolean              // resource.ownerId === user.id || admin
```

## 6. REST API 契约（按任务）

> 约定：除 `POST /api/auth/login` 与 `GET /api/health` 外，全部需要 `Authorization: Bearer <token>`。
> 响应错误统一 `{ error: { code, message } }`；状态：401 未登录/失效、403 无权限、404 不存在、400 参数错/归档组操作/重复添加（沿用既有惯例，TfaServerError 无 conflict 码）、413 超限。

### 6.1 认证与用户（已落地，列出以供前端对齐）

```
POST   /api/auth/login            { username, password } → { token, user, expiresAt }
POST   /api/auth/logout
GET    /api/auth/me               → { user }
GET    /api/admin/users           (admin) → { users }
POST   /api/admin/users           (admin) { username, password, role } → 201 { user }
PATCH  /api/admin/users/:id       (admin) { status?, password? }
```

### 6.2 课程组（T4，`groups.ts`）

```
GET    /api/groups                      → { groups: GroupView[] }
POST   /api/groups                      (teacher/admin) { name, type: course|research } → 201 { group }
GET    /api/groups/:id                  (组内/admin) → { group: GroupDetail }
PATCH  /api/groups/:id                  (owner/admin) { name? }
DELETE /api/groups/:id                  (owner/admin) → 归档（status=archived）
POST   /api/groups/:id/members          (owner/admin，组须 active) { userId } → 201
PATCH  /api/groups/:id/members/:userId  (owner/admin) { role: owner|member } → 任免组长
DELETE /api/groups/:id/members/:userId  (owner/admin；不可移除 owner) → 移除成员
```

- `GroupView = { id, name, type, status, ownerId, memberCount, myRole: owner|member|null }`
- `GroupDetail = GroupView & { members: [{ userId, username, role }] }`
- 列表语义：返回「我是 owner 或 member」的组（含 archived，标注 status）；admin 返回全部
- 归档组：POST members → 400；任免组长在 archived 组允许（恢复治理能力），建项目被 T5 拦截
- 移除成员：项目保留，但成员立即失去组库可见性（存储层查询天然生效）

### 6.3 项目（T5，`project-store.ts` + `projects.ts`）

`ProjectView`（扩展现有 `ProjectInfo`）：

```ts
interface ProjectView {
  id: string; name: string; folder?: string;
  ownerId: string;
  scope: { type: "private" } | { type: "group"; groupId: string; groupName: string };
  createdAt: number; updatedAt: number; sessionIds: string[];
}
```

```
GET    /api/projects                    → 当前用户可见项目（5.1）
POST   /api/projects                    { name, folder?, scope?: { type:'private' } | { type:'group', groupId } } → 201
PATCH  /api/projects/:id                (owner/admin) { name?, folder? }   // scope 不可改
DELETE /api/projects/:id                (owner/admin) → 解除会话链接并终止其 running 会话
POST   /api/sessions/:id/project        （已有路由）加 ACL：仅项目可见用户可挂载
```

- scope 缺省 = private；scope=group 须为组 active 成员（**A4 默认允许**，含学生）；归档组 → 400
- scope 不可改：避免组库可见性漂移（改作用域 = 删除重建，设计决策）

#> **T5 实现备注（2026-08-21）**：已按本契约落地，其中「归档组历史项目对现有成员仍可见（只读）」为当日决策调整（原 §5.1 要求 active 组）；会话路由（`/api/sessions*`）暂未加鉴权，会话归属与 `POST /api/sessions` 的 projectId 用户校验随 T8b 补齐。
> **T9 补充（2026-08-21）**：新增 `GET /api/projects/:id/session-context` → `{ skills, datasets }`（该项目会话可见资源：私有 ∪ 挂组则该组库 ∪ 公开；不可见项目 404）。前端据此展示/默认注入；实际进程注入在会话路径切桥接后生效。
> **T12 实现备注（2026-08-24，spec v0.2 项目工作台）**：项目文件工作空间已落地（`src/server/project-files.ts`）。工作目录 = 平台托管 `{agentDir}/workspaces/{projectId}/`（决策 A）；ACL 复用 `ProjectStore.getForUser`（不可见 → 404）；`safeResolve` 强制路径落在工作区内（防 `../` 与绝对路径）；接口：目录树 / 文本读写 / multipart 上传 / 下载 / 删除；无大小上限（需求确认）。
> **T13 实现备注（2026-08-24）**：三栏项目工作台已落地（前端 `ProjectWorkbench.vue` + `MonacoEditor.vue` 懒加载）。后端配套：项目会话 cwd 指向 `{agentDir}/workspaces/{projectId}`（与 T12 文件 API 同根，Agent 可在工作区读写文件）。文本/代码 Monaco 编辑保存、PDF/图片/CSV 预览、右侧精简对话面板；通用会话保持全屏聊天。手动走查待用户本地验证。
> **T14 实现备注（2026-08-24，会话权限模式）**：`sessions.permission_mode`（004 迁移）持久化；创建默认项目会话 `request_approval`、通用 `full_access`；`GET/POST /api/sessions/:id/permission-mode` + 批准状态机（`ApprovalService`）与 `approvals` API，均需登录 + 会话归属校验（创建时带 token 记录 userId）。executor 隔离参数（工作区挂载 + `--no-skills --skill` + 数据集只读）不因权限模式减弱（FakeRunner 验证）。执行面拦截（request_approval 挂起工具调用）待 T8b 桥接接入。
> **T16 实现备注（2026-08-24，智能体广场 FR-9）**：`005_agent_listings.sql` + `agents.ts` AgentService（三层可见性存储层强制、公开直发不审核 A7、owner/admin 编辑下架、URL http(s) 校验、availability 字段默认 unknown）；API `GET/POST /api/agents`、`PATCH/DELETE /api/agents/:id`；前端 AgentPlaza + 导航入口。T17 健康检查填充 availability/lastCheckedAt。
> **T17 实现备注（2026-08-24，智能体健康检查 FR-9）**：`agent-health.ts` AgentHealthChecker（HEAD + `AbortSignal.timeout(5s)`、有界并发 3；2xx/3xx 在线，其余不可用）；`AgentService.runHealthCheck` 持久化 availability/lastCheckedAt；`ShellSessionService.startAgentHealthCheck(5min)`（close 清理），server-entry 启动时调用。参数可注入（fetchImpl/超时/并发）。前端 AgentPlaza 展示徽标与检测时间，失败不删卡片。
## 6.4 技能库（T6，`skills.ts`）

```
GET    /api/skills                          → 当前用户可见技能（5.1）
POST   /api/skills                          { name?, repoUrl?, visibility, groupId? } | { name, visibility, groupId?, upload: true }
POST   /api/skills/:id/update               (owner/admin) → git pull --ff-only（无 repoUrl → 400）
PATCH  /api/skills/:id                      (owner/admin) { name?, visibility?, groupId? }
DELETE /api/skills/:id                      (owner/admin) → status=removed（软删）
POST   /api/admin/skills/:id/review         (admin) { reviewState: approved|rejected }
```

- `SkillView = { id, name, repoUrl, version, visibility, groupId?, groupName?, ownerId, path, reviewState, createdAt }`
- visibility=group 必须带 groupId 且我是组 active 成员；visibility=public → `reviewState=pending`（A7 先上线后审核）
- 改 visibility→public：owner 可改但置 pending；admin 可改任意
- repoUrl 安装：`git clone --depth 1 <url> <storage>/skills/<id>`；update = `git -C <path> pull --ff-only`；上传型（upload:true）无 repoUrl，不可 update

#> **T6 实现备注（2026-08-21）**：已按本契约落地。可见性查询（`skillListVisible`）含归档组成员（组库对归档成员只读，与 T5 项目决策一致）；上传型仅建记录+空目录，文件上传接口留给前端阶段；`path` 列存绝对路径（便于执行器挂载），不出现在 SkillView。
## 6.5 数据集库（T7，`datasets.ts`）

```
GET    /api/datasets                    → 当前用户可见数据集（5.1）
POST   /api/datasets                    multipart { file, name?, visibility, groupId? } → 201
PATCH  /api/datasets/:id                (owner/admin) { name?, visibility?, groupId? }
DELETE /api/datasets/:id                (owner/admin) → status=removed
GET    /api/datasets/:id/download       (可见用户，只读；可选，MVP 可省)
```

- 限额（**A5 默认值**）：单文件 ≤500MB（413 `quota_exceeded`）；每用户累计 ≤20GB（上传时校验）
- 存储：`<storage>/datasets/<id>/<安全文件名>`；`sizeBytes` 落库
- 组内只读：组 member 可下载/被挂载，不可编辑（编辑仅 owner/admin）

#> **T7/T10 实现备注（2026-08-21）**：已按本契约落地。配额经 `QuotaExceededError` → 413（新增 `errors.ts`，TfaServerError 无 quota 码）；上传为 multipart（`POST /api/datasets`，`file/name/visibility/groupId` 字段）；`GET /api/datasets/:id/download` 对可见用户只读提供；存储 `{agentDir}/platform/datasets/{id}/{安全文件名}`，`path` 列存绝对路径（供执行器只读挂载）。**T10 起数据集与技能一致支持审核**（`003_datasets_review.sql` 迁移 + `POST /api/admin/datasets/:id/review`，公开创建/改公开 → pending）。
## 6.6 会话执行与监控（T8a/T8b/T11）

**既有路由保持兼容**，`POST /api/sessions` 扩展 `{ projectId? }`（须为当前用户可见项目；启动时注入该项目可见 skills + 数据集只读挂载）：

```
POST   /api/sessions                       { projectId? } → 201 { session }
GET    /api/sessions                       → 当前用户会话
GET    /api/sessions/:id                   → 会话快照
GET    /api/sessions/:id/events            SSE（桥接容器内 tfa 事件）
POST   /api/sessions/:id/prompt|steer|abort|model|thinking|rename
DELETE /api/sessions/:id                   → 销毁容器 + 删除会话文件
GET    /api/sessions/:id/export

GET    /api/admin/sessions                 (admin) → 全部会话 + 容器状态（T11）
POST   /api/admin/sessions/:id/terminate   (admin) → 终止容器并置 ended（T11）
```

#> **T10 实现备注（2026-08-21）**：后端治理能力落地：用户（T3）、组归档/任免（T4）、技能审核+下架（T6）、数据集审核+下架（T7/T10）。未新增冗余 admin 聚合模块——各领域服务的 admin 分支已覆盖全部治理路径；前端 AdminPanel.vue 属 Phase 1 UI 待补。会话监控（列表+终止）归 T11。
> **T11 实现备注（2026-08-21）**：会话监控已落地。平台 `sessions` 表为数据源（会话创建/恢复落表、删除/终止置 ended）；`GET /api/admin/sessions` 返回全量会话 + 用户/项目名 + 容器状态（executor inspect）；`POST /api/admin/sessions/:id/terminate` 停容器（docker stop/rm）+ 销毁 live/rpc 会话 + 置 ended。执行器经 `ShellSessionServiceOptions.executor` 可注入（测试用 FakeRunner）。前端管理页待补。
## 6.7 管理端（T10）汇总

- 用户：6.1 已有
- 组：6.2 已含 admin 归档任意组、任免组长
- 资源治理：6.4 review、admin 可 DELETE 任意 skill/dataset（下架立即生效）
- 会话：6.6 admin sessions/terminate

## 7. 存储布局

```
{agentDir}/platform/
  skills/{id}/        # 每个技能一个目录（git 仓库或上传目录）；path 列存相对路径
  datasets/{id}/      # 每个数据集一个文件/目录
  trash/              # 下架资源移入（可选，MVP 可保留原地仅标记 removed）
```

**设计决策（对 plan 决策 4 的细化）**：可见性强制以 DB ACL 为准，目录按资源 id 平铺，不按三层分目录——避免「改可见性=搬目录」的迁移成本与竞态。如需纵深防御，二期可再加目录分层（控制面先保证 SQL 只返回可见资源）。

## 8. 执行层设计（T8a/T8b）

### 8.1 执行器接口（`src/server/executor/types.ts`）

```ts
interface SessionExecutor {
  start(opts: ExecutorStartOptions): Promise<ExecutorHandle>;
  stop(id: string): Promise<void>;                 // docker stop --time / kill
  inspect(id: string): Promise<ExecutorStatus>;    // 容器状态（admin 监控）
  list(): Promise<ExecutorStatus[]>;
}
interface ExecutorStartOptions {
  sessionId: string;
  project: ProjectView;
  skills: string[];           // 可见技能目录（容器内路径）
  datasets: MountSpec[];      // { hostPath, containerPath, readonly: true }
  env: Record<string, string>; // OLLAMA_BASE_URL / OLLAMA_API_KEY 或 BYOK key
  limits: { memoryBytes: number; cpuCount: number; timeoutMs: number };
  cwd: string;                // 项目目录（容器内 /work）
}
```

### 8.2 实现

- **`LocalExecutor`（dev/test 过渡）**：宿主直接 spawn tfa（或复用现有 in-process AgentSession 路径），用于 T8b 先跑通「选项目→注入技能/数据→聊天」，再切容器
- **`DockerExecutor`（生产）**：镜像内装 tfa CLI（T13 出 Dockerfile）
  - `docker run -d --name tfa-<sessionId> --memory 2g --cpus 2`
  - 挂载：项目目录 `/work`（读写，会话产出落此）、数据集 `<storage>/datasets/<id>` → `/data/<id>:ro`
  - 注入：`tfa --no-skills --skill <可见技能目录...>`（**A1 结论：必须 `--no-skills` + 显式 `--skill`，禁用默认全局技能目录，否则跨会话泄漏**）
  - env：Ollama 端点 / BYOK key（仅注入本人容器，不落共享磁盘——spec NFR）
  - 超时：服务端定时器扫描 running 会话，超时 `docker stop --time 30` + `sessions.status=ended`
  - 限额可配：`EXECUTOR_MEMORY=2g`、`EXECUTOR_CPUS=2`、`SESSION_TIMEOUT_MS=1800000`

> **T8a 实现备注（2026-08-21）**：执行器层已落地（`src/server/executor/`）。命令执行经可注入 `CommandRunner`，单测用 FakeRunner 校验构造（不依赖真实 Docker）；`opts.limits.timeoutMs` 按会话设超时，`pruneExpired` 供服务层定时器；并发上限默认 2/用户（`MAX_CONCURRENT_SESSIONS_PER_USER`）。`LocalExecutor` 为 T8b 过渡占位（宿主 spawn+RPC 在 T8b 接入）。http-server 起停端点随 T8b 桥接落地；容器实机冒烟待 Docker 环境（T13）。
### 8.3 会话桥接（T8b）

- 容器内 tfa 以 RPC 子进程模式启动（`packages/coding-agent/src/rpc-entry.ts`）；控制面新增 `ProxyRuntime implements TfaSessionRuntime`，把 prompt/steer/abort/setModel 序列化为 RPC 消息经容器通道转发，事件回传由现有 `GET /api/sessions/:id/events` SSE 透出
- 会话持久化：会话文件写容器内 `/work`（= 宿主项目目录），容器销毁后数据保留（T8b 验收「关页面→重开会话恢复」）
- 先宿主直连（LocalExecutor）再容器化（DockerExecutor），复用现有 RPC/SSE 协议，风险可控


> **T8b 实现备注（2026-08-21）**：桥接层已落地（`src/server/bridge.ts`）：`ProxyRuntime` 复用 `@earendil-works/tfa-coding-agent` 的 `RpcClient`（spawn `node <cli> --mode rpc` + JSONL），转发 prompt/steer/abort/setModel/setThinking，RPC 事件映射为 snapshot 事件（流式 progress 精化留待 T9）；快照由 `get_state`+`get_messages` 构建。`ShellSessionService.createRpcSession()` 提供宿主直连桥接会话（默认会话路径仍 in-process，不破坏现有流程）。剩余：会话路径整体切桥接、恢复（switch_session）、容器化挂 executor——待 T9 与 Docker 环境（T13）。
## 9. 错误处理与校验

- 保持 `app.onError` 现有映射；扩展状态码：`quota_exceeded`→413（T6/T7 上传超限）。重复添加/重名冲突沿用既有惯例 400（`TfaServerError` 的 code 类型仅 busy/session_locked/not_found/invalid_request/not_implemented，无 conflict）
- 新增错误码枚举放 `src/server/errors.ts`（或并入 auth.ts 的 `AuthError` 模式），service 层抛、路由层只透传
- 输入校验与现有风格一致：service 层手写，先 trim；长度/枚举/必填在 service 收口，路由层只做类型粗检
- 文件上传错误：超限 413、非法文件名 400（白名单字符）、存储写失败 500

## 10. 测试计划

> 约定：测试用仓库根 `./test.sh`；单测可 `cd packages/shell && node ../../node_modules/vitest/dist/cli.js --run`。

| 任务 | 测试文件 | 覆盖点 |
|---|---|---|
| T4 | `test/groups.test.ts` | 建组（teacher/admin）、学生建组 403、加/移成员、重复加 409、任免组长、归档后不可加人/建项目、非本组老师 403 |
| T5 | `test/projects-scope.test.ts` | 私有/挂组可见性（5.1）、学生建挂组项目（A4）、归档组拒新建、scope 不可改、非 owner PATCH 403 |
| T6 | `test/skills.test.ts` | git clone 安装（本地裸仓库 fixture）、上传型、可见性三态、public→pending、update、下架后消失、admin review |
| T7 | `test/datasets.test.ts` | 上传/大小上限 413/配额 20GB、可见性、组内只读、下架 |
| T8a | `test/executor.test.ts` | docker 命令构造（mock spawn）、限额参数、超时扫描、stop/terminate、错误清理 |
| T8b | `test/session-bridge.test.ts` | ProxyRuntime 转发 prompt/abort、事件流回传、会话持久化恢复、容器销毁数据保留 |
| T10 | `test/admin.test.ts` | admin 归档任意组、任免组长、下架任意资源、review |
| T11 | `test/sessions-admin.test.ts` | 全量会话列表含容器状态、terminate 立即生效 |

## 11. 开放问题（施工前确认）

| # | 问题 | 本设计默认值 | 备注 |
|---|---|---|---|
| A2 | 数据集只读挂载 vs 复制 | 只读 bind mount | T7 后验证存储与并发 |
| A3 | GPU 内存（Ollama + 多容器） | 并发上限 ~10 会话 | T13 部署实测 |
| A4 | 学生可创建挂组项目 | **允许** | 已按允许施工 |
| A5 | 数据集上限 | 500MB/文件、20GB/用户 | 已按默认施工 |
| A7 | 先上线后审核 | reviewState=pending + admin 下架兜底 | 已按默认施工 |
| — | 删除项目是否级联终止 running 会话 | **是**（终止容器 + 解除链接） | 备选：仅解除链接 |
| — | datasets 是否加 review_state | ✅ 已加（003 迁移，T10） | 与技能一致：公开创建/改公开 → pending |
| — | 目录按 id 平铺 vs 三层分层 | 按 id 平铺 + DB ACL 强制 | 对 plan 决策 4 的细化，见第 7 节 |

## 12. 与既有文档的关系

- 本设计是 `tasks/plan.md` 架构决策的落地细化；plan 中「目录三层布局」被第 7 节细化为「DB ACL 强制 + 按 id 平铺」（可回退）
- API 契约以本文档为准，前端按 6.x 实现
- 施工顺序：T4 → T5（后端 + 测试先行）→ T6/T7 可并行 → T8a/T8b → T10/T11

