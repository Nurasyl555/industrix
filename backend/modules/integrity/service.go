package integrity

import (
	"context"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
)

// Service defines the integrity service interface
type Service interface {
	CreateCompany(ctx context.Context, company *Company) error
	GetCompany(ctx context.Context, id string) (*Company, error)
	UpdateCompany(ctx context.Context, company *Company) error

	// Contracts
	contracts.CompanyProvider
}

type service struct {
	repo *Repository
}

// NewService creates a new integrity service
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateCompany(ctx context.Context, company *Company) error {
	if len(company.BIN) != 12 {
		return errors.New(errors.CodeValidation, "Invalid BIN format")
	}

	existing, err := s.repo.GetCompanyByBIN(ctx, company.BIN)
	if err == nil && existing != nil {
		return errors.New(errors.CodeConflict, "Company with this BIN already exists")
	}

	return s.repo.CreateCompany(ctx, company)
}

func (s *service) GetCompany(ctx context.Context, id string) (*Company, error) {
	return s.repo.GetCompanyByID(ctx, id)
}

func (s *service) UpdateCompany(ctx context.Context, company *Company) error {
	return s.repo.UpdateCompany(ctx, company)
}

// === Contracts (CompanyProvider) ===

func (s *service) GetCompanyBasic(ctx context.Context, companyID string) (*contracts.CompanyBasic, error) {
	c, err := s.repo.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return &contracts.CompanyBasic{
		ID:       c.ID,
		Name:     c.Name,
		Verified: c.Verified,
	}, nil
}
