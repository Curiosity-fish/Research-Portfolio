package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/notification"
	"github.com/school-api/school-api-v1/ent/notificationread"
)

// CreateNotificationInput is the data required to create a notification.
type CreateNotificationInput struct {
	UserID  *uuid.UUID
	Type    string
	Title   string
	Content string
}

// ListNotificationFilter controls the admin list query.
type ListNotificationFilter struct {
	UserID *uuid.UUID
	Type   *string
	Offset int
	Limit  int
}

// NotificationRepository provides data access for notifications.
type NotificationRepository interface {
	Create(ctx context.Context, input CreateNotificationInput) (*ent.Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Notification, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.Notification, int, error)
	List(ctx context.Context, filter ListNotificationFilter) ([]*ent.Notification, int, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) (bool, error)
	UnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	ReadAll(ctx context.Context, userID uuid.UUID) error
}

// EntNotificationRepository implements NotificationRepository using Ent.
type EntNotificationRepository struct {
	client *ent.Client
}

// NewEntNotificationRepository creates a new Ent-backed notification repository.
func NewEntNotificationRepository(client *ent.Client) *EntNotificationRepository {
	return &EntNotificationRepository{client: client}
}

// Create inserts a new notification.
func (r *EntNotificationRepository) Create(ctx context.Context, input CreateNotificationInput) (*ent.Notification, error) {
	b := r.client.Notification.Create().
		SetType(notification.Type(input.Type)).
		SetTitle(input.Title).
		SetContent(input.Content)
	if input.UserID != nil {
		b.SetUserID(*input.UserID)
	}

	n, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return n, nil
}

// GetByID returns a notification by ID.
func (r *EntNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Notification, error) {
	n, err := r.client.Notification.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get notification: %w", err)
	}
	return n, nil
}

// ListByUserID returns notifications visible to a user (personal + broadcast).
func (r *EntNotificationRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.Notification, int, error) {
	q := r.client.Notification.Query().
		Where(notification.Or(
			notification.UserIDEQ(userID),
			notification.UserIDIsNil(),
		))

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	list, err := q.Order(ent.Desc(notification.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}

	// Broadcast read state is tracked per user in notification_reads; fold it
	// into IsRead so callers see a single unified flag.
	if len(list) > 0 {
		ids := make([]uuid.UUID, len(list))
		for i, n := range list {
			ids[i] = n.ID
		}
		reads, err := r.client.NotificationRead.Query().
			Where(
				notificationread.UserIDEQ(userID),
				notificationread.NotificationIDIn(ids...),
			).
			All(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("list notification reads: %w", err)
		}
		readSet := make(map[uuid.UUID]struct{}, len(reads))
		for _, rd := range reads {
			readSet[rd.NotificationID] = struct{}{}
		}
		for _, n := range list {
			if _, ok := readSet[n.ID]; ok {
				n.IsRead = true
			}
		}
	}

	return list, total, nil
}

// List returns a paginated list of notifications for the admin panel.
func (r *EntNotificationRepository) List(ctx context.Context, filter ListNotificationFilter) ([]*ent.Notification, int, error) {
	q := r.client.Notification.Query()
	if filter.UserID != nil {
		q = q.Where(notification.UserIDEQ(*filter.UserID))
	}
	if filter.Type != nil {
		q = q.Where(notification.TypeEQ(notification.Type(*filter.Type)))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	list, err := q.Order(ent.Desc(notification.FieldCreatedAt)).
		Offset(filter.Offset).
		Limit(filter.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}

	return list, total, nil
}

// MarkRead marks a notification as read for the given user.
// Personal notifications flip their is_read flag; broadcast notifications
// record a per-user read so other users are unaffected.
func (r *EntNotificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) (bool, error) {
	n, err := r.client.Notification.Query().
		Where(
			notification.IDEQ(id),
			notification.Or(
				notification.UserIDEQ(userID),
				notification.UserIDIsNil(),
			),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("get notification for mark read: %w", err)
	}

	if n.UserID == nil {
		exists, err := r.client.NotificationRead.Query().
			Where(
				notificationread.NotificationIDEQ(id),
				notificationread.UserIDEQ(userID),
			).
			Exist(ctx)
		if err != nil {
			return false, fmt.Errorf("check notification read: %w", err)
		}
		if exists {
			return true, nil
		}

		err = r.client.NotificationRead.Create().
			SetNotificationID(id).
			SetUserID(userID).
			Exec(ctx)
		if err != nil {
			// A concurrent mark-read may have inserted the same row; the
			// unique (notification_id, user_id) constraint makes that a no-op.
			if ent.IsConstraintError(err) {
				return true, nil
			}
			return false, fmt.Errorf("record broadcast notification read: %w", err)
		}
		return true, nil
	}

	if n.IsRead {
		return true, nil
	}

	if _, err := r.client.Notification.UpdateOneID(id).SetIsRead(true).Save(ctx); err != nil {
		return false, fmt.Errorf("mark notification read: %w", err)
	}
	return true, nil
}

// UnreadCount returns the number of notifications the user has not read:
// unread personal notifications plus broadcast notifications without a
// per-user read record.
func (r *EntNotificationRepository) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	personalUnread, err := r.client.Notification.Query().
		Where(
			notification.UserIDEQ(userID),
			notification.IsReadEQ(false),
		).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count unread personal notifications: %w", err)
	}

	broadcastIDs, err := r.client.Notification.Query().
		Where(notification.UserIDIsNil()).
		IDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("list broadcast notification ids: %w", err)
	}

	broadcastUnread := 0
	if len(broadcastIDs) > 0 {
		readCount, err := r.client.NotificationRead.Query().
			Where(
				notificationread.UserIDEQ(userID),
				notificationread.NotificationIDIn(broadcastIDs...),
			).
			Count(ctx)
		if err != nil {
			return 0, fmt.Errorf("count broadcast reads: %w", err)
		}
		broadcastUnread = len(broadcastIDs) - readCount
		if broadcastUnread < 0 {
			broadcastUnread = 0
		}
	}

	return personalUnread + broadcastUnread, nil
}

// ReadAll marks every notification visible to the user as read: personal
// notifications flip their flag; broadcast notifications get a per-user read
// record. Concurrent read-all calls are safe: duplicate inserts hit the
// unique constraint and are tolerated.
func (r *EntNotificationRepository) ReadAll(ctx context.Context, userID uuid.UUID) error {
	if _, err := r.client.Notification.Update().
		Where(
			notification.UserIDEQ(userID),
			notification.IsReadEQ(false),
		).
		SetIsRead(true).
		Save(ctx); err != nil {
		return fmt.Errorf("mark all personal notifications read: %w", err)
	}

	// Broadcast IDs the user has not read yet.
	readRows, err := r.client.NotificationRead.Query().
		Where(notificationread.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("list user read records: %w", err)
	}
	readSet := make(map[uuid.UUID]struct{}, len(readRows))
	for _, rd := range readRows {
		readSet[rd.NotificationID] = struct{}{}
	}

	broadcastIDs, err := r.client.Notification.Query().
		Where(notification.UserIDIsNil()).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("list broadcast notification ids: %w", err)
	}

	for _, id := range broadcastIDs {
		if _, ok := readSet[id]; ok {
			continue
		}
		err := r.client.NotificationRead.Create().
			SetNotificationID(id).
			SetUserID(userID).
			Exec(ctx)
		if err != nil && !ent.IsConstraintError(err) {
			return fmt.Errorf("record broadcast read: %w", err)
		}
	}
	return nil
}

var _ NotificationRepository = (*EntNotificationRepository)(nil)
