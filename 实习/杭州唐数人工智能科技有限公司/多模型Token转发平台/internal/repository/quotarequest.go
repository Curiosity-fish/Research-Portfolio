package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/quotarecord"
	"github.com/school-api/school-api-v1/ent/quotarequest"
)

// QuotaRequestStatus represents the status of a quota request.
type QuotaRequestStatus string

const (
	QuotaRequestStatusPending  QuotaRequestStatus = "pending"
	QuotaRequestStatusApproved QuotaRequestStatus = "approved"
	QuotaRequestStatusRejected QuotaRequestStatus = "rejected"
)

// CreateQuotaRequestInput is the data required to create a quota request.
type CreateQuotaRequestInput struct {
	UserID          uuid.UUID
	TokenID         uuid.UUID
	RequestedAmount int64
	Reason          *string
}

// UpdateQuotaRequestStatusInput is the data required to approve/reject a request.
type UpdateQuotaRequestStatusInput struct {
	Status     QuotaRequestStatus
	ReviewedBy uuid.UUID
	ReviewedAt time.Time
}

// ListQuotaRequestFilter controls the admin list query.
type ListQuotaRequestFilter struct {
	Status   *QuotaRequestStatus
	Page     int
	PageSize int
}

// QuotaRequestRepository provides data access for quota requests.
type QuotaRequestRepository interface {
	Create(ctx context.Context, input CreateQuotaRequestInput) (*ent.QuotaRequest, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.QuotaRequest, error)
	GetPendingByUserID(ctx context.Context, userID uuid.UUID) (*ent.QuotaRequest, error)
	List(ctx context.Context, filter ListQuotaRequestFilter) ([]*ent.QuotaRequest, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateQuotaRequestStatusInput) (*ent.QuotaRequest, error)
	Approve(ctx context.Context, id, adminID uuid.UUID) (*ent.QuotaRequest, error)
}

// EntQuotaRequestRepository implements QuotaRequestRepository using Ent.
type EntQuotaRequestRepository struct {
	client *ent.Client
}

// NewEntQuotaRequestRepository creates a new Ent-backed quota request repository.
func NewEntQuotaRequestRepository(client *ent.Client) *EntQuotaRequestRepository {
	return &EntQuotaRequestRepository{client: client}
}

// Create inserts a new quota request.
func (r *EntQuotaRequestRepository) Create(ctx context.Context, input CreateQuotaRequestInput) (*ent.QuotaRequest, error) {
	b := r.client.QuotaRequest.Create().
		SetUserID(input.UserID).
		SetTokenID(input.TokenID).
		SetRequestedAmount(input.RequestedAmount)
	if input.Reason != nil {
		b.SetReason(*input.Reason)
	}

	req, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create quota request: %w", err)
	}
	return req, nil
}

// GetByID returns a quota request by ID.
func (r *EntQuotaRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.QuotaRequest, error) {
	req, err := r.client.QuotaRequest.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get quota request: %w", err)
	}
	return req, nil
}

// GetPendingByUserID returns the pending request for a user, if any.
func (r *EntQuotaRequestRepository) GetPendingByUserID(ctx context.Context, userID uuid.UUID) (*ent.QuotaRequest, error) {
	req, err := r.client.QuotaRequest.Query().
		Where(quotarequest.UserIDEQ(userID), quotarequest.StatusEQ(quotarequest.StatusPending)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get pending quota request: %w", err)
	}
	return req, nil
}

// List returns a paginated list of quota requests matching the filter and the total count.
func (r *EntQuotaRequestRepository) List(ctx context.Context, filter ListQuotaRequestFilter) ([]*ent.QuotaRequest, int, error) {
	q := r.client.QuotaRequest.Query()
	if filter.Status != nil {
		q = q.Where(quotarequest.StatusEQ(quotarequest.Status(*filter.Status)))
	}

	page, pageSize := normalizePagination(filter.Page, filter.PageSize)

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count quota requests: %w", err)
	}

	reqs, err := q.Order(ent.Desc(quotarequest.FieldCreatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list quota requests: %w", err)
	}

	return reqs, total, nil
}

// UpdateStatus changes the status of a quota request and records the reviewer.
func (r *EntQuotaRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateQuotaRequestStatusInput) (*ent.QuotaRequest, error) {
	b := r.client.QuotaRequest.UpdateOneID(id).
		SetStatus(quotarequest.Status(input.Status)).
		SetReviewedBy(input.ReviewedBy).
		SetReviewedAt(input.ReviewedAt)

	req, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update quota request status: %w", err)
	}
	return req, nil
}

// Approve marks a pending request as approved, increases the token's quota_limit,
// and writes a quota record. All operations happen in a single transaction.
func (r *EntQuotaRequestRepository) Approve(ctx context.Context, id, adminID uuid.UUID) (*ent.QuotaRequest, error) {
	req, err := r.client.QuotaRequest.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get quota request: %w", err)
	}
	if req.Status != quotarequest.StatusPending {
		return nil, fmt.Errorf("quota request is not pending")
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start transaction: %w", err)
	}

	rollback := func(err error) (*ent.QuotaRequest, error) {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return nil, err
	}

	now := time.Now()
	updated, err := tx.QuotaRequest.UpdateOneID(id).
		SetStatus(quotarequest.StatusApproved).
		SetReviewedBy(adminID).
		SetReviewedAt(now).
		Save(ctx)
	if err != nil {
		return rollback(fmt.Errorf("approve quota request: %w", err))
	}

	if _, err := tx.UserToken.UpdateOneID(req.TokenID).AddQuotaLimit(req.RequestedAmount).Save(ctx); err != nil {
		return rollback(fmt.Errorf("increase token quota: %w", err))
	}

	tok, err := tx.UserToken.Get(ctx, req.TokenID)
	if err != nil {
		return rollback(fmt.Errorf("get updated token: %w", err))
	}

	if _, err := tx.QuotaRecord.Create().
		SetUserID(req.UserID).
		SetTokenID(req.TokenID).
		SetType(quotarecord.TypeRequestApproved).
		SetAmount(req.RequestedAmount).
		SetQuotaAfter(tok.QuotaUsed).
		Save(ctx); err != nil {
		return rollback(fmt.Errorf("create approval quota record: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit quota request approval: %w", err)
	}
	return updated, nil
}

var _ QuotaRequestRepository = (*EntQuotaRequestRepository)(nil)
