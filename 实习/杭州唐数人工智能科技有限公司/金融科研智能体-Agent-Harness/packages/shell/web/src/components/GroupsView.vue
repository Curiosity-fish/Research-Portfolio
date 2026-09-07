<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import type { GroupDetail } from "../api/types.ts";
import AppIcon from "./AppIcon.vue";
import { api, addGroupMember, archiveGroup, createGroup, removeGroupMember, store, updateGroup, updateGroupMemberRole } from "../state/session-store.ts";

const selectedId = ref<string | null>(null);
const detail = ref<GroupDetail | null>(null);

const showCreate = ref(false);
const createName = ref("");
const createType = ref<"course" | "research">("course");
const createBusy = ref(false);

const showRename = ref(false);
const renameName = ref("");

const showAdd = ref(false);
const addUserId = ref("");
const addBusy = ref(false);

const isAdmin = computed(() => store.user?.role === "admin");

function canManage(group: { myRole: "owner" | "member" | null }): boolean {
	return group.myRole === "owner" || isAdmin.value;
}

function typeLabel(type: string): string {
	return type === "course" ? "课程组" : "课题组";
}

function statusBadge(status: string): string {
	return status === "archived" ? "已归档" : "进行中";
}

function openDetail(groupId: string): void {
	selectedId.value = groupId;
	detail.value = null;
	void loadDetail(groupId);
}

async function loadDetail(groupId: string): Promise<void> {
	try {
		detail.value = await api.getGroup(groupId);
	} catch {
		// 错误已写入 store
	}
}

function backToList(): void {
	selectedId.value = null;
	detail.value = null;
}

async function submitCreate(): Promise<void> {
	if (!createName.value.trim()) return;
	createBusy.value = true;
	try {
		await createGroup({ name: createName.value.trim(), type: createType.value });
		showCreate.value = false;
		createName.value = "";
	} finally {
		createBusy.value = false;
	}
}

function openRename(groupId: string): void {
	const group = store.groups.find((item) => item.id === groupId);
	renameName.value = group?.name ?? "";
	showRename.value = true;
	selectedId.value = groupId;
}

async function submitRename(): Promise<void> {
	if (!renameName.value.trim() || !selectedId.value) return;
	await updateGroup(selectedId.value, { name: renameName.value.trim() });
	showRename.value = false;
	if (detail.value) await loadDetail(selectedId.value);
}

async function submitArchive(groupId: string): Promise<void> {
	if (!window.confirm("确定归档该组？归档后项目与库保留，但不可再添加新成员。")) return;
	await archiveGroup(groupId);
	if (selectedId.value === groupId) await loadDetail(groupId);
}

function openAdd(): void {
	addUserId.value = "";
	showAdd.value = true;
}

async function submitAdd(): Promise<void> {
	if (!addUserId.value.trim() || !selectedId.value) return;
	addBusy.value = true;
	try {
		await addGroupMember(selectedId.value, addUserId.value.trim());
		showAdd.value = false;
		await loadDetail(selectedId.value);
	} finally {
		addBusy.value = false;
	}
}

async function setLeader(userId: string): Promise<void> {
	if (!selectedId.value) return;
	if (!window.confirm("确定将该成员设为组长？")) return;
	await updateGroupMemberRole(selectedId.value, userId, "owner");
	await loadDetail(selectedId.value);
}

async function demoteLeader(userId: string): Promise<void> {
	if (!selectedId.value) return;
	await updateGroupMemberRole(selectedId.value, userId, "member");
	await loadDetail(selectedId.value);
}

async function kickMember(userId: string): Promise<void> {
	if (!selectedId.value) return;
	if (!window.confirm("确定将该成员移出组？")) return;
	await removeGroupMember(selectedId.value, userId);
	await loadDetail(selectedId.value);
}

</script>

