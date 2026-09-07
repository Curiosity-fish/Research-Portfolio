package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AdminUser holds the schema definition for the AdminUser entity.
type AdminUser struct {
	ent.Schema
}

// Fields of the AdminUser.
func (AdminUser) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("username").
			MaxLen(50).
			NotEmpty().
			Unique(),
		field.String("password_hash").
			MaxLen(255).
			NotEmpty(),
		field.Enum("role").
			Values("super_admin", "admin").
			Default("admin"),
		field.Enum("status").
			Values("active", "inactive").
			Default("active"),
		field.Time("last_login_at").
			Optional().
			Nillable(),
	}
}

// Indexes of the AdminUser.
func (AdminUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").
			Unique(),
		index.Fields("status"),
	}
}

// Mixin of the AdminUser.
func (AdminUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
