<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import AdminView from "./components/AdminView.vue";
import AgentPlaza from "./components/AgentPlaza.vue";
import AppIcon from "./components/AppIcon.vue";
import ChatView from "./components/ChatView.vue";
import DatasetsView from "./components/DatasetsView.vue";
import EmptyState from "./components/EmptyState.vue";
import GroupsView from "./components/GroupsView.vue";
import MessageItem from "./components/MessageItem.vue";
import ModelSettingsDrawer from "./components/ModelSettingsDrawer.vue";
import MonitorView from "./components/MonitorView.vue";
import ProjectPicker from "./components/ProjectPicker.vue";
import ProjectWorkbench from "./components/ProjectWorkbench.vue";
import PromptBox from "./components/PromptBox.vue";
import SessionList from "./components/SessionList.vue";
import SkillsView from "./components/SkillsView.vue";
import ToolsDrawer from "./components/ToolsDrawer.vue";
import { groupSessions, type ProjectWithSessions } from "./state/projects.ts";
import {
	abortCurrent,
	bootstrap,
	createProject,
	createSession,
	deleteProject,
	deleteSession,
	exportSession,
	login,
	logout,
	moveSessionProject,
	openSession,
	openWorkbench,
	closeWorkbench,
	refreshContext,
	refreshSessions,
	renameSession,
	sendPrompt,
	sendSteer,
	store,
	switchModel,
	transcriptItems,
	updateProject,
} from "./state/session-store.ts";
import { sessionProjectId } from "./state/projects.ts";

const items = computed(() => transcriptItems());
const currentModel = computed(() => store.snapshot?.model ?? null);
const sidebarOpen = ref(false);
const settingsOpen = ref(false);
const toolsOpen = ref(false);
const conversationDraft = ref("");
const view = ref<"home" | "chat" | "workbench" | "groups" | "skills" | "datasets" | "agents" | "admin" | "monitor">("home");

const pickerOpen = ref(false);
const pickerMode = ref<"create" | "move">("create");
const moveSessionId = ref<string | null>(null);

// 登录
const loginUsername = ref("");
const loginPassword = ref("");
const loginBusy = ref(false);

async function handleLogin(): Promise<void> {
	loginBusy.value = true;
	try {
		await login(loginUsername.value, loginPassword.value);
		loginPassword.value = "";
	} catch {
		// 错误已写入 store.error
	} finally {
		loginBusy.value = false;
	}
}

function handleLogout(): void {
	void logout();
}

/** 当前会话所属项目(用于展示会话上下文)。 */
const contextProject = computed(() =>
	store.contextProjectId ? (store.projects.find((item) => item.id === store.contextProjectId) ?? null) : null,
);

const pageTitle = computed(() => {
	switch (view.value) {
		case "chat":
			return store.snapshot?.name || "AI Research Workbench";
		case "groups":
			return "课程组";
		case "skills":
			return "技能库";
		case "datasets":
			return "数据集库";
		case "agents":
			return "智能体广场";
		case "admin":
			return "管理端";
		case "monitor":
			return "会话监控";
		default:
			return "Research Cockpit";
	}
});

/** 会话归属分组:无项目 + 按项目分组 */
const grouping = computed(() => groupSessions(store.sessions, store.projects));

/** 首页最近项目(最多 4 个,已按项目 updatedAt 倒序) */
const recentProjects = computed<ProjectWithSessions[]>(() => grouping.value.projects.slice(0, 4));

/** 首页(对话框 + 最近项目) */
const showHome = computed(() => view.value === "home");

/** 会话是否正在生成(供插话/停止按钮使用;store.busy 仅覆盖请求期间,phase 覆盖流式生成期) */
const isGenerating = computed(() => {
	if (store.busy) return true;
	const phase = store.snapshot?.phase;
	return phase === "turn" || phase === "compaction" || phase === "retry";
});

const phaseLabel = computed(() => {
	if (!store.snapshot) return "";
	switch (store.snapshot.phase) {
		case "turn":
			return "生成中";
		case "compaction":
			return "压缩中";
		case "retry":
			return "重试中";
		default:
			return store.busy ? "处理中" : "";
	}
});

const lastUsageTokens = computed(() => {
	const last = items.value.at(-1);
	return last && last.role !== "user" ? last.usage?.totalTokens : null;
});

onMounted(() => {
	void bootstrap();
});

