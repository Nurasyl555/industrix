package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/industrix/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	// Connection string takes precedence if set
	DSN         string
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
		DSN:         getEnv("DB_DSN", ""),
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
	cfg  *Config
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("postgres-client")

	var connStr string
	if cfg.DSN != "" {
		connStr = cfg.DSN
	} else {
		connStr = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
		)
	}

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
		Str("dsn_masked", maskDSN(connStr)).
		Msg("PostgreSQL client connected")

	return &Client{
		pool: pool,
		log:  log,
		cfg:  cfg,
	}, nil
}

func maskDSN(dsn string) string {
	// Simple masking for logging
	if len(dsn) > 10 {
		return dsn[:10] + "..."
	}
	return "..."
}

func (c *Client) Pool() *pgxpool.Pool { return c.pool }

func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
		c.log.Info().Msg("PostgreSQL connection pool closed")
	}
}

func (c *Client) HealthCheck(ctx context.Context) error { return c.pool.Ping(ctx) }

func (c *Client) RunMigrations(migrationsPath string) error {
	var connStr string
	if c.cfg.DSN != "" {
		connStr = c.cfg.DSN
	} else {
		connStr = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			c.cfg.Username, c.cfg.Password, c.cfg.Host, c.cfg.Port, c.cfg.Database,
		)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		connStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	c.log.Info().Str("path", migrationsPath).Msg("Migrations applied successfully")
	return nil
}

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
