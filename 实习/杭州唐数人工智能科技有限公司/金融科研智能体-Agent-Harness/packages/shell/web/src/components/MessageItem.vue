<script setup lang="ts">
import { computed, ref } from "vue";
import type { TranscriptItem } from "../api/types.ts";
import AppIcon from "./AppIcon.vue";
import ToolCallCard from "./ToolCallCard.vue";

const props = defineProps<{
	item: TranscriptItem;
}>();

const thinkingExpanded = ref(false);
const copied = ref(false);

function escapeHtml(value: string): string {
	return value
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#39;");
}

function renderInline(value: string): string {
	let out = escapeHtml(value);
	out = out.replace(/`([^`]+)`/g, (_match, code: string) => `<code>${code}</code>`);
	out = out.replace(/\*\*([^*]+)\*\*/g, (_match, text: string) => `<strong>${text}</strong>`);
	out = out.replace(/\*([^*]+)\*/g, (_match, text: string) => `<em>${text}</em>`);
	out = out.replace(/\[([^\]]+)\]\((https?:[^)\s]+)\)/g, (_match, label: string, url: string) =>
		`<a href="${url}" target="_blank" rel="noreferrer">${label}</a>`,
	);
	return out;
}

function isTableSeparator(cells: string[]): boolean {
	return cells.length > 0 && cells.every((cell) => /^:?-{2,}:?$/.test(cell));
}

function renderMarkdown(value: string): string {
	const escaped = escapeHtml(value);
	const blocks = escaped.split(/(```[\s\S]*?```)/g);
	return blocks
		.map((block) => {
			if (block.startsWith("```")) {
				const body = block.slice(3, -3).replace(/^\n/, "");
				return `<pre class="markdown-code"><code>${body}</code></pre>`;
			}
			const lines = block.split("\n");
			let html = "";
			let inList = false;
			let listTag: "ul" | "ol" | null = null;
			let tableHeader: string[] | null = null;
			const tableRows: string[][] = [];

			const flushTable = () => {
				if (!tableHeader && tableRows.length === 0) return;
				const head = tableHeader
					? `<thead><tr>${tableHeader.map((cell) => `<th>${cell}</th>`).join("")}</tr></thead>`
					: "";
				const body =
					tableRows.length > 0
						? `<tbody>${tableRows
								.map((row) => `<tr>${row.map((cell) => `<td>${cell}</td>`).join("")}</tr>`)
								.join("")}</tbody>`
						: "";
				html += `<div class="table-wrap"><table>${head}${body}</table></div>`;
				tableHeader = null;
				tableRows.length = 0;
			};

			const flushList = () => {
				if (inList) {
					html += `</${listTag}>`;
					inList = false;
					listTag = null;
				}
			};

			for (const line of lines) {
				const tableMatch = line.match(/^\|(.+)\|$/);
				if (tableMatch) {
					flushList();
					const cells = tableMatch[1].split("|").map((cell) => cell.trim());
					if (isTableSeparator(cells)) continue;
					if (tableHeader) {
						tableRows.push(cells);
					} else {
						tableHeader = cells;
					}
					continue;
				}
				flushTable();

				const heading = line.match(/^#{1,4}\s/);
				if (heading) {
					flushList();
					const level = heading[0].length - 1;
					html += `<h${level}>${renderInline(line.replace(/^#{1,4}\s/, ""))}</h${level}>`;
					continue;
				}
				if (/^\s*[-*]\s/.test(line)) {
					if (!inList || listTag !== "ul") {
						flushList();
						html += "<ul>";
						inList = true;
						listTag = "ul";
					}
					html += `<li>${renderInline(line.replace(/^\s*[-*]\s/, ""))}</li>`;
					continue;
				}
				if (/^\s*\d+\.\s/.test(line)) {
					if (!inList || listTag !== "ol") {
						flushList();
						html += "<ol>";
						inList = true;
						listTag = "ol";
					}
					html += `<li>${renderInline(line.replace(/^\s*\d+\.\s/, ""))}</li>`;
					continue;
				}
				if (/^\s*>\s?/.test(line)) {
					flushList();
					html += `<blockquote>${renderInline(line.replace(/^\s*>\s?/, ""))}</blockquote>`;
					continue;
				}
				flushList();
				html += line ? `<p>${renderInline(line)}</p>` : "";
			}
			flushTable();
			flushList();
			return html;
		})
		.join("");
}

const assistantText = computed(() => {
	if (props.item.role !== "assistant") return "";
	return props.item.content
		.filter((part) => part.type === "text")
		.map((part) => part.text ?? "")
		.join("\n");
});

