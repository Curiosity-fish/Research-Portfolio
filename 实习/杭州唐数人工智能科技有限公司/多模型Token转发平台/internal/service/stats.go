package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// DashboardStatsResponse is the public representation of overall dashboard stats.
type DashboardStatsResponse struct {
	Calls            int64 `json:"calls"`
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
	RechargeAmount   int64 `json:"recharge_amount"`
	ConsumeAmount    int64 `json:"consume_amount"`
}

// UsageByModelResponse is the public representation of per-model usage.
type UsageByModelResponse struct {
	Model            string `json:"model"`
	Calls            int    `json:"calls"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
}

// UsageByUserResponse is the public representation of per-user usage.
type UsageByUserResponse struct {
	UserID      string `json:"user_id"`
	Calls       int    `json:"calls"`
	TotalTokens int64  `json:"total_tokens"`
}

// UsageByDayResponse is the public representation of per-day usage.
type UsageByDayResponse struct {
	Day         string `json:"day"`
	Calls       int64  `json:"calls"`
	TotalTokens int64  `json:"total_tokens"`
}

// UserStatsResponse is the public representation of a single user's stats.
type UserStatsResponse struct {
	Calls            int64 `json:"calls"`
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
	ConsumeAmount    int64 `json:"consume_amount"`
}

// RankingRowResponse is one row of the user leaderboard.
type RankingRowResponse struct {
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	DepartmentID   string `json:"department_id,omitempty"`
	DepartmentName string `json:"department_name"`
	Calls          int    `json:"calls"`
	TotalTokens    int64  `json:"total_tokens"`
}

// TrendPointResponse is one bucket of a usage trend.
type TrendPointResponse struct {
	Hour             int    `json:"hour,omitempty"`
	Label            string `json:"label,omitempty"`
	Date             string `json:"date,omitempty"`
	Calls            int    `json:"calls"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
}

// ModelDistItemResponse is one slice of the model distribution.
type ModelDistItemResponse struct {
	Model       string `json:"model"`
	Calls       int    `json:"calls"`
	TotalTokens int64  `json:"total_tokens"`
}

// DeptDistItemResponse is one bar of the department distribution.
type DeptDistItemResponse struct {
	DepartmentID   string `json:"department_id,omitempty"`
	DepartmentName string `json:"department_name"`
	Calls          int    `json:"calls"`
	TotalTokens    int64  `json:"total_tokens"`
}

// FinanceSummaryResponse is the platform-wide money overview for a window.
type FinanceSummaryResponse struct {
	RechargeAmount  int64  `json:"recharge_amount"`
	ConsumeAmount   int64  `json:"consume_amount"`
	RefundAmount    int64  `json:"refund_amount"`
	AdjustAmount    int64  `json:"adjust_amount"`
	OrderCount      int    `json:"order_count"`
	ActiveUserCount int    `json:"active_user_count"`
	TotalBalance    int64  `json:"total_balance"`
	PeriodStart     string `json:"period_start"`
	PeriodEnd       string `json:"period_end"`
}

// StatsService handles usage and billing statistics.
type StatsService struct {
	repo repository.StatsRepository
}

// NewStatsService creates a new StatsService.
func NewStatsService(repo repository.StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

// DashboardStats returns overall statistics for the admin dashboard.
func (s *StatsService) DashboardStats(ctx context.Context, start, end *time.Time) (*DashboardStatsResponse, error) {
	stats, err := s.repo.OverallStats(ctx, repository.StatsFilter{Start: start, End: end})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	return &DashboardStatsResponse{
		Calls:            int64(stats.Calls),
		PromptTokens:     stats.PromptTokens,
		CompletionTokens: stats.CompletionTokens,
		TotalTokens:      stats.TotalTokens,
		RechargeAmount:   stats.RechargeAmount,
		ConsumeAmount:    stats.ConsumeAmount,
	}, nil
}

// UsageByModel returns usage aggregated by model.
func (s *StatsService) UsageByModel(ctx context.Context, start, end *time.Time) ([]UsageByModelResponse, error) {
	rows, err := s.repo.UsageByModel(ctx, repository.StatsFilter{Start: start, End: end})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]UsageByModelResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, UsageByModelResponse{
			Model:            r.Model,
			Calls:            r.Calls,
			PromptTokens:     r.PromptTokens,
			CompletionTokens: r.CompletionTokens,
			TotalTokens:      r.TotalTokens,
		})
	}
	return resp, nil
}

