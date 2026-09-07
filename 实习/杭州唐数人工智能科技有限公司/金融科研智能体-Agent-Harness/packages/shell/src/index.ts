/**
 * tfa-shell 公共导出面
 *
 * 壳层对外暴露:
 * - CLI 入口 shellMain(cli.ts)
 * - 会话服务 ShellSessionService(server/session-service.ts)
 * - HTTP 服务 startHttpServer(server/http-server.ts)
 * - 金融扩展机制 FintechToolRegistry / MarketDataSource(fintech/)
 */

export type { FauxSetup } from "./core/faux.ts";
export { setupFauxModel } from "./core/faux.ts";
export type { ShellRuntimeOptions } from "./core/runtime-factory.ts";
export { createShellRuntimeFactory } from "./core/runtime-factory.ts";
export { shellMain } from "./entry.ts";
export type { FinancialReport, MarketDataSource, Quote } from "./fintech/datasource.ts";
export { DataSourceUnavailableError } from "./fintech/datasource.ts";
export { echoMarketTool } from "./fintech/echo-market.ts";
export type { FintechToolMeta } from "./fintech/registry.ts";
export { FintechToolRegistry } from "./fintech/registry.ts";
export type { HttpServerHandle, HttpServerOptions } from "./server/http-server.ts";
export { createApp, startHttpServer } from "./server/http-server.ts";
export type { ShellSessionServiceOptions } from "./server/session-service.ts";
export { ShellSessionService } from "./server/session-service.ts";
