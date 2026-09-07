import type {
	ExactTelemetryAttributes,
	SchemaTelemetrySpan,
	TelemetryContext,
	TelemetrySchemaDefinition,
	TelemetrySchemaSpanEndAttributes,
	TelemetrySchemaSpanEventAttributes,
	TelemetrySchemaSpanEventName,
	TelemetrySchemaSpanName,
	TelemetrySchemaSpanStartAttributes,
	TelemetrySchemaSpanUnion,
	TelemetrySpan,
} from "@earendil-works/tfa-telemetry";

export type {
	AttributeValue,
	ExactTelemetryAttributes,
	SchemaTelemetrySpan,
	SpanAttributes,
	SpanOptions,
	SpanStatus,
	TelemetryAttributeDefinition,
	TelemetryAttributeMetadata,
	TelemetryAttributeType,
	TelemetryContext,
	TelemetryEventAttributeDefinition,
	TelemetryEventDefinition,
	TelemetryParentDefinition,
	TelemetrySchemaDefinition,
	TelemetrySchemaSpanEndAttributes,
	TelemetrySchemaSpanEventAttributes,
	TelemetrySchemaSpanEventName,
	TelemetrySchemaSpanName,
	TelemetrySchemaSpanStartAttributes,
	TelemetrySchemaSpanUnion,
	TelemetrySpan,
	TelemetrySpanDefinition,
	TelemetryStartAttributeDefinition,
} from "@earendil-works/tfa-telemetry";

export const AI_TELEMETRY_SCHEMA = {
	version: 1,
	spans: {
		"tfa.ai.request": {
			description: "One logical request to an AI provider",
			parents: { kind: "any" },
			startAttributes: {
				"tfa.ai.operation": {
					type: "string",
					required: true,
					values: ["stream", "fetch_deferred", "cancel_deferred", "generate_images"],
					description: "Logical provider operation",
				},
				"tfa.ai.provider": {
					type: "string",
					required: true,
					description: "Selected provider id",
				},
				"tfa.ai.model": {
					type: "string",
					required: true,
					description: "Requested model id",
				},
				"tfa.ai.api": {
					type: "string",
					required: true,
					description: "Provider API id",
				},
				"tfa.ai.streaming": {
					type: "boolean",
					required: true,
					description: "Whether this operation returns a stream",
				},
				"tfa.ai.deferred": {
					type: "boolean",
					required: false,
					description: "Whether the operation requests or participates in deferred execution",
				},
			},
			endAttributes: {
				"tfa.ai.response.model": { type: "string", description: "Concrete response model" },
				"tfa.ai.response.id": {
					type: "string",
					cardinality: "high",
					description: "Provider response id",
				},
				"tfa.ai.response.stop_reason": {
					type: "string",
					values: ["stop", "length", "tool_use", "error", "aborted", "deferred"],
					description: "Normalized terminal response reason",
				},
				"tfa.ai.http.status_code": { type: "number", description: "Final HTTP status" },
				"tfa.ai.usage.input_tokens": { type: "number", description: "Reported input tokens" },
				"tfa.ai.usage.output_tokens": { type: "number", description: "Reported output tokens" },
				"tfa.ai.usage.cache_read_tokens": { type: "number", description: "Reported cache-read tokens" },
				"tfa.ai.usage.cache_write_tokens": {
					type: "number",
					description: "Reported cache-write tokens",
				},
				"tfa.ai.usage.reasoning_tokens": { type: "number", description: "Reported reasoning tokens" },
				"tfa.ai.usage.total_tokens": { type: "number", description: "Reported total tokens" },
				"tfa.ai.usage.cost": { type: "number", description: "Reported total cost" },
				"tfa.ai.stream.chunk_count": { type: "number", description: "Streamed update chunk count" },
				"tfa.ai.stream.time_to_first_chunk_ms": {
					type: "number",
					description: "Elapsed milliseconds to first update chunk",
				},
				"tfa.ai.error.type": {
					type: "string",
					cardinality: "low",
					description: "Provider or transport error class",
				},
			},
			status: { default: "ok", errorWhen: "The operation throws or returns an error result" },
		},
	},
} as const satisfies TelemetrySchemaDefinition;

export type AiSpanName = TelemetrySchemaSpanName<typeof AI_TELEMETRY_SCHEMA>;
export type AiSpanStartAttributes<Name extends AiSpanName> = TelemetrySchemaSpanStartAttributes<
	typeof AI_TELEMETRY_SCHEMA,
	Name
