import { randomUUID } from "node:crypto";
import type {
	AgentAvailability,
	AgentListingRecord,
	AuthSessionRecord,
	CourseGroupRecord,
	DatasetRecord,
	GroupMemberRecord,
	GroupMemberRole,
	GroupStatus,
	GroupType,
	PlatformDatabase,
	PlatformSessionRecord,
	PlatformSessionStatus,
	ProjectRecord,
	ProjectScopeType,
	ResourceStatus,
	ResourceVisibility,
	SessionPermissionMode,
	SkillRecord,
	SkillReviewState,
	UserRecord,
	UserRole,
	UserStatus,
} from "./types.ts";

function now(): number {
	return Date.now();
}

function rowToProject(row: {
	id: string;
	name: string;
	folder: string | null;
	owner_id: string | null;
	scope_type: ProjectScopeType;
	group_id: string | null;
	created_at: number;
	updated_at: number;
}): ProjectRecord {
	return {
		id: row.id,
		name: row.name,
		folder: row.folder,
		ownerId: row.owner_id,
		scopeType: row.scope_type,
		groupId: row.group_id,
		createdAt: row.created_at,
		updatedAt: row.updated_at,
	};
}

function rowToUser(row: {
	id: string;
	username: string;
	password_hash: string;
	role: UserRole;
	status: UserStatus;
	created_at: number;
}): UserRecord {
	return {
		id: row.id,
		username: row.username,
		passwordHash: row.password_hash,
		role: row.role,
		status: row.status,
		createdAt: row.created_at,
	};
}

function rowToGroup(row: {
	id: string;
	name: string;
	type: GroupType;
	owner_id: string;
	status: GroupStatus;
	created_at: number;
}): CourseGroupRecord {
	return {
		id: row.id,
		name: row.name,
		type: row.type,
		ownerId: row.owner_id,
		status: row.status,
		createdAt: row.created_at,
	};
}

function rowToSkill(row: {
	id: string;
	name: string;
	repo_url: string | null;
	version: string | null;
	visibility: ResourceVisibility;
	group_id: string | null;
	owner_id: string | null;
	path: string;
	status: ResourceStatus;
	review_state: SkillReviewState;
	created_at: number;
}): SkillRecord {
	return {
		id: row.id,
		name: row.name,
		repoUrl: row.repo_url,
		version: row.version,
		visibility: row.visibility,
		groupId: row.group_id,
		ownerId: row.owner_id,
		path: row.path,
		status: row.status,
		reviewState: row.review_state,
		createdAt: row.created_at,
	};
}

function rowToDataset(row: {
	id: string;
	name: string;
	size_bytes: number;
	visibility: ResourceVisibility;
	group_id: string | null;
	owner_id: string | null;
	path: string;
	review_state: SkillReviewState;
	status: ResourceStatus;
	created_at: number;
}): DatasetRecord {
	return {
		id: row.id,
		name: row.name,
		sizeBytes: row.size_bytes,
		visibility: row.visibility,
		groupId: row.group_id,
		ownerId: row.owner_id,
		path: row.path,
		status: row.status,
		reviewState: row.review_state,
		createdAt: row.created_at,
	};
}

function rowToSession(row: {
	id: string;
	project_id: string | null;
	user_id: string | null;
	container_id: string | null;
	status: PlatformSessionStatus;
	permission_mode: SessionPermissionMode;
	cwd: string | null;
	started_at: number | null;
	ended_at: number | null;
}): PlatformSessionRecord {
	return {
		id: row.id,
		projectId: row.project_id,
		userId: row.user_id,
		containerId: row.container_id,
		status: row.status,
		permissionMode: row.permission_mode,
		cwd: row.cwd,
		startedAt: row.started_at,
		endedAt: row.ended_at,
	};
}

