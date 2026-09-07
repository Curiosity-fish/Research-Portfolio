package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// AuditLogService is the subset of the audit log service used by AuditLogHandler.
type AuditLogService interface {
	ListAuditLogsAdmin(ctx context.Context, actorType *string, actorID *uuid.UUID, action *string, page, pageSize int) ([]service.AuditLogResponse, int, error)
}

// AuditLogHandler handles audit log endpoints.
type AuditLogHandler struct {
	svc AuditLogService
}

// NewAuditLogHandler creates a new AuditLogHandler.
func NewAuditLogHandler(svc AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{svc: svc}
}

// ListAdmin handles GET /api/v1/admin/audit-logs.
func (h *AuditLogHandler) ListAdmin(c *gin.Context) {
	var q struct {
		ActorType string `form:"actor_type"`
		ActorID   string `form:"actor_id"`
		Action    string `form:"action"`
		Page      int    `form:"page"`
		PageSize  int    `form:"page_size"`
	}
	_ = c.ShouldBindQuery(&q)

	var actorType *string
	if q.ActorType != "" {
		actorType = &q.ActorType
	}

	var actorID *uuid.UUID
	if q.ActorID != "" {
		id, err := uuid.Parse(q.ActorID)
		if err != nil {
			respond.Error(c, domain.NewValidationError(map[string]string{"actor_id": "actor_id 格式错误"}))
			return
		}
		actorID = &id
	}

	var action *string
	if q.Action != "" {
		action = &q.Action
	}

	logs, total, err := h.svc.ListAuditLogsAdmin(c.Request.Context(), actorType, actorID, action, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": logs})
}
