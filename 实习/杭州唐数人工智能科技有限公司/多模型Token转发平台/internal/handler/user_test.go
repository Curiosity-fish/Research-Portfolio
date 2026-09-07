package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockUserService struct {
	createFunc       func(ctx context.Context, input service.CreateUserInput) (*service.UserResponse, error)
	listFunc         func(ctx context.Context, filter service.ListUsersFilter) (*service.ListUsersResponse, error)
	getFunc          func(ctx context.Context, id uuid.UUID) (*service.UserResponse, error)
	updateFunc       func(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*service.UserResponse, error)
	updateStatusFunc func(ctx context.Context, id uuid.UUID, status user.Status) (*service.UserResponse, error)
	bulkImportFunc   func(ctx context.Context, input service.BulkImportInput) (*service.BulkImportResult, error)
}

func (m *mockUserService) CreateUser(ctx context.Context, input service.CreateUserInput) (*service.UserResponse, error) {
	return m.createFunc(ctx, input)
}

func (m *mockUserService) ListUsers(ctx context.Context, filter service.ListUsersFilter) (*service.ListUsersResponse, error) {
	return m.listFunc(ctx, filter)
}

func (m *mockUserService) GetUser(ctx context.Context, id uuid.UUID) (*service.UserResponse, error) {
	return m.getFunc(ctx, id)
}

func (m *mockUserService) UpdateUser(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*service.UserResponse, error) {
	return m.updateFunc(ctx, id, input)
}

func (m *mockUserService) UpdateUserStatus(ctx context.Context, id uuid.UUID, status user.Status) (*service.UserResponse, error) {
	return m.updateStatusFunc(ctx, id, status)
}

func (m *mockUserService) BulkImport(ctx context.Context, input service.BulkImportInput) (*service.BulkImportResult, error) {
	return m.bulkImportFunc(ctx, input)
}

func newUserEngine(t *testing.T, svc UserService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewUserHandler(svc)

	engine.POST("/api/v1/admin/users", handler.Create)
	engine.GET("/api/v1/admin/users", handler.List)
	engine.GET("/api/v1/admin/users/:id", handler.Get)
	engine.PUT("/api/v1/admin/users/:id", handler.Update)
	engine.PATCH("/api/v1/admin/users/:id/status", handler.UpdateStatus)
	// Static segments alongside the /users/:id wildcard: registering them here
	// proves the router accepts both without panicking.
	engine.POST("/api/v1/admin/users/bulk-import", handler.BulkImport)
	engine.POST("/api/v1/admin/users/bulk-import-file", handler.BulkImportFile)

	return engine
}

func TestUserHandler_Create(t *testing.T) {
	svc := &mockUserService{
		createFunc: func(ctx context.Context, input service.CreateUserInput) (*service.UserResponse, error) {
			if input.Username != "newuser" {
				t.Errorf("expected username newuser, got %s", input.Username)
			}
			return &service.UserResponse{ID: uuid.New().String(), Username: input.Username, Name: input.Name, Status: "active"}, nil
		},
	}

	engine := newUserEngine(t, svc)
	body := `{"username":"newuser","password":"password123","name":"New User","email":"new@example.com","role":"teacher"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_CreateValidationError(t *testing.T) {
	svc := &mockUserService{}
	engine := newUserEngine(t, svc)

	body := `{"username":"ab","password":"short","name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_List(t *testing.T) {
	svc := &mockUserService{
		listFunc: func(ctx context.Context, filter service.ListUsersFilter) (*service.ListUsersResponse, error) {
			return &service.ListUsersResponse{
				Total:    1,
				Page:     1,
				PageSize: 20,
				List:     []service.UserResponse{{ID: uuid.New().String(), Username: "alice"}},
			}, nil
		},
	}

	engine := newUserEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?keyword=alice&status=active", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_Get(t *testing.T) {
	id := uuid.New()
	svc := &mockUserService{
		getFunc: func(ctx context.Context, uid uuid.UUID) (*service.UserResponse, error) {
			if uid != id {
				t.Errorf("expected id %s, got %s", id, uid)
			}
			return &service.UserResponse{ID: id.String(), Username: "alice"}, nil
		},
	}

	engine := newUserEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/"+id.String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_GetNotFound(t *testing.T) {
	svc := &mockUserService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*service.UserResponse, error) {
			return nil, domain.ErrNotFound
		},
	}

	engine := newUserEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_Update(t *testing.T) {
	id := uuid.New()
	newName := "Updated"
	svc := &mockUserService{
		updateFunc: func(ctx context.Context, uid uuid.UUID, input service.UpdateUserInput) (*service.UserResponse, error) {
			if uid != id {
				t.Errorf("expected id %s, got %s", id, uid)
			}
			if input.Name == nil || *input.Name != newName {
				t.Errorf("expected name %q, got %v", newName, input.Name)
			}
			return &service.UserResponse{ID: id.String(), Name: newName}, nil
		},
	}

	engine := newUserEngine(t, svc)
	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/"+id.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_UpdateStatus(t *testing.T) {
	id := uuid.New()
	svc := &mockUserService{
		updateStatusFunc: func(ctx context.Context, uid uuid.UUID, status user.Status) (*service.UserResponse, error) {
			if uid != id {
				t.Errorf("expected id %s, got %s", id, uid)
			}
			if status != user.StatusInactive {
				t.Errorf("expected inactive, got %s", status)
			}
			return &service.UserResponse{ID: id.String(), Status: string(status)}, nil
		},
	}

	engine := newUserEngine(t, svc)
	body := `{"status":"inactive"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+id.String()+"/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_ServiceError(t *testing.T) {
	svc := &mockUserService{
		createFunc: func(ctx context.Context, input service.CreateUserInput) (*service.UserResponse, error) {
			return nil, errors.New("boom")
		},
	}

	engine := newUserEngine(t, svc)
	body := `{"username":"newuser","password":"password123","name":"New User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func decodeUserResponse(t *testing.T, body []byte) service.UserResponse {
	t.Helper()
	var wrapper struct {
		Data service.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return wrapper.Data
}
