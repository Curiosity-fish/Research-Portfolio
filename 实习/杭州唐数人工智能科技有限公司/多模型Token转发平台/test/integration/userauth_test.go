//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

// buildUserAuthEngine assembles the user-portal account routes plus the
// minimal admin routes needed to seed a user and token.
func buildUserAuthEngine(t *testing.T, client *ent.Client) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	tokenManager := auth.NewTokenManager("user-auth-integration-secret-32b", time.Hour)

	adminRepo := repository.NewEntAdminRepository(client)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.DefaultBcryptCost, time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	userRepo := repository.NewEntUserRepository(client)
	userService := service.NewUserService(userRepo, auth.DefaultBcryptCost)
	userHandler := handler.NewUserHandler(userService)

	userTokenService := service.NewUserTokenService(
		repository.NewEntUserTokenRepository(client),
		userRepo,
		[]byte("0123456789abcdef0123456789abcdef"),
		nil,
	)
	userTokenHandler := handler.NewUserTokenHandler(userTokenService)

	userAuthService := service.NewUserAuthService(
		userRepo,
		repository.NewEntUserTokenRepository(client),
		tokenManager,
		auth.DefaultBcryptCost,
		[]byte("0123456789abcdef0123456789abcdef"),
		int64((time.Hour).Seconds()),
	)
	userAuthHandler := handler.NewUserAuthHandler(userAuthService)

	engine := gin.New()
	engine.Use(middleware.Recovery(slog.New(slog.NewTextHandler(io.Discard, nil))))
	engine.POST("/api/v1/admin/login", authHandler.Login)
	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager))
	adminGroup.POST("/users", userHandler.Create)
	adminGroup.POST("/users/:id/tokens", userTokenHandler.Create)
	adminGroup.GET("/users", userHandler.List)

	engine.POST("/api/v1/user/login", userAuthHandler.Login)
	userGroup := engine.Group("/api/v1/user", middleware.UserJWTAuth(tokenManager))
	userGroup.POST("/change-password", userAuthHandler.ChangePassword)
	userGroup.POST("/token/reveal", userAuthHandler.RevealToken)

	return engine
}

func bytesReader(s string) *bytes.Reader {
	return bytes.NewReader([]byte(s))
}

func userAuthRequest(t *testing.T, engine *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytesReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

// TestUserAuthFlow covers login, audience isolation, token reveal and
// password change end to end.
func TestUserAuthFlow(t *testing.T) {
	ctx := context.Background()
	client := openOpsTestClient(t)
	engine := buildUserAuthEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	// Seed a user and an API token via the admin API.
	suffix := uuid.NewString()[:8]
	createBody := `{"username":"portal-` + suffix + `","password":"password123","name":"门户用户","role":"student"}`
	rec := userAuthRequest(t, engine, http.MethodPost, "/api/v1/admin/users", adminToken, createBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var userResp struct {
		Data service.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &userResp); err != nil {
		t.Fatalf("decode user: %v", err)
	}

	rec = userAuthRequest(t, engine, http.MethodPost, "/api/v1/admin/users/"+userResp.Data.ID+"/tokens", adminToken, `{"name":"portal"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create token expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var tokenResp struct {
		Data handler.CreateUserTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if tokenResp.Data.Plaintext == "" {
		t.Fatal("expected one-time plaintext token")
	}

	// Login returns a user JWT.
	rec = userAuthRequest(t, engine, http.MethodPost, "/api/v1/user/login", "", `{"username":"portal-`+suffix+`","password":"password123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var loginResp struct {
		Data service.UserLoginResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	userToken := loginResp.Data.AccessToken
	if userToken == "" || loginResp.Data.ExpiresIn == 0 {
		t.Fatalf("unexpected login result: %+v", loginResp.Data)
	}

	// Wrong password produces the same error as unknown user (no enumeration).
	rec = userAuthRequest(t, engine, http.MethodPost, "/api/v1/user/login", "", `{"username":"portal-`+suffix+`","password":"wrong-password"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password expected 401, got %d", rec.Code)
	}
	var errResp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	rec2 := userAuthRequest(t, engine, http.MethodPost, "/api/v1/user/login", "", `{"username":"ghost-`+suffix+`","password":"password123"}`)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("unknown user expected 401, got %d", rec2.Code)
	}
	var errResp2 struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &errResp2); err != nil {
		t.Fatalf("decode error2: %v", err)
	}
	if errResp.Error.Message != errResp2.Error.Message {
		t.Fatalf("login errors must be identical, got %q vs %q", errResp.Error.Message, errResp2.Error.Message)
	}

	// Audience isolation: user JWT cannot call admin routes.
	rec = userAuthRequest(t, engine, http.MethodGet, "/api/v1/admin/users", userToken, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("user JWT on admin route expected 401, got %d", rec.Code)
	}

	// Token reveal requires the login password and returns the original key.
	rec = userAuthRequest(t, engine, http.MethodPost, "/api/v1/user/token/reveal", userToken,
		`{"token_id":"`+tokenResp.Data.ID+`","password":"wrong-password"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("reveal with wrong password expected 401, got %d", rec.Code)
	}

	rec = userAuthRequest(t, engine, http.MethodPost, "/api/v1/user/token/reveal", userToken,
		`{"token_id":"`+tokenResp.Data.ID+`","password":"password123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("reveal expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var revealResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &revealResp); err != nil {
		t.Fatalf("decode reveal: %v", err)
	}
	if revealResp.Data.Token != tokenResp.Data.Plaintext {
		t.Fatal("revealed token must equal the one-time plaintext from creation")
	}

	// Change password: old fails, new works.
	rec = userAuthRequest(t, engine, http.MethodPost, "/api/v1/user/change-password", userToken,
		`{"old_password":"password123","password":"newpassword9"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("change password expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = userAuthRequest(t, engine, http.MethodPost, "/api/v1/user/login", "", `{"username":"portal-`+suffix+`","password":"password123"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("old password expected 401, got %d", rec.Code)
	}
	rec = userAuthRequest(t, engine, http.MethodPost, "/api/v1/user/login", "", `{"username":"portal-`+suffix+`","password":"newpassword9"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("new password expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
