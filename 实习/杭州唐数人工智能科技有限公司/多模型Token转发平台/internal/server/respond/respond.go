package respond

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/requestmeta"
)

// ErrorResponse is the unified error response format returned by all endpoints.
type ErrorResponse struct {
	Error struct {
		Code      string            `json:"code"`
		Message   string            `json:"message"`
		RequestID string            `json:"request_id,omitempty"`
		Details   map[string]string `json:"details,omitempty"`
	} `json:"error"`
}

// OK writes a 200 JSON response.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Created writes a 201 JSON response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

// NoContent writes a 204 response with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error writes an error response using the standard format.
// If err is an AppError, its code, message and details are used; otherwise a
// generic internal error is returned. The request ID is included whenever the
// RequestID middleware has set one, so clients can quote it for support.
// The handler chain is aborted before writing; the JSON body is still
// written. Callers should still return immediately (Abort does not stop the
// current handler's own code path).
func Error(c *gin.Context, err error) {
	appErr := domain.AsAppError(err)
	var resp ErrorResponse
	resp.Error.Code = appErr.BizCode
	resp.Error.Message = appErr.Message
	resp.Error.RequestID = requestmeta.RequestID(c)
	resp.Error.Details = appErr.Details
	c.Abort()
	c.JSON(appErr.Code, resp)
}

// ErrorWithStatus writes an error response with an explicit HTTP status.
func ErrorWithStatus(c *gin.Context, status int, bizCode, message string) {
	var resp ErrorResponse
	resp.Error.Code = bizCode
	resp.Error.Message = message
	resp.Error.RequestID = requestmeta.RequestID(c)
	c.Abort()
	c.JSON(status, resp)
}
