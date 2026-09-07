import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
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
	const cwd = mkdtempSync(join(tmpdir(), "tfa-admin-test-"));
	cleanupDirs.push(cwd);
	const service = new ShellSessionService({ cwd, agentDir: join(cwd, "agent") });
	services.push(service);
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

function formRequest(app: ReturnType<typeof createApp>, url: string, token: string, form: FormData): Promise<Response> {
	const headers = new Headers();
	headers.set("Authorization", `Bearer ${token}`);
	return Promise.resolve(app.fetch(new Request(`http://test.local${url}`, { method: "POST", headers, body: form })));
}

function uploadForm(content: string, visibility: "private" | "group" | "public", groupId?: string): FormData {
	const form = new FormData();
	form.append("file", new File([content], "data.csv"));
	form.append("visibility", visibility);
	if (groupId) form.append("groupId", groupId);
	return form;
}

async function readJson(response: Response): Promise<Record<string, unknown>> {
	return (await response.json()) as Record<string, unknown>;
}

async function login(app: ReturnType<typeof createApp>, username: string, password: string): Promise<string> {
	const res = await jsonRequest(app, "/api/auth/login", { method: "POST", body: { username, password } });
	const body = await readJson(res);
	return body.token as string;
}

async function createUser(
	app: ReturnType<typeof createApp>,
	adminToken: string,
	username: string,
	role: "admin" | "teacher" | "student",
): Promise<{ id: string }> {
	const res = await jsonRequest(app, "/api/admin/users", {
		method: "POST",
		token: adminToken,
		body: { username, password: "pass123", role },
	});
	const body = await readJson(res);
	return body.user as { id: string };
}

async function createGroup(app: ReturnType<typeof createApp>, token: string, name: string): Promise<{ id: string }> {
	const res = await jsonRequest(app, "/api/groups", { method: "POST", token, body: { name, type: "course" } });
	const body = await readJson(res);
	return body.group as { id: string };
}

