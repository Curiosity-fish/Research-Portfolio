import type { AgentMessage } from "@earendil-works/tfa-agent-core";
import type { ToolCall } from "@earendil-works/tfa-ai";
import type { AgentSession } from "@earendil-works/tfa-coding-agent";
import type { SessionPhase, SessionSnapshot, TranscriptItem } from "@earendil-works/tfa-protocol";
import {
	toProtocolAssistantMessage,
	toProtocolToolResultMessage,
	toProtocolUserMessage,
} from "@earendil-works/tfa-server";

/**
 * 会话 phase 判定:与协议 SessionPhase 词汇对齐。
 */
export function getSessionPhase(session: AgentSession): SessionPhase {
	if (session.isCompacting) return "compaction";
	if (session.isRetrying) return "retry";
	if (session.isStreaming) return "turn";
	return "idle";
}

/**
 * 确定性消息 id:user/assistant 按 transcript 索引,tool 按 toolCallId。
 * snapshot 与增量事件用同一算法,保证 id 一致。
 */
export function transcriptItemId(message: AgentMessage, index: number): string {
	if (message.role === "toolResult") {
		return `tool-${message.toolCallId}`;
	}
	return `${message.role}-${index}`;
}

/**
 * 解析消息在 transcript 中的确定性 id。
 * agent 在 message_end 时才把消息 push 进 state.messages,流式期间消息仅存在于
 * state.streamingMessage——此时其最终位置是 messages.length,id 与入列后一致。
 */
export function resolveTranscriptItemId(session: AgentSession, message: AgentMessage): string | undefined {
	const index = session.messages.indexOf(message);
	if (index >= 0) {
		return transcriptItemId(message, index);
	}
	if (session.state.streamingMessage === message) {
		return transcriptItemId(message, session.messages.length);
	}
	return undefined;
}

/**
 * 构建会话快照:把 agent 状态中的 messages 经协议桥映射为 TranscriptItem 数组。
 */
export function buildSessionSnapshot(session: AgentSession, revision: number): SessionSnapshot {
	const messages = session.messages;
	const toolCallsById = new Map<string, ToolCall>();
	const transcript: TranscriptItem[] = [];

	for (const message of messages) {
		if (message.role === "assistant") {
			for (const part of message.content) {
				if (part.type === "toolCall") {
					toolCallsById.set(part.id, part);
				}
			}
		}
	}

	for (let index = 0; index < messages.length; index++) {
		const message = messages[index];
		const id = transcriptItemId(message, index);
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

	const header = session.sessionManager.getHeader();
	const createdAt = header ? parseHeaderTimestamp(header.timestamp) : firstMessageTimestamp(messages);
	const updatedAt = lastMessageTimestamp(messages) ?? Date.now();

	return {
		id: session.sessionId,
		name: session.sessionName,
		cwd: session.sessionManager.getCwd(),
		createdAt,
		updatedAt,
		phase: getSessionPhase(session),
		model: {
			provider: session.model?.provider ?? "unknown",
			id: session.model?.id ?? "unknown",
		},
		thinkingLevel: session.thinkingLevel,
		attached: true,
		locked: true,
		revision,
		transcript,
		queuedSteer: session.getSteeringMessages().map((text, index) => ({
			id: `steer-${index}`,
			role: "user" as const,
			content: [{ type: "text" as const, text }],
			timestamp: Date.now(),
		})),
		queuedSteerCount: 0,
	};
}

function parseHeaderTimestamp(timestamp: string): number {
	const parsed = Date.parse(timestamp);
	return Number.isFinite(parsed) ? parsed : Date.now();
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
