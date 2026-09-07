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

type mockGroupPlatformService struct {
	bindFunc   func(ctx context.Context, groupID, platformID uuid.UUID) error
	unbindFunc func(ctx context.Context, groupID, platformID uuid.UUID) error
	listFunc   func(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error)
}

func (m *mockGroupPlatformService) BindPlatform(ctx context.Context, groupID, platformID uuid.UUID) error {
	return m.bindFunc(ctx, groupID, platformID)
}

func (m *mockGroupPlatformService) UnbindPlatform(ctx context.Context, groupID, platformID uuid.UUID) error {
	return m.unbindFunc(ctx, groupID, platformID)
}

func (m *mockGroupPlatformService) ListPlatformsByGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	return m.listFunc(ctx, groupID)
}

func newGroupPlatformEngine(t *testing.T, svc service.GroupPlatformService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewGroupPlatformHandler(svc)

	engine.POST("/api/v1/admin/groups/:id/platforms", h.Bind)
	engine.GET("/api/v1/admin/groups/:id/platforms", h.List)
	engine.DELETE("/api/v1/admin/groups/:id/platforms/:platform_id", h.Unbind)

	return engine
}

func TestGroupPlatformHandler_Bind(t *testing.T) {
	groupID := uuid.New()
	platformID := uuid.New()

	svc := &mockGroupPlatformService{
		bindFunc: func(ctx context.Context, gid, pid uuid.UUID) error {
			if gid != groupID || pid != platformID {
				t.Errorf("unexpected ids: %s %s", gid, pid)
			}
			return nil
		},
	}

	engine := newGroupPlatformEngine(t, svc)
	body := `{"platform_id":"` + platformID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/groups/"+groupID.String()+"/platforms", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGroupPlatformHandler_BindGroupNotFound(t *testing.T) {
	svc := &mockGroupPlatformService{
		bindFunc: func(ctx context.Context, gid, pid uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	engine := newGroupPlatformEngine(t, svc)
	body := `{"platform_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/groups/"+uuid.New().String()+"/platforms", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
