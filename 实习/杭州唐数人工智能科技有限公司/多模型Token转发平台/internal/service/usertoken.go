package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/encryption"
	"github.com/school-api/school-api-v1/internal/repository"
)

// CreateUserTokenInput is the service-level input for creating a token.
type CreateUserTokenInput struct {
	UserID     uuid.UUID
	Name       string
	QuotaLimit *int64
	ExpiresAt  *time.Time
}

// UserTokenResponse is the public representation of an API token.
// The plaintext key is returned only once during creation and is never stored.
type UserTokenResponse struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	TokenLast4 string     `json:"token_last4"`
	QuotaLimit *int64     `json:"quota_limit,omitempty"`
	QuotaUsed  int64      `json:"quota_used"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	IsEnabled  bool       `json:"is_enabled"`
	CreatedAt  string     `json:"created_at"`
	UpdatedAt  string     `json:"updated_at"`
}

// UserTokenRepository defines the persistence operations needed by UserTokenService.
type UserTokenRepository interface {
	Create(ctx context.Context, input repository.CreateUserTokenInput) (*ent.UserToken, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.UserToken, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.UserToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, enabled bool) (*ent.UserToken, error)
}

// UserTokenService handles API token management business logic.
type UserTokenService struct {
	tokenRepo  UserTokenRepository
	userRepo   UserRepository
	cipherKey  []byte
	settings   DefaultQuotaProvider
}

// NewUserTokenService creates a new UserTokenService. cipherKey is the
// AES-256 key used to encrypt plaintext tokens for later reveal; if nil,
// tokens are created without a retrievable ciphertext (reveal returns 409).
// settings may be nil, in which case no default quota is applied to new
// tokens that were created without an explicit quota limit.
func NewUserTokenService(tokenRepo UserTokenRepository, userRepo UserRepository, cipherKey []byte, settings DefaultQuotaProvider) *UserTokenService {
	return &UserTokenService{tokenRepo: tokenRepo, userRepo: userRepo, cipherKey: cipherKey, settings: settings}
}

// CreateUserToken creates a new API token for the given user. It returns the
// stored token metadata and the one-time plaintext key.
func (s *UserTokenService) CreateUserToken(ctx context.Context, input CreateUserTokenInput) (*UserTokenResponse, string, error) {
	if _, err := s.userRepo.GetByID(ctx, input.UserID); err != nil {
		if ent.IsNotFound(err) {
			return nil, "", domain.ErrNotFound
		}
		return nil, "", domain.WrapInternal(err)
	}

	plaintext, hash, err := auth.GenerateAPIKey()
	if err != nil {
		return nil, "", domain.WrapInternal(err)
	}

	// Store an encrypted copy so the portal can reveal the key later after a
	// password confirmation. Failure to encrypt must not fail token creation.
	var cipher *string
	if s.cipherKey != nil {
		if enc, err := encryption.Encrypt(s.cipherKey, plaintext); err == nil {
			cipher = &enc
		}
	}

	// Fall back to the configured default quota when the admin did not
	// specify one explicitly.
	quotaLimit := input.QuotaLimit
	if quotaLimit == nil && s.settings != nil {
		defaultLimit, err := s.settings.DefaultQuotaLimit(ctx)
		if err != nil {
			return nil, "", err
		}
		quotaLimit = defaultLimit
	}

	tok, err := s.tokenRepo.Create(ctx, repository.CreateUserTokenInput{
		UserID:       input.UserID,
		Name:         input.Name,
		TokenHash:    hash,
		TokenLast4:   plaintext[len(plaintext)-4:],
		APIKeyCipher: cipher,
		QuotaLimit:   quotaLimit,
		ExpiresAt:    input.ExpiresAt,
	})
	if err != nil {
		return nil, "", domain.WrapInternal(err)
	}

	return toUserTokenResponse(tok), plaintext, nil
}

// ListUserTokens returns all tokens for a user.
func (s *UserTokenService) ListUserTokens(ctx context.Context, userID uuid.UUID) ([]UserTokenResponse, error) {
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	toks, err := s.tokenRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]UserTokenResponse, 0, len(toks))
	for _, tok := range toks {
		resp = append(resp, *toUserTokenResponse(tok))
	}
	return resp, nil
}

// DeleteUserToken permanently deletes a token after verifying it belongs to the user.
func (s *UserTokenService) DeleteUserToken(ctx context.Context, userID, tokenID uuid.UUID) error {
	tok, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.WrapInternal(err)
	}
	if tok.UserID != userID {
		return domain.ErrNotFound
	}

	if err := s.tokenRepo.Delete(ctx, tokenID); err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// UpdateUserTokenStatus enables or disables a token.
func (s *UserTokenService) UpdateUserTokenStatus(ctx context.Context, userID, tokenID uuid.UUID, enabled bool) (*UserTokenResponse, error) {
	tok, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	if tok.UserID != userID {
		return nil, domain.ErrNotFound
	}

	updated, err := s.tokenRepo.UpdateStatus(ctx, tokenID, enabled)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return toUserTokenResponse(updated), nil
}

func toUserTokenResponse(tok *ent.UserToken) *UserTokenResponse {
	resp := &UserTokenResponse{
		ID:         tok.ID.String(),
		UserID:     tok.UserID.String(),
		Name:       tok.Name,
		TokenLast4: tok.TokenLast4,
		QuotaUsed:  tok.QuotaUsed,
		IsEnabled:  tok.IsEnabled,
		CreatedAt:  tok.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  tok.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if tok.QuotaLimit != nil {
		resp.QuotaLimit = tok.QuotaLimit
	}
	if tok.ExpiresAt != nil && !tok.ExpiresAt.IsZero() {
		resp.ExpiresAt = tok.ExpiresAt
	}
	return resp
}

// Compile-time interface check.
var _ repository.UserTokenRepository = (*repository.EntUserTokenRepository)(nil)
