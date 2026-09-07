import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { setupFauxModel } from "../src/core/faux.ts";
import { createApp } from "../src/server/http-server.ts";
import { ShellSessionService } from "../src/server/session-service.ts";

const cleanupDirs: string[] = [];
const services: ShellSessionService[] = [];

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

async function createTestContext(): Promise<{
	app: ReturnType<typeof createApp>;
	faux: ReturnType<typeof setupFauxModel>;
}> {
	const cwd = mkdtempSync(join(tmpdir(), "tfa-http-test-"));
	cleanupDirs.push(cwd);
	const faux = setupFauxModel();
	const service = new ShellSessionService({ cwd, agentDir: join(cwd, "agent"), fauxSetup: faux });
	services.push(service);
	await service.createUser({ username: "admin", password: "admin123", role: "admin" });
	const app = createApp(service, { staticDir: join(cwd, "no-such-dir") });
	return { app, faux };
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

async function readJson(response: Response): Promise<unknown> {
	return response.json();
}

async function login(app: ReturnType<typeof createApp>, username: string, password: string): Promise<string> {
	const res = await jsonRequest(app, "/api/auth/login", { method: "POST", body: { username, password } });
	const body = (await readJson(res)) as { token: string };
	return body.token;
}

/** 读取 SSE 流,收集 event 帧,直到连接关闭。 */
async function readSse(response: Response): Promise<Array<{ event: string; data: string }>> {
	const frames: Array<{ event: string; data: string }> = [];
	if (!response.body) return frames;
	const reader = response.body.getReader();
	const decoder = new TextDecoder();
	let buffer = "";
	for (;;) {
		const { done, value } = await reader.read();
		if (done) break;
		buffer += decoder.decode(value, { stream: true });
		// 按空行切分 SSE 帧
		let separator = buffer.indexOf("\n\n");
		while (separator !== -1) {
			const frame = buffer.slice(0, separator);
			buffer = buffer.slice(separator + 2);
			const eventLine = frame.split("\n").find((line) => line.startsWith("event:"));
			const dataLine = frame.split("\n").find((line) => line.startsWith("data:"));
			if (eventLine && dataLine) {
				frames.push({ event: eventLine.slice(7), data: dataLine.slice(6) });
			}
			separator = buffer.indexOf("\n\n");
		}
	}
	return frames;
}

describe("HTTP 服务", () => {
	it("health 检查", async () => {
		const { app, faux } = await createTestContext();
		const response = await jsonRequest(app, "/api/health");
		expect(response.status).toBe(200);
		expect(await readJson(response)).toEqual({ ok: true });
		await faux.unregister();
	});

	it("创建会话并 prompt,SSE 收到事件流", async () => {
		const { app, faux } = await createTestContext();
		const created = await jsonRequest(app, "/api/sessions", { method: "POST", body: { name: "测试会话" } });
		expect(created.status).toBe(201);
		const createdBody = (await readJson(created)) as { session: { id: string } };
		const sessionId = createdBody.session.id;
		expect(sessionId).toBeTruthy();

		// 并发打开 SSE
		const ssePromise = jsonRequest(app, `/api/sessions/${sessionId}/events`);
		const promptPromise = jsonRequest(app, `/api/sessions/${sessionId}/prompt`, {
			method: "POST",
			body: { text: "你好" },
		});
		const [sseResponse, promptResponse] = await Promise.all([ssePromise, promptPromise]);
		expect(promptResponse.status).toBe(200);

		// 等流式完成后读 SSE 帧(响应体流在 prompt 完成后仍在 keep-alive;读到 progress 帧即验证)
		const frames: Array<{ event: string; data: string }> = [];
		if (sseResponse.body) {
			const reader = sseResponse.body.getReader();
			const decoder = new TextDecoder();
			let buffer = "";
			let gotProgress = false;
			for (let attempts = 0; attempts < 400 && !gotProgress; attempts++) {
				const { done, value } = await reader.read();
				if (done) break;
				buffer += decoder.decode(value, { stream: true });
				const separator = buffer.indexOf("\n\n");
				if (separator !== -1) {
					const frame = buffer.slice(0, separator);
					buffer = buffer.slice(separator + 2);
					const eventLine = frame.split("\n").find((line) => line.startsWith("event:"));
					const dataLine = frame.split("\n").find((line) => line.startsWith("data:"));
					if (eventLine && dataLine) {
						frames.push({ event: eventLine.slice(7), data: dataLine.slice(6) });
						if (eventLine.slice(7) === "progress") gotProgress = true;
					}
				}
			}
			reader.cancel();
		}
		expect(frames.some((frame) => frame.event === "progress")).toBe(true);
		expect(frames.some((frame) => frame.event === "snapshot")).toBe(true);
		await faux.unregister();
	});

	it("模型列表包含 faux 模型", async () => {
		const { app, faux } = await createTestContext();
		const response = await jsonRequest(app, "/api/models");
		expect(response.status).toBe(200);
		const body = (await readJson(response)) as { models: Array<{ provider: string; id: string }> };
		expect(body.models.some((model) => model.provider === "faux")).toBe(true);
		await faux.unregister();
	});

	it("会话列表与删除", async () => {
		const { app, faux } = await createTestContext();
		const created = await jsonRequest(app, "/api/sessions", { method: "POST", body: {} });
		const sessionId = ((await readJson(created)) as { session: { id: string } }).session.id;

		const listResponse = await jsonRequest(app, "/api/sessions");
		const list = (await readJson(listResponse)) as { sessions: Array<{ id: string }> };
		expect(list.sessions.some((session) => session.id === sessionId)).toBe(true);

		const deleted = await jsonRequest(app, `/api/sessions/${sessionId}`, { method: "DELETE" });
		expect(deleted.status).toBe(200);

		const after = await jsonRequest(app, `/api/sessions`);
		const listAfter = (await readJson(after)) as { sessions: Array<{ id: string }> };
		expect(listAfter.sessions.some((session) => session.id === sessionId)).toBe(false);
		await faux.unregister();
	});

	it("未知会话返回 404", async () => {
		const { app, faux } = await createTestContext();
		const response = await jsonRequest(app, "/api/sessions/does-not-exist");
		expect(response.status).toBe(404);
		await faux.unregister();
	});

	it("prompt 空文本返回 400", async () => {
		const { app, faux } = await createTestContext();
		const created = await jsonRequest(app, "/api/sessions", { method: "POST", body: {} });
		const sessionId = ((await readJson(created)) as { session: { id: string } }).session.id;
		const response = await jsonRequest(app, `/api/sessions/${sessionId}/prompt`, {
			method: "POST",
			body: { text: "" },
		});
		expect(response.status).toBe(400);
		await faux.unregister();
	});
});

// readSse 供后续扩展使用(保留引用避免未使用告警)
void readSse;

describe("项目 API", () => {
	it("创建/列出/重命名/删除项目,创建会话可归属项目", async () => {
		const { app, faux } = await createTestContext();
		const token = await login(app, "admin", "admin123");

		const created = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token,
			body: { name: "新能源行业", folder: "D:\\research\\新能源" },
		});
		expect(created.status).toBe(201);
		const project = (
			(await readJson(created)) as { project: { id: string; name: string; folder?: string; sessionIds: string[] } }
		).project;
		expect(project.name).toBe("新能源行业");
		expect(project.folder).toBe("D:\\research\\新能源");
		expect(project.sessionIds).toEqual([]);

		// 名称必填
		const bad = await jsonRequest(app, "/api/projects", { method: "POST", token, body: { name: "  " } });
		expect(bad.status).toBe(400);

		const listed = await jsonRequest(app, "/api/projects", { token });
		expect(listed.status).toBe(200);
		expect(((await readJson(listed)) as { projects: Array<{ id: string }> }).projects).toHaveLength(1);

		// 重命名
		const renamed = await jsonRequest(app, `/api/projects/${project.id}`, {
			method: "PATCH",
			token,
			body: { name: "新能源与储能" },
		});
		expect(renamed.status).toBe(200);
		expect(((await readJson(renamed)) as { project: { name: string } }).project.name).toBe("新能源与储能");

		// 归属创建会话
		const createdSession = await jsonRequest(app, "/api/sessions", {
			method: "POST",
			body: { name: "项目内会话", projectId: project.id },
		});
		expect(createdSession.status).toBe(201);
		const sessionId = ((await readJson(createdSession)) as { session: { id: string } }).session.id;
		const afterCreate = await jsonRequest(app, "/api/projects", { token });
		const projectAfter = ((await readJson(afterCreate)) as { projects: Array<{ id: string; sessionIds: string[] }> })
			.projects[0];
		expect(projectAfter.sessionIds).toContain(sessionId);

		// 移动会话到无项目
		const moved = await jsonRequest(app, `/api/sessions/${sessionId}/project`, {
			method: "POST",
			token,
			body: { projectId: null },
		});
		expect(moved.status).toBe(200);
		const afterMove = await jsonRequest(app, "/api/projects", { token });
		expect(
			((await readJson(afterMove)) as { projects: Array<{ sessionIds: string[] }> }).projects[0].sessionIds,
		).toEqual([]);

		// 删除项目
		const removed = await jsonRequest(app, `/api/projects/${project.id}`, { method: "DELETE", token });
		expect(removed.status).toBe(200);
		const afterDelete = await jsonRequest(app, "/api/projects", { token });
		expect(((await readJson(afterDelete)) as { projects: unknown[] }).projects).toHaveLength(0);

		await faux.unregister();
	});

	it("删除项目会话后,项目 sessionIds 被清理", async () => {
		const { app, faux } = await createTestContext();
		const token = await login(app, "admin", "admin123");
		const created = await jsonRequest(app, "/api/projects", { method: "POST", token, body: { name: "测试项目" } });
		const project = ((await readJson(created)) as { project: { id: string } }).project;
		const createdSession = await jsonRequest(app, "/api/sessions", {
			method: "POST",
			body: { projectId: project.id },
		});
		const sessionId = ((await readJson(createdSession)) as { session: { id: string } }).session.id;

		const del = await jsonRequest(app, `/api/sessions/${sessionId}`, { method: "DELETE" });
		expect(del.status).toBe(200);
		const listed = await jsonRequest(app, "/api/projects", { token });
		const projectAfter = ((await readJson(listed)) as { projects: Array<{ id: string; sessionIds: string[] }> })
			.projects[0];
		expect(projectAfter.sessionIds).toEqual([]);

		await faux.unregister();
	});
});
