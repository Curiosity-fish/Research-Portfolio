package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/guidesection"
)

// CreateGuideSectionInput is the data required to create a guide section.
type CreateGuideSectionInput struct {
	Title     string
	ContentMD string
	Audience  string
	SortOrder int
	IsEnabled bool
}

// UpdateGuideSectionInput carries the fields updatable by the admin. Nil
// pointer fields leave the corresponding column untouched.
type UpdateGuideSectionInput struct {
	Title     *string
	ContentMD *string
	Audience  *string
	SortOrder *int
	IsEnabled *bool
}

// ListGuideSectionFilter controls the admin list query.
type ListGuideSectionFilter struct {
	Audience *string
	Offset   int
	Limit    int
}

// GuideSectionRepository provides data access for guide sections.
type GuideSectionRepository interface {
	Create(ctx context.Context, input CreateGuideSectionInput) (*ent.GuideSection, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.GuideSection, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateGuideSectionInput) (*ent.GuideSection, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListGuideSectionFilter) ([]*ent.GuideSection, int, error)
	// ListEnabledByAudience returns enabled sections for the given audience
	// ordered by sort_order ascending (then created_at as a stable tiebreak).
	ListEnabledByAudience(ctx context.Context, audience string) ([]*ent.GuideSection, error)
}

// EntGuideSectionRepository implements GuideSectionRepository using Ent.
type EntGuideSectionRepository struct {
	client *ent.Client
}

// NewEntGuideSectionRepository creates a new Ent-backed guide section repository.
func NewEntGuideSectionRepository(client *ent.Client) *EntGuideSectionRepository {
	return &EntGuideSectionRepository{client: client}
}

// Create inserts a new guide section.
func (r *EntGuideSectionRepository) Create(ctx context.Context, input CreateGuideSectionInput) (*ent.GuideSection, error) {
	s, err := r.client.GuideSection.Create().
		SetTitle(input.Title).
		SetContentMd(input.ContentMD).
		SetAudience(guidesection.Audience(input.Audience)).
		SetSortOrder(input.SortOrder).
		SetIsEnabled(input.IsEnabled).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create guide section: %w", err)
	}
	return s, nil
}

// GetByID returns a guide section by ID.
func (r *EntGuideSectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.GuideSection, error) {
	s, err := r.client.GuideSection.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get guide section: %w", err)
	}
	return s, nil
}

// Update applies the non-nil fields of the input to the guide section.
func (r *EntGuideSectionRepository) Update(ctx context.Context, id uuid.UUID, input UpdateGuideSectionInput) (*ent.GuideSection, error) {
	b := r.client.GuideSection.UpdateOneID(id)
	if input.Title != nil {
		b.SetTitle(*input.Title)
	}
	if input.ContentMD != nil {
		b.SetContentMd(*input.ContentMD)
	}
	if input.Audience != nil {
		b.SetAudience(guidesection.Audience(*input.Audience))
	}
	if input.SortOrder != nil {
		b.SetSortOrder(*input.SortOrder)
	}
	if input.IsEnabled != nil {
		b.SetIsEnabled(*input.IsEnabled)
	}
	s, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update guide section: %w", err)
	}
	return s, nil
}

// Delete permanently removes a guide section.
func (r *EntGuideSectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.client.GuideSection.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete guide section: %w", err)
	}
	return nil
}

// List returns a paginated list of guide sections for the admin panel.
func (r *EntGuideSectionRepository) List(ctx context.Context, filter ListGuideSectionFilter) ([]*ent.GuideSection, int, error) {
	q := r.client.GuideSection.Query()
	if filter.Audience != nil {
		q = q.Where(guidesection.AudienceEQ(guidesection.Audience(*filter.Audience)))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count guide sections: %w", err)
	}

	list, err := q.
		Order(ent.Asc(guidesection.FieldSortOrder), ent.Asc(guidesection.FieldCreatedAt)).
		Offset(filter.Offset).
		Limit(filter.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list guide sections: %w", err)
	}
	return list, total, nil
}

// ListEnabledByAudience returns enabled sections for the given audience
// ordered by sort_order ascending.
func (r *EntGuideSectionRepository) ListEnabledByAudience(ctx context.Context, audience string) ([]*ent.GuideSection, error) {
	list, err := r.client.GuideSection.Query().
		Where(
			guidesection.AudienceEQ(guidesection.Audience(audience)),
			guidesection.IsEnabledEQ(true),
		).
		Order(ent.Asc(guidesection.FieldSortOrder), ent.Asc(guidesection.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled guide sections: %w", err)
	}
	return list, nil
}

var _ GuideSectionRepository = (*EntGuideSectionRepository)(nil)
