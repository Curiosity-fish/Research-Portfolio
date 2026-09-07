//go:build integration

package integration

import (
	"context"
	"database/sql"
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
	_ "github.com/jackc/pgx/v5/stdlib"

	"entgo.io/ent/dialect"
	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/platform"
	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

func buildStatsEngine(t *testing.T, client *ent.Client, db *sql.DB) *gin.Engine {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gin.SetMode(gin.TestMode)

	adminRepo := repository.NewEntAdminRepository(client)
	tokenManager := auth.NewTokenManager("stats-integration-secret-32bytes", time.Hour)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.BcryptCostForEnv("test"), time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	statsRepo := repository.NewEntStatsRepositoryWithDB(client, db, dialect.Postgres)
	statsService := service.NewStatsService(statsRepo)
	statsHandler := handler.NewStatsHandler(statsService)

	apiKeyLookup := repository.NewEntAPIKeyLookup(client)
	apiKeyValidator := auth.NewAPIKeyValidator(apiKeyLookup)

	engine := gin.New()
	engine.Use(middleware.Recovery(logger), middleware.RequestID())
	engine.POST("/api/v1/admin/login", authHandler.Login)

	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager))
	adminGroup.GET("/stats/rankings", statsHandler.Rankings)
	adminGroup.GET("/stats/trend", statsHandler.AdminTrend)
	adminGroup.GET("/stats/model-dist", statsHandler.ModelDist)
	adminGroup.GET("/stats/dept-dist", statsHandler.DeptDist)
	adminGroup.GET("/finance/summary", statsHandler.FinanceSummary)

	apiGroup := engine.Group("/api/v1", middleware.APIKeyAuth(apiKeyValidator))
	apiGroup.GET("/stats/trend", statsHandler.UserTrend)
	apiGroup.GET("/stats/model-stats", statsHandler.UserModelStats)
	return engine
}

