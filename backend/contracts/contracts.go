package contracts

import "context"

// UserProvider is implemented by the identity module, consumed by other modules
type UserProvider interface {
	GetUserBasic(ctx context.Context, userID string) (*UserBasic, error)
}

// CompanyProvider is implemented by the integrity module, consumed by other modules
type CompanyProvider interface {
	GetCompanyBasic(ctx context.Context, companyID string) (*CompanyBasic, error)
}

// EquipmentProvider is implemented by the catalog module, consumed by other modules
type EquipmentProvider interface {
	GetEquipmentBasic(ctx context.Context, equipmentID string) (*EquipmentBasic, error)
}

// ListingProvider is implemented by the listing module, consumed by other modules
type ListingProvider interface {
	GetListingBasic(ctx context.Context, listingID string) (*ListingBasic, error)
}

// UserBasic is a minimal user DTO for cross-module communication
type UserBasic struct {
	ID        string
	FirstName string
	LastName  string
	AvatarURL string
}

// CompanyBasic is a minimal company DTO for cross-module communication
type CompanyBasic struct {
	ID       string
	Name     string
	Verified bool
}

// EquipmentBasic is a minimal equipment DTO for cross-module communication
type EquipmentBasic struct {
	ID      string
	Title   string
	OwnerID string
}

// ListingBasic is a minimal listing DTO for cross-module communication
type ListingBasic struct {
	ID          string
	EquipmentID string
	SellerID    string
	Status      string
}
