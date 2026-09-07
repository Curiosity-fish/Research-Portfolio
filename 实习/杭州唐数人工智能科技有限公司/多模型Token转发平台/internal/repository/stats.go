package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/calllog"
	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/ent/user"
)

// StatsFilter controls the date range and optional user filter for statistics.
type StatsFilter struct {
	Start  *time.Time
	End    *time.Time
	UserID *uuid.UUID
}

// OverallStats holds the aggregated dashboard numbers.
type OverallStats struct {
	Calls            int   `sql:"calls"`
	PromptTokens     int64 `sql:"prompt_tokens"`
	CompletionTokens int64 `sql:"completion_tokens"`
	TotalTokens      int64 `sql:"total_tokens"`
	RechargeAmount   int64
	ConsumeAmount    int64
}

// ModelUsageStat holds usage aggregated by model.
type ModelUsageStat struct {
	Model            string `sql:"model"`
	Calls            int    `sql:"calls"`
	PromptTokens     int64  `sql:"prompt_tokens"`
	CompletionTokens int64  `sql:"completion_tokens"`
	TotalTokens      int64  `sql:"total_tokens"`
}

// UserUsageStat holds usage aggregated by user.
type UserUsageStat struct {
	UserID      uuid.UUID `sql:"user_id"`
	Calls       int       `sql:"calls"`
	TotalTokens int64     `sql:"total_tokens"`
}

// DayUsageStat holds usage aggregated by day.
type DayUsageStat struct {
	Day         string `json:"day"`
	Calls       int64  `json:"calls"`
	TotalTokens int64  `json:"total_tokens"`
}

// UserRankingRow is one row of the user leaderboard. DepartmentName is empty
// when the user has no department.
type UserRankingRow struct {
	UserID         uuid.UUID `sql:"user_id"`
	Username       string    `sql:"username"`
	Name           string    `sql:"name"`
	DepartmentID   uuid.UUID `sql:"department_id"`
	DepartmentName string    `sql:"department_name"`
	Calls          int       `sql:"calls"`
	TotalTokens    int64     `sql:"total_tokens"`
}

// ModelDistRow is one slice of the per-model distribution pie.
type ModelDistRow struct {
	Model      string `sql:"model"`
	Calls      int    `sql:"calls"`
	TotalTokens int64 `sql:"total_tokens"`
}

// DeptDistRow is one bar of the per-department distribution. DepartmentID is
// nil for users without a department (grouped under empty name).
type DeptDistRow struct {
	DepartmentID   *uuid.UUID `sql:"department_id"`
	DepartmentName string     `sql:"department_name"`
	Calls          int        `sql:"calls"`
	TotalTokens    int64      `sql:"total_tokens"`
}

// HourBucket is one hour slot of the hourly trend.
type HourBucket struct {
	Hour            int   `sql:"hour"`
	Calls           int   `sql:"calls"`
	PromptTokens    int64 `sql:"prompt_tokens"`
	CompletionTokens int64 `sql:"completion_tokens"`
	TotalTokens     int64 `sql:"total_tokens"`
}

// DayBucket is one day slot of the daily trend.
type DayBucket struct {
	Day              string `sql:"day"`
	Calls            int    `sql:"calls"`
	PromptTokens     int64  `sql:"prompt_tokens"`
	CompletionTokens int64  `sql:"completion_tokens"`
	TotalTokens      int64  `sql:"total_tokens"`
}

// FinanceSummary is the platform-wide money overview for a window.
type FinanceSummary struct {
	RechargeAmount  int64 `sql:"recharge_amount"`
	ConsumeAmount   int64 `sql:"consume_amount"`
	RefundAmount    int64 `sql:"refund_amount"`
	AdjustAmount    int64 `sql:"adjust_amount"`
	OrderCount      int   `sql:"order_count"`
	ActiveUserCount int   `sql:"active_user_count"`
	TotalBalance    int64 `sql:"total_balance"`
}

