package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Setting holds the schema definition for the Setting entity.
type Setting struct {
	ent.Schema
}

// Fields of the Setting.
func (Setting) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("key").
			MaxLen(128).
			NotEmpty().
			Unique(),
		field.Text("value").
			NotEmpty(),
		field.Enum("type").
			Values("string", "int", "bool", "json").
			Default("string"),
		field.String("description").
			MaxLen(255).
			Default(""),
		field.Bool("is_public").
			Default(false),
	}
}

// Indexes of the Setting.
func (Setting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key").
			Unique(),
	}
}

// Mixin of the Setting.
func (Setting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
