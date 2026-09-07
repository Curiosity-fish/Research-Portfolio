package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisHealthChecker checks Redis connectivity.
type RedisHealthChecker struct {
	client *redis.Client
}

// NewRedisHealthChecker creates a new health checker for Redis.
func NewRedisHealthChecker(client *redis.Client) *RedisHealthChecker {
	return &RedisHealthChecker{client: client}
}

// HealthCheck verifies that Redis responds to a ping.
func (h *RedisHealthChecker) HealthCheck(ctx context.Context) error {
	if h.client == nil {
		return fmt.Errorf("redis client is nil")
	}
	return h.client.Ping(ctx).Err()
}
