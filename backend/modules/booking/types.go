package booking

import "time"

// Booking is a confirmed rental reservation of a listing for a date range.
type Booking struct {
	ID         string    `json:"id"`
	ListingID  string    `json:"listing_id"`
	RenterID   string    `json:"renter_id"`
	OwnerID    string    `json:"owner_id"`
	StartDate  string    `json:"start_date"` // YYYY-MM-DD
	EndDate    string    `json:"end_date"`   // YYYY-MM-DD
	Status     string    `json:"status"`     // confirmed | cancelled
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
}

// Quote is a price estimate for renting a listing over a date range, without
// creating a booking.
type Quote struct {
	ListingID string  `json:"listing_id"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Days      int     `json:"days"`
	Period    string  `json:"period"`
	UnitPrice float64 `json:"unit_price"`
	Units     int     `json:"units"`
	Total     float64 `json:"total"`
}

// DateRange is a compact booked-interval used to render availability.
type DateRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// CreateBookingRequest is a rental reservation request.
type CreateBookingRequest struct {
	ListingID string `json:"listing_id"`
	StartDate string `json:"start_date"` // YYYY-MM-DD
	EndDate   string `json:"end_date"`   // YYYY-MM-DD
}
