package catalog

import (
	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the catalog module's public components
type Module struct {
	Handler *Handler
	Service Service
}

// NewModule wires internal dependencies and returns the module. events is the
// platform event publisher — catalog emits equipment.* events for search
// indexing and analytics.
func NewModule(pg *postgres.Client, events contracts.EventPublisher) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, events)
	handler := NewHandler(svc)
	return &Module{
		Handler: handler,
		Service: svc,
	}
}
