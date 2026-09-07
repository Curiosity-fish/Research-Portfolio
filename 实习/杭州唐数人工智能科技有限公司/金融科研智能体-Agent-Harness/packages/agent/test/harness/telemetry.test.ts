import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { NOOP_TELEMETRY_CONTEXT, type TelemetryContext } from "@earendil-works/tfa-telemetry";
import { describe, expect, expectTypeOf, it } from "vitest";
import { renderAgentTelemetrySchemaMarkdown } from "../../scripts/generate-telemetry-docs.ts";
import {
	AI_TELEMETRY_SCHEMA,
	type AiSpanEndAttributes,
	type AiSpanStartAttributes,
	HARNESS_TELEMETRY_SCHEMA,
	type HarnessSpanEndAttributes,
	type HarnessSpanStartAttributes,
	startAiSpan,
	startHarnessSpan,
} from "../../src/harness/telemetry.ts";

describe("agent telemetry schemas", () => {
	it("serializes both schemas and generates the checked-in reference", () => {
		expect(() => JSON.stringify(AI_TELEMETRY_SCHEMA)).not.toThrow();
		expect(() => JSON.stringify(HARNESS_TELEMETRY_SCHEMA)).not.toThrow();
		expect(Object.keys(HARNESS_TELEMETRY_SCHEMA.spans)).toEqual([
			"tfa.harness.run",
			"tfa.harness.compaction",
			"tfa.harness.navigation",
			"tfa.harness.checkpoint",
			"tfa.harness.turn",
			"tfa.harness.step",
			"tfa.harness.tool",
			"tfa.harness.hook",
			"tfa.harness.sleep",
			"tfa.harness.event_handler",
			"tfa.session.write",
		]);
		const actual = readFileSync(resolve(import.meta.dirname, "../../docs/telemetry-schema.md"), "utf8");
		expect(actual).toBe(renderAgentTelemetrySchemaMarkdown());
	});

	it("infers exact AI start and optional end attributes", async () => {
		type Start = AiSpanStartAttributes<"tfa.ai.request">;
		type End = AiSpanEndAttributes<"tfa.ai.request">;
		expectTypeOf<Start>().toMatchTypeOf<{
			"tfa.ai.operation": "stream" | "fetch_deferred" | "cancel_deferred" | "generate_images";
			"tfa.ai.provider": string;
			"tfa.ai.model": string;
			"tfa.ai.api": string;
			"tfa.ai.streaming": boolean;
			"tfa.ai.deferred"?: boolean;
		}>();
		expectTypeOf<End["tfa.ai.response.stop_reason"]>().toEqualTypeOf<
			"stop" | "length" | "tool_use" | "error" | "aborted" | "deferred" | undefined
		>();

		const telemetryContext: TelemetryContext = NOOP_TELEMETRY_CONTEXT;
		await startAiSpan(
			telemetryContext,
			"tfa.ai.request",
			{
				"tfa.ai.operation": "stream",
				"tfa.ai.provider": "provider",
				"tfa.ai.model": "model",
				"tfa.ai.api": "api",
				"tfa.ai.streaming": true,
			},
			(span) => {
				span.setAttributes({ "tfa.ai.response.stop_reason": "tool_use" });
				// @ts-expect-error tfa.ai.request declares no span events
				span.addEvent("chunk");
			},
		);

		const compileTimeFailures = () => {
			const extraAttributes = {
				"tfa.ai.operation": "stream",
				"tfa.ai.provider": "provider",
				"tfa.ai.model": "model",
				"tfa.ai.api": "api",
				"tfa.ai.streaming": true,
				"tfa.ai.unknown": true,
			} as const;
			// @ts-expect-error variables with unknown attributes are rejected
			void startAiSpan(telemetryContext, "tfa.ai.request", extraAttributes, () => {});
			// @ts-expect-error missing required start attributes
			void startAiSpan(telemetryContext, "tfa.ai.request", { "tfa.ai.operation": "stream" }, () => {});
		};
		expectTypeOf(compileTimeFailures).toBeFunction();
	});

	it("infers per-span harness literals and optional completion enrichment", async () => {
		type RunStart = HarnessSpanStartAttributes<"tfa.harness.run">;
		type RunEnd = HarnessSpanEndAttributes<"tfa.harness.run">;
		expectTypeOf<RunStart["tfa.operation.kind"]>().toEqualTypeOf<"run">();
		expectTypeOf<RunEnd["tfa.operation.outcome"]>().toEqualTypeOf<
			"completed" | "aborted" | "failed" | "suspended" | undefined
		>();

		const telemetryContext: TelemetryContext = NOOP_TELEMETRY_CONTEXT;
		await startHarnessSpan(
			telemetryContext,
			"tfa.harness.run",
			{
				"tfa.session.id": "session",
				"tfa.lane.name": "main",
				"tfa.operation.id": "operation",
				"tfa.operation.kind": "run",
				"tfa.operation.recovery": false,
			},
			(span) => {
				span.setAttributes({ "tfa.operation.outcome": "completed" });
				span.setAttributes({});
				// @ts-expect-error the harness schema declares no span events
				span.addEvent("result");
			},
		);

		const compileTimeFailures = () => {
			const extraRunAttributes = {
				"tfa.session.id": "session",
				"tfa.lane.name": "main",
				"tfa.operation.id": "operation",
				"tfa.operation.kind": "run",
				"tfa.operation.recovery": false,
				"tfa.unknown": true,
			} as const;
			// @ts-expect-error variables with unknown attributes are rejected
			void startHarnessSpan(telemetryContext, "tfa.harness.run", extraRunAttributes, () => {});
			void startHarnessSpan(
				telemetryContext,
				"tfa.harness.checkpoint",
				{
					"tfa.lane.name": "main",
					"tfa.operation.id": "operation",
					"tfa.checkpoint.kind": "normal",
				},
				(span) => {
					// @ts-expect-error empty end schemas reject every attribute
					span.setAttributes({ "tfa.unknown": true });
				},
			);
			void startHarnessSpan(
				telemetryContext,
				"tfa.harness.run",
				{
					"tfa.session.id": "session",
					"tfa.lane.name": "main",
					"tfa.operation.id": "operation",
					// @ts-expect-error run spans accept only the run operation kind
					"tfa.operation.kind": "navigation",
					"tfa.operation.recovery": false,
				},
				() => {},
			);
			// @ts-expect-error missing required run start attributes
			void startHarnessSpan(telemetryContext, "tfa.harness.run", {}, () => {});
		};
		expectTypeOf(compileTimeFailures).toBeFunction();
	});
});
