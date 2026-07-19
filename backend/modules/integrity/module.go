package integrity

import (
	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the integrity module's public components
type Module struct {
	Handler *Handler
	Service Service // also satisfies contracts.CompanyProvider
}

// NewModule wires internal dependencies and returns the module. The billing
// dependency is injected later via Service.SetCharger — see the comment there.
func NewModule(pg *postgres.Client, notifier contracts.Notifier, events contracts.EventPublisher) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, notifier, events)
	handler := NewHandler(svc)
	return &Module{
		Handler: handler,
		Service: svc,
	}
}
