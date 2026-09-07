import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import type { Model } from "@earendil-works/tfa-ai";
import {
	type AgentSession,
	type AgentSessionEvent,
	createAgentSession,
	getAgentDir,
	ModelRuntime,
	SessionManager,
	type ToolDefinition,
} from "@earendil-works/tfa-coding-agent";
import type {
	ModelMetadata,
	ModelRef,
	SessionPhase,
	SessionSnapshot,
	SessionSummary,
	ThinkingLevel,
} from "@earendil-works/tfa-protocol";
import {
	type CreateSessionOptions,
	type PromptInput,
	SessionNotFoundError,
	type SteerInput,
	TfaServerError,
	type TfaServerService,
	type TfaSessionRuntime,
	type TfaSessionRuntimeEvent,
	toProtocolModelMetadata,
} from "@earendil-works/tfa-server";
import type { FauxSetup } from "../core/faux.ts";
import { AgentService } from "./agents.ts";
import { type ApprovalRequest, ApprovalService } from "./approvals.ts";
import {
	AUTH_SESSION_TTL_MS,
	AuthError,
	type AuthUserView,
	generateToken,
	hashPassword,
	hashToken,
	verifyPassword,
} from "./auth.ts";
import { buildSkillCliArgs, createRpcSessionClient, ProxyRuntime } from "./bridge.ts";
import { DatasetService, type DatasetView } from "./datasets.ts";
import type { PlatformRepos } from "./db/repos.ts";
import type { PlatformSessionRecord, SessionPermissionMode, UserRecord, UserRole, UserStatus } from "./db/types.ts";
import { loadExecutorConfig } from "./executor/config.ts";
import { DockerExecutor } from "./executor/docker.ts";
import type { SessionExecutor } from "./executor/types.ts";
import { GroupService } from "./groups.ts";
import { type ProgressState, toTranscriptProgress } from "./progress.ts";
import { ProjectFileService } from "./project-files.ts";
import { type CreateProjectInput, ProjectStore, type ProjectView, type UpdateProjectInput } from "./project-store.ts";
import { SkillService, type SkillView } from "./skills.ts";
import { buildSessionSnapshot, getSessionPhase } from "./snapshot.ts";

export interface ShellSessionServiceOptions {
	/** 默认工作目录 */
	cwd: string;
	/** 全局配置目录,默认 getAgentDir() */
	agentDir?: string;
	/** 自定义(金融)工具 */
	customTools?: ToolDefinition[];
	/** 工具允许名单 */
	toolAllowlist?: string[];
	/** 默认模型 */
	defaultModel?: Model<any>;
	/** faux 离线验证设置 */
	fauxSetup?: FauxSetup;
	/** 数据集单文件上限(默认 500MB)。 */
	datasetMaxFileBytes?: number;
	/** 数据集每用户累计上限(默认 20GB)。 */
	datasetMaxUserBytes?: number;
	/** 会话执行器(容器监控/终止);缺省按环境配置创建 DockerExecutor。 */
	executor?: SessionExecutor;
}

export interface ProviderConfigInfo {
	provider: string;
	configured: boolean;
	source: string;
}

export interface ShellConfigInfo {
	agentDir: string;
	cwd: string;
	authPath: string;
	modelsPath: string;
	providers: ProviderConfigInfo[];
}

export interface ToolInfo {
	name: string;
	label: string;
	description: string;
}

export interface RpcSessionCreateOptions {
	sessionId: string;
	/** tfa CLI 入口。 */
	cliPath: string;
	/** 会话工作目录(项目目录)。 */
	cwd: string;
	env: Record<string, string>;
	provider?: string;
	model?: string;
	/** 项目可见技能目录(宿主路径),经 --no-skills --skill 注入。 */
	skills?: string[];
}

export interface AdminSessionView {
	id: string;
	userId: string | null;
	username: string | null;
	projectId: string | null;
	projectName: string | null;
	containerId: string | null;
	/** 容器状态:running/exited/missing/expired/error;非容器会话为 ended/null。 */
	containerState: string | null;
	status: string;
	cwd: string | null;
	startedAt: number | null;
	endedAt: number | null;
}

export type { CreateSessionOptions, SessionSnapshot, SessionSummary, TfaServerService, TfaSessionRuntime };

function parseHeaderTimestamp(timestamp: string): number {
	const parsed = Date.parse(timestamp);
	return Number.isFinite(parsed) ? parsed : Date.now();
}

