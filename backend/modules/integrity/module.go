package integrity

import "github.com/industrix/backend/pkg/postgres"

// Module holds the integrity module's public components
type Module struct {
	Handler *Handler
	Service Service // also satisfies contracts.CompanyProvider
}

// NewModule wires internal dependencies and returns the module
func NewModule(pg *postgres.Client) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo)
	handler := NewHandler(svc)
	return &Module{
		Handler: handler,
		Service: svc,
	}
}
