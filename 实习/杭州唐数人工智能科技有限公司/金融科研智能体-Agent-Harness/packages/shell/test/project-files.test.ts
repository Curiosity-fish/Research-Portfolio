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
	const cwd = mkdtempSync(join(tmpdir(), "tfa-files-test-"));
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

function uploadForm(content: string, fileName: string, dir?: string): FormData {
	const form = new FormData();
	form.append("file", new File([content], fileName));
	if (dir) form.append("path", dir);
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

async function createProject(app: ReturnType<typeof createApp>, token: string, name: string): Promise<{ id: string }> {
	const res = await jsonRequest(app, "/api/projects", { method: "POST", token, body: { name } });
	const body = await readJson(res);
	return body.project as { id: string };
}

describe("项目文件工作空间 API (T12)", () => {
	it("未认证 401;越权项目 404", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "teacher2", "teacher");
		const t1 = await login(app, "teacher1", "pass123");
		const t2 = await login(app, "teacher2", "pass123");

		const project = await createProject(app, t1, "项目A");
		expect((await jsonRequest(app, `/api/projects/${project.id}/files`)).status).toBe(401);
		expect((await jsonRequest(app, `/api/projects/${project.id}/files`, { token: t2 })).status).toBe(404);
		expect((await jsonRequest(app, `/api/projects/${project.id}/files`, { token: t1 })).status).toBe(200);
	});

	it("新项目列表为空;读写文本后列表与内容一致", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t1 = await login(app, "teacher1", "pass123");
		const project = await createProject(app, t1, "项目A");

		const empty = (await readJson(await jsonRequest(app, `/api/projects/${project.id}/files`, { token: t1 }))) as {
			files: unknown[];
		};
		expect(empty.files).toEqual([]);

		const write = await jsonRequest(app, `/api/projects/${project.id}/files/content`, {
			method: "PUT",
			token: t1,
			body: { path: "notes/研究.md", content: "# 研究\nhello" },
		});
		expect(write.status).toBe(200);
		expect(((await readJson(write)).file as { size: number }).size).toBe(Buffer.byteLength("# 研究\nhello", "utf8"));

		const read = await jsonRequest(app, `/api/projects/${project.id}/files/content?path=notes/研究.md`, {
			token: t1,
		});
		expect(read.status).toBe(200);
		const readBody = (await readJson(read)).file as { content: string; mimeType: string };
		expect(readBody.content).toBe("# 研究\nhello");
		expect(readBody.mimeType).toBe("text/markdown");

		const list = (await readJson(
			await jsonRequest(app, `/api/projects/${project.id}/files?path=notes`, { token: t1 }),
		)) as {
			files: Array<{ name: string; type: string }>;
		};
		expect(list.files.find((f) => f.name === "研究.md")?.type).toBe("file");

		// 不存在路径 404
		expect(
			(await jsonRequest(app, `/api/projects/${project.id}/files/content?path=nope.txt`, { token: t1 })).status,
		).toBe(404);
	});

	it("路径穿越/绝对路径被拒", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t1 = await login(app, "teacher1", "pass123");
		const project = await createProject(app, t1, "项目A");

		const escapeWrite = await jsonRequest(app, `/api/projects/${project.id}/files/content`, {
			method: "PUT",
			token: t1,
			body: { path: "../escape.txt", content: "x" },
		});
		expect(escapeWrite.status).toBe(400);

		const readEscape = await jsonRequest(app, `/api/projects/${project.id}/files/content?path=../secret`, {
			token: t1,
		});
		expect(readEscape.status).toBe(400);

		const abs = await jsonRequest(app, `/api/projects/${project.id}/files/content`, {
			method: "PUT",
			token: t1,
			body: { path: "/etc/passwd", content: "x" },
		});
		expect(abs.status).toBe(400);
	});

	it("上传文件可读回并列入列表", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t1 = await login(app, "teacher1", "pass123");
		const project = await createProject(app, t1, "项目A");

		const upload = await formRequest(
			app,
			`/api/projects/${project.id}/files`,
			t1,
			uploadForm("1,2,3\n4,5,6", "data.csv", "data"),
		);
		expect(upload.status).toBe(201);
		expect(((await readJson(upload)).file as { mimeType: string }).mimeType).toBe("text/csv");

		const read = await jsonRequest(app, `/api/projects/${project.id}/files/content?path=data/data.csv`, {
			token: t1,
		});
		expect(read.status).toBe(200);
		expect(((await readJson(read)).file as { content: string }).content).toBe("1,2,3\n4,5,6");

		const list = (await readJson(
			await jsonRequest(app, `/api/projects/${project.id}/files?path=data`, { token: t1 }),
		)) as {
			files: Array<{ name: string }>;
		};
		expect(list.files.map((f) => f.name)).toContain("data.csv");
	});

	it("下载返回内容与 MIME;删除后列表消失", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		const t1 = await login(app, "teacher1", "pass123");
		const project = await createProject(app, t1, "项目A");

		const upload = await formRequest(
			app,
			`/api/projects/${project.id}/files`,
			t1,
			uploadForm("binary-ish", "img.png"),
		);
		expect(upload.status).toBe(201);

		const download = await jsonRequest(app, `/api/projects/${project.id}/files/download?path=img.png`, { token: t1 });
		expect(download.status).toBe(200);
		expect(download.headers.get("Content-Type")).toContain("image/png");
		expect(await download.text()).toBe("binary-ish");

		const del = await jsonRequest(app, `/api/projects/${project.id}/files?path=img.png`, {
			method: "DELETE",
			token: t1,
		});
		expect(del.status).toBe(200);
		const list = (await readJson(await jsonRequest(app, `/api/projects/${project.id}/files`, { token: t1 }))) as {
			files: unknown[];
		};
		expect(list.files).toEqual([]);
	});
});
