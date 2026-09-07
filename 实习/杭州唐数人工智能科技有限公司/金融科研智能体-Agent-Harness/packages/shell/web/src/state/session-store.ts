/**
 * 全局会话状态:会话列表 + 当前会话 + transcript + 连接状态。
 */

import { shallowReactive } from "vue";
import { ApiClient } from "../api/client.ts";
import type {
	ApprovalRequestView,
	AdminSessionView,
	DatasetView,
	GroupView,
	ModelMetadata,
	ProjectFileEntry,
	ProjectFileMeta,
	ProjectInfo,
	SessionContext,
	SessionEvent,
	SkillView,
	SessionSnapshot,
	SessionSummary,
	ShellConfig,
	ThinkingLevel,
	ToolInfo,
	UserInfo,
} from "../api/types.ts";
import {
	applyTranscriptProgress,
	applyTranscriptSnapshot,
	createTranscriptState,
	selectTranscript,
	type TranscriptState,
} from "./transcript.ts";

export const api = new ApiClient();

const TOKEN_KEY = "tfa-token";
const storedToken = localStorage.getItem(TOKEN_KEY);
if (storedToken) api.setToken(storedToken);

export interface WorkbenchState {
	projectId: string | null;
	/** 目录路径 → 条目列表("" 为根);按需懒加载。 */
	tree: Record<string, ProjectFileEntry[]>;
	expandedDirs: string[];
	selectedPath: string | null;
	fileMeta: ProjectFileMeta | null;
	fileContent: string | null;
	loading: boolean;
	error: string | null;
}

export interface SessionStoreState {
	user: UserInfo | null;
	workbench: WorkbenchState;
	sessions: SessionSummary[];
	models: ModelMetadata[];
	config: ShellConfig | null;
	tools: ToolInfo[];
	projects: ProjectInfo[];
	groups: GroupView[];
	skills: SkillView[];
	datasets: DatasetView[];
	users: UserInfo[];
	adminSessions: AdminSessionView[];
	sessionContext: SessionContext | null;
	contextProjectId: string | null;
	permissionMode: string | null;
	pendingApprovals: ApprovalRequestView[];
	currentId: string | null;
	snapshot: SessionSnapshot | null;
	transcriptState: TranscriptState;
	connected: boolean;
	busy: boolean;
	error: string | null;
}

export const store = shallowReactive<SessionStoreState>({
	user: null,
	workbench: {
		projectId: null,
		tree: {},
		expandedDirs: [],
		selectedPath: null,
		fileMeta: null,
		fileContent: null,
		loading: false,
		error: null,
	},
	sessions: [],
	models: [],
	config: null,
	tools: [],
	projects: [],
	groups: [],
	skills: [],
	datasets: [],
	users: [],
	adminSessions: [],
	sessionContext: null,
	contextProjectId: null,
	permissionMode: null,
	pendingApprovals: [],
	currentId: null,
	snapshot: null,
	transcriptState: createTranscriptState(),
	connected: false,
	busy: false,
	error: null,
});

export const transcriptItems = () => selectTranscript(store.transcriptState);

let unsubscribeEvents: (() => void) | undefined;

export async function bootstrap(): Promise<void> {
	if (!localStorage.getItem(TOKEN_KEY)) {
		store.user = null;
		return;
	}
	try {
		store.user = await api.getMe();
		await refreshAll();
	} catch {
		store.user = null;
	}
}

