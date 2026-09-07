package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// GuideSectionService is the subset of the guide section service used by GuideSectionHandler.
type GuideSectionService interface {
	CreateGuideSection(ctx context.Context, input service.CreateGuideSectionInput) (*service.GuideSectionResponse, error)
	GetGuideSection(ctx context.Context, id uuid.UUID) (*service.GuideSectionResponse, error)
	UpdateGuideSection(ctx context.Context, id uuid.UUID, input service.UpdateGuideSectionInput) (*service.GuideSectionResponse, error)
	DeleteGuideSection(ctx context.Context, id uuid.UUID) error
	ListGuideSectionsAdmin(ctx context.Context, audience *string, page, pageSize int) ([]service.GuideSectionResponse, int, error)
	ListEnabledUserSections(ctx context.Context) ([]service.GuideSectionResponse, error)
}

// CreateGuideSectionRequest is the request body for creating a guide section.
type CreateGuideSectionRequest struct {
	Title     string `json:"title" binding:"required"`
	ContentMD string `json:"content_md" binding:"required"`
	Audience  string `json:"audience" binding:"required"`
	SortOrder int    `json:"sort_order"`
	IsEnabled *bool  `json:"is_enabled"`
}

// UpdateGuideSectionRequest is the request body for updating a guide section.
// Pointer fields are applied only when present.
type UpdateGuideSectionRequest struct {
	Title     *string `json:"title"`
	ContentMD *string `json:"content_md"`
	Audience  *string `json:"audience"`
	SortOrder *int    `json:"sort_order"`
	IsEnabled *bool   `json:"is_enabled"`
}

// GuideSectionHandler handles guide section endpoints.
type GuideSectionHandler struct {
	svc GuideSectionService
}

// NewGuideSectionHandler creates a new GuideSectionHandler.
func NewGuideSectionHandler(svc GuideSectionService) *GuideSectionHandler {
	return &GuideSectionHandler{svc: svc}
}

// Create handles POST /api/v1/admin/guide-sections.
func (h *GuideSectionHandler) Create(c *gin.Context) {
	var req CreateGuideSectionRequest
	if !bindAndValidate(c, &req) {
		return
	}

	resp, err := h.svc.CreateGuideSection(c.Request.Context(), service.CreateGuideSectionInput{
		Title:     req.Title,
		ContentMD: req.ContentMD,
		Audience:  req.Audience,
		SortOrder: req.SortOrder,
		IsEnabled: req.IsEnabled == nil || *req.IsEnabled,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.Created(c, resp)
}

// Get handles GET /api/v1/admin/guide-sections/:id.
func (h *GuideSectionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "指南 ID 格式错误"}))
		return
	}
	resp, err := h.svc.GetGuideSection(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// Update handles PUT /api/v1/admin/guide-sections/:id.
func (h *GuideSectionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "指南 ID 格式错误"}))
		return
	}

	var req UpdateGuideSectionRequest
	if !bindAndValidate(c, &req) {
		return
	}

	resp, err := h.svc.UpdateGuideSection(c.Request.Context(), id, service.UpdateGuideSectionInput{
		Title:     req.Title,
		ContentMD: req.ContentMD,
		Audience:  req.Audience,
		SortOrder: req.SortOrder,
		IsEnabled: req.IsEnabled,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// Delete handles DELETE /api/v1/admin/guide-sections/:id.
func (h *GuideSectionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "指南 ID 格式错误"}))
		return
	}
	if err := h.svc.DeleteGuideSection(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}
	respond.NoContent(c)
}

// ListAdmin handles GET /api/v1/admin/guide-sections.
func (h *GuideSectionHandler) ListAdmin(c *gin.Context) {
	var q struct {
		Audience string `form:"audience"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	_ = c.ShouldBindQuery(&q)

	var audience *string
	if q.Audience != "" {
		audience = &q.Audience
	}

	list, total, err := h.svc.ListGuideSectionsAdmin(c.Request.Context(), audience, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": list})
}

// ListUser handles GET /api/v1/guide/sections for the user portal: enabled
// user-facing sections ordered by sort_order.
func (h *GuideSectionHandler) ListUser(c *gin.Context) {
	list, err := h.svc.ListEnabledUserSections(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, list)
}
