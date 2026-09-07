package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/calllog"
)

// CreateCallLogInput is the data required to create a call log record.
type CreateCallLogInput struct {
	UserID           uuid.UUID
	TokenID          uuid.UUID
	PlatformID       uuid.UUID
	AccountID        uuid.UUID
	Model            string
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
	LatencyMs        int64
	StatusCode       int
	ErrorMsg         *string
}

// CallLogRepository provides data access for relay call logs.
type CallLogRepository interface {
	Create(ctx context.Context, input CreateCallLogInput) (*ent.CallLog, error)
	// DeleteOlderThan removes every log created before cutoff and returns the
	// number of deleted rows. created_at is indexed, so the delete does not
	// require a full table scan.
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int, error)
}

// EntCallLogRepository implements CallLogRepository using Ent.
type EntCallLogRepository struct {
	client *ent.Client
}

// NewEntCallLogRepository creates a new Ent-backed call log repository.
func NewEntCallLogRepository(client *ent.Client) *EntCallLogRepository {
	return &EntCallLogRepository{client: client}
}

// Create inserts a new call log record.
func (r *EntCallLogRepository) Create(ctx context.Context, input CreateCallLogInput) (*ent.CallLog, error) {
	b := r.client.CallLog.Create().
		SetUserID(input.UserID).
		SetTokenID(input.TokenID).
		SetPlatformID(input.PlatformID).
		SetAccountID(input.AccountID).
		SetModel(input.Model).
		SetPromptTokens(input.PromptTokens).
		SetCompletionTokens(input.CompletionTokens).
		SetTotalTokens(input.TotalTokens).
		SetLatencyMs(input.LatencyMs).
		SetStatusCode(input.StatusCode)

	if input.ErrorMsg != nil {
		b.SetErrorMsg(*input.ErrorMsg)
	}

	log, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create call log: %w", err)
	}
	return log, nil
}

// DeleteOlderThan deletes all call logs whose created_at is strictly before
// the cutoff and reports how many rows were removed.
func (r *EntCallLogRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int, error) {
	n, err := r.client.CallLog.Delete().
		Where(calllog.CreatedAtLT(cutoff)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete call logs older than %s: %w", cutoff, err)
	}
	return n, nil
}