>;
export type AiSpanEndAttributes<Name extends AiSpanName> = TelemetrySchemaSpanEndAttributes<
	typeof AI_TELEMETRY_SCHEMA,
	Name
>;
export type AiSpanAttributes<Name extends AiSpanName> = AiSpanStartAttributes<Name> & AiSpanEndAttributes<Name>;
export type AiSpanEventName<Name extends AiSpanName> = TelemetrySchemaSpanEventName<typeof AI_TELEMETRY_SCHEMA, Name>;
export type AiSpanEventAttributes<
	Name extends AiSpanName,
	EventName extends AiSpanEventName<Name>,
> = TelemetrySchemaSpanEventAttributes<typeof AI_TELEMETRY_SCHEMA, Name, EventName>;
export type AiTelemetrySpan<Name extends AiSpanName> = SchemaTelemetrySpan<typeof AI_TELEMETRY_SCHEMA, Name>;
export type AiSpan = TelemetrySchemaSpanUnion<typeof AI_TELEMETRY_SCHEMA>;

export function startAiSpan<Name extends AiSpanName, const Attributes extends AiSpanStartAttributes<Name>, Result>(
	telemetryContext: TelemetryContext,
	name: Name,
	attributes: ExactTelemetryAttributes<AiSpanStartAttributes<Name>, Attributes>,
	callback: (span: AiTelemetrySpan<Name>) => Result | Promise<Result>,
): Promise<Result> {
	return telemetryContext.startSpan({ name, attributes }, (span) => callback(span as AiTelemetrySpan<Name>));
}

const HOOK_NAMES = [
	"before_run",
	"before_resume",
	"before_run_end",
	"transform_context",
	"before_request",
	"before_payload",
	"after_response",
	"before_tool",
	"after_tool",
	"before_compaction",
	"before_navigation",
] as const;

const EVENT_TYPES = [
	"run_start",
	"run_resume",
	"run_suspend",
	"run_abort",
	"run_end",
	"fault",
	"handler_error",
	"turn_start",
	"turn_end",
	"retry_scheduled",
	"retry_start",
	"retry_end",
	"message_start",
	"message_update",
	"message_end",
	"tool_start",
	"tool_update",
	"tool_end",
	"entry_added",
	"write_pending",
	"queue_update",
	"fact_update",
	"config_update",
	"compaction_start",
	"compaction_end",
	"navigation_start",
	"navigation_end",
	"lane_created",
	"usage",
] as const;

const operationStartAttributes = {
	"tfa.session.id": {
		type: "string",
		required: true,
		cardinality: "high",
		description: "Session id",
	},
	"tfa.lane.name": {
		type: "string",
		required: true,
		cardinality: "high",
		description: "Lane name",
	},
	"tfa.operation.id": {
		type: "string",
		required: true,
		cardinality: "high",
		description: "Durable operation id",
	},
	"tfa.operation.recovery": {
		type: "boolean",
		required: true,
		description: "Whether this invocation resumes durable work",
	},
} as const;

const operationErrorAttributes = {
	"tfa.error.code": {
		type: "string",
		cardinality: "low",
		description: "Stable operation error code",
	},
	"tfa.error.type": {
		type: "string",
		cardinality: "low",
		description: "Low-cardinality operation error class",
	},
} as const;

