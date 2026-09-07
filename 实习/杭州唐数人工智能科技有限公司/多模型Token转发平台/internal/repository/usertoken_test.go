package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntUserTokenRepository_CreateAndList(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	user, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)

	repo := NewEntUserTokenRepository(client)
	plaintext, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}

	quota := int64(1000)
	expiry := time.Now().UTC().Add(time.Hour)
	tok, err := repo.Create(ctx, CreateUserTokenInput{
		UserID:     user.ID,
		Name:       "secondary",
		TokenHash:  hash,
		TokenLast4: plaintext[len(plaintext)-4:],
		QuotaLimit: &quota,
		ExpiresAt:  &expiry,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if tok.Name != "secondary" {
		t.Errorf("expected name secondary, got %s", tok.Name)
	}

	toks, err := repo.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	if len(toks) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(toks))
	}
}

func TestEntUserTokenRepository_GetAndDelete(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	_, _, existing, _ := testutil.CreateTestAPIKey(t, ctx, client)

	repo := NewEntUserTokenRepository(client)
	got, err := repo.GetByID(ctx, existing.ID)
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if got.ID != existing.ID {
		t.Errorf("expected token id %s, got %s", existing.ID, got.ID)
	}

	if err := repo.Delete(ctx, existing.ID); err != nil {
		t.Fatalf("delete token: %v", err)
	}

	_, err = repo.GetByID(ctx, existing.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestEntUserTokenRepository_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	_, _, existing, _ := testutil.CreateTestAPIKey(t, ctx, client)

	repo := NewEntUserTokenRepository(client)
	updated, err := repo.UpdateStatus(ctx, existing.ID, false)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.IsEnabled {
		t.Error("expected token disabled")
	}
}

func TestEntUserTokenRepository_CreateDuplicateHash(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	user1, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	user2, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)

	repo := NewEntUserTokenRepository(client)
	plaintext, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}

	_, err = repo.Create(ctx, CreateUserTokenInput{
		UserID:     user1.ID,
		Name:       "dup1",
		TokenHash:  hash,
		TokenLast4: plaintext[len(plaintext)-4:],
	})
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = repo.Create(ctx, CreateUserTokenInput{
		UserID:     user2.ID,
		Name:       "dup2",
		TokenHash:  hash,
		TokenLast4: plaintext[len(plaintext)-4:],
	})
	if err == nil {
		t.Fatal("expected error for duplicate token hash")
	}
}

func TestEntUserTokenRepository_ListWrongUser(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	user, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)

	repo := NewEntUserTokenRepository(client)
	toks, err := repo.ListByUserID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	if len(toks) != 0 {
		t.Errorf("expected 0 tokens for unrelated user, got %d", len(toks))
	}
	_ = user
}
