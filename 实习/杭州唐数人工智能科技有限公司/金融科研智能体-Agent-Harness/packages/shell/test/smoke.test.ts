import { describe, expect, it } from "vitest";
import {
	createShellRuntimeFactory,
	FintechToolRegistry,
	ShellSessionService,
	shellMain,
	startHttpServer,
} from "../src/index.ts";

describe("tfa-shell exports", () => {
	it("index exports load", () => {
		expect(typeof shellMain).toBe("function");
		expect(typeof createShellRuntimeFactory).toBe("function");
		expect(typeof ShellSessionService).toBe("function");
		expect(typeof startHttpServer).toBe("function");
		expect(typeof FintechToolRegistry).toBe("function");
	});
});