function lastMessageTimestamp(session: AgentSession): number | undefined {
	const messages = session.messages;
	for (let index = messages.length - 1; index >= 0; index--) {
		const message = messages[index];
		if (message.role !== "toolResult") return message.timestamp;
	}
	return undefined;
}

/**
 * 一个 live 会话:持有 AgentSession,实现 TfaSessionRuntime 契约。
 * 冲突操作拒绝(reject)而非排队,与协议契约一致。
 */
class LiveShellSession implements TfaSessionRuntime {
	private readonly session: AgentSession;
	private readonly modelRuntime: ModelRuntime;
	private readonly progressState: ProgressState = { toolItemsStarted: new Set() };
	private readonly listeners = new Set<(event: TfaSessionRuntimeEvent) => void>();
	private readonly unsubscribe: () => void;
	private revision = 0;

	constructor(session: AgentSession, modelRuntime: ModelRuntime) {
		this.session = session;
		this.modelRuntime = modelRuntime;
		this.unsubscribe = session.subscribe((event) => {
			this.handleEvent(event);
		});
	}

	private handleEvent(event: AgentSessionEvent): void {
		if (event.type === "message_update") {
			const progress = toTranscriptProgress(event, this.session, this.progressState);
			if (progress) {
				this.emit({ type: "progress", progress });
			}
			return;
		}
		const progress = toTranscriptProgress(event, this.session, this.progressState);
		if (progress) {
			this.emit({ type: "progress", progress });
		} else {
			// agent_start/end、queue_update、compaction 等结构性事件 → 触发快照广播
			this.emit({ type: "snapshot" });
		}
	}

	private emit(event: TfaSessionRuntimeEvent): void {
		for (const listener of this.listeners) {
			listener(event);
		}
	}

	get agentSession(): AgentSession {
		return this.session;
	}

	snapshot(): SessionSnapshot {
		this.revision += 1;
		return buildSessionSnapshot(this.session, this.revision);
	}

	getPhase(): SessionPhase {
		return getSessionPhase(this.session);
	}

	async prompt(input: PromptInput): Promise<void> {
		if (this.session.isStreaming) {
			throw new TfaServerError("busy", "Session is busy");
		}
		await this.session.prompt(input.text);
	}

	async steer(input: SteerInput): Promise<void> {
		if (!this.session.isStreaming) {
			throw new TfaServerError("invalid_request", "Session is not streaming; steer requires an active turn");
		}
		await this.session.steer(input.text);
	}

	async abort(): Promise<void> {
		if (this.session.isIdle) {
			throw new TfaServerError("invalid_request", "Session is idle; nothing to abort");
		}
		await this.session.abort();
	}

	async setModel(model: ModelRef): Promise<void> {
		const found = await this.findModel(model.provider, model.id);
		if (!found) {
			throw new TfaServerError("invalid_request", `Unknown model: ${model.provider}/${model.id}`);
		}
		await this.session.setModel(found);
	}

	async setThinking(thinkingLevel: ThinkingLevel): Promise<void> {
		this.session.setThinkingLevel(thinkingLevel);
	}

	subscribe(listener: (event: TfaSessionRuntimeEvent) => void): () => void {
		this.listeners.add(listener);
		return () => {
			this.listeners.delete(listener);
		};
	}

	async dispose(): Promise<void> {
		this.unsubscribe();
		this.session.dispose();
	}

	private async findModel(provider: string, modelId: string): Promise<Model<any> | undefined> {
		const models = await this.modelRuntime.getAvailable(undefined, { signal: AbortSignal.timeout(15_000) });
		return models.find((model) => model.provider === provider && model.id === modelId);
	}
}

/**
 * 会话服务:实现 TfaServerService 契约,内部用 createAgentSession 驱动。
 * 同一 ModelRuntime 在服务与所有会话间共享(faux 注册只需一次)。
 */
function toAuthUserView(record: UserRecord): AuthUserView {
	return { id: record.id, username: record.username, role: record.role, status: record.status };
}

