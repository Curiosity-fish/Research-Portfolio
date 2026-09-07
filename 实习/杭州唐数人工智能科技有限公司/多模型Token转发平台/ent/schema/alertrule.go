package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertRule defines thresholds used to detect operational anomalies.
type AlertRule struct {
	ent.Schema
}

func (AlertRule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			MaxLen(200),
		field.Enum("metric").
			Values("balance_low", "quota_low", "error_rate", "cost_spike"),
		field.Int64("threshold").
			Positive(),
		field.Bool("enabled").
			Default(true),
		field.String("description").
			MaxLen(500).
			Optional().
			Nillable(),
	}
}

func (AlertRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("metric"),
		index.Fields("enabled"),
		index.Fields("created_at"),
	}
}

func (AlertRule) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
