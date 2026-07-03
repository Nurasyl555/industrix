package deal

import "time"

// Deal represents a buyer's inquiry on a listing. Deliberately minimal for
// MVP — no state machine, no escrow/payments (see docs/impl-plan.md Phase 3
// for the full scope this grows into).
type Deal struct {
	ID        string    `json:"id"`
	ListingID string    `json:"listing_id"`
	BuyerID   string    `json:"buyer_id"`
	SellerID  string    `json:"seller_id"`
	Message   string    `json:"message"`
	Status    string    `json:"status"` // inquiry, closed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DealView is a Deal annotated with the caller's role in it.
type DealView struct {
	Deal
	Role string `json:"role"` // "buyer" or "seller", relative to the caller
}

// CreateDealRequest represents a request to inquire about a listing
type CreateDealRequest struct {
	ListingID string `json:"listing_id"`
	Message   string `json:"message"`
}
