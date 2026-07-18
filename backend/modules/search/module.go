package search

import (
	"context"

	"github.com/industrix/backend/pkg/logger"
	"github.com/industrix/backend/pkg/redis"
)

// Config holds the search module's external dependencies.
type Config struct {
	OpenSearchHosts    string
	OpenSearchUser     string
	OpenSearchPassword string
	Brokers            []string
	ConsumerGroup      string
}

// Module holds the search module's public components.
type Module struct {
	Handler  *Handler
	Service  Service
	Consumer *Consumer // may be nil if Kafka is unavailable
}

// NewModule wires the OpenSearch client, cache-backed service, HTTP handler and
// the Kafka consumer that syncs the index. Index creation and consumer setup
// are best-effort so the service still boots with a degraded search subsystem.
func NewModule(ctx context.Context, cfg Config, cache *redis.Client) *Module {
	log := logger.New("search-module")

	osClient := NewOpenSearchClient(cfg.OpenSearchHosts, cfg.OpenSearchUser, cfg.OpenSearchPassword)
	if err := osClient.EnsureIndex(ctx); err != nil {
		log.Warn().Err(err).Msg("could not ensure OpenSearch index — search may be degraded")
	}

	svc := NewService(osClient, cache)
	consumer := NewConsumer(ctx, svc, cfg.Brokers, cfg.ConsumerGroup)

	return &Module{
		Handler:  NewHandler(svc),
		Service:  svc,
		Consumer: consumer,
	}
}
