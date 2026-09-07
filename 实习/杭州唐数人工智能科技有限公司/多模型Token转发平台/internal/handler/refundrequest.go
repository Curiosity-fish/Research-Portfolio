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

// RefundRequestService is the subset of the refund request service used by the handler.
type RefundRequestService interface {
	CreateRefundRequest(ctx context.Context, input service.CreateRefundRequestInput) (*service.RefundRequestResponse, error)
	ListMyRefundRequests(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.RefundRequestResponse, int, error)
	CancelRefundRequest(ctx context.Context, userID, requestID uuid.UUID) error
	ListRefundRequests(ctx context.Context, filter service.ListRefundRequestFilter) ([]service.RefundRequestResponse, int, error)
	ApproveRefundRequest(ctx context.Context, requestID, adminID uuid.UUID) (*service.RefundRequestResponse, error)
	RejectRefundRequest(ctx context.Context, requestID, adminID uuid.UUID) (*service.RefundRequestResponse, error)
}

// RefundRequestHandler handles the user-initiated refund flow.
type RefundRequestHandler struct {
	svc RefundRequestService
}

// NewRefundRequestHandler creates a new RefundRequestHandler.
func NewRefundRequestHandler(svc RefundRequestService) *RefundRequestHandler {
	return &RefundRequestHandler{svc: svc}
}

// CreateRefundRequestRequest is the user-facing body for opening a request.
type CreateRefundRequestRequest struct {
	OrderID string  `json:"order_id" binding:"required,uuid"`
	Amount  int64   `json:"amount" binding:"required,gt=0"`
	Reason  *string `json:"reason,omitempty" binding:"omitempty,max=500"`
}

// Create handles POST /api/v1/refund-requests (API Key auth).
func (h *RefundRequestHandler) Create(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}
	var req CreateRefundRequestRequest
	if !bindAndValidate(c, &req) {
		return
	}
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"order_id": "订单 ID 格式错误"}))
		return
	}
	resp, err := h.svc.CreateRefundRequest(c.Request.Context(), service.CreateRefundRequestInput{
		UserID:  userID,
		OrderID: orderID,
		Amount:  req.Amount,
		Reason:  req.Reason,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.Created(c, resp)
}

// ListMy handles GET /api/v1/refund-requests (API Key auth).
func (h *RefundRequestHandler) ListMy(c *gin.Context) {
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
	list, total, err := h.svc.ListMyRefundRequests(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}
	page, pageSize := normalizeListPage(q.Page, q.PageSize)
	respond.OK(c, gin.H{"total": total, "page": page, "page_size": pageSize, "list": list})
}

// Cancel handles DELETE /api/v1/refund-requests/:id (API Key auth).
func (h *RefundRequestHandler) Cancel(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "申请 ID 格式错误"}))
		return
	}
	if err := h.svc.CancelRefundRequest(c.Request.Context(), userID, id); err != nil {
		respond.Error(c, err)
		return
	}
	respond.NoContent(c)
}

// ListAdmin handles GET /api/v1/admin/refund-requests.
func (h *RefundRequestHandler) ListAdmin(c *gin.Context) {
	var q struct {
		Status   string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
		UserID   string `form:"user_id"`
		OrderID  string `form:"order_id"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"query": "查询参数无效"}))
		return
	}
	var userID, orderID *uuid.UUID
	if q.UserID != "" {
		id, err := uuid.Parse(q.UserID)
		if err != nil {
			respond.Error(c, domain.NewValidationError(map[string]string{"user_id": "用户 ID 格式错误"}))
			return
		}
		userID = &id
	}
	if q.OrderID != "" {
		id, err := uuid.Parse(q.OrderID)
		if err != nil {
			respond.Error(c, domain.NewValidationError(map[string]string{"order_id": "订单 ID 格式错误"}))
			return
		}
		orderID = &id
	}
	list, total, err := h.svc.ListRefundRequests(c.Request.Context(), service.ListRefundRequestFilter{
		Status:   optionalString(q.Status),
		UserID:   userID,
		OrderID:  orderID,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}
	page, pageSize := normalizeListPage(q.Page, q.PageSize)
	respond.OK(c, gin.H{"total": total, "page": page, "page_size": pageSize, "list": list})
}

// Approve handles POST /api/v1/admin/refund-requests/:id/approve.
func (h *RefundRequestHandler) Approve(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "申请 ID 格式错误"}))
		return
	}
	resp, err := h.svc.ApproveRefundRequest(c.Request.Context(), id, adminID)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// Reject handles POST /api/v1/admin/refund-requests/:id/reject.
func (h *RefundRequestHandler) Reject(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "申请 ID 格式错误"}))
		return
	}
	resp, err := h.svc.RejectRefundRequest(c.Request.Context(), id, adminID)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// optionalString returns nil for an empty filter value.
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// normalizeListPage applies the shared pagination defaults.
func normalizeListPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
