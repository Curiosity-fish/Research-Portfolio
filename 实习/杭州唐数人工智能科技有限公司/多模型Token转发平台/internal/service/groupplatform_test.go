package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
)

type mockGroupPlatformRepo struct {
	bindFunc                 func(ctx context.Context, groupID, platformID uuid.UUID) (*ent.GroupPlatform, error)
	unbindFunc               func(ctx context.Context, groupID, platformID uuid.UUID) error
	listPlatformsByGroupFunc func(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error)
}

func (m *mockGroupPlatformRepo) Bind(ctx context.Context, groupID, platformID uuid.UUID) (*ent.GroupPlatform, error) {
	return m.bindFunc(ctx, groupID, platformID)
}

func (m *mockGroupPlatformRepo) Unbind(ctx context.Context, groupID, platformID uuid.UUID) error {
	return m.unbindFunc(ctx, groupID, platformID)
}

func (m *mockGroupPlatformRepo) ListPlatformsByGroupID(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	return m.listPlatformsByGroupFunc(ctx, groupID)
}

func (m *mockGroupPlatformRepo) ListGroupsByPlatformID(ctx context.Context, platformID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

type mockGroupRepo struct {
	getFunc func(ctx context.Context, id uuid.UUID) (*ent.Group, error)
}

func (m *mockGroupRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Group, error) {
	return m.getFunc(ctx, id)
}

func (m *mockGroupRepo) List(ctx context.Context) ([]*ent.Group, error) {
	return nil, nil
}

func TestGroupPlatformService_Bind(t *testing.T) {
	groupID := uuid.New()
	platformID := uuid.New()

	repo := &mockGroupPlatformRepo{
		bindFunc: func(ctx context.Context, gid, pid uuid.UUID) (*ent.GroupPlatform, error) {
			if gid != groupID || pid != platformID {
				t.Errorf("unexpected ids: %s %s", gid, pid)
			}
			return &ent.GroupPlatform{GroupID: gid, PlatformID: pid}, nil
		},
	}
	groupRepo := &mockGroupRepo{
		getFunc: func(ctx context.Context, id uuid.UUID) (*ent.Group, error) {
			return &ent.Group{ID: id}, nil
		},
	}
	platformRepo := &mockPlatformRepo{
		getFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: id}, nil
		},
	}

	svc := NewGroupPlatformService(repo, groupRepo, platformRepo)
	if err := svc.BindPlatform(context.Background(), groupID, platformID); err != nil {
		t.Fatalf("bind: %v", err)
	}
}

func TestGroupPlatformService_BindGroupNotFound(t *testing.T) {
	repo := &mockGroupPlatformRepo{}
	groupRepo := &mockGroupRepo{
		getFunc: func(ctx context.Context, id uuid.UUID) (*ent.Group, error) {
			return nil, &ent.NotFoundError{}
		},
	}
	platformRepo := &mockPlatformRepo{}

	svc := NewGroupPlatformService(repo, groupRepo, platformRepo)
	err := svc.BindPlatform(context.Background(), uuid.New(), uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
