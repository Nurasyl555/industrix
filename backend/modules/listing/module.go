package listing

import (
	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the listing module's public components
type Module struct {
	Handler *Handler
	Service Service
}

// NewModule wires internal dependencies and returns the module.
// equipment is the catalog module's EquipmentProvider — listing never
// imports catalog directly, only the shared contract. events is the platform
// event publisher — listing emits listing.* lifecycle events.
func NewModule(pg *postgres.Client, equipment contracts.EquipmentProvider, notifier contracts.Notifier, events contracts.EventPublisher) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, equipment, notifier, events)
	handler := NewHandler(svc)
	return &Module{
		Handler: handler,
		Service: svc,
	}
}
