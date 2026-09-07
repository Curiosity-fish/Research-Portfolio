package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// RefundRequestResponse is the public representation of a refund request.
type RefundRequestResponse struct {
	ID         string     `json:"id"`
	OrderID    string     `json:"order_id"`
	UserID     string     `json:"user_id"`
	Amount     int64      `json:"amount"`
	Reason     *string    `json:"reason,omitempty"`
	Status     string     `json:"status"`
	ReviewedBy *string    `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt  string     `json:"created_at"`
	UpdatedAt  string     `json:"updated_at"`
}

// CreateRefundRequestInput is the service-level input for opening a request.
type CreateRefundRequestInput struct {
	UserID uuid.UUID
	OrderID uuid.UUID
	Amount int64
	Reason *string
}

// ListRefundRequestFilter is the service-level filter for listing requests.
type ListRefundRequestFilter struct {
	Status   *string
	UserID   *uuid.UUID
	OrderID  *uuid.UUID
	Page     int
	PageSize int
}

// RefundRequestService handles the user-initiated refund flow.
type RefundRequestService struct {
	requestRepo    repository.RefundRequestRepository
	reconciliation repository.ReconciliationRepository
}

// NewRefundRequestService creates a new RefundRequestService.
func NewRefundRequestService(requestRepo repository.RefundRequestRepository, reconciliation repository.ReconciliationRepository) *RefundRequestService {
	return &RefundRequestService{requestRepo: requestRepo, reconciliation: reconciliation}
}

// CreateRefundRequest opens a refund request for one of the user's paid orders.
// At most one pending request per order is allowed.
func (s *RefundRequestService) CreateRefundRequest(ctx context.Context, input CreateRefundRequestInput) (*RefundRequestResponse, error) {
	if input.Amount <= 0 {
		return nil, domain.NewValidationError(map[string]string{"amount": "退款金额必须大于 0"})
	}

	order, _, err := s.reconciliation.GetRechargeOrderWithRecords(ctx, input.OrderID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	// Only the owner may request a refund for their own order.
	if order.UserID != input.UserID {
		return nil, domain.ErrNotFound
	}

	existing, err := s.requestRepo.GetPendingByOrderID(ctx, input.OrderID)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	if existing != nil {
		return nil, domain.NewAppError(409, "PENDING_REFUND_REQUEST_EXISTS", "该订单已有待审批的退款申请")
	}

	req, err := s.requestRepo.Create(ctx, repository.CreateRefundRequestInput{
		OrderID: input.OrderID,
		UserID:  input.UserID,
		Amount:  input.Amount,
		Reason:  input.Reason,
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return toRefundRequestResponse(req), nil
}

// ListMyRefundRequests returns the requests submitted by the given user.
func (s *RefundRequestService) ListMyRefundRequests(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]RefundRequestResponse, int, error) {
	reqs, total, err := s.requestRepo.List(ctx, repository.ListRefundRequestFilter{
		UserID:   &userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}
	resp := make([]RefundRequestResponse, 0, len(reqs))
	for _, r := range reqs {
		resp = append(resp, *toRefundRequestResponse(r))
	}
	return resp, total, nil
}

// CancelRefundRequest lets the owner withdraw a still-pending request.
func (s *RefundRequestService) CancelRefundRequest(ctx context.Context, userID, requestID uuid.UUID) error {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.WrapInternal(err)
	}
	if req.UserID != userID {
		return domain.ErrNotFound
	}
	// Cancellation is modelled as a rejection with no reviewer: the money path
	// is untouched and the state machine stays at three values. The
	// conditional transition loses against a concurrent admin review.
	_, err = s.requestRepo.Transition(ctx, requestID, repository.RefundRequestStatusPending, repository.RefundRequestStatusRejected, nil)
	return mapTransitionError(err)
}

// ListRefundRequests returns a paginated admin list.
func (s *RefundRequestService) ListRefundRequests(ctx context.Context, filter ListRefundRequestFilter) ([]RefundRequestResponse, int, error) {
	var status *repository.RefundRequestStatus
	if filter.Status != nil {
		st := repository.RefundRequestStatus(*filter.Status)
		status = &st
	}
	reqs, total, err := s.requestRepo.List(ctx, repository.ListRefundRequestFilter{
		Status:   status,
		UserID:   filter.UserID,
		OrderID:  filter.OrderID,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	})
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}
	resp := make([]RefundRequestResponse, 0, len(reqs))
	for _, r := range reqs {
		resp = append(resp, *toRefundRequestResponse(r))
	}
	return resp, total, nil
}

// ApproveRefundRequest first wins the pending->approved transition (a
// conditional update, so concurrent approvals settle exactly once) and only
// then runs the money path. If the refund fails after the transition, the
// request is reverted to pending so the approval can be retried.
func (s *RefundRequestService) ApproveRefundRequest(ctx context.Context, requestID, adminID uuid.UUID) (*RefundRequestResponse, error) {
	existing, err := s.getExisting(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if _, err := s.requestRepo.Transition(ctx, requestID, repository.RefundRequestStatusPending, repository.RefundRequestStatusApproved, &adminID); err != nil {
		return nil, mapTransitionError(err)
	}

	if _, _, err := s.reconciliation.RefundOrder(ctx, existing.OrderID, existing.Amount); err != nil {
		// Money did not move; reopen the request so a later approval can retry.
		// A revert failure leaves an approved-but-unrefunded request, which is
		// preferable to silently double-refunding.
		_, _ = s.requestRepo.Transition(ctx, requestID, repository.RefundRequestStatusApproved, repository.RefundRequestStatusPending, nil)
		return nil, domain.AsAppError(err)
	}

	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return toRefundRequestResponse(req), nil
}

// RejectRefundRequest marks a pending request rejected without moving money.
func (s *RefundRequestService) RejectRefundRequest(ctx context.Context, requestID, adminID uuid.UUID) (*RefundRequestResponse, error) {
	if _, err := s.getExisting(ctx, requestID); err != nil {
		return nil, err
	}
	updated, err := s.requestRepo.Transition(ctx, requestID, repository.RefundRequestStatusPending, repository.RefundRequestStatusRejected, &adminID)
	if err != nil {
		return nil, mapTransitionError(err)
	}
	return toRefundRequestResponse(updated), nil
}

// getExisting loads a request, mapping not-found to a 404.
func (s *RefundRequestService) getExisting(ctx context.Context, requestID uuid.UUID) (*ent.RefundRequest, error) {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return req, nil
}

// mapTransitionError converts the repository conflict sentinel into the
// public 409 error; anything else is an internal failure.
func mapTransitionError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrRefundRequestConflict) {
		return domain.NewAppError(409, "REFUND_REQUEST_NOT_PENDING", "该申请已处理")
	}
	return domain.WrapInternal(err)
}

func toRefundRequestResponse(req *ent.RefundRequest) *RefundRequestResponse {
	resp := &RefundRequestResponse{
		ID:        req.ID.String(),
		OrderID:   req.OrderID.String(),
		UserID:    req.UserID.String(),
		Amount:    req.Amount,
		Status:    string(req.Status),
		CreatedAt: req.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: req.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if req.Reason != nil && *req.Reason != "" {
		resp.Reason = req.Reason
	}
	if req.ReviewedBy != nil {
		s := req.ReviewedBy.String()
		resp.ReviewedBy = &s
	}
	if req.ReviewedAt != nil {
		resp.ReviewedAt = req.ReviewedAt
	}
	return resp
}
