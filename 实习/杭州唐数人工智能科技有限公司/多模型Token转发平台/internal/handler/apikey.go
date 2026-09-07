package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/server/respond"
)

// APIKeyMeResponse is the response body for the API key identity endpoint.
type APIKeyMeResponse struct {
	UserID    string `json:"user_id"`
	TokenID   string `json:"token_id"`
	GroupID   string `json:"group_id,omitempty"`
	GroupCode string `json:"group_code,omitempty"`
}

// APIKeyHandler exposes endpoints for API-key-authenticated clients.
type APIKeyHandler struct{}

// NewAPIKeyHandler creates a new APIKeyHandler.
func NewAPIKeyHandler() *APIKeyHandler {
	return &APIKeyHandler{}
}

// Me returns the current API key identity.
func (h *APIKeyHandler) Me(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.ErrorWithStatus(c, http.StatusUnauthorized, "UNAUTHORIZED", "未授权")
		return
	}

	tokenID, _ := auth.TokenID(c)
	groupID, _ := auth.GroupID(c)
	groupCode, _ := auth.GroupCode(c)

	resp := APIKeyMeResponse{
		UserID:    userID.String(),
		TokenID:   tokenID.String(),
		GroupCode: groupCode,
	}
	if groupID != (uuid.UUID{}) {
		resp.GroupID = groupID.String()
	}

	respond.OK(c, resp)
}
