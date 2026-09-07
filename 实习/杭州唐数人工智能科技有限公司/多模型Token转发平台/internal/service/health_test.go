package service

import (
	"context"
	"testing"
	"time"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestHealthService_AllUp(t *testing.T) {
	svc := NewHealthService(map[string]domain.HealthChecker{
		"postgres": &testutil.HealthyChecker{},
		"redis":    &testutil.HealthyChecker{},
	})

	result := svc.Check(context.Background())

	if result.Status != HealthStatusOK {
		t.Fatalf("expected status ok, got %s", result.Status)
	}
	if len(result.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(result.Components))
	}
	for name, status := range result.Components {
		if status.Status != ComponentStatusUp {
			t.Errorf("expected component %s to be up, got %s", name, status.Status)
		}
	}
}

func TestHealthService_OneDown(t *testing.T) {
	svc := NewHealthService(map[string]domain.HealthChecker{
		"postgres": &testutil.HealthyChecker{},
		"redis":    &testutil.UnhealthyChecker{},
	})

	result := svc.Check(context.Background())

	if result.Status != HealthStatusUnavailable {
		t.Fatalf("expected status unavailable, got %s", result.Status)
	}
	if result.Components["postgres"].Status != ComponentStatusUp {
		t.Errorf("expected postgres up, got %s", result.Components["postgres"].Status)
	}
	if result.Components["redis"].Status != ComponentStatusDown {
		t.Errorf("expected redis down, got %s", result.Components["redis"].Status)
	}
	if result.Components["redis"].Error == "" {
		t.Error("expected error message for redis")
	}
}

func TestHealthService_AllDown(t *testing.T) {
	svc := NewHealthService(map[string]domain.HealthChecker{
		"postgres": &testutil.UnhealthyChecker{},
		"redis":    &testutil.UnhealthyChecker{},
	})

	result := svc.Check(context.Background())

	if result.Status != HealthStatusUnavailable {
		t.Fatalf("expected status unavailable, got %s", result.Status)
	}
	for name, status := range result.Components {
		if status.Status != ComponentStatusDown {
			t.Errorf("expected component %s to be down, got %s", name, status.Status)
		}
	}
}

func TestHealthService_PanicRecovered(t *testing.T) {
	svc := NewHealthService(map[string]domain.HealthChecker{
		"postgres": &testutil.HealthyChecker{},
		"redis":    &testutil.PanickingChecker{},
	})

	result := svc.Check(context.Background())

	if result.Status != HealthStatusUnavailable {
		t.Fatalf("expected status unavailable, got %s", result.Status)
	}
	if result.Components["redis"].Status != ComponentStatusDown {
		t.Errorf("expected redis down after panic, got %s", result.Components["redis"].Status)
	}
	if result.Components["redis"].Error != "health check panic" {
		t.Errorf("expected panic error message, got %s", result.Components["redis"].Error)
	}
}

func TestHealthService_Timeout(t *testing.T) {
	svc := NewHealthService(map[string]domain.HealthChecker{
		"postgres": &testutil.SlowChecker{Timeout: 10 * time.Second},
	})

	start := time.Now()
	result := svc.Check(context.Background())
	elapsed := time.Since(start)

	if result.Status != HealthStatusUnavailable {
		t.Fatalf("expected status unavailable, got %s", result.Status)
	}
	if result.Components["postgres"].Status != ComponentStatusDown {
		t.Errorf("expected postgres down due to timeout, got %s", result.Components["postgres"].Status)
	}
	if elapsed >= HealthCheckTimeout+time.Second {
		t.Errorf("expected timeout around %v, got %v", HealthCheckTimeout, elapsed)
	}
}
