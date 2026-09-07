import { BaseExecutor, type TrackedSession } from "./base.ts";
import type { ExecutorHandle, ExecutorStartOptions, ExecutorStatus } from "./types.ts";

/**
 * 宿主机直连执行器(dev/test 过渡,设计 §8.2)。
 * T8a 先落地状态/并发/超时跟踪;T8b 接入宿主 tfa spawn 与 RPC 桥接(ProxyRuntime)。
 */
export class LocalExecutor extends BaseExecutor {
	async start(opts: ExecutorStartOptions): Promise<ExecutorHandle> {
		this.assertConcurrency(opts.userId);
		return this.register(opts, null);
	}

	async stop(id: string, _timeoutSeconds?: number): Promise<void> {
		this.tracked.delete(id);
	}

	protected async doInspect(id: string, entry: TrackedSession): Promise<ExecutorStatus> {
		return {
			id,
			containerId: null,
			state: this.isExpired(entry) ? "expired" : "running",
			startedAt: entry.startedAt,
			expiresAt: entry.expiresAt,
		};
	}
}
