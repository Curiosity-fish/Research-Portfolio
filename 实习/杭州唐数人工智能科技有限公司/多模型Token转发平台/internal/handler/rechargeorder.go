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

// RechargeOrderService is the subset of the recharge order service used by RechargeOrderHandler.
type RechargeOrderService interface {
	CreateRechargeOrder(ctx context.Context, userID uuid.UUID, amount int64) (*service.RechargeOrderResponse, error)
	GetRechargeOrder(ctx context.Context, userID, orderID uuid.UUID) (*service.RechargeOrderResponse, error)
	ListRechargeOrders(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.RechargeOrderResponse, int, error)
	ListRechargeOrdersAdmin(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]service.RechargeOrderResponse, int, error)
	ProcessMockCallback(ctx context.Context, userID, orderID uuid.UUID, success bool) (*service.RechargeOrderResponse, error)
}

// CreateRechargeOrderRequest is the request body for creating a recharge order.
type CreateRechargeOrderRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

// MockCallbackRequest is the request body for simulating a payment callback.
// Success is a pointer so that an explicit "success": false passes validation
// (bool's zero value would be rejected by `required`).
type MockCallbackRequest struct {
	Success *bool `json:"success" binding:"required"`
}

// RechargeOrderHandler handles recharge order endpoints.
type RechargeOrderHandler struct {
	svc RechargeOrderService
}

// NewRechargeOrderHandler creates a new RechargeOrderHandler.
func NewRechargeOrderHandler(svc RechargeOrderService) *RechargeOrderHandler {
	return &RechargeOrderHandler{svc: svc}
}

// Create handles POST /api/v1/recharge/orders.
func (h *RechargeOrderHandler) Create(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	var req CreateRechargeOrderRequest
	if !bindAndValidate(c, &req) {
		return
	}

	resp, err := h.svc.CreateRechargeOrder(c.Request.Context(), userID, req.Amount)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, resp)
}

// Get handles GET /api/v1/recharge/orders/:id.
func (h *RechargeOrderHandler) Get(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "订单 ID 格式错误"}))
		return
	}

	resp, err := h.svc.GetRechargeOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// List handles GET /api/v1/recharge/orders.
func (h *RechargeOrderHandler) List(c *gin.Context) {
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

	orders, total, err := h.svc.ListRechargeOrders(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": orders})
}

// MockCallback handles POST /api/v1/recharge/orders/:id/mock-callback.
func (h *RechargeOrderHandler) MockCallback(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "订单 ID 格式错误"}))
		return
	}

	var req MockCallbackRequest
	if !bindAndValidate(c, &req) {
		return
	}

	resp, err := h.svc.ProcessMockCallback(c.Request.Context(), userID, orderID, *req.Success)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// ListAdmin handles GET /api/v1/admin/recharge-orders.
func (h *RechargeOrderHandler) ListAdmin(c *gin.Context) {
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

	orders, total, err := h.svc.ListRechargeOrdersAdmin(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": orders})
}
