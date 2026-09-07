<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { api } from "../state/session-store.ts";
import { store } from "../state/session-store.ts";
import type { AgentView } from "../api/types.ts";

const agents = ref<AgentView[]>([]);
const error = ref<string | null>(null);
const showForm = ref(false);
const editingId = ref<string | null>(null);
const name = ref("");
const description = ref("");
const url = ref("");
const visibility = ref<"private" | "group" | "public">("private");
const groupId = ref("");

const activeGroups = computed(() => store.groups.filter((g) => g.status === "active"));

async function load(): Promise<void> {
	try {
		const list = await api.listAgents();
		// 下架(软删 status=removed)的智能体不再出现在广场
		agents.value = list.filter((agent) => agent.status !== "removed");
		error.value = null;
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e);
	}
}

function resetForm(): void {
	editingId.value = null;
	name.value = "";
	description.value = "";
	url.value = "";
	visibility.value = "private";
	groupId.value = "";
}

function openCreate(): void {
	resetForm();
	showForm.value = true;
}

function openEdit(agent: AgentView): void {
	editingId.value = agent.id;
	name.value = agent.name;
	description.value = agent.description ?? "";
	url.value = agent.url;
	visibility.value = agent.visibility;
	groupId.value = agent.groupId ?? "";
	showForm.value = true;
}

async function submit(): Promise<void> {
	try {
		const input = {
			name: name.value.trim(),
			url: url.value.trim(),
			description: description.value.trim() || undefined,
			visibility: visibility.value,
			...(visibility.value === "group" && groupId.value ? { groupId: groupId.value } : {}),
		};
		if (editingId.value) {
			await api.updateAgent(editingId.value, input);
		} else {
			await api.createAgent(input);
		}
		showForm.value = false;
		error.value = null;
		await load();
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e);
	}
}

async function removeAgent(agent: AgentView): Promise<void> {
	if (!window.confirm(`确定下架智能体「${agent.name}」？`)) return;
	try {
		await api.deleteAgent(agent.id);
		error.value = null;
		await load();
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e);
	}
}

function availabilityLabel(status: AgentView["availability"]): string {
	switch (status) {
		case "online":
			return "在线";
		case "unavailable":
			return "暂时不可用";
		default:
			return "尚未检测";
	}
}

onMounted(() => {
	void load();
});
</script>

<template>
	<div class="plaza">
		<header class="plaza-header">
			<h2 class="plaza-title">智能体广场</h2>
			<button class="btn btn-primary" type="button" @click="openCreate">发布智能体</button>
		</header>

		<p v-if="error" class="plaza-error" role="alert">{{ error }}</p>

		<div v-if="showForm" class="plaza-form">
			<div class="plaza-field">
				<label>名称</label>
				<input v-model="name" type="text" placeholder="例如 研报分析助手" />
			</div>
			<div class="plaza-field">
				<label>URL</label>
				<input v-model="url" type="text" placeholder="https://..." />
			</div>
			<div class="plaza-field">
				<label>简介</label>
				<textarea v-model="description" rows="2" placeholder="一句话描述" />
			</div>
			<div class="plaza-field">
				<label>可见性</label>
				<select v-model="visibility">
					<option value="private">私有</option>
					<option value="group">课程组</option>
					<option value="public">公开</option>
				</select>
				<select v-if="visibility === 'group'" v-model="groupId">
					<option value="">选择课程组</option>
					<option v-for="g in activeGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
				</select>
			</div>
			<div class="plaza-form-actions">
				<button class="btn btn-ghost" type="button" @click="showForm = false">取消</button>
				<button class="btn btn-primary" type="button" :disabled="!name.trim() || !url.trim()" @click="submit">
					{{ editingId ? "保存" : "发布" }}
				</button>
			</div>
		</div>

		<div v-if="agents.length === 0 && !showForm" class="plaza-empty">暂无智能体，点击右上角发布</div>

		<div class="plaza-grid">
			<a
				v-for="agent in agents"
				:key="agent.id"
				class="plaza-card"
				:href="agent.url"
				target="_blank"
				rel="noopener noreferrer"
			>
				<div class="plaza-card-head">
					<span class="plaza-card-name">{{ agent.name }}</span>
					<span class="plaza-badge" :class="`badge-${agent.availability}`">{{ availabilityLabel(agent.availability) }}</span>
				</div>
				<p class="plaza-card-desc">{{ agent.description || "无简介" }}</p>
				<div class="plaza-card-meta">
					<span class="plaza-visibility">{{ agent.visibility === "public" ? "公开" : agent.visibility === "group" ? (agent.groupName ?? "课程组") : "私有" }}</span>
					<span v-if="agent.lastCheckedAt" class="plaza-checked">检测于 {{ new Date(agent.lastCheckedAt).toLocaleString() }}</span>
				</div>
				<div class="plaza-card-actions" @click.prevent>
					<button class="btn btn-ghost" type="button" @click="openEdit(agent)">编辑</button>
					<button class="btn btn-ghost" type="button" @click="removeAgent(agent)">下架</button>
				</div>
			</a>
		</div>
	</div>
