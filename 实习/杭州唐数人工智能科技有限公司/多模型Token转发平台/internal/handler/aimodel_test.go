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

type mockAIModelService struct {
	createFunc func(ctx context.Context, input service.CreateAIModelInput) (*service.AIModelResponse, error)
	listFunc   func(ctx context.Context) ([]service.AIModelResponse, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*service.AIModelResponse, error)
	updateFunc func(ctx context.Context, id uuid.UUID, input service.UpdateAIModelInput) (*service.AIModelResponse, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockAIModelService) CreateAIModel(ctx context.Context, input service.CreateAIModelInput) (*service.AIModelResponse, error) {
	return m.createFunc(ctx, input)
}

func (m *mockAIModelService) ListAIModels(ctx context.Context) ([]service.AIModelResponse, error) {
	return m.listFunc(ctx)
}

func (m *mockAIModelService) GetAIModel(ctx context.Context, id uuid.UUID) (*service.AIModelResponse, error) {
	return m.getFunc(ctx, id)
}

func (m *mockAIModelService) UpdateAIModel(ctx context.Context, id uuid.UUID, input service.UpdateAIModelInput) (*service.AIModelResponse, error) {
	return m.updateFunc(ctx, id, input)
}

func (m *mockAIModelService) DeleteAIModel(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

func newAIModelEngine(t *testing.T, svc service.AIModelService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewAIModelHandler(svc)

	engine.POST("/api/v1/admin/models", h.Create)
	engine.GET("/api/v1/admin/models", h.List)
	engine.GET("/api/v1/admin/models/:id", h.Get)
	engine.PUT("/api/v1/admin/models/:id", h.Update)
	engine.DELETE("/api/v1/admin/models/:id", h.Delete)

	return engine
}

func TestAIModelHandler_Create(t *testing.T) {
	svc := &mockAIModelService{
		createFunc: func(ctx context.Context, input service.CreateAIModelInput) (*service.AIModelResponse, error) {
			if input.Name != "gpt-4o" {
				t.Errorf("expected name gpt-4o, got %s", input.Name)
			}
			return &service.AIModelResponse{ID: uuid.New().String(), Name: input.Name, UpstreamName: input.UpstreamName}, nil
		},
	}

	engine := newAIModelEngine(t, svc)
	body := `{"name":"gpt-4o","upstream_name":"gpt-4o-2024-08-06","input_price":2500,"output_price":10000}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/models", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAIModelHandler_GetNotFound(t *testing.T) {
	svc := &mockAIModelService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*service.AIModelResponse, error) {
			return nil, domain.ErrNotFound
		},
	}

	engine := newAIModelEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/models/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
