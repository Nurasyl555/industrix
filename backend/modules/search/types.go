package search

// Doc is a single searchable listing document. It denormalises equipment
// fields (from equipment.* events) together with the active listing's price
// and type (from listing.* events), so a buyer's query hits one index without
// any cross-module joins. The OpenSearch document _id is the equipment id.
type Doc struct {
	EquipmentID string  `json:"equipment_id"`
	ListingID   string  `json:"listing_id,omitempty"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	CategoryID  string  `json:"category_id"`
	Region      string  `json:"region"`
	Condition   string  `json:"condition"`
	ImageURL    string  `json:"image_url,omitempty"`
	SellerID    string  `json:"seller_id"`
	ListingType string  `json:"listing_type,omitempty"`
	Price       float64 `json:"price,omitempty"`
	PricePeriod string  `json:"price_period,omitempty"`
	// Active is true only while a published listing exists for this equipment.
	// Search results are restricted to active docs.
	Active bool `json:"active"`
}

// Query captures the buyer-facing search parameters.
type Query struct {
	Text        string
	CategoryID  string
	Region      string
	Condition   string
	ListingType string
	PriceMin    float64
	PriceMax    float64
	Sort        string // "price_asc", "price_desc", "newest" (default: relevance)
	Page        int
	Limit       int
}

// Result is the response returned to the search endpoint.
type Result struct {
	Items  []Doc            `json:"items"`
	Total  int64            `json:"total"`
	Page   int              `json:"page"`
	Limit  int              `json:"limit"`
	Facets map[string]Facet `json:"facets"`
}

// Facet is the set of value→count buckets for one aggregatable field.
type Facet map[string]int64
