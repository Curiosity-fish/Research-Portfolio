# 前后端接口配合清单 (Backend ↔ Frontend API Contract)

> 更新日期:2026-08-24 | 适用范围:shell 后端 + tfc-web 前端
> 后端入口:`packages/shell/src/server/http-server.ts` | 前端客户端:`packages/shell/web/src/api/client.ts`
> 默认监听:`127.0.0.1:8787`(可通过环境变量覆盖),所有 REST 路径前缀 `/api`。

## 0. 通用约定

- **认证**:除 `POST /api/auth/login` 与 `GET /api/health` 外,所有 `/api` 路由都要求请求头 `Authorization: Bearer <token>`(登录后返回)。
- **响应包裹**:单资源返回 `{ xxx }`,列表返回 `{ xxxs: [] }`,无资源操作返回 `{ ok: true }`。
- **错误格式**:统一 `{ error: { code, message } }`;HTTP 状态:401 未登录 / 403 无权限 / 404 不存在 / 409 冲突或忙 / 413 超配额 / 400 参数错误 / 500 内部错误。
- **上传**:数据集上传与项目文件上传使用 `multipart/form-data`(FormData),非 JSON。
- **实时事件**:会话进度走 SSE(`GET /api/sessions/:id/events`),前端用 `EventSource`,断线自动重连。
- **权限模式**:项目会话默认 `request_approval`(需要批准),通用会话默认 `full_access`。

---

## 1. 认证与用户

| 方法 | 路径 | 鉴权 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|---|
| POST | `/api/auth/login` | 公开 | `{ username, password }` | `{ token, user, expiresAt }` | `App.vue` 登录页;登录后保存 token 并写入 `ApiClient` |
| POST | `/api/auth/logout` | 已登录 | - | `{ ok: true }` | 退出登录;失败也清本地 token |
| GET | `/api/auth/me` | 已登录 | - | `{ user }` | 页面刷新恢复会话 / 校验登录态 |

> 账号:admin(管理员)/ teacher(老师)/ student(学生),密码均为 `000000`。

## 2. 系统与配置

| 方法 | 路径 | 鉴权 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/health` | 公开 | `{ ok: true }` | 健康检查(部署探针) |
| GET | `/api/models` | 已登录 | `{ models }` | `ModelPicker.vue` 模型下拉 |
| GET | `/api/config` | 已登录 | `{ config }` | 读取是否已配置 API Key、默认模型等 |
| GET | `/api/tools` | 已登录 | `{ tools }` | `ToolsDrawer.vue` 工具抽屉 |
| POST | `/api/config/api-key` | 已登录 | `{ provider, apiKey }` | `ModelSettingsDrawer.vue` 保存 DeepSeek 等 API Key(在线模式) |
| DELETE | `/api/config/api-key?provider=xxx` | 已登录 | `{ ok: true }` | 移除 API Key(回到离线) |

## 3. 管理后台(仅 admin)

| 方法 | 路径 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/admin/users` | - | `{ users }` | `AdminView.vue` 用户管理列表 |
| POST | `/api/admin/users` | `{ username, password, role }` | `{ user }` | 新建管理员/老师/学生 |
| PATCH | `/api/admin/users/:id` | `{ status?, password?, role? }` | `{ user }` | 启用/禁用、重置密码、改角色 |
| GET | `/api/admin/sessions` | - | `{ sessions }` | `AdminView.vue` 全局会话监控 |
| POST | `/api/admin/sessions/:id/terminate` | - | `{ ok: true }` | 强制终止某会话 |
| POST | `/api/admin/skills/:id/review` | `{ reviewState: approved|rejected }` | `{ skill }` | 技能审核 |
| POST | `/api/admin/datasets/:id/review` | `{ reviewState }` | `{ dataset }` | 数据集审核 |

## 4. 课程组 / 科研组

