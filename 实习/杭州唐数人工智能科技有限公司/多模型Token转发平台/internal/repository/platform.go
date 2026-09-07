package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/platform"
)

// CreatePlatformInput is the data required to create a platform.
type CreatePlatformInput struct {
	Name    string
	Code    string
	Type    platform.Type
	BaseURL string
	Status  platform.Status
}

// UpdatePlatformInput is the data that can be updated for a platform.
// Pointer fields that are nil are left unchanged.
type UpdatePlatformInput struct {
	Name    *string
	Code    *string
	Type    *platform.Type
	BaseURL *string
	Status  *platform.Status
}

// PlatformRepository provides data access for platforms.
type PlatformRepository interface {
	Create(ctx context.Context, input CreatePlatformInput) (*ent.Platform, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Platform, error)
	GetByCode(ctx context.Context, code string) (*ent.Platform, error)
	List(ctx context.Context) ([]*ent.Platform, error)
	Update(ctx context.Context, id uuid.UUID, input UpdatePlatformInput) (*ent.Platform, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// EntPlatformRepository implements PlatformRepository using Ent.
type EntPlatformRepository struct {
	client *ent.Client
}

// NewEntPlatformRepository creates a new Ent-backed platform repository.
func NewEntPlatformRepository(client *ent.Client) *EntPlatformRepository {
	return &EntPlatformRepository{client: client}
}

// Create inserts a new platform.
func (r *EntPlatformRepository) Create(ctx context.Context, input CreatePlatformInput) (*ent.Platform, error) {
	p, err := r.client.Platform.Create().
		SetName(input.Name).
		SetCode(input.Code).
		SetType(input.Type).
		SetBaseURL(input.BaseURL).
		SetStatus(input.Status).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create platform: %w", err)
	}
	return p, nil
}

// GetByID returns a platform by ID.
func (r *EntPlatformRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
	p, err := r.client.Platform.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get platform: %w", err)
	}
	return p, nil
}

// GetByCode returns a platform by its unique code.
func (r *EntPlatformRepository) GetByCode(ctx context.Context, code string) (*ent.Platform, error) {
	p, err := r.client.Platform.Query().
		Where(platform.CodeEQ(code)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get platform by code: %w", err)
	}
	return p, nil
}

// List returns all platforms ordered by name.
func (r *EntPlatformRepository) List(ctx context.Context) ([]*ent.Platform, error) {
	platforms, err := r.client.Platform.Query().
		Order(ent.Asc(platform.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list platforms: %w", err)
	}
	return platforms, nil
}

// Update modifies a platform.
func (r *EntPlatformRepository) Update(ctx context.Context, id uuid.UUID, input UpdatePlatformInput) (*ent.Platform, error) {
	b := r.client.Platform.UpdateOneID(id)
	if input.Name != nil {
		b.SetName(*input.Name)
	}
	if input.Code != nil {
		b.SetCode(*input.Code)
	}
	if input.Type != nil {
		b.SetType(*input.Type)
	}
	if input.BaseURL != nil {
		b.SetBaseURL(*input.BaseURL)
	}
	if input.Status != nil {
		b.SetStatus(*input.Status)
	}

	p, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update platform: %w", err)
	}
	return p, nil
}

// Delete removes a platform. Accounts and group_platforms referencing it are
// deleted by CASCADE.
func (r *EntPlatformRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.client.Platform.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete platform: %w", err)
	}
	return nil
}
