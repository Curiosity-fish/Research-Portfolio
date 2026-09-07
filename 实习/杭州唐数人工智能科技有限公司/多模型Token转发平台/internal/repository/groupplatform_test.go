package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/group"
	"github.com/school-api/school-api-v1/ent/platform"
)

func TestEntGroupPlatformRepository_BindUnbind(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)

	g, err := client.Group.Create().
		SetName("Default").
		SetCode("default").
		SetStatus(group.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	p, err := client.Platform.Create().
		SetName("OpenAI").
		SetCode("openai").
		SetType(platform.TypeOpenai).
		SetBaseURL("https://api.openai.com/v1").
		SetStatus(platform.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}

	repo := NewEntGroupPlatformRepository(client)

	gp, err := repo.Bind(ctx, g.ID, p.ID)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if gp.GroupID != g.ID || gp.PlatformID != p.ID {
		t.Error("binding ids mismatch")
	}

	// Idempotent: binding again should not error.
	_, err = repo.Bind(ctx, g.ID, p.ID)
	if err != nil {
		t.Fatalf("re-bind: %v", err)
	}

	platforms, err := repo.ListPlatformsByGroupID(ctx, g.ID)
	if err != nil {
		t.Fatalf("list platforms by group: %v", err)
	}
	if len(platforms) != 1 || platforms[0] != p.ID {
		t.Errorf("expected 1 platform %s, got %v", p.ID, platforms)
	}

	groups, err := repo.ListGroupsByPlatformID(ctx, p.ID)
	if err != nil {
		t.Fatalf("list groups by platform: %v", err)
	}
	if len(groups) != 1 || groups[0] != g.ID {
		t.Errorf("expected 1 group %s, got %v", g.ID, groups)
	}

	if err := repo.Unbind(ctx, g.ID, p.ID); err != nil {
		t.Fatalf("unbind: %v", err)
	}

	platforms, err = repo.ListPlatformsByGroupID(ctx, g.ID)
	if err != nil {
		t.Fatalf("list platforms after unbind: %v", err)
	}
	if len(platforms) != 0 {
		t.Errorf("expected 0 platforms after unbind, got %d", len(platforms))
	}
}

func TestEntGroupPlatformRepository_BindMissingGroup(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntGroupPlatformRepository(client)

	p, err := client.Platform.Create().
		SetName("OpenAI").
		SetCode("openai").
		SetType(platform.TypeOpenai).
		SetBaseURL("https://api.openai.com/v1").
		SetStatus(platform.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}

	_, err = repo.Bind(ctx, uuid.New(), p.ID)
	if err == nil {
		t.Error("expected error binding non-existent group")
	}
}
