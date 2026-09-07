package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockUserTokenService struct {
	createFunc       func(ctx context.Context, input service.CreateUserTokenInput) (*service.UserTokenResponse, string, error)
	listFunc         func(ctx context.Context, userID uuid.UUID) ([]service.UserTokenResponse, error)
	deleteFunc       func(ctx context.Context, userID, tokenID uuid.UUID) error
	updateStatusFunc func(ctx context.Context, userID, tokenID uuid.UUID, enabled bool) (*service.UserTokenResponse, error)
}

func (m *mockUserTokenService) CreateUserToken(ctx context.Context, input service.CreateUserTokenInput) (*service.UserTokenResponse, string, error) {
	return m.createFunc(ctx, input)
}

func (m *mockUserTokenService) ListUserTokens(ctx context.Context, userID uuid.UUID) ([]service.UserTokenResponse, error) {
	return m.listFunc(ctx, userID)
}

func (m *mockUserTokenService) DeleteUserToken(ctx context.Context, userID, tokenID uuid.UUID) error {
	return m.deleteFunc(ctx, userID, tokenID)
}

func (m *mockUserTokenService) UpdateUserTokenStatus(ctx context.Context, userID, tokenID uuid.UUID, enabled bool) (*service.UserTokenResponse, error) {
	return m.updateStatusFunc(ctx, userID, tokenID, enabled)
}

func newUserTokenEngine(t *testing.T, svc UserTokenService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewUserTokenHandler(svc)

	engine.POST("/api/v1/admin/users/:id/tokens", h.Create)
	engine.GET("/api/v1/admin/users/:id/tokens", h.List)
	engine.DELETE("/api/v1/admin/users/:id/tokens/:token_id", h.Delete)
	engine.PATCH("/api/v1/admin/users/:id/tokens/:token_id/status", h.UpdateStatus)

	return engine
}

func TestUserTokenHandler_Create(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserTokenService{
		createFunc: func(ctx context.Context, input service.CreateUserTokenInput) (*service.UserTokenResponse, string, error) {
			if input.UserID != userID {
				t.Errorf("expected user id %s, got %s", userID, input.UserID)
			}
			if input.Name != "default" {
				t.Errorf("expected name default, got %s", input.Name)
			}
			return &service.UserTokenResponse{ID: uuid.New().String(), Name: input.Name}, "sk-xxxx", nil
		},
	}

	engine := newUserTokenEngine(t, svc)
	body := `{"name":"default","quota_limit":1000,"expires_at":"2026-12-31T23:59:59Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+userID.String()+"/tokens", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserTokenHandler_CreateValidationError(t *testing.T) {
	svc := &mockUserTokenService{}
	engine := newUserTokenEngine(t, svc)

	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+uuid.New().String()+"/tokens", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserTokenHandler_List(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserTokenService{
		listFunc: func(ctx context.Context, uid uuid.UUID) ([]service.UserTokenResponse, error) {
			if uid != userID {
				t.Errorf("expected user id %s, got %s", userID, uid)
			}
			return []service.UserTokenResponse{{ID: uuid.New().String(), Name: "default"}}, nil
		},
	}

	engine := newUserTokenEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/"+userID.String()+"/tokens", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserTokenHandler_Delete(t *testing.T) {
	userID := uuid.New()
	tokenID := uuid.New()
	svc := &mockUserTokenService{
		deleteFunc: func(ctx context.Context, uid, tid uuid.UUID) error {
			if uid != userID || tid != tokenID {
				t.Errorf("expected user %s token %s, got %s %s", userID, tokenID, uid, tid)
			}
			return nil
		},
	}

	engine := newUserTokenEngine(t, svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/"+userID.String()+"/tokens/"+tokenID.String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserTokenHandler_UpdateStatus(t *testing.T) {
	userID := uuid.New()
	tokenID := uuid.New()
	svc := &mockUserTokenService{
		updateStatusFunc: func(ctx context.Context, uid, tid uuid.UUID, enabled bool) (*service.UserTokenResponse, error) {
			if uid != userID || tid != tokenID {
				t.Errorf("expected user %s token %s, got %s %s", userID, tokenID, uid, tid)
			}
			if enabled {
				t.Error("expected disabled")
			}
			return &service.UserTokenResponse{ID: tid.String(), IsEnabled: enabled}, nil
		},
	}

	engine := newUserTokenEngine(t, svc)
	body := `{"is_enabled":false}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+userID.String()+"/tokens/"+tokenID.String()+"/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserTokenHandler_NotFound(t *testing.T) {
	svc := &mockUserTokenService{
		listFunc: func(ctx context.Context, userID uuid.UUID) ([]service.UserTokenResponse, error) {
			return nil, domain.ErrNotFound
		},
	}

	engine := newUserTokenEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/"+uuid.New().String()+"/tokens", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserTokenHandler_InvalidUserID(t *testing.T) {
	svc := &mockUserTokenService{}
	engine := newUserTokenEngine(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/not-a-uuid/tokens", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// satisfy unused imports in case test evolves
var _ = time.Now
var _ = errors.New
