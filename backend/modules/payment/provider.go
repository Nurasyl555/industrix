package payment

import (
	"context"

	"github.com/google/uuid"

	"github.com/industrix/backend/pkg/logger"
)

// Provider abstracts an escrow-capable payment gateway. The escrow pattern is
// three phases: Hold captures and ring-fences the buyer's funds; Release pays
// them out to the seller; Refund returns them to the buyer.
//
// Real CIS-local providers (Kaspi Pay, Halyk Bank, Uzcard/Humo) plug in here.
type Provider interface {
	Name() string
	Hold(ctx context.Context, amount float64, currency string) (ref string, err error)
	Release(ctx context.Context, ref string) error
	Refund(ctx context.Context, ref string) error
}

// KaspiProvider is a stub implementation of the Kaspi Pay escrow flow. It does
// not talk to a real gateway — it returns a synthetic reference and logs each
// phase — so the module is end-to-end testable without live credentials. Swap
// the method bodies for real HTTP calls when integrating the Kaspi merchant API.
type KaspiProvider struct {
	log *logger.Logger
}

func NewKaspiProvider() *KaspiProvider {
	return &KaspiProvider{log: logger.New("kaspi-provider")}
}

func (p *KaspiProvider) Name() string { return "kaspi" }

func (p *KaspiProvider) Hold(_ context.Context, amount float64, currency string) (string, error) {
	ref := "kaspi_" + uuid.NewString()
	p.log.Info().Float64("amount", amount).Str("currency", currency).Str("ref", ref).Msg("escrow hold (stub)")
	return ref, nil
}

func (p *KaspiProvider) Release(_ context.Context, ref string) error {
	p.log.Info().Str("ref", ref).Msg("escrow release (stub)")
	return nil
}

func (p *KaspiProvider) Refund(_ context.Context, ref string) error {
	p.log.Info().Str("ref", ref).Msg("escrow refund (stub)")
	return nil
}