export class ShellSessionService implements TfaServerService {
	private readonly cwd: string;
	private readonly agentDir: string;
	private readonly customTools: ToolDefinition[] | undefined;
	private readonly toolAllowlist: string[] | undefined;
	private readonly defaultModel: Model<any> | undefined;
	private readonly fauxSetup: FauxSetup | undefined;
	private readonly sessionDir: string;
	private readonly projects: ProjectStore;
	private readonly groupsService: GroupService;
	private readonly skillsService: SkillService;
	private readonly datasetsService: DatasetService;
	private readonly live = new Map<string, LiveShellSession>();
	private readonly rpcSessions = new Set<ProxyRuntime>();
	private readonly sessionExecutor: SessionExecutor;
	private readonly filesService: ProjectFileService;
	private readonly approvalsService = new ApprovalService();
	private readonly agentsService: AgentService;
	private readonly approvalResponders = new Map<string, { runtime: ProxyRuntime; rpcId: string }>();
	private healthCheckTimer: ReturnType<typeof setInterval> | undefined;
	private modelRuntimePromise: Promise<ModelRuntime> | undefined;
	private fauxApplied = false;

	constructor(options: ShellSessionServiceOptions) {
		this.cwd = resolve(options.cwd);
		this.agentDir = options.agentDir ? resolve(options.agentDir) : getAgentDir();
		this.customTools = options.customTools;
		this.toolAllowlist = options.toolAllowlist;
		// faux 模式下会话必须用 faux 模型,否则会落到 settings 默认模型导致认证失败
		this.defaultModel = options.defaultModel ?? options.fauxSetup?.defaultModel;
		this.fauxSetup = options.fauxSetup;
		this.sessionDir = join(this.agentDir, "sessions");
		this.projects = new ProjectStore(this.agentDir);
		this.groupsService = new GroupService(this.projects.getRepos());
		this.skillsService = new SkillService({
			repos: this.projects.getRepos(),
			storageRoot: join(this.agentDir, "platform", "skills"),
		});
		this.datasetsService = new DatasetService({
			repos: this.projects.getRepos(),
			storageRoot: join(this.agentDir, "platform", "datasets"),
			...(options.datasetMaxFileBytes !== undefined ? { maxFileBytes: options.datasetMaxFileBytes } : {}),
			...(options.datasetMaxUserBytes !== undefined ? { maxUserBytes: options.datasetMaxUserBytes } : {}),
		});
		this.sessionExecutor =
			options.executor ??
			new DockerExecutor(loadExecutorConfig(process.env, join(this.agentDir, "platform", "skills")));
		this.filesService = new ProjectFileService({
			projects: this.projects,
			workspacesRoot: join(this.agentDir, "workspaces"),
		});
		this.agentsService = new AgentService(this.projects.getRepos());
	}

	async listSessions(): Promise<SessionSummary[]> {
		// live 会话优先(可能尚未落盘),磁盘会话合并(去重,避免快照 revision 副作用)
		const summaries = new Map<string, SessionSummary>();
		for (const [id, live] of this.live) {
			const session = live.agentSession;
			const header = session.sessionManager.getHeader();
			const createdAt = header ? parseHeaderTimestamp(header.timestamp) : Date.now();
			summaries.set(id, {
				id,
				name: session.sessionName,
				cwd: session.sessionManager.getCwd(),
				createdAt,
				updatedAt: lastMessageTimestamp(session) ?? Date.now(),
				phase: live.getPhase(),
				model: {
					provider: session.model?.provider ?? "unknown",
					id: session.model?.id ?? "unknown",
				},
				thinkingLevel: session.thinkingLevel,
				attached: false,
				locked: false,
			});
		}
		const sessions = await SessionManager.listAll(this.sessionDir);
		for (const info of sessions) {
			if (summaries.has(info.id)) continue;
			summaries.set(info.id, {
				id: info.id,
				name: info.name,
				cwd: info.cwd,
				createdAt: info.created.getTime(),
				updatedAt: info.modified.getTime(),
				phase: "idle",
				model: { provider: "unknown", id: "unknown" },
				thinkingLevel: "medium",
				attached: false,
				locked: false,
			});
		}
		return [...summaries.values()];
	}

	async listModels(): Promise<ModelMetadata[]> {
		const modelRuntime = await this.getModelRuntime();
		const models = await modelRuntime.getAvailable(undefined, { signal: AbortSignal.timeout(15_000) });
		return models.map((model) =>
			toProtocolModelMetadata(model, modelRuntime.getProviderAuthStatus(model.provider).configured),
		);
	}

