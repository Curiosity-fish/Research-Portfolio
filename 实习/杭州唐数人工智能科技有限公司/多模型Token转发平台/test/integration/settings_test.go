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
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/setting"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

// buildSettingsEngine assembles the admin settings routes together with the
// user-facing recharge / quota-request routes so the feature-mode and limit
// enforcement can be exercised end to end.
func buildSettingsEngine(t *testing.T, client *ent.Client) *gin.Engine {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gin.SetMode(gin.TestMode)

	adminRepo := repository.NewEntAdminRepository(client)
	tokenManager := auth.NewTokenManager("settings-integration-secret-32b", time.Hour)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.DefaultBcryptCost, time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	userRepo := repository.NewEntUserRepository(client)
	userService := service.NewUserService(userRepo, auth.DefaultBcryptCost)
	userHandler := handler.NewUserHandler(userService)

	settingsService := service.NewSettingsService(repository.NewEntSettingRepository(client))
	settingsHandler := handler.NewSettingsHandler(settingsService)

	userTokenRepo := repository.NewEntUserTokenRepository(client)
	userTokenService := service.NewUserTokenService(
		userTokenRepo, userRepo,
		[]byte("0123456789abcdef0123456789abcdef"),
		settingsService,
	)
	userTokenHandler := handler.NewUserTokenHandler(userTokenService)

	rechargeOrderRepo := repository.NewEntRechargeOrderRepository(client)
	rechargeOrderService := service.NewRechargeOrderService(rechargeOrderRepo, settingsService)
	rechargeOrderHandler := handler.NewRechargeOrderHandler(rechargeOrderService)

	quotaRequestRepo := repository.NewEntQuotaRequestRepository(client)
	quotaRequestService := service.NewQuotaRequestService(quotaRequestRepo, userTokenRepo, settingsService)
	quotaRequestHandler := handler.NewQuotaRequestHandler(quotaRequestService)

	apiKeyLookup := repository.NewEntAPIKeyLookup(client)
	apiKeyValidator := auth.NewAPIKeyValidator(apiKeyLookup)

	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)

	engine.POST("/api/v1/admin/login", authHandler.Login)

	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager))
	adminGroup.POST("/users", userHandler.Create)
	adminGroup.POST("/users/:id/tokens", userTokenHandler.Create)
	adminGroup.GET("/users/:id/tokens", userTokenHandler.List)
	adminGroup.GET("/settings/feature-mode", settingsHandler.GetFeatureMode)
	adminGroup.PUT("/settings/feature-mode", settingsHandler.PutFeatureMode)
	adminGroup.GET("/settings/default-quota", settingsHandler.GetDefaultQuota)
	adminGroup.PUT("/settings/default-quota", settingsHandler.PutDefaultQuota)
	adminGroup.GET("/settings/quota-range", settingsHandler.GetQuotaRange)
	adminGroup.PUT("/settings/quota-range", settingsHandler.PutQuotaRange)
	adminGroup.GET("/settings/recharge", settingsHandler.GetRechargeLimits)
	adminGroup.PUT("/settings/recharge", settingsHandler.PutRechargeLimits)

	apiGroup := engine.Group("/api/v1", middleware.APIKeyAuth(apiKeyValidator))
	apiGroup.POST("/recharge/orders", rechargeOrderHandler.Create)
	apiGroup.GET("/recharge/config", settingsHandler.GetPublicRechargeConfig)
	apiGroup.POST("/quota-requests", quotaRequestHandler.Create)

	return engine
}

