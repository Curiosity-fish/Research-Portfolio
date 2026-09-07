package service

import (
	"context"
	stdsql "database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/ent/alertrule"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/testutil"
)

// openAlertServiceTestClient returns an Ent client backed by in-memory SQLite
// so EvaluateRules can be exercised against real repositories.
func openAlertServiceTestClient(t *testing.T) *ent.Client {
	t.Helper()

	db, err := stdsql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	t.Cleanup(func() { _ = client.Close() })

	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return client
}

type mockAlertRuleRepository struct {
	createFunc func(ctx context.Context, input repository.CreateAlertRuleInput) (*ent.AlertRule, error)
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*ent.AlertRule, error)
	listFunc    func(ctx context.Context, enabledOnly bool, offset, limit int) ([]*ent.AlertRule, int, error)
	updateFunc  func(ctx context.Context, id uuid.UUID, input repository.UpdateAlertRuleInput) (*ent.AlertRule, error)
	deleteFunc  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockAlertRuleRepository) Create(ctx context.Context, input repository.CreateAlertRuleInput) (*ent.AlertRule, error) {
	return m.createFunc(ctx, input)
}

func (m *mockAlertRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.AlertRule, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockAlertRuleRepository) List(ctx context.Context, enabledOnly bool, offset, limit int) ([]*ent.AlertRule, int, error) {
	return m.listFunc(ctx, enabledOnly, offset, limit)
}

func (m *mockAlertRuleRepository) Update(ctx context.Context, id uuid.UUID, input repository.UpdateAlertRuleInput) (*ent.AlertRule, error) {
	return m.updateFunc(ctx, id, input)
}

