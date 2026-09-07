package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/repository"
)

type mockStatsRepository struct {
	overallStatsFunc func(ctx context.Context, filter repository.StatsFilter) (*repository.OverallStats, error)
	usageByModelFunc func(ctx context.Context, filter repository.StatsFilter) ([]repository.ModelUsageStat, error)
	usageByUserFunc  func(ctx context.Context, filter repository.StatsFilter) ([]repository.UserUsageStat, error)
	usageByDayFunc   func(ctx context.Context, filter repository.StatsFilter) ([]repository.DayUsageStat, error)
	rankingsFunc     func(ctx context.Context, filter repository.StatsFilter, metric string, limit int) ([]repository.UserRankingRow, error)
	hourlyTrendFunc  func(ctx context.Context, filter repository.StatsFilter) ([]repository.HourBucket, error)
	dailyTrendFunc   func(ctx context.Context, filter repository.StatsFilter) ([]repository.DayBucket, error)
	financeFunc      func(ctx context.Context, filter repository.StatsFilter) (*repository.FinanceSummary, error)
}

func (m *mockStatsRepository) OverallStats(ctx context.Context, filter repository.StatsFilter) (*repository.OverallStats, error) {
	return m.overallStatsFunc(ctx, filter)
}

func (m *mockStatsRepository) UsageByModel(ctx context.Context, filter repository.StatsFilter) ([]repository.ModelUsageStat, error) {
	return m.usageByModelFunc(ctx, filter)
}

func (m *mockStatsRepository) UsageByUser(ctx context.Context, filter repository.StatsFilter) ([]repository.UserUsageStat, error) {
	return m.usageByUserFunc(ctx, filter)
}

func (m *mockStatsRepository) UsageByDay(ctx context.Context, filter repository.StatsFilter) ([]repository.DayUsageStat, error) {
	return m.usageByDayFunc(ctx, filter)
}

func (m *mockStatsRepository) Rankings(ctx context.Context, filter repository.StatsFilter, metric string, limit int) ([]repository.UserRankingRow, error) {
	if m.rankingsFunc == nil {
		return nil, nil
	}
	return m.rankingsFunc(ctx, filter, metric, limit)
}

func (m *mockStatsRepository) ModelDist(ctx context.Context, filter repository.StatsFilter, limit int) ([]repository.ModelDistRow, error) {
	return nil, nil
}

func (m *mockStatsRepository) UserModelDist(ctx context.Context, filter repository.StatsFilter, limit int) ([]repository.ModelDistRow, error) {
	return nil, nil
}

func (m *mockStatsRepository) DeptDist(ctx context.Context, filter repository.StatsFilter, limit int) ([]repository.DeptDistRow, error) {
	return nil, nil
}

func (m *mockStatsRepository) HourlyTrend(ctx context.Context, filter repository.StatsFilter) ([]repository.HourBucket, error) {
	if m.hourlyTrendFunc == nil {
		return nil, nil
	}
	return m.hourlyTrendFunc(ctx, filter)
}

func (m *mockStatsRepository) DailyTrend(ctx context.Context, filter repository.StatsFilter) ([]repository.DayBucket, error) {
	if m.dailyTrendFunc == nil {
		return nil, nil
	}
	return m.dailyTrendFunc(ctx, filter)
}

func (m *mockStatsRepository) FinanceSummary(ctx context.Context, filter repository.StatsFilter) (*repository.FinanceSummary, error) {
	if m.financeFunc == nil {
		return &repository.FinanceSummary{}, nil
	}
	return m.financeFunc(ctx, filter)
}

func TestStatsService_DashboardStats(t *testing.T) {
	repo := &mockStatsRepository{
		overallStatsFunc: func(ctx context.Context, filter repository.StatsFilter) (*repository.OverallStats, error) {
			return &repository.OverallStats{Calls: 5, TotalTokens: 1000, RechargeAmount: 500, ConsumeAmount: 200}, nil
		},
	}
	svc := NewStatsService(repo)

	resp, err := svc.DashboardStats(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("dashboard stats: %v", err)
	}
	if resp.Calls != 5 {
		t.Fatalf("expected calls 5, got %d", resp.Calls)
	}
	if resp.RechargeAmount != 500 {
		t.Fatalf("expected recharge 500, got %d", resp.RechargeAmount)
	}
}

