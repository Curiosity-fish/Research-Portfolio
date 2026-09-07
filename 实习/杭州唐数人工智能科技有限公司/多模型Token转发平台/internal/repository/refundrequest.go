package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/predicate"
	"github.com/school-api/school-api-v1/ent/refundrequest"
)

// RefundRequestStatus mirrors the ent enum at the repository boundary.
type RefundRequestStatus string

const (
	RefundRequestStatusPending  RefundRequestStatus = "pending"
	RefundRequestStatusApproved RefundRequestStatus = "approved"
	RefundRequestStatusRejected RefundRequestStatus = "rejected"
)

// ErrRefundRequestConflict is returned when a state transition finds the
// request is no longer in the expected state (a concurrent review settled it).
var ErrRefundRequestConflict = errors.New("refund request state conflict")

// CreateRefundRequestInput is the data required to open a refund request.
type CreateRefundRequestInput struct {
	OrderID uuid.UUID
	UserID  uuid.UUID
	Amount  int64
	Reason  *string
}

// ListRefundRequestFilter controls refund request list queries.
type ListRefundRequestFilter struct {
	Status   *RefundRequestStatus
	UserID   *uuid.UUID
	OrderID  *uuid.UUID
	Page     int
	PageSize int
}

// RefundRequestRepository provides data access for refund requests.
type RefundRequestRepository interface {
	Create(ctx context.Context, input CreateRefundRequestInput) (*ent.RefundRequest, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.RefundRequest, error)
	// GetPendingByOrderID returns the single pending request for an order, or
	// nil when none exists (an order may have at most one open request).
	GetPendingByOrderID(ctx context.Context, orderID uuid.UUID) (*ent.RefundRequest, error)
	List(ctx context.Context, filter ListRefundRequestFilter) ([]*ent.RefundRequest, int, error)
	// Transition moves a request from the expected state to the target state
	// using a conditional update, so a concurrent transition loses instead of
	// overwriting. It returns ErrRefundRequestConflict when the request is not
	// in the expected state. A nil reviewedBy clears the reviewer fields
	// (user-initiated cancellation and reverting a failed approval); a non-nil
	// reviewedBy records the reviewer.
	Transition(ctx context.Context, id uuid.UUID, expect, status RefundRequestStatus, reviewedBy *uuid.UUID) (*ent.RefundRequest, error)
}

// EntRefundRequestRepository implements RefundRequestRepository using Ent.
type EntRefundRequestRepository struct {
	client *ent.Client
}

// NewEntRefundRequestRepository creates a new Ent-backed refund request repository.
func NewEntRefundRequestRepository(client *ent.Client) *EntRefundRequestRepository {
	return &EntRefundRequestRepository{client: client}
}

// Create inserts a new refund request in pending state.
func (r *EntRefundRequestRepository) Create(ctx context.Context, input CreateRefundRequestInput) (*ent.RefundRequest, error) {
	b := r.client.RefundRequest.Create().
		SetOrderID(input.OrderID).
		SetUserID(input.UserID).
		SetAmount(input.Amount).
		SetStatus(refundrequest.StatusPending)
	if input.Reason != nil {
		b.SetReason(*input.Reason)
	}
	req, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create refund request: %w", err)
	}
	return req, nil
}

// GetByID looks up a refund request by ID.
func (r *EntRefundRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.RefundRequest, error) {
	req, err := r.client.RefundRequest.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get refund request: %w", err)
	}
	return req, nil
}

// GetPendingByOrderID returns the open request for an order, if any.
func (r *EntRefundRequestRepository) GetPendingByOrderID(ctx context.Context, orderID uuid.UUID) (*ent.RefundRequest, error) {
	req, err := r.client.RefundRequest.Query().
		Where(
			refundrequest.OrderIDEQ(orderID),
			refundrequest.StatusEQ(refundrequest.StatusPending),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get pending refund request: %w", err)
	}
	return req, nil
}

// List returns a paginated list of refund requests and the total count.
func (r *EntRefundRequestRepository) List(ctx context.Context, filter ListRefundRequestFilter) ([]*ent.RefundRequest, int, error) {
	var preds []predicate.RefundRequest
	if filter.Status != nil {
		preds = append(preds, refundrequest.StatusEQ(refundrequest.Status(*filter.Status)))
	}
	if filter.UserID != nil {
		preds = append(preds, refundrequest.UserIDEQ(*filter.UserID))
	}
	if filter.OrderID != nil {
		preds = append(preds, refundrequest.OrderIDEQ(*filter.OrderID))
	}

	page, pageSize := normalizePagination(filter.Page, filter.PageSize)
	reqs, err := r.client.RefundRequest.Query().
		Where(preds...).
		Order(ent.Desc(refundrequest.FieldCreatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list refund requests: %w", err)
	}
	total, err := r.client.RefundRequest.Query().Where(preds...).Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count refund requests: %w", err)
	}
	return reqs, total, nil
}

// Transition applies a conditional state change; see the interface contract.
func (r *EntRefundRequestRepository) Transition(ctx context.Context, id uuid.UUID, expect, status RefundRequestStatus, reviewedBy *uuid.UUID) (*ent.RefundRequest, error) {
	updater := r.client.RefundRequest.Update().
		Where(
			refundrequest.IDEQ(id),
			refundrequest.StatusEQ(refundrequest.Status(expect)),
		).
		SetStatus(refundrequest.Status(status))
	if reviewedBy != nil {
		updater.SetReviewedBy(*reviewedBy).SetReviewedAt(time.Now())
	} else {
		updater.ClearReviewedBy().ClearReviewedAt()
	}
	affected, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("transition refund request: %w", err)
	}
	if affected == 0 {
		return nil, ErrRefundRequestConflict
	}
	req, err := r.client.RefundRequest.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get transitioned refund request: %w", err)
	}
	return req, nil
}
