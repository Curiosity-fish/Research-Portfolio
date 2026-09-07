package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/service"
)

type mockAuditLogService struct {
	listFunc func(ctx context.Context, actorType *string, actorID *uuid.UUID, action *string, page, pageSize int) ([]service.AuditLogResponse, int, error)
}

func (m *mockAuditLogService) ListAuditLogsAdmin(ctx context.Context, actorType *string, actorID *uuid.UUID, action *string, page, pageSize int) ([]service.AuditLogResponse, int, error) {
	return m.listFunc(ctx, actorType, actorID, action, page, pageSize)
}

func newAuditLogEngine(t *testing.T, svc AuditLogService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewAuditLogHandler(svc)

	engine.GET("/api/v1/admin/audit-logs", handler.ListAdmin)

	return engine
}

func TestAuditLogHandler_ListAdmin(t *testing.T) {
	svc := &mockAuditLogService{
		listFunc: func(ctx context.Context, actorType *string, actorID *uuid.UUID, action *string, page, pageSize int) ([]service.AuditLogResponse, int, error) {
			return []service.AuditLogResponse{{ID: uuid.New().String(), Action: "GET /api/v1/admin/users"}}, 1, nil
		},
	}

	engine := newAuditLogEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs?actor_type=admin", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
