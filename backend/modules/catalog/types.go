package catalog

import "time"

// Category represents an equipment category
type Category struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	ParentID *string `json:"parent_id,omitempty"`
}

// Equipment represents a piece of industrial equipment in the catalog
type Equipment struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	CategoryID  string    `json:"category_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Condition   string    `json:"condition"` // new, used
	Region      string    `json:"region"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateEquipmentRequest represents a request to add equipment to the catalog
type CreateEquipmentRequest struct {
	CategoryID  string `json:"category_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Condition   string `json:"condition"`
	Region      string `json:"region"`
	ImageURL    string `json:"image_url"`
}

// UpdateEquipmentRequest represents a request to update equipment details
type UpdateEquipmentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Condition   string `json:"condition"`
	Region      string `json:"region"`
	ImageURL    string `json:"image_url"`
}

// ListEquipmentFilter represents query filters for listing equipment
type ListEquipmentFilter struct {
	CategoryID string
	Region     string
	Search     string
	Page       int
	Limit      int
}
