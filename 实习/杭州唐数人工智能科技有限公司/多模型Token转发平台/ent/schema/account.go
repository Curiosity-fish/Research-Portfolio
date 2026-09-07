package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Account is a set of credentials for one upstream AI provider account.
// api_key_encrypted stores AES-256-GCM(nonce||ciphertext) base64url so the
// plaintext key never appears in the database.
type Account struct {
	ent.Schema
}

func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("platform_id", uuid.UUID{}),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		// api_key_encrypted: AES-256-GCM encrypted, base64url encoded.
		field.String("api_key_encrypted").
			MaxLen(1000).
			NotEmpty(),
		field.Int("weight").
			Default(1),
		// max_rpm is a soft ceiling used by the relay; 0 means unlimited.
		field.Int("max_rpm").
			Default(0),
		field.Enum("status").
			Values("active", "inactive", "error").
			Default("active"),
		// error_count tracks consecutive upstream failures for soft circuit-breaking.
		field.Int("error_count").
			Default(0),
	}
}

func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("platform", Platform.Type).
			Ref("accounts").
			Field("platform_id").
			Unique().
			Required(),
		edge.To("call_logs", CallLog.Type),
	}
}

func (Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform_id"),
		index.Fields("status"),
	}
}

func (Account) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
