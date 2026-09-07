/**
 * Titlebar Spinner Extension
 *
 * Shows a braille spinner animation in the terminal title while the agent is working.
 * Uses `ctx.ui.setTitle()` to update the terminal title via the extension API.
 *
 * Usage:
 *   tfa --extension examples/extensions/titlebar-spinner.ts
 */

import path from "node:path";
import type { ExtensionAPI, ExtensionContext } from "@earendil-works/tfa-coding-agent";

const BRAILLE_FRAMES = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];

function getBaseTitle(tfa: ExtensionAPI): string {
	const cwd = path.basename(process.cwd());
	const session = tfa.getSessionName();
	return session ? `TFA - ${session} - ${cwd}` : `TFA - ${cwd}`;
}

export default function (tfa: ExtensionAPI) {
	let timer: ReturnType<typeof setInterval> | null = null;
	let frameIndex = 0;

	function stopAnimation(ctx: ExtensionContext) {
		if (timer) {
			clearInterval(timer);
			timer = null;
		}
		frameIndex = 0;
		ctx.ui.setTitle(getBaseTitle(tfa));
	}

	function startAnimation(ctx: ExtensionContext) {
		stopAnimation(ctx);
		timer = setInterval(() => {
			const frame = BRAILLE_FRAMES[frameIndex % BRAILLE_FRAMES.length];
			const cwd = path.basename(process.cwd());
			const session = tfa.getSessionName();
			const title = session ? `${frame} TFA - ${session} - ${cwd}` : `${frame} TFA - ${cwd}`;
			ctx.ui.setTitle(title);
			frameIndex++;
		}, 80);
	}

	tfa.on("agent_start", async (_event, ctx) => {
		startAnimation(ctx);
	});

	tfa.on("agent_end", async (_event, ctx) => {
		stopAnimation(ctx);
	});

	tfa.on("session_shutdown", async (_event, ctx) => {
		stopAnimation(ctx);
	});
}