	/** Persist an API key to auth.json and refresh the model runtime snapshot. */
	async configureApiKey(provider: string, apiKey: string): Promise<void> {
		const cleanProvider = provider.trim();
		const cleanKey = apiKey.trim();
		if (!cleanProvider || !cleanKey) throw new TfaServerError("invalid_request", "provider 和 apiKey 不能为空");
		await this.writeCredential(cleanProvider, { type: "api_key", key: cleanKey });
		const modelRuntime = await this.getModelRuntime();
		await modelRuntime.setRuntimeApiKey(cleanProvider, cleanKey);
	}

	/** Remove a stored API key from auth.json and refresh the model runtime snapshot. */
	async removeApiKey(provider: string): Promise<void> {
		const cleanProvider = provider.trim();
		if (!cleanProvider) throw new TfaServerError("invalid_request", "provider 不能为空");
		await this.deleteCredential(cleanProvider);
		const modelRuntime = await this.getModelRuntime();
		await modelRuntime.removeRuntimeApiKey(cleanProvider);
	}

	/** Runtime configuration summary used by the Web settings panel. */
	async getConfig(): Promise<ShellConfigInfo> {
		const modelRuntime = await this.getModelRuntime();
		const models = await modelRuntime.getAvailable(undefined, { signal: AbortSignal.timeout(15_000) });
		const providers = new Map<string, { configured: boolean; source: string }>();
		for (const model of models) {
			const status = modelRuntime.getProviderAuthStatus(model.provider);
			providers.set(model.provider, { configured: status.configured, source: status.source ?? "none" });
		}
		return {
			agentDir: this.agentDir,
			cwd: this.cwd,
			authPath: join(this.agentDir, "auth.json"),
			modelsPath: join(this.agentDir, "models.json"),
			providers: [...providers.entries()].map(([provider, info]) => ({ provider, ...info })),
		};
	}

	/** List registered financial/custom tools for the Web tool panel. */
	async listTools(): Promise<ToolInfo[]> {
		return (this.customTools ?? []).map((tool) => ({
			name: tool.name,
			label: tool.label,
			description: tool.description,
		}));
	}

	/** 项目实体:可见范围按用户过滤(owner ∪ 挂组 active 成员 ∪ admin)。 */
	async listProjects(user: AuthUserView): Promise<ProjectView[]> {
		return this.projects.listForUser(user);
	}

	async createProject(user: AuthUserView, input: CreateProjectInput): Promise<ProjectView> {
		return this.projects.createForUser(user, input);
	}

	async updateProject(user: AuthUserView, id: string, input: UpdateProjectInput): Promise<ProjectView> {
		return this.projects.updateForUser(user, id, input);
	}

	async deleteProject(user: AuthUserView, id: string): Promise<void> {
		await this.projects.deleteForUser(user, id);
	}

	/** 移动会话到项目(projectId 为 null 表示无项目)。只改归属,不改会话 cwd;目标项目须对用户可见。 */
	async moveSessionToProject(user: AuthUserView, sessionId: string, projectId: string | null): Promise<void> {
		if (projectId !== null) {
			const project = await this.projects.getForUser(user, projectId);
			if (!project) throw new TfaServerError("not_found", `项目不存在: ${projectId}`);
		}
		await this.projects.moveSession(sessionId, projectId);
	}

	/** Rename a session. */
	async renameSession(sessionId: string, name: string): Promise<void> {
		const clean = name.replace(/[\r\n]+/g, " ").trim();
		if (!clean) throw new TfaServerError("invalid_request", "name 不能为空");
		const runtime = await this.openSession(sessionId);
		runtime.agentSession.setSessionName(clean);
	}

	/** Export a session as JSONL text. Live sessions serialize from memory; persisted sessions read from disk. */
	async exportSession(sessionId: string): Promise<string> {
		const live = this.live.get(sessionId);
		if (live) {
			const sessionManager = live.agentSession.sessionManager;
			const header = sessionManager.getHeader();
			const lines = [header, ...sessionManager.getEntries()].map((entry) => JSON.stringify(entry));
			return lines.join("\n");
		}
		const sessions = await SessionManager.listAll(this.sessionDir);
		const info = sessions.find((session) => session.id === sessionId);
		if (!info) throw new TfaServerError("not_found", `会话不存在: ${sessionId}`);
		return readFile(info.path, "utf8");
	}

	private async writeCredential(provider: string, credential: { type: "api_key"; key: string }): Promise<void> {
		const authPath = join(this.agentDir, "auth.json");
		await mkdir(dirname(authPath), { recursive: true });
		let data: Record<string, unknown> = {};
		try {
			data = JSON.parse(await readFile(authPath, "utf8")) as Record<string, unknown>;
		} catch {
			// 文件不存在或不可解析时从空对象开始
		}
		data[provider] = credential;
		await writeFile(authPath, JSON.stringify(data, null, 2), "utf8");
	}