function rowToAgent(row: {
	id: string;
	name: string;
	description: string | null;
	url: string;
	visibility: ResourceVisibility;
	group_id: string | null;
	owner_id: string | null;
	status: ResourceStatus;
	availability: AgentAvailability;
	last_checked_at: number | null;
	created_at: number;
	updated_at: number;
}): AgentListingRecord {
	return {
		id: row.id,
		name: row.name,
		description: row.description,
		url: row.url,
		visibility: row.visibility,
		groupId: row.group_id,
		ownerId: row.owner_id,
		status: row.status,
		availability: row.availability,
		lastCheckedAt: row.last_checked_at,
		createdAt: row.created_at,
		updatedAt: row.updated_at,
	};
}

export interface AgentCreateInput {
	id?: string;
	name: string;
	description?: string | null;
	url: string;
	visibility: ResourceVisibility;
	groupId?: string | null;
	ownerId?: string | null;
	status?: ResourceStatus;
	availability?: AgentAvailability;
	createdAt?: number;
	updatedAt?: number;
}

function rowToAuthSession(row: {
	token_hash: string;
	user_id: string;
	created_at: number;
	expires_at: number;
}): AuthSessionRecord {
	return {
		tokenHash: row.token_hash,
		userId: row.user_id,
		createdAt: row.created_at,
		expiresAt: row.expires_at,
	};
}

export interface UserCreateInput {
	id?: string;
	username: string;
	passwordHash: string;
	role: UserRole;
	status?: UserStatus;
	createdAt?: number;
}

export interface GroupCreateInput {
	id?: string;
	name: string;
	type: GroupType;
	ownerId: string;
	status?: GroupStatus;
	createdAt?: number;
}

export interface ProjectCreateInput {
	id?: string;
	name: string;
	folder?: string | null;
	ownerId?: string | null;
	scopeType?: ProjectScopeType;
	groupId?: string | null;
	createdAt?: number;
	updatedAt?: number;
}

export interface SkillCreateInput {
	id?: string;
	name: string;
	repoUrl?: string | null;
	version?: string | null;
	visibility: ResourceVisibility;
	groupId?: string | null;
	ownerId?: string | null;
	path: string;
	status?: ResourceStatus;
	reviewState?: SkillReviewState;
	createdAt?: number;
}

export interface DatasetCreateInput {
	id?: string;
	name: string;
	sizeBytes?: number;
	visibility: ResourceVisibility;
	groupId?: string | null;
	ownerId?: string | null;
	path: string;
	status?: ResourceStatus;
	reviewState?: SkillReviewState;
	createdAt?: number;
}

export interface PlatformSessionCreateInput {
	id?: string;
	projectId?: string | null;
	userId?: string | null;
	containerId?: string | null;
	status?: PlatformSessionStatus;
	permissionMode?: SessionPermissionMode;
	cwd?: string | null;
	startedAt?: number | null;
	endedAt?: number | null;
}

/** 平台仓储：users / course_groups / projects / skills / datasets / sessions 的 CRUD。 */
export class PlatformRepos {
	private readonly db: PlatformDatabase;

	constructor(db: PlatformDatabase) {
		this.db = db;
	}

	// ---- users ----

	userCreate(input: UserCreateInput): UserRecord {
		const id = input.id ?? randomUUID();
		const createdAt = input.createdAt ?? now();
		this.db
			.prepare("INSERT INTO users (id, username, password_hash, role, status, created_at) VALUES (?, ?, ?, ?, ?, ?)")
			.run(id, input.username, input.passwordHash, input.role, input.status ?? "active", createdAt);
		return {
			id,
			username: input.username,
			passwordHash: input.passwordHash,
			role: input.role,
			status: input.status ?? "active",
			createdAt,
		};
	}

	userList(): UserRecord[] {
		return this.db
			.prepare("SELECT id, username, password_hash, role, status, created_at FROM users ORDER BY created_at")
			.all()
			.map((row) => rowToUser(row as Parameters<typeof rowToUser>[0]));
	}

	userGetById(id: string): UserRecord | undefined {
		const row = this.db
			.prepare("SELECT id, username, password_hash, role, status, created_at FROM users WHERE id = ?")
			.get(id) as Parameters<typeof rowToUser>[0] | undefined;
		return row ? rowToUser(row) : undefined;
	}

