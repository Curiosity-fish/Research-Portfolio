package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/aimodel"
)

// CreateAIModelInput is the data required to create an AI model.
type CreateAIModelInput struct {
	Name         string
	UpstreamName string
	Type         aimodel.Type
	InputPrice   int64
	OutputPrice  int64
	IsEnabled    bool
}

// UpdateAIModelInput is the data that can be updated for an AI model.
type UpdateAIModelInput struct {
	Name         *string
	UpstreamName *string
	Type         *aimodel.Type
	InputPrice   *int64
	OutputPrice  *int64
	IsEnabled    *bool
}

// AIModelRepository provides data access for AI models.
type AIModelRepository interface {
	Create(ctx context.Context, input CreateAIModelInput) (*ent.AIModel, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.AIModel, error)
	GetByName(ctx context.Context, name string) (*ent.AIModel, error)
	List(ctx context.Context) ([]*ent.AIModel, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateAIModelInput) (*ent.AIModel, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// EntAIModelRepository implements AIModelRepository using Ent.
type EntAIModelRepository struct {
	client *ent.Client
}

// NewEntAIModelRepository creates a new Ent-backed AI model repository.
func NewEntAIModelRepository(client *ent.Client) *EntAIModelRepository {
	return &EntAIModelRepository{client: client}
}

// Create inserts a new AI model.
func (r *EntAIModelRepository) Create(ctx context.Context, input CreateAIModelInput) (*ent.AIModel, error) {
	m, err := r.client.AIModel.Create().
		SetName(input.Name).
		SetUpstreamName(input.UpstreamName).
		SetType(input.Type).
		SetInputPrice(input.InputPrice).
		SetOutputPrice(input.OutputPrice).
		SetIsEnabled(input.IsEnabled).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create ai model: %w", err)
	}
	return m, nil
}

// GetByID returns an AI model by ID.
func (r *EntAIModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.AIModel, error) {
	m, err := r.client.AIModel.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ai model: %w", err)
	}
	return m, nil
}

// GetByName returns an AI model by its unique name.
func (r *EntAIModelRepository) GetByName(ctx context.Context, name string) (*ent.AIModel, error) {
	m, err := r.client.AIModel.Query().
		Where(aimodel.NameEQ(name)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get ai model by name: %w", err)
	}
	return m, nil
}

// List returns all AI models ordered by name.
func (r *EntAIModelRepository) List(ctx context.Context) ([]*ent.AIModel, error) {
	models, err := r.client.AIModel.Query().
		Order(ent.Asc(aimodel.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list ai models: %w", err)
	}
	return models, nil
}

// Update modifies an AI model.
func (r *EntAIModelRepository) Update(ctx context.Context, id uuid.UUID, input UpdateAIModelInput) (*ent.AIModel, error) {
	b := r.client.AIModel.UpdateOneID(id)
	if input.Name != nil {
		b.SetName(*input.Name)
	}
	if input.UpstreamName != nil {
		b.SetUpstreamName(*input.UpstreamName)
	}
	if input.Type != nil {
		b.SetType(*input.Type)
	}
	if input.InputPrice != nil {
		b.SetInputPrice(*input.InputPrice)
	}
	if input.OutputPrice != nil {
		b.SetOutputPrice(*input.OutputPrice)
	}
	if input.IsEnabled != nil {
		b.SetIsEnabled(*input.IsEnabled)
	}

	m, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update ai model: %w", err)
	}
	return m, nil
}

// Delete removes an AI model.
func (r *EntAIModelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.client.AIModel.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete ai model: %w", err)
	}
	return nil
}
