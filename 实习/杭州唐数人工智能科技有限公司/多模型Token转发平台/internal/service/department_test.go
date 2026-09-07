package service

import (
	"context"
	"testing"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/department"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

func newDepartmentService(t *testing.T, client *ent.Client) *DepartmentService {
	t.Helper()
	return NewDepartmentService(repository.NewEntDepartmentRepository(client))
}

func TestDepartmentService_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newDepartmentService(t, client)

	d, err := svc.CreateDepartment(ctx, CreateDepartmentInput{
		Name:   "教务处",
		Code:   "jwc",
		Status: department.StatusActive,
	})
	if err != nil {
		t.Fatalf("create department: %v", err)
	}
	if d.Code != "jwc" {
		t.Errorf("expected code jwc, got %s", d.Code)
	}

	got, err := svc.GetDepartment(ctx, parseUUID(t, d.ID))
	if err != nil {
		t.Fatalf("get department: %v", err)
	}
	if got.ID != d.ID {
		t.Errorf("expected id %s, got %s", d.ID, got.ID)
	}
}

func TestDepartmentService_CreateDefaults(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newDepartmentService(t, client)

	d, err := svc.CreateDepartment(ctx, CreateDepartmentInput{
		Name: "图书馆",
		Code: "lib",
	})
	if err != nil {
		t.Fatalf("create department: %v", err)
	}
	if d.Status != string(department.StatusActive) {
		t.Errorf("expected active, got %s", d.Status)
	}
	if d.Level != 0 {
		t.Errorf("expected level 0, got %d", d.Level)
	}
}

func TestDepartmentService_DuplicateCode(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newDepartmentService(t, client)

	input := CreateDepartmentInput{Name: "A", Code: "dup", Status: department.StatusActive}
	if _, err := svc.CreateDepartment(ctx, input); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := svc.CreateDepartment(ctx, input)
	if !domain.IsAppError(err) {
		t.Fatalf("expected app error, got %v", err)
	}
	if !domain.ErrConflict.Is(err.(*domain.AppError)) {
		t.Errorf("expected conflict, got %v", err)
	}
}

func TestDepartmentService_List(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newDepartmentService(t, client)

	for _, code := range []string{"a", "b"} {
		_, err := svc.CreateDepartment(ctx, CreateDepartmentInput{
			Name:   code,
			Code:   code,
			Status: department.StatusActive,
		})
		if err != nil {
			t.Fatalf("create department %s: %v", code, err)
		}
	}

	status := department.StatusActive
	list, err := svc.ListDepartments(ctx, &status)
	if err != nil {
		t.Fatalf("list departments: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 departments, got %d", len(list))
	}
}

func TestDepartmentService_Update(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newDepartmentService(t, client)

	d, err := svc.CreateDepartment(ctx, CreateDepartmentInput{
		Name:   "Old",
		Code:   "old",
		Status: department.StatusActive,
	})
	if err != nil {
		t.Fatalf("create department: %v", err)
	}

	newName := "New"
	newCode := "new"
	updated, err := svc.UpdateDepartment(ctx, parseUUID(t, d.ID), UpdateDepartmentInput{
		Name: &newName,
		Code: &newCode,
	})
	if err != nil {
		t.Fatalf("update department: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}
	if updated.Code != newCode {
		t.Errorf("expected code %q, got %q", newCode, updated.Code)
	}
}

func TestDepartmentService_Delete(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := newDepartmentService(t, client)

	d, err := svc.CreateDepartment(ctx, CreateDepartmentInput{
		Name:   "Del",
		Code:   "del",
		Status: department.StatusActive,
	})
	if err != nil {
		t.Fatalf("create department: %v", err)
	}

	if err := svc.DeleteDepartment(ctx, parseUUID(t, d.ID)); err != nil {
		t.Fatalf("delete department: %v", err)
	}

	_, err = svc.GetDepartment(ctx, parseUUID(t, d.ID))
	if !domain.IsAppError(err) {
		t.Fatalf("expected app error, got %v", err)
	}
	if !domain.ErrNotFound.Is(err.(*domain.AppError)) {
		t.Errorf("expected not found, got %v", err)
	}
}
