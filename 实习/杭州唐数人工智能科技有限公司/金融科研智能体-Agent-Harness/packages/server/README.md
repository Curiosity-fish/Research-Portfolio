# @earendil-works/tfa-server

Experimental. This package is under active development and may change or be removed without notice. Its APIs and behavior are not yet stable.

Server package for tfa.

## Session server core

The package exports the `TfaServer` session server.

```ts
import type { TfaServerService } from "@earendil-works/tfa-server";
import { createUnixServer } from "@earendil-works/tfa-server/unix";

const service: TfaServerService = {
  async listSessions() {
    return storage.listSessions();
  },
  async listModels() {
    return modelRegistry.listModels();
  },
  async createSession(options) {
    return storage.createAndOpen(options);
  },
  async openSession(sessionId) {
    return storage.open(sessionId);
  },
};

const server = createUnixServer(service, {
  path: "/tmp/tfa/server.sock",
});
await server.start();
```

`TfaServer` composes transport listeners through the `TfaServerListener` interface. Each listener must complete any transport-specific authentication and authorization before passing a connection to `TfaServer`. For example, a WebSocket listener can validate credentials during the HTTP upgrade, while the Unix listener relies on socket filesystem permissions. The Unix submodule exports the `createUnixListener()` building block and `createUnixServer()` preset, keeping the common case concise without coupling the primary server to Unix sockets. The listener uses length-prefixed CBOR messages from `@earendil-works/tfa-protocol`.

This package does not provide a standalone CLI or coding-agent service. Applications supply the `TfaServerService` implementation.

## Transport testing

Custom transports can use `@earendil-works/tfa-server/testing` for deterministic protocol conformance tests. It exports `createTestServer()`, `TestServerService`, `ProtocolTestClient`, and the transport-neutral `WireChannel` contract. `connectUnixTestClient()` is provided for Unix transport tests.

## `tfa-ai` protocol bridge

`@earendil-works/tfa-ai` domain objects and `@earendil-works/tfa-protocol` wire DTOs remain independent. This package owns their boundary and exports `toProtocolModelMetadata()`, `toProtocolAssistantMessage()`, `toProtocolUserMessage()`, and `toProtocolToolResultMessage()`.

The adapters reject invalid tool inputs, identifiers, timestamps, and mismatched tool results; `toProtocolToolResultMessage()` requires the original `ToolCall` so it can verify the association and convert its arguments itself. Diagnostic details are explicitly sanitized. Closed `tfa-ai` unions are mapped exhaustively, and compile-time field manifests enumerate current `tfa-ai` properties so additions require an explicit review. The protocol mirrors `tfa-ai` vocabulary such as `toolCall` and `toolUse` where the semantics are identical. Protocol schemas enforce consistent lifecycle states, and tests encode adapter output through the runtime schemas so incompatible changes fail in the bridging package.