func (m *mockAlertRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

type mockAlertRecordRepository struct {
	createFunc           func(ctx context.Context, input repository.CreateAlertRecordInput) (*ent.AlertRecord, error)
	createBatchFunc      func(ctx context.Context, inputs []repository.CreateAlertRecordInput) (int, error)
	existsUnresolvedFunc func(ctx context.Context, ruleID uuid.UUID, userID, tokenID, accountID *uuid.UUID) (bool, error)
	getByIDFunc          func(ctx context.Context, id uuid.UUID) (*ent.AlertRecord, error)
	listFunc             func(ctx context.Context, filter repository.ListAlertRecordFilter) ([]*ent.AlertRecord, int, error)
	resolveFunc          func(ctx context.Context, id uuid.UUID, input repository.ResolveAlertRecordInput) (*ent.AlertRecord, error)
}

func (m *mockAlertRecordRepository) Create(ctx context.Context, input repository.CreateAlertRecordInput) (*ent.AlertRecord, error) {
	return m.createFunc(ctx, input)
}

func (m *mockAlertRecordRepository) CreateBatch(ctx context.Context, inputs []repository.CreateAlertRecordInput) (int, error) {
	return m.createBatchFunc(ctx, inputs)
}

func (m *mockAlertRecordRepository) ExistsUnresolved(ctx context.Context, ruleID uuid.UUID, userID, tokenID, accountID *uuid.UUID) (bool, error) {
	return m.existsUnresolvedFunc(ctx, ruleID, userID, tokenID, accountID)
}

func (m *mockAlertRecordRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.AlertRecord, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockAlertRecordRepository) List(ctx context.Context, filter repository.ListAlertRecordFilter) ([]*ent.AlertRecord, int, error) {
	return m.listFunc(ctx, filter)
}

func (m *mockAlertRecordRepository) Resolve(ctx context.Context, id uuid.UUID, input repository.ResolveAlertRecordInput) (*ent.AlertRecord, error) {
	return m.resolveFunc(ctx, id, input)
}

func TestAlertService_CreateAlertRuleInvalidMetric(t *testing.T) {
	svc := NewAlertService(&mockAlertRuleRepository{}, &mockAlertRecordRepository{}, nil)
	_, err := svc.CreateAlertRule(context.Background(), CreateAlertRuleInput{
		Name:      "test",
		Metric:    "unknown",
		Threshold: 1,
	})
	if err == nil {
		t.Fatal("expected error for invalid metric")
	}
}

func TestAlertService_DeleteAlertRuleNotFound(t *testing.T) {
	rid := uuid.New()
	repo := &mockAlertRuleRepository{
		deleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return &ent.NotFoundError{}
		},
	}
	svc := NewAlertService(repo, &mockAlertRecordRepository{}, nil)
	err := svc.DeleteAlertRule(context.Background(), rid)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAlertService_CreateAlertRuleSuccess(t *testing.T) {
	repo := &mockAlertRuleRepository{
		createFunc: func(ctx context.Context, input repository.CreateAlertRuleInput) (*ent.AlertRule, error) {
			return &ent.AlertRule{ID: uuid.New(), Name: input.Name, Metric: alertrule.Metric(input.Metric), Threshold: input.Threshold, Enabled: input.Enabled}, nil
		},
	}
	svc := NewAlertService(repo, &mockAlertRecordRepository{}, nil)

	resp, err := svc.CreateAlertRule(context.Background(), CreateAlertRuleInput{
		Name:      "余额低",
		Metric:    "balance_low",
		Threshold: 100,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.Metric != "balance_low" {
		t.Fatalf("unexpected metric %s", resp.Metric)
	}
}

func TestAlertService_EvaluateRulesSkipsUnresolvedDuplicates(t *testing.T) {
	ctx := context.Background()
	client := openAlertServiceTestClient(t)

	ruleRepo := repository.NewEntAlertRuleRepository(client)
	recordRepo := repository.NewEntAlertRecordRepository(client)
	svc := NewAlertService(ruleRepo, recordRepo, client)

	// Two active users with low balances both trigger the rule once.
	u1, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	if _, err := client.User.UpdateOneID(u1.ID).SetBalance(10).Save(ctx); err != nil {
		t.Fatalf("set balance u1: %v", err)
	}
	u2, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	if _, err := client.User.UpdateOneID(u2.ID).SetBalance(50).Save(ctx); err != nil {
		t.Fatalf("set balance u2: %v", err)
	}

	if _, err := ruleRepo.Create(ctx, repository.CreateAlertRuleInput{
		Name:      "余额低",
		Metric:    string(alertrule.MetricBalanceLow),
		Threshold: 100,
		Enabled:   true,
	}); err != nil {
		t.Fatalf("create rule: %v", err)
	}

	created, err := svc.EvaluateRules(ctx)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if created != 2 {
		t.Fatalf("expected 2 records created, got %d", created)
	}

	// Second run must not stack duplicate unresolved alerts.
	created, err = svc.EvaluateRules(ctx)
	if err != nil {
		t.Fatalf("evaluate again: %v", err)
	}
	if created != 0 {
		t.Fatalf("expected 0 records on second run, got %d", created)
	}

	total, err := client.AlertRecord.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count records: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 total records, got %d", total)
	}

	// Resolving a record allows the rule to fire again for that target.
	records, err := client.AlertRecord.Query().All(ctx)
	if err != nil {
		t.Fatalf("list records: %v", err)
	}
	admin, err := client.AdminUser.Create().
		SetUsername("admin-" + uuid.NewString()[:8]).
		SetPasswordHash("hash").
		SetRole(adminuser.RoleAdmin).
		SetStatus(adminuser.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	for _, rec := range records {
		if _, err := recordRepo.Resolve(ctx, rec.ID, repository.ResolveAlertRecordInput{ResolvedBy: admin.ID, ResolvedAt: time.Now()}); err != nil {
			t.Fatalf("resolve: %v", err)
		}
	}
	created, err = svc.EvaluateRules(ctx)
	if err != nil {
		t.Fatalf("evaluate after resolve: %v", err)
	}
	if created != 2 {
		t.Fatalf("expected 2 records after resolve, got %d", created)
	}
}
