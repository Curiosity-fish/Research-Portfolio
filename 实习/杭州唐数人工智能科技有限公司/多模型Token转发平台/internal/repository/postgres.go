package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresHealthChecker checks PostgreSQL connectivity.
type PostgresHealthChecker struct {
	pool *pgxpool.Pool
}

// NewPostgresHealthChecker creates a new health checker for PostgreSQL.
func NewPostgresHealthChecker(pool *pgxpool.Pool) *PostgresHealthChecker {
	return &PostgresHealthChecker{pool: pool}
}

// HealthCheck verifies that PostgreSQL responds to a ping.
func (h *PostgresHealthChecker) HealthCheck(ctx context.Context) error {
	if h.pool == nil {
		return fmt.Errorf("postgres pool is nil")
	}
	return h.pool.Ping(ctx)
}
