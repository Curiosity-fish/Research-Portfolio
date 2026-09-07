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
	const cwd = mkdtempSync(join(tmpdir(), "tfa-projects-test-"));
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

async function createGroup(app: ReturnType<typeof createApp>, token: string, name: string): Promise<{ id: string }> {
	const res = await jsonRequest(app, "/api/groups", { method: "POST", token, body: { name, type: "course" } });
	const body = await readJson(res);
	return body.group as { id: string };
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

describe("项目作用域 (FR-3)", () => {
	it("私有项目仅 owner 可见;非 owner 修改/删除被拒;未认证 401", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		const teacher = await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "student1", "student");
		const t = await login(app, "teacher1", "pass123");
		const s = await login(app, "student1", "pass123");

		// 未认证 401
		expect((await jsonRequest(app, "/api/projects")).status).toBe(401);

		const created = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t,
			body: { name: "私有研究", scope: { type: "private" } },
		});
		expect(created.status).toBe(201);
		const project = (await readJson(created)).project as {
			id: string;
			ownerId: string;
			scope: { type: string };
			sessionIds: string[];
		};
		expect(project.ownerId).toBe(teacher.id);
		expect(project.scope.type).toBe("private");
		expect(project.sessionIds).toEqual([]);

		// 非 owner 列表看不到
		const sList = (await readJson(await jsonRequest(app, "/api/projects", { token: s }))) as {
			projects: unknown[];
		};
		expect(sList.projects).toHaveLength(0);

		// 非 owner 改/删被拒
		const patchDenied = await jsonRequest(app, `/api/projects/${project.id}`, {
			method: "PATCH",
			token: s,
			body: { name: "x" },
		});
		expect(patchDenied.status).toBe(403);
		const deleteDenied = await jsonRequest(app, `/api/projects/${project.id}`, { method: "DELETE", token: s });
		expect(deleteDenied.status).toBe(403);

		// owner 改名成功
		const renamed = await jsonRequest(app, `/api/projects/${project.id}`, {
			method: "PATCH",
			token: t,
			body: { name: "私有研究V2" },
		});
		expect(renamed.status).toBe(200);
		expect(((await readJson(renamed)).project as { name: string }).name).toBe("私有研究V2");
	});

	it("挂组项目:组内成员可见,组外成员不可见/不可创建", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "teacher2", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const s2 = await createUser(app, adminToken, "student2", "student");
		const t1 = await login(app, "teacher1", "pass123");
		const t2 = await login(app, "teacher2", "pass123");
		const s1Token = await login(app, "student1", "pass123");
		const s2Token = await login(app, s2.username, "pass123");

		const group = await createGroup(app, t1, "金融数据分析");
		await jsonRequest(app, `/api/groups/${group.id}/members`, {
			method: "POST",
			token: t1,
			body: { userId: s1.id },
		});

		// 组长创建挂组项目
		const created = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t1,
			body: { name: "组内课题", scope: { type: "group", groupId: group.id } },
		});
		expect(created.status).toBe(201);
		const project = (await readJson(created)).project as {
			id: string;
			scope: { type: string; groupId: string; groupName: string };
		};
		expect(project.scope).toEqual({ type: "group", groupId: group.id, groupName: "金融数据分析" });

		// 组内成员可见
		const s1List = (await readJson(await jsonRequest(app, "/api/projects", { token: s1Token }))) as {
			projects: Array<{ id: string; scope: { type: string; groupId: string } }>;
		};
		expect(s1List.projects.map((p) => p.id)).toContain(project.id);
		expect(s1List.projects.find((p) => p.id === project.id)?.scope.groupId).toBe(group.id);

		// 组外学生/老师不可见
		const s2List = (await readJson(await jsonRequest(app, "/api/projects", { token: s2Token }))) as {
			projects: Array<{ id: string }>;
		};
		expect(s2List.projects).toHaveLength(0);
		const t2List = (await readJson(await jsonRequest(app, "/api/projects", { token: t2 }))) as {
			projects: Array<{ id: string }>;
		};
		expect(t2List.projects).toHaveLength(0);

		// 组外成员在该组下创建项目被拒
		const createDeniedS = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: s2Token,
			body: { name: "越权项目", scope: { type: "group", groupId: group.id } },
		});
		expect(createDeniedS.status).toBe(403);
		const createDeniedT = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t2,
			body: { name: "越权项目", scope: { type: "group", groupId: group.id } },
		});
		expect(createDeniedT.status).toBe(403);
	});

	it("学生可创建挂组项目(A4 默认允许)", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const t = await login(app, "teacher1", "pass123");

		const group = await createGroup(app, t, "课程组");
		await jsonRequest(app, `/api/groups/${group.id}/members`, {
			method: "POST",
			token: t,
			body: { userId: s1.id },
		});

		const s1Token = await login(app, "student1", "pass123");
		const created = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: s1Token,
			body: { name: "学生课题", scope: { type: "group", groupId: group.id } },
		});
		expect(created.status).toBe(201);
		const project = (await readJson(created)).project as { ownerId: string; scope: { type: string } };
		expect(project.ownerId).toBe(s1.id);
		expect(project.scope.type).toBe("group");
	});

	it("归档组:不可新建挂组项目,历史项目对成员仍可见(只读);scope 不可修改", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const t = await login(app, "teacher1", "pass123");

		const group = await createGroup(app, t, "将被归档");
		await jsonRequest(app, `/api/groups/${group.id}/members`, {
			method: "POST",
			token: t,
			body: { userId: s1.id },
		});

		// 归档前创建挂组项目
		const groupProject = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t,
			body: { name: "组内历史课题", scope: { type: "group", groupId: group.id } },
		});
		expect(groupProject.status).toBe(201);
		const groupProjectId = ((await readJson(groupProject)).project as { id: string }).id;

		// 归档
		await jsonRequest(app, `/api/groups/${group.id}`, { method: "DELETE", token: t });

		// 归档组下新建项目被拒
		const createArchived = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t,
			body: { name: "归档组项目", scope: { type: "group", groupId: group.id } },
		});
		expect(createArchived.status).toBe(400);

		// 成员在归档后仍可见历史项目(只读)
		const s1Token = await login(app, "student1", "pass123");
		const s1List = (await readJson(await jsonRequest(app, "/api/projects", { token: s1Token }))) as {
			projects: Array<{ id: string }>;
		};
		expect(s1List.projects.map((p) => p.id)).toContain(groupProjectId);
		const patchDenied = await jsonRequest(app, `/api/projects/${groupProjectId}`, {
			method: "PATCH",
			token: s1Token,
			body: { name: "x" },
		});
		expect(patchDenied.status).toBe(403);

		// 已存在项目:scope 不可改
		const created = await jsonRequest(app, "/api/projects", { method: "POST", token: t, body: { name: "私有项目" } });
		const projectId = ((await readJson(created)).project as { id: string }).id;
		const patchScope = await jsonRequest(app, `/api/projects/${projectId}`, {
			method: "PATCH",
			token: t,
			body: { scope: { type: "group", groupId: group.id } },
		});
		expect(patchScope.status).toBe(400);

		// 不存在的组
		const createMissingGroup = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t,
			body: { name: "x", scope: { type: "group", groupId: "no-such-group" } },
		});
		expect(createMissingGroup.status).toBe(404);
	});

	it("admin 可见全部项目", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t = await login(app, "teacher1", "pass123");

		await jsonRequest(app, "/api/projects", { method: "POST", token: t, body: { name: "私有A" } });
		const group = await createGroup(app, t, "组B");
		await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t,
			body: { name: "挂组B", scope: { type: "group", groupId: group.id } },
		});

		const adminList = (await readJson(await jsonRequest(app, "/api/projects", { token: adminToken }))) as {
			projects: Array<{ name: string }>;
		};
		expect(adminList.projects.map((p) => p.name).sort()).toEqual(["挂组B", "私有A"]);
	});
	it("项目会话上下文:私有项目只见私有+公开,挂组项目含该组库", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		await createUser(app, adminToken, "student2", "student");
		const t1 = await login(app, "teacher1", "pass123");
		const s1Token = await login(app, "student1", "pass123");
		const s2Token = await login(app, "student2", "pass123");

		const group = await createGroup(app, t1, "课程组");
		await jsonRequest(app, `/api/groups/${group.id}/members`, { method: "POST", token: t1, body: { userId: s1.id } });

		const privSkill = (
			(
				await readJson(
					await jsonRequest(app, "/api/skills", {
						method: "POST",
						token: t1,
						body: { name: "私有技能", visibility: "private", upload: true },
					}),
				)
			).skill as { id: string }
		).id;
		const groupSkill = (
			(
				await readJson(
					await jsonRequest(app, "/api/skills", {
						method: "POST",
						token: t1,
						body: { name: "组技能", visibility: "group", groupId: group.id, upload: true },
					}),
				)
			).skill as { id: string }
		).id;
		const pubSkill = (
			(
				await readJson(
					await jsonRequest(app, "/api/skills", {
						method: "POST",
						token: t1,
						body: { name: "公开技能", visibility: "public", upload: true },
					}),
				)
			).skill as { id: string }
		).id;
		const privDs = (
			(await readJson(await formRequest(app, "/api/datasets", t1, uploadForm("p", "private")))).dataset as {
				id: string;
			}
		).id;
		const groupDs = (
			(await readJson(await formRequest(app, "/api/datasets", t1, uploadForm("g", "group", group.id)))).dataset as {
				id: string;
			}
		).id;
		const pubDs = (
			(await readJson(await formRequest(app, "/api/datasets", t1, uploadForm("u", "public")))).dataset as {
				id: string;
			}
		).id;

		const privRes = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t1,
			body: { name: "私有项目" },
		});
		const privProject = ((await readJson(privRes)).project as { id: string }).id;
		const privCtx = (await readJson(
			await jsonRequest(app, `/api/projects/${privProject}/session-context`, { token: t1 }),
		)) as {
			skills: Array<{ id: string }>;
			datasets: Array<{ id: string }>;
		};
		expect(privCtx.skills.map((s) => s.id).sort()).toEqual([privSkill, pubSkill].sort());
		expect(privCtx.datasets.map((d) => d.id).sort()).toEqual([privDs, pubDs].sort());

		const groupRes = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t1,
			body: { name: "挂组项目", scope: { type: "group", groupId: group.id } },
		});
		const groupProject = ((await readJson(groupRes)).project as { id: string }).id;
		const groupCtx = (await readJson(
			await jsonRequest(app, `/api/projects/${groupProject}/session-context`, { token: t1 }),
		)) as {
			skills: Array<{ id: string }>;
			datasets: Array<{ id: string }>;
		};
		expect(groupCtx.skills.map((s) => s.id).sort()).toEqual([privSkill, groupSkill, pubSkill].sort());
		expect(groupCtx.datasets.map((d) => d.id).sort()).toEqual([privDs, groupDs, pubDs].sort());

		const memberCtx = (await readJson(
			await jsonRequest(app, `/api/projects/${groupProject}/session-context`, { token: s1Token }),
		)) as {
			skills: Array<{ id: string }>;
			datasets: Array<{ id: string }>;
		};
		expect(memberCtx.skills.map((s) => s.id).sort()).toEqual([groupSkill, pubSkill].sort());
		expect(memberCtx.datasets.map((d) => d.id).sort()).toEqual([groupDs, pubDs].sort());

		const denied = await jsonRequest(app, `/api/projects/${groupProject}/session-context`, { token: s2Token });
		expect(denied.status).toBe(404);
	});
});
