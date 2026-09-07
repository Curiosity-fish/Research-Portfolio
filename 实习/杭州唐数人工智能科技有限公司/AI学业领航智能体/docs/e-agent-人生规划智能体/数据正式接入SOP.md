# 学业数据正式接入 SOP

## 1. 目标

让真实学业数据以可控、可校验、可审计的方式进入系统，并被 E-Agent 使用。

正式数据分三类，分别走不同通道：

| 数据 | 正式通道 | 是否上传到 E-Agent |
|---|---|---|
| GPA、画像、预警、课程、成绩等实时数据 | 后端 API / 数据库视图 | 否，E-Agent 实时调用 |
| 培养方案、政策、FAQ、规划方法论 | E-Agent 知识库文件 | 是，上传文档/QA |
| 历史数据、批量导入、校方数据 | MySQL 导入或 School API 同步 | 不直接传 E-Agent |

## 2. 当前正式链路

```text
数据源
  -> 数据标准化 / 字段映射
  -> academic_nav MySQL
  -> Spring Boot 业务 API
  -> 反向隧道 127.0.0.1:18080
  -> E-Agent API 工具
  -> AI 人生规划智能体
```

静态知识：

```text
knowledge/*.md
  -> E-Agent 文档知识库 / QA 知识库
  -> 智能体知识检索节点
```

## 3. 正式数据包导出

运行：

```powershell
.\scripts\export-eagent-formal-data.ps1 `
  -OutDir "E:\Academic Navigation-test\Academic Navigation - 副本\outputs\eagent-formal-data"
```

脚本会生成：

```text
eagent-formal-data/
├─ manifest.json
├─ tables/
│  ├─ t_department.tsv
│  ├─ t_major.tsv
│  ├─ t_class.tsv
│  ├─ t_teacher.tsv
│  ├─ t_student.tsv
│  ├─ t_user_public.tsv
│  ├─ t_course.tsv
│  ├─ t_course_class.tsv
│  ├─ t_grade.tsv
│  ├─ t_gpa_history.tsv
│  ├─ t_profile_score.tsv
│  ├─ t_health_report.tsv
│  ├─ t_alert.tsv
│  ├─ t_attendance.tsv
│  ├─ t_psychology.tsv
│  ├─ t_volunteer.tsv
│  ├─ t_competition.tsv
│  ├─ t_job_cache.tsv
│  ├─ t_school_cache.tsv
│  └─ t_major_plan.tsv
├─ 人生规划安全FAQ.xlsx
└─ knowledge_package/
   ├─ 01_人生规划方法论.md
   ├─ 02_大学生发展路径与阶段任务.md
   ├─ 03_考研留学就业考公决策框架.md
   ├─ 04_浙师大学业与就业政策FAQ.md
   └─ 05_心理与学业风险识别.md
```

敏感字段处理：

- 不导出 `t_user.password`、`avatar_url`。
- `t_user_public` 只保留账号、姓名、角色、学院、专业、班级、年级、状态。
- 心理、体测、预警等数据保留在本地库和受控 API 中，不进入 E-Agent 文档知识库。

## 4. 数据校验

导出后至少检查：

1. `manifest.json` 中每张表行数与源库一致。
2. `t_student.student_id` 与 `t_user.account` 能关联。
3. `t_class.major_id`、`t_student.class_id` 能关联到有效专业和班级。
4. `t_grade.course_class_id` 能关联到有效开课记录。
5. `term` 格式为 `YYYY-YYYY-1` 或 `YYYY-YYYY-2`。
6. 枚举值符合：
   - `role`: student/teacher/course_teacher/department/dean
   - `level`: yellow/orange/red
   - `status`: pending/processing/resolved/passed/failed/retake/active/finished
7. 不存在重复的 `t_student.student_id`、`t_grade(student_id, course_class_id)`、`t_gpa_history(student_id, term)`。
8. 不包含手机号、身份证号、家庭地址等未授权字段。

## 5. 本地正式导入

### 方式 A：直接导入当前 `academic_nav`

如果已有合法数据文件，按 `docs/数据接入说明.md` 的字段要求导入对应表。推荐先导入基础档案，再导入课程成绩，最后触发派生数据计算。

### 方式 B：通过 School API 同步

如果使用标准化 `sample_data.json` 或真实校方 API：

1. 启动学校数据接口：

```powershell
& "C:\Users\BenBen\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe" `
  tools/sample-data/school-api-server.mjs `
  --data "E:\Academic Navigation-test\Academic Navigation - 副本\outputs\formal_data.json" `
  --port 9090
```

2. 用院长账号调用同步：

```text
POST http://localhost:8080/api/school-sync
Authorization: Bearer <院长 token>
```

3. 触发派生数据计算：

```text
POST http://localhost:8080/api/school-sync/compute
Authorization: Bearer <院长 token>
```

4. 检查现有业务接口，数据应立即更新。

## 6. E-Agent 端配置

### 实时结构化数据

1. 启动当前拓扑的两段反向隧道：

```powershell
.\scripts\start-eagent-chain-tunnel.ps1 -Action Start
```

2. E-Agent API 工具 Base URL：

```text
http://127.0.0.1:18080
```

3. 在 `agent` 节点中配置 6 个只读工具：

- `get_student_dashboard`
- `get_student_profile`
- `get_gpa_trend`
- `get_student_alerts`
- `get_development_paths`
- `analyze_development_goal`

### 静态知识

1. 登录 E-Agent。
2. 进入“文件库”。
3. 创建文档知识库和 QA 知识库。
4. 上传 `knowledge_package/` 中的 Markdown 文件。
5. 给文件添加 `domain`、`grade`、`risk_level`、`audience` 元数据。

QA 知识库导入时使用：

```text
outputs/eagent-formal-data/人生规划安全FAQ.xlsx
```

该文件第一张表只包含 `question`、`answer` 两列，符合 E-Agent QA 知识库的 XLSX 导入格式。

## 7. 未来接学校 API

学校提供 API 后，按以下顺序做，不改变前端和 E-Agent：

1. 获取 API 文档、鉴权方式、字段样例、更新频率。
2. 写“校方字段 -> 本地字段”映射表。
3. 修改或新增 `SchoolApiClient` 请求地址和鉴权头。
4. 保留 `SchoolApiSyncService` 的 upsert 与分页逻辑。
5. 按频率同步：
   - 成绩、GPA：每小时或每天增量。
   - 画像、预警：每天批处理。
   - 基础档案：变更时同步。
6. 同步完成后验证 `counts`、抽样学生和前端展示。

## 8. 验收清单

- [ ] 数据包字段与源库一致。
- [ ] 无未授权敏感字段导出。
- [ ] 数据能导入本地 MySQL 或通过 School API 同步。
- [ ] 后端接口能返回真实数据。
- [ ] E-Agent 工具能通过 `127.0.0.1:18080` 读取数据。
- [ ] 知识库文件已上传并完成索引。
- [ ] 学生身份隔离生效，不能跨学号查询。
- [ ] 心理、健康、预警高风险回答包含转介提示。
- [ ] 日志可追踪数据来源、同步批次和接口调用。