	userGetByUsername(username: string): UserRecord | undefined {
		const row = this.db
			.prepare("SELECT id, username, password_hash, role, status, created_at FROM users WHERE username = ?")
			.get(username) as Parameters<typeof rowToUser>[0] | undefined;
		return row ? rowToUser(row) : undefined;
	}

	userSetStatus(id: string, status: UserStatus): void {
		this.db.prepare("UPDATE users SET status = ? WHERE id = ?").run(status, id);
	}

	userDelete(id: string): void {
		this.db.prepare("DELETE FROM users WHERE id = ?").run(id);
	}

	// ---- course_groups ----

	groupCreate(input: GroupCreateInput): CourseGroupRecord {
		const id = input.id ?? randomUUID();
		const createdAt = input.createdAt ?? now();
		this.db
			.prepare("INSERT INTO course_groups (id, name, type, owner_id, status, created_at) VALUES (?, ?, ?, ?, ?, ?)")
			.run(id, input.name, input.type, input.ownerId, input.status ?? "active", createdAt);
		return {
			id,
			name: input.name,
			type: input.type,
			ownerId: input.ownerId,
			status: input.status ?? "active",
			createdAt,
		};
	}

	groupList(): CourseGroupRecord[] {
		return this.db
			.prepare("SELECT id, name, type, owner_id, status, created_at FROM course_groups ORDER BY created_at")
			.all()
			.map((row) => rowToGroup(row as Parameters<typeof rowToGroup>[0]));
	}

	groupGet(id: string): CourseGroupRecord | undefined {
		const row = this.db
			.prepare("SELECT id, name, type, owner_id, status, created_at FROM course_groups WHERE id = ?")
			.get(id) as Parameters<typeof rowToGroup>[0] | undefined;
		return row ? rowToGroup(row) : undefined;
	}

	groupSetStatus(id: string, status: GroupStatus): void {
		this.db.prepare("UPDATE course_groups SET status = ? WHERE id = ?").run(status, id);
	}

	groupUpdateName(id: string, name: string): void {
		this.db.prepare("UPDATE course_groups SET name = ? WHERE id = ?").run(name, id);
	}

	groupSetOwner(id: string, ownerId: string): void {
		this.db.prepare("UPDATE course_groups SET owner_id = ? WHERE id = ?").run(ownerId, id);
	}

	/** 设置成员角色(upsert,已在组内则更新 role)。 */
	groupSetMemberRole(groupId: string, userId: string, role: GroupMemberRole): void {
		this.db
			.prepare(
				"INSERT INTO group_members (group_id, user_id, role) VALUES (?, ?, ?) ON CONFLICT(group_id, user_id) DO UPDATE SET role = excluded.role",
			)
			.run(groupId, userId, role);
	}

	/** 单一组长语义:晋升新组长,原组长(若非同一人)降为 member,并同步 owner_id。 */
	groupSetOwnerRole(groupId: string, newOwnerId: string): void {
		this.db.transaction(() => {
			const currentOwner = this.groupGet(groupId)?.ownerId;
			if (currentOwner && currentOwner !== newOwnerId) {
				this.groupSetMemberRole(groupId, currentOwner, "member");
			}
			this.groupSetMemberRole(groupId, newOwnerId, "owner");
			this.groupSetOwner(groupId, newOwnerId);
		});
	}

	/** 建组并把 owner 登记为组成员(保证 groupListByMember 对组长可见)。 */
	groupCreateWithOwner(input: GroupCreateInput): CourseGroupRecord {
		return this.db.transaction(() => {
			const group = this.groupCreate(input);
			this.groupAddMember(group.id, group.ownerId, "owner");
			return group;
		});
	}

	groupDelete(id: string): void {
		this.db.prepare("DELETE FROM course_groups WHERE id = ?").run(id);
	}

	groupAddMember(groupId: string, userId: string, role: GroupMemberRole): void {
		this.db
			.prepare("INSERT INTO group_members (group_id, user_id, role) VALUES (?, ?, ?)")
			.run(groupId, userId, role);
	}

