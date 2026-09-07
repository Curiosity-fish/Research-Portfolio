<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from "vue";
import type { ModelMetadata, ModelRef } from "../api/types.ts";
import AppIcon from "./AppIcon.vue";
import ModelPicker from "./ModelPicker.vue";

const props = withDefaults(
	defineProps<{
		busy: boolean;
		models: ModelMetadata[];
		current: ModelRef | null;
		modelValue?: string;
		placeholder?: string;
		autofocus?: boolean;
	}>(),
	{
		modelValue: "",
		placeholder: "输入你的金融研究问题……",
		autofocus: false,
	},
);

const emit = defineEmits<{
	send: [text: string];
	steer: [text: string];
	abort: [];
	"update:modelValue": [value: string];
	"model-change": [model: ModelRef];
	"open-tools": [];
}>();

const textareaEl = ref<HTMLTextAreaElement | null>(null);
const composing = ref(false);

const hasText = () => props.modelValue.trim().length > 0;
const canSend = () => hasText() && !props.busy;
const canSteer = () => hasText() && props.busy;

function resize(): void {
	const el = textareaEl.value;
	if (!el) return;
	el.style.height = "auto";
	el.style.height = `${Math.min(el.scrollHeight, 128)}px`;
}

function onInput(event: Event): void {
	emit("update:modelValue", (event.target as HTMLTextAreaElement).value);
	resize();
}

function submit(): void {
	if (!canSend()) return;
	emit("send", props.modelValue.trim());
	emit("update:modelValue", "");
	nextTick(resize);
}

function submitSteer(): void {
	if (!canSteer()) return;
	emit("steer", props.modelValue.trim());
	emit("update:modelValue", "");
	nextTick(resize);
}

function onKeydown(event: KeyboardEvent): void {
	if (event.key === "Enter" && !event.shiftKey && !composing.value) {
		event.preventDefault();
		if (props.busy) {
			submitSteer();
		} else {
			submit();
		}
	}
}

function focus(): void {
	textareaEl.value?.focus();
}

defineExpose({ focus });

onMounted(() => {
	resize();
	if (props.autofocus) focus();
});

watch(
	() => props.modelValue,
	() => {
		nextTick(resize);
	},
);
</script>

<template>
	<div class="composer">
		<textarea
			ref="textareaEl"
			:value="modelValue"
			name="prompt"
			rows="1"
			aria-label="消息输入"
			:placeholder="busy ? '生成中，输入指令可插话……' : placeholder"
			@input="onInput"
			@keydown="onKeydown"
			@compositionstart="composing = true"
			@compositionend="composing = false"
		/>
		<div class="composer-toolbar">
			<div class="composer-toolbar-left">
				<button class="composer-action" type="button" disabled aria-label="添加附件（开发中）" title="附件功能开发中">
					<AppIcon name="paperclip" :size="16" />
					<span class="composer-action-label">附件</span>
				</button>
				<button class="composer-action" type="button" aria-label="打开金融工具" title="查看已注册的金融工具" @click="emit('open-tools')">
					<AppIcon name="wrench" :size="16" />
					<span class="composer-action-label">工具</span>
				</button>
			</div>
			<div class="composer-toolbar-right">
				<ModelPicker :models="models" :current="current" @change="emit('model-change', $event)" />
				<template v-if="busy">
					<button
						class="composer-send composer-send--steer"
						type="button"
						:disabled="!canSteer()"
						aria-label="插话发送"
						title="生成中插话"
						@click="submitSteer"
					>
						<AppIcon name="arrow-up" :size="18" />
					</button>
					<button
						class="composer-send composer-send--stop"
						type="button"
						aria-label="停止生成"
						title="停止生成"
						@click="emit('abort')"
					>
						<AppIcon name="stop" :size="17" />
					</button>
				</template>
				<button
					v-else
					class="composer-send"
					type="button"
					:disabled="!canSend()"
					aria-label="发送"
					title="发送"
					@click="submit"
				>
					<AppIcon name="arrow-up" :size="18" />
				</button>
			</div>
		</div>
	</div>
</template>

