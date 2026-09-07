package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// QuotaRecord records every change to a token's quota usage.
// It is append-only; the type field explains the direction of the change.
type QuotaRecord struct {
	ent.Schema
}

func (QuotaRecord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("token_id", uuid.UUID{}),
		// call_log_id links the quota change to a relay call when applicable.
		field.UUID("call_log_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Enum("type").
			Values("pre_deduct", "consume", "refund", "admin_adjust", "request_approved"),
		// amount is always a positive number; the type indicates the direction.
		field.Int64("amount").
			Positive(),
		// quota_after is the token's quota_used after this record is applied.
		field.Int64("quota_after").
			Default(0),
	}
}

func (QuotaRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.To("token", UserToken.Type).
			Field("token_id").
			Unique().
			Required(),
		edge.To("call_log", CallLog.Type).
			Field("call_log_id").
			Unique(),
	}
}

func (QuotaRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("token_id"),
		index.Fields("call_log_id"),
		index.Fields("created_at"),
	}
}

func (QuotaRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
