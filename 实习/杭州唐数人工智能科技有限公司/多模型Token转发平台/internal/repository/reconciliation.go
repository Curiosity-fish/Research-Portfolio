package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/domain"
)

// ReconciliationRepository provides data access for order reconciliation and refunds.
type ReconciliationRepository interface {
	GetRechargeOrderWithRecords(ctx context.Context, orderID uuid.UUID) (*ent.RechargeOrder, []*ent.BalanceRecord, error)
	RefundOrder(ctx context.Context, orderID uuid.UUID, amount int64) (*ent.RechargeOrder, *ent.BalanceRecord, error)
}

// EntReconciliationRepository implements ReconciliationRepository using Ent transactions.
type EntReconciliationRepository struct {
	client *ent.Client
}

// NewEntReconciliationRepository creates a new Ent-backed reconciliation repository.
func NewEntReconciliationRepository(client *ent.Client) *EntReconciliationRepository {
	return &EntReconciliationRepository{client: client}
}

// GetRechargeOrderWithRecords returns an order and all balance records linked to it.
func (r *EntReconciliationRepository) GetRechargeOrderWithRecords(ctx context.Context, orderID uuid.UUID) (*ent.RechargeOrder, []*ent.BalanceRecord, error) {
	order, err := r.client.RechargeOrder.Get(ctx, orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("get recharge order: %w", err)
	}

	records, err := r.client.BalanceRecord.Query().
		Where(balancerecord.RelatedOrderIDEQ(orderID)).
		Order(ent.Desc(balancerecord.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list balance records for order: %w", err)
	}

	return order, records, nil
}

// RefundOrder refunds part of a paid recharge order.
// It updates the order's refunded_amount, decreases the user's balance, and writes a refund record.
//
// The order and balance mutations use conditional updates inside the transaction
// so concurrent refunds cannot over-refund or drive the balance negative.
func (r *EntReconciliationRepository) RefundOrder(ctx context.Context, orderID uuid.UUID, amount int64) (*ent.RechargeOrder, *ent.BalanceRecord, error) {
	order, err := r.client.RechargeOrder.Get(ctx, orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("get recharge order: %w", err)
	}
	if order.Status != rechargeorder.StatusPaid {
		return nil, nil, domain.NewAppError(400, "ORDER_NOT_PAID", "只有已支付订单可以申请退款")
	}

	maxRefund := order.Amount - order.RefundedAmount
	if amount <= 0 || amount > maxRefund {
		return nil, nil, domain.NewAppError(400, "INVALID_REFUND_AMOUNT", "退款金额超出可退范围")
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("start refund transaction: %w", err)
	}

	rollback := func(err error) (*ent.RechargeOrder, *ent.BalanceRecord, error) {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return nil, nil, err
	}

	orderAffected, err := tx.RechargeOrder.Update().
		Where(
			rechargeorder.IDEQ(orderID),
			rechargeorder.StatusEQ(rechargeorder.StatusPaid),
			rechargeorder.RefundedAmountLTE(order.Amount-amount),
		).
		AddRefundedAmount(amount).
		Save(ctx)
	if err != nil {
		return rollback(fmt.Errorf("update refunded amount: %w", err))
	}
	if orderAffected == 0 {
		return rollback(domain.NewAppError(400, "INVALID_REFUND_AMOUNT", "退款金额超出可退范围"))
	}

	balanceAffected, err := tx.User.Update().
		Where(user.IDEQ(order.UserID), user.BalanceGTE(amount)).
		AddBalance(-amount).
		Save(ctx)
	if err != nil {
		return rollback(fmt.Errorf("decrease user balance: %w", err))
	}
	if balanceAffected == 0 {
		return rollback(domain.NewAppError(400, "INSUFFICIENT_BALANCE_FOR_REFUND", "用户余额不足以退款"))
	}

	updatedUser, err := tx.User.Get(ctx, order.UserID)
	if err != nil {
		return rollback(fmt.Errorf("get updated user: %w", err))
	}

	rec, err := tx.BalanceRecord.Create().
		SetUserID(order.UserID).
		SetType(balancerecord.TypeRefund).
		SetAmount(amount).
		SetBalanceAfter(updatedUser.Balance).
		SetRelatedOrderID(orderID).
		Save(ctx)
	if err != nil {
		return rollback(fmt.Errorf("create refund balance record: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit refund: %w", err)
	}

	updatedOrder, err := r.client.RechargeOrder.Get(ctx, orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("get updated recharge order: %w", err)
	}
	return updatedOrder, rec, nil
}

var _ ReconciliationRepository = (*EntReconciliationRepository)(nil)