// UsageByUser returns usage aggregated by user.
func (s *StatsService) UsageByUser(ctx context.Context, start, end *time.Time) ([]UsageByUserResponse, error) {
	rows, err := s.repo.UsageByUser(ctx, repository.StatsFilter{Start: start, End: end})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]UsageByUserResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, UsageByUserResponse{
			UserID:      r.UserID.String(),
			Calls:       r.Calls,
			TotalTokens: r.TotalTokens,
		})
	}
	return resp, nil
}

// UsageByDay returns usage aggregated by day.
func (s *StatsService) UsageByDay(ctx context.Context, start, end *time.Time) ([]UsageByDayResponse, error) {
	rows, err := s.repo.UsageByDay(ctx, repository.StatsFilter{Start: start, End: end})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]UsageByDayResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, UsageByDayResponse{
			Day:         r.Day,
			Calls:       r.Calls,
			TotalTokens: r.TotalTokens,
		})
	}
	return resp, nil
}

// UserStats returns statistics for a single user.
func (s *StatsService) UserStats(ctx context.Context, userID uuid.UUID, start, end *time.Time) (*UserStatsResponse, error) {
	stats, err := s.repo.OverallStats(ctx, repository.StatsFilter{Start: start, End: end, UserID: &userID})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	return &UserStatsResponse{
		Calls:            int64(stats.Calls),
		PromptTokens:     stats.PromptTokens,
		CompletionTokens: stats.CompletionTokens,
		TotalTokens:      stats.TotalTokens,
		ConsumeAmount:    stats.ConsumeAmount,
	}, nil
}

// Clamp bounds used by the trend/distribution windows.
const (
	statsMinDays = 7
	statsMaxDays = 90
)

// normalizeStatsDays clamps the requested window to [statsMinDays, statsMaxDays].
func normalizeStatsDays(days int) int {
	if days < statsMinDays {
		return statsMinDays
	}
	if days > statsMaxDays {
		return statsMaxDays
	}
	return days
}

// Rankings returns the user leaderboard over the last `days` days. metric is
// "tokens" (default) or "calls"; limit is clamped to [1, 100].
func (s *StatsService) Rankings(ctx context.Context, days int, metric string, limit int) ([]RankingRowResponse, error) {
	days = normalizeStatsDays(days)
	if metric != "calls" && metric != "tokens" {
		return nil, domain.NewValidationError(map[string]string{"metric": "metric 必须是 calls 或 tokens"})
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	now := time.Now()
	start := now.AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	end := now.Truncate(24*time.Hour).AddDate(0, 0, 1)
	rows, err := s.repo.Rankings(ctx, repository.StatsFilter{Start: &start, End: &end}, metric, limit)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]RankingRowResponse, 0, len(rows))
	for _, r := range rows {
		row := RankingRowResponse{
			UserID:         r.UserID.String(),
			Username:       r.Username,
			Name:           r.Name,
			DepartmentName: r.DepartmentName,
			Calls:          r.Calls,
			TotalTokens:    r.TotalTokens,
		}
		if r.DepartmentID != uuid.Nil {
			row.DepartmentID = r.DepartmentID.String()
		}
		resp = append(resp, row)
	}
	return resp, nil
}

// ModelDist returns the top models by tokens over the last `days` days.
func (s *StatsService) ModelDist(ctx context.Context, days, limit int) ([]ModelDistItemResponse, error) {
	days = normalizeStatsDays(days)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	now := time.Now()
	start := now.AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	end := now.Truncate(24*time.Hour).AddDate(0, 0, 1)
	rows, err := s.repo.ModelDist(ctx, repository.StatsFilter{Start: &start, End: &end}, limit)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]ModelDistItemResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, ModelDistItemResponse{
			Model:       r.Model,
			Calls:       r.Calls,
			TotalTokens: r.TotalTokens,
		})
	}
	return resp, nil
}

// DeptDist returns usage aggregated by user department over the last `days` days.
func (s *StatsService) DeptDist(ctx context.Context, days, limit int) ([]DeptDistItemResponse, error) {
	days = normalizeStatsDays(days)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	now := time.Now()
	start := now.AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	end := now.Truncate(24*time.Hour).AddDate(0, 0, 1)
	rows, err := s.repo.DeptDist(ctx, repository.StatsFilter{Start: &start, End: &end}, limit)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]DeptDistItemResponse, 0, len(rows))
	for _, r := range rows {
		item := DeptDistItemResponse{
			DepartmentName: r.DepartmentName,
			Calls:          r.Calls,
			TotalTokens:    r.TotalTokens,
		}
		if r.DepartmentID != nil {
			item.DepartmentID = r.DepartmentID.String()
		}
		resp = append(resp, item)
	}
	return resp, nil
}

