package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/industrix/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config holds MinIO connection configuration
type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	BucketName string
	Location   string
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
		Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:  getEnv("MINIO_ACCESS_KEY", "minio"),
		SecretKey:  getEnv("MINIO_SECRET_KEY", "minio123"),
		UseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",
		BucketName: "",
		Location:   "kz-almaty-1",
	}
}

// Client wraps minio.Client and provides additional functionality
type Client struct {
	client *minio.Client
	bucket string
	log    *logger.Logger
}

// NewClient creates a new MinIO client
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("minio-client")

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Verify connection
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists && cfg.BucketName != "" {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{
			Region:        cfg.Location,
			ObjectLocking: false,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info().Str("bucket", cfg.BucketName).Msg("Bucket created")
	}

	log.Info().
		Str("endpoint", cfg.Endpoint).
		Str("bucket", cfg.BucketName).
		Msg("MinIO client connected")

	return &Client{
		client: client,
		bucket: cfg.BucketName,
		log:    log,
	}, nil
}

// Client returns the underlying MinIO client
func (c *Client) Client() *minio.Client {
	return c.client
}

// Bucket returns the configured bucket name
func (c *Client) Bucket() string {
	return c.bucket
}

// UploadFile uploads a file to MinIO
func (c *Client) UploadFile(ctx context.Context, objectName string, filePath string, contentType string) (minio.UploadInfo, error) {
	return c.client.FPutObject(ctx, c.bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
}

// UploadData uploads data to MinIO
func (c *Client) UploadData(ctx context.Context, objectName string, data []byte, contentType string) (minio.UploadInfo, error) {
	return c.client.PutObject(ctx, c.bucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
}

// DownloadFile downloads a file from MinIO
func (c *Client) DownloadFile(ctx context.Context, objectName string, filePath string) error {
	return c.client.FGetObject(ctx, c.bucket, objectName, filePath, minio.GetObjectOptions{})
}

// DownloadData downloads data from MinIO
func (c *Client) DownloadData(ctx context.Context, objectName string) ([]byte, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return io.ReadAll(obj)
}

// DeleteFile deletes a file from MinIO
func (c *Client) DeleteFile(ctx context.Context, objectName string) error {
	return c.client.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
}

// GetFileURL gets a public URL for a file
func (c *Client) GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	return c.client.PresignedGetObject(ctx, c.bucket, objectName, expiry, nil)
}

// PresignPutURL generates a presigned URL for PUT operations (upload)
func (c *Client) PresignPutURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	reqParams := make(minio.Map)
	return c.client.PresignedPutObject(ctx, c.bucket, objectName, expiry, reqParams)
}

// PresignGetURL generates a presigned URL for GET operations (download)
func (c *Client) PresignGetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	reqParams := make(minio.Map)
	return c.client.PresignedGetObject(ctx, c.bucket, objectName, expiry, reqParams)
}

// GetObject gets an object
func (c *Client) GetObject(ctx context.Context, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return c.client.GetObject(ctx, c.bucket, objectName, opts)
}

// PutObject puts an object
func (c *Client) PutObject(ctx context.Context, objectName string, reader *bytes.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return c.client.PutObject(ctx, c.bucket, objectName, reader, objectSize, opts)
}

// ListObjects lists objects
func (c *Client) ListObjects(ctx context.Context, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return c.client.ListObjects(ctx, c.bucket, opts)
}
