<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import type { SessionSummary, UserInfo } from "../api/types.ts";
import type { SessionGrouping } from "../state/projects.ts";
import AppIcon, { type IconName } from "./AppIcon.vue";
import ConnectionBadge from "./ConnectionBadge.vue";
import ThemeToggle from "./ThemeToggle.vue";

const props = defineProps<{
	grouping: SessionGrouping;
	currentId: string | null;
	open: boolean;
	connected: boolean;
	user: UserInfo | null;
	view: string;
}>();

const emit = defineEmits<{
	select: [sessionId: string];
	home: [];
	remove: [sessionId: string];
	rename: [sessionId: string];
	move: [sessionId: string];
	export: [sessionId: string];
	close: [];
	"open-tools": [];
	"open-settings": [];
	navigate: [view: "groups" | "skills" | "datasets" | "agents" | "admin" | "monitor"];
	"new-chat-in-project": [projectId: string];
	"rename-project": [projectId: string];
	"delete-project": [projectId: string];
}>();

type PlatformView = "groups" | "skills" | "datasets" | "agents" | "admin" | "monitor";

const NAV_ITEMS: Array<{ view: PlatformView; label: string; icon: IconName; adminOnly?: boolean }> = [
	{ view: "groups", label: "课程组", icon: "users" },
	{ view: "skills", label: "技能库", icon: "layers" },
	{ view: "datasets", label: "数据集库", icon: "database" },
	{ view: "agents", label: "智能体广场", icon: "eye" },
	{ view: "monitor", label: "会话监控", icon: "terminal", adminOnly: true },
	{ view: "admin", label: "管理端", icon: "shield", adminOnly: true },
];

const visibleNav = computed(() => NAV_ITEMS.filter((item) => !item.adminOnly || props.user?.role === "admin"));

const searchQuery = ref("");
const historyProjectsCollapsed = ref(true);
const historyChatsCollapsed = ref(true);
const projectCollapsed = ref<Record<string, boolean>>({});

/** 右键菜单(fixed 定位,避免被侧边栏滚动裁剪) */
const menu = ref<null | {
	kind: "session" | "project";
	id: string;
	left: number;
	top: number;
}>(null);
let menuTarget: string | null = null;

const filteredUnassigned = computed<SessionSummary[]>(() => {
	const q = searchQuery.value.trim().toLowerCase();
	if (!q) return props.grouping.unassigned;
	return props.grouping.unassigned.filter(
		(session) => (session.name ?? "").toLowerCase().includes(q) || (session.cwd ?? "").toLowerCase().includes(q),
	);
});

const filteredProjects = computed(() => {
	const q = searchQuery.value.trim().toLowerCase();
	if (!q) return props.grouping.projects;
	const result: SessionGrouping["projects"] = [];
	for (const group of props.grouping.projects) {
		const matched = group.sessions.filter(
			(session) =>
				(session.name ?? "").toLowerCase().includes(q) ||
				(session.cwd ?? "").toLowerCase().includes(q) ||
				group.project.name.toLowerCase().includes(q),
		);
		if (matched.length > 0 || group.project.name.toLowerCase().includes(q)) {
			result.push({ project: group.project, sessions: matched });
		}
	}
	return result;
});

const hasResults = computed(
	() => filteredUnassigned.value.length > 0 || filteredProjects.value.length > 0,
);

function isProjectCollapsed(id: string): boolean {
	return !!projectCollapsed.value[id];
}

function toggleProject(id: string): void {
	projectCollapsed.value[id] = !projectCollapsed.value[id];
}

function openMenu(kind: "session" | "project", id: string, event: MouseEvent): void {
	const button = event.currentTarget as HTMLElement;
	const rect = button.getBoundingClientRect();
	const width = 172;
	let left = rect.right - width;
	if (left < 8) left = 8;
	if (left + width > window.innerWidth - 8) left = window.innerWidth - width - 8;
	menuTarget = id;
		menu.value = { kind, id, left, top: rect.bottom + 4 };
}

function closeMenu(): void {
	menu.value = null;
}

function onWindowClick(event: MouseEvent): void {
	const target = event.target as HTMLElement | null;
	if (target && target.closest(".row-menu")) return;
	closeMenu();
}

function onWindowKeydown(event: KeyboardEvent): void {
	if (event.key === "Escape") closeMenu();
}

