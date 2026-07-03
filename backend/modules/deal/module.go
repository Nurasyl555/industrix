package deal

import (
	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the deal module's public components
type Module struct {
	Handler *Handler
	Service Service
}

// NewModule wires internal dependencies and returns the module.
// listings is the listing module's ListingProvider — deal never imports
// listing directly, only the shared contract.
func NewModule(pg *postgres.Client, listings contracts.ListingProvider) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, listings)
	handler := NewHandler(svc)
	return &Module{
		Handler: handler,
		Service: svc,
	}
}
