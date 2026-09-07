import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import type { ExecutorConfig } from "../src/server/executor/config.ts";
import { DockerExecutor } from "../src/server/executor/docker.ts";
import type { CommandRunner, CommandRunOptions } from "../src/server/executor/types.ts";
import { createApp } from "../src/server/http-server.ts";
import { ShellSessionService } from "../src/server/session-service.ts";

class FakeRunner implements CommandRunner {
	calls: Array<{ command: string; args: string[]; options?: CommandRunOptions }> = [];

	run(command: string, args: string[], options?: CommandRunOptions): Promise<{ stdout: string; stderr: string }> {
		this.calls.push({ command, args, options });
		return Promise.resolve({ stdout: "cid\n", stderr: "" });
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

async function createTestContext(): Promise<{ app: ReturnType<typeof createApp>; service: ShellSessionService }> {
	const cwd = mkdtempSync(join(tmpdir(), "tfa-perm-test-"));
	cleanupDirs.push(cwd);
	const service = new ShellSessionService({ cwd, agentDir: join(cwd, "agent") });
	services.push(service);
	await service.createUser({ username: "admin", password: "admin123", role: "admin" });
	const app = createApp(service, { staticDir: join(cwd, "no-such-dir") });
	return { app, service };
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

describe("会话权限模式 (T14)", () => {
	it("创建默认模式:项目会话 request_approval,通用会话 full_access;归属记录", async () => {
		const { app, service } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		const teacher = await createUser(app, adminToken, "teacher1", "teacher");
		const t1 = await login(app, "teacher1", "pass123");

		// 项目会话
		const projectRes = await jsonRequest(app, "/api/projects", {
			method: "POST",
			token: t1,
			body: { name: "项目A" },
		});
		const projectId = ((await readJson(projectRes)).project as { id: string }).id;
		const ps = await jsonRequest(app, "/api/sessions", { method: "POST", token: t1, body: { projectId } });
		expect(ps.status).toBe(201);
		const psId = ((await readJson(ps)).session as { id: string }).id;
		const psRecord = service.getRepos().sessionGet(psId);
		expect(psRecord?.permissionMode).toBe("request_approval");
		expect(psRecord?.userId).toBe(teacher.id);

		// 通用会话
		const gs = await jsonRequest(app, "/api/sessions", { method: "POST", token: t1, body: {} });
		const gsId = ((await readJson(gs)).session as { id: string }).id;
		expect(service.getRepos().sessionGet(gsId)?.permissionMode).toBe("full_access");
	});

	it("切换权限模式:owner 可切;未登录 401;他人 403;非法 400", async () => {
		const { app } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "teacher2", "teacher");
		const t1 = await login(app, "teacher1", "pass123");
		const t2 = await login(app, "teacher2", "pass123");

		const created = await jsonRequest(app, "/api/sessions", { method: "POST", token: t1, body: {} });
		const sessionId = ((await readJson(created)).session as { id: string }).id;

		// 未登录 401
		expect(
			(
				await jsonRequest(app, `/api/sessions/${sessionId}/permission-mode`, {
					method: "POST",
					body: { mode: "request_approval" },
				})
			).status,
		).toBe(401);

		// 非法 400
		const bad = await jsonRequest(app, `/api/sessions/${sessionId}/permission-mode`, {
			method: "POST",
			token: t1,
			body: { mode: "super" },
		});
		expect(bad.status).toBe(400);

		// 他人 403
		const denied = await jsonRequest(app, `/api/sessions/${sessionId}/permission-mode`, {
			method: "POST",
			token: t2,
			body: { mode: "full_access" },
		});
		expect(denied.status).toBe(403);

		// owner 切换成功
		const ok = await jsonRequest(app, `/api/sessions/${sessionId}/permission-mode`, {
			method: "POST",
			token: t1,
			body: { mode: "request_approval" },
		});
		expect(ok.status).toBe(200);
		expect((await readJson(ok)).mode as string).toBe("request_approval");
		const read = await jsonRequest(app, `/api/sessions/${sessionId}/permission-mode`, { token: t1 });
		expect((await readJson(read)).mode as string).toBe("request_approval");
	});

	it("批准状态机:审批/拒绝/幂等/超时默认拒绝", async () => {
		const { app, service } = await createTestContext();

		const created = await jsonRequest(app, "/api/sessions", { method: "POST", body: {} });
		const sessionId = ((await readJson(created)).session as { id: string }).id;

		const req = service.approvals.createRequest(sessionId, {
			tool: "write",
			target: "notes/a.md",
			risk: "修改项目文件",
		});
		expect(req.status).toBe("pending");
		expect(service.approvals.listPending(sessionId).map((r) => r.id)).toContain(req.id);

		// 批准 + 幂等
		expect(service.approvals.approve(req.id).status).toBe("approved");
		expect(service.approvals.approve(req.id).status).toBe("approved");

		// 已批准再拒绝 → 400
		expect(() => service.approvals.reject(req.id)).toThrow();

		// 拒绝路径
		const req2 = service.approvals.createRequest(sessionId, { tool: "bash", target: "rm -rf x", risk: "删除" });
		expect(service.approvals.reject(req2.id).status).toBe("rejected");
		expect(service.approvals.reject(req2.id).status).toBe("rejected");

		// 超时 → expired
		const req3 = service.approvals.createRequest(sessionId, { tool: "write", target: "x", risk: "y" });
		service.approvals.expireOverdue(req3.expiresAt + 1);
		expect(service.approvals.get(req3.id)?.status).toBe("expired");

		// 未知 id → 404
		expect(() => service.approvals.approve("no-such")).toThrow();
	});

	it("批准 API:列表/批准/拒绝,非 owner 403", async () => {
		const { app, service } = await createTestContext();
		const adminToken = await login(app, "admin", "admin123");
		await createUser(app, adminToken, "teacher1", "teacher");
		await createUser(app, adminToken, "teacher2", "teacher");
		const t1 = await login(app, "teacher1", "pass123");
		const t2 = await login(app, "teacher2", "pass123");

		const created = await jsonRequest(app, "/api/sessions", { method: "POST", token: t1, body: {} });
		const sessionId = ((await readJson(created)).session as { id: string }).id;
		const req = service.approvals.createRequest(sessionId, { tool: "write", target: "a.txt", risk: "x" });

		// 非 owner 403
		expect((await jsonRequest(app, `/api/sessions/${sessionId}/approvals`, { token: t2 })).status).toBe(403);

		// owner 列表
		const list = (await readJson(await jsonRequest(app, `/api/sessions/${sessionId}/approvals`, { token: t1 }))) as {
			approvals: Array<{ id: string; status: string }>;
		};
		expect(list.approvals.map((a) => a.id)).toContain(req.id);

		// 批准
		const approve = await jsonRequest(app, `/api/sessions/${sessionId}/approvals/${req.id}/approve`, {
			method: "POST",
			token: t1,
		});
		expect(approve.status).toBe(200);
		expect(((await readJson(approve)).approval as { status: string }).status).toBe("approved");
	});

	it("executor:两种权限模式下的隔离参数一致(工作区挂载 + --no-skills --skill)", async () => {
		const runner = new FakeRunner();
		const executor = new DockerExecutor(makeConfig(), runner);
		const base = {
			sessionId: "s-x",
			userId: "u-1",
			projectHostPath: "/home/p1",
			cwd: "/work",
			skills: ["/srv/skills/sk1"],
			datasets: [{ hostPath: "/data/d1", containerPath: "/data/d1", readonly: true }],
			env: {},
			limits: { memoryBytes: 1, cpuCount: 1, timeoutMs: 60_000 },
		};

		await executor.start({ ...base, permissionMode: "request_approval" });
		await executor.start({ ...base, permissionMode: "full_access", sessionId: "s-y" });

		expect(runner.calls).toHaveLength(2);
		for (const call of runner.calls) {
			expect(call.args).toEqual(
				expect.arrayContaining([
					"-v",
					"/home/p1:/work",
					"-v",
					"/srv/skills:/skills:ro",
					"-v",
					"/data/d1:/data/d1:ro",
					"tfa-executor",
					"tfa",
					"--no-skills",
					"--skill",
					"/skills/sk1",
				]),
			);
		}
	});
});
