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

func buildUserEngine(t *testing.T, client *ent.Client) *gin.Engine {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gin.SetMode(gin.TestMode)

	adminRepo := repository.NewEntAdminRepository(client)
	tokenManager := auth.NewTokenManager("integration-test-secret-32bytes", time.Hour)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.DefaultBcryptCost, time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	userRepo := repository.NewEntUserRepository(client)
	userService := service.NewUserService(userRepo, auth.DefaultBcryptCost)
	userHandler := handler.NewUserHandler(userService)

	userTokenRepo := repository.NewEntUserTokenRepository(client)
	userTokenService := service.NewUserTokenService(userTokenRepo, userRepo, []byte("0123456789abcdef0123456789abcdef"), nil)
	userTokenHandler := handler.NewUserTokenHandler(userTokenService)

	departmentRepo := repository.NewEntDepartmentRepository(client)
	departmentService := service.NewDepartmentService(departmentRepo)
	departmentHandler := handler.NewDepartmentHandler(departmentService)

	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)

	engine.POST("/api/v1/admin/login", authHandler.Login)
	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager))
	adminGroup.GET("/me", authHandler.Me)
	adminGroup.POST("/users", userHandler.Create)
	adminGroup.GET("/users", userHandler.List)
	adminGroup.GET("/users/:id", userHandler.Get)
	adminGroup.PUT("/users/:id", userHandler.Update)
	adminGroup.PATCH("/users/:id/status", userHandler.UpdateStatus)

	adminGroup.POST("/users/:id/tokens", userTokenHandler.Create)
	adminGroup.GET("/users/:id/tokens", userTokenHandler.List)
	adminGroup.DELETE("/users/:id/tokens/:token_id", userTokenHandler.Delete)
	adminGroup.PATCH("/users/:id/tokens/:token_id/status", userTokenHandler.UpdateStatus)

	adminGroup.POST("/departments", departmentHandler.Create)
	adminGroup.GET("/departments", departmentHandler.List)
	adminGroup.GET("/departments/:id", departmentHandler.Get)
	adminGroup.PUT("/departments/:id", departmentHandler.Update)
	adminGroup.DELETE("/departments/:id", departmentHandler.Delete)

	return engine
}

func createAdminAndLogin(t *testing.T, ctx context.Context, client *ent.Client, engine *gin.Engine) string {
	t.Helper()

	username := "iu-" + uuid.NewString()[:8]
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
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	return loginResp.Data.AccessToken
}

func TestUser_CRUD(t *testing.T) {
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

	engine := buildUserEngine(t, client)
	token := createAdminAndLogin(t, ctx, client, engine)

	// Create user.
	username := "int-user-" + uuid.NewString()[:8]
	createBody := fmt.Sprintf(`{"username":%q,"password":"password123","name":"Integration User","email":"int1@example.com","role":"teacher"}`, username)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var createResp struct {
		Data service.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	userID := createResp.Data.ID
	if createResp.Data.Role != "teacher" {
		t.Errorf("expected role teacher, got %s", createResp.Data.Role)
	}
	t.Cleanup(func() {
		_ = client.User.DeleteOneID(uuid.MustParse(userID)).Exec(context.Background())
	})

	// Get user.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get user expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Update user.
	updateBody := `{"name":"Updated User"}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/"+userID, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update user expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var updateResp struct {
		Data service.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &updateResp); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updateResp.Data.Name != "Updated User" {
		t.Errorf("expected name Updated User, got %s", updateResp.Data.Name)
	}

	// Update status.
	statusBody := `{"status":"inactive"}`
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+userID+"/status", bytes.NewReader([]byte(statusBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// List users.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?keyword="+username, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list users expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var listResp struct {
		Data service.ListUsersResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.Data.Total != 1 {
		t.Errorf("expected total 1, got %d", listResp.Data.Total)
	}
}

func TestUser_Unauthorized(t *testing.T) {
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

	engine := buildUserEngine(t, client)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", rec.Code)
	}
}