func TestStatsService_UsageByModel(t *testing.T) {
	repo := &mockStatsRepository{
		usageByModelFunc: func(ctx context.Context, filter repository.StatsFilter) ([]repository.ModelUsageStat, error) {
			return []repository.ModelUsageStat{{Model: "gpt-4", Calls: 1, TotalTokens: 100}}, nil
		},
	}
	svc := NewStatsService(repo)

	resp, err := svc.UsageByModel(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("usage by model: %v", err)
	}
	if len(resp) != 1 || resp[0].Model != "gpt-4" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestStatsService_UserStats(t *testing.T) {
	uid := uuid.New()
	repo := &mockStatsRepository{
		overallStatsFunc: func(ctx context.Context, filter repository.StatsFilter) (*repository.OverallStats, error) {
			if filter.UserID == nil || *filter.UserID != uid {
				t.Fatalf("expected user filter %s, got %v", uid, filter.UserID)
			}
			return &repository.OverallStats{Calls: 3, TotalTokens: 300, ConsumeAmount: 50}, nil
		},
	}
	svc := NewStatsService(repo)

	resp, err := svc.UserStats(context.Background(), uid, nil, nil)
	if err != nil {
		t.Fatalf("user stats: %v", err)
	}
	if resp.Calls != 3 {
		t.Fatalf("expected calls 3, got %d", resp.Calls)
	}
}

func TestStatsService_DateFilterPassThrough(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	repo := &mockStatsRepository{
		overallStatsFunc: func(ctx context.Context, filter repository.StatsFilter) (*repository.OverallStats, error) {
			if filter.Start == nil || !filter.Start.Equal(start) {
				t.Fatalf("expected start filter")
			}
			if filter.End == nil || !filter.End.Equal(end) {
				t.Fatalf("expected end filter")
			}
			return &repository.OverallStats{}, nil
		},
	}
	svc := NewStatsService(repo)

	_, err := svc.DashboardStats(context.Background(), &start, &end)
	if err != nil {
		t.Fatalf("dashboard stats: %v", err)
	}
}

func TestStatsService_DailyTrendZeroFill(t *testing.T) {
	repo := &mockStatsRepository{
		dailyTrendFunc: func(ctx context.Context, filter repository.StatsFilter) ([]repository.DayBucket, error) {
			// One bucket on the first day of the requested window; the rest of
			// the window must be zero-filled by the service.
			return []repository.DayBucket{
				{Day: filter.Start.Format("2006-01-02"), Calls: 2, TotalTokens: 200},
			}, nil
		},
	}
	svc := NewStatsService(repo)

	points, err := svc.DailyTrend(context.Background(), nil, 7)
	if err != nil {
		t.Fatalf("daily trend: %v", err)
	}
	if len(points) != 7 {
		t.Fatalf("expected 7 zero-filled points, got %d", len(points))
	}
	if points[0].Calls != 2 || points[0].TotalTokens != 200 {
		t.Fatalf("expected first bucket populated, got %+v", points[0])
	}
	for _, p := range points[1:] {
		if p.Calls != 0 || p.TotalTokens != 0 {
			t.Fatalf("expected zero-filled bucket, got %+v", p)
		}
	}
}

func TestStatsService_HourlyTrendZeroFill(t *testing.T) {
	repo := &mockStatsRepository{
		hourlyTrendFunc: func(ctx context.Context, filter repository.StatsFilter) ([]repository.HourBucket, error) {
			return []repository.HourBucket{{Hour: 5, Calls: 1, TotalTokens: 42}}, nil
		},
	}
	svc := NewStatsService(repo)

	points, err := svc.HourlyTrend(context.Background(), nil)
	if err != nil {
		t.Fatalf("hourly trend: %v", err)
	}
	if len(points) != 24 {
		t.Fatalf("expected 24 points, got %d", len(points))
	}
	if points[5].Calls != 1 || points[5].TotalTokens != 42 || points[5].Label != "05:00" {
		t.Fatalf("unexpected 05:00 point: %+v", points[5])
	}
	if points[6].Calls != 0 {
		t.Fatalf("expected zero-filled neighbour, got %+v", points[6])
	}
}

func TestStatsService_RankingsValidation(t *testing.T) {
	repo := &mockStatsRepository{}
	svc := NewStatsService(repo)

	if _, err := svc.Rankings(context.Background(), 14, "bogus", 20); err == nil {
		t.Fatal("expected validation error for unknown metric")
	}
	if _, err := svc.Rankings(context.Background(), 14, "tokens", 20); err != nil {
		t.Fatalf("expected ok for tokens metric, got %v", err)
	}
}

func TestStatsService_FinanceSummaryPassThrough(t *testing.T) {
	repo := &mockStatsRepository{
		financeFunc: func(ctx context.Context, filter repository.StatsFilter) (*repository.FinanceSummary, error) {
			return &repository.FinanceSummary{
				RechargeAmount: 1000, ConsumeAmount: -300, RefundAmount: -200,
				OrderCount: 2, ActiveUserCount: 5, TotalBalance: 9000,
			}, nil
		},
	}
	svc := NewStatsService(repo)

	resp, err := svc.FinanceSummary(context.Background(), 30)
	if err != nil {
		t.Fatalf("finance summary: %v", err)
	}
	if resp.RechargeAmount != 1000 || resp.RefundAmount != -200 || resp.OrderCount != 2 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.PeriodStart == "" || resp.PeriodEnd == "" {
		t.Fatalf("expected period bounds, got %+v", resp)
	}
}
