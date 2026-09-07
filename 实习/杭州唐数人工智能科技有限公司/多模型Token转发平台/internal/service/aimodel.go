package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/aimodel"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// CreateAIModelInput is the service-level input for creating an AI model.
type CreateAIModelInput struct {
	Name         string
	UpstreamName string
	Type         aimodel.Type
	InputPrice   int64
	OutputPrice  int64
	IsEnabled    bool
}

// UpdateAIModelInput is the service-level input for updating an AI model.
type UpdateAIModelInput struct {
	Name         *string
	UpstreamName *string
	Type         *aimodel.Type
	InputPrice   *int64
	OutputPrice  *int64
	IsEnabled    *bool
}

// AIModelResponse is the public representation of an AI model.
type AIModelResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	UpstreamName string `json:"upstream_name"`
	Type         string `json:"type"`
	InputPrice   int64  `json:"input_price"`
	OutputPrice  int64  `json:"output_price"`
	IsEnabled    bool   `json:"is_enabled"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// AIModelService defines AI model management business logic.
type AIModelService interface {
	CreateAIModel(ctx context.Context, input CreateAIModelInput) (*AIModelResponse, error)
	ListAIModels(ctx context.Context) ([]AIModelResponse, error)
	GetAIModel(ctx context.Context, id uuid.UUID) (*AIModelResponse, error)
	UpdateAIModel(ctx context.Context, id uuid.UUID, input UpdateAIModelInput) (*AIModelResponse, error)
	DeleteAIModel(ctx context.Context, id uuid.UUID) error
}

type aiModelService struct {
	repo repository.AIModelRepository
}

// NewAIModelService creates a new AIModelService.
func NewAIModelService(repo repository.AIModelRepository) AIModelService {
	return &aiModelService{repo: repo}
}

// CreateAIModel creates a new AI model.
func (s *aiModelService) CreateAIModel(ctx context.Context, input CreateAIModelInput) (*AIModelResponse, error) {
	if input.Type == "" {
		input.Type = aimodel.TypeChat
	}

	m, err := s.repo.Create(ctx, repository.CreateAIModelInput{
		Name:         input.Name,
		UpstreamName: input.UpstreamName,
		Type:         input.Type,
		InputPrice:   input.InputPrice,
		OutputPrice:  input.OutputPrice,
		IsEnabled:    input.IsEnabled,
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, domain.ErrConflict
		}
		return nil, domain.WrapInternal(err)
	}
	return toAIModelResponse(m), nil
}

// ListAIModels returns all AI models.
func (s *aiModelService)ListAIModels(ctx context.Context) ([]AIModelResponse, error) {
	models, err := s.repo.List(ctx)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]AIModelResponse, 0, len(models))
	for _, m := range models {
		resp = append(resp, *toAIModelResponse(m))
	}
	return resp, nil
}

// GetAIModel returns an AI model by ID.
func (s *aiModelService)GetAIModel(ctx context.Context, id uuid.UUID) (*AIModelResponse, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toAIModelResponse(m), nil
}

// UpdateAIModel updates an existing AI model.
func (s *aiModelService)UpdateAIModel(ctx context.Context, id uuid.UUID, input UpdateAIModelInput) (*AIModelResponse, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	m, err := s.repo.Update(ctx, id, repository.UpdateAIModelInput{
		Name:         input.Name,
		UpstreamName: input.UpstreamName,
		Type:         input.Type,
		InputPrice:   input.InputPrice,
		OutputPrice:  input.OutputPrice,
		IsEnabled:    input.IsEnabled,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		if ent.IsConstraintError(err) {
			return nil, domain.ErrConflict
		}
		return nil, domain.WrapInternal(err)
	}
	return toAIModelResponse(m), nil
}

// DeleteAIModel removes an AI model by ID.
func (s *aiModelService)DeleteAIModel(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.WrapInternal(err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

func toAIModelResponse(m *ent.AIModel) *AIModelResponse {
	return &AIModelResponse{
		ID:           m.ID.String(),
		Name:         m.Name,
		UpstreamName: m.UpstreamName,
		Type:         m.Type.String(),
		InputPrice:   m.InputPrice,
		OutputPrice:  m.OutputPrice,
		IsEnabled:    m.IsEnabled,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
