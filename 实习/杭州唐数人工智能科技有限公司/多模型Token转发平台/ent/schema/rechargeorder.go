package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RechargeOrder represents a user's request to add balance to their account.
// In Phase 5 only a mock provider is supported.
type RechargeOrder struct {
	ent.Schema
}

func (RechargeOrder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		// amount is the recharge amount in micro-currency.
		field.Int64("amount").
			Positive(),
		field.Enum("status").
			Values("pending", "paid", "failed", "cancelled").
			Default("pending"),
		field.Enum("provider").
			Values("mock").
			Default("mock"),
		// provider_order_id is the identifier returned by the payment provider.
		field.String("provider_order_id").
			MaxLen(255).
			Optional().
			Nillable(),
		field.Time("paid_at").
			Optional().
			Nillable(),
		// refunded_amount accumulates the total amount already refunded for this order.
		field.Int64("refunded_amount").
			Default(0),
	}
}

func (RechargeOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

func (RechargeOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("provider_order_id"),
	}
}

func (RechargeOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
