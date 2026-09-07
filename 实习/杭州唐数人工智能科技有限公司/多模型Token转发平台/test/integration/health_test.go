//go:build integration

// Package integration verifies the full middleware -> handler -> service ->
// repository chain against real PostgreSQL and Redis instances. Run via
// docker-compose.test.yml (make test-integration), which provides the
// APP_DATABASE__URL and APP_REDIS__URL environment variables.
package integration

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/server/requestmeta"
	"github.com/school-api/school-api-v1/internal/service"
)

type healthResponse struct {
	Status     string `json:"status"`
	Components map[string]struct {
		Status string `json:"status"`
		Error  string `json:"error"`
	} `json:"components"`
}

// buildEngine wires the same middleware stack and /health route as main.go,
// backed by real PostgreSQL and Redis clients.
func buildEngine(t *testing.T, dbURL, redisURL string) *gin.Engine {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	t.Cleanup(dbPool.Close)

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse redis url: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	t.Cleanup(func() { redisClient.Close() })

	healthSvc := service.NewHealthService(map[string]domain.HealthChecker{
		"postgres": repository.NewPostgresHealthChecker(dbPool),
		"redis":    repository.NewRedisHealthChecker(redisClient),
	})
	healthHandler := handler.NewHealthHandler(healthSvc)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)
	engine.GET("/health", healthHandler.Health)
	return engine
}

func getHealth(t *testing.T, engine *gin.Engine) (*httptest.ResponseRecorder, healthResponse) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	var body healthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response %q: %v", recorder.Body.String(), err)
	}
	return recorder, body
}

func TestHealth_AllDependenciesUp(t *testing.T) {
	dbURL, redisURL := os.Getenv("APP_DATABASE__URL"), os.Getenv("APP_REDIS__URL")
	if dbURL == "" || redisURL == "" {
		t.Skip("APP_DATABASE__URL / APP_REDIS__URL not set")
	}

	engine := buildEngine(t, dbURL, redisURL)
	recorder, body := getHealth(t, engine)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if body.Status != service.HealthStatusOK {
		t.Errorf("expected status ok, got %s", body.Status)
	}
	for _, name := range []string{"postgres", "redis"} {
		component, ok := body.Components[name]
		if !ok {
			t.Fatalf("expected components to include %s", name)
		}
		if component.Status != service.ComponentStatusUp {
			t.Errorf("expected %s up, got %s", name, component.Status)
		}
	}
	// The RequestID middleware must stamp the response even on this path.
	if recorder.Header().Get(requestmeta.RequestIDHeader) == "" {
		t.Error("expected X-Request-ID response header")
	}
}

func TestHealth_RedisDown(t *testing.T) {
	dbURL := os.Getenv("APP_DATABASE__URL")
	if dbURL == "" {
		t.Skip("APP_DATABASE__URL not set")
	}

	engine := buildEngine(t, dbURL, "redis://localhost:1/0")
	recorder, body := getHealth(t, engine)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
	if body.Status != service.HealthStatusUnavailable {
		t.Errorf("expected status unavailable, got %s", body.Status)
	}
	if got := body.Components["postgres"].Status; got != service.ComponentStatusUp {
		t.Errorf("expected postgres up, got %s", got)
	}
	redisComponent, ok := body.Components["redis"]
	if !ok {
		t.Fatal("expected components to include redis")
	}
	if redisComponent.Status != service.ComponentStatusDown {
		t.Errorf("expected redis down, got %s", redisComponent.Status)
	}
	if redisComponent.Error == "" {
		t.Error("expected error message for redis")
	}
}
