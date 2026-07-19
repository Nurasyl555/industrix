package analytics

import (
	"context"

	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the analytics module's public components.
type Module struct {
	Handler  *Handler
	Service  Service
	Consumer *Consumer // may be nil if Kafka is unavailable
}

// NewModule wires the dashboard service and, when brokers are supplied, the
// Kafka consumer that feeds the event store from every aggregated topic.
func NewModule(ctx context.Context, pg *postgres.Client, brokers []string) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo)
	consumer := NewConsumer(ctx, svc, brokers, "industrix-analytics")
	return &Module{
		Handler:  NewHandler(svc),
		Service:  svc,
		Consumer: consumer,
	}
}
