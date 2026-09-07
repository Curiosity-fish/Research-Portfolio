//go:build integration

// Package integration verifies that golang-migrate migrations apply and roll
// back cleanly against a real PostgreSQL instance.
package integration

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/school-api/school-api-v1/internal/migrate"
)

func migrationPath(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}
	return filepath.Join(file, "..", "..", "..", "migrations")
}

func TestMigrations_UpAndDown(t *testing.T) {
	dbURL := os.Getenv("APP_DATABASE__URL")
	if dbURL == "" {
		t.Skip("APP_DATABASE__URL not set")
	}

	m, err := migrate.New(dbURL, migrationPath(t))
	if err != nil {
		t.Fatalf("create migrate runner: %v", err)
	}

	ctx := context.Background()

	// Ensure the database is clean even if the test fails partway through.
	t.Cleanup(func() {
		if err := m.Down(ctx); err != nil {
			t.Logf("cleanup rollback failed: %v", err)
		}
	})

	// Start from a clean state.
	if err := m.Down(ctx); err != nil {
		t.Fatalf("roll back migrations: %v", err)
	}

	if err := m.Up(ctx); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	tables := []string{"departments", "settings", "admin_users", "users", "groups", "user_tokens"}
	for _, table := range tables {
		var exists bool
		query := "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)"
		if err := pool.QueryRow(ctx, query, table).Scan(&exists); err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("expected table %s to exist after migrations", table)
		}
	}

	version, dirty, err := m.Version(ctx)
	if err != nil {
		t.Fatalf("get migration version: %v", err)
	}
	if dirty {
		t.Fatalf("expected migration version to be clean, got dirty at %d", version)
	}
	if version == 0 {
		t.Fatal("expected a non-zero migration version after applying migrations")
	}
	t.Logf("migration version: %d", version)

	if err := m.Down(ctx); err != nil {
		t.Fatalf("roll back migrations: %v", err)
	}

	for _, table := range tables {
		var exists bool
		query := "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)"
		if err := pool.QueryRow(ctx, query, table).Scan(&exists); err != nil {
			t.Fatalf("check table %s after rollback: %v", table, err)
		}
		if exists {
			t.Errorf("expected table %s to be dropped after rollback", table)
		}
	}

	finalVersion, _, err := m.Version(ctx)
	if err != nil {
		// After a full rollback the schema_migrations table is gone, which the
		// migrate library reports as "no migration". Treat that as version 0.
		if err.Error() != "no migration" {
			t.Fatalf("get final migration version: %v", err)
		}
		finalVersion = 0
	}
	if finalVersion != 0 {
		t.Errorf("expected migration version 0 after rollback, got %d", finalVersion)
	}
}