	groupRemoveMember(groupId: string, userId: string): void {
		this.db.prepare("DELETE FROM group_members WHERE group_id = ? AND user_id = ?").run(groupId, userId);
	}

	groupListMembers(groupId: string): GroupMemberRecord[] {
		return this.db
			.prepare("SELECT group_id, user_id, role FROM group_members WHERE group_id = ?")
			.all(groupId)
			.map((row) => {
				const typed = row as { group_id: string; user_id: string; role: GroupMemberRole };
				return { groupId: typed.group_id, userId: typed.user_id, role: typed.role };
			});
	}

	groupMemberRole(groupId: string, userId: string): GroupMemberRole | undefined {
		const row = this.db
			.prepare("SELECT role FROM group_members WHERE group_id = ? AND user_id = ?")
			.get(groupId, userId) as { role: GroupMemberRole } | undefined;
		return row?.role;
	}

	groupListByMember(userId: string): CourseGroupRecord[] {
		return this.db
			.prepare(
				`SELECT g.id, g.name, g.type, g.owner_id, g.status, g.created_at
				 FROM course_groups g
				 JOIN group_members m ON m.group_id = g.id
				 WHERE m.user_id = ?
				 ORDER BY g.created_at`,
			)
			.all(userId)
			.map((row) => rowToGroup(row as Parameters<typeof rowToGroup>[0]));
	}

	// ---- projects ----

