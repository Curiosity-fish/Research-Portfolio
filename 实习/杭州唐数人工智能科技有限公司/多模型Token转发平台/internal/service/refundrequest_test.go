package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/refundrequest"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// fakeRefundRequestRepo is an in-memory stand-in for the refund request repo.
type fakeRefundRequestRepo struct {
	byID    map[uuid.UUID]*ent.RefundRequest
	pending map[uuid.UUID]*ent.RefundRequest // keyed by order_id
}

func newFakeRefundRequestRepo() *fakeRefundRequestRepo {
	return &fakeRefundRequestRepo{byID: map[uuid.UUID]*ent.RefundRequest{}, pending: map[uuid.UUID]*ent.RefundRequest{}}
}

func (f *fakeRefundRequestRepo) Create(_ context.Context, input repository.CreateRefundRequestInput) (*ent.RefundRequest, error) {
	req := &ent.RefundRequest{
		ID:        uuid.New(),
		OrderID:   input.OrderID,
		UserID:    input.UserID,
		Amount:    input.Amount,
		Reason:    input.Reason,
		Status:    refundrequest.StatusPending,
		CreatedAt: time.Now(),
	}
	req.UpdatedAt = req.CreatedAt
	f.byID[req.ID] = req
	f.pending[req.OrderID] = req
	return req, nil
}

func (f *fakeRefundRequestRepo) GetByID(_ context.Context, id uuid.UUID) (*ent.RefundRequest, error) {
	if r, ok := f.byID[id]; ok {
		return r, nil
	}
	return nil, &ent.NotFoundError{}
}

func (f *fakeRefundRequestRepo) GetPendingByOrderID(_ context.Context, orderID uuid.UUID) (*ent.RefundRequest, error) {
	return f.pending[orderID], nil
}

func (f *fakeRefundRequestRepo) List(_ context.Context, filter repository.ListRefundRequestFilter) ([]*ent.RefundRequest, int, error) {
	var out []*ent.RefundRequest
	for _, r := range f.byID {
		if filter.Status != nil && string(r.Status) != string(*filter.Status) {
			continue
		}
		if filter.UserID != nil && r.UserID != *filter.UserID {
			continue
		}
		if filter.OrderID != nil && r.OrderID != *filter.OrderID {
			continue
		}
		out = append(out, r)
	}
	return out, len(out), nil
}

func (f *fakeRefundRequestRepo) Transition(_ context.Context, id uuid.UUID, expect, status repository.RefundRequestStatus, reviewedBy *uuid.UUID) (*ent.RefundRequest, error) {
	r, ok := f.byID[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if string(r.Status) != string(expect) {
		return nil, repository.ErrRefundRequestConflict
	}
	r.Status = refundrequest.Status(status)
	if reviewedBy != nil {
		r.ReviewedBy = reviewedBy
		now := time.Now()
		r.ReviewedAt = &now
	} else {
		r.ReviewedBy = nil
		r.ReviewedAt = nil
	}
	if status == repository.RefundRequestStatusPending {
		f.pending[r.OrderID] = r
	} else {
		delete(f.pending, r.OrderID)
	}
	return r, nil
}

// fakeReconciliation stubs the money path.
type fakeReconciliation struct {
	orderByID   map[uuid.UUID]*ent.RechargeOrder
	refundCalls int
	refundErr   error
}

func (f *fakeReconciliation) GetRechargeOrderWithRecords(_ context.Context, orderID uuid.UUID) (*ent.RechargeOrder, []*ent.BalanceRecord, error) {
	o, ok := f.orderByID[orderID]
	if !ok {
		return nil, nil, &ent.NotFoundError{}
	}
	return o, nil, nil
}

func (f *fakeReconciliation) RefundOrder(_ context.Context, orderID uuid.UUID, amount int64) (*ent.RechargeOrder, *ent.BalanceRecord, error) {
	f.refundCalls++
	if f.refundErr != nil {
		return nil, nil, f.refundErr
	}
	o := f.orderByID[orderID]
	o.RefundedAmount += amount
	return o, &ent.BalanceRecord{Amount: amount}, nil
}

func paidOrder(id, userID uuid.UUID, amount int64) *ent.RechargeOrder {
	return &ent.RechargeOrder{ID: id, UserID: userID, Amount: amount, Status: "paid"}
}

func TestRefundRequestService_CreateValidatesOwnershipAndPending(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	orderID := uuid.New()
	otherUser := uuid.New()

	repo := newFakeRefundRequestRepo()
	recon := &fakeReconciliation{orderByID: map[uuid.UUID]*ent.RechargeOrder{orderID: paidOrder(orderID, userID, 10000)}}
	svc := NewRefundRequestService(repo, recon)

	// Owner can open a request.
	resp, err := svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: userID, OrderID: orderID, Amount: 3000})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.Status != "pending" {
		t.Fatalf("expected pending, got %s", resp.Status)
	}

	// A second pending request for the same order is rejected.
	_, err = svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: userID, OrderID: orderID, Amount: 1000})
	if err == nil {
		t.Fatal("expected duplicate-pending rejection")
	}

	// Another user gets 404, not a leak of the order's existence.
	_, err = svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: otherUser, OrderID: orderID, Amount: 1000})
	appErr, ok := err.(*domain.AppError)
	if !ok || appErr.Code != 404 {
		t.Fatalf("expected 404 for non-owner, got %v", err)
	}

	// Non-positive amount is rejected before touching the repos.
	_, err = svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: userID, OrderID: uuid.New(), Amount: 0})
	if err == nil {
		t.Fatal("expected amount validation error")
	}
}

