package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/service"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestHealthHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	healthSvc := service.NewHealthService(map[string]domain.HealthChecker{
		"postgres": &testutil.HealthyChecker{},
		"redis":    &testutil.HealthyChecker{},
	})
	handler := NewHealthHandler(healthSvc)
	engine.GET("/health", handler.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var body struct {
		Status     string `json:"status"`
		Components map[string]struct {
			Status string `json:"status"`
		} `json:"components"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if body.Status != service.HealthStatusOK {
		t.Errorf("expected status ok, got %s", body.Status)
	}
	if len(body.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(body.Components))
	}
}

func TestHealthHandler_Unavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	healthSvc := service.NewHealthService(map[string]domain.HealthChecker{
		"postgres": &testutil.HealthyChecker{},
		"redis":    &testutil.UnhealthyChecker{},
	})
	handler := NewHealthHandler(healthSvc)
	engine.GET("/health", handler.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}

	var body struct {
		Status     string `json:"status"`
		Components map[string]struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"components"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if body.Status != service.HealthStatusUnavailable {
		t.Errorf("expected status unavailable, got %s", body.Status)
	}
	// Even on 503 the per-component detail must be present so operators can
	// see which dependency failed.
	redis, ok := body.Components["redis"]
	if !ok {
		t.Fatal("expected components to include redis")
	}
	if redis.Status != service.ComponentStatusDown {
		t.Errorf("expected redis down, got %s", redis.Status)
	}
	if redis.Error == "" {
		t.Error("expected error message for redis")
	}
	if pg, ok := body.Components["postgres"]; !ok || pg.Status != service.ComponentStatusUp {
		t.Errorf("expected postgres up, got %+v", body.Components["postgres"])
	}
}
