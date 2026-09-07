package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// QuotaRequest allows a user to ask for more quota on a specific token.
// Admins can approve or reject the request; approval increases the token's quota_limit.
type QuotaRequest struct {
	ent.Schema
}

func (QuotaRequest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("token_id", uuid.UUID{}),
		// requested_amount is the additional quota (in tokens) requested.
		field.Int64("requested_amount").
			Positive(),
		field.String("reason").
			MaxLen(500).
			Optional().
			Nillable(),
		field.Enum("status").
			Values("pending", "approved", "rejected").
			Default("pending"),
		// reviewed_by identifies the admin who processed the request.
		field.UUID("reviewed_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("reviewed_at").
			Optional().
			Nillable(),
	}
}

func (QuotaRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.To("token", UserToken.Type).
			Field("token_id").
			Unique().
			Required(),
		edge.To("reviewer", AdminUser.Type).
			Field("reviewed_by").
			Unique(),
	}
}

func (QuotaRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("token_id"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}

func (QuotaRequest) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
