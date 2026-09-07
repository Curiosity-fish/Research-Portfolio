package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/ent/platform"
)

func TestEntAccountRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)

	p, err := client.Platform.Create().
		SetName("OpenAI").
		SetCode("openai").
		SetType(platform.TypeOpenai).
		SetBaseURL("https://api.openai.com/v1").
		SetStatus(platform.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}

	repo := NewEntAccountRepository(client)
	a, err := repo.Create(ctx, CreateAccountInput{
		PlatformID:      p.ID,
		Name:            "Primary",
		APIKeyEncrypted: "encrypted-key",
		Weight:          2,
		MaxRPM:          60,
		Status:          account.StatusActive,
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	if a.Weight != 2 {
		t.Errorf("expected weight 2, got %d", a.Weight)
	}

	got, err := repo.GetByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.PlatformID != p.ID {
		t.Errorf("expected platform id %s, got %s", p.ID, got.PlatformID)
	}

	byPlatform, err := repo.ListByPlatformID(ctx, p.ID)
	if err != nil {
		t.Fatalf("list by platform: %v", err)
	}
	if len(byPlatform) != 1 {
		t.Errorf("expected 1 account by platform, got %d", len(byPlatform))
	}

	updated, err := repo.Update(ctx, a.ID, UpdateAccountInput{
		Weight: intPtr(5),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Weight != 5 {
		t.Errorf("expected weight 5, got %d", updated.Weight)
	}

	if err := repo.Delete(ctx, a.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestEntAccountRepository_PlatformNotFound(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntAccountRepository(client)

	_, err := repo.Create(ctx, CreateAccountInput{
		PlatformID:      uuid.New(),
		Name:            "Orphan",
		APIKeyEncrypted: "x",
	})
	if err == nil {
		t.Error("expected error for non-existent platform")
	}
}

func TestEntAccountRepository_IncrementErrorCount(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)

	p, err := client.Platform.Create().
		SetName("OpenAI").
		SetCode("openai").
		SetType(platform.TypeOpenai).
		SetBaseURL("https://api.openai.com/v1").
		SetStatus(platform.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}

	repo := NewEntAccountRepository(client)
	a, err := repo.Create(ctx, CreateAccountInput{
		PlatformID:      p.ID,
		Name:            "Primary",
		APIKeyEncrypted: "encrypted-key",
		Weight:          1,
		MaxRPM:          0,
		Status:          account.StatusActive,
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	if err := repo.IncrementErrorCount(ctx, a.ID); err != nil {
		t.Fatalf("increment error count: %v", err)
	}

	got, err := repo.GetByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.ErrorCount != 1 {
		t.Errorf("expected error_count 1, got %d", got.ErrorCount)
	}
}

func intPtr(i int) *int {
	return &i
}
