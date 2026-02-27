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
