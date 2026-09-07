package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/ent/alertrule"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntAlertRecordRepository_ListPagination(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	ruleRepo := NewEntAlertRuleRepository(client)
	recordRepo := NewEntAlertRecordRepository(client)

	rule, err := ruleRepo.Create(ctx, CreateAlertRuleInput{
		Name:      "测试规则",
		Metric:    string(alertrule.MetricBalanceLow),
		Threshold: 100,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}

	for i := 0; i < 3; i++ {
		u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
		if _, err := recordRepo.Create(ctx, CreateAlertRecordInput{
			RuleID:         &rule.ID,
			Metric:         string(rule.Metric),
			UserID:         &u.ID,
			TriggeredValue: 50,
			Message:        "余额不足",
		}); err != nil {
			t.Fatalf("create record %d: %v", i, err)
		}
	}

	list, total, err := recordRepo.List(ctx, ListAlertRecordFilter{Offset: 2, Limit: 2})
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if total != 3 || len(list) != 1 {
		t.Fatalf("expected total=3 len=1, got total=%d len=%d", total, len(list))
	}
}

func TestEntAlertRecordRepository_ExistsUnresolvedAndBatch(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	ruleRepo := NewEntAlertRuleRepository(client)
	recordRepo := NewEntAlertRecordRepository(client)

	rule, err := ruleRepo.Create(ctx, CreateAlertRuleInput{
		Name:      "测试规则",
		Metric:    string(alertrule.MetricBalanceLow),
		Threshold: 100,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	uid := u.ID

	exists, err := recordRepo.ExistsUnresolved(ctx, rule.ID, &uid, nil, nil)
	if err != nil {
		t.Fatalf("exists unresolved: %v", err)
	}
	if exists {
		t.Fatal("expected no unresolved record before creation")
	}

	n, err := recordRepo.CreateBatch(ctx, []CreateAlertRecordInput{{
		RuleID:         &rule.ID,
		Metric:         string(rule.Metric),
		UserID:         &uid,
		TriggeredValue: 50,
		Message:        "余额不足",
	}})
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 created, got %d", n)
	}

	exists, err = recordRepo.ExistsUnresolved(ctx, rule.ID, &uid, nil, nil)
	if err != nil {
		t.Fatalf("exists unresolved after create: %v", err)
	}
	if !exists {
		t.Fatal("expected unresolved record after creation")
	}

	// A nil-target lookup must not match a record that has a user target.
	exists, err = recordRepo.ExistsUnresolved(ctx, rule.ID, nil, nil, nil)
	if err != nil {
		t.Fatalf("exists unresolved nil target: %v", err)
	}
	if exists {
		t.Fatal("expected nil target not to match user-targeted record")
	}
}

func TestEntAlertRecordRepository_CreateAndResolve(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	ruleRepo := NewEntAlertRuleRepository(client)
	recordRepo := NewEntAlertRecordRepository(client)

	rule, err := ruleRepo.Create(ctx, CreateAlertRuleInput{
		Name:      "测试规则",
		Metric:    string(alertrule.MetricBalanceLow),
		Threshold: 100,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	admin, err := client.AdminUser.Create().
		SetUsername("admin-" + uuid.NewString()[:8]).
		SetPasswordHash("hash").
		SetRole(adminuser.RoleAdmin).
		SetStatus(adminuser.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}

	rec, err := recordRepo.Create(ctx, CreateAlertRecordInput{
		RuleID:         &rule.ID,
		Metric:         string(rule.Metric),
		UserID:         &u.ID,
		TriggeredValue: 50,
		Message:        "余额不足",
	})
	if err != nil {
		t.Fatalf("create record: %v", err)
	}
	if rec.IsResolved {
		t.Fatal("expected unresolved")
	}

	list, total, err := recordRepo.List(ctx, ListAlertRecordFilter{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 record, got total=%d len=%d", total, len(list))
	}

	resolved, err := recordRepo.Resolve(ctx, rec.ID, ResolveAlertRecordInput{
		ResolvedBy: admin.ID,
		ResolvedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !resolved.IsResolved {
		t.Fatal("expected resolved")
	}

	resolvedOnly := true
	list, total, err = recordRepo.List(ctx, ListAlertRecordFilter{IsResolved: &resolvedOnly, Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("list resolved: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 resolved record, got total=%d len=%d", total, len(list))
	}
}