describe("管理端 (FR-7)", () => {
	it("用户治理:建号/禁用/重置密码立即生效", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const student = await createUser(app, adminToken, "student1", "student");

		// 禁用学生 → 立即无法登录
		const disable = await jsonRequest(app, `/api/admin/users/${student.id}`, {
			method: "PATCH",
			token: adminToken,
			body: { status: "disabled" },
		});
		expect(disable.status).toBe(200);
		const disabledLogin = await jsonRequest(app, "/api/auth/login", {
			method: "POST",
			body: { username: "student1", password: "pass123" },
		});
		expect(disabledLogin.status).toBe(401);

		// 重置老师密码 → 旧密码失效、新密码可用
		const teacher = await createUser(app, adminToken, "teacher2", "teacher");
		const reset = await jsonRequest(app, `/api/admin/users/${teacher.id}`, {
			method: "PATCH",
			token: adminToken,
			body: { password: "newpass456" },
		});
		expect(reset.status).toBe(200);
		const oldLogin = await jsonRequest(app, "/api/auth/login", {
			method: "POST",
			body: { username: "teacher2", password: "pass123" },
		});
		expect(oldLogin.status).toBe(401);
		const newLogin = await jsonRequest(app, "/api/auth/login", {
			method: "POST",
			body: { username: "teacher2", password: "newpass456" },
		});
		expect(newLogin.status).toBe(200);
	});

	it("非管理员访问管理接口 403", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t1 = await login(app, "teacher1", "pass123");

		expect((await jsonRequest(app, "/api/admin/users", { token: t1 })).status).toBe(403);

		// 老师建公开数据集,再以老师身份调审核接口 → 403
		const created = await formRequest(app, "/api/datasets", t1, uploadForm("data", "public"));
		const datasetId = ((await readJson(created)).dataset as { id: string }).id;
		const reviewDenied = await jsonRequest(app, `/api/admin/datasets/${datasetId}/review`, {
			method: "POST",
			token: t1,
			body: { reviewState: "approved" },
		});
		expect(reviewDenied.status).toBe(403);
	});

	it("组治理:admin 可归档任意组并任免组长", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const s2 = await createUser(app, adminToken, "student2", "student");
		const t1 = await login(app, "teacher1", "pass123");

		const group = await createGroup(app, t1, "课程组");
		await jsonRequest(app, `/api/groups/${group.id}/members`, { method: "POST", token: t1, body: { userId: s1.id } });
		await jsonRequest(app, `/api/groups/${group.id}/members`, { method: "POST", token: t1, body: { userId: s2.id } });

		// admin 归档非本人组
		const archive = await jsonRequest(app, `/api/groups/${group.id}`, { method: "DELETE", token: adminToken });
		expect(archive.status).toBe(200);
		expect(((await readJson(archive)).group as { status: string }).status).toBe("archived");

		// admin 在归档组任免组长(s2 已是成员)
		const promote = await jsonRequest(app, `/api/groups/${group.id}/members/${s2.id}`, {
			method: "PATCH",
			token: adminToken,
			body: { role: "owner" },
		});
		expect(promote.status).toBe(200);
		const detail = (await readJson(await jsonRequest(app, `/api/groups/${group.id}`, { token: adminToken }))) as {
			group: { ownerId: string };
		};
		expect(detail.group.ownerId).toBe(s2.id);
	});

	it("资源治理:admin 下架公开技能/数据集后全员不可见", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "student1", "student");
		const t1 = await login(app, "teacher1", "pass123");
		const s1Token = await login(app, "student1", "pass123");

		// 老师发布公开技能 + 公开数据集
		const skillRes = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t1,
			body: { name: "公开技能", visibility: "public", upload: true },
		});
		const skillId = ((await readJson(skillRes)).skill as { id: string }).id;
		const dsRes = await formRequest(app, "/api/datasets", t1, uploadForm("data", "public"));
		const dsId = ((await readJson(dsRes)).dataset as { id: string }).id;

		// 下架前全员可见
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/skills", { token: s1Token }))) as {
					skills: Array<{ id: string }>;
				}
			).skills.map((s) => s.id),
		).toContain(skillId);
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/datasets", { token: s1Token }))) as {
					datasets: Array<{ id: string }>;
				}
			).datasets.map((d) => d.id),
		).toContain(dsId);

		// admin 下架 → 全员不可见
		expect((await jsonRequest(app, `/api/skills/${skillId}`, { method: "DELETE", token: adminToken })).status).toBe(
			200,
		);
		expect((await jsonRequest(app, `/api/datasets/${dsId}`, { method: "DELETE", token: adminToken })).status).toBe(
			200,
		);
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/skills", { token: s1Token }))) as {
					skills: Array<{ id: string }>;
				}
			).skills.map((s) => s.id),
		).not.toContain(skillId);
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/datasets", { token: s1Token }))) as {
					datasets: Array<{ id: string }>;
				}
			).datasets.map((d) => d.id),
		).not.toContain(dsId);
	});

	it("审核:admin 审核公开技能与数据集 pending → approved", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t1 = await login(app, "teacher1", "pass123");

		// 公开技能 → pending
		const skillRes = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t1,
			body: { name: "公开技能", visibility: "public", upload: true },
		});
		const skill = (await readJson(skillRes)).skill as { id: string; reviewState: string };
		expect(skill.reviewState).toBe("pending");

		const skillReview = await jsonRequest(app, `/api/admin/skills/${skill.id}/review`, {
			method: "POST",
			token: adminToken,
			body: { reviewState: "approved" },
		});
		expect(skillReview.status).toBe(200);
		expect(((await readJson(skillReview)).skill as { reviewState: string }).reviewState).toBe("approved");

		// 公开数据集 → pending
		const dsRes = await formRequest(app, "/api/datasets", t1, uploadForm("data", "public"));
		const ds = (await readJson(dsRes)).dataset as { id: string; reviewState: string };
		expect(ds.reviewState).toBe("pending");

		const dsReview = await jsonRequest(app, `/api/admin/datasets/${ds.id}/review`, {
			method: "POST",
			token: adminToken,
			body: { reviewState: "approved" },
		});
		expect(dsReview.status).toBe(200);
		expect(((await readJson(dsReview)).dataset as { reviewState: string }).reviewState).toBe("approved");
	});
});
