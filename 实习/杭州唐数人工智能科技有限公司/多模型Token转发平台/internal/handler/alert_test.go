package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockAlertService struct {
	createRuleFunc     func(ctx context.Context, input service.CreateAlertRuleInput) (*service.AlertRuleResponse, error)
	getRuleFunc        func(ctx context.Context, id uuid.UUID) (*service.AlertRuleResponse, error)
	listRulesFunc      func(ctx context.Context, enabledOnly bool, page, pageSize int) ([]service.AlertRuleResponse, int, error)
	updateRuleFunc     func(ctx context.Context, id uuid.UUID, input service.UpdateAlertRuleInput) (*service.AlertRuleResponse, error)
	deleteRuleFunc     func(ctx context.Context, id uuid.UUID) error
	listRecordsFunc    func(ctx context.Context, ruleID *uuid.UUID, isResolved *bool, page, pageSize int) ([]service.AlertRecordResponse, int, error)
	resolveRecordFunc  func(ctx context.Context, id, adminID uuid.UUID) (*service.AlertRecordResponse, error)
	evaluateRulesFunc  func(ctx context.Context) (int, error)
}

func (m *mockAlertService) CreateAlertRule(ctx context.Context, input service.CreateAlertRuleInput) (*service.AlertRuleResponse, error) {
	return m.createRuleFunc(ctx, input)
}

func (m *mockAlertService) GetAlertRule(ctx context.Context, id uuid.UUID) (*service.AlertRuleResponse, error) {
	return m.getRuleFunc(ctx, id)
}

func (m *mockAlertService) ListAlertRules(ctx context.Context, enabledOnly bool, page, pageSize int) ([]service.AlertRuleResponse, int, error) {
	return m.listRulesFunc(ctx, enabledOnly, page, pageSize)
}

func (m *mockAlertService) UpdateAlertRule(ctx context.Context, id uuid.UUID, input service.UpdateAlertRuleInput) (*service.AlertRuleResponse, error) {
	return m.updateRuleFunc(ctx, id, input)
}

func (m *mockAlertService) DeleteAlertRule(ctx context.Context, id uuid.UUID) error {
	return m.deleteRuleFunc(ctx, id)
}

func (m *mockAlertService) ListAlertRecords(ctx context.Context, ruleID *uuid.UUID, isResolved *bool, page, pageSize int) ([]service.AlertRecordResponse, int, error) {
	return m.listRecordsFunc(ctx, ruleID, isResolved, page, pageSize)
}

func (m *mockAlertService) ResolveAlertRecord(ctx context.Context, id, adminID uuid.UUID) (*service.AlertRecordResponse, error) {
	return m.resolveRecordFunc(ctx, id, adminID)
}

func (m *mockAlertService) EvaluateRules(ctx context.Context) (int, error) {
	return m.evaluateRulesFunc(ctx)
}

func newAlertEngine(t *testing.T, svc AlertService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewAlertHandler(svc)

	adminAuth := func(c *gin.Context) {
		auth.SetAdminContext(c, &auth.Claims{AdminID: uuid.MustParse("22222222-2222-2222-2222-222222222222")})
		c.Next()
	}

	engine.POST("/api/v1/admin/alert-rules", handler.CreateRule)
	engine.GET("/api/v1/admin/alert-rules", handler.ListRules)
	engine.GET("/api/v1/admin/alert-rules/:id", handler.GetRule)
	engine.PUT("/api/v1/admin/alert-rules/:id", handler.UpdateRule)
	engine.DELETE("/api/v1/admin/alert-rules/:id", handler.DeleteRule)
	engine.POST("/api/v1/admin/alerts/evaluate", handler.Evaluate)
	engine.GET("/api/v1/admin/alerts", handler.ListRecords)
	engine.PATCH("/api/v1/admin/alerts/:id/resolve", adminAuth, handler.ResolveRecord)

	return engine
}

func TestAlertHandler_CreateRule(t *testing.T) {
	svc := &mockAlertService{
		createRuleFunc: func(ctx context.Context, input service.CreateAlertRuleInput) (*service.AlertRuleResponse, error) {
			return &service.AlertRuleResponse{ID: uuid.New().String(), Name: input.Name, Metric: input.Metric}, nil
		},
	}

	engine := newAlertEngine(t, svc)
	body := `{"name":"余额告警","metric":"balance_low","threshold":100,"enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/alert-rules", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAlertHandler_Evaluate(t *testing.T) {
	svc := &mockAlertService{
		evaluateRulesFunc: func(ctx context.Context) (int, error) {
			return 3, nil
		},
	}

	engine := newAlertEngine(t, svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/alerts/evaluate", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAlertHandler_ResolveRecord(t *testing.T) {
	svc := &mockAlertService{
		resolveRecordFunc: func(ctx context.Context, id, adminID uuid.UUID) (*service.AlertRecordResponse, error) {
			return &service.AlertRecordResponse{ID: id.String(), IsResolved: true}, nil
		},
	}

	engine := newAlertEngine(t, svc)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/alerts/"+uuid.New().String()+"/resolve", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
