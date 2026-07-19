package engagement

import (
	"context"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the engagement module's public components.
type Module struct {
	Handler  *Handler
	Service  Service
	Consumer *Consumer // may be nil if Kafka is unavailable
}

// NewModule wires the watchlist/price-history service and, when brokers are
// supplied, the Kafka consumer that records price history and fires price-drop
// alerts from listing.price_changed events.
func NewModule(ctx context.Context, pg *postgres.Client, notifier contracts.Notifier, brokers []string) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, notifier)
	consumer := NewConsumer(ctx, svc, brokers, "industrix-engagement")
	return &Module{
		Handler:  NewHandler(svc),
		Service:  svc,
		Consumer: consumer,
	}
}
