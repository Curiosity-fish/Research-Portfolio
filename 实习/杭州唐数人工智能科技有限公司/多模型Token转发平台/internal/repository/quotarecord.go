package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/quotarecord"
)

// QuotaRecordRepository provides data access for quota records.
type QuotaRecordRepository interface {
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.QuotaRecord, int, error)
	ListByTokenID(ctx context.Context, tokenID uuid.UUID, offset, limit int) ([]*ent.QuotaRecord, int, error)
	List(ctx context.Context, userID *uuid.UUID, offset, limit int) ([]*ent.QuotaRecord, int, error)
}

// EntQuotaRecordRepository implements QuotaRecordRepository using Ent.
type EntQuotaRecordRepository struct {
	client *ent.Client
}

// NewEntQuotaRecordRepository creates a new Ent-backed quota record repository.
func NewEntQuotaRecordRepository(client *ent.Client) *EntQuotaRecordRepository {
	return &EntQuotaRecordRepository{client: client}
}

// ListByUserID returns a paginated list of quota records for a user.
func (r *EntQuotaRecordRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.QuotaRecord, int, error) {
	q := r.client.QuotaRecord.Query().Where(quotarecord.UserIDEQ(userID))

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count quota records: %w", err)
	}

	records, err := q.Order(ent.Desc(quotarecord.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list quota records: %w", err)
	}

	return records, total, nil
}

// ListByTokenID returns a paginated list of quota records for a token.
func (r *EntQuotaRecordRepository) ListByTokenID(ctx context.Context, tokenID uuid.UUID, offset, limit int) ([]*ent.QuotaRecord, int, error) {
	q := r.client.QuotaRecord.Query().Where(quotarecord.TokenIDEQ(tokenID))

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count quota records by token: %w", err)
	}

	records, err := q.Order(ent.Desc(quotarecord.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list quota records by token: %w", err)
	}

	return records, total, nil
}

// List returns a paginated list of quota records, optionally filtered by user.
func (r *EntQuotaRecordRepository) List(ctx context.Context, userID *uuid.UUID, offset, limit int) ([]*ent.QuotaRecord, int, error) {
	q := r.client.QuotaRecord.Query()
	if userID != nil {
		q = q.Where(quotarecord.UserIDEQ(*userID))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count quota records: %w", err)
	}

	records, err := q.Order(ent.Desc(quotarecord.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list quota records: %w", err)
	}

	return records, total, nil
}

var _ QuotaRecordRepository = (*EntQuotaRecordRepository)(nil)
