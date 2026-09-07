<script setup lang="ts">
import { onMounted, ref } from "vue";
import type { AdminSessionView } from "../api/types.ts";
import AppIcon from "./AppIcon.vue";
import { refreshAdminSessions, store, terminateAdminSession } from "../state/session-store.ts";

const refreshing = ref(false);

onMounted(() => {
	void refreshAdminSessions();
});

function statusLabel(status: string): string {
	switch (status) {
		case "created":
			return "已创建";
		case "running":
			return "运行中";
		case "ended":
			return "已结束";
		case "error":
			return "异常";
		default:
			return status;
	}
}

function statusBadgeClass(status: string): string {
	switch (status) {
		case "running":
			return "badge--blue";
		case "ended":
			return "badge--gray";
		case "error":
			return "badge--red";
		default:
			return "badge--amber";
	}
}

function containerLabel(state: string | null): string {
	if (!state) return "无容器";
	switch (state) {
		case "running":
			return "运行中";
		case "exited":
			return "已退出";
		case "missing":
			return "缺失";
		case "error":
			return "异常";
		case "expired":
			return "已过期";
		default:
			return state;
	}
}

function containerBadgeClass(state: string | null): string {
	if (!state || state === "missing") return "badge--gray";
	if (state === "running") return "badge--blue";
	return "badge--red";
}

function formatTime(ts: number | null): string {
	if (!ts) return "—";
	const date = new Date(ts);
	const sameDay = date.toDateString() === new Date().toDateString();
	if (sameDay) return date.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
	return date.toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}

async function terminate(session: AdminSessionView): Promise<void> {
	if (!window.confirm(`确定终止会话 ${session.id.slice(0, 8)}？容器将被销毁。`)) return;
	await terminateAdminSession(session.id);
}

async function doRefresh(): Promise<void> {
	refreshing.value = true;
	try {
		await refreshAdminSessions();
	} finally {
		refreshing.value = false;
	}
}
</script>

<template>
	<section class="page">
		<header class="page-header">
			<div>
				<h1 class="page-title">会话监控</h1>
				<p class="page-subtitle">全量平台会话与容器状态，可强制终止</p>
			</div>
			<div class="page-actions">
				<button class="btn btn-ghost" type="button" :disabled="refreshing" @click="doRefresh">
					<AppIcon name="refresh" :size="15" />
					<span>刷新</span>
				</button>
			</div>
		</header>

		<div class="page-body">
			<div v-if="store.adminSessions.length === 0" class="page-empty">暂无平台会话记录。</div>
			<table v-else class="data-table">
				<thead>
					<tr>
						<th>会话</th>
						<th>用户</th>
						<th>项目</th>
						<th>容器</th>
						<th>状态</th>
						<th>开始</th>
						<th>操作</th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="session in store.adminSessions" :key="session.id">
						<td class="mono-cell">{{ session.id.slice(0, 8) }}</td>
						<td>{{ session.username ?? "—" }}</td>
						<td>{{ session.projectName ?? "—" }}</td>
						<td>
							<span class="badge" :class="containerBadgeClass(session.containerState)">
								{{ containerLabel(session.containerState) }}
							</span>
							<div v-if="session.containerId" class="cell-sub mono-cell">{{ session.containerId.slice(0, 12) }}</div>
						</td>
						<td>
							<span class="badge" :class="statusBadgeClass(session.status)">{{ statusLabel(session.status) }}</span>
						</td>
						<td class="mono-cell">{{ formatTime(session.startedAt) }}</td>
						<td>
							<button
								v-if="session.status !== 'ended'"
								class="text-btn danger"
								type="button"
								@click="terminate(session)"
							>
								<AppIcon name="stop" :size="13" />
								终止
							</button>
							<span v-else class="cell-sub">—</span>
						</td>
					</tr>
				</tbody>
			</table>
		</div>
	</section>
</template>

<style scoped>
.mono-cell {
	font-family: var(--font-mono);
	font-size: 12px;
	color: var(--text-secondary);
	white-space: nowrap;
}

.cell-sub {
	font-size: 11px;
	color: var(--text-muted);
	margin-top: 2px;
}

.text-btn {
	display: inline-flex;
	align-items: center;
	gap: 4px;
}
</style>