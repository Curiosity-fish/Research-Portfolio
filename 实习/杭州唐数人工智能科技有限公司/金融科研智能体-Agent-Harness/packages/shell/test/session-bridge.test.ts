import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { RpcClient } from "@earendil-works/tfa-coding-agent";
import { afterEach, describe, expect, it } from "vitest";
import { buildSkillCliArgs, createRpcSessionClient, ProxyRuntime } from "../src/server/bridge.ts";
import { ShellSessionService } from "../src/server/session-service.ts";

const fakeCli = join(import.meta.dirname, "fixtures", "fake-rpc-agent.cjs");

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

const sleep = (ms: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, ms));

describe("会话桥接 (T8b)", () => {
	it("createRpcSessionClient 携带 --no-skills --skill / provider / model / env", async () => {
		const tmp = mkdtempSync(join(tmpdir(), "tfa-bridge-argv-"));
		cleanupDirs.push(tmp);
		const argvFile = join(tmp, "argv.json");

		const client = createRpcSessionClient({
			cliPath: fakeCli,
			cwd: tmp,
			env: { FAKE_RPC_ARGV_FILE: argvFile, OLLAMA_BASE_URL: "http://host:11434/v1" },
			provider: "ollama",
			model: "qwen3:0.6b",
			args: buildSkillCliArgs(["/libs/sk1", "/libs/sk2"]),
		});
		await client.start();

		const argv = JSON.parse(readFileSync(argvFile, "utf8")) as string[];
		expect(argv).toEqual(
			expect.arrayContaining([
				"--mode",
				"rpc",
				"--provider",
				"ollama",
				"--model",
				"qwen3:0.6b",
				"--no-skills",
				"--skill",
				"/libs/sk1",
				"--skill",
				"/libs/sk2",
			]),
		);
		await client.stop();
	});

	it("ProxyRuntime 转发 prompt/steer/abort/setModel/setThinking 并映射事件为 snapshot", async () => {
		const tmp = mkdtempSync(join(tmpdir(), "tfa-bridge-proxy-"));
		cleanupDirs.push(tmp);
		const client = new RpcClient({ cliPath: fakeCli, cwd: tmp, env: {} });
		await client.start();
		const runtime = new ProxyRuntime({ client, sessionId: "s-1", cwd: tmp });

		const events: string[] = [];
		runtime.subscribe((event) => events.push(event.type));

		await runtime.prompt({ text: "hello" });
		await sleep(50);
		expect(events).toContain("snapshot");
		expect(runtime.getPhase()).toBe("idle");

		await runtime.steer({ text: "more" });
		await runtime.setModel({ provider: "fake", id: "m2" });
		await runtime.setThinking("high");
		await runtime.abort();

		const snap = await runtime.snapshot();
		expect(snap.id).toBe("s-1");
		expect(snap.cwd).toBe(tmp);
		expect(snap.transcript).toEqual([]);

		await runtime.dispose();
	});

	it("ShellSessionService.createRpcSession 走桥接,close 销毁进程", async () => {
		const tmp = mkdtempSync(join(tmpdir(), "tfa-bridge-svc-"));
		cleanupDirs.push(tmp);
		const argvFile = join(tmp, "argv.json");
		const service = new ShellSessionService({ cwd: tmp, agentDir: join(tmp, "agent") });
		services.push(service);

		const runtime = await service.createRpcSession({
			sessionId: "s-1",
			cliPath: fakeCli,
			cwd: tmp,
			env: { FAKE_RPC_ARGV_FILE: argvFile },
			provider: "ollama",
			model: "qwen3:0.6b",
			skills: ["/libs/sk1"],
		});

		const events: string[] = [];
		runtime.subscribe((event) => events.push(event.type));
		await runtime.prompt({ text: "hi" });
		await sleep(50);
		expect(events).toContain("snapshot");

		const argv = JSON.parse(readFileSync(argvFile, "utf8")) as string[];
		expect(argv).toContain("--no-skills");
		expect(argv).toContain("--skill");
		expect(argv).toContain("/libs/sk1");

		await service.close();
	});
});

it("RPC 审批表面化:agent 的 approval_request 进入 ApprovalService,decideApproval 生效", async () => {
	const tmp = mkdtempSync(join(tmpdir(), "tfa-bridge-approval-"));
	cleanupDirs.push(tmp);
	const service = new ShellSessionService({ cwd: tmp, agentDir: join(tmp, "agent") });
	services.push(service);
	const runtime = await service.createRpcSession({ sessionId: "s-appr", cliPath: fakeCli, cwd: tmp, env: {} });

	await runtime.prompt({ text: "hi" });
	await sleep(100);

	const approvals = service.approvals.listPending("s-appr");
	expect(approvals.length).toBe(1);
	expect(approvals[0].tool).toBe("write");
	expect(approvals[0].target).toBe("a.txt");

	// decideApproval:会话无归属用户(null),getOwnedSession 放行
	const admin = await service.createUser({ username: "admin", password: "admin123", role: "admin" });
	const decided = await service.decideApproval(admin, "s-appr", approvals[0].id, true);
	expect(decided.status).toBe("approved");
	expect(service.approvals.listPending("s-appr")).toHaveLength(0);

	await service.close();
});