| 方法 | 路径 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/groups` | - | `{ groups }` | `GroupsView.vue` 组列表 |
| POST | `/api/groups` | `{ name, type: course|research }` | `{ group }` | 建组(老师/管理员) |
| GET | `/api/groups/:id` | - | `{ group }` | 组详情(含成员) |
| PATCH | `/api/groups/:id` | `{ name? }` | `{ group }` | 重命名 |
| DELETE | `/api/groups/:id` | - | `{ group }` | **归档**组(归档后成员仍可见历史项目) |
| POST | `/api/groups/:id/members` | `{ userId }` | `{ ok: true }` | 添加成员 |
| PATCH | `/api/groups/:id/members/:userId` | `{ role: owner|member }` | `{ ok: true }` | 改成员角色 |
| DELETE | `/api/groups/:id/members/:userId` | - | `{ ok: true }` | 移除成员 |

## 5. 技能库

| 方法 | 路径 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/skills` | - | `{ skills }` | `SkillsView.vue` 技能列表(按可见性过滤) |
| POST | `/api/skills` | `{ name?, repoUrl?, visibility, groupId?, upload? }` | `{ skill }` | 从仓库导入或本地上传 |
| POST | `/api/skills/:id/update` | - | `{ skill }` | 从 git 仓库拉取更新 |
| PATCH | `/api/skills/:id` | `{ name?, visibility?, groupId? }` | `{ skill }` | 改名 / 改可见性 |
| DELETE | `/api/skills/:id` | - | `{ ok: true }` | 删除(owner/admin) |

## 6. 数据集库

| 方法 | 路径 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/datasets` | - | `{ datasets }` | `DatasetsView.vue` 数据集列表 |
| POST | `/api/datasets` | multipart:`file` 必填,`name?/visibility/groupId?` | `{ dataset }` (201) | **文件上传**(无大小上限) |
| PATCH | `/api/datasets/:id` | `{ name?, visibility?, groupId? }` | `{ dataset }` | 改名 / 改可见性 |
| DELETE | `/api/datasets/:id` | - | `{ ok: true }` | 删除 |
| GET | `/api/datasets/:id/download` | - | 文件 blob(Content-Disposition) | `downloadDataset` 下载 |

## 7. 智能体广场(FR-9 / T16)

| 方法 | 路径 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/agents` | - | `{ agents }` | `AgentPlaza.vue` 广场卡片(含 availability 在线状态) |
| POST | `/api/agents` | `{ name, url, description?, visibility, groupId? }` | `{ agent }` (201) | 发布智能体(公开直发免审核,组需为组成员) |
| PATCH | `/api/agents/:id` | `{ name?, url?, description?, visibility?, groupId? }` | `{ agent }` | 编辑(owner/admin) |
| DELETE | `/api/agents/:id` | - | `{ ok: true }` | 下架/删除 |

## 8. 项目与工作区

| 方法 | 路径 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/projects` | - | `{ projects }` | `ProjectPicker.vue` / `SessionList.vue` 项目列表 |
| POST | `/api/projects` | `{ name, folder?, scope?: {type:'private'} | {type:'group',groupId} }` | `{ project }` (201) | 新建项目(私有/组) |
| PATCH | `/api/projects/:id` | `{ name?, folder? }` | `{ project }` | 重命名(scope 不可改) |
| DELETE | `/api/projects/:id` | - | `{ ok: true }` | 删除项目 |
| GET | `/api/projects/:id/session-context` | - | SessionContext 对象 | 进入工作台时组装 Agent 上下文 |

### 8.1 项目文件(方案 A:平台管理目录,无大小上限)

| 方法 | 路径 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/projects/:id/files?path=` | query `path`(相对目录) | `{ files }` | `ProjectWorkbench.vue` 左栏文件树 |
| GET | `/api/projects/:id/files/content?path=` | query `path` | `{ file: { ...meta, content } }` | Monaco 编辑器读取文本 |
| PUT | `/api/projects/:id/files/content` | `{ path, content }` | `{ file }` | 编辑器保存文本 |
| POST | `/api/projects/:id/files` | multipart:`file` 必填,`path?` 目标目录 | `{ file }` (201) | **文件上传** |
| GET | `/api/projects/:id/files/download?path=` | query `path` | 文件 blob | 下载/PDF 图片预览 |
| DELETE | `/api/projects/:id/files?path=` | query `path` | `{ ok: true }` | 删除文件 |

