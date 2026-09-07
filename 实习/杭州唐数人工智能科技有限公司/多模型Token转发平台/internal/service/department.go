package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/department"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// CreateDepartmentInput is the service-level input for creating a department.
type CreateDepartmentInput struct {
	Name      string
	Code      string
	ParentID  *uuid.UUID
	SortOrder int
	Status    department.Status
}

// UpdateDepartmentInput is the service-level input for updating a department.
type UpdateDepartmentInput struct {
	Name      *string
	Code      *string
	ParentID  *uuid.UUID
	SortOrder *int
	Status    *department.Status
}

// DepartmentResponse is the public representation of a department.
type DepartmentResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	ParentID  string `json:"parent_id,omitempty"`
	Level     int    `json:"level"`
	Path      string `json:"path"`
	SortOrder int    `json:"sort_order"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// DepartmentService handles department management business logic.
type DepartmentService struct {
	repo repository.DepartmentRepository
}

// NewDepartmentService creates a new DepartmentService.
func NewDepartmentService(repo repository.DepartmentRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

// CreateDepartment creates a new department.
func (s *DepartmentService) CreateDepartment(ctx context.Context, input CreateDepartmentInput) (*DepartmentResponse, error) {
	if input.Status == "" {
		input.Status = department.StatusActive
	}

	d, err := s.repo.Create(ctx, repository.CreateDepartmentInput{
		Name:      input.Name,
		Code:      input.Code,
		ParentID:  input.ParentID,
		SortOrder: input.SortOrder,
		Status:    input.Status,
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, domain.ErrConflict
		}
		return nil, domain.AsAppError(err)
	}
	return toDepartmentResponse(d), nil
}

// ListDepartments returns all departments ordered by path.
func (s *DepartmentService) ListDepartments(ctx context.Context, status *department.Status) ([]DepartmentResponse, error) {
	deps, err := s.repo.List(ctx, repository.ListDepartmentFilter{Status: status})
	if err != nil {
		return nil, domain.AsAppError(err)
	}

	resp := make([]DepartmentResponse, 0, len(deps))
	for _, d := range deps {
		resp = append(resp, *toDepartmentResponse(d))
	}
	return resp, nil
}

// GetDepartment returns a department by ID.
func (s *DepartmentService) GetDepartment(ctx context.Context, id uuid.UUID) (*DepartmentResponse, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.AsAppError(err)
	}
	return toDepartmentResponse(d), nil
}

// UpdateDepartment updates an existing department.
func (s *DepartmentService) UpdateDepartment(ctx context.Context, id uuid.UUID, input UpdateDepartmentInput) (*DepartmentResponse, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.AsAppError(err)
	}

	d, err := s.repo.Update(ctx, id, repository.UpdateDepartmentInput{
		Name:      input.Name,
		Code:      input.Code,
		ParentID:  input.ParentID,
		SortOrder: input.SortOrder,
		Status:    input.Status,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		if ent.IsConstraintError(err) {
			return nil, domain.ErrConflict
		}
		return nil, domain.AsAppError(err)
	}
	return toDepartmentResponse(d), nil
}

// DeleteDepartment removes a department by ID.
func (s *DepartmentService) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.AsAppError(err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return domain.AsAppError(err)
	}
	return nil
}

func toDepartmentResponse(d *ent.Department) *DepartmentResponse {
	resp := &DepartmentResponse{
		ID:        d.ID.String(),
		Name:      d.Name,
		Code:      d.Code,
		Level:     d.Level,
		Path:      d.Path,
		SortOrder: d.SortOrder,
		Status:    d.Status.String(),
		CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if d.ParentID != nil {
		resp.ParentID = d.ParentID.String()
	}
	return resp
}
