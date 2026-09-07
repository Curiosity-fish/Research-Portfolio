<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import type { ProjectFileEntry, ProjectInfo } from "../api/types.ts";
import {
	abortCurrent,
	deleteEntry,
	refreshApprovals,
	refreshDir,
	refreshPermissionMode,
	saveCurrentFile,
	selectFile,
	setPermissionMode,
	store,
	sendPrompt,
	sendSteer,
	toggleDir,
	transcriptItems,
	uploadFiles,
} from "../state/session-store.ts";
import ApprovalDialog from "./ApprovalDialog.vue";
import MessageItem from "./MessageItem.vue";
import MonacoEditor from "./MonacoEditor.vue";

const props = defineProps<{ project: ProjectInfo | null }>();
const emit = defineEmits<{ close: [] }>();

const workbench = computed(() => store.workbench);
const items = computed(() => transcriptItems());
const draft = ref("");

const selectedPath = computed(() => workbench.value.selectedPath);
const fileName = computed(() => selectedPath.value?.split("/").pop() ?? "");
const ext = computed(() => {
	const dot = fileName.value.lastIndexOf(".");
	return dot >= 0 ? fileName.value.slice(dot).toLowerCase() : "";
});

const editable = computed(() =>
	[
		".txt", ".md", ".json", ".csv", ".py", ".ts", ".tsx", ".js", ".mjs", ".cjs", ".jsx",
		".vue", ".html", ".css", ".yaml", ".yml", ".toml", ".ini", ".sh", ".ps1", ".bat",
		".sql", ".log", ".xml", ".svg",
	].includes(ext.value),
);
const isImage = computed(() => [".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"].includes(ext.value));
const isPdf = computed(() => ext.value === ".pdf");
const isCsv = computed(() => ext.value === ".csv");

const previewUrl = computed(() =>
	props.project && selectedPath.value
		? `/api/projects/${props.project.id}/files/download?path=${encodeURIComponent(selectedPath.value)}`
		: "",
);

const currentDir = computed(() => {
	const path = selectedPath.value;
	if (!path) return "";
	const index = path.lastIndexOf("/");
	return index > 0 ? path.slice(0, index) : "";
});

const isGenerating = computed(() => {
	if (store.busy) return true;
	const phase = store.snapshot?.phase;
	return phase === "turn" || phase === "compaction" || phase === "retry";
});

const csvRows = computed(() => {
	if (!isCsv.value || !workbench.value.fileContent) return [];
	return workbench.value.fileContent
		.split(/\r?\n/)
		.filter((line) => line.length > 0)
		.slice(0, 200)
		.map((line) => line.split(","));
});

function visibleRows(): Array<{ entry: ProjectFileEntry; depth: number }> {
	const rows: Array<{ entry: ProjectFileEntry; depth: number }> = [];
	const walk = (dir: string, depth: number): void => {
		for (const entry of workbench.value.tree[dir] ?? []) {
			rows.push({ entry, depth });
			if (entry.type === "dir" && workbench.value.expandedDirs.includes(entry.path)) walk(entry.path, depth + 1);
		}
	};
	walk("", 0);
	return rows;
}

function send(): void {
	const text = draft.value.trim();
	if (!text || !store.currentId) return;
	draft.value = "";
	void sendPrompt(text);
}

function steer(): void {
	const text = draft.value.trim();
	if (!text || !store.currentId) return;
	draft.value = "";
	void sendSteer(text);
}

function onUpload(event: Event): void {
	const input = event.target as HTMLInputElement;
	if (input.files) void uploadFiles(input.files, currentDir.value || undefined);
	input.value = "";
}

function changePermissionMode(event: Event): void {
	const mode = (event.target as HTMLSelectElement).value;
	void setPermissionMode(mode);
}

let approvalTimer: ReturnType<typeof setInterval> | undefined;

onMounted(() => {
	void refreshPermissionMode();
	approvalTimer = setInterval(() => {
		if (store.currentId) void refreshApprovals();
	}, 2000);
});

onBeforeUnmount(() => {
	if (approvalTimer) clearInterval(approvalTimer);
});

function onDelete(path: string): void {
	if (!window.confirm(`确定删除 ${path}？`)) return;
	void deleteEntry(path);
}
</script>

