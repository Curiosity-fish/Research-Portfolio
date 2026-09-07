package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// nowUTC returns the current time in UTC. It is used as the default for the
// timestamp fields so the database and logs have a consistent time zone
// regardless of the host configuration.
func nowUTC() time.Time {
	return time.Now().UTC()
}

// TimeMixin adds created_at and updated_at timestamps to an entity.
// It mirrors the naming used in the legacy schema instead of ent's default
// create_time / update_time, and stores values in UTC.
type TimeMixin struct {
	mixin.Schema
}

// Fields of the TimeMixin.
func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}
