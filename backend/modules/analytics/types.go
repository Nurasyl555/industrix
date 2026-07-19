package analytics

import "time"

// Event is one recorded domain event in the analytics store.
type Event struct {
	EventType  string
	EntityID   string
	SellerID   string
	Amount     float64
	OccurredAt time.Time
}

// SellerStats is a seller's funnel over the requested window.
type SellerStats struct {
	Days              int     `json:"days"`
	ListingsPublished int     `json:"listings_published"`
	Inquiries         int     `json:"inquiries"`
	DealsCompleted    int     `json:"deals_completed"`
	DealsCancelled    int     `json:"deals_cancelled"`
	Revenue           float64 `json:"revenue"`
	// ConversionRate is completed deals / inquiries (0 when there are none).
	ConversionRate float64 `json:"conversion_rate"`
}

// AdminStats is the platform-wide view over the requested window.
type AdminStats struct {
	Days              int              `json:"days"`
	GMV               float64          `json:"gmv"`
	PaymentsCompleted int              `json:"payments_completed"`
	PaymentsRefunded  int              `json:"payments_refunded"`
	ListingsPublished int              `json:"listings_published"`
	Inquiries         int              `json:"inquiries"`
	DealsCompleted    int              `json:"deals_completed"`
	ActiveSellers     int              `json:"active_sellers"`
	EventsByType      map[string]int   `json:"events_by_type"`
	Daily             []DailyGMVBucket `json:"daily_gmv"`
}

// DailyGMVBucket is one day of gross merchandise value.
type DailyGMVBucket struct {
	Day string  `json:"day"` // YYYY-MM-DD
	GMV float64 `json:"gmv"`
}
