// Package testutil provides shared test helpers used across packages.
package testutil

import (
	"context"
	"errors"
	"time"
)

// HealthyChecker is a domain.HealthChecker that always succeeds.
type HealthyChecker struct{}

// HealthCheck always returns nil.
func (h *HealthyChecker) HealthCheck(ctx context.Context) error {
	return nil
}

// UnhealthyChecker is a domain.HealthChecker that always fails.
type UnhealthyChecker struct{}

// HealthCheck always returns an error.
func (h *UnhealthyChecker) HealthCheck(ctx context.Context) error {
	return errors.New("connection refused")
}

// PanickingChecker is a domain.HealthChecker that always panics.
type PanickingChecker struct{}

// HealthCheck always panics.
func (h *PanickingChecker) HealthCheck(ctx context.Context) error {
	panic("health check panic")
}

// SlowChecker is a domain.HealthChecker that blocks until the context is done
// or the timeout elapses, whichever comes first.
type SlowChecker struct{ Timeout time.Duration }

// HealthCheck blocks until ctx is cancelled or the configured timeout passes.
func (h *SlowChecker) HealthCheck(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(h.Timeout):
		return nil
	}
}
