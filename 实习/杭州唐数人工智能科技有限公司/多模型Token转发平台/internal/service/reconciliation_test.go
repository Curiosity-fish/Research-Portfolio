package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/internal/domain"
)

type mockReconciliationRepository struct {
	getWithRecordsFunc func(ctx context.Context, orderID uuid.UUID) (*ent.RechargeOrder, []*ent.BalanceRecord, error)
	refundOrderFunc    func(ctx context.Context, orderID uuid.UUID, amount int64) (*ent.RechargeOrder, *ent.BalanceRecord, error)
}

func (m *mockReconciliationRepository) GetRechargeOrderWithRecords(ctx context.Context, orderID uuid.UUID) (*ent.RechargeOrder, []*ent.BalanceRecord, error) {
	return m.getWithRecordsFunc(ctx, orderID)
}

func (m *mockReconciliationRepository) RefundOrder(ctx context.Context, orderID uuid.UUID, amount int64) (*ent.RechargeOrder, *ent.BalanceRecord, error) {
	return m.refundOrderFunc(ctx, orderID, amount)
}

func TestReconciliationService_GetOrderReconciliation(t *testing.T) {
	orderID := uuid.New()
	repo := &mockReconciliationRepository{
		getWithRecordsFunc: func(ctx context.Context, id uuid.UUID) (*ent.RechargeOrder, []*ent.BalanceRecord, error) {
			if id != orderID {
				t.Fatalf("expected order id %s, got %s", orderID, id)
			}
			order := &ent.RechargeOrder{ID: orderID, UserID: uuid.New(), Amount: 1000, Status: rechargeorder.StatusPaid, RefundedAmount: 0}
			return order, []*ent.BalanceRecord{}, nil
		},
	}
	svc := NewReconciliationService(repo)

	resp, err := svc.GetOrderReconciliation(context.Background(), orderID)
	if err != nil {
		t.Fatalf("get reconciliation: %v", err)
	}
	if resp.ID != orderID.String() {
		t.Fatalf("unexpected id %s", resp.ID)
	}
	if resp.RefundableAmount != 1000 {
		t.Fatalf("expected refundable 1000, got %d", resp.RefundableAmount)
	}
}

func TestReconciliationService_GetOrderReconciliationNotFound(t *testing.T) {
	repo := &mockReconciliationRepository{
		getWithRecordsFunc: func(ctx context.Context, id uuid.UUID) (*ent.RechargeOrder, []*ent.BalanceRecord, error) {
			return nil, nil, &ent.NotFoundError{}
		},
	}
	svc := NewReconciliationService(repo)

	_, err := svc.GetOrderReconciliation(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestReconciliationService_RefundOrderInvalidAmount(t *testing.T) {
	repo := &mockReconciliationRepository{
		refundOrderFunc: func(ctx context.Context, orderID uuid.UUID, amount int64) (*ent.RechargeOrder, *ent.BalanceRecord, error) {
			return nil, nil, domain.NewAppError(400, "INVALID_REFUND_AMOUNT", "invalid")
		},
	}
	svc := NewReconciliationService(repo)

	_, err := svc.RefundOrder(context.Background(), uuid.New(), uuid.New(), 2000, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	appErr := domain.AsAppError(err)
	if appErr.Code != 400 {
		t.Fatalf("expected 400, got %d", appErr.Code)
	}
}
