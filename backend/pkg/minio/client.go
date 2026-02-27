package minio

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/industrix/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minio"),
		SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minio123"),
		UseSSL:          false,
		Region:          getEnv("MINIO_REGION", "us-east-1"),
	}
}

type Client struct {
	client *minio.Client
	config *Config
	log    *logger.Logger
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("minio-client")

	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Test connection
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO: %w", err)
	}

	log.Info().
		Str("endpoint", cfg.Endpoint).
		Msg("MinIO client connected")

	return &Client{
		client: minioClient,
		config: cfg,
		log:    log,
	}, nil
}

func (c *Client) Client() *minio.Client {
	return c.client
}

func (c *Client) PresignPutURL(ctx context.Context, bucketName, objectName string, expires time.Duration) (string, error) {
	url, err := c.client.PresignedPutObject(ctx, bucketName, objectName, expires)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("bucket", bucketName).
			Str("object", objectName).
			Msg("Failed to generate presigned PUT URL")
		return "", err
	}
	return url.String(), nil
}

func (c *Client) PresignGetURL(ctx context.Context, bucketName, objectName string, expires time.Duration) (string, error) {
	url, err := c.client.PresignedGetObject(ctx, bucketName, objectName, expires, nil)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("bucket", bucketName).
			Str("object", objectName).
			Msg("Failed to generate presigned GET URL")
		return "", err
	}
	return url.String(), nil
}

func (c *Client) CreateBucket(ctx context.Context, bucketName string) error {
	err := c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: c.config.Region})
	if err != nil {
		c.log.Error().
			Err(err).
			Str("bucket", bucketName).
			Msg("Failed to create bucket")
		return err
	}
	c.log.Info().
		Str("bucket", bucketName).
		Msg("Bucket created")
	return nil
}

func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("bucket", bucketName).
			Msg("Failed to check bucket existence")
		return false, err
	}
	return exists, nil
}