func TestStatsReports_EndToEnd(t *testing.T) {
	dbURL := os.Getenv("APP_DATABASE__URL")
	if dbURL == "" {
		t.Skip("APP_DATABASE__URL not set")
	}
	ctx := context.Background()
	client, db, err := repository.OpenEntClientWithDB(dbURL)
	if err != nil {
		t.Fatalf("open ent client: %v", err)
	}
	defer client.Close()
	defer db.Close()
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	engine := buildStatsEngine(t, client, db)

	// Seed: two users in different departments, call logs at known hours,
	// balance records of every flow type.
	p, err := client.Platform.Create().
		SetName("stats-p").SetCode("st-"+uuid.NewString()[:6]).
		SetType(platform.TypeOpenai).
		SetBaseURL("https://api.openai.com/v1").
		SetStatus(platform.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}
	a, err := client.Account.Create().
		SetPlatform(p).SetName("stats-a").SetAPIKeyEncrypted("enc").
		SetWeight(1).SetMaxRpm(0).SetStatus(account.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	deptCS, err := client.Department.Create().SetName("计算机系").SetCode("cs-"+uuid.NewString()[:6]).Save(ctx)
	if err != nil {
		t.Fatalf("create dept cs: %v", err)
	}
	deptMA, err := client.Department.Create().SetName("数学系").SetCode("ma-"+uuid.NewString()[:6]).Save(ctx)
	if err != nil {
		t.Fatalf("create dept ma: %v", err)
	}

	mkUser := func(name string, deptID uuid.UUID, balance int64) (*ent.User, *ent.UserToken, string) {
		t.Helper()
		u, err := client.User.Create().
			SetUsername(name + "-" + uuid.NewString()[:6]).
			SetPasswordHash("h").SetName(name).
			SetRole(user.RoleStudent).SetStatus(user.StatusActive).
			SetDepartmentID(deptID).SetBalance(balance).
			Save(ctx)
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
		key, hash, err := auth.GenerateAPIKey()
		if err != nil {
			t.Fatalf("generate key: %v", err)
		}
		tok, err := client.UserToken.Create().
			SetUserID(u.ID).SetName("default").SetTokenHash(hash).
			SetTokenLast4(key[len(key)-4:]).SetIsEnabled(true).
			Save(ctx)
		if err != nil {
			t.Fatalf("create token: %v", err)
		}
		return u, tok, key
	}

	u1, tok1, key1 := mkUser("stats-cs", deptCS.ID, 700)
	u2, tok2, key2 := mkUser("stats-ma", deptMA.ID, 0)
	_ = key2

	callLogRepo := repository.NewEntCallLogRepository(client)
	now := time.Now().UTC()
	mkLog := func(u *ent.User, tok *ent.UserToken, model string, at time.Time, total int64) {
		t.Helper()
		log, err := callLogRepo.Create(ctx, repository.CreateCallLogInput{
			UserID: u.ID, TokenID: tok.ID, PlatformID: p.ID, AccountID: a.ID,
			Model: model, StatusCode: 200,
		})
		if err != nil {
			t.Fatalf("create call log: %v", err)
		}
		// Token counts and created_at are immutable through Ent; backfill via
		// raw SQL keyed by the row ID.
		if _, err := db.ExecContext(ctx,
			`UPDATE call_logs SET created_at = $1, prompt_tokens = $2, completion_tokens = $3, total_tokens = $4 WHERE id = $5`,
			at.Format(time.RFC3339Nano), total/2, total-total/2, total, log.ID,
		); err != nil {
			t.Fatalf("backfill call log: %v", err)
		}
	}

	today := now.Truncate(24 * time.Hour)
	mkLog(u1, tok1, "gpt-4", today.Add(2*time.Hour), 100)
	mkLog(u1, tok1, "gpt-4", today.Add(2*time.Hour+time.Minute), 100)
	mkLog(u1, tok1, "gpt-4o", today.Add(9*time.Hour), 300)
	mkLog(u2, tok2, "gpt-4", today.AddDate(0, 0, -1), 50)

	mkRecord := func(u *ent.User, typ balancerecord.Type, amount, after int64) {
		t.Helper()
		if _, err := client.BalanceRecord.Create().
			SetUserID(u.ID).SetType(typ).SetAmount(amount).SetBalanceAfter(after).
			Save(ctx); err != nil {
			t.Fatalf("create balance record: %v", err)
		}
	}
	mkRecord(u1, balancerecord.TypeRecharge, 1000, 1000)
	mkRecord(u1, balancerecord.TypeConsume, -300, 700)
	mkRecord(u1, balancerecord.TypeRefund, -200, 500)
	mkRecord(u1, balancerecord.TypeAdminAdjust, 200, 700)
	if _, err := client.RechargeOrder.Create().
		SetUserID(u1.ID).SetAmount(1000).
		SetStatus(rechargeorder.StatusPaid).SetProvider(rechargeorder.ProviderMock).
		Save(ctx); err != nil {
		t.Fatalf("create order: %v", err)
	}

	adminToken := createAdminAndLogin(t, ctx, client, engine)

	// Rankings: u1 must appear with its full usage. (Other tests leave their
	// own call logs in the shared database, so assert by user, not by rank.)
	rec := statsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/admin/stats/rankings?days=7&metric=tokens&limit=100", adminToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("rankings expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var rankings struct {
		Data struct {
			List []struct {
				UserID         string `json:"user_id"`
				DepartmentName string `json:"department_name"`
				Calls          int    `json:"calls"`
				TotalTokens    int64  `json:"total_tokens"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &rankings); err != nil {
		t.Fatalf("decode rankings: %v", err)
	}
	foundU1 := false
	for _, row := range rankings.Data.List {
		if row.UserID == u1.ID.String() {
			foundU1 = true
			if row.TotalTokens != 500 || row.Calls != 3 {
				t.Fatalf("unexpected u1 ranking row: %+v", row)
			}
			if row.DepartmentName != "计算机系" {
				t.Fatalf("expected department name on u1 row: %+v", row)
			}
		}
	}
	if !foundU1 {
		t.Fatalf("u1 missing from rankings: %+v", rankings.Data.List)
	}

	// Hourly trend: days=1 → 24 points, 02:00 and 09:00 populated.
	rec = statsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/admin/stats/trend?days=1", adminToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("hourly trend expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var hourly struct {
		Data struct {
			Granularity string `json:"granularity"`
			List        []struct {
				Hour        int    `json:"hour"`
				Label       string `json:"label"`
				Calls       int    `json:"calls"`
				TotalTokens int64  `json:"total_tokens"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &hourly); err != nil {
		t.Fatalf("decode hourly trend: %v", err)
	}
	if hourly.Data.Granularity != "hour" || len(hourly.Data.List) != 24 {
		t.Fatalf("expected 24 hourly points, got %+v", hourly.Data)
	}
	if hourly.Data.List[2].Calls != 2 || hourly.Data.List[2].TotalTokens != 200 {
		t.Fatalf("expected 02:00 bucket with 2 calls/200 tokens, got %+v", hourly.Data.List[2])
	}
	if hourly.Data.List[9].TotalTokens != 300 {
		t.Fatalf("expected 09:00 bucket with 300 tokens, got %+v", hourly.Data.List[9])
	}

	// Daily trend: 7 days, yesterday and today populated.
	rec = statsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/admin/stats/trend?days=7", adminToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("daily trend expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var daily struct {
		Data struct {
			Granularity string `json:"granularity"`
			List        []struct {
				Date        string `json:"date"`
				Calls       int    `json:"calls"`
				TotalTokens int64  `json:"total_tokens"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &daily); err != nil {
		t.Fatalf("decode daily trend: %v", err)
	}
	if daily.Data.Granularity != "day" || len(daily.Data.List) != 7 {
		t.Fatalf("expected 7 daily points, got %+v", daily.Data)
	}
	nonZeroDays := 0
	for _, point := range daily.Data.List {
		if point.Calls > 0 {
			nonZeroDays++
		}
	}
	if nonZeroDays != 2 {
		t.Fatalf("expected 2 non-zero days, got %d (%+v)", nonZeroDays, daily.Data.List)
	}

	// Model dist: assert by model name (other tests leave their own logs).
	rec = statsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/admin/stats/model-dist?days=7&limit=100", adminToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("model dist expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var modelDist struct {
		Data struct {
			List []struct {
				Model       string `json:"model"`
				TotalTokens int64  `json:"total_tokens"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &modelDist); err != nil {
		t.Fatalf("decode model dist: %v", err)
	}
	modelTokens := map[string]int64{}
	for _, item := range modelDist.Data.List {
		modelTokens[item.Model] = item.TotalTokens
	}
	// gpt-4: 100+100 (u1 today) + 50 (u2 yesterday); gpt-4o: 300 (u1 today).
	if modelTokens["gpt-4"] != 250 || modelTokens["gpt-4o"] != 300 {
		t.Fatalf("unexpected model dist: %+v", modelDist.Data.List)
	}

	// Dept dist: assert by department name (other tests create their own
	// departments and logs).
	rec = statsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/admin/stats/dept-dist?days=7&limit=100", adminToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("dept dist expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var deptDist struct {
		Data struct {
			List []struct {
				DepartmentName string `json:"department_name"`
				TotalTokens    int64  `json:"total_tokens"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &deptDist); err != nil {
		t.Fatalf("decode dept dist: %v", err)
	}
	deptTokens := map[string]int64{}
	for _, item := range deptDist.Data.List {
		deptTokens[item.DepartmentName] = item.TotalTokens
	}
	// 计算机系: 500 (u1), 数学系: 50 (u2).
	if deptTokens["计算机系"] != 500 || deptTokens["数学系"] != 50 {
		t.Fatalf("unexpected dept dist: %+v", deptDist.Data.List)
	}

	// Finance summary: the endpoint aggregates platform-wide sums, so other
	// tests' records pollute every total. Cross-check the response against an
	// independent ground-truth query over the same window instead of asserting
	// absolute numbers (exact aggregation semantics are covered by the repo
	// unit test on a fresh database).
	rec = statsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/admin/finance/summary?days=7", adminToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("finance summary expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var finance struct {
		Data struct {
			RechargeAmount  int64 `json:"recharge_amount"`
			ConsumeAmount   int64 `json:"consume_amount"`
			RefundAmount    int64 `json:"refund_amount"`
			AdjustAmount    int64 `json:"adjust_amount"`
			OrderCount      int   `json:"order_count"`
			ActiveUserCount int   `json:"active_user_count"`
			TotalBalance    int64 `json:"total_balance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &finance); err != nil {
		t.Fatalf("decode finance summary: %v", err)
	}

	windowNow := time.Now()
	windowStart := windowNow.AddDate(0, 0, -6).Truncate(24 * time.Hour)
	windowEnd := windowNow.Truncate(24 * time.Hour).AddDate(0, 0, 1)
	want := financeGroundTruth(t, ctx, db, windowStart, windowEnd)

	if finance.Data.RechargeAmount != want.recharge ||
		finance.Data.ConsumeAmount != want.consume ||
		finance.Data.RefundAmount != want.refund ||
		finance.Data.AdjustAmount != want.adjust {
		t.Fatalf("finance flows mismatch, got %+v want %+v", finance.Data, want)
	}
	if finance.Data.OrderCount != want.orderCount {
		t.Fatalf("expected %d paid orders, got %+v", want.orderCount, finance.Data)
	}
	if finance.Data.ActiveUserCount != want.activeUsers {
		t.Fatalf("expected %d active users, got %+v", want.activeUsers, finance.Data)
	}
	if finance.Data.TotalBalance != want.totalBalance {
		t.Fatalf("expected total balance %d, got %+v", want.totalBalance, finance.Data)
	}
	// Sanity: our own seeded records must be inside the reported aggregates.
	if want.recharge < 1000 || want.orderCount < 1 || want.totalBalance < 700 {
		t.Fatalf("seeded finance data missing from aggregates: %+v", want)
	}

	// User-side trend + model-stats (API key auth, scoped to u1).
	rec = statsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/stats/trend?days=7", key1)
	if rec.Code != http.StatusOK {
		t.Fatalf("user trend expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var userDaily struct {
		Data struct {
			List []struct {
				Date        string `json:"date"`
				Calls       int    `json:"calls"`
				TotalTokens int64  `json:"total_tokens"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &userDaily); err != nil {
		t.Fatalf("decode user trend: %v", err)
	}
	userTokens := int64(0)
	for _, point := range userDaily.Data.List {
		userTokens += point.TotalTokens
	}
	if userTokens != 500 {
		t.Fatalf("expected user tokens 500, got %d", userTokens)
	}

	rec = statsIntegrationRequest(t, engine, http.MethodGet, "/api/v1/stats/model-stats?days=7", key1)
	if rec.Code != http.StatusOK {
		t.Fatalf("user model-stats expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var userModels struct {
		Data struct {
			List []struct {
				Model       string `json:"model"`
				TotalTokens int64  `json:"total_tokens"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &userModels); err != nil {
		t.Fatalf("decode user model-stats: %v", err)
	}
	if len(userModels.Data.List) != 2 || userModels.Data.List[0].Model != "gpt-4o" || userModels.Data.List[0].TotalTokens != 300 {
		t.Fatalf("unexpected user model dist: %+v", userModels.Data.List)
	}
}

// financeGroundTruth mirrors the endpoint's aggregation semantics with
// independent SQL so the integration test stays correct while other tests
// share the same database.
type financeTruth struct {
	recharge, consume, refund, adjust int64
	orderCount, activeUsers           int
	totalBalance                      int64
}

func financeGroundTruth(t *testing.T, ctx context.Context, db *sql.DB, start, end time.Time) financeTruth {
	t.Helper()
	var truth financeTruth
	rows, err := db.QueryContext(ctx,
		`SELECT type, COALESCE(SUM(amount), 0) FROM balance_records
		 WHERE created_at >= $1 AND created_at < $2 GROUP BY type`,
		start, end)
	if err != nil {
		t.Fatalf("ground truth flows: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var typ string
		var sum int64
		if err := rows.Scan(&typ, &sum); err != nil {
			t.Fatalf("scan ground truth flow: %v", err)
		}
		switch typ {
		case string(balancerecord.TypeRecharge):
			truth.recharge = sum
		case string(balancerecord.TypeConsume):
			truth.consume = sum
		case string(balancerecord.TypeRefund):
			truth.refund = sum
		case string(balancerecord.TypeAdminAdjust):
			truth.adjust = sum
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate ground truth flows: %v", err)
	}
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM recharge_orders
		 WHERE status = $1 AND created_at >= $2 AND created_at < $3`,
		string(rechargeorder.StatusPaid), start, end).Scan(&truth.orderCount); err != nil {
		t.Fatalf("ground truth order count: %v", err)
	}
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM balance_records`).Scan(&truth.activeUsers); err != nil {
		t.Fatalf("ground truth active users: %v", err)
	}
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(balance), 0) FROM users`).Scan(&truth.totalBalance); err != nil {
		t.Fatalf("ground truth total balance: %v", err)
	}
	return truth
}

func statsIntegrationRequest(t *testing.T, engine *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}