function textOf(content: string | Array<{ type: string; text?: string }>): string {
	if (typeof content === "string") return content;
	return content
		.filter((part) => part.type === "text")
		.map((part) => part.text ?? "")
		.join("\n");
}

function formatTime(timestamp: number): string {
	return new Date(timestamp).toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
}

async function copyAssistant(): Promise<void> {
	if (!assistantText.value) return;
	try {
		await navigator.clipboard.writeText(assistantText.value);
		copied.value = true;
		setTimeout(() => {
			copied.value = false;
		}, 1500);
	} catch {
		// 剪贴板不可用时静默失败
	}
}
</script>

<template>
	<div class="message" :class="`role-${item.role}`">
		<template v-if="item.role === 'user'">
			<div class="user-bubble">{{ textOf(item.content) }}</div>
		</template>

		<template v-else-if="item.role === 'assistant'">
			<div class="assistant-block">
				<div class="assistant-rail" aria-hidden="true">
					<div class="assistant-avatar">
						<AppIcon name="trending" :size="15" />
					</div>
				</div>
				<div class="research-answer">
					<div class="assistant-head">
					<span class="assistant-label">
						Agent
						<span v-if="item.status === 'streaming'" class="streaming-dot" />
					</span>
						<span v-if="item.responseModel || item.model?.id" class="assistant-model mono">
							{{ item.responseModel || item.model?.id }}
						</span>
					<button v-if="assistantText" class="copy-button" type="button" @click="copyAssistant">
						<AppIcon :name="copied ? 'check' : 'more'" :size="13" />
						{{ copied ? "已复制" : "复制" }}
					</button>
				</div>
					<template v-for="(part, index) in item.content" :key="index">
						<div v-if="part.type === 'thinking'" class="thinking-block">
							<button class="thinking-toggle" type="button" @click="thinkingExpanded = !thinkingExpanded">
								<span class="thinking-icon" aria-hidden="true">{{ thinkingExpanded ? "▾" : "▸" }}</span>
								思考过程
							</button>
							<pre v-if="thinkingExpanded" class="thinking-text">{{ part.thinking }}</pre>
						</div>
						<div v-else-if="part.type === 'toolCall'" class="tool-call-inline">
							<span class="tool-call-name">工具 {{ part.toolName }}</span>
							<span class="tool-call-state" :class="{ streaming: item.status === 'streaming' }">
								{{ item.status === "streaming" ? "调用中…" : "已调用" }}
							</span>
						</div>
						<div v-else class="assistant-text" v-html="renderMarkdown(part.text ?? '')" />
					</template>
					<div class="answer-footer">
						<span class="usage-row mono">{{ item.usage?.totalTokens ? `${item.usage.totalTokens} tokens` : "研究输出" }}</span>
						<span class="answer-time">{{ formatTime(item.timestamp) }}</span>
					</div>
					<div v-if="item.status === 'error' && item.errorMessage" class="error-text">
						错误：{{ item.errorMessage }}
					</div>
				</div>
			</div>
		</template>

		<template v-else>
			<ToolCallCard :item="item" />
		</template>
	</div>
</template>

<style scoped>
.message {
	margin: 16px 0;
}

/* 用户消息:靠右浅灰气泡 */
.message.role-user {
	display: flex;
	justify-content: flex-end;
}

.user-bubble {
	background: var(--user-bubble);
	color: var(--text-black);
	border: 1px solid var(--border);
	border-radius: var(--radius-md) var(--radius-md) 4px var(--radius-md);
	padding: 10px 14px;
	max-width: 84%;
	white-space: pre-wrap;
	font-size: 14px;
	line-height: 1.65;
	overflow-wrap: anywhere;
}

/* Agent 消息:靠左 Avatar + 正文 */
.assistant-block {
	display: flex;
	gap: 12px;
	align-items: flex-start;
}

.assistant-avatar {
	width: 28px;
	height: 28px;
	flex-shrink: 0;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 9px;
	background: var(--accent-soft);
	color: var(--accent);
	margin-top: 2px;
}

.assistant-head {
	display: flex;
	align-items: center;
	gap: 8px;
	margin-bottom: 2px;
}

.assistant-label {
	display: inline-flex;
	align-items: center;
	gap: 6px;
	font-size: 12px;
	font-weight: 600;
	color: var(--text-secondary);
	letter-spacing: 0.01em;
}

.streaming-dot {
	width: 7px;
	height: 7px;
	border-radius: 50%;
	background: var(--accent);
	animation: pulse 1.2s ease-in-out infinite;
}

