import { describe, expect, it } from "vitest";
import { resolveModelSelection } from "../src/tfa-harness.ts";

describe("resolveModelSelection", () => {
	it("prefers an explicit harness model over environment defaults", () => {
		expect(
			resolveModelSelection(
				{ provider: "anthropic", id: "claude-opus-4-6" },
				{ TFA_PROVIDER: "openai-codex", TFA_MODEL: "gpt-5.6-sol" },
			),
		).toEqual({ provider: "anthropic", id: "claude-opus-4-6" });
	});

	it("uses trimmed environment defaults when the harness has no explicit model", () => {
		expect(resolveModelSelection(undefined, { TFA_PROVIDER: " openai-codex ", TFA_MODEL: " gpt-5.6-sol " })).toEqual({
			provider: "openai-codex",
			id: "gpt-5.6-sol",
		});
	});

	it.each([
		[undefined, {}],
		[undefined, { TFA_PROVIDER: "openai-codex" }],
		[undefined, { TFA_MODEL: "gpt-5.6-sol" }],
		[
			{ provider: "", id: "gpt-5.6-sol" },
			{ TFA_PROVIDER: "openai-codex", TFA_MODEL: "gpt-5.6-sol" },
		],
	] as const)("rejects an incomplete model selection", (explicitModel, environment) => {
		expect(() => resolveModelSelection(explicitModel, environment)).toThrow(
			"Select a harness model explicitly or set both TFA_PROVIDER and TFA_MODEL as defaults.",
		);
	});
});
