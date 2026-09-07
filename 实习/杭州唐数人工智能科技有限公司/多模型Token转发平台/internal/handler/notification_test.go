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
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockNotificationService struct {
	createFunc       func(ctx context.Context, input service.CreateNotificationInput) (*service.NotificationResponse, error)
	listMyFunc       func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.NotificationResponse, int, error)
	listAdminFunc    func(ctx context.Context, userID *uuid.UUID, notificationType *string, page, pageSize int) ([]service.NotificationResponse, int, error)
	markReadFunc     func(ctx context.Context, userID, notificationID uuid.UUID) error
}

func (m *mockNotificationService) CreateNotification(ctx context.Context, input service.CreateNotificationInput) (*service.NotificationResponse, error) {
	return m.createFunc(ctx, input)
}

func (m *mockNotificationService) ListMyNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.NotificationResponse, int, error) {
	return m.listMyFunc(ctx, userID, page, pageSize)
}

func (m *mockNotificationService) ListNotificationsAdmin(ctx context.Context, userID *uuid.UUID, notificationType *string, page, pageSize int) ([]service.NotificationResponse, int, error) {
	return m.listAdminFunc(ctx, userID, notificationType, page, pageSize)
}

func (m *mockNotificationService) MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	return m.markReadFunc(ctx, userID, notificationID)
}

func newNotificationEngine(t *testing.T, svc NotificationService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewNotificationHandler(svc)

	// Simulate API key authentication for user-facing routes.
	userAuth := func(c *gin.Context) {
		auth.SetAPIKeyContext(c, &auth.APIKeyInfo{UserID: uuid.MustParse("11111111-1111-1111-1111-111111111111")})
		c.Next()
	}

	engine.GET("/api/v1/notifications", userAuth, handler.ListMy)
	engine.PATCH("/api/v1/notifications/:id/read", userAuth, handler.MarkRead)
	engine.POST("/api/v1/admin/notifications", handler.Create)
	engine.GET("/api/v1/admin/notifications", handler.ListAdmin)

	return engine
}

func TestNotificationHandler_ListMy(t *testing.T) {
	svc := &mockNotificationService{
		listMyFunc: func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.NotificationResponse, int, error) {
			return []service.NotificationResponse{{ID: uuid.New().String(), Title: "通知"}}, 1, nil
		},
	}

	engine := newNotificationEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNotificationHandler_CreateAdmin(t *testing.T) {
	svc := &mockNotificationService{
		createFunc: func(ctx context.Context, input service.CreateNotificationInput) (*service.NotificationResponse, error) {
			return &service.NotificationResponse{ID: uuid.New().String(), Title: input.Title}, nil
		},
	}

	engine := newNotificationEngine(t, svc)
	body := `{"title":"公告","content":"内容","type":"announcement"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/notifications", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNotificationHandler_MarkRead(t *testing.T) {
	svc := &mockNotificationService{
		markReadFunc: func(ctx context.Context, userID, notificationID uuid.UUID) error {
			return nil
		},
	}

	engine := newNotificationEngine(t, svc)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+uuid.New().String()+"/read", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNotificationHandler_MarkReadNotFound(t *testing.T) {
	svc := &mockNotificationService{
		markReadFunc: func(ctx context.Context, userID, notificationID uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	engine := newNotificationEngine(t, svc)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+uuid.New().String()+"/read", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func (m *mockNotificationService) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockNotificationService) ReadAll(ctx context.Context, userID uuid.UUID) error {
	return nil
}
