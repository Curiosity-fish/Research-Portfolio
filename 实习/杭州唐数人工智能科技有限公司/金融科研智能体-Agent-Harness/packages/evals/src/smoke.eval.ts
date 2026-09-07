import { expect } from "vitest";
import { describeEval } from "vitest-evals";
import { createTfaCodingAgentHarness } from "./tfa-harness.ts";

const tfaCodingAgentHarness = createTfaCodingAgentHarness({ noTools: "all" });

describeEval("TFA Coding Agent smoke", { harness: tfaCodingAgentHarness }, (it) => {
	it("runs a basic prompt end to end", async ({ run }) => {
		const result = await run("What's the capital of France? Respond with only the city name.");

		expect(result.output.trim()).toBe("Paris");
		expect(result.errors).toEqual([]);
		expect(result.usage.provider).toBe(process.env.TFA_PROVIDER);
		expect(result.usage.model).toBe(process.env.TFA_MODEL);
		expect(result.usage.totalTokens).toBeGreaterThan(0);
	});
});
