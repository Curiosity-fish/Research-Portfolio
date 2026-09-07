package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/server/respond"
)

// LogRetentionService is the subset of the retention service used by the
// maintenance handler.
type LogRetentionService interface {
	SweepOnce(ctx context.Context) (int, error)
}

// MaintenanceHandler exposes admin-triggered maintenance actions. The manual
// sweep exists so operators can verify retention immediately after a deploy
// instead of waiting for the next scheduled pass.
type MaintenanceHandler struct {
	retention LogRetentionService
}

// NewMaintenanceHandler creates a new MaintenanceHandler.
func NewMaintenanceHandler(retention LogRetentionService) *MaintenanceHandler {
	return &MaintenanceHandler{retention: retention}
}

// SweepCallLogs handles POST /api/v1/admin/maintenance/call-logs/sweep.
func (h *MaintenanceHandler) SweepCallLogs(c *gin.Context) {
	deleted, err := h.retention.SweepOnce(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"deleted": deleted})
}
