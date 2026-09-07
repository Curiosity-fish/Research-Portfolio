package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// RelayHandler exposes OpenAI-compatible relay endpoints.
type RelayHandler struct {
	svc service.RelayService
}

// NewRelayHandler creates a new RelayHandler.
func NewRelayHandler(svc service.RelayService) *RelayHandler {
	return &RelayHandler{svc: svc}
}

// ChatCompletions handles POST /v1/chat/completions.
func (h *RelayHandler) ChatCompletions(c *gin.Context) {
	userID, tokenID, groupID, ok := extractRelayAuth(c)
	if !ok {
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respond.OpenAIError(c, 400, "invalid_request_error", "bad_request", "Failed to read request body", "")
		return
	}

	if isStreamRequest(body) {
		if err := h.svc.ChatCompletionsStream(c.Request.Context(), service.RelayInput{
			UserID:      userID,
			TokenID:     tokenID,
			GroupID:     groupID,
			RequestBody: body,
		}, c.Writer); err != nil {
			appErr := domain.AsAppError(err)
			respond.OpenAIErrorFromAppError(c, appErr)
		}
		return
	}

	result, err := h.svc.ChatCompletions(c.Request.Context(), service.RelayInput{
		UserID:      userID,
		TokenID:     tokenID,
		GroupID:     groupID,
		RequestBody: body,
	})
	if err != nil {
		appErr := domain.AsAppError(err)
		respond.OpenAIErrorFromAppError(c, appErr)
		return
	}

	copyHeaders(c.Writer.Header(), result.Header)
	c.Status(result.StatusCode)
	_, _ = c.Writer.Write(result.Body)
}

// Messages handles POST /v1/messages (Anthropic Messages API).
func (h *RelayHandler) Messages(c *gin.Context) {
	userID, tokenID, groupID, ok := extractRelayAuth(c)
	if !ok {
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respond.OpenAIError(c, 400, "invalid_request_error", "bad_request", "Failed to read request body", "")
		return
	}

	if isStreamRequest(body) {
		if err := h.svc.MessagesStream(c.Request.Context(), service.RelayInput{
			UserID:      userID,
			TokenID:     tokenID,
			GroupID:     groupID,
			RequestBody: body,
		}, c.Writer); err != nil {
			appErr := domain.AsAppError(err)
			respond.OpenAIErrorFromAppError(c, appErr)
		}
		return
	}

	result, err := h.svc.Messages(c.Request.Context(), service.RelayInput{
		UserID:      userID,
		TokenID:     tokenID,
		GroupID:     groupID,
		RequestBody: body,
	})
	if err != nil {
		appErr := domain.AsAppError(err)
		respond.OpenAIErrorFromAppError(c, appErr)
		return
	}

	copyHeaders(c.Writer.Header(), result.Header)
	c.Status(result.StatusCode)
	_, _ = c.Writer.Write(result.Body)
}

// extractRelayAuth reads API key identity from the gin context.
func extractRelayAuth(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.OpenAIError(c, 401, "invalid_request_error", "invalid_api_key", "Authentication required", "")
		return uuid.UUID{}, uuid.UUID{}, uuid.UUID{}, false
	}
	tokenID, ok := auth.TokenID(c)
	if !ok {
		respond.OpenAIError(c, 401, "invalid_request_error", "invalid_api_key", "Authentication required", "")
		return uuid.UUID{}, uuid.UUID{}, uuid.UUID{}, false
	}
	groupID, ok := auth.GroupID(c)
	if !ok {
		respond.OpenAIError(c, 401, "invalid_request_error", "invalid_api_key", "Authentication required", "")
		return uuid.UUID{}, uuid.UUID{}, uuid.UUID{}, false
	}
	return userID, tokenID, groupID, true
}

// ListModels handles GET /v1/models.
func (h *RelayHandler) ListModels(c *gin.Context) {
	groupID, ok := auth.GroupID(c)
	if !ok {
		respond.OpenAIError(c, 401, "invalid_request_error", "invalid_api_key", "Authentication required", "")
		return
	}

	models, err := h.svc.ListModels(c.Request.Context(), groupID)
	if err != nil {
		appErr := domain.AsAppError(err)
		respond.OpenAIErrorFromAppError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}

// isStreamRequest reports whether the incoming JSON body requests streaming.
func isStreamRequest(body []byte) bool {
	var req struct {
		Stream bool `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return false
	}
	return req.Stream
}

// copyHeaders copies selected upstream headers to the downstream response.
func copyHeaders(dst, src http.Header) {
	for _, key := range []string{"Content-Type", "Cache-Control", "X-Request-ID"} {
		if v := src.Get(key); v != "" {
			dst.Set(key, v)
		}
	}
}