	private async deleteCredential(provider: string): Promise<void> {
		const authPath = join(this.agentDir, "auth.json");
		let data: Record<string, unknown> = {};
		try {
			data = JSON.parse(await readFile(authPath, "utf8")) as Record<string, unknown>;
		} catch {
			// 文件不存在或不可解析时视为空
		}
		delete data[provider];
		await writeFile(authPath, JSON.stringify(data, null, 2), "utf8");
	}

	async createSession(
		options: CreateSessionOptions & { projectId?: string; userId?: string | null },
	): Promise<TfaSessionRuntime> {
		const existing = this.live.get(options.id);
		if (existing) return existing;
		// 项目归属:项目绑定文件夹且未显式指定 cwd 时,用项目文件夹作为会话工作目录
		let cwd = options.cwd ? resolve(options.cwd) : this.cwd;
		if (options.projectId) {
			const project = await this.projects.get(options.projectId);
			if (!project) throw new TfaServerError("not_found", `项目不存在: ${options.projectId}`);
			// 项目会话工作目录 = 平台托管工作空间(与 T12 文件 API 一致),Agent 在此读写项目文件
			const workspace = join(this.agentDir, "workspaces", options.projectId);
			await this.files.ensureWorkspace(options.projectId);
			if (!options.cwd) cwd = resolve(workspace);
		}
		const modelRuntime = await this.getModelRuntime();
		const sessionManager = SessionManager.create(cwd, this.sessionDir, { id: options.id });
		const { session } = await createAgentSession({
			cwd,
			agentDir: this.agentDir,
			sessionManager,
			modelRuntime,
			...(options.model ? { model: await this.resolveModel(modelRuntime, options.model) } : {}),
			...(this.defaultModel ? { model: this.defaultModel } : {}),
			...(options.thinkingLevel ? { thinkingLevel: options.thinkingLevel } : {}),
			...(this.toolAllowlist ? { tools: this.toolAllowlist } : {}),
			...(this.customTools ? { customTools: this.customTools } : {}),
		});
		if (options.name) {
			sessionManager.appendSessionInfo(options.name);
		}
		const live = new LiveShellSession(session, modelRuntime);
		this.live.set(options.id, live);
		this.recordSessionRunning(options.id, {
			projectId: options.projectId ?? null,
			cwd,
			userId: options.userId ?? null,
			permissionMode: options.projectId ? "request_approval" : "full_access",
		});
		if (options.projectId) {
			await this.projects.moveSession(options.id, options.projectId);
		}
		return live;
	}

	async openSession(sessionId: string): Promise<LiveShellSession> {
		const existing = this.live.get(sessionId);
		if (existing) return existing;
		const sessions = await SessionManager.listAll(this.sessionDir);
		const info = sessions.find((session) => session.id === sessionId);
		if (!info) {
			throw new SessionNotFoundError(`Session was not found: ${sessionId}`);
		}
		const cwd = info.cwd || this.cwd;
		const modelRuntime = await this.getModelRuntime();
		const sessionManager = SessionManager.open(info.path, this.sessionDir, cwd);
		const { session } = await createAgentSession({
			cwd,
			agentDir: this.agentDir,
			sessionManager,
			modelRuntime,
			...(this.defaultModel ? { model: this.defaultModel } : {}),
			...(this.toolAllowlist ? { tools: this.toolAllowlist } : {}),
			...(this.customTools ? { customTools: this.customTools } : {}),
		});
		const live = new LiveShellSession(session, modelRuntime);
		this.live.set(sessionId, live);
		this.recordSessionRunning(sessionId, { projectId: null, cwd });
		return live;
	}

	/** 壳扩展:删除会话(销毁 live 实例并删除持久化文件)。 */
	async deleteSession(sessionId: string): Promise<void> {
		const live = this.live.get(sessionId);
		if (live) {
			await live.dispose();
			this.live.delete(sessionId);
		}
		const sessions = await SessionManager.listAll(this.sessionDir);
		const info = sessions.find((session) => session.id === sessionId);
		if (info) {
			await rm(info.path, { force: true });
		}
		await this.projects.removeSession(sessionId);
		this.authRepos().sessionUpdate(sessionId, { status: "ended", endedAt: Date.now() });
	}

