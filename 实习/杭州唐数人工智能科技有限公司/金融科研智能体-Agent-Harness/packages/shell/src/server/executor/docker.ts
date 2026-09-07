import { execFile } from "node:child_process";
import { basename } from "node:path";
import { promisify } from "node:util";
import { BaseExecutor, type TrackedSession } from "./base.ts";
import type { ExecutorConfig } from "./config.ts";
import type {
	CommandRunner,
	ExecutorHandle,
	ExecutorStartOptions,
	ExecutorStatus,
	ExecutorStatusKind,
} from "./types.ts";

const execFileAsync = promisify(execFile);

const defaultDockerRunner: CommandRunner = {
	async run(command, args, options) {
		const { stdout, stderr } = await execFileAsync(command, args, {
			...(options?.cwd !== undefined ? { cwd: options.cwd } : {}),
			...(options?.env ? { env: { ...process.env, ...options.env } } : {}),
			timeout: 30_000,
			windowsHide: true,
		});
		return { stdout, stderr };
	},
};

/**
 * Docker 会话执行器:每个会话一个容器,跑 tfa CLI。
 * 注入项目可见技能(--no-skills + 显式 --skill,防全局技能泄漏)、数据集只读挂载、限额与超时。
 */
export class DockerExecutor extends BaseExecutor {
	private readonly runner: CommandRunner;

	constructor(config: ExecutorConfig, runner?: CommandRunner) {
		super(config);
		this.runner = runner ?? defaultDockerRunner;
	}

	private containerName(sessionId: string): string {
		return `tfa-${sessionId}`;
	}

	async start(opts: ExecutorStartOptions): Promise<ExecutorHandle> {
		this.assertConcurrency(opts.userId);
		const args = [
			"run",
			"-d",
			"--name",
			this.containerName(opts.sessionId),
			"--memory",
			String(opts.limits.memoryBytes),
			"--cpus",
			String(opts.limits.cpuCount),
			"-v",
			`${opts.projectHostPath}:${opts.cwd}`,
			"-v",
			`${this.config.skillsHostDir}:/skills:ro`,
			"-w",
			opts.cwd,
		];
		for (const mount of opts.datasets) {
			args.push("-v", `${mount.hostPath}:${mount.containerPath}:ro`);
		}
		for (const [key, value] of Object.entries(opts.env)) {
			args.push("-e", `${key}=${value}`);
		}
		// A1 结论:必须 --no-skills + 显式 --skill,禁用默认全局技能目录,防止跨会话泄漏
		args.push(this.config.image, "tfa", "--no-skills");
		for (const skillHostPath of opts.skills) {
			args.push("--skill", `/skills/${basename(skillHostPath)}`);
		}
		if (opts.command) args.push(...opts.command);

		const { stdout } = await this.runner.run("docker", args);
		const containerId = stdout.trim() || null;
		return this.register(opts, containerId);
	}

	/** 终止并销毁:docker stop --time + rm -f。 */
	async stop(id: string, timeoutSeconds = 30): Promise<void> {
		const name = this.containerName(id);
		await this.runner.run("docker", ["stop", "--time", String(timeoutSeconds), name]).catch(() => {
			// 容器可能已不存在
		});
		await this.runner.run("docker", ["rm", "-f", name]).catch(() => {
			// 容器可能已不存在
		});
		this.tracked.delete(id);
	}

	protected async doInspect(id: string, entry: TrackedSession): Promise<ExecutorStatus> {
		if (this.isExpired(entry)) {
			return {
				id,
				containerId: entry.containerId,
				state: "expired",
				startedAt: entry.startedAt,
				expiresAt: entry.expiresAt,
			};
		}
		const name = this.containerName(id);
		try {
			const { stdout } = await this.runner.run("docker", ["inspect", "--format", "{{.State.Status}}", name]);
			const state: ExecutorStatusKind = stdout.trim() === "running" ? "running" : "exited";
			return {
				id,
				containerId: entry.containerId,
				state,
				startedAt: entry.startedAt,
				expiresAt: entry.expiresAt,
			};
		} catch {
			return {
				id,
				containerId: entry.containerId,
				state: "missing",
				startedAt: entry.startedAt,
				expiresAt: entry.expiresAt,
			};
		}
	}
}
