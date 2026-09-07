import { execFile } from "node:child_process";
import { randomUUID } from "node:crypto";
import { mkdir } from "node:fs/promises";
import { join } from "node:path";
import { promisify } from "node:util";
import { TfaServerError } from "@earendil-works/tfa-server";
import { AuthError, type AuthUserView } from "./auth.ts";
import type { PlatformRepos } from "./db/repos.ts";
import type { ResourceVisibility, SkillRecord, SkillReviewState } from "./db/types.ts";

const execFileAsync = promisify(execFile);

async function runGit(args: string[], cwd?: string): Promise<string> {
	try {
		const { stdout } = await execFileAsync("git", args, { cwd, timeout: 30_000, windowsHide: true });
		return stdout.trim();
	} catch (error) {
		const message = error instanceof Error ? error.message : String(error);
		throw new TfaServerError("invalid_request", `git 操作失败: ${message}`);
	}
}

export function isVisibility(value: unknown): value is ResourceVisibility {
	return value === "private" || value === "group" || value === "public";
}

export interface SkillView {
	id: string;
	name: string;
	repoUrl: string | null;
	version: string | null;
	visibility: ResourceVisibility;
	groupId: string | null;
	groupName: string | undefined;
	ownerId: string | null;
	status: string;
	reviewState: SkillReviewState;
	createdAt: number;
}

export interface CreateSkillInput {
	name?: string;
	repoUrl?: string;
	visibility: ResourceVisibility;
	groupId?: string;
	/** 自研/上传型(无 repoUrl,不参与 git 更新)。 */
	upload?: boolean;
}

export interface UpdateSkillInput {
	name?: string;
	visibility?: ResourceVisibility;
	groupId?: string;
}

export interface SkillServiceOptions {
	repos: PlatformRepos;
	/** 技能存储根目录: {agentDir}/platform/skills,每个技能一个 {id}/ 目录。 */
	storageRoot: string;
}

/**
 * 技能库服务:方法以当前用户为第一参数,服务层强制 ACL。
 * 安装 = git clone 进存储目录;可见性 = 私有/组/公开(组库对归档组成员仍只读可见)。
 */
export class SkillService {
	private readonly repos: PlatformRepos;
	private readonly storageRoot: string;

	constructor(options: SkillServiceOptions) {
		this.repos = options.repos;
		this.storageRoot = options.storageRoot;
		void mkdir(this.storageRoot, { recursive: true }).catch(() => {
			// 目录创建失败延迟到实际写入时报错
		});
	}

	private getSkillOrThrow(id: string): SkillRecord {
		const skill = this.repos.skillGet(id);
		if (!skill) throw new TfaServerError("not_found", `技能不存在: ${id}`);
		return skill;
	}

	private toSkillView(record: SkillRecord): SkillView {
		return {
			id: record.id,
			name: record.name,
			repoUrl: record.repoUrl,
			version: record.version,
			visibility: record.visibility,
			groupId: record.groupId,
			groupName: record.groupId ? (this.repos.groupGet(record.groupId)?.name ?? "未知组") : undefined,
			ownerId: record.ownerId,
			status: record.status,
			reviewState: record.reviewState,
			createdAt: record.createdAt,
		};
	}

	/** 当前用户可见技能:本人私有 ∪ 所在组库 ∪ 公开(admin 另见全部含已下架)。 */
	listSkills(user: AuthUserView): SkillView[] {
		const records = user.role === "admin" ? this.repos.skillList() : this.repos.skillListVisible(user.id);
		return records.map((record) => this.toSkillView(record));
	}

	/** 某项目作用域下可见技能:本人私有 ∪ (挂组则仅该组库) ∪ 公开(admin 同规则,含已下架由调用方展示)。 */
	listForProject(user: AuthUserView, scope: { type: "private" } | { type: "group"; groupId: string }): SkillView[] {
		const records = user.role === "admin" ? this.repos.skillList() : this.repos.skillListVisible(user.id);
		const groupId = scope.type === "group" ? scope.groupId : null;
		return records
			.filter((record) => {
				if (record.status !== "active") return false;
				if (record.visibility === "public") return true;
				if (record.visibility === "private" && record.ownerId === user.id) return true;
				return record.visibility === "group" && groupId !== null && record.groupId === groupId;
			})
			.map((record) => this.toSkillView(record));
	}
	/** 安装(GitHub clone)或自研/上传;公开技能置 reviewState=pending(A7 先上线后审核)。 */
	async createSkill(user: AuthUserView, input: CreateSkillInput): Promise<SkillView> {
		if (!isVisibility(input.visibility)) {
			throw new TfaServerError("invalid_request", "visibility 必须是 private/group/public");
		}
		const repoUrl = input.repoUrl?.trim() || null;
		if (repoUrl && input.upload) {
			throw new TfaServerError("invalid_request", "repoUrl 与 upload 不能同时提供");
		}
		if (!repoUrl && !input.upload) {
			throw new TfaServerError("invalid_request", "需要 repoUrl 或 upload");
		}
		let groupId: string | null = null;
		if (input.visibility === "group") {
			if (!input.groupId) throw new TfaServerError("invalid_request", "组技能需要 groupId");
			const group = this.repos.groupGet(input.groupId);
			if (!group) throw new TfaServerError("not_found", `组不存在: ${input.groupId}`);
			if (group.status === "archived") throw new TfaServerError("invalid_request", "归档组不可新增技能");
			if (user.role !== "admin" && this.repos.groupMemberRole(group.id, user.id) === undefined) {
				throw new AuthError("forbidden", "不是该组成员,不能在该组共享技能");
			}
			groupId = group.id;
		}
		const name = input.name?.trim() || (repoUrl ? (repoUrl.replace(/\/+$/, "").split(/[/\\]/).pop() ?? "") : "");
		if (!name) throw new TfaServerError("invalid_request", "技能名称不能为空");

		const id = randomUUID();
		const dir = join(this.storageRoot, id);
		await mkdir(dir, { recursive: true });
		let version: string | null = null;
		if (repoUrl) {
			await runGit(["clone", "--depth", "1", repoUrl, dir]);
			version = await runGit(["rev-parse", "--short", "HEAD"], dir).catch(() => null);
		}
		const reviewState = input.visibility === "public" ? "pending" : "none";
		const record = this.repos.skillCreate({
			id,
			name,
			repoUrl,
			version,
			visibility: input.visibility,
			groupId,
			ownerId: user.id,
			path: dir,
			reviewState,
		});
		return this.toSkillView(record);
	}