function onWindowScroll(): void {
	if (menu.value) closeMenu();
}

watch(menu, (value) => {
	if (value) {
		window.addEventListener("click", onWindowClick);
		window.addEventListener("keydown", onWindowKeydown);
		window.addEventListener("scroll", onWindowScroll, true);
		window.addEventListener("resize", onWindowScroll);
	} else {
		window.removeEventListener("click", onWindowClick);
		window.removeEventListener("keydown", onWindowKeydown);
		window.removeEventListener("scroll", onWindowScroll, true);
		window.removeEventListener("resize", onWindowScroll);
	}
});

onBeforeUnmount(closeMenu);

function formatTime(timestamp: number): string {
	const date = new Date(timestamp);
	const now = new Date();
	const sameDay = date.toDateString() === now.toDateString();
	if (sameDay) {
		return date.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
	}
	return date.toLocaleDateString("zh-CN", { month: "2-digit", day: "2-digit" });
}

function phaseLabel(phase: string): string {
	switch (phase) {
		case "turn":
			return "生成中";
		case "compaction":
			return "压缩中";
		case "retry":
			return "重试中";
		default:
			return "";
	}
}

function runAction(action: (id: string) => void): void {
	const id = menuTarget;
	closeMenu();
	if (id !== null) action(id);
}
</script>

