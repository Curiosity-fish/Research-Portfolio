package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Context keys use a project-specific prefix to avoid collisions with
// third-party middleware or handlers that store values in gin.Context.
const (
	adminIDKey     = "school-api-v1/admin_id"
	adminRoleKey   = "school-api-v1/admin_role"
	adminStatusKey = "school-api-v1/admin_status"
)

// SetAdminContext stores admin identity information in the gin context.
func SetAdminContext(c *gin.Context, claims *Claims) {
	c.Set(adminIDKey, claims.AdminID)
	c.Set(adminRoleKey, claims.Role)
	c.Set(adminStatusKey, claims.Status)
}

// AdminID returns the admin ID stored in the context, if any.
func AdminID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(adminIDKey)
	if !ok {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// AdminRole returns the admin role stored in the context, if any.
func AdminRole(c *gin.Context) (string, bool) {
	v, ok := c.Get(adminRoleKey)
	if !ok {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}

// AdminStatus returns the admin status stored in the context, if any.
func AdminStatus(c *gin.Context) (string, bool) {
	v, ok := c.Get(adminStatusKey)
	if !ok {
		return "", false
	}
	status, ok := v.(string)
	return status, ok
}