func TestRefundRequestService_ApproveMovesMoneyOnce(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	orderID := uuid.New()
	adminID := uuid.New()

	repo := newFakeRefundRequestRepo()
	recon := &fakeReconciliation{orderByID: map[uuid.UUID]*ent.RechargeOrder{orderID: paidOrder(orderID, userID, 10000)}}
	svc := NewRefundRequestService(repo, recon)

	resp, err := svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: userID, OrderID: orderID, Amount: 4000})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	reqID := uuid.MustParse(resp.ID)

	approved, err := svc.ApproveRefundRequest(ctx, reqID, adminID)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if approved.Status != "approved" {
		t.Fatalf("expected approved, got %s", approved.Status)
	}
	if recon.refundCalls != 1 {
		t.Fatalf("expected money path called once, got %d", recon.refundCalls)
	}
	if recon.orderByID[orderID].RefundedAmount != 4000 {
		t.Fatalf("expected refunded_amount 4000, got %d", recon.orderByID[orderID].RefundedAmount)
	}

	// A second approval attempt must not move money again.
	if _, err := svc.ApproveRefundRequest(ctx, reqID, adminID); err == nil {
		t.Fatal("expected second approval to fail")
	}
	if recon.refundCalls != 1 {
		t.Fatalf("money path must stay single-writer, got %d calls", recon.refundCalls)
	}
}

func TestRefundRequestService_RejectDoesNotMoveMoney(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	orderID := uuid.New()
	adminID := uuid.New()

	repo := newFakeRefundRequestRepo()
	recon := &fakeReconciliation{orderByID: map[uuid.UUID]*ent.RechargeOrder{orderID: paidOrder(orderID, userID, 10000)}}
	svc := NewRefundRequestService(repo, recon)

	resp, err := svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: userID, OrderID: orderID, Amount: 2000})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	reqID := uuid.MustParse(resp.ID)

	rejected, err := svc.RejectRefundRequest(ctx, reqID, adminID)
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if rejected.Status != "rejected" {
		t.Fatalf("expected rejected, got %s", rejected.Status)
	}
	if recon.refundCalls != 0 {
		t.Fatalf("reject must not move money, got %d refund calls", recon.refundCalls)
	}

	// After rejection a new request may be opened for the same order.
	if _, err := svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: userID, OrderID: orderID, Amount: 1000}); err != nil {
		t.Fatalf("re-create after reject: %v", err)
	}
}

// When the money path fails after the approval transition, the request is
// reopened so a later approval can retry.
func TestRefundRequestService_ApproveFailureRevertsToPending(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	orderID := uuid.New()
	adminID := uuid.New()

	repo := newFakeRefundRequestRepo()
	recon := &fakeReconciliation{
		orderByID: map[uuid.UUID]*ent.RechargeOrder{orderID: paidOrder(orderID, userID, 10000)},
		refundErr: domain.NewAppError(400, "INSUFFICIENT_BALANCE_FOR_REFUND", "用户余额不足以退款"),
	}
	svc := NewRefundRequestService(repo, recon)

	resp, err := svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: userID, OrderID: orderID, Amount: 4000})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	reqID := uuid.MustParse(resp.ID)

	if _, err := svc.ApproveRefundRequest(ctx, reqID, adminID); err == nil {
		t.Fatal("expected approval to fail when the money path fails")
	}
	if got := repo.byID[reqID].Status; got != refundrequest.StatusPending {
		t.Fatalf("expected reverted to pending, got %s", got)
	}
	if repo.byID[reqID].ReviewedBy != nil {
		t.Fatal("revert must clear the reviewer")
	}

	// The refund succeeds on retry once the money path recovers.
	recon.refundErr = nil
	if _, err := svc.ApproveRefundRequest(ctx, reqID, adminID); err != nil {
		t.Fatalf("retry approve: %v", err)
	}
	if recon.refundCalls != 2 {
		t.Fatalf("expected 2 refund attempts total, got %d", recon.refundCalls)
	}
}

func TestRefundRequestService_Cancel(t *testing.T) {	ctx := context.Background()
	userID := uuid.New()
	otherUser := uuid.New()
	orderID := uuid.New()

	repo := newFakeRefundRequestRepo()
	recon := &fakeReconciliation{orderByID: map[uuid.UUID]*ent.RechargeOrder{orderID: paidOrder(orderID, userID, 10000)}}
	svc := NewRefundRequestService(repo, recon)

	resp, err := svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: userID, OrderID: orderID, Amount: 5000})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	reqID := uuid.MustParse(resp.ID)

	// Another user cannot cancel the request (404, not a leak).
	if err := svc.CancelRefundRequest(ctx, otherUser, reqID); err == nil {
		t.Fatal("expected cancel by non-owner to fail")
	}
	if repo.byID[reqID].Status != refundrequest.StatusPending {
		t.Fatal("non-owner cancel must leave the request pending")
	}

	// The owner cancels; cancellation is recorded as a rejection without a reviewer.
	if err := svc.CancelRefundRequest(ctx, userID, reqID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if got := repo.byID[reqID].Status; got != refundrequest.StatusRejected {
		t.Fatalf("expected cancelled->rejected, got %s", got)
	}
	if repo.byID[reqID].ReviewedBy != nil {
		t.Fatal("user-initiated cancel must not set a reviewer")
	}

	// A cancelled request cannot be cancelled again.
	if err := svc.CancelRefundRequest(ctx, userID, reqID); err == nil {
		t.Fatal("expected cancelling a settled request to fail")
	}

	// Cancellation frees the order for a new request.
	if _, err := svc.CreateRefundRequest(ctx, CreateRefundRequestInput{UserID: userID, OrderID: orderID, Amount: 1000}); err != nil {
		t.Fatalf("re-create after cancel: %v", err)
	}
}
