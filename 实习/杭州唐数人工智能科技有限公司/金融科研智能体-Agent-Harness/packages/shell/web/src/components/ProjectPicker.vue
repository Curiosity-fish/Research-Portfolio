<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import type { GroupView, ProjectInfo } from "../api/types.ts";
import AppIcon from "./AppIcon.vue";

const props = defineProps<{
	open: boolean;
	mode: "create" | "move";
	projects: ProjectInfo[];
	/** 我所在的课程/课题组(建组时可选挂靠)。 */
	groups: GroupView[];
}>();

const emit = defineEmits<{
	close: [];
	/** 创建项目 */
	create: [input: { name: string; folder?: string }];
	/** 移动会话到项目(projectId 为 null 表示无项目) */
	move: [projectId: string | null];
}>();

const modalRef = ref<HTMLElement | null>(null);
const restoreFocus = ref<HTMLElement | null>(null);

// create 模式
const projectName = ref("");
const projectFolder = ref("");

// move 模式:null = 明确选择「无项目」,undefined 状态用 selectedNone 区分
const selectedProjectId = ref<string | null>(null);
const selectedGroupId = ref("");
const selectedNone = ref(false);

const canConfirm = computed(() => {
	if (props.mode === "create") return projectName.value.trim().length > 0;
	return selectedNone.value || selectedProjectId.value !== null;
});

/** 可选挂靠的 active 课程组。 */
const activeGroups = computed(() => props.groups.filter((group) => group.status === "active"));

/** move 模式:按 私有 / 课程组 分组展示项目。 */
const moveSections = computed(() => {
	const privateProjects: ProjectInfo[] = [];
	const byGroup = new Map<string, ProjectInfo[]>();
	for (const project of props.projects) {
		if (project.scope.type === "private") {
			privateProjects.push(project);
		} else {
			const list = byGroup.get(project.scope.groupId) ?? [];
			list.push(project);
			byGroup.set(project.scope.groupId, list);
		}
	}
	const sections: Array<{ label: string; projects: ProjectInfo[] }> = [];
	if (privateProjects.length > 0) sections.push({ label: "私有项目", projects: privateProjects });
	for (const [groupId, projects] of byGroup) {
		const group = props.groups.find((item) => item.id === groupId);
		sections.push({ label: group ? group.name : "课程组", projects });
	}
	return sections;
});

const title = computed(() => (props.mode === "create" ? "创建项目" : "移动到项目"));
const desc = computed(() =>
	props.mode === "create"
		? "项目是对话的分类；挂靠课程组后可加载该组技能与数据集"
		: "把该对话移动到所选项目；选择「无项目」可移出项目",
);

function reset(): void {
	projectName.value = "";
	projectFolder.value = "";
	selectedProjectId.value = null;
	selectedNone.value = false;
	selectedGroupId.value = "";
}

function selectProject(id: string): void {
	selectedProjectId.value = id;
	selectedNone.value = false;
}

function selectNone(): void {
	selectedProjectId.value = null;
	selectedNone.value = true;
}

function confirm(): void {
	if (props.mode === "create") {
		emit("create", {
			name: projectName.value.trim(),
			folder: projectFolder.value.trim() || undefined,
			...(selectedGroupId.value ? { scope: { type: "group" as const, groupId: selectedGroupId.value } } : {}),
		});
		return;
	}
	if (selectedNone.value) {
		emit("move", null);
		return;
	}
	if (selectedProjectId.value !== null) {
		emit("move", selectedProjectId.value);
	}
}

function onKeydown(event: KeyboardEvent): void {
	if (event.key === "Escape") {
		event.preventDefault();
		emit("close");
		return;
	}
	if (event.key === "Tab") {
		const modal = modalRef.value;
		if (!modal) return;
		const focusable = Array.from(
			modal.querySelectorAll<HTMLElement>('button, input, [tabindex]:not([tabindex="-1"])'),
		).filter((el) => !el.hasAttribute("disabled"));
		if (focusable.length === 0) return;
		const first = focusable[0];
		const last = focusable[focusable.length - 1];
		if (event.shiftKey && document.activeElement === first) {
			event.preventDefault();
			last.focus();
		} else if (!event.shiftKey && document.activeElement === last) {
			event.preventDefault();
			first.focus();
		}
	}
}

watch(
	() => props.open,
	(open) => {
		if (open) {
			restoreFocus.value = document.activeElement as HTMLElement | null;
			nextTick(() => {
				const modal = modalRef.value;
				if (!modal) return;
				const first = modal.querySelector<HTMLElement>('button:not([disabled]), input');
				(first ?? modal).focus();
			});
		} else {
			reset();
			restoreFocus.value?.focus?.();
			restoreFocus.value = null;
		}
	},
);
</script>

