package service

import (
	"context"
	"time"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// dummyHash is a valid bcrypt hash used to keep failed login attempts
// timing-equivalent regardless of whether the username exists. It must never
// match a real user password.
const dummyHash = "$2a$10$aybCWfZVR9yo2Wk2OuoZFuM4.G9Bly0Cr9YBLOmi9ARR1c9s9xBNK"

// AdminLoginInput is the service-level input for admin login.
type AdminLoginInput struct {
	Username string
	Password string
}

// AdminLoginResult is the service-level output for a successful login.
type AdminLoginResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// AuthService handles administrator authentication.
type AuthService struct {
	repo         repository.AdminRepository
	tokenManager auth.TokenManager
	bcryptCost   int
	tokenExpiry  time.Duration
}

// NewAuthService creates an AuthService.
func NewAuthService(repo repository.AdminRepository, tm auth.TokenManager, bcryptCost int, tokenExpiry time.Duration) *AuthService {
	if bcryptCost < auth.MinBcryptCost {
		bcryptCost = auth.DefaultBcryptCost
	}
	return &AuthService{
		repo:         repo,
		tokenManager: tm,
		bcryptCost:   bcryptCost,
		tokenExpiry:  tokenExpiry,
	}
}

// Login validates the credentials and returns a JWT.
// It always returns ErrUnauthorized for invalid username, password, or inactive
// account to avoid information leakage. Failed attempts perform the same
// bcrypt comparison so response timing does not reveal whether a username
// exists or is active.
func (s *AuthService) Login(ctx context.Context, input AdminLoginInput) (AdminLoginResult, error) {
	var result AdminLoginResult

	admin, err := s.repo.GetByUsername(ctx, input.Username)
	if err != nil {
		if ent.IsNotFound(err) {
			// Run a dummy comparison so this path is timing-equivalent to a
			// real password check.
			_ = auth.ComparePassword(dummyHash, input.Password)
			return result, domain.ErrUnauthorized
		}
		return result, domain.WrapInternal(err)
	}

	if admin.Status != adminuser.StatusActive {
		_ = auth.ComparePassword(admin.PasswordHash, input.Password)
		return result, domain.ErrUnauthorized
	}

	if !auth.ComparePassword(admin.PasswordHash, input.Password) {
		return result, domain.ErrUnauthorized
	}

	tokenStr, err := s.tokenManager.Generate(ctx, admin)
	if err != nil {
		return result, domain.WrapInternal(err)
	}

	// Update last_login only after a token has been successfully issued so
	// audit state matches the outcome seen by the client.
	if err := s.repo.UpdateLastLoginAt(ctx, admin.ID); err != nil {
		return result, domain.WrapInternal(err)
	}

	return AdminLoginResult{
		AccessToken: tokenStr,
		TokenType:   auth.TokenType,
		ExpiresIn:   int64(s.tokenExpiry.Seconds()),
	}, nil
}
