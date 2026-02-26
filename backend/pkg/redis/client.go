package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/industrix/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Config holds Redis connection configuration
type Config struct {
	Host           string
	Port           int
	Password       string
	DB             int
	PoolSize       int
	MinIdleConns   int
	MaxRetries     int
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	PoolTimeout    time.Duration
	SentinelMaster string
	SentinelAddrs  []string
	UseSentinel    bool
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Host:         getEnv("REDIS_HOST", "localhost"),
		Port:         6379,
		Password:     getEnv("REDIS_PASSWORD", ""),
		DB:           0,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  5 * time.Second,
		UseSentinel:  false,
	}
}

// Client wraps redis.Client and provides additional functionality
type Client struct {
	client *redis.Client
	log    *logger.Logger
}

// NewClient creates a new Redis client
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("redis-client")

	var rdb *redis.Client

	if cfg.UseSentinel {
		// Sentinel mode for high availability
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    cfg.SentinelMaster,
			SentinelAddrs: cfg.SentinelAddrs,
			Password:      cfg.Password,
			DB:            cfg.DB,
			PoolSize:      cfg.PoolSize,
			MinIdleConns:  cfg.MinIdleConns,
			MaxRetries:    cfg.MaxRetries,
			DialTimeout:   cfg.DialTimeout,
			ReadTimeout:   cfg.ReadTimeout,
			WriteTimeout:  cfg.WriteTimeout,
			PoolTimeout:   cfg.PoolTimeout,
		})
	} else {
		// Standalone mode
		rdb = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			PoolTimeout:  cfg.PoolTimeout,
		})
	}

	// Verify connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Int("db", cfg.DB).
		Msg("Redis client connected")

	return &Client{
		client: rdb,
		log:    log,
	}, nil
}

// Client returns the underlying Redis client
func (c *Client) Client() *redis.Client {
	return c.client
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if c.client != nil {
		err := c.client.Close()
		if err != nil {
			return err
		}
		c.log.Info("Redis connection closed")
	}
	return nil
}

// HealthCheck verifies the Redis connection
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// KeyHelpers provides typed key generation
type KeyHelpers struct {
	prefix string
}

// NewKeyHelpers creates a new KeyHelpers instance with a prefix
func NewKeyHelpers(prefix string) *KeyHelpers {
	return &KeyHelpers{prefix: prefix}
}

// UserKey generates user-related keys
func (k *KeyHelpers) UserKey(userID string) string {
	return fmt.Sprintf("%s:user:%s", k.prefix, userID)
}

// SessionKey generates session keys
func (k *KeyHelpers) SessionKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s", k.prefix, sessionID)
}

// TokenKey generates token keys (for refresh token storage)
func (k *KeyHelpers) TokenKey(tokenID string) string {
	return fmt.Sprintf("%s:token:%s", k.prefix, tokenID)
}

// CacheKey generates cache keys
func (k *KeyHelpers) CacheKey(prefix string, id string) string {
	return fmt.Sprintf("%s:cache:%s:%s", k.prefix, prefix, id)
}

// ListingKey generates listing-related keys
func (k *KeyHelpers) ListingKey(listingID string) string {
	return fmt.Sprintf("%s:listing:%s", k.prefix, listingID)
}

// SearchCacheKey generates search cache keys
func (k *KeyHelpers) SearchCacheKey(queryHash string) string {
	return fmt.Sprintf("%s:search:%s", k.prefix, queryHash)
}

// RateLimitKey generates rate limit keys
func (k *KeyHelpers) RateLimitKey(identifier string, window string) string {
	return fmt.Sprintf("%s:ratelimit:%s:%s", k.prefix, window, identifier)
}

// LockKey generates distributed lock keys
func (k *KeyHelpers) LockKey(resource string) string {
	return fmt.Sprintf("%s:lock:%s", k.prefix, resource)
}

// Setnx executes SETNX command with expiration
func (c *Client) Setnx(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// GetDel executes GET and DEL in a single command
func (c *Client) GetDel(ctx context.Context, key string) (string, error) {
	return c.client.GetDel(ctx, key).Result()
}

// IncrByFloat executes INCRBYFLOAT command
func (c *Client) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) {
	return c.client.IncrByFloat(ctx, key, value).Result()
}

// ExpireIfNotExists sets expiration only if key doesn't exist
func (c *Client) ExpireIfNotExists(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.client.ExpireIfNotExists(ctx, key, expiration).Result()
}

// ScanKeys iterates over keys matching a pattern
func (c *Client) ScanKeys(ctx context.Context, pattern string, count int64) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var batch []string
		var err error
		batch, cursor, err = c.client.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}
