package payment

import (
	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the payment module's public components.
type Module struct {
	Handler *Handler
	Service Service
}

// NewModule wires internal dependencies and returns the module. deals is the
// deal module's DealProvider (payment never imports deal directly, only the
// shared contract); events publishes payment.* topics; notifier emits
// user-facing payment updates. The escrow provider defaults to the Kaspi stub.
func NewModule(pg *postgres.Client, deals contracts.DealProvider, events contracts.EventPublisher, notifier contracts.Notifier) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, deals, NewKaspiProvider(), events, notifier)
	return &Module{
		Handler: NewHandler(svc),
		Service: svc,
	}
}
