package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// AuthService is the subset of the auth service used by AuthHandler.
type AuthService interface {
	Login(ctx context.Context, input service.AdminLoginInput) (service.AdminLoginResult, error)
}

// AdminLoginRequest is the request body for the admin login endpoint.
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// AdminMeResponse is the response body for the admin profile endpoint.
type AdminMeResponse struct {
	AdminID string `json:"admin_id"`
	Role    string `json:"role"`
	Status  string `json:"status"`
}

// AuthHandler handles administrator authentication endpoints.
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login issues a JWT for a valid username/password pair.
func (h *AuthHandler) Login(c *gin.Context) {
	var req AdminLoginRequest
	if !bindAndValidate(c, &req) {
		return
	}

	// bcrypt operates on bytes; reject passwords that would be silently
	// truncated beyond the 72-byte limit.
	if len([]byte(req.Password)) > 72 {
		respond.Error(c, domain.NewValidationError(map[string]string{"password": "密码长度超过最大字节限制"}))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), service.AdminLoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, result)
}

// Me returns the current authenticated admin's identity.
func (h *AuthHandler) Me(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		respond.ErrorWithStatus(c, http.StatusUnauthorized, "UNAUTHORIZED", "未授权")
		return
	}

	role, _ := auth.AdminRole(c)
	status, _ := auth.AdminStatus(c)

	respond.OK(c, AdminMeResponse{
		AdminID: adminID.String(),
		Role:    role,
		Status:  status,
	})
}
