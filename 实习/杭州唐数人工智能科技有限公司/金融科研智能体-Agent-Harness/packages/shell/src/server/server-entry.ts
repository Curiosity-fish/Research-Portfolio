import { getAgentDir, type ToolDefinition } from "@earendil-works/tfa-coding-agent";
import { setupFauxModel } from "../core/faux.ts";
import type { ShellRuntimeOptions } from "../core/runtime-factory.ts";
import { echoMarketTool } from "../fintech/echo-market.ts";
import { FintechToolRegistry } from "../fintech/registry.ts";
import { startHttpServer } from "./http-server.ts";
import { ShellSessionService } from "./session-service.ts";

/**
 * Web 服务入口:读取环境变量装配 ShellSessionService 并启动 HTTP 服务。
 *
 * 环境变量:
 *   TFA_SHELL_PORT      监听端口(默认 8787)
 *   TFA_SHELL_HOST      监听地址(默认 127.0.0.1)
 *   TFA_SHELL_CWD       会话默认工作目录(默认 process.cwd())
 *   TFA_SHELL_AGENT_DIR 全局配置目录(默认 getAgentDir())
 *   TFA_SHELL_FAUX      离线模式:注册 faux provider 作为默认模型(无 API key 验证用)
 */
export async function startWebServer(
	argv: string[],
	shellRuntimeOptions: Omit<ShellRuntimeOptions, "cwd" | "agentDir"> | undefined,
): Promise<number> {
	const args = parseWebArgs(argv);
	const cwd = process.env.TFA_SHELL_CWD ?? process.cwd();
	const agentDir = process.env.TFA_SHELL_AGENT_DIR ?? getAgentDir();
	const port = args.port ?? parsePort(process.env.TFA_SHELL_PORT);
	const host = process.env.TFA_SHELL_HOST ?? "127.0.0.1";

	const fauxSetup = isTruthyEnvFlag(process.env.TFA_SHELL_FAUX) ? setupFauxModel() : undefined;
	if (fauxSetup) {
		console.log("[tfa-shell] TFA_SHELL_FAUX 已启用:使用 faux provider(离线模式)");
	}

	const service = new ShellSessionService({
		cwd,
		agentDir,
		customTools: shellRuntimeOptions?.customTools ?? defaultFintechTools(),
		toolAllowlist: shellRuntimeOptions?.toolAllowlist,
		defaultModel: fauxSetup?.defaultModel,
		fauxSetup,
	});

	const handle = await startHttpServer({ service, port, host });
	// 智能体广场链接健康检查(周期默认 5 分钟)
	service.startAgentHealthCheck();
	console.log(`[tfa-shell] Web 服务已启动: ${handle.url}`);

	await new Promise<void>((resolve) => {
		const shutdown = async () => {
			console.log("[tfa-shell] 正在关闭...");
			await handle.shutdown();
			resolve();
		};
		process.once("SIGINT", shutdown);
		process.once("SIGTERM", shutdown);
	});
	return 0;
}

function parseWebArgs(argv: string[]): { port: number | undefined } {
	let port: number | undefined;
	for (let index = 0; index < argv.length; index++) {
		const arg = argv[index];
		if ((arg === "--port" || arg === "-p") && index + 1 < argv.length) {
			port = parsePort(argv[++index]);
		}
	}
	return { port };
}

function parsePort(value: string | undefined): number {
	if (value === undefined) return 8787;
	const parsed = Number.parseInt(value, 10);
	if (!Number.isInteger(parsed) || parsed < 1 || parsed > 65535) {
		throw new Error(`无效端口: ${value}`);
	}
	return parsed;
}

function isTruthyEnvFlag(value: string | undefined): boolean {
	return value === "1" || value === "true" || value === "yes";
}

function defaultFintechTools(): ToolDefinition[] {
	const registry = new FintechToolRegistry();
	registry.register(echoMarketTool);
	return registry.toCustomTools();
}
