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
	const cwd = mkdtempSync(join(tmpdir(), "tfa-groups-test-"));
	cleanupDirs.push(cwd);
	const service = new ShellSessionService({ cwd, agentDir: join(cwd, "agent") });
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
): Promise<{ id: string; username: string; role: string }> {
	const res = await jsonRequest(app, "/api/admin/users", {
		method: "POST",
		token: adminToken,
		body: { username, password: "pass123", role },
	});
	const body = await readJson(res);
	return body.user as { id: string; username: string; role: string };
}

async function createGroup(
	app: ReturnType<typeof createApp>,
	token: string,
	name: string,
	type: "course" | "research",
): Promise<{ id: string; ownerId: string }> {
	const res = await jsonRequest(app, "/api/groups", { method: "POST", token, body: { name, type } });
	const body = await readJson(res);
	return body.group as { id: string; ownerId: string };
}

describe("课程组 API (FR-2)", () => {
	it("老师建组成功(owner 计入成员),学生建组 403,未登录 401,空名称 400", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		const teacher = await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "student1", "student");

		const teacherToken = await login(app, "teacher1", "pass123");
		const res = await jsonRequest(app, "/api/groups", {
			method: "POST",
			token: teacherToken,
			body: { name: "金融数据分析", type: "course" },
		});
		expect(res.status).toBe(201);
		const group = (await readJson(res)).group as { id: string; ownerId: string; memberCount: number; myRole: string };
		expect(group.ownerId).toBe(teacher.id);
		expect(group.memberCount).toBe(1);
		expect(group.myRole).toBe("owner");

		const studentToken = await login(app, "student1", "pass123");
		const denied = await jsonRequest(app, "/api/groups", {
			method: "POST",
			token: studentToken,
			body: { name: "x", type: "course" },
		});
		expect(denied.status).toBe(403);

		const unauth = await jsonRequest(app, "/api/groups");
		expect(unauth.status).toBe(401);

		const bad = await jsonRequest(app, "/api/groups", {
			method: "POST",
			token: teacherToken,
			body: { name: "  ", type: "course" },
		});
		expect(bad.status).toBe(400);
	});

	it("列表可见性:非成员不可见,admin 可见全部;非成员看详情 403", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "teacher2", "teacher");
		const t1 = await login(app, "teacher1", "pass123");
		const t2 = await login(app, "teacher2", "pass123");

		const groupId = (await createGroup(app, t1, "组A", "research")).id;

		const list1 = (await readJson(await jsonRequest(app, "/api/groups", { token: t1 }))) as {
			groups: Array<{ id: string }>;
		};
		expect(list1.groups.map((g) => g.id)).toContain(groupId);

		const list2 = (await readJson(await jsonRequest(app, "/api/groups", { token: t2 }))) as {
			groups: Array<{ id: string }>;
		};
		expect(list2.groups).toHaveLength(0);

		const listAdmin = (await readJson(await jsonRequest(app, "/api/groups", { token: adminToken }))) as {
			groups: Array<{ id: string }>;
		};
		expect(listAdmin.groups.map((g) => g.id)).toContain(groupId);

		const detailDenied = await jsonRequest(app, `/api/groups/${groupId}`, { token: t2 });
		expect(detailDenied.status).toBe(403);

		const detail = await jsonRequest(app, `/api/groups/${groupId}`, { token: t1 });
		expect(detail.status).toBe(200);
		const detailBody = (await readJson(detail)).group as { members: unknown[] };
		expect(detailBody.members).toHaveLength(1);
	});

	it("添加/移除成员:重复 400、不存在 404、非组长 403、不能移除组长", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		const teacher = await createUser(app, adminToken, "teacher1", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const s2 = await createUser(app, adminToken, "student2", "student");
		const t = await login(app, "teacher1", "pass123");

		const groupId = (await createGroup(app, t, "组", "course")).id;

		const add = await jsonRequest(app, `/api/groups/${groupId}/members`, {
			method: "POST",
			token: t,
			body: { userId: s1.id },
		});
		expect(add.status).toBe(201);

		const dup = await jsonRequest(app, `/api/groups/${groupId}/members`, {
			method: "POST",
			token: t,
			body: { userId: s1.id },
		});
		expect(dup.status).toBe(400);

		const missing = await jsonRequest(app, `/api/groups/${groupId}/members`, {
			method: "POST",
			token: t,
			body: { userId: "no-such-user" },
		});
		expect(missing.status).toBe(404);

		const detail = (await readJson(await jsonRequest(app, `/api/groups/${groupId}`, { token: t }))) as {
			group: { members: Array<{ userId: string; username: string; role: string }> };
		};
		expect(detail.group.members.find((m) => m.userId === s1.id)?.username).toBe("student1");

		// 学生加入后可见该组(myRole=member)
		const s1Token = await login(app, "student1", "pass123");
		const s1List = (await readJson(await jsonRequest(app, "/api/groups", { token: s1Token }))) as {
			groups: Array<{ id: string; myRole: string }>;
		};
		expect(s1List.groups.find((g) => g.id === groupId)?.myRole).toBe("member");

		// 非组长(其他学生)加人 403
		const s2Token = await login(app, s2.username, "pass123");
		const denied = await jsonRequest(app, `/api/groups/${groupId}/members`, {
			method: "POST",
			token: s2Token,
			body: { userId: s1.id },
		});
		expect(denied.status).toBe(403);

		// 不能移除组长
		const removeOwner = await jsonRequest(app, `/api/groups/${groupId}/members/${teacher.id}`, {
			method: "DELETE",
			token: t,
		});
		expect(removeOwner.status).toBe(400);

		// 移除成员后学生不可见
		const remove = await jsonRequest(app, `/api/groups/${groupId}/members/${s1.id}`, {
			method: "DELETE",
			token: t,
		});
		expect(remove.status).toBe(200);
		const s1ListAfter = (await readJson(await jsonRequest(app, "/api/groups", { token: s1Token }))) as {
			groups: unknown[];
		};
		expect(s1ListAfter.groups).toHaveLength(0);
	});

	it("归档:不可加人,成员仍可见,非组长不可归档", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "teacher2", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const s2 = await createUser(app, adminToken, "student2", "student");
		const t1 = await login(app, "teacher1", "pass123");
		const t2 = await login(app, "teacher2", "pass123");

		const groupId = (await createGroup(app, t1, "组", "course")).id;
		await jsonRequest(app, `/api/groups/${groupId}/members`, { method: "POST", token: t1, body: { userId: s1.id } });

		// 非组长归档 403
		const denied = await jsonRequest(app, `/api/groups/${groupId}`, { method: "DELETE", token: t2 });
		expect(denied.status).toBe(403);

		const archived = await jsonRequest(app, `/api/groups/${groupId}`, { method: "DELETE", token: t1 });
		expect(archived.status).toBe(200);
		expect(((await readJson(archived)).group as { status: string }).status).toBe("archived");

		// 归档后不可加人
		const addAfter = await jsonRequest(app, `/api/groups/${groupId}/members`, {
			method: "POST",
			token: t1,
			body: { userId: s2.id },
		});
		expect(addAfter.status).toBe(400);

		// 现有成员仍可见(列表含归档组)
		const s1Token = await login(app, "student1", "pass123");
		const s1List = (await readJson(await jsonRequest(app, "/api/groups", { token: s1Token }))) as {
			groups: Array<{ id: string; status: string }>;
		};
		expect(s1List.groups.find((g) => g.id === groupId)?.status).toBe("archived");
	});

	it("任免组长:晋升后新组长可管理,旧组长失去管理权,不能直接降级组长", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		const teacher = await createUser(app, adminToken, "teacher1", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const s2 = await createUser(app, adminToken, "student2", "student");
		const t = await login(app, "teacher1", "pass123");

		const groupId = (await createGroup(app, t, "组", "course")).id;
		await jsonRequest(app, `/api/groups/${groupId}/members`, { method: "POST", token: t, body: { userId: s1.id } });

		// 不能直接降级组长
		const demoteOwner = await jsonRequest(app, `/api/groups/${groupId}/members/${teacher.id}`, {
			method: "PATCH",
			token: t,
			body: { role: "member" },
		});
		expect(demoteOwner.status).toBe(400);

		// 晋升 s1 为组长
		const promote = await jsonRequest(app, `/api/groups/${groupId}/members/${s1.id}`, {
			method: "PATCH",
			token: t,
			body: { role: "owner" },
		});
		expect(promote.status).toBe(200);

		// 新组长可管理(加人)
		const s1Token = await login(app, "student1", "pass123");
		const addByNewOwner = await jsonRequest(app, `/api/groups/${groupId}/members`, {
			method: "POST",
			token: s1Token,
			body: { userId: s2.id },
		});
		expect(addByNewOwner.status).toBe(201);

		// 旧组长失去管理权(改名 403)
		const renameDenied = await jsonRequest(app, `/api/groups/${groupId}`, {
			method: "PATCH",
			token: t,
			body: { name: "改名" },
		});
		expect(renameDenied.status).toBe(403);

		// 新组长改名成功
		const rename = await jsonRequest(app, `/api/groups/${groupId}`, {
			method: "PATCH",
			token: s1Token,
			body: { name: "新组名" },
		});
		expect(rename.status).toBe(200);
		expect(((await readJson(rename)).group as { name: string }).name).toBe("新组名");
	});

	it("成员不能改名;admin 可归档任意组并任免组长", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const s2 = await createUser(app, adminToken, "student2", "student");
		const t = await login(app, "teacher1", "pass123");

		const groupId = (await createGroup(app, t, "组", "course")).id;
		await jsonRequest(app, `/api/groups/${groupId}/members`, { method: "POST", token: t, body: { userId: s1.id } });
		await jsonRequest(app, `/api/groups/${groupId}/members`, { method: "POST", token: t, body: { userId: s2.id } });

		// 成员(学生)不能改名
		const s1Token = await login(app, "student1", "pass123");
		const memberRename = await jsonRequest(app, `/api/groups/${groupId}`, {
			method: "PATCH",
			token: s1Token,
			body: { name: "x" },
		});
		expect(memberRename.status).toBe(403);

		// admin 归档任意组
		const archive = await jsonRequest(app, `/api/groups/${groupId}`, { method: "DELETE", token: adminToken });
		expect(archive.status).toBe(200);
		expect(((await readJson(archive)).group as { status: string }).status).toBe("archived");

		// admin 在归档组任免组长(治理能力保留)
		const promoteByAdmin = await jsonRequest(app, `/api/groups/${groupId}/members/${s2.id}`, {
			method: "PATCH",
			token: adminToken,
			body: { role: "owner" },
		});
		expect(promoteByAdmin.status).toBe(200);
		const detail = (await readJson(await jsonRequest(app, `/api/groups/${groupId}`, { token: adminToken }))) as {
			group: { ownerId: string };
		};
		expect(detail.group.ownerId).toBe(s2.id);
	});
});