@keyframes pulse {
	0%,
	100% {
		opacity: 1;
	}
	50% {
		opacity: 0.35;
	}
}

.copy-button {
	display: inline-flex;
	align-items: center;
	gap: 5px;
	background: none;
	border: none;
	border-radius: 6px;
	color: var(--text-muted);
	font-size: 11px;
	height: 24px;
	padding: 0 6px;
	cursor: pointer;
	transition:
		color 120ms ease,
		background 120ms ease;
}

.copy-button:hover {
	color: var(--text-primary);
	background: var(--bg-hover);
}

.assistant-body {
	min-width: 0;
	flex: 1;
}

.assistant-text {
	margin: 4px 0;
	font-size: 14.5px;
	line-height: 1.75;
	color: var(--text-black);
	overflow-wrap: break-word;
	text-wrap: pretty;
	max-width: 96%;
}

.assistant-text :deep(h1),
.assistant-text :deep(h2),
.assistant-text :deep(h3),
.assistant-text :deep(h4) {
	margin: 18px 0 8px;
	font-weight: 600;
	line-height: 1.4;
	letter-spacing: -0.01em;
}

.assistant-text :deep(h1) {
	font-size: 18px;
}

.assistant-text :deep(h2) {
	font-size: 16px;
}

.assistant-text :deep(h3),
.assistant-text :deep(h4) {
	font-size: 15px;
}

.assistant-text :deep(p) {
	margin: 8px 0;
}

.assistant-text :deep(ul),
.assistant-text :deep(ol) {
	margin: 8px 0;
	padding-left: 24px;
}

.assistant-text :deep(li) {
	margin: 4px 0;
}

.assistant-text :deep(a) {
	color: var(--info);
	text-decoration: none;
	border-bottom: 1px solid transparent;
}

.assistant-text :deep(a:hover) {
	border-bottom-color: currentColor;
}

.assistant-text :deep(code) {
	background: var(--bg-hover);
	border: 1px solid var(--border);
	border-radius: 5px;
	padding: 1px 5px;
	font-size: 12.5px;
}

.assistant-text :deep(blockquote) {
	margin: 10px 0;
	padding: 2px 0 2px 14px;
	border-left: 3px solid var(--border-strong);
	color: var(--text-secondary);
}

.assistant-text :deep(strong) {
	font-weight: 600;
}

.markdown-code {
	background: var(--bg-input);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 12px 14px;
	overflow-x: auto;
	font-size: 12.5px;
	line-height: 1.6;
	margin: 10px 0;
}

.table-wrap {
	margin: 12px 0;
	overflow-x: auto;
	border: 1px solid var(--table-border);
	border-radius: var(--radius-sm);
}

.assistant-text :deep(table) {
	width: 100%;
	border-collapse: collapse;
	font-size: 13px;
	line-height: 1.55;
	color: var(--text-primary);
}

.assistant-text :deep(th),
.assistant-text :deep(td) {
	border-bottom: 1px solid var(--table-border);
	padding: 8px 12px;
	text-align: left;
	vertical-align: top;
}

.assistant-text :deep(thead th) {
	background: var(--table-head-bg);
	font-weight: 600;
	color: var(--text-secondary);
	white-space: nowrap;
}

.assistant-text :deep(tbody tr:hover td) {
	background: var(--table-hover);
}

.assistant-text :deep(tbody tr:last-child td) {
	border-bottom: none;
}

.thinking-block {
	margin: 8px 0;
	border-left: 2px solid var(--border-strong);
	padding-left: 12px;
}

.thinking-toggle {
	background: none;
	border: none;
	color: var(--text-muted);
	font-size: 12px;
	font-weight: 500;
	cursor: pointer;
	padding: 3px 0;
	display: inline-flex;
	align-items: center;
	gap: 5px;
}

.thinking-toggle:hover {
	color: var(--text-secondary);
}

.thinking-icon {
	color: var(--text-muted);
	font-size: 10px;
}

.thinking-text {
	margin: 6px 0 0;
	white-space: pre-wrap;
	font-size: 12px;
	line-height: 1.6;
	color: var(--text-muted);
	font-family: var(--font-mono);
}

.tool-call-inline {
	display: flex;
	align-items: center;
	gap: 8px;
	margin: 8px 0;
	font-size: 12px;
}

.tool-call-name {
	color: var(--text-secondary);
	font-family: var(--font-mono);
}

.tool-call-state {
	color: var(--text-muted);
	font-size: 11px;
}

.tool-call-state.streaming {
	color: var(--accent);
}

