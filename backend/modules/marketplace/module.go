package marketplace

import "github.com/industrix/backend/pkg/postgres"

// Module holds the marketplace module's public components
type Module struct {
	Handler *Handler
	Service Service
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
