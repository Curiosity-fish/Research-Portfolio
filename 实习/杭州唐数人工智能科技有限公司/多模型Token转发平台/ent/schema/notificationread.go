package schema

import (
	"time"

	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationRead records that a user has read a broadcast notification.
// Read state is tracked per user so one user marking a broadcast as read
// does not affect other users.
type NotificationRead struct {
	ent.Schema
}

func (NotificationRead) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("notification_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Time("read_at").
			Default(time.Now).
			Immutable(),
	}
}

func (NotificationRead) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("notification", Notification.Type).
			Field("notification_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (NotificationRead) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("notification_id", "user_id").Unique(),
		index.Fields("user_id"),
	}
}
