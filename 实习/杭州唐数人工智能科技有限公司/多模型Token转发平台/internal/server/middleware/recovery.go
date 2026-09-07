package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/requestmeta"
	"github.com/school-api/school-api-v1/internal/server/respond"
)

// Recovery returns a middleware that recovers from panics and returns a 500 response.
// It should be registered as the outermost middleware so that panics in any
// later middleware or handler are captured.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					slog.Any("error", err),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("request_id", requestmeta.RequestID(c)),
				)

				respond.ErrorWithStatus(c,
					http.StatusInternalServerError,
					domain.ErrInternal.BizCode,
					domain.ErrInternal.Message,
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}
