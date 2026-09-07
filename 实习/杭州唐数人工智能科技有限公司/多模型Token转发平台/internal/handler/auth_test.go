package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockAuthService struct {
	login func(ctx context.Context, input service.AdminLoginInput) (service.AdminLoginResult, error)
}

func (m *mockAuthService) Login(ctx context.Context, input service.AdminLoginInput) (service.AdminLoginResult, error) {
	if m.login != nil {
		return m.login(ctx, input)
	}
	return service.AdminLoginResult{}, errors.New("not implemented")
}

func newAuthTestEngine(t *testing.T, svc AuthService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine := gin.New()
	engine.Use(middleware.Recovery(logger), middleware.RequestID(), middleware.RequestLogger(logger))

	handler := NewAuthHandler(svc)
	engine.POST("/api/v1/admin/login", handler.Login)
	return engine
}

func TestAuthHandler_Login_Success(t *testing.T) {
	svc := &mockAuthService{
		login: func(ctx context.Context, input service.AdminLoginInput) (service.AdminLoginResult, error) {
			return service.AdminLoginResult{
				AccessToken: "access-token",
				TokenType:   "Bearer",
				ExpiresIn:   86400,
			}, nil
		},
	}
	engine := newAuthTestEngine(t, svc)

	body := `{"username":"admin","password":"valid-password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data service.AdminLoginResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.AccessToken != "access-token" {
		t.Errorf("expected access-token, got %s", resp.Data.AccessToken)
	}
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	svc := &mockAuthService{}
	engine := newAuthTestEngine(t, svc)

	body := `{"username":"ab","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Error.Code)
	}
}

func TestAuthHandler_Login_Unauthorized(t *testing.T) {
	svc := &mockAuthService{
		login: func(ctx context.Context, input service.AdminLoginInput) (service.AdminLoginResult, error) {
			return service.AdminLoginResult{}, domain.ErrUnauthorized
		},
	}
	engine := newAuthTestEngine(t, svc)

	body := `{"username":"admin","password":"wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}
