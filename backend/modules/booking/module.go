package booking

import (
	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the booking module's public components.
type Module struct {
	Handler *Handler
	Service Service
}

// NewModule wires the booking module. listings is the listing module's
// ListingProvider — booking validates rentals through the contract, never a
// direct import.
func NewModule(pg *postgres.Client, listings contracts.ListingProvider, notifier contracts.Notifier) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, listings, notifier)
	return &Module{Handler: NewHandler(svc), Service: svc}
}
