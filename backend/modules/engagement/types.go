package engagement

import "time"

// Favorite is a user's watchlist entry for a listing.
type Favorite struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ListingID string    `json:"listing_id"`
	CreatedAt time.Time `json:"created_at"`
}

// FavoriteListing is a favorited listing joined with its display fields — what
// the watchlist page renders.
type FavoriteListing struct {
	ListingID   string    `json:"listing_id"`
	EquipmentID string    `json:"equipment_id"`
	Title       string    `json:"title"`
	ImageURL    string    `json:"image_url"`
	Price       float64   `json:"price"`
	PricePeriod string    `json:"price_period,omitempty"`
	ListingType string    `json:"listing_type"`
	Status      string    `json:"status"`
	FavoritedAt time.Time `json:"favorited_at"`
}

// PriceHistoryEntry is one recorded price change for a listing.
type PriceHistoryEntry struct {
	ListingID string    `json:"listing_id"`
	OldPrice  float64   `json:"old_price"`
	NewPrice  float64   `json:"new_price"`
	ChangedAt time.Time `json:"changed_at"`
}

// AddFavoriteRequest adds a listing to the caller's watchlist.
type AddFavoriteRequest struct {
	ListingID string `json:"listing_id"`
}