export const HARNESS_TELEMETRY_SCHEMA = {
	version: 1,
	spans: {
		"tfa.harness.run": {
			description: "One admitted in-process run invocation",
			parents: { kind: "root_or_external" },
			startAttributes: {
				...operationStartAttributes,
				"tfa.operation.kind": {
					type: "string",
					required: true,
					values: ["run"],
					description: "Run operation kind",
				},
			},
			endAttributes: {
				"tfa.operation.outcome": {
					type: "string",
					values: ["completed", "aborted", "failed", "suspended"],
					description: "Run invocation outcome",
				},
				...operationErrorAttributes,
			},
			status: { default: "ok", errorWhen: "The run fails or throws" },
		},
		"tfa.harness.compaction": {
			description: "One admitted in-process manual compaction invocation",
			parents: { kind: "root_or_external" },
			startAttributes: {
				...operationStartAttributes,
				"tfa.operation.kind": {
					type: "string",
					required: true,
					values: ["compaction"],
					description: "Compaction operation kind",
				},
			},
			endAttributes: {
				"tfa.operation.outcome": {
					type: "string",
					values: ["completed", "declined", "aborted", "failed"],
					description: "Compaction invocation outcome",
				},
				...operationErrorAttributes,
			},
			status: { default: "ok", errorWhen: "The compaction fails or throws" },
		},
		"tfa.harness.navigation": {
			description: "One admitted in-process navigation invocation",
			parents: { kind: "root_or_external" },
			startAttributes: {
				...operationStartAttributes,
				"tfa.operation.kind": {
					type: "string",
					required: true,
					values: ["navigation"],
					description: "Navigation operation kind",
				},
			},
			endAttributes: {
				"tfa.operation.outcome": {
					type: "string",
					values: ["completed", "declined", "aborted", "failed"],
					description: "Navigation invocation outcome",
				},
				...operationErrorAttributes,
			},
			status: { default: "ok", errorWhen: "The navigation fails or throws" },
		},
		"tfa.harness.checkpoint": {
			description: "One run checkpoint",
			parents: { kind: "spans", spans: ["tfa.harness.run"] },
			startAttributes: {
				"tfa.lane.name": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Lane name",
				},
				"tfa.operation.id": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Durable operation id",
				},
				"tfa.checkpoint.kind": {
					type: "string",
					required: true,
					values: ["normal", "failure_drain", "abort_reconcile"],
					description: "Checkpoint purpose",
				},
			},
			endAttributes: {},
			status: { default: "ok", errorWhen: "Checkpoint work throws" },
		},
		"tfa.harness.turn": {
			description: "One assistant response and its tool batch",
			parents: { kind: "spans", spans: ["tfa.harness.run"] },
			startAttributes: {
				"tfa.lane.name": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Lane name",
				},
				"tfa.operation.id": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Durable operation id",
				},
				"tfa.turn.id": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Invocation-local turn id",
				},
			},
			endAttributes: {},
			status: { default: "ok", errorWhen: "Turn work throws" },
		},
		"tfa.harness.step": {
			description: "One durable retry attempt",
			parents: {
				kind: "spans",
				spans: ["tfa.harness.turn", "tfa.harness.checkpoint", "tfa.harness.compaction", "tfa.harness.navigation"],
			},
			startAttributes: {
				"tfa.lane.name": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Lane name",
				},
				"tfa.operation.id": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Durable operation id",
				},
				"tfa.step.kind": {
					type: "string",
					required: true,
					values: ["assistant", "compaction", "branch_summary"],
					description: "Retryable step kind",
				},
				"tfa.step.attempt": {
					type: "number",
					required: true,
					description: "One-based durable attempt number",
				},
				"tfa.compaction.reason": {
					type: "string",
					required: false,
					values: ["manual", "threshold", "overflow"],
					description: "Compaction trigger",
				},
			},
			endAttributes: {
				"tfa.step.outcome": {
					type: "string",
					values: ["succeeded", "retry", "failed", "aborted", "deferred", "overflow"],
					description: "Attempt outcome",
				},
			},
			status: { default: "ok", errorWhen: "The attempt retries, fails, or throws" },
		},
		"tfa.harness.tool": {
			description: "One raw phase-2 tool execution",
			parents: { kind: "spans", spans: ["tfa.harness.turn", "tfa.harness.run"] },
			startAttributes: {
				"tfa.lane.name": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Lane name",
				},
				"tfa.operation.id": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Durable operation id",
				},
				"tfa.turn.id": {
					type: "string",
					required: false,
					cardinality: "high",
					description: "Invocation-local live turn id",
				},
				"tfa.tool.name": {
					type: "string",
					required: true,
					description: "Tool name",
				},
				"tfa.tool.call_id": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Tool call id",
				},
				"tfa.tool.replay": {
					type: "string",
					required: true,
					values: ["never", "safe"],
					description: "Declared replay policy",
				},
				"tfa.tool.recovery": {
					type: "boolean",
					required: true,
					description: "Whether this is recovery execution",
				},
			},
			endAttributes: {
				"tfa.tool.is_error": {
					type: "boolean",
					description: "Whether raw phase-2 execution returned an error",
				},
			},
			status: { default: "ok", errorWhen: "Raw phase-2 execution returns an error" },
		},
		"tfa.harness.hook": {
			description: "One registered hook handler invocation",
			parents: { kind: "any" },
			startAttributes: {
				"tfa.lane.name": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Lane name",
				},
				"tfa.operation.id": {
					type: "string",
					required: false,
					cardinality: "high",
					description: "Durable operation id when accepted",
				},
				"tfa.hook.name": {
					type: "string",
					required: true,
					values: HOOK_NAMES,
					description: "Hook name",
				},
				"tfa.hook.registration_id": {
					type: "string",
					required: false,
					description: "Stable hook registration id",
				},
			},
			endAttributes: {
				"tfa.hook.outcome": {
					type: "string",
					values: ["completed", "skipped", "blocked", "failed"],
					description: "Handler outcome",
				},
			},
			status: { default: "ok", errorWhen: "The handler throws" },
		},
		"tfa.harness.sleep": {
			description: "One retry delay",
			parents: { kind: "spans", spans: ["tfa.harness.step", "tfa.harness.run"] },
			startAttributes: {
				"tfa.operation.id": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Durable operation id",
				},
				"tfa.sleep.delay_ms": {
					type: "number",
					required: true,
					description: "Requested delay in milliseconds",
				},
			},
			endAttributes: {
				"tfa.sleep.outcome": {
					type: "string",
					values: ["elapsed", "aborted"],
					description: "Delay outcome",
				},
			},
			status: { default: "ok", errorWhen: "Sleep work throws" },
		},
		"tfa.harness.event_handler": {
			description: "One passive event listener invocation",
			parents: { kind: "any" },
			startAttributes: {
				"tfa.event.type": {
					type: "string",
					required: true,
					cardinality: "low",
					values: EVENT_TYPES,
					description: "Delivered harness event type",
				},
				"tfa.lane.name": {
					type: "string",
					required: false,
					cardinality: "high",
					description: "Lane name for lane-scoped events",
				},
			},
			endAttributes: {},
			status: { default: "ok", errorWhen: "The listener throws" },
		},
		"tfa.session.write": {
			description: "One committed session mutation",
			parents: { kind: "any" },
			startAttributes: {
				"tfa.lane.name": {
					type: "string",
					required: true,
					cardinality: "high",
					description: "Lane name",
				},
				"tfa.operation.id": {
					type: "string",
					required: false,
					cardinality: "high",
					description: "Durable operation id when accepted",
				},
				"tfa.session.mutation": {
					type: "string",
					required: true,
					values: ["entry", "record", "lane", "fact"],
					description: "Session mutation kind",
				},
				"tfa.session.item_type": {
					type: "string",
					required: false,
					description: "Entry, record, lane, or fact subtype",
				},
			},
			endAttributes: {
				"tfa.session.seq": {
					type: "number",
					description: "Committed session sequence when exposed",
				},
			},
			status: { default: "ok", errorWhen: "Storage rejects the mutation" },
		},
	},
} as const satisfies TelemetrySchemaDefinition;

