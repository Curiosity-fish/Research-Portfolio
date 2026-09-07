package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/department"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockDepartmentService struct {
	createFunc func(ctx context.Context, input service.CreateDepartmentInput) (*service.DepartmentResponse, error)
	listFunc   func(ctx context.Context, status *department.Status) ([]service.DepartmentResponse, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*service.DepartmentResponse, error)
	updateFunc func(ctx context.Context, id uuid.UUID, input service.UpdateDepartmentInput) (*service.DepartmentResponse, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockDepartmentService) CreateDepartment(ctx context.Context, input service.CreateDepartmentInput) (*service.DepartmentResponse, error) {
	return m.createFunc(ctx, input)
}

func (m *mockDepartmentService) ListDepartments(ctx context.Context, status *department.Status) ([]service.DepartmentResponse, error) {
	return m.listFunc(ctx, status)
}

func (m *mockDepartmentService) GetDepartment(ctx context.Context, id uuid.UUID) (*service.DepartmentResponse, error) {
	return m.getFunc(ctx, id)
}

func (m *mockDepartmentService) UpdateDepartment(ctx context.Context, id uuid.UUID, input service.UpdateDepartmentInput) (*service.DepartmentResponse, error) {
	return m.updateFunc(ctx, id, input)
}

func (m *mockDepartmentService) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

func newDepartmentEngine(t *testing.T, svc DepartmentService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewDepartmentHandler(svc)

	engine.POST("/api/v1/admin/departments", h.Create)
	engine.GET("/api/v1/admin/departments", h.List)
	engine.GET("/api/v1/admin/departments/:id", h.Get)
	engine.PUT("/api/v1/admin/departments/:id", h.Update)
	engine.DELETE("/api/v1/admin/departments/:id", h.Delete)

	return engine
}

func TestDepartmentHandler_Create(t *testing.T) {
	svc := &mockDepartmentService{
		createFunc: func(ctx context.Context, input service.CreateDepartmentInput) (*service.DepartmentResponse, error) {
			if input.Name != "教务处" {
				t.Errorf("expected name 教务处, got %s", input.Name)
			}
			return &service.DepartmentResponse{ID: uuid.New().String(), Name: input.Name, Code: input.Code}, nil
		},
	}

	engine := newDepartmentEngine(t, svc)
	body := `{"name":"教务处","code":"jwc","sort_order":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/departments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentHandler_CreateValidationError(t *testing.T) {
	svc := &mockDepartmentService{}
	engine := newDepartmentEngine(t, svc)
	body := `{"name":"","code":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/departments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentHandler_List(t *testing.T) {
	svc := &mockDepartmentService{
		listFunc: func(ctx context.Context, status *department.Status) ([]service.DepartmentResponse, error) {
			return []service.DepartmentResponse{{ID: uuid.New().String(), Name: "教务处"}}, nil
		},
	}

	engine := newDepartmentEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/departments?status=active", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentHandler_Get(t *testing.T) {
	id := uuid.New()
	svc := &mockDepartmentService{
		getFunc: func(ctx context.Context, did uuid.UUID) (*service.DepartmentResponse, error) {
			if did != id {
				t.Errorf("expected id %s, got %s", id, did)
			}
			return &service.DepartmentResponse{ID: id.String(), Name: "教务处"}, nil
		},
	}

	engine := newDepartmentEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/departments/"+id.String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentHandler_GetNotFound(t *testing.T) {
	svc := &mockDepartmentService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*service.DepartmentResponse, error) {
			return nil, domain.ErrNotFound
		},
	}

	engine := newDepartmentEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/departments/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentHandler_Update(t *testing.T) {
	id := uuid.New()
	newName := "Updated"
	svc := &mockDepartmentService{
		updateFunc: func(ctx context.Context, did uuid.UUID, input service.UpdateDepartmentInput) (*service.DepartmentResponse, error) {
			if did != id {
				t.Errorf("expected id %s, got %s", id, did)
			}
			if input.Name == nil || *input.Name != newName {
				t.Errorf("expected name %q, got %v", newName, input.Name)
			}
			return &service.DepartmentResponse{ID: id.String(), Name: newName}, nil
		},
	}

	engine := newDepartmentEngine(t, svc)
	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/departments/"+id.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentHandler_Delete(t *testing.T) {
	id := uuid.New()
	svc := &mockDepartmentService{
		deleteFunc: func(ctx context.Context, did uuid.UUID) error {
			if did != id {
				t.Errorf("expected id %s, got %s", id, did)
			}
			return nil
		},
	}

	engine := newDepartmentEngine(t, svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/departments/"+id.String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentHandler_ServiceError(t *testing.T) {
	svc := &mockDepartmentService{
		createFunc: func(ctx context.Context, input service.CreateDepartmentInput) (*service.DepartmentResponse, error) {
			return nil, errors.New("boom")
		},
	}

	engine := newDepartmentEngine(t, svc)
	body := `{"name":"教务处","code":"jwc"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/departments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
