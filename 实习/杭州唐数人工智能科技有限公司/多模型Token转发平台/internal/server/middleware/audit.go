package middleware

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/service"
)

// AuditLogService is the subset of the audit log service needed by the middleware.
type AuditLogService interface {
	Log(ctx context.Context, input service.CreateAuditLogInput) error
}

// AuditMiddleware returns a middleware that writes an audit log for each admin
// request after the handler has run. Audit failures are logged but never fail
// the request, to avoid breaking operational endpoints because of logging issues.
func AuditMiddleware(svc AuditLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		adminID, ok := auth.AdminID(c)
		if !ok {
			return
		}

		details := map[string]any{
			"path":   c.Request.URL.Path,
			"query":  c.Request.URL.RawQuery,
			"method": c.Request.Method,
			"status": c.Writer.Status(),
		}

		ip := c.ClientIP()
		ua := c.Request.UserAgent()

		if err := svc.Log(c.Request.Context(), service.CreateAuditLogInput{
			ActorType: "admin",
			ActorID:   adminID,
			Action:    c.Request.Method + " " + c.Request.URL.Path,
			Details:   details,
			IP:        &ip,
			UserAgent: &ua,
		}); err != nil {
			slog.Error("failed to write audit log", "error", err, "path", c.Request.URL.Path)
		}
	}
}
