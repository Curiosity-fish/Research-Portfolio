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

// ReconciliationService is the subset of the reconciliation service used by ReconciliationHandler.
type ReconciliationService interface {
	GetOrderReconciliation(ctx context.Context, orderID uuid.UUID) (*service.ReconciliationOrderResponse, error)
	RefundOrder(ctx context.Context, adminID, orderID uuid.UUID, amount int64, reason *string) (*service.RefundOrderResponse, error)
}

// RefundOrderRequest is the request body for refunding an order.
type RefundOrderRequest struct {
	Amount int64  `json:"amount" binding:"required,gt=0"`
	Reason string `json:"reason,omitempty" binding:"omitempty,max=500"`
}

// ReconciliationHandler handles order reconciliation and refund endpoints.
type ReconciliationHandler struct {
	svc ReconciliationService
}

// NewReconciliationHandler creates a new ReconciliationHandler.
func NewReconciliationHandler(svc ReconciliationService) *ReconciliationHandler {
	return &ReconciliationHandler{svc: svc}
}

// GetReconciliation handles GET /api/v1/admin/recharge-orders/:id/reconciliation.
func (h *ReconciliationHandler) GetReconciliation(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "订单 ID 格式错误"}))
		return
	}

	resp, err := h.svc.GetOrderReconciliation(c.Request.Context(), orderID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// Refund handles POST /api/v1/admin/recharge-orders/:id/refunds.
func (h *ReconciliationHandler) Refund(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "订单 ID 格式错误"}))
		return
	}

	var req RefundOrderRequest
	if !bindAndValidate(c, &req) {
		return
	}

	var reason *string
	if req.Reason != "" {
		reason = &req.Reason
	}

	resp, err := h.svc.RefundOrder(c.Request.Context(), adminID, orderID, req.Amount, reason)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}