// UserModelStats returns the model distribution for a single user over the
// last `days` days.
func (s *StatsService) UserModelStats(ctx context.Context, userID uuid.UUID, days, limit int) ([]ModelDistItemResponse, error) {
	days = normalizeStatsDays(days)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	now := time.Now()
	start := now.AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	end := now.Truncate(24*time.Hour).AddDate(0, 0, 1)
	rows, err := s.repo.UserModelDist(ctx, repository.StatsFilter{Start: &start, End: &end, UserID: &userID}, limit)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]ModelDistItemResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, ModelDistItemResponse{
			Model:       r.Model,
			Calls:       r.Calls,
			TotalTokens: r.TotalTokens,
		})
	}
	return resp, nil
}

// HourlyTrend returns today's usage bucketed by hour (UTC). The caller passes
// days=1 from the request; any other value falls back to the daily view.
func (s *StatsService) HourlyTrend(ctx context.Context, userID *uuid.UUID) ([]TrendPointResponse, error) {
	now := time.Now()
	start := now.Truncate(24 * time.Hour)
	end := start.AddDate(0, 0, 1)

	buckets, err := s.repo.HourlyTrend(ctx, repository.StatsFilter{Start: &start, End: &end, UserID: userID})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	byHour := make(map[int]repository.HourBucket, len(buckets))
	for _, b := range buckets {
		byHour[b.Hour] = b
	}

	resp := make([]TrendPointResponse, 0, 24)
	for h := 0; h < 24; h++ {
		point := TrendPointResponse{
			Hour:  h,
			Label: fmt.Sprintf("%02d:00", h),
		}
		if b, ok := byHour[h]; ok {
			point.Calls = b.Calls
			point.PromptTokens = b.PromptTokens
			point.CompletionTokens = b.CompletionTokens
			point.TotalTokens = b.TotalTokens
		}
		resp = append(resp, point)
	}
	return resp, nil
}

// DailyTrend returns usage bucketed by day (UTC) over the last `days` days,
// with zero-filled buckets so the chart renders a continuous line.
func (s *StatsService) DailyTrend(ctx context.Context, userID *uuid.UUID, days int) ([]TrendPointResponse, error) {
	days = normalizeStatsDays(days)
	now := time.Now()
	end := now.Truncate(24*time.Hour).AddDate(0, 0, 1)
	start := end.AddDate(0, 0, -days)

	buckets, err := s.repo.DailyTrend(ctx, repository.StatsFilter{Start: &start, End: &end, UserID: userID})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	byDay := make(map[string]repository.DayBucket, len(buckets))
	for _, b := range buckets {
		byDay[b.Day] = b
	}

	resp := make([]TrendPointResponse, 0, days)
	for i := days - 1; i >= 0; i-- {
		day := end.AddDate(0, 0, -i-1).Format("2006-01-02")
		point := TrendPointResponse{Date: day}
		if b, ok := byDay[day]; ok {
			point.Calls = b.Calls
			point.PromptTokens = b.PromptTokens
			point.CompletionTokens = b.CompletionTokens
			point.TotalTokens = b.TotalTokens
		}
		resp = append(resp, point)
	}
	return resp, nil
}

// FinanceSummary returns the platform-wide money overview for the last `days` days.
func (s *StatsService) FinanceSummary(ctx context.Context, days int) (*FinanceSummaryResponse, error) {
	days = normalizeStatsDays(days)
	now := time.Now()
	start := now.AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	end := now.Truncate(24*time.Hour).AddDate(0, 0, 1)

	summary, err := s.repo.FinanceSummary(ctx, repository.StatsFilter{Start: &start, End: &end})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	return &FinanceSummaryResponse{
		RechargeAmount:  summary.RechargeAmount,
		ConsumeAmount:   summary.ConsumeAmount,
		RefundAmount:    summary.RefundAmount,
		AdjustAmount:    summary.AdjustAmount,
		OrderCount:      summary.OrderCount,
		ActiveUserCount: summary.ActiveUserCount,
		TotalBalance:    summary.TotalBalance,
		PeriodStart:     start.Format("2006-01-02"),
		PeriodEnd:       end.AddDate(0, 0, -1).Format("2006-01-02"),
	}, nil
}
