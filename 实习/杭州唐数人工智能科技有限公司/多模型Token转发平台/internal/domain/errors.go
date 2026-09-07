package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is the standard application error used across the codebase.
// It carries an HTTP status code, a machine-readable business code, and a
// human-readable message. Details optionally carries field-level validation
// errors; cause keeps the original error for logging and errors.Is/As
// traversal without exposing it in responses.
type AppError struct {
	Code    int               // HTTP status code
	BizCode string            // Business error code
	Message string            // User-facing message
	Details map[string]string // Optional field-level error details

	cause error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.BizCode, e.Message)
}

// Unwrap returns the original error that caused this AppError, if any.
func (e *AppError) Unwrap() error {
	return e.cause
}

// Is reports whether target is an equivalent AppError.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code && e.BizCode == t.BizCode
}

// NewAppError creates a new AppError.
func NewAppError(code int, bizCode, message string) *AppError {
	return &AppError{
		Code:    code,
		BizCode: bizCode,
		Message: message,
	}
}

// NewValidationError creates a 400 AppError carrying field-level details.
func NewValidationError(details map[string]string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		BizCode: "VALIDATION_ERROR",
		Message: "请求参数校验失败",
		Details: details,
	}
}

// WrapInternal creates an internal AppError that keeps err as its cause.
// The cause is available via Unwrap for logging but never serialized.
func WrapInternal(err error) *AppError {
	return &AppError{
		Code:    ErrInternal.Code,
		BizCode: ErrInternal.BizCode,
		Message: ErrInternal.Message,
		cause:   err,
	}
}

// Predefined application errors.
var (
	ErrInvalidRequest      = NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "请求参数无效")
	ErrUnauthorized        = NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "未授权")
	ErrForbidden           = NewAppError(http.StatusForbidden, "FORBIDDEN", "禁止访问")
	ErrNotFound            = NewAppError(http.StatusNotFound, "NOT_FOUND", "资源不存在")
	ErrConflict            = NewAppError(http.StatusConflict, "CONFLICT", "资源冲突")
	ErrTooMany             = NewAppError(http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "请求过于频繁")
	ErrInternal            = NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误")
	ErrUnavailable         = NewAppError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "服务不可用")
	ErrInsufficientQuota   = NewAppError(http.StatusPaymentRequired, "INSUFFICIENT_QUOTA", "配额不足")
	ErrInsufficientBalance = NewAppError(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "余额不足")
	ErrFeatureDisabled     = NewAppError(http.StatusForbidden, "FEATURE_DISABLED", "当前功能模式已禁用此操作")
)

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to AppError. Unknown errors become a new
// internal error instance that retains the original error as its cause, so
// callers can still log the root cause after conversion.
func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return WrapInternal(err)
}
