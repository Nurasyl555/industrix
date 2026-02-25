package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis             *redis.Client
	requestsPerMinute int
	burst             int
}

func NewRateLimiter(redisClient *redis.Client, requestsPerMinute, burst int) *RateLimiter {
	return &RateLimiter{
		redis:             redisClient,
		requestsPerMinute: requestsPerMinute,
		burst:             burst,
	}
}

// SlidingWindow implements Redis-based sliding window rate limiter
func (rl *RateLimiter) SlidingWindow() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get unique key for rate limiting
		// Use userID if authenticated, otherwise use IP
		key := c.IP()
		if userID := GetUserID(c); userID != "" {
			key = fmt.Sprintf("ratelimit:user:%s", userID)
		} else {
			key = fmt.Sprintf("ratelimit:ip:%s", c.IP())
		}

		ctx := context.Background()

		// Get current timestamp
		now := time.Now().Unix()
		windowSize := int64(60) // 60 seconds window

		// Remove old entries outside the window
		rl.redis.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-windowSize))

		// Count requests in current window
		count, err := rl.redis.ZCard(ctx, key).Result()
		if err != nil {
			// Fail open - allow request if Redis fails
			return c.Next()
		}

		// Check if rate limit exceeded
		if count >= int64(rl.requestsPerMinute) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded",
				"retry_after": 60,
			})
		}

		// Add current request to the sorted set
		score := float64(now)
		member := redis.Z{Score: score, Member: fmt.Sprintf("%d-%s", now, c.IP())}
		rl.redis.ZAdd(ctx, key, member)

		// Set expiry on the key
		rl.redis.Expire(ctx, key, time.Duration(windowSize)*time.Second)

		// Set rate limit headers
		remaining := rl.requestsPerMinute - int(count) - 1
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.requestsPerMinute))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", now+windowSize))

		return c.Next()
	}
}

// SlidingWindowPerRoute implements per-route rate limiting
func (rl *RateLimiter) SlidingWindowPerRoute(route string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get unique key for rate limiting per route
		key := fmt.Sprintf("ratelimit:route:%s:%s", route, c.IP())
		if userID := GetUserID(c); userID != "" {
			key = fmt.Sprintf("ratelimit:route:%s:user:%s", route, userID)
		}

		ctx := context.Background()
		now := time.Now().Unix()
		windowSize := int64(60)

		rl.redis.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-windowSize))

		count, err := rl.redis.ZCard(ctx, key).Result()
		if err != nil {
			return c.Next()
		}

		if count >= int64(rl.requestsPerMinute) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded for this endpoint",
				"retry_after": 60,
			})
		}

		score := float64(now)
		member := redis.Z{Score: score, Member: fmt.Sprintf("%d-%s", now, c.IP())}
		rl.redis.ZAdd(ctx, key, member)
		rl.redis.Expire(ctx, key, time.Duration(windowSize)*time.Second)

		return c.Next()
	}
}

// AdminRateLimiter creates a rate limiter with higher limits for admin routes
func AdminRateLimiter(redisClient *redis.Client) fiber.Handler {
	adminLimiter := NewRateLimiter(redisClient, 120, 20)
	return adminLimiter.SlidingWindow()
}

// PublicRateLimiter creates a rate limiter for public routes
func PublicRateLimiter(redisClient *redis.Client) fiber.Handler {
	publicLimiter := NewRateLimiter(redisClient, 30, 5)
	return publicLimiter.SlidingWindow()
}