	/** 改名/改可见性;owner 或 admin。改 public → pending;组可见性切换须同时给 groupId。 */
	updateSkill(user: AuthUserView, id: string, input: UpdateSkillInput): SkillView {
		const skill = this.getSkillOrThrow(id);
		if (skill.ownerId !== user.id && user.role !== "admin") {
			throw new AuthError("forbidden", "只有技能 owner 或管理员可以修改");
		}
		let visibility = skill.visibility;
		let groupId = skill.groupId;
		if (input.visibility !== undefined) {
			if (!isVisibility(input.visibility)) {
				throw new TfaServerError("invalid_request", "visibility 必须是 private/group/public");
			}
			visibility = input.visibility;
			if (visibility === "group") {
				if (!input.groupId) throw new TfaServerError("invalid_request", "组技能需要 groupId");
				const group = this.repos.groupGet(input.groupId);
				if (!group) throw new TfaServerError("not_found", `组不存在: ${input.groupId}`);
				if (group.status === "archived") throw new TfaServerError("invalid_request", "归档组不可新增技能");
				if (user.role !== "admin" && this.repos.groupMemberRole(group.id, user.id) === undefined) {
					throw new AuthError("forbidden", "不是该组成员,不能在该组共享技能");
				}
				groupId = group.id;
			} else {
				groupId = null;
			}
		} else if (input.groupId !== undefined) {
			throw new TfaServerError("invalid_request", "仅当设置 visibility=group 时可指定 groupId");
		}
		const name = input.name?.trim() ?? skill.name;
		if (!name) throw new TfaServerError("invalid_request", "技能名称不能为空");

		this.repos.skillUpdate(id, { name, visibility, groupId });
		if (visibility === "public" && skill.visibility !== "public") {
			this.repos.skillSetReview(id, "pending");
		}
		return this.toSkillView(this.repos.skillGet(id)!);
	}

	/** 下架(软删);owner 或 admin。 */
	deleteSkill(user: AuthUserView, id: string): void {
		const skill = this.getSkillOrThrow(id);
		if (skill.ownerId !== user.id && user.role !== "admin") {
			throw new AuthError("forbidden", "只有技能 owner 或管理员可以下架");
		}
		this.repos.skillSetStatus(id, "removed");
	}

	/** git 安装型技能拉取更新;owner 或 admin。 */
	async updateSkillFromRepo(user: AuthUserView, id: string): Promise<SkillView> {
		const skill = this.getSkillOrThrow(id);
		if (skill.ownerId !== user.id && user.role !== "admin") {
			throw new AuthError("forbidden", "只有技能 owner 或管理员可以更新");
		}
		if (!skill.repoUrl || !skill.path) {
			throw new TfaServerError("invalid_request", "非 git 安装的技能不可更新");
		}
		await runGit(["pull", "--ff-only"], skill.path);
		const version = await runGit(["rev-parse", "--short", "HEAD"], skill.path).catch(() => null);
		this.repos.skillUpdate(id, { version });
		return this.toSkillView(this.repos.skillGet(id)!);
	}

	/** 审核公开技能(admin):pending → approved/rejected。 */
	reviewSkill(user: AuthUserView, id: string, reviewState: SkillReviewState): SkillView {
		if (user.role !== "admin") throw new AuthError("forbidden", "需要管理员权限");
		if (reviewState !== "approved" && reviewState !== "rejected") {
			throw new TfaServerError("invalid_request", "reviewState 必须是 approved 或 rejected");
		}
		const skill = this.getSkillOrThrow(id);
		this.repos.skillSetReview(skill.id, reviewState);
		const updated = this.repos.skillGet(id);
		return this.toSkillView(updated!);
	}
}
