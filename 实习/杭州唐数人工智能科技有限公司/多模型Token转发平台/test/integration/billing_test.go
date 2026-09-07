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
	"github.com/school-api/school-api-v1/ent/quotarequest"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

func buildBillingEngine(t *testing.T, client *ent.Client) *gin.Engine {
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

	billingRepo := repository.NewEntBillingRepository(client)

	rechargeOrderRepo := repository.NewEntRechargeOrderRepository(client)
	rechargeOrderService := service.NewRechargeOrderService(rechargeOrderRepo, nil)
	rechargeOrderHandler := handler.NewRechargeOrderHandler(rechargeOrderService)

	balanceRecordRepo := repository.NewEntBalanceRecordRepository(client)
	balanceRecordService := service.NewBalanceRecordService(balanceRecordRepo, billingRepo)
	balanceRecordHandler := handler.NewBalanceRecordHandler(balanceRecordService)

	quotaRequestRepo := repository.NewEntQuotaRequestRepository(client)
	quotaRequestService := service.NewQuotaRequestService(quotaRequestRepo, userTokenRepo, nil)
	quotaRequestHandler := handler.NewQuotaRequestHandler(quotaRequestService)

	quotaRecordRepo := repository.NewEntQuotaRecordRepository(client)
	quotaRecordService := service.NewQuotaRecordService(quotaRecordRepo)
	quotaRecordHandler := handler.NewQuotaRecordHandler(quotaRecordService)

	apiKeyLookup := repository.NewEntAPIKeyLookup(client)
	apiKeyValidator := auth.NewAPIKeyValidator(apiKeyLookup)
	apiKeyHandler := handler.NewAPIKeyHandler()

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
	adminGroup.POST("/users/:id/tokens", userTokenHandler.Create)
	adminGroup.POST("/users/:id/balance-adjust", balanceRecordHandler.AdminAdjust)
	adminGroup.GET("/balance-records", balanceRecordHandler.ListAdmin)
	adminGroup.GET("/quota-requests", quotaRequestHandler.ListAdmin)
	adminGroup.PATCH("/quota-requests/:id/approve", quotaRequestHandler.Approve)
	adminGroup.PATCH("/quota-requests/:id/reject", quotaRequestHandler.Reject)
	adminGroup.GET("/quota-records", quotaRecordHandler.ListAdmin)

	apiGroup := engine.Group("/api/v1", middleware.APIKeyAuth(apiKeyValidator))
	apiGroup.GET("/me", apiKeyHandler.Me)
	apiGroup.POST("/recharge/orders", rechargeOrderHandler.Create)
	apiGroup.GET("/recharge/orders", rechargeOrderHandler.List)
	apiGroup.GET("/recharge/orders/:id", rechargeOrderHandler.Get)
	apiGroup.POST("/recharge/orders/:id/mock-callback", rechargeOrderHandler.MockCallback)
	apiGroup.GET("/balance-records", balanceRecordHandler.ListMy)
	apiGroup.GET("/quota-records", quotaRecordHandler.ListMy)
	apiGroup.POST("/quota-requests", quotaRequestHandler.Create)
	apiGroup.GET("/quota-requests", quotaRequestHandler.ListMy)

	return engine
}

