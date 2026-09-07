import { mkdirSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { TfaServerError } from "@earendil-works/tfa-server";
import { AuthError, type AuthUserView } from "./auth.ts";
import { openPlatformDatabase } from "./db/database.ts";
import { PlatformRepos } from "./db/repos.ts";
import type { PlatformDatabase, ProjectRecord, ProjectScopeType } from "./db/types.ts";

/** 项目作用域(输出视图):私有或挂靠某课程/课题组。 */
export type ProjectScope = { type: "private" } | { type: "group"; groupId: string; groupName: string };

/** 创建时的作用域输入:挂组须提供 groupId(缺省 private)。 */
export type ProjectScopeInput = { type: "private" } | { type: "group"; groupId: string };

export interface ProjectView {
	id: string;
	name: string;
	folder?: string;
	/** 创建者;旧数据导入/存储层直建为 null。 */
	ownerId: string | null;
	scope: ProjectScope;
	createdAt: number;
	updatedAt: number;
	sessionIds: string[];
}

export interface CreateProjectInput {
	name: string;
	folder?: string;
	/** 作用域;缺省 private。 */
	scope?: ProjectScopeInput;
}

export interface UpdateProjectInput {
	name?: string;
	folder?: string;
}

interface LegacyProjectFileShape {
	version: number;
	projects: Array<{
		id: string;
		name: string;
		folder?: string;
		createdAt: number;
		updatedAt: number;
		sessionIds: string[];
	}>;
}

function cleanName(name: string): string {
	return name.replace(/[\r\n]+/g, " ").trim();
}

/**
 * 项目实体存储:基于平台 SQLite(projects + project_session_links)。
 * 存储层方法(list/get/create/update/delete)返回全量;用户感知方法(*ForUser)在服务层强制 ACL。
 * 首次打开时若 DB 为空且存在旧 projects.json,则导入旧数据(不删除原文件)。
 */
export class ProjectStore {
	private readonly db: PlatformDatabase;
	private readonly repos: PlatformRepos;
	private closed = false;

	constructor(agentDir: string, dbPath?: string) {
		const resolvedDbPath = dbPath ?? join(agentDir, "platform.db");
		mkdirSync(dirname(resolvedDbPath), { recursive: true });
		this.db = openPlatformDatabase(resolvedDbPath);
		this.repos = new PlatformRepos(this.db);
		this.importLegacyProjects(agentDir);
	}

	getRepos(): PlatformRepos {
		return this.repos;
	}

	close(): void {
		if (this.closed) return;
		this.closed = true;
		this.db.close();
	}

	private importLegacyProjects(agentDir: string): void {
		if (this.repos.projectList().length > 0) return;
		const legacyPath = join(agentDir, "projects.json");
		let shape: LegacyProjectFileShape;
		try {
			shape = JSON.parse(readFileSync(legacyPath, "utf8")) as LegacyProjectFileShape;
		} catch {
			return;
		}
		if (!Array.isArray(shape.projects) || shape.projects.length === 0) return;
		this.db.transaction(() => {
			for (const legacy of shape.projects) {
				this.repos.projectCreate({
					id: legacy.id,
					name: legacy.name,
					folder: legacy.folder?.trim() || null,
					createdAt: legacy.createdAt,
					updatedAt: legacy.updatedAt,
				});
				for (const sessionId of legacy.sessionIds) {
					this.repos.projectAddSession(legacy.id, sessionId);
				}
			}
		});
	}

	private toProjectView(record: ProjectRecord): ProjectView {
		return {
			id: record.id,
			name: record.name,
			folder: record.folder ?? undefined,
			ownerId: record.ownerId,
			scope:
				record.scopeType === "group" && record.groupId
					? {
							type: "group",
							groupId: record.groupId,
							groupName: this.repos.groupGet(record.groupId)?.name ?? "未知组",
						}
					: { type: "private" },
			createdAt: record.createdAt,
			updatedAt: record.updatedAt,
			sessionIds: this.repos.projectListSessionIds(record.id),
		};
	}

	private canSee(user: AuthUserView, record: ProjectRecord): boolean {
		if (record.ownerId === user.id) return true;
		if (record.scopeType === "group" && record.groupId) {
			// 组成员即可见(含归档组,历史项目对成员保留);改/删仍限 owner/admin → 天然只读
			const group = this.repos.groupGet(record.groupId);
			if (group && this.repos.groupMemberRole(record.groupId, user.id) !== undefined) {
				return true;
			}
		}
		return false;
	}

	// ---- 存储层(全量,供内部/管理端使用) ----

	async list(): Promise<ProjectView[]> {
		return this.repos.projectList().map((record) => this.toProjectView(record));
	}

	async get(id: string): Promise<ProjectView | undefined> {
		const record = this.repos.projectGet(id);
		return record ? this.toProjectView(record) : undefined;
	}

	async create(input: CreateProjectInput): Promise<ProjectView> {
		const name = cleanName(input.name);
		if (!name) throw new TfaServerError("invalid_request", "项目名称不能为空");
		const record = this.repos.projectCreate({
			name,
			folder: input.folder?.trim() || null,
		});
		return this.toProjectView(record);
	}

	async update(id: string, input: UpdateProjectInput): Promise<ProjectView> {
		const existing = this.repos.projectGet(id);
		if (!existing) throw new TfaServerError("not_found", `项目不存在: ${id}`);
		const patch: { name?: string; folder?: string | null } = {};
		if (input.name !== undefined) {
			const name = cleanName(input.name);
			if (!name) throw new TfaServerError("invalid_request", "项目名称不能为空");
			patch.name = name;
		}
		if (input.folder !== undefined) {
			patch.folder = input.folder.trim() || null;
		}
		const updated = this.repos.projectUpdate(id, patch);
		if (!updated) throw new TfaServerError("not_found", `项目不存在: ${id}`);
		return this.toProjectView(updated);
	}

	async delete(id: string): Promise<void> {
		const existing = this.repos.projectGet(id);
		if (!existing) throw new TfaServerError("not_found", `项目不存在: ${id}`);
		this.repos.sessionClearProject(id);
		this.repos.projectDelete(id);
	}

	/** 移动会话到项目(projectId 为 null 表示移出项目/无项目)。 */
	async moveSession(sessionId: string, projectId: string | null): Promise<void> {
		if (projectId !== null) {
			const target = this.repos.projectGet(projectId);
			if (!target) throw new TfaServerError("not_found", `项目不存在: ${projectId}`);
		}
		this.db.transaction(() => {
			this.repos.projectRemoveSession(sessionId);
			if (projectId !== null) {
				this.repos.projectAddSession(projectId, sessionId);
				this.repos.projectUpdate(projectId, {});
			}
		});
	}

	/** 从所有项目移除会话(会话被删除时调用)。 */
	async removeSession(sessionId: string): Promise<void> {
		this.repos.projectRemoveSession(sessionId);
	}

	// ---- 用户感知(ACL 强制) ----

	/** 当前用户可见项目:owner ∪ 挂组且为 active 组成员 ∪ admin 全量。 */
	async listForUser(user: AuthUserView): Promise<ProjectView[]> {
		const records = user.role === "admin" ? this.repos.projectList() : this.repos.projectListVisible(user.id);
		return records.map((record) => this.toProjectView(record));
	}

	/** 不可见项目返回 undefined(不泄露存在性)。 */
	async getForUser(user: AuthUserView, id: string): Promise<ProjectView | undefined> {
		const record = this.repos.projectGet(id);
		if (!record) return undefined;
		if (user.role !== "admin" && !this.canSee(user, record)) return undefined;
		return this.toProjectView(record);
	}

	async createForUser(user: AuthUserView, input: CreateProjectInput): Promise<ProjectView> {
		const name = cleanName(input.name);
		if (!name) throw new TfaServerError("invalid_request", "项目名称不能为空");
		const scope = input.scope ?? { type: "private" as const };
		let scopeType: ProjectScopeType = "private";
		let groupId: string | null = null;
		if (scope.type === "group") {
			const group = this.repos.groupGet(scope.groupId);
			if (!group) throw new TfaServerError("not_found", `组不存在: ${scope.groupId}`);
			if (group.status === "archived") throw new TfaServerError("invalid_request", "归档组不可新建项目");
			if (user.role !== "admin" && this.repos.groupMemberRole(group.id, user.id) === undefined) {
				throw new AuthError("forbidden", "不是该组成员,不能在该组下创建项目");
			}
			scopeType = "group";
			groupId = group.id;
		}
		const record = this.repos.projectCreate({
			name,
			folder: input.folder?.trim() || null,
			ownerId: user.id,
			scopeType,
			groupId,
		});
		return this.toProjectView(record);
	}

	/** 修改仅 owner 或 admin;作用域不可改。 */
	async updateForUser(user: AuthUserView, id: string, input: UpdateProjectInput): Promise<ProjectView> {
		const record = this.repos.projectGet(id);
		if (!record) throw new TfaServerError("not_found", `项目不存在: ${id}`);
		if (record.ownerId !== user.id && user.role !== "admin") {
			throw new AuthError("forbidden", "只有项目 owner 或管理员可以修改项目");
		}
		return this.update(id, input);
	}

	/** 删除仅 owner 或 admin。 */
	async deleteForUser(user: AuthUserView, id: string): Promise<void> {
		const record = this.repos.projectGet(id);
		if (!record) throw new TfaServerError("not_found", `项目不存在: ${id}`);
		if (record.ownerId !== user.id && user.role !== "admin") {
			throw new AuthError("forbidden", "只有项目 owner 或管理员可以删除项目");
		}
		this.repos.sessionClearProject(id);
		this.repos.projectDelete(id);
	}
}
