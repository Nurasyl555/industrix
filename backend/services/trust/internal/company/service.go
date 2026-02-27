package company

import (
	"context"
	"encoding/json"
	"time"

	"github.com/industrix/pkg/errors"
	"github.com/industrix/pkg/kafka"
)

type Company struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BIN       string `json:"bin"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Website   string `json:"website"`
	Verified  bool   `json:"verified"`
	CreatedAt string `json:"created_at"`
	OwnerID   string `json:"owner_id"`
}

type Repository interface {
	CreateCompany(ctx context.Context, company *Company) error
	GetCompanyByID(ctx context.Context, id string) (*Company, error)
	GetCompanyByBIN(ctx context.Context, bin string) (*Company, error)
	UpdateCompany(ctx context.Context, company *Company) error
}

type Service interface {
	CreateCompany(ctx context.Context, company *Company) error
	GetCompany(ctx context.Context, id string) (*Company, error)
	UpdateCompany(ctx context.Context, company *Company) error
}

type service struct {
	repo     Repository
	producer kafka.Producer
}

func NewService(repo Repository, producer kafka.Producer) Service {
	return &service{repo: repo, producer: producer}
}

func (s *service) CreateCompany(ctx context.Context, company *Company) error {
	// Validate BIN format (12 digits)
	if len(company.BIN) != 12 {
		return errors.New(errors.CodeValidation, "Invalid BIN format")
	}

	// Check for duplicates
	existing, err := s.repo.GetCompanyByBIN(ctx, company.BIN)
	if err == nil && existing != nil {
		return errors.New(errors.CodeConflict, "Company with this BIN already exists")
	}

	if err := s.repo.CreateCompany(ctx, company); err != nil {
		return err
	}

	return nil
}

func (s *service) GetCompany(ctx context.Context, id string) (*Company, error) {
	return s.repo.GetCompanyByID(ctx, id)
}

func (s *service) UpdateCompany(ctx context.Context, company *Company) error {
	if err := s.repo.UpdateCompany(ctx, company); err != nil {
		return err
	}

	// If verified status changed, emit event
	if company.Verified && s.producer != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"company_id": company.ID,
			"verified":   true,
			"timestamp":  time.Now(),
		})
		s.producer.SendMessage(ctx, "company.verified", []byte(company.ID), payload)
	}

	return nil
}
