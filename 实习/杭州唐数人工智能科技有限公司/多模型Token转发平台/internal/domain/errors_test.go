package domain

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestIsAppError(t *testing.T) {
	if !IsAppError(ErrNotFound) {
		t.Error("expected ErrNotFound to be an AppError")
	}
	if IsAppError(errors.New("other")) {
		t.Error("expected non-AppError to be false")
	}
}

func TestAsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected *AppError
	}{
		{
			name:     "app error passes through",
			err:      ErrNotFound,
			expected: ErrNotFound,
		},
		{
			name:     "wrapped app error",
			err:      fmt.Errorf("wrap: %w", ErrNotFound),
			expected: ErrNotFound,
		},
		{
			name:     "unknown error",
			err:      errors.New("unknown"),
			expected: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AsAppError(tt.err)
			if got.BizCode != tt.expected.BizCode {
				t.Errorf("expected biz code %s, got %s", tt.expected.BizCode, got.BizCode)
			}
			if got.Code != tt.expected.Code {
				t.Errorf("expected http code %d, got %d", tt.expected.Code, got.Code)
			}
		})
	}
}

func TestAsAppError_UnknownKeepsCause(t *testing.T) {
	cause := errors.New("db connection refused")
	got := AsAppError(fmt.Errorf("query failed: %w", cause))

	if got == ErrInternal {
		t.Error("expected a new instance, not the shared ErrInternal")
	}
	if !errors.Is(got, cause) {
		t.Error("expected original error to be reachable via errors.Is")
	}
	if !errors.Is(got, ErrInternal) {
		t.Error("expected converted error to still match ErrInternal")
	}
}

func TestNewValidationError(t *testing.T) {
	details := map[string]string{"username": "不能为空", "age": "必须为正数"}
	err := NewValidationError(details)

	if err.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", err.Code)
	}
	if err.BizCode != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", err.BizCode)
	}
	if len(err.Details) != 2 || err.Details["username"] != "不能为空" {
		t.Errorf("unexpected details: %v", err.Details)
	}
	if errors.Is(err, ErrInvalidRequest) {
		// same HTTP code but different biz code: must NOT match
		t.Error("validation error should not match ErrInvalidRequest")
	}
}

func TestErrorCodes(t *testing.T) {
	if ErrInvalidRequest.Code != http.StatusBadRequest {
		t.Error("ErrInvalidRequest should be 400")
	}
	if ErrUnauthorized.Code != http.StatusUnauthorized {
		t.Error("ErrUnauthorized should be 401")
	}
	if ErrNotFound.Code != http.StatusNotFound {
		t.Error("ErrNotFound should be 404")
	}
	if ErrInternal.Code != http.StatusInternalServerError {
		t.Error("ErrInternal should be 500")
	}
}

func TestErrorsIs(t *testing.T) {
	if !errors.Is(fmt.Errorf("wrap: %w", ErrNotFound), ErrNotFound) {
		t.Error("expected errors.Is to find ErrNotFound after wrapping")
	}
	if errors.Is(ErrNotFound, ErrUnauthorized) {
		t.Error("expected different AppErrors not to match")
	}
}
