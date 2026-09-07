package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

type mockPlatformRepo struct {
	createFunc func(ctx context.Context, input repository.CreatePlatformInput) (*ent.Platform, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*ent.Platform, error)
	listFunc   func(ctx context.Context) ([]*ent.Platform, error)
	updateFunc func(ctx context.Context, id uuid.UUID, input repository.UpdatePlatformInput) (*ent.Platform, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockPlatformRepo) Create(ctx context.Context, input repository.CreatePlatformInput) (*ent.Platform, error) {
	return m.createFunc(ctx, input)
}

func (m *mockPlatformRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
	return m.getFunc(ctx, id)
}

func (m *mockPlatformRepo) GetByCode(ctx context.Context, code string) (*ent.Platform, error) {
	return nil, nil
}

func (m *mockPlatformRepo) List(ctx context.Context) ([]*ent.Platform, error) {
	return m.listFunc(ctx)
}

func (m *mockPlatformRepo) Update(ctx context.Context, id uuid.UUID, input repository.UpdatePlatformInput) (*ent.Platform, error) {
	return m.updateFunc(ctx, id, input)
}

func (m *mockPlatformRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

func TestPlatformService_Create(t *testing.T) {
	repo := &mockPlatformRepo{
		createFunc: func(ctx context.Context, input repository.CreatePlatformInput) (*ent.Platform, error) {
			return &ent.Platform{
				ID:      uuid.New(),
				Name:    input.Name,
				Code:    input.Code,
				Type:    input.Type,
				BaseURL: input.BaseURL,
				Status:  input.Status,
			}, nil
		},
	}

	svc := NewPlatformService(repo)
	resp, err := svc.CreatePlatform(context.Background(), CreatePlatformInput{
		Name:    "OpenAI",
		Code:    "openai",
		BaseURL: "https://api.openai.com/v1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.Type != "openai" {
		t.Errorf("expected type openai, got %s", resp.Type)
	}
	if resp.Status != "active" {
		t.Errorf("expected status active, got %s", resp.Status)
	}
}

func TestPlatformService_GetNotFound(t *testing.T) {
	repo := &mockPlatformRepo{
		getFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return nil, &ent.NotFoundError{}
		},
	}

	svc := NewPlatformService(repo)
	_, err := svc.GetPlatform(context.Background(), uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
