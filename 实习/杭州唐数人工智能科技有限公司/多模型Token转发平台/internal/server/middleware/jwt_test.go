package middleware

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/auth"
)

type mockJWTTokenManager struct {
	parse func(ctx context.Context, tokenStr string) (*auth.Claims, error)
}

func (m *mockJWTTokenManager) Generate(ctx context.Context, admin *ent.AdminUser) (string, error) {
	return "", errors.New("not implemented")
}

func (m *mockJWTTokenManager) GenerateUser(ctx context.Context, user *ent.User) (string, error) {
	return "", errors.New("not implemented")
}

func (m *mockJWTTokenManager) Parse(ctx context.Context, tokenStr string) (*auth.Claims, error) {
	if m.parse != nil {
		return m.parse(ctx, tokenStr)
	}
	return nil, errors.New("not implemented")
}

func newJWTTestEngine(t *testing.T, tm auth.TokenManager) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine := gin.New()
	engine.Use(Recovery(logger), RequestID(), RequestLogger(logger))

	protected := engine.Group("/api/v1/admin", JWTAuth(tm))
	protected.GET("/me", func(c *gin.Context) {
		id, _ := auth.AdminID(c)
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"admin_id": id.String()}})
	})
	return engine
}

func TestJWTAuth_Success(t *testing.T) {
	adminID := uuid.New()
	tm := &mockJWTTokenManager{
		parse: func(ctx context.Context, tokenStr string) (*auth.Claims, error) {
			return &auth.Claims{
				AdminID: adminID,
				Role:    "super_admin",
				Status:  "active",
				RegisteredClaims: jwt.RegisteredClaims{
					Audience: jwt.ClaimStrings{auth.AudienceAdmin},
				},
			}, nil
		},
	}
	engine := newJWTTestEngine(t, tm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	engine := newJWTTestEngine(t, &mockJWTTokenManager{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestJWTAuth_InvalidScheme(t *testing.T) {
	engine := newJWTTestEngine(t, &mockJWTTokenManager{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Basic invalid")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	tm := &mockJWTTokenManager{
		parse: func(ctx context.Context, tokenStr string) (*auth.Claims, error) {
			return nil, errors.New("invalid token")
		},
	}
	engine := newJWTTestEngine(t, tm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestJWTAuth_InactiveAccount(t *testing.T) {
	tm := &mockJWTTokenManager{
		parse: func(ctx context.Context, tokenStr string) (*auth.Claims, error) {
			return &auth.Claims{
				AdminID: uuid.New(),
				Role:    "admin",
				Status:  "inactive",
			}, nil
		},
	}
	engine := newJWTTestEngine(t, tm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

// userAudienceEngine mounts one admin route and one user route to verify that
// tokens cannot cross audiences in either direction.
func userAudienceEngine(t *testing.T, tm auth.TokenManager) *gin.Engine {
	t.Helper()
	engine := newJWTTestEngine(t, tm)
	admin := engine.Group("/admin", JWTAuth(tm))
	admin.GET("/me", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	user := engine.Group("/user", UserJWTAuth(tm))
	user.GET("/me", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return engine
}

func TestAudienceIsolation_UserTokenRejectedByAdmin(t *testing.T) {
	tm := auth.NewTokenManager("test-secret-32bytes-for-jwt-tests", time.Hour)
	userToken, err := tm.GenerateUser(context.Background(), &ent.User{ID: uuid.New(), Status: "active"})
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}
	engine := userAudienceEngine(t, tm)

	req := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected user token rejected on admin route, got %d", rec.Code)
	}
}

func TestAudienceIsolation_AdminTokenRejectedByUser(t *testing.T) {
	tm := auth.NewTokenManager("test-secret-32bytes-for-jwt-tests", time.Hour)
	adminToken, err := tm.Generate(context.Background(), &ent.AdminUser{ID: uuid.New(), Status: "active"})
	if err != nil {
		t.Fatalf("generate admin token: %v", err)
	}
	engine := userAudienceEngine(t, tm)

	req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected admin token rejected on user route, got %d", rec.Code)
	}
}

func TestUserJWTAuth_Success(t *testing.T) {
	tm := auth.NewTokenManager("test-secret-32bytes-for-jwt-tests", time.Hour)
	userID := uuid.New()
	userToken, err := tm.GenerateUser(context.Background(), &ent.User{ID: userID, Status: "active"})
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}
	engine := userAudienceEngine(t, tm)

	req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
