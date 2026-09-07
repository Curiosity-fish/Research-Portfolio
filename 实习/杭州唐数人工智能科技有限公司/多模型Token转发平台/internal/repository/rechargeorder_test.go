package repository

import (
	"context"
	"testing"

	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntRechargeOrderRepository_ListPagination(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntRechargeOrderRepository(client)

	for i := 0; i < 3; i++ {
		if _, err := repo.Create(ctx, CreateRechargeOrderInput{
			UserID:   u.ID,
			Amount:   1000,
			Provider: string(rechargeorder.ProviderMock),
		}); err != nil {
			t.Fatalf("create order %d: %v", i, err)
		}
	}

	orders, total, err := repo.List(ctx, nil, 2, 2)
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if total != 3 || len(orders) != 1 {
		t.Fatalf("expected total=3 len=1, got total=%d len=%d", total, len(orders))
	}

	orders, total, err = repo.ListByUserID(ctx, u.ID, 2, 2)
	if err != nil {
		t.Fatalf("list by user page 2: %v", err)
	}
	if total != 3 || len(orders) != 1 {
		t.Fatalf("expected total=3 len=1, got total=%d len=%d", total, len(orders))
	}
}

func TestEntRechargeOrderRepository_MarkPaidIdempotent(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntRechargeOrderRepository(client)

	order, err := repo.Create(ctx, CreateRechargeOrderInput{
		UserID:   u.ID,
		Amount:   1000,
		Provider: string(rechargeorder.ProviderMock),
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	paid, err := repo.MarkPaid(ctx, order.ID, "mock-1")
	if err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if paid.Status != rechargeorder.StatusPaid {
		t.Fatalf("expected status paid, got %s", paid.Status)
	}

	// A duplicate callback must not credit the balance twice.
	again, err := repo.MarkPaid(ctx, order.ID, "mock-2")
	if err != nil {
		t.Fatalf("mark paid again: %v", err)
	}
	if again.Status != rechargeorder.StatusPaid {
		t.Fatalf("expected status paid, got %s", again.Status)
	}

	fresh, err := client.User.Get(ctx, u.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if fresh.Balance != 1000 {
		t.Fatalf("expected balance credited once (1000), got %d", fresh.Balance)
	}

	records, err := client.BalanceRecord.Query().
		Where(
			balancerecord.UserIDEQ(u.ID),
			balancerecord.TypeEQ(balancerecord.TypeRecharge),
		).
		Count(ctx)
	if err != nil {
		t.Fatalf("count balance records: %v", err)
	}
	if records != 1 {
		t.Fatalf("expected 1 recharge balance record, got %d", records)
	}
}
