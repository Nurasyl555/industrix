package deal

import (
	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/jwt"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the deal module's public components
type Module struct {
	Handler *Handler
	Service Service
}

// NewModule wires internal dependencies and returns the module.
// listings is the listing module's ListingProvider — deal never imports
// listing directly, only the shared contract. jwtClient is used to
// authenticate WebSocket upgrades from the access_token cookie.
func NewModule(pg *postgres.Client, listings contracts.ListingProvider, jwtClient jwt.Client, notifier contracts.Notifier, events contracts.EventPublisher) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, listings, notifier, events)
	handler := NewHandler(svc, NewHub(), jwtClient)
	return &Module{
		Handler: handler,
		Service: svc,
	}
}
