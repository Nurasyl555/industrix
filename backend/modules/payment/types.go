package payment

import "time"

// Payment status values model a simple escrow lifecycle:
//
//	pending -> held -> released   (happy path: funds captured then paid out)
//	pending -> failed             (provider hold failed)
//	held    -> refunded           (deal fell through, money returned to buyer)
const (
	StatusPending  = "pending"
	StatusHeld     = "held"
	StatusReleased = "released"
	StatusRefunded = "refunded"
	StatusFailed   = "failed"
)

// Payment kinds. Escrow is held then released to a seller; a subscription
// charge is captured immediately and has no deal or payee.
const (
	KindEscrow       = "escrow"
	KindSubscription = "subscription"
)

// Payment is a money movement. For escrow it is tied to a deal: the buyer
// (payer) funds it and the seller (payee) receives it on release. For a
// subscription charge there is no deal and no payee — the platform is paid.
type Payment struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	Description string    `json:"description,omitempty"`
	DealID      string    `json:"deal_id,omitempty"`
	PayerID     string    `json:"payer_id"`
	PayeeID     string    `json:"payee_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Provider    string    `json:"provider"`
	Status      string    `json:"status"`
	ProviderRef string    `json:"provider_ref,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreatePaymentRequest funds a deal into escrow.
type CreatePaymentRequest struct {
	DealID   string  `json:"deal_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"` // defaults to KZT
}

// paymentEvent is the payload published on payment.* topics.
type paymentEvent struct {
	ID       string  `json:"id"`
	DealID   string  `json:"deal_id"`
	PayerID  string  `json:"payer_id"`
	PayeeID  string  `json:"payee_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
}

func toEvent(p *Payment) paymentEvent {
	return paymentEvent{
		ID: p.ID, DealID: p.DealID, PayerID: p.PayerID, PayeeID: p.PayeeID,
		Amount: p.Amount, Currency: p.Currency, Status: p.Status,
	}
}
