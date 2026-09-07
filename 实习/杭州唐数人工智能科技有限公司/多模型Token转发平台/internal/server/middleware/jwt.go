package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
)

const bearerPrefix = "Bearer "

// audienceContains reports whether audience is present in audiences.
// jwt.ClaimStrings in jwt/v5 does not provide a Contains helper.
func audienceContains(audiences jwt.ClaimStrings, audience string) bool {
	for _, a := range audiences {
		if a == audience {
			return true
		}
	}
	return false
}

// JWTAuth returns a middleware that validates admin Bearer JWTs (audience
// "admin") and stores the admin identity in the gin context. Inactive
// accounts are rejected. User-portal tokens are rejected by the audience
// check, preventing token cross-use.
func JWTAuth(tm auth.TokenManager) gin.HandlerFunc {
	return jwtAuthWithAudience(tm, auth.AudienceAdmin, true)
}

// UserJWTAuth returns a middleware that validates user-portal Bearer JWTs
// (audience "user") and stores the user ID in the gin context so downstream
// handlers can use auth.UserID exactly as with API key auth.
func UserJWTAuth(tm auth.TokenManager) gin.HandlerFunc {
	return jwtAuthWithAudience(tm, auth.AudienceUser, false)
}

func jwtAuthWithAudience(tm auth.TokenManager, audience string, admin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, bearerPrefix) {
			respond.Error(c, domain.ErrUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(header, bearerPrefix)
		claims, err := tm.Parse(c.Request.Context(), tokenStr)
		if err != nil {
			respond.Error(c, domain.ErrUnauthorized)
			return
		}

		if !audienceContains(claims.Audience, audience) {
			respond.Error(c, domain.ErrUnauthorized)
			return
		}

		if claims.Status != string(adminuser.StatusActive) {
			respond.Error(c, domain.ErrUnauthorized)
			return
		}

		if admin {
			auth.SetAdminContext(c, claims)
		} else {
			auth.SetUserIDContext(c, claims.UserID)
		}
		c.Next()
	}
}
