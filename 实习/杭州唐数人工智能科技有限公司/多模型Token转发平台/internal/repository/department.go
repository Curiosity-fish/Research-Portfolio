package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/department"
	"github.com/school-api/school-api-v1/ent/predicate"
)

// CreateDepartmentInput is the data required to create a department.
type CreateDepartmentInput struct {
	Name       string
	Code       string
	ParentID   *uuid.UUID
	SortOrder  int
	Status     department.Status
}

// UpdateDepartmentInput is the data that can be updated for a department.
// Pointer fields that are nil are left unchanged.
type UpdateDepartmentInput struct {
	Name      *string
	Code      *string
	ParentID  *uuid.UUID
	SortOrder *int
	Status    *department.Status
}

// ListDepartmentFilter controls department list queries.
type ListDepartmentFilter struct {
	Status *department.Status
}

// DepartmentRepository provides data access for departments.
type DepartmentRepository interface {
	Create(ctx context.Context, input CreateDepartmentInput) (*ent.Department, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Department, error)
	GetByCode(ctx context.Context, code string) (*ent.Department, error)
	List(ctx context.Context, filter ListDepartmentFilter) ([]*ent.Department, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateDepartmentInput) (*ent.Department, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// EntDepartmentRepository implements DepartmentRepository using Ent.
type EntDepartmentRepository struct {
	client *ent.Client
}

// NewEntDepartmentRepository creates a new Ent-backed department repository.
func NewEntDepartmentRepository(client *ent.Client) *EntDepartmentRepository {
	return &EntDepartmentRepository{client: client}
}

// Create inserts a new department and computes its level/path based on the parent.
func (r *EntDepartmentRepository) Create(ctx context.Context, input CreateDepartmentInput) (*ent.Department, error) {
	id := uuid.New()
	level := 0
	path := id.String()

	if input.ParentID != nil && *input.ParentID != (uuid.UUID{}) {
		parent, err := r.client.Department.Get(ctx, *input.ParentID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, fmt.Errorf("parent department not found")
			}
			return nil, fmt.Errorf("query parent department: %w", err)
		}
		level = parent.Level + 1
		path = parent.Path + "/" + id.String()
	}

	d, err := r.client.Department.Create().
		SetID(id).
		SetName(input.Name).
		SetCode(input.Code).
		SetLevel(level).
		SetPath(path).
		SetSortOrder(input.SortOrder).
		SetStatus(input.Status).
		SetNillableParentID(input.ParentID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}
	return d, nil
}

// GetByID returns a department by ID.
func (r *EntDepartmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Department, error) {
	d, err := r.client.Department.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get department: %w", err)
	}
	return d, nil
}

// GetByCode returns a department by its unique code.
func (r *EntDepartmentRepository) GetByCode(ctx context.Context, code string) (*ent.Department, error) {
	d, err := r.client.Department.Query().
		Where(department.CodeEQ(code)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get department by code: %w", err)
	}
	return d, nil
}

// List returns all departments ordered by path (topological/depth order).
func (r *EntDepartmentRepository) List(ctx context.Context, filter ListDepartmentFilter) ([]*ent.Department, error) {
	preds := []predicate.Department{}
	if filter.Status != nil {
		preds = append(preds, department.StatusEQ(*filter.Status))
	}

	deps, err := r.client.Department.Query().
		Where(preds...).
		Order(ent.Asc(department.FieldPath), ent.Asc(department.FieldSortOrder)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	return deps, nil
}

// Update modifies a department. Changing the parent triggers a recursive update
// of the department's level and path, plus all descendants.
func (r *EntDepartmentRepository) Update(ctx context.Context, id uuid.UUID, input UpdateDepartmentInput) (*ent.Department, error) {
	current, err := r.client.Department.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get department: %w", err)
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start transaction: %w", err)
	}

	// A helper to roll back on failure without shadowing the original error.
	rollback := func(err error) (*ent.Department, error) {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return nil, err
	}

	b := tx.Department.UpdateOneID(id)
	if input.Name != nil {
		b.SetName(*input.Name)
	}
	if input.Code != nil {
		b.SetCode(*input.Code)
	}
	if input.SortOrder != nil {
		b.SetSortOrder(*input.SortOrder)
	}
	if input.Status != nil {
		b.SetStatus(*input.Status)
	}

	parentChanged := input.ParentID != nil && (current.ParentID == nil || *input.ParentID != *current.ParentID)
	if parentChanged {
		newParentID := *input.ParentID
		newLevel := 0
		newPath := id.String()

		if newParentID != (uuid.UUID{}) {
			parent, err := tx.Department.Get(ctx, newParentID)
			if err != nil {
				return rollback(fmt.Errorf("query new parent: %w", err))
			}
			if strings.HasPrefix(parent.Path, current.Path+"/") || parent.ID == current.ID {
				return rollback(fmt.Errorf("cannot move department under itself or a descendant"))
			}
			newLevel = parent.Level + 1
			newPath = parent.Path + "/" + id.String()
		}

		if newParentID == (uuid.UUID{}) {
			b.ClearParentID()
		} else {
			b.SetParentID(newParentID)
		}
		b.SetLevel(newLevel).SetPath(newPath)

		// Update descendants' level and path.
		oldPath := current.Path
		descendants, err := tx.Department.Query().
			Where(department.PathHasPrefix(oldPath + "/")).
			All(ctx)
		if err != nil {
			return rollback(fmt.Errorf("query descendants: %w", err))
		}

		levelDelta := newLevel - current.Level
		for _, desc := range descendants {
			suffix := strings.TrimPrefix(desc.Path, oldPath+"/")
			updatedPath := newPath + "/" + suffix
			if err := tx.Department.UpdateOneID(desc.ID).
				SetLevel(desc.Level + levelDelta).
				SetPath(updatedPath).
				Exec(ctx); err != nil {
				return rollback(fmt.Errorf("update descendant %s: %w", desc.ID, err))
			}
		}
	}

	updated, err := b.Save(ctx)
	if err != nil {
		return rollback(fmt.Errorf("update department: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return updated, nil
}

// Delete removes a department. Children become root departments because the
// foreign key uses ON DELETE SET NULL.
func (r *EntDepartmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.client.Department.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete department: %w", err)
	}
	return nil
}
