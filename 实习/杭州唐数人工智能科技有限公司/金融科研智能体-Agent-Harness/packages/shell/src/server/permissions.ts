import { AuthError, type AuthUserView } from "./auth.ts";
import type { CourseGroupRecord, UserRole } from "./db/types.ts";

/** 角色门槛:user.role 不在 roles 内 → 403。 */
export function requireRole(user: AuthUserView, roles: readonly UserRole[]): void {
	if (!roles.includes(user.role)) {
		throw new AuthError("forbidden", "没有权限执行此操作");
	}
}

/** 组管理权:组 owner 或 admin。 */
export function canManageGroup(user: AuthUserView, group: CourseGroupRecord): boolean {
	return user.role === "admin" || group.ownerId === user.id;
}
