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
	GetMyCompany(ctx context.Context, ownerID string) (*Company, error)
	UpdateCompany(ctx context.Context, company *Company) error

	// Admin
	ListCompaniesByStatus(ctx context.Context, status string) ([]*Company, error)
	SetCompanyStatus(ctx context.Context, id string, status CompanyStatus, note string) error

	// Contracts
	contracts.CompanyProvider
}

type service struct {
	repo     *Repository
	notifier contracts.Notifier
}

// NewService creates a new integrity service
func NewService(repo *Repository, notifier contracts.Notifier) Service {
	return &service{repo: repo, notifier: notifier}
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

func (s *service) GetMyCompany(ctx context.Context, ownerID string) (*Company, error) {
	return s.repo.GetCompanyByOwner(ctx, ownerID)
}

func (s *service) UpdateCompany(ctx context.Context, company *Company) error {
	return s.repo.UpdateCompany(ctx, company)
}

// === Admin ===

func (s *service) ListCompaniesByStatus(ctx context.Context, status string) ([]*Company, error) {
	return s.repo.ListByStatus(ctx, status)
}

func (s *service) SetCompanyStatus(ctx context.Context, id string, status CompanyStatus, note string) error {
	if status != StatusVerified && status != StatusRejected && status != StatusPending {
		return errors.New(errors.CodeValidation, "Invalid status")
	}
	company, err := s.repo.GetCompanyByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.SetStatus(ctx, id, status, note); err != nil {
		return err
	}
	if s.notifier != nil && company.OwnerID != "" {
		switch status {
		case StatusVerified:
			s.notifier.Notify(ctx, company.OwnerID, "company_verified", "Your company was verified", "/account/company")
		case StatusRejected:
			s.notifier.Notify(ctx, company.OwnerID, "company_rejected", "Your company registration was rejected", "/account/company")
		}
	}
	return nil
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