	/** 壳扩展:关闭服务,销毁所有 live 会话并反注册 faux。 */
	/** 会话权限模式:仅 owner 可读。 */
	getSessionPermissionMode(user: AuthUserView, sessionId: string): SessionPermissionMode {
		return this.getOwnedSession(user, sessionId).permissionMode;
	}

	/** 切换会话权限模式:仅 owner 可切换;校验取值。 */
	setSessionPermissionMode(user: AuthUserView, sessionId: string, mode: SessionPermissionMode): SessionPermissionMode {
		if (mode !== "request_approval" && mode !== "full_access") {
			throw new TfaServerError("invalid_request", "permissionMode 必须是 request_approval 或 full_access");
		}
		this.getOwnedSession(user, sessionId);
		this.authRepos().sessionUpdate(sessionId, { permissionMode: mode });
		return mode;
	}

	/** 会话待批准请求列表(仅 owner)。 */
	listSessionApprovals(user: AuthUserView, sessionId: string): ApprovalRequest[] {
		this.getOwnedSession(user, sessionId);
		return this.approvals.listPending(sessionId);
	}

	/** 审批决策:仅 owner;批准/拒绝后回传 RPC 侧的 agent(若桥接支持)。 */
	async decideApproval(
		user: AuthUserView,
		sessionId: string,
		approvalId: string,
		approved: boolean,
	): Promise<ApprovalRequest> {
		this.getOwnedSession(user, sessionId);
		const decision = approved ? this.approvals.approve(approvalId) : this.approvals.reject(approvalId);
		const responder = this.approvalResponders.get(approvalId);
		if (responder) {
			this.approvalResponders.delete(approvalId);
			await responder.runtime.decideApproval(responder.rpcId, approved).catch(() => {
				// agent 进程可能已退出
			});
		}
		return decision;
	}

	/** 桥接会话:经 RPC 子进程跑 tfa(宿主直连;容器化由 executor 承接)。 */
	/** 项目会话上下文:该项目会话可见的技能与数据集(私有 ∪ 挂组则该组库 ∪ 公开)。 */
	/** 管理端:全部平台会话 + 容器状态 + 用户/项目名。 */
	async listAdminSessions(): Promise<AdminSessionView[]> {
		const repos = this.authRepos();
		const views: AdminSessionView[] = [];
		for (const record of repos.sessionList()) {
			const user = record.userId ? repos.userGetById(record.userId) : undefined;
			const project = record.projectId ? repos.projectGet(record.projectId) : undefined;
			let containerId = record.containerId;
			let containerState: string | null = null;
			const execStatus = await this.executor.inspect(record.id).catch(() => undefined);
			if (execStatus && execStatus.state !== "missing") {
				containerId = execStatus.containerId ?? containerId;
				containerState = execStatus.state;
			}
			if (containerState === null && this.live.has(record.id)) containerState = "running";
			views.push({
				id: record.id,
				userId: record.userId,
				username: user?.username ?? null,
				projectId: record.projectId,
				projectName: project?.name ?? null,
				containerId,
				containerState,
				status: record.status,
				cwd: record.cwd,
				startedAt: record.startedAt,
				endedAt: record.endedAt,
			});
		}
		return views;
	}

	/** 管理端:终止会话(停容器 + 销毁 live/rpc 会话 + 置 ended)。 */
	async terminateAdminSession(id: string): Promise<void> {
		const repos = this.authRepos();
		if (!repos.sessionGet(id)) throw new TfaServerError("not_found", `会话不存在: ${id}`);
		await this.executor.stop(id).catch(() => {
			// 容器可能不存在
		});
		const live = this.live.get(id);
		if (live) {
			await live.dispose().catch(() => {
				// 已关闭
			});
			this.live.delete(id);
		}
		const rpc = [...this.rpcSessions].find((runtime) => runtime.id === id);
		if (rpc) {
			await rpc.dispose().catch(() => {
				// 进程可能已退出
			});
			this.rpcSessions.delete(rpc);
		}
		repos.sessionUpdate(id, { status: "ended", endedAt: Date.now() });
	}

