package minio

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/industrix/backend/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey: getEnv("MINIO_ROOT_USER", "minio"),
		SecretKey: getEnv("MINIO_ROOT_PASSWORD", "minio123"),
		UseSSL:    false,
	}
}

type Client struct {
	client *minio.Client
	log    *logger.Logger
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	log := logger.New("minio-client")

	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	log.Info().Str("endpoint", cfg.Endpoint).Msg("MinIO client created")
	return &Client{client: minioClient, log: log}, nil
}

func (c *Client) PresignPutURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	url, err := c.client.PresignedPutObject(ctx, bucketName, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}
	return url.String(), nil
}

func (c *Client) PresignGetURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	reqParams := make(map[string][]string)
	url, err := c.client.PresignedGetObject(ctx, bucketName, objectName, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}
	return url.String(), nil
}

func (c *Client) HealthCheck(ctx context.Context) error {
	// MinIO Go SDK doesn't have a direct ping, so we list buckets as a liveness check
	_, err := c.client.ListBuckets(ctx)
	return err
}
