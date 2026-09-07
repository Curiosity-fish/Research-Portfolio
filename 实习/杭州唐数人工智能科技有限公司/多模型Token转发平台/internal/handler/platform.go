package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/platform"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// PlatformHandler exposes platform management endpoints.
type PlatformHandler struct {
	svc service.PlatformService
}

// NewPlatformHandler creates a new PlatformHandler.
func NewPlatformHandler(svc service.PlatformService) *PlatformHandler {
	return &PlatformHandler{svc: svc}
}

// CreatePlatformRequest is the request body for creating a platform.
type CreatePlatformRequest struct {
	Name    string        `json:"name" binding:"required,max=100"`
	Code    string        `json:"code" binding:"required,max=50"`
	Type    platform.Type `json:"type" binding:"omitempty,oneof=openai anthropic"`
	BaseURL string        `json:"base_url" binding:"required,max=500,url"`
	Status  platform.Status `json:"status" binding:"omitempty,oneof=active inactive"`
}

// UpdatePlatformRequest is the request body for updating a platform.
type UpdatePlatformRequest struct {
	Name    *string        `json:"name" binding:"omitempty,max=100"`
	Code    *string        `json:"code" binding:"omitempty,max=50"`
	Type    *platform.Type `json:"type" binding:"omitempty,oneof=openai anthropic"`
	BaseURL *string        `json:"base_url" binding:"omitempty,max=500,url"`
	Status  *platform.Status `json:"status" binding:"omitempty,oneof=active inactive"`
}

// Create handles POST /api/v1/admin/platforms.
func (h *PlatformHandler) Create(c *gin.Context) {
	var req CreatePlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"body": err.Error()}))
		return
	}

	resp, err := h.svc.CreatePlatform(c.Request.Context(), service.CreatePlatformInput{
		Name:    req.Name,
		Code:    req.Code,
		Type:    req.Type,
		BaseURL: req.BaseURL,
		Status:  req.Status,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.Created(c, resp)
}

// List handles GET /api/v1/admin/platforms.
func (h *PlatformHandler) List(c *gin.Context) {
	platforms, err := h.svc.ListPlatforms(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, platforms)
}

// Get handles GET /api/v1/admin/platforms/:id.
func (h *PlatformHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "平台 ID 格式错误"}))
		return
	}

	resp, err := h.svc.GetPlatform(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// Update handles PUT /api/v1/admin/platforms/:id.
func (h *PlatformHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "平台 ID 格式错误"}))
		return
	}

	var req UpdatePlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"body": err.Error()}))
		return
	}

	resp, err := h.svc.UpdatePlatform(c.Request.Context(), id, service.UpdatePlatformInput(req))
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// Delete handles DELETE /api/v1/admin/platforms/:id.
func (h *PlatformHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "平台 ID 格式错误"}))
		return
	}

	if err := h.svc.DeletePlatform(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}
	respond.NoContent(c)
}
