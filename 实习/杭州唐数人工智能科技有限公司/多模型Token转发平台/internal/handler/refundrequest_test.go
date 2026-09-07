package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockRefundRequestService struct {
	createFunc  func(ctx context.Context, input service.CreateRefundRequestInput) (*service.RefundRequestResponse, error)
	listMyFunc  func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.RefundRequestResponse, int, error)
	cancelFunc  func(ctx context.Context, userID, requestID uuid.UUID) error
	listFunc    func(ctx context.Context, filter service.ListRefundRequestFilter) ([]service.RefundRequestResponse, int, error)
	approveFunc func(ctx context.Context, requestID, adminID uuid.UUID) (*service.RefundRequestResponse, error)
	rejectFunc  func(ctx context.Context, requestID, adminID uuid.UUID) (*service.RefundRequestResponse, error)
}

func (m *mockRefundRequestService) CreateRefundRequest(ctx context.Context, input service.CreateRefundRequestInput) (*service.RefundRequestResponse, error) {
	return m.createFunc(ctx, input)
}

func (m *mockRefundRequestService) ListMyRefundRequests(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.RefundRequestResponse, int, error) {
	return m.listMyFunc(ctx, userID, page, pageSize)
}

func (m *mockRefundRequestService) CancelRefundRequest(ctx context.Context, userID, requestID uuid.UUID) error {
	return m.cancelFunc(ctx, userID, requestID)
}

func (m *mockRefundRequestService) ListRefundRequests(ctx context.Context, filter service.ListRefundRequestFilter) ([]service.RefundRequestResponse, int, error) {
	return m.listFunc(ctx, filter)
}

func (m *mockRefundRequestService) ApproveRefundRequest(ctx context.Context, requestID, adminID uuid.UUID) (*service.RefundRequestResponse, error) {
	return m.approveFunc(ctx, requestID, adminID)
}

func (m *mockRefundRequestService) RejectRefundRequest(ctx context.Context, requestID, adminID uuid.UUID) (*service.RefundRequestResponse, error) {
	return m.rejectFunc(ctx, requestID, adminID)
}

