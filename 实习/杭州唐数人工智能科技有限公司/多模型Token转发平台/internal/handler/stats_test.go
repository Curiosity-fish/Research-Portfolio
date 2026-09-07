package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockStatsService struct {
	dashboardStatsFunc func(ctx context.Context, start, end *time.Time) (*service.DashboardStatsResponse, error)
	usageByModelFunc   func(ctx context.Context, start, end *time.Time) ([]service.UsageByModelResponse, error)
	usageByUserFunc    func(ctx context.Context, start, end *time.Time) ([]service.UsageByUserResponse, error)
	usageByDayFunc     func(ctx context.Context, start, end *time.Time) ([]service.UsageByDayResponse, error)
	userStatsFunc      func(ctx context.Context, userID uuid.UUID, start, end *time.Time) (*service.UserStatsResponse, error)
	rankingsFunc       func(ctx context.Context, days int, metric string, limit int) ([]service.RankingRowResponse, error)
}

func (m *mockStatsService) DashboardStats(ctx context.Context, start, end *time.Time) (*service.DashboardStatsResponse, error) {
	return m.dashboardStatsFunc(ctx, start, end)
}

func (m *mockStatsService) UsageByModel(ctx context.Context, start, end *time.Time) ([]service.UsageByModelResponse, error) {
	return m.usageByModelFunc(ctx, start, end)
}

func (m *mockStatsService) UsageByUser(ctx context.Context, start, end *time.Time) ([]service.UsageByUserResponse, error) {
	return m.usageByUserFunc(ctx, start, end)
}

func (m *mockStatsService) UsageByDay(ctx context.Context, start, end *time.Time) ([]service.UsageByDayResponse, error) {
	return m.usageByDayFunc(ctx, start, end)
}

func (m *mockStatsService) UserStats(ctx context.Context, userID uuid.UUID, start, end *time.Time) (*service.UserStatsResponse, error) {
	return m.userStatsFunc(ctx, userID, start, end)
}

func (m *mockStatsService) Rankings(ctx context.Context, days int, metric string, limit int) ([]service.RankingRowResponse, error) {
	if m.rankingsFunc == nil {
		return nil, nil
	}
	return m.rankingsFunc(ctx, days, metric, limit)
}

func (m *mockStatsService) ModelDist(ctx context.Context, days, limit int) ([]service.ModelDistItemResponse, error) {
	return nil, nil
}

func (m *mockStatsService) DeptDist(ctx context.Context, days, limit int) ([]service.DeptDistItemResponse, error) {
	return nil, nil
}

func (m *mockStatsService) HourlyTrend(ctx context.Context, userID *uuid.UUID) ([]service.TrendPointResponse, error) {
	return nil, nil
}

func (m *mockStatsService) DailyTrend(ctx context.Context, userID *uuid.UUID, days int) ([]service.TrendPointResponse, error) {
	return nil, nil
}

func (m *mockStatsService) FinanceSummary(ctx context.Context, days int) (*service.FinanceSummaryResponse, error) {
	return nil, nil
}

func (m *mockStatsService) UserModelStats(ctx context.Context, userID uuid.UUID, days, limit int) ([]service.ModelDistItemResponse, error) {
	return nil, nil
}

func newStatsEngine(t *testing.T, svc StatsService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewStatsHandler(svc)

	userAuth := func(c *gin.Context) {
		auth.SetAPIKeyContext(c, &auth.APIKeyInfo{UserID: uuid.MustParse("11111111-1111-1111-1111-111111111111")})
		c.Next()
	}

	engine.GET("/api/v1/admin/dashboard/stats", handler.DashboardStats)
	engine.GET("/api/v1/admin/stats/usage", handler.AdminUsage)
	engine.GET("/api/v1/admin/stats/rankings", handler.Rankings)
	engine.GET("/api/v1/admin/stats/trend", handler.AdminTrend)
	engine.GET("/api/v1/admin/stats/model-dist", handler.ModelDist)
	engine.GET("/api/v1/admin/stats/dept-dist", handler.DeptDist)
	engine.GET("/api/v1/admin/finance/summary", handler.FinanceSummary)
	engine.GET("/api/v1/stats/usage", userAuth, handler.UserStats)
	engine.GET("/api/v1/stats/trend", userAuth, handler.UserTrend)
	engine.GET("/api/v1/stats/model-stats", userAuth, handler.UserModelStats)

	return engine
}

func TestStatsHandler_DashboardStats(t *testing.T) {
	svc := &mockStatsService{
		dashboardStatsFunc: func(ctx context.Context, start, end *time.Time) (*service.DashboardStatsResponse, error) {
			return &service.DashboardStatsResponse{Calls: 10, TotalTokens: 5000}, nil
		},
	}

	engine := newStatsEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStatsHandler_UsageByModel(t *testing.T) {
	svc := &mockStatsService{
		usageByModelFunc: func(ctx context.Context, start, end *time.Time) ([]service.UsageByModelResponse, error) {
			return []service.UsageByModelResponse{{Model: "gpt-4", Calls: 2, TotalTokens: 200}}, nil
		},
	}

	engine := newStatsEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats/usage?group_by=model", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStatsHandler_UserStats(t *testing.T) {
	svc := &mockStatsService{
		userStatsFunc: func(ctx context.Context, userID uuid.UUID, start, end *time.Time) (*service.UserStatsResponse, error) {
			return &service.UserStatsResponse{Calls: 1, TotalTokens: 100}, nil
		},
	}

	engine := newStatsEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/usage", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStatsHandler_AdminUsageInvalidGroupBy(t *testing.T) {
	svc := &mockStatsService{}
	engine := newStatsEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats/usage?group_by=unknown", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStatsHandler_ReportEndpoints(t *testing.T) {
	svc := &mockStatsService{
		rankingsFunc: func(ctx context.Context, days int, metric string, limit int) ([]service.RankingRowResponse, error) {
			if metric != "calls" {
				t.Fatalf("expected metric forwarded, got %s", metric)
			}
			return []service.RankingRowResponse{{UserID: uuid.New().String(), Username: "u1", Calls: 9}}, nil
		},
	}
	engine := newStatsEngine(t, svc)

	adminRoutes := []struct {
		path string
	}{
		{"/api/v1/admin/stats/rankings?metric=calls&days=14&limit=20"},
		{"/api/v1/admin/stats/trend?days=1"},
		{"/api/v1/admin/stats/trend?days=14"},
		{"/api/v1/admin/stats/model-dist?days=14"},
		{"/api/v1/admin/stats/dept-dist?days=14"},
		{"/api/v1/admin/finance/summary?days=30"},
	}
	for _, route := range adminRoutes {
		req := httptest.NewRequest(http.MethodGet, route.path, nil)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected 200, got %d: %s", route.path, rec.Code, rec.Body.String())
		}
	}

	// The user-facing trend endpoint requires API key identity.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/trend?days=14", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("user trend expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/stats/model-stats?days=14", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("user model-stats expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
