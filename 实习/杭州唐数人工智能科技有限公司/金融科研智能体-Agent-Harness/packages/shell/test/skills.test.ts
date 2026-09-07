import { execFileSync } from "node:child_process";
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
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
	const cwd = mkdtempSync(join(tmpdir(), "tfa-skills-test-"));
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

// ---- 本地 git fixture(离线验证 clone / pull) ----

function runGit(args: string[], cwd?: string): void {
	execFileSync("git", args, {
		cwd,
		stdio: "ignore",
		env: {
			...process.env,
			GIT_AUTHOR_NAME: "fixture",
			GIT_AUTHOR_EMAIL: "fixture@test",
			GIT_COMMITTER_NAME: "fixture",
			GIT_COMMITTER_EMAIL: "fixture@test",
		},
	});
}

function createFixtureRepo(parent: string): string {
	mkdirSync(parent, { recursive: true });
	const dir = join(parent, "repo");
	runGit(["init", "-q", dir]);
	writeFileSync(join(dir, "SKILL.md"), "# fixture skill\n");
	runGit(["-C", dir, "add", "SKILL.md"]);
	runGit(["-C", dir, "commit", "-qm", "init"]);
	return dir;
}

function addCommit(dir: string): void {
	writeFileSync(join(dir, "SKILL.md"), "# fixture skill v2\n");
	runGit(["-C", dir, "add", "SKILL.md"]);
	runGit(["-C", dir, "commit", "-qm", "v2"]);
}

