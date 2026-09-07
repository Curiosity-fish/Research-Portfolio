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

async function createTestContext(
	options: { datasetMaxFileBytes?: number; datasetMaxUserBytes?: number } = {},
): Promise<{ app: ReturnType<typeof createApp> }> {
	const cwd = mkdtempSync(join(tmpdir(), "tfa-datasets-test-"));
	cleanupDirs.push(cwd);
	const service = new ShellSessionService({ cwd, agentDir: join(cwd, "agent"), ...options });
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

function uploadForm(
	content: string,
	overrides: { name?: string; visibility?: string; groupId?: string } = {},
): FormData {
	const form = new FormData();
	form.append("file", new File([content], "data.csv"));
	if (overrides.name) form.append("name", overrides.name);
	form.append("visibility", overrides.visibility ?? "private");
	if (overrides.groupId) form.append("groupId", overrides.groupId);
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

describe("数据集库 API (FR-5)", () => {
	it("上传私有/组/公开;可见性规则;下载只读", async () => {
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

		// 未认证 401
		expect((await jsonRequest(app, "/api/datasets")).status).toBe(401);

		// 私有
		const privRes = await formRequest(app, "/api/datasets", t1, uploadForm("private-data"));
		expect(privRes.status).toBe(201);
		const priv = (await readJson(privRes)).dataset as { id: string; sizeBytes: number; visibility: string };
		expect(priv.sizeBytes).toBe("private-data".length);
		expect(priv.visibility).toBe("private");
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/datasets", { token: t2 }))) as {
					datasets: Array<{ id: string }>;
				}
			).datasets.map((d) => d.id),
		).not.toContain(priv.id);

		// 组
		const group = await createGroup(app, t1, "课程组");
		await jsonRequest(app, `/api/groups/${group.id}/members`, { method: "POST", token: t1, body: { userId: s1.id } });
		const groupRes = await formRequest(
			app,
			"/api/datasets",
			t1,
			uploadForm("group-data", { visibility: "group", groupId: group.id }),
		);
		expect(groupRes.status).toBe(201);
		const groupDataset = (await readJson(groupRes)).dataset as { id: string };
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/datasets", { token: s1Token }))) as {
					datasets: Array<{ id: string }>;
				}
			).datasets.map((d) => d.id),
		).toContain(groupDataset.id);
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/datasets", { token: s2Token }))) as {
					datasets: Array<{ id: string }>;
				}
			).datasets.map((d) => d.id),
		).not.toContain(groupDataset.id);

		// 公开
		const pubRes = await formRequest(app, "/api/datasets", t1, uploadForm("public-data", { visibility: "public" }));
		expect(pubRes.status).toBe(201);
		const pub = (await readJson(pubRes)).dataset as { id: string };
		expect(
			(
				(await readJson(await jsonRequest(app, "/api/datasets", { token: t2 }))) as {
					datasets: Array<{ id: string }>;
				}
			).datasets.map((d) => d.id),
		).toContain(pub.id);

		// 下载:公开可见可下载;私有对他人 404
		const download = await jsonRequest(app, `/api/datasets/${pub.id}/download`, { token: t2 });
		expect(download.status).toBe(200);
		expect(await download.text()).toBe("public-data");
		const privDownload = await jsonRequest(app, `/api/datasets/${priv.id}/download`, { token: t2 });
		expect(privDownload.status).toBe(404);
	});

	it("组权限:非组成员上传 403,归档组 400,归档后成员仍可见(只读)", async () => {
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

		// 非组成员上传组数据集 403
		const denied = await formRequest(
			app,
			"/api/datasets",
			t2,
			uploadForm("x", { visibility: "group", groupId: group.id }),
		);
		expect(denied.status).toBe(403);

		// 成员可共享
		const shared = await formRequest(
			app,
			"/api/datasets",
			s1Token,
			uploadForm("s1-data", { visibility: "group", groupId: group.id }),
		);
		expect(shared.status).toBe(201);
		const sharedId = ((await readJson(shared)).dataset as { id: string }).id;

		// 归档
		await jsonRequest(app, `/api/groups/${group.id}`, { method: "DELETE", token: t1 });
		const afterArchive = await formRequest(
			app,
			"/api/datasets",
			t1,
			uploadForm("y", { visibility: "group", groupId: group.id }),
		);
		expect(afterArchive.status).toBe(400);

		// 归档后成员仍可见;非 owner(t1) 改被拒 → 只读
		const s1List = (await readJson(await jsonRequest(app, "/api/datasets", { token: s1Token }))) as {
			datasets: Array<{ id: string }>;
		};
		expect(s1List.datasets.map((d) => d.id)).toContain(sharedId);
		const patchDenied = await jsonRequest(app, `/api/datasets/${sharedId}`, {
			method: "PATCH",
			token: t1,
			body: { name: "x" },
		});
		expect(patchDenied.status).toBe(403);
		// owner(s1) 仍可管理
		const ownerPatch = await jsonRequest(app, `/api/datasets/${sharedId}`, {
			method: "PATCH",
			token: s1Token,
			body: { name: "s1-data-v2" },
		});
		expect(ownerPatch.status).toBe(200);
	});

	it("改名/改可见性;非 owner 403;admin 治理", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const s1 = await createUser(app, adminToken, "student1", "student");
		const t1 = await login(app, "teacher1", "pass123");
		const s1Token = await login(app, "student1", "pass123");

		const group = await createGroup(app, t1, "课程组");
		await jsonRequest(app, `/api/groups/${group.id}/members`, { method: "POST", token: t1, body: { userId: s1.id } });

		const created = await formRequest(app, "/api/datasets", t1, uploadForm("data", { name: "数据集A" }));
		const datasetId = ((await readJson(created)).dataset as { id: string }).id;

		// 改名
		const renamed = await jsonRequest(app, `/api/datasets/${datasetId}`, {
			method: "PATCH",
			token: t1,
			body: { name: "数据集A2" },
		});
		expect(renamed.status).toBe(200);
		expect(((await readJson(renamed)).dataset as { name: string }).name).toBe("数据集A2");

		// 改可见性 group
		const toGroup = await jsonRequest(app, `/api/datasets/${datasetId}`, {
			method: "PATCH",
			token: t1,
			body: { visibility: "group", groupId: group.id },
		});
		expect(toGroup.status).toBe(200);
		expect(((await readJson(toGroup)).dataset as { groupName: string | undefined }).groupName).toBe("课程组");

		// 非 owner(组内成员)不能改
		const memberPatch = await jsonRequest(app, `/api/datasets/${datasetId}`, {
			method: "PATCH",
			token: s1Token,
			body: { name: "x" },
		});
		expect(memberPatch.status).toBe(403);

		// admin 可治理
		const adminPatch = await jsonRequest(app, `/api/datasets/${datasetId}`, {
			method: "PATCH",
			token: adminToken,
			body: { name: "管理员改名" },
		});
		expect(adminPatch.status).toBe(200);
		expect(((await readJson(adminPatch)).dataset as { name: string }).name).toBe("管理员改名");
	});

	it("配额:单文件超限 413,每用户累计超限 413", async () => {
		const { app } = await createTestContext({ datasetMaxFileBytes: 10, datasetMaxUserBytes: 15 });
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t1 = await login(app, "teacher1", "pass123");

		// 单文件超限(20 字节 > 10)
		const tooBig = await formRequest(app, "/api/datasets", t1, uploadForm("a".repeat(20)));
		expect(tooBig.status).toBe(413);
		expect(((await readJson(tooBig)).error as { code: string }).code).toBe("quota_exceeded");

		// 累计超限:8 + 8 > 15
		const first = await formRequest(app, "/api/datasets", t1, uploadForm("b".repeat(8)));
		expect(first.status).toBe(201);
		const second = await formRequest(app, "/api/datasets", t1, uploadForm("c".repeat(8)));
		expect(second.status).toBe(413);
	});

	it("缺 file 字段 400", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t1 = await login(app, "teacher1", "pass123");

		const form = new FormData();
		form.append("visibility", "private");
		const res = await formRequest(app, "/api/datasets", t1, form);
		expect(res.status).toBe(400);
	});
});