<template>
	<div v-if="open" class="picker-layer" @click.self="emit('close')">
		<div
			ref="modalRef"
			class="picker"
			role="dialog"
			aria-modal="true"
			:aria-label="title"
			tabindex="-1"
			@keydown="onKeydown"
		>
			<header class="picker-header">
				<div class="picker-heading">
					<h2 class="picker-title">{{ title }}</h2>
					<p class="picker-desc">{{ desc }}</p>
				</div>
				<button class="picker-close" type="button" aria-label="关闭" @click="emit('close')">
					<AppIcon name="x" :size="16" />
				</button>
			</header>

			<div class="picker-body">
				<template v-if="mode === 'create'">
					<div class="picker-field">
						<label for="project-name-input">项目名称</label>
						<input
							id="project-name-input"
							v-model="projectName"
							type="text"
							autocomplete="off"
							spellcheck="false"
							placeholder="例如 新能源行业研究"
							@keydown.enter.prevent="canConfirm && confirm()"
						/>
					</div>
					<div class="picker-field">
						<label for="project-folder-input">本地文件夹（可选）</label>
						<input
							id="project-folder-input"
							v-model="projectFolder"
							type="text"
							autocomplete="off"
							spellcheck="false"
							placeholder="例如 C:\Users\...\研究\新能源"
						/>
						<p class="picker-hint">留空则该项目对话使用默认工作目录</p>
					</div>
					<div class="picker-field">
						<label for="project-group-select">所属课程组（可选）</label>
						<select id="project-group-select" v-model="selectedGroupId" class="picker-field-input">
							<option value="">私有项目</option>
							<option v-for="group in activeGroups" :key="group.id" :value="group.id">{{ group.name }}</option>
						</select>
						<p class="picker-hint">挂靠课程组后，会话可加载该组技能与数据集</p>
					</div>
				</template>

				<template v-else>
					<p class="picker-label">选择目标项目</p>
					<template v-for="section in moveSections" :key="section.label">
						<p class="picker-sub-label">{{ section.label }}</p>
						<ul class="project-options">
							<li v-for="project in section.projects" :key="project.id">
								<button
									class="project-option"
									type="button"
									:class="{ selected: selectedProjectId === project.id && !selectedNone }"
									:aria-pressed="selectedProjectId === project.id && !selectedNone"
									@click="selectProject(project.id)"
								>
									<AppIcon name="folder" :size="16" class="project-option-icon" />
									<span class="project-option-text">
										<span class="project-option-name">{{ project.name }}</span>
										<span class="project-option-path" :title="project.folder">{{ project.folder || "未绑定文件夹" }}</span>
									</span>
									<span class="project-option-count">{{ project.sessionIds.length }}</span>
									<AppIcon
										v-if="selectedProjectId === project.id && !selectedNone"
										name="check"
										:size="14"
										class="project-option-check"
									/>
								</button>
							</li>
						</ul>
					</template>
					<div class="picker-divider" />
					<button
						class="project-option project-option-none"
						type="button"
						:class="{ selected: selectedNone }"
						:aria-pressed="selectedNone"
						@click="selectNone"
					>
						<AppIcon name="x" :size="16" class="project-option-icon" />
						<span class="project-option-text">
							<span class="project-option-name">无项目</span>
							<span class="project-option-path">移出项目，仅保留在历史对话</span>
						</span>
						<AppIcon v-if="selectedNone" name="check" :size="14" class="project-option-check" />
					</button>
				</template>
			</div>

			<footer class="picker-footer">
				<button class="btn btn-ghost" type="button" @click="emit('close')">取消</button>
				<button class="btn btn-primary picker-confirm" type="button" :disabled="!canConfirm" @click="confirm">
					<AppIcon name="check" :size="16" />
					<span>{{ mode === "create" ? "创建项目" : "移动" }}</span>
				</button>
			</footer>
		</div>
	</div>
</template>

<style scoped>
.picker-layer {
	position: fixed;
	inset: 0;
	z-index: var(--z-drawer);
	background: rgba(24, 42, 62, 0.45);
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 20px;
	animation: picker-fade 160ms ease;
}

.picker {
	width: min(460px, 100%);
	max-height: min(600px, 100%);
	display: flex;
	flex-direction: column;
	background: var(--bg-surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-float);
	animation: picker-pop 180ms ease;
	outline: none;
}

@keyframes picker-fade {
	from {
		opacity: 0;
	}
	to {
		opacity: 1;
	}
}

@keyframes picker-pop {
	from {
		transform: translateY(8px) scale(0.98);
		opacity: 0.6;
	}
	to {
		transform: translateY(0) scale(1);
		opacity: 1;
	}
}

