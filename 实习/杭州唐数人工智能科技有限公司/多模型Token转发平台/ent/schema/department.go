package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Department holds the schema definition for the Department entity.
type Department struct {
	ent.Schema
}

// Fields of the Department.
func (Department) Fields() []ent.Field {
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
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int("level").
			Default(0),
		field.String("path").
			MaxLen(2000).
			Default(""),
		field.Int("sort_order").
			Default(0),
		field.Enum("status").
			Values("active", "inactive").
			Default("active"),
	}
}

// Edges of the Department.
func (Department) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Department.Type).
			From("parent").
			Field("parent_id").
			Unique(),
		edge.From("users", User.Type).
			Ref("department"),
	}
}

// Indexes of the Department.
func (Department) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").
			Unique(),
		index.Fields("parent_id"),
		index.Fields("path"),
		index.Fields("status"),
	}
}

// Mixin of the Department.
func (Department) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
