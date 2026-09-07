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

// AlertService is the subset of the alert service used by AlertHandler.
type AlertService interface {
	CreateAlertRule(ctx context.Context, input service.CreateAlertRuleInput) (*service.AlertRuleResponse, error)
	GetAlertRule(ctx context.Context, id uuid.UUID) (*service.AlertRuleResponse, error)
	ListAlertRules(ctx context.Context, enabledOnly bool, page, pageSize int) ([]service.AlertRuleResponse, int, error)
	UpdateAlertRule(ctx context.Context, id uuid.UUID, input service.UpdateAlertRuleInput) (*service.AlertRuleResponse, error)
	DeleteAlertRule(ctx context.Context, id uuid.UUID) error
	ListAlertRecords(ctx context.Context, ruleID *uuid.UUID, isResolved *bool, page, pageSize int) ([]service.AlertRecordResponse, int, error)
	ResolveAlertRecord(ctx context.Context, id, adminID uuid.UUID) (*service.AlertRecordResponse, error)
	EvaluateRules(ctx context.Context) (int, error)
}

// CreateAlertRuleRequest is the request body for creating an alert rule.
type CreateAlertRuleRequest struct {
	Name        string  `json:"name" binding:"required,max=200"`
	Metric      string  `json:"metric" binding:"required,oneof=balance_low quota_low error_rate cost_spike"`
	Threshold   int64   `json:"threshold" binding:"required,gt=0"`
	Enabled     bool    `json:"enabled"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
}

// UpdateAlertRuleRequest is the request body for updating an alert rule.
type UpdateAlertRuleRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=200"`
	Metric      *string `json:"metric,omitempty" binding:"omitempty,oneof=balance_low quota_low error_rate cost_spike"`
	Threshold   *int64  `json:"threshold,omitempty" binding:"omitempty,gt=0"`
	Enabled     *bool   `json:"enabled,omitempty"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
}

// AlertHandler handles alert rule and alert record endpoints.
type AlertHandler struct {
	svc AlertService
}

// NewAlertHandler creates a new AlertHandler.
func NewAlertHandler(svc AlertService) *AlertHandler {
	return &AlertHandler{svc: svc}
}

// CreateRule handles POST /api/v1/admin/alert-rules.
func (h *AlertHandler) CreateRule(c *gin.Context) {
	var req CreateAlertRuleRequest
	if !bindAndValidate(c, &req) {
		return
	}

	resp, err := h.svc.CreateAlertRule(c.Request.Context(), service.CreateAlertRuleInput{
		Name:        req.Name,
		Metric:      req.Metric,
		Threshold:   req.Threshold,
		Enabled:     req.Enabled,
		Description: req.Description,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, resp)
}

// ListRules handles GET /api/v1/admin/alert-rules.
func (h *AlertHandler) ListRules(c *gin.Context) {
	var q struct {
		Enabled  *bool `form:"enabled"`
		Page     int   `form:"page"`
		PageSize int   `form:"page_size"`
	}
	_ = c.ShouldBindQuery(&q)

	enabledOnly := false
	if q.Enabled != nil {
		enabledOnly = *q.Enabled
	}

	list, total, err := h.svc.ListAlertRules(c.Request.Context(), enabledOnly, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": list})
}

// GetRule handles GET /api/v1/admin/alert-rules/:id.
func (h *AlertHandler) GetRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "规则 ID 格式错误"}))
		return
	}

	resp, err := h.svc.GetAlertRule(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// UpdateRule handles PUT /api/v1/admin/alert-rules/:id.
func (h *AlertHandler) UpdateRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "规则 ID 格式错误"}))
		return
	}

	var req UpdateAlertRuleRequest
	if !bindAndValidate(c, &req) {
		return
	}

	resp, err := h.svc.UpdateAlertRule(c.Request.Context(), id, service.UpdateAlertRuleInput{
		Name:        req.Name,
		Metric:      req.Metric,
		Threshold:   req.Threshold,
		Enabled:     req.Enabled,
		Description: req.Description,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}

// DeleteRule handles DELETE /api/v1/admin/alert-rules/:id.
func (h *AlertHandler) DeleteRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "规则 ID 格式错误"}))
		return
	}

	if err := h.svc.DeleteAlertRule(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}

	respond.NoContent(c)
}

// Evaluate handles POST /api/v1/admin/alerts/evaluate.
func (h *AlertHandler) Evaluate(c *gin.Context) {
	count, err := h.svc.EvaluateRules(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"created": count})
}

// ListRecords handles GET /api/v1/admin/alerts.
func (h *AlertHandler) ListRecords(c *gin.Context) {
	var q struct {
		RuleID     string `form:"rule_id"`
		IsResolved *bool  `form:"is_resolved"`
		Page       int    `form:"page"`
		PageSize   int    `form:"page_size"`
	}
	_ = c.ShouldBindQuery(&q)

	var ruleID *uuid.UUID
	if q.RuleID != "" {
		id, err := uuid.Parse(q.RuleID)
		if err != nil {
			respond.Error(c, domain.NewValidationError(map[string]string{"rule_id": "规则 ID 格式错误"}))
			return
		}
		ruleID = &id
	}

	list, total, err := h.svc.ListAlertRecords(c.Request.Context(), ruleID, q.IsResolved, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": list})
}

// ResolveRecord handles PATCH /api/v1/admin/alerts/:id/resolve.
func (h *AlertHandler) ResolveRecord(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "告警 ID 格式错误"}))
		return
	}

	resp, err := h.svc.ResolveAlertRecord(c.Request.Context(), id, adminID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, resp)
}
