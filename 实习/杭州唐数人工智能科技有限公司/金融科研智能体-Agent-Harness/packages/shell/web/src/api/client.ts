import type {
	AgentView,
	ApprovalRequestView,
	AdminSessionView,
	DatasetView,
	GroupDetail,
	GroupMemberView,
	GroupView,
	ModelMetadata,
	ModelRef,
	ProjectFileEntry,
	ProjectFileMeta,
	ProjectInfo,
	SessionContext,
	SessionEvent,
	SessionSnapshot,
	SessionSummary,
	ShellConfig,
	SkillView,
	ThinkingLevel,
	ToolInfo,
	TranscriptProgress,
	UserInfo,
} from "./types.ts";

export interface ApiError {
	code: string;
	message: string;
}

let authToken: string | null = null;

function setAuthToken(token: string | null): void {
	authToken = token;
}

async function toApiError(response: Response): Promise<Error> {
	let message = `HTTP ${response.status}`;
	try {
		const body = (await response.json()) as { error?: { message?: string } };
		if (body.error?.message) message = body.error.message;
	} catch {
		// 保留默认
	}
	return new Error(message);
}

async function request<T>(url: string, options: { method?: string; body?: unknown } = {}): Promise<T> {
	const headers = new Headers();
	if (authToken) headers.set("Authorization", `Bearer ${authToken}`);
	let body: string | undefined;
	if (options.body !== undefined) {
		headers.set("Content-Type", "application/json");
		body = JSON.stringify(options.body);
	}
	const response = await fetch(url, { method: options.method ?? "GET", headers, body });
	if (!response.ok) {
		let error: ApiError = { code: "http_error", message: `HTTP ${response.status}` };
		try {
			const body = (await response.json()) as { error?: ApiError };
			if (body.error) error = body.error;
		} catch {
			// 保留默认错误
		}
		throw error;
	}
	return (await response.json()) as T;
}

export class ApiClient {
	setToken(token: string | null): void {
		setAuthToken(token);
	}

	async login(username: string, password: string): Promise<{ token: string; user: UserInfo; expiresAt: number }> {
		const result = await request<{ token: string; user: UserInfo; expiresAt: number }>("/api/auth/login", {
			method: "POST",
			body: { username, password },
		});
		setAuthToken(result.token);
		return result;
	}

	async getMe(): Promise<UserInfo> {
		const result = await request<{ user: UserInfo }>("/api/auth/me");
		return result.user;
	}

	async logout(): Promise<void> {
		try {
			await request<{ ok: boolean }>("/api/auth/logout", { method: "POST" });
		} catch {
			// 忽略登出失败,本地仍清除 token
		}
		setAuthToken(null);
	}

	async listGroups(): Promise<GroupView[]> {
		const result = await request<{ groups: GroupView[] }>("/api/groups");
		return result.groups;
	}

	async getGroup(groupId: string): Promise<GroupDetail> {
		const result = await request<{ group: GroupDetail }>(`/api/groups/${groupId}`);
		return result.group;
	}

	async createGroup(input: { name: string; type: "course" | "research" }): Promise<GroupView> {
		const result = await request<{ group: GroupView }>("/api/groups", { method: "POST", body: input });
		return result.group;
	}

	async updateGroup(groupId: string, input: { name?: string }): Promise<GroupView> {
		const result = await request<{ group: GroupView }>(`/api/groups/${groupId}`, { method: "PATCH", body: input });
		return result.group;
	}

	async archiveGroup(groupId: string): Promise<GroupView> {
		const result = await request<{ group: GroupView }>(`/api/groups/${groupId}`, { method: "DELETE" });
		return result.group;
	}

