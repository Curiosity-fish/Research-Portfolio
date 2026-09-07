<script setup lang="ts">
import { computed } from "vue";
import type { ToolTranscriptItem } from "../api/types.ts";

const props = defineProps<{
	item: ToolTranscriptItem;
}>();

const statusLabel = computed(() => {
	switch (props.item.status) {
		case "running":
			return "执行中";
		case "complete":
			return props.item.isError ? "失败" : "完成";
		case "error":
			return "错误";
	}
});

const resultText = computed(() => {
	return props.item.content
		.map((part) => (part.type === "text" ? part.text : `[图片 ${part.mimeType}]`))
		.join("\n");
});

const inputText = computed(() => {
	if (props.item.input === null || typeof props.item.input !== "object") {
		return JSON.stringify(props.item.input);
	}
	return JSON.stringify(props.item.input, null, 2);
});
</script>

<template>
	<div class="tool-card" :class="{ running: item.status === 'running', error: item.isError }">
		<div class="tool-header">
			<span class="tool-name mono">{{ item.toolName }}</span>
			<span class="tool-status" :class="{ running: item.status === 'running' }">{{ statusLabel }}</span>
		</div>
		<pre v-if="inputText && inputText !== '{}'" class="tool-input">{{ inputText }}</pre>
		<div v-if="resultText" class="tool-result">
			<div class="tool-result-label">结果</div>
			<pre class="tool-result-text">{{ resultText }}</pre>
		</div>
	</div>
</template>

<style scoped>
.tool-card {
	border: 1px solid var(--border);
	border-radius: var(--radius-md);
	padding: 12px 14px;
	margin: 8px 0 8px 40px;
	background: var(--bg-surface);
	font-size: 12px;
	transition: border-color 140ms ease;
}

.tool-card.running {
	border-color: var(--accent);
}

.tool-card.error {
	border-color: var(--danger);
}

.tool-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 8px;
}

.tool-name {
	font-size: 12px;
	font-weight: 600;
	color: var(--text-secondary);
}

.tool-status {
	font-size: 11px;
	font-weight: 500;
	color: var(--text-muted);
	background: var(--bg-hover);
	padding: 2px 8px;
	border-radius: var(--radius-pill);
}

.tool-status.running {
	color: var(--accent);
	background: var(--accent-soft);
}

.tool-input {
	margin: 10px 0 0;
	background: var(--bg-input);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 10px 12px;
	font-size: 11.5px;
	line-height: 1.55;
	overflow-x: auto;
	color: var(--text-secondary);
	font-family: var(--font-mono);
}

.tool-result {
	margin-top: 10px;
}

.tool-result-label {
	font-size: 11px;
	font-weight: 500;
	color: var(--text-muted);
	margin-bottom: 6px;
}

.tool-result-text {
	margin: 0;
	background: var(--bg-input);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 10px 12px;
	font-size: 11.5px;
	line-height: 1.55;
	white-space: pre-wrap;
	overflow-wrap: anywhere;
	color: var(--text-primary);
	font-family: var(--font-mono);
}

@media (max-width: 640px) {
	.tool-card {
		margin-left: 0;
	}
}
</style>
