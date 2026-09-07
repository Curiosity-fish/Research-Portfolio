package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/server/requestmeta"
)

// RequestID returns a middleware that assigns a unique request ID to each request.
// If the client provides X-Request-ID, it is reused; otherwise a UUID is generated.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestmeta.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set(requestmeta.RequestIDKey, requestID)
		c.Header(requestmeta.RequestIDHeader, requestID)
		ctx := context.WithValue(c.Request.Context(), requestmeta.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// RequestLogger returns a middleware that logs each request with structured fields.
// Only the URL path is logged; the raw query string is deliberately omitted
// because it may carry sensitive parameters (tokens, codes, credentials).
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("http_request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.String("request_id", requestmeta.RequestID(c)),
		)
	}
}
