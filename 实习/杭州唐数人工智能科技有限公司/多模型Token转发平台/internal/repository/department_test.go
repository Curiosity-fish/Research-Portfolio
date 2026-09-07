package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/department"
)

func TestEntDepartmentRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntDepartmentRepository(client)

	root, err := repo.Create(ctx, CreateDepartmentInput{
		Name:      "教务处",
		Code:      "jwc",
		SortOrder: 1,
		Status:    department.StatusActive,
	})
	if err != nil {
		t.Fatalf("create root department: %v", err)
	}
	if root.Level != 0 {
		t.Errorf("expected level 0, got %d", root.Level)
	}
	if root.Path != root.ID.String() {
		t.Errorf("expected path %s, got %s", root.ID.String(), root.Path)
	}

	child, err := repo.Create(ctx, CreateDepartmentInput{
		Name:      "教学科",
		Code:      "jxk",
		ParentID:  &root.ID,
		SortOrder: 1,
		Status:    department.StatusActive,
	})
	if err != nil {
		t.Fatalf("create child department: %v", err)
	}
	if child.Level != 1 {
		t.Errorf("expected level 1, got %d", child.Level)
	}
	if child.ParentID == nil || *child.ParentID != root.ID {
		t.Errorf("expected parent_id %s, got %v", root.ID, child.ParentID)
	}
	expectedPath := root.Path + "/" + child.ID.String()
	if child.Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, child.Path)
	}

	got, err := repo.GetByID(ctx, child.ID)
	if err != nil {
		t.Fatalf("get department: %v", err)
	}
	if got.ID != child.ID {
		t.Errorf("expected id %s, got %s", child.ID, got.ID)
	}

	byCode, err := repo.GetByCode(ctx, "jxk")
	if err != nil {
		t.Fatalf("get by code: %v", err)
	}
	if byCode.ID != child.ID {
		t.Errorf("expected id %s by code, got %s", child.ID, byCode.ID)
	}
}

func TestEntDepartmentRepository_CreateParentNotFound(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntDepartmentRepository(client)

	parentID := uuid.New()
	_, err := repo.Create(ctx, CreateDepartmentInput{
		Name:     "孤儿科",
		Code:     "orphan",
		ParentID: &parentID,
		Status:   department.StatusActive,
	})
	if err == nil {
		t.Fatal("expected error for missing parent")
	}
}

func TestEntDepartmentRepository_List(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntDepartmentRepository(client)

	for _, code := range []string{"a", "b"} {
		_, err := repo.Create(ctx, CreateDepartmentInput{
			Name:   "Dept " + code,
			Code:   code,
			Status: department.StatusActive,
		})
		if err != nil {
			t.Fatalf("create department %s: %v", code, err)
		}
	}

	status := department.StatusActive
	deps, err := repo.List(ctx, ListDepartmentFilter{Status: &status})
	if err != nil {
		t.Fatalf("list departments: %v", err)
	}
	if len(deps) != 2 {
		t.Errorf("expected 2 departments, got %d", len(deps))
	}
}

func TestEntDepartmentRepository_Update(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntDepartmentRepository(client)

	d, err := repo.Create(ctx, CreateDepartmentInput{
		Name:   "Original",
		Code:   "orig",
		Status: department.StatusActive,
	})
	if err != nil {
		t.Fatalf("create department: %v", err)
	}

	newName := "Updated"
	newSort := 5
	updated, err := repo.Update(ctx, d.ID, UpdateDepartmentInput{
		Name:      &newName,
		SortOrder: &newSort,
	})
	if err != nil {
		t.Fatalf("update department: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}
	if updated.SortOrder != newSort {
		t.Errorf("expected sort_order %d, got %d", newSort, updated.SortOrder)
	}
}

func TestEntDepartmentRepository_MoveParent(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntDepartmentRepository(client)

	root1, err := repo.Create(ctx, CreateDepartmentInput{Name: "Root1", Code: "r1", Status: department.StatusActive})
	if err != nil {
		t.Fatalf("create root1: %v", err)
	}
	root2, err := repo.Create(ctx, CreateDepartmentInput{Name: "Root2", Code: "r2", Status: department.StatusActive})
	if err != nil {
		t.Fatalf("create root2: %v", err)
	}
	child, err := repo.Create(ctx, CreateDepartmentInput{
		Name:     "Child",
		Code:     "c1",
		ParentID: &root1.ID,
		Status:   department.StatusActive,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	moved, err := repo.Update(ctx, child.ID, UpdateDepartmentInput{ParentID: &root2.ID})
	if err != nil {
		t.Fatalf("move parent: %v", err)
	}
	if moved.Level != 1 {
		t.Errorf("expected level 1 after move, got %d", moved.Level)
	}
	expectedPath := root2.Path + "/" + child.ID.String()
	if moved.Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, moved.Path)
	}
}

func TestEntDepartmentRepository_MoveToRoot(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntDepartmentRepository(client)

	root, err := repo.Create(ctx, CreateDepartmentInput{Name: "Root", Code: "root", Status: department.StatusActive})
	if err != nil {
		t.Fatalf("create root: %v", err)
	}
	child, err := repo.Create(ctx, CreateDepartmentInput{
		Name:     "Child",
		Code:     "child",
		ParentID: &root.ID,
		Status:   department.StatusActive,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	zero := uuid.UUID{}
	moved, err := repo.Update(ctx, child.ID, UpdateDepartmentInput{ParentID: &zero})
	if err != nil {
		t.Fatalf("move to root: %v", err)
	}
	if moved.Level != 0 {
		t.Errorf("expected level 0, got %d", moved.Level)
	}
	if moved.Path != child.ID.String() {
		t.Errorf("expected path %s, got %s", child.ID.String(), moved.Path)
	}
}

func TestEntDepartmentRepository_MoveCycle(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntDepartmentRepository(client)

	root, err := repo.Create(ctx, CreateDepartmentInput{Name: "Root", Code: "root", Status: department.StatusActive})
	if err != nil {
		t.Fatalf("create root: %v", err)
	}
	child, err := repo.Create(ctx, CreateDepartmentInput{
		Name:     "Child",
		Code:     "child",
		ParentID: &root.ID,
		Status:   department.StatusActive,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	_, err = repo.Update(ctx, root.ID, UpdateDepartmentInput{ParentID: &child.ID})
	if err == nil {
		t.Fatal("expected error for cycle")
	}
}

func TestEntDepartmentRepository_Delete(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntDepartmentRepository(client)

	d, err := repo.Create(ctx, CreateDepartmentInput{Name: "ToDelete", Code: "del", Status: department.StatusActive})
	if err != nil {
		t.Fatalf("create department: %v", err)
	}

	if err := repo.Delete(ctx, d.ID); err != nil {
		t.Fatalf("delete department: %v", err)
	}

	_, err = repo.GetByID(ctx, d.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}
