package ent_test

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/ent/department"
	"github.com/school-api/school-api-v1/ent/setting"
	"github.com/school-api/school-api-v1/ent/user"
	_ "modernc.org/sqlite"
)

// openTestClient creates an in-memory SQLite client and runs auto-migration.
// It uses modernc.org/sqlite (pure Go, no CGO) but tells ent the dialect is
// "sqlite3" so the migration layer recognizes it.
func openTestClient(t *testing.T) *ent.Client {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	stdDB, err := stdsql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// Use a single connection so the PRAGMA foreign_keys=ON setting persists
	// across the schema creation queries.
	stdDB.SetMaxOpenConns(1)
	if _, err := stdDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	db := sql.OpenDB(dialect.SQLite, stdDB)
	client := ent.NewClient(ent.Driver(db))
	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return client
}

func TestDepartment_TreeAndUsers(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	// Create parent department.
	parent, err := client.Department.Create().
		SetName("计算机学院").
		SetCode("CS").
		SetLevel(0).
		SetPath("CS.").
		Save(ctx)
	if err != nil {
		t.Fatalf("create parent department: %v", err)
	}

	// Create child department.
	child, err := client.Department.Create().
		SetName("软件工程系").
		SetCode("SE").
		SetParentID(parent.ID).
		SetLevel(1).
		SetPath("CS.SE.").
		Save(ctx)
	if err != nil {
		t.Fatalf("create child department: %v", err)
	}

	// Verify parent-child relationship.
	parentReload, err := client.Department.Query().
		Where(department.ID(parent.ID)).
		WithChildren().
		Only(ctx)
	if err != nil {
		t.Fatalf("query parent with children: %v", err)
	}
	if len(parentReload.Edges.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parentReload.Edges.Children))
	}
	if parentReload.Edges.Children[0].ID != child.ID {
		t.Errorf("expected child %v, got %v", child.ID, parentReload.Edges.Children[0].ID)
	}

	// Create user in child department.
	u, err := client.User.Create().
		SetUsername("zhangsan").
		SetPasswordHash("hash").
		SetName("张三").
		SetDepartmentID(child.ID).
		SetRole(user.RoleStudent).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Verify user-department relationship.
	uReload, err := client.User.Query().
		Where(user.ID(u.ID)).
		WithDepartment().
		Only(ctx)
	if err != nil {
		t.Fatalf("query user with department: %v", err)
	}
	if uReload.Edges.Department == nil {
		t.Fatal("expected user to have department")
	}
	if uReload.Edges.Department.ID != child.ID {
		t.Errorf("expected department %v, got %v", child.ID, uReload.Edges.Department.ID)
	}
}

func TestDepartment_CodeUnique(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	_, err := client.Department.Create().
		SetName("学院A").
		SetCode("UNIQUE_CODE").
		Save(ctx)
	if err != nil {
		t.Fatalf("create first department: %v", err)
	}

	_, err = client.Department.Create().
		SetName("学院B").
		SetCode("UNIQUE_CODE").
		Save(ctx)
	if err == nil {
		t.Fatal("expected unique constraint error for duplicate department code")
	}
}

func TestUser_UsernameEmailPhoneUnique(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	createUser := func(username, email, phone string) error {
		return client.User.Create().
			SetUsername(username).
			SetPasswordHash("hash").
			SetName(username).
			SetEmail(email).
			SetPhone(phone).
			Exec(ctx)
	}

	if err := createUser("alice", "alice@example.com", "13800138000"); err != nil {
		t.Fatalf("create first user: %v", err)
	}

	cases := []struct {
		name    string
		email   string
		phone   string
		wantErr bool
	}{
		{"alice2", "alice@example.com", "13800138001", true}, // duplicate email
		{"alice3", "bob@example.com", "13800138000", true},   // duplicate phone
		{"alice", "carol@example.com", "13800138002", true},  // duplicate username
		{"bob", "bob@example.com", "13800138003", false},     // ok
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := createUser(tc.name, tc.email, tc.phone)
			if tc.wantErr && err == nil {
				t.Fatal("expected unique constraint error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAdminUser_CRUD(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	admin, err := client.AdminUser.Create().
		SetUsername("admin").
		SetPasswordHash("hash").
		SetRole(adminuser.RoleSuperAdmin).
		Save(ctx)
	if err != nil {
		t.Fatalf("create admin user: %v", err)
	}

	if admin.Role != adminuser.RoleSuperAdmin {
		t.Errorf("expected role super_admin, got %s", admin.Role)
	}
	if admin.Status != adminuser.StatusActive {
		t.Errorf("expected default status active, got %s", admin.Status)
	}

	// Duplicate username should fail.
	_, err = client.AdminUser.Create().
		SetUsername("admin").
		SetPasswordHash("hash2").
		Save(ctx)
	if err == nil {
		t.Fatal("expected unique constraint error for duplicate admin username")
	}
}

func TestSetting_KeyValue(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	s, err := client.Setting.Create().
		SetKey("site.name").
		SetValue("School API").
		SetType(setting.TypeString).
		SetIsPublic(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("create setting: %v", err)
	}

	if s.Key != "site.name" {
		t.Errorf("expected key site.name, got %s", s.Key)
	}
	if !s.IsPublic {
		t.Error("expected setting to be public")
	}

	// Duplicate key should fail.
	_, err = client.Setting.Create().
		SetKey("site.name").
		SetValue("Duplicate").
		Save(ctx)
	if err == nil {
		t.Fatal("expected unique constraint error for duplicate setting key")
	}
}

func TestMixin_Timestamps(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	admin, err := client.AdminUser.Create().
		SetUsername("ts-admin").
		SetPasswordHash("hash").
		Save(ctx)
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}

	if admin.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if admin.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
	if admin.CreatedAt.After(admin.UpdatedAt) {
		t.Errorf("expected created_at not after updated_at on creation, got %v / %v", admin.CreatedAt, admin.UpdatedAt)
	}
	if admin.UpdatedAt.Sub(admin.CreatedAt) > time.Second {
		t.Errorf("expected created_at and updated_at to be within 1 second on creation, got %v / %v", admin.CreatedAt, admin.UpdatedAt)
	}
}
