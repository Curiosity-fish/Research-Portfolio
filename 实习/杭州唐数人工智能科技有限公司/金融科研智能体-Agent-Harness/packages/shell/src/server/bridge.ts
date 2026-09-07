import type { AgentMessage } from "@earendil-works/tfa-agent-core";
import type { ToolCall } from "@earendil-works/tfa-ai";
import {
	type JsonAgentSessionEvent,
	RpcClient,
	type RpcClientOptions,
	type RpcSessionState,
} from "@earendil-works/tfa-coding-agent";
import type {
	ModelRef,
	SessionPhase,
	SessionSnapshot,
	ThinkingLevel,
	TranscriptItem,
} from "@earendil-works/tfa-protocol";
import {
	type PromptInput,
	type SteerInput,
	type TfaSessionRuntime,
	type TfaSessionRuntimeEvent,
	toProtocolAssistantMessage,
	toProtocolToolResultMessage,
	toProtocolUserMessage,
} from "@earendil-works/tfa-server";

export interface RpcSessionProcessOptions {
	/** tfa CLI 入口(cli.js 或任意可执行脚本)。 */
	cliPath: string;
	/** 会话工作目录(项目目录)。 */
	cwd: string;
	/** 注入环境变量(Ollama 端点 / BYOK key,仅本会话进程)。 */
	env: Record<string, string>;
	provider?: string;
	model?: string;
	/** 额外 CLI 参数:如 --no-skills --skill <dir>。 */
	args?: string[];
}

export interface ProxyRuntimeOptions {
	/** 已启动的 RPC 会话客户端。 */
	client: RpcClient;
	sessionId: string;
	cwd: string;
	/** 收到 RPC approval_request 事件时回调(用于登记审批并转 Web)。 */
	onApprovalRequest?: (request: { id: string; tool: string; target: string; risk: string }) => void;
}

/**
 * 构造 RPC 会话客户端:spawn `node <cliPath> --mode rpc [--provider] [--model] [args]`。
 */
export function createRpcSessionClient(options: RpcSessionProcessOptions): RpcClient {
	const clientOptions: RpcClientOptions = {
		cliPath: options.cliPath,
		cwd: options.cwd,
		env: options.env,
		...(options.provider ? { provider: options.provider } : {}),
		...(options.model ? { model: options.model } : {}),
		...(options.args && options.args.length > 0 ? { args: options.args } : {}),
	};
	return new RpcClient(clientOptions);
}

/**
 * A1 结论:技能注入必须显式 --no-skills + --skill <目录>,禁用默认全局技能目录防跨会话泄漏。
 */
export function buildSkillCliArgs(skillDirs: string[]): string[] {
	const args = ["--no-skills"];
	for (const dir of skillDirs) args.push("--skill", dir);
	return args;
}

function firstMessageTimestamp(messages: AgentMessage[]): number {
	if (messages.length === 0) return Date.now();
	return messages[0].timestamp;
}

function lastMessageTimestamp(messages: AgentMessage[]): number | undefined {
	for (let index = messages.length - 1; index >= 0; index--) {
		const message = messages[index];
		if (message.role !== "toolResult") return message.timestamp;
	}
	return undefined;
}

function proxyPhase(state: RpcSessionState, tracked: SessionPhase): SessionPhase {
	if (state.isCompacting) return "compaction";
	if (state.isStreaming) return "turn";
	return tracked;
}

/** 由 RPC 状态 + 消息构建协议会话快照。 */
function buildProxySnapshot(input: {
	sessionId: string;
	cwd: string;
	state: RpcSessionState;
	messages: AgentMessage[];
	revision: number;
	phase: SessionPhase;
}): SessionSnapshot {
	const toolCallsById = new Map<string, ToolCall>();
	const transcript: TranscriptItem[] = [];
	for (const message of input.messages) {
		if (message.role === "assistant") {
			for (const part of message.content) {
				if (part.type === "toolCall") toolCallsById.set(part.id, part);
			}
		}
	}
	for (let index = 0; index < input.messages.length; index++) {
		const message = input.messages[index];
		const id = `proxy-${message.role}-${index}`;
		if (message.role === "user") {
			transcript.push(toProtocolUserMessage(message, { id }));
		} else if (message.role === "assistant") {
			transcript.push(toProtocolAssistantMessage(message, { id }));
		} else if (message.role === "toolResult") {
			const call = toolCallsById.get(message.toolCallId);
			transcript.push(
				toProtocolToolResultMessage(message, {
					id,
					call: call ?? { type: "toolCall", id: message.toolCallId, name: message.toolName, arguments: {} },
				}),
			);
		}
	}
	const createdAt = firstMessageTimestamp(input.messages);
	const updatedAt = lastMessageTimestamp(input.messages) ?? Date.now();
	return {
		id: input.sessionId,
		name: input.state.sessionName,
		cwd: input.cwd,
		createdAt,
		updatedAt,
		phase: proxyPhase(input.state, input.phase),
		model: {
			provider: input.state.model?.provider ?? "unknown",
			id: input.state.model?.id ?? "unknown",
		},
		thinkingLevel: input.state.thinkingLevel,
		attached: true,
		locked: true,
		revision: input.revision,
		transcript,
		queuedSteer: [],
		queuedSteerCount: 0,
	};
}

