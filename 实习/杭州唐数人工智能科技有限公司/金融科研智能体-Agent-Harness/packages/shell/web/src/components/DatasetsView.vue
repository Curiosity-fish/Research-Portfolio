<script setup lang="ts">
import { computed, ref } from "vue";
import type { DatasetView, ResourceVisibility } from "../api/types.ts";
import AppIcon from "./AppIcon.vue";
import {
	createDataset,
	deleteDataset,
	downloadDataset,
	refreshDatasets,
	store,
	updateDataset,
} from "../state/session-store.ts";

const showUpload = ref(false);
const uploadBusy = ref(false);
const uploadFile = ref<File | null>(null);
const uploadName = ref("");
const uploadVisibility = ref<ResourceVisibility>("private");
const uploadGroupId = ref<string>("");
const fileInput = ref<HTMLInputElement | null>(null);

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

function canEdit(dataset: DatasetView): boolean {
	return store.user?.role === "admin" || dataset.ownerId === store.user?.id;
}

function formatSize(bytes: number): string {
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
	return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`;
}

function onFileChange(event: Event): void {
	const input = event.target as HTMLInputElement;
	uploadFile.value = input.files?.[0] ?? null;
	if (uploadFile.value && !uploadName.value) uploadName.value = uploadFile.value.name;
}

function resetUpload(): void {
	uploadFile.value = null;
	uploadName.value = "";
	uploadVisibility.value = "private";
	uploadGroupId.value = "";
	if (fileInput.value) fileInput.value.value = "";
}

async function submitUpload(): Promise<void> {
	if (!uploadFile.value) return;
	uploadBusy.value = true;
	try {
		await createDataset({
			file: uploadFile.value,
			name: uploadName.value.trim() || undefined,
			visibility: uploadVisibility.value,
			groupId: uploadVisibility.value === "group" && uploadGroupId.value ? uploadGroupId.value : undefined,
		});
		showUpload.value = false;
		resetUpload();
	} finally {
		uploadBusy.value = false;
	}
}

async function handleVisibilityChange(dataset: DatasetView, visibility: ResourceVisibility): Promise<void> {
	await updateDataset(dataset.id, { visibility });
}

async function handleDownload(dataset: DatasetView): Promise<void> {
	await downloadDataset(dataset.id);
}

async function handleDelete(dataset: DatasetView): Promise<void> {
	if (!window.confirm(`确定删除数据集「${dataset.name}」？`)) return;
	await deleteDataset(dataset.id);
}

async function doRefresh(): Promise<void> {
	refreshing.value = true;
	try {
		await refreshDatasets();
	} finally {
		refreshing.value = false;
	}
}
</script>

<template>
	<section class="page">
		<header class="page-header">
			<div>
				<h1 class="page-title">数据集库</h1>
				<p class="page-subtitle">上传共享数据集，供项目会话只读挂载</p>
			</div>
			<div class="page-actions">
				<button class="btn btn-ghost" type="button" :disabled="refreshing" @click="doRefresh">
					<AppIcon name="refresh" :size="15" />
					<span>刷新</span>
				</button>
				<button class="btn btn-primary" type="button" @click="showUpload = true">
					<AppIcon name="upload" :size="15" />
					<span>上传数据集</span>
				</button>
			</div>
		</header>

		<div class="page-body">
			<div v-if="store.datasets.length === 0" class="page-empty">
				数据集库为空。点击「上传数据集」上传第一个文件。
			</div>
			<div v-else class="card-grid">
				<div v-for="dataset in store.datasets" :key="dataset.id" class="platform-card">
					<div class="card-title-row">
						<span class="card-icon"><AppIcon name="database" :size="16" /></span>
						<span class="card-title" :title="dataset.name">{{ dataset.name }}</span>
					</div>
					<div class="card-desc mono">{{ formatSize(dataset.sizeBytes) }}</div>
					<div class="card-badges">
						<span class="badge" :class="visBadgeClass(dataset.visibility)">{{ visLabel(dataset.visibility) }}</span>
						<span class="badge" :class="reviewBadge(dataset.reviewState).cls">{{ reviewBadge(dataset.reviewState).text }}</span>
						<span v-if="dataset.groupName" class="badge badge--gray">{{ dataset.groupName }}</span>
					</div>
					<div class="card-actions">
						<select
							v-if="canEdit(dataset)"
							class="vis-select"
							:value="dataset.visibility"
							:aria-label="`${dataset.name} 可见性`"
							@change="handleVisibilityChange(dataset, ($event.target as HTMLSelectElement).value as ResourceVisibility)"
						>
							<option value="private">私有</option>
							<option value="group">组共享</option>
							<option value="public">公开</option>
						</select>
						<div class="spacer" />
						<button class="btn btn-ghost btn-sm" type="button" @click="handleDownload(dataset)">
							<AppIcon name="download" :size="13" />
							下载
						</button>
						<button v-if="canEdit(dataset)" class="text-btn danger" type="button" @click="handleDelete(dataset)">
							<AppIcon name="trash" :size="13" />
							删除
						</button>
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- 上传数据集(抽屉) -->
	<div v-if="showUpload" class="drawer-layer" @click.self="showUpload = false">
		<aside class="drawer" role="dialog" aria-modal="true" aria-label="上传数据集">
			<header class="drawer-header">
				<h2>上传数据集</h2>
				<button class="btn btn-ghost btn-sm" type="button" @click="showUpload = false">关闭</button>
			</header>
			<div class="drawer-body">
				<div class="modal-field">
					<label for="dataset-file">文件</label>
					<input ref="fileInput" id="dataset-file" type="file" accept="*/*" @change="onFileChange" />
				</div>
				<div class="modal-field">
					<label for="dataset-name">数据集名称（可选）</label>
					<input id="dataset-name" v-model="uploadName" type="text" autocomplete="off" spellcheck="false" placeholder="留空则用文件名" />
				</div>
				<div class="modal-field">
					<label for="dataset-vis">可见性</label>
					<select id="dataset-vis" v-model="uploadVisibility">
						<option value="private">私有</option>
						<option value="group">组共享</option>
						<option value="public">公开</option>
					</select>
				</div>
				<div v-if="uploadVisibility === 'group'" class="modal-field">
					<label for="dataset-group">所属组</label>
					<select id="dataset-group" v-model="uploadGroupId">
						<option value="" disabled>选择组</option>
						<option v-for="g in myGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
					</select>
				</div>
				<button
					class="btn btn-primary drawer-submit"
					type="button"
					:disabled="!uploadFile || uploadBusy || (uploadVisibility === 'group' && !uploadGroupId)"
					@click="submitUpload"
				>
					{{ uploadBusy ? "上传中…" : "上传数据集" }}
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