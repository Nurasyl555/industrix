package media

import (
	"context"
	"time"

	"github.com/industrix/backend/pkg/logger"
	"github.com/industrix/backend/pkg/minio"
)

const bucketName = "equipment-media"

// Module holds the media module's public components.
type Module struct {
	Handler *Handler
}

// Config carries the two MinIO endpoints. They differ in Docker: the backend
// reaches MinIO over the internal docker-network host (minio:9000) for bucket
// operations, but presigned URLs must be signed for the host the BROWSER can
// reach (localhost:9000 in dev, the public S3 domain in prod) — the SigV4
// signature is bound to the host, so it can't just be string-swapped later.
type Config struct {
	InternalEndpoint string // e.g. minio:9000 — backend → minio
	PublicEndpoint   string // e.g. localhost:9000 — browser → minio
	AccessKey        string
	SecretKey        string
	UseSSL           bool
}

// NewModule wires the media module. It ensures the public bucket exists (via
// the internal client) and keeps a presign client bound to the public host.
func NewModule(cfg Config) (*Module, error) {
	log := logger.New("media-module")

	// Ops client — actually connects to MinIO to create/configure the bucket.
	ops, err := minio.NewClient(&minio.Config{
		Endpoint:  cfg.InternalEndpoint,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		UseSSL:    cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ops.EnsurePublicBucket(ctx, bucketName); err != nil {
		return nil, err
	}
	log.Info().Str("bucket", bucketName).Msg("media bucket ready")

	// Presign client — bound to the public host. Presigning is offline (no
	// network call), so it never needs to actually reach that host.
	presign, err := minio.NewClient(&minio.Config{
		Endpoint:  cfg.PublicEndpoint,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		UseSSL:    cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	scheme := "http"
	if cfg.UseSSL {
		scheme = "https"
	}
	publicBase := scheme + "://" + cfg.PublicEndpoint + "/" + bucketName

	return &Module{Handler: NewHandler(presign, publicBase)}, nil
}