func settingsIntegrationRequest(t *testing.T, engine *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func decodeSettingsEnvelope(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var parsed struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("decode envelope: %v; body=%s", err, rec.Body.String())
	}
	var data map[string]interface{}
	if err := json.Unmarshal(parsed.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	return data
}

func TestSettings_FeatureModeLimitsAndDefaultQuotaFlow(t *testing.T) {
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

	// The settings rows are global; remove them afterwards so other tests in
	// the shared database never observe this test's configuration.
	t.Cleanup(func() {
		_, _ = client.Setting.Delete().
			Where(setting.KeyIn(
				service.SettingKeyFeatureMode,
				service.SettingKeyDefaultQuotaLimit,
				service.SettingKeyQuotaRange,
				service.SettingKeyRechargeLimits,
			)).
			Exec(context.Background())
	})

	engine := buildSettingsEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	username := "settings-user-" + uuid.NewString()[:8]
	createBody, _ := json.Marshal(map[string]interface{}{
		"username": username,
		"password": "password123",
		"name":     "Settings User",
		"role":     "student",
	})
	rec := settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/admin/users", adminToken, string(createBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var userResp struct {
		Data service.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &userResp); err != nil {
		t.Fatalf("decode user response: %v", err)
	}
	userID := uuid.MustParse(userResp.Data.ID)
	t.Cleanup(func() {
		_ = client.User.DeleteOneID(userID).Exec(context.Background())
	})

	// Create the first token with an explicit quota so quota requests work.
	tokenResp := createTokenForSettings(t, engine, adminToken, userID, `{"name":"primary","quota_limit":1000}`)
	apiKey := tokenResp.Plaintext
	tokenID := tokenResp.ID

	// Default mode: recharge order allowed within default limits.
	rec = settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders", apiKey, `{"amount":5000000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("default mode recharge expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// Public recharge config exposes defaults and the mock channel.
	rec = settingsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/recharge/config", apiKey, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("recharge config expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	configData := decodeSettingsEnvelope(t, rec)
	if configData["min_amount"] != float64(1_000_000) {
		t.Fatalf("expected default min_amount 1000000, got %v", configData["min_amount"])
	}
	channels, ok := configData["channels"].([]interface{})
	if !ok || len(channels) != 1 || channels[0] != "mock" {
		t.Fatalf("expected channels [mock], got %v", configData["channels"])
	}

	// quota_only disables recharge orders but keeps quota requests open.
	rec = settingsIntegrationRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/feature-mode", adminToken, `{"mode":"quota_only"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set quota_only expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders", apiKey, `{"amount":5000000}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("recharge under quota_only expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/quota-requests", apiKey,
		`{"token_id":"`+tokenID+`","requested_amount":50}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("quota request under quota_only expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// recharge_only disables quota requests but keeps recharge open.
	rec = settingsIntegrationRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/feature-mode", adminToken, `{"mode":"recharge_only"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set recharge_only expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/quota-requests", apiKey,
		`{"token_id":"`+tokenID+`","requested_amount":50}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("quota request under recharge_only expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders", apiKey, `{"amount":5000000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("recharge under recharge_only expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// Custom amount limits are enforced on order creation.
	rec = settingsIntegrationRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/recharge", adminToken,
		`{"min_amount":2000000,"max_amount":10000000,"quick_amounts":[5000000]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set recharge limits expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, amount := range []string{"1000000", "20000000"} {
		rec = settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders", apiKey, `{"amount":`+amount+`}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("amount %s outside limits expected 400, got %d: %s", amount, rec.Code, rec.Body.String())
		}
	}
	rec = settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders", apiKey, `{"amount":5000000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("amount within limits expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// Feature mode round trip: back to both, invalid value rejected.
	rec = settingsIntegrationRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/feature-mode", adminToken, `{"mode":"both"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("reset feature mode expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = settingsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/admin/settings/feature-mode", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get feature mode expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if data := decodeSettingsEnvelope(t, rec); data["mode"] != "both" {
		t.Fatalf("expected mode both, got %v", data["mode"])
	}
	rec = settingsIntegrationRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/feature-mode", adminToken, `{"mode":"nonsense"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid mode expected 400, got %d", rec.Code)
	}

	// Shrink the quota range first so the small values used below are valid;
	// the default quota must then stay inside it.
	rec = settingsIntegrationRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/quota-range", adminToken, `{"min":100,"max":10000}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set quota range expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = settingsIntegrationRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/default-quota", adminToken, `{"value":99}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("default quota below range expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	// Default quota applies to new tokens without an explicit limit and does
	// not override an explicit one.
	rec = settingsIntegrationRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/default-quota", adminToken, `{"value":4321}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set default quota expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	createTokenForSettings(t, engine, adminToken, userID, `{"name":"defaulted"}`)
	createTokenForSettings(t, engine, adminToken, userID, `{"name":"explicit","quota_limit":99}`)

	rec = settingsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/admin/users/"+userID.String()+"/tokens", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list tokens expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tokensResp struct {
		Data struct {
			List []service.UserTokenResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &tokensResp); err != nil {
		t.Fatalf("decode tokens: %v", err)
	}
	found := map[string]*int64{}
	for _, tok := range tokensResp.Data.List {
		name := tok.Name
		found[name] = tok.QuotaLimit
	}
	if got := found["defaulted"]; got == nil || *got != 4321 {
		t.Fatalf("expected defaulted token quota 4321, got %v", got)
	}
	if got := found["explicit"]; got == nil || *got != 99 {
		t.Fatalf("expected explicit token quota 99, got %v", got)
	}
	if got := found["primary"]; got == nil || *got != 1000 {
		t.Fatalf("expected primary token quota 1000, got %v", got)
	}
}

type settingsTokenResponse struct {
	ID        string `json:"id"`
	Plaintext string `json:"plaintext_token"`
}

func createTokenForSettings(t *testing.T, engine *gin.Engine, adminToken string, userID uuid.UUID, body string) settingsTokenResponse {
	t.Helper()
	rec := settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/admin/users/"+userID.String()+"/tokens", adminToken, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create token expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data settingsTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	if resp.Data.Plaintext == "" {
		t.Fatal("expected plaintext token")
	}
	return resp.Data
}
