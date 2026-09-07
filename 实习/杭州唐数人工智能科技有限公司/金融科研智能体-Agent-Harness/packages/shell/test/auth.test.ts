import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { setupFauxModel } from "../src/core/faux.ts";
import { hashPassword, verifyPassword } from "../src/server/auth.ts";
import { createApp } from "../src/server/http-server.ts";
import { ShellSessionService } from "../src/server/session-service.ts";

const services: ShellSessionService[] = [];
const cleanupDirs: string[] = [];

afterEach(async () => {
	for (const service of services) {
		await service.close().catch(() => {
			// 已关闭的句柄忽略
		});
	}
	services.length = 0;
	for (const dir of cleanupDirs) {
		rmSync(dir, { recursive: true, force: true });
	}
	cleanupDirs.length = 0;
});

async function createTestContext(): Promise<{ app: ReturnType<typeof createApp> }> {
	const cwd = mkdtempSync(join(tmpdir(), "tfa-auth-test-"));
	cleanupDirs.push(cwd);
	const faux = setupFauxModel();
	const service = new ShellSessionService({ cwd, agentDir: join(cwd, "agent"), fauxSetup: faux });
	services.push(service);
	// 种子管理员
	await service.createUser({ username: "admin", password: "admin123", role: "admin" });
	const app = createApp(service, { staticDir: join(cwd, "no-such-dir") });
	return { app };
}

function jsonRequest(
	app: ReturnType<typeof createApp>,
	url: string,
	options: { method?: string; body?: unknown; token?: string } = {},
): Promise<Response> {
	const headers = new Headers();
	let body: string | undefined;
	if (options.body !== undefined) {
		headers.set("Content-Type", "application/json");
		body = JSON.stringify(options.body);
	}
	if (options.token) {
		headers.set("Authorization", `Bearer ${options.token}`);
	}
	return Promise.resolve(
		app.fetch(new Request(`http://test.local${url}`, { method: options.method ?? "GET", headers, body })),
	);
}

async function readJson(response: Response): Promise<Record<string, unknown>> {
	return (await response.json()) as Record<string, unknown>;
}

describe("auth 密码哈希", () => {
	it("hashPassword/verifyPassword 往返", () => {
		const stored = hashPassword("secret123");
		expect(stored.startsWith("scrypt$")).toBe(true);
		expect(verifyPassword("secret123", stored)).toBe(true);
		expect(verifyPassword("wrong", stored)).toBe(false);
		expect(verifyPassword("secret123", "not-a-hash")).toBe(false);
	});
});

describe("认证 API", () => {
	it("登录成功返回 token 与用户", async () => {
		const { app } = await createTestContext();
		const res = await jsonRequest(app, "/api/auth/login", {
			method: "POST",
			body: { username: "admin", password: "admin123" },
		});
		expect(res.status).toBe(200);
		const body = await readJson(res);
		expect(typeof body.token).toBe("string");
		expect((body.user as { role: string }).role).toBe("admin");
	});

	it("密码错误登录返回 401", async () => {
		const { app } = await createTestContext();
		const res = await jsonRequest(app, "/api/auth/login", {
			method: "POST",
			body: { username: "admin", password: "wrong-pass" },
		});
		expect(res.status).toBe(401);
	});

	it("me 无 token 返回 401,带 token 返回用户", async () => {
		const { app } = await createTestContext();
		const unauth = await jsonRequest(app, "/api/auth/me");
		expect(unauth.status).toBe(401);

		const login = await readJson(
			await jsonRequest(app, "/api/auth/login", {
				method: "POST",
				body: { username: "admin", password: "admin123" },
			}),
		);
		const token = login.token as string;
		const me = await jsonRequest(app, "/api/auth/me", { token });
		expect(me.status).toBe(200);
		expect(((await readJson(me)).user as { username: string }).username).toBe("admin");
	});

	it("logout 后 token 失效", async () => {
		const { app } = await createTestContext();
		const login = await readJson(
			await jsonRequest(app, "/api/auth/login", {
				method: "POST",
				body: { username: "admin", password: "admin123" },
			}),
		);
		const token = login.token as string;

		const logout = await jsonRequest(app, "/api/auth/logout", { method: "POST", token });
		expect(logout.status).toBe(200);

		const me = await jsonRequest(app, "/api/auth/me", { token });
		expect(me.status).toBe(401);
	});

	it("禁用用户无法登录", async () => {
		const { app } = await createTestContext();
		const login = await readJson(
			await jsonRequest(app, "/api/auth/login", {
				method: "POST",
				body: { username: "admin", password: "admin123" },
			}),
		);
		const adminToken = login.token as string;

		const created = await readJson(
			await jsonRequest(app, "/api/admin/users", {
				method: "POST",
				token: adminToken,
				body: { username: "teacher1", password: "pass123", role: "teacher" },
			}),
		);
		const userId = (created.user as { id: string }).id;

		const disable = await jsonRequest(app, `/api/admin/users/${userId}`, {
			method: "PATCH",
			token: adminToken,
			body: { status: "disabled" },
		});
		expect(disable.status).toBe(200);

		const loginTeacher = await jsonRequest(app, "/api/auth/login", {
			method: "POST",
			body: { username: "teacher1", password: "pass123" },
		});
		expect(loginTeacher.status).toBe(401);
	});
});

