package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/requestmeta"
)

func TestRequestID_GeneratesUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	requestID := recorder.Header().Get(requestmeta.RequestIDHeader)
	if requestID == "" {
		t.Fatal("expected request id header to be set")
	}
	if len(requestID) != 36 {
		t.Errorf("expected UUID format request id, got %q", requestID)
	}
}

func TestRequestID_ReusesHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(requestmeta.RequestIDHeader, "client-provided-id")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	requestID := recorder.Header().Get(requestmeta.RequestIDHeader)
	if requestID != "client-provided-id" {
		t.Errorf("expected reused request id, got %q", requestID)
	}
}

func TestRequestLogger_LogsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	engine := gin.New()
	engine.Use(RequestID(), RequestLogger(logger))
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "http_request") {
		t.Errorf("expected log to contain 'http_request', got %q", logOutput)
	}
	if !strings.Contains(logOutput, "status=201") {
		t.Errorf("expected log to contain status 201, got %q", logOutput)
	}
}

func TestRequestLogger_OmitsQueryString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	engine := gin.New()
	engine.Use(RequestID(), RequestLogger(logger))
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test?api_key=sk-secret&foo=bar", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	logOutput := buf.String()
	if strings.Contains(logOutput, "sk-secret") {
		t.Errorf("expected query string to be omitted from log, got %q", logOutput)
	}
	if !strings.Contains(logOutput, `path=/test`) {
		t.Errorf("expected log to contain path /test, got %q", logOutput)
	}
}

func TestRecovery_HandlesPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	engine := gin.New()
	engine.Use(Recovery(logger), RequestID())
	engine.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}

	var body struct {
		Error struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if body.Error.Code != domain.ErrInternal.BizCode {
		t.Errorf("expected code %s, got %s", domain.ErrInternal.BizCode, body.Error.Code)
	}
	if body.Error.Message != domain.ErrInternal.Message {
		t.Errorf("expected message %s, got %s", domain.ErrInternal.Message, body.Error.Message)
	}
	if body.Error.RequestID == "" {
		t.Error("expected request_id in error response")
	}
}

func TestRecovery_CatchesPanicInLaterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	engine := gin.New()
	engine.Use(Recovery(logger), RequestID(), func(c *gin.Context) {
		panic("middleware panic")
	})
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	if !strings.Contains(buf.String(), "middleware panic") {
		t.Errorf("expected panic to be logged, got %q", buf.String())
	}
}