export type HarnessSpanName = TelemetrySchemaSpanName<typeof HARNESS_TELEMETRY_SCHEMA>;
export type HarnessSpanStartAttributes<Name extends HarnessSpanName> = TelemetrySchemaSpanStartAttributes<
	typeof HARNESS_TELEMETRY_SCHEMA,
	Name
>;
export type HarnessSpanEndAttributes<Name extends HarnessSpanName> = TelemetrySchemaSpanEndAttributes<
	typeof HARNESS_TELEMETRY_SCHEMA,
	Name
>;
export type HarnessSpanAttributes<Name extends HarnessSpanName> = HarnessSpanStartAttributes<Name> &
	HarnessSpanEndAttributes<Name>;
export type HarnessSpanEventName<Name extends HarnessSpanName> = TelemetrySchemaSpanEventName<
	typeof HARNESS_TELEMETRY_SCHEMA,
	Name
>;
export type HarnessSpanEventAttributes<
	Name extends HarnessSpanName,
	EventName extends HarnessSpanEventName<Name>,
> = TelemetrySchemaSpanEventAttributes<typeof HARNESS_TELEMETRY_SCHEMA, Name, EventName>;
export type HarnessTelemetrySpan<Name extends HarnessSpanName> = SchemaTelemetrySpan<
	typeof HARNESS_TELEMETRY_SCHEMA,
	Name
>;
export type HarnessSpan = TelemetrySchemaSpanUnion<typeof HARNESS_TELEMETRY_SCHEMA>;

export function startHarnessSpan<
	Name extends HarnessSpanName,
	const Attributes extends HarnessSpanStartAttributes<Name>,
	Result,
>(
	telemetryContext: TelemetryContext,
	name: Name,
	attributes: ExactTelemetryAttributes<HarnessSpanStartAttributes<Name>, Attributes>,
	callback: (span: HarnessTelemetrySpan<Name>) => Result | Promise<Result>,
): Promise<Result> {
	return telemetryContext.startSpan({ name, attributes }, (span: TelemetrySpan) =>
		callback(span as HarnessTelemetrySpan<Name>),
	);
}
