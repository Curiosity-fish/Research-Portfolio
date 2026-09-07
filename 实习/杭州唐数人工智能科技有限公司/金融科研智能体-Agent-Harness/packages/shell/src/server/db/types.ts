/** 平台 SQLite 数据库的语句能力（镜像 @earendil-works/tfa-session-backend-sqlite-node 的抽象，shell 不依赖该包）。 */
export interface PlatformStatement {
	run(...params: unknown[]): PlatformRunResult;
	get<TRow extends object>(...params: unknown[]): TRow | undefined;
	all<TRow extends object>(...params: unknown[]): TRow[];
}

export interface PlatformRunResult {
	changes: number;
	lastInsertRowid?: number;
}

export interface PlatformDatabase {
	exec(sql: string): void;
	prepare(sql: string): PlatformStatement;
	/** 同步写事务；回调不得返回 Promise。 */
	transaction<T>(fn: () => T): T;
	close(): void;
}

export type UserRole = "admin" | "teacher" | "student";
export type UserStatus = "active" | "disabled";
export type GroupType = "course" | "research";
export type GroupStatus = "active" | "archived";
export type GroupMemberRole = "owner" | "member";
export type ProjectScopeType = "private" | "group";
export type ResourceVisibility = "private" | "group" | "public";
export type ResourceStatus = "active" | "removed";
export type SkillReviewState = "none" | "pending" | "approved" | "rejected";
export type PlatformSessionStatus = "created" | "running" | "ended" | "error";
export type SessionPermissionMode = "request_approval" | "full_access";
export type AgentAvailability = "online" | "unavailable" | "unknown";

export interface UserRecord {
	id: string;
	username: string;
	passwordHash: string;
	role: UserRole;
	status: UserStatus;
	createdAt: number;
}

export interface CourseGroupRecord {
	id: string;
	name: string;
	type: GroupType;
	ownerId: string;
	status: GroupStatus;
	createdAt: number;
}

export interface GroupMemberRecord {
	groupId: string;
	userId: string;
	role: GroupMemberRole;
}

export interface ProjectRecord {
	id: string;
	name: string;
	folder: string | null;
	ownerId: string | null;
	scopeType: ProjectScopeType;
	groupId: string | null;
	createdAt: number;
	updatedAt: number;
}

export interface SkillRecord {
	id: string;
	name: string;
	repoUrl: string | null;
	version: string | null;
	visibility: ResourceVisibility;
	groupId: string | null;
	ownerId: string | null;
	path: string;
	status: ResourceStatus;
	reviewState: SkillReviewState;
	createdAt: number;
}

export interface DatasetRecord {
	id: string;
	name: string;
	sizeBytes: number;
	visibility: ResourceVisibility;
	groupId: string | null;
	ownerId: string | null;
	path: string;
	status: ResourceStatus;
	reviewState: SkillReviewState;
	createdAt: number;
}

export interface PlatformSessionRecord {
	id: string;
	projectId: string | null;
	userId: string | null;
	containerId: string | null;
	status: PlatformSessionStatus;
	permissionMode: SessionPermissionMode;
	cwd: string | null;
	startedAt: number | null;
	endedAt: number | null;
}

export interface AgentListingRecord {
	id: string;
	name: string;
	description: string | null;
	url: string;
	visibility: ResourceVisibility;
	groupId: string | null;
	ownerId: string | null;
	status: ResourceStatus;
	availability: AgentAvailability;
	lastCheckedAt: number | null;
	createdAt: number;
	updatedAt: number;
}

export interface AuthSessionRecord {
	tokenHash: string;
	userId: string;
	createdAt: number;
	expiresAt: number;
}
