package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockAuditMiddlewareService struct {
	logFunc func(ctx context.Context, input service.CreateAuditLogInput) error
}

func (m *mockAuditMiddlewareService) Log(ctx context.Context, input service.CreateAuditLogInput) error {
	return m.logFunc(ctx, input)
}

func TestAuditMiddleware_LogsAdminRequest(t *testing.T) {
	called := false
	svc := &mockAuditMiddlewareService{
		logFunc: func(ctx context.Context, input service.CreateAuditLogInput) error {
			called = true
			if input.ActorType != "admin" {
				t.Fatalf("expected actor type admin, got %s", input.ActorType)
			}
			if input.Action == "" {
				t.Fatal("expected action")
			}
			return nil
		},
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		auth.SetAdminContext(c, &auth.Claims{AdminID: uuid.MustParse("22222222-2222-2222-2222-222222222222")})
		c.Next()
	})
	engine.Use(AuditMiddleware(svc))
	engine.GET("/api/v1/admin/users", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !called {
		t.Fatal("expected audit log to be written")
	}
}

func TestAuditMiddleware_SkipsWithoutAdmin(t *testing.T) {
	called := false
	svc := &mockAuditMiddlewareService{
		logFunc: func(ctx context.Context, input service.CreateAuditLogInput) error {
			called = true
			return nil
		},
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(AuditMiddleware(svc))
	engine.GET("/api/v1/admin/users", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if called {
		t.Fatal("expected no audit log without admin context")
	}
}

func TestAuditMiddleware_DoesNotFailRequestOnError(t *testing.T) {
	svc := &mockAuditMiddlewareService{
		logFunc: func(ctx context.Context, input service.CreateAuditLogInput) error {
			return context.Canceled
		},
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		auth.SetAdminContext(c, &auth.Claims{AdminID: uuid.MustParse("22222222-2222-2222-2222-222222222222")})
		c.Next()
	})
	engine.Use(AuditMiddleware(svc))
	engine.GET("/api/v1/admin/users", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
