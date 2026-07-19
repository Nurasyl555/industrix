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

// planPrices is the fee per billing period, in KZT. Free costs nothing and is
// never charged or expired.
var planPrices = map[string]float64{
	PlanFree:       0,
	PlanBasic:      5000,
	PlanPro:        20000,
	PlanEnterprise: 50000,
}

// planPeriod is how long a paid plan lasts before it needs renewing.
const planPeriod = 30 * 24 * time.Hour

func priceFor(plan string) float64 { return planPrices[plan] }

var validPlans = map[string]bool{
	PlanFree: true, PlanBasic: true, PlanPro: true, PlanEnterprise: true,
}

// Subscription is a user's current tariff plan.
type Subscription struct {
	UserID       string     `json:"user_id"`
	Plan         string     `json:"plan"`
	Status       string     `json:"status"`
	Price        float64    `json:"price"`         // fee per period, KZT
	ListingLimit int        `json:"listing_limit"` // -1 = unlimited
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// PlanOption describes a plan on the pricing page.
type PlanOption struct {
	Plan         string  `json:"plan"`
	Price        float64 `json:"price"`
	ListingLimit int     `json:"listing_limit"`
}

// Plans lists every plan with its price and limit, cheapest first.
func Plans() []PlanOption {
	order := []string{PlanFree, PlanBasic, PlanPro, PlanEnterprise}
	out := make([]PlanOption, 0, len(order))
	for _, p := range order {
		out = append(out, PlanOption{Plan: p, Price: planPrices[p], ListingLimit: planListingLimits[p]})
	}
	return out
}

// ChangePlanRequest changes a user's subscription plan.
type ChangePlanRequest struct {
	Plan string `json:"plan"`
}
