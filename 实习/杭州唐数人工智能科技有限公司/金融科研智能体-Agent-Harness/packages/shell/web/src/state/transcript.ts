/**
 * transcript reducer:镜像 coding-agent/src/client/transcript.ts 的语义。
 * snapshot 为权威基线,progress 增量覆盖;前端渲染 selectTranscript 的结果。
 */

import type { JsonValue, SessionSnapshot, TranscriptItem, TranscriptProgress } from "../api/types.ts";

export interface TranscriptState {
	readonly snapshot: SessionSnapshot | null;
	readonly progressItems: Map<string, TranscriptItem>;
	readonly progressOrder: string[];
	readonly toolCallBuffers: Map<string, string>;
}

export function createTranscriptState(): TranscriptState {
	return { snapshot: null, progressItems: new Map(), progressOrder: [], toolCallBuffers: new Map() };
}

export function applyTranscriptSnapshot(state: TranscriptState, snapshot: SessionSnapshot): TranscriptState {
	if (state.snapshot && state.snapshot.id === snapshot.id && snapshot.revision < state.snapshot.revision) {
		return state;
	}
	return {
		snapshot: structuredClone(snapshot),
		progressItems: new Map(),
		progressOrder: [],
		toolCallBuffers: new Map(),
	};
}

function isJsonValue(value: unknown): value is JsonValue {
	if (value === null || typeof value === "boolean" || typeof value === "string") return true;
	if (typeof value === "number") return Number.isFinite(value);
	if (Array.isArray(value)) return value.every(isJsonValue);
	if (typeof value !== "object" || Object.getPrototypeOf(value) !== Object.prototype) return false;
	return Object.values(value).every(isJsonValue);
}

function parsePartialToolInput(value: string): JsonValue {
	try {
		const parsed: unknown = JSON.parse(value);
		if (isJsonValue(parsed)) return parsed;
	} catch {
		// 工具参数流式未完整时保留原始前缀
	}
	return value;
}

export function applyTranscriptProgress(state: TranscriptState, progress: TranscriptProgress): TranscriptState {
	if (progress.type === "item_started" || progress.type === "item_updated") {
		return setProgressItem(state, progress.item);
	}
	if (progress.type === "item_finished") {
		const toolCallBuffers = new Map(state.toolCallBuffers);
		for (const key of toolCallBuffers.keys()) {
			if (key.startsWith(`${progress.item.id}:`)) toolCallBuffers.delete(key);
		}
		return setProgressItem({ ...state, toolCallBuffers }, progress.item);
	}

	const item =
		state.progressItems.get(progress.messageId) ??
		state.snapshot?.transcript.find((entry) => entry.id === progress.messageId);
	if (!item || item.role !== "assistant") return state;

	let toolCallBuffers = state.toolCallBuffers;
	const content = item.content.map((part, index) => {
		if (index !== progress.contentIndex) return structuredClone(part);
		if (progress.kind === "text" && part.type === "text") {
			return { ...part, text: part.text + progress.delta };
		}
		if (progress.kind === "thinking" && part.type === "thinking") {
			return { ...part, thinking: part.thinking + progress.delta };
		}
		if (progress.kind === "toolCall" && part.type === "toolCall") {
			const key = `${progress.messageId}:${progress.contentIndex}`;
			const existing = state.toolCallBuffers.get(key) ?? (typeof part.input === "string" ? part.input : "");
			const buffer = existing + progress.delta;
			toolCallBuffers = new Map(state.toolCallBuffers).set(key, buffer);
			return { ...part, input: parsePartialToolInput(buffer) };
		}
		return structuredClone(part);
	});
	return setProgressItem({ ...state, toolCallBuffers }, { ...item, content });
}

/** 渲染用的 transcript 视图:snapshot 基底 + progress 增量覆盖 + 新增项追加。 */
export function selectTranscript(state: TranscriptState): TranscriptItem[] {
	if (!state.snapshot) {
		return [...state.progressOrder]
			.map((id) => state.progressItems.get(id))
			.filter((item): item is TranscriptItem => item !== undefined);
	}
	const transcript = state.snapshot.transcript.map((item) => state.progressItems.get(item.id) ?? item);
	const ids = new Set(transcript.map((item) => item.id));
	for (const id of state.progressOrder) {
		if (ids.has(id)) continue;
		const item = state.progressItems.get(id);
		if (item) {
			transcript.push(item);
			ids.add(id);
		}
	}
	for (const item of state.snapshot.queuedSteer) {
		if (ids.has(item.id)) continue;
		transcript.push(item);
		ids.add(item.id);
	}
	return transcript;
}

function setProgressItem(state: TranscriptState, item: TranscriptItem): TranscriptState {
	const progressItems = new Map(state.progressItems);
	const progressOrder = progressItems.has(item.id) ? state.progressOrder : [...state.progressOrder, item.id];
	progressItems.set(item.id, structuredClone(item));
	return { ...state, progressItems, progressOrder };
}
