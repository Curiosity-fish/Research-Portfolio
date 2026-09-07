import { TfaServerError } from "@earendil-works/tfa-server";
import { AuthError, type AuthUserView } from "./auth.ts";
import type { PlatformRepos } from "./db/repos.ts";
import type { CourseGroupRecord, GroupMemberRole, GroupStatus, GroupType } from "./db/types.ts";
import { canManageGroup, requireRole } from "./permissions.ts";

export interface GroupView {
	id: string;
	name: string;
	type: GroupType;
	status: GroupStatus;
	ownerId: string;
	memberCount: number;
	/** 当前用户在组内的角色;非成员为 null。 */
	myRole: GroupMemberRole | null;
}

export interface GroupMemberView {
	userId: string;
	username: string;
	role: GroupMemberRole;
}

export interface GroupDetail extends GroupView {
	members: GroupMemberView[];
}

function cleanName(name: string): string {
	return name.replace(/[\r\n]+/g, " ").trim();
}

/**
 * 课程/课题组服务:方法以当前用户为第一参数,服务层强制 ACL。
 * repos 为同步 CRUD,本服务同步返回视图。
 */
export class GroupService {
	private readonly repos: PlatformRepos;

	constructor(repos: PlatformRepos) {
		this.repos = repos;
	}

	private getGroupOrThrow(groupId: string): CourseGroupRecord {
		const group = this.repos.groupGet(groupId);
		if (!group) throw new TfaServerError("not_found", `组不存在: ${groupId}`);
		return group;
	}

	private toGroupView(group: CourseGroupRecord, userId: string): GroupView {
		return {
			id: group.id,
			name: group.name,
			type: group.type,
			status: group.status,
			ownerId: group.ownerId,
			memberCount: this.repos.groupListMembers(group.id).length,
			myRole: this.repos.groupMemberRole(group.id, userId) ?? null,
		};
	}

	/** 我可见的组:admin 全部;其他人是自己作为 owner/member 的组(含归档)。 */
	listGroups(user: AuthUserView): GroupView[] {
		const groups = user.role === "admin" ? this.repos.groupList() : this.repos.groupListByMember(user.id);
		return groups.map((group) => this.toGroupView(group, user.id));
	}

	createGroup(user: AuthUserView, input: { name: string; type: GroupType }): GroupView {
		requireRole(user, ["admin", "teacher"]);
		const name = cleanName(input.name);
		if (!name) throw new TfaServerError("invalid_request", "组名称不能为空");
		if (input.type !== "course" && input.type !== "research") {
			throw new TfaServerError("invalid_request", "type 必须是 course 或 research");
		}
		const group = this.repos.groupCreateWithOwner({ name, type: input.type, ownerId: user.id });
		return this.toGroupView(group, user.id);
	}

	/** 组详情(含成员):组成员或 admin。 */
	getGroup(user: AuthUserView, groupId: string): GroupDetail {
		const group = this.getGroupOrThrow(groupId);
		if (user.role !== "admin" && this.repos.groupMemberRole(groupId, user.id) === undefined) {
			throw new AuthError("forbidden", "不是该组成员");
		}
		const members = this.repos.groupListMembers(groupId).map((member) => ({
			userId: member.userId,
			username: this.repos.userGetById(member.userId)?.username ?? "未知用户",
			role: member.role,
		}));
		return { ...this.toGroupView(group, user.id), members };
	}

	updateGroup(user: AuthUserView, groupId: string, input: { name?: string }): GroupView {
		const group = this.getGroupOrThrow(groupId);
		if (!canManageGroup(user, group)) throw new AuthError("forbidden", "只有组长或管理员可以管理该组");
		if (input.name !== undefined) {
			const name = cleanName(input.name);
			if (!name) throw new TfaServerError("invalid_request", "组名称不能为空");
			this.repos.groupUpdateName(groupId, name);
		}
		return this.toGroupView(this.repos.groupGet(groupId)!, user.id);
	}

	/** 归档(软删):项目与库保留,组库对现有成员只读;归档后不可加新成员。 */
	archiveGroup(user: AuthUserView, groupId: string): GroupView {
		const group = this.getGroupOrThrow(groupId);
		if (!canManageGroup(user, group)) throw new AuthError("forbidden", "只有组长或管理员可以归档该组");
		this.repos.groupSetStatus(groupId, "archived");
		return this.toGroupView(this.repos.groupGet(groupId)!, user.id);
	}

	addMember(user: AuthUserView, groupId: string, input: { userId: string }): void {
		const group = this.getGroupOrThrow(groupId);
		if (!canManageGroup(user, group)) throw new AuthError("forbidden", "只有组长或管理员可以添加成员");
		if (group.status === "archived") throw new TfaServerError("invalid_request", "归档组不可添加新成员");
		const target = this.repos.userGetById(input.userId);
		if (!target) throw new TfaServerError("not_found", `用户不存在: ${input.userId}`);
		if (this.repos.groupMemberRole(groupId, input.userId) !== undefined) {
			throw new TfaServerError("invalid_request", "该用户已是组成员");
		}
		this.repos.groupAddMember(groupId, input.userId, "member");
	}

	/**
	 * 任免组长:role=owner 晋升目标为组长(原组长自动降为 member,保持单一组长);
	 * role=member 时不能直接降级组长,须先任免新组长。
	 */
	setMemberRole(user: AuthUserView, groupId: string, targetUserId: string, role: GroupMemberRole): void {
		const group = this.getGroupOrThrow(groupId);
		if (!canManageGroup(user, group)) throw new AuthError("forbidden", "只有组长或管理员可以任免组长");
		const currentRole = this.repos.groupMemberRole(groupId, targetUserId);
		if (currentRole === undefined) throw new TfaServerError("not_found", `该用户不是组成员: ${targetUserId}`);
		if (role === "owner") {
			this.repos.groupSetOwnerRole(groupId, targetUserId);
			return;
		}
		if (group.ownerId === targetUserId) {
			throw new TfaServerError("invalid_request", "不能直接降级组长,请先任免新组长");
		}
		this.repos.groupSetMemberRole(groupId, targetUserId, "member");
	}

	removeMember(user: AuthUserView, groupId: string, targetUserId: string): void {
		const group = this.getGroupOrThrow(groupId);
		if (!canManageGroup(user, group)) throw new AuthError("forbidden", "只有组长或管理员可以移除成员");
		const currentRole = this.repos.groupMemberRole(groupId, targetUserId);
		if (currentRole === undefined) throw new TfaServerError("not_found", `该用户不是组成员: ${targetUserId}`);
		if (group.ownerId === targetUserId) {
			throw new TfaServerError("invalid_request", "不能移除组长,请先任免新组长");
		}
		this.repos.groupRemoveMember(groupId, targetUserId);
	}
}
