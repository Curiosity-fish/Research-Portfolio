package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/adminuser"
)

func newManager(t *testing.T, expiry time.Duration) TokenManager {
	t.Helper()
	return NewTokenManager("test-secret-at-least-32-bytes-long", expiry)
}

func TestTokenManager_GenerateAndParse(t *testing.T) {
	manager := newManager(t, time.Hour)
	admin := &ent.AdminUser{
		ID:     uuid.New(),
		Role:   adminuser.RoleSuperAdmin,
		Status: adminuser.StatusActive,
	}

	tokenStr, err := manager.Generate(context.Background(), admin)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := manager.Parse(context.Background(), tokenStr)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.AdminID != admin.ID {
		t.Errorf("expected admin id %v, got %v", admin.ID, claims.AdminID)
	}
	if claims.Role != string(admin.Role) {
		t.Errorf("expected role %s, got %s", admin.Role, claims.Role)
	}
	if claims.Status != string(admin.Status) {
		t.Errorf("expected status %s, got %s", admin.Status, claims.Status)
	}
	if claims.Subject != admin.ID.String() {
		t.Errorf("expected subject %s, got %s", admin.ID.String(), claims.Subject)
	}
	if claims.Issuer != Issuer {
		t.Errorf("expected issuer %s, got %s", Issuer, claims.Issuer)
	}
}

func TestTokenManager_Parse_Expired(t *testing.T) {
	manager := newManager(t, -time.Hour)
	admin := &ent.AdminUser{
		ID:     uuid.New(),
		Role:   adminuser.RoleAdmin,
		Status: adminuser.StatusActive,
	}

	tokenStr, err := manager.Generate(context.Background(), admin)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	_, err = manager.Parse(context.Background(), tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestTokenManager_Parse_WrongSecret(t *testing.T) {
	manager := newManager(t, time.Hour)
	admin := &ent.AdminUser{
		ID:     uuid.New(),
		Role:   adminuser.RoleAdmin,
		Status: adminuser.StatusActive,
	}

	tokenStr, err := manager.Generate(context.Background(), admin)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	other := NewTokenManager("other-secret-at-least-32-bytes-long", time.Hour)
	_, err = other.Parse(context.Background(), tokenStr)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}

func TestTokenManager_Parse_Invalid(t *testing.T) {
	manager := newManager(t, time.Hour)

	_, err := manager.Parse(context.Background(), "not-a-valid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
