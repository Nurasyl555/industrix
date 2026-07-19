package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/redis"
)

type RateLimitMiddleware struct {
	redisClient *redis.Client
	limit       int
	window      time.Duration
}

func NewRateLimit(client *redis.Client, limit int, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient: client,
		limit:       limit,
		window:      window,
	}
}

func (m *RateLimitMiddleware) SlidingWindow() fiber.Handler {
	return func(c *fiber.Ctx) error {
		identifier := c.IP()
		if userID := c.Locals("user_id"); userID != nil {
			identifier = userID.(string)
		}

		key := fmt.Sprintf("ratelimit:%s:%s", identifier, c.Path())

		count, err := m.redisClient.Incr(context.Background(), key)
		if err != nil {
			return c.Next() // Fail open
		}

		if count == 1 {
			// Best-effort TTL on the first hit; a failure just means the window
			// key lingers until Redis evicts it.
			_, _ = m.redisClient.Expire(context.Background(), key, m.window)
		}

		if count > int64(m.limit) {
			return c.Status(fiber.StatusTooManyRequests).SendString("Rate limit exceeded")
		}

		return c.Next()
	}
}
