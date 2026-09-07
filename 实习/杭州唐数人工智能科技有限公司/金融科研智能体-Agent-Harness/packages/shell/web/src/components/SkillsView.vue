<script setup lang="ts">
import { computed, ref } from "vue";
import type { ResourceVisibility, SkillView } from "../api/types.ts";
import AppIcon from "./AppIcon.vue";
import {
	createSkill,
	deleteSkill,
	refreshSkills,
	store,
	updateSkill,
	updateSkillFromRepo,
} from "../state/session-store.ts";

const showInstall = ref(false);
const installBusy = ref(false);
const installRepoUrl = ref("");
const installName = ref("");
const installVisibility = ref<ResourceVisibility>("private");
const installGroupId = ref<string>("");

const showUpload = ref(false);
const uploadBusy = ref(false);
const uploadName = ref("");
const uploadVisibility = ref<ResourceVisibility>("private");
const uploadGroupId = ref<string>("");

const refreshing = ref(false);

const myGroups = computed(() => store.groups.filter((g) => g.myRole === "owner" || g.myRole === "member"));

function visLabel(v: ResourceVisibility): string {
	return v === "private" ? "私有" : v === "group" ? "组共享" : "公开";
}

function visBadgeClass(v: ResourceVisibility): string {
	return v === "public" ? "badge--blue" : v === "group" ? "badge--amber" : "badge--gray";
}

function reviewBadge(s: string): { text: string; cls: string } {
	switch (s) {
		case "pending":
			return { text: "待审核", cls: "badge--amber" };
		case "approved":
			return { text: "已通过", cls: "badge--blue" };
		case "rejected":
			return { text: "已驳回", cls: "badge--red" };
		default:
			return { text: "未审核", cls: "badge--gray" };
	}
}

function canEdit(skill: SkillView): boolean {
	return store.user?.role === "admin" || skill.ownerId === store.user?.id;
}

function resetInstall(): void {
	installRepoUrl.value = "";
	installName.value = "";
	installVisibility.value = "private";
	installGroupId.value = "";
}

function resetUpload(): void {
	uploadName.value = "";
	uploadVisibility.value = "private";
	uploadGroupId.value = "";
}

async function submitInstall(): Promise<void> {
	if (!installRepoUrl.value.trim()) return;
	installBusy.value = true;
	try {
		await createSkill({
			repoUrl: installRepoUrl.value.trim(),
			name: installName.value.trim() || undefined,
			visibility: installVisibility.value,
			groupId: installVisibility.value === "group" && installGroupId.value ? installGroupId.value : undefined,
		});
		showInstall.value = false;
		resetInstall();
	} finally {
		installBusy.value = false;
	}
}

async function submitUpload(): Promise<void> {
	if (!uploadName.value.trim()) return;
	uploadBusy.value = true;
	try {
		await createSkill({
			name: uploadName.value.trim(),
			upload: true,
			visibility: uploadVisibility.value,
			groupId: uploadVisibility.value === "group" && uploadGroupId.value ? uploadGroupId.value : undefined,
		});
		showUpload.value = false;
		resetUpload();
	} finally {
		uploadBusy.value = false;
	}
}

async function handleUpdate(skill: SkillView): Promise<void> {
	if (!window.confirm(`确定从仓库更新「${skill.name}」？`)) return;
	await updateSkillFromRepo(skill.id);
}

async function handleVisibilityChange(skill: SkillView, visibility: ResourceVisibility): Promise<void> {
	await updateSkill(skill.id, { visibility });
}

async function handleDelete(skill: SkillView): Promise<void> {
	if (!window.confirm(`确定删除技能「${skill.name}」？`)) return;
	await deleteSkill(skill.id);
}

async function doRefresh(): Promise<void> {
	refreshing.value = true;
	try {
		await refreshSkills();
	} finally {
		refreshing.value = false;
	}
}
</script>

