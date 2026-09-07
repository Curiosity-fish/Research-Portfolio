package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertRecord stores a single triggered alert produced by evaluating AlertRules.
type AlertRecord struct {
	ent.Schema
}

func (AlertRecord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("rule_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("metric").
			MaxLen(50),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("token_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("account_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int64("triggered_value").
			Default(0),
		field.String("message").
			MaxLen(500),
		field.Bool("is_resolved").
			Default(false),
		field.Time("resolved_at").
			Optional().
			Nillable(),
		field.UUID("resolved_by", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (AlertRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("rule", AlertRule.Type).
			Field("rule_id").
			Unique(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique(),
		edge.To("token", UserToken.Type).
			Field("token_id").
			Unique(),
		edge.To("account", Account.Type).
			Field("account_id").
			Unique(),
		edge.To("resolver", AdminUser.Type).
			Field("resolved_by").
			Unique(),
	}
}

func (AlertRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("rule_id"),
		index.Fields("metric"),
		index.Fields("user_id"),
		index.Fields("is_resolved"),
		index.Fields("created_at"),
	}
}

func (AlertRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
