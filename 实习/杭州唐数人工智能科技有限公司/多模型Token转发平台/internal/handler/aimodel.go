package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/aimodel"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// AIModelHandler exposes AI model management endpoints.
type AIModelHandler struct {
	svc service.AIModelService
}

// NewAIModelHandler creates a new AIModelHandler.
func NewAIModelHandler(svc service.AIModelService) *AIModelHandler {
	return &AIModelHandler{svc: svc}
}

// CreateAIModelRequest is the request body for creating an AI model.
type CreateAIModelRequest struct {
	Name         string       `json:"name" binding:"required,max=100"`
	UpstreamName string       `json:"upstream_name" binding:"required,max=100"`
	Type         aimodel.Type `json:"type" binding:"omitempty,oneof=chat embedding image"`
	InputPrice   int64        `json:"input_price" binding:"omitempty,min=0"`
	OutputPrice  int64        `json:"output_price" binding:"omitempty,min=0"`
	IsEnabled    bool         `json:"is_enabled"`
}

// UpdateAIModelRequest is the request body for updating an AI model.
type UpdateAIModelRequest struct {
	Name         *string       `json:"name" binding:"omitempty,max=100"`
	UpstreamName *string       `json:"upstream_name" binding:"omitempty,max=100"`
	Type         *aimodel.Type `json:"type" binding:"omitempty,oneof=chat embedding image"`
	InputPrice   *int64        `json:"input_price" binding:"omitempty,min=0"`
	OutputPrice  *int64        `json:"output_price" binding:"omitempty,min=0"`
	IsEnabled    *bool         `json:"is_enabled"`
}

// Create handles POST /api/v1/admin/models.
func (h *AIModelHandler) Create(c *gin.Context) {
	var req CreateAIModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"body": err.Error()}))
		return
	}

	resp, err := h.svc.CreateAIModel(c.Request.Context(), service.CreateAIModelInput(req))
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.Created(c, resp)
}

// List handles GET /api/v1/admin/models.
func (h *AIModelHandler) List(c *gin.Context) {
	models, err := h.svc.ListAIModels(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, models)
}

// Get handles GET /api/v1/admin/models/:id.
func (h *AIModelHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "模型 ID 格式错误"}))
		return
	}

	resp, err := h.svc.GetAIModel(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// Update handles PUT /api/v1/admin/models/:id.
func (h *AIModelHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "模型 ID 格式错误"}))
		return
	}

	var req UpdateAIModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"body": err.Error()}))
		return
	}

	resp, err := h.svc.UpdateAIModel(c.Request.Context(), id, service.UpdateAIModelInput(req))
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// Delete handles DELETE /api/v1/admin/models/:id.
func (h *AIModelHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "模型 ID 格式错误"}))
		return
	}

	if err := h.svc.DeleteAIModel(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}
	respond.NoContent(c)
}