<template>
	<section class="page">
		<header class="page-header">
			<div>
				<h1 class="page-title">技能库</h1>
				<p class="page-subtitle">从 git 仓库安装或自研上传技能，按可见性共享给组或公开</p>
			</div>
			<div class="page-actions">
				<button class="btn btn-ghost" type="button" :disabled="refreshing" @click="doRefresh">
					<AppIcon name="refresh" :size="15" />
					<span>刷新</span>
				</button>
				<button class="btn btn-primary" type="button" @click="showInstall = true">
					<AppIcon name="download" :size="15" />
					<span>安装技能</span>
				</button>
				<button class="btn btn-ghost" type="button" @click="showUpload = true">
					<AppIcon name="upload" :size="15" />
					<span>自研上传</span>
				</button>
			</div>
		</header>

		<div class="page-body">
			<div v-if="store.skills.length === 0" class="page-empty">
				技能库为空。点击「安装技能」从 git 仓库安装，或「自研上传」登记自研技能。
			</div>
			<div v-else class="card-grid">
				<div v-for="skill in store.skills" :key="skill.id" class="platform-card">
					<div class="card-title-row">
						<span class="card-icon"><AppIcon name="layers" :size="16" /></span>
						<span class="card-title">{{ skill.name }}</span>
					</div>
					<div class="card-desc mono" :title="skill.repoUrl ?? ''">
						{{ skill.repoUrl || "自研上传" }}
					</div>
					<div class="card-badges">
						<span class="badge" :class="visBadgeClass(skill.visibility)">{{ visLabel(skill.visibility) }}</span>
						<span class="badge" :class="reviewBadge(skill.reviewState).cls">{{ reviewBadge(skill.reviewState).text }}</span>
						<span v-if="skill.version" class="badge badge--gray mono">v{{ skill.version }}</span>
						<span v-if="skill.groupName" class="badge badge--gray">{{ skill.groupName }}</span>
					</div>
					<div class="card-actions">
						<select
							v-if="canEdit(skill)"
							class="vis-select"
							:value="skill.visibility"
							:aria-label="`${skill.name} 可见性`"
							@change="handleVisibilityChange(skill, ($event.target as HTMLSelectElement).value as ResourceVisibility)"
						>
							<option value="private">私有</option>
							<option value="group">组共享</option>
							<option value="public">公开</option>
						</select>
						<div class="spacer" />
						<button v-if="skill.repoUrl && canEdit(skill)" class="text-btn" type="button" @click="handleUpdate(skill)">
							<AppIcon name="refresh" :size="13" />
							更新
						</button>
						<button v-if="canEdit(skill)" class="text-btn danger" type="button" @click="handleDelete(skill)">
							<AppIcon name="trash" :size="13" />
							删除
						</button>
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- 安装技能(抽屉) -->
	<div v-if="showInstall" class="drawer-layer" @click.self="showInstall = false">
		<aside class="drawer" role="dialog" aria-modal="true" aria-label="安装技能">
			<header class="drawer-header">
				<h2>安装技能</h2>
				<button class="btn btn-ghost btn-sm" type="button" @click="showInstall = false">关闭</button>
			</header>
			<div class="drawer-body">
				<div class="modal-field">
					<label for="skill-repo">仓库地址</label>
					<input id="skill-repo" v-model="installRepoUrl" type="text" autocomplete="off" spellcheck="false" placeholder="https://github.com/xxx/skill" />
				</div>
				<div class="modal-field">
					<label for="skill-name">技能名称（可选）</label>
					<input id="skill-name" v-model="installName" type="text" autocomplete="off" spellcheck="false" placeholder="留空则取仓库名" />
				</div>
				<div class="modal-field">
					<label for="skill-vis">可见性</label>
					<select id="skill-vis" v-model="installVisibility">
						<option value="private">私有</option>
						<option value="group">组共享</option>
						<option value="public">公开</option>
					</select>
				</div>
				<div v-if="installVisibility === 'group'" class="modal-field">
					<label for="skill-group">所属组</label>
					<select id="skill-group" v-model="installGroupId">
						<option value="" disabled>选择组</option>
						<option v-for="g in myGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
					</select>
				</div>
				<button
					class="btn btn-primary drawer-submit"
					type="button"
					:disabled="!installRepoUrl.trim() || installBusy || (installVisibility === 'group' && !installGroupId)"
					@click="submitInstall"
				>
					{{ installBusy ? "安装中…" : "安装技能" }}
				</button>
			</div>
		</aside>
	</div>

	<!-- 自研上传(抽屉) -->
	<div v-if="showUpload" class="drawer-layer" @click.self="showUpload = false">
		<aside class="drawer" role="dialog" aria-modal="true" aria-label="自研上传技能">
			<header class="drawer-header">
				<h2>自研上传技能</h2>
				<button class="btn btn-ghost btn-sm" type="button" @click="showUpload = false">关闭</button>
			</header>
			<div class="drawer-body">
				<div class="modal-field">
					<label for="skill-upload-name">技能名称</label>
					<input id="skill-upload-name" v-model="uploadName" type="text" autocomplete="off" spellcheck="false" placeholder="例如 财报分析技能" />
				</div>
				<div class="modal-field">
					<label for="skill-upload-vis">可见性</label>
					<select id="skill-upload-vis" v-model="uploadVisibility">
						<option value="private">私有</option>
						<option value="group">组共享</option>
						<option value="public">公开</option>
					</select>
				</div>
				<div v-if="uploadVisibility === 'group'" class="modal-field">
					<label for="skill-upload-group">所属组</label>
					<select id="skill-upload-group" v-model="uploadGroupId">
						<option value="" disabled>选择组</option>
						<option v-for="g in myGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
					</select>
				</div>
				<button
					class="btn btn-primary drawer-submit"
					type="button"
					:disabled="!uploadName.trim() || uploadBusy || (uploadVisibility === 'group' && !uploadGroupId)"
					@click="submitUpload"
				>
					{{ uploadBusy ? "登记中…" : "登记技能" }}
				</button>
			</div>
		</aside>
	</div>
</template>

<style scoped>
.mono {
	font-family: var(--font-mono);
	font-size: 11px;
}

.card-desc {
	font-family: var(--font-mono);
	font-size: 11px;
}

.card-badges {
	display: flex;
	align-items: center;
	flex-wrap: wrap;
	gap: 6px;
	margin-top: 2px;
}

.card-actions {
	display: flex;
	align-items: center;
	gap: 6px;
	margin-top: 10px;
	padding-top: 10px;
	border-top: 1px solid var(--border);
}

.vis-select {
	background: var(--bg-input);
	color: var(--text-primary);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 4px 8px;
	font-size: 12px;
}

.spacer {
	flex: 1;
}

.text-btn {
	display: inline-flex;
	align-items: center;
	gap: 4px;
}

.drawer-submit {
	margin-top: 16px;
	width: 100%;
	height: 40px;
}
</style>