var (
	refundTestUserID  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	refundTestAdminID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func newRefundRequestEngine(t *testing.T, svc RefundRequestService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewRefundRequestHandler(svc)

	userAuth := func(c *gin.Context) {
		auth.SetUserIDContext(c, refundTestUserID)
		c.Next()
	}
	adminAuth := func(c *gin.Context) {
		auth.SetAdminContext(c, &auth.Claims{AdminID: refundTestAdminID})
		c.Next()
	}

	engine.POST("/api/v1/refund-requests", userAuth, h.Create)
	engine.GET("/api/v1/refund-requests", userAuth, h.ListMy)
	engine.DELETE("/api/v1/refund-requests/:id", userAuth, h.Cancel)
	engine.GET("/api/v1/admin/refund-requests", adminAuth, h.ListAdmin)
	engine.POST("/api/v1/admin/refund-requests/:id/approve", adminAuth, h.Approve)
	engine.POST("/api/v1/admin/refund-requests/:id/reject", adminAuth, h.Reject)
	return engine
}

func serveRefundRequest(engine *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var reader *bytes.Buffer
	if body != "" {
		reader = bytes.NewBufferString(body)
	} else {
		reader = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestRefundRequestHandler_Create(t *testing.T) {
	svc := &mockRefundRequestService{
		createFunc: func(_ context.Context, input service.CreateRefundRequestInput) (*service.RefundRequestResponse, error) {
			if input.UserID != refundTestUserID {
				t.Fatal("handler must forward the authenticated user id")
			}
			return &service.RefundRequestResponse{
				ID:      uuid.New().String(),
				OrderID: input.OrderID.String(),
				Amount:  input.Amount,
				Status:  "pending",
			}, nil
		},
	}
	engine := newRefundRequestEngine(t, svc)

	rec := serveRefundRequest(engine, http.MethodPost, "/api/v1/refund-requests",
		fmt.Sprintf(`{"order_id":%q,"amount":300,"reason":"重复充值"}`, uuid.New().String()))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// Missing/invalid fields fail binding.
	rec = serveRefundRequest(engine, http.MethodPost, "/api/v1/refund-requests", `{"amount":300}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing order_id, got %d", rec.Code)
	}
	rec = serveRefundRequest(engine, http.MethodPost, "/api/v1/refund-requests",
		fmt.Sprintf(`{"order_id":%q,"amount":0}`, uuid.New().String()))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-positive amount, got %d", rec.Code)
	}
}

func TestRefundRequestHandler_ListMy(t *testing.T) {
	svc := &mockRefundRequestService{
		listMyFunc: func(_ context.Context, userID uuid.UUID, page, pageSize int) ([]service.RefundRequestResponse, int, error) {
			return []service.RefundRequestResponse{{ID: uuid.New().String(), Status: "pending"}}, 1, nil
		},
	}
	engine := newRefundRequestEngine(t, svc)

	rec := serveRefundRequest(engine, http.MethodGet, "/api/v1/refund-requests?page=1&page_size=10", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRefundRequestHandler_Cancel(t *testing.T) {
	cancelled := false
	svc := &mockRefundRequestService{
		cancelFunc: func(_ context.Context, userID, requestID uuid.UUID) error {
			cancelled = true
			return nil
		},
	}
	engine := newRefundRequestEngine(t, svc)

	rec := serveRefundRequest(engine, http.MethodDelete, "/api/v1/refund-requests/"+uuid.New().String(), "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if !cancelled {
		t.Fatal("service cancel was not invoked")
	}

	rec = serveRefundRequest(engine, http.MethodDelete, "/api/v1/refund-requests/not-a-uuid", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad id, got %d", rec.Code)
	}
}

func TestRefundRequestHandler_ListAdmin(t *testing.T) {
	svc := &mockRefundRequestService{
		listFunc: func(_ context.Context, filter service.ListRefundRequestFilter) ([]service.RefundRequestResponse, int, error) {
			if filter.Status == nil || *filter.Status != "pending" {
				t.Fatal("status filter not forwarded")
			}
			return []service.RefundRequestResponse{{ID: uuid.New().String(), Status: "pending"}}, 1, nil
		},
	}
	engine := newRefundRequestEngine(t, svc)

	rec := serveRefundRequest(engine, http.MethodGet, "/api/v1/admin/refund-requests?status=pending", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Unknown status values fail validation.
	rec = serveRefundRequest(engine, http.MethodGet, "/api/v1/admin/refund-requests?status=weird", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad status, got %d", rec.Code)
	}
}

func TestRefundRequestHandler_ApproveReject(t *testing.T) {
	svc := &mockRefundRequestService{
		approveFunc: func(_ context.Context, requestID, adminID uuid.UUID) (*service.RefundRequestResponse, error) {
			if adminID != refundTestAdminID {
				t.Fatal("handler must forward the authenticated admin id")
			}
			return &service.RefundRequestResponse{ID: requestID.String(), Status: "approved"}, nil
		},
		rejectFunc: func(_ context.Context, requestID, adminID uuid.UUID) (*service.RefundRequestResponse, error) {
			return &service.RefundRequestResponse{ID: requestID.String(), Status: "rejected"}, nil
		},
	}
	engine := newRefundRequestEngine(t, svc)
	id := uuid.New().String()

	rec := serveRefundRequest(engine, http.MethodPost, "/api/v1/admin/refund-requests/"+id+"/approve", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = serveRefundRequest(engine, http.MethodPost, "/api/v1/admin/refund-requests/"+id+"/reject", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = serveRefundRequest(engine, http.MethodPost, "/api/v1/admin/refund-requests/bad-id/approve", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad id, got %d", rec.Code)
	}
}
