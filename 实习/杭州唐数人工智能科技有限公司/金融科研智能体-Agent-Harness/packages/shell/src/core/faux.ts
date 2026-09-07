import { type Context, fauxAssistantMessage, type Model } from "@earendil-works/tfa-ai";
import { registerFauxProvider } from "@earendil-works/tfa-ai/compat";
import type { ModelRuntime } from "@earendil-works/tfa-coding-agent";

/** 从消息内容提取纯文本(离线回复的回显用) */
function messageToText(message: { role: string; content: unknown }): string {
	const content = message.content;
	if (typeof content === "string") return content;
	if (Array.isArray(content)) {
		return content
			.map((block) => {
				if (block && typeof block === "object" && "type" in block && block.type === "text") {
					return String((block as { text: unknown }).text);
				}
				return "";
			})
			.join("\n");
	}
	return "";
}

/**
 * faux 离线验证设置:TFA_SHELL_FAUX=1 时注册 faux provider 作为默认模型。
 *
 * registerFauxProvider 只注册 API 到全局 api registry,ModelRuntime 的 auth 检查
 * 仍需通过 registerProvider 注册 provider 配置(带占位 apiKey),模型才会被识别为已认证。
 */
export interface FauxSetup {
	/** 默认模型(显式传入 createAgentSession,不查 auth.json) */
	defaultModel: Model<any>;
	/** 把 faux provider 注册到 ModelRuntime */
	applyTo(modelRuntime: ModelRuntime): void;
	/** 反注册(进程退出/测试清理) */
	unregister(): void;
}

export function setupFauxModel(options: { tokensPerSecond?: number } = {}): FauxSetup {
	const faux = registerFauxProvider(
		options.tokensPerSecond === undefined ? {} : { tokensPerSecond: options.tokensPerSecond },
	);
	const defaultModel = faux.getModel("faux-1");
	if (!defaultModel) {
		throw new Error("faux provider 未提供默认模型");
	}
	// 默认动态响应:回显最后一条用户消息,保证离线对话不因响应耗尽而报错
	faux.setResponses([
		(context: Context): ReturnType<typeof fauxAssistantMessage> => {
			const lastUser = [...context.messages].reverse().find((message) => message.role === "user");
			const echoed = lastUser ? messageToText(lastUser) : "(无用户消息)";
			return fauxAssistantMessage(`[faux 离线回复] 收到:${echoed}`);
		},
	]);
	return {
		defaultModel,
		applyTo(modelRuntime) {
			modelRuntime.registerProvider(defaultModel.provider, {
				baseUrl: defaultModel.baseUrl,
				apiKey: "faux-key",
				api: faux.api,
				models: faux.models.map((model) => ({
					id: model.id,
					name: model.name,
					api: model.api,
					reasoning: model.reasoning,
					input: model.input,
					cost: model.cost,
					contextWindow: model.contextWindow,
					maxTokens: model.maxTokens,
					baseUrl: model.baseUrl,
				})),
			});
		},
		unregister() {
			faux.unregister();
		},
	};
}
