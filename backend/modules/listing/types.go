package listing

import "time"

// Listing represents a sale/rental listing for a piece of equipment
type Listing struct {
	ID          string    `json:"id"`
	EquipmentID string    `json:"equipment_id"`
	SellerID    string    `json:"seller_id"`
	ListingType string    `json:"listing_type"` // sale, rental
	Price       float64   `json:"price"`
	PricePeriod string    `json:"price_period,omitempty"` // day, week, month — rental only
	Status      string    `json:"status"`                 // draft, active, archived
	ViewCount   int       `json:"view_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListingView is a Listing joined with its equipment — what buyers browse.
type ListingView struct {
	ID          string    `json:"id"`
	EquipmentID string    `json:"equipment_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CategoryID  string    `json:"category_id"`
	Region      string    `json:"region"`
	Condition   string    `json:"condition"`
	ImageURL    string    `json:"image_url"`
	SellerID    string    `json:"seller_id"`
	ListingType string    `json:"listing_type"`
	Price       float64   `json:"price"`
	PricePeriod string    `json:"price_period,omitempty"`
	Status      string    `json:"status"`
	ViewCount   int       `json:"view_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateListingRequest represents a request to list equipment for sale/rental
type CreateListingRequest struct {
	EquipmentID string  `json:"equipment_id"`
	ListingType string  `json:"listing_type"`
	Price       float64 `json:"price"`
	PricePeriod string  `json:"price_period"`
}

// UpdateListingRequest represents a request to update price/type
type UpdateListingRequest struct {
	Price       float64 `json:"price"`
	PricePeriod string  `json:"price_period"`
}

// ListListingsFilter represents query filters for browsing active listings
type ListListingsFilter struct {
	CategoryID  string
	Region      string
	ListingType string
	Condition   string
	Search      string
	PriceMin    float64
	PriceMax    float64
	Sort        string // "price_asc", "price_desc", "newest" (default)
	Page        int
	Limit       int
}
