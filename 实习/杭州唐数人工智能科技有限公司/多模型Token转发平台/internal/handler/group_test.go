package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockGroupService struct {
	listFunc func(ctx context.Context) ([]service.GroupResponse, error)
}

func (m *mockGroupService) ListGroups(ctx context.Context) ([]service.GroupResponse, error) {
	return m.listFunc(ctx)
}

func newGroupEngine(t *testing.T, svc GroupService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewGroupHandler(svc)

	engine.GET("/api/v1/admin/groups", h.List)

	return engine
}

func TestGroupHandler_List(t *testing.T) {
	desc := "测试分组"
	svc := &mockGroupService{
		listFunc: func(ctx context.Context) ([]service.GroupResponse, error) {
			return []service.GroupResponse{
				{ID: "g1", Name: "默认分组", Code: "default", Description: &desc, SortOrder: 0, Status: "active"},
			}, nil
		},
	}

	engine := newGroupEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/groups", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGroupHandler_ListError(t *testing.T) {
	svc := &mockGroupService{
		listFunc: func(ctx context.Context) ([]service.GroupResponse, error) {
			return nil, domain.WrapInternal(errors.New("db down"))
		},
	}

	engine := newGroupEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/groups", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
