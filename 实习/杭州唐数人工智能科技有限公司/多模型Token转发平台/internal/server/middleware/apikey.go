package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
)

// APIKeyAuth returns a middleware that validates Bearer API keys (sk-xxx) and
// stores the resulting identity in the gin context.
func APIKeyAuth(validator auth.APIKeyValidator) gin.HandlerFunc {
	return buildAPIKeyAuth(validator, false)
}

// OpenAIAPIKeyAuth is the same as APIKeyAuth but returns OpenAI-compatible
// error responses on authentication failures. It is intended for `/v1/*`
// routes consumed by OpenAI SDKs.
func OpenAIAPIKeyAuth(validator auth.APIKeyValidator) gin.HandlerFunc {
	return buildAPIKeyAuth(validator, true)
}

func buildAPIKeyAuth(validator auth.APIKeyValidator, openAIErrors bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, bearerPrefix) {
			unauthorized(c, openAIErrors)
			return
		}

		key := strings.TrimPrefix(header, bearerPrefix)
		if !strings.HasPrefix(key, auth.APIKeyPrefix) {
			unauthorized(c, openAIErrors)
			return
		}

		info, err := validator.Validate(c.Request.Context(), key)
		if err != nil {
			unauthorized(c, openAIErrors)
			return
		}

		auth.SetAPIKeyContext(c, info)
		c.Next()
	}
}

func unauthorized(c *gin.Context, openAIErrors bool) {
	if openAIErrors {
		respond.OpenAIError(c, 401, "invalid_request_error", "invalid_api_key", "Incorrect API key provided", "")
		return
	}
	respond.Error(c, domain.ErrUnauthorized)
}
