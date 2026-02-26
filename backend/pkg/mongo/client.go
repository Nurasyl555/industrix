package mongo

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/industrix/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config holds MongoDB connection configuration
type Config struct {
	URI                    string
	Database               string
	DirectConnect          bool
	ReplicaSet             string
	MaxPoolSize            uint64
	MinPoolSize            uint64
	MaxConnIdleTime        time.Duration
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
	HeartbeatInterval      time.Duration
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
		URI:                    getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		Database:               getEnv("MONGODB_DATABASE", "industrix"),
		DirectConnect:          false,
		ReplicaSet:             getEnv("MONGODB_REPLICA_SET", ""),
		MaxPoolSize:            100,
		MinPoolSize:            10,
		MaxConnIdleTime:        30 * time.Minute,
		ConnectTimeout:         10 * time.Second,
		ServerSelectionTimeout: 10 * time.Second,
		HeartbeatInterval:      5 * time.Second,
	}
}

// Client wraps mongo.Client and provides additional functionality
type Client struct {
	client   *mongo.Client
	database *mongo.Database
	log      *logger.Logger
}

// NewClient creates a new MongoDB client
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("mongo-client")

	clientOpts := options.Client().
		SetURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime).
		SetConnectTimeout(cfg.ConnectTimeout).
		SetServerSelectionTimeout(cfg.ServerSelectionTimeout).
		SetHeartbeatInterval(cfg.HeartbeatInterval)

	if cfg.ReplicaSet != "" {
		clientOpts.SetReplicaSet(cfg.ReplicaSet)
	}

	if cfg.DirectConnect {
		clientOpts.SetDirect(cfg.DirectConnect)
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(cfg.Database)

	log.Info().
		Str("uri", cfg.URI).
		Str("database", cfg.Database).
		Msg("MongoDB client connected")

	return &Client{
		client:   client,
		database: database,
		log:      log,
	}, nil
}

// Client returns the underlying MongoDB client
func (c *Client) Client() *mongo.Client {
	return c.client
}

// Database returns a handle to the configured database
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection returns a handle to a specific collection
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	if c.client != nil {
		err := c.client.Disconnect(ctx)
		if err != nil {
			return err
		}
		c.log.Info("MongoDB connection closed")
	}
	return nil
}

// HealthCheck verifies the MongoDB connection
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

// WithTimeout creates a context with timeout
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

// WithDeadline creates a context with deadline
func WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(parent, deadline)
}
