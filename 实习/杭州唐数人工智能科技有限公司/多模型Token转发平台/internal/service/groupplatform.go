package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// GroupPlatformService manages group-platform bindings.
type GroupPlatformService interface {
	BindPlatform(ctx context.Context, groupID, platformID uuid.UUID) error
	UnbindPlatform(ctx context.Context, groupID, platformID uuid.UUID) error
	ListPlatformsByGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error)
}

type groupPlatformService struct {
	repo         repository.GroupPlatformRepository
	groupRepo    repository.GroupRepository
	platformRepo repository.PlatformRepository
}

// NewGroupPlatformService creates a new GroupPlatformService.
func NewGroupPlatformService(
	repo repository.GroupPlatformRepository,
	groupRepo repository.GroupRepository,
	platformRepo repository.PlatformRepository,
) GroupPlatformService {
	return &groupPlatformService{
		repo:         repo,
		groupRepo:    groupRepo,
		platformRepo: platformRepo,
	}
}

// BindPlatform binds a platform to a group.
func (s *groupPlatformService) BindPlatform(ctx context.Context, groupID, platformID uuid.UUID) error {
	if _, err := s.groupRepo.GetByID(ctx, groupID); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.WrapInternal(err)
	}
	if _, err := s.platformRepo.GetByID(ctx, platformID); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.WrapInternal(err)
	}

	if _, err := s.repo.Bind(ctx, groupID, platformID); err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// UnbindPlatform removes a platform binding from a group.
func (s *groupPlatformService)UnbindPlatform(ctx context.Context, groupID, platformID uuid.UUID) error {
	if err := s.repo.Unbind(ctx, groupID, platformID); err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// ListPlatformsByGroup returns platform IDs bound to the given group.
func (s *groupPlatformService)ListPlatformsByGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	if _, err := s.groupRepo.GetByID(ctx, groupID); err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	ids, err := s.repo.ListPlatformsByGroupID(ctx, groupID)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return ids, nil
}
