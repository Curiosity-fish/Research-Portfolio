package respond

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/requestmeta"
)

func TestOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	OK(c, map[string]string{"key": "value"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	expected := `{"data":{"key":"value"}}`
	if recorder.Body.String() != expected {
		t.Errorf("expected body %s, got %s", expected, recorder.Body.String())
	}
}

func TestCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	Created(c, map[string]string{"id": "123"})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}
}

func TestNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.DELETE("/resource", func(c *gin.Context) {
		NoContent(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/resource", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", recorder.Body.String())
	}
}

func TestError_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	Error(c, domain.ErrNotFound)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), domain.ErrNotFound.BizCode) {
		t.Errorf("expected body to contain %s, got %s", domain.ErrNotFound.BizCode, recorder.Body.String())
	}
}

func TestError_UnknownError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	Error(c, errors.New("unknown"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), domain.ErrInternal.BizCode) {
		t.Errorf("expected body to contain %s, got %s", domain.ErrInternal.BizCode, recorder.Body.String())
	}
	// The raw cause must never leak into the response body.
	if strings.Contains(recorder.Body.String(), "unknown") {
		t.Errorf("expected internal error detail to be hidden, got %s", recorder.Body.String())
	}
}

func TestError_IncludesRequestIDAndDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(requestmeta.RequestIDKey, "req-123")

	err := domain.NewValidationError(map[string]string{"username": "不能为空"})
	Error(c, err)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var body struct {
		Error struct {
			Code      string            `json:"code"`
			RequestID string            `json:"request_id"`
			Details   map[string]string `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", body.Error.Code)
	}
	if body.Error.RequestID != "req-123" {
		t.Errorf("expected request_id req-123, got %s", body.Error.RequestID)
	}
	if body.Error.Details["username"] != "不能为空" {
		t.Errorf("expected field detail, got %v", body.Error.Details)
	}
}

func TestError_OmitsRequestIDWhenAbsent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	Error(c, domain.ErrNotFound)

	if strings.Contains(recorder.Body.String(), "request_id") {
		t.Errorf("expected request_id to be omitted, got %s", recorder.Body.String())
	}
}

func TestErrorWithStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(requestmeta.RequestIDKey, "req-456")

	ErrorWithStatus(c, http.StatusBadRequest, "TEST_ERROR", "test message")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "TEST_ERROR") {
		t.Errorf("expected body to contain TEST_ERROR, got %s", body)
	}
	if !strings.Contains(body, "req-456") {
		t.Errorf("expected body to contain request_id, got %s", body)
	}
}
