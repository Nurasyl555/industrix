package integrity

import "time"

// Subscription plans. Absence of a row implies PlanFree.
const (
	PlanFree       = "free"
	PlanBasic      = "basic"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"
)

// planListingLimits maps each plan to its max number of live/pending listings.
// -1 means unlimited.
var planListingLimits = map[string]int{
	PlanFree:       3,
	PlanBasic:      10,
	PlanPro:        50,
	PlanEnterprise: -1,
}

// listingLimitFor returns the listing cap for a plan (free's cap for unknown).
func listingLimitFor(plan string) int {
	if limit, ok := planListingLimits[plan]; ok {
		return limit
	}
	return planListingLimits[PlanFree]
}

var validPlans = map[string]bool{
	PlanFree: true, PlanBasic: true, PlanPro: true, PlanEnterprise: true,
}

// Subscription is a user's current tariff plan.
type Subscription struct {
	UserID       string     `json:"user_id"`
	Plan         string     `json:"plan"`
	Status       string     `json:"status"`
	ListingLimit int        `json:"listing_limit"` // -1 = unlimited
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ChangePlanRequest changes a user's subscription plan.
type ChangePlanRequest struct {
	Plan string `json:"plan"`
}
