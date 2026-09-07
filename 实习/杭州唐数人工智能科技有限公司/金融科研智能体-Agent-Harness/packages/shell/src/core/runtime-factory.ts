import type { Model } from "@earendil-works/tfa-ai";
import {
	type CreateAgentSessionRuntimeFactory,
	createAgentSessionFromServices,
	createAgentSessionServices,
	type InlineExtension,
	SettingsManager,
	type ToolDefinition,
} from "@earendil-works/tfa-coding-agent";
import type { FauxSetup } from "./faux.ts";

export interface ShellRuntimeOptions {
	/** 默认工作目录,默认 process.cwd() */
	cwd: string;
	/** 全局配置目录,默认 getAgentDir() */
	agentDir?: string;
	/** 自定义(金融)工具,经 createAgentSession 的 customTools 注入 */
	customTools?: ToolDefinition[];
	/** 工具允许名单;提供时只启用名单内工具 */
	toolAllowlist?: string[];
	/** 默认模型(离线 faux 验证用;显式传入时 SDK 不查 auth.json) */
	defaultModel?: Model<any>;
	/** faux 离线验证设置(见 core/faux.ts) */
	fauxSetup?: FauxSetup;
	/** 扩展工厂 */
	extensionFactories?: InlineExtension[];
	/** 附加扩展路径(.tfa/extensions/) */
	additionalExtensionPaths?: string[];
}

/**
 * 创建会话 runtime 工厂:复用 SDK 的 createAgentSessionServices/FromServices。
 *
 * 壳统一信任项目资源(projectTrusted: true),避免 Web 服务场景出现交互式信任提示。
 * 金融工具经 customTools 注入、经 toolAllowlist 控制启用。
 */
export function createShellRuntimeFactory(options: ShellRuntimeOptions): CreateAgentSessionRuntimeFactory {
	return async ({ cwd, agentDir, sessionManager, sessionStartEvent }) => {
		const settingsManager = SettingsManager.create(cwd, agentDir, { projectTrusted: true });
		const services = await createAgentSessionServices({
			cwd,
			agentDir,
			settingsManager,
			resourceLoaderOptions: {
				...(options.additionalExtensionPaths ? { additionalExtensionPaths: options.additionalExtensionPaths } : {}),
				...(options.extensionFactories ? { extensionFactories: options.extensionFactories } : {}),
			},
		});
		options.fauxSetup?.applyTo(services.modelRuntime);
		const model = options.fauxSetup?.defaultModel ?? options.defaultModel;
		const created = await createAgentSessionFromServices({
			services,
			sessionManager,
			sessionStartEvent,
			...(model ? { model } : {}),
			...(options.toolAllowlist ? { tools: options.toolAllowlist } : {}),
			...(options.customTools ? { customTools: options.customTools } : {}),
		});
		return {
			...created,
			services,
			diagnostics: services.diagnostics,
		};
	};
}