	/** 项目会话上下文:该项目会话可见的技能与数据集(私有 ∪ 挂组则该组库 ∪ 公开)。 */
	async getProjectSessionContext(
		user: AuthUserView,
		projectId: string,
	): Promise<{ skills: SkillView[]; datasets: DatasetView[] }> {
		const project = await this.projects.getForUser(user, projectId);
		if (!project) throw new TfaServerError("not_found", `项目不存在: ${projectId}`);
		const skills = this.skills.listForProject(user, project.scope);
		const datasets = this.datasets.listForProject(user, project.scope);
		return { skills, datasets };
	}

	async createRpcSession(options: RpcSessionCreateOptions): Promise<TfaSessionRuntime> {
		const client = createRpcSessionClient({
			cliPath: options.cliPath,
			cwd: options.cwd,
			env: options.env,
			...(options.provider ? { provider: options.provider } : {}),
			...(options.model ? { model: options.model } : {}),
			...(options.skills && options.skills.length > 0 ? { args: buildSkillCliArgs(options.skills) } : {}),
		});
		await client.start();
		const runtime = new ProxyRuntime({
			client,
			sessionId: options.sessionId,
			cwd: options.cwd,
			onApprovalRequest: (request) => {
				const approval = this.approvals.createRequest(options.sessionId, {
					tool: request.tool,
					target: request.target,
					risk: request.risk,
				});
				this.approvalResponders.set(approval.id, { runtime, rpcId: request.id });
			},
		});
		this.rpcSessions.add(runtime);
		this.recordSessionRunning(options.sessionId, { projectId: null, cwd: options.cwd });
		return runtime;
	}

	async close(): Promise<void> {
		const lives = [...this.live.values()];
		this.live.clear();
		for (const live of lives) {
			await live.dispose();
		}
		this.fauxSetup?.unregister();
		for (const runtime of [...this.rpcSessions]) {
			await runtime.dispose().catch(() => {
				// 进程可能已退出
			});
		}
		this.rpcSessions.clear();
		if (this.healthCheckTimer) {
			clearInterval(this.healthCheckTimer);
			this.healthCheckTimer = undefined;
		}
		const repos = this.authRepos();
		for (const session of repos.sessionList()) {
			if (session.status === "running" || session.status === "created") {
				repos.sessionUpdate(session.id, { status: "ended", endedAt: Date.now() });
			}
		}
		this.projects.close();
	}

	/** 课程/课题组服务(组 CRUD + 成员管理,ACL 在服务层强制)。 */
	/** 技能库服务(安装/可见性/审核/下架)。 */
	/** 数据集库服务(上传/可见性/下架/下载)。 */
	/** 会话执行器(容器状态/终止)。 */
	get executor(): SessionExecutor {
		return this.sessionExecutor;
	}

	/** 项目文件工作空间服务(文件树/读写/上传/下载/删除)。 */
	get files(): ProjectFileService {
		return this.filesService;
	}

	/** 请求批准状态机(request_approval 模式)。 */
	get approvals(): ApprovalService {
		return this.approvalsService;
	}

	/** 智能体广场服务(外部网页目录资源)。 */
	get agents(): AgentService {
		return this.agentsService;
	}

	/** 启动智能体健康检查定时任务(周期默认 5 分钟);close 时清理。 */
	startAgentHealthCheck(intervalMs = 5 * 60 * 1000): void {
		if (this.healthCheckTimer) return;
		this.healthCheckTimer = setInterval(() => {
			void this.agents.runHealthCheck().catch(() => {
				// 单轮检查失败不中断定时任务
			});
		}, intervalMs);
	}

	/** 平台仓储(管理端/测试用)。 */
	getRepos(): PlatformRepos {
		return this.authRepos();
	}

	get datasets(): DatasetService {
		return this.datasetsService;
	}

	get skills(): SkillService {
		return this.skillsService;
	}

	get groups(): GroupService {
		return this.groupsService;
	}

	// ---- 认证与账号 ----

	private authRepos() {
		return this.projects.getRepos();
	}

	private recordSessionRunning(
		id: string,
		options: {
			projectId: string | null;
			cwd: string;
			userId?: string | null;
			permissionMode?: SessionPermissionMode;
		},
	): void {
		const repos = this.authRepos();
		const existing = repos.sessionGet(id);
		if (existing) {
			repos.sessionUpdate(id, { status: "running", endedAt: null });
		} else {
			repos.sessionCreate({
				id,
				projectId: options.projectId,
				userId: options.userId ?? null,
				cwd: options.cwd,
				status: "running",
				permissionMode: options.permissionMode ?? "full_access",
				startedAt: Date.now(),
			});
		}
	}

