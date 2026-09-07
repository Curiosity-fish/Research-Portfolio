<script setup lang="ts">
import { computed } from "vue";
import type { ModelMetadata, ModelRef } from "../api/types.ts";

const props = defineProps<{
	models: ModelMetadata[];
	current: ModelRef | null;
}>();

const emit = defineEmits<{
	change: [model: ModelRef];
}>();

const groups = computed(() => {
	const byProvider = new Map<string, ModelMetadata[]>();
	for (const model of props.models) {
		const list = byProvider.get(model.provider) ?? [];
		list.push(model);
		byProvider.set(model.provider, list);
	}
	return [...byProvider.entries()].map(([provider, models]) => ({ provider, models }));
});

function onSelect(event: Event): void {
	const value = (event.target as HTMLSelectElement).value;
	const [provider, id] = value.split("|", 2);
	emit("change", { provider, id });
}
</script>

<template>
	<div class="model-picker-wrap">
		<select
			class="model-picker"
			name="model"
			:value="current ? `${current.provider}|${current.id}` : ''"
			:aria-label="'选择模型'"
			@change="onSelect"
		>
			<option value="" disabled>选择模型</option>
			<template v-for="group in groups" :key="group.provider">
				<optgroup :label="group.provider">
					<option v-for="model in group.models" :key="model.id" :value="`${model.provider}|${model.id}`">
						{{ model.name }}{{ model.authenticated ? "" : " (未认证)" }}
					</option>
				</optgroup>
			</template>
		</select>
	</div>
</template>

<style scoped>
.model-picker-wrap {
	position: relative;
	display: inline-flex;
	align-items: center;
}

.model-picker {
	appearance: none;
	-webkit-appearance: none;
	background-color: var(--picker-bg);
	background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
	background-repeat: no-repeat;
	background-position: right 9px center;
	background-size: 14px;
	color: var(--picker-text);
	border: 1px solid var(--picker-border);
	border-radius: 7px;
	height: 36px;
	padding: 0 28px 0 10px;
	font-size: 13px;
	font-weight: 500;
	max-width: 240px;
	cursor: pointer;
	transition:
		background 150ms ease,
		border-color 150ms ease,
		color 150ms ease;
}

.model-picker:hover {
	background: var(--picker-hover-bg);
	border-color: var(--picker-hover-border);
	color: var(--picker-text);
}

.model-picker:focus {
	outline: none;
	border-color: var(--accent);
}

.model-picker::after {
	content: "";
}
</style>
