package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/usertoken"
)

// CreateUserTokenInput is the data required to create a user token.
type CreateUserTokenInput struct {
	UserID        uuid.UUID
	Name          string
	TokenHash     string
	TokenLast4    string
	APIKeyCipher  *string
	QuotaLimit    *int64
	ExpiresAt     *time.Time
}

// UserTokenRepository provides data access for user_tokens.
type UserTokenRepository interface {
	Create(ctx context.Context, input CreateUserTokenInput) (*ent.UserToken, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.UserToken, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.UserToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, enabled bool) (*ent.UserToken, error)
	AddQuotaLimit(ctx context.Context, id uuid.UUID, amount int64) (*ent.UserToken, error)
}

// EntUserTokenRepository implements UserTokenRepository using Ent.
type EntUserTokenRepository struct {
	client *ent.Client
}

// NewEntUserTokenRepository creates a new Ent-backed user token repository.
func NewEntUserTokenRepository(client *ent.Client) *EntUserTokenRepository {
	return &EntUserTokenRepository{client: client}
}

// Create inserts a new token for a user.
func (r *EntUserTokenRepository) Create(ctx context.Context, input CreateUserTokenInput) (*ent.UserToken, error) {
	b := r.client.UserToken.Create().
		SetUserID(input.UserID).
		SetName(input.Name).
		SetTokenHash(input.TokenHash).
		SetTokenLast4(input.TokenLast4)

	if input.APIKeyCipher != nil {
		b.SetAPIKeyCipher(*input.APIKeyCipher)
	}
	if input.QuotaLimit != nil {
		b.SetQuotaLimit(*input.QuotaLimit)
	}
	if input.ExpiresAt != nil {
		b.SetExpiresAt(*input.ExpiresAt)
	}

	tok, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create user token: %w", err)
	}
	return tok, nil
}

// ListByUserID returns all tokens for a user, newest first.
func (r *EntUserTokenRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.UserToken, error) {
	toks, err := r.client.UserToken.Query().
		Where(usertoken.UserIDEQ(userID)).
		Order(ent.Desc(usertoken.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list user tokens: %w", err)
	}
	return toks, nil
}

// GetByID returns a token by ID.
func (r *EntUserTokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.UserToken, error) {
	tok, err := r.client.UserToken.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user token: %w", err)
	}
	return tok, nil
}

// Delete removes a token permanently.
func (r *EntUserTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.client.UserToken.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete user token: %w", err)
	}
	return nil
}

// UpdateStatus enables or disables a token.
func (r *EntUserTokenRepository) UpdateStatus(ctx context.Context, id uuid.UUID, enabled bool) (*ent.UserToken, error) {
	tok, err := r.client.UserToken.UpdateOneID(id).
		SetIsEnabled(enabled).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update user token status: %w", err)
	}
	return tok, nil
}

// AddQuotaLimit atomically increases a token's quota limit by the given amount.
func (r *EntUserTokenRepository) AddQuotaLimit(ctx context.Context, id uuid.UUID, amount int64) (*ent.UserToken, error) {
	tok, err := r.client.UserToken.UpdateOneID(id).
		AddQuotaLimit(amount).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("add user token quota limit: %w", err)
	}
	return tok, nil
}
