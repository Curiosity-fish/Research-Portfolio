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
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/platform"
	"github.com/school-api/school-api-v1/ent/usertoken"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

// buildOpsEngine assembles the Phase 5/6 routes (notifications, alerts, stats,
// audit logs, reconciliation) on a single engine backed by the test database.
func buildOpsEngine(t *testing.T, client *ent.Client) *gin.Engine {
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

	apiKeyLookup := repository.NewEntAPIKeyLookup(client)
	apiKeyValidator := auth.NewAPIKeyValidator(apiKeyLookup)
	apiKeyHandler := handler.NewAPIKeyHandler()

	rechargeOrderRepo := repository.NewEntRechargeOrderRepository(client)
	rechargeOrderService := service.NewRechargeOrderService(rechargeOrderRepo, nil)
	rechargeOrderHandler := handler.NewRechargeOrderHandler(rechargeOrderService)

	notificationRepo := repository.NewEntNotificationRepository(client)
	notificationService := service.NewNotificationService(notificationRepo, userRepo)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	alertRuleRepo := repository.NewEntAlertRuleRepository(client)
	alertRecordRepo := repository.NewEntAlertRecordRepository(client)
	alertService := service.NewAlertService(alertRuleRepo, alertRecordRepo, client)
	alertHandler := handler.NewAlertHandler(alertService)

	statsRepo := repository.NewEntStatsRepository(client)
	statsService := service.NewStatsService(statsRepo)
	statsHandler := handler.NewStatsHandler(statsService)

	auditLogRepo := repository.NewEntAuditLogRepository(client)
	auditLogService := service.NewAuditLogService(auditLogRepo)
	auditLogHandler := handler.NewAuditLogHandler(auditLogService)

	reconciliationRepo := repository.NewEntReconciliationRepository(client)
	reconciliationService := service.NewReconciliationService(reconciliationRepo)
	reconciliationHandler := handler.NewReconciliationHandler(reconciliationService)

	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)

	engine.POST("/api/v1/admin/login", authHandler.Login)

	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager), middleware.AuditMiddleware(auditLogService))
	adminGroup.POST("/users", userHandler.Create)
	adminGroup.POST("/users/:id/tokens", userTokenHandler.Create)
	adminGroup.POST("/notifications", notificationHandler.Create)
	adminGroup.GET("/notifications", notificationHandler.ListAdmin)
	adminGroup.POST("/alert-rules", alertHandler.CreateRule)
	adminGroup.POST("/alerts/evaluate", alertHandler.Evaluate)
	adminGroup.GET("/alerts", alertHandler.ListRecords)
	adminGroup.PATCH("/alerts/:id/resolve", alertHandler.ResolveRecord)
	adminGroup.GET("/dashboard/stats", statsHandler.DashboardStats)
	adminGroup.GET("/stats/usage", statsHandler.AdminUsage)
	adminGroup.GET("/audit-logs", auditLogHandler.ListAdmin)
	adminGroup.GET("/recharge-orders/:id/reconciliation", reconciliationHandler.GetReconciliation)
	adminGroup.POST("/recharge-orders/:id/refunds", reconciliationHandler.Refund)

	apiGroup := engine.Group("/api/v1", middleware.APIKeyAuth(apiKeyValidator))
	apiGroup.GET("/me", apiKeyHandler.Me)
	apiGroup.GET("/notifications", notificationHandler.ListMy)
	apiGroup.PATCH("/notifications/:id/read", notificationHandler.MarkRead)
	apiGroup.GET("/stats/usage", statsHandler.UserStats)
	apiGroup.POST("/recharge/orders", rechargeOrderHandler.Create)
	apiGroup.POST("/recharge/orders/:id/mock-callback", rechargeOrderHandler.MockCallback)

	return engine
}

