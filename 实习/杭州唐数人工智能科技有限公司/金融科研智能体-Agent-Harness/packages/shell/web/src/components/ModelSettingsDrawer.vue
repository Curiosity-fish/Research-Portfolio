<script setup lang="ts">
import { computed, ref } from "vue";
import { api, refreshConfig, refreshModels, removeApiKey, setThinking, store } from "../state/session-store.ts";
import type { ThinkingLevel } from "../api/types.ts";

defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: [] }>();

const KNOWN_PROVIDERS = [
	"openai",
	"anthropic",
	"deepseek",
	"google",
	"groq",
	"openrouter",
	"xai",
	"mistral",
	"zai-coding-cn",
	"qwen-token-plan-cn",
	"minimax-cn",
	"xiaomi-token-plan-cn",
	"kimi-coding",
];

const provider = ref("deepseek");
const apiKey = ref("");
const saving = ref(false);
const message = ref("");

const providerOptions = computed(() => {
	const known = new Set(KNOWN_PROVIDERS);
	for (const entry of store.config?.providers ?? []) known.add(entry.provider);
	return [...known];
});

const configuredProviders = computed(() => {
	const map = new Map<string, boolean>();
	for (const entry of store.config?.providers ?? []) map.set(entry.provider, entry.configured);
	return [...map.entries()].map(([provider, configured]) => ({ provider, configured }));
});

async function save(): Promise<void> {
	const cleanProvider = provider.value.trim();
	const cleanKey = apiKey.value.trim();
	if (!cleanProvider || !cleanKey) {
		message.value = "请填写供应商和 API Key";
		return;
	}
	saving.value = true;
	message.value = "";
	try {
		await api.configureApiKey(cleanProvider, cleanKey);
		await refreshModels();
		await refreshConfig();
		apiKey.value = "";
		message.value = "已保存并刷新模型列表";
	} catch (error) {
		message.value = error instanceof Error ? error.message : String(error);
	} finally {
		saving.value = false;
	}
}

async function remove(providerId: string): Promise<void> {
	if (!window.confirm(`移除 ${providerId} 的 API Key？`)) return;
	message.value = "";
	try {
		await removeApiKey(providerId);
		message.value = "已移除";
	} catch (error) {
		message.value = error instanceof Error ? error.message : String(error);
	}
}

const ALL_THINKING_LEVELS: ThinkingLevel[] = ["off", "minimal", "low", "medium", "high", "xhigh", "max"];

const THINKING_LABELS: Record<ThinkingLevel, string> = {
	off: "关闭",
	minimal: "极简",
	low: "低",
	medium: "中",
	high: "高",
	xhigh: "极高",
	max: "最高",
};

const currentModelMeta = computed(() => {
	const model = store.snapshot?.model;
	if (!model) return undefined;
	return store.models.find((m) => m.provider === model.provider && m.id === model.id);
});

const thinkingOptions = computed<ThinkingLevel[]>(() => {
	const supported = currentModelMeta.value?.supportedThinkingLevels;
	return supported && supported.length > 0 ? supported : ALL_THINKING_LEVELS;
});

async function onThinkingChange(event: Event): Promise<void> {
	const level = (event.target as HTMLSelectElement).value as ThinkingLevel;
	message.value = "";
	try {
		await setThinking(level);
		message.value = "已更新思考级别";
	} catch (error) {
		message.value = error instanceof Error ? error.message : String(error);
	}
}
</script>

<template>
	<div v-if="open" class="drawer-layer" @click.self="emit('close')">
		<aside class="drawer" role="dialog" aria-label="模型设置">
			<header class="drawer-header">
				<h2>模型设置</h2>
				<button class="btn btn-ghost btn-sm" type="button" @click="emit('close')">关闭</button>
			</header>
			<div class="drawer-body">
				<div class="field">
					<label for="provider-select">供应商</label>
					<select id="provider-select" name="provider" v-model="provider">
						<option v-for="item in providerOptions" :key="item" :value="item">{{ item }}</option>
					</select>
				</div>
				<div class="field">
					<label for="api-key-input">API Key</label>
					<input id="api-key-input" name="apiKey" v-model="apiKey" type="password" placeholder="sk-..." autocomplete="off" />
				</div>
				<button class="btn btn-primary" type="button" :disabled="saving" @click="save">
					{{ saving ? "保存中…" : "保存并刷新" }}
				</button>
				<p v-if="message" class="drawer-message">{{ message }}</p>

				<section class="drawer-section">
					<h3>当前会话</h3>
					<div class="field">
						<label for="thinking-select">思考级别</label>
						<select
							id="thinking-select"
							name="thinkingLevel"
							:value="store.snapshot?.thinkingLevel ?? ''"
							:disabled="!store.currentId"
							@change="onThinkingChange"
						>
							<option value="" disabled>选择思考级别</option>
							<option v-for="level in thinkingOptions" :key="level" :value="level">
								{{ THINKING_LABELS[level] }}
							</option>
						</select>
						<p v-if="!store.currentId" class="drawer-message">打开一个会话后可设置思考级别</p>
					</div>
				</section>

				<section class="drawer-section">
					<h3>已配置供应商</h3>
					<p v-if="configuredProviders.length === 0" class="empty-state">还没有配置 API Key</p>
					<ul v-else class="provider-list">
						<li v-for="entry in configuredProviders" :key="entry.provider">
							<span class="provider-id">{{ entry.provider }}</span>
							<span class="status" :class="{ ok: entry.configured }">{{ entry.configured ? "已认证" : "未认证" }}</span>
							<button v-if="entry.configured" class="text-button danger" type="button" @click="remove(entry.provider)">移除</button>
						</li>
					</ul>
				</section>

				<section class="drawer-section">
					<h3>配置路径</h3>
					<dl class="config-list">
						<dt>配置目录</dt>
						<dd>{{ store.config?.agentDir ?? "加载中…" }}</dd>
						<dt>认证文件</dt>
						<dd>{{ store.config?.authPath ?? "加载中…" }}</dd>
						<dt>模型文件</dt>
						<dd>{{ store.config?.modelsPath ?? "加载中…" }}</dd>
					</dl>
				</section>
			</div>
		</aside>
	</div>
</template>
