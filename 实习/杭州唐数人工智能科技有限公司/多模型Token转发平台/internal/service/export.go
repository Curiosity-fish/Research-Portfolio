package service

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// exportMaxRows caps every export sheet so a single request cannot materialise
// an unbounded spreadsheet. 10000 data rows plus the header is well within
// excelize's in-memory limits.
const exportMaxRows = 10000

// exportTimeLayout renders timestamps in exported sheets. Times are stored in
// UTC; the layout keeps the value unambiguous without an explicit zone suffix.
const exportTimeLayout = "2006-01-02 15:04:05"

// ExportSheet is one worksheet of an exported workbook: a name plus rows of
// plain strings (row 0 is always the header).
type ExportSheet struct {
	Name string
	Rows [][]string
}

// ExportService builds XLSX-ready row data for the export endpoints. Handlers
// own the excelize wiring; keeping this layer pure [][]string makes the rows
// trivially unit-testable.
type ExportService struct {
	stats       *StatsService
	balanceRepo repository.BalanceRecordRepository
	orderRepo   repository.RechargeOrderRepository
}

// NewExportService creates a new ExportService.
func NewExportService(stats *StatsService, balanceRepo repository.BalanceRecordRepository, orderRepo repository.RechargeOrderRepository) *ExportService {
	return &ExportService{stats: stats, balanceRepo: balanceRepo, orderRepo: orderRepo}
}

// UsageExport returns the usage-statistics sheet grouped by model, user or
// day. groupBy semantics match StatsHandler.AdminUsage.
func (s *ExportService) UsageExport(ctx context.Context, start, end *time.Time, groupBy string) ([][]string, error) {
	switch groupBy {
	case "model":
		rows, err := s.stats.UsageByModel(ctx, start, end)
		if err != nil {
			return nil, err
		}
		sheet := [][]string{{"模型", "调用次数", "提示 Tokens", "补全 Tokens", "总 Tokens"}}
		for _, r := range rows {
			sheet = append(sheet, []string{
				r.Model,
				strconv.Itoa(r.Calls),
				strconv.FormatInt(r.PromptTokens, 10),
				strconv.FormatInt(r.CompletionTokens, 10),
				strconv.FormatInt(r.TotalTokens, 10),
			})
		}
		return sheet, nil
	case "user":
		rows, err := s.stats.UsageByUser(ctx, start, end)
		if err != nil {
			return nil, err
		}
		sheet := [][]string{{"用户 ID", "调用次数", "总 Tokens"}}
		for _, r := range rows {
			sheet = append(sheet, []string{
				r.UserID,
				strconv.Itoa(r.Calls),
				strconv.FormatInt(r.TotalTokens, 10),
			})
		}
		return sheet, nil
	case "day":
		rows, err := s.stats.UsageByDay(ctx, start, end)
		if err != nil {
			return nil, err
		}
		sheet := [][]string{{"日期", "调用次数", "总 Tokens"}}
		for _, r := range rows {
			sheet = append(sheet, []string{
				r.Day,
				strconv.FormatInt(r.Calls, 10),
				strconv.FormatInt(r.TotalTokens, 10),
			})
		}
		return sheet, nil
	default:
		return nil, domain.NewAppError(400, "INVALID_GROUP_BY", "group_by 必须是 model、user 或 day")
	}
}

// SummaryExport returns the finance summary as a metrics sheet plus the
// zero-filled daily trend for the same window.
func (s *ExportService) SummaryExport(ctx context.Context, days int) ([]ExportSheet, error) {
	summary, err := s.stats.FinanceSummary(ctx, days)
	if err != nil {
		return nil, err
	}
	trend, err := s.stats.DailyTrend(ctx, nil, days)
	if err != nil {
		return nil, err
	}

	metrics := [][]string{{"指标", "数值"}}
	metrics = append(metrics,
		[]string{"统计开始日期", summary.PeriodStart},
		[]string{"统计结束日期", summary.PeriodEnd},
		[]string{"充值金额", strconv.FormatInt(summary.RechargeAmount, 10)},
		[]string{"消费金额", strconv.FormatInt(summary.ConsumeAmount, 10)},
		[]string{"退款金额", strconv.FormatInt(summary.RefundAmount, 10)},
		[]string{"管理员调整金额", strconv.FormatInt(summary.AdjustAmount, 10)},
		[]string{"支付订单数", strconv.Itoa(summary.OrderCount)},
		[]string{"活跃用户总数", strconv.Itoa(summary.ActiveUserCount)},
		[]string{"平台余额总额", strconv.FormatInt(summary.TotalBalance, 10)},
	)

	daily := [][]string{{"日期", "调用次数", "提示 Tokens", "补全 Tokens", "总 Tokens"}}
	for _, p := range trend {
		daily = append(daily, []string{
			p.Date,
			strconv.Itoa(p.Calls),
			strconv.FormatInt(p.PromptTokens, 10),
			strconv.FormatInt(p.CompletionTokens, 10),
			strconv.FormatInt(p.TotalTokens, 10),
		})
	}

	return []ExportSheet{
		{Name: "财务汇总", Rows: metrics},
		{Name: "按天趋势", Rows: daily},
	}, nil
}

// BalanceRecordsExport returns the balance-record sheet. A nil userID exports
// the whole platform; the caller passes the authenticated user for the
// user-facing endpoint.
func (s *ExportService) BalanceRecordsExport(ctx context.Context, userID *uuid.UUID) ([][]string, error) {
	records, err := s.balanceRepo.ListRecentWithUser(ctx, userID, exportMaxRows)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	sheet := [][]string{{"记录 ID", "用户 ID", "用户名", "类型", "金额", "变动后余额", "备注", "时间"}}
	for _, rec := range records {
		username := ""
		if rec.Edges.User != nil {
			username = rec.Edges.User.Username
		}
		remark := ""
		if rec.Remark != nil {
			remark = *rec.Remark
		}
		sheet = append(sheet, []string{
			rec.ID.String(),
			rec.UserID.String(),
			username,
			string(rec.Type),
			strconv.FormatInt(rec.Amount, 10),
			strconv.FormatInt(rec.BalanceAfter, 10),
			remark,
			rec.CreatedAt.Format(exportTimeLayout),
		})
	}
	return sheet, nil
}

// RechargeOrdersExport returns the recharge-order sheet for the whole
// platform.
func (s *ExportService) RechargeOrdersExport(ctx context.Context) ([][]string, error) {
	orders, err := s.orderRepo.ListRecentWithUser(ctx, exportMaxRows)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	sheet := [][]string{{"订单 ID", "用户 ID", "用户名", "金额", "状态", "渠道", "支付时间", "创建时间"}}
	for _, order := range orders {
		username := ""
		if order.Edges.User != nil {
			username = order.Edges.User.Username
		}
		paidAt := ""
		if order.PaidAt != nil {
			paidAt = order.PaidAt.Format(exportTimeLayout)
		}
		sheet = append(sheet, []string{
			order.ID.String(),
			order.UserID.String(),
			username,
			strconv.FormatInt(order.Amount, 10),
			string(order.Status),
			string(order.Provider),
			paidAt,
			order.CreatedAt.Format(exportTimeLayout),
		})
	}
	return sheet, nil
}
