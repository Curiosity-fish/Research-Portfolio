/** 执行器资源限额。 */
export interface ExecutorLimits {
	memoryBytes: number;
	cpuCount: number;
	/** 会话超时(ms),超时自动销毁。 */
	timeoutMs: number;
}

/** 数据集/目录挂载规格。 */
export interface MountSpec {
	hostPath: string;
	containerPath: string;
	readonly: boolean;
}

/** 启动一个执行会话所需的全部上下文。 */
export interface ExecutorStartOptions {
	sessionId: string;
	/** 归属用户(用于并发上限)。 */
	userId: string;
	/** 宿主项目目录(读写挂载到容器 cwd)。 */
	projectHostPath: string;
	/** 容器内工作目录,例如 /work。 */
	cwd: string;
	/** 可见技能目录(宿主路径);容器内映射到 /skills/<目录名>。 */
	skills: string[];
	/** 数据集只读挂载。 */
	datasets: MountSpec[];
	/** 注入环境变量(Ollama 端点 / BYOK key,仅本会话容器)。 */
	env: Record<string, string>;
	/** 会话权限模式(不改变隔离参数;执行面拦截由桥接层处理)。 */
	permissionMode?: "request_approval" | "full_access";
	limits: ExecutorLimits;
	/** tfa 附加参数,如 --provider ollama --model xxx。 */
	command?: string[];
}

export type ExecutorStatusKind = "running" | "exited" | "missing" | "error" | "expired";

export interface ExecutorStatus {
	id: string;
	containerId: string | null;
	state: ExecutorStatusKind;
	startedAt: number | null;
	expiresAt: number | null;
}

export interface ExecutorHandle {
	id: string;
	containerId: string | null;
}

/** 命令执行抽象:生产用真实进程,测试注入 fake 以校验命令构造。 */
export interface CommandRunOptions {
	cwd?: string;
	env?: Record<string, string>;
}

export interface CommandRunner {
	run(command: string, args: string[], options?: CommandRunOptions): Promise<{ stdout: string; stderr: string }>;
}

export interface SessionExecutor {
	start(opts: ExecutorStartOptions): Promise<ExecutorHandle>;
	/** 终止并销毁(容器 stop + rm)。 */
	stop(id: string, timeoutSeconds?: number): Promise<void>;
	inspect(id: string): Promise<ExecutorStatus>;
	list(): Promise<ExecutorStatus[]>;
	/** 停止并清理所有超时会话。 */
	pruneExpired(): Promise<void>;
}
