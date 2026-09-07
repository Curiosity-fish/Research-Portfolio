package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RefundRequest is a user-initiated request to refund part or all of a paid
// recharge order. Approval goes through the existing reconciliation refund
// transaction, so the money path stays single-writer.
type RefundRequest struct {
	ent.Schema
}

func (RefundRequest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("order_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		// amount is the requested refund in micro-currency.
		field.Int64("amount").
			Positive(),
		field.String("reason").
			MaxLen(500).
			Optional().
			Nillable(),
		field.Enum("status").
			Values("pending", "approved", "rejected").
			Default("pending"),
		field.UUID("reviewed_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("reviewed_at").
			Optional().
			Nillable(),
	}
}

func (RefundRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("order", RechargeOrder.Type).
			Field("order_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

func (RefundRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("user_id"),
		index.Fields("status"),
	}
}

func (RefundRequest) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
