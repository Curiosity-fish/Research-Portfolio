import { describe, expect, it } from "vitest";
import type { ExecutorConfig } from "../src/server/executor/config.ts";
import { loadExecutorConfig, parseBytes } from "../src/server/executor/config.ts";
import { DockerExecutor } from "../src/server/executor/docker.ts";
import { LocalExecutor } from "../src/server/executor/local.ts";
import type { CommandRunner, CommandRunOptions, ExecutorStartOptions } from "../src/server/executor/types.ts";

class FakeRunner implements CommandRunner {
	calls: Array<{ command: string; args: string[]; options?: CommandRunOptions }> = [];
	queue: string[] = [];
	failQueue: Error[] = [];

	run(command: string, args: string[], options?: CommandRunOptions): Promise<{ stdout: string; stderr: string }> {
		this.calls.push({ command, args, options });
		const error = this.failQueue.shift();
		if (error) return Promise.reject(error);
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

function startOpts(
	sessionId: string,
	userId = "u-1",
	overrides: Partial<ExecutorStartOptions> = {},
): ExecutorStartOptions {
	return {
		sessionId,
		userId,
		projectHostPath: "/home/user/proj",
		cwd: "/work",
		skills: ["/srv/skills/sk1", "/srv/skills/sk2"],
		datasets: [{ hostPath: "/data/d1", containerPath: "/data/d1", readonly: true }],
		env: { OLLAMA_BASE_URL: "http://host:11434/v1" },
		limits: { memoryBytes: 2 * 1024 ** 3, cpuCount: 2, timeoutMs: 60_000 },
		...overrides,
	};
}

const sleep = (ms: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, ms));

describe("DockerExecutor", () => {
	it("start 构造正确的 docker run 命令(限额/挂载/技能注入/env)", async () => {
		const runner = new FakeRunner();
		runner.queue.push("container-abc\n");
		const executor = new DockerExecutor(makeConfig(), runner);

		const handle = await executor.start(startOpts("s-1"));
		expect(handle).toEqual({ id: "s-1", containerId: "container-abc" });
		expect(runner.calls).toHaveLength(1);

		const call = runner.calls[0];
		expect(call.command).toBe("docker");
		expect(call.args).toEqual([
			"run",
			"-d",
			"--name",
			"tfa-s-1",
			"--memory",
			String(2 * 1024 ** 3),
			"--cpus",
			"2",
			"-v",
			"/home/user/proj:/work",
			"-v",
			"/srv/skills:/skills:ro",
			"-w",
			"/work",
			"-v",
			"/data/d1:/data/d1:ro",
			"-e",
			"OLLAMA_BASE_URL=http://host:11434/v1",
			"tfa-executor",
			"tfa",
			"--no-skills",
			"--skill",
			"/skills/sk1",
			"--skill",
			"/skills/sk2",
		]);
	});

	it("同一用户并发上限生效,不同用户不受限", async () => {
		const runner = new FakeRunner();
		const executor = new DockerExecutor(makeConfig(), runner);

		await executor.start(startOpts("s-1", "u-1"));
		await executor.start(startOpts("s-2", "u-1"));
		await expect(executor.start(startOpts("s-3", "u-1"))).rejects.toMatchObject({ code: "invalid_request" });

		await expect(executor.start(startOpts("s-4", "u-2"))).resolves.toBeTruthy();
	});

	it("stop 执行 docker stop --time 30 + rm -f,并从跟踪移除", async () => {
		const runner = new FakeRunner();
		const executor = new DockerExecutor(makeConfig(), runner);
		await executor.start(startOpts("s-1"));

		await executor.stop("s-1");
		const stopCall = runner.calls.find((c) => c.args[0] === "stop");
		expect(stopCall?.args).toEqual(["stop", "--time", "30", "tfa-s-1"]);
		const rmCall = runner.calls.find((c) => c.args[0] === "rm");
		expect(rmCall?.args).toEqual(["rm", "-f", "tfa-s-1"]);

		expect((await executor.inspect("s-1")).state).toBe("missing");
	});

	it("inspect:running / missing / expired", async () => {
		const runner = new FakeRunner();
		const executor = new DockerExecutor(makeConfig(), runner);

		// running:start 消费容器 ID,inspect 消费状态
		runner.queue.push("cid1\n", "running\n");
		await executor.start(startOpts("s-1"));
		expect((await executor.inspect("s-1")).state).toBe("running");

		// missing:docker inspect 失败
		const executor2 = new DockerExecutor(makeConfig(), new FakeRunner());
		await executor2.start(startOpts("s-2"));
		(executor2 as unknown as { runner: FakeRunner }).runner.failQueue.push(new Error("No such container"));
		expect((await executor2.inspect("s-2")).state).toBe("missing");

		// 未跟踪的 id
		expect((await executor.inspect("no-such")).state).toBe("missing");

		// expired:超时后状态为 expired(超时来自 opts.limits.timeoutMs)
		const executor3 = new DockerExecutor(makeConfig(), new FakeRunner());
		await executor3.start(
			startOpts("s-3", "u-1", { limits: { memoryBytes: 2 * 1024 ** 3, cpuCount: 2, timeoutMs: 1 } }),
		);
		await sleep(10);
		expect((await executor3.inspect("s-3")).state).toBe("expired");
	});

	it("pruneExpired 停止并清理超时会话", async () => {
		const runner = new FakeRunner();
		const executor = new DockerExecutor(makeConfig(), runner);
		await executor.start(
			startOpts("s-1", "u-1", { limits: { memoryBytes: 2 * 1024 ** 3, cpuCount: 2, timeoutMs: 1 } }),
		);
		await sleep(10);

		await executor.pruneExpired();
		const stopCall = runner.calls.find((c) => c.args[0] === "stop");
		expect(stopCall?.args).toEqual(["stop", "--time", "30", "tfa-s-1"]);
		expect(await executor.list()).toEqual([]);
	});
});

describe("LocalExecutor", () => {
	it("start 跟踪 running 状态并执行并发上限", async () => {
		const executor = new LocalExecutor(makeConfig());
		await executor.start(startOpts("s-1"));
		expect((await executor.inspect("s-1")).state).toBe("running");

		await executor.start(startOpts("s-2", "u-1"));
		await expect(executor.start(startOpts("s-3", "u-1"))).rejects.toMatchObject({ code: "invalid_request" });

		await executor.stop("s-1");
		expect((await executor.inspect("s-1")).state).toBe("missing");
	});
});

describe("executor config", () => {
	it("parseBytes 解析 500m/2g/纯数字", () => {
		expect(parseBytes("2g", 0)).toBe(2 * 1024 ** 3);
		expect(parseBytes("500m", 0)).toBe(500 * 1024 ** 2);
		expect(parseBytes("8", 0)).toBe(8);
		expect(parseBytes(undefined, 123)).toBe(123);
		expect(parseBytes("junk", 7)).toBe(7);
	});

	it("loadExecutorConfig 默认值与环境覆盖", () => {
		const defaults = loadExecutorConfig({}, "/srv/skills");
		expect(defaults.memoryBytes).toBe(2 * 1024 ** 3);
		expect(defaults.cpuCount).toBe(2);
		expect(defaults.timeoutMs).toBe(30 * 60 * 1000);
		expect(defaults.maxConcurrentPerUser).toBe(2);
		expect(defaults.image).toBe("tfa-executor");
		expect(defaults.skillsHostDir).toBe("/srv/skills");

		const overridden = loadExecutorConfig(
			{
				EXECUTOR_MEMORY: "1g",
				EXECUTOR_CPUS: "4",
				SESSION_TIMEOUT_MS: "1000",
				MAX_CONCURRENT_SESSIONS_PER_USER: "5",
				EXECUTOR_IMAGE: "img",
			},
			"/srv/skills",
		);
		expect(overridden).toMatchObject({
			memoryBytes: 1024 ** 3,
			cpuCount: 4,
			timeoutMs: 1000,
			maxConcurrentPerUser: 5,
			image: "img",
		});
	});
});
