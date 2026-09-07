package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CallLog records every upstream relay attempt. It is append-only and never
// updated. Only token counts and the outcome are stored; request/response
// bodies are deliberately excluded.
type CallLog struct {
	ent.Schema
}

func (CallLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("token_id", uuid.UUID{}),
		field.UUID("platform_id", uuid.UUID{}),
		field.UUID("account_id", uuid.UUID{}),
		field.String("model").
			MaxLen(100),
		field.Int64("prompt_tokens").
			Default(0),
		field.Int64("completion_tokens").
			Default(0),
		field.Int64("total_tokens").
			Default(0),
		// latency_ms is end-to-end relay latency in milliseconds.
		field.Int64("latency_ms").
			Default(0),
		field.Int("status_code").
			Default(0),
		field.String("error_msg").
			MaxLen(500).
			Optional().
			Nillable(),
	}
}

func (CallLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).
			Ref("call_logs").
			Field("account_id").
			Unique().
			Required(),
	}
}

func (CallLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("token_id"),
		index.Fields("account_id"),
		index.Fields("platform_id"),
		index.Fields("created_at"),
	}
}

func (CallLog) Mixin() []ent.Mixin {
	// CallLog is immutable; updated_at is not meaningful here but TimeMixin
	// is kept for consistency. The created_at field is the only timestamp used.
	return []ent.Mixin{TimeMixin{}}
}
