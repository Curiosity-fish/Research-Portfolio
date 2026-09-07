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
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/ent/refundrequest"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

func buildRefundRequestEngine(t *testing.T, client *ent.Client) *gin.Engine {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gin.SetMode(gin.TestMode)

	adminRepo := repository.NewEntAdminRepository(client)
	tokenManager := auth.NewTokenManager("refund-integration-secret-32bytes", time.Hour)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.BcryptCostForEnv("test"), time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	refundRequestRepo := repository.NewEntRefundRequestRepository(client)
	reconciliationRepo := repository.NewEntReconciliationRepository(client)
	refundRequestService := service.NewRefundRequestService(refundRequestRepo, reconciliationRepo)
	refundRequestHandler := handler.NewRefundRequestHandler(refundRequestService)

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
	adminGroup.GET("/refund-requests", refundRequestHandler.ListAdmin)
	adminGroup.POST("/refund-requests/:id/approve", refundRequestHandler.Approve)
	adminGroup.POST("/refund-requests/:id/reject", refundRequestHandler.Reject)

	apiGroup := engine.Group("/api/v1", middleware.APIKeyAuth(apiKeyValidator))
	apiGroup.POST("/refund-requests", refundRequestHandler.Create)
	apiGroup.GET("/refund-requests", refundRequestHandler.ListMy)
	apiGroup.DELETE("/refund-requests/:id", refundRequestHandler.Cancel)

	return engine
}

// seedRefundUser creates a group, a user with the given balance, an API key,
// and one paid recharge order of orderAmount.
func seedRefundUser(t *testing.T, ctx context.Context, client *ent.Client, prefix string, balance, orderAmount int64) (userID uuid.UUID, apiKey string, orderID uuid.UUID) {
	t.Helper()

	group, err := client.Group.Create().SetName("g").SetCode("g-" + uuid.NewString()[:8]).Save(ctx)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	u, err := client.User.Create().
		SetUsername(prefix + "-" + uuid.NewString()[:8]).
		SetPasswordHash("hash").
		SetName(prefix).
		SetRole(user.RoleStudent).
		SetStatus(user.StatusActive).
		SetGroupID(group.ID).
		SetBalance(balance).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _ = client.User.DeleteOneID(u.ID).Exec(context.Background()) })

	key, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}
	if _, err := client.UserToken.Create().
		SetUserID(u.ID).
		SetName("default").
		SetTokenHash(hash).
		SetTokenLast4(key[len(key)-4:]).
		SetIsEnabled(true).
		Save(ctx); err != nil {
		t.Fatalf("create token: %v", err)
	}

	order, err := client.RechargeOrder.Create().
		SetUserID(u.ID).
		SetAmount(orderAmount).
		SetStatus(rechargeorder.StatusPaid).
		SetProvider(rechargeorder.ProviderMock).
		Save(ctx)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	return u.ID, key, order.ID
}

