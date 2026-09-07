package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// Guide section audiences.
const (
	GuideAudienceUser  = "user"
	GuideAudienceAdmin = "admin"
)

// GuideSectionResponse is the public representation of a guide section.
type GuideSectionResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ContentMD string `json:"content_md"`
	Audience  string `json:"audience"`
	SortOrder int    `json:"sort_order"`
	IsEnabled bool   `json:"is_enabled"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateGuideSectionInput is the service-level input for creating a section.
type CreateGuideSectionInput struct {
	Title     string
	ContentMD string
	Audience  string
	SortOrder int
	IsEnabled bool
}

// UpdateGuideSectionInput is the service-level input for updating a section.
// Nil pointer fields leave the corresponding field untouched.
type UpdateGuideSectionInput struct {
	Title     *string
	ContentMD *string
	Audience  *string
	SortOrder *int
	IsEnabled *bool
}

// GuideSectionService handles usage guide business logic.
type GuideSectionService struct {
	repo repository.GuideSectionRepository
}

// NewGuideSectionService creates a new GuideSectionService.
func NewGuideSectionService(repo repository.GuideSectionRepository) *GuideSectionService {
	return &GuideSectionService{repo: repo}
}

// CreateGuideSection validates and persists a new guide section.
func (s *GuideSectionService) CreateGuideSection(ctx context.Context, input CreateGuideSectionInput) (*GuideSectionResponse, error) {
	if err := validateGuideSectionInput(input.Title, input.ContentMD, input.Audience); err != nil {
		return nil, err
	}
	section, err := s.repo.Create(ctx, repository.CreateGuideSectionInput{
		Title:     input.Title,
		ContentMD: input.ContentMD,
		Audience:  input.Audience,
		SortOrder: input.SortOrder,
		IsEnabled: input.IsEnabled,
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return toGuideSectionResponse(section), nil
}

// GetGuideSection returns a guide section by ID.
func (s *GuideSectionService) GetGuideSection(ctx context.Context, id uuid.UUID) (*GuideSectionResponse, error) {
	section, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toGuideSectionResponse(section), nil
}

// UpdateGuideSection applies the non-nil fields after validation.
func (s *GuideSectionService) UpdateGuideSection(ctx context.Context, id uuid.UUID, input UpdateGuideSectionInput) (*GuideSectionResponse, error) {
	if input.Title != nil || input.ContentMD != nil || input.Audience != nil {
		title, contentMD, audience := input.Title, input.ContentMD, input.Audience
		existing, err := s.repo.GetByID(ctx, id)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, domain.ErrNotFound
			}
			return nil, domain.WrapInternal(err)
		}
		if title == nil {
			title = &existing.Title
		}
		if contentMD == nil {
			contentMD = &existing.ContentMd
		}
		if audience == nil {
			a := string(existing.Audience)
			audience = &a
		}
		if err := validateGuideSectionInput(*title, *contentMD, *audience); err != nil {
			return nil, err
		}
	}
	if input.Audience != nil && !isValidGuideAudience(*input.Audience) {
		return nil, domain.NewValidationError(map[string]string{"audience": "受众必须是 user 或 admin"})
	}

	section, err := s.repo.Update(ctx, id, repository.UpdateGuideSectionInput{
		Title:     input.Title,
		ContentMD: input.ContentMD,
		Audience:  input.Audience,
		SortOrder: input.SortOrder,
		IsEnabled: input.IsEnabled,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toGuideSectionResponse(section), nil
}

// DeleteGuideSection permanently removes a guide section.
func (s *GuideSectionService) DeleteGuideSection(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.WrapInternal(err)
	}
	return nil
}

// ListGuideSectionsAdmin returns a paginated list for the admin panel.
func (s *GuideSectionService) ListGuideSectionsAdmin(ctx context.Context, audience *string, page, pageSize int) ([]GuideSectionResponse, int, error) {
	if audience != nil && !isValidGuideAudience(*audience) {
		return nil, 0, domain.NewValidationError(map[string]string{"audience": "受众必须是 user 或 admin"})
	}
	page, pageSize = normalizePage(page, pageSize)
	list, total, err := s.repo.List(ctx, repository.ListGuideSectionFilter{
		Audience: audience,
		Offset:   (page - 1) * pageSize,
		Limit:    pageSize,
	})
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}
	resp := make([]GuideSectionResponse, 0, len(list))
	for _, section := range list {
		resp = append(resp, *toGuideSectionResponse(section))
	}
	return resp, total, nil
}

// ListEnabledUserSections returns the enabled user-facing guide sections
// ordered by sort_order.
func (s *GuideSectionService) ListEnabledUserSections(ctx context.Context) ([]GuideSectionResponse, error) {
	list, err := s.repo.ListEnabledByAudience(ctx, GuideAudienceUser)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	resp := make([]GuideSectionResponse, 0, len(list))
	for _, section := range list {
		resp = append(resp, *toGuideSectionResponse(section))
	}
	return resp, nil
}

func validateGuideSectionInput(title, contentMD, audience string) error {
	if title == "" {
		return domain.NewValidationError(map[string]string{"title": "标题不能为空"})
	}
	if contentMD == "" {
		return domain.NewValidationError(map[string]string{"content_md": "内容不能为空"})
	}
	if !isValidGuideAudience(audience) {
		return domain.NewValidationError(map[string]string{"audience": "受众必须是 user 或 admin"})
	}
	return nil
}

func isValidGuideAudience(audience string) bool {
	return audience == GuideAudienceUser || audience == GuideAudienceAdmin
}

func toGuideSectionResponse(s *ent.GuideSection) *GuideSectionResponse {
	return &GuideSectionResponse{
		ID:        s.ID.String(),
		Title:     s.Title,
		ContentMD: s.ContentMd,
		Audience:  string(s.Audience),
		SortOrder: s.SortOrder,
		IsEnabled: s.IsEnabled,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