## 9. 会话(工作台核心)

| 方法 | 路径 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/sessions` | - | `{ sessions }` | `SessionList.vue` 会话列表 |
| POST | `/api/sessions` | `{ name?, cwd?, projectId?, model?, thinkingLevel? }` | `{ session }` (201) | 新建会话/项目会话(带 projectId 自动 request_approval) |
| GET | `/api/sessions/:id` | - | `{ session }` | 拉取快照 |
| DELETE | `/api/sessions/:id` | - | `{ ok: true }` | 删除会话 |
| POST | `/api/sessions/:id/prompt` | `{ text }` | `{ ok: true }` | 发送消息(ChatView/PromptBox) |
| POST | `/api/sessions/:id/steer` | `{ text }` | `{ ok: true }` | 中途干预 |
| POST | `/api/sessions/:id/abort` | - | `{ ok: true }` | 中止生成 |
| POST | `/api/sessions/:id/model` | `{ model: { provider, id } }` | `{ ok: true }` | 切换模型 |
| POST | `/api/sessions/:id/thinking` | `{ thinkingLevel }` | `{ ok: true }` | 切换思考级别 |
| POST | `/api/sessions/:id/rename` | `{ name }` | `{ ok: true }` | 重命名会话 |
| POST | `/api/sessions/:id/project` | `{ projectId: string | null }` | `{ ok: true }` | 绑定/解绑项目(cwd → 工作区) |
| GET | `/api/sessions/:id/export` | - | text/plain | 导出对话记录 |
| GET | `/api/sessions/:id/events` | - | SSE 流 | `subscribeEvents` 订阅 snapshot/progress/error 事件 |

## 10. 权限模式与请求批准(T14/T15)

| 方法 | 路径 | 请求 | 响应 | 前端配合点 |
|---|---|---|---|---|
| GET | `/api/sessions/:id/permission-mode` | - | `{ mode }` | 工作台顶部显示当前权限模式 |
| POST | `/api/sessions/:id/permission-mode` | `{ mode: request_approval | full_access }` | `{ mode }` | 切换权限模式下拉 |
| GET | `/api/sessions/:id/approvals` | - | `{ approvals }` | `ApprovalDialog.vue` 轮询待批准请求 |
| POST | `/api/sessions/:id/approvals/:approvalId/approve` | - | `{ approval }` | 批准高风险操作(write/delete/bash/skill) |
| POST | `/api/sessions/:id/approvals/:approvalId/reject` | - | `{ approval }` | 拒绝;超时/断线默认拒绝 |

## 11. 静态托管 / SPA

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/*` | 优先返回 `packages/shell/web/dist` 静态资源,未知路径回退 `index.html`(前端路由) |

---

## 12. 前端配合要点(对接时注意)

1. **Token 管理**:`ApiClient.setToken()` 在登录/登出/恢复时调用;所有请求自动带 `Authorization` 头,SSE 的 EventSource 不带头,依赖 Cookie 或后端豁免(当前 events 路由未强制登录)。
2. **上传协议**:数据集与项目文件上传必须用 FormData + 原生 `fetch`,不能走通用 JSON `request()`。
3. **错误处理**:统一捕获 `{ error: { code, message } }`,401 跳登录,403 提示无权限,413 提示超配额。
4. **权限模式联动**:进入项目会话时先 `GET permission-mode`,若为 `request_approval` 则挂载 `ApprovalDialog` 并轮询 `approvals`(建议 2–3s),决策后回写 `approve/reject`。
5. **Agent 可用性**:`agents[].availability` 由后端健康检查(HEAD,5s 超时,每 5 分钟)维护;前端对 `unavailable` 的卡片仍可点击进入(打开 URL),仅展示状态徽章。
6. **归档组**:`DELETE /api/groups/:id` 是归档语义;归档后组内历史项目对成员仍只读可见,前端保留浏览入口。
7. **角色导航**:按 `user.role` 控制入口——admin 可见管理后台;teacher 可建组/建项目/审核? (审核仅 admin);student 仅使用自己的会话与公开/组内资源。
