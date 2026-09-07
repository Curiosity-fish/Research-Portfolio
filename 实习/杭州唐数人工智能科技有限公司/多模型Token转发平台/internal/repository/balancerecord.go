package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/balancerecord"
)

// CreateBalanceRecordInput is the data required to create a balance record.
type CreateBalanceRecordInput struct {
	UserID         uuid.UUID
	CallLogID      *uuid.UUID
	RelatedOrderID *uuid.UUID
	Type           balancerecord.Type
	Amount         int64
	BalanceAfter   int64
	Remark         *string
}

// BalanceRecordRepository provides data access for balance records.
type BalanceRecordRepository interface {
	Create(ctx context.Context, input CreateBalanceRecordInput) (*ent.BalanceRecord, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.BalanceRecord, int, error)
	List(ctx context.Context, userID *uuid.UUID, offset, limit int) ([]*ent.BalanceRecord, int, error)
	// ListRecentWithUser returns the most recent records (newest first) with
	// the owning user eager-loaded, for XLSX export. A nil userID exports the
	// whole platform. Callers enforce the export row cap via limit.
	ListRecentWithUser(ctx context.Context, userID *uuid.UUID, limit int) ([]*ent.BalanceRecord, error)
}

// EntBalanceRecordRepository implements BalanceRecordRepository using Ent.
type EntBalanceRecordRepository struct {
	client *ent.Client
}

// NewEntBalanceRecordRepository creates a new Ent-backed balance record repository.
func NewEntBalanceRecordRepository(client *ent.Client) *EntBalanceRecordRepository {
	return &EntBalanceRecordRepository{client: client}
}

// Create inserts a new balance record.
func (r *EntBalanceRecordRepository) Create(ctx context.Context, input CreateBalanceRecordInput) (*ent.BalanceRecord, error) {
	b := r.client.BalanceRecord.Create().
		SetUserID(input.UserID).
		SetType(input.Type).
		SetAmount(input.Amount).
		SetBalanceAfter(input.BalanceAfter)
	if input.CallLogID != nil {
		b.SetCallLogID(*input.CallLogID)
	}
	if input.RelatedOrderID != nil {
		b.SetRelatedOrderID(*input.RelatedOrderID)
	}
	if input.Remark != nil {
		b.SetRemark(*input.Remark)
	}

	rec, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create balance record: %w", err)
	}
	return rec, nil
}

// ListByUserID returns a paginated list of balance records for a user.
func (r *EntBalanceRecordRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.BalanceRecord, int, error) {
	q := r.client.BalanceRecord.Query().Where(balancerecord.UserIDEQ(userID))

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count balance records: %w", err)
	}

	records, err := q.Order(ent.Desc(balancerecord.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list balance records: %w", err)
	}

	return records, total, nil
}

// List returns a paginated list of balance records, optionally filtered by user.
func (r *EntBalanceRecordRepository) List(ctx context.Context, userID *uuid.UUID, offset, limit int) ([]*ent.BalanceRecord, int, error) {
	q := r.client.BalanceRecord.Query()
	if userID != nil {
		q = q.Where(balancerecord.UserIDEQ(*userID))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count balance records: %w", err)
	}

	records, err := q.Order(ent.Desc(balancerecord.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list balance records: %w", err)
	}

	return records, total, nil
}

// ListRecentWithUser returns recent records with the owning user eager-loaded.
func (r *EntBalanceRecordRepository) ListRecentWithUser(ctx context.Context, userID *uuid.UUID, limit int) ([]*ent.BalanceRecord, error) {
	q := r.client.BalanceRecord.Query().WithUser()
	if userID != nil {
		q = q.Where(balancerecord.UserIDEQ(*userID))
	}

	records, err := q.Order(ent.Desc(balancerecord.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list balance records for export: %w", err)
	}
	return records, nil
}

var _ BalanceRecordRepository = (*EntBalanceRecordRepository)(nil)
