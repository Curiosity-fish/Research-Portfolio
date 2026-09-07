package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/usertoken"
	"github.com/school-api/school-api-v1/internal/auth"
)

// EntAPIKeyLookup implements auth.APIKeyLookup using the generated Ent client.
type EntAPIKeyLookup struct {
	client *ent.Client
}

// NewEntAPIKeyLookup creates a new Ent-backed API key lookup.
func NewEntAPIKeyLookup(client *ent.Client) *EntAPIKeyLookup {
	return &EntAPIKeyLookup{client: client}
}

// GetByTokenHash looks up a user token by hash and returns the identity data
// needed to validate the key.
func (r *EntAPIKeyLookup) GetByTokenHash(ctx context.Context, hash string) (*auth.APIKeyLookupResult, error) {
	tok, err := r.client.UserToken.Query().
		Where(usertoken.TokenHash(hash)).
		WithUser(func(q *ent.UserQuery) {
			q.WithGroup()
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("query token: %w", err)
	}

	u := tok.Edges.User
	if u == nil {
		return nil, fmt.Errorf("token user not loaded")
	}

	result := &auth.APIKeyLookupResult{
		UserID:     u.ID,
		TokenID:    tok.ID,
		IsEnabled:  tok.IsEnabled,
		ExpiresAt:  nullableTime(tok.ExpiresAt),
		UserStatus: string(u.Status),
	}

	if g := u.Edges.Group; g != nil {
		result.GroupID = g.ID
		result.GroupCode = g.Code
	}

	return result, nil
}

func nullableTime(t *time.Time) *time.Time {
	if t == nil || t.IsZero() {
		return nil
	}
	return t
}

// Compile-time interface check.
var _ auth.APIKeyLookup = (*EntAPIKeyLookup)(nil)
