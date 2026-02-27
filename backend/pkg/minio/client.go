package minio

import (
	"context"
	"os"
	"time"

	"github.com/industrix/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	BucketName string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists { return value }
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey: getEnv("MINIO_ACCESS_KEY", "minio"),
		SecretKey: getEnv("MINIO_SECRET_KEY", "minio123"),
		UseSSL:    false,
	}
}

type Client struct {
	client *minio.Client
	bucket string
	log    *logger.Logger
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil { cfg = DefaultConfig() }
	log := logger.New("minio-client")
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil { return nil, err }
	return &Client{client: client, bucket: cfg.BucketName, log: log}, nil
}

func (c *Client) PresignPutURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	u, err := c.client.PresignedPutObject(ctx, c.bucket, objectName, expiry)
	if err != nil { return "", err }
	return u.String(), nil
}

func (c *Client) PresignGetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	u, err := c.client.PresignedGetObject(ctx, c.bucket, objectName, expiry, nil)
	if err != nil { return "", err }
	return u.String(), nil
}