/**
 * 会话桥接(设计 §8.3):实现 TfaSessionRuntime,把 prompt/steer/abort/setModel/setThinking
 * 转发到容器/宿主机内的 tfa RPC 进程,事件经现有 SSE 透出。
 * MVP:事件统一映射为 snapshot(前端重取快照);流式 progress 精化留待 T9。
 */
export class ProxyRuntime implements TfaSessionRuntime {
	private readonly client: RpcClient;
	private readonly sessionId: string;
	private readonly cwd: string;
	private readonly onApprovalRequest:
		| ((request: { id: string; tool: string; target: string; risk: string }) => void)
		| undefined;
	private readonly listeners = new Set<(event: TfaSessionRuntimeEvent) => void>();
	private readonly unsubscribe: () => void;
	private phase: SessionPhase = "idle";
	private revision = 0;

	constructor(options: ProxyRuntimeOptions) {
		this.client = options.client;
		this.sessionId = options.sessionId;
		this.cwd = options.cwd;
		this.onApprovalRequest = options.onApprovalRequest;
		this.unsubscribe = this.client.onEvent((event) => {
			this.handleEvent(event);
		});
	}

	private handleEvent(event: JsonAgentSessionEvent): void {
		if (event.type === "agent_start") {
			this.phase = "turn";
		} else if (event.type === "agent_settled") {
			this.phase = "idle";
		} else if ((event as { type?: string }).type === "approval_request" && this.onApprovalRequest) {
			const candidate = event as unknown as { id: string; tool: string; target: string; risk: string };
			this.onApprovalRequest({
				id: candidate.id,
				tool: candidate.tool,
				target: candidate.target,
				risk: candidate.risk,
			});
		}
		this.emit({ type: "snapshot" });
	}

	private emit(event: TfaSessionRuntimeEvent): void {
		for (const listener of this.listeners) listener(event);
	}

	get id(): string {
		return this.sessionId;
	}

	getPhase(): SessionPhase {
		return this.phase;
	}

	async snapshot(): Promise<SessionSnapshot> {
		const [state, messages] = await Promise.all([this.client.getState(), this.client.getMessages()]);
		this.revision += 1;
		return buildProxySnapshot({
			sessionId: this.sessionId,
			cwd: this.cwd,
			state,
			messages,
			revision: this.revision,
			phase: this.getPhase(),
		});
	}

	async prompt(input: PromptInput): Promise<void> {
		await this.client.prompt(input.text);
	}

	async steer(input: SteerInput): Promise<void> {
		await this.client.steer(input.text);
	}

	async abort(): Promise<void> {
		await this.client.abort();
	}

	async setModel(model: ModelRef): Promise<void> {
		await this.client.setModel(model.provider, model.id);
	}

	async setThinking(thinkingLevel: ThinkingLevel): Promise<void> {
		await this.client.setThinkingLevel(thinkingLevel);
	}

	subscribe(listener: (event: TfaSessionRuntimeEvent) => void): () => void {
		this.listeners.add(listener);
		return () => {
			this.listeners.delete(listener);
		};
	}

	/** 向 agent 回传审批决策(需 RpcClient 支持 respondApproval;旧 dist 下静默跳过)。 */
	async decideApproval(rpcId: string, approved: boolean): Promise<void> {
		const responder = (
			this.client as unknown as { respondApproval?: (id: string, approved: boolean) => Promise<void> }
		).respondApproval;
		if (typeof responder === "function") {
			await responder.call(this.client, rpcId, approved);
		}
	}

	async dispose(): Promise<void> {
		this.unsubscribe();
		await this.client.stop();
	}
}
