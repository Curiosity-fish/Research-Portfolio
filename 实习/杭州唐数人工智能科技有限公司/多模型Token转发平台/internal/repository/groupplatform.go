package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/groupplatform"
)

// GroupPlatformRepository provides data access for group-platform bindings.
type GroupPlatformRepository interface {
	// Bind links a platform to a group. It is idempotent: duplicate bindings return no error.
	Bind(ctx context.Context, groupID, platformID uuid.UUID) (*ent.GroupPlatform, error)
	// Unbind removes the link between a group and a platform.
	Unbind(ctx context.Context, groupID, platformID uuid.UUID) error
	// ListPlatformsByGroupID returns all platform IDs bound to the given group.
	ListPlatformsByGroupID(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error)
	// ListGroupsByPlatformID returns all group IDs bound to the given platform.
	ListGroupsByPlatformID(ctx context.Context, platformID uuid.UUID) ([]uuid.UUID, error)
}

// EntGroupPlatformRepository implements GroupPlatformRepository using Ent.
type EntGroupPlatformRepository struct {
	client *ent.Client
}

// NewEntGroupPlatformRepository creates a new Ent-backed group-platform repository.
func NewEntGroupPlatformRepository(client *ent.Client) *EntGroupPlatformRepository {
	return &EntGroupPlatformRepository{client: client}
}

// Bind links a platform to a group, ignoring duplicate key conflicts.
func (r *EntGroupPlatformRepository) Bind(ctx context.Context, groupID, platformID uuid.UUID) (*ent.GroupPlatform, error) {
	gp, err := r.client.GroupPlatform.Create().
		SetGroupID(groupID).
		SetPlatformID(platformID).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			// Already bound: return the existing record.
			existing, qerr := r.client.GroupPlatform.Query().
				Where(
					groupplatform.GroupIDEQ(groupID),
					groupplatform.PlatformIDEQ(platformID),
				).
				Only(ctx)
			if qerr != nil {
				return nil, fmt.Errorf("query existing group-platform binding: %w", qerr)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("bind group-platform: %w", err)
	}
	return gp, nil
}

// Unbind removes the link between a group and a platform.
func (r *EntGroupPlatformRepository) Unbind(ctx context.Context, groupID, platformID uuid.UUID) error {
	_, err := r.client.GroupPlatform.Delete().
		Where(
			groupplatform.GroupIDEQ(groupID),
			groupplatform.PlatformIDEQ(platformID),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("unbind group-platform: %w", err)
	}
	return nil
}

// ListPlatformsByGroupID returns all platform IDs bound to the given group.
func (r *EntGroupPlatformRepository) ListPlatformsByGroupID(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.client.GroupPlatform.Query().
		Where(groupplatform.GroupIDEQ(groupID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list platforms by group: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, gp := range rows {
		ids = append(ids, gp.PlatformID)
	}
	return ids, nil
}

// ListGroupsByPlatformID returns all group IDs bound to the given platform.
func (r *EntGroupPlatformRepository) ListGroupsByPlatformID(ctx context.Context, platformID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.client.GroupPlatform.Query().
		Where(groupplatform.PlatformIDEQ(platformID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups by platform: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, gp := range rows {
		ids = append(ids, gp.GroupID)
	}
	return ids, nil
}
