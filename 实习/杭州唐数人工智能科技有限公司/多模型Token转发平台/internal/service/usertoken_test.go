package service

import (
	"context"
	"testing"
	"time"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

func newUserTokenService(t *testing.T, client *ent.Client) *UserTokenService {
	t.Helper()
	return NewUserTokenService(
		repository.NewEntUserTokenRepository(client),
		repository.NewEntUserRepository(client),
		testCipherKey,
		nil,
	)
}

// testCipherKey is a fixed 32-byte AES key for encrypting token plaintext in tests.
var testCipherKey = []byte("0123456789abcdef0123456789abcdef")

func TestUserTokenService_CreateAndList(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserTokenService(t, client)

	u, err := repository.NewEntUserRepository(client).Create(ctx, repository.CreateUserInput{
		Username:     "tok-user-1",
		PasswordHash: "hash",
		Name:         "Token User",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	quota := int64(500)
	expiry := time.Now().UTC().Add(time.Hour)
	resp, plaintext, err := svc.CreateUserToken(ctx, CreateUserTokenInput{
		UserID:     u.ID,
		Name:       "default",
		QuotaLimit: &quota,
		ExpiresAt:  &expiry,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if plaintext == "" {
		t.Fatal("expected plaintext token")
	}
	if resp.TokenLast4 != plaintext[len(plaintext)-4:] {
		t.Errorf("expected last4 %s, got %s", plaintext[len(plaintext)-4:], resp.TokenLast4)
	}
	if resp.QuotaLimit == nil || *resp.QuotaLimit != quota {
		t.Errorf("expected quota %d, got %v", quota, resp.QuotaLimit)
	}

	list, err := svc.ListUserTokens(ctx, u.ID)
	if err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 token, got %d", len(list))
	}
}

func TestUserTokenService_CreateUserNotFound(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserTokenService(t, client)

	_, _, err := svc.CreateUserToken(ctx, CreateUserTokenInput{
		UserID: newUUID(t),
		Name:   "default",
	})
	if !domain.IsAppError(err) {
		t.Fatalf("expected app error, got %v", err)
	}
	if !domain.ErrNotFound.Is(err.(*domain.AppError)) {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestUserTokenService_Delete(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserTokenService(t, client)

	u, err := repository.NewEntUserRepository(client).Create(ctx, repository.CreateUserInput{
		Username:     "tok-user-2",
		PasswordHash: "hash",
		Name:         "Token User 2",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	resp, _, err := svc.CreateUserToken(ctx, CreateUserTokenInput{UserID: u.ID, Name: "to-delete"})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	tokenID := parseUUID(t, resp.ID)

	if err := svc.DeleteUserToken(ctx, u.ID, tokenID); err != nil {
		t.Fatalf("delete token: %v", err)
	}

	list, err := svc.ListUserTokens(ctx, u.ID)
	if err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 tokens after delete, got %d", len(list))
	}
}

func TestUserTokenService_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserTokenService(t, client)

	u, err := repository.NewEntUserRepository(client).Create(ctx, repository.CreateUserInput{
		Username:     "tok-user-3",
		PasswordHash: "hash",
		Name:         "Token User 3",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	resp, _, err := svc.CreateUserToken(ctx, CreateUserTokenInput{UserID: u.ID, Name: "status-test"})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	updated, err := svc.UpdateUserTokenStatus(ctx, u.ID, parseUUID(t, resp.ID), false)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.IsEnabled {
		t.Error("expected token disabled")
	}
}

func TestUserTokenService_DeleteWrongUser(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserTokenService(t, client)

	u1, err := repository.NewEntUserRepository(client).Create(ctx, repository.CreateUserInput{
		Username:     "tok-user-4",
		PasswordHash: "hash",
		Name:         "User 4",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	u2, err := repository.NewEntUserRepository(client).Create(ctx, repository.CreateUserInput{
		Username:     "tok-user-5",
		PasswordHash: "hash",
		Name:         "User 5",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	resp, _, err := svc.CreateUserToken(ctx, CreateUserTokenInput{UserID: u1.ID, Name: "owned"})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	err = svc.DeleteUserToken(ctx, u2.ID, parseUUID(t, resp.ID))
	if !domain.IsAppError(err) {
		t.Fatalf("expected app error, got %v", err)
	}
	if !domain.ErrNotFound.Is(err.(*domain.AppError)) {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestUserTokenService_ValidatePlaintext(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserTokenService(t, client)

	u, err := repository.NewEntUserRepository(client).Create(ctx, repository.CreateUserInput{
		Username:     "tok-user-6",
		PasswordHash: "hash",
		Name:         "User 6",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	resp, plaintext, err := svc.CreateUserToken(ctx, CreateUserTokenInput{UserID: u.ID, Name: "validate"})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	lookup := repository.NewEntAPIKeyLookup(client)
	info, err := auth.NewAPIKeyValidator(lookup).Validate(ctx, plaintext)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if info.UserID != u.ID {
		t.Errorf("expected user_id %s, got %s", u.ID, info.UserID)
	}
	if info.TokenID != parseUUID(t, resp.ID) {
		t.Errorf("expected token_id %s, got %s", resp.ID, info.TokenID)
	}
}
