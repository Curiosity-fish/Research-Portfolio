<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import type { UserInfo } from "../api/types.ts";
import AppIcon from "./AppIcon.vue";
import {
	createUser,
	refreshDatasets,
	refreshSkills,
	refreshUsers,
	reviewDataset,
	reviewSkill,
	store,
	updateUser,
} from "../state/session-store.ts";

type Tab = "users" | "skills" | "datasets";
const tab = ref<Tab>("users");
const refreshing = ref(false);

const showCreate = ref(false);
const createBusy = ref(false);
const createUsername = ref("");
const createPassword = ref("");
const createRole = ref<"admin" | "teacher" | "student">("student");

const pendingSkills = computed(() => store.skills.filter((s) => s.reviewState === "pending"));

onMounted(() => {
	void refreshUsers();
});
const pendingDatasets = computed(() => store.datasets.filter((d) => d.reviewState === "pending"));

function roleLabel(role: string): string {
	return role === "admin" ? "管理员" : role === "teacher" ? "教师" : "学生";
}

function resetCreate(): void {
	createUsername.value = "";
	createPassword.value = "";
	createRole.value = "student";
}

async function submitCreate(): Promise<void> {
	if (!createUsername.value.trim() || !createPassword.value) return;
	createBusy.value = true;
	try {
		await createUser({
			username: createUsername.value.trim(),
			password: createPassword.value,
			role: createRole.value,
		});
		showCreate.value = false;
		resetCreate();
	} finally {
		createBusy.value = false;
	}
}

async function toggleStatus(user: UserInfo): Promise<void> {
	await updateUser(user.id, { status: user.status === "active" ? "disabled" : "active" });
}

async function changeRole(user: UserInfo, role: "admin" | "teacher" | "student"): Promise<void> {
	if (role === user.role) return;
	await updateUser(user.id, { role });
}

async function resetPassword(user: UserInfo): Promise<void> {
	const password = window.prompt(`为「${user.username}」设置新密码`);
	if (!password) return;
	await updateUser(user.id, { password });
}

async function handleReviewSkill(id: string, approved: boolean): Promise<void> {
	await reviewSkill(id, approved ? "approved" : "rejected");
}

async function handleReviewDataset(id: string, approved: boolean): Promise<void> {
	await reviewDataset(id, approved ? "approved" : "rejected");
}

async function doRefresh(): Promise<void> {
	refreshing.value = true;
	try {
		await Promise.all([refreshUsers(), refreshSkills(), refreshDatasets()]);
	} finally {
		refreshing.value = false;
	}
}
</script>

