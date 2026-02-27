package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/industrix/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host        string
	Port        int
	Database    string
	Username    string
	Password    string
	MaxConns    int32
	MinConns    int32
	MaxConnLife time.Duration
	MaxConnIdle time.Duration
	HealthCheck time.Duration
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		Host:        getEnv("POSTGRES_HOST", "localhost"),
		Port:        5432,
		Database:    "postgres",
		Username:    "postgres",
		Password:    "devpassword",
		MaxConns:    20,
		MinConns:    5,
		MaxConnLife: time.Hour,
		MaxConnIdle: 30 * time.Minute,
		HealthCheck: 30 * time.Second,
	}
}

type Client struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("postgres-client")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLife
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdle

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Database).
		Msg("PostgreSQL client connected")

	return &Client{
		pool: pool,
		log:  log,
	}, nil
}

func (c *Client) Pool() *pgxpool.Pool { return c.pool }

func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
		c.log.Info().Msg("PostgreSQL connection pool closed")
	}
}

func (c *Client) HealthCheck(ctx context.Context) error { return c.pool.Ping(ctx) }
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return c.pool.Exec(ctx, query, args...)
}
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return c.pool.Query(ctx, query, args...)
}
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return c.pool.QueryRow(ctx, query, args...)
}
func (c *Client) Begin(ctx context.Context) (pgx.Tx, error) { return c.pool.Begin(ctx) }
