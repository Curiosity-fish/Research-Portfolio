package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// QuotaRecordService is the subset of the quota record service used by QuotaRecordHandler.
type QuotaRecordService interface {
	ListQuotaRecords(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.QuotaRecordResponse, int, error)
	ListQuotaRecordsAdmin(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]service.QuotaRecordResponse, int, error)
}

// QuotaRecordHandler handles quota record endpoints.
type QuotaRecordHandler struct {
	svc QuotaRecordService
}

// NewQuotaRecordHandler creates a new QuotaRecordHandler.
func NewQuotaRecordHandler(svc QuotaRecordService) *QuotaRecordHandler {
	return &QuotaRecordHandler{svc: svc}
}

// ListMy handles GET /api/v1/quota-records.
func (h *QuotaRecordHandler) ListMy(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	var q struct {
		Page     int `form:"page"`
		PageSize int `form:"page_size"`
	}
	_ = c.ShouldBindQuery(&q)

	records, total, err := h.svc.ListQuotaRecords(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": records})
}

// ListAdmin handles GET /api/v1/admin/quota-records.
func (h *QuotaRecordHandler) ListAdmin(c *gin.Context) {
	var q struct {
		UserID   string `form:"user_id"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	_ = c.ShouldBindQuery(&q)

	var userID *uuid.UUID
	if q.UserID != "" {
		id, err := uuid.Parse(q.UserID)
		if err != nil {
			respond.Error(c, domain.NewValidationError(map[string]string{"user_id": "用户 ID 格式错误"}))
			return
		}
		userID = &id
	}

	records, total, err := h.svc.ListQuotaRecordsAdmin(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": records})
}