func refundRequestHTTP(t *testing.T, engine *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func createRefundRequest(t *testing.T, engine *gin.Engine, apiKey string, orderID uuid.UUID, amount int64) string {
	t.Helper()
	rec := refundRequestHTTP(t, engine, http.MethodPost, "/api/v1/refund-requests", apiKey,
		fmt.Sprintf(`{"order_id":%q,"amount":%d,"reason":"测试退款"}`, orderID.String(), amount))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create refund request expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data service.RefundRequestResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if resp.Data.Status != "pending" {
		t.Fatalf("expected pending, got %s", resp.Data.Status)
	}
	return resp.Data.ID
}

// TestRefundRequest_ApproveConcurrencyAndIdempotency is the SPEC Task 36
// acceptance: concurrent approvals of one request settle it exactly once and
// move money exactly once.
func TestRefundRequest_ApproveConcurrencyAndIdempotency(t *testing.T) {
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

	engine := buildRefundRequestEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	userID, apiKey, orderID := seedRefundUser(t, ctx, client, "refund-conc", 1000, 1000)
	requestID := createRefundRequest(t, engine, apiKey, orderID, 400)

	// A duplicate pending request for the same order is rejected.
	rec := refundRequestHTTP(t, engine, http.MethodPost, "/api/v1/refund-requests", apiKey,
		fmt.Sprintf(`{"order_id":%q,"amount":100}`, orderID.String()))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate pending request, got %d: %s", rec.Code, rec.Body.String())
	}

	// Admin sees the pending request.
	rec = refundRequestHTTP(t, engine, http.MethodGet, "/api/v1/admin/refund-requests?status=pending", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("admin list expected 200, got %d", rec.Code)
	}
	var listResp struct {
		Data struct {
			Total int                              `json:"total"`
			List  []service.RefundRequestResponse  `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode admin list: %v", err)
	}
	found := false
	for _, item := range listResp.Data.List {
		if item.ID == requestID {
			found = true
		}
	}
	if !found {
		t.Fatal("admin pending list does not contain the request")
	}

	// Fire 8 concurrent approvals; exactly one may win.
	const approvers = 8
	codes := make([]int, approvers)
	var wg sync.WaitGroup
	for i := 0; i < approvers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rec := refundRequestHTTP(t, engine, http.MethodPost, "/api/v1/admin/refund-requests/"+requestID+"/approve", adminToken, "")
			codes[i] = rec.Code
		}(i)
	}
	wg.Wait()

	okCount := 0
	for _, code := range codes {
		if code == http.StatusOK {
			okCount++
		} else if code != http.StatusConflict {
			t.Fatalf("unexpected status %d in concurrent approvals (codes=%v)", code, codes)
		}
	}
	if okCount != 1 {
		t.Fatalf("expected exactly one successful approval, got %d (codes=%v)", okCount, codes)
	}

	// Money moved exactly once: order refunded 400, balance 600, one record.
	order, err := client.RechargeOrder.Get(ctx, orderID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if order.RefundedAmount != 400 {
		t.Fatalf("expected refunded_amount 400, got %d", order.RefundedAmount)
	}
	u, err := client.User.Get(ctx, userID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if u.Balance != 600 {
		t.Fatalf("expected balance 600, got %d", u.Balance)
	}
	refundRecords, err := client.BalanceRecord.Query().
		Where(
			balancerecord.RelatedOrderIDEQ(orderID),
			balancerecord.TypeEQ(balancerecord.TypeRefund),
		).
		Count(ctx)
	if err != nil {
		t.Fatalf("count refund records: %v", err)
	}
	if refundRecords != 1 {
		t.Fatalf("expected exactly 1 refund record, got %d", refundRecords)
	}

	// The settled request rejects any further review.
	rec = refundRequestHTTP(t, engine, http.MethodPost, "/api/v1/admin/refund-requests/"+requestID+"/reject", adminToken, "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 rejecting a settled request, got %d", rec.Code)
	}
}

// TestRefundRequest_RejectAndCancel covers the no-money paths.
func TestRefundRequest_RejectAndCancel(t *testing.T) {
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

	engine := buildRefundRequestEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	userID, apiKey, orderID := seedRefundUser(t, ctx, client, "refund-rej", 1000, 1000)

	// Another user cannot open a request for this order (404, no leak).
	_, otherKey, _ := seedRefundUser(t, ctx, client, "refund-other", 0, 1000)
	rec := refundRequestHTTP(t, engine, http.MethodPost, "/api/v1/refund-requests", otherKey,
		fmt.Sprintf(`{"order_id":%q,"amount":100}`, orderID.String()))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-owner request, got %d", rec.Code)
	}

	// Reject path: no money moves, and the order can be requested again.
	requestID := createRefundRequest(t, engine, apiKey, orderID, 200)
	rec = refundRequestHTTP(t, engine, http.MethodPost, "/api/v1/admin/refund-requests/"+requestID+"/reject", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("reject expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	u, err := client.User.Get(ctx, userID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if u.Balance != 1000 {
		t.Fatalf("reject must not move money, balance=%d", u.Balance)
	}
	createRefundRequest(t, engine, apiKey, orderID, 300)

	// Cancel path: the owner withdraws the pending request, then can re-open.
	rec = refundRequestHTTP(t, engine, http.MethodDelete, "/api/v1/refund-requests/"+requestID, apiKey, "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 cancelling a rejected request, got %d", rec.Code)
	}

	// Cancel the second (pending) request via its id from the user's list.
	rec = refundRequestHTTP(t, engine, http.MethodGet, "/api/v1/refund-requests", apiKey, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list my requests expected 200, got %d", rec.Code)
	}
	var myResp struct {
		Data struct {
			Total int                             `json:"total"`
			List  []service.RefundRequestResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &myResp); err != nil {
		t.Fatalf("decode my list: %v", err)
	}
	var pendingID string
	for _, item := range myResp.Data.List {
		if item.Status == "pending" {
			pendingID = item.ID
		}
	}
	if pendingID == "" {
		t.Fatal("expected a pending request in the user's list")
	}
	rec = refundRequestHTTP(t, engine, http.MethodDelete, "/api/v1/refund-requests/"+pendingID, apiKey, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("cancel expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	// Cancel records a rejection with no reviewer.
	cancelled, err := client.RefundRequest.Get(ctx, uuid.MustParse(pendingID))
	if err != nil {
		t.Fatalf("get cancelled request: %v", err)
	}
	if cancelled.Status != refundrequest.StatusRejected || cancelled.ReviewedBy != nil {
		t.Fatalf("expected rejected-without-reviewer, got %s reviewer=%v", cancelled.Status, cancelled.ReviewedBy)
	}
	// The order is free for a new request again.
	createRefundRequest(t, engine, apiKey, orderID, 100)
}
