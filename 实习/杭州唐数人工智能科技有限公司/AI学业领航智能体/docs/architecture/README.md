# 学业领航系统架构图

这组图以当前仓库中的运行链为准：

`academic-navigation-demo`（Vue 3 + Vite 前端） → `/api` Vite Proxy → `academic-navigation-backend`（Spring Boot，8080）。

## 图纸

- [流程图 PNG](./academic-navigation-flow.png)：从五类角色登录开始，展示 JWT 鉴权、领域服务、MySQL 查询、派生计算、校方同步和 AI 输出的闭环。
- [流程图源文件](./academic-navigation-flow.excalidraw)：可在 Excalidraw 中继续编辑。
- [整体架构图 PNG](./academic-navigation-architecture.png)：按体验层、接入安全层、领域应用层、计算适配层、持久化与外部系统分层。
- [整体架构图源文件](./academic-navigation-architecture.excalidraw)：可在 Excalidraw 中继续编辑。

## 阅读重点

1. 当前默认数据源是 `MockAcademicDataSource`；配置切换为 `provider=school` 后，使用 `SchoolApiDataSource` 调校方 REST API。
2. `AcademicComputeService` 和 `AcademicAlertPolicy` 将成绩、考勤等原始数据计算为 GPA 历史、五维画像和三级预警，再回写 `academic_nav`。
3. 前端 `AgentChat` 的 Dify/E-Agent 调用是外部直连流式通道；后端 `/api/ai/interpret` 是本地 LLM（默认 Ollama）解读通道，两者是两条不同的 AI 路径。
4. 历史的 `academic-navigation/` 精简 Spring Boot 模块的源代码与配置已清理；当前运行链只使用 `academic-navigation-backend`。
