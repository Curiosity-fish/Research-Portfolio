package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/platform"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// CreatePlatformInput is the service-level input for creating a platform.
type CreatePlatformInput struct {
	Name    string
	Code    string
	Type    platform.Type
	BaseURL string
	Status  platform.Status
}

// UpdatePlatformInput is the service-level input for updating a platform.
type UpdatePlatformInput struct {
	Name    *string
	Code    *string
	Type    *platform.Type
	BaseURL *string
	Status  *platform.Status
}

// PlatformResponse is the public representation of a platform.
type PlatformResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Type      string `json:"type"`
	BaseURL   string `json:"base_url"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PlatformService defines platform management business logic.
type PlatformService interface {
	CreatePlatform(ctx context.Context, input CreatePlatformInput) (*PlatformResponse, error)
	ListPlatforms(ctx context.Context) ([]PlatformResponse, error)
	GetPlatform(ctx context.Context, id uuid.UUID) (*PlatformResponse, error)
	UpdatePlatform(ctx context.Context, id uuid.UUID, input UpdatePlatformInput) (*PlatformResponse, error)
	DeletePlatform(ctx context.Context, id uuid.UUID) error
}

type platformService struct {
	repo repository.PlatformRepository
}

// NewPlatformService creates a new PlatformService.
func NewPlatformService(repo repository.PlatformRepository) PlatformService {
	return &platformService{repo: repo}
}

// CreatePlatform creates a new platform.
func (s *platformService) CreatePlatform(ctx context.Context, input CreatePlatformInput) (*PlatformResponse, error) {
	if input.Type == "" {
		input.Type = platform.TypeOpenai
	}
	if input.Status == "" {
		input.Status = platform.StatusActive
	}

	p, err := s.repo.Create(ctx, repository.CreatePlatformInput{
		Name:    input.Name,
		Code:    input.Code,
		Type:    input.Type,
		BaseURL: input.BaseURL,
		Status:  input.Status,
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, domain.ErrConflict
		}
		return nil, domain.WrapInternal(err)
	}
	return toPlatformResponse(p), nil
}

// ListPlatforms returns all platforms.
func (s *platformService)ListPlatforms(ctx context.Context) ([]PlatformResponse, error) {
	platforms, err := s.repo.List(ctx)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]PlatformResponse, 0, len(platforms))
	for _, p := range platforms {
		resp = append(resp, *toPlatformResponse(p))
	}
	return resp, nil
}

// GetPlatform returns a platform by ID.
func (s *platformService)GetPlatform(ctx context.Context, id uuid.UUID) (*PlatformResponse, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toPlatformResponse(p), nil
}

// UpdatePlatform updates an existing platform.
func (s *platformService)UpdatePlatform(ctx context.Context, id uuid.UUID, input UpdatePlatformInput) (*PlatformResponse, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	p, err := s.repo.Update(ctx, id, repository.UpdatePlatformInput{
		Name:    input.Name,
		Code:    input.Code,
		Type:    input.Type,
		BaseURL: input.BaseURL,
		Status:  input.Status,
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
	return toPlatformResponse(p), nil
}

// DeletePlatform removes a platform by ID.
func (s *platformService)DeletePlatform(ctx context.Context, id uuid.UUID) error {
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

func toPlatformResponse(p *ent.Platform) *PlatformResponse {
	return &PlatformResponse{
		ID:        p.ID.String(),
		Name:      p.Name,
		Code:      p.Code,
		Type:      p.Type.String(),
		BaseURL:   p.BaseURL,
		Status:    p.Status.String(),
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