export async function login(username: string, password: string): Promise<void> {
	try {
		const result = await api.login(username, password);
		localStorage.setItem(TOKEN_KEY, result.token);
		store.user = result.user;
		store.error = null;
		await refreshAll();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function logout(): Promise<void> {
	localStorage.removeItem(TOKEN_KEY);
	await api.logout();
	store.user = null;
	store.currentId = null;
	store.snapshot = null;
	store.transcriptState = createTranscriptState();
	store.sessions = [];
	store.projects = [];
	store.groups = [];
	store.sessionContext = null;
	store.contextProjectId = null;
}

export async function refreshAll(): Promise<void> {
	await Promise.all([refreshSessions(), refreshModels(), refreshConfig(), refreshTools(), refreshProjects(), refreshGroups(), refreshSkills(), refreshDatasets()]);
}

export async function refreshGroups(): Promise<void> {
	try {
		store.groups = await api.listGroups();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

/** 拉取项目会话上下文(可见技能/数据集);null 清空。 */
export async function refreshContext(projectId: string | null): Promise<void> {
	if (!projectId) {
		store.sessionContext = null;
		store.contextProjectId = null;
		return;
	}
	try {
		store.sessionContext = await api.getProjectSessionContext(projectId);
		store.contextProjectId = projectId;
	} catch (error) {
		store.sessionContext = null;
		store.contextProjectId = null;
		store.error = errorMessage(error);
	}
}
export async function refreshSessions(): Promise<void> {
	try {
		store.sessions = await api.listSessions();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function refreshModels(): Promise<void> {
	try {
		store.models = await api.listModels();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function refreshConfig(): Promise<void> {
	try {
		store.config = await api.getConfig();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function refreshTools(): Promise<void> {
	try {
		store.tools = await api.listTools();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function refreshProjects(): Promise<void> {
	try {
		store.projects = await api.listProjects();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function configureApiKey(provider: string, apiKey: string): Promise<void> {
	try {
		await api.configureApiKey(provider, apiKey);
		store.error = null;
		await refreshModels();
		await refreshConfig();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function removeApiKey(provider: string): Promise<void> {
	try {
		await api.removeApiKey(provider);
		store.error = null;
		await refreshModels();
		await refreshConfig();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function renameSession(sessionId: string, name: string): Promise<void> {
	try {
		await api.renameSession(sessionId, name);
		await refreshSessions();
		if (store.currentId === sessionId) {
			await openSession(sessionId);
		}
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function exportSession(sessionId: string): Promise<string> {
	return api.exportSession(sessionId);
}

export async function openSession(sessionId: string): Promise<void> {
	unsubscribeEvents?.();
	store.currentId = sessionId;
	store.error = null;
	store.busy = false;
	try {
		const snapshot = await api.getSnapshot(sessionId);
		applySnapshot(snapshot);
	} catch (error) {
		store.error = errorMessage(error);
	}
	unsubscribeEvents = api.subscribeEvents(sessionId, handleEvent, (connected) => {
		store.connected = connected;
	});
}

export async function createSession(options: { name?: string; cwd?: string; projectId?: string } = {}): Promise<void> {
	try {
		const snapshot = await api.createSession({
			name: options.name?.trim() || nextSessionName(),
			cwd: options.cwd,
			projectId: options.projectId,
		});
		await refreshSessions();
		await refreshProjects();
		await openSession(snapshot.id);
	} catch (error) {
		store.error = errorMessage(error);
	}
}

/** 按现有会话名生成下一个「会话N」名称。 */
function nextSessionName(): string {
	let max = 0;
	for (const session of store.sessions) {
		const match = /^会话(\d+)$/.exec(session.name ?? "");
		if (match) {
			const number = Number.parseInt(match[1], 10);
			if (Number.isFinite(number) && number > max) max = number;
		}
	}
	return `会话${max + 1}`;
}

export async function deleteSession(sessionId: string): Promise<void> {
	try {
		await api.deleteSession(sessionId);
		if (store.currentId === sessionId) {
			unsubscribeEvents?.();
			unsubscribeEvents = undefined;
			store.currentId = null;
			store.snapshot = null;
			store.transcriptState = createTranscriptState();
		}
		await refreshSessions();
		await refreshProjects();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function deleteCurrentSession(): Promise<void> {
	if (!store.currentId) return;
	await deleteSession(store.currentId);
}

export async function createProject(input: { name: string; folder?: string; scope?: { type: "private" } | { type: "group"; groupId: string } }): Promise<void> {
	try {
		await api.createProject(input);
		store.error = null;
		await refreshProjects();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function updateProject(id: string, input: { name?: string; folder?: string }): Promise<void> {
	try {
		await api.updateProject(id, input);
		store.error = null;
		await refreshProjects();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function deleteProject(id: string): Promise<void> {
	try {
		await api.deleteProject(id);
		store.error = null;
		await refreshProjects();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function moveSessionProject(sessionId: string, projectId: string | null): Promise<void> {
	try {
		await api.moveSessionProject(sessionId, projectId);
		store.error = null;
		await refreshProjects();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function sendPrompt(text: string): Promise<void> {
	if (!store.currentId || store.busy) return;
	store.error = null;
	store.busy = true;
	try {
		await api.prompt(store.currentId, text);
	} catch (error) {
		store.error = errorMessage(error);
	} finally {
		store.busy = false;
	}
}

export async function sendSteer(text: string): Promise<void> {
	if (!store.currentId) return;
	store.error = null;
	try {
		await api.steer(store.currentId, text);
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function abortCurrent(): Promise<void> {
	if (!store.currentId) return;
	try {
		await api.abort(store.currentId);
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function switchModel(provider: string, id: string): Promise<void> {
	if (!store.currentId) return;
	try {
		await api.setModel(store.currentId, { provider, id });
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function setThinking(thinkingLevel: ThinkingLevel): Promise<void> {
	if (!store.currentId) return;
	store.error = null;
	try {
		await api.setThinking(store.currentId, thinkingLevel);
		await refreshSnapshot();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

function applySnapshot(snapshot: SessionSnapshot): void {
	store.snapshot = snapshot;
	store.transcriptState = applyTranscriptSnapshot(store.transcriptState, snapshot);
}

function handleEvent(event: SessionEvent): void {
	if (event.type === "snapshot") {
		void refreshSnapshot();
		return;
	}
	if (event.type === "progress") {
		store.transcriptState = applyTranscriptProgress(store.transcriptState, event.progress);
		return;
	}
	store.error = event.error.message;
}

async function refreshSnapshot(): Promise<void> {
	if (!store.currentId) return;
	try {
		const snapshot = await api.getSnapshot(store.currentId);
		applySnapshot(snapshot);
	} catch (error) {
		store.error = errorMessage(error);
	}
}

function errorMessage(error: unknown): string {
	if (error && typeof error === "object" && "message" in error) {
		return String((error as { message: unknown }).message);
	}
	return String(error);
}

// ============ 平台:课程组 / 技能库 / 数据集库 / 管理端 / 监控 ============

export async function refreshSkills(): Promise<void> {
	try {
		store.skills = await api.listSkills();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function refreshDatasets(): Promise<void> {
	try {
		store.datasets = await api.listDatasets();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function refreshUsers(): Promise<void> {
	try {
		store.users = await api.listUsers();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function refreshAdminSessions(): Promise<void> {
	try {
		store.adminSessions = await api.listAdminSessions();
	} catch (error) {
		store.error = errorMessage(error);
	}
}

// ---- 课程组 ----
export async function createGroup(input: { name: string; type: "course" | "research" }): Promise<void> {
	try {
		await api.createGroup(input);
		store.error = null;
		await refreshGroups();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function updateGroup(groupId: string, input: { name?: string }): Promise<void> {
	try {
		await api.updateGroup(groupId, input);
		store.error = null;
		await refreshGroups();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function archiveGroup(groupId: string): Promise<void> {
	try {
		await api.archiveGroup(groupId);
		store.error = null;
		await refreshGroups();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function addGroupMember(groupId: string, userId: string): Promise<void> {
	try {
		await api.addGroupMember(groupId, userId);
		store.error = null;
		await refreshGroups();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function updateGroupMemberRole(groupId: string, userId: string, role: "owner" | "member"): Promise<void> {
	try {
		await api.updateGroupMemberRole(groupId, userId, role);
		store.error = null;
		await refreshGroups();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function removeGroupMember(groupId: string, userId: string): Promise<void> {
	try {
		await api.removeGroupMember(groupId, userId);
		store.error = null;
		await refreshGroups();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

// ---- 技能库 ----
export async function createSkill(input: { name?: string; repoUrl?: string; visibility: "private" | "group" | "public"; groupId?: string; upload?: boolean }): Promise<void> {
	try {
		await api.createSkill(input);
		store.error = null;
		await refreshSkills();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function updateSkillFromRepo(skillId: string): Promise<void> {
	try {
		await api.updateSkillFromRepo(skillId);
		store.error = null;
		await refreshSkills();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function updateSkill(skillId: string, input: { name?: string; visibility?: "private" | "group" | "public"; groupId?: string }): Promise<void> {
	try {
		await api.updateSkill(skillId, input);
		store.error = null;
		await refreshSkills();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function deleteSkill(skillId: string): Promise<void> {
	try {
		await api.deleteSkill(skillId);
		store.error = null;
		await refreshSkills();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function reviewSkill(skillId: string, reviewState: "approved" | "rejected"): Promise<void> {
	try {
		await api.reviewSkill(skillId, reviewState);
		store.error = null;
		await Promise.all([refreshSkills(), refreshDatasets()]);
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

// ---- 数据集库 ----
export async function createDataset(input: { file: File; name?: string; visibility: "private" | "group" | "public"; groupId?: string }): Promise<void> {
	try {
		await api.createDataset(input);
		store.error = null;
		await refreshDatasets();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function updateDataset(datasetId: string, input: { name?: string; visibility?: "private" | "group" | "public"; groupId?: string }): Promise<void> {
	try {
		await api.updateDataset(datasetId, input);
		store.error = null;
		await refreshDatasets();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function deleteDataset(datasetId: string): Promise<void> {
	try {
		await api.deleteDataset(datasetId);
		store.error = null;
		await refreshDatasets();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function reviewDataset(datasetId: string, reviewState: "approved" | "rejected"): Promise<void> {
	try {
		await api.reviewDataset(datasetId, reviewState);
		store.error = null;
		await Promise.all([refreshSkills(), refreshDatasets()]);
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function downloadDataset(datasetId: string): Promise<void> {
	try {
		await api.downloadDataset(datasetId);
	} catch (error) {
		store.error = errorMessage(error);
	}
}

// ---- 管理端 ----
export async function createUser(input: { username: string; password: string; role: "admin" | "teacher" | "student" }): Promise<void> {
	try {
		await api.createUser(input);
		store.error = null;
		await refreshUsers();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function updateUser(userId: string, input: { status?: "active" | "disabled"; password?: string; role?: "admin" | "teacher" | "student" }): Promise<void> {
	try {
		await api.updateUser(userId, input);
		store.error = null;
		await refreshUsers();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

export async function terminateAdminSession(sessionId: string): Promise<void> {
	try {
		await api.terminateAdminSession(sessionId);
		store.error = null;
		await refreshAdminSessions();
	} catch (error) {
		store.error = errorMessage(error);
		throw error;
	}
}

function dirnamePath(path: string): string {
	const index = path.lastIndexOf("/");
	return index > 0 ? path.slice(0, index) : "";
}

export async function openWorkbench(projectId: string): Promise<void> {
	store.workbench.projectId = projectId;
	store.workbench.selectedPath = null;
	store.workbench.fileMeta = null;
	store.workbench.fileContent = null;
	store.workbench.tree = {};
	store.workbench.expandedDirs = [];
	await refreshDir("");
}

export function closeWorkbench(): void {
	store.workbench.projectId = null;
	store.workbench.tree = {};
	store.workbench.expandedDirs = [];
	store.workbench.selectedPath = null;
	store.workbench.fileMeta = null;
	store.workbench.fileContent = null;
}

export async function refreshDir(path: string): Promise<void> {
	const projectId = store.workbench.projectId;
	if (!projectId) return;
	store.workbench.loading = true;
	try {
		store.workbench.tree[path] = await api.listProjectFiles(projectId, path || undefined);
		store.workbench.error = null;
	} catch (error) {
		store.workbench.error = errorMessage(error);
	} finally {
		store.workbench.loading = false;
	}
}

export async function toggleDir(path: string): Promise<void> {
	if (store.workbench.expandedDirs.includes(path)) {
		store.workbench.expandedDirs = store.workbench.expandedDirs.filter((item) => item !== path);
		return;
	}
	store.workbench.expandedDirs = [...store.workbench.expandedDirs, path];
	await refreshDir(path);
}

export async function selectFile(path: string): Promise<void> {
	const projectId = store.workbench.projectId;
	if (!projectId) return;
	store.workbench.selectedPath = path;
	store.workbench.loading = true;
	try {
		const file = await api.readProjectFile(projectId, path);
		store.workbench.fileMeta = { path: file.path, name: file.name, size: file.size, mimeType: file.mimeType };
		store.workbench.fileContent = file.content;
		store.workbench.error = null;
	} catch (error) {
		store.workbench.error = errorMessage(error);
	} finally {
		store.workbench.loading = false;
	}
}

export async function saveCurrentFile(): Promise<void> {
	const projectId = store.workbench.projectId;
	const path = store.workbench.selectedPath;
	const content = store.workbench.fileContent;
	if (!projectId || !path || content === null) return;
	try {
		store.workbench.fileMeta = await api.writeProjectFile(projectId, path, content);
		store.workbench.error = null;
		await refreshDir(dirnamePath(path));
	} catch (error) {
		store.workbench.error = errorMessage(error);
	}
}

export async function deleteEntry(path: string): Promise<void> {
	const projectId = store.workbench.projectId;
	if (!projectId) return;
	try {
		await api.deleteProjectFile(projectId, path);
		if (store.workbench.selectedPath === path) {
			store.workbench.selectedPath = null;
			store.workbench.fileMeta = null;
			store.workbench.fileContent = null;
		}
		store.workbench.error = null;
		await refreshDir(dirnamePath(path));
	} catch (error) {
		store.workbench.error = errorMessage(error);
	}
}

export async function uploadFiles(files: FileList | File[], dir?: string): Promise<void> {
	const projectId = store.workbench.projectId;
	if (!projectId) return;
	store.workbench.loading = true;
	try {
		for (const file of Array.from(files)) {
			await api.uploadProjectFile(projectId, file, dir);
		}
		await refreshDir(dir ?? "");
		store.workbench.error = null;
	} catch (error) {
		store.workbench.error = errorMessage(error);
	} finally {
		store.workbench.loading = false;
	}
}

export async function refreshPermissionMode(): Promise<void> {
	const sessionId = store.currentId;
	if (!sessionId) return;
	try {
		store.permissionMode = await api.getPermissionMode(sessionId);
		store.error = null;
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function setPermissionMode(mode: string): Promise<void> {
	const sessionId = store.currentId;
	if (!sessionId) return;
	try {
		store.permissionMode = await api.setPermissionMode(sessionId, mode);
		store.error = null;
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function refreshApprovals(): Promise<void> {
	const sessionId = store.currentId;
	if (!sessionId) return;
	try {
		store.pendingApprovals = await api.listApprovals(sessionId);
		store.error = null;
	} catch (error) {
		store.error = errorMessage(error);
	}
}

export async function decideApproval(approvalId: string, approved: boolean): Promise<void> {
	const sessionId = store.currentId;
	if (!sessionId) return;
	try {
		await api.decideApproval(sessionId, approvalId, approved);
		store.error = null;
		await refreshApprovals();
	} catch (error) {
		store.error = errorMessage(error);
	}
}
