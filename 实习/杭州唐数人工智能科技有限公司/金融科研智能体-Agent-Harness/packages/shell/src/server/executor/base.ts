import { TfaServerError } from "@earendil-works/tfa-server";
import type { ExecutorConfig } from "./config.ts";
import type { ExecutorHandle, ExecutorStartOptions, ExecutorStatus, SessionExecutor } from "./types.ts";

export interface TrackedSession {
	userId: string;
	containerId: string | null;
	startedAt: number;
	expiresAt: number;
}

/** 共享的会话跟踪/并发上限/超时逻辑;具体启动/停止由子类实现。 */
export abstract class BaseExecutor implements SessionExecutor {
	protected readonly config: ExecutorConfig;
	protected readonly tracked = new Map<string, TrackedSession>();

	constructor(config: ExecutorConfig) {
		this.config = config;
	}

	abstract start(opts: ExecutorStartOptions): Promise<ExecutorHandle>;
	abstract stop(id: string, timeoutSeconds?: number): Promise<void>;

	protected isExpired(entry: TrackedSession, now = Date.now()): boolean {
		return now >= entry.expiresAt;
	}

	protected activeCountForUser(userId: string, now = Date.now()): number {
		let count = 0;
		for (const entry of this.tracked.values()) {
			if (entry.userId === userId && !this.isExpired(entry, now)) count += 1;
		}
		return count;
	}

	protected assertConcurrency(userId: string): void {
		const active = this.activeCountForUser(userId);
		if (active >= this.config.maxConcurrentPerUser) {
			throw new TfaServerError("invalid_request", `同一用户并发会话数超过上限 ${this.config.maxConcurrentPerUser}`);
		}
	}

	protected register(opts: ExecutorStartOptions, containerId: string | null): ExecutorHandle {
		const now = Date.now();
		this.tracked.set(opts.sessionId, {
			userId: opts.userId,
			containerId,
			startedAt: now,
			expiresAt: now + opts.limits.timeoutMs,
		});
		return { id: opts.sessionId, containerId };
	}

	async inspect(id: string): Promise<ExecutorStatus> {
		const entry = this.tracked.get(id);
		if (!entry) {
			return { id, containerId: null, state: "missing", startedAt: null, expiresAt: null };
		}
		return this.doInspect(id, entry);
	}

	protected abstract doInspect(id: string, entry: TrackedSession): Promise<ExecutorStatus>;

	async list(): Promise<ExecutorStatus[]> {
		return Promise.all([...this.tracked.keys()].map((id) => this.inspect(id)));
	}

	/** 停止并清理所有超时会话(供服务层定时器调用)。 */
	async pruneExpired(): Promise<void> {
		const now = Date.now();
		for (const [id, entry] of [...this.tracked]) {
			if (this.isExpired(entry, now)) {
				await this.stop(id).catch(() => {
					// 清理失败不阻塞其他会话
				});
				this.tracked.delete(id);
			}
		}
	}
}
