package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("department_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("group_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("username").
			MaxLen(50).
			NotEmpty().
			Unique(),
		field.String("password_hash").
			MaxLen(255).
			NotEmpty(),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("email").
			MaxLen(255).
			Optional().
			Nillable().
			Unique(),
		field.String("phone").
			MaxLen(20).
			Optional().
			Nillable().
			Unique(),
		field.Enum("role").
			Values("student", "teacher", "staff").
			Default("student"),
		field.String("gender").
			MaxLen(10).
			Optional().
			Nillable(),
		field.Enum("status").
			Values("active", "inactive", "banned").
			Default("active"),
		// balance is the user's account balance in micro-currency.
		field.Int64("balance").
			Default(0),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("department", Department.Type).
			Field("department_id").
			Unique(),
		edge.To("group", Group.Type).
			Field("group_id").
			Unique(),
		edge.From("tokens", UserToken.Type).
			Ref("user"),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").
			Unique(),
		index.Fields("email").
			Unique(),
		index.Fields("phone").
			Unique(),
		index.Fields("department_id"),
		index.Fields("group_id"),
		index.Fields("status"),
	}
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