	projectCreate(input: ProjectCreateInput): ProjectRecord {
		const id = input.id ?? randomUUID();
		const createdAt = input.createdAt ?? now();
		const updatedAt = input.updatedAt ?? createdAt;
		this.db
			.prepare(
				"INSERT INTO projects (id, name, folder, owner_id, scope_type, group_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			)
			.run(
				id,
				input.name,
				input.folder ?? null,
				input.ownerId ?? null,
				input.scopeType ?? "private",
				input.groupId ?? null,
				createdAt,
				updatedAt,
			);
		return {
			id,
			name: input.name,
			folder: input.folder ?? null,
			ownerId: input.ownerId ?? null,
			scopeType: input.scopeType ?? "private",
			groupId: input.groupId ?? null,
			createdAt,
			updatedAt,
		};
	}

	projectList(): ProjectRecord[] {
		return this.db
			.prepare(
				"SELECT id, name, folder, owner_id, scope_type, group_id, created_at, updated_at FROM projects ORDER BY updated_at DESC",
			)
			.all()
			.map((row) => rowToProject(row as Parameters<typeof rowToProject>[0]));
	}

	/** 用户可见项目(存储层强制):owner 或 挂组且为 active 组成员;admin 用 projectList 全量。 */
	/** 用户可见项目(存储层强制):owner 或 挂组且为组成员(含归档组,历史项目对成员保留可见);admin 用 projectList 全量。 */
	projectListVisible(userId: string): ProjectRecord[] {
		return this.db
			.prepare(
				`SELECT id, name, folder, owner_id, scope_type, group_id, created_at, updated_at
				 FROM projects
				 WHERE owner_id = ?
				    OR (scope_type = 'group' AND group_id IN (
				        SELECT group_id FROM group_members WHERE user_id = ?
				    ))
				 ORDER BY updated_at DESC`,
			)
			.all(userId, userId)
			.map((row) => rowToProject(row as Parameters<typeof rowToProject>[0]));
	}
	projectGet(id: string): ProjectRecord | undefined {
		const row = this.db
			.prepare(
				"SELECT id, name, folder, owner_id, scope_type, group_id, created_at, updated_at FROM projects WHERE id = ?",
			)
			.get(id) as Parameters<typeof rowToProject>[0] | undefined;
		return row ? rowToProject(row) : undefined;
	}

	projectUpdate(id: string, patch: { name?: string; folder?: string | null }): ProjectRecord | undefined {
		const existing = this.projectGet(id);
		if (!existing) return undefined;
		const name = patch.name !== undefined ? patch.name : existing.name;
		const folder = patch.folder !== undefined ? patch.folder : existing.folder;
		this.db
			.prepare("UPDATE projects SET name = ?, folder = ?, updated_at = ? WHERE id = ?")
			.run(name, folder, now(), id);
		return this.projectGet(id);
	}

	projectDelete(id: string): void {
		this.db.prepare("DELETE FROM projects WHERE id = ?").run(id);
	}

	projectListSessionIds(projectId: string): string[] {
		const rows = this.db
			.prepare("SELECT session_id FROM project_session_links WHERE project_id = ?")
			.all(projectId) as unknown as Array<{ session_id: string }>;
		return rows.map((row) => row.session_id);
	}

	projectAddSession(projectId: string, sessionId: string): void {
		this.db
			.prepare("INSERT OR IGNORE INTO project_session_links (project_id, session_id) VALUES (?, ?)")
			.run(projectId, sessionId);
	}

	projectRemoveSession(sessionId: string): void {
		this.db.prepare("DELETE FROM project_session_links WHERE session_id = ?").run(sessionId);
	}

	// ---- skills ----

	skillCreate(input: SkillCreateInput): SkillRecord {
		const id = input.id ?? randomUUID();
		const createdAt = input.createdAt ?? now();
		this.db
			.prepare(
				"INSERT INTO skills (id, name, repo_url, version, visibility, group_id, owner_id, path, status, review_state, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			)
			.run(
				id,
				input.name,
				input.repoUrl ?? null,
				input.version ?? null,
				input.visibility,
				input.groupId ?? null,
				input.ownerId ?? null,
				input.path,
				input.status ?? "active",
				input.reviewState ?? "none",
				createdAt,
			);
		return {
			id,
			name: input.name,
			repoUrl: input.repoUrl ?? null,
			version: input.version ?? null,
			visibility: input.visibility,
			groupId: input.groupId ?? null,
			ownerId: input.ownerId ?? null,
			path: input.path,
			status: input.status ?? "active",
			reviewState: input.reviewState ?? "none",
			createdAt,
		};
	}

	skillList(): SkillRecord[] {
		return this.db
			.prepare(
				"SELECT id, name, repo_url, version, visibility, group_id, owner_id, path, status, review_state, created_at FROM skills ORDER BY created_at",
			)
			.all()
			.map((row) => rowToSkill(row as Parameters<typeof rowToSkill>[0]));
	}

	skillGet(id: string): SkillRecord | undefined {
		const row = this.db
			.prepare(
				"SELECT id, name, repo_url, version, visibility, group_id, owner_id, path, status, review_state, created_at FROM skills WHERE id = ?",
			)
			.get(id) as Parameters<typeof rowToSkill>[0] | undefined;
		return row ? rowToSkill(row) : undefined;
	}

	skillSetStatus(id: string, status: ResourceStatus): void {
		this.db.prepare("UPDATE skills SET status = ? WHERE id = ?").run(status, id);
	}

	skillSetReview(id: string, reviewState: SkillReviewState): void {
		this.db.prepare("UPDATE skills SET review_state = ? WHERE id = ?").run(reviewState, id);
	}

	/** 更新技能元数据(name/visibility/groupId/version);不存在则忽略。 */
	skillUpdate(
		id: string,
		patch: {
			name?: string;
			visibility?: ResourceVisibility;
			groupId?: string | null;
			version?: string | null;
		},
	): void {
		const existing = this.skillGet(id);
		if (!existing) return;
		this.db
			.prepare("UPDATE skills SET name = ?, visibility = ?, group_id = ?, version = ? WHERE id = ?")
			.run(
				patch.name ?? existing.name,
				patch.visibility ?? existing.visibility,
				patch.groupId !== undefined ? patch.groupId : existing.groupId,
				patch.version !== undefined ? patch.version : existing.version,
				id,
			);
	}

	/** 用户可见技能(存储层强制):active 且 (public ∨ 本人私有 ∨ 所在组库,含归档组成员)。 */
	skillListVisible(userId: string): SkillRecord[] {
		return this.db
			.prepare(
				`SELECT id, name, repo_url, version, visibility, group_id, owner_id, path, status, review_state, created_at
				 FROM skills
				 WHERE status = 'active'
				   AND (visibility = 'public'
				        OR (visibility = 'private' AND owner_id = ?)
				        OR (visibility = 'group' AND group_id IN (SELECT group_id FROM group_members WHERE user_id = ?)))
				 ORDER BY created_at`,
			)
			.all(userId, userId)
			.map((row) => rowToSkill(row as Parameters<typeof rowToSkill>[0]));
	}

	skillDelete(id: string): void {
		this.db.prepare("DELETE FROM skills WHERE id = ?").run(id);
	}

	// ---- datasets ----

	datasetCreate(input: DatasetCreateInput): DatasetRecord {
		const id = input.id ?? randomUUID();
		const createdAt = input.createdAt ?? now();
		this.db
			.prepare(
				"INSERT INTO datasets (id, name, size_bytes, visibility, group_id, owner_id, path, status, review_state, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			)
			.run(
				id,
				input.name,
				input.sizeBytes ?? 0,
				input.visibility,
				input.groupId ?? null,
				input.ownerId ?? null,
				input.path,
				input.status ?? "active",
				input.reviewState ?? "none",
				createdAt,
			);
		return {
			id,
			name: input.name,
			sizeBytes: input.sizeBytes ?? 0,
			visibility: input.visibility,
			groupId: input.groupId ?? null,
			ownerId: input.ownerId ?? null,
			path: input.path,
			status: input.status ?? "active",
			reviewState: input.reviewState ?? "none",
			createdAt,
		};
	}

	datasetList(): DatasetRecord[] {
		return this.db
			.prepare(
				"SELECT id, name, size_bytes, visibility, group_id, owner_id, path, status, review_state, created_at FROM datasets ORDER BY created_at",
			)
			.all()
			.map((row) => rowToDataset(row as Parameters<typeof rowToDataset>[0]));
	}

	datasetGet(id: string): DatasetRecord | undefined {
		const row = this.db
			.prepare(
				"SELECT id, name, size_bytes, visibility, group_id, owner_id, path, status, review_state, created_at FROM datasets WHERE id = ?",
			)
			.get(id) as Parameters<typeof rowToDataset>[0] | undefined;
		return row ? rowToDataset(row) : undefined;
	}

	datasetSetStatus(id: string, status: ResourceStatus): void {
		this.db.prepare("UPDATE datasets SET status = ? WHERE id = ?").run(status, id);
	}

	/** 更新数据集元数据(name/visibility/groupId);不存在则忽略。 */
	datasetUpdate(
		id: string,
		patch: {
			name?: string;
			visibility?: ResourceVisibility;
			groupId?: string | null;
		},
	): void {
		const existing = this.datasetGet(id);
		if (!existing) return;
		this.db
			.prepare("UPDATE datasets SET name = ?, visibility = ?, group_id = ? WHERE id = ?")
			.run(
				patch.name ?? existing.name,
				patch.visibility ?? existing.visibility,
				patch.groupId !== undefined ? patch.groupId : existing.groupId,
				id,
			);
	}

	/** 用户可见数据集(存储层强制):active 且 (public ∨ 本人私有 ∨ 所在组库,含归档组成员)。 */
	datasetListVisible(userId: string): DatasetRecord[] {
		return this.db
			.prepare(
				`SELECT id, name, size_bytes, visibility, group_id, owner_id, path, status, review_state, created_at
				 FROM datasets
				 WHERE status = 'active'
				   AND (visibility = 'public'
				        OR (visibility = 'private' AND owner_id = ?)
				        OR (visibility = 'group' AND group_id IN (SELECT group_id FROM group_members WHERE user_id = ?)))
				 ORDER BY created_at`,
			)
			.all(userId, userId)
			.map((row) => rowToDataset(row as Parameters<typeof rowToDataset>[0]));
	}

	/** 用户 active 数据集累计字节数(配额检查用)。 */
	datasetTotalSizeByOwner(userId: string): number {
		const row = this.db
			.prepare("SELECT COALESCE(SUM(size_bytes), 0) AS total FROM datasets WHERE owner_id = ? AND status = 'active'")
			.get(userId) as { total: number } | undefined;
		return row?.total ?? 0;
	}

	datasetSetReview(id: string, reviewState: SkillReviewState): void {
		this.db.prepare("UPDATE datasets SET review_state = ? WHERE id = ?").run(reviewState, id);
	}

	datasetDelete(id: string): void {
		this.db.prepare("DELETE FROM datasets WHERE id = ?").run(id);
	}

	// ---- agent_listings (智能体广场) ----

	agentCreate(input: AgentCreateInput): AgentListingRecord {
		const id = input.id ?? randomUUID();
		const now = Date.now();
		const createdAt = input.createdAt ?? now;
		this.db
			.prepare(
				"INSERT INTO agent_listings (id, name, description, url, visibility, group_id, owner_id, status, availability, last_checked_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			)
			.run(
				id,
				input.name,
				input.description ?? null,
				input.url,
				input.visibility,
				input.groupId ?? null,
				input.ownerId ?? null,
				input.status ?? "active",
				input.availability ?? "unknown",
				null,
				createdAt,
				input.updatedAt ?? createdAt,
			);
		return this.agentGet(id)!;
	}

	agentList(): AgentListingRecord[] {
		return this.db
			.prepare(
				"SELECT id, name, description, url, visibility, group_id, owner_id, status, availability, last_checked_at, created_at, updated_at FROM agent_listings ORDER BY created_at",
			)
			.all()
			.map((row) => rowToAgent(row as Parameters<typeof rowToAgent>[0]));
	}

	agentGet(id: string): AgentListingRecord | undefined {
		const row = this.db
			.prepare(
				"SELECT id, name, description, url, visibility, group_id, owner_id, status, availability, last_checked_at, created_at, updated_at FROM agent_listings WHERE id = ?",
			)
			.get(id) as Parameters<typeof rowToAgent>[0] | undefined;
		return row ? rowToAgent(row) : undefined;
	}

	/** 用户可见智能体(存储层强制):active 且 (public ∨ 本人私有 ∨ 所在组库,含归档组成员)。 */
	agentListVisible(userId: string): AgentListingRecord[] {
		return this.db
			.prepare(
				`SELECT id, name, description, url, visibility, group_id, owner_id, status, availability, last_checked_at, created_at, updated_at
				 FROM agent_listings
				 WHERE status = 'active'
				   AND (visibility = 'public'
				        OR (visibility = 'private' AND owner_id = ?)
				        OR (visibility = 'group' AND group_id IN (SELECT group_id FROM group_members WHERE user_id = ?)))
				 ORDER BY created_at`,
			)
			.all(userId, userId)
			.map((row) => rowToAgent(row as Parameters<typeof rowToAgent>[0]));
	}

	agentUpdate(
		id: string,
		patch: {
			name?: string;
			description?: string | null;
			url?: string;
			visibility?: ResourceVisibility;
			groupId?: string | null;
		},
	): void {
		const existing = this.agentGet(id);
		if (!existing) return;
		this.db
			.prepare(
				"UPDATE agent_listings SET name = ?, description = ?, url = ?, visibility = ?, group_id = ?, updated_at = ? WHERE id = ?",
			)
			.run(
				patch.name ?? existing.name,
				patch.description !== undefined ? patch.description : existing.description,
				patch.url ?? existing.url,
				patch.visibility ?? existing.visibility,
				patch.groupId !== undefined ? patch.groupId : existing.groupId,
				Date.now(),
				id,
			);
	}

	agentSetStatus(id: string, status: ResourceStatus): void {
		this.db.prepare("UPDATE agent_listings SET status = ?, updated_at = ? WHERE id = ?").run(status, Date.now(), id);
	}

	agentSetAvailability(id: string, availability: AgentAvailability, checkedAt: number): void {
		this.db
			.prepare("UPDATE agent_listings SET availability = ?, last_checked_at = ?, updated_at = ? WHERE id = ?")
			.run(availability, checkedAt, Date.now(), id);
	}

	agentDelete(id: string): void {
		this.db.prepare("DELETE FROM agent_listings WHERE id = ?").run(id);
	}

	// ---- sessions (executor 跟踪) ----

	sessionCreate(input: PlatformSessionCreateInput): PlatformSessionRecord {
		const id = input.id ?? randomUUID();
		this.db
			.prepare(
				"INSERT INTO sessions (id, project_id, user_id, container_id, status, permission_mode, cwd, started_at, ended_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			)
			.run(
				id,
				input.projectId ?? null,
				input.userId ?? null,
				input.containerId ?? null,
				input.status ?? "created",
				input.permissionMode ?? "full_access",
				input.cwd ?? null,
				input.startedAt ?? null,
				input.endedAt ?? null,
			);
		return {
			id,
			projectId: input.projectId ?? null,
			userId: input.userId ?? null,
			containerId: input.containerId ?? null,
			status: input.status ?? "created",
			permissionMode: input.permissionMode ?? "full_access",
			cwd: input.cwd ?? null,
			startedAt: input.startedAt ?? null,
			endedAt: input.endedAt ?? null,
		};
	}

	sessionList(): PlatformSessionRecord[] {
		return this.db
			.prepare(
				"SELECT id, project_id, user_id, container_id, status, permission_mode, cwd, started_at, ended_at FROM sessions ORDER BY COALESCE(started_at, 0) DESC",
			)
			.all()
			.map((row) => rowToSession(row as Parameters<typeof rowToSession>[0]));
	}

	sessionGet(id: string): PlatformSessionRecord | undefined {
		const row = this.db
			.prepare(
				"SELECT id, project_id, user_id, container_id, status, permission_mode, cwd, started_at, ended_at FROM sessions WHERE id = ?",
			)
			.get(id) as Parameters<typeof rowToSession>[0] | undefined;
		return row ? rowToSession(row) : undefined;
	}

	sessionUpdate(
		id: string,
		patch: {
			status?: PlatformSessionStatus;
			containerId?: string | null;
			permissionMode?: SessionPermissionMode;
			endedAt?: number | null;
		},
	): void {
		const existing = this.sessionGet(id);
		if (!existing) return;
		this.db
			.prepare("UPDATE sessions SET status = ?, container_id = ?, permission_mode = ?, ended_at = ? WHERE id = ?")
			.run(
				patch.status ?? existing.status,
				patch.containerId !== undefined ? patch.containerId : existing.containerId,
				patch.permissionMode !== undefined ? patch.permissionMode : existing.permissionMode,
				patch.endedAt !== undefined ? patch.endedAt : existing.endedAt,
				id,
			);
	}

	/** 删除项目前解除平台会话对项目的引用(外键无级联)。 */
	sessionClearProject(projectId: string): void {
		this.db.prepare("UPDATE sessions SET project_id = NULL WHERE project_id = ?").run(projectId);
	}

	sessionDelete(id: string): void {
		this.db.prepare("DELETE FROM sessions WHERE id = ?").run(id);
	}

	// ---- auth sessions ----

	authSessionCreate(input: { tokenHash: string; userId: string; createdAt?: number; expiresAt: number }): void {
		this.db
			.prepare("INSERT INTO auth_sessions (token_hash, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)")
			.run(input.tokenHash, input.userId, input.createdAt ?? now(), input.expiresAt);
	}

	authSessionGet(tokenHash: string): AuthSessionRecord | undefined {
		const row = this.db
			.prepare("SELECT token_hash, user_id, created_at, expires_at FROM auth_sessions WHERE token_hash = ?")
			.get(tokenHash) as Parameters<typeof rowToAuthSession>[0] | undefined;
		return row ? rowToAuthSession(row) : undefined;
	}

	authSessionDelete(tokenHash: string): void {
		this.db.prepare("DELETE FROM auth_sessions WHERE token_hash = ?").run(tokenHash);
	}

	userSetPassword(id: string, passwordHash: string): void {
		this.db.prepare("UPDATE users SET password_hash = ? WHERE id = ?").run(passwordHash, id);
	}
}