func TestBilling_RechargeBalanceQuotaRequestFlow(t *testing.T) {
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

	// The pending-request assertion below counts globally, so stale rows from
	// an earlier aborted run would break it. Clear them instead of relying on
	// execution order.
	if _, err := client.QuotaRequest.Delete().
		Where(quotarequest.StatusEQ(quotarequest.StatusPending)).
		Exec(ctx); err != nil {
		t.Fatalf("clear stale pending quota requests: %v", err)
	}

	engine := buildBillingEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	// Create a user to own tokens and orders.
	username := "billing-user-" + uuid.NewString()[:8]
	createBody := map[string]interface{}{
		"username": username,
		"password": "password123",
		"name":     "Billing User",
		"role":     "student",
	}
	bodyBytes, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
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
	userID := uuid.MustParse(userResp.Data.ID)
	t.Cleanup(func() {
		_ = client.User.DeleteOneID(userID).Exec(context.Background())
	})

	// Create an API token with an initial quota limit.
	tokenBody := `{"name":"billing","quota_limit":1000}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+userResp.Data.ID+"/tokens", bytes.NewReader([]byte(tokenBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
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
	apiKey := tokenResp.Data.Plaintext
	tokenID := uuid.MustParse(tokenResp.Data.ID)

	// Create a recharge order.
	req = httptest.NewRequest(http.MethodPost, "/api/v1/recharge/orders", bytes.NewReader([]byte(`{"amount":5000}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create recharge order expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var orderResp struct {
		Data service.RechargeOrderResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &orderResp); err != nil {
		t.Fatalf("decode order response: %v", err)
	}
	if orderResp.Data.Status != "pending" {
		t.Errorf("expected pending order, got %s", orderResp.Data.Status)
	}
	orderID := orderResp.Data.ID

	// Get order.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/recharge/orders/"+orderID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get order expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// List orders.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/recharge/orders", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list orders expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var listOrdersResp struct {
		Data struct {
			Total int                             `json:"total"`
			List  []service.RechargeOrderResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listOrdersResp); err != nil {
		t.Fatalf("decode list orders: %v", err)
	}
	if listOrdersResp.Data.Total != 1 {
		t.Errorf("expected 1 order, got %d", listOrdersResp.Data.Total)
	}

	// Mock payment callback success.
	req = httptest.NewRequest(http.MethodPost, "/api/v1/recharge/orders/"+orderID+"/mock-callback", bytes.NewReader([]byte(`{"success":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("mock callback expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var callbackResp struct {
		Data service.RechargeOrderResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &callbackResp); err != nil {
		t.Fatalf("decode callback response: %v", err)
	}
	if callbackResp.Data.Status != "paid" {
		t.Errorf("expected paid order, got %s", callbackResp.Data.Status)
	}

	// Verify balance records from user's perspective.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/balance-records", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list balance records expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var balanceResp struct {
		Data struct {
			Total int                           `json:"total"`
			List  []service.BalanceRecordResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &balanceResp); err != nil {
		t.Fatalf("decode balance records: %v", err)
	}
	if balanceResp.Data.Total != 1 {
		t.Errorf("expected 1 balance record, got %d", balanceResp.Data.Total)
	}
	if len(balanceResp.Data.List) != 1 || balanceResp.Data.List[0].Type != "recharge" || balanceResp.Data.List[0].Amount != 5000 {
		t.Errorf("unexpected balance record: %+v", balanceResp.Data.List)
	}

	// Verify user's balance in database.
	u, err := client.User.Get(ctx, userID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if u.Balance != 5000 {
		t.Errorf("expected user balance 5000, got %d", u.Balance)
	}

	// Admin deducts balance.
	adjustBody := `{"amount":-1000,"remark":"test deduction"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+userResp.Data.ID+"/balance-adjust", bytes.NewReader([]byte(adjustBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("balance adjust expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	u, err = client.User.Get(ctx, userID)
	if err != nil {
		t.Fatalf("get user after adjust: %v", err)
	}
	if u.Balance != 4000 {
		t.Errorf("expected user balance 4000, got %d", u.Balance)
	}

	// Admin list balance records should include both recharge and admin_adjust.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/balance-records?user_id="+userResp.Data.ID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin list balance records expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var adminBalanceResp struct {
		Data struct {
			Total int                           `json:"total"`
			List  []service.BalanceRecordResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &adminBalanceResp); err != nil {
		t.Fatalf("decode admin balance records: %v", err)
	}
	if adminBalanceResp.Data.Total != 2 {
		t.Errorf("expected 2 balance records, got %d", adminBalanceResp.Data.Total)
	}

	// Submit a quota request.
	quotaBody := map[string]interface{}{
		"token_id":         tokenResp.Data.ID,
		"requested_amount": 500,
		"reason":           "need more quota",
	}
	bodyBytes, _ = json.Marshal(quotaBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/quota-requests", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create quota request expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var quotaReqResp struct {
		Data service.QuotaRequestResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &quotaReqResp); err != nil {
		t.Fatalf("decode quota request: %v", err)
	}
	if quotaReqResp.Data.Status != "pending" {
		t.Errorf("expected pending quota request, got %s", quotaReqResp.Data.Status)
	}
	requestID := quotaReqResp.Data.ID

	// Admin lists pending quota requests.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/quota-requests?status=pending", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin list quota requests expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var adminQuotaList struct {
		Data struct {
			Total int                            `json:"total"`
			List  []service.QuotaRequestResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &adminQuotaList); err != nil {
		t.Fatalf("decode admin quota list: %v", err)
	}
	if adminQuotaList.Data.Total != 1 {
		t.Errorf("expected 1 pending quota request, got %d", adminQuotaList.Data.Total)
	}

	// Admin approves the quota request.
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/quota-requests/"+requestID+"/approve", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("approve quota request expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var approveResp struct {
		Data service.QuotaRequestResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &approveResp); err != nil {
		t.Fatalf("decode approve response: %v", err)
	}
	if approveResp.Data.Status != "approved" {
		t.Errorf("expected approved quota request, got %s", approveResp.Data.Status)
	}

	// Verify token quota limit increased.
	tok, err := client.UserToken.Get(ctx, tokenID)
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if tok.QuotaLimit == nil || *tok.QuotaLimit != 1500 {
		limit := int64(-1)
		if tok.QuotaLimit != nil {
			limit = *tok.QuotaLimit
		}
		t.Errorf("expected quota_limit 1500, got %d", limit)
	}

	// Verify quota records contain the approved request.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/quota-records", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list quota records expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var quotaRecordResp struct {
		Data struct {
			Total int                          `json:"total"`
			List  []service.QuotaRecordResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &quotaRecordResp); err != nil {
		t.Fatalf("decode quota records: %v", err)
	}
	if quotaRecordResp.Data.Total != 1 {
		t.Errorf("expected 1 quota record, got %d", quotaRecordResp.Data.Total)
	}
	if len(quotaRecordResp.Data.List) != 1 || quotaRecordResp.Data.List[0].Type != "request_approved" || quotaRecordResp.Data.List[0].Amount != 500 {
		t.Errorf("unexpected quota record: %+v", quotaRecordResp.Data.List)
	}
}

func TestBilling_OnlyOwnerCanAccessOrder(t *testing.T) {
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

	engine := buildBillingEngine(t, client)

	// Create two users and tokens directly to avoid relying on admin login.
	group, err := client.Group.Create().SetName("g").SetCode("g-"+uuid.NewString()[:8]).Save(ctx)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	owner, err := client.User.Create().
		SetUsername("owner-"+uuid.NewString()[:8]).
		SetPasswordHash("hash").
		SetName("Owner").
		SetRole(user.RoleStudent).
		SetStatus(user.StatusActive).
		SetGroupID(group.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	other, err := client.User.Create().
		SetUsername("other-"+uuid.NewString()[:8]).
		SetPasswordHash("hash").
		SetName("Other").
		SetRole(user.RoleStudent).
		SetStatus(user.StatusActive).
		SetGroupID(group.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("create other: %v", err)
	}
	ownerKey, ownerHash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate owner key: %v", err)
	}
	otherKey, otherHash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate other key: %v", err)
	}
	ownerToken, err := client.UserToken.Create().
		SetUserID(owner.ID).
		SetName("default").
		SetTokenHash(ownerHash).
		SetTokenLast4(ownerKey[len(ownerKey)-4:]).
		SetIsEnabled(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("create owner token: %v", err)
	}
	_, err = client.UserToken.Create().
		SetUserID(other.ID).
		SetName("default").
		SetTokenHash(otherHash).
		SetTokenLast4(otherKey[len(otherKey)-4:]).
		SetIsEnabled(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("create other token: %v", err)
	}

	order, err := client.RechargeOrder.Create().
		SetUserID(owner.ID).
		SetAmount(1000).
		SetProvider("mock").
		Save(ctx)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recharge/orders/"+order.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+otherKey)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for other user accessing order, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/recharge/orders/"+order.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+ownerKey)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for owner accessing order, got %d: %s", rec.Code, rec.Body.String())
	}

	_ = ownerToken
}

func TestBilling_MockCallbackRequiresValidOrder(t *testing.T) {
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

	engine := buildBillingEngine(t, client)

	group, err := client.Group.Create().SetName("g").SetCode("g-"+uuid.NewString()[:8]).Save(ctx)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	u, err := client.User.Create().
		SetUsername("mock-"+uuid.NewString()[:8]).
		SetPasswordHash("hash").
		SetName("Mock").
		SetRole(user.RoleStudent).
		SetStatus(user.StatusActive).
		SetGroupID(group.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	key, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	_, err = client.UserToken.Create().
		SetUserID(u.ID).
		SetName("default").
		SetTokenHash(hash).
		SetTokenLast4(key[len(key)-4:]).
		SetIsEnabled(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recharge/orders/"+uuid.NewString()+"/mock-callback", bytes.NewReader([]byte(`{"success":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown order, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBilling_CreateRechargeOrderRequiresPositiveAmount(t *testing.T) {
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

	engine := buildBillingEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	username := "billing-val-" + uuid.NewString()[:8]
	createBody := map[string]interface{}{
		"username": username,
		"password": "password123",
		"name":     "Validation User",
		"role":     "student",
	}
	bodyBytes, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
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

	tokenBody := `{"name":"billing"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+userResp.Data.ID+"/tokens", bytes.NewReader([]byte(tokenBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create token expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var tokenResp struct {
		Data handler.CreateUserTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("decode token response: %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/recharge/orders", bytes.NewReader([]byte(`{"amount":0}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenResp.Data.Plaintext)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for zero amount, got %d: %s", rec.Code, rec.Body.String())
	}
}
