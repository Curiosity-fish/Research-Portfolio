package service

import (
	"context"
	stdsql "database/sql"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

func openTestClient(t *testing.T) *ent.Client {
	t.Helper()

	db, err := stdsql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	drv := sql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	t.Cleanup(func() { _ = client.Close() })

	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return client
}

func parseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", s, err)
	}
	return id
}

func newUUID(t *testing.T) uuid.UUID {
	t.Helper()
	return parseUUID(t, uuid.New().String())
}

func TestUserService_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), auth.DefaultBcryptCost)

	u, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "svc-user-1",
		Password: "password123",
		Name:     "Service User",
		Email:    "svc1@example.com",
		Role:     user.RoleTeacher,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if u.Username != "svc-user-1" {
		t.Errorf("expected username svc-user-1, got %s", u.Username)
	}
	if u.Role != string(user.RoleTeacher) {
		t.Errorf("expected role teacher, got %s", u.Role)
	}

	got, err := svc.GetUser(ctx, parseUUID(t, u.ID))
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("expected id %s, got %s", u.ID, got.ID)
	}
}

func TestUserService_CreateDefaults(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), auth.DefaultBcryptCost)

	u, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "svc-user-2",
		Password: "password123",
		Name:     "Default User",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if u.Status != string(user.StatusActive) {
		t.Errorf("expected active status, got %s", u.Status)
	}
	if u.Role != string(user.RoleStudent) {
		t.Errorf("expected student role, got %s", u.Role)
	}
}

func TestUserService_DuplicateUsername(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), auth.DefaultBcryptCost)

	input := CreateUserInput{
		Username: "dup-user",
		Password: "password123",
		Name:     "Dup",
	}
	if _, err := svc.CreateUser(ctx, input); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := svc.CreateUser(ctx, input)
	if !domain.IsAppError(err) {
		t.Fatalf("expected app error, got %v", err)
	}
	appErr := err.(*domain.AppError)
	if !domain.ErrConflict.Is(appErr) {
		t.Errorf("expected conflict, got %v", err)
	}
}

func TestUserService_List(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), auth.DefaultBcryptCost)

	for _, name := range []string{"alpha", "beta"} {
		_, err := svc.CreateUser(ctx, CreateUserInput{
			Username: "list-" + name,
			Password: "password123",
			Name:     name,
		})
		if err != nil {
			t.Fatalf("create user %s: %v", name, err)
		}
	}

	status := user.StatusActive
	resp, err := svc.ListUsers(ctx, ListUsersFilter{
		Keyword:  "alpha",
		Status:   &status,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
	if len(resp.List) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.List))
	}
}

func TestUserService_Update(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), auth.DefaultBcryptCost)

	u, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "update-user",
		Password: "password123",
		Name:     "Before",
		Gender:   "male",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	newName := "After"
	newEmail := "after@example.com"
	role := user.RoleStaff
	updated, err := svc.UpdateUser(ctx, parseUUID(t, u.ID), UpdateUserInput{
		Name:  &newName,
		Email: &newEmail,
		Role:  &role,
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
	if updated.Role != string(role) {
		t.Errorf("expected role %q, got %q", role, updated.Role)
	}
	if updated.Gender == nil || *updated.Gender != "male" {
		t.Errorf("expected gender unchanged, got %v", updated.Gender)
	}
}

func TestUserService_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), auth.DefaultBcryptCost)

	u, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "status-user",
		Password: "password123",
		Name:     "Status",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	updated, err := svc.UpdateUserStatus(ctx, parseUUID(t, u.ID), user.StatusBanned)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.Status != string(user.StatusBanned) {
		t.Errorf("expected banned, got %s", updated.Status)
	}
}

func TestUserService_GetNotFound(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), auth.DefaultBcryptCost)

	_, err := svc.GetUser(ctx, newUUID(t))
	if !domain.IsAppError(err) {
		t.Fatalf("expected app error, got %v", err)
	}
	appErr := err.(*domain.AppError)
	if !domain.ErrNotFound.Is(appErr) {
		t.Errorf("expected not found, got %v", err)
	}
}
