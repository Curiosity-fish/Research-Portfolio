package respond

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/requestmeta"
)

// OpenAIErrorObject is the error object used in OpenAI-compatible responses.
type OpenAIErrorObject struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// OpenAIErrorResponse is the top-level error envelope returned by /v1 routes.
type OpenAIErrorResponse struct {
	Error OpenAIErrorObject `json:"error"`
}

// OpenAIError writes an OpenAI-compatible error response and aborts the handler chain.
func OpenAIError(c *gin.Context, status int, errType, code, message, param string) {
	resp := OpenAIErrorResponse{
		Error: OpenAIErrorObject{
			Message: message,
			Type:    errType,
			Param:   param,
			Code:    code,
		},
	}
	c.Abort()
	c.JSON(status, resp)
}

// OpenAIErrorFromAppError converts a domain.AppError to an OpenAI-compatible response.
// It maps HTTP status ranges to OpenAI error types:
//   - 402 -> insufficient_quota
//   - 4xx -> invalid_request_error
//   - 5xx -> api_error
func OpenAIErrorFromAppError(c *gin.Context, appErr *domain.AppError) {
	errType := "api_error"
	switch {
	case appErr.Code == http.StatusPaymentRequired:
		errType = "insufficient_quota"
	case appErr.Code >= 400 && appErr.Code < 500:
		errType = "invalid_request_error"
	}

	// Add request_id to response headers so clients can quote it for support.
	c.Header("X-Request-ID", requestmeta.RequestID(c))

	OpenAIError(c, appErr.Code, errType, appErr.BizCode, appErr.Message, "")
}
