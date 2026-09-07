package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GuideSection stores a section of the platform usage guide rendered by the
// user portal or the admin panel, depending on the audience.
type GuideSection struct {
	ent.Schema
}

func (GuideSection) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			MaxLen(255).
			NotEmpty(),
		field.Text("content_md").
			NotEmpty(),
		field.Enum("audience").
			Values("user", "admin").
			Default("user"),
		field.Int("sort_order").
			Default(0),
		field.Bool("is_enabled").
			Default(true),
	}
}

func (GuideSection) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("audience"),
		index.Fields("is_enabled"),
		index.Fields("sort_order"),
	}
}

func (GuideSection) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
