package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AuditLog records significant operational actions for traceability.
// It intentionally does not store request bodies to avoid leaking secrets.
type AuditLog struct {
	ent.Schema
}

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("actor_type").
			Values("admin", "user", "system"),
		field.UUID("actor_id", uuid.UUID{}),
		field.String("action").
			MaxLen(100),
		field.String("target_type").
			MaxLen(50).
			Optional().
			Nillable(),
		field.UUID("target_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.JSON("details", map[string]any{}).
			Default(map[string]any{}),
		field.String("ip").
			MaxLen(50).
			Optional().
			Nillable(),
		field.String("user_agent").
			MaxLen(500).
			Optional().
			Nillable(),
	}
}

func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("actor_type"),
		index.Fields("actor_id"),
		index.Fields("action"),
		index.Fields("created_at"),
	}
}

func (AuditLog) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