describe("管理员用户 API", () => {
	it("未认证访问返回 401,非管理员返回 403", async () => {
		const { app } = await createTestContext();
		const unauth = await jsonRequest(app, "/api/admin/users");
		expect(unauth.status).toBe(401);

		// 建一个老师并登录
		const adminLogin = await readJson(
			await jsonRequest(app, "/api/auth/login", {
				method: "POST",
				body: { username: "admin", password: "admin123" },
			}),
		);
		const adminToken = adminLogin.token as string;
		await jsonRequest(app, "/api/admin/users", {
			method: "POST",
			token: adminToken,
			body: { username: "teacher1", password: "pass123", role: "teacher" },
		});
		const teacherLogin = await readJson(
			await jsonRequest(app, "/api/auth/login", {
				method: "POST",
				body: { username: "teacher1", password: "pass123" },
			}),
		);
		const forbidden = await jsonRequest(app, "/api/admin/users", { token: teacherLogin.token as string });
		expect(forbidden.status).toBe(403);
	});

	it("建号、列用户、重复用户名拒绝", async () => {
		const { app } = await createTestContext();
		const login = await readJson(
			await jsonRequest(app, "/api/auth/login", {
				method: "POST",
				body: { username: "admin", password: "admin123" },
			}),
		);
		const token = login.token as string;

		const created = await jsonRequest(app, "/api/admin/users", {
			method: "POST",
			token,
			body: { username: "student1", password: "pass123", role: "student" },
		});
		expect(created.status).toBe(201);
		expect(((await readJson(created)).user as { role: string }).role).toBe("student");

		const list = await readJson(await jsonRequest(app, "/api/admin/users", { token }));
		const users = list.users as Array<{ username: string }>;
		expect(users.map((u) => u.username).sort()).toEqual(["admin", "student1"]);

		const dup = await jsonRequest(app, "/api/admin/users", {
			method: "POST",
			token,
			body: { username: "student1", password: "pass123", role: "student" },
		});
		expect(dup.status).toBe(400);
	});

	it("重置密码后旧密码失效新密码可用", async () => {
		const { app } = await createTestContext();
		const adminLogin = await readJson(
			await jsonRequest(app, "/api/auth/login", {
				method: "POST",
				body: { username: "admin", password: "admin123" },
			}),
		);
		const adminToken = adminLogin.token as string;
		const created = await readJson(
			await jsonRequest(app, "/api/admin/users", {
				method: "POST",
				token: adminToken,
				body: { username: "teacher1", password: "pass123", role: "teacher" },
			}),
		);
		const userId = (created.user as { id: string }).id;

		const reset = await jsonRequest(app, `/api/admin/users/${userId}`, {
			method: "PATCH",
			token: adminToken,
			body: { password: "newpass456" },
		});
		expect(reset.status).toBe(200);

		const oldLogin = await jsonRequest(app, "/api/auth/login", {
			method: "POST",
			body: { username: "teacher1", password: "pass123" },
		});
		expect(oldLogin.status).toBe(401);

		const newLogin = await jsonRequest(app, "/api/auth/login", {
			method: "POST",
			body: { username: "teacher1", password: "newpass456" },
		});
		expect(newLogin.status).toBe(200);
	});
});