<style scoped>
.composer {
	width: 100%;
	min-height: 110px;
	max-height: 160px;
	display: flex;
	flex-direction: column;
	background: var(--bg-surface);
	border: 1px solid var(--border);
	border-radius: 10px;
	padding: 14px 14px 10px;
	box-shadow: none;
	transition:
		border-color 160ms ease,
		box-shadow 160ms ease;
}

.composer:focus-within {
	border-color: var(--accent);
	box-shadow: 0 0 0 4px rgba(116, 185, 235, 0.1);
}

textarea {
	width: 100%;
	background: transparent;
	color: var(--text-black);
	border: none;
	outline: none;
	resize: none;
	font-size: 15px;
	line-height: 1.6;
	padding: 2px 2px 8px;
	flex: 1;
	min-height: 40px;
}

textarea::placeholder {
	color: var(--text-disabled);
	font-weight: 400;
}

.composer-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 12px;
	padding-top: 8px;
	border-top: 1px solid var(--border);
}

.composer-toolbar-left,
.composer-toolbar-right {
	display: flex;
	align-items: center;
	gap: 4px;
	min-width: 0;
}

.composer-action {
	display: inline-flex;
	align-items: center;
	gap: 6px;
	height: 32px;
	padding: 0 10px;
	border: none;
	border-radius: var(--radius-sm);
	background: transparent;
	color: var(--text-secondary);
	font-size: 13px;
	font-weight: 400;
	cursor: pointer;
	transition:
		background 150ms ease,
		color 150ms ease;
}

.composer-action svg {
	color: var(--text-muted);
}

.composer-action:hover:not(:disabled) {
	background: var(--bg-hover);
	color: var(--text-primary);
}

.composer-action:hover:not(:disabled) svg {
	color: var(--accent-strong);
}

.composer-action:disabled {
	opacity: 0.45;
	cursor: not-allowed;
}

.composer-send {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	width: 38px;
	height: 38px;
	flex-shrink: 0;
	border: none;
	border-radius: 8px;
	background: var(--accent-strong);
	color: #ffffff;
	cursor: pointer;
	transition:
		background 150ms ease,
		opacity 150ms ease;
}

.composer-send:hover:not(:disabled) {
	background: #0052ff;
}

.composer-send:active:not(:disabled) {
	background: #0044dd;
}

.composer-send:disabled {
	opacity: 0.45;
	cursor: not-allowed;
}

.composer-send--steer {
	background: transparent;
	color: var(--accent-strong);
	border: 1px solid var(--accent);
}

.composer-send--steer:hover:not(:disabled) {
	background: var(--accent-soft);
	color: var(--accent-strong);
}

.composer-send--stop {
	background: var(--danger);
}

.composer-send--stop:hover:not(:disabled) {
	background: var(--danger-strong);
}

@media (max-width: 640px) {
	.composer-action-label {
		display: none;
	}

	.composer {
		padding: 12px 12px 8px;
		min-height: 96px;
	}
}

/* Research Cockpit command composer overrides */
.composer {
	min-height: 124px;
	border-radius: var(--radius-lg);
	border-color: color-mix(in srgb, var(--accent) 10%, var(--border));
	box-shadow: var(--shadow-composer);
	padding: 15px 15px 11px;
}

.composer:focus-within {
	border-color: color-mix(in srgb, var(--accent) 62%, var(--border));
	box-shadow: 0 0 0 4px color-mix(in srgb, var(--accent) 12%, transparent);
}

textarea {
	color: var(--text-primary);
	font-size: 16px;
	line-height: 1.62;
	padding: 0 0 10px;
}

textarea::placeholder {
	color: var(--text-muted);
}

.composer-toolbar {
	padding-top: 10px;
	border-top: 1px solid color-mix(in srgb, var(--border) 78%, transparent);
}

.composer-action {
	font-weight: 600;
}

.composer-send {
	width: 40px;
	height: 40px;
	background: var(--accent-strong);
	box-shadow: 0 4px 14px color-mix(in srgb, var(--accent) 22%, transparent);
}

.composer-send:hover:not(:disabled) {
	background: var(--accent);
}

.composer-send--stop {
	background: var(--danger-strong);
}
</style>
