<script setup lang="ts">
import { computed } from "vue";
import { decideApproval, store } from "../state/session-store.ts";

const approvals = computed(() => store.pendingApprovals);
</script>

<template>
	<div v-if="approvals.length > 0" class="approval-layer">
		<div class="approval-card" role="dialog" aria-modal="true" aria-label="操作审批">
			<header class="approval-header">
				<span class="approval-title">需要审批的操作</span>
				<span class="approval-count">{{ approvals.length }}</span>
			</header>
			<div class="approval-list">
				<div v-for="request in approvals" :key="request.id" class="approval-item">
					<div class="approval-meta">
						<span class="approval-tool">{{ request.tool }}</span>
						<span class="approval-target" :title="request.target">{{ request.target }}</span>
						<span class="approval-risk">{{ request.risk }}</span>
					</div>
					<div class="approval-actions">
						<button class="btn btn-ghost" type="button" @click="decideApproval(request.id, false)">拒绝</button>
						<button class="btn btn-primary" type="button" @click="decideApproval(request.id, true)">批准</button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.approval-layer {
	position: fixed;
	inset: 0;
	z-index: 60;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 24px;
	background: rgba(15, 23, 42, 0.5);
}

.approval-card {
	width: 100%;
	max-width: 520px;
	background: var(--bg-card);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-float);
	overflow: hidden;
}

.approval-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 14px 18px;
	border-bottom: 1px solid var(--border);
}

.approval-title {
	font-weight: 700;
	font-size: 14px;
	color: var(--text-primary);
}

.approval-count {
	font-size: 12px;
	color: var(--accent-strong);
	background: var(--accent-soft);
	padding: 2px 8px;
	border-radius: var(--radius-pill);
}

.approval-list {
	padding: 12px 18px 18px;
	display: flex;
	flex-direction: column;
	gap: 12px;
	max-height: 50vh;
	overflow-y: auto;
}

.approval-item {
	display: flex;
	align-items: center;
	gap: 12px;
	border: 1px solid var(--border);
	border-radius: var(--radius-md);
	padding: 10px 12px;
	background: var(--bg-app);
}

.approval-meta {
	flex: 1;
	min-width: 0;
	display: flex;
	flex-direction: column;
	gap: 4px;
}

.approval-tool {
	font-family: var(--font-mono);
	font-size: 12px;
	font-weight: 600;
	color: var(--accent-strong);
}

.approval-target {
	font-family: var(--font-mono);
	font-size: 12px;
	color: var(--text-primary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.approval-risk {
	font-size: 11px;
	color: var(--text-muted);
}

.approval-actions {
	display: flex;
	gap: 8px;
	flex-shrink: 0;
}
</style>
