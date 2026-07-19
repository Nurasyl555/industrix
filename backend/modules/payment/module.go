package payment

import (
	"context"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the payment module's public components.
type Module struct {
	Handler  *Handler
	Service  Service
	Consumer *Consumer // may be nil if Kafka is unavailable
}

// NewModule wires internal dependencies and returns the module. deals is the
// deal module's DealProvider (payment never imports deal directly, only the
// shared contract); events publishes payment.* topics; notifier emits
// user-facing payment updates. The escrow provider defaults to the Kaspi stub.
// When brokers are supplied, a Kafka consumer auto-settles escrow on
// deal.status.changed (release on completed, refund on cancelled).
func NewModule(ctx context.Context, pg *postgres.Client, deals contracts.DealProvider, events contracts.EventPublisher, notifier contracts.Notifier, brokers []string) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, deals, NewKaspiProvider(), events, notifier)
	consumer := NewConsumer(ctx, svc, brokers, "industrix-payment")
	return &Module{
		Handler:  NewHandler(svc),
		Service:  svc,
		Consumer: consumer,
	}
}
