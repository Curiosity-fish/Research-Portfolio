package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestEntAuditLogRepository_CreateAndList(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	repo := NewEntAuditLogRepository(client)

	actorID := uuid.New()
	details := map[string]any{"path": "/api/v1/admin/users", "method": "POST"}
	ip := "127.0.0.1"
	ua := "test-agent"
	log, err := repo.Create(ctx, CreateAuditLogInput{
		ActorType: "admin",
		ActorID:   actorID,
		Action:    "POST /api/v1/admin/users",
		Details:   details,
		IP:        &ip,
		UserAgent: &ua,
	})
	if err != nil {
		t.Fatalf("create audit log: %v", err)
	}
	if log.ActorID != actorID {
		t.Fatal("actor id mismatch")
	}

	actorType := "admin"
	list, total, err := repo.List(ctx, ListAuditLogFilter{ActorType: &actorType, Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 log, got total=%d len=%d", total, len(list))
	}
}

func TestEntAuditLogRepository_ListFilterByAction(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	repo := NewEntAuditLogRepository(client)
	_, err := repo.Create(ctx, CreateAuditLogInput{
		ActorType: "admin",
		ActorID:   uuid.New(),
		Action:    "GET /api/v1/admin/audit-logs",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	action := "GET /api/v1/admin/audit-logs"
	list, total, err := repo.List(ctx, ListAuditLogFilter{Action: &action, Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 log, got total=%d len=%d", total, len(list))
	}
}

func TestEntAuditLogRepository_ListPagination(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	repo := NewEntAuditLogRepository(client)
	for i := 0; i < 3; i++ {
		if _, err := repo.Create(ctx, CreateAuditLogInput{
			ActorType: "admin",
			ActorID:   uuid.New(),
			Action:    "TEST",
		}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	list, total, err := repo.List(ctx, ListAuditLogFilter{Offset: 2, Limit: 2})
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if total != 3 || len(list) != 1 {
		t.Fatalf("expected total=3 len=1, got total=%d len=%d", total, len(list))
	}
}

func TestEntAuditLogRepository_ActorTypeValidation(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	repo := NewEntAuditLogRepository(client)
	_, err := repo.Create(ctx, CreateAuditLogInput{
		ActorType: "invalid",
		ActorID:   uuid.New(),
		Action:    "TEST",
	})
	if err == nil {
		t.Fatal("expected error for invalid actor type")
	}
}
