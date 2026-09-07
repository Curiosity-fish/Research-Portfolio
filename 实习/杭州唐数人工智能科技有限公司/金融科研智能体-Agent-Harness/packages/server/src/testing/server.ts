import { TfaServer } from "../server.ts";
import type { TfaServerOptions, TfaServerService } from "../types.ts";
import { TestServerService } from "./service.ts";

export interface TestServerOptions extends TfaServerOptions {
	service?: TfaServerService;
}

export interface TestServer {
	server: TfaServer;
	service: TfaServerService;
}

/** Create an unstarted TfaServer with deterministic defaults for transport conformance tests. */
export function createTestServer(options: TestServerOptions): TestServer {
	const service = options.service ?? new TestServerService();
	return {
		server: new TfaServer(service, {
			listeners: options.listeners,
			maxFrameLength: options.maxFrameLength,
			handshakeTimeoutMs: options.handshakeTimeoutMs,
			serverId: options.serverId,
			onError: options.onError,
		}),
		service,
	};
}
