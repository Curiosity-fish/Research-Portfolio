<script setup lang="ts">
import { store } from "../state/session-store.ts";

defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: [] }>();
</script>

<template>
	<div v-if="open" class="drawer-layer" @click.self="emit('close')">
		<aside class="drawer" role="dialog" aria-label="金融工具">
			<header class="drawer-header">
				<h2>金融工具</h2>
				<button class="btn btn-ghost btn-sm" type="button" @click="emit('close')">关闭</button>
			</header>
			<div class="drawer-body">
				<p v-if="store.tools.length === 0" class="empty-state">
					当前服务没有注册金融工具。可在 packages/shell/src/fintech 注册后重启服务。
				</p>
				<ul v-else class="tool-list">
					<li v-for="tool in store.tools" :key="tool.name" class="tool-entry">
						<div class="tool-title">{{ tool.label }}</div>
						<div class="tool-name">{{ tool.name }}</div>
						<p class="tool-description">{{ tool.description }}</p>
					</li>
				</ul>
			</div>
		</aside>
	</div>
</template>
