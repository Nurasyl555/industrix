package deal

import "time"

// Deal status values form a state machine (Phase 3). A deal starts as an
// inquiry and advances toward completion, or is cancelled from any live state.
// completed/cancelled are terminal.
const (
	StatusInquiry     = "inquiry"
	StatusNegotiation = "negotiation"
	StatusConfirmed   = "confirmed"
	StatusInProgress  = "in_progress"
	StatusCompleted   = "completed"
	StatusCancelled   = "cancelled"
)

// allowedTransitions maps each status to the states it may move to.
var allowedTransitions = map[string][]string{
	StatusInquiry:     {StatusNegotiation, StatusCancelled},
	StatusNegotiation: {StatusConfirmed, StatusCancelled},
	StatusConfirmed:   {StatusInProgress, StatusCancelled},
	StatusInProgress:  {StatusCompleted, StatusCancelled},
	StatusCompleted:   {},
	StatusCancelled:   {},
}

// canTransition reports whether a deal may move from one status to another.
func canTransition(from, to string) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// isTerminal reports whether a status admits no further transitions.
func isTerminal(status string) bool {
	return status == StatusCompleted || status == StatusCancelled
}

// Deal represents a buyer's inquiry on a listing that progresses through the
// deal state machine (see status constants above and docs/impl-plan.md Phase 3).
type Deal struct {
	ID        string `json:"id"`
	ListingID string `json:"listing_id"`
	BuyerID   string `json:"buyer_id"`
	SellerID  string `json:"seller_id"`
	Message   string `json:"message"`
	Status    string `json:"status"` // see status constants (inquiry … completed/cancelled)
	// Disputed freezes the deal while an arbitrator reviews an open dispute.
	Disputed  bool      `json:"disputed"`
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

// DealMessage is a single message in a deal's conversation thread.
type DealMessage struct {
	ID        string    `json:"id"`
	DealID    string    `json:"deal_id"`
	SenderID  string    `json:"sender_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// PostMessageRequest represents a reply within a deal thread.
type PostMessageRequest struct {
	Body string `json:"body"`
}

// TransitionRequest moves a deal to a new status in the state machine.
type TransitionRequest struct {
	Status string `json:"status"`
}
