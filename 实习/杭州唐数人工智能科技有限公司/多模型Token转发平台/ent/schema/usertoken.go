package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserToken holds the schema definition for user-created API keys.
// A user may have multiple tokens, each with its own quota and expiry.
type UserToken struct {
	ent.Schema
}

// Fields of the UserToken.
func (UserToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("token_hash").
			MaxLen(255).
			NotEmpty().
			Unique(),
		field.String("token_last4").
			MaxLen(10).
			NotEmpty(),
		// api_key_cipher is the AES-256-GCM encrypted plaintext key. It exists
		// only for tokens created after this column was introduced; tokens
		// without it cannot be revealed via the user portal (409). Nullable
		// so legacy rows stay valid.
		field.String("api_key_cipher").
			MaxLen(512).
			Optional().
			Nillable(),
		field.Int64("quota_limit").
			Optional().
			Nillable(),
		field.Int64("quota_used").
			Default(0),
		field.Time("expires_at").
			Optional().
			Nillable(),
		field.Bool("is_enabled").
			Default(true),
	}
}

// Edges of the UserToken.
func (UserToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes of the UserToken.
func (UserToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token_hash").
			Unique(),
		index.Fields("user_id"),
		index.Fields("expires_at"),
		index.Fields("is_enabled"),
	}
}

// Mixin of the UserToken.
func (UserToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
