package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// StatsService is the subset of the stats service used by StatsHandler.
type StatsService interface {
	DashboardStats(ctx context.Context, start, end *time.Time) (*service.DashboardStatsResponse, error)
	UsageByModel(ctx context.Context, start, end *time.Time) ([]service.UsageByModelResponse, error)
	UsageByUser(ctx context.Context, start, end *time.Time) ([]service.UsageByUserResponse, error)
	UsageByDay(ctx context.Context, start, end *time.Time) ([]service.UsageByDayResponse, error)
	UserStats(ctx context.Context, userID uuid.UUID, start, end *time.Time) (*service.UserStatsResponse, error)
	Rankings(ctx context.Context, days int, metric string, limit int) ([]service.RankingRowResponse, error)
	ModelDist(ctx context.Context, days, limit int) ([]service.ModelDistItemResponse, error)
	DeptDist(ctx context.Context, days, limit int) ([]service.DeptDistItemResponse, error)
	HourlyTrend(ctx context.Context, userID *uuid.UUID) ([]service.TrendPointResponse, error)
	DailyTrend(ctx context.Context, userID *uuid.UUID, days int) ([]service.TrendPointResponse, error)
	FinanceSummary(ctx context.Context, days int) (*service.FinanceSummaryResponse, error)
	UserModelStats(ctx context.Context, userID uuid.UUID, days, limit int) ([]service.ModelDistItemResponse, error)
}

// StatsHandler handles statistics endpoints.
type StatsHandler struct {
	svc StatsService
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(svc StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

// DashboardStats handles GET /api/v1/admin/dashboard/stats.
func (h *StatsHandler) DashboardStats(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.svc.DashboardStats(c.Request.Context(), start, end)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// UsageByModel handles GET /api/v1/admin/stats/usage?group_by=model.
func (h *StatsHandler) UsageByModel(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.svc.UsageByModel(c.Request.Context(), start, end)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"list": resp})
}

// UsageByUser handles GET /api/v1/admin/stats/usage?group_by=user.
func (h *StatsHandler) UsageByUser(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.svc.UsageByUser(c.Request.Context(), start, end)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"list": resp})
}

// UsageByDay handles GET /api/v1/admin/stats/usage?group_by=day.
func (h *StatsHandler) UsageByDay(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.svc.UsageByDay(c.Request.Context(), start, end)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"list": resp})
}

// UserStats handles GET /api/v1/stats/usage.
func (h *StatsHandler) UserStats(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	start, end, err := parseDateRange(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.svc.UserStats(c.Request.Context(), userID, start, end)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

func parseDateRange(c *gin.Context) (*time.Time, *time.Time, error) {
	var q struct {
		StartDate string `form:"start_date"`
		EndDate   string `form:"end_date"`
	}
	_ = c.ShouldBindQuery(&q)

	var start, end *time.Time
	if q.StartDate != "" {
		t, err := time.Parse("2006-01-02", q.StartDate)
		if err != nil {
			return nil, nil, domain.NewValidationError(map[string]string{"start_date": "日期格式错误，应为 YYYY-MM-DD"})
		}
		start = &t
	}
	if q.EndDate != "" {
		t, err := time.Parse("2006-01-02", q.EndDate)
		if err != nil {
			return nil, nil, domain.NewValidationError(map[string]string{"end_date": "日期格式错误，应为 YYYY-MM-DD"})
		}
		end = &t
	}
	return start, end, nil
}

// AdminUsage handles GET /api/v1/admin/stats/usage with group_by dispatch.
func (h *StatsHandler) AdminUsage(c *gin.Context) {
	groupBy := c.Query("group_by")
	switch groupBy {
	case "model":
		h.UsageByModel(c)
	case "user":
		h.UsageByUser(c)
	case "day":
		h.UsageByDay(c)
	default:
		respond.ErrorWithStatus(c, http.StatusBadRequest, "INVALID_GROUP_BY", "group_by 必须是 model、user 或 day")
	}
}

// parseStatsWindow reads and clamps the days/limit parameters shared by the
// report endpoints.
func parseStatsWindow(c *gin.Context) (days, limit int, err error) {
	var q struct {
		Days  int `form:"days"`
		Limit int `form:"limit"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		return 0, 0, domain.NewValidationError(map[string]string{"query": "查询参数无效"})
	}
	if q.Days <= 0 {
		q.Days = 14
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	return q.Days, q.Limit, nil
}

// Rankings handles GET /api/v1/admin/stats/rankings.
func (h *StatsHandler) Rankings(c *gin.Context) {
	days, limit, err := parseStatsWindow(c)
	if err != nil {
		respond.Error(c, err)
		return
	}
	metric := c.DefaultQuery("metric", "tokens")

	resp, err := h.svc.Rankings(c.Request.Context(), days, metric, limit)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"list": resp})
}

// ModelDist handles GET /api/v1/admin/stats/model-dist.
func (h *StatsHandler) ModelDist(c *gin.Context) {
	days, limit, err := parseStatsWindow(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.svc.ModelDist(c.Request.Context(), days, limit)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"list": resp})
}

// DeptDist handles GET /api/v1/admin/stats/dept-dist.
func (h *StatsHandler) DeptDist(c *gin.Context) {
	days, limit, err := parseStatsWindow(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.svc.DeptDist(c.Request.Context(), days, limit)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"list": resp})
}

// AdminTrend handles GET /api/v1/admin/stats/trend. days=1 switches to the
// hourly (today, UTC) view; any other value returns the daily view.
func (h *StatsHandler) AdminTrend(c *gin.Context) {
	days, _, err := parseStatsWindow(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	if days == 1 {
		resp, err := h.svc.HourlyTrend(c.Request.Context(), nil)
		if err != nil {
			respond.Error(c, err)
			return
		}
		respond.OK(c, gin.H{"list": resp, "granularity": "hour"})
		return
	}
	resp, err := h.svc.DailyTrend(c.Request.Context(), nil, days)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"list": resp, "granularity": "day"})
}

// FinanceSummary handles GET /api/v1/admin/finance/summary.
func (h *StatsHandler) FinanceSummary(c *gin.Context) {
	days, _, err := parseStatsWindow(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.svc.FinanceSummary(c.Request.Context(), days)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// UserTrend handles GET /api/v1/stats/trend (API Key auth). Same granularity
// switch as the admin view but scoped to the authenticated user.
func (h *StatsHandler) UserTrend(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	days, _, err := parseStatsWindow(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	if days == 1 {
		resp, err := h.svc.HourlyTrend(c.Request.Context(), &userID)
		if err != nil {
			respond.Error(c, err)
			return
		}
		respond.OK(c, gin.H{"list": resp, "granularity": "hour"})
		return
	}
	resp, err := h.svc.DailyTrend(c.Request.Context(), &userID, days)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"list": resp, "granularity": "day"})
}

// UserModelStats handles GET /api/v1/stats/model-stats (API Key auth).
func (h *StatsHandler) UserModelStats(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	days, limit, err := parseStatsWindow(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.svc.UserModelStats(c.Request.Context(), userID, days, limit)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"list": resp})
}
