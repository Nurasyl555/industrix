package dispute

import (
	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Module holds the dispute module's public components.
type Module struct {
	Handler *Handler
	Service Service
}

// NewModule wires the module. deals validates that the filer is party to the
// deal; escrow is the payment module's settler, used to move held funds the way
// arbitration decided — dispute never imports either package directly.
func NewModule(pg *postgres.Client, deals contracts.DealProvider, escrow contracts.EscrowSettler,
	events contracts.EventPublisher, notifier contracts.Notifier) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo, deals, escrow, events, notifier)
	return &Module{Handler: NewHandler(svc), Service: svc}
}
