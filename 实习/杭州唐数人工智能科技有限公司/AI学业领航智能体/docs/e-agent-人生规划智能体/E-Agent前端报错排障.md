# E-Agent 前端报错排障

## TypeError: Cannot read properties of undefined (reading 'length')

已知触发位置：QA 知识库的“导入 XLSX”流程。

前端代码在提交导入后执行了类似：

```js
const firstError = result.errors[0];
firstError.length
```

当导入成功、`errors` 为空数组时，`firstError` 是 `undefined`，页面就会崩到全局错误页：

```text
TypeError: Cannot read properties of undefined (reading 'length')
```

这不是你的 XLSX 文件格式错误。

## 处理顺序

1. 点击页面上的 `Reset Application`。
2. 回到 QA 知识库页面。
3. 查看导入的数据是否已经出现：
   - 如果已经出现，说明导入实际成功，只是前端提示层崩了。
   - 如果没有出现，继续使用下面的绕行方案。

## 绕行方案 A：使用“新建 QA”

不要使用“导入”，直接：

1. 进入 QA 知识库。
2. 点击“新建 QA”。
3. 把 XLSX 里的每组 `question` 和 `answer` 手动复制进去。

适合数据量较少的情况。

## 绕行方案 B：使用 E-Agent 提供的示例文件

1. 点击 QA 导入页面的“示例文件”。
2. 下载并打开官方 XLSX。
3. 把我们的 `question`、`answer` 两列内容按相同结构复制进去。
4. 重新导入。

如果官方示例文件也触发同样错误，可以确认是当前 E-Agent 版本的前端缺陷。

## 绕行方案 C：通过 E-Agent 后端 API 导入

如果你有 E-Agent 登录后的 `ws_token`，可以绕过前端直接调用导入接口。这里只做模板，正式使用需要替换 token、知识库 ID 和文件路径：

```powershell
$headers = @{
    Authorization = "Bearer <E-Agent ws_token>"
}

$form = @{
    knowledge_id = "<QA知识库ID>"
    file = Get-Item "E:\Academic Navigation-test\Academic Navigation - 副本\outputs\eagent-formal-data\人生规划安全FAQ.xlsx"
}

Invoke-RestMethod `
    -Method Post `
    -Uri "http://localhost:3001/api/v1/knowledge/qa/import/<QA知识库ID>" `
    -Headers $headers `
    -Form $form
```

> 注意：`localhost:3001` 只是你本机访问 E-Agent 的入口。API 导入是 E-Agent 后端执行，和前端无关。

## 建议

当前 E-Agent 前端这个错误属于平台缺陷。正式上线前，建议向 E-Agent 管理员反馈该导入错误，或在升级平台版本后重新测试 QA 导入。

## 创建助手时提示 error：'description'

这是创建“助手”应用时的字段校验，不是系统崩溃。

E-Agent 对助手描述有要求：

- 描述不能为空。
- 描述必须大于 20 个中文字符。

如果你在“角色和任务”框中只填了几个字，或者没填，保存时就会报 `description`。

解决方法：

1. 回到创建助手页面。
2. 找到“你希望助手的角色是什么，具体完成什么任务？”输入框。
3. 填写一段超过 20 个字的描述。

推荐描述：

```text
你是浙江师范大学学生端 AI 人生规划助手，根据学生学业画像、GPA 趋势和发展目标，为学生提供考研、留学、就业、考公考编等方向分析，并制定可执行的大学阶段行动路线。
```

名称和描述都填写后，再点击“创建”。

## 工作流执行失败：'description'

如果已经能保存工作流，但点击运行或对话时报：

```text
工作流执行任务失败：'description'
```

通常是某个节点的“描述”字段为空。E-Agent 工作流节点顶部名称下方有一个小号描述输入框，后端运行时会校验它。

重点检查：

1. 选中“助手”节点。
2. 看节点名称下方的描述是否为空。
3. 如果为空，填一段简要描述。

推荐描述：

```text
根据学生画像和补充信息，生成人生规划方向与可执行建议。
```

如果画布中有多个“助手”节点，每个都要检查。工具节点同样需要在工具管理页面填写描述。

快速定位：

1. 回到工作流画布。
2. 一个一个选中节点。
3. 删除空白描述，填写简短中文说明。
4. 保存工作流后重新运行。

如果还报错，点击该节点右上角的“运行此节点”，看日志中具体是哪一条参数缺失。

## 已补节点描述，仍报 'description'

如果已经给“助手”节点填了描述，运行工作流仍然提示 `description`，重点检查 API 工具本身的描述。

操作：

1. 进入“构建 -> 工具 -> API工具”。
2. 找到 `check_local_database`。
3. 打开编辑。
4. 找到“描述”或 `description` 输入框。
5. 填写：

```text
查询本地学业数据库连接状态，返回数据库名称、时间和关键表行数。
```

6. 保存工具。
7. 回到工作流，重新把该工具加入“助手”节点的工具列表。
8. 保存工作流，再重新运行。

如果工具管理页面没有单独的描述输入框，就在 OpenAPI Schema 的 `info.description` 和该 GET 接口的 `summary` 中补上文字，然后保存。

## 连接数据库时报“服务器错误”

三层验证已经通过：

```text
本机 http://localhost:8080/api/eagent/health            -> 200
node1 http://127.0.0.1:18080/api/eagent/health          -> 200
E-Agent 主机 http://127.0.0.1:18080/api/eagent/health   -> 200
```

所以数据库和后端本身正常。E-Agent 仍报服务器错误，通常有两个原因：

### 1. API 工具的 server_host / 访问根地址不对

打开 `check_local_database` 工具，确认：

```text
server_host = http://127.0.0.1:18080
path = /api/eagent/health
method = GET
鉴权方式 = 无
```

不要把 server_host 写成：

```text
http://localhost:8080
```

因为 E-Agent 后端执行工具调用时，`localhost` 是 E-Agent 服务器自己，不是你 Windows 本机。

### 2. 两段反向隧道没有保持运行

每次运行智能体前，隧道必须存在。可在本机另开终端运行：

```powershell
.\scripts\start-eagent-chain-tunnel.ps1 -Action Start
```

状态检查：

```powershell
.\scripts\start-eagent-chain-tunnel.ps1 -Action Status
```

确认两项都返回 `HTTP 200`。
