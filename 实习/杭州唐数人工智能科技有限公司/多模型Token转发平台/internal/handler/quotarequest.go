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

// QuotaRequestService is the subset of the quota request service used by QuotaRequestHandler.
type QuotaRequestService interface {
	CreateQuotaRequest(ctx context.Context, input service.CreateQuotaRequestInput) (*service.QuotaRequestResponse, error)
	ListMyQuotaRequests(ctx context.Context, userID uuid.UUID) ([]service.QuotaRequestResponse, error)
	ListQuotaRequests(ctx context.Context, filter service.ListQuotaRequestFilter) ([]service.QuotaRequestResponse, int, error)
	ApproveQuotaRequest(ctx context.Context, requestID, adminID uuid.UUID) (*service.QuotaRequestResponse, error)
	RejectQuotaRequest(ctx context.Context, requestID, adminID uuid.UUID) (*service.QuotaRequestResponse, error)
}

// CreateQuotaRequestRequest is the request body for submitting a quota request.
type CreateQuotaRequestRequest struct {
	TokenID         string  `json:"token_id" binding:"required,uuid"`
	RequestedAmount int64   `json:"requested_amount" binding:"required,gt=0"`
	Reason          *string `json:"reason" binding:"omitempty,max=500"`
}

// QuotaRequestHandler handles quota request endpoints.
type QuotaRequestHandler struct {
	svc QuotaRequestService
}

// NewQuotaRequestHandler creates a new QuotaRequestHandler.
func NewQuotaRequestHandler(svc QuotaRequestService) *QuotaRequestHandler {
	return &QuotaRequestHandler{svc: svc}
}

// Create handles POST /api/v1/quota-requests.
func (h *QuotaRequestHandler) Create(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	var req CreateQuotaRequestRequest
	if !bindAndValidate(c, &req) {
		return
	}

	tokenID, err := uuid.Parse(req.TokenID)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"token_id": "Token ID 格式错误"}))
		return
	}

	resp, err := h.svc.CreateQuotaRequest(c.Request.Context(), service.CreateQuotaRequestInput{
		UserID:          userID,
		TokenID:         tokenID,
		RequestedAmount: req.RequestedAmount,
		Reason:          req.Reason,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, resp)
}

// ListMy handles GET /api/v1/quota-requests.
func (h *QuotaRequestHandler) ListMy(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	reqs, err := h.svc.ListMyQuotaRequests(c.Request.Context(), userID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"list": reqs})
}

// ListAdmin handles GET /api/v1/admin/quota-requests.
func (h *QuotaRequestHandler) ListAdmin(c *gin.Context) {
	var q struct {
		Status   string `form:"status"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"query": "查询参数格式错误"}))
		return
	}

	filter := service.ListQuotaRequestFilter{
		Page:     q.Page,
		PageSize: q.PageSize,
	}
	if q.Status != "" {
		filter.Status = &q.Status
	}

	reqs, total, err := h.svc.ListQuotaRequests(c.Request.Context(), filter)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": filter.Page, "page_size": filter.PageSize, "list": reqs})
}

// Approve handles PATCH /api/v1/admin/quota-requests/:id/approve.
func (h *QuotaRequestHandler) Approve(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "申请 ID 格式错误"}))
		return
	}

	resp, err := h.svc.ApproveQuotaRequest(c.Request.Context(), requestID, adminID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// Reject handles PATCH /api/v1/admin/quota-requests/:id/reject.
func (h *QuotaRequestHandler) Reject(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "申请 ID 格式错误"}))
		return
	}

	resp, err := h.svc.RejectQuotaRequest(c.Request.Context(), requestID, adminID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}
