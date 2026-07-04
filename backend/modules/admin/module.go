package admin

import (
	"github.com/industrix/backend/modules/integrity"
	"github.com/industrix/backend/modules/listing"
)

// Module holds the admin module's public components
type Module struct {
	Handler *Handler
}

// NewModule wires the admin aggregator over the domain services it moderates.
func NewModule(integritySvc integrity.Service, listingSvc listing.Service) *Module {
	return &Module{Handler: NewHandler(integritySvc, listingSvc)}
}
