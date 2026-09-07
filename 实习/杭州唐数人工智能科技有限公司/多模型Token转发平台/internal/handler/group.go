package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// GroupService is the subset of the group service used by GroupHandler.
type GroupService interface {
	ListGroups(ctx context.Context) ([]service.GroupResponse, error)
}

// GroupHandler exposes read-only group endpoints for admin pickers.
type GroupHandler struct {
	groupService GroupService
}

// NewGroupHandler creates a new GroupHandler.
func NewGroupHandler(groupService GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

// List handles GET /api/v1/admin/groups.
func (h *GroupHandler) List(c *gin.Context) {
	resp, err := h.groupService.ListGroups(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}
