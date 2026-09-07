package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

func newUserAuthService(t *testing.T, client *ent.Client) *UserAuthService {
	t.Helper()
	return NewUserAuthService(
		repository.NewEntUserRepository(client),
		repository.NewEntUserTokenRepository(client),
		auth.NewTokenManager("user-auth-test-secret-32bytes", time.Hour),
		auth.DefaultBcryptCost,
		testCipherKey,
		int64((time.Hour).Seconds()),
	)
}

func createPasswordUser(t *testing.T, ctx context.Context, client *ent.Client, username, password string, status user.Status) *ent.User {
	t.Helper()
	hash, err := auth.HashPassword(auth.DefaultBcryptCost, password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	u, err := repository.NewEntUserRepository(client).Create(ctx, repository.CreateUserInput{
		Username:     username,
		PasswordHash: hash,
		Name:         "Portal User",
		Role:         user.RoleStudent,
		Status:       status,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func TestUserAuthService_LoginSuccess(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserAuthService(t, client)
	createPasswordUser(t, ctx, client, "portal-1", "password123", user.StatusActive)

	result, err := svc.Login(ctx, "portal-1", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.AccessToken == "" || result.TokenType != "Bearer" || result.ExpiresIn == 0 {
		t.Fatalf("unexpected login result: %+v", result)
	}

	// The returned token must be a user-audience token.
	tm := auth.NewTokenManager("user-auth-test-secret-32bytes", time.Hour)
	claims, err := tm.Parse(ctx, result.AccessToken)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != auth.AudienceUser {
		t.Fatalf("expected user audience, got %v", claims.Audience)
	}
}

func TestUserAuthService_LoginUnknownUserSameError(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserAuthService(t, client)
	createPasswordUser(t, ctx, client, "portal-2", "password123", user.StatusActive)

	// Unknown username and wrong password must produce the identical error to
	// prevent username enumeration.
	_, errUnknown := svc.Login(ctx, "no-such-user", "password123")
	_, errWrong := svc.Login(ctx, "portal-2", "wrong-password")
	if errUnknown == nil || errWrong == nil {
		t.Fatal("expected both logins to fail")
	}
	if errUnknown.Error() != errWrong.Error() {
		t.Fatalf("error text must be identical, got %q vs %q", errUnknown.Error(), errWrong.Error())
	}
}

func TestUserAuthService_LoginInactiveUser(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserAuthService(t, client)
	createPasswordUser(t, ctx, client, "portal-3", "password123", user.StatusInactive)

	if _, err := svc.Login(ctx, "portal-3", "password123"); err == nil {
		t.Fatal("expected inactive user login to fail")
	}
}

func TestUserAuthService_ChangePassword(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserAuthService(t, client)
	u := createPasswordUser(t, ctx, client, "portal-4", "password123", user.StatusActive)

	if err := svc.ChangePassword(ctx, u.ID, "wrong-old", "newpassword1"); err == nil {
		t.Fatal("expected wrong old password to fail")
	}

	if err := svc.ChangePassword(ctx, u.ID, "password123", "newpassword1"); err != nil {
		t.Fatalf("change password: %v", err)
	}

	if _, err := svc.Login(ctx, "portal-4", "password123"); err == nil {
		t.Fatal("expected old password to be rejected after change")
	}
	if _, err := svc.Login(ctx, "portal-4", "newpassword1"); err != nil {
		t.Fatalf("expected new password to work: %v", err)
	}
}

func TestUserAuthService_RevealToken(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserAuthService(t, client)
	u := createPasswordUser(t, ctx, client, "portal-5", "password123", user.StatusActive)

	tokenSvc := newUserTokenService(t, client)
	resp, plaintext, err := tokenSvc.CreateUserToken(ctx, CreateUserTokenInput{UserID: u.ID, Name: "reveal"})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	tokenID, err := uuid.Parse(resp.ID)
	if err != nil {
		t.Fatalf("parse token id: %v", err)
	}

	// Wrong password is rejected with the unified credentials error.
	if _, err := svc.RevealToken(ctx, u.ID, tokenID, "wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail")
	}

	revealed, err := svc.RevealToken(ctx, u.ID, tokenID, "password123")
	if err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if revealed != plaintext {
		t.Fatal("expected revealed plaintext to match the one-time plaintext")
	}

	// Another user's token is reported as not found.
	other := createPasswordUser(t, ctx, client, "portal-6", "password123", user.StatusActive)
	if _, err := svc.RevealToken(ctx, other.ID, tokenID, "password123"); err == nil {
		t.Fatal("expected cross-user reveal to fail")
	}
}

func TestUserAuthService_RevealTokenWithoutCipher(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newUserAuthService(t, client)
	u := createPasswordUser(t, ctx, client, "portal-7", "password123", user.StatusActive)

	// Create a token without a cipher key (legacy behaviour).
	legacySvc := NewUserTokenService(
		repository.NewEntUserTokenRepository(client),
		repository.NewEntUserRepository(client),
		nil,
		nil,
	)
	resp, _, err := legacySvc.CreateUserToken(ctx, CreateUserTokenInput{UserID: u.ID, Name: "legacy"})
	if err != nil {
		t.Fatalf("create legacy token: %v", err)
	}
	tokenID, err := uuid.Parse(resp.ID)
	if err != nil {
		t.Fatalf("parse token id: %v", err)
	}

	_, err = svc.RevealToken(ctx, u.ID, tokenID, "password123")
	if err == nil {
		t.Fatal("expected reveal of cipher-less token to fail")
	}
	appErr := domain.AsAppError(err)
	if appErr.Code != 409 {
		t.Fatalf("expected 409 TOKEN_REVEAL_UNAVAILABLE, got %d", appErr.Code)
	}
}
