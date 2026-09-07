package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// UserTokenService is the subset of the token service used by UserTokenHandler.
type UserTokenService interface {
	CreateUserToken(ctx context.Context, input service.CreateUserTokenInput) (*service.UserTokenResponse, string, error)
	ListUserTokens(ctx context.Context, userID uuid.UUID) ([]service.UserTokenResponse, error)
	DeleteUserToken(ctx context.Context, userID, tokenID uuid.UUID) error
	UpdateUserTokenStatus(ctx context.Context, userID, tokenID uuid.UUID, enabled bool) (*service.UserTokenResponse, error)
}

// CreateUserTokenRequest is the request body for creating a token.
type CreateUserTokenRequest struct {
	Name       string `json:"name" binding:"required,max=100"`
	QuotaLimit *int64 `json:"quota_limit" binding:"omitempty,gte=0"`
	ExpiresAt  string `json:"expires_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

// CreateUserTokenResponse wraps the one-time plaintext key.
type CreateUserTokenResponse struct {
	*service.UserTokenResponse
	Plaintext string `json:"plaintext_token"`
}

// UpdateUserTokenStatusRequest is the request body for enabling/disabling a token.
type UpdateUserTokenStatusRequest struct {
	IsEnabled *bool `json:"is_enabled" binding:"required"`
}

// UserTokenHandler handles API token management endpoints.
type UserTokenHandler struct {
	tokenService UserTokenService
}

// NewUserTokenHandler creates a new UserTokenHandler.
func NewUserTokenHandler(tokenService UserTokenService) *UserTokenHandler {
	return &UserTokenHandler{tokenService: tokenService}
}

// Create handles POST /api/v1/admin/users/:id/tokens.
func (h *UserTokenHandler) Create(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "用户 ID 格式错误"}))
		return
	}

	var req CreateUserTokenRequest
	if !bindAndValidate(c, &req) {
		return
	}

	expiresAt, err := parseOptionalTime(req.ExpiresAt)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"expires_at": "时间格式错误"}))
		return
	}

	resp, plaintext, err := h.tokenService.CreateUserToken(c.Request.Context(), service.CreateUserTokenInput{
		UserID:     userID,
		Name:       req.Name,
		QuotaLimit: req.QuotaLimit,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, CreateUserTokenResponse{
		UserTokenResponse: resp,
		Plaintext:         plaintext,
	})
}

// List handles GET /api/v1/admin/users/:id/tokens.
func (h *UserTokenHandler) List(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "用户 ID 格式错误"}))
		return
	}

	toks, err := h.tokenService.ListUserTokens(c.Request.Context(), userID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"list": toks})
}

// Delete handles DELETE /api/v1/admin/users/:id/tokens/:token_id.
func (h *UserTokenHandler) Delete(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "用户 ID 格式错误"}))
		return
	}
	tokenID, err := uuid.Parse(c.Param("token_id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"token_id": "Token ID 格式错误"}))
		return
	}

	if err := h.tokenService.DeleteUserToken(c.Request.Context(), userID, tokenID); err != nil {
		respond.Error(c, err)
		return
	}

	respond.NoContent(c)
}

// UpdateStatus handles PATCH /api/v1/admin/users/:id/tokens/:token_id/status.
func (h *UserTokenHandler) UpdateStatus(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "用户 ID 格式错误"}))
		return
	}
	tokenID, err := uuid.Parse(c.Param("token_id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"token_id": "Token ID 格式错误"}))
		return
	}

	var req UpdateUserTokenStatusRequest
	if !bindAndValidate(c, &req) {
		return
	}

	resp, err := h.tokenService.UpdateUserTokenStatus(c.Request.Context(), userID, tokenID, *req.IsEnabled)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

func parseOptionalTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