watch(
	() => store.currentId,
	() => {
		conversationDraft.value = "";
		void refreshSessions();
		const projectId = store.currentId ? sessionProjectId(store.currentId, store.projects) : null;
		void refreshContext(projectId);
	},
);

/** 首页对话框发送:总是新建无项目会话,立即进入对话页并异步发送 */
async function handleSendFromHome(text: string): Promise<void> {
	await createSession();
	view.value = "chat";
	if (store.currentId) {
		void sendPrompt(text);
	}
}

async function handleSend(text: string): Promise<void> {
	if (!store.currentId) {
		await createSession();
	}
	if (store.currentId) {
		await sendPrompt(text);
	}
}

async function handleSteer(text: string): Promise<void> {
	if (!store.currentId) return;
	await sendSteer(text);
}

/** 打开项目工作台:确保该项目有会话,进入三栏 */
async function handleOpenProject(projectId: string): Promise<void> {
	const currentProject = store.currentId ? sessionProjectId(store.currentId, store.projects) : null;
	if (currentProject !== projectId) {
		await createSession({ projectId });
	}
	if (!store.currentId) return;
	await openSession(store.currentId);
	await refreshContext(projectId);
	await openWorkbench(projectId);
	view.value = "workbench";
}

/** 打开历史会话:项目会话进工作台,通用会话进对话,空会话回首页 */
async function handleOpen(sessionId: string): Promise<void> {
	await openSession(sessionId);
	const projectId = sessionProjectId(sessionId, store.projects);
	if (projectId) {
		await refreshContext(projectId);
		await openWorkbench(projectId);
		view.value = "workbench";
	} else {
		view.value = items.value.length > 0 ? "chat" : "home";
	}
}

function handleCloseWorkbench(): void {
	closeWorkbench();
	view.value = "home";
}

/** 当前工作台项目 */
const workbenchProject = computed(() =>
	store.workbench.projectId ? (store.projects.find((item) => item.id === store.workbench.projectId) ?? null) : null,
);

/** 在指定项目下新建对话(folder 作为 cwd 由后端解析) */
async function handleNewChatInProject(projectId: string): Promise<void> {
	await handleOpenProject(projectId);
}

function openCreateProject(): void {
	pickerMode.value = "create";
	pickerOpen.value = true;
}

function openMoveSession(sessionId: string): void {
	pickerMode.value = "move";
	moveSessionId.value = sessionId;
	pickerOpen.value = true;
}

async function handleCreateProject(input: {
	name: string;
	folder?: string;
	scope?: { type: "private" } | { type: "group"; groupId: string };
}): Promise<void> {
	pickerOpen.value = false;
	await createProject(input);
}

async function handleMoveSession(projectId: string | null): Promise<void> {
	pickerOpen.value = false;
	const sessionId = moveSessionId.value;
	moveSessionId.value = null;
	if (sessionId) {
		await moveSessionProject(sessionId, projectId);
	}
}

async function handleRenameProject(projectId: string): Promise<void> {
	const project = store.projects.find((item) => item.id === projectId);
	const name = window.prompt("请输入新的项目名称", project?.name ?? "");
	if (!name || !name.trim()) return;
	await updateProject(projectId, { name: name.trim() });
}

async function handleDeleteProject(projectId: string): Promise<void> {
	if (!window.confirm("确定删除该项目？项目内会话将变为无项目，保留在「历史对话」中。")) return;
	await deleteProject(projectId);
}

async function handleRemoveSession(sessionId: string): Promise<void> {
	if (!window.confirm("确定删除该会话？此操作不可恢复。")) return;
	const wasCurrent = store.currentId === sessionId;
	await deleteSession(sessionId);
	if (wasCurrent) view.value = "home";
}

async function handleRename(sessionId: string): Promise<void> {
	const name = window.prompt("请输入新的会话名称");
	if (!name || !name.trim()) return;
	try {
		await renameSession(sessionId, name.trim());
	} catch {
		// 错误已写入 store.error
	}
}

async function handleExport(sessionId: string): Promise<void> {
	try {
		const content = await exportSession(sessionId);
		const blob = new Blob([content], { type: "text/plain;charset=utf-8" });
		const url = URL.createObjectURL(blob);
		const anchor = document.createElement("a");
		anchor.href = url;
		anchor.download = `tfa-session-${sessionId.slice(0, 8)}.jsonl`;
		anchor.click();
		URL.revokeObjectURL(url);
	} catch {
		// 错误已写入 store.error
	}
}
</script>

