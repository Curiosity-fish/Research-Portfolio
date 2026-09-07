package repository

import (
	"context"
	"testing"

	"github.com/school-api/school-api-v1/ent/group"
)

func TestEntGroupRepository_List(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntGroupRepository(client)

	second, err := client.Group.Create().
		SetName("第二分组").
		SetCode("group-b").
		SetSortOrder(2).
		SetStatus(group.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create second group: %v", err)
	}

	first, err := client.Group.Create().
		SetName("第一分组").
		SetCode("group-a").
		SetSortOrder(1).
		SetStatus(group.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create first group: %v", err)
	}

	got, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list groups: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(got))
	}
	if got[0].ID != first.ID || got[1].ID != second.ID {
		t.Errorf("expected sort_order ascending, got %s then %s", got[0].Name, got[1].Name)
	}
}
