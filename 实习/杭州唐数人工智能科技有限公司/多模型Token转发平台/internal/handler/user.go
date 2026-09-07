package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// UserService is the subset of the user service used by UserHandler.
type UserService interface {
	CreateUser(ctx context.Context, input service.CreateUserInput) (*service.UserResponse, error)
	ListUsers(ctx context.Context, filter service.ListUsersFilter) (*service.ListUsersResponse, error)
	GetUser(ctx context.Context, id uuid.UUID) (*service.UserResponse, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*service.UserResponse, error)
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status user.Status) (*service.UserResponse, error)
	BulkImport(ctx context.Context, input service.BulkImportInput) (*service.BulkImportResult, error)
}

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Username     string `json:"username" binding:"required,min=3,max=50"`
	Password     string `json:"password" binding:"required,min=8,max=72"`
	Name         string `json:"name" binding:"required,max=100"`
	Email        string `json:"email" binding:"omitempty,max=255,email"`
	Phone        string `json:"phone" binding:"omitempty,max=20"`
	Role         string `json:"role" binding:"omitempty,oneof=student teacher staff"`
	Gender       string `json:"gender" binding:"omitempty,max=10"`
	Status       string `json:"status" binding:"omitempty,oneof=active inactive banned"`
	DepartmentID string `json:"department_id" binding:"omitempty,uuid"`
	GroupID      string `json:"group_id" binding:"omitempty,uuid"`
}

// UpdateUserRequest is the request body for updating a user. All fields are
// optional; nil fields are left unchanged.
type UpdateUserRequest struct {
	Name         *string `json:"name" binding:"omitempty,max=100"`
	Password     *string `json:"password" binding:"omitempty,min=8,max=72"`
	Email        *string `json:"email" binding:"omitempty,max=255,email"`
	Phone        *string `json:"phone" binding:"omitempty,max=20"`
	Role         *string `json:"role" binding:"omitempty,oneof=student teacher staff"`
	Gender       *string `json:"gender" binding:"omitempty,max=10"`
	DepartmentID *string `json:"department_id" binding:"omitempty,uuid"`
	GroupID      *string `json:"group_id" binding:"omitempty,uuid"`
}

// UpdateUserStatusRequest is the request body for the status-only update endpoint.
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive banned"`
}

