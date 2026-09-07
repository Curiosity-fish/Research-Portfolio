package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/repository"
)

type mockAIModelRepo struct {
	createFunc func(ctx context.Context, input repository.CreateAIModelInput) (*ent.AIModel, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*ent.AIModel, error)
	listFunc   func(ctx context.Context) ([]*ent.AIModel, error)
	updateFunc func(ctx context.Context, id uuid.UUID, input repository.UpdateAIModelInput) (*ent.AIModel, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockAIModelRepo) Create(ctx context.Context, input repository.CreateAIModelInput) (*ent.AIModel, error) {
	return m.createFunc(ctx, input)
}

func (m *mockAIModelRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.AIModel, error) {
	return m.getFunc(ctx, id)
}

func (m *mockAIModelRepo) GetByName(ctx context.Context, name string) (*ent.AIModel, error) {
	return nil, nil
}

func (m *mockAIModelRepo) List(ctx context.Context) ([]*ent.AIModel, error) {
	return m.listFunc(ctx)
}

func (m *mockAIModelRepo) Update(ctx context.Context, id uuid.UUID, input repository.UpdateAIModelInput) (*ent.AIModel, error) {
	return m.updateFunc(ctx, id, input)
}

func (m *mockAIModelRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

func TestAIModelService_Create(t *testing.T) {
	repo := &mockAIModelRepo{
		createFunc: func(ctx context.Context, input repository.CreateAIModelInput) (*ent.AIModel, error) {
			return &ent.AIModel{
				ID:           uuid.New(),
				Name:         input.Name,
				UpstreamName: input.UpstreamName,
				Type:         input.Type,
				InputPrice:   input.InputPrice,
				OutputPrice:  input.OutputPrice,
				IsEnabled:    input.IsEnabled,
			}, nil
		},
	}

	svc := NewAIModelService(repo)
	resp, err := svc.CreateAIModel(context.Background(), CreateAIModelInput{
		Name:         "gpt-4o",
		UpstreamName: "gpt-4o-2024-08-06",
		InputPrice:   2500,
		OutputPrice:  10000,
		IsEnabled:    true,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.Type != "chat" {
		t.Errorf("expected default type chat, got %s", resp.Type)
	}
	if resp.InputPrice != 2500 {
		t.Errorf("expected input price 2500, got %d", resp.InputPrice)
	}
}

func TestAIModelService_UpdatePrice(t *testing.T) {
	id := uuid.New()
	repo := &mockAIModelRepo{
		getFunc: func(ctx context.Context, uid uuid.UUID) (*ent.AIModel, error) {
			return &ent.AIModel{ID: id}, nil
		},
		updateFunc: func(ctx context.Context, uid uuid.UUID, input repository.UpdateAIModelInput) (*ent.AIModel, error) {
			return &ent.AIModel{ID: id, OutputPrice: *input.OutputPrice}, nil
		},
	}

	svc := NewAIModelService(repo)
	newPrice := int64(12000)
	resp, err := svc.UpdateAIModel(context.Background(), id, UpdateAIModelInput{OutputPrice: &newPrice})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if resp.OutputPrice != 12000 {
		t.Errorf("expected output price 12000, got %d", resp.OutputPrice)
	}
}
