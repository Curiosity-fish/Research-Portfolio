package service

import (
	"context"
	"sync"
	"time"

	"github.com/school-api/school-api-v1/internal/domain"
)

const (
	// HealthStatusOK indicates the service and all dependencies are healthy.
	HealthStatusOK = "ok"
	// HealthStatusUnavailable indicates at least one dependency is unhealthy.
	HealthStatusUnavailable = "unavailable"
	// ComponentStatusUp indicates a dependency is healthy.
	ComponentStatusUp = "up"
	// ComponentStatusDown indicates a dependency is unhealthy.
	ComponentStatusDown = "down"

	// HealthCheckTimeout is the maximum time allowed for all health checks.
	HealthCheckTimeout = 5 * time.Second
)

// HealthStatus describes the state of a single dependency.
type HealthStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthResult is the aggregate health response.
type HealthResult struct {
	Status     string                  `json:"status"`
	Components map[string]HealthStatus `json:"components"`
}

// HealthService checks the health of all registered dependencies.
type HealthService struct {
	checkers map[string]domain.HealthChecker
}

// NewHealthService creates a HealthService with the provided checkers.
func NewHealthService(checkers map[string]domain.HealthChecker) *HealthService {
	return &HealthService{checkers: checkers}
}

// Check runs all health checks concurrently and returns the aggregate result.
func (s *HealthService) Check(ctx context.Context) *HealthResult {
	ctx, cancel := context.WithTimeout(ctx, HealthCheckTimeout)
	defer cancel()

	result := &HealthResult{
		Status:     HealthStatusOK,
		Components: make(map[string]HealthStatus, len(s.checkers)),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, checker := range s.checkers {
		wg.Add(1)
		go func(name string, checker domain.HealthChecker) {
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					result.Components[name] = HealthStatus{
						Status: ComponentStatusDown,
						Error:  "health check panic",
					}
					result.Status = HealthStatusUnavailable
					mu.Unlock()
				}
				wg.Done()
			}()

			status := HealthStatus{Status: ComponentStatusUp}
			if err := checker.HealthCheck(ctx); err != nil {
				status = HealthStatus{Status: ComponentStatusDown, Error: "dependency check failed"}
			}

			mu.Lock()
			result.Components[name] = status
			if status.Status == ComponentStatusDown {
				result.Status = HealthStatusUnavailable
			}
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()
	return result
}
