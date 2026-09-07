import type { ToolDefinition } from "@earendil-works/tfa-coding-agent";

export interface FintechToolMeta {
	/** 工具分类:行情 / 报告 / 分析 */
	category?: "market" | "report" | "analysis";
}

/**
 * 金融工具注册表:后期接入金融工具的扩展点。
 *
 * 使用方式:
 *   const registry = new FintechToolRegistry();
 *   registry.register(createQuoteTool(datasource));
 *   // 接入 CLI/Web:
 *   createShellRuntimeFactory({ cwd, customTools: registry.toCustomTools() });
 *   new ShellSessionService({ cwd, customTools: registry.toCustomTools() });
 */
export class FintechToolRegistry {
	private readonly entries = new Map<string, { tool: ToolDefinition; meta?: FintechToolMeta }>();

	register(tool: ToolDefinition, meta?: FintechToolMeta): void {
		this.entries.set(tool.name, { tool, meta });
	}

	unregister(name: string): void {
		this.entries.delete(name);
	}

	list(): Array<{ tool: ToolDefinition; meta?: FintechToolMeta }> {
		return [...this.entries.values()];
	}

	/** 转成 createAgentSession 的 customTools 参数。 */
	toCustomTools(): ToolDefinition[] {
		return this.list().map((entry) => entry.tool);
	}

	/** 工具名列表,追加到 tools 允许名单。 */
	toolNames(): string[] {
		return [...this.entries.keys()];
	}
}
