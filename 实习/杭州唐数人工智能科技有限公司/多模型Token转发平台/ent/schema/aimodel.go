package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AIModel maps a display name (visible to callers) to the upstream model name
// used in requests to the provider, along with per-token pricing.
type AIModel struct {
	ent.Schema
}

func (AIModel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		// name is the public identifier callers use in their requests.
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Unique(),
		// upstream_name is the actual model id sent to the provider.
		field.String("upstream_name").
			MaxLen(100).
			NotEmpty(),
		field.Enum("type").
			Values("chat", "embedding", "image").
			Default("chat"),
		// input_price and output_price are in units per 1 000 tokens (micro-currency
		// integers to avoid float precision issues).
		field.Int64("input_price").
			Default(0),
		field.Int64("output_price").
			Default(0),
		field.Bool("is_enabled").
			Default(true),
	}
}

func (AIModel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
		index.Fields("is_enabled"),
	}
}

func (AIModel) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
