package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Platform represents an upstream AI provider (OpenAI, Anthropic, etc.).
type Platform struct {
	ent.Schema
}

func (Platform) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("code").
			MaxLen(50).
			NotEmpty().
			Unique(),
		// type distinguishes which protocol adapter to use during relay.
		field.Enum("type").
			Values("openai", "anthropic").
			Default("openai"),
		field.String("base_url").
			MaxLen(500).
			NotEmpty(),
		field.Enum("status").
			Values("active", "inactive").
			Default("active"),
	}
}

func (Platform) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("accounts", Account.Type),
		edge.To("group_platforms", GroupPlatform.Type),
	}
}

func (Platform) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("status"),
	}
}

func (Platform) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
