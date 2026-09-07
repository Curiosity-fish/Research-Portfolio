import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { setupFauxModel } from "../src/core/faux.ts";
import type { ExecutorConfig } from "../src/server/executor/config.ts";
import { DockerExecutor } from "../src/server/executor/docker.ts";
import type { CommandRunner, CommandRunOptions } from "../src/server/executor/types.ts";
import { createApp } from "../src/server/http-server.ts";
import { ShellSessionService } from "../src/server/session-service.ts";

class FakeRunner implements CommandRunner {
	calls: Array<{ command: string; args: string[]; options?: CommandRunOptions }> = [];
	queue: string[] = [];

	run(command: string, args: string[], options?: CommandRunOptions): Promise<{ stdout: string; stderr: string }> {
		this.calls.push({ command, args, options });
		return Promise.resolve({ stdout: this.queue.shift() ?? "", stderr: "" });
	}
}

function makeConfig(overrides: Partial<ExecutorConfig> = {}): ExecutorConfig {
	return {
		image: "tfa-executor",
		skillsHostDir: "/srv/skills",
		memoryBytes: 2 * 1024 ** 3,
		cpuCount: 2,
		timeoutMs: 60_000,
		maxConcurrentPerUser: 2,
		localBinary: "tfa",
		...overrides,
	};
}

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

async function createTestContext(options: { executor?: DockerExecutor } = {}): Promise<{
	app: ReturnType<typeof createApp>;
}> {
	const cwd = mkdtempSync(join(tmpdir(), "tfa-admin-sess-"));
	cleanupDirs.push(cwd);
	const faux = setupFauxModel();
	const service = new ShellSessionService({
		cwd,
		agentDir: join(cwd, "agent"),
		fauxSetup: faux,
		...(options.executor ? { executor: options.executor } : {}),
	});
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

describe("会话监控 (T11)", () => {
	it("未认证 401,非管理员 403", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await jsonRequest(app, "/api/admin/users", {
			method: "POST",
			token: adminToken,
			body: { username: "teacher1", password: "pass123", role: "teacher" },
		});
		const t1 = await login(app, "teacher1", "pass123");

		expect((await jsonRequest(app, "/api/admin/sessions")).status).toBe(401);
		expect((await jsonRequest(app, "/api/admin/sessions", { token: t1 })).status).toBe(403);
	});

	it("创建会话后 admin 列表可见(status running)", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");

		const created = await jsonRequest(app, "/api/sessions", { method: "POST", body: { name: "监控会话" } });
		expect(created.status).toBe(201);
		const sessionId = ((await readJson(created)).session as { id: string }).id;

		const list = (await readJson(await jsonRequest(app, "/api/admin/sessions", { token: adminToken }))) as {
			sessions: Array<{ id: string; status: string; cwd: string | null }>;
		};
		const view = list.sessions.find((s) => s.id === sessionId);
		expect(view).toBeDefined();
		expect(view?.status).toBe("running");
		expect(view?.cwd).toBeTruthy();
	});

	it("terminate 结束 live 会话并置 ended", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");

		const created = await jsonRequest(app, "/api/sessions", { method: "POST", body: {} });
		const sessionId = ((await readJson(created)).session as { id: string }).id;

		const terminate = await jsonRequest(app, `/api/admin/sessions/${sessionId}/terminate`, {
			method: "POST",
			token: adminToken,
		});
		expect(terminate.status).toBe(200);

		const list = (await readJson(await jsonRequest(app, "/api/admin/sessions", { token: adminToken }))) as {
			sessions: Array<{ id: string; status: string }>;
		};
		expect(list.sessions.find((s) => s.id === sessionId)?.status).toBe("ended");
	});

	it("terminate 不存在的会话返回 404", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		const res = await jsonRequest(app, "/api/admin/sessions/no-such/terminate", {
			method: "POST",
			token: adminToken,
		});
		expect(res.status).toBe(404);
	});

	it("executor 集成:容器会话状态展示,terminate 调 docker stop/rm 并置 ended", async () => {
		const runner = new FakeRunner();
		runner.queue.push("cid-123\n");
		const executor = new DockerExecutor(makeConfig(), runner);
		const { app } = await createTestContext({ executor });
		const adminToken = await login(app, "admin", "admin123");

		// 登记平台会话行 + 执行器跟踪(模拟容器已启动)
		const service = services[services.length - 1];
		service.getRepos().sessionCreate({
			id: "s-c1",
			userId: null,
			cwd: "/work",
			status: "running",
			startedAt: Date.now(),
		});
		await executor.start({
			sessionId: "s-c1",
			userId: "u-1",
			projectHostPath: "/home/p1",
			cwd: "/work",
			skills: [],
			datasets: [],
			env: {},
			limits: { memoryBytes: 1, cpuCount: 1, timeoutMs: 60_000 },
		});
		// inspect 响应(start 已消费 "cid-123",这里补 "running")
		runner.queue.push("running\n");

		// admin 列表:容器状态 running,容器 id
		const list = (await readJson(await jsonRequest(app, "/api/admin/sessions", { token: adminToken }))) as {
			sessions: Array<{ id: string; containerId: string | null; containerState: string | null }>;
		};
		const view = list.sessions.find((s) => s.id === "s-c1");
		expect(view?.containerId).toBe("cid-123");
		expect(view?.containerState).toBe("running");

		// terminate:executor stop/rm + 平台行 ended
		const terminate = await jsonRequest(app, "/api/admin/sessions/s-c1/terminate", {
			method: "POST",
			token: adminToken,
		});
		expect(terminate.status).toBe(200);
		expect(runner.calls.some((call) => call.args[0] === "stop" && call.args[3] === "tfa-s-c1")).toBe(true);
		expect(runner.calls.some((call) => call.args[0] === "rm" && call.args[2] === "tfa-s-c1")).toBe(true);

		const after = (await readJson(await jsonRequest(app, "/api/admin/sessions", { token: adminToken }))) as {
			sessions: Array<{ id: string; status: string }>;
		};
		expect(after.sessions.find((s) => s.id === "s-c1")?.status).toBe("ended");
	});
});
