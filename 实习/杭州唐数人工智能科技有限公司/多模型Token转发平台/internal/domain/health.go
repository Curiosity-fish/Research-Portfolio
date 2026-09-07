package domain

import "context"

// HealthChecker defines a dependency that can report its health status.
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}
