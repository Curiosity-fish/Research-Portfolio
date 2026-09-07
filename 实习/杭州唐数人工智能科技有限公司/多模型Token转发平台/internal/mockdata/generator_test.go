package mockdata

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/internal/auth"
	_ "modernc.org/sqlite"

	stdsql "database/sql"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
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

func TestGenerator_CreateAdmin(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	g := NewGenerator(client, DefaultSeed)

	admin, err := g.CreateAdmin(ctx, "mockgen-admin", "MockGen@123", adminuser.RoleAdmin)
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if admin.Username != "mockgen-admin" {
		t.Errorf("expected username mockgen-admin, got %s", admin.Username)
	}
	if admin.Role != adminuser.RoleAdmin {
		t.Errorf("expected role admin, got %s", admin.Role)
	}
}

func TestGenerator_CreateGroups(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	g := NewGenerator(client, DefaultSeed)

	groups, err := g.CreateGroups(ctx, 3)
	if err != nil {
		t.Fatalf("create groups: %v", err)
	}
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	for _, gr := range groups {
		if !isTagged(gr.Code) {
			t.Errorf("group code %q is not tagged", gr.Code)
		}
	}
}

func TestGenerator_CreateDepartments(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	g := NewGenerator(client, DefaultSeed)

	deps, err := g.CreateDepartments(ctx, 2, 2)
	if err != nil {
		t.Fatalf("create departments: %v", err)
	}
	if len(deps) != 6 {
		t.Fatalf("expected 6 departments, got %d", len(deps))
	}

	var roots, children int
	for _, d := range deps {
		if !isTagged(d.Code) {
			t.Errorf("department code %q is not tagged", d.Code)
		}
		if d.Level == 0 {
			roots++
		} else if d.Level == 1 {
			children++
		} else {
			t.Errorf("unexpected department level %d", d.Level)
		}
	}
	if roots != 2 {
		t.Errorf("expected 2 roots, got %d", roots)
	}
	if children != 4 {
		t.Errorf("expected 4 children, got %d", children)
	}
}

func TestGenerator_CreateUsers(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	g := NewGenerator(client, DefaultSeed)

	groups, err := g.CreateGroups(ctx, 2)
	if err != nil {
		t.Fatalf("create groups: %v", err)
	}
	deps, err := g.CreateDepartments(ctx, 1, 1)
	if err != nil {
		t.Fatalf("create departments: %v", err)
	}

	users, err := g.CreateUsers(ctx, 5, groups, deps)
	if err != nil {
		t.Fatalf("create users: %v", err)
	}
	if len(users) != 5 {
		t.Fatalf("expected 5 users, got %d", len(users))
	}

	for _, u := range users {
		if !isTagged(u.Username) {
			t.Errorf("username %q is not tagged", u.Username)
		}
		if u.GroupID == nil {
			t.Error("expected user to have a group")
		}
		if u.DepartmentID == nil {
			t.Error("expected user to have a department")
		}
	}
}

func TestGenerator_CreateTokens(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	g := NewGenerator(client, DefaultSeed)

	users, err := g.CreateUsers(ctx, 2, nil, nil)
	if err != nil {
		t.Fatalf("create users: %v", err)
	}

	tokens, err := g.CreateTokens(ctx, 2, users)
	if err != nil {
		t.Fatalf("create tokens: %v", err)
	}
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}

	for _, tok := range tokens {
		if tok.UserID == uuid.Nil {
			t.Error("expected token user id")
		}
		if tok.TokenID == uuid.Nil {
			t.Error("expected token id")
		}
		if !isTagged(tok.Name) {
			t.Errorf("token name %q is not tagged", tok.Name)
		}
		if tok.Plaintext == "" {
			t.Error("expected non-empty plaintext token")
		}
		if !strings.HasPrefix(tok.Plaintext, auth.APIKeyPrefix) {
			t.Errorf("expected token to start with %q, got %q", auth.APIKeyPrefix, tok.Plaintext)
		}
	}
}

func TestGenerator_GenerateAndCleanup(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	g := NewGenerator(client, DefaultSeed)

	set, err := g.Generate(ctx, DefaultScenario())
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if set.Admin == nil {
		t.Fatal("expected admin")
	}
	if len(set.Groups) != DefaultScenario().GroupCount {
		t.Errorf("expected %d groups, got %d", DefaultScenario().GroupCount, len(set.Groups))
	}
	if len(set.Users) != DefaultScenario().UserCount {
		t.Errorf("expected %d users, got %d", DefaultScenario().UserCount, len(set.Users))
	}

	if err := g.Cleanup(ctx); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	if n, _ := client.AdminUser.Query().Count(ctx); n != 0 {
		t.Errorf("expected 0 admins after cleanup, got %d", n)
	}
	if n, _ := client.User.Query().Count(ctx); n != 0 {
		t.Errorf("expected 0 users after cleanup, got %d", n)
	}
	if n, _ := client.UserToken.Query().Count(ctx); n != 0 {
		t.Errorf("expected 0 tokens after cleanup, got %d", n)
	}
	if n, _ := client.Department.Query().Count(ctx); n != 0 {
		t.Errorf("expected 0 departments after cleanup, got %d", n)
	}
	if n, _ := client.Group.Query().Count(ctx); n != 0 {
		t.Errorf("expected 0 groups after cleanup, got %d", n)
	}
}

func TestMustBeNonProduction(t *testing.T) {
	cases := []struct {
		env  string
		want bool
	}{
		{"production", true},
		{"prod", true},
		{"Production", true},
		{"development", false},
		{"test", false},
		{"staging", false},
		{"", false},
	}
	for _, c := range cases {
		err := MustBeNonProduction(c.env)
		if c.want && err == nil {
			t.Errorf("expected error for env %q", c.env)
		}
		if !c.want && err != nil {
			t.Errorf("unexpected error for env %q: %v", c.env, err)
		}
	}
}

func TestDefaultScenario(t *testing.T) {
	cfg := DefaultScenario()
	if cfg.GroupCount <= 0 {
		t.Error("expected positive group count")
	}
	if cfg.UserCount <= 0 {
		t.Error("expected positive user count")
	}
}

func isTagged(s string) bool {
	return len(s) >= len(Prefix) && s[:len(Prefix)] == Prefix
}
