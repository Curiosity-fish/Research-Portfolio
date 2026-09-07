//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

func buildAuthEngine(t *testing.T, client *ent.Client) *gin.Engine {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gin.SetMode(gin.TestMode)

	adminRepo := repository.NewEntAdminRepository(client)
	tokenManager := auth.NewTokenManager("integration-test-secret-32bytes", time.Hour)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.DefaultBcryptCost, time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)

	engine.POST("/api/v1/admin/login", authHandler.Login)
	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager))
	adminGroup.GET("/me", authHandler.Me)

	return engine
}

func TestAuth_LoginAndProtectedRoute(t *testing.T) {
	dbURL := os.Getenv("APP_DATABASE__URL")
	if dbURL == "" {
		t.Skip("APP_DATABASE__URL not set")
	}

	ctx := context.Background()

	client, err := repository.OpenEntClient(dbURL)
	if err != nil {
		t.Fatalf("open ent client: %v", err)
	}
	defer client.Close()

	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	// Seed an active admin with a unique username so repeated runs do not
	// collide on the unique constraint.
	username := "ia-" + uuid.NewString()[:8]
	repo := repository.NewEntAdminRepository(client)
	hash, err := auth.HashPassword(auth.DefaultBcryptCost, "integration-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	admin, err := repo.CreateAdmin(ctx, username, hash, adminuser.RoleSuperAdmin)
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	t.Cleanup(func() {
		_ = client.AdminUser.DeleteOneID(admin.ID).Exec(context.Background())
	})

	engine := buildAuthEngine(t, client)

	// Login.
	loginBody := fmt.Sprintf(`{"username":%q,"password":"integration-password"}`, username)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewReader([]byte(loginBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var loginResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int64  `json:"expires_in"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResp.Data.AccessToken == "" {
		t.Fatal("expected access token")
	}

	// Access protected route.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Data.AccessToken)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("me expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var meResp struct {
		Data struct {
			AdminID string `json:"admin_id"`
			Role    string `json:"role"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &meResp); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if meResp.Data.Role != string(adminuser.RoleSuperAdmin) {
		t.Errorf("expected role super_admin, got %s", meResp.Data.Role)
	}

	// Invalid token should be rejected.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", rec.Code)
	}
}
