package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// BalanceRecordResponse is the public representation of a balance record.
type BalanceRecordResponse struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Type            string     `json:"type"`
	Amount          int64      `json:"amount"`
	BalanceAfter    int64      `json:"balance_after"`
	CallLogID       *string    `json:"call_log_id,omitempty"`
	RelatedOrderID  *string    `json:"related_order_id,omitempty"`
	Remark          *string    `json:"remark,omitempty"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

// BalanceRecordService handles balance record business logic.
type BalanceRecordService struct {
	repo    repository.BalanceRecordRepository
	billing repository.BillingRepository
}

// NewBalanceRecordService creates a new BalanceRecordService.
func NewBalanceRecordService(repo repository.BalanceRecordRepository, billing repository.BillingRepository) *BalanceRecordService {
	return &BalanceRecordService{repo: repo, billing: billing}
}

// ListBalanceRecords returns a paginated list of balance records for a user.
func (s *BalanceRecordService) ListBalanceRecords(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]BalanceRecordResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	records, total, err := s.repo.ListByUserID(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]BalanceRecordResponse, 0, len(records))
	for _, rec := range records {
		resp = append(resp, *toBalanceRecordResponse(rec))
	}
	return resp, total, nil
}

// ListBalanceRecordsAdmin returns a paginated list of balance records for admin, optionally filtered by user.
func (s *BalanceRecordService) ListBalanceRecordsAdmin(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]BalanceRecordResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	records, total, err := s.repo.List(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]BalanceRecordResponse, 0, len(records))
	for _, rec := range records {
		resp = append(resp, *toBalanceRecordResponse(rec))
	}
	return resp, total, nil
}

// AdminAdjustBalance adjusts a user's balance by a signed amount and writes a balance record.
func (s *BalanceRecordService) AdminAdjustBalance(ctx context.Context, userID, adminID uuid.UUID, amount int64, remark *string) (*BalanceRecordResponse, error) {
	if amount == 0 {
		return nil, domain.NewAppError(400, "INVALID_REQUEST", "调整金额不能为 0")
	}

	u, err := s.billing.AdminAdjustBalance(ctx, userID, adminID, amount, remark)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	return &BalanceRecordResponse{
		UserID:       u.ID.String(),
		Type:         "admin_adjust",
		Amount:       amount,
		BalanceAfter: u.Balance,
		CreatedAt:    u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func toBalanceRecordResponse(rec *ent.BalanceRecord) *BalanceRecordResponse {
	resp := &BalanceRecordResponse{
		ID:           rec.ID.String(),
		UserID:       rec.UserID.String(),
		Type:         string(rec.Type),
		Amount:       rec.Amount,
		BalanceAfter: rec.BalanceAfter,
		CreatedAt:    rec.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    rec.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if rec.CallLogID != nil {
		s := rec.CallLogID.String()
		resp.CallLogID = &s
	}
	if rec.RelatedOrderID != nil {
		s := rec.RelatedOrderID.String()
		resp.RelatedOrderID = &s
	}
	if rec.Remark != nil && *rec.Remark != "" {
		resp.Remark = rec.Remark
	}
	return resp
}
