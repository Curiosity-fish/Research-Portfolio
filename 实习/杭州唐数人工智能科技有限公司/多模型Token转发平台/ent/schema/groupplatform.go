package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GroupPlatform is the join table for the Group <-> Platform many-to-many
// relationship. No extra payload fields are needed at this stage.
type GroupPlatform struct {
	ent.Schema
}

func (GroupPlatform) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("group_id", uuid.UUID{}),
		field.UUID("platform_id", uuid.UUID{}),
	}
}

func (GroupPlatform) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("group_platforms").
			Field("group_id").
			Unique().
			Required(),
		edge.From("platform", Platform.Type).
			Ref("group_platforms").
			Field("platform_id").
			Unique().
			Required(),
	}
}

func (GroupPlatform) Indexes() []ent.Index {
	return []ent.Index{
		// Composite unique: a group can only be linked to a platform once.
		index.Fields("group_id", "platform_id").Unique(),
	}
}

func (GroupPlatform) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