// ListUsersQuery holds the query parameters for the user list endpoint.
type ListUsersQuery struct {
	Keyword      string `form:"keyword"`
	Status       string `form:"status" binding:"omitempty,oneof=active inactive banned"`
	Role         string `form:"role" binding:"omitempty,oneof=student teacher staff"`
	DepartmentID string `form:"department_id" binding:"omitempty,uuid"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

// UserHandler handles user management endpoints for administrators.
type UserHandler struct {
	userService UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Create handles POST /api/v1/admin/users.
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if !bindAndValidate(c, &req) {
		return
	}

	if len([]byte(req.Password)) > 72 {
		respond.Error(c, domain.NewValidationError(map[string]string{"password": "密码长度超过最大字节限制"}))
		return
	}

	role, err := parseUserRole(req.Role)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"role": err.Error()}))
		return
	}
	status, err := parseUserStatus(req.Status)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"status": err.Error()}))
		return
	}

	deptID, groupID, err := parseOptionalUUIDs(req.DepartmentID, req.GroupID)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": err.Error()}))
		return
	}

	resp, err := h.userService.CreateUser(c.Request.Context(), service.CreateUserInput{
		Username:     req.Username,
		Password:     req.Password,
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Role:         role,
		Gender:       req.Gender,
		Status:       status,
		DepartmentID: deptID,
		GroupID:      groupID,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, resp)
}

// List handles GET /api/v1/admin/users.
func (h *UserHandler) List(c *gin.Context) {
	var q ListUsersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"query": "查询参数无效"}))
		return
	}

	status, err := parseUserStatusPtr(q.Status)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"status": err.Error()}))
		return
	}
	role, err := parseUserRolePtr(q.Role)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"role": err.Error()}))
		return
	}
	var deptID *uuid.UUID
	if q.DepartmentID != "" {
		id, err := uuid.Parse(q.DepartmentID)
		if err != nil {
			respond.Error(c, domain.NewValidationError(map[string]string{"department_id": "部门 ID 格式错误"}))
			return
		}
		deptID = &id
	}

	resp, err := h.userService.ListUsers(c.Request.Context(), service.ListUsersFilter{
		Keyword:      q.Keyword,
		Status:       status,
		Role:         role,
		DepartmentID: deptID,
		Page:         q.Page,
		PageSize:     q.PageSize,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// Get handles GET /api/v1/admin/users/:id.
func (h *UserHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "用户 ID 格式错误"}))
		return
	}

	resp, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// Update handles PUT /api/v1/admin/users/:id.
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "用户 ID 格式错误"}))
		return
	}

	var req UpdateUserRequest
	if !bindAndValidate(c, &req) {
		return
	}

	if req.Password != nil && len([]byte(*req.Password)) > 72 {
		respond.Error(c, domain.NewValidationError(map[string]string{"password": "密码长度超过最大字节限制"}))
		return
	}

	input, err := h.toUpdateInput(req)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"body": err.Error()}))
		return
	}

	resp, err := h.userService.UpdateUser(c.Request.Context(), id, input)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// UpdateStatus handles PATCH /api/v1/admin/users/:id/status.
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "用户 ID 格式错误"}))
		return
	}

	var req UpdateUserStatusRequest
	if !bindAndValidate(c, &req) {
		return
	}

	status, err := parseUserStatus(req.Status)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"status": err.Error()}))
		return
	}

	resp, err := h.userService.UpdateUserStatus(c.Request.Context(), id, status)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

func (h *UserHandler) toUpdateInput(req UpdateUserRequest) (service.UpdateUserInput, error) {
	var input service.UpdateUserInput

	if req.Name != nil {
		input.Name = req.Name
	}
	if req.Password != nil {
		input.Password = req.Password
	}
	if req.Email != nil {
		input.Email = req.Email
	}
	if req.Phone != nil {
		input.Phone = req.Phone
	}
	if req.Gender != nil {
		input.Gender = req.Gender
	}
	if req.Role != nil {
		role, err := parseUserRole(*req.Role)
		if err != nil {
			return input, err
		}
		input.Role = &role
	}
	if req.DepartmentID != nil {
		id, err := uuid.Parse(*req.DepartmentID)
		if err != nil {
			return input, err
		}
		input.DepartmentID = &id
	}
	if req.GroupID != nil {
		id, err := uuid.Parse(*req.GroupID)
		if err != nil {
			return input, err
		}
		input.GroupID = &id
	}

	return input, nil
}

func parseUserRole(s string) (user.Role, error) {
	if s == "" {
		return user.RoleStudent, nil
	}
	switch s {
	case "student":
		return user.RoleStudent, nil
	case "teacher":
		return user.RoleTeacher, nil
	case "staff":
		return user.RoleStaff, nil
	}
	return "", domain.ErrInvalidRequest
}

func parseUserStatus(s string) (user.Status, error) {
	if s == "" {
		return user.StatusActive, nil
	}
	switch s {
	case "active":
		return user.StatusActive, nil
	case "inactive":
		return user.StatusInactive, nil
	case "banned":
		return user.StatusBanned, nil
	}
	return "", domain.ErrInvalidRequest
}

func parseUserRolePtr(s string) (*user.Role, error) {
	if s == "" {
		return nil, nil
	}
	role, err := parseUserRole(s)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func parseUserStatusPtr(s string) (*user.Status, error) {
	if s == "" {
		return nil, nil
	}
	status, err := parseUserStatus(s)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func parseOptionalUUIDs(deptID, groupID string) (*uuid.UUID, *uuid.UUID, error) {
	var d, g *uuid.UUID
	if deptID != "" {
		id, err := uuid.Parse(deptID)
		if err != nil {
			return nil, nil, err
		}
		d = &id
	}
	if groupID != "" {
		id, err := uuid.Parse(groupID)
		if err != nil {
			return nil, nil, err
		}
		g = &id
	}
	return d, g, nil
}
