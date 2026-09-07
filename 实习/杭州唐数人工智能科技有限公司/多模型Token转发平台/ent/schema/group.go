package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Group holds the schema definition for a user group used by the AI relay
// routing layer (Group -> Platform -> Account pool).
type Group struct {
	ent.Schema
}

// Fields of the Group.
func (Group) Fields() []ent.Field {
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
		field.String("description").
			MaxLen(500).
			Optional().
			Nillable().
			Default(""),
		field.Int("sort_order").
			Default(0),
		field.Enum("status").
			Values("active", "inactive").
			Default("active"),
	}
}

// Edges of the Group.
func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).
			Ref("group"),
		edge.To("group_platforms", GroupPlatform.Type),
	}
}

// Indexes of the Group.
func (Group) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").
			Unique(),
		index.Fields("status"),
	}
}

// Mixin of the Group.
func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