describe("技能库 API (FR-4)", () => {
	it("repoUrl 安装成功,公开置 pending,git 更新版本变化;upload 型不可 git 更新", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t = await login(app, "teacher1", "pass123");

		// 未认证 401
		expect((await jsonRequest(app, "/api/skills")).status).toBe(401);

		const fixtureParent = mkdtempSync(join(tmpdir(), "tfa-skill-fixture-"));
		cleanupDirs.push(fixtureParent);
		const fixture = createFixtureRepo(fixtureParent);

		// git 安装
		const res = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t,
			body: { name: "pdf-extract", repoUrl: fixture, visibility: "public" },
		});
		expect(res.status).toBe(201);
		const skill = (await readJson(res)).skill as {
			id: string;
			repoUrl: string | null;
			version: string | null;
			reviewState: string;
		};
		expect(skill.repoUrl).toBe(fixture);
		expect(skill.version).toBeTruthy();
		expect(skill.reviewState).toBe("pending");

		// 更新拉取新版本
		addCommit(fixture);
		const updated = await jsonRequest(app, `/api/skills/${skill.id}/update`, { method: "POST", token: t });
		expect(updated.status).toBe(200);
		const updatedSkill = (await readJson(updated)).skill as { version: string | null };
		expect(updatedSkill.version).not.toBe(skill.version);

		// upload 型
		const uploadRes = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t,
			body: { name: "自研技能", visibility: "private", upload: true },
		});
		expect(uploadRes.status).toBe(201);
		const uploadSkill = (await readJson(uploadRes)).skill as { id: string; repoUrl: string | null };
		expect(uploadSkill.repoUrl).toBeNull();
		const updateUpload = await jsonRequest(app, `/api/skills/${uploadSkill.id}/update`, { method: "POST", token: t });
		expect(updateUpload.status).toBe(400);

		// 缺 repoUrl/upload → 400
		const bad = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t,
			body: { name: "x", visibility: "private" },
		});
		expect(bad.status).toBe(400);
	});

	it("可见性:私有仅 owner,组内可见,公开全员,下架后不可见", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "teacher2", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		await createUser(app, adminToken, "student2", "student");
		const t1 = await login(app, "teacher1", "pass123");
		const t2 = await login(app, "teacher2", "pass123");
		const s1Token = await login(app, "student1", "pass123");
		const s2Token = await login(app, "student2", "pass123");

		// 私有技能
		const privRes = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t1,
			body: { name: "私有技能", visibility: "private", upload: true },
		});
		const privId = ((await readJson(privRes)).skill as { id: string }).id;
		const t2List = (await readJson(await jsonRequest(app, "/api/skills", { token: t2 }))) as {
			skills: Array<{ id: string }>;
		};
		expect(t2List.skills.map((s) => s.id)).not.toContain(privId);

		// 组技能:组内可见,组外不可见
		const group = await createGroup(app, t1, "课程组");
		await jsonRequest(app, `/api/groups/${group.id}/members`, { method: "POST", token: t1, body: { userId: s1.id } });
		const groupSkillRes = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t1,
			body: { name: "组技能", visibility: "group", groupId: group.id, upload: true },
		});
		const groupSkillId = ((await readJson(groupSkillRes)).skill as { id: string }).id;
		const s1List = (await readJson(await jsonRequest(app, "/api/skills", { token: s1Token }))) as {
			skills: Array<{ id: string }>;
		};
		expect(s1List.skills.map((s) => s.id)).toContain(groupSkillId);
		const s2List = (await readJson(await jsonRequest(app, "/api/skills", { token: s2Token }))) as {
			skills: Array<{ id: string }>;
		};
		expect(s2List.skills.map((s) => s.id)).not.toContain(groupSkillId);

		// 公开技能:全员可见
		const pubRes = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t1,
			body: { name: "公开技能", visibility: "public", upload: true },
		});
		const pubId = ((await readJson(pubRes)).skill as { id: string }).id;
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/skills", { token: t2 }))) as { skills: Array<{ id: string }> }
			).skills.map((s) => s.id),
		).toContain(pubId);
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/skills", { token: s2Token }))) as {
					skills: Array<{ id: string }>;
				}
			).skills.map((s) => s.id),
		).toContain(pubId);

		// owner 下架公开技能 → 不可见
		const del = await jsonRequest(app, `/api/skills/${pubId}`, { method: "DELETE", token: t1 });
		expect(del.status).toBe(200);
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/skills", { token: t2 }))) as { skills: Array<{ id: string }> }
			).skills.map((s) => s.id),
		).not.toContain(pubId);

		// admin 下架组技能 → 组内不可见
		const delGroup = await jsonRequest(app, `/api/skills/${groupSkillId}`, { method: "DELETE", token: adminToken });
		expect(delGroup.status).toBe(200);
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/skills", { token: s1Token }))) as {
					skills: Array<{ id: string }>;
				}
			).skills.map((s) => s.id),
		).not.toContain(groupSkillId);
	});

	it("组技能权限:非组成员创建 403,归档组创建 400,归档后成员仍可见(只读)", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "teacher2", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const t1 = await login(app, "teacher1", "pass123");
		const t2 = await login(app, "teacher2", "pass123");
		const s1Token = await login(app, "student1", "pass123");

		const group = await createGroup(app, t1, "课程组");
		await jsonRequest(app, `/api/groups/${group.id}/members`, { method: "POST", token: t1, body: { userId: s1.id } });

		// 非组成员创建组技能 403
		const denied = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t2,
			body: { name: "越权", visibility: "group", groupId: group.id, upload: true },
		});
		expect(denied.status).toBe(403);

		// 组成员可在组内共享技能
		const shared = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: s1Token,
			body: { name: "学生共享技能", visibility: "group", groupId: group.id, upload: true },
		});
		expect(shared.status).toBe(201);
		const sharedId = ((await readJson(shared)).skill as { id: string }).id;

		// 归档
		await jsonRequest(app, `/api/groups/${group.id}`, { method: "DELETE", token: t1 });

		// 归档组不可新增组技能
		const afterArchive = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t1,
			body: { name: "归档新增", visibility: "group", groupId: group.id, upload: true },
		});
		expect(afterArchive.status).toBe(400);

		// 归档后成员仍可见组技能;非 owner 成员(t1)改被拒 → 只读
		const s1List = (await readJson(await jsonRequest(app, "/api/skills", { token: s1Token }))) as {
			skills: Array<{ id: string }>;
		};
		expect(s1List.skills.map((s) => s.id)).toContain(sharedId);
		const patchDenied = await jsonRequest(app, `/api/skills/${sharedId}`, {
			method: "PATCH",
			token: t1,
			body: { name: "x" },
		});
		expect(patchDenied.status).toBe(403);

		// owner(s1) 在归档后仍可管理自己创建的组技能
		const ownerPatch = await jsonRequest(app, `/api/skills/${sharedId}`, {
			method: "PATCH",
			token: s1Token,
			body: { name: "学生共享技能v2" },
		});
		expect(ownerPatch.status).toBe(200);
	});

	it("改名/改可见性触发 pending;非 owner 403;admin 审核与治理", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const t1 = await login(app, "teacher1", "pass123");
		const s1Token = await login(app, "student1", "pass123");

		const group = await createGroup(app, t1, "课程组");
		await jsonRequest(app, `/api/groups/${group.id}/members`, { method: "POST", token: t1, body: { userId: s1.id } });

		const created = await jsonRequest(app, "/api/skills", {
			method: "POST",
			token: t1,
			body: { name: "技能A", visibility: "private", upload: true },
		});
		const skillId = ((await readJson(created)).skill as { id: string }).id;

		// 改名
		const renamed = await jsonRequest(app, `/api/skills/${skillId}`, {
			method: "PATCH",
			token: t1,
			body: { name: "技能A2" },
		});
		expect(renamed.status).toBe(200);
		expect(((await readJson(renamed)).skill as { name: string }).name).toBe("技能A2");

		// 改可见性 group(带 groupId)
		const toGroup = await jsonRequest(app, `/api/skills/${skillId}`, {
			method: "PATCH",
			token: t1,
			body: { visibility: "group", groupId: group.id },
		});
		expect(toGroup.status).toBe(200);
		const groupSkill = (await readJson(toGroup)).skill as { groupId: string | null; groupName: string | undefined };
		expect(groupSkill.groupId).toBe(group.id);
		expect(groupSkill.groupName).toBe("课程组");

		// 改 public → pending
		const toPublic = await jsonRequest(app, `/api/skills/${skillId}`, {
			method: "PATCH",
			token: t1,
			body: { visibility: "public" },
		});
		expect(toPublic.status).toBe(200);
		expect(((await readJson(toPublic)).skill as { reviewState: string }).reviewState).toBe("pending");

		// 非 owner(组内学生)不能改
		const memberPatch = await jsonRequest(app, `/api/skills/${skillId}`, {
			method: "PATCH",
			token: s1Token,
			body: { name: "x" },
		});
		expect(memberPatch.status).toBe(403);

		// admin 审核 pending → approved
		const review = await jsonRequest(app, `/api/admin/skills/${skillId}/review`, {
			method: "POST",
			token: adminToken,
			body: { reviewState: "approved" },
		});
		expect(review.status).toBe(200);
		expect(((await readJson(review)).skill as { reviewState: string }).reviewState).toBe("approved");

		// admin 可治理任意技能
		const adminPatch = await jsonRequest(app, `/api/skills/${skillId}`, {
			method: "PATCH",
			token: adminToken,
			body: { name: "管理员改名" },
		});
		expect(adminPatch.status).toBe(200);
		expect(((await readJson(adminPatch)).skill as { name: string }).name).toBe("管理员改名");

		// 非 admin 调审核接口 403
		const reviewDenied = await jsonRequest(app, `/api/admin/skills/${skillId}/review`, {
			method: "POST",
			token: t1,
			body: { reviewState: "approved" },
		});
		expect(reviewDenied.status).toBe(403);
	});
});
