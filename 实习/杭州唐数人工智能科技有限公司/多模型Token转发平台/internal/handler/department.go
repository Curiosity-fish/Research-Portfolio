package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/department"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// DepartmentService is the subset of the department service used by DepartmentHandler.
type DepartmentService interface {
	CreateDepartment(ctx context.Context, input service.CreateDepartmentInput) (*service.DepartmentResponse, error)
	ListDepartments(ctx context.Context, status *department.Status) ([]service.DepartmentResponse, error)
	GetDepartment(ctx context.Context, id uuid.UUID) (*service.DepartmentResponse, error)
	UpdateDepartment(ctx context.Context, id uuid.UUID, input service.UpdateDepartmentInput) (*service.DepartmentResponse, error)
	DeleteDepartment(ctx context.Context, id uuid.UUID) error
}

// CreateDepartmentRequest is the request body for creating a department.
type CreateDepartmentRequest struct {
	Name      string `json:"name" binding:"required,max=100"`
	Code      string `json:"code" binding:"required,max=50"`
	ParentID  string `json:"parent_id" binding:"omitempty,uuid"`
	SortOrder int    `json:"sort_order"`
	Status    string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// UpdateDepartmentRequest is the request body for updating a department.
type UpdateDepartmentRequest struct {
	Name      *string `json:"name" binding:"omitempty,max=100"`
	Code      *string `json:"code" binding:"omitempty,max=50"`
	ParentID  *string `json:"parent_id" binding:"omitempty,uuid"`
	SortOrder *int    `json:"sort_order"`
	Status    *string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// ListDepartmentQuery holds query parameters for listing departments.
type ListDepartmentQuery struct {
	Status string `form:"status" binding:"omitempty,oneof=active inactive"`
}

// DepartmentHandler handles department management endpoints.
type DepartmentHandler struct {
	departmentService DepartmentService
}

// NewDepartmentHandler creates a new DepartmentHandler.
func NewDepartmentHandler(departmentService DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{departmentService: departmentService}
}

// Create handles POST /api/v1/admin/departments.
func (h *DepartmentHandler) Create(c *gin.Context) {
	var req CreateDepartmentRequest
	if !bindAndValidate(c, &req) {
		return
	}

	status, err := parseDepartmentStatus(req.Status)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"status": "状态值无效"}))
		return
	}

	parentID, err := parseOptionalUUID(req.ParentID)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"parent_id": "父部门 ID 格式错误"}))
		return
	}

	resp, err := h.departmentService.CreateDepartment(c.Request.Context(), service.CreateDepartmentInput{
		Name:      req.Name,
		Code:      req.Code,
		ParentID:  parentID,
		SortOrder: req.SortOrder,
		Status:    status,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, resp)
}

// List handles GET /api/v1/admin/departments.
func (h *DepartmentHandler) List(c *gin.Context) {
	var q ListDepartmentQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"query": "查询参数无效"}))
		return
	}

	status, err := parseDepartmentStatusPtr(q.Status)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"status": "状态值无效"}))
		return
	}

	resp, err := h.departmentService.ListDepartments(c.Request.Context(), status)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"list": resp})
}

// Get handles GET /api/v1/admin/departments/:id.
func (h *DepartmentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "部门 ID 格式错误"}))
		return
	}

	resp, err := h.departmentService.GetDepartment(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// Update handles PUT /api/v1/admin/departments/:id.
func (h *DepartmentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "部门 ID 格式错误"}))
		return
	}

	var req UpdateDepartmentRequest
	if !bindAndValidate(c, &req) {
		return
	}

	input, err := h.toUpdateInput(req)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"body": err.Error()}))
		return
	}

	resp, err := h.departmentService.UpdateDepartment(c.Request.Context(), id, input)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// Delete handles DELETE /api/v1/admin/departments/:id.
func (h *DepartmentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "部门 ID 格式错误"}))
		return
	}

	if err := h.departmentService.DeleteDepartment(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}

	respond.NoContent(c)
}

func (h *DepartmentHandler) toUpdateInput(req UpdateDepartmentRequest) (service.UpdateDepartmentInput, error) {
	var input service.UpdateDepartmentInput

	if req.Name != nil {
		input.Name = req.Name
	}
	if req.Code != nil {
		input.Code = req.Code
	}
	if req.SortOrder != nil {
		input.SortOrder = req.SortOrder
	}
	if req.Status != nil {
		status, err := parseDepartmentStatus(*req.Status)
		if err != nil {
			return input, err
		}
		input.Status = &status
	}
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return input, err
		}
		input.ParentID = &id
	}

	return input, nil
}

func parseDepartmentStatus(s string) (department.Status, error) {
	if s == "" {
		return department.StatusActive, nil
	}
	switch s {
	case "active":
		return department.StatusActive, nil
	case "inactive":
		return department.StatusInactive, nil
	}
	return "", domain.ErrInvalidRequest
}

func parseDepartmentStatusPtr(s string) (*department.Status, error) {
	if s == "" {
		return nil, nil
	}
	status, err := parseDepartmentStatus(s)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func parseOptionalUUID(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
