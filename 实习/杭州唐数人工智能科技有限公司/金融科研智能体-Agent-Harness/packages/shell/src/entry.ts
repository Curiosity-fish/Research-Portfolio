import {
	type Args,
	createAgentSessionRuntime,
	getAgentDir,
	InteractiveMode,
	parseArgs,
	runPrintMode,
	SessionManager,
} from "@earendil-works/tfa-coding-agent";
import { type FauxSetup, setupFauxModel } from "./core/faux.ts";
import { createShellRuntimeFactory, type ShellRuntimeOptions } from "./core/runtime-factory.ts";
import { startWebServer } from "./server/server-entry.ts";

/** 壳版本号,与仓库 lockstep 一致 */
export const SHELL_VERSION = "0.83.0";

export interface ShellCliOptions {
	/** 传入会话创建的可选参数(金融工具、默认模型等) */
	shellRuntimeOptions?: Omit<ShellRuntimeOptions, "cwd" | "agentDir">;
}

function resolveAppMode(parsed: Args, stdinIsTTY: boolean, stdoutIsTTY: boolean): "interactive" | "print" | "rpc" {
	if (parsed.mode === "rpc") {
		return "rpc";
	}
	if (parsed.mode === "json") {
		return "print";
	}
	if (parsed.print || !stdinIsTTY || !stdoutIsTTY) {
		return "print";
	}
	return "interactive";
}

function printShellHelp(): void {
	console.log(`tfa-shell ${SHELL_VERSION} - 金融科技 agent 壳`);
	console.log("");
	console.log("用法:");
	console.log("  tfa-shell [选项] [消息...]        交互模式(无消息时)或单发模式(有消息时)");
	console.log("  tfa-shell web [--port N]        启动 Web 服务(默认 127.0.0.1:8787)");
	console.log("");
	console.log("常用选项:");
	console.log("  -p, --print                    单发模式,输出结果后退出");
	console.log("  --mode json                    单发模式,输出 JSON 事件流");
	console.log("  --model <id>                   指定模型");
	console.log("  --provider <id>                指定厂商");
	console.log("  --list-models [pattern]        列出可用模型");
	console.log("  -c, --continue                 继续最近的会话");
	console.log("  -r, --resume                   恢复指定会话");
	console.log("  --session <path>               指定会话文件");
	console.log("  --tools <name1,name2>          工具允许名单");
	console.log("  --offline                      离线模式(不刷新模型目录)");
	console.log("  -h, --help                     显示帮助");
	console.log("  -v, --version                  显示版本");
}

/**
 * 壳主入口:argv 分发(web 子命令 / 交互模式 / 单发模式)。
 * 由 cli.ts 调用,返回退出码,不直接 process.exit。
 */
export async function shellMain(argv: string[], options: ShellCliOptions = {}): Promise<number> {
	if (argv.length > 0 && argv[0] === "web") {
		return startWebServer(argv.slice(1), options.shellRuntimeOptions);
	}

	const parsed = parseArgs(argv);
	if (parsed.diagnostics.some((diagnostic) => diagnostic.type === "error")) {
		for (const diagnostic of parsed.diagnostics) {
			console.error(`${diagnostic.type === "error" ? "Error" : "Warning"}: ${diagnostic.message}`);
		}
		return 1;
	}

	if (parsed.version) {
		console.log(SHELL_VERSION);
		return 0;
	}

	if (parsed.help) {
		printShellHelp();
		return 0;
	}

	const cwd = process.cwd();
	const agentDir = getAgentDir();

	// 离线模式:TFA_SHELL_FAUX=1 时注册 faux provider 作为默认模型(无 API key 验证用)
	let fauxSetup: FauxSetup | undefined;
	if (isTruthyEnvFlag(process.env.TFA_SHELL_FAUX)) {
		fauxSetup = setupFauxModel();
		console.log("[tfa-shell] TFA_SHELL_FAUX 已启用:使用 faux provider(离线模式)");
	}

	const runtime = await createAgentSessionRuntime(
		createShellRuntimeFactory({
			cwd,
			agentDir,
			fauxSetup,
			...options.shellRuntimeOptions,
		}),
		{
			cwd,
			agentDir,
			sessionManager: SessionManager.create(cwd),
		},
	);

	if (parsed.listModels !== undefined) {
		const searchPattern = typeof parsed.listModels === "string" ? parsed.listModels : undefined;
		const models = await runtime.services.modelRuntime.getAvailable(undefined, {
			signal: AbortSignal.timeout(15_000),
		});
		const filtered = searchPattern
			? models.filter((model) => `${model.provider} ${model.id}`.includes(searchPattern))
			: models;
		if (filtered.length === 0) {
			console.log("No models available (auth.json / models.json 未配置或未登录)");
		} else {
			console.log(
				`${"provider".padEnd(12)} ${"model".padEnd(32)} ${"context".padEnd(8)} ${"max-out".padEnd(8)} thinking`,
			);
			for (const model of filtered) {
				console.log(
					`${model.provider.padEnd(12)} ${model.id.padEnd(32)} ${String(model.contextWindow).padEnd(8)} ${String(
						model.maxTokens,
					).padEnd(8)} ${model.reasoning ? "yes" : "no"}`,
				);
			}
		}
		return 0;
	}

	const appMode = resolveAppMode(parsed, process.stdin.isTTY, process.stdout.isTTY);

	if (appMode === "rpc") {
		console.error("tfa-shell 不支持 rpc 模式(该模式为编码 agent 内部使用)");
		return 1;
	}

	if (appMode === "interactive") {
		const interactiveMode = new InteractiveMode(runtime, {
			initialMessage: parsed.messages[0],
			initialMessages: parsed.messages.slice(1),
			tuiMode: parsed.tuiMode,
			verbose: parsed.verbose,
		});
		await interactiveMode.run();
		return 0;
	}

	const exitCode = await runPrintMode(runtime, {
		mode: parsed.mode === "json" ? "json" : "text",
		messages: parsed.messages.slice(1),
		initialMessage: parsed.messages[0],
	});
	return exitCode;
}

function isTruthyEnvFlag(value: string | undefined): boolean {
	return value === "1" || value === "true" || value === "yes";
}