<template>
	<div class="workbench">
		<aside class="wb-side">
			<header class="wb-side-header">
				<span class="wb-project-name" :title="project?.name">{{ project?.name ?? "项目" }}</span>
				<div class="wb-side-actions">
					<button class="wb-icon-btn" type="button" title="刷新" @click="refreshDir('')">↻</button>
					<label class="wb-icon-btn wb-upload" title="上传文件">
						⬆
						<input class="wb-file-input" type="file" multiple @change="onUpload" />
					</label>
					<button class="wb-icon-btn" type="button" title="关闭工作台" @click="emit('close')">✕</button>
				</div>
			</header>
			<div class="wb-tree">
				<p v-if="workbench.loading && !selectedPath" class="wb-hint">加载中…</p>
				<p v-else-if="visibleRows().length === 0" class="wb-hint">工作区为空，点击上方 ⬆ 上传文件</p>
				<div
					v-for="{ entry, depth } in visibleRows()"
					:key="entry.path"
					class="wb-tree-row"
					:class="{ selected: selectedPath === entry.path }"
					:style="{ paddingLeft: `${8 + depth * 14}px` }"
					@click="entry.type === 'dir' ? toggleDir(entry.path) : selectFile(entry.path)"
				>
					<span class="wb-tree-icon">{{ entry.type === "dir" ? (workbench.expandedDirs.includes(entry.path) ? "▾" : "▸") : "·" }}</span>
					<span class="wb-tree-name">{{ entry.name }}</span>
					<span v-if="entry.type === 'file'" class="wb-tree-size">{{ entry.size ?? "" }}</span>
				</div>
			</div>
		</aside>

		<main class="wb-center">
			<template v-if="!selectedPath">
				<p class="wb-hint">从左侧选择文件，或上传文件到项目工作区</p>
			</template>
			<template v-else-if="editable">
				<header class="wb-center-header">
					<span class="wb-file-name">{{ fileName }}</span>
					<span class="wb-file-meta">{{ workbench.fileMeta?.size ?? 0 }} B</span>
					<div class="wb-center-actions">
						<button class="btn btn-primary wb-save" type="button" @click="saveCurrentFile()">保存</button>
						<button class="btn btn-ghost" type="button" @click="onDelete(selectedPath)">删除</button>
					</div>
				</header>
				<div class="wb-editor-host">
					<MonacoEditor :model-value="workbench.fileContent ?? ''" :language="ext.slice(1) || 'plaintext'" @update:model-value="workbench.fileContent = $event" />
				</div>
			</template>
			<template v-else-if="isCsv">
				<header class="wb-center-header">
					<span class="wb-file-name">{{ fileName }}</span>
					<span class="wb-file-meta">{{ workbench.fileMeta?.size ?? 0 }} B</span>
				</header>
				<div class="wb-csv-preview">
					<table>
						<tbody>
							<tr v-for="(row, i) in csvRows" :key="i">
								<td v-for="(cell, j) in row" :key="j">{{ cell }}</td>
							</tr>
						</tbody>
					</table>
				</div>
			</template>
			<template v-else-if="isPdf">
				<iframe class="wb-preview" :src="previewUrl" title="PDF 预览" />
			</template>
			<template v-else-if="isImage">
				<div class="wb-image-preview">
					<img :src="previewUrl" :alt="fileName" />
				</div>
			</template>
			<template v-else>
				<header class="wb-center-header">
					<span class="wb-file-name">{{ fileName }}</span>
					<span class="wb-file-meta">{{ workbench.fileMeta?.size ?? 0 }} B · {{ workbench.fileMeta?.mimeType }}</span>
				</header>
				<div class="wb-download">
					<a class="btn btn-primary" :href="previewUrl" download>下载文件</a>
					<button class="btn btn-ghost" type="button" @click="onDelete(selectedPath)">删除</button>
				</div>
			</template>
		</main>

		<aside class="wb-chat">
			<header class="wb-chat-header">
				<span>Agent 对话</span>
				<div class="wb-chat-header-right">
					<span v-if="isGenerating" class="wb-phase">生成中…</span>
					<select class="wb-mode-select" :value="store.permissionMode ?? 'full_access'" @change="changePermissionMode" title="会话权限模式">
						<option value="full_access">完全访问</option>
						<option value="request_approval">请求批准</option>
					</select>
				</div>
			</header>
			<div class="wb-chat-messages">
				<MessageItem v-for="item in items" :key="item.id" :item="item" />
				<p v-if="items.length === 0" class="wb-hint">向 Agent 提问，让它读写项目文件</p>
			</div>
			<div class="wb-composer">
				<textarea
					v-model="draft"
					rows="3"
					placeholder="输入消息，Enter 发送，Shift+Enter 换行"
					@keydown.enter.exact.prevent="send"
				/>
				<div class="wb-composer-actions">
					<button v-if="isGenerating" class="btn btn-ghost" type="button" @click="abortCurrent()">停止</button>
					<button v-else class="btn btn-ghost" type="button" @click="steer">插话</button>
					<button class="btn btn-primary" type="button" :disabled="!draft.trim() || !store.currentId" @click="send">发送</button>
				</div>
			</div>
		</aside>

		<ApprovalDialog />
	</div>
</template>

<style scoped>
.workbench {
	display: grid;
	grid-template-columns: 240px 1fr 360px;
	height: 100%;
	min-height: 0;
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	overflow: hidden;
	background: var(--bg-app);
}

