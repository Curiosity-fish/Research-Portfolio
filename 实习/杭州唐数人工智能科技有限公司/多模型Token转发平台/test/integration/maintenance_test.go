//go:build integration

package integration

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/ent/platform"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/config"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

func buildMaintenanceEngine(t *testing.T, client *ent.Client, retentionCfg config.RetentionConfig) *gin.Engine {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	gin.SetMode(gin.TestMode)

	adminRepo := repository.NewEntAdminRepository(client)
	tokenManager := auth.NewTokenManager("maintenance-integration-secret", time.Hour)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.BcryptCostForEnv("test"), time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	callLogRepo := repository.NewEntCallLogRepository(client)
	retention := service.NewLogRetentionService(callLogRepo, retentionCfg)
	maintenanceHandler := handler.NewMaintenanceHandler(retention)

	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)
	engine.POST("/api/v1/admin/login", authHandler.Login)

	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager))
	adminGroup.POST("/maintenance/call-logs/sweep", maintenanceHandler.SweepCallLogs)

	return engine
}

func TestMaintenance_CallLogSweep(t *testing.T) {
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

	// A second raw connection is used to backdate created_at, which Ent treats
	// as immutable.
	rawDB, err := stdsql.Open("pgx", dbURL)
	if err != nil {
		t.Fatalf("open raw db: %v", err)
	}
	defer rawDB.Close()

	// Seed FK parents.
	p, err := client.Platform.Create().
		SetName("p").SetCode("p-" + uuid.NewString()[:6]).
		SetType(platform.TypeOpenai).
		SetBaseURL("https://api.openai.com/v1").
		SetStatus(platform.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}
	t.Cleanup(func() { _ = client.Platform.DeleteOneID(p.ID).Exec(context.Background()) })

	a, err := client.Account.Create().
		SetPlatform(p).
		SetName("acc").
		SetAPIKeyEncrypted("enc").
		SetWeight(1).SetMaxRpm(0).
		SetStatus(account.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	t.Cleanup(func() { _ = client.Account.DeleteOneID(a.ID).Exec(context.Background()) })

	userRepo := repository.NewEntUserRepository(client)
	u, err := userRepo.Create(ctx, repository.CreateUserInput{
		Username: "sweep-user-" + uuid.NewString()[:6], PasswordHash: "h", Name: "sweep",
		Role: "student", Status: "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _ = client.User.DeleteOneID(u.ID).Exec(context.Background()) })
	tok, err := client.UserToken.Create().
		SetUserID(u.ID).SetName("t").SetTokenHash(uuid.NewString()).SetTokenLast4("zzzz").
		Save(ctx)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	t.Cleanup(func() { _ = client.UserToken.DeleteOneID(tok.ID).Exec(context.Background()) })

	repo := repository.NewEntCallLogRepository(client)
	mk := func(model string) *ent.CallLog {
		t.Helper()
		log, err := repo.Create(ctx, repository.CreateCallLogInput{
			UserID: u.ID, TokenID: tok.ID, PlatformID: p.ID, AccountID: a.ID,
			Model: model, StatusCode: 200,
		})
		if err != nil {
			t.Fatalf("create log %s: %v", model, err)
		}
		return log
	}
	oldLog := mk("old")    // backdated 40 days
	recentLog := mk("new") // stays recent
	t.Cleanup(func() {
		_, _ = client.CallLog.Delete().Where().Exec(context.Background())
	})

	if _, err := rawDB.ExecContext(ctx,
		"UPDATE call_logs SET created_at = $1 WHERE id = $2",
		time.Now().AddDate(0, 0, -40), oldLog.ID,
	); err != nil {
		t.Fatalf("backdate old log: %v", err)
	}

	engine := buildMaintenanceEngine(t, client, config.RetentionConfig{CallLogDays: 30, SweepInterval: time.Hour})
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	rec := settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/admin/maintenance/call-logs/sweep", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("sweep expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Deleted int `json:"deleted"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode sweep response: %v", err)
	}
	if resp.Data.Deleted < 1 {
		t.Fatalf("expected at least the old log deleted, got %d", resp.Data.Deleted)
	}

	// Old log gone, recent survives.
	if _, err := client.CallLog.Get(ctx, oldLog.ID); !ent.IsNotFound(err) {
		t.Fatalf("expected old log deleted, got err=%v", err)
	}
	if _, err := client.CallLog.Get(ctx, recentLog.ID); err != nil {
		t.Fatalf("expected recent log to survive: %v", err)
	}
}
