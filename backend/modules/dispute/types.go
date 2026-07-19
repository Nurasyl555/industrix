package dispute

import "time"

// Dispute statuses. A dispute is opened by a participant and closed by an
// admin, whose decision also settles the escrow.
const (
	StatusOpen = "open"
	// The buyer was right: held funds go back to them.
	StatusResolvedRefund = "resolved_refund"
	// The seller was right: held funds are paid out.
	StatusResolvedRelease = "resolved_release"
	// The complaint had no merit; escrow is left as it is.
	StatusRejected = "rejected"
)

// resolutions maps an admin decision to the status it produces.
var resolutions = map[string]string{
	"refund":  StatusResolvedRefund,
	"release": StatusResolvedRelease,
	"reject":  StatusRejected,
}

// Dispute is a complaint about a deal, awaiting or carrying an arbitration
// decision.
type Dispute struct {
	ID             string    `json:"id"`
	DealID         string    `json:"deal_id"`
	FiledBy        string    `json:"filed_by"`
	Reason         string    `json:"reason"`
	EvidenceURLs   []string  `json:"evidence_urls"`
	Status         string    `json:"status"`
	ResolutionNote string    `json:"resolution_note,omitempty"`
	ResolvedBy     string    `json:"resolved_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// FileDisputeRequest opens a dispute on a deal.
type FileDisputeRequest struct {
	DealID       string   `json:"deal_id"`
	Reason       string   `json:"reason"`
	EvidenceURLs []string `json:"evidence_urls"`
}

// ResolveRequest is the admin's decision: refund, release or reject.
type ResolveRequest struct {
	Resolution string `json:"resolution"`
	Note       string `json:"note"`
}

// disputeEvent is the payload published on dispute.* topics.
type disputeEvent struct {
	ID        string `json:"id"`
	DealID    string `json:"deal_id"`
	FiledBy   string `json:"filed_by"`
	Status    string `json:"status"`
	DecidedBy string `json:"decided_by,omitempty"`
}
