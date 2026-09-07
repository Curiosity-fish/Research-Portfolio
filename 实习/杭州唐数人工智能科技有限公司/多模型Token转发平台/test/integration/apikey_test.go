//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func buildAPIKeyEngine(t *testing.T) *gin.Engine {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)

	return engine
}

func TestAPIKey_ValidAndInvalidKeys(t *testing.T) {
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

	user, group, token, plaintext := testutil.CreateTestAPIKey(t, ctx, client)
	lookup := repository.NewEntAPIKeyLookup(client)
	validator := auth.NewAPIKeyValidator(lookup)

	apiHandler := handler.NewAPIKeyHandler()
	engine := buildAPIKeyEngine(t)
	apiGroup := engine.Group("/api/v1", middleware.APIKeyAuth(validator))
	apiGroup.GET("/me", apiHandler.Me)

	// Valid key returns identity fields.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+plaintext)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid key, got %d: %s", rec.Code, rec.Body.String())
	}

	var meResp struct {
		Data handler.APIKeyMeResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &meResp); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if meResp.Data.UserID != user.ID.String() {
		t.Errorf("expected user_id %s, got %s", user.ID.String(), meResp.Data.UserID)
	}
	if meResp.Data.TokenID != token.ID.String() {
		t.Errorf("expected token_id %s, got %s", token.ID.String(), meResp.Data.TokenID)
	}
	if meResp.Data.GroupID != group.ID.String() {
		t.Errorf("expected group_id %s, got %s", group.ID.String(), meResp.Data.GroupID)
	}
	if meResp.Data.GroupCode != group.Code {
		t.Errorf("expected group_code %s, got %s", group.Code, meResp.Data.GroupCode)
	}

	// Missing header returns 401.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing header, got %d", rec.Code)
	}

	// Wrong scheme returns 401.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Basic "+plaintext)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for Basic scheme, got %d", rec.Code)
	}

	// Tampered key returns 401.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+plaintext+"x")
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for tampered key, got %d", rec.Code)
	}
}
