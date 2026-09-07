//go:build integration

package integration

import (
	"context"
	"encoding/json"
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

func buildQuotaBatchEngine(t *testing.T, client *ent.Client) *gin.Engine {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	adminRepo := repository.NewEntAdminRepository(client)
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.BcryptCostForEnv("test"), time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	settingRepo := repository.NewEntSettingRepository(client)
	settingsService := service.NewSettingsService(settingRepo)
	settingsHandler := handler.NewSettingsHandler(settingsService)

	quotaBatchRepo := repository.NewEntQuotaBatchRepository(client)
	quotaBatchService := service.NewQuotaBatchService(quotaBatchRepo, settingsService)
	quotaBatchHandler := handler.NewQuotaBatchHandler(quotaBatchService)

	quotaRecordRepo := repository.NewEntQuotaRecordRepository(client)
	quotaRecordService := service.NewQuotaRecordService(quotaRecordRepo)
	quotaRecordHandler := handler.NewQuotaRecordHandler(quotaRecordService)

	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)

	engine.POST("/api/v1/admin/login", authHandler.Login)

	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager))
	adminGroup.PUT("/settings/quota-range", settingsHandler.PutQuotaRange)
	adminGroup.POST("/quota-batch", quotaBatchHandler.Apply)
	adminGroup.GET("/quota-records", quotaRecordHandler.ListAdmin)

	return engine
}

// quotaBatchRequest reuses the settings integration request helper (same
// envelope and auth header conventions).
func quotaBatchRequest(t *testing.T, engine *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	return settingsIntegrationRequest(t, engine, method, path, token, body)
}

func TestQuotaBatch_SetAndAddByUsersAndDepartment(t *testing.T) {
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

	t.Cleanup(func() {
		_, _ = client.Setting.Delete().
			Where(setting.KeyEQ(service.SettingKeyQuotaRange)).
			Exec(context.Background())
	})

	engine := buildQuotaBatchEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	// Quota range the subsequent values must respect.
	rec := quotaBatchRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/quota-range", adminToken,
		`{"min":1000,"max":10000}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set quota range expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Two students in one class; the second one has no token yet.
	dep, err := client.Department.Create().
		SetName("批量测试班-" + uuid.NewString()[:8]).
		SetCode("batch-" + uuid.NewString()[:8]).
		Save(ctx)
	if err != nil {
		t.Fatalf("create department: %v", err)
	}
	t.Cleanup(func() { _ = client.Department.DeleteOneID(dep.ID).Exec(context.Background()) })

	user1 := createQuotaBatchUser(t, ctx, client, dep.ID)
	user2 := createQuotaBatchUser(t, ctx, client, dep.ID)

	tok1, err := client.UserToken.Create().
		SetUserID(user1).
		SetName("primary").
		SetTokenHash(uuid.NewString()).
		SetTokenLast4("abcd").
		SetQuotaLimit(1000).
		Save(ctx)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	// Set mode: below-range value rejected with 400.
	rec = quotaBatchRequest(t, engine, http.MethodPost, "/api/v1/admin/quota-batch", adminToken,
		`{"user_ids":["`+user1.String()+`"],"mode":"set","value":999}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("set below range expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	// Both targets at once is rejected.
	rec = quotaBatchRequest(t, engine, http.MethodPost, "/api/v1/admin/quota-batch", adminToken,
		`{"user_ids":["`+user1.String()+`"],"department_id":"`+dep.ID.String()+`","mode":"set","value":5000}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("dual target expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	// Set by user IDs: user1's token is updated, user2 (no token) is skipped.
	rec = quotaBatchRequest(t, engine, http.MethodPost, "/api/v1/admin/quota-batch", adminToken,
		`{"user_ids":["`+user1.String()+`","`+user2.String()+`"],"mode":"set","value":5000}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set by user ids expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var setResult struct {
		Data repository.QuotaBatchResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &setResult); err != nil {
		t.Fatalf("decode set result: %v; body=%s", err, rec.Body.String())
	}
	if setResult.Data.MatchedUsers != 2 || setResult.Data.UpdatedTokens != 1 || setResult.Data.SkippedUsers != 1 {
		t.Fatalf("unexpected set result: %+v", setResult.Data)
	}

	after, err := client.UserToken.Get(ctx, tok1.ID)
	if err != nil {
		t.Fatalf("reload token: %v", err)
	}
	if after.QuotaLimit == nil || *after.QuotaLimit != 5000 {
		t.Fatalf("expected limit 5000 after set, got %v", after.QuotaLimit)
	}

	// Add by department: -2000 lands within the range.
	rec = quotaBatchRequest(t, engine, http.MethodPost, "/api/v1/admin/quota-batch", adminToken,
		`{"department_id":"`+dep.ID.String()+`","mode":"add","value":-2000}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("add by department expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	after, err = client.UserToken.Get(ctx, tok1.ID)
	if err != nil {
		t.Fatalf("reload token after add: %v", err)
	}
	if after.QuotaLimit == nil || *after.QuotaLimit != 3000 {
		t.Fatalf("expected limit 3000 after add, got %v", after.QuotaLimit)
	}

	// The admin quota-record list shows one admin_adjust row per change.
	rec = quotaBatchRequest(t, engine, http.MethodGet,
		"/api/v1/admin/quota-records?user_id="+user1.String()+"&page_size=10", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list quota records expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var records struct {
		Data struct {
			List []struct {
				Type   string `json:"type"`
				Amount int64  `json:"amount"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &records); err != nil {
		t.Fatalf("decode quota records: %v; body=%s", err, rec.Body.String())
	}
	if len(records.Data.List) != 2 {
		t.Fatalf("expected 2 quota records, got %d", len(records.Data.List))
	}
	for _, r := range records.Data.List {
		if r.Type != "admin_adjust" || r.Amount <= 0 {
			t.Fatalf("unexpected record: %+v", r)
		}
	}
}

func createQuotaBatchUser(t *testing.T, ctx context.Context, client *ent.Client, depID uuid.UUID) uuid.UUID {
	t.Helper()
	u, err := client.User.Create().
		SetUsername("batch-user-" + uuid.NewString()[:8]).
		SetPasswordHash("hash").
		SetName("批量用户").
		SetDepartmentID(depID).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _ = client.User.DeleteOneID(u.ID).Exec(context.Background()) })
	return u.ID
}
