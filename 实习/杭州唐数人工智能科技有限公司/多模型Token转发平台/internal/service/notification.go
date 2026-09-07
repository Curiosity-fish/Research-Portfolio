package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/notification"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// NotificationResponse is the public representation of a notification.
type NotificationResponse struct {
	ID        string  `json:"id"`
	UserID    *string `json:"user_id,omitempty"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	IsRead    bool    `json:"is_read"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CreateNotificationInput is the data required to create a notification.
type CreateNotificationInput struct {
	UserID  *uuid.UUID
	Type    string
	Title   string
	Content string
}

// NotificationService handles notification business logic.
type NotificationService struct {
	repo       repository.NotificationRepository
	userRepo   repository.UserRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(repo repository.NotificationRepository, userRepo repository.UserRepository) *NotificationService {
	return &NotificationService{repo: repo, userRepo: userRepo}
}

// CreateNotification creates a notification or announcement.
func (s *NotificationService) CreateNotification(ctx context.Context, input CreateNotificationInput) (*NotificationResponse, error) {
	if input.Type == "" {
		input.Type = string(notification.TypeAnnouncement)
	}
	if input.Type != string(notification.TypeAnnouncement) && input.Type != string(notification.TypeSystem) {
		return nil, domain.NewAppError(400, "INVALID_NOTIFICATION_TYPE", "通知类型无效")
	}
	if input.Title == "" {
		return nil, domain.NewValidationError(map[string]string{"title": "标题不能为空"})
	}
	if input.Content == "" {
		return nil, domain.NewValidationError(map[string]string{"content": "内容不能为空"})
	}

	if input.UserID != nil {
		if _, err := s.userRepo.GetByID(ctx, *input.UserID); err != nil {
			if ent.IsNotFound(err) {
				return nil, domain.ErrNotFound
			}
			return nil, domain.WrapInternal(err)
		}
	}

	n, err := s.repo.Create(ctx, repository.CreateNotificationInput{
		UserID:  input.UserID,
		Type:    input.Type,
		Title:   input.Title,
		Content: input.Content,
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	return toNotificationResponse(n), nil
}

// ListMyNotifications returns notifications visible to the current user.
func (s *NotificationService) ListMyNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]NotificationResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	list, total, err := s.repo.ListByUserID(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]NotificationResponse, 0, len(list))
	for _, n := range list {
		resp = append(resp, *toNotificationResponse(n))
	}
	return resp, total, nil
}

// ListNotificationsAdmin returns notifications for the admin panel.
func (s *NotificationService) ListNotificationsAdmin(ctx context.Context, userID *uuid.UUID, notificationType *string, page, pageSize int) ([]NotificationResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	list, total, err := s.repo.List(ctx, repository.ListNotificationFilter{
		UserID: userID,
		Type:   notificationType,
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	})
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]NotificationResponse, 0, len(list))
	for _, n := range list {
		resp = append(resp, *toNotificationResponse(n))
	}
	return resp, total, nil
}

// MarkRead marks a notification as read for the current user.
func (s *NotificationService) MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	ok, err := s.repo.MarkRead(ctx, notificationID, userID)
	if err != nil {
		return domain.WrapInternal(err)
	}
	if !ok {
		return domain.ErrNotFound
	}
	return nil
}

// UnreadCount returns the current user's unread notification count.
func (s *NotificationService) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := s.repo.UnreadCount(ctx, userID)
	if err != nil {
		return 0, domain.WrapInternal(err)
	}
	return count, nil
}

// ReadAll marks all notifications visible to the user as read.
func (s *NotificationService) ReadAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.ReadAll(ctx, userID); err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

func toNotificationResponse(n *ent.Notification) *NotificationResponse {
	resp := &NotificationResponse{
		ID:        n.ID.String(),
		Type:      string(n.Type),
		Title:     n.Title,
		Content:   n.Content,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: n.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if n.UserID != nil {
		s := n.UserID.String()
		resp.UserID = &s
	}
	return resp
}

var _ repository.NotificationRepository = (*repository.EntNotificationRepository)(nil)
