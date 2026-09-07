<script setup lang="ts">
import { nextTick, ref, watch } from "vue";
import type { TranscriptItem } from "../api/types.ts";

const props = defineProps<{
	items: TranscriptItem[];
	busy: boolean;
}>();

const scrollArea = ref<HTMLElement | null>(null);

watch(
	() => props.items.length,
	async () => {
		await nextTick();
		scrollArea.value?.scrollTo({ top: scrollArea.value.scrollHeight });
	},
);

watch(
	() => JSON.stringify(props.items.at(-1)?.content ?? null),
	async () => {
		await nextTick();
		scrollArea.value?.scrollTo({ top: scrollArea.value.scrollHeight });
	},
);
</script>

<template>
	<div ref="scrollArea" class="chat-view">
		<div class="chat-column">
			<header v-if="$slots.header" class="chat-header">
				<slot name="header" />
			</header>
			<div v-if="items.length === 0" class="chat-empty">
				<h2>开始一次研究</h2>
				<p>输入公司、行业、财报或市场问题。Agent 会调用工具、整理数据并给出可继续追问的结论。</p>
			</div>
			<div v-for="item in items" :key="item.id" class="chat-item">
				<slot name="message" :item="item" />
			</div>
			<div v-if="busy" class="work-status" aria-live="polite">
				<span class="work-spinner" aria-hidden="true" />
				Agent 正在处理，工具调用和研究输出会实时出现在下方。
			</div>
		</div>
	</div>
</template>

<style scoped>
.chat-view {
	flex: 1;
	overflow-y: auto;
	display: flex;
	flex-direction: column;
}

.chat-column {
	width: 100%;
	max-width: 1680px;
	margin: 0 auto;
	padding: 28px clamp(24px, 4vw, 64px) 18px;
	display: flex;
	flex-direction: column;
}

.chat-header {
	margin-bottom: 16px;
}

.chat-empty {
	margin: 36px 0;
	padding: 28px;
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	background: color-mix(in srgb, var(--bg-surface) 68%, transparent);
}

.chat-empty h2 {
	margin: 0 0 8px;
	font-size: 20px;
	font-weight: 700;
	letter-spacing: -0.03em;
	color: var(--text-primary);
}

.chat-empty p {
	margin: 0;
	max-width: 680px;
	font-size: 14px;
	line-height: 1.65;
	color: var(--text-secondary);
}

.work-status {
	display: flex;
	align-items: center;
	gap: 10px;
	margin-top: 18px;
	padding: 12px 14px;
	border: 1px solid color-mix(in srgb, var(--accent) 18%, var(--border));
	border-radius: var(--radius-md);
	background: color-mix(in srgb, var(--accent-soft) 52%, var(--bg-surface));
	color: var(--accent-strong);
	font-size: 13px;
	font-weight: 600;
}

.work-spinner {
	width: 14px;
	height: 14px;
	border: 2px solid color-mix(in srgb, var(--accent) 30%, transparent);
	border-top-color: var(--accent);
	border-radius: 50%;
	animation: spin 900ms linear infinite;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}
.chat-item {
	width: 100%;
}

@media (max-width: 1200px) {
	.chat-column {
		padding: 20px 32px 12px;
	}
}

@media (max-width: 860px) {
	.chat-column {
		padding: 20px 16px 12px;
	}
}
</style>
