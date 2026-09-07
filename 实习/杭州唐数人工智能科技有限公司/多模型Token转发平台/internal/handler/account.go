package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// AccountHandler exposes account management endpoints.
type AccountHandler struct {
	svc service.AccountService
}

// NewAccountHandler creates a new AccountHandler.
func NewAccountHandler(svc service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

// CreateAccountRequest is the request body for creating an account.
type CreateAccountRequest struct {
	PlatformID uuid.UUID     `json:"platform_id" binding:"required"`
	Name       string        `json:"name" binding:"required,max=100"`
	APIKey     string        `json:"api_key" binding:"required"`
	Weight     int           `json:"weight" binding:"omitempty,min=1"`
	MaxRPM     int           `json:"max_rpm" binding:"omitempty,min=0"`
	Status     account.Status `json:"status" binding:"omitempty,oneof=active inactive error"`
}

// UpdateAccountRequest is the request body for updating an account.
type UpdateAccountRequest struct {
	Name   *string        `json:"name" binding:"omitempty,max=100"`
	APIKey *string        `json:"api_key" binding:"omitempty"`
	Weight *int           `json:"weight" binding:"omitempty,min=1"`
	MaxRPM *int           `json:"max_rpm" binding:"omitempty,min=0"`
	Status *account.Status `json:"status" binding:"omitempty,oneof=active inactive error"`
}

// Create handles POST /api/v1/admin/accounts.
func (h *AccountHandler) Create(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"body": err.Error()}))
		return
	}

	resp, err := h.svc.CreateAccount(c.Request.Context(), service.CreateAccountInput{
		PlatformID: req.PlatformID,
		Name:       req.Name,
		APIKey:     req.APIKey,
		Weight:     req.Weight,
		MaxRPM:     req.MaxRPM,
		Status:     req.Status,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.Created(c, resp)
}

// List handles GET /api/v1/admin/accounts.
func (h *AccountHandler) List(c *gin.Context) {
	accounts, err := h.svc.ListAccounts(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, accounts)
}

// Get handles GET /api/v1/admin/accounts/:id.
func (h *AccountHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "账号 ID 格式错误"}))
		return
	}

	resp, err := h.svc.GetAccount(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// Update handles PUT /api/v1/admin/accounts/:id.
func (h *AccountHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "账号 ID 格式错误"}))
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"body": err.Error()}))
		return
	}

	resp, err := h.svc.UpdateAccount(c.Request.Context(), id, service.UpdateAccountInput(req))
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// Delete handles DELETE /api/v1/admin/accounts/:id.
func (h *AccountHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "账号 ID 格式错误"}))
		return
	}

	if err := h.svc.DeleteAccount(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}
	respond.NoContent(c)
}