.wb-side {
	display: flex;
	flex-direction: column;
	border-right: 1px solid var(--border);
	background: var(--bg-card);
	min-height: 0;
}

.wb-side-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 8px;
	padding: 10px 12px;
	border-bottom: 1px solid var(--border);
}

.wb-project-name {
	font-weight: 600;
	font-size: 13px;
	color: var(--text-primary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.wb-side-actions {
	display: flex;
	gap: 4px;
}

.wb-icon-btn {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	width: 24px;
	height: 24px;
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	background: var(--bg-card);
	color: var(--text-muted);
	font-size: 12px;
	cursor: pointer;
}

.wb-icon-btn:hover {
	color: var(--text-primary);
	border-color: var(--border-strong);
}

.wb-upload {
	position: relative;
	overflow: hidden;
}

.wb-file-input {
	position: absolute;
	inset: 0;
	opacity: 0;
	cursor: pointer;
}

.wb-tree {
	flex: 1;
	overflow-y: auto;
	padding: 6px 0;
	font-family: var(--font-mono);
	font-size: 12px;
}

.wb-tree-row {
	display: flex;
	align-items: center;
	gap: 6px;
	padding: 4px 8px;
	cursor: pointer;
	color: var(--text-secondary);
	white-space: nowrap;
}

.wb-tree-row:hover {
	background: var(--bg-hover);
}

.wb-tree-row.selected {
	background: var(--picker-hover-bg);
	color: var(--text-primary);
}

.wb-tree-icon {
	flex-shrink: 0;
	width: 12px;
	color: var(--text-muted);
}

.wb-tree-name {
	flex: 1;
	overflow: hidden;
	text-overflow: ellipsis;
}

.wb-tree-size {
	flex-shrink: 0;
	font-size: 11px;
	color: var(--text-disabled);
}

.wb-center {
	display: flex;
	flex-direction: column;
	min-height: 0;
	border-right: 1px solid var(--border);
	background: var(--bg-app);
}

.wb-center-header {
	display: flex;
	align-items: center;
	gap: 12px;
	padding: 8px 14px;
	border-bottom: 1px solid var(--border);
	background: var(--bg-card);
}

.wb-file-name {
	font-weight: 600;
	font-size: 13px;
	color: var(--text-primary);
}

.wb-file-meta {
	font-size: 11px;
	color: var(--text-muted);
}

.wb-center-actions {
	margin-left: auto;
	display: flex;
	gap: 8px;
}

.wb-save {
	height: 28px;
	padding: 0 14px;
}

.wb-editor-host {
	flex: 1;
	min-height: 0;
	padding: 4px;
}

.wb-csv-preview {
	flex: 1;
	overflow: auto;
	padding: 12px;
}

.wb-csv-preview table {
	border-collapse: collapse;
	font-size: 12px;
}

.wb-csv-preview td {
	border: 1px solid var(--border);
	padding: 4px 8px;
}

.wb-preview {
	flex: 1;
	width: 100%;
	border: none;
}

.wb-image-preview {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 16px;
	overflow: auto;
}

.wb-image-preview img {
	max-width: 100%;
	max-height: 100%;
}

.wb-download {
	display: flex;
	gap: 10px;
	align-items: center;
	padding: 20px;
}

.wb-chat {
	display: flex;
	flex-direction: column;
	min-height: 0;
	background: var(--bg-card);
}

.wb-chat-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 10px 12px;
	border-bottom: 1px solid var(--border);
	font-weight: 600;
	font-size: 13px;
	color: var(--text-primary);
}

.wb-phase {
	font-size: 11px;
	font-weight: 400;
	color: var(--accent-strong);
}

.wb-chat-header-right {
	display: flex;
	align-items: center;
	gap: 10px;
}

.wb-mode-select {
	background: var(--bg-input);
	color: var(--text-primary);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 4px 8px;
	font-size: 12px;
	font-family: var(--font-mono);
}

.wb-mode-select:focus {
	outline: none;
	border-color: var(--accent);
}

.wb-chat-messages {
	flex: 1;
	overflow-y: auto;
	padding: 12px;
}

.wb-composer {
	border-top: 1px solid var(--border);
	padding: 10px;
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.wb-composer textarea {
	width: 100%;
	background: var(--bg-input);
	color: var(--text-primary);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 8px 10px;
	font-size: 13px;
	font-family: var(--font-mono);
	resize: vertical;
}

.wb-composer textarea:focus {
	outline: none;
	border-color: var(--accent);
}

.wb-composer-actions {
	display: flex;
	justify-content: flex-end;
	gap: 8px;
}

.wb-hint {
	margin: 16px;
	font-size: 12px;
	color: var(--text-disabled);
}

@media (max-width: 900px) {
	.workbench {
		grid-template-columns: 1fr;
		grid-template-rows: 200px 1fr 280px;
	}
}
</style>