.usage-row {
	font-size: 11px;
	color: var(--text-muted);
	margin-top: 10px;
	padding-top: 8px;
	border-top: 1px solid var(--border);
}

.error-text {
	color: var(--danger);
	font-size: 13px;
	margin-top: 8px;
	background: var(--danger-bg);
	border-radius: var(--radius-sm);
	padding: 8px 10px;
}

@media (max-width: 640px) {
	.user-bubble {
		max-width: 88%;
	}

	.assistant-text {
		font-size: 14px;
	}
}

/* Research Workbench answer overrides */
.message { margin: 20px 0; }
.message.role-user { display: flex; justify-content: flex-end; }
.user-bubble {
	max-width: min(680px, 84%);
	background: color-mix(in srgb, var(--accent-soft) 72%, var(--bg-surface));
	border-color: color-mix(in srgb, var(--accent) 18%, var(--border));
	box-shadow: 0 1px 2px rgba(16, 28, 48, 0.04);
	font-weight: 500;
}

.assistant-block {
	display: grid;
	grid-template-columns: 34px minmax(0, 1fr);
	gap: 12px;
}

.assistant-rail { display: flex; justify-content: center; padding-top: 2px; }
.assistant-avatar {
	width: 32px;
	height: 32px;
	border: 1px solid color-mix(in srgb, var(--accent) 18%, var(--border));
	background: linear-gradient(135deg, color-mix(in srgb, var(--accent) 92%, #fff), #6c99ff);
	color: #fff;
	box-shadow: 0 1px 2px rgba(16, 28, 48, 0.08);
}

.research-answer {
	min-width: 0;
	padding: 16px 18px;
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	background: var(--bg-surface);
	box-shadow: var(--shadow-card);
}

.assistant-head {
	display: flex;
	align-items: center;
	gap: 8px;
	margin: 0 0 12px;
	padding-bottom: 10px;
	border-bottom: 1px solid var(--border);
}

.assistant-label {
	color: var(--accent-strong);
	font-size: 11px;
	font-weight: 750;
	letter-spacing: 0.08em;
	text-transform: uppercase;
}

.assistant-model {
	max-width: 220px;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	font-size: 10px;
	color: var(--text-muted);
}

.copy-button { margin-left: auto; height: 26px; border: 1px solid transparent; }
.copy-button:hover { border-color: var(--border); background: var(--bg-surface); }
.assistant-body { min-width: 0; }
.assistant-text { max-width: none; font-size: 14px; line-height: 1.72; color: var(--text-primary); }
.assistant-text :deep(h1), .assistant-text :deep(h2) { letter-spacing: -0.025em; }
.assistant-text :deep(h1) { font-size: 20px; margin-top: 20px; }
.assistant-text :deep(h2) { font-size: 17px; margin-top: 18px; padding-bottom: 5px; border-bottom: 1px solid var(--border); }
.assistant-text :deep(h3), .assistant-text :deep(h4) { font-size: 15px; letter-spacing: -0.015em; }
.assistant-text :deep(a) { color: var(--accent); }
.markdown-code { border-color: color-mix(in srgb, var(--accent) 10%, var(--border)); background: color-mix(in srgb, var(--bg-app) 54%, var(--bg-surface)); }
.table-wrap { box-shadow: 0 1px 2px rgba(16, 28, 48, 0.04); }
.assistant-text :deep(th), .assistant-text :deep(td) { padding: 9px 11px; font-size: 13px; }
.assistant-text :deep(thead th) { font-size: 11px; letter-spacing: 0.04em; text-transform: uppercase; }
.thinking-block { border-left: 2px solid color-mix(in srgb, var(--accent) 45%, var(--border)); padding-left: 12px; }
.thinking-toggle { color: var(--text-secondary); font-weight: 600; }
.tool-call-inline { padding: 7px 9px; border: 1px solid var(--border); border-radius: var(--radius-sm); background: color-mix(in srgb, var(--bg-app) 62%, var(--bg-surface)); }
.answer-footer { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: 14px; padding-top: 10px; border-top: 1px solid var(--border); }
.usage-row { margin: 0; padding: 0; border: 0; font-size: 10px; letter-spacing: 0.06em; text-transform: uppercase; color: var(--text-muted); }
.answer-time { font-size: 10px; color: var(--text-muted); }

@media (max-width: 640px) {
	.assistant-block { grid-template-columns: minmax(0, 1fr); }
	.assistant-rail { display: none; }
	.research-answer { padding: 14px; }
	.assistant-model { display: none; }
}
</style>
