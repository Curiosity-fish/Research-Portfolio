package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
)

const (
	loginRateLimitMax    int64         = 5
	loginRateLimitWindow time.Duration = time.Minute
	loginRateLimitPrefix               = "ratelimit:login:"
)

// loginRateLimiterScript atomically increments a per-key counter and sets its
// TTL on the first request in the window. It returns the current count.
var loginRateLimiterScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return current
`)

// LoginRateLimit limits login attempts per IP address using a Redis-backed
// fixed window. If Redis is unreachable the request is rejected to avoid
// opening a bypass path.
func LoginRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := loginRateLimitPrefix + c.ClientIP()

		allowed, err := allowLoginAttempt(c.Request.Context(), rdb, key)
		if err != nil || !allowed {
			respond.Error(c, domain.ErrTooMany)
			return
		}

		c.Next()
	}
}

func allowLoginAttempt(ctx context.Context, rdb *redis.Client, key string) (bool, error) {
	count, err := loginRateLimiterScript.Run(ctx, rdb, []string{key}, int64(loginRateLimitWindow.Seconds())).Int64()
	if err != nil {
		return false, fmt.Errorf("rate limit check: %w", err)
	}
	return count <= loginRateLimitMax, nil
}
