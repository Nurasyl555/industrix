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
	"github.com/jackc/pgx/v5/tracelog"
)

// Config holds PostgreSQL connection configuration
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

// Client wraps pgxpool and provides additional functionality
type Client struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

// NewClient creates a new PostgreSQL client with connection pooling
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("postgres-client")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLife
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdle

	// Add logging tracer
	poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger: tracelog.NewStdLogger(log.Logger),
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
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

// Pool returns the underlying connection pool
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// Close closes the connection pool
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
		c.log.Info("PostgreSQL connection pool closed")
	}
}

// HealthCheck verifies the database connection
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

// Exec executes a query without returning rows
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return c.pool.Exec(ctx, query, args...)
}

// Query executes a query that returns rows
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return c.pool.Query(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return c.pool.QueryRow(ctx, query, args...)
}

// Begin starts a new transaction
func (c *Client) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.pool.Begin(ctx)
}

// BeginTx starts a new transaction with options
func (c *Client) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return c.pool.BeginTx(ctx, txOptions)
}

// CopyFrom copies data from a slice of values to a table
func (c *Client) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return c.pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

// Acquire gets a connection from the pool
func (c *Client) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return c.pool.Acquire(ctx)
}
