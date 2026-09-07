package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

type mockAdminRepo struct {
	getByUsername   func(ctx context.Context, username string) (*ent.AdminUser, error)
	updateLastLogin func(ctx context.Context, id uuid.UUID) error
	createAdmin     func(ctx context.Context, username, passwordHash string, role adminuser.Role) (*ent.AdminUser, error)
}

func (m *mockAdminRepo) GetByUsername(ctx context.Context, username string) (*ent.AdminUser, error) {
	if m.getByUsername != nil {
		return m.getByUsername(ctx, username)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAdminRepo) UpdateLastLoginAt(ctx context.Context, id uuid.UUID) error {
	if m.updateLastLogin != nil {
		return m.updateLastLogin(ctx, id)
	}
	return nil
}

func (m *mockAdminRepo) CreateAdmin(ctx context.Context, username, passwordHash string, role adminuser.Role) (*ent.AdminUser, error) {
	if m.createAdmin != nil {
		return m.createAdmin(ctx, username, passwordHash, role)
	}
	return nil, errors.New("not implemented")
}

type mockTokenManager struct {
	generate func(ctx context.Context, admin *ent.AdminUser) (string, error)
	parse    func(ctx context.Context, tokenStr string) (*auth.Claims, error)
}

func (m *mockTokenManager) Generate(ctx context.Context, admin *ent.AdminUser) (string, error) {
	if m.generate != nil {
		return m.generate(ctx, admin)
	}
	return "", errors.New("not implemented")
}

func (m *mockTokenManager) GenerateUser(ctx context.Context, user *ent.User) (string, error) {
	return "", errors.New("not implemented")
}

func (m *mockTokenManager) Parse(ctx context.Context, tokenStr string) (*auth.Claims, error) {
	if m.parse != nil {
		return m.parse(ctx, tokenStr)
	}
	return nil, errors.New("not implemented")
}

func TestAuthService_Login_Success(t *testing.T) {
	ctx := context.Background()
	password := "valid-password"
	hash, err := auth.HashPassword(auth.DefaultBcryptCost, password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	admin := &ent.AdminUser{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: hash,
		Role:         adminuser.RoleSuperAdmin,
		Status:       adminuser.StatusActive,
	}

	repo := &mockAdminRepo{
		getByUsername: func(ctx context.Context, username string) (*ent.AdminUser, error) {
			return admin, nil
		},
		updateLastLogin: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	tm := &mockTokenManager{
		generate: func(ctx context.Context, a *ent.AdminUser) (string, error) {
			return "test-token", nil
		},
	}

	svc := NewAuthService(repo, tm, auth.DefaultBcryptCost, time.Hour)
	result, err := svc.Login(ctx, AdminLoginInput{Username: "admin", Password: password})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.AccessToken != "test-token" {
		t.Errorf("expected token test-token, got %s", result.AccessToken)
	}
	if result.TokenType != auth.TokenType {
		t.Errorf("expected token type %s, got %s", auth.TokenType, result.TokenType)
	}
	if result.ExpiresIn != 3600 {
		t.Errorf("expected expires_in 3600, got %d", result.ExpiresIn)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	hash, err := auth.HashPassword(auth.DefaultBcryptCost, "valid-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &mockAdminRepo{
		getByUsername: func(ctx context.Context, username string) (*ent.AdminUser, error) {
			return &ent.AdminUser{
				ID:           uuid.New(),
				Username:     "admin",
				PasswordHash: hash,
				Role:         adminuser.RoleAdmin,
				Status:       adminuser.StatusActive,
			}, nil
		},
	}

	svc := NewAuthService(repo, &mockTokenManager{}, auth.DefaultBcryptCost, time.Hour)
	_, err = svc.Login(ctx, AdminLoginInput{Username: "admin", Password: "wrong-password"})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockAdminRepo{
		getByUsername: func(ctx context.Context, username string) (*ent.AdminUser, error) {
			return nil, &ent.NotFoundError{}
		},
	}

	svc := NewAuthService(repo, &mockTokenManager{}, auth.DefaultBcryptCost, time.Hour)
	_, err := svc.Login(ctx, AdminLoginInput{Username: "missing", Password: "any"})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthService_Login_InactiveAccount(t *testing.T) {
	ctx := context.Background()
	password := "valid-password"
	hash, err := auth.HashPassword(auth.DefaultBcryptCost, password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &mockAdminRepo{
		getByUsername: func(ctx context.Context, username string) (*ent.AdminUser, error) {
			return &ent.AdminUser{
				ID:           uuid.New(),
				Username:     "admin",
				PasswordHash: hash,
				Role:         adminuser.RoleAdmin,
				Status:       adminuser.StatusInactive,
			}, nil
		},
	}

	svc := NewAuthService(repo, &mockTokenManager{}, auth.DefaultBcryptCost, time.Hour)
	_, err = svc.Login(ctx, AdminLoginInput{Username: "admin", Password: password})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthService_Login_TokenGenerationFailure(t *testing.T) {
	ctx := context.Background()
	password := "valid-password"
	hash, err := auth.HashPassword(auth.DefaultBcryptCost, password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &mockAdminRepo{
		getByUsername: func(ctx context.Context, username string) (*ent.AdminUser, error) {
			return &ent.AdminUser{
				ID:           uuid.New(),
				Username:     "admin",
				PasswordHash: hash,
				Role:         adminuser.RoleAdmin,
				Status:       adminuser.StatusActive,
			}, nil
		},
		updateLastLogin: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	tm := &mockTokenManager{
		generate: func(ctx context.Context, a *ent.AdminUser) (string, error) {
			return "", errors.New("token failure")
		},
	}

	svc := NewAuthService(repo, tm, auth.DefaultBcryptCost, time.Hour)
	_, err = svc.Login(ctx, AdminLoginInput{Username: "admin", Password: password})
	if err == nil {
		t.Fatal("expected error when token generation fails")
	}
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrInternal.Code {
		t.Fatalf("expected internal error, got %v", err)
	}
}

// Compile-time check that mockAdminRepo satisfies the interface.
var _ repository.AdminRepository = (*mockAdminRepo)(nil)