// opsRequest performs one HTTP request against the integration engine and
// returns the recorded response.
func opsRequest(t *testing.T, engine *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
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

// createOpsUser creates a user and an API token via the admin API and returns
// the user ID and API key.
func createOpsUser(t *testing.T, engine *gin.Engine, adminToken, suffix string) (string, string) {
	t.Helper()

	createBody := fmt.Sprintf(`{"username":"ops-%s-%s","password":"password123","name":"Ops User","role":"student"}`, suffix, uuid.NewString()[:8])
	rec := opsRequest(t, engine, http.MethodPost, "/api/v1/admin/users", adminToken, createBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var userResp struct {
		Data service.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &userResp); err != nil {
		t.Fatalf("decode user response: %v", err)
	}

	tokenBody := `{"name":"ops"}`
	rec = opsRequest(t, engine, http.MethodPost, "/api/v1/admin/users/"+userResp.Data.ID+"/tokens", adminToken, tokenBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create token expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var tokenResp struct {
		Data handler.CreateUserTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	if tokenResp.Data.Plaintext == "" {
		t.Fatal("expected plaintext token")
	}
	return userResp.Data.ID, tokenResp.Data.Plaintext
}

func openOpsTestClient(t *testing.T) *ent.Client {
	t.Helper()

	dbURL := os.Getenv("APP_DATABASE__URL")
	if dbURL == "" {
		t.Skip("APP_DATABASE__URL not set")
	}
	client, err := repository.OpenEntClient(dbURL)
	if err != nil {
		t.Fatalf("open ent client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return client
}

type notificationListData struct {
	Total int                            `json:"total"`
	List  []service.NotificationResponse `json:"list"`
}

func listNotifications(t *testing.T, engine *gin.Engine, apiKey, query string) notificationListData {
	t.Helper()

	path := "/api/v1/notifications"
	if query != "" {
		path += "?" + query
	}
	rec := opsRequest(t, engine, http.MethodGet, path, apiKey, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list notifications expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data notificationListData `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode notifications: %v", err)
	}
	return resp.Data
}

// TestOps_NotificationFlow covers broadcast isolation, personal notifications
// and the page-2 pagination regression.
func TestOps_NotificationFlow(t *testing.T) {
	ctx := context.Background()
	client := openOpsTestClient(t)
	engine := buildOpsEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	userAID, apiKeyA := createOpsUser(t, engine, adminToken, "a")
	userBID, apiKeyB := createOpsUser(t, engine, adminToken, "b")
	_ = userBID

	// Admin creates one broadcast and one personal notification for A.
	rec := opsRequest(t, engine, http.MethodPost, "/api/v1/admin/notifications", adminToken, `{"type":"announcement","title":"系统公告","content":"全体可见"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create broadcast expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var broadcastResp struct {
		Data service.NotificationResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &broadcastResp); err != nil {
		t.Fatalf("decode broadcast: %v", err)
	}

	personalBody := fmt.Sprintf(`{"type":"system","title":"个人通知","content":"仅 A 可见","user_id":%q}`, userAID)
	rec = opsRequest(t, engine, http.MethodPost, "/api/v1/admin/notifications", adminToken, personalBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create personal expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// A sees both, B sees only the broadcast.
	dataA := listNotifications(t, engine, apiKeyA, "")
	if dataA.Total != 2 {
		t.Fatalf("expected A total 2, got %d", dataA.Total)
	}
	dataB := listNotifications(t, engine, apiKeyB, "")
	if dataB.Total != 1 {
		t.Fatalf("expected B total 1, got %d", dataB.Total)
	}

	// A marks the broadcast read.
	rec = opsRequest(t, engine, http.MethodPatch, "/api/v1/notifications/"+broadcastResp.Data.ID+"/read", apiKeyA, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("mark read expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	// A sees it read, B still sees it unread.
	dataA = listNotifications(t, engine, apiKeyA, "")
	for _, n := range dataA.List {
		if n.ID == broadcastResp.Data.ID && !n.IsRead {
			t.Fatal("expected A to see broadcast as read")
		}
	}
	dataB = listNotifications(t, engine, apiKeyB, "")
	for _, n := range dataB.List {
		if n.ID == broadcastResp.Data.ID && n.IsRead {
			t.Fatal("expected B to see broadcast as unread")
		}
	}

	// Pagination regression: page 2 must not fail on the count query.
	for i := 0; i < 3; i++ {
		personalBody := fmt.Sprintf(`{"type":"system","title":"分页通知 %d","content":"内容","user_id":%q}`, i, userAID)
		rec = opsRequest(t, engine, http.MethodPost, "/api/v1/admin/notifications", adminToken, personalBody)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create paged notification %d: %d", i, rec.Code)
		}
	}
	dataA = listNotifications(t, engine, apiKeyA, "page=2&page_size=2")
	if dataA.Total != 5 {
		t.Fatalf("expected A total 5, got %d", dataA.Total)
	}
	if len(dataA.List) != 2 {
		t.Fatalf("expected 2 items on page 2, got %d", len(dataA.List))
	}
}

// TestOps_AlertFlow covers rule creation, evaluation dedupe and resolve.
func TestOps_AlertFlow(t *testing.T) {
	ctx := context.Background()
	client := openOpsTestClient(t)
	engine := buildOpsEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	// A fresh user starts with balance 0 and therefore triggers balance_low.
	_, apiKey := createOpsUser(t, engine, adminToken, "alert")
	_ = apiKey

	ruleBody := `{"name":"余额低","metric":"balance_low","threshold":100,"enabled":true}`
	rec := opsRequest(t, engine, http.MethodPost, "/api/v1/admin/alert-rules", adminToken, ruleBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create rule expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var ruleResp struct {
		Data service.AlertRuleResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &ruleResp); err != nil {
		t.Fatalf("decode rule: %v", err)
	}

	evaluate := func() int {
		t.Helper()
		rec := opsRequest(t, engine, http.MethodPost, "/api/v1/admin/alerts/evaluate", adminToken, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("evaluate expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Data struct {
				Created int `json:"created"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode evaluate: %v", err)
		}
		return resp.Data.Created
	}

	if created := evaluate(); created < 1 {
		t.Fatalf("expected at least 1 alert record, got %d", created)
	}
	if created := evaluate(); created != 0 {
		t.Fatalf("expected duplicate evaluation to create 0 records, got %d", created)
	}

	// List unresolved records for the rule.
	rec = opsRequest(t, engine, http.MethodGet, "/api/v1/admin/alerts?rule_id="+ruleResp.Data.ID+"&is_resolved=false", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list alerts expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var listResp struct {
		Data struct {
			Total int                          `json:"total"`
			List  []service.AlertRecordResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode alerts: %v", err)
	}
	if listResp.Data.Total < 1 {
		t.Fatalf("expected at least 1 unresolved record, got %d", listResp.Data.Total)
	}

	// Resolve everything, then evaluation fires again.
	for _, record := range listResp.Data.List {
		rec = opsRequest(t, engine, http.MethodPatch, "/api/v1/admin/alerts/"+record.ID+"/resolve", adminToken, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("resolve expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	}
	if created := evaluate(); created < 1 {
		t.Fatalf("expected evaluation after resolve to create records, got %d", created)
	}
}

// TestOps_StatsFlow covers dashboard stats, usage groupings and the user-scope
// stats boundary.
func TestOps_StatsFlow(t *testing.T) {
	ctx := context.Background()
	client := openOpsTestClient(t)
	engine := buildOpsEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	userID, apiKey := createOpsUser(t, engine, adminToken, "stats")
	uid := uuid.MustParse(userID)

	// Seed usage data directly: one platform + account, three call logs.
	p, err := client.Platform.Create().
		SetName("统计平台").
		SetCode("stats-platform-" + uuid.NewString()[:8]).
		SetType(platform.TypeOpenai).
		SetBaseURL("https://example.invalid").
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}
	a, err := client.Account.Create().
		SetName("统计账号").
		SetPlatformID(p.ID).
		SetAPIKeyEncrypted("encrypted-key").
		SetStatus(account.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	tokens := client.UserToken.Query().Where(usertoken.UserIDEQ(uid)).FirstX(ctx)
	models := []string{"gpt-stats-a", "gpt-stats-a", "gpt-stats-b"}
	for _, m := range models {
		if err := client.CallLog.Create().
			SetUserID(uid).
			SetTokenID(tokens.ID).
			SetPlatformID(p.ID).
			SetAccountID(a.ID).
			SetModel(m).
			SetPromptTokens(10).
			SetCompletionTokens(20).
			SetTotalTokens(30).
			SetStatusCode(200).
			Exec(ctx); err != nil {
			t.Fatalf("create call log: %v", err)
		}
	}

	rec := opsRequest(t, engine, http.MethodGet, "/api/v1/admin/dashboard/stats", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard stats expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var dashboard struct {
		Data service.DashboardStatsResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	if dashboard.Data.Calls < 3 {
		t.Fatalf("expected at least 3 calls, got %d", dashboard.Data.Calls)
	}

	// group_by=model aggregates the two identical models.
	rec = opsRequest(t, engine, http.MethodGet, "/api/v1/admin/stats/usage?group_by=model", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("usage by model expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var byModel struct {
		Data struct {
			List []service.UsageByModelResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &byModel); err != nil {
		t.Fatalf("decode usage by model: %v", err)
	}
	found := false
	for _, row := range byModel.Data.List {
		if row.Model == "gpt-stats-a" && row.Calls == 2 && row.TotalTokens == 60 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected gpt-stats-a with 2 calls in %+v", byModel.Data.List)
	}

	rec = opsRequest(t, engine, http.MethodGet, "/api/v1/admin/stats/usage?group_by=user", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("usage by user expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = opsRequest(t, engine, http.MethodGet, "/api/v1/admin/stats/usage?group_by=day", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("usage by day expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = opsRequest(t, engine, http.MethodGet, "/api/v1/admin/stats/usage?group_by=bogus", adminToken, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("usage by bogus group expected 400, got %d", rec.Code)
	}

	// User-scope stats only count the caller's own usage.
	rec = opsRequest(t, engine, http.MethodGet, "/api/v1/stats/usage", apiKey, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("user stats expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var userStats struct {
		Data service.UserStatsResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &userStats); err != nil {
		t.Fatalf("decode user stats: %v", err)
	}
	if userStats.Data.Calls != 3 {
		t.Fatalf("expected user calls 3, got %d", userStats.Data.Calls)
	}
}

// TestOps_AuditLogFlow verifies the audit middleware records admin mutations
// and the audit log list supports pagination.
func TestOps_AuditLogFlow(t *testing.T) {
	ctx := context.Background()
	client := openOpsTestClient(t)
	engine := buildOpsEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	createOpsUser(t, engine, adminToken, "audit")

	action := url.QueryEscape("POST /api/v1/admin/users")
	rec := opsRequest(t, engine, http.MethodGet, "/api/v1/admin/audit-logs?action="+action, adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("audit logs expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var auditResp struct {
		Data struct {
			Total int                         `json:"total"`
			List  []service.AuditLogResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &auditResp); err != nil {
		t.Fatalf("decode audit logs: %v", err)
	}
	if auditResp.Data.Total < 1 {
		t.Fatalf("expected at least 1 audit log for user creation, got %d", auditResp.Data.Total)
	}
	if auditResp.Data.List[0].ActorType != "admin" {
		t.Fatalf("expected actor_type admin, got %s", auditResp.Data.List[0].ActorType)
	}

	// Pagination regression: page 2 must not fail on the count query.
	rec = opsRequest(t, engine, http.MethodGet, "/api/v1/admin/audit-logs?page=2&page_size=2", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("audit logs page 2 expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestOps_ReconciliationRefundFlow covers order reconciliation and the
// over-refund rejection.
func TestOps_ReconciliationRefundFlow(t *testing.T) {
	ctx := context.Background()
	client := openOpsTestClient(t)
	engine := buildOpsEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	userID, apiKey := createOpsUser(t, engine, adminToken, "refund")

	// Create a recharge order and pay it via the mock callback.
	rec := opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders", apiKey, `{"amount":5000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create order expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var orderResp struct {
		Data service.RechargeOrderResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &orderResp); err != nil {
		t.Fatalf("decode order: %v", err)
	}

	rec = opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders/"+orderResp.Data.ID+"/mock-callback", apiKey, `{"success":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("mock callback expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Reconciliation view shows the full refundable amount.
	rec = opsRequest(t, engine, http.MethodGet, "/api/v1/admin/recharge-orders/"+orderResp.Data.ID+"/reconciliation", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("reconciliation expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var recon struct {
		Data service.ReconciliationOrderResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &recon); err != nil {
		t.Fatalf("decode reconciliation: %v", err)
	}
	if recon.Data.Amount != 5000 || recon.Data.RefundableAmount != 5000 {
		t.Fatalf("unexpected reconciliation: %+v", recon.Data)
	}

	// Partial refund succeeds.
	rec = opsRequest(t, engine, http.MethodPost, "/api/v1/admin/recharge-orders/"+orderResp.Data.ID+"/refunds", adminToken, `{"amount":2000,"reason":"部分退款"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("refund expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var refundResp struct {
		Data service.RefundOrderResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &refundResp); err != nil {
		t.Fatalf("decode refund: %v", err)
	}
	if refundResp.Data.Order.RefundedAmount != 2000 || refundResp.Data.Order.RefundableAmount != 3000 {
		t.Fatalf("unexpected refund result: %+v", refundResp.Data.Order)
	}
	if refundResp.Data.BalanceRecord.Type != "refund" || refundResp.Data.BalanceRecord.Amount != 2000 {
		t.Fatalf("unexpected refund balance record: %+v", refundResp.Data.BalanceRecord)
	}

	// Over-refund beyond the remaining amount is rejected.
	rec = opsRequest(t, engine, http.MethodPost, "/api/v1/admin/recharge-orders/"+orderResp.Data.ID+"/refunds", adminToken, `{"amount":4000}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("over-refund expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	// User balance reflects the refund.
	u, err := client.User.Get(ctx, uuid.MustParse(userID))
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if u.Balance != 3000 {
		t.Fatalf("expected balance 3000, got %d", u.Balance)
	}
}

// TestOps_ConcurrentCallback verifies that racing mock-callbacks for the same
// order credit the balance exactly once. The in-transaction conditional update
// in MarkPaid is the only guard that serializes the races, so this case cannot
// be covered by single-connection SQLite unit tests.
func TestOps_ConcurrentCallback(t *testing.T) {
	ctx := context.Background()
	client := openOpsTestClient(t)
	engine := buildOpsEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	userID, apiKey := createOpsUser(t, engine, adminToken, "conc-pay")

	rec := opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders", apiKey, `{"amount":7000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create order expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var orderResp struct {
		Data service.RechargeOrderResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &orderResp); err != nil {
		t.Fatalf("decode order: %v", err)
	}

	const goroutines = 8
	statuses := make([]int, goroutines)
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rec := opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders/"+orderResp.Data.ID+"/mock-callback", apiKey, `{"success":true}`)
			statuses[idx] = rec.Code
		}(i)
	}
	wg.Wait()

	successes := 0
	for _, code := range statuses {
		if code == http.StatusOK {
			successes++
		}
	}
	// Every racing callback must be accepted (all idempotent), but the balance
	// must be credited exactly once.
	if successes != goroutines {
		t.Fatalf("expected all %d callbacks to return 200, got %d", goroutines, successes)
	}

	u, err := client.User.Get(ctx, uuid.MustParse(userID))
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if u.Balance != 7000 {
		t.Fatalf("expected balance credited once (7000), got %d", u.Balance)
	}

	uid := uuid.MustParse(userID)
	records, err := client.BalanceRecord.Query().
		Where(
			balancerecord.UserIDEQ(uid),
			balancerecord.TypeEQ(balancerecord.TypeRecharge),
			balancerecord.AmountEQ(7000),
		).
		All(ctx)
	if err != nil {
		t.Fatalf("list balance records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected exactly 1 recharge balance record, got %d", len(records))
	}
}

// TestOps_ConcurrentRefund verifies that racing refunds for the same order
// cannot exceed the refundable amount. Only the in-transaction
// RefundedAmountLTE condition update serializes the races, so this case needs
// a real PostgreSQL.
func TestOps_ConcurrentRefund(t *testing.T) {
	ctx := context.Background()
	client := openOpsTestClient(t)
	engine := buildOpsEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	userID, apiKey := createOpsUser(t, engine, adminToken, "conc-refund")

	rec := opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders", apiKey, `{"amount":1000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create order expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var orderResp struct {
		Data service.RechargeOrderResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &orderResp); err != nil {
		t.Fatalf("decode order: %v", err)
	}

	rec = opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders/"+orderResp.Data.ID+"/mock-callback", apiKey, `{"success":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("mock callback expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Six refunds of 400 race against a refundable amount of 1000: at most
	// two can win (800), the rest must be rejected with 400.
	const goroutines = 6
	statuses := make([]int, goroutines)
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rec := opsRequest(t, engine, http.MethodPost, "/api/v1/admin/recharge-orders/"+orderResp.Data.ID+"/refunds", adminToken, `{"amount":400}`)
			statuses[idx] = rec.Code
		}(i)
	}
	wg.Wait()

	approved := 0
	for _, code := range statuses {
		if code == http.StatusOK {
			approved++
		}
	}
	if approved > 2 {
		t.Fatalf("expected at most 2 racing refunds to succeed, got %d", approved)
	}

	order, err := client.RechargeOrder.Get(ctx, uuid.MustParse(orderResp.Data.ID))
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if order.RefundedAmount > order.Amount {
		t.Fatalf("refunded amount %d exceeds order amount %d", order.RefundedAmount, order.Amount)
	}

	// Balance invariant: paid 1000, refunded approved*400, balance must match.
	u, err := client.User.Get(ctx, uuid.MustParse(userID))
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if u.Balance != 1000-order.RefundedAmount {
		t.Fatalf("expected balance %d, got %d", 1000-order.RefundedAmount, u.Balance)
	}
}

// TestOps_MockCallbackOwnership verifies that one user cannot drive the
// mock-callback of another user's order.
func TestOps_MockCallbackOwnership(t *testing.T) {
	ctx := context.Background()
	client := openOpsTestClient(t)
	engine := buildOpsEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	_, apiKeyOwner := createOpsUser(t, engine, adminToken, "owner")
	_, apiKeyOther := createOpsUser(t, engine, adminToken, "other")

	rec := opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders", apiKeyOwner, `{"amount":1500}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create order expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var orderResp struct {
		Data service.RechargeOrderResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &orderResp); err != nil {
		t.Fatalf("decode order: %v", err)
	}

	rec = opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders/"+orderResp.Data.ID+"/mock-callback", apiKeyOther, `{"success":true}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("foreign mock-callback expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	// A repeated success callback on an already-paid order is idempotent 200
	// (payment providers retry callbacks); a failed callback on a terminal
	// order is a 409 business conflict, not a 500.
	rec = opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders/"+orderResp.Data.ID+"/mock-callback", apiKeyOwner, `{"success":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("repeat success callback expected idempotent 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = opsRequest(t, engine, http.MethodPost, "/api/v1/recharge/orders/"+orderResp.Data.ID+"/mock-callback", apiKeyOwner, `{"success":false}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("failed callback on paid order expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}