</template>

<style scoped>
.plaza {
	max-width: 1100px;
	margin: 0 auto;
	padding: 24px;
	width: 100%;
	display: flex;
	flex-direction: column;
	gap: 16px;
	overflow-y: auto;
}

.plaza-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.plaza-title {
	margin: 0;
	font-size: 18px;
	font-weight: 700;
	color: var(--text-primary);
}

.plaza-error {
	margin: 0;
	font-size: 12px;
	color: var(--danger, #e06c75);
}

.plaza-form {
	display: flex;
	flex-direction: column;
	gap: 12px;
	padding: 16px;
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	background: var(--bg-card);
}

.plaza-field {
	display: flex;
	flex-direction: column;
	gap: 6px;
}

.plaza-field label {
	font-size: 12px;
	font-weight: 500;
	color: var(--text-muted);
}

.plaza-field input,
.plaza-field textarea,
.plaza-field select {
	width: 100%;
	background: var(--bg-input);
	color: var(--text-primary);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 8px 10px;
	font-size: 13px;
	font-family: var(--font-mono);
}

.plaza-field input:focus,
.plaza-field textarea:focus,
.plaza-field select:focus {
	outline: none;
	border-color: var(--accent);
}

.plaza-form-actions {
	display: flex;
	justify-content: flex-end;
	gap: 10px;
}

.plaza-empty {
	font-size: 13px;
	color: var(--text-disabled);
	text-align: center;
	padding: 40px 0;
}

.plaza-grid {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
	gap: 14px;
}

.plaza-card {
	display: flex;
	flex-direction: column;
	gap: 8px;
	padding: 14px;
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	background: var(--bg-card);
	text-decoration: none;
	color: var(--text-primary);
	transition:
		border-color 140ms ease,
		transform 140ms ease;
}

.plaza-card:hover {
	border-color: var(--border-strong);
	transform: translateY(-1px);
}

.plaza-card-head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 8px;
}

.plaza-card-name {
	font-weight: 600;
	font-size: 14px;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.plaza-badge {
	font-size: 11px;
	padding: 2px 8px;
	border-radius: var(--radius-pill);
	flex-shrink: 0;
}

.badge-online {
	color: #2ecc71;
	background: rgba(46, 204, 113, 0.12);
}

.badge-unavailable {
	color: #e67e22;
	background: rgba(230, 126, 34, 0.12);
}

.badge-unknown {
	color: var(--text-muted);
	background: var(--bg-hover);
}

.plaza-card-desc {
	margin: 0;
	font-size: 12px;
	color: var(--text-muted);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.plaza-card-meta {
	display: flex;
	align-items: center;
	justify-content: space-between;
	font-size: 11px;
	color: var(--text-disabled);
}

.plaza-card-actions {
	display: flex;
	justify-content: flex-end;
	gap: 8px;
	border-top: 1px solid var(--border);
	padding-top: 10px;
}
</style>
