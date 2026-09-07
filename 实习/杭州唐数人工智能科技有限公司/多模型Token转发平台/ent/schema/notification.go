package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Notification stores system announcements and user-specific notifications.
// A nil user_id represents a broadcast announcement visible to all users.
type Notification struct {
	ent.Schema
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		// user_id is nil for broadcast announcements.
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Enum("type").
			Values("announcement", "system").
			Default("announcement"),
		field.String("title").
			MaxLen(200),
		field.String("content").
			MaxLen(5000),
		field.Bool("is_read").
			Default(false),
	}
}

func (Notification) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("type"),
		index.Fields("is_read"),
		index.Fields("created_at"),
	}
}

func (Notification) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
