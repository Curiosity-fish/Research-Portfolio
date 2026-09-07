import type { JsonValue, ProtocolError, ProtocolErrorCode } from "@earendil-works/tfa-protocol";

export class TfaServerError extends Error {
	readonly code: ProtocolErrorCode;
	readonly details: JsonValue | undefined;

	constructor(error: ProtocolError) {
		super(error.message);
		this.name = "TfaServerError";
		this.code = error.code;
		this.details = error.details;
	}
}

export class TfaDisconnectedError extends Error {
	constructor(message = "TFA client is disconnected") {
		super(message);
		this.name = "TfaDisconnectedError";
	}
}

export class TfaClientDisposedError extends Error {
	constructor() {
		super("TFA client is disposed");
		this.name = "TfaClientDisposedError";
	}
}

export class TfaSessionOwnershipError extends Error {
	readonly sessionId: string;

	constructor(sessionId: string, message: string) {
		super(message);
		this.name = "TfaSessionOwnershipError";
		this.sessionId = sessionId;
	}
}

export class TfaSessionDetachedError extends Error {
	readonly sessionId: string;

	constructor(sessionId: string) {
		super(`Session ${sessionId} is not attached`);
		this.name = "TfaSessionDetachedError";
		this.sessionId = sessionId;
	}
}

export function toError(error: unknown): Error {
	return error instanceof Error ? error : new Error(String(error));
}

export function toDisconnectedError(error: unknown): TfaDisconnectedError {
	const cause = toError(error);
	return cause instanceof TfaDisconnectedError ? cause : new TfaDisconnectedError(cause.message);
}
