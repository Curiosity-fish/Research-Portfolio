import { createHash, randomBytes, scryptSync, timingSafeEqual } from "node:crypto";

import type { Context } from "hono";
import type { UserRole, UserStatus } from "./db/types.ts";

const SCRYPT_KEYLEN = 32;
const SCRYPT_PREFIX = "scrypt";

/** 登录会话有效期:7 天 */
export const AUTH_SESSION_TTL_MS = 7 * 24 * 60 * 60 * 1000;
export type AuthErrorCode = "unauthorized" | "forbidden";

export class AuthError extends Error {
	readonly code: AuthErrorCode;

	constructor(code: AuthErrorCode, message: string) {
		super(message);
		this.name = "AuthError";
		this.code = code;
	}
}

export interface AuthUserView {
	id: string;
	username: string;
	role: UserRole;
	status: UserStatus;
}

/** scrypt 加盐哈希,存储格式 scrypt$<salt hex>$<hash hex>。 */
export function hashPassword(password: string): string {
	const salt = randomBytes(16).toString("hex");
	const hash = scryptSync(password, salt, SCRYPT_KEYLEN).toString("hex");
	return `${SCRYPT_PREFIX}$${salt}$${hash}`;
}

export function verifyPassword(password: string, stored: string): boolean {
	const parts = stored.split("$");
	if (parts.length !== 3 || parts[0] !== SCRYPT_PREFIX) return false;
	const [, salt, hashHex] = parts;
	const expected = Buffer.from(hashHex, "hex");
	const actual = scryptSync(password, salt, SCRYPT_KEYLEN);
	if (actual.length !== expected.length) return false;
	return timingSafeEqual(actual, expected);
}

export function generateToken(): string {
	return randomBytes(32).toString("base64url");
}

export function hashToken(token: string): string {
	return createHash("sha256").update(token).digest("hex");
}

/** 提供 resolveAuth 的对象(ShellSessionService 结构性满足,避免循环依赖)。 */
export interface AuthResolver {
	resolveAuth(token: string): Promise<AuthUserView | undefined>;
}

/**
 * 从 Authorization: Bearer <token> 解析当前用户。
 * 缺失/失效/账号禁用时抛 unauthorized(由 app.onError 映射为 401)。
 */
export async function authenticateRequest(c: Context, resolver: AuthResolver): Promise<AuthUserView> {
	const header = c.req.header("authorization");
	const token = header?.startsWith("Bearer ") ? header.slice("Bearer ".length) : undefined;
	if (!token) throw new AuthError("unauthorized", "未登录");
	const user = await resolver.resolveAuth(token);
	if (!user) throw new AuthError("unauthorized", "登录已失效或账号被禁用");
	return user;
}
