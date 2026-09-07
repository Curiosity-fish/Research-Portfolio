package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func newRateLimitTestEngine(t *testing.T, rdb *redis.Client) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine := gin.New()
	engine.Use(Recovery(logger))
	engine.POST("/api/v1/admin/login", LoginRateLimit(rdb), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return engine
}

func parseRedisURL(t *testing.T) *redis.Client {
	t.Helper()
	redisURL := os.Getenv("APP_REDIS__URL")
	if redisURL == "" {
		t.Skip("APP_REDIS__URL not set")
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse redis url: %v", err)
	}
	return redis.NewClient(opts)
}

func TestLoginRateLimit_AllowsUnderLimit(t *testing.T) {
	rdb := parseRedisURL(t)
	ctx := context.Background()
	key := loginRateLimitPrefix + "10.0.0.1"
	t.Cleanup(func() { _ = rdb.Del(ctx, key).Err() })
	_ = rdb.Del(ctx, key)

	engine := newRateLimitTestEngine(t, rdb)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestLoginRateLimit_BlocksOverLimit(t *testing.T) {
	rdb := parseRedisURL(t)
	ctx := context.Background()
	key := loginRateLimitPrefix + "10.0.0.2"
	t.Cleanup(func() { _ = rdb.Del(ctx, key).Err() })
	_ = rdb.Del(ctx, key)

	// Seed the counter over the limit.
	if err := rdb.Set(ctx, key, loginRateLimitMax+1, loginRateLimitWindow).Err(); err != nil {
		t.Fatalf("seed rate limit key: %v", err)
	}

	engine := newRateLimitTestEngine(t, rdb)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.2")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}
}
