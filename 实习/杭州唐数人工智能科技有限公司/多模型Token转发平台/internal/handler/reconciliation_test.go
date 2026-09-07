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

type mockReconciliationService struct {
	getFunc   func(ctx context.Context, orderID uuid.UUID) (*service.ReconciliationOrderResponse, error)
	refundFunc func(ctx context.Context, adminID, orderID uuid.UUID, amount int64, reason *string) (*service.RefundOrderResponse, error)
}

func (m *mockReconciliationService) GetOrderReconciliation(ctx context.Context, orderID uuid.UUID) (*service.ReconciliationOrderResponse, error) {
	return m.getFunc(ctx, orderID)
}

func (m *mockReconciliationService) RefundOrder(ctx context.Context, adminID, orderID uuid.UUID, amount int64, reason *string) (*service.RefundOrderResponse, error) {
	return m.refundFunc(ctx, adminID, orderID, amount, reason)
}

func newReconciliationEngine(t *testing.T, svc ReconciliationService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewReconciliationHandler(svc)

	adminAuth := func(c *gin.Context) {
		auth.SetAdminContext(c, &auth.Claims{AdminID: uuid.MustParse("22222222-2222-2222-2222-222222222222")})
		c.Next()
	}

	engine.GET("/api/v1/admin/recharge-orders/:id/reconciliation", handler.GetReconciliation)
	engine.POST("/api/v1/admin/recharge-orders/:id/refunds", adminAuth, handler.Refund)

	return engine
}

func TestReconciliationHandler_GetReconciliation(t *testing.T) {
	orderID := uuid.New()
	svc := &mockReconciliationService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*service.ReconciliationOrderResponse, error) {
			return &service.ReconciliationOrderResponse{ID: id.String(), Amount: 1000, RefundedAmount: 0}, nil
		},
	}

	engine := newReconciliationEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/recharge-orders/"+orderID.String()+"/reconciliation", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestReconciliationHandler_Refund(t *testing.T) {
	orderID := uuid.New()
	svc := &mockReconciliationService{
		refundFunc: func(ctx context.Context, adminID, id uuid.UUID, amount int64, reason *string) (*service.RefundOrderResponse, error) {
			return &service.RefundOrderResponse{
				Order:         service.ReconciliationOrderResponse{ID: id.String(), RefundedAmount: amount},
				BalanceRecord: service.BalanceRecordResponse{ID: uuid.New().String(), Amount: amount},
			}, nil
		},
	}

	engine := newReconciliationEngine(t, svc)
	body := `{"amount":300,"reason":"测试退款"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/recharge-orders/"+orderID.String()+"/refunds", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
