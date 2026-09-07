package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// QuotaBatchService is the subset of the batch service used by QuotaBatchHandler.
type QuotaBatchService interface {
	Apply(ctx context.Context, input service.ApplyQuotaBatchInput) (*repository.QuotaBatchResult, error)
}

// ApplyQuotaBatchRequest is the request body for a batch quota operation.
// Exactly one of user_ids and department_id must be present; cross-field and
// range validation happens in the service.
type ApplyQuotaBatchRequest struct {
	UserIDs      []string `json:"user_ids" binding:"omitempty,min=1,max=1000,dive,uuid"`
	DepartmentID string   `json:"department_id" binding:"omitempty,uuid"`
	Mode         string   `json:"mode" binding:"required,oneof=set add"`
	Value        int64    `json:"value" binding:"required"`
}

// QuotaBatchHandler handles batch quota endpoints.
type QuotaBatchHandler struct {
	svc QuotaBatchService
}

// NewQuotaBatchHandler creates a new QuotaBatchHandler.
func NewQuotaBatchHandler(svc QuotaBatchService) *QuotaBatchHandler {
	return &QuotaBatchHandler{svc: svc}
}

// Apply handles POST /api/v1/admin/quota-batch.
func (h *QuotaBatchHandler) Apply(c *gin.Context) {
	var req ApplyQuotaBatchRequest
	if !bindAndValidate(c, &req) {
		return
	}

	input := service.ApplyQuotaBatchInput{Mode: req.Mode, Value: req.Value}
	if len(req.UserIDs) > 0 {
		input.UserIDs = make([]uuid.UUID, 0, len(req.UserIDs))
		for _, raw := range req.UserIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				respond.Error(c, domain.NewValidationError(map[string]string{
					"user_ids": "用户 ID 格式错误: " + raw,
				}))
				return
			}
			input.UserIDs = append(input.UserIDs, id)
		}
	}
	if req.DepartmentID != "" {
		id, err := uuid.Parse(req.DepartmentID)
		if err != nil {
			respond.Error(c, domain.NewValidationError(map[string]string{
				"department_id": "部门 ID 格式错误",
			}))
			return
		}
		input.DepartmentID = &id
	}

	result, err := h.svc.Apply(c.Request.Context(), input)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, result)
}
