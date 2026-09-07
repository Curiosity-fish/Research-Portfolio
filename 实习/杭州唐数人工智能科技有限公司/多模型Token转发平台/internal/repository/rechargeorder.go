package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/rechargeorder"
	"github.com/school-api/school-api-v1/internal/domain"
)

// RechargeOrderStatus represents the status of a recharge order.
type RechargeOrderStatus string

const (
	RechargeOrderStatusPending   RechargeOrderStatus = "pending"
	RechargeOrderStatusPaid      RechargeOrderStatus = "paid"
	RechargeOrderStatusFailed    RechargeOrderStatus = "failed"
	RechargeOrderStatusCancelled RechargeOrderStatus = "cancelled"
)

// CreateRechargeOrderInput is the data required to create a recharge order.
type CreateRechargeOrderInput struct {
	UserID   uuid.UUID
	Amount   int64
	Provider string
}

// RechargeOrderRepository provides data access for recharge orders.
type RechargeOrderRepository interface {
	Create(ctx context.Context, input CreateRechargeOrderInput) (*ent.RechargeOrder, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.RechargeOrder, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.RechargeOrder, int, error)
	List(ctx context.Context, userID *uuid.UUID, offset, limit int) ([]*ent.RechargeOrder, int, error)
	// ListRecentWithUser returns the most recent orders (newest first) with the
	// owning user eager-loaded, for XLSX export. Callers enforce the export row
	// cap via limit.
	ListRecentWithUser(ctx context.Context, limit int) ([]*ent.RechargeOrder, error)
	MarkPaid(ctx context.Context, id uuid.UUID, providerOrderID string) (*ent.RechargeOrder, error)
	MarkFailed(ctx context.Context, id uuid.UUID) (*ent.RechargeOrder, error)
}

// EntRechargeOrderRepository implements RechargeOrderRepository using Ent.
type EntRechargeOrderRepository struct {
	client *ent.Client
}

// NewEntRechargeOrderRepository creates a new Ent-backed recharge order repository.
func NewEntRechargeOrderRepository(client *ent.Client) *EntRechargeOrderRepository {
	return &EntRechargeOrderRepository{client: client}
}

// Create inserts a new recharge order.
func (r *EntRechargeOrderRepository) Create(ctx context.Context, input CreateRechargeOrderInput) (*ent.RechargeOrder, error) {
	order, err := r.client.RechargeOrder.Create().
		SetUserID(input.UserID).
		SetAmount(input.Amount).
		SetProvider(rechargeorder.Provider(input.Provider)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create recharge order: %w", err)
	}
	return order, nil
}

// GetByID returns a recharge order by ID.
func (r *EntRechargeOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.RechargeOrder, error) {
	order, err := r.client.RechargeOrder.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get recharge order: %w", err)
	}
	return order, nil
}

// ListByUserID returns a paginated list of recharge orders for a user.
func (r *EntRechargeOrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.RechargeOrder, int, error) {
	q := r.client.RechargeOrder.Query().Where(rechargeorder.UserIDEQ(userID))

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count recharge orders: %w", err)
	}

	orders, err := q.Order(ent.Desc(rechargeorder.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list recharge orders: %w", err)
	}

	return orders, total, nil
}

// MarkPaid marks an order as paid, increases the user's balance, and writes a balance record.
// It is idempotent: if the order is already paid, it returns the existing order unchanged.
func (r *EntRechargeOrderRepository) MarkPaid(ctx context.Context, id uuid.UUID, providerOrderID string) (*ent.RechargeOrder, error) {
	order, err := r.client.RechargeOrder.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get recharge order: %w", err)
	}
	if order.Status == rechargeorder.StatusPaid {
		return order, nil
	}
	if order.Status != rechargeorder.StatusPending {
		return nil, domain.NewAppError(409, "ORDER_NOT_PENDING", "订单状态不允许该操作")
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start transaction: %w", err)
	}

	rollback := func(err error) (*ent.RechargeOrder, error) {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return nil, err
	}

	// Conditional update so concurrent callbacks for the same order cannot
	// both credit the balance; only a pending order transitions to paid.
	affected, err := tx.RechargeOrder.Update().
		Where(
			rechargeorder.IDEQ(id),
			rechargeorder.StatusEQ(rechargeorder.StatusPending),
		).
		SetStatus(rechargeorder.StatusPaid).
		SetProviderOrderID(providerOrderID).
		SetPaidAt(time.Now()).
		Save(ctx)
	if err != nil {
		return rollback(fmt.Errorf("mark order paid: %w", err))
	}
	if affected == 0 {
		// Lost the race against another callback: re-read and return the
		// current row so the caller stays idempotent.
		latest, gerr := r.client.RechargeOrder.Get(ctx, id)
		if gerr == nil && latest.Status == rechargeorder.StatusPaid {
			// The transaction only ran a no-op UPDATE; roll it back so the
			// underlying *sql.Tx releases its connection back to the pool.
			_ = tx.Rollback()
			return latest, nil
		}
		return rollback(domain.NewAppError(409, "ORDER_NOT_PENDING", "订单状态不允许该操作"))
	}

	if _, err := tx.User.UpdateOneID(order.UserID).AddBalance(order.Amount).Save(ctx); err != nil {
		return rollback(fmt.Errorf("increase user balance: %w", err))
	}

	u, err := tx.User.Get(ctx, order.UserID)
	if err != nil {
		return rollback(fmt.Errorf("get updated user: %w", err))
	}

	if _, err := tx.BalanceRecord.Create().
		SetUserID(order.UserID).
		SetType(balancerecord.TypeRecharge).
		SetAmount(order.Amount).
		SetBalanceAfter(u.Balance).
		SetRelatedOrderID(id).
		Save(ctx); err != nil {
		return rollback(fmt.Errorf("create recharge balance record: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit recharge payment: %w", err)
	}

	updated, err := r.client.RechargeOrder.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get updated recharge order: %w", err)
	}
	return updated, nil
}

// MarkFailed marks a pending order as failed.
func (r *EntRechargeOrderRepository) MarkFailed(ctx context.Context, id uuid.UUID) (*ent.RechargeOrder, error) {
	order, err := r.client.RechargeOrder.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get recharge order: %w", err)
	}
	if order.Status != rechargeorder.StatusPending {
		return nil, domain.NewAppError(409, "ORDER_NOT_PENDING", "订单状态不允许该操作")
	}

	updated, err := r.client.RechargeOrder.UpdateOneID(id).
		SetStatus(rechargeorder.StatusFailed).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("mark order failed: %w", err)
	}
	return updated, nil
}

// List returns a paginated list of recharge orders, optionally filtered by user.
func (r *EntRechargeOrderRepository) List(ctx context.Context, userID *uuid.UUID, offset, limit int) ([]*ent.RechargeOrder, int, error) {
	q := r.client.RechargeOrder.Query()
	if userID != nil {
		q = q.Where(rechargeorder.UserIDEQ(*userID))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count recharge orders: %w", err)
	}

	orders, err := q.Order(ent.Desc(rechargeorder.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list recharge orders: %w", err)
	}

	return orders, total, nil
}

// ListRecentWithUser returns recent orders with the owning user eager-loaded.
func (r *EntRechargeOrderRepository) ListRecentWithUser(ctx context.Context, limit int) ([]*ent.RechargeOrder, error) {
	orders, err := r.client.RechargeOrder.Query().
		WithUser().
		Order(ent.Desc(rechargeorder.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recharge orders for export: %w", err)
	}
	return orders, nil
}

var _ RechargeOrderRepository = (*EntRechargeOrderRepository)(nil)
