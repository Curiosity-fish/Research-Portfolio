package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/service"
)

// HealthCheckerService is the interface expected by HealthHandler.
type HealthCheckerService interface {
	Check(ctx context.Context) *service.HealthResult
}

// HealthHandler handles health check requests.
type HealthHandler struct {
	healthService HealthCheckerService
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(healthService HealthCheckerService) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

// Health returns the health status of the service and its dependencies.
// The response body shape is identical for 200 and 503 so that callers can
// always inspect which component failed.
func (h *HealthHandler) Health(c *gin.Context) {
	result := h.healthService.Check(c.Request.Context())

	status := http.StatusOK
	if result.Status == service.HealthStatusUnavailable {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, result)
}
