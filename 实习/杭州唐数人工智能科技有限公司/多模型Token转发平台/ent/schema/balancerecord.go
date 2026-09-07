package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BalanceRecord records every change to a user's balance.
// It is append-only; the balance_after field is the authoritative snapshot.
type BalanceRecord struct {
	ent.Schema
}

func (BalanceRecord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		// call_log_id links the balance change to a relay call when applicable.
		field.UUID("call_log_id", uuid.UUID{}).
			Optional().
			Nillable(),
		// related_order_id links the record to a recharge or refund order.
		field.UUID("related_order_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Enum("type").
			Values("recharge", "consume", "refund", "admin_adjust"),
		// amount is the signed change. Positive for balance increase, negative for decrease.
		field.Int64("amount"),
		// balance_after is the user's balance after this record is applied.
		field.Int64("balance_after").
			Default(0),
		// remark carries an optional human-readable note.
		field.String("remark").
			MaxLen(500).
			Optional().
			Nillable(),
	}
}

func (BalanceRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.To("call_log", CallLog.Type).
			Field("call_log_id").
			Unique(),
		edge.To("related_order", RechargeOrder.Type).
			Field("related_order_id").
			Unique(),
	}
}

func (BalanceRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("call_log_id"),
		index.Fields("related_order_id"),
		index.Fields("created_at"),
	}
}

func (BalanceRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
