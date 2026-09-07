package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/encryption"
	"github.com/school-api/school-api-v1/internal/repository"
)

// errInvalidCredentials is the only login/identity error exposed to clients.
// A single message prevents username enumeration through error text.
var errInvalidCredentials = domain.NewAppError(401, "INVALID_CREDENTIALS", "账号或密码错误")

// UserLoginResult is the response of a successful user login.
type UserLoginResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// UserAuthService handles user-portal account flows: password login, password
// change and API token reveal. Security properties mirror the admin login:
// unified error text, constant-cost failure path, and password confirmation
// before any secret material is exposed.
type UserAuthService struct {
	userRepo   UserRepository
	tokenRepo  UserTokenRepository
	tm         auth.TokenManager
	bcryptCost int
	cipherKey  []byte
	expiry     int64 // seconds, reported to clients as expires_in
}

// NewUserAuthService creates a new UserAuthService.
func NewUserAuthService(userRepo UserRepository, tokenRepo UserTokenRepository, tm auth.TokenManager, bcryptCost int, cipherKey []byte, expirySeconds int64) *UserAuthService {
	return &UserAuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		tm:         tm,
		bcryptCost: bcryptCost,
		cipherKey:  cipherKey,
		expiry:     expirySeconds,
	}
}

// Login validates username+password and returns a user-portal JWT.
// The failure path always runs one bcrypt comparison (against dummyHash) so
// response timing does not reveal whether the username exists.
func (s *UserAuthService) Login(ctx context.Context, username, password string) (UserLoginResult, error) {
	u, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if ent.IsNotFound(err) {
			// Constant-cost path: compare against a hash that never matches.
			_ = auth.ComparePassword(dummyHash, password)
			return UserLoginResult{}, errInvalidCredentials
		}
		return UserLoginResult{}, domain.WrapInternal(err)
	}

	if !auth.ComparePassword(u.PasswordHash, password) {
		return UserLoginResult{}, errInvalidCredentials
	}

	if u.Status != user.StatusActive {
		return UserLoginResult{}, errInvalidCredentials
	}

	token, err := s.tm.GenerateUser(ctx, u)
	if err != nil {
		return UserLoginResult{}, domain.WrapInternal(err)
	}

	return UserLoginResult{
		AccessToken: token,
		TokenType:   auth.TokenType,
		ExpiresIn:   s.expiry,
	}, nil
}

// ChangePassword verifies the old password and replaces it with the new one.
func (s *UserAuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.WrapInternal(err)
	}

	if !auth.ComparePassword(u.PasswordHash, oldPassword) {
		return errInvalidCredentials
	}

	hash, err := auth.HashPassword(s.bcryptCost, newPassword)
	if err != nil {
		return domain.WrapInternal(err)
	}

	if _, err := s.userRepo.Update(ctx, userID, repository.UpdateUserInput{PasswordHash: &hash}); err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// RevealToken returns the plaintext of one of the user's API tokens after a
// password confirmation. Tokens created before ciphertext storage was added
// return 409 TOKEN_REVEAL_UNAVAILABLE.
func (s *UserAuthService) RevealToken(ctx context.Context, userID, tokenID uuid.UUID, password string) (string, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", domain.ErrNotFound
		}
		return "", domain.WrapInternal(err)
	}

	if !auth.ComparePassword(u.PasswordHash, password) {
		return "", errInvalidCredentials
	}

	tok, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", domain.ErrNotFound
		}
		return "", domain.WrapInternal(err)
	}
	if tok.UserID != userID {
		// Do not leak token existence across users.
		return "", domain.ErrNotFound
	}

	if tok.APIKeyCipher == nil || *tok.APIKeyCipher == "" {
		return "", domain.NewAppError(409, "TOKEN_REVEAL_UNAVAILABLE", "该 Token 无法恢复，请删除后重新创建")
	}

	plaintext, err := encryption.Decrypt(s.cipherKey, *tok.APIKeyCipher)
	if err != nil {
		// Corrupt ciphertext (e.g. key rotation) is a business conflict,
		// not a server fault.
		return "", domain.NewAppError(409, "TOKEN_REVEAL_UNAVAILABLE", "该 Token 密文无法解密，请删除后重新创建")
	}

	return plaintext, nil
}
