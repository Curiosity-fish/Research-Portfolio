import type {
	Command,
	ModelMetadata,
	ModelRef,
	SessionPhase,
	SessionSnapshot,
	SessionSummary,
	ThinkingLevel,
	TranscriptProgress,
} from "@earendil-works/tfa-protocol";
import type { TfaServerError } from "./errors.ts";
import type { TfaServerListener } from "./listener.ts";

export interface TfaServerOptions {
	listeners: readonly TfaServerListener[];
	maxFrameLength?: number;
	handshakeTimeoutMs?: number;
	serverId?: string;
	onError?: (error: Error) => void;
}

export type MaybePromise<T> = T | Promise<T>;

export type PromptInput = Omit<Extract<Command, { command: "prompt" }>, "command" | "sessionId">;
export type SteerInput = Omit<Extract<Command, { command: "steer" }>, "command" | "sessionId">;

export interface CreateSessionOptions {
	/** A collision-resistant ID assigned by TfaServer. The service must persist this exact ID. */
	id: string;
	cwd?: string;
	name?: string;
	model?: ModelRef;
	thinkingLevel?: ThinkingLevel;
}

export type TfaSessionRuntimeEvent =
	| { type: "snapshot" }
	| { type: "progress"; progress: TranscriptProgress }
	| { type: "error"; error: TfaServerError };

/** One acquired durable session. Conflicting operations must reject rather than queue. */
export interface TfaSessionRuntime {
	snapshot(): MaybePromise<SessionSnapshot>;
	getPhase(): SessionPhase;
	prompt(input: PromptInput): Promise<void>;
	steer(input: SteerInput): Promise<void>;
	abort(): Promise<void>;
	setModel(model: ModelRef): Promise<void>;
	setThinking(thinkingLevel: ThinkingLevel): Promise<void>;
	subscribe(listener: (event: TfaSessionRuntimeEvent) => void): () => void;
	dispose(): Promise<void>;
}

/** Service boundary for durable sessions and exclusively acquired runtimes. */
export interface TfaServerService {
	listSessions(): Promise<SessionSummary[]>;
	listModels(): Promise<ModelMetadata[]>;
	createSession(options: CreateSessionOptions): Promise<TfaSessionRuntime>;
	openSession(sessionId: string): Promise<TfaSessionRuntime>;
}

export type SessionRuntime = TfaSessionRuntime;
export type SessionRuntimeEvent = TfaSessionRuntimeEvent;