<template>
	<section class="page">
		<template v-if="!selectedId">
			<header class="page-header">
				<div>
					<h1 class="page-title">课程组</h1>
					<p class="page-subtitle">按课程/课题分组管理成员，组内可共享项目、技能与数据集</p>
				</div>
				<div class="page-actions">
					<button
						v-if="isAdmin || store.user?.role === 'teacher'"
						class="btn btn-primary"
						type="button"
						@click="showCreate = true"
					>
						<AppIcon name="plus" :size="15" />
						<span>新建组</span>
					</button>
				</div>
			</header>

			<div class="page-body">
				<div v-if="store.groups.length === 0" class="page-empty">
					还没有课程组。点击「新建组」创建第一个分组。
				</div>
				<div v-else class="card-grid">
					<div
						v-for="group in store.groups"
						:key="group.id"
						class="platform-card"
						@click="openDetail(group.id)"
					>
						<div class="card-title-row">
							<span class="card-icon"><AppIcon name="users" :size="16" /></span>
							<span class="card-title">{{ group.name }}</span>
						</div>
						<div class="card-desc">
							{{ typeLabel(group.type) }}
						</div>
						<div class="card-meta">
							<span class="badge" :class="group.status === 'archived' ? 'badge--gray' : 'badge--blue'">
								{{ statusBadge(group.status) }}
							</span>
							<span>{{ group.memberCount }} 名成员</span>
							<span v-if="group.myRole === 'owner'" class="badge badge--amber">组长</span>
							<span v-else-if="group.myRole === 'member'" class="badge badge--gray">成员</span>
						</div>
					</div>
				</div>
			</div>
		</template>

		<template v-else-if="detail">
			<header class="page-header">
				<div>
					<button class="text-btn back-btn" type="button" @click="backToList">
						<AppIcon name="arrow-right" :size="14" />
						<span>返回列表</span>
					</button>
					<h1 class="page-title">{{ detail.name }}</h1>
					<p class="page-subtitle">
						{{ typeLabel(detail.type) }} · {{ statusBadge(detail.status) }} · {{ detail.memberCount }} 名成员
					</p>
				</div>
				<div class="page-actions">
					<button v-if="canManage(detail)" class="btn btn-ghost" type="button" @click="openRename(detail.id)">
						<AppIcon name="pen" :size="14" />
						<span>改名</span>
					</button>
					<button
						v-if="canManage(detail) && detail.status === 'active'"
						class="btn btn-ghost"
						type="button"
						@click="submitArchive(detail.id)"
					>
						<AppIcon name="archive" :size="14" />
						<span>归档</span>
					</button>
					<button v-if="canManage(detail) && detail.status === 'active'" class="btn btn-primary" type="button" @click="openAdd">
						<AppIcon name="plus" :size="15" />
						<span>添加成员</span>
					</button>
				</div>
			</header>

			<div class="page-body">
				<table class="data-table">
					<thead>
						<tr>
							<th>用户名</th>
							<th>角色</th>
							<th v-if="canManage(detail)">操作</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="member in detail.members" :key="member.userId">
							<td>
								<span class="member-name">{{ member.username }}</span>
								<span v-if="member.userId === detail.ownerId" class="badge badge--amber">组长</span>
							</td>
							<td>
								<span class="badge" :class="member.role === 'owner' ? 'badge--blue' : 'badge--gray'">
									{{ member.role === "owner" ? "组长" : "成员" }}
								</span>
							</td>
							<td v-if="canManage(detail)">
								<div class="row-actions">
									<button
										v-if="member.role !== 'owner'"
										class="text-btn"
										type="button"
										@click="setLeader(member.userId)"
									>
										设为组长
									</button>
									<button
										v-else-if="member.userId !== detail.ownerId"
										class="text-btn"
										type="button"
										@click="demoteLeader(member.userId)"
									>
										降为成员
									</button>
									<button
										v-if="member.userId !== store.user?.id"
										class="text-btn danger"
										type="button"
										@click="kickMember(member.userId)"
									>
										移除
									</button>
								</div>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</template>
	</section>

	<!-- 新建组 -->
	<div v-if="showCreate" class="picker-layer" @click.self="showCreate = false">
		<div class="picker" role="dialog" aria-modal="true" aria-label="新建课程组">
			<header class="picker-header">
				<div class="picker-heading">
					<h2 class="picker-title">新建组</h2>
					<p class="picker-desc">创建课程组或课题组</p>
				</div>
				<button class="picker-close" type="button" aria-label="关闭" @click="showCreate = false">
					<AppIcon name="x" :size="16" />
				</button>
			</header>
			<div class="picker-body">
				<div class="modal-field">
					<label for="group-name">组名称</label>
					<input id="group-name" v-model="createName" type="text" autocomplete="off" spellcheck="false" placeholder="例如 机器学习课程" />
				</div>
				<div class="modal-field">
					<label for="group-type">类型</label>
					<select id="group-type" v-model="createType">
						<option value="course">课程组</option>
						<option value="research">课题组</option>
					</select>
				</div>
			</div>
			<footer class="picker-footer">
				<button class="btn btn-ghost" type="button" @click="showCreate = false">取消</button>
				<button class="btn btn-primary" type="button" :disabled="!createName.trim() || createBusy" @click="submitCreate">
					{{ createBusy ? "创建中…" : "创建" }}
				</button>
			</footer>
		</div>
	</div>

	<!-- 改名 -->
	<div v-if="showRename" class="picker-layer" @click.self="showRename = false">
		<div class="picker" role="dialog" aria-modal="true" aria-label="重命名组">
			<header class="picker-header">
				<div class="picker-heading">
					<h2 class="picker-title">重命名组</h2>
				</div>
				<button class="picker-close" type="button" aria-label="关闭" @click="showRename = false">
					<AppIcon name="x" :size="16" />
				</button>
			</header>
			<div class="picker-body">
				<div class="modal-field">
					<label for="group-rename">组名称</label>
					<input id="group-rename" v-model="renameName" type="text" autocomplete="off" spellcheck="false" />
				</div>
			</div>
			<footer class="picker-footer">
				<button class="btn btn-ghost" type="button" @click="showRename = false">取消</button>
				<button class="btn btn-primary" type="button" :disabled="!renameName.trim()" @click="submitRename">保存</button>
			</footer>
		</div>
	</div>

	<!-- 添加成员 -->
	<div v-if="showAdd" class="picker-layer" @click.self="showAdd = false">
		<div class="picker" role="dialog" aria-modal="true" aria-label="添加成员">
			<header class="picker-header">
				<div class="picker-heading">
					<h2 class="picker-title">添加成员</h2>
					<p class="picker-desc">输入要加入的用户 ID（管理端可查看用户列表）</p>
				</div>
				<button class="picker-close" type="button" aria-label="关闭" @click="showAdd = false">
					<AppIcon name="x" :size="16" />
				</button>
			</header>
			<div class="picker-body">
				<div class="modal-field">
					<label for="member-userid">用户 ID</label>
					<input id="member-userid" v-model="addUserId" type="text" autocomplete="off" spellcheck="false" placeholder="粘贴用户 ID" />
				</div>
			</div>
			<footer class="picker-footer">
				<button class="btn btn-ghost" type="button" @click="showAdd = false">取消</button>
				<button class="btn btn-primary" type="button" :disabled="!addUserId.trim() || addBusy" @click="submitAdd">
					{{ addBusy ? "添加中…" : "添加" }}
				</button>
			</footer>
		</div>
	</div>
</template>

<style scoped>
.back-btn {
	display: inline-flex;
	align-items: center;
	gap: 4px;
	margin-bottom: 6px;
	padding: 0;
}

.back-btn svg {
	transform: rotate(180deg);
	color: var(--text-muted);
}

.member-name {
	font-weight: 500;
	color: var(--text-primary);
	margin-right: 8px;
}

.row-actions {
	display: flex;
	align-items: center;
	gap: 4px;
	flex-wrap: wrap;
}
</style>