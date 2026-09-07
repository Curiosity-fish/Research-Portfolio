package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/notification"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

type mockNotificationRepository struct {
	createFunc      func(ctx context.Context, input repository.CreateNotificationInput) (*ent.Notification, error)
	getByIDFunc     func(ctx context.Context, id uuid.UUID) (*ent.Notification, error)
	listByUserIDFunc func(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.Notification, int, error)
	listFunc        func(ctx context.Context, filter repository.ListNotificationFilter) ([]*ent.Notification, int, error)
	markReadFunc    func(ctx context.Context, id, userID uuid.UUID) (bool, error)
}

func (m *mockNotificationRepository) Create(ctx context.Context, input repository.CreateNotificationInput) (*ent.Notification, error) {
	return m.createFunc(ctx, input)
}

func (m *mockNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Notification, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockNotificationRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*ent.Notification, int, error) {
	return m.listByUserIDFunc(ctx, userID, offset, limit)
}

func (m *mockNotificationRepository) List(ctx context.Context, filter repository.ListNotificationFilter) ([]*ent.Notification, int, error) {
	return m.listFunc(ctx, filter)
}

func (m *mockNotificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) (bool, error) {
	return m.markReadFunc(ctx, id, userID)
}

type mockUserRepository struct {
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*ent.User, error)
}

func (m *mockUserRepository) Create(ctx context.Context, input repository.CreateUserInput) (*ent.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) List(ctx context.Context, filter repository.ListUserFilter) ([]*ent.User, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *mockUserRepository) Update(ctx context.Context, id uuid.UUID, input repository.UpdateUserInput) (*ent.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status user.Status) (*ent.User, error) {
	return nil, errors.New("not implemented")
}

func TestNotificationService_CreateBroadcast(t *testing.T) {
	repo := &mockNotificationRepository{
		createFunc: func(ctx context.Context, input repository.CreateNotificationInput) (*ent.Notification, error) {
			if input.UserID != nil {
				t.Fatalf("expected broadcast user_id nil, got %v", input.UserID)
			}
			return &ent.Notification{ID: uuid.New(), Type: notification.Type(input.Type), Title: input.Title, Content: input.Content}, nil
		},
	}
	svc := NewNotificationService(repo, nil)

	resp, err := svc.CreateNotification(context.Background(), CreateNotificationInput{
		Type:    "announcement",
		Title:   "公告",
		Content: "内容",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.Title != "公告" {
		t.Fatalf("unexpected title %s", resp.Title)
	}
}

func TestNotificationService_CreatePersonalUserNotFound(t *testing.T) {
	uid := uuid.New()
	repo := &mockNotificationRepository{}
	userRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.User, error) {
			return nil, &ent.NotFoundError{}
		},
	}
	svc := NewNotificationService(repo, userRepo)

	_, err := svc.CreateNotification(context.Background(), CreateNotificationInput{
		UserID:  &uid,
		Type:    "system",
		Title:   "通知",
		Content: "内容",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestNotificationService_MarkReadNotFound(t *testing.T) {
	uid := uuid.New()
	nid := uuid.New()
	repo := &mockNotificationRepository{
		markReadFunc: func(ctx context.Context, id, userID uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	svc := NewNotificationService(repo, nil)

	err := svc.MarkRead(context.Background(), uid, nid)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestNotificationService_InvalidType(t *testing.T) {
	svc := NewNotificationService(&mockNotificationRepository{}, nil)
	_, err := svc.CreateNotification(context.Background(), CreateNotificationInput{
		Type:    "invalid",
		Title:   "t",
		Content: "c",
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func (m *mockNotificationRepository) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockNotificationRepository) ReadAll(ctx context.Context, userID uuid.UUID) error {
	return nil
}
