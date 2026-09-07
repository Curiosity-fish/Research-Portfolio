package repository

import (
	"context"
	"testing"

	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntReconciliationRepository_GetRechargeOrderWithRecords(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	order, err := client.RechargeOrder.Create().
		SetUserID(u.ID).
		SetAmount(1000).
		SetStatus(rechargeorder.StatusPaid).
		SetProvider(rechargeorder.ProviderMock).
		Save(ctx)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	repo := NewEntReconciliationRepository(client)
	got, records, err := repo.GetRechargeOrderWithRecords(ctx, order.ID)
	if err != nil {
		t.Fatalf("get order with records: %v", err)
	}
	if got.ID != order.ID {
		t.Fatal("order id mismatch")
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}
}

func TestEntReconciliationRepository_RefundOrder(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	_, err := client.User.UpdateOneID(u.ID).SetBalance(1000).Save(ctx)
	if err != nil {
		t.Fatalf("set balance: %v", err)
	}

	order, err := client.RechargeOrder.Create().
		SetUserID(u.ID).
		SetAmount(1000).
		SetStatus(rechargeorder.StatusPaid).
		SetProvider(rechargeorder.ProviderMock).
		Save(ctx)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	repo := NewEntReconciliationRepository(client)
	updated, rec, err := repo.RefundOrder(ctx, order.ID, 300)
	if err != nil {
		t.Fatalf("refund: %v", err)
	}
	if updated.RefundedAmount != 300 {
		t.Fatalf("expected refunded 300, got %d", updated.RefundedAmount)
	}
	if rec.Amount != 300 {
		t.Fatalf("expected record amount 300, got %d", rec.Amount)
	}

	freshUser, err := client.User.Get(ctx, u.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if freshUser.Balance != 700 {
		t.Fatalf("expected balance 700, got %d", freshUser.Balance)
	}
}

func TestEntReconciliationRepository_RefundOrderExceedsAmount(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	_, err := client.User.UpdateOneID(u.ID).SetBalance(1000).Save(ctx)
	if err != nil {
		t.Fatalf("set balance: %v", err)
	}

	order, err := client.RechargeOrder.Create().
		SetUserID(u.ID).
		SetAmount(1000).
		SetStatus(rechargeorder.StatusPaid).
		SetProvider(rechargeorder.ProviderMock).
		Save(ctx)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	repo := NewEntReconciliationRepository(client)
	_, _, err = repo.RefundOrder(ctx, order.ID, 1200)
	if err == nil {
		t.Fatal("expected error for refund exceeding amount")
	}
}

func TestEntReconciliationRepository_RefundOrderNotPaid(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	order, err := client.RechargeOrder.Create().
		SetUserID(u.ID).
		SetAmount(1000).
		SetStatus(rechargeorder.StatusPending).
		SetProvider(rechargeorder.ProviderMock).
		Save(ctx)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	repo := NewEntReconciliationRepository(client)
	_, _, err = repo.RefundOrder(ctx, order.ID, 100)
	if err == nil {
		t.Fatal("expected error for unpaid order refund")
	}
}

func TestEntReconciliationRepository_RefundOrderPartialThenOverRefund(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	_, err := client.User.UpdateOneID(u.ID).SetBalance(1000).Save(ctx)
	if err != nil {
		t.Fatalf("set balance: %v", err)
	}

	order, err := client.RechargeOrder.Create().
		SetUserID(u.ID).
		SetAmount(1000).
		SetStatus(rechargeorder.StatusPaid).
		SetProvider(rechargeorder.ProviderMock).
		Save(ctx)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	repo := NewEntReconciliationRepository(client)
	updated, _, err := repo.RefundOrder(ctx, order.ID, 600)
	if err != nil {
		t.Fatalf("refund 600: %v", err)
	}
	if updated.RefundedAmount != 600 {
		t.Fatalf("expected refunded 600, got %d", updated.RefundedAmount)
	}

	// Remaining refundable is 400; refunding more must fail and must not
	// mutate the order or the balance.
	_, _, err = repo.RefundOrder(ctx, order.ID, 500)
	if err == nil {
		t.Fatal("expected error for refund beyond remaining amount")
	}

	freshOrder, err := client.RechargeOrder.Get(ctx, order.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if freshOrder.RefundedAmount != 600 {
		t.Fatalf("expected refunded_amount unchanged at 600, got %d", freshOrder.RefundedAmount)
	}

	freshUser, err := client.User.Get(ctx, u.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if freshUser.Balance != 400 {
		t.Fatalf("expected balance 400, got %d", freshUser.Balance)
	}

	// Refunding the exact remaining amount succeeds; after that the order is
	// fully refunded and any further refund is rejected.
	if _, _, err := repo.RefundOrder(ctx, order.ID, 400); err != nil {
		t.Fatalf("refund remaining 400: %v", err)
	}
	if _, _, err := repo.RefundOrder(ctx, order.ID, 1); err == nil {
		t.Fatal("expected error for refund after full refund")
	}
}

func TestEntReconciliationRepository_RefundOrderInsufficientBalance(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	order, err := client.RechargeOrder.Create().
		SetUserID(u.ID).
		SetAmount(1000).
		SetStatus(rechargeorder.StatusPaid).
		SetProvider(rechargeorder.ProviderMock).
		Save(ctx)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	repo := NewEntReconciliationRepository(client)
	_, _, err = repo.RefundOrder(ctx, order.ID, 100)
	if err == nil {
		t.Fatal("expected error for insufficient user balance")
	}
}