<template>
	<aside class="sidebar" :class="{ open }">
		<div class="brand">
			<img class="brand-logo" src="/logo.png" alt="金融科技 Agent Logo" />
			<div class="brand-text">
				<div class="brand-title">金融科技 Agent</div>
				<span class="brand-sub">Financial Intelligence</span>
			</div>
			<button class="close-sidebar" type="button" aria-label="收起会话列表" @click="emit('close')">
				<AppIcon name="chevron-right" :size="16" />
			</button>
		</div>

		<div class="sidebar-scroll">
			<button class="home-link" type="button" @click="emit('home')">
				<AppIcon name="home" :size="16" />
				<span>回到首页</span>
			</button>

			<div class="search-box">
				<AppIcon name="search" :size="14" />
				<input
					v-model="searchQuery"
					type="search"
					aria-label="搜索会话"
					placeholder="搜索会话…"
					autocomplete="off"
				/>
			</div>

			<nav class="nav" aria-label="主导航">
				<p class="nav-label">导航</p>
				<button
					v-for="item in visibleNav"
					:key="item.view"
					class="nav-item"
					:class="{ active: props.view === item.view }"
					type="button"
					@click="emit('navigate', item.view)"
				>
					<AppIcon :name="item.icon" :size="16" />
					<span>{{ item.label }}</span>
				</button>
			</nav>

			<section class="history" aria-label="历史项目">
				<button
					class="section-toggle"
					type="button"
					:aria-expanded="historyProjectsCollapsed ? 'false' : 'true'"
					@click="historyProjectsCollapsed = !historyProjectsCollapsed"
				>
					<AppIcon
						name="chevron-right"
						:size="14"
						class="section-chevron"
						:class="{ open: !historyProjectsCollapsed }"
					/>
					<span>历史项目</span>
					<span class="section-count">{{ grouping.projects.length }}</span>
				</button>
				<div v-if="!historyProjectsCollapsed" class="section-body">
					<div v-if="grouping.projects.length === 0" class="history-empty">
						<span>还没有项目</span>
						<small>去首页点「创建项目」开始分类管理</small>
					</div>
					<div v-else-if="searchQuery && filteredProjects.length === 0" class="history-empty">
						<span>没有匹配的项目</span>
					</div>
					<div v-else class="project-list">
						<div v-for="group in filteredProjects" :key="group.project.id" class="project-group">
							<button
								class="project-header"
								type="button"
								:aria-expanded="isProjectCollapsed(group.project.id) ? 'false' : 'true'"
								@click="toggleProject(group.project.id)"
							>
								<AppIcon
									name="chevron-right"
									:size="13"
									class="project-chevron"
									:class="{ open: !isProjectCollapsed(group.project.id) }"
								/>
								<span class="project-header-name" :title="group.project.folder || group.project.name">{{
									group.project.name
								}}</span>
								<span class="project-header-count">{{ group.project.sessionIds.length }}</span>
								<span class="project-header-actions">
									<button
										type="button"
										aria-label="在该项目新建对话"
										title="新建对话"
										@click.stop="emit('new-chat-in-project', group.project.id)"
									>
										<AppIcon name="plus" :size="13" />
									</button>
									<button
										type="button"
										aria-label="项目操作"
										title="项目操作"
										@click.stop="openMenu('project', group.project.id, $event)"
									>
										<AppIcon name="more" :size="13" />
									</button>
								</span>
							</button>
							<ul v-if="!isProjectCollapsed(group.project.id) || searchQuery" class="session-scroll">
								<li
									v-for="session in group.sessions"
									:key="session.id"
									:class="{ active: session.id === currentId }"
									@click="emit('select', session.id)"
								>
									<div class="session-row">
										<span class="session-name">{{ session.name || session.id.slice(0, 8) }}</span>
										<button
											class="session-more"
											type="button"
											aria-label="会话操作"
											title="会话操作"
											@click.stop="openMenu('session', session.id, $event)"
										>
											<AppIcon name="more" :size="14" />
										</button>
									</div>
									<div class="session-meta">
										<span v-if="phaseLabel(session.phase)" class="phase-badge">{{ phaseLabel(session.phase) }}</span>
										<span class="session-time">{{ formatTime(session.updatedAt) }}</span>
									</div>
								</li>
								<li v-if="group.sessions.length === 0" class="session-empty">该项目还没有对话</li>
							</ul>
						</div>
					</div>
				</div>
			</section>

			<section class="history" aria-label="历史对话">
				<button
					class="section-toggle"
					type="button"
					:aria-expanded="historyChatsCollapsed ? 'false' : 'true'"
					@click="historyChatsCollapsed = !historyChatsCollapsed"
				>
					<AppIcon
						name="chevron-right"
						:size="14"
						class="section-chevron"
						:class="{ open: !historyChatsCollapsed }"
					/>
					<span>历史对话</span>
					<span class="section-count">{{ grouping.unassigned.length }}</span>
				</button>
				<div v-if="!historyChatsCollapsed" class="section-body">
					<div v-if="grouping.unassigned.length === 0" class="history-empty">
						<span>没有无项目会话</span>
						<small>首页直接输入的对话会出现在这里</small>
					</div>
					<div v-else-if="searchQuery && filteredUnassigned.length === 0" class="history-empty">
						<span>没有匹配的会话</span>
					</div>
					<ul v-else class="session-scroll">
						<li
							v-for="session in filteredUnassigned"
							:key="session.id"
							:class="{ active: session.id === currentId }"
							@click="emit('select', session.id)"
						>
							<div class="session-row">
								<span class="session-name">{{ session.name || session.id.slice(0, 8) }}</span>
								<button
									class="session-more"
									type="button"
									aria-label="会话操作"
									title="会话操作"
									@click.stop="openMenu('session', session.id, $event)"
								>
									<AppIcon name="more" :size="14" />
								</button>
							</div>
							<div class="session-meta">
								<span v-if="phaseLabel(session.phase)" class="phase-badge">{{ phaseLabel(session.phase) }}</span>
								<span class="session-time">{{ formatTime(session.updatedAt) }}</span>
							</div>
						</li>
					</ul>
				</div>
			</section>

			<div v-if="searchQuery && !hasResults" class="history-empty search-none">
				<span>没有匹配的会话</span>
			</div>
		</div>

		<footer class="sidebar-footer">
			<div class="footer-actions">
				<button class="footer-btn" type="button" @click="emit('open-tools')">
					<AppIcon name="wrench" :size="16" />
					<span>工具</span>
				</button>
				<button class="footer-btn" type="button" @click="emit('open-settings')">
					<AppIcon name="sliders" :size="16" />
					<span>设置</span>
				</button>
				<ThemeToggle class="footer-theme" />
			</div>
			<div class="footer-status">
				<ConnectionBadge :connected="connected" :has-session="!!currentId" />
			</div>
		</footer>

		<div v-if="menu" class="row-menu" :style="{ left: menu.left + 'px', top: menu.top + 'px' }" role="menu">
			<template v-if="menu.kind === 'session'">
				<button type="button" role="menuitem" @click="runAction((id) => emit('rename', id))">
					<AppIcon name="pen" :size="14" />
					<span>重命名</span>
				</button>
				<button type="button" role="menuitem" @click="runAction((id) => emit('move', id))">
					<AppIcon name="folder" :size="14" />
					<span>移动到项目</span>
				</button>
				<button type="button" role="menuitem" @click="runAction((id) => emit('export', id))">
					<AppIcon name="file-text" :size="14" />
					<span>导出</span>
				</button>
				<div class="row-menu-divider" />
				<button type="button" role="menuitem" class="danger" @click="runAction((id) => emit('remove', id))">
					<AppIcon name="x" :size="14" />
					<span>删除</span>
				</button>
			</template>
			<template v-else>
				<button type="button" role="menuitem" @click="runAction((id) => emit('rename-project', id))">
					<AppIcon name="pen" :size="14" />
					<span>重命名项目</span>
				</button>
				<div class="row-menu-divider" />
				<button type="button" role="menuitem" class="danger" @click="runAction((id) => emit('delete-project', id))">
					<AppIcon name="x" :size="14" />
					<span>删除项目</span>
				</button>
			</template>
		</div>
	</aside>
