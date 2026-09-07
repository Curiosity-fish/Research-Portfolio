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

type mockAccountService struct {
	createFunc func(ctx context.Context, input service.CreateAccountInput) (*service.AccountResponse, error)
	listFunc   func(ctx context.Context) ([]service.AccountResponse, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*service.AccountResponse, error)
	updateFunc func(ctx context.Context, id uuid.UUID, input service.UpdateAccountInput) (*service.AccountResponse, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockAccountService) CreateAccount(ctx context.Context, input service.CreateAccountInput) (*service.AccountResponse, error) {
	return m.createFunc(ctx, input)
}

func (m *mockAccountService) ListAccounts(ctx context.Context) ([]service.AccountResponse, error) {
	return m.listFunc(ctx)
}

func (m *mockAccountService) ListAccountsByPlatform(ctx context.Context, platformID uuid.UUID) ([]service.AccountResponse, error) {
	return nil, nil
}

func (m *mockAccountService) GetAccount(ctx context.Context, id uuid.UUID) (*service.AccountResponse, error) {
	return m.getFunc(ctx, id)
}

func (m *mockAccountService) UpdateAccount(ctx context.Context, id uuid.UUID, input service.UpdateAccountInput) (*service.AccountResponse, error) {
	return m.updateFunc(ctx, id, input)
}

func (m *mockAccountService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

func (m *mockAccountService) DecryptAPIKey(ctx context.Context, accountID uuid.UUID) (string, error) {
	return "", nil
}

func newAccountEngine(t *testing.T, svc service.AccountService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewAccountHandler(svc)

	engine.POST("/api/v1/admin/accounts", h.Create)
	engine.GET("/api/v1/admin/accounts", h.List)
	engine.GET("/api/v1/admin/accounts/:id", h.Get)
	engine.PUT("/api/v1/admin/accounts/:id", h.Update)
	engine.DELETE("/api/v1/admin/accounts/:id", h.Delete)

	return engine
}

func TestAccountHandler_Create(t *testing.T) {
	platformID := uuid.New()
	svc := &mockAccountService{
		createFunc: func(ctx context.Context, input service.CreateAccountInput) (*service.AccountResponse, error) {
			if input.PlatformID != platformID {
				t.Errorf("expected platform id %s, got %s", platformID, input.PlatformID)
			}
			return &service.AccountResponse{ID: uuid.New().String(), PlatformID: platformID.String(), Name: input.Name}, nil
		},
	}

	engine := newAccountEngine(t, svc)
	body := `{"platform_id":"` + platformID.String() + `","name":"Primary","api_key":"sk-test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAccountHandler_GetNotFound(t *testing.T) {
	svc := &mockAccountService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*service.AccountResponse, error) {
			return nil, domain.ErrNotFound
		},
	}

	engine := newAccountEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
