/**
 * 协议 DTO 类型镜像:与 @earendil-works/tfa-protocol 的 JSON 结构一致。
 * 前端不直接依赖 protocol 包(cbor 部分有 node 依赖面),手工镜像保持最小依赖。
 */

export type JsonValue = null | boolean | number | string | JsonValue[] | { [key: string]: JsonValue };

export type ThinkingLevel = "off" | "minimal" | "low" | "medium" | "high" | "xhigh" | "max";

export type SessionPhase = "idle" | "turn" | "compaction" | "branch_summary" | "retry";

export interface ModelRef {
	provider: string;
	id: string;
}

export interface ModelMetadata {
	provider: string;
	id: string;
	name: string;
	api: string;
	reasoning: boolean;
	input: Array<"text" | "image">;
	contextWindow: number;
	maxTokens: number;
	cost: { input: number; output: number; cacheRead: number; cacheWrite: number };
	supportedThinkingLevels: ThinkingLevel[];
	authenticated: boolean;
}

export interface TextContent {
	type: "text";
	text: string;
}

export interface ThinkingContent {
	type: "thinking";
	thinking: string;
	redacted?: boolean;
}

export interface ImageContent {
	type: "image";
	data: string;
	mimeType: string;
}

export interface ToolCallContent {
	type: "toolCall";
	toolCallId: string;
	toolName: string;
	input: JsonValue;
}

export interface Usage {
	input: number;
	output: number;
	cacheRead: number;
	cacheWrite: number;
	reasoning?: number;
	totalTokens: number;
	cost: { input: number; output: number; cacheRead: number; cacheWrite: number; total: number };
}

export interface UserTranscriptItem {
	id: string;
	role: "user";
	content: Array<TextContent | ImageContent>;
	timestamp: number;
}

export interface AssistantTranscriptItem {
	id: string;
	role: "assistant";
	content: Array<TextContent | ThinkingContent | ToolCallContent>;
	model: ModelRef;
	responseModel?: string;
	usage?: Usage;
	timestamp: number;
	status: "streaming" | "complete" | "error" | "aborted";
	stopReason: "stop" | "length" | "toolUse" | "error" | "aborted";
	errorMessage?: string;
}

export interface ToolTranscriptItem {
	id: string;
	role: "tool";
	toolCallId: string;
	toolName: string;
	input: JsonValue;
	content: Array<TextContent | ImageContent>;
	details?: JsonValue;
	usage?: Usage;
	timestamp: number;
	status: "running" | "complete" | "error";
	isError: boolean;
}

export type TranscriptItem = UserTranscriptItem | AssistantTranscriptItem | ToolTranscriptItem;

export type TranscriptProgress =
	| { type: "item_started"; item: TranscriptItem }
	| {
			type: "assistant_delta";
			messageId: string;
			contentIndex: number;
			kind: "text" | "thinking" | "toolCall";
			delta: string;
	  }
	| { type: "item_updated"; item: AssistantTranscriptItem | ToolTranscriptItem }
	| { type: "item_finished"; item: AssistantTranscriptItem | ToolTranscriptItem };

export interface SessionSummary {
	id: string;
	name?: string;
	cwd: string;
	createdAt: number;
	updatedAt: number;
	phase: SessionPhase;
	model: ModelRef;
	thinkingLevel: ThinkingLevel;
	attached: boolean;
	locked: boolean;
}

export interface SessionSnapshot extends SessionSummary {
	revision: number;
	transcript: TranscriptItem[];
	queuedSteer: UserTranscriptItem[];
	queuedSteerCount: number;
}

export type SessionEvent =
	| { type: "snapshot" }
	| { type: "progress"; progress: TranscriptProgress }
	| { type: "error"; error: { code: string; message: string } };

export interface ProviderConfigInfo {
	provider: string;
	configured: boolean;
	source: string;
}

export interface ShellConfig {
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

/** 项目作用域:私有或挂靠某课程/课题组。 */
export type ProjectScope = { type: "private" } | { type: "group"; groupId: string; groupName: string };

/** 项目分类实体:可绑定可选本地文件夹,会话归属通过 sessionIds 记录;多用户含 owner 与作用域。 */
export interface ProjectInfo {
	id: string;
	name: string;
	folder?: string;
	ownerId: string | null;
	scope: ProjectScope;
	createdAt: number;
	updatedAt: number;
	sessionIds: string[];
}

export interface UserInfo {
	id: string;
	username: string;
	role: "admin" | "teacher" | "student";
	status: "active" | "disabled";
}

export interface GroupView {
	id: string;
	name: string;
	type: "course" | "research";
	status: "active" | "archived";
	ownerId: string;
	memberCount: number;
	myRole: "owner" | "member" | null;
}

export type ResourceVisibility = "private" | "group" | "public";
export type ReviewState = "none" | "pending" | "approved" | "rejected";

export interface SkillView {
	id: string;
	name: string;
	repoUrl: string | null;
	version: string | null;
	visibility: ResourceVisibility;
	groupId: string | null;
	groupName?: string;
	ownerId: string | null;
	status: string;
	reviewState: ReviewState;
	createdAt: number;
}

export interface DatasetView {
	id: string;
	name: string;
	sizeBytes: number;
	visibility: ResourceVisibility;
	groupId: string | null;
	groupName?: string;
	ownerId: string | null;
	status: string;
	reviewState: ReviewState;
	createdAt: number;
}

/** 项目会话上下文:该项目会话可见的技能与数据集。 */
export interface SessionContext {
	skills: SkillView[];
	datasets: DatasetView[];
}

/** 项目文件工作空间条目(目录树)。 */
export interface ProjectFileEntry {
	name: string;
	/** 相对项目工作区的路径(正斜杠)。 */
	path: string;
	type: "file" | "dir";
	size: number | null;
	mtime: number | null;
}

export interface AgentView {
	id: string;
	name: string;
	description: string | null;
	url: string;
	visibility: ResourceVisibility;
	groupId: string | null;
	groupName?: string;
	ownerId: string | null;
	status: string;
	availability: "online" | "unavailable" | "unknown";
	lastCheckedAt: number | null;
	createdAt: number;
	updatedAt: number;
}
export type ApprovalStatus = "pending" | "approved" | "rejected" | "expired";

export interface ApprovalRequestView {
	id: string;
	sessionId: string;
	tool: string;
	target: string;
	risk: string;
	status: ApprovalStatus;
	createdAt: number;
	expiresAt: number;
	decidedAt: number | null;
}
export interface ProjectFileMeta {
	path: string;
	name: string;
	size: number;
	mimeType: string;
}

/** 组内成员视图。 */
export interface GroupMemberView {
	userId: string;
	username: string;
	role: "owner" | "member";
}

/** 课程/课题组详情(含成员)。 */
export interface GroupDetail extends GroupView {
	members: GroupMemberView[];
}

/** 平台会话容器状态。 */
export type AdminSessionState = "created" | "running" | "ended" | "error";
export type ContainerState = "running" | "exited" | "missing" | "error" | "expired" | null;

/** 管理端会话监控视图。 */
export interface AdminSessionView {
	id: string;
	userId: string | null;
	username: string | null;
	projectId: string | null;
	projectName: string | null;
	containerId: string | null;
	containerState: ContainerState;
	status: AdminSessionState;
	cwd: string | null;
	startedAt: number | null;
	endedAt: number | null;
}
