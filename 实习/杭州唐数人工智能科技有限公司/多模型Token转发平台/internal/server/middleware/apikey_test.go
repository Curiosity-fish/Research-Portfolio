package middleware

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
)

type mockAPIKeyValidator struct {
	info *auth.APIKeyInfo
	err  error
}

func (m *mockAPIKeyValidator) Validate(ctx context.Context, key string) (*auth.APIKeyInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.info, nil
}

func newAPIKeyTestEngine(t *testing.T, v auth.APIKeyValidator) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine := gin.New()
	engine.Use(Recovery(logger), RequestID(), RequestLogger(logger))

	protected := engine.Group("/api/v1", APIKeyAuth(v))
	protected.GET("/me", func(c *gin.Context) {
		userID, _ := auth.UserID(c)
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"user_id": userID.String()}})
	})
	return engine
}

func TestAPIKeyAuth_Success(t *testing.T) {
	userID := uuid.New()
	v := &mockAPIKeyValidator{
		info: &auth.APIKeyInfo{
			UserID:    userID,
			TokenID:   uuid.New(),
			GroupCode: "g1",
		},
	}
	engine := newAPIKeyTestEngine(t, v)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer sk-valid-key")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAPIKeyAuth_MissingHeader(t *testing.T) {
	engine := newAPIKeyTestEngine(t, &mockAPIKeyValidator{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAPIKeyAuth_InvalidScheme(t *testing.T) {
	engine := newAPIKeyTestEngine(t, &mockAPIKeyValidator{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Basic invalid")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAPIKeyAuth_WrongPrefix(t *testing.T) {
	engine := newAPIKeyTestEngine(t, &mockAPIKeyValidator{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer not-an-sk-key")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	v := &mockAPIKeyValidator{err: errors.New("invalid key")}
	engine := newAPIKeyTestEngine(t, v)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer sk-invalid-key")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}
