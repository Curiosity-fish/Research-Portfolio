package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// RechargeOrderResponse is the public representation of a recharge order.
type RechargeOrderResponse struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Amount          int64      `json:"amount"`
	Status          string     `json:"status"`
	Provider        string     `json:"provider"`
	ProviderOrderID *string    `json:"provider_order_id,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

// RechargeOrderService handles recharge order business logic.
type RechargeOrderService struct {
	repo     repository.RechargeOrderRepository
	settings RechargePolicy
}

// NewRechargeOrderService creates a new RechargeOrderService. settings may be
// nil, in which case feature-mode gating and amount-limit checks are skipped.
func NewRechargeOrderService(repo repository.RechargeOrderRepository, settings RechargePolicy) *RechargeOrderService {
	return &RechargeOrderService{repo: repo, settings: settings}
}

// CreateRechargeOrder creates a pending recharge order for a user.
func (s *RechargeOrderService) CreateRechargeOrder(ctx context.Context, userID uuid.UUID, amount int64) (*RechargeOrderResponse, error) {
	if amount <= 0 {
		return nil, domain.NewAppError(400, "INVALID_REQUEST", "充值金额必须大于 0")
	}

	if err := s.checkPolicy(ctx, amount); err != nil {
		return nil, err
	}

	order, err := s.repo.Create(ctx, repository.CreateRechargeOrderInput{
		UserID:   userID,
		Amount:   amount,
		Provider: "mock",
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return toRechargeOrderResponse(order), nil
}

// checkPolicy enforces the feature mode and the configured amount limits.
// A nil settings provider means the platform runs unrestricted.
func (s *RechargeOrderService) checkPolicy(ctx context.Context, amount int64) error {
	if s.settings == nil {
		return nil
	}
	mode, err := s.settings.FeatureMode(ctx)
	if err != nil {
		return err
	}
	if mode == FeatureModeQuotaOnly {
		return domain.ErrFeatureDisabled
	}

	limits, err := s.settings.RechargeLimits(ctx)
	if err != nil {
		return err
	}
	if limits == nil {
		return nil
	}
	if amount < limits.MinAmount {
		return domain.NewAppError(400, "AMOUNT_OUT_OF_RANGE",
			fmt.Sprintf("充值金额不能低于 %.2f 元", float64(limits.MinAmount)/1e6))
	}
	if amount > limits.MaxAmount {
		return domain.NewAppError(400, "AMOUNT_OUT_OF_RANGE",
			fmt.Sprintf("充值金额不能高于 %.2f 元", float64(limits.MaxAmount)/1e6))
	}
	return nil
}

// GetRechargeOrder returns a recharge order if it belongs to the user.
func (s *RechargeOrderService) GetRechargeOrder(ctx context.Context, userID, orderID uuid.UUID) (*RechargeOrderResponse, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	if order.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return toRechargeOrderResponse(order), nil
}

// ListRechargeOrders returns a paginated list of recharge orders for a user.
func (s *RechargeOrderService) ListRechargeOrders(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]RechargeOrderResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	orders, total, err := s.repo.ListByUserID(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]RechargeOrderResponse, 0, len(orders))
	for _, order := range orders {
		resp = append(resp, *toRechargeOrderResponse(order))
	}
	return resp, total, nil
}

// ListRechargeOrdersAdmin returns a paginated list of recharge orders for admin, optionally filtered by user.
func (s *RechargeOrderService) ListRechargeOrdersAdmin(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]RechargeOrderResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	orders, total, err := s.repo.List(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]RechargeOrderResponse, 0, len(orders))
	for _, order := range orders {
		resp = append(resp, *toRechargeOrderResponse(order))
	}
	return resp, total, nil
}

// ProcessMockCallback simulates a payment provider callback.
// The order must belong to the user; a mismatched order is reported as
// not found so the endpoint does not leak order existence across users.
func (s *RechargeOrderService) ProcessMockCallback(ctx context.Context, userID, orderID uuid.UUID, success bool) (*RechargeOrderResponse, error) {
	existing, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	if existing.UserID != userID {
		return nil, domain.ErrNotFound
	}

	var order *ent.RechargeOrder
	if success {
		providerOrderID := fmt.Sprintf("mock-%s", uuid.NewString())
		order, err = s.repo.MarkPaid(ctx, orderID, providerOrderID)
	} else {
		order, err = s.repo.MarkFailed(ctx, orderID)
	}
	if err != nil {
		// Repository business errors (e.g. order no longer pending) are
		// AppErrors and must pass through instead of becoming a 500.
		return nil, domain.AsAppError(err)
	}
	return toRechargeOrderResponse(order), nil
}

func toRechargeOrderResponse(order *ent.RechargeOrder) *RechargeOrderResponse {
	resp := &RechargeOrderResponse{
		ID:        order.ID.String(),
		UserID:    order.UserID.String(),
		Amount:    order.Amount,
		Status:    string(order.Status),
		Provider:  string(order.Provider),
		CreatedAt: order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if order.ProviderOrderID != nil {
		resp.ProviderOrderID = order.ProviderOrderID
	}
	if order.PaidAt != nil {
		resp.PaidAt = order.PaidAt
	}
	return resp
}
