import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

export const workspaceSourcePaths = {
	telemetryIndex: fileURLToPath(new URL("./packages/telemetry/src/index.ts", import.meta.url)),
	aiIndex: fileURLToPath(new URL("./packages/ai/src/index.ts", import.meta.url)),
	aiCompat: fileURLToPath(new URL("./packages/ai/src/compat.ts", import.meta.url)),
	aiOAuth: fileURLToPath(new URL("./packages/ai/src/oauth.ts", import.meta.url)),
	aiProviders: fileURLToPath(new URL("./packages/ai/src/providers", import.meta.url)),
	agentIndex: fileURLToPath(new URL("./packages/agent/src/index.ts", import.meta.url)),
	codingAgentIndex: fileURLToPath(new URL("./packages/coding-agent/src/index.ts", import.meta.url)),
	tuiIndex: fileURLToPath(new URL("./packages/tui/src/index.ts", import.meta.url)),
} as const;

export default defineConfig({
	resolve: {
		alias: [
			{ find: /^@earendil-works\/tfa-telemetry$/, replacement: workspaceSourcePaths.telemetryIndex },
			{ find: /^@earendil-works\/tfa-ai$/, replacement: workspaceSourcePaths.aiIndex },
			{ find: /^@earendil-works\/tfa-ai\/compat$/, replacement: workspaceSourcePaths.aiCompat },
			{ find: /^@earendil-works\/tfa-ai\/oauth$/, replacement: workspaceSourcePaths.aiOAuth },
			{
				find: /^@earendil-works\/tfa-ai\/providers\/(.+)$/,
				replacement: `${workspaceSourcePaths.aiProviders}/$1.ts`,
			},
			{ find: /^@earendil-works\/tfa-agent-core$/, replacement: workspaceSourcePaths.agentIndex },
			{ find: /^@earendil-works\/tfa-tui$/, replacement: workspaceSourcePaths.tuiIndex },
		],
	},
});
