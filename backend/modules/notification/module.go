package notification

import "github.com/industrix/backend/pkg/postgres"

// Module holds the notification module's public components.
type Module struct {
	Handler *Handler
	Service Service // also satisfies contracts.Notifier for other modules
}

func NewModule(pg *postgres.Client) *Module {
	repo := NewRepository(pg)
	svc := NewService(repo)
	return &Module{Handler: NewHandler(svc), Service: svc}
}