.picker-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 12px;
	padding: 18px 20px 14px;
	border-bottom: 1px solid var(--border);
}

.picker-title {
	margin: 0;
	font-size: 16px;
	font-weight: 600;
	letter-spacing: -0.01em;
	color: var(--text-primary);
}

.picker-desc {
	margin: 4px 0 0;
	font-size: 12px;
	font-weight: 400;
	color: var(--text-muted);
	line-height: 1.55;
}

.picker-close {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	width: 30px;
	height: 30px;
	flex-shrink: 0;
	background: transparent;
	border: none;
	border-radius: var(--radius-sm);
	color: var(--text-muted);
	cursor: pointer;
	transition:
		background 140ms ease,
		color 140ms ease;
}

.picker-close:hover {
	background: var(--bg-hover);
	color: var(--text-primary);
}

.picker-body {
	flex: 1;
	overflow-y: auto;
	padding: 16px 20px 8px;
	overscroll-behavior: contain;
}

.picker-label {
	margin: 0 0 8px;
	font-size: 11px;
	font-weight: 500;
	letter-spacing: 0.04em;
	color: var(--text-muted);
}

.picker-field {
	display: flex;
	flex-direction: column;
	gap: 7px;
	margin-bottom: 14px;
}

.picker-field label {
	font-size: 12px;
	font-weight: 500;
	color: var(--text-muted);
}

.picker-field input {
	width: 100%;
	background: var(--bg-input);
	color: var(--text-primary);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 9px 11px;
	font-size: 13px;
	font-family: var(--font-mono);
	transition: border-color 140ms ease;
}

.picker-field input:hover {
	border-color: var(--border-strong);
}

.picker-field input:focus {
	outline: none;
	border-color: var(--accent);
	box-shadow: 0 0 0 4px rgba(116, 185, 235, 0.1);
}

.picker-field input::placeholder {
	color: var(--text-disabled);
}

.picker-hint {
	margin: 0;
	font-size: 11px;
	font-weight: 400;
	color: var(--text-muted);
}

.picker-field-input {
	width: 100%;
	background: var(--bg-input);
	color: var(--text-primary);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 9px 11px;
	font-size: 13px;
	font-family: var(--font-mono);
}

.picker-field-input:focus {
	outline: none;
	border-color: var(--accent);
}

.picker-sub-label {
	margin: 10px 0 6px;
	font-size: 11px;
	font-weight: 600;
	letter-spacing: 0.04em;
	color: var(--accent-strong);
}

.project-options {
	list-style: none;
	margin: 0;
	padding: 0;
	display: flex;
	flex-direction: column;
	gap: 6px;
}

.project-option {
	display: flex;
	align-items: center;
	gap: 10px;
	width: 100%;
	padding: 10px 12px;
	border: 1px solid var(--border);
	border-radius: var(--radius-md);
	background: var(--bg-card);
	color: var(--text-primary);
	text-align: left;
	cursor: pointer;
	transition:
		border-color 140ms ease,
		background 140ms ease;
}

.project-option:hover {
	border-color: var(--border-strong);
	background: var(--picker-hover-bg);
}

.project-option.selected {
	border-color: var(--accent);
	background: var(--picker-hover-bg);
}

.project-option-icon {
	flex-shrink: 0;
	color: var(--accent-strong);
}

.project-option-none {
	margin-top: 2px;
}

.project-option-none .project-option-icon {
	color: var(--text-muted);
}

.project-option-text {
	min-width: 0;
	flex: 1;
	display: flex;
	flex-direction: column;
	gap: 2px;
}

.project-option-name {
	font-size: 13px;
	font-weight: 600;
	color: var(--text-primary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.project-option-path {
	font-size: 11px;
	font-weight: 400;
	color: var(--text-muted);
	font-family: var(--font-mono);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.project-option-count {
	flex-shrink: 0;
	font-size: 11px;
	font-weight: 500;
	color: var(--text-muted);
}

.project-option-check {
	flex-shrink: 0;
	color: var(--accent-strong);
}

.picker-divider {
	height: 1px;
	background: var(--border);
	margin: 14px 0;
}

.picker-footer {
	display: flex;
	align-items: center;
	justify-content: flex-end;
	gap: 10px;
	padding: 14px 20px 18px;
	border-top: 1px solid var(--border);
}

.picker-confirm {
	height: 38px;
	padding: 0 18px;
}

@media (max-width: 640px) {
	.picker-layer {
		padding: 12px;
	}

	.picker-footer {
		flex-direction: column-reverse;
		align-items: stretch;
	}

	.picker-footer .btn {
		width: 100%;
	}
}
</style>