	/** 会话归属校验:仅会话 owner(或未绑定用户的旧会话)可操作。 */
	private getOwnedSession(user: AuthUserView, sessionId: string): PlatformSessionRecord {
		const session = this.authRepos().sessionGet(sessionId);
		if (!session) throw new TfaServerError("not_found", `会话不存在: ${sessionId}`);
		if (session.userId && session.userId !== user.id) {
			throw new AuthError("forbidden", "只能操作自己的会话");
		}
		return session;
	}

	async login(username: string, password: string): Promise<{ token: string; user: AuthUserView; expiresAt: number }> {
		const repos = this.authRepos();
		const user = repos.userGetByUsername(username.trim());
		if (!user || !verifyPassword(password, user.passwordHash)) {
			throw new AuthError("unauthorized", "用户名或密码错误");
		}
		if (user.status === "disabled") {
			throw new AuthError("unauthorized", "账号已禁用");
		}
		const token = generateToken();
		const expiresAt = Date.now() + AUTH_SESSION_TTL_MS;
		repos.authSessionCreate({ tokenHash: hashToken(token), userId: user.id, expiresAt });
		return { token, user: toAuthUserView(user), expiresAt };
	}

	async resolveAuth(token: string): Promise<AuthUserView | undefined> {
		const repos = this.authRepos();
		const session = repos.authSessionGet(hashToken(token));
		if (!session) return undefined;
		if (session.expiresAt < Date.now()) {
			repos.authSessionDelete(session.tokenHash);
			return undefined;
		}
		const user = repos.userGetById(session.userId);
		if (!user || user.status === "disabled") return undefined;
		return toAuthUserView(user);
	}

	async logout(token: string): Promise<void> {
		this.authRepos().authSessionDelete(hashToken(token));
	}

	async createUser(input: { username: string; password: string; role: UserRole }): Promise<AuthUserView> {
		const repos = this.authRepos();
		const username = input.username.trim();
		if (!username) throw new TfaServerError("invalid_request", "用户名不能为空");
		if (!input.password || input.password.length < 6) throw new TfaServerError("invalid_request", "密码至少 6 位");
		if (input.role !== "admin" && input.role !== "teacher" && input.role !== "student") {
			throw new TfaServerError("invalid_request", "角色必须是 admin/teacher/student");
		}
		if (repos.userGetByUsername(username)) throw new TfaServerError("invalid_request", "用户名已存在");
		const record = repos.userCreate({ username, passwordHash: hashPassword(input.password), role: input.role });
		return toAuthUserView(record);
	}

	async listUsers(): Promise<AuthUserView[]> {
		return this.authRepos().userList().map(toAuthUserView);
	}

	async setUserStatus(id: string, status: UserStatus): Promise<void> {
		const repos = this.authRepos();
		if (!repos.userGetById(id)) throw new TfaServerError("not_found", `用户不存在: ${id}`);
		repos.userSetStatus(id, status);
	}

	async resetPassword(id: string, newPassword: string): Promise<void> {
		if (!newPassword || newPassword.length < 6) throw new TfaServerError("invalid_request", "密码至少 6 位");
		const repos = this.authRepos();
		if (!repos.userGetById(id)) throw new TfaServerError("not_found", `用户不存在: ${id}`);
		repos.userSetPassword(id, hashPassword(newPassword));
	}

	private getModelRuntime(): Promise<ModelRuntime> {
		if (!this.modelRuntimePromise) {
			this.modelRuntimePromise = ModelRuntime.create({
				authPath: join(this.agentDir, "auth.json"),
				modelsPath: join(this.agentDir, "models.json"),
			});
		}
		return this.modelRuntimePromise.then((modelRuntime) => {
			this.applyFaux(modelRuntime);
			return modelRuntime;
		});
	}

	private applyFaux(modelRuntime: ModelRuntime): void {
		if (!this.fauxSetup || this.fauxApplied) return;
		this.fauxApplied = true;
		this.fauxSetup.applyTo(modelRuntime);
	}

	private async resolveModel(modelRuntime: ModelRuntime, model: ModelRef): Promise<Model<any>> {
		const models = await modelRuntime.getAvailable(undefined, { signal: AbortSignal.timeout(15_000) });
		const found = models.find((candidate) => candidate.provider === model.provider && candidate.id === model.id);
		if (!found) {
			throw new TfaServerError("invalid_request", `Unknown model: ${model.provider}/${model.id}`);
		}
		return found;
	}
}
