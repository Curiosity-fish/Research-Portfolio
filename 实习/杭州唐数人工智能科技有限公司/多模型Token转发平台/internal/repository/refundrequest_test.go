package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/ent/refundrequest"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntRefundRequestRepository_CreateAndGet(t *testing.T) {
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

	repo := NewEntRefundRequestRepository(client)
	reason := "买错了"
	req, err := repo.Create(ctx, CreateRefundRequestInput{
		OrderID: order.ID,
		UserID:  u.ID,
		Amount:  300,
		Reason:  &reason,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if req.Status != refundrequest.StatusPending {
		t.Fatalf("expected pending, got %s", req.Status)
	}

	got, err := repo.GetByID(ctx, req.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Amount != 300 || got.Reason == nil || *got.Reason != reason {
		t.Fatalf("unexpected row: %+v", got)
	}

	pending, err := repo.GetPendingByOrderID(ctx, order.ID)
	if err != nil {
		t.Fatalf("get pending: %v", err)
	}
	if pending == nil || pending.ID != req.ID {
		t.Fatal("expected the pending request for the order")
	}

	// No pending request for an unrelated order ID.
	none, err := repo.GetPendingByOrderID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("get pending for unknown order: %v", err)
	}
	if none != nil {
		t.Fatal("expected nil for an order without a pending request")
	}
}

func TestEntRefundRequestRepository_ListFilters(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	u2, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	newOrder := func(userID uuid.UUID) uuid.UUID {
		o, err := client.RechargeOrder.Create().
			SetUserID(userID).
			SetAmount(1000).
			SetStatus(rechargeorder.StatusPaid).
			SetProvider(rechargeorder.ProviderMock).
			Save(ctx)
		if err != nil {
			t.Fatalf("create order: %v", err)
		}
		return o.ID
	}
	orderA, orderB := newOrder(u.ID), newOrder(u2.ID)

	repo := NewEntRefundRequestRepository(client)
	if _, err := repo.Create(ctx, CreateRefundRequestInput{OrderID: orderA, UserID: u.ID, Amount: 100}); err != nil {
		t.Fatalf("create A: %v", err)
	}
	if _, err := repo.Create(ctx, CreateRefundRequestInput{OrderID: orderB, UserID: u2.ID, Amount: 200}); err != nil {
		t.Fatalf("create B: %v", err)
	}

	all, total, err := repo.List(ctx, ListRefundRequestFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if total != 2 || len(all) != 2 {
		t.Fatalf("expected 2 rows, got total=%d len=%d", total, len(all))
	}

	byUser, total, err := repo.List(ctx, ListRefundRequestFilter{UserID: &u.ID, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list by user: %v", err)
	}
	if total != 1 || len(byUser) != 1 || byUser[0].UserID != u.ID {
		t.Fatalf("user filter mismatch: total=%d", total)
	}

	byOrder, total, err := repo.List(ctx, ListRefundRequestFilter{OrderID: &orderB, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list by order: %v", err)
	}
	if total != 1 || len(byOrder) != 1 || byOrder[0].OrderID != orderB {
		t.Fatalf("order filter mismatch: total=%d", total)
	}

	pending := RefundRequestStatusPending
	byStatus, total, err := repo.List(ctx, ListRefundRequestFilter{Status: &pending, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list by status: %v", err)
	}
	if total != 2 || len(byStatus) != 2 {
		t.Fatalf("expected 2 pending, got total=%d len=%d", total, len(byStatus))
	}
}

func TestEntRefundRequestRepository_Transition(t *testing.T) {
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

	repo := NewEntRefundRequestRepository(client)
	req, err := repo.Create(ctx, CreateRefundRequestInput{OrderID: order.ID, UserID: u.ID, Amount: 100})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Admin review sets the reviewer and the timestamp.
	reviewer := uuid.New()
	updated, err := repo.Transition(ctx, req.ID, RefundRequestStatusPending, RefundRequestStatusApproved, &reviewer)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if updated.Status != refundrequest.StatusApproved {
		t.Fatalf("expected approved, got %s", updated.Status)
	}
	if updated.ReviewedBy == nil || *updated.ReviewedBy != reviewer {
		t.Fatal("reviewer not recorded")
	}
	if updated.ReviewedAt == nil {
		t.Fatal("reviewed_at not recorded")
	}

	// Repeating the same transition conflicts: the request already settled.
	if _, err := repo.Transition(ctx, req.ID, RefundRequestStatusPending, RefundRequestStatusApproved, &reviewer); !errors.Is(err, ErrRefundRequestConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}

	// After settling, the order has no pending request anymore.
	pending, err := repo.GetPendingByOrderID(ctx, order.ID)
	if err != nil {
		t.Fatalf("get pending: %v", err)
	}
	if pending != nil {
		t.Fatal("expected no pending request after approval")
	}

	// A nil reviewer clears the reviewer fields (user-initiated cancel).
	req2, err := repo.Create(ctx, CreateRefundRequestInput{OrderID: order.ID, UserID: u.ID, Amount: 50})
	if err != nil {
		t.Fatalf("create second: %v", err)
	}
	cancelled, err := repo.Transition(ctx, req2.ID, RefundRequestStatusPending, RefundRequestStatusRejected, nil)
	if err != nil {
		t.Fatalf("cancel transition: %v", err)
	}
	if cancelled.Status != refundrequest.StatusRejected {
		t.Fatalf("expected rejected, got %s", cancelled.Status)
	}
	if cancelled.ReviewedBy != nil || cancelled.ReviewedAt != nil {
		t.Fatal("nil reviewer must clear reviewer fields")
	}

	// Reverting an approved request back to pending (failed-refund path).
	req3, err := repo.Create(ctx, CreateRefundRequestInput{OrderID: order.ID, UserID: u.ID, Amount: 10})
	if err != nil {
		t.Fatalf("create third: %v", err)
	}
	if _, err := repo.Transition(ctx, req3.ID, RefundRequestStatusPending, RefundRequestStatusApproved, &reviewer); err != nil {
		t.Fatalf("approve third: %v", err)
	}
	reverted, err := repo.Transition(ctx, req3.ID, RefundRequestStatusApproved, RefundRequestStatusPending, nil)
	if err != nil {
		t.Fatalf("revert third: %v", err)
	}
	if reverted.Status != refundrequest.StatusPending || reverted.ReviewedBy != nil || reverted.ReviewedAt != nil {
		t.Fatalf("expected clean pending after revert, got %+v", reverted)
	}
}
