import type { ExecutorConfig } from "./config.ts";
import { DockerExecutor } from "./docker.ts";
import { LocalExecutor } from "./local.ts";
import type { CommandRunner, SessionExecutor } from "./types.ts";

export type { ExecutorConfig } from "./config.ts";
export { loadExecutorConfig, parseBytes } from "./config.ts";
export { DockerExecutor } from "./docker.ts";
export { LocalExecutor } from "./local.ts";
export type {
	CommandRunner,
	CommandRunOptions,
	ExecutorHandle,
	ExecutorLimits,
	ExecutorStartOptions,
	ExecutorStatus,
	ExecutorStatusKind,
	MountSpec,
	SessionExecutor,
} from "./types.ts";

export function createSessionExecutor(
	kind: "docker" | "local",
	config: ExecutorConfig,
	runner?: CommandRunner,
): SessionExecutor {
	return kind === "docker" ? new DockerExecutor(config, runner) : new LocalExecutor(config);
}
