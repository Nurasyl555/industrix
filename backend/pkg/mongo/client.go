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

type Config struct {
	URI            string
	ConnectTimeout time.Duration
	PingTimeout    time.Duration
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		URI:            getEnv("MONGO_URI", "mongodb://localhost:27017"),
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	}
}

type Client struct {
	client *mongo.Client
	log    *logger.Logger
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	log := logger.New("mongo-client")

	clientOptions := options.Client().ApplyURI(cfg.URI).
		SetConnectTimeout(cfg.ConnectTimeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Info().Str("uri", cfg.URI).Msg("MongoDB client connected")
	return &Client{client: client, log: log}, nil
}

func (c *Client) HealthCheck(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

func (c *Client) Database(name string) *mongo.Database {
	return c.client.Database(name)
}

func (c *Client) Close(ctx context.Context) error {
	if c.client != nil {
		c.log.Info().Msg("MongoDB connection closed")
		return c.client.Disconnect(ctx)
	}
	return nil
}
