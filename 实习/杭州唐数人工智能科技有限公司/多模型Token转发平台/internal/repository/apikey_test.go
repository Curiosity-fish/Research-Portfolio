package repository

import (
	"context"
	stdsql "database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/group"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/testutil"
	_ "modernc.org/sqlite"
)

// openTestClient returns an Ent client backed by an in-memory SQLite database.
// It uses the CGO-free modernc.org/sqlite driver so the test also passes in
// containers built with CGO_ENABLED=0.
func openTestClient(t *testing.T) *ent.Client {
	t.Helper()
	client, _ := openTestClientDB(t)
	return client
}

// openTestClientDB additionally returns the underlying *sql.DB so tests can
// backdate rows through raw SQL (created_at is immutable via Ent builders).
func openTestClientDB(t *testing.T) (*ent.Client, *stdsql.DB) {
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
	return client, db
}

func TestEntAPIKeyLookup_Success(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)

	u, g, _, plaintext := testutil.CreateTestAPIKey(t, ctx, client)
	lookup := NewEntAPIKeyLookup(client)

	result, err := lookup.GetByTokenHash(ctx, auth.HashAPIKey(plaintext))
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if result.UserID != u.ID {
		t.Errorf("expected user_id %s, got %s", u.ID, result.UserID)
	}
	if result.GroupID != g.ID {
		t.Errorf("expected group_id %s, got %s", g.ID, result.GroupID)
	}
	if result.GroupCode != g.Code {
		t.Errorf("expected group_code %s, got %s", g.Code, result.GroupCode)
	}
	if result.UserStatus != string(user.StatusActive) {
		t.Errorf("expected user status active, got %s", result.UserStatus)
	}
	if !result.IsEnabled {
		t.Error("expected token enabled")
	}
}

func TestEntAPIKeyLookup_NotFound(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)

	lookup := NewEntAPIKeyLookup(client)
	_, err := lookup.GetByTokenHash(ctx, auth.HashAPIKey("sk-notexist"))
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestEntAPIKeyLookup_WithExpiry(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)

	g, err := client.Group.Create().
		SetName("测试分组").
		SetCode("test-group-" + uuid.NewString()[:8]).
		SetStatus(group.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	u, err := client.User.Create().
		SetUsername("test-user-" + uuid.NewString()[:8]).
		SetPasswordHash("hash").
		SetName("测试用户").
		SetRole(user.RoleStudent).
		SetStatus(user.StatusActive).
		SetGroupID(g.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	plaintext, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}

	future := time.Now().UTC().Add(time.Hour)
	_, err = client.UserToken.Create().
		SetUserID(u.ID).
		SetName("default").
		SetTokenHash(hash).
		SetTokenLast4(plaintext[len(plaintext)-4:]).
		SetIsEnabled(true).
		SetExpiresAt(future).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user token: %v", err)
	}

	lookup := NewEntAPIKeyLookup(client)
	result, err := lookup.GetByTokenHash(ctx, auth.HashAPIKey(plaintext))
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if result.ExpiresAt == nil {
		t.Fatal("expected expires_at")
	}
	if !result.ExpiresAt.Equal(future) {
		t.Errorf("expected expires_at %v, got %v", future, *result.ExpiresAt)
	}
}
