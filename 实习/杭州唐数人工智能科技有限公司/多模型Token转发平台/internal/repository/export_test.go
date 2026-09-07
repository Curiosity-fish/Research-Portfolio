package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntBalanceRecordRepository_ListRecentWithUser(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u1, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	u2, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntBalanceRecordRepository(client)

	mk := func(userID uuid.UUID, typ balancerecord.Type, amount int64) {
		t.Helper()
		if _, err := repo.Create(ctx, CreateBalanceRecordInput{
			UserID:       userID,
			Type:         typ,
			Amount:       amount,
			BalanceAfter: amount,
		}); err != nil {
			t.Fatalf("create balance record: %v", err)
		}
	}
	mk(u1.ID, balancerecord.TypeRecharge, 100)
	mk(u1.ID, balancerecord.TypeConsume, -30)
	mk(u2.ID, balancerecord.TypeRecharge, 200)

	// Platform-wide export: every record carries its owning user.
	all, err := repo.ListRecentWithUser(ctx, nil, 100)
	if err != nil {
		t.Fatalf("list all for export: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 records, got %d", len(all))
	}
	for _, rec := range all {
		if rec.Edges.User == nil {
			t.Fatalf("expected eager-loaded user on record %s", rec.ID)
		}
	}

	// User filter.
	mine, err := repo.ListRecentWithUser(ctx, &u1.ID, 100)
	if err != nil {
		t.Fatalf("list by user for export: %v", err)
	}
	if len(mine) != 2 {
		t.Fatalf("expected 2 records for u1, got %d", len(mine))
	}
	for _, rec := range mine {
		if rec.UserID != u1.ID {
			t.Fatalf("unexpected record for user %s", rec.UserID)
		}
		if rec.Edges.User.Username != u1.Username {
			t.Fatalf("expected username %q, got %q", u1.Username, rec.Edges.User.Username)
		}
	}

	// Limit caps the newest rows.
	limited, err := repo.ListRecentWithUser(ctx, nil, 2)
	if err != nil {
		t.Fatalf("list limited for export: %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("expected 2 records with limit, got %d", len(limited))
	}
}

func TestEntRechargeOrderRepository_ListRecentWithUser(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u1, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	u2, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntRechargeOrderRepository(client)

	for _, uid := range []uuid.UUID{u1.ID, u2.ID} {
		if _, err := repo.Create(ctx, CreateRechargeOrderInput{
			UserID:   uid,
			Amount:   500,
			Provider: string(rechargeorder.ProviderMock),
		}); err != nil {
			t.Fatalf("create recharge order: %v", err)
		}
	}

	orders, err := repo.ListRecentWithUser(ctx, 100)
	if err != nil {
		t.Fatalf("list orders for export: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
	for _, order := range orders {
		if order.Edges.User == nil {
			t.Fatalf("expected eager-loaded user on order %s", order.ID)
		}
	}

	limited, err := repo.ListRecentWithUser(ctx, 1)
	if err != nil {
		t.Fatalf("list limited orders for export: %v", err)
	}
	if len(limited) != 1 {
		t.Fatalf("expected 1 order with limit, got %d", len(limited))
	}
}
