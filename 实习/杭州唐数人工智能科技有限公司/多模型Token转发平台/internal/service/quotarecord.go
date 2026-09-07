package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// QuotaRecordResponse is the public representation of a quota record.
type QuotaRecordResponse struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	TokenID    string     `json:"token_id"`
	CallLogID  *string    `json:"call_log_id,omitempty"`
	Type       string     `json:"type"`
	Amount     int64      `json:"amount"`
	QuotaAfter int64      `json:"quota_after"`
	CreatedAt  string     `json:"created_at"`
	UpdatedAt  string     `json:"updated_at"`
}

// QuotaRecordService handles quota record business logic.
type QuotaRecordService struct {
	repo repository.QuotaRecordRepository
}

// NewQuotaRecordService creates a new QuotaRecordService.
func NewQuotaRecordService(repo repository.QuotaRecordRepository) *QuotaRecordService {
	return &QuotaRecordService{repo: repo}
}

// ListQuotaRecords returns a paginated list of quota records for a user.
func (s *QuotaRecordService) ListQuotaRecords(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]QuotaRecordResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	records, total, err := s.repo.ListByUserID(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]QuotaRecordResponse, 0, len(records))
	for _, rec := range records {
		resp = append(resp, *toQuotaRecordResponse(rec))
	}
	return resp, total, nil
}

// ListQuotaRecordsAdmin returns a paginated list of quota records for admin, optionally filtered by user.
func (s *QuotaRecordService) ListQuotaRecordsAdmin(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]QuotaRecordResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	records, total, err := s.repo.List(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]QuotaRecordResponse, 0, len(records))
	for _, rec := range records {
		resp = append(resp, *toQuotaRecordResponse(rec))
	}
	return resp, total, nil
}

func toQuotaRecordResponse(rec *ent.QuotaRecord) *QuotaRecordResponse {
	resp := &QuotaRecordResponse{
		ID:         rec.ID.String(),
		UserID:     rec.UserID.String(),
		TokenID:    rec.TokenID.String(),
		Type:       string(rec.Type),
		Amount:     rec.Amount,
		QuotaAfter: rec.QuotaAfter,
		CreatedAt:  rec.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  rec.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if rec.CallLogID != nil {
		s := rec.CallLogID.String()
		resp.CallLogID = &s
	}
	return resp
}
