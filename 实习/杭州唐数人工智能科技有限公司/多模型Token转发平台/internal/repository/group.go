package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/group"
)

// GroupRepository provides minimal data access for groups.
type GroupRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Group, error)
	// List returns all groups ordered by sort_order for admin pickers.
	List(ctx context.Context) ([]*ent.Group, error)
}

// EntGroupRepository implements GroupRepository using Ent.
type EntGroupRepository struct {
	client *ent.Client
}

// NewEntGroupRepository creates a new Ent-backed group repository.
func NewEntGroupRepository(client *ent.Client) *EntGroupRepository {
	return &EntGroupRepository{client: client}
}

// GetByID returns a group by ID.
func (r *EntGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Group, error) {
	g, err := r.client.Group.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}
	return g, nil
}

// List returns all groups ordered by sort_order ascending.
func (r *EntGroupRepository) List(ctx context.Context) ([]*ent.Group, error) {
	groups, err := r.client.Group.Query().
		Order(ent.Asc(group.FieldSortOrder)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	return groups, nil
}
