package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/group"
)

// mockGroupListRepo 只实现 List：GroupService 目前只读列表。
type mockGroupListRepo struct {
	listFunc func(ctx context.Context) ([]*ent.Group, error)
}

func (m *mockGroupListRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Group, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGroupListRepo) List(ctx context.Context) ([]*ent.Group, error) {
	return m.listFunc(ctx)
}

func TestGroupService_ListGroups(t *testing.T) {
	desc := "描述"
	repo := &mockGroupListRepo{
		listFunc: func(ctx context.Context) ([]*ent.Group, error) {
			return []*ent.Group{
				{
					ID:          uuid.New(),
					Name:        "默认分组",
					Code:        "default",
					Description: &desc,
					SortOrder:   0,
					Status:      group.StatusActive,
				},
			}, nil
		},
	}

	svc := NewGroupService(repo)
	got, err := svc.ListGroups(context.Background())
	if err != nil {
		t.Fatalf("list groups: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 group, got %d", len(got))
	}
	if got[0].Code != "default" || got[0].Description == nil || *got[0].Description != "描述" {
		t.Errorf("unexpected response: %+v", got[0])
	}
}

func TestGroupService_ListGroupsError(t *testing.T) {
	repo := &mockGroupListRepo{
		listFunc: func(ctx context.Context) ([]*ent.Group, error) {
			return nil, errors.New("db down")
		},
	}

	svc := NewGroupService(repo)
	if _, err := svc.ListGroups(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}
