package notification

import (
	"context"

	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the notification module's public components.
type Module struct {
	Handler  *Handler
	Service  Service   // also satisfies contracts.Notifier for other modules
	Consumer *Consumer // may be nil if Kafka is unavailable
}

// NewModule wires the notification service and, when brokers are supplied, the
// Kafka consumer for notification.dispatch multi-channel fan-out.
func NewModule(ctx context.Context, pg *postgres.Client, brokers []string, groupID string) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo)
	consumer := NewConsumer(ctx, svc, brokers, groupID)
	return &Module{Handler: NewHandler(svc), Service: svc, Consumer: consumer}
}
