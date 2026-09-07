package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockAPIKeyLookup struct {
	result *APIKeyLookupResult
	err    error
}

func (m *mockAPIKeyLookup) GetByTokenHash(ctx context.Context, hash string) (*APIKeyLookupResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestGenerateAPIKey(t *testing.T) {
	plaintext, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}
	if !strings.HasPrefix(plaintext, APIKeyPrefix) {
		t.Errorf("expected prefix %s, got %s", APIKeyPrefix, plaintext)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if HashAPIKey(plaintext) != hash {
		t.Error("hash mismatch")
	}
}

func TestAPIKeyValidator_Valid(t *testing.T) {
	groupID := uuid.New()
	lookup := &mockAPIKeyLookup{
		result: &APIKeyLookupResult{
			UserID:     uuid.New(),
			TokenID:    uuid.New(),
			IsEnabled:  true,
			UserStatus: "active",
			GroupID:    groupID,
			GroupCode:  "g1",
		},
	}
	v := NewAPIKeyValidator(lookup)

	info, err := v.Validate(context.Background(), "sk-test")
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if info.GroupCode != "g1" {
		t.Errorf("expected group code g1, got %s", info.GroupCode)
	}
}

func TestAPIKeyValidator_Disabled(t *testing.T) {
	lookup := &mockAPIKeyLookup{
		result: &APIKeyLookupResult{
			UserID:    uuid.New(),
			TokenID:   uuid.New(),
			IsEnabled: false,
		},
	}
	v := NewAPIKeyValidator(lookup)

	_, err := v.Validate(context.Background(), "sk-test")
	if err == nil {
		t.Fatal("expected error for disabled token")
	}
}

func TestAPIKeyValidator_Expired(t *testing.T) {
	past := time.Now().UTC().Add(-time.Hour)
	lookup := &mockAPIKeyLookup{
		result: &APIKeyLookupResult{
			UserID:    uuid.New(),
			TokenID:   uuid.New(),
			IsEnabled: true,
			ExpiresAt: &past,
		},
	}
	v := NewAPIKeyValidator(lookup)

	_, err := v.Validate(context.Background(), "sk-test")
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestAPIKeyValidator_UserInactive(t *testing.T) {
	lookup := &mockAPIKeyLookup{
		result: &APIKeyLookupResult{
			UserID:     uuid.New(),
			TokenID:    uuid.New(),
			IsEnabled:  true,
			UserStatus: "banned",
		},
	}
	v := NewAPIKeyValidator(lookup)

	_, err := v.Validate(context.Background(), "sk-test")
	if err == nil {
		t.Fatal("expected error for inactive user")
	}
}
