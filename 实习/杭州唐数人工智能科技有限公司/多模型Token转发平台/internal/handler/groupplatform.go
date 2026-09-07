package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// GroupPlatformHandler exposes group-platform binding endpoints.
type GroupPlatformHandler struct {
	svc service.GroupPlatformService
}

// NewGroupPlatformHandler creates a new GroupPlatformHandler.
func NewGroupPlatformHandler(svc service.GroupPlatformService) *GroupPlatformHandler {
	return &GroupPlatformHandler{svc: svc}
}

// BindPlatformRequest is the request body for binding a platform to a group.
type BindPlatformRequest struct {
	PlatformID uuid.UUID `json:"platform_id" binding:"required"`
}

// Bind handles POST /api/v1/admin/groups/:id/platforms.
func (h *GroupPlatformHandler) Bind(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "分组 ID 格式错误"}))
		return
	}

	var req BindPlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"body": err.Error()}))
		return
	}

	if err := h.svc.BindPlatform(c.Request.Context(), groupID, req.PlatformID); err != nil {
		respond.Error(c, err)
		return
	}
	respond.NoContent(c)
}

// Unbind handles DELETE /api/v1/admin/groups/:id/platforms/:platform_id.
func (h *GroupPlatformHandler) Unbind(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "分组 ID 格式错误"}))
		return
	}
	platformID, err := uuid.Parse(c.Param("platform_id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "平台 ID 格式错误"}))
		return
	}

	if err := h.svc.UnbindPlatform(c.Request.Context(), groupID, platformID); err != nil {
		respond.Error(c, err)
		return
	}
	respond.NoContent(c)
}

// List handles GET /api/v1/admin/groups/:id/platforms.
func (h *GroupPlatformHandler) List(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "分组 ID 格式错误"}))
		return
	}

	ids, err := h.svc.ListPlatformsByGroup(c.Request.Context(), groupID)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, ids)
}