// StatsRepository provides aggregated statistics queries.
type StatsRepository interface {
	OverallStats(ctx context.Context, filter StatsFilter) (*OverallStats, error)
	UsageByModel(ctx context.Context, filter StatsFilter) ([]ModelUsageStat, error)
	UsageByUser(ctx context.Context, filter StatsFilter) ([]UserUsageStat, error)
	UsageByDay(ctx context.Context, filter StatsFilter) ([]DayUsageStat, error)
	// Rankings returns the top users by calls or total tokens within the window.
	Rankings(ctx context.Context, filter StatsFilter, metric string, limit int) ([]UserRankingRow, error)
	// ModelDist returns the top models by tokens within the window.
	ModelDist(ctx context.Context, filter StatsFilter, limit int) ([]ModelDistRow, error)
	// UserModelDist returns the top models by tokens for a single user within
	// the window.
	UserModelDist(ctx context.Context, filter StatsFilter, limit int) ([]ModelDistRow, error)
	// DeptDist returns usage aggregated by user department within the window.
	DeptDist(ctx context.Context, filter StatsFilter, limit int) ([]DeptDistRow, error)
	// HourlyTrend buckets usage by hour. Window must be a single calendar day;
	// the caller aligns it to a day boundary.
	HourlyTrend(ctx context.Context, filter StatsFilter) ([]HourBucket, error)
	// DailyTrend buckets usage by UTC day within the window.
	DailyTrend(ctx context.Context, filter StatsFilter) ([]DayBucket, error)
	// FinanceSummary aggregates balance records and orders in the window plus
	// the platform-wide balance and active-user totals.
	FinanceSummary(ctx context.Context, filter StatsFilter) (*FinanceSummary, error)
}

// EntStatsRepository implements StatsRepository using Ent plus raw SQL for
// the report-style aggregations (joins and time bucketing that ent's group-by
// cannot express portably).
type EntStatsRepository struct {
	client *ent.Client
	// db and dialectName back the raw aggregation queries. They are optional:
	// the legacy pure-ent constructor leaves them nil and only the raw queries
	// require them.
	db          *sql.DB
	dialectName string
}

// NewEntStatsRepository creates a new Ent-backed stats repository.
//
// Deprecated: prefer NewEntStatsRepositoryWithDB so the report aggregations
// backed by raw SQL are available.
func NewEntStatsRepository(client *ent.Client) *EntStatsRepository {
	return &EntStatsRepository{client: client}
}

// NewEntStatsRepositoryWithDB creates an Ent-backed stats repository that can
// also run raw SQL through db. dialectName is dialect.Postgres or
// dialect.SQLite and selects the time-bucketing syntax.
func NewEntStatsRepositoryWithDB(client *ent.Client, db *sql.DB, dialectName string) *EntStatsRepository {
	return &EntStatsRepository{client: client, db: db, dialectName: dialectName}
}

// OverallStats returns the aggregated call and balance totals.
func (r *EntStatsRepository) OverallStats(ctx context.Context, filter StatsFilter) (*OverallStats, error) {
	cq := r.applyCallLogFilter(filter)

	var rows []struct {
		Calls            int   `sql:"calls"`
		PromptTokens     int64 `sql:"prompt_tokens"`
		CompletionTokens int64 `sql:"completion_tokens"`
		TotalTokens      int64 `sql:"total_tokens"`
	}
	if err := cq.Aggregate(
		ent.As(ent.Count(), "calls"),
		ent.As(ent.Sum(calllog.FieldPromptTokens), "prompt_tokens"),
		ent.As(ent.Sum(calllog.FieldCompletionTokens), "completion_tokens"),
		ent.As(ent.Sum(calllog.FieldTotalTokens), "total_tokens"),
	).Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("aggregate call logs: %w", err)
	}

	stats := &OverallStats{}
	if len(rows) > 0 {
		stats.Calls = rows[0].Calls
		stats.PromptTokens = rows[0].PromptTokens
		stats.CompletionTokens = rows[0].CompletionTokens
		stats.TotalTokens = rows[0].TotalTokens
	}

	bq := r.applyBalanceRecordFilter(filter)
	var balanceRows []struct {
		Type   string `sql:"type"`
		Amount int64  `sql:"amount"`
	}
	if err := bq.GroupBy(balancerecord.FieldType).Aggregate(
		ent.As(ent.Sum(balancerecord.FieldAmount), "amount"),
	).Scan(ctx, &balanceRows); err != nil {
		return nil, fmt.Errorf("aggregate balance records: %w", err)
	}
	for _, row := range balanceRows {
		switch row.Type {
		case string(balancerecord.TypeRecharge):
			stats.RechargeAmount = row.Amount
		case string(balancerecord.TypeConsume):
			stats.ConsumeAmount = row.Amount
		}
	}

	return stats, nil
}

