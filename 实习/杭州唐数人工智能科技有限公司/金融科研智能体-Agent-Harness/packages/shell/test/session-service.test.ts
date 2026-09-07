import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { setupFauxModel } from "../src/core/faux.ts";
import { type SessionSnapshot, ShellSessionService, type TfaSessionRuntime } from "../src/server/session-service.ts";

let tempDir = "";
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

function createTempDir(): string {
	tempDir = mkdtempSync(join(tmpdir(), "tfa-shell-test-"));
	cleanupDirs.push(tempDir);
	return tempDir;
}

async function createService(options?: { tokensPerSecond?: number }): Promise<{
	service: ShellSessionService;
	faux: ReturnType<typeof setupFauxModel>;
}> {
	const cwd = createTempDir();
	const faux = setupFauxModel(options);
	const service = new ShellSessionService({ cwd, agentDir: join(cwd, "agent"), fauxSetup: faux });
	services.push(service);
	return { service, faux };
}

function waitForIdle(runtime: TfaSessionRuntime, timeoutMs = 10_000): Promise<void> {
	return new Promise((resolve, reject) => {
		const started = Date.now();
		const timer = setInterval(() => {
			if (runtime.getPhase() === "idle") {
				clearInterval(timer);
				resolve();
			} else if (Date.now() - started > timeoutMs) {
				clearInterval(timer);
				reject(new Error("会话未在超时内回到 idle"));
			}
		}, 25);
	});
}

describe("ShellSessionService", () => {
	it("create → prompt → snapshot 包含回复", async () => {
		const { service, faux } = await createService();
		const runtime = await service.createSession({ id: "test-1" });
		const events: string[] = [];
		runtime.subscribe((event) => {
			events.push(event.type);
		});

		expect(runtime.getPhase()).toBe("idle");
		await runtime.prompt({ text: "你好" });
		await waitForIdle(runtime);

		const snapshot = runtime.snapshot() as SessionSnapshot;
		expect(snapshot.id).toBe("test-1");
		expect(snapshot.transcript.length).toBeGreaterThanOrEqual(2);
		const roles = snapshot.transcript.map((item) => item.role);
		expect(roles).toContain("user");
		expect(roles).toContain("assistant");
		expect(events).toContain("progress");
		await service.close();
		faux.unregister();
	});

	it("prompt 期间第二次 prompt 返回 busy", async () => {
		// 慢速响应确保第一次 prompt 未完成时发起第二次
		const { service, faux } = await createService({ tokensPerSecond: 20 });
		const runtime = await service.createSession({ id: "test-2" });
		const first = runtime.prompt({ text: "第一次" });
		await new Promise((resolve) => setTimeout(resolve, 50));
		await expect(runtime.prompt({ text: "第二次" })).rejects.toMatchObject({ code: "busy" });
		await first;
		await waitForIdle(runtime);
		await service.close();
		faux.unregister();
	});

	it("openSession 恢复持久化会话", async () => {
		const { service, faux } = await createService();
		const runtime = await service.createSession({ id: "test-3", name: "恢复测试" });
		await runtime.prompt({ text: "你好" });
		await waitForIdle(runtime);

		const reopened = await service.openSession("test-3");
		const snapshot = reopened.snapshot() as SessionSnapshot;
		expect(snapshot.id).toBe("test-3");
		expect(snapshot.transcript.length).toBeGreaterThanOrEqual(2);
		await service.close();
		faux.unregister();
	});

	it("deleteSession 后 listSessions 不含该会话", async () => {
		const { service, faux } = await createService();
		const runtime = await service.createSession({ id: "test-4" });
		await runtime.prompt({ text: "你好" });
		await waitForIdle(runtime);

		const before = await service.listSessions();
		expect(before.some((session) => session.id === "test-4")).toBe(true);

		await service.deleteSession("test-4");
		const after = await service.listSessions();
		expect(after.some((session) => session.id === "test-4")).toBe(false);
		await service.close();
		faux.unregister();
	});

	it("listModels 返回 faux 模型且已认证", async () => {
		const { service, faux } = await createService();
		const models = await service.listModels();
		expect(models.length).toBeGreaterThan(0);
		const fauxModel = models.find((model) => model.provider === "faux");
		expect(fauxModel).toBeDefined();
		expect(fauxModel?.authenticated).toBe(true);
		await service.close();
		faux.unregister();
	});

	it("progress 事件流包含 item_started/assistant_delta/item_finished", async () => {
		const { service, faux } = await createService();
		const runtime = await service.createSession({ id: "test-5" });
		const progressTypes: string[] = [];
		runtime.subscribe((event) => {
			if (event.type === "progress") progressTypes.push(event.progress.type);
		});
		await runtime.prompt({ text: "你好" });
		await waitForIdle(runtime);
		expect(progressTypes).toContain("item_started");
		expect(progressTypes).toContain("assistant_delta");
		expect(progressTypes).toContain("item_finished");
		await service.close();
		faux.unregister();
	});
});
