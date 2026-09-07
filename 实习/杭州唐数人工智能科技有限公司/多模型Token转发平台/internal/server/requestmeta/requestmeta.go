// Package requestmeta holds per-request metadata shared between middleware
// (which produces it) and respond/handlers (which consume it). Keeping the
// keys here avoids an import cycle: middleware already imports respond for
// error responses, so respond cannot import middleware for the key.
package requestmeta

import "github.com/gin-gonic/gin"

// RequestIDKey is the gin context key under which the request ID is stored.
const RequestIDKey = "request_id"

// RequestIDHeader is the HTTP header that carries the request ID.
const RequestIDHeader = "X-Request-ID"

// RequestID returns the request ID stored in the context, or "" if absent.
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
