package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/pkg/redis"
)

type RateLimitMiddleware struct {
	redisClient *redis.Client
	limit       int           // Requests per window
	window      time.Duration // Window duration
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
		// Identify by IP or UserID if available
		identifier := c.IP()
		if userID := c.Locals("user_id"); userID != nil {
			identifier = userID.(string)
		}

		key := fmt.Sprintf("ratelimit:%s:%s", identifier, c.Path())

		// Use redis generic client or specific rate limit logic
		// Simplified: increment counter with expiration
		count, err := m.redisClient.Incr(context.Background(), key)
		if err != nil {
			return c.Next() // Fail open
		}

		if count == 1 {
			m.redisClient.Expire(context.Background(), key, m.window)
		}

		if count > int64(m.limit) {
			return c.Status(fiber.StatusTooManyRequests).SendString("Rate limit exceeded")
		}

		return c.Next()
	}
}
