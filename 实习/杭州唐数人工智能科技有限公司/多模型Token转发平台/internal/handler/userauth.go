package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// UserAuthService is the subset of the user account service used by UserAuthHandler.
type UserAuthService interface {
	Login(ctx context.Context, username, password string) (service.UserLoginResult, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	RevealToken(ctx context.Context, userID, tokenID uuid.UUID, password string) (string, error)
}

// UserAuthHandler handles user-portal account endpoints (login, password,
// token reveal).
type UserAuthHandler struct {
	svc UserAuthService
}

// NewUserAuthHandler creates a new UserAuthHandler.
func NewUserAuthHandler(svc UserAuthService) *UserAuthHandler {
	return &UserAuthHandler{svc: svc}
}

// UserLoginRequest is the body for POST /api/v1/user/login.
type UserLoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=1,max=72"`
}

// ChangePasswordRequest is the body for POST /api/v1/user/change-password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=1,max=72"`
	NewPassword string `json:"password" binding:"required,min=8,max=72"`
}

// RevealTokenRequest is the body for POST /api/v1/user/token/reveal.
type RevealTokenRequest struct {
	TokenID  string `json:"token_id" binding:"required,uuid"`
	Password string `json:"password" binding:"required,min=1,max=72"`
}

// Login handles POST /api/v1/user/login.
func (h *UserAuthHandler) Login(c *gin.Context) {
	var req UserLoginRequest
	if !bindAndValidate(c, &req) {
		return
	}

	result, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, result)
}

// ChangePassword handles POST /api/v1/user/change-password.
func (h *UserAuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if !bindAndValidate(c, &req) {
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"changed": true})
}

// RevealToken handles POST /api/v1/user/token/reveal.
func (h *UserAuthHandler) RevealToken(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	var req RevealTokenRequest
	if !bindAndValidate(c, &req) {
		return
	}

	tokenID, err := uuid.Parse(req.TokenID)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"token_id": "Token ID 格式错误"}))
		return
	}

	plaintext, err := h.svc.RevealToken(c.Request.Context(), userID, tokenID, req.Password)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"token": plaintext})
}