<template>
	<section class="page">
		<header class="page-header">
			<div>
				<h1 class="page-title">管理端</h1>
				<p class="page-subtitle">用户管理、资源审核与平台治理</p>
			</div>
			<div class="page-actions">
				<button class="btn btn-ghost" type="button" :disabled="refreshing" @click="doRefresh">
					<AppIcon name="refresh" :size="15" />
					<span>刷新</span>
				</button>
			</div>
		</header>

		<div class="tabs" role="tablist" aria-label="管理端分区">
			<button class="tab-btn" :class="{ active: tab === 'users' }" type="button" role="tab" @click="tab = 'users'">
				用户管理
			</button>
			<button class="tab-btn" :class="{ active: tab === 'skills' }" type="button" role="tab" @click="tab = 'skills'">
				技能审核
				<span v-if="pendingSkills.length" class="tab-count">{{ pendingSkills.length }}</span>
			</button>
			<button class="tab-btn" :class="{ active: tab === 'datasets' }" type="button" role="tab" @click="tab = 'datasets'">
				数据集审核
				<span v-if="pendingDatasets.length" class="tab-count">{{ pendingDatasets.length }}</span>
			</button>
		</div>

		<div class="page-body">
			<template v-if="tab === 'users'">
				<div class="page-actions">
					<button class="btn btn-primary" type="button" @click="showCreate = true">
						<AppIcon name="plus" :size="15" />
						<span>创建用户</span>
					</button>
				</div>
				<table class="data-table">
					<thead>
						<tr>
							<th>用户名</th>
							<th>角色</th>
							<th>状态</th>
							<th>操作</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="user in store.users" :key="user.id">
							<td class="user-name">{{ user.username }}</td>
							<td>
								<select class="vis-select" :value="user.role" @change="changeRole(user, ($event.target as HTMLSelectElement).value as 'admin' | 'teacher' | 'student')">
									<option value="admin">管理员</option>
									<option value="teacher">教师</option>
									<option value="student">学生</option>
								</select>
							</td>
							<td>
								<span class="badge" :class="user.status === 'active' ? 'badge--blue' : 'badge--red'">
									{{ user.status === "active" ? "启用" : "禁用" }}
								</span>
							</td>
							<td>
								<div class="row-actions">
									<button class="text-btn" type="button" @click="toggleStatus(user)">
										{{ user.status === "active" ? "禁用" : "启用" }}
									</button>
									<button class="text-btn" type="button" @click="resetPassword(user)">重置密码</button>
								</div>
							</td>
						</tr>
					</tbody>
				</table>
			</template>

			<template v-else-if="tab === 'skills'">
				<div v-if="pendingSkills.length === 0" class="page-empty">暂无待审核的技能。</div>
				<table v-else class="data-table">
					<thead>
						<tr>
							<th>技能</th>
							<th>来源</th>
							<th>可见性</th>
							<th>操作</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="skill in pendingSkills" :key="skill.id">
							<td class="user-name">{{ skill.name }}</td>
							<td class="mono-cell">{{ skill.repoUrl || "自研上传" }}</td>
							<td>
								<span class="badge badge--gray">{{ skill.visibility }}</span>
							</td>
							<td>
								<div class="row-actions">
									<button class="text-btn" type="button" @click="handleReviewSkill(skill.id, true)">通过</button>
									<button class="text-btn danger" type="button" @click="handleReviewSkill(skill.id, false)">驳回</button>
								</div>
							</td>
						</tr>
					</tbody>
				</table>
			</template>

			<template v-else>
				<div v-if="pendingDatasets.length === 0" class="page-empty">暂无待审核的数据集。</div>
				<table v-else class="data-table">
					<thead>
						<tr>
							<th>数据集</th>
							<th>大小</th>
							<th>可见性</th>
							<th>操作</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="dataset in pendingDatasets" :key="dataset.id">
							<td class="user-name">{{ dataset.name }}</td>
							<td class="mono-cell">{{ (dataset.sizeBytes / 1024).toFixed(1) }} KB</td>
							<td>
								<span class="badge badge--gray">{{ dataset.visibility }}</span>
							</td>
							<td>
								<div class="row-actions">
									<button class="text-btn" type="button" @click="handleReviewDataset(dataset.id, true)">通过</button>
									<button class="text-btn danger" type="button" @click="handleReviewDataset(dataset.id, false)">驳回</button>
								</div>
							</td>
						</tr>
					</tbody>
				</table>
			</template>
		</div>
	</section>

	<!-- 创建用户 -->
	<div v-if="showCreate" class="picker-layer" @click.self="showCreate = false">
		<div class="picker" role="dialog" aria-modal="true" aria-label="创建用户">
			<header class="picker-header">
				<div class="picker-heading">
					<h2 class="picker-title">创建用户</h2>
				</div>
				<button class="picker-close" type="button" aria-label="关闭" @click="showCreate = false">
					<AppIcon name="x" :size="16" />
				</button>
			</header>
			<div class="picker-body">
				<div class="modal-field">
					<label for="new-username">用户名</label>
					<input id="new-username" v-model="createUsername" type="text" autocomplete="off" spellcheck="false" />
				</div>
				<div class="modal-field">
					<label for="new-password">密码</label>
					<input id="new-password" v-model="createPassword" type="password" autocomplete="new-password" />
				</div>
				<div class="modal-field">
					<label for="new-role">角色</label>
					<select id="new-role" v-model="createRole">
						<option value="student">学生</option>
						<option value="teacher">教师</option>
						<option value="admin">管理员</option>
					</select>
				</div>
			</div>
			<footer class="picker-footer">
				<button class="btn btn-ghost" type="button" @click="showCreate = false">取消</button>
				<button class="btn btn-primary" type="button" :disabled="!createUsername.trim() || !createPassword || createBusy" @click="submitCreate">
					{{ createBusy ? "创建中…" : "创建" }}
				</button>
			</footer>
		</div>
	</div>
</template>

<style scoped>
.user-name {
	font-weight: 500;
	color: var(--text-primary);
}

.mono-cell {
	font-family: var(--font-mono);
	font-size: 12px;
	color: var(--text-secondary);
}

.vis-select {
	background: var(--bg-input);
	color: var(--text-primary);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 4px 8px;
	font-size: 12px;
}

.row-actions {
	display: flex;
	align-items: center;
	gap: 4px;
	flex-wrap: wrap;
}

.tab-count {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	min-width: 16px;
	height: 16px;
	padding: 0 4px;
	margin-left: 4px;
	border-radius: 8px;
	background: var(--danger);
	color: #fff;
	font-size: 10px;
	font-weight: 600;
}
</style>