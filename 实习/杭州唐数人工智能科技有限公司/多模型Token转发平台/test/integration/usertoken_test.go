//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

func TestUserToken_CRUD(t *testing.T) {
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

	// Create a user to own tokens.
	username := "tok-owner-" + uuid.NewString()[:8]
	createBody := map[string]interface{}{
		"username": username,
		"password": "password123",
		"name":     "Token Owner",
	}
	bodyBytes, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var userResp struct {
		Data service.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &userResp); err != nil {
		t.Fatalf("decode user response: %v", err)
	}
	userID := userResp.Data.ID
	t.Cleanup(func() {
		_ = client.User.DeleteOneID(uuid.MustParse(userID)).Exec(context.Background())
	})

	// Create token.
	tokenBody := `{"name":"default","quota_limit":1000,"expires_at":"2026-12-31T23:59:59Z"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+userID+"/tokens", bytes.NewReader([]byte(tokenBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create token expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var createResp struct {
		Data handler.CreateUserTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create token response: %v", err)
	}
	if createResp.Data.Plaintext == "" {
		t.Fatal("expected plaintext token")
	}
	if createResp.Data.TokenLast4 != createResp.Data.Plaintext[len(createResp.Data.Plaintext)-4:] {
		t.Errorf("token_last4 mismatch")
	}
	tokenID := createResp.Data.ID

	// Validate the created token via API Key middleware.
	lookup := repository.NewEntAPIKeyLookup(client)
	validator := auth.NewAPIKeyValidator(lookup)
	apiHandler := handler.NewAPIKeyHandler()
	apiEngine := buildAPIKeyEngine(t)
	apiGroup := apiEngine.Group("/api/v1", middleware.APIKeyAuth(validator))
	apiGroup.GET("/me", apiHandler.Me)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+createResp.Data.Plaintext)
	rec = httptest.NewRecorder()
	apiEngine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for created token, got %d: %s", rec.Code, rec.Body.String())
	}

	// List tokens.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/"+userID+"/tokens", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list tokens expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var listResp struct {
		Data struct {
			List []service.UserTokenResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Data.List) != 1 {
		t.Errorf("expected 1 token, got %d", len(listResp.Data.List))
	}

	// Disable token.
	statusBody := `{"is_enabled":false}`
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+userID+"/tokens/"+tokenID+"/status", bytes.NewReader([]byte(statusBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update token status expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Disabled token should be rejected.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+createResp.Data.Plaintext)
	rec = httptest.NewRecorder()
	apiEngine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for disabled token, got %d", rec.Code)
	}

	// Delete token.
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/"+userID+"/tokens/"+tokenID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete token expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	_ = time.Now
	_ = user.RoleStudent
}
