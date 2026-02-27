package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/industrix/backend/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     6379,
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}
}

type Client struct {
	client *redis.Client
	log    *logger.Logger
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	log := logger.New("redis-client")
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Info().Str("host", cfg.Host).Msg("Redis client connected")
	return &Client{client: rdb, log: log}, nil
}

func (c *Client) Close() error {
	if c.client != nil {
		err := c.client.Close()
		c.log.Info().Msg("Redis connection closed")
		return err
	}
	return nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}
func (c *Client) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, expiration).Result()
}

// KeyBuilder helper for generating standardized Redis keys
type KeyBuilder struct{}

func (k *KeyBuilder) Session(userID string) string {
	return fmt.Sprintf("session:%s", userID)
}

func (k *KeyBuilder) RateLimit(ip string, path string) string {
	return fmt.Sprintf("ratelimit:%s:%s", ip, path)
}
