package repository

import (
	"context"
	"testing"

	"github.com/school-api/school-api-v1/ent/aimodel"
)

func TestEntAIModelRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntAIModelRepository(client)

	m, err := repo.Create(ctx, CreateAIModelInput{
		Name:         "gpt-4o",
		UpstreamName: "gpt-4o-2024-08-06",
		Type:         aimodel.TypeChat,
		InputPrice:   2500,
		OutputPrice:  10000,
		IsEnabled:    true,
	})
	if err != nil {
		t.Fatalf("create ai model: %v", err)
	}
	if m.Name != "gpt-4o" {
		t.Errorf("expected name gpt-4o, got %s", m.Name)
	}

	got, err := repo.GetByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.UpstreamName != "gpt-4o-2024-08-06" {
		t.Errorf("expected upstream name, got %s", got.UpstreamName)
	}

	byName, err := repo.GetByName(ctx, "gpt-4o")
	if err != nil {
		t.Fatalf("get by name: %v", err)
	}
	if byName.ID != m.ID {
		t.Error("get by name returned wrong model")
	}

	updated, err := repo.Update(ctx, m.ID, UpdateAIModelInput{
		OutputPrice: int64Ptr(12000),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.OutputPrice != 12000 {
		t.Errorf("expected output price 12000, got %d", updated.OutputPrice)
	}

	if err := repo.Delete(ctx, m.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}
