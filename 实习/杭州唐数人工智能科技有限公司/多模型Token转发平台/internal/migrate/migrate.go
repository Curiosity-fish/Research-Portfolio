// Package migrate wraps golang-migrate for the school-api-v1 project.
// It provides a small, testable API around the migrate library so that the
// command-line entry point and integration tests share the same logic.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	gmigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Migrate wraps a migrate source/destination pair. Each method opens its own
// database connection so callers do not need to manage sql.DB lifecycle.
type Migrate struct {
	dbURL        string
	sourceURL    string
	databaseName string
}

// New creates a Migrate runner for the given database URL and migrations
// directory. The directory may be relative; it is converted to an absolute
// file:// URL for the migrate library.
func New(dbURL, migrationsPath string) (*Migrate, error) {
	if dbURL == "" {
		return nil, errors.New("database URL is required")
	}

	dbName, err := databaseName(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse database name: %w", err)
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("resolve migrations path: %w", err)
	}

	source := url.URL{Scheme: "file", Path: filepath.ToSlash(absPath)}

	return &Migrate{
		dbURL:        dbURL,
		sourceURL:    source.String(),
		databaseName: dbName,
	}, nil
}

// Up applies all pending migrations.
func (m *Migrate) Up(ctx context.Context) error {
	return m.run(ctx, func(mig *gmigrate.Migrate) error {
		if err := mig.Up(); err != nil && !errors.Is(err, gmigrate.ErrNoChange) {
			return err
		}
		return nil
	})
}

// Down rolls back all migrations.
func (m *Migrate) Down(ctx context.Context) error {
	return m.run(ctx, func(mig *gmigrate.Migrate) error {
		if err := mig.Down(); err != nil && !errors.Is(err, gmigrate.ErrNoChange) {
			return err
		}
		return nil
	})
}

// Steps applies N migrations forward (N > 0) or backward (N < 0).
func (m *Migrate) Steps(ctx context.Context, n int) error {
	return m.run(ctx, func(mig *gmigrate.Migrate) error {
		if err := mig.Steps(n); err != nil && !errors.Is(err, gmigrate.ErrNoChange) {
			return err
		}
		return nil
	})
}

// Version returns the current migration version and dirty flag.
func (m *Migrate) Version(ctx context.Context) (version uint, dirty bool, err error) {
	db, err := sql.Open("pgx", m.dbURL)
	if err != nil {
		return 0, false, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return 0, false, fmt.Errorf("ping database: %w", err)
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{DatabaseName: m.databaseName})
	if err != nil {
		return 0, false, fmt.Errorf("create migrate driver: %w", err)
	}

	mig, err := gmigrate.NewWithDatabaseInstance(m.sourceURL, m.databaseName, driver)
	if err != nil {
		return 0, false, fmt.Errorf("create migrate instance: %w", err)
	}
	defer mig.Close()

	return mig.Version()
}

func (m *Migrate) run(ctx context.Context, fn func(*gmigrate.Migrate) error) error {
	db, err := sql.Open("pgx", m.dbURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{DatabaseName: m.databaseName})
	if err != nil {
		return fmt.Errorf("create migrate driver: %w", err)
	}

	mig, err := gmigrate.NewWithDatabaseInstance(m.sourceURL, m.databaseName, driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer mig.Close()

	return fn(mig)
}

// databaseName extracts the database name from a PostgreSQL connection URL.
func databaseName(dbURL string) (string, error) {
	u, err := url.Parse(dbURL)
	if err != nil {
		return "", err
	}

	name := strings.TrimPrefix(u.Path, "/")
	if name == "" {
		return "", errors.New("database URL does not contain a database name")
	}

	// Strip query parameters that some tools append to the path.
	if i := strings.IndexAny(name, "?#"); i >= 0 {
		name = name[:i]
	}

	return name, nil
}
