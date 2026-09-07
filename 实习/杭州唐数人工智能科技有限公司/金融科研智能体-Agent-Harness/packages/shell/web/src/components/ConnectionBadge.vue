<script setup lang="ts">
withDefaults(
	defineProps<{
		connected: boolean;
		hasSession?: boolean;
	}>(),
	{ hasSession: true },
);
</script>

<template>
	<span v-if="hasSession" class="connection" :class="{ disconnected: !connected }">
		<span class="dot" />
		<span>{{ connected ? "Agent 服务已连接" : "Agent 服务连接失败" }}</span>
	</span>
	<span v-else class="connection idle">
		<span class="dot" />
		<span>服务就绪</span>
	</span>
</template>

<style scoped>
.connection {
	display: inline-flex;
	align-items: center;
	gap: 8px;
	font-size: 11px;
	font-weight: 400;
	color: #8399ab;
}

.dot {
	width: 7px;
	height: 7px;
	border-radius: 50%;
	background: var(--accent);
	flex-shrink: 0;
}

.connection.disconnected .dot {
	background: var(--danger);
}

.connection.disconnected {
	color: var(--danger);
}

.connection.idle {
	color: #8399ab;
}

.connection.idle .dot {
	background: var(--text-muted);
}
</style>
