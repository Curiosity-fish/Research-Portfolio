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

// BalanceRecordService is the subset of the balance record service used by BalanceRecordHandler.
type BalanceRecordService interface {
	ListBalanceRecords(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.BalanceRecordResponse, int, error)
	ListBalanceRecordsAdmin(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]service.BalanceRecordResponse, int, error)
	AdminAdjustBalance(ctx context.Context, userID, adminID uuid.UUID, amount int64, remark *string) (*service.BalanceRecordResponse, error)
}

// AdminAdjustBalanceRequest is the request body for admin balance adjustment.
type AdminAdjustBalanceRequest struct {
	Amount int64   `json:"amount" binding:"required"`
	Remark *string `json:"remark" binding:"omitempty,max=500"`
}

// BalanceRecordHandler handles balance record endpoints.
type BalanceRecordHandler struct {
	svc BalanceRecordService
}

// NewBalanceRecordHandler creates a new BalanceRecordHandler.
func NewBalanceRecordHandler(svc BalanceRecordService) *BalanceRecordHandler {
	return &BalanceRecordHandler{svc: svc}
}

// ListMy handles GET /api/v1/balance-records.
func (h *BalanceRecordHandler) ListMy(c *gin.Context) {
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

	records, total, err := h.svc.ListBalanceRecords(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": records})
}

// AdminAdjust handles POST /api/v1/admin/users/:id/balance-adjust.
func (h *BalanceRecordHandler) AdminAdjust(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "用户 ID 格式错误"}))
		return
	}

	var req AdminAdjustBalanceRequest
	if !bindAndValidate(c, &req) {
		return
	}

	resp, err := h.svc.AdminAdjustBalance(c.Request.Context(), userID, adminID, req.Amount, req.Remark)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// ListAdmin handles GET /api/v1/admin/balance-records.
func (h *BalanceRecordHandler) ListAdmin(c *gin.Context) {
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

	records, total, err := h.svc.ListBalanceRecordsAdmin(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": records})
}