// UsageByModel returns usage aggregated by model.
func (r *EntStatsRepository) UsageByModel(ctx context.Context, filter StatsFilter) ([]ModelUsageStat, error) {
	q := r.applyCallLogFilter(filter)

	var rows []ModelUsageStat
	if err := q.GroupBy(calllog.FieldModel).Aggregate(
		ent.As(ent.Count(), "calls"),
		ent.As(ent.Sum(calllog.FieldPromptTokens), "prompt_tokens"),
		ent.As(ent.Sum(calllog.FieldCompletionTokens), "completion_tokens"),
		ent.As(ent.Sum(calllog.FieldTotalTokens), "total_tokens"),
	).Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("usage by model: %w", err)
	}
	return rows, nil
}

// UsageByUser returns usage aggregated by user.
func (r *EntStatsRepository) UsageByUser(ctx context.Context, filter StatsFilter) ([]UserUsageStat, error) {
	q := r.applyCallLogFilter(filter)

	var rows []UserUsageStat
	if err := q.GroupBy(calllog.FieldUserID).Aggregate(
		ent.As(ent.Count(), "calls"),
		ent.As(ent.Sum(calllog.FieldTotalTokens), "total_tokens"),
	).Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("usage by user: %w", err)
	}
	return rows, nil
}

// UsageByDay returns usage aggregated by UTC day.
func (r *EntStatsRepository) UsageByDay(ctx context.Context, filter StatsFilter) ([]DayUsageStat, error) {
	q := r.applyCallLogFilter(filter)

	logs, err := q.Order(ent.Asc(calllog.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("usage by day: %w", err)
	}

	byDay := make(map[string]*DayUsageStat)
	for _, log := range logs {
		day := log.CreatedAt.UTC().Format("2006-01-02")
		s, ok := byDay[day]
		if !ok {
			s = &DayUsageStat{Day: day}
			byDay[day] = s
		}
		s.Calls++
		s.TotalTokens += log.TotalTokens
	}

	result := make([]DayUsageStat, 0, len(byDay))
	for _, s := range byDay {
		result = append(result, *s)
	}
	return result, nil
}

// Rankings returns the top users by calls or total tokens within the window,
// joined with username/name/department for display.
func (r *EntStatsRepository) Rankings(ctx context.Context, filter StatsFilter, metric string, limit int) ([]UserRankingRow, error) {
	orderCol := "total_tokens"
	if metric == "calls" {
		orderCol = "calls"
	}
	// Both candidate order columns are fixed literals above; the window
	// predicates are the only user-controlled parts and they bind as
	// parameters.
	query := fmt.Sprintf(`SELECT
		cl.user_id,
		u.username,
		u.name,
		u.department_id,
		COALESCE(d.name, '') AS department_name,
		COUNT(*) AS calls,
		COALESCE(SUM(cl.total_tokens), 0) AS total_tokens
	FROM call_logs cl
	JOIN users u ON u.id = cl.user_id
	LEFT JOIN departments d ON d.id = u.department_id
	WHERE cl.created_at >= $1 AND cl.created_at < $2
	GROUP BY cl.user_id, u.username, u.name, u.department_id, d.name
	ORDER BY %s DESC, cl.user_id ASC
	LIMIT $3`, orderCol)

	rows, err := r.db.QueryContext(ctx, query, filter.Start, filter.End, limit)
	if err != nil {
		return nil, fmt.Errorf("rankings: %w", err)
	}
	defer rows.Close()

	var out []UserRankingRow
	for rows.Next() {
		var row UserRankingRow
		var deptID sql.NullString
		if err := rows.Scan(&row.UserID, &row.Username, &row.Name, &deptID, &row.DepartmentName, &row.Calls, &row.TotalTokens); err != nil {
			return nil, fmt.Errorf("rankings scan: %w", err)
		}
		if deptID.Valid {
			id, err := uuid.Parse(deptID.String)
			if err != nil {
				return nil, fmt.Errorf("rankings department id: %w", err)
			}
			row.DepartmentID = id
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// ModelDist returns the top models by tokens within the window.
func (r *EntStatsRepository) ModelDist(ctx context.Context, filter StatsFilter, limit int) ([]ModelDistRow, error) {
	q := r.client.CallLog.Query().
		Where(
			calllog.CreatedAtGTE(*filter.Start),
			calllog.CreatedAtLT(*filter.End),
		)

	var rows []ModelDistRow
	if err := q.GroupBy(calllog.FieldModel).Aggregate(
		ent.As(ent.Count(), "calls"),
		ent.As(ent.Sum(calllog.FieldTotalTokens), "total_tokens"),
	).Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("model dist: %w", err)
	}
	sortModelDist(rows)
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows, nil
}

// UserModelDist returns the top models by tokens for a single user.
func (r *EntStatsRepository) UserModelDist(ctx context.Context, filter StatsFilter, limit int) ([]ModelDistRow, error) {
	q := r.client.CallLog.Query().
		Where(
			calllog.CreatedAtGTE(*filter.Start),
			calllog.CreatedAtLT(*filter.End),
			calllog.UserIDEQ(*filter.UserID),
		)

	var rows []ModelDistRow
	if err := q.GroupBy(calllog.FieldModel).Aggregate(
		ent.As(ent.Count(), "calls"),
		ent.As(ent.Sum(calllog.FieldTotalTokens), "total_tokens"),
	).Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("user model dist: %w", err)
	}
	sortModelDist(rows)
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows, nil
}

// DeptDist aggregates call logs by the user's department. Users without a
// department are grouped under an empty name.
func (r *EntStatsRepository) DeptDist(ctx context.Context, filter StatsFilter, limit int) ([]DeptDistRow, error) {
	query := `SELECT
		d.id AS department_id,
		COALESCE(d.name, '') AS department_name,
		COUNT(*) AS calls,
		COALESCE(SUM(cl.total_tokens), 0) AS total_tokens
	FROM call_logs cl
	JOIN users u ON u.id = cl.user_id
	LEFT JOIN departments d ON d.id = u.department_id
	WHERE cl.created_at >= $1 AND cl.created_at < $2
	GROUP BY d.id, d.name
	ORDER BY total_tokens DESC
	LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, filter.Start, filter.End, limit)
	if err != nil {
		return nil, fmt.Errorf("dept dist: %w", err)
	}
	defer rows.Close()

	var out []DeptDistRow
	for rows.Next() {
		var row DeptDistRow
		var deptID sql.NullString
		if err := rows.Scan(&deptID, &row.DepartmentName, &row.Calls, &row.TotalTokens); err != nil {
			return nil, fmt.Errorf("dept dist scan: %w", err)
		}
		if deptID.Valid {
			id, err := uuid.Parse(deptID.String)
			if err != nil {
				return nil, fmt.Errorf("dept dist department id: %w", err)
			}
			row.DepartmentID = &id
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// HourlyTrend buckets usage by hour of day. The caller is responsible for
// aligning the window to a single calendar day.
func (r *EntStatsRepository) HourlyTrend(ctx context.Context, filter StatsFilter) ([]HourBucket, error) {
	hourExpr := "CAST(STRFTIME('%H', created_at) AS INTEGER)"
	if r.dialectName == dialect.Postgres {
		hourExpr = "EXTRACT(HOUR FROM created_at)"
	}
	query := fmt.Sprintf(`SELECT
		%s AS hour,
		COUNT(*) AS calls,
		COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
		COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
		COALESCE(SUM(total_tokens), 0) AS total_tokens
	FROM call_logs
	WHERE created_at >= $1 AND created_at < $2
	`, hourExpr)

	var args []any
	if filter.UserID != nil {
		query += " AND user_id = $3"
		args = append(args, filter.Start, filter.End, *filter.UserID)
	} else {
		args = append(args, filter.Start, filter.End)
	}
	query += " GROUP BY hour ORDER BY hour ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("hourly trend: %w", err)
	}
	defer rows.Close()

	var out []HourBucket
	for rows.Next() {
		var b HourBucket
		if err := rows.Scan(&b.Hour, &b.Calls, &b.PromptTokens, &b.CompletionTokens, &b.TotalTokens); err != nil {
			return nil, fmt.Errorf("hourly trend scan: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// DailyTrend buckets usage by UTC day within the window.
func (r *EntStatsRepository) DailyTrend(ctx context.Context, filter StatsFilter) ([]DayBucket, error) {
	dayExpr := "STRFTIME('%Y-%m-%d', created_at)"
	if r.dialectName == dialect.Postgres {
		// to_char returns text in exactly YYYY-MM-DD, avoiding driver-specific
		// time.Time scanning of the date type.
		dayExpr = "TO_CHAR(created_at, 'YYYY-MM-DD')"
	}
	query := fmt.Sprintf(`SELECT
		%s AS day,
		COUNT(*) AS calls,
		COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
		COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
		COALESCE(SUM(total_tokens), 0) AS total_tokens
	FROM call_logs
	WHERE created_at >= $1 AND created_at < $2
	`, dayExpr)

	var args []any
	if filter.UserID != nil {
		query += " AND user_id = $3"
		args = append(args, filter.Start, filter.End, *filter.UserID)
	} else {
		args = append(args, filter.Start, filter.End)
	}
	query += " GROUP BY day ORDER BY day ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("daily trend: %w", err)
	}
	defer rows.Close()

	var out []DayBucket
	for rows.Next() {
		var b DayBucket
		if err := rows.Scan(&b.Day, &b.Calls, &b.PromptTokens, &b.CompletionTokens, &b.TotalTokens); err != nil {
			return nil, fmt.Errorf("daily trend scan: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// FinanceSummary aggregates recharge/consume/refund/adjust flows from balance
// records in the window, plus paid-order count, active users (any balance
// movement ever), and the platform-wide balance total.
func (r *EntStatsRepository) FinanceSummary(ctx context.Context, filter StatsFilter) (*FinanceSummary, error) {
	var flows []struct {
		Type   string `sql:"type"`
		Amount int64  `sql:"amount"`
	}
	bq := r.client.BalanceRecord.Query().
		Where(
			balancerecord.CreatedAtGTE(*filter.Start),
			balancerecord.CreatedAtLT(*filter.End),
		)
	if err := bq.GroupBy(balancerecord.FieldType).Aggregate(
		ent.As(ent.Sum(balancerecord.FieldAmount), "amount"),
	).Scan(ctx, &flows); err != nil {
		return nil, fmt.Errorf("finance summary flows: %w", err)
	}

	summary := &FinanceSummary{}
	for _, f := range flows {
		switch f.Type {
		case string(balancerecord.TypeRecharge):
			summary.RechargeAmount = f.Amount
		case string(balancerecord.TypeConsume):
			summary.ConsumeAmount = f.Amount
		case string(balancerecord.TypeRefund):
			summary.RefundAmount = f.Amount
		case string(balancerecord.TypeAdminAdjust):
			summary.AdjustAmount = f.Amount
		}
	}

	oq := r.client.RechargeOrder.Query().
		Where(
			rechargeorder.StatusEQ(rechargeorder.StatusPaid),
			rechargeorder.CreatedAtGTE(*filter.Start),
			rechargeorder.CreatedAtLT(*filter.End),
		)
	paidCount, err := oq.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("finance summary order count: %w", err)
	}
	summary.OrderCount = paidCount

	// Active users: anyone with any balance record ever (ent's group-by cannot
	// count distinct values across the whole table).
	if r.db == nil {
		return nil, fmt.Errorf("finance summary requires a raw SQL DB handle")
	}
	var activeCount int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT user_id) FROM balance_records").Scan(&activeCount); err != nil {
		return nil, fmt.Errorf("finance summary active users: %w", err)
	}
	summary.ActiveUserCount = activeCount

	var balanceRow []struct {
		Total int64 `sql:"total"`
	}
	if err := r.client.User.Query().
		Aggregate(ent.As(ent.Sum(user.FieldBalance), "total")).
		Scan(ctx, &balanceRow); err != nil {
		return nil, fmt.Errorf("finance summary total balance: %w", err)
	}
	if len(balanceRow) > 0 {
		summary.TotalBalance = balanceRow[0].Total
	}

	return summary, nil
}

// sortModelDist orders by tokens descending, then model name ascending for
// deterministic output (SQL group-by order is otherwise unspecified).
func sortModelDist(rows []ModelDistRow) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0; j-- {
			a, b := rows[j-1], rows[j]
			if a.TotalTokens > b.TotalTokens || (a.TotalTokens == b.TotalTokens && a.Model <= b.Model) {
				break
			}
			rows[j-1], rows[j] = rows[j], rows[j-1]
		}
	}
}

func (r *EntStatsRepository) applyCallLogFilter(filter StatsFilter) *ent.CallLogQuery {
	q := r.client.CallLog.Query()
	if filter.Start != nil {
		q = q.Where(calllog.CreatedAtGTE(*filter.Start))
	}
	if filter.End != nil {
		q = q.Where(calllog.CreatedAtLTE(*filter.End))
	}
	if filter.UserID != nil {
		q = q.Where(calllog.UserIDEQ(*filter.UserID))
	}
	return q
}

func (r *EntStatsRepository) applyBalanceRecordFilter(filter StatsFilter) *ent.BalanceRecordQuery {
	q := r.client.BalanceRecord.Query()
	if filter.Start != nil {
		q = q.Where(balancerecord.CreatedAtGTE(*filter.Start))
	}
	if filter.End != nil {
		q = q.Where(balancerecord.CreatedAtLTE(*filter.End))
	}
	if filter.UserID != nil {
		q = q.Where(balancerecord.UserIDEQ(*filter.UserID))
	}
	return q
}

var _ StatsRepository = (*EntStatsRepository)(nil)
