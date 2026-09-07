package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockSettingsService struct {
	featureMode       func(ctx context.Context) (string, error)
	setFeatureMode    func(ctx context.Context, mode string) error
	defaultQuotaLimit func(ctx context.Context) (*int64, error)
	setDefaultQuota   func(ctx context.Context, value int64) error
	quotaRange        func(ctx context.Context) (*service.QuotaRange, error)
	setQuotaRange     func(ctx context.Context, rng service.QuotaRange) error
	rechargeLimits    func(ctx context.Context) (*service.RechargeLimits, error)
	setRechargeLimits func(ctx context.Context, limits service.RechargeLimits) error
	publicConfig      func(ctx context.Context) (*service.PublicRechargeConfig, error)
}

func (m *mockSettingsService) FeatureMode(ctx context.Context) (string, error) {
	return m.featureMode(ctx)
}

func (m *mockSettingsService) SetFeatureMode(ctx context.Context, mode string) error {
	return m.setFeatureMode(ctx, mode)
}

func (m *mockSettingsService) DefaultQuotaLimit(ctx context.Context) (*int64, error) {
	return m.defaultQuotaLimit(ctx)
}

func (m *mockSettingsService) SetDefaultQuotaLimit(ctx context.Context, value int64) error {
	return m.setDefaultQuota(ctx, value)
}

func (m *mockSettingsService) QuotaRange(ctx context.Context) (*service.QuotaRange, error) {
	return m.quotaRange(ctx)
}

func (m *mockSettingsService) SetQuotaRange(ctx context.Context, rng service.QuotaRange) error {
	return m.setQuotaRange(ctx, rng)
}

func (m *mockSettingsService) RechargeLimits(ctx context.Context) (*service.RechargeLimits, error) {
	return m.rechargeLimits(ctx)
}

func (m *mockSettingsService) SetRechargeLimits(ctx context.Context, limits service.RechargeLimits) error {
	return m.setRechargeLimits(ctx, limits)
}

func (m *mockSettingsService) PublicRechargeConfig(ctx context.Context) (*service.PublicRechargeConfig, error) {
	return m.publicConfig(ctx)
}

func newSettingsEngine(svc SettingsService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewSettingsHandler(svc)

	engine.GET("/api/v1/admin/settings/feature-mode", h.GetFeatureMode)
	engine.PUT("/api/v1/admin/settings/feature-mode", h.PutFeatureMode)
	engine.GET("/api/v1/admin/settings/default-quota", h.GetDefaultQuota)
	engine.PUT("/api/v1/admin/settings/default-quota", h.PutDefaultQuota)
	engine.GET("/api/v1/admin/settings/quota-range", h.GetQuotaRange)
	engine.PUT("/api/v1/admin/settings/quota-range", h.PutQuotaRange)
	engine.GET("/api/v1/admin/settings/recharge", h.GetRechargeLimits)
	engine.PUT("/api/v1/admin/settings/recharge", h.PutRechargeLimits)
	engine.GET("/api/v1/recharge/config", h.GetPublicRechargeConfig)

	return engine
}

func settingsRequest(t *testing.T, engine *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestSettingsHandler_FeatureModeRoundTrip(t *testing.T) {
	stored := service.FeatureModeBoth
	svc := &mockSettingsService{
		featureMode: func(context.Context) (string, error) { return stored, nil },
		setFeatureMode: func(_ context.Context, mode string) error {
			stored = mode
			return nil
		},
	}
	engine := newSettingsEngine(svc)

	rec := settingsRequest(t, engine, http.MethodGet, "/api/v1/admin/settings/feature-mode", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = settingsRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/feature-mode", `{"mode":"quota_only"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("put: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if stored != service.FeatureModeQuotaOnly {
		t.Fatalf("expected mode persisted, got %s", stored)
	}
}

func TestSettingsHandler_PutFeatureModeValidation(t *testing.T) {
	svc := &mockSettingsService{
		setFeatureMode: func(context.Context, string) error { return nil },
	}
	engine := newSettingsEngine(svc)

	rec := settingsRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/feature-mode", `{"mode":""}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty mode, got %d", rec.Code)
	}
	rec = settingsRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/feature-mode", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing mode, got %d", rec.Code)
	}
}

