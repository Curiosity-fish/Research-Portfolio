package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/auditlog"
	"github.com/school-api/school-api-v1/internal/repository"
)

type mockAuditLogRepository struct {
	createFunc func(ctx context.Context, input repository.CreateAuditLogInput) (*ent.AuditLog, error)
	listFunc   func(ctx context.Context, filter repository.ListAuditLogFilter) ([]*ent.AuditLog, int, error)
}

func (m *mockAuditLogRepository) Create(ctx context.Context, input repository.CreateAuditLogInput) (*ent.AuditLog, error) {
	return m.createFunc(ctx, input)
}

func (m *mockAuditLogRepository) List(ctx context.Context, filter repository.ListAuditLogFilter) ([]*ent.AuditLog, int, error) {
	return m.listFunc(ctx, filter)
}

func TestAuditLogService_Log(t *testing.T) {
	called := false
	repo := &mockAuditLogRepository{
		createFunc: func(ctx context.Context, input repository.CreateAuditLogInput) (*ent.AuditLog, error) {
			called = true
			if input.ActorType != "admin" {
				t.Fatalf("expected actor type admin, got %s", input.ActorType)
			}
			return &ent.AuditLog{ID: uuid.New(), ActorType: auditlog.ActorType(input.ActorType), Action: input.Action}, nil
		},
	}
	svc := NewAuditLogService(repo)

	err := svc.Log(context.Background(), CreateAuditLogInput{
		ActorType: "admin",
		ActorID:   uuid.New(),
		Action:    "GET /api/v1/admin/users",
	})
	if err != nil {
		t.Fatalf("log: %v", err)
	}
	if !called {
		t.Fatal("expected repository create to be called")
	}
}

func TestAuditLogService_ListAuditLogsAdmin(t *testing.T) {
	repo := &mockAuditLogRepository{
		listFunc: func(ctx context.Context, filter repository.ListAuditLogFilter) ([]*ent.AuditLog, int, error) {
			return []*ent.AuditLog{
				{ID: uuid.New(), ActorType: auditlog.ActorTypeAdmin, Action: "GET /api/v1/admin/users"},
			}, 1, nil
		},
	}
	svc := NewAuditLogService(repo)

	logs, total, err := svc.ListAuditLogsAdmin(context.Background(), nil, nil, nil, 1, 20)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(logs) != 1 {
		t.Fatalf("expected 1 log, got total=%d len=%d", total, len(logs))
	}
}