</template>

<style scoped>
.sidebar {
	display: flex;
	flex-direction: column;
	width: 240px;
	min-width: 240px;
	background: var(--bg-sidebar);
	border-right: 1px solid var(--border);
	height: 100%;
}

/* 品牌 */
.brand {
	display: flex;
	align-items: center;
	gap: 10px;
	padding: 20px 16px 16px;
}

.brand-logo {
	width: 32px;
	height: 32px;
	flex-shrink: 0;
	object-fit: contain;
}

.brand-text {
	min-width: 0;
	display: flex;
	flex-direction: column;
}

.brand-title {
	font-size: 16px;
	font-weight: 600;
	letter-spacing: -0.01em;
	line-height: 1.3;
	color: var(--text-primary);
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.brand-sub {
	font-size: 11px;
	font-weight: 400;
	color: var(--text-muted);
	margin-top: 1px;
	letter-spacing: 0.015em;
}

.close-sidebar {
	display: none;
	margin-left: auto;
	background: none;
	border: none;
	color: var(--text-muted);
	cursor: pointer;
	padding: 4px;
	border-radius: var(--radius-xs);
}

.close-sidebar:hover {
	color: var(--text-primary);
	background: var(--bg-hover);
}

/* 滚动区 */
.sidebar-scroll {
	flex: 1;
	overflow-y: auto;
	padding: 4px 12px 12px;
}

/* 回到首页(第一) */
.home-link {
	display: flex;
	align-items: center;
	justify-content: flex-start;
	gap: 10px;
	width: 100%;
	height: 40px;
	padding: 0 12px;
	border: none;
	border-radius: 7px;
	background: transparent;
	color: #304a5e;
	font-size: 14px;
	font-weight: 500;
	cursor: pointer;
	transition:
		background 150ms ease,
		color 150ms ease;
}

.home-link svg {
	color: #6dafe0;
	flex-shrink: 0;
}

.home-link:hover {
	background: #e8f3fa;
}

.home-link:active {
	background: #ddeffa;
}

/* 搜索 */
.search-box {
	display: flex;
	align-items: center;
	gap: 8px;
	height: 34px;
	padding: 0 10px;
	margin: 8px 0 4px;
	border: 1px solid var(--border);
	border-radius: 7px;
	background: var(--bg-surface);
	color: var(--text-muted);
	transition: border-color 140ms ease;
}

.search-box:focus-within {
	border-color: var(--accent);
}

.search-box svg {
	flex-shrink: 0;
}

.search-box input {
	flex: 1;
	min-width: 0;
	background: transparent;
	border: none;
	outline: none;
	color: var(--text-primary);
	font-size: 13px;
}

.search-box input::placeholder {
	color: var(--text-disabled);
}

.search-box input::-webkit-search-cancel-button {
	-webkit-appearance: none;
}

/* 导航骨架 */
.nav {
	margin-top: 12px;
	display: flex;
	flex-direction: column;
	gap: 2px;
}

.nav-item {
	display: flex;
	align-items: center;
	gap: 10px;
	width: 100%;
	height: 38px;
	padding: 0 12px;
	border: none;
	border-radius: var(--radius-sm);
	background: transparent;
	color: #647c90;
	font-size: 13px;
	font-weight: 500;
	cursor: pointer;
}

	.nav-item:hover {
		background: var(--bg-hover);
		color: #304a5e;
	}

	.nav-item.active {
		background: #eaf4fa;
		color: #304a5e;
	}

	.nav-item.active svg {
		color: var(--accent-strong);
	}

.nav-item svg {
	color: #86b3d3;
	flex-shrink: 0;
}

.nav-badge {
	margin-left: auto;
	font-size: 10px;
	font-weight: 400;
	color: #8ca3b5;
	background: #eaf3f9;
	border-radius: 4px;
	padding: 2px 5px;
	white-space: nowrap;
}

/* 历史分区 */
.history {
	margin-top: 16px;
}

.section-toggle {
	display: flex;
	align-items: center;
	gap: 6px;
	width: 100%;
	height: 30px;
	padding: 0 8px;
	border: none;
	border-radius: var(--radius-sm);
	background: transparent;
	color: #647c90;
	font-size: 12px;
	font-weight: 500;
	cursor: pointer;
	text-align: left;
	transition:
		background 150ms ease,
		color 150ms ease;
}

.section-toggle:hover {
	background: var(--bg-hover);
	color: #304a5e;
}

.section-chevron {
	flex-shrink: 0;
	color: #8ca3b5;
	transition: transform 160ms ease;
}

.section-chevron.open {
	transform: rotate(90deg);
}

.section-count {
	flex-shrink: 0;
	font-size: 10px;
	font-weight: 500;
	color: #8ca3b5;
	background: #eaf3f9;
	border-radius: 4px;
	padding: 1px 6px;
}

.section-body {
	padding-top: 4px;
}

.history-empty {
	display: flex;
	flex-direction: column;
	gap: 3px;
	padding: 4px 12px 8px;
	font-size: 13px;
	font-weight: 500;
	color: #667d90;
}

.history-empty small {
	font-size: 11px;
	font-weight: 400;
	color: #98aabc;
	line-height: 1.5;
}

.search-none {
	margin-top: 12px;
}

/* 项目分组 */
.project-list {
	display: flex;
	flex-direction: column;
	gap: 2px;
}

.project-group {
	display: flex;
	flex-direction: column;
}

.project-header {
	display: flex;
	align-items: center;
	gap: 6px;
	width: 100%;
	height: 30px;
	padding: 0 8px;
	margin: 2px 0;
	border: none;
	border-radius: var(--radius-sm);
	background: transparent;
	color: #647c90;
	font-size: 12px;
	font-weight: 500;
	cursor: pointer;
	text-align: left;
	transition:
		background 150ms ease,
		color 150ms ease;
}

.project-header:hover {
	background: var(--bg-hover);
	color: #304a5e;
}

.project-chevron {
	flex-shrink: 0;
	color: #8ca3b5;
	transition: transform 160ms ease;
}

.project-chevron.open {
	transform: rotate(90deg);
}

.project-header-name {
	flex: 1;
	min-width: 0;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.project-header-count {
	flex-shrink: 0;
	font-size: 10px;
	font-weight: 500;
	color: #8ca3b5;
	background: #eaf3f9;
	border-radius: 4px;
	padding: 1px 6px;
}

.project-header-actions {
	display: flex;
	align-items: center;
	gap: 2px;
	flex-shrink: 0;
	opacity: 0;
	transition: opacity 120ms ease;
}

.project-header:hover .project-header-actions {
	opacity: 1;
}

.project-header-actions button {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	width: 22px;
	height: 22px;
	background: transparent;
	border: none;
	border-radius: var(--radius-xs);
	color: var(--text-muted);
	cursor: pointer;
}

.project-header-actions button:hover {
	color: var(--text-primary);
	background: var(--bg-active);
}

/* 会话行 */
.session-scroll {
	list-style: none;
	margin: 0 0 4px;
	padding: 0;
}

.session-scroll li {
	position: relative;
	padding: 8px 12px;
	border-radius: var(--radius-sm);
	cursor: pointer;
	margin-bottom: 1px;
	transition: background 150ms ease;
}

.session-scroll li:hover {
	background: var(--bg-hover);
}

.session-scroll li.active {
	background: #eaf4fa;
}

.session-scroll li.active::before {
	content: "";
	position: absolute;
	left: -12px;
	top: 6px;
	bottom: 6px;
	width: 2px;
	border-radius: 2px;
	background: var(--accent);
}

.session-empty {
	display: block;
	padding: 6px 12px;
	font-size: 11px;
	font-weight: 400;
	color: #98aabc;
	cursor: default;
}

.session-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 8px;
}

.session-name {
	font-size: 13px;
	font-weight: 500;
	color: #344d61;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.session-more {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	width: 24px;
	height: 24px;
	flex-shrink: 0;
	background: transparent;
	border: none;
	border-radius: var(--radius-xs);
	color: var(--text-muted);
	cursor: pointer;
	opacity: 0;
	transition:
		opacity 120ms ease,
		background 120ms ease,
		color 120ms ease;
}

.session-scroll li:hover .session-more,
.session-scroll li.active .session-more {
	opacity: 1;
}

.session-more:hover {
	color: var(--text-primary);
	background: var(--bg-active);
}

.session-meta {
	display: flex;
	align-items: center;
	gap: 8px;
	margin-top: 2px;
}

.phase-badge {
	font-size: 11px;
	color: var(--accent-strong);
}

.session-time {
	font-size: 11px;
	font-weight: 400;
	color: #9aaaba;
}

/* 右键菜单 */
.row-menu {
	position: fixed;
	z-index: 200;
	min-width: 172px;
	padding: 6px;
	background: var(--bg-surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-md);
	box-shadow: var(--shadow-float);
	display: flex;
	flex-direction: column;
	gap: 1px;
}

.row-menu button {
	display: flex;
	align-items: center;
	gap: 9px;
	height: 32px;
	padding: 0 10px;
	border: none;
	border-radius: var(--radius-sm);
	background: transparent;
	color: var(--text-primary);
	font-size: 13px;
	font-weight: 400;
	text-align: left;
	cursor: pointer;
	transition:
		background 120ms ease,
		color 120ms ease;
}

.row-menu button svg {
	color: var(--text-muted);
	flex-shrink: 0;
}

.row-menu button:hover {
	background: var(--bg-hover);
	color: var(--text-primary);
}

.row-menu button:hover svg {
	color: var(--accent-strong);
}

.row-menu button.danger {
	color: var(--danger);
}

.row-menu button.danger svg {
	color: var(--danger);
}

.row-menu button.danger:hover {
	background: var(--danger-bg);
	color: var(--danger-strong);
}

.row-menu-divider {
	height: 1px;
	background: var(--border);
	margin: 4px 2px;
}

/* 底部 */
.sidebar-footer {
	padding: 10px 12px 14px;
	border-top: 1px solid var(--border);
	display: flex;
	flex-direction: column;
	gap: 10px;
}

.footer-actions {
	display: flex;
	align-items: center;
	gap: 2px;
}

.footer-btn {
	display: inline-flex;
	align-items: center;
	gap: 8px;
	height: 34px;
	padding: 0 10px;
	border: none;
	border-radius: var(--radius-sm);
	background: transparent;
	color: #526c81;
	font-size: 13px;
	font-weight: 500;
	cursor: pointer;
	transition:
		background 150ms ease,
		color 150ms ease;
}

.footer-btn:hover {
	background: var(--bg-hover);
	color: var(--text-primary);
}

.footer-theme {
	margin-left: auto;
}

.footer-status {
	padding: 0 4px;
}

@media (max-width: 860px) {
	.sidebar {
		position: fixed;
		left: 0;
		top: 0;
		bottom: 0;
		z-index: var(--z-sidebar);
		transform: translateX(-100%);
		transition: transform 200ms ease;
		box-shadow: var(--shadow-float);
	}

	.sidebar.open {
		transform: translateX(0);
	}

	.close-sidebar {
		display: inline-flex;
	}
}

@media (min-width: 861px) and (max-width: 1200px) {
	.sidebar {
		width: 220px;
		min-width: 220px;
	}
}

/* 深色导航骨架：让历史与功能入口在浅色研究画布上保持清晰层级 */
.sidebar :is(.brand-title, .home-link, .nav-item, .section-toggle, .project-header, .session-name, .footer-btn) { color: #e7edf7; }
.sidebar :is(.brand-sub, .section-count, .project-header-count, .session-time, .history-empty, .search-box) { color: #9aa9c0; }
.sidebar .brand-sub { color: #91a3bf; }
.sidebar .home-link:hover, .sidebar .nav-item:hover, .sidebar .section-toggle:hover, .sidebar .project-header:hover, .sidebar .session-scroll li:hover { background: rgba(255, 255, 255, 0.08); }
.sidebar .nav-item.active, .sidebar .session-scroll li.active { background: rgba(74, 137, 255, 0.2); }
.sidebar .home-link svg, .sidebar .nav-item svg, .sidebar .section-chevron, .sidebar .project-chevron { color: #76a6ff; }
.sidebar .search-box { background: rgba(255, 255, 255, 0.06); border-color: rgba(255, 255, 255, 0.14); }
.sidebar .search-box input { color: #eef3fb; }
.sidebar .search-box input::placeholder { color: #7f8da5; }
.sidebar .nav-badge, .sidebar .section-count, .sidebar .project-header-count { background: rgba(255, 255, 255, 0.09); color: #a8b7ce; }
.sidebar .sidebar-footer { border-top-color: rgba(255, 255, 255, 0.12); }

/* Research Cockpit command rail overrides */
.sidebar {
	width: 280px;
	min-width: 280px;
	background: linear-gradient(180deg, var(--shell-surface), var(--shell-bg));
	border-right-color: var(--shell-border);
}

.sidebar::before {
	content: "";
	position: absolute;
	top: 0;
	left: 0;
	width: 2px;
	height: 100%;
	background: linear-gradient(180deg, transparent, color-mix(in srgb, var(--accent) 44%, transparent), transparent);
	pointer-events: none;
}

.sidebar .brand { padding: 18px 16px; border-bottom: 1px solid var(--shell-border); }
.sidebar .brand-title { color: var(--shell-text); font-size: 17px; letter-spacing: -0.03em; }
.sidebar .brand-sub { color: var(--shell-text-secondary); }
.sidebar .sidebar-scroll { padding: 14px 12px 16px; }
.sidebar .home-link {
	height: 42px;
	background: color-mix(in srgb, var(--accent) 14%, transparent);
	color: var(--shell-text);
	font-weight: 650;
	border: 1px solid color-mix(in srgb, var(--accent) 24%, transparent);
}

.sidebar .home-link:hover { background: color-mix(in srgb, var(--accent) 22%, transparent); }
.sidebar .search-box { height: 36px; background: rgba(255, 255, 255, 0.055); border: 1px solid var(--shell-border); color: var(--shell-text-secondary); }
.sidebar .nav-label { color: #7c8ea9; font-size: 10px; font-weight: 750; letter-spacing: 0.14em; text-transform: uppercase; }
.sidebar .nav-item { height: 38px; color: #c8d4e6; border-radius: var(--radius-sm); font-weight: 550; }
.sidebar .nav-item:hover { background: rgba(255, 255, 255, 0.07); color: var(--shell-text); }
.sidebar .nav-item.active { background: color-mix(in srgb, var(--accent) 24%, transparent); color: #fff; box-shadow: inset 2px 0 0 var(--shell-accent); }
.sidebar .section-toggle { height: 36px; color: #b7c4d8; font-size: 11px; font-weight: 750; letter-spacing: 0.08em; text-transform: uppercase; }
.sidebar .project-header { min-height: 38px; color: #d8e2f0; font-weight: 600; }
.sidebar .session-scroll li { border: 1px solid transparent; border-radius: var(--radius-sm); }
.sidebar .session-scroll li.active { border-color: color-mix(in srgb, var(--accent) 38%, transparent); background: color-mix(in srgb, var(--accent) 20%, transparent); }
.sidebar .session-name { color: #e4ebf8; font-weight: 550; }
.sidebar .session-time { color: #8b9cb5; }
.sidebar .history-empty { color: #8b9cb5; }
.sidebar .sidebar-footer { background: rgba(255, 255, 255, 0.025); border-top: 1px solid var(--shell-border); }
.sidebar .footer-btn { color: #b7c4d8; font-weight: 600; }
.sidebar .footer-btn:hover { background: rgba(255, 255, 255, 0.07); color: var(--shell-text); }

@media (min-width: 861px) and (max-width: 1200px) {
	.sidebar { width: 248px; min-width: 248px; }
}
</style>