func TestSettingsHandler_DefaultQuotaNilShownAsZero(t *testing.T) {
	svc := &mockSettingsService{
		defaultQuotaLimit: func(context.Context) (*int64, error) { return nil, nil },
	}
	engine := newSettingsEngine(svc)

	rec := settingsRequest(t, engine, http.MethodGet, "/api/v1/admin/settings/default-quota", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"value":0`)) {
		t.Fatalf("expected value 0 in body, got %s", rec.Body.String())
	}
}

func TestSettingsHandler_PutDefaultQuotaRejectsNegative(t *testing.T) {
	called := false
	svc := &mockSettingsService{
		setDefaultQuota: func(context.Context, int64) error {
			called = true
			return nil
		},
	}
	engine := newSettingsEngine(svc)

	rec := settingsRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/default-quota", `{"value":-5}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative value, got %d", rec.Code)
	}
	if called {
		t.Fatal("service must not be called for invalid input")
	}
}

func TestSettingsHandler_PutRechargeLimitsValidation(t *testing.T) {
	calls := 0
	svc := &mockSettingsService{
		setRechargeLimits: func(context.Context, service.RechargeLimits) error {
			calls++
			return nil
		},
	}
	engine := newSettingsEngine(svc)

	// min > max passes binding and is rejected by the service instead.
	rec := settingsRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/recharge",
		`{"min_amount":20,"max_amount":10,"quick_amounts":[15]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected binding to accept min>max (service validates), got %d", rec.Code)
	}

	// empty quick amounts
	rec = settingsRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/recharge",
		`{"min_amount":1,"max_amount":10,"quick_amounts":[]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty quick amounts, got %d", rec.Code)
	}

	// non-positive quick amount
	rec = settingsRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/recharge",
		`{"min_amount":1,"max_amount":10,"quick_amounts":[0]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero quick amount, got %d", rec.Code)
	}

	if calls != 1 {
		t.Fatalf("service must only be called for the binding-valid request, got %d calls", calls)
	}
}

func TestSettingsHandler_PutRechargeLimitsServiceValidationSurfaces(t *testing.T) {
	svc := &mockSettingsService{
		setRechargeLimits: func(context.Context, service.RechargeLimits) error {
			return domain.NewValidationError(map[string]string{"min_amount": "最小充值金额不能大于最大充值金额"})
		},
	}
	engine := newSettingsEngine(svc)

	rec := settingsRequest(t, engine, http.MethodPut, "/api/v1/admin/settings/recharge",
		`{"min_amount":20,"max_amount":10,"quick_amounts":[15]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSettingsHandler_GetRechargeLimits(t *testing.T) {
	svc := &mockSettingsService{
		rechargeLimits: func(context.Context) (*service.RechargeLimits, error) {
			return &service.RechargeLimits{
				MinAmount:    2_000_000,
				MaxAmount:    20_000_000,
				QuickAmounts: []int64{2_000_000},
			}, nil
		},
	}
	engine := newSettingsEngine(svc)

	rec := settingsRequest(t, engine, http.MethodGet, "/api/v1/admin/settings/recharge", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"min_amount":2000000`)) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSettingsHandler_PublicRechargeConfig(t *testing.T) {
	svc := &mockSettingsService{
		publicConfig: func(context.Context) (*service.PublicRechargeConfig, error) {
			return &service.PublicRechargeConfig{
				MinAmount:    1_000_000,
				MaxAmount:    100_000_000,
				QuickAmounts: []int64{1_000_000},
				Channels:     []string{"mock"},
			}, nil
		},
	}
	engine := newSettingsEngine(svc)

	rec := settingsRequest(t, engine, http.MethodGet, "/api/v1/recharge/config", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"channels":["mock"]`)) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSettingsHandler_ServiceErrorPropagates(t *testing.T) {
	svc := &mockSettingsService{
		featureMode: func(context.Context) (string, error) {
			return "", domain.WrapInternal(assertErr("boom"))
		},
	}
	engine := newSettingsEngine(svc)

	rec := settingsRequest(t, engine, http.MethodGet, "/api/v1/admin/settings/feature-mode", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
