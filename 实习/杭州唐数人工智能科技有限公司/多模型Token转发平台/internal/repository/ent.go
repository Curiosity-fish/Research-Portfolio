package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/school-api/school-api-v1/ent"
)

// OpenEntClient opens an Ent client using the pgx v5 driver.
// It pings the database immediately and configures a conservative connection
// pool so bad DATABASE_URLs fail fast at startup.
func OpenEntClient(dbURL string) (*ent.Client, error) {
	client, _, err := OpenEntClientWithDB(dbURL)
	return client, err
}

// OpenEntClientWithDB opens an Ent client and also returns the underlying
// *sql.DB so repositories that need raw SQL (e.g. report-style statistics)
// share the same connection pool instead of opening a second one.
func OpenEntClientWithDB(dbURL string) (*ent.Client, *sql.DB, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, nil, fmt.Errorf("open sql db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	return ent.NewClient(ent.Driver(drv)), db, nil
}
