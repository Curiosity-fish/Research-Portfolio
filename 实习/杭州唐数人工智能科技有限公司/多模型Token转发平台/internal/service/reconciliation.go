package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// ReconciliationOrderResponse is the public representation of an order reconciliation.
type ReconciliationOrderResponse struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Amount          int64                  `json:"amount"`
	RefundedAmount  int64                  `json:"refunded_amount"`
	RefundableAmount int64                 `json:"refundable_amount"`
	Status          string                 `json:"status"`
	Provider        string                 `json:"provider"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
	BalanceRecords  []BalanceRecordResponse `json:"balance_records"`
}

// RefundOrderResponse is the public representation of a refund result.
type RefundOrderResponse struct {
	Order         ReconciliationOrderResponse `json:"order"`
	BalanceRecord BalanceRecordResponse       `json:"balance_record"`
}

// RefundOrderRequest is the request body for refunding an order.
type RefundOrderRequest struct {
	Amount int64  `json:"amount" binding:"required,gt=0"`
	Reason string `json:"reason,omitempty" binding:"omitempty,max=500"`
}

// ReconciliationService handles order reconciliation and refunds.
type ReconciliationService struct {
	repo repository.ReconciliationRepository
}

// NewReconciliationService creates a new ReconciliationService.
func NewReconciliationService(repo repository.ReconciliationRepository) *ReconciliationService {
	return &ReconciliationService{repo: repo}
}

// GetOrderReconciliation returns the reconciliation details for a recharge order.
func (s *ReconciliationService) GetOrderReconciliation(ctx context.Context, orderID uuid.UUID) (*ReconciliationOrderResponse, error) {
	order, records, err := s.repo.GetRechargeOrderWithRecords(ctx, orderID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	return toReconciliationOrderResponse(order, records), nil
}

// RefundOrder processes a refund for a paid recharge order.
func (s *ReconciliationService) RefundOrder(ctx context.Context, adminID, orderID uuid.UUID, amount int64, reason *string) (*RefundOrderResponse, error) {
	order, rec, err := s.repo.RefundOrder(ctx, orderID, amount)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.AsAppError(err)
	}

	_, records, err := s.repo.GetRechargeOrderWithRecords(ctx, order.ID)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	_ = adminID
	_ = reason

	return &RefundOrderResponse{
		Order:         *toReconciliationOrderResponse(order, records),
		BalanceRecord: *toBalanceRecordResponse(rec),
	}, nil
}

func toReconciliationOrderResponse(order *ent.RechargeOrder, records []*ent.BalanceRecord) *ReconciliationOrderResponse {
	resp := &ReconciliationOrderResponse{
		ID:               order.ID.String(),
		UserID:           order.UserID.String(),
		Amount:           order.Amount,
		RefundedAmount:   order.RefundedAmount,
		RefundableAmount: order.Amount - order.RefundedAmount,
		Status:           string(order.Status),
		Provider:         string(order.Provider),
		CreatedAt:        order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		BalanceRecords:   make([]BalanceRecordResponse, 0, len(records)),
	}
	for _, rec := range records {
		resp.BalanceRecords = append(resp.BalanceRecords, *toBalanceRecordResponse(rec))
	}
	return resp
}