<template>
	<div v-if="store.user" class="app">
		<a class="skip-link" href="#main-content">跳到主要内容</a>
		<div v-if="sidebarOpen" class="sidebar-backdrop" @click="sidebarOpen = false" />
		<SessionList
			:grouping="grouping"
			:current-id="store.currentId"
			:open="sidebarOpen"
			:connected="store.connected"
			:user="store.user"
			:view="view"
			@select="handleOpen($event)"
			@navigate="view = $event"
			@home="view = 'home'"
			@remove="handleRemoveSession($event)"
			@rename="handleRename($event)"
			@move="openMoveSession($event)"
			@export="handleExport($event)"
			@close="sidebarOpen = false"
			@open-tools="toolsOpen = true"
			@open-settings="settingsOpen = true"
			@new-chat-in-project="handleNewChatInProject($event)"
			@rename-project="handleRenameProject($event)"
			@delete-project="handleDeleteProject($event)"
		/>
		<main id="main-content" class="workspace" tabindex="-1">
			<header class="global-bar" role="banner">
				<button class="menu-button" type="button" aria-label="打开会话列表" @click="sidebarOpen = true">
					<AppIcon name="message" :size="18" />
				</button>
				<div class="global-bar-main">
					<span class="page-kicker">Financial Research</span>
					<h1 class="page-title">{{ pageTitle }}</h1>
				</div>
				<div class="global-bar-status">
					<span class="status-chip" :class="store.connected ? 'online' : 'offline'">
						<i aria-hidden="true" />
						{{ store.connected ? "已连接" : "离线" }}
					</span>
					<div class="user-chip">
						<span class="user-avatar" aria-hidden="true">{{ store.user?.username.slice(0, 1).toUpperCase() }}</span>
						<span class="user-detail">
							<strong>{{ store.user?.username }}</strong>
							<small>{{ store.user?.role }}</small>
						</span>
					</div>
					<button class="logout-button" type="button" @click="handleLogout">退出</button>
				</div>
			</header>

			<div v-if="store.error" class="error-banner" role="alert">
				<span class="error-text">{{ store.error }}</span>
				<button class="error-dismiss" type="button" @click="store.error = null">关闭</button>
			</div>

			<template v-if="showHome">
				<EmptyState
					:projects="recentProjects"
					:total-projects="store.projects.length"
					:sessions="grouping.unassigned.slice(0, 3)"
					:total-sessions="store.sessions.length"
					:skill-count="store.skills.length"
					:dataset-count="store.datasets.length"
					:user="store.user"
					:connected="store.connected"
					:models="store.models"
					:current-model="currentModel"
					@send="handleSendFromHome($event)"
					@open-session="handleOpen($event)"
					@new-chat-in-project="handleNewChatInProject($event)"
					@create-project="openCreateProject()"
				/>
			</template>
			<template v-else-if="view === 'chat'">
				<div class="research-layout">
					<ChatView :items="items" :busy="store.busy">
						<template #header>
							<div class="conversation-heading">
								<span class="conversation-kicker">Research Transcript</span>
								<div class="conversation-title">
									<span>{{ store.snapshot?.name || "当前会话" }}</span>
									<span v-if="phaseLabel" class="conversation-phase">{{ phaseLabel }}</span>
								</div>
							</div>
						</template>
						<template #message="{ item }">
							<MessageItem :item="item" />
						</template>
					</ChatView>
					<aside class="context-panel" aria-label="研究上下文">
						<section class="context-card">
							<h2>项目上下文</h2>
							<p v-if="contextProject" class="context-primary">{{ contextProject.name }}</p>
							<p v-else class="context-muted">无项目独立研究</p>
							<small v-if="contextProject?.folder" class="context-path mono">{{ contextProject.folder }}</small>
						</section>
						<section class="context-card">
							<h2>可用技能</h2>
							<template v-if="store.sessionContext?.skills.length">
								<ul class="context-list">
									<li v-for="skill in store.sessionContext.skills.slice(0, 5)" :key="skill.id">
										<strong>{{ skill.name }}</strong>
										<span>{{ skill.visibility }}</span>
									</li>
								</ul>
							</template>
							<p v-else class="context-muted">当前项目无注入技能</p>
						</section>
						<section class="context-card">
							<h2>可用数据集</h2>
							<template v-if="store.sessionContext?.datasets.length">
								<ul class="context-list">
									<li v-for="dataset in store.sessionContext.datasets.slice(0, 5)" :key="dataset.id">
										<strong>{{ dataset.name }}</strong>
										<span>{{ dataset.visibility }}</span>
									</li>
								</ul>
							</template>
							<p v-else class="context-muted">当前项目无注入数据集</p>
						</section>
						<section class="context-card">
							<h2>运行状态</h2>
							<div class="status-grid">
								<div>
									<small>阶段</small>
									<strong>{{ phaseLabel || "空闲" }}</strong>
								</div>
								<div>
									<small>消息</small>
									<strong>{{ items.length }}</strong>
								</div>
								<div>
									<small>Tokens</small>
									<strong>{{ lastUsageTokens ?? "—" }}</strong>
								</div>
							</div>
						</section>
					</aside>
				</div>
				<div class="composer-dock">
					<PromptBox
						:busy="isGenerating"
						:models="store.models"
						:current="currentModel"
						v-model="conversationDraft"
						@send="handleSend($event)"
						@steer="handleSteer($event)"
						@abort="abortCurrent()"
						@model-change="switchModel($event.provider, $event.id)"
						@open-tools="toolsOpen = true"
					/>
				</div>
			</template>
			<template v-else>
				<ProjectWorkbench v-if="view === 'workbench'" :project="workbenchProject" @close="handleCloseWorkbench" />
				<GroupsView v-else-if="view === 'groups'" />
				<SkillsView v-else-if="view === 'skills'" />
				<DatasetsView v-else-if="view === 'datasets'" />
				<AgentPlaza v-else-if="view === 'agents'" />
				<AdminView v-else-if="view === 'admin'" />
				<MonitorView v-else-if="view === 'monitor'" />
			</template>
		</main>

		<ProjectPicker
			:open="pickerOpen"
			:mode="pickerMode"
			:projects="store.projects"
			:groups="store.groups"
			@close="pickerOpen = false"
			@create="handleCreateProject($event)"
			@move="handleMoveSession($event)"
		/>
		<ModelSettingsDrawer :open="settingsOpen" @close="settingsOpen = false" />
		<ToolsDrawer :open="toolsOpen" @close="toolsOpen = false" />
	</div>

	<div v-else class="login-screen">
		<form class="login-card" @submit.prevent="handleLogin">
			<h1 class="login-title">TFA 科研 Agent 平台</h1>
			<p class="login-desc">登录后管理项目、课程组、技能与数据集</p>
			<p v-if="store.error" class="login-error" role="alert">{{ store.error }}</p>
			<label class="login-label" for="login-username">用户名</label>
			<input id="login-username" v-model="loginUsername" type="text" autocomplete="username" spellcheck="false" />
			<label class="login-label" for="login-password">密码</label>
			<input id="login-password" v-model="loginPassword" type="password" autocomplete="current-password" />
			<button class="btn btn-primary login-submit" type="submit" :disabled="loginBusy || !loginUsername.trim() || !loginPassword">
				{{ loginBusy ? "登录中…" : "登录" }}
			</button>
		</form>
	</div>
