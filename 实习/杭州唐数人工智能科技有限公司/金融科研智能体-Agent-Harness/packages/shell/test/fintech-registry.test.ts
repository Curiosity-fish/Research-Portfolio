import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { setupFauxModel } from "../src/core/faux.ts";
import { DataSourceUnavailableError, type MarketDataSource, type Quote } from "../src/fintech/datasource.ts";
import { echoMarketTool } from "../src/fintech/echo-market.ts";
import { FintechToolRegistry } from "../src/fintech/registry.ts";
import { ShellSessionService } from "../src/server/session-service.ts";

const cleanupDirs: string[] = [];

afterEach(() => {
	for (const dir of cleanupDirs) {
		rmSync(dir, { recursive: true, force: true });
	}
	cleanupDirs.length = 0;
});

describe("FintechToolRegistry", () => {
	it("注册/注销/列出/转换", () => {
		const registry = new FintechToolRegistry();
		registry.register(echoMarketTool, { category: "market" });
		expect(registry.toolNames()).toEqual(["echo_market"]);
		expect(registry.list()).toHaveLength(1);
		expect(registry.toCustomTools()[0].name).toBe("echo_market");
		registry.unregister("echo_market");
		expect(registry.toolNames()).toEqual([]);
	});

	it("重复注册覆盖同名工具", () => {
		const registry = new FintechToolRegistry();
		registry.register(echoMarketTool);
		registry.register(echoMarketTool, { category: "analysis" });
		expect(registry.list()).toHaveLength(1);
		expect(registry.list()[0].meta?.category).toBe("analysis");
	});

	it("示例工具 execute 返回占位结果", async () => {
		const result = await echoMarketTool.execute("call-1", { symbol: "600519" }, undefined, undefined, {} as never);
		const part = result.content[0];
		expect(part.type).toBe("text");
		if (part.type === "text") {
			expect(part.text).toContain("600519");
		}
	});

	it("数据源接口可被实现(占位数据源示例)", async () => {
		const datasource: MarketDataSource = {
			name: "placeholder",
			async getQuote(symbol: string): Promise<Quote> {
				return { symbol, price: 0, currency: "CNY", timestamp: Date.now() };
			},
			async getQuotes(symbols: string[]): Promise<Quote[]> {
				return symbols.map((symbol) => ({ symbol, price: 0, currency: "CNY", timestamp: Date.now() }));
			},
			async getFinancialReport(): Promise<never> {
				throw new DataSourceUnavailableError("财报数据源未接入");
			},
		};
		const quote = await datasource.getQuote("600519");
		expect(quote.symbol).toBe("600519");
		await expect(datasource.getFinancialReport("600519")).rejects.toBeInstanceOf(DataSourceUnavailableError);
	});

	it("注册表工具经 ShellSessionService 注入会话,会话可用工具包含 echo_market", async () => {
		const cwd = mkdtempSync(join(tmpdir(), "tfa-fintech-test-"));
		cleanupDirs.push(cwd);
		const faux = setupFauxModel();
		const registry = new FintechToolRegistry();
		registry.register(echoMarketTool);
		const service = new ShellSessionService({
			cwd,
			agentDir: join(cwd, "agent"),
			fauxSetup: faux,
			customTools: registry.toCustomTools(),
		});
		const runtime = await service.createSession({ id: "fintech-1" });
		const runtime2 = runtime as unknown as { agentSession: { getAllTools(): Array<{ name: string }> } };
		const tools = runtime2.agentSession.getAllTools().map((tool) => tool.name);
		expect(tools).toContain("echo_market");
		await service.close();
		faux.unregister();
	});
});
