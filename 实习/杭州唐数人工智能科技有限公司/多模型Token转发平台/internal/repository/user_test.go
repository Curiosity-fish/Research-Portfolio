package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/school-api/school-api-v1/ent/department"
	"github.com/school-api/school-api-v1/ent/group"
	"github.com/school-api/school-api-v1/ent/user"
)

func TestEntUserRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)

	repo := NewEntUserRepository(client)

	input := CreateUserInput{
		Username:     "alice",
		PasswordHash: "hash",
		Name:         "Alice",
		Email:        "alice@example.com",
		Phone:        "13800138000",
		Role:         user.RoleStudent,
		Gender:       "female",
		Status:       user.StatusActive,
	}

	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created.Username != input.Username {
		t.Errorf("expected username %q, got %q", input.Username, created.Username)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected user id %s, got %s", created.ID, got.ID)
	}

	byName, err := repo.GetByUsername(ctx, input.Username)
	if err != nil {
		t.Fatalf("get user by username: %v", err)
	}
	if byName.ID != created.ID {
		t.Errorf("expected user id %s by username, got %s", created.ID, byName.ID)
	}
}

func TestEntUserRepository_CreateDuplicateUsername(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntUserRepository(client)

	input := CreateUserInput{
		Username:     "bob",
		PasswordHash: "hash",
		Name:         "Bob",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	}
	if _, err := repo.Create(ctx, input); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.Create(ctx, input)
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}

func TestEntUserRepository_List(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntUserRepository(client)

	for _, name := range []string{"carol", "dave"} {
		_, err := repo.Create(ctx, CreateUserInput{
			Username:     name,
			PasswordHash: "hash",
			Name:         name,
			Role:         user.RoleStudent,
			Status:       user.StatusActive,
		})
		if err != nil {
			t.Fatalf("create user %s: %v", name, err)
		}
	}

	active := user.StatusActive
	users, total, err := repo.List(ctx, ListUserFilter{
		Keyword:  "carol",
		Status:   &active,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestEntUserRepository_ListByDepartment(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntUserRepository(client)

	dep, err := client.Department.Create().
		SetName("过滤测试班").
		SetCode("flt-" + uuid.NewString()[:8]).
		Save(ctx)
	if err != nil {
		t.Fatalf("create department: %v", err)
	}

	for _, name := range []string{"gina", "heidi"} {
		_, err := repo.Create(ctx, CreateUserInput{
			Username:     name,
			PasswordHash: "hash",
			Name:         name,
			Role:         user.RoleStudent,
			Status:       user.StatusActive,
			DepartmentID: &dep.ID,
		})
		if err != nil {
			t.Fatalf("create user %s: %v", name, err)
		}
	}
	if _, err := repo.Create(ctx, CreateUserInput{
		Username:     "ivan",
		PasswordHash: "hash",
		Name:         "ivan",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	}); err != nil {
		t.Fatalf("create user ivan: %v", err)
	}

	users, total, err := repo.List(ctx, ListUserFilter{DepartmentID: &dep.ID, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list by department: %v", err)
	}
	if total != 2 || len(users) != 2 {
		t.Fatalf("expected 2 users in department, got total=%d len=%d", total, len(users))
	}
	for _, u := range users {
		if u.DepartmentID == nil || *u.DepartmentID != dep.ID {
			t.Fatalf("user %s not in department", u.Username)
		}
	}
}

func TestEntUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntUserRepository(client)

	created, err := repo.Create(ctx, CreateUserInput{
		Username:     "eve",
		PasswordHash: "hash",
		Name:         "Eve",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	newName := "Eve Updated"
	newEmail := "eve@example.com"
	role := user.RoleTeacher
	updated, err := repo.Update(ctx, created.ID, UpdateUserInput{
		Name:   &newName,
		Email:  &newEmail,
		Role:   &role,
		Gender: new(string), // clear gender
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}
	if updated.Email == nil || *updated.Email != newEmail {
		t.Errorf("expected email %q, got %v", newEmail, updated.Email)
	}
	if updated.Role != role {
		t.Errorf("expected role %q, got %q", role, updated.Role)
	}
	if updated.Gender != nil && *updated.Gender != "" {
		t.Errorf("expected gender cleared, got %v", updated.Gender)
	}
}

func TestEntUserRepository_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntUserRepository(client)

	created, err := repo.Create(ctx, CreateUserInput{
		Username:     "frank",
		PasswordHash: "hash",
		Name:         "Frank",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	updated, err := repo.UpdateStatus(ctx, created.ID, user.StatusInactive)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.Status != user.StatusInactive {
		t.Errorf("expected status inactive, got %q", updated.Status)
	}
}

func TestEntUserRepository_UpdateClearsFKs(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntUserRepository(client)

	dept, err := client.Department.Create().
		SetName("测试部门").
		SetCode("dept-" + uuid.NewString()[:8]).
		SetStatus(department.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create department: %v", err)
	}

	grp, err := client.Group.Create().
		SetName("测试分组").
		SetCode("group-" + uuid.NewString()[:8]).
		SetStatus(group.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	created, err := repo.Create(ctx, CreateUserInput{
		Username:     "grace",
		PasswordHash: "hash",
		Name:         "Grace",
		Role:         user.RoleStudent,
		Status:       user.StatusActive,
		DepartmentID: &dept.ID,
		GroupID:      &grp.ID,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created.DepartmentID == nil || *created.DepartmentID != dept.ID {
		t.Fatal("expected department_id set")
	}
	if created.GroupID == nil || *created.GroupID != grp.ID {
		t.Fatal("expected group_id set")
	}

	zero := uuid.UUID{}
	updated, err := repo.Update(ctx, created.ID, UpdateUserInput{
		DepartmentID: &zero,
		GroupID:      &zero,
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.DepartmentID != nil {
		t.Errorf("expected department_id cleared, got %s", *updated.DepartmentID)
	}
	if updated.GroupID != nil {
		t.Errorf("expected group_id cleared, got %s", *updated.GroupID)
	}
}
