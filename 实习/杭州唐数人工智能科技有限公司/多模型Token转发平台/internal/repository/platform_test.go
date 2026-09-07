package repository

import (
	"context"
	"testing"

	"github.com/school-api/school-api-v1/ent/platform"
)

func TestEntPlatformRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntPlatformRepository(client)

	p, err := repo.Create(ctx, CreatePlatformInput{
		Name:    "OpenAI",
		Code:    "openai",
		Type:    platform.TypeOpenai,
		BaseURL: "https://api.openai.com/v1",
		Status:  platform.StatusActive,
	})
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}
	if p.Code != "openai" {
		t.Errorf("expected code openai, got %s", p.Code)
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected id %s, got %s", p.ID, got.ID)
	}

	byCode, err := repo.GetByCode(ctx, "openai")
	if err != nil {
		t.Fatalf("get by code: %v", err)
	}
	if byCode.ID != p.ID {
		t.Errorf("get by code returned wrong platform")
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 platform, got %d", len(list))
	}

	updated, err := repo.Update(ctx, p.ID, UpdatePlatformInput{
		Name: strPtr("OpenAI Updated"),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "OpenAI Updated" {
		t.Errorf("expected updated name, got %s", updated.Name)
	}

	if err := repo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = repo.GetByID(ctx, p.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func strPtr(s string) *string {
	return &s
}