	async addGroupMember(groupId: string, userId: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/groups/${groupId}/members`, { method: "POST", body: { userId } });
	}

	async updateGroupMemberRole(groupId: string, userId: string, role: "owner" | "member"): Promise<void> {
		await request<{ ok: boolean }>(`/api/groups/${groupId}/members/${userId}`, { method: "PATCH", body: { role } });
	}

	async removeGroupMember(groupId: string, userId: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/groups/${groupId}/members/${userId}`, { method: "DELETE" });
	}

	async listSkills(): Promise<SkillView[]> {
		const result = await request<{ skills: SkillView[] }>("/api/skills");
		return result.skills;
	}

	async createSkill(input: { name?: string; repoUrl?: string; visibility: "private" | "group" | "public"; groupId?: string; upload?: boolean }): Promise<SkillView> {
		const result = await request<{ skill: SkillView }>("/api/skills", { method: "POST", body: input });
		return result.skill;
	}

	async updateSkillFromRepo(skillId: string): Promise<SkillView> {
		const result = await request<{ skill: SkillView }>(`/api/skills/${skillId}/update`, { method: "POST" });
		return result.skill;
	}

	async updateSkill(skillId: string, input: { name?: string; visibility?: "private" | "group" | "public"; groupId?: string }): Promise<SkillView> {
		const result = await request<{ skill: SkillView }>(`/api/skills/${skillId}`, { method: "PATCH", body: input });
		return result.skill;
	}

	async deleteSkill(skillId: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/skills/${skillId}`, { method: "DELETE" });
	}

	async reviewSkill(skillId: string, reviewState: "approved" | "rejected"): Promise<SkillView> {
		const result = await request<{ skill: SkillView }>(`/api/admin/skills/${skillId}/review`, { method: "POST", body: { reviewState } });
		return result.skill;
	}

	async listDatasets(): Promise<DatasetView[]> {
		const result = await request<{ datasets: DatasetView[] }>("/api/datasets");
		return result.datasets;
	}

	async createDataset(input: { file: File; name?: string; visibility: "private" | "group" | "public"; groupId?: string }): Promise<DatasetView> {
		const form = new FormData();
		form.append("file", input.file);
		if (input.name) form.append("name", input.name);
		form.append("visibility", input.visibility);
		if (input.groupId) form.append("groupId", input.groupId);
		const headers = new Headers();
		if (authToken) headers.set("Authorization", "Bearer " + authToken);
		const response = await fetch("/api/datasets", { method: "POST", headers, body: form });
		if (!response.ok) {
			let error: ApiError = { code: "http_error", message: `HTTP ${response.status}` };
			try { const body = (await response.json()) as { error?: ApiError }; if (body.error) error = body.error; } catch { /* 保留默认 */ }
			throw error;
		}
		const result = (await response.json()) as { dataset: DatasetView };
		return result.dataset;
	}

	async updateDataset(datasetId: string, input: { name?: string; visibility?: "private" | "group" | "public"; groupId?: string }): Promise<DatasetView> {
		const result = await request<{ dataset: DatasetView }>(`/api/datasets/${datasetId}`, { method: "PATCH", body: input });
		return result.dataset;
	}

	async deleteDataset(datasetId: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/datasets/${datasetId}`, { method: "DELETE" });
	}

	async reviewDataset(datasetId: string, reviewState: "approved" | "rejected"): Promise<DatasetView> {
		const result = await request<{ dataset: DatasetView }>(`/api/admin/datasets/${datasetId}/review`, { method: "POST", body: { reviewState } });
		return result.dataset;
	}

	async downloadDataset(datasetId: string): Promise<void> {
		const headers = new Headers();
		if (authToken) headers.set("Authorization", "Bearer " + authToken);
		const response = await fetch(`/api/datasets/${datasetId}/download`, { headers });
		if (!response.ok) {
			let error: ApiError = { code: "http_error", message: `HTTP ${response.status}` };
			try { const body = (await response.json()) as { error?: ApiError }; if (body.error) error = body.error; } catch { /* 保留默认 */ }
			throw error;
		}
		const blob = await response.blob();
		const disposition = response.headers.get("Content-Disposition") ?? "";
		const m = /filename\*=UTF-8\x27\x27([^;]+)/i.exec(disposition);
		const name = m ? decodeURIComponent(m[1]) : "dataset";
		const url = URL.createObjectURL(blob);
		const anchor = document.createElement("a");
		anchor.href = url;
		anchor.download = name;
		anchor.click();
		URL.revokeObjectURL(url);
	}

	async getPermissionMode(sessionId: string): Promise<string> {
		const result = await request<{ mode: string }>(`/api/sessions/${sessionId}/permission-mode`);
		return result.mode;
	}

	async setPermissionMode(sessionId: string, mode: string): Promise<string> {
		const result = await request<{ mode: string }>(`/api/sessions/${sessionId}/permission-mode`, {
			method: "POST",
			body: { mode },
		});
		return result.mode;
	}

	async listApprovals(sessionId: string): Promise<ApprovalRequestView[]> {
		const result = await request<{ approvals: ApprovalRequestView[] }>(`/api/sessions/${sessionId}/approvals`);
		return result.approvals;
	}

	async decideApproval(sessionId: string, approvalId: string, approved: boolean): Promise<ApprovalRequestView> {
		const action = approved ? "approve" : "reject";
		const result = await request<{ approval: ApprovalRequestView }>(
			`/api/sessions/${sessionId}/approvals/${approvalId}/${action}`,
			{ method: "POST" },
		);
		return result.approval;
	}
	async listAgents(): Promise<AgentView[]> {
		const result = await request<{ agents: AgentView[] }>("/api/agents");
		return result.agents;
	}

	async createAgent(input: { name: string; url: string; description?: string; visibility: string; groupId?: string }): Promise<AgentView> {
		const result = await request<{ agent: AgentView }>("/api/agents", { method: "POST", body: input });
		return result.agent;
	}

	async updateAgent(id: string, input: { name?: string; url?: string; description?: string | null; visibility?: string; groupId?: string }): Promise<AgentView> {
		const result = await request<{ agent: AgentView }>(`/api/agents/${id}`, { method: "PATCH", body: input });
		return result.agent;
	}

	async deleteAgent(id: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/agents/${id}`, { method: "DELETE" });
	}
	async getProjectSessionContext(projectId: string): Promise<SessionContext> {
		return request<SessionContext>(`/api/projects/${projectId}/session-context`);
	}

	async listProjectFiles(projectId: string, path?: string): Promise<ProjectFileEntry[]> {
		const q = path && path.length > 0 ? `?path=${encodeURIComponent(path)}` : "";
		const result = await request<{ files: ProjectFileEntry[] }>(`/api/projects/${projectId}/files${q}`);
		return result.files;
	}

	async readProjectFile(projectId: string, path: string): Promise<ProjectFileMeta & { content: string }> {
		const result = await request<{ file: ProjectFileMeta & { content: string } }>(
			`/api/projects/${projectId}/files/content?path=${encodeURIComponent(path)}`,
		);
		return result.file;
	}

	async writeProjectFile(projectId: string, path: string, content: string): Promise<ProjectFileMeta> {
		const result = await request<{ file: ProjectFileMeta }>(`/api/projects/${projectId}/files/content`, {
			method: "PUT",
			body: { path, content },
		});
		return result.file;
	}

	async uploadProjectFile(projectId: string, file: File, dir?: string): Promise<ProjectFileMeta> {
		const form = new FormData();
		form.append("file", file);
		if (dir) form.append("path", dir);
		const headers = new Headers();
		if (authToken) headers.set("Authorization", `Bearer ${authToken}`);
		const response = await fetch(`/api/projects/${projectId}/files`, { method: "POST", headers, body: form });
		if (!response.ok) throw await toApiError(response);
		const body = (await response.json()) as { file: ProjectFileMeta };
		return body.file;
	}

	async downloadProjectFile(projectId: string, path: string): Promise<{ blob: Blob; name: string; mimeType: string }> {
		const headers = new Headers();
		if (authToken) headers.set("Authorization", `Bearer ${authToken}`);
		const response = await fetch(`/api/projects/${projectId}/files/download?path=${encodeURIComponent(path)}`, { headers });
		if (!response.ok) throw await toApiError(response);
		const disposition = response.headers.get("Content-Disposition") ?? "";
		const match = /filename\*=UTF-8''([^;]+)/i.exec(disposition);
		const name = match ? decodeURIComponent(match[1]) : path.split("/").pop() ?? path;
		return { blob: await response.blob(), name, mimeType: response.headers.get("Content-Type") ?? "application/octet-stream" };
	}

	async deleteProjectFile(projectId: string, path: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/projects/${projectId}/files?path=${encodeURIComponent(path)}`, { method: "DELETE" });
	}

	async listUsers(): Promise<UserInfo[]> {
		const result = await request<{ users: UserInfo[] }>("/api/admin/users");
		return result.users;
	}

	async createUser(input: { username: string; password: string; role: "admin" | "teacher" | "student" }): Promise<UserInfo> {
		const result = await request<{ user: UserInfo }>("/api/admin/users", { method: "POST", body: input });
		return result.user;
	}

	async updateUser(userId: string, input: { status?: "active" | "disabled"; password?: string; role?: "admin" | "teacher" | "student" }): Promise<UserInfo> {
		const result = await request<{ user: UserInfo }>(`/api/admin/users/${userId}`, { method: "PATCH", body: input });
		return result.user;
	}

	async listAdminSessions(): Promise<AdminSessionView[]> {
		const result = await request<{ sessions: AdminSessionView[] }>("/api/admin/sessions");
		return result.sessions;
	}

	async terminateAdminSession(sessionId: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/admin/sessions/${sessionId}/terminate`, { method: "POST" });
	}

	async configureApiKey(provider: string, apiKey: string): Promise<void> {
		await request<{ ok: boolean }>("/api/config/api-key", { method: "POST", body: { provider, apiKey } });
	}
	async listSessions(): Promise<SessionSummary[]> {
		const result = await request<{ sessions: SessionSummary[] }>("/api/sessions");
		return result.sessions;
	}

	async createSession(options: { name?: string; cwd?: string; projectId?: string; model?: ModelRef; thinkingLevel?: ThinkingLevel } = {}): Promise<SessionSnapshot> {
		const result = await request<{ session: SessionSnapshot }>("/api/sessions", { method: "POST", body: options });
		return result.session;
	}

	async deleteSession(sessionId: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/sessions/${sessionId}`, { method: "DELETE" });
	}

	async listModels(): Promise<ModelMetadata[]> {
		const result = await request<{ models: ModelMetadata[] }>("/api/models");
		return result.models;
	}

	async getSnapshot(sessionId: string): Promise<SessionSnapshot> {
		const result = await request<{ session: SessionSnapshot }>(`/api/sessions/${sessionId}`);
		return result.session;
	}

	async prompt(sessionId: string, text: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/sessions/${sessionId}/prompt`, { method: "POST", body: { text } });
	}

	async steer(sessionId: string, text: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/sessions/${sessionId}/steer`, { method: "POST", body: { text } });
	}

	async abort(sessionId: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/sessions/${sessionId}/abort`, { method: "POST" });
	}

	async setModel(sessionId: string, model: ModelRef): Promise<void> {
		await request<{ ok: boolean }>(`/api/sessions/${sessionId}/model`, { method: "POST", body: { model } });
	}

	async setThinking(sessionId: string, thinkingLevel: ThinkingLevel): Promise<void> {
		await request<{ ok: boolean }>(`/api/sessions/${sessionId}/thinking`, {
			method: "POST",
			body: { thinkingLevel },
		});
	}

	async getConfig(): Promise<ShellConfig> {
		const result = await request<{ config: ShellConfig }>("/api/config");
		return result.config;
	}

	async listTools(): Promise<ToolInfo[]> {
		const result = await request<{ tools: ToolInfo[] }>("/api/tools");
		return result.tools;
	}

	async listProjects(): Promise<ProjectInfo[]> {
		const result = await request<{ projects: ProjectInfo[] }>("/api/projects");
		return result.projects;
	}

	async createProject(input: { name: string; folder?: string; scope?: { type: "private" } | { type: "group"; groupId: string } }): Promise<ProjectInfo> {
		const result = await request<{ project: ProjectInfo }>("/api/projects", { method: "POST", body: input });
		return result.project;
	}

	async updateProject(id: string, input: { name?: string; folder?: string }): Promise<ProjectInfo> {
		const result = await request<{ project: ProjectInfo }>(`/api/projects/${id}`, { method: "PATCH", body: input });
		return result.project;
	}

	async deleteProject(id: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/projects/${id}`, { method: "DELETE" });
	}

	async moveSessionProject(sessionId: string, projectId: string | null): Promise<void> {
		await request<{ ok: boolean }>(`/api/sessions/${sessionId}/project`, { method: "POST", body: { projectId } });
	}

	async removeApiKey(provider: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/config/api-key?provider=${encodeURIComponent(provider)}`, { method: "DELETE" });
	}

	async renameSession(sessionId: string, name: string): Promise<void> {
		await request<{ ok: boolean }>(`/api/sessions/${sessionId}/rename`, { method: "POST", body: { name } });
	}

	async exportSession(sessionId: string): Promise<string> {
		const response = await fetch(`/api/sessions/${sessionId}/export`);
		if (!response.ok) {
			let message = `HTTP ${response.status}`;
			try {
				const body = (await response.json()) as { error?: ApiError };
				if (body.error) message = body.error.message;
			} catch {
				// 保留默认错误
			}
			throw new Error(message);
		}
		return response.text();
	}

	/**
	 * 订阅会话事件流(SSE,EventSource 自动重连)。
	 * 返回退订函数;会话切换时应先退订旧连接。
	 */
	subscribeEvents(sessionId: string, onEvent: (event: SessionEvent) => void, onStatus?: (connected: boolean) => void): () => void {
		const source = new EventSource(`/api/sessions/${sessionId}/events`);
		source.onopen = () => onStatus?.(true);
		source.onerror = () => onStatus?.(false);
		source.addEventListener("snapshot", (message) => {
			onEvent({ type: "snapshot" });
		});
		source.addEventListener("progress", (message) => {
			try {
				const data = JSON.parse((message as MessageEvent<string>).data) as { progress: TranscriptProgress };
				onEvent({ type: "progress", progress: data.progress });
			} catch {
				// 忽略解析失败的帧
			}
		});
		source.addEventListener("error", (message) => {
			try {
				const data = JSON.parse((message as MessageEvent<string>).data) as { error: { code: string; message: string } };
				onEvent({ type: "error", error: data.error });
			} catch {
				// 忽略解析失败的帧
			}
		});
		return () => {
			source.close();
			onStatus?.(false);
		};
	}
}
