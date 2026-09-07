package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockPlatformService struct {
	createFunc func(ctx context.Context, input service.CreatePlatformInput) (*service.PlatformResponse, error)
	listFunc   func(ctx context.Context) ([]service.PlatformResponse, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*service.PlatformResponse, error)
	updateFunc func(ctx context.Context, id uuid.UUID, input service.UpdatePlatformInput) (*service.PlatformResponse, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockPlatformService) CreatePlatform(ctx context.Context, input service.CreatePlatformInput) (*service.PlatformResponse, error) {
	return m.createFunc(ctx, input)
}

func (m *mockPlatformService) ListPlatforms(ctx context.Context) ([]service.PlatformResponse, error) {
	return m.listFunc(ctx)
}

func (m *mockPlatformService) GetPlatform(ctx context.Context, id uuid.UUID) (*service.PlatformResponse, error) {
	return m.getFunc(ctx, id)
}

func (m *mockPlatformService) UpdatePlatform(ctx context.Context, id uuid.UUID, input service.UpdatePlatformInput) (*service.PlatformResponse, error) {
	return m.updateFunc(ctx, id, input)
}

func (m *mockPlatformService) DeletePlatform(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

func newPlatformEngine(t *testing.T, svc service.PlatformService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewPlatformHandler(svc)

	engine.POST("/api/v1/admin/platforms", h.Create)
	engine.GET("/api/v1/admin/platforms", h.List)
	engine.GET("/api/v1/admin/platforms/:id", h.Get)
	engine.PUT("/api/v1/admin/platforms/:id", h.Update)
	engine.DELETE("/api/v1/admin/platforms/:id", h.Delete)

	return engine
}

func TestPlatformHandler_Create(t *testing.T) {
	svc := &mockPlatformService{
		createFunc: func(ctx context.Context, input service.CreatePlatformInput) (*service.PlatformResponse, error) {
			if input.Name != "OpenAI" {
				t.Errorf("expected name OpenAI, got %s", input.Name)
			}
			return &service.PlatformResponse{ID: uuid.New().String(), Name: input.Name, Code: input.Code}, nil
		},
	}

	engine := newPlatformEngine(t, svc)
	body := `{"name":"OpenAI","code":"openai","base_url":"https://api.openai.com/v1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/platforms", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPlatformHandler_GetNotFound(t *testing.T) {
	svc := &mockPlatformService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*service.PlatformResponse, error) {
			return nil, domain.ErrNotFound
		},
	}

	engine := newPlatformEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/platforms/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPlatformHandler_CreateValidationError(t *testing.T) {
	svc := &mockPlatformService{}
	engine := newPlatformEngine(t, svc)

	body := `{"name":"","code":"","base_url":"not-a-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/platforms", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
