package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/alertrule"
)

func TestEntAlertRuleRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	repo := NewEntAlertRuleRepository(client)

	rule, err := repo.Create(ctx, CreateAlertRuleInput{
		Name:      "余额告警",
		Metric:    string(alertrule.MetricBalanceLow),
		Threshold: 100,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if rule.Metric != alertrule.MetricBalanceLow {
		t.Fatalf("unexpected metric %s", rule.Metric)
	}

	got, err := repo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != rule.ID {
		t.Fatal("id mismatch")
	}

	desc := "updated"
	updated, err := repo.Update(ctx, rule.ID, UpdateAlertRuleInput{
		Threshold:   ptrInt64(200),
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Threshold != 200 {
		t.Fatalf("expected threshold 200, got %d", updated.Threshold)
	}

	list, total, err := repo.List(ctx, true, 0, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 rule, got total=%d len=%d", total, len(list))
	}

	if err := repo.Delete(ctx, rule.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = repo.GetByID(ctx, rule.ID)
	if err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestEntAlertRuleRepository_ListPagination(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	repo := NewEntAlertRuleRepository(client)
	for i := 0; i < 3; i++ {
		if _, err := repo.Create(ctx, CreateAlertRuleInput{
			Name:      "规则",
			Metric:    string(alertrule.MetricBalanceLow),
			Threshold: 100,
			Enabled:   true,
		}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	list, total, err := repo.List(ctx, true, 2, 2)
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if total != 3 || len(list) != 1 {
		t.Fatalf("expected total=3 len=1, got total=%d len=%d", total, len(list))
	}
}

func TestEntAlertRuleRepository_GetByIDNotFound(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	repo := NewEntAlertRuleRepository(client)
	_, err := repo.GetByID(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error")
	}
}

func ptrInt64(v int64) *int64 { return &v }
