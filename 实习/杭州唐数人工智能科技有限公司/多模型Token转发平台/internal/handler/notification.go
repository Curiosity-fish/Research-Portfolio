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

// NotificationService is the subset of the notification service used by NotificationHandler.
type NotificationService interface {
	CreateNotification(ctx context.Context, input service.CreateNotificationInput) (*service.NotificationResponse, error)
	ListMyNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]service.NotificationResponse, int, error)
	ListNotificationsAdmin(ctx context.Context, userID *uuid.UUID, notificationType *string, page, pageSize int) ([]service.NotificationResponse, int, error)
	MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error
	UnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	ReadAll(ctx context.Context, userID uuid.UUID) error
}

// CreateNotificationRequest is the request body for creating a notification.
type CreateNotificationRequest struct {
	UserID  *string `json:"user_id,omitempty"`
	Type    string  `json:"type"`
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
}

// NotificationHandler handles notification endpoints.
type NotificationHandler struct {
	svc NotificationService
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(svc NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// ListMy handles GET /api/v1/notifications.
func (h *NotificationHandler) ListMy(c *gin.Context) {
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

	list, total, err := h.svc.ListMyNotifications(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": list})
}

// MarkRead handles PATCH /api/v1/notifications/:id/read.
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"id": "通知 ID 格式错误"}))
		return
	}

	if err := h.svc.MarkRead(c.Request.Context(), userID, id); err != nil {
		respond.Error(c, err)
		return
	}

	respond.NoContent(c)
}

// UnreadCount handles GET /api/v1/notifications/unread-count.
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	count, err := h.svc.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"count": count})
}

// ReadAll handles POST /api/v1/notifications/read-all.
func (h *NotificationHandler) ReadAll(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}

	if err := h.svc.ReadAll(c.Request.Context(), userID); err != nil {
		respond.Error(c, err)
		return
	}

	respond.NoContent(c)
}

// Create handles POST /api/v1/admin/notifications.
func (h *NotificationHandler) Create(c *gin.Context) {
	var req CreateNotificationRequest
	if !bindAndValidate(c, &req) {
		return
	}

	input := service.CreateNotificationInput{
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
	}
	if req.UserID != nil && *req.UserID != "" {
		uid, err := uuid.Parse(*req.UserID)
		if err != nil {
			respond.Error(c, domain.NewValidationError(map[string]string{"user_id": "用户 ID 格式错误"}))
			return
		}
		input.UserID = &uid
	}

	resp, err := h.svc.CreateNotification(c.Request.Context(), input)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, resp)
}

// ListAdmin handles GET /api/v1/admin/notifications.
func (h *NotificationHandler) ListAdmin(c *gin.Context) {
	var q struct {
		UserID string `form:"user_id"`
		Type   string `form:"type"`
		Page   int    `form:"page"`
		PageSize int  `form:"page_size"`
	}
	_ = c.ShouldBindQuery(&q)

	var userID *uuid.UUID
	if q.UserID != "" {
		uid, err := uuid.Parse(q.UserID)
		if err != nil {
			respond.Error(c, domain.NewValidationError(map[string]string{"user_id": "用户 ID 格式错误"}))
			return
		}
		userID = &uid
	}

	var notificationType *string
	if q.Type != "" {
		notificationType = &q.Type
	}

	list, total, err := h.svc.ListNotificationsAdmin(c.Request.Context(), userID, notificationType, q.Page, q.PageSize)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, gin.H{"total": total, "page": q.Page, "page_size": q.PageSize, "list": list})
}
