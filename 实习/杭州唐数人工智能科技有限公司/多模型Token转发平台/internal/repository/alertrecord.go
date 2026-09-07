package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/alertrecord"
)

// CreateAlertRecordInput is the data required to create an alert record.
type CreateAlertRecordInput struct {
	RuleID         *uuid.UUID
	Metric         string
	UserID         *uuid.UUID
	TokenID        *uuid.UUID
	AccountID      *uuid.UUID
	TriggeredValue int64
	Message        string
}

// ListAlertRecordFilter controls the alert record list query.
type ListAlertRecordFilter struct {
	RuleID     *uuid.UUID
	IsResolved *bool
	Offset     int
	Limit      int
}

// ResolveAlertRecordInput is the data required to resolve an alert.
type ResolveAlertRecordInput struct {
	ResolvedBy uuid.UUID
	ResolvedAt time.Time
}

// AlertRecordRepository provides data access for alert records.
type AlertRecordRepository interface {
	Create(ctx context.Context, input CreateAlertRecordInput) (*ent.AlertRecord, error)
	CreateBatch(ctx context.Context, inputs []CreateAlertRecordInput) (int, error)
	ExistsUnresolved(ctx context.Context, ruleID uuid.UUID, userID, tokenID, accountID *uuid.UUID) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.AlertRecord, error)
	List(ctx context.Context, filter ListAlertRecordFilter) ([]*ent.AlertRecord, int, error)
	Resolve(ctx context.Context, id uuid.UUID, input ResolveAlertRecordInput) (*ent.AlertRecord, error)
}

// EntAlertRecordRepository implements AlertRecordRepository using Ent.
type EntAlertRecordRepository struct {
	client *ent.Client
}

// NewEntAlertRecordRepository creates a new Ent-backed alert record repository.
func NewEntAlertRecordRepository(client *ent.Client) *EntAlertRecordRepository {
	return &EntAlertRecordRepository{client: client}
}

// Create inserts a new alert record.
func (r *EntAlertRecordRepository) Create(ctx context.Context, input CreateAlertRecordInput) (*ent.AlertRecord, error) {
	b := r.client.AlertRecord.Create().
		SetMetric(input.Metric).
		SetTriggeredValue(input.TriggeredValue).
		SetMessage(input.Message)
	if input.RuleID != nil {
		b.SetRuleID(*input.RuleID)
	}
	if input.UserID != nil {
		b.SetUserID(*input.UserID)
	}
	if input.TokenID != nil {
		b.SetTokenID(*input.TokenID)
	}
	if input.AccountID != nil {
		b.SetAccountID(*input.AccountID)
	}

	ar, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create alert record: %w", err)
	}
	return ar, nil
}

// CreateBatch inserts alert records in a single transaction so a failure
// halfway through does not leave partially written records for one rule.
func (r *EntAlertRecordRepository) CreateBatch(ctx context.Context, inputs []CreateAlertRecordInput) (int, error) {
	if len(inputs) == 0 {
		return 0, nil
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, fmt.Errorf("start alert record batch transaction: %w", err)
	}
	rollback := func(err error) (int, error) {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return 0, err
	}

	builders := make([]*ent.AlertRecordCreate, len(inputs))
	for i, input := range inputs {
		b := tx.AlertRecord.Create().
			SetMetric(input.Metric).
			SetTriggeredValue(input.TriggeredValue).
			SetMessage(input.Message)
		if input.RuleID != nil {
			b.SetRuleID(*input.RuleID)
		}
		if input.UserID != nil {
			b.SetUserID(*input.UserID)
		}
		if input.TokenID != nil {
			b.SetTokenID(*input.TokenID)
		}
		if input.AccountID != nil {
			b.SetAccountID(*input.AccountID)
		}
		builders[i] = b
	}

	if err := tx.AlertRecord.CreateBulk(builders...).Exec(ctx); err != nil {
		return rollback(fmt.Errorf("create alert records: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit alert record batch: %w", err)
	}
	return len(inputs), nil
}

// ExistsUnresolved reports whether an unresolved record already exists for the
// same rule and target, so evaluation does not stack duplicate alerts.
func (r *EntAlertRecordRepository) ExistsUnresolved(ctx context.Context, ruleID uuid.UUID, userID, tokenID, accountID *uuid.UUID) (bool, error) {
	q := r.client.AlertRecord.Query().
		Where(
			alertrecord.RuleIDEQ(ruleID),
			alertrecord.IsResolvedEQ(false),
		)
	if userID == nil {
		q = q.Where(alertrecord.UserIDIsNil())
	} else {
		q = q.Where(alertrecord.UserIDEQ(*userID))
	}
	if tokenID == nil {
		q = q.Where(alertrecord.TokenIDIsNil())
	} else {
		q = q.Where(alertrecord.TokenIDEQ(*tokenID))
	}
	if accountID == nil {
		q = q.Where(alertrecord.AccountIDIsNil())
	} else {
		q = q.Where(alertrecord.AccountIDEQ(*accountID))
	}

	exists, err := q.Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("check unresolved alert record: %w", err)
	}
	return exists, nil
}

// GetByID returns an alert record by ID.
func (r *EntAlertRecordRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.AlertRecord, error) {
	ar, err := r.client.AlertRecord.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get alert record: %w", err)
	}
	return ar, nil
}

// List returns a paginated list of alert records.
func (r *EntAlertRecordRepository) List(ctx context.Context, filter ListAlertRecordFilter) ([]*ent.AlertRecord, int, error) {
	q := r.client.AlertRecord.Query()
	if filter.RuleID != nil {
		q = q.Where(alertrecord.RuleIDEQ(*filter.RuleID))
	}
	if filter.IsResolved != nil {
		q = q.Where(alertrecord.IsResolvedEQ(*filter.IsResolved))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count alert records: %w", err)
	}

	list, err := q.Order(ent.Desc(alertrecord.FieldCreatedAt)).
		Offset(filter.Offset).
		Limit(filter.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list alert records: %w", err)
	}

	return list, total, nil
}

// Resolve marks an alert record as resolved.
func (r *EntAlertRecordRepository) Resolve(ctx context.Context, id uuid.UUID, input ResolveAlertRecordInput) (*ent.AlertRecord, error) {
	ar, err := r.client.AlertRecord.UpdateOneID(id).
		SetIsResolved(true).
		SetResolvedAt(input.ResolvedAt).
		SetResolvedBy(input.ResolvedBy).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve alert record: %w", err)
	}
	return ar, nil
}

var _ AlertRecordRepository = (*EntAlertRecordRepository)(nil)
