import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

const telemetrySrcIndex = fileURLToPath(new URL("../telemetry/src/index.ts", import.meta.url));
const aiSrcIndex = fileURLToPath(new URL("../ai/src/index.ts", import.meta.url));
const aiSrcCompat = fileURLToPath(new URL("../ai/src/compat.ts", import.meta.url));
const agentSrcIndex = fileURLToPath(new URL("./src/index.ts", import.meta.url));

export default defineConfig({
	test: {
		globals: true,
		environment: "node",
		testTimeout: 30000, // 30 seconds for API calls
		reporters: process.env.GITHUB_ACTIONS ? ["dot", "github-actions"] : ["dot"],
		silent: "passed-only",
	},
	resolve: {
		alias: [
			{ find: /^@earendil-works\/tfa-telemetry$/, replacement: telemetrySrcIndex },
			{ find: /^@earendil-works\/tfa-agent-core$/, replacement: agentSrcIndex },
			{ find: /^@earendil-works\/tfa-ai$/, replacement: aiSrcIndex },
			{ find: /^@earendil-works\/tfa-ai\/compat$/, replacement: aiSrcCompat },
		],
	},
});
