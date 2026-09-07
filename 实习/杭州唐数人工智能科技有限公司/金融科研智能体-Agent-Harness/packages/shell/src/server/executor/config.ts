export interface ExecutorConfig {
	/** 执行镜像,默认 tfa-executor(T13 出 Dockerfile)。 */
	image: string;
	/** 宿主技能库根目录(挂载到容器 /skills:ro)。 */
	skillsHostDir: string;
	memoryBytes: number;
	cpuCount: number;
	timeoutMs: number;
	/** 同一用户最大并发活跃会话。 */
	maxConcurrentPerUser: number;
	/** LocalExecutor 使用的 tfa 可执行文件。 */
	localBinary: string;
}

/** 解析 "500m"/"2g"/纯数字 为字节。 */
export function parseBytes(value: string | undefined, fallback: number): number {
	if (!value) return fallback;
	const match = /^(\d+(?:\.\d+)?)\s*([kmg]?)$/i.exec(value.trim());
	if (!match) return fallback;
	const n = Number(match[1]);
	const suffix = match[2].toLowerCase();
	const mult = suffix === "k" ? 1024 : suffix === "m" ? 1024 ** 2 : suffix === "g" ? 1024 ** 3 : 1;
	return Math.floor(n * mult);
}

/** 从环境变量加载执行器配置(可注入便于测试)。 */
export function loadExecutorConfig(env: Record<string, string | undefined>, skillsHostDir: string): ExecutorConfig {
	return {
		image: env.EXECUTOR_IMAGE ?? "tfa-executor",
		skillsHostDir,
		memoryBytes: parseBytes(env.EXECUTOR_MEMORY, 2 * 1024 ** 3),
		cpuCount: Number(env.EXECUTOR_CPUS ?? 2),
		timeoutMs: Number(env.SESSION_TIMEOUT_MS ?? 30 * 60 * 1000),
		maxConcurrentPerUser: Number(env.MAX_CONCURRENT_SESSIONS_PER_USER ?? 2),
		localBinary: env.EXECUTOR_LOCAL_BINARY ?? "tfa",
	};
}
