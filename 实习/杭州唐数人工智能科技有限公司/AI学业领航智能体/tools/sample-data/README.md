# 样本数据生成工具

用于为 Academic Navigation 生成可导入的测试样本，输出两套一致的数据：

- `sample_data.sql`：MySQL 8 可直接执行的 INSERT 脚本（含全部业务表）。
- `sample_data.xlsx`：与 SQL 完全同源的 Excel 工作簿，每个业务表一个 Sheet，方便用 ETL/Excel 导入工具做正规数据接入测试。

## 运行

脚本依赖 `@oai/artifact-tool` 生成 Excel，当前仓库已在 `tools/sample-data/node_modules` 建立指向 Codex 运行时依赖的 junction。用以下命令可重新生成：

```powershell
& "C:\Users\BenBen\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe" `
  tools/sample-data/generate_sample_data.mjs `
  --out "E:\Academic Navigation-test\Academic Navigation - 副本\outputs\sample-data" `
  --departments 2 --students 120 --classes 4 --majors 4 --courses 12 --terms 8 --seed 20260812
```

常用参数：

| 参数 | 默认 | 说明 |
|---|---|---|
| `--departments` | 2 | 学院数量，专业、班级、教师按学院归属 |
| `--students` | 120 | 学生数量，会在班级间尽量平均分配 |
| `--classes` | 4 | 班级数量，每班自动生成一名班主任 |
| `--majors` | 4 | 专业数量，班级按专业循环归属 |
| `--courses` | 12 | 课程数量，每学期每班都会开设 |
| `--terms` | 8 | 学期数量，格式自动生成如 `2024-2025-1` |
| `--first-year` | 2024 | 起始学年 |
| `--id-offset` | 100000 | 显式 ID 起始偏移，避免与 Flyway 种子数据冲突 |
| `--seed` | 20260812 | 随机种子，固定后可复现同一份数据 |
| `--out` | `outputs/sample-data` | 输出目录 |

校验生成的 Excel：

```powershell
& "C:\Users\BenBen\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe" `
  tools/sample-data/verify_workbook.mjs "路径\sample_data.xlsx"
```

## 数据约定

- 学生账号为纯数字（如 `20260001`），教师/任课教师/系主任/院长账号分别以 `T`、`C`、`D`、`L` 开头，符合后端 `RoleDetector` 的自动识别规则。
- 默认生成 2 个学院：专业、班级、班主任、任课教师、系主任都按学院归属；院长为统一的 `L` 账号，可查看全校数据。
- 密码统一是 `123456`，`t_user.password` 写入与 Flyway 种子一致的 BCrypt 哈希。
- `term` 使用 `2024-2025-1` 这种格式；`failed_courses`、`items`、`methods`、`details` 等字段写入合法 JSON 字符串，非挂科预警的 `failed_courses` 为 `[]`。
- GPA 由当学期成绩按学分加权计算，`rank` 按专业内 GPA 排序生成；画像、体测、预警数据都关联到真实存在的学生 ID。

## 导入 SQL

SQL 使用显式 ID，适合导入独立测试库，或先清空对应表再导入：

```sql
CREATE DATABASE academic_nav_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

然后执行 Flyway 的 `V1__init_user.sql`、`V2__add_course_teacher.sql`、`V3__init_business_data.sql`，最后执行生成的 `sample_data.sql`。导入后可登录新增账号，验证学生端、班主任端、预警模块以及五角色权限。

## 接口接入模式

生成器同时会输出 `sample_data.json`，可以把它当“校方教务系统”的模拟数据源，让后端通过 REST 接口拉取并同步到本地库。

1. 启动模拟教务接口服务（默认端口 9090）：

```powershell
& "C:\Users\BenBen\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe" `
  tools/sample-data/school-api-server.mjs `
  --data "E:\Academic Navigation-test\Academic Navigation - 副本\outputs\019ff3a5-6093-7392-a178-984c45645bd0\sample_data.json" `
  --port 9090
```

2. 后端已实现 `SchoolApiDataSource` + `SchoolApiSyncService`，用院长账号触发一次全量同步：

```text
POST http://localhost:8080/api/school-sync
Authorization: Bearer <院长 token>
```

同步会按 `t_department -> t_major -> t_class -> t_user -> t_teacher -> t_student -> 课程/开课 -> 成绩/GPA/画像/体测/预警/干预` 的顺序 upsert 到 `academic_nav`，完成后现有业务接口立即读到新数据。

3. 需要启动时自动同步，把后端配置切到 school 数据源：

```text
ACADEMIC_DATASOURCE_PROVIDER=school
SCHOOL_API_BASE_URL=http://localhost:9090
```

对应配置项在 `application.yml` 的 `academic.datasource.provider` 和 `academic.school.*`；默认仍是 `mock`，不自动拉取。

模拟接口已提供以下端点：`/api/v1/departments`、`/api/v1/majors`、`/api/v1/classes`、`/api/v1/users`、`/api/v1/teachers`、`/api/v1/students`、`/api/v1/courses`、`/api/v1/course-classes`、`/api/v1/alerts`、`/api/v1/interventions`，以及按学生维度查询的成绩、GPA、画像、体测、预警子接口。将来接真实校方接口时，只需替换 `SchoolApiClient` 的请求地址和字段映射。

## 后续正规数据接入

校方正式数据接入时，建议以本工具生成的 `字段字典` Sheet 为映射基准：

1. 把校方接口/数据源字段映射到 `t_*` 表的列名。
2. 用 Excel 样本跑通 ETL 或批量导入，再做字段类型、枚举值、JSON 格式校验。
3. 确认后再实现 `SchoolApiDataSource` 适配层，保持前端类型不变。
