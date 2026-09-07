# 金融数据源接入指南

本目录是后期接入真实金融数据源的扩展点。当前**没有预置任何数据源**,以下说明接入路径。

## 三步接入法

1. **实现 `MarketDataSource` 接口**(`../datasource.ts`)

   新建 `src/fintech/datasources/<name>.ts`,实现:

   ```ts
   import type { MarketDataSource, Quote, FinancialReport } from "../datasource.ts";

   export class XxxDataSource implements MarketDataSource {
     readonly name = "xxx";
     async getQuote(symbol: string, opts?: { signal?: AbortSignal }): Promise<Quote> {
       // 调用行情 HTTP 接口,返回标准化 Quote
     }
     async getQuotes(symbols: string[], opts?: { signal?: AbortSignal }): Promise<Quote[]> { ... }
     async getFinancialReport(symbol: string, period?: string, opts?: { signal?: AbortSignal }): Promise<FinancialReport> { ... }
   }
   ```

2. **注册到 `FintechToolRegistry`**

   在 `src/fintech/index.ts`(或壳装配处)注册:

   ```ts
   const registry = new FintechToolRegistry();
   registry.register(createQuoteTool(new XxxDataSource()), { category: "market" });
   ```

3. **注入会话**

   CLI 与 Web 共用同一个注册表:

   ```ts
   // CLI:createShellRuntimeFactory
   createShellRuntimeFactory({ cwd, customTools: registry.toCustomTools() });
   // Web:ShellSessionService
   new ShellSessionService({ cwd, customTools: registry.toCustomTools() });
   ```

   注册后的工具自动出现在:
   - CLI 交互终端的可用工具列表
   - Web 前端 `/api/models` 之外的工具调用(前端 `ToolCallCard` 渲染工具名/参数/结果)

## 常见数据源形态

| 数据源 | 形态 | 接入要点 |
| -- | -- | -- |
| 新浪/腾讯行情 | HTTP JSON 接口 | 简单 GET,无鉴权,注意限频 |
| 东方财富 | HTTP JSON 接口 | 字段复杂,需要归一化映射 |
| AkShare / Tushare | Python 库 | 需经子进程或独立服务桥接 |
| 财报数据 | PDF/Excel 文件 | 需解析 + 结构化,建议单独工具 |
| 商业数据服务 | REST API + key | 密钥放配置/环境变量,不得写死 |

## 错误处理约定

- 数据源不可用或数据缺失时抛 `DataSourceUnavailableError`,工具执行层统一转为失败结果
- 工具内使用 `opts?.signal` 支持中止(用户点击"停止生成"时)
- 网络请求建议设置超时与重试,避免阻塞 agent 回合