</template>

<style scoped>
.login-screen {
	height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 24px;
	background: var(--bg-app);
}

.login-card {
	width: 100%;
	max-width: 340px;
	display: flex;
	flex-direction: column;
	gap: 10px;
	padding: 28px 26px;
	background: var(--bg-card);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
}

.login-title {
	margin: 0;
	font-size: 20px;
	font-weight: 700;
	color: var(--text-primary);
}

.login-desc {
	margin: 0 0 8px;
	font-size: 12px;
	color: var(--text-muted);
}

.login-error {
	margin: 0;
	font-size: 12px;
	color: var(--danger, #e06c75);
}

.login-label {
	font-size: 12px;
	font-weight: 500;
	color: var(--text-muted);
}

.login-card input {
	width: 100%;
	background: var(--bg-input);
	color: var(--text-primary);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	padding: 9px 11px;
	font-size: 13px;
	font-family: var(--font-mono);
}

.login-card input:focus {
	outline: none;
	border-color: var(--accent);
}

.login-submit {
	margin-top: 8px;
	height: 38px;
}

.logout-button {
	display: inline-flex;
	align-items: center;
	gap: 6px;
	height: 30px;
	padding: 0 10px;
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
	background: var(--bg-card);
	color: var(--text-muted);
	font-size: 12px;
	cursor: pointer;
}

.logout-button:hover {
	color: var(--text-primary);
	border-color: var(--border-strong);
}

.logout-text {
	max-width: 120px;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.session-context-bar {
	display: flex;
	align-items: center;
	gap: 12px;
	padding: 8px 16px;
	border-bottom: 1px solid var(--border);
	background: var(--bg-card);
	font-size: 12px;
	color: var(--text-muted);
}

.context-project {
	font-weight: 600;
	color: var(--text-primary);
}

.context-empty {
	color: var(--text-disabled);
}
.skip-link {
	position: fixed;
	top: -40px;
	left: 12px;
	z-index: 100;
	background: var(--accent);
	color: #ffffff;
	padding: 8px 14px;
	border-radius: 0 0 var(--radius-md) var(--radius-md);
	font-size: 13px;
	font-weight: 600;
	text-decoration: none;
	transition: top 150ms ease;
}

.skip-link:focus {
	top: 0;
}

.app {
	display: flex;
	height: 100dvh;
	position: relative;
	background: var(--bg-app);
}

.workspace {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-width: 0;
	position: relative;
}

.menu-button {
	display: none;
	position: absolute;
	top: 14px;
	left: 14px;
	z-index: 5;
	width: 38px;
	height: 38px;
	align-items: center;
	justify-content: center;
	background: var(--bg-surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-md);
	color: var(--text-secondary);
	cursor: pointer;
	box-shadow: var(--shadow-float);
}

.sidebar-backdrop {
	display: none;
}

.error-banner {
	background: var(--danger-bg);
	color: var(--danger);
	font-size: 13px;
	padding: 10px 24px;
	border-bottom: 1px solid var(--danger);
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 12px;
}

.error-text {
	overflow-wrap: anywhere;
}

.error-dismiss {
	background: none;
	border: none;
	color: inherit;
	font-size: 12px;
	cursor: pointer;
	flex-shrink: 0;
}

.composer-dock {
	padding: 16px clamp(24px, 4vw, 64px) 24px;
	border-top: 1px solid var(--border);
	background: color-mix(in srgb, var(--bg-app) 92%, var(--accent-soft));
}

.composer-dock :deep(.composer) {
	max-width: 1680px;
	margin-right: auto;
}

.conversation-title {
	display: flex;
	align-items: center;
	justify-content: flex-start;
	gap: 10px;
	font-size: 15px;
	font-weight: 600;
	color: var(--text-primary);
}

.conversation-phase {
	font-size: 11px;
	color: var(--accent);
	background: var(--accent-soft);
	padding: 2px 8px;
	border-radius: var(--radius-pill);
}

@media (max-width: 860px) {
	.menu-button {
		display: inline-flex;
	}

	.sidebar-backdrop {
		display: block;
		position: fixed;
		inset: 0;
		z-index: var(--z-backdrop);
		background: rgba(15, 23, 42, 0.45);
	}

	.composer-dock {
		padding: 8px 16px 14px;
	}
}

/* Research Cockpit shell overrides */
.global-bar {
	display: flex;
	align-items: center;
	gap: 16px;
	min-height: 64px;
	padding: 0 clamp(20px, 3vw, 40px);
	background: color-mix(in srgb, var(--bg-app) 88%, #ffffff);
	border-bottom: 1px solid var(--border);
}

.global-bar-main { display: flex; flex-direction: column; min-width: 0; }
.page-kicker { font-size: 10px; font-weight: 700; letter-spacing: 0.16em; text-transform: uppercase; color: var(--accent); }
.page-title { margin: 2px 0 0; font-size: 18px; line-height: 1.25; font-weight: 700; letter-spacing: -0.03em; color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.global-bar-status { display: flex; align-items: center; gap: 10px; margin-left: auto; }
.status-chip { display: inline-flex; align-items: center; gap: 6px; height: 28px; padding: 0 10px; border: 1px solid var(--border); border-radius: var(--radius-sm); background: var(--bg-surface); color: var(--text-secondary); font-size: 11px; font-weight: 600; }
.status-chip i { width: 7px; height: 7px; border-radius: 50%; background: var(--accent); }
.status-chip.offline i { background: var(--danger); }
.user-chip { display: flex; align-items: center; gap: 8px; padding: 4px 8px 4px 4px; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-surface); }
.user-avatar { display: inline-flex; align-items: center; justify-content: center; width: 26px; height: 26px; border-radius: 6px; background: linear-gradient(135deg, var(--accent), #6c99ff); color: #fff; font-size: 12px; font-weight: 700; }
.user-detail { display: flex; flex-direction: column; line-height: 1.15; }
.user-detail strong { font-size: 11px; color: var(--text-primary); }
.user-detail small { font-size: 9px; letter-spacing: 0.08em; text-transform: uppercase; color: var(--text-muted); }

.workspace .menu-button { display: none; position: static; width: 38px; height: 38px; flex-shrink: 0; }
.workspace .logout-button { height: 32px; padding: 0 10px; border: 1px solid var(--border); border-radius: var(--radius-sm); background: var(--bg-surface); color: var(--text-secondary); font-size: 12px; font-weight: 600; cursor: pointer; }
.workspace .logout-button:hover { border-color: var(--danger); color: var(--danger); }

.research-layout { flex: 1; display: grid; grid-template-columns: minmax(0, 1fr) 320px; min-height: 0; }
.context-panel { min-width: 0; overflow-y: auto; padding: 24px; border-left: 1px solid var(--border); background: color-mix(in srgb, var(--bg-app) 84%, #ffffff); display: flex; flex-direction: column; gap: 12px; }
.context-card { padding: 16px; background: var(--bg-surface); border: 1px solid var(--border); border-radius: var(--radius-lg); box-shadow: var(--shadow-card); }
.context-card h2 { margin: 0 0 10px; font-size: 11px; font-weight: 700; letter-spacing: 0.1em; text-transform: uppercase; color: var(--text-muted); }
.context-primary { margin: 0; font-size: 16px; font-weight: 700; color: var(--text-primary); }
.context-muted { margin: 0; font-size: 13px; color: var(--text-muted); }
.context-path { display: block; margin-top: 8px; font-size: 11px; color: var(--text-muted); overflow-wrap: anywhere; }
.context-list { margin: 0; padding: 0; list-style: none; display: flex; flex-direction: column; gap: 8px; }
.context-list li { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding-bottom: 8px; border-bottom: 1px solid var(--border); }
.context-list li:last-child { border-bottom: 0; padding-bottom: 0; }
.context-list strong { font-size: 13px; color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.context-list span { flex-shrink: 0; font-size: 10px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--accent); }
.status-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.status-grid div { padding: 8px; border-radius: var(--radius-sm); background: color-mix(in srgb, var(--accent-soft) 48%, var(--bg-surface)); }
.status-grid small { display: block; margin-bottom: 3px; font-size: 10px; color: var(--text-muted); }
.status-grid strong { font-size: 12px; color: var(--text-primary); }
.conversation-heading { display: flex; flex-direction: column; gap: 4px; }
.conversation-kicker { font-size: 10px; font-weight: 700; letter-spacing: 0.14em; text-transform: uppercase; color: var(--accent); }
.workspace .conversation-title { font-size: 18px; font-weight: 700; letter-spacing: -0.025em; }
.workspace .conversation-phase { font-size: 10px; color: var(--accent-strong); background: var(--accent-soft); padding: 3px 7px; border-radius: var(--radius-sm); font-weight: 700; }
.workspace .composer-dock { padding: 14px clamp(24px, 4vw, 64px) 22px; border-top: 1px solid var(--border); background: color-mix(in srgb, var(--bg-app) 92%, #ffffff); }
.workspace .composer-dock :deep(.composer) { max-width: 1680px; margin-right: auto; }

@media (max-width: 1280px) { .research-layout { grid-template-columns: minmax(0, 1fr) 280px; } }
@media (max-width: 1080px) { .research-layout { grid-template-columns: minmax(0, 1fr); } .context-panel { display: none; } }
@media (max-width: 860px) {
	.global-bar { min-height: 56px; padding: 0 14px; gap: 10px; }
	.workspace .menu-button { display: inline-flex; }
	.user-detail, .page-kicker { display: none; }
	.sidebar-backdrop { display: block; position: fixed; inset: 0; z-index: var(--z-backdrop); background: rgba(15, 23, 42, 0.45); }
}
</style>
