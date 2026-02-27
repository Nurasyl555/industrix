package identity

import (
	"github.com/industrix/backend/pkg/jwt"
	"github.com/industrix/backend/pkg/postgres"
	"github.com/industrix/backend/pkg/redis"
)

// Module holds the identity module's public components
type Module struct {
	Handler *Handler
	Service Service // also satisfies contracts.UserProvider
}

// NewModule wires internal dependencies and returns the module
func NewModule(pg *postgres.Client, redis *redis.Client, jwtClient jwt.Client) *Module {
	repo := NewRepository(pg, redis)
	svc := NewService(repo, jwtClient)
	handler := NewHandler(svc)
	return &Module{
		Handler: handler,
		Service: svc,
	}
}
