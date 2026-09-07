package repository

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/platform"
	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func createTestCallLog(t *testing.T, ctx context.Context, client *ent.Client, u *ent.User, tok *ent.UserToken, model string, total int64) {
	t.Helper()

	p, err := client.Platform.Create().
		SetName("测试平台" + uuid.NewString()[:4]).
		SetCode("test-" + uuid.NewString()[:4]).
		SetType(platform.TypeOpenai).
		SetBaseURL("https://example.com").
		SetStatus(platform.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}

	a, err := client.Account.Create().
		SetPlatformID(p.ID).
		SetName("测试账号").
		SetAPIKeyEncrypted("encrypted").
		SetWeight(1).
		SetStatus(account.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err = client.CallLog.Create().
		SetUserID(u.ID).
		SetTokenID(tok.ID).
		SetPlatformID(p.ID).
		SetAccountID(a.ID).
		SetModel(model).
		SetPromptTokens(total / 2).
		SetCompletionTokens(total - total/2).
		SetTotalTokens(total).
		SetLatencyMs(10).
		SetStatusCode(200).
		Save(ctx)
	if err != nil {
		t.Fatalf("create call log: %v", err)
	}
}

func TestEntStatsRepository_OverallStats(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	createTestCallLog(t, ctx, client, u, tok, "gpt-4", 100)
	createTestCallLog(t, ctx, client, u, tok, "gpt-4", 200)

	_, err := client.BalanceRecord.Create().
		SetUserID(u.ID).
		SetType(balancerecord.TypeRecharge).
		SetAmount(500).
		SetBalanceAfter(500).
		Save(ctx)
	if err != nil {
		t.Fatalf("create balance record: %v", err)
	}

	repo := NewEntStatsRepository(client)
	stats, err := repo.OverallStats(ctx, StatsFilter{})
	if err != nil {
		t.Fatalf("overall stats: %v", err)
	}
	if stats.Calls != 2 {
		t.Fatalf("expected calls 2, got %d", stats.Calls)
	}
	if stats.TotalTokens != 300 {
		t.Fatalf("expected total tokens 300, got %d", stats.TotalTokens)
	}
	if stats.RechargeAmount != 500 {
		t.Fatalf("expected recharge 500, got %d", stats.RechargeAmount)
	}
}

func TestEntStatsRepository_UsageByModel(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	createTestCallLog(t, ctx, client, u, tok, "gpt-4", 100)
	createTestCallLog(t, ctx, client, u, tok, "gpt-3.5", 50)

	repo := NewEntStatsRepository(client)
	rows, err := repo.UsageByModel(ctx, StatsFilter{})
	if err != nil {
		t.Fatalf("usage by model: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 models, got %d", len(rows))
	}
}

func TestEntStatsRepository_UsageByDay(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	createTestCallLog(t, ctx, client, u, tok, "gpt-4", 100)

	repo := NewEntStatsRepository(client)
	rows, err := repo.UsageByDay(ctx, StatsFilter{})
	if err != nil {
		t.Fatalf("usage by day: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 day, got %d", len(rows))
	}
	if rows[0].Day != time.Now().UTC().Format("2006-01-02") {
		t.Fatalf("unexpected day %s", rows[0].Day)
	}
}

func TestEntStatsRepository_UserFilter(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u1, _, tok1, _ := testutil.CreateTestAPIKey(t, ctx, client)
	u2, _, tok2, _ := testutil.CreateTestAPIKey(t, ctx, client)
	createTestCallLog(t, ctx, client, u1, tok1, "gpt-4", 100)
	createTestCallLog(t, ctx, client, u2, tok2, "gpt-4", 200)

	repo := NewEntStatsRepository(client)
	stats, err := repo.OverallStats(ctx, StatsFilter{UserID: &u1.ID})
	if err != nil {
		t.Fatalf("overall stats: %v", err)
	}
	if stats.Calls != 1 {
		t.Fatalf("expected calls 1, got %d", stats.Calls)
	}
	if stats.TotalTokens != 100 {
		t.Fatalf("expected total tokens 100, got %d", stats.TotalTokens)
	}
}

// openStatsTestRepo opens a test client plus the raw *sql.DB handle that the
// report-style statistics queries require.
func openStatsTestRepo(t *testing.T) (*ent.Client, *EntStatsRepository) {
	t.Helper()
	client, db := openTestClientDB(t)
	repo := NewEntStatsRepositoryWithDB(client, db, dialect.SQLite)
	return client, repo
}

// wideWindow covers every timestamp the tests produce.
func wideWindow() StatsFilter {
	start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	return StatsFilter{Start: &start, End: &end}
}

func TestEntStatsRepository_Rankings(t *testing.T) {
	ctx := context.Background()
	client, repo := openStatsTestRepo(t)
	defer client.Close()

	u1, _, tok1, _ := testutil.CreateTestAPIKey(t, ctx, client)
	u2, _, tok2, _ := testutil.CreateTestAPIKey(t, ctx, client)
	createTestCallLog(t, ctx, client, u1, tok1, "gpt-4", 100)
	createTestCallLog(t, ctx, client, u1, tok1, "gpt-4", 200)
	createTestCallLog(t, ctx, client, u2, tok2, "gpt-4", 50)

	dept, err := client.Department.Create().SetName("数学系").SetCode("math-"+uuid.NewString()[:4]).Save(ctx)
	if err != nil {
		t.Fatalf("create department: %v", err)
	}
	if _, err := client.User.UpdateOneID(u1.ID).SetDepartmentID(dept.ID).Save(ctx); err != nil {
		t.Fatalf("assign department: %v", err)
	}

	rows, err := repo.Rankings(ctx, wideWindow(), "tokens", 10)
	if err != nil {
		t.Fatalf("rankings: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	// u1 leads on tokens: 300 vs 50.
	if rows[0].UserID != u1.ID || rows[0].TotalTokens != 300 || rows[0].Calls != 2 {
		t.Fatalf("unexpected leader row: %+v", rows[0])
	}
	if rows[0].DepartmentName != "数学系" {
		t.Fatalf("expected department name, got %+v", rows[0])
	}

	// Metric switch reorders by call count.
	rows, err = repo.Rankings(ctx, wideWindow(), "calls", 10)
	if err != nil {
		t.Fatalf("rankings by calls: %v", err)
	}
	if rows[0].UserID != u1.ID || rows[0].Calls != 2 {
		t.Fatalf("unexpected calls ranking: %+v", rows[0])
	}
}

func TestEntStatsRepository_Dists(t *testing.T) {
	ctx := context.Background()
	client, repo := openStatsTestRepo(t)
	defer client.Close()

	u1, _, tok1, _ := testutil.CreateTestAPIKey(t, ctx, client)
	u2, _, tok2, _ := testutil.CreateTestAPIKey(t, ctx, client)
	createTestCallLog(t, ctx, client, u1, tok1, "gpt-4", 100)
	createTestCallLog(t, ctx, client, u1, tok1, "gpt-4o", 300)
	createTestCallLog(t, ctx, client, u2, tok2, "gpt-4", 50)

	models, err := repo.ModelDist(ctx, wideWindow(), 10)
	if err != nil {
		t.Fatalf("model dist: %v", err)
	}
	if len(models) != 2 || models[0].Model != "gpt-4o" || models[0].TotalTokens != 300 {
		t.Fatalf("unexpected model dist: %+v", models)
	}

	userModels, err := repo.UserModelDist(ctx, StatsFilter{Start: wideWindow().Start, End: wideWindow().End, UserID: &u1.ID}, 10)
	if err != nil {
		t.Fatalf("user model dist: %v", err)
	}
	if len(userModels) != 2 {
		t.Fatalf("expected 2 models for user, got %+v", userModels)
	}

	depts, err := repo.DeptDist(ctx, wideWindow(), 10)
	if err != nil {
		t.Fatalf("dept dist: %v", err)
	}
	// Both users lack a department: they collapse into one empty-name bucket.
	if len(depts) != 1 || depts[0].DepartmentName != "" || depts[0].Calls != 3 {
		t.Fatalf("unexpected dept dist: %+v", depts)
	}
}

func TestEntStatsRepository_Trends(t *testing.T) {
	ctx := context.Background()
	client, db := openTestClientDB(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	logRepo := NewEntCallLogRepository(client)
	mk := func(at time.Time, total int64) {
		t.Helper()
		p, err := client.Platform.Create().
			SetName("p" + uuid.NewString()[:4]).
			SetCode("p-" + uuid.NewString()[:4]).
			SetType(platform.TypeOpenai).
			SetBaseURL("https://example.com").
			SetStatus(platform.StatusActive).
			Save(ctx)
		if err != nil {
			t.Fatalf("create platform: %v", err)
		}
		a, err := client.Account.Create().
			SetPlatformID(p.ID).
			SetName("a").
			SetAPIKeyEncrypted("enc").
			SetWeight(1).
			SetStatus(account.StatusActive).
			Save(ctx)
		if err != nil {
			t.Fatalf("create account: %v", err)
		}
		log, err := logRepo.Create(ctx, CreateCallLogInput{
			UserID: u.ID, TokenID: tok.ID, PlatformID: p.ID, AccountID: a.ID,
			Model: "gpt-4", StatusCode: 200,
		})
		if err != nil {
			t.Fatalf("create call log: %v", err)
		}
		// created_at is immutable through Ent; backfill via raw SQL.
		if _, err := db.ExecContext(ctx,
			"UPDATE call_logs SET created_at = $1, total_tokens = $2, prompt_tokens = $3, completion_tokens = $4 WHERE id = $5",
			at.UTC().Format(time.RFC3339Nano), total, total/2, total-total/2, log.ID,
		); err != nil {
			t.Fatalf("backfill call log: %v", err)
		}
	}

	day := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	mk(day.Add(3*time.Hour), 100) // 03:00
	mk(day.Add(3*time.Hour+time.Minute), 50) // 03:01, same bucket
	mk(day.Add(15*time.Hour), 70) // 15:00
	mk(day.AddDate(0, 0, -1), 40) // previous day

	repo := NewEntStatsRepositoryWithDB(client, db, dialect.SQLite)
	hours, err := repo.HourlyTrend(ctx, StatsFilter{Start: &day, End: func() *time.Time { e := day.AddDate(0, 0, 1); return &e }(), UserID: &u.ID})
	if err != nil {
		t.Fatalf("hourly trend: %v", err)
	}
	if len(hours) != 2 {
		t.Fatalf("expected 2 hour buckets, got %+v", hours)
	}
	if hours[0].Hour != 3 || hours[0].Calls != 2 || hours[0].TotalTokens != 150 {
		t.Fatalf("unexpected 03:00 bucket: %+v", hours[0])
	}
	if hours[1].Hour != 15 || hours[1].TotalTokens != 70 {
		t.Fatalf("unexpected 15:00 bucket: %+v", hours[1])
	}

	start := day.AddDate(0, 0, -1)
	end := day.AddDate(0, 0, 1)
	days, err := repo.DailyTrend(ctx, StatsFilter{Start: &start, End: &end, UserID: &u.ID})
	if err != nil {
		t.Fatalf("daily trend: %v", err)
	}
	if len(days) != 2 {
		t.Fatalf("expected 2 day buckets, got %+v", days)
	}
	if days[0].Day != "2026-09-05" || days[0].TotalTokens != 40 {
		t.Fatalf("unexpected first day bucket: %+v", days[0])
	}
	if days[1].Day != "2026-09-06" || days[1].TotalTokens != 220 {
		t.Fatalf("unexpected second day bucket: %+v", days[1])
	}
}

func TestEntStatsRepository_FinanceSummary(t *testing.T) {
	ctx := context.Background()
	client, repo := openStatsTestRepo(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	if _, err := client.User.UpdateOneID(u.ID).SetBalance(700).Save(ctx); err != nil {
		t.Fatalf("set balance: %v", err)
	}
	mk := func(typ balancerecord.Type, amount int64) {
		t.Helper()
		if _, err := client.BalanceRecord.Create().
			SetUserID(u.ID).
			SetType(typ).
			SetAmount(amount).
			SetBalanceAfter(amount).
			Save(ctx); err != nil {
			t.Fatalf("create balance record: %v", err)
		}
	}
	mk(balancerecord.TypeRecharge, 1000)
	mk(balancerecord.TypeConsume, -300)
	mk(balancerecord.TypeRefund, -200)
	mk(balancerecord.TypeAdminAdjust, 50)

	if _, err := client.RechargeOrder.Create().
		SetUserID(u.ID).
		SetAmount(1000).
		SetStatus(rechargeorder.StatusPaid).
		SetProvider(rechargeorder.ProviderMock).
		Save(ctx); err != nil {
		t.Fatalf("create paid order: %v", err)
	}

	summary, err := repo.FinanceSummary(ctx, wideWindow())
	if err != nil {
		t.Fatalf("finance summary: %v", err)
	}
	if summary.RechargeAmount != 1000 || summary.ConsumeAmount != -300 || summary.RefundAmount != -200 || summary.AdjustAmount != 50 {
		t.Fatalf("unexpected flows: %+v", summary)
	}
	if summary.OrderCount != 1 {
		t.Fatalf("expected 1 paid order, got %d", summary.OrderCount)
	}
	if summary.ActiveUserCount != 1 {
		t.Fatalf("expected 1 active user, got %d", summary.ActiveUserCount)
	}
	if summary.TotalBalance != 700 {
		t.Fatalf("expected total balance 700, got %d", summary.TotalBalance)
	}
}
