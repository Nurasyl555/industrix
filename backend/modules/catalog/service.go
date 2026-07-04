package catalog

import (
	"context"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
)

// Service defines the catalog service interface
type Service interface {
	ListCategories(ctx context.Context) ([]*Category, error)

	CreateEquipment(ctx context.Context, ownerID string, req CreateEquipmentRequest) (*Equipment, error)
	GetEquipment(ctx context.Context, id string) (*Equipment, error)
	ListEquipment(ctx context.Context, f ListEquipmentFilter) ([]*Equipment, int64, error)
	UpdateEquipment(ctx context.Context, id, ownerID string, req UpdateEquipmentRequest) (*Equipment, error)
	DeleteEquipment(ctx context.Context, id, ownerID string) error

	// Contracts
	contracts.EquipmentProvider
}

type service struct {
	repo *Repository
}

// NewService creates a new catalog service
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

var validConditions = map[string]bool{"new": true, "used": true}

func (s *service) ListCategories(ctx context.Context) ([]*Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *service) CreateEquipment(ctx context.Context, ownerID string, req CreateEquipmentRequest) (*Equipment, error) {
	if req.Title == "" {
		return nil, errors.New(errors.CodeValidation, "Title is required")
	}
	if req.CategoryID == "" {
		return nil, errors.New(errors.CodeValidation, "Category is required")
	}
	if req.Condition == "" {
		req.Condition = "used"
	}
	if !validConditions[req.Condition] {
		return nil, errors.New(errors.CodeValidation, "Condition must be 'new' or 'used'")
	}

	exists, err := s.repo.CategoryExists(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New(errors.CodeValidation, "Category does not exist")
	}

	eq := &Equipment{
		OwnerID:     ownerID,
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Description: req.Description,
		Condition:   req.Condition,
		Region:      req.Region,
		ImageURL:    req.ImageURL,
	}
	if err := s.repo.CreateEquipment(ctx, eq); err != nil {
		return nil, err
	}
	return eq, nil
}

func (s *service) GetEquipment(ctx context.Context, id string) (*Equipment, error) {
	return s.repo.GetEquipmentByID(ctx, id)
}

func (s *service) ListEquipment(ctx context.Context, f ListEquipmentFilter) ([]*Equipment, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.repo.ListEquipment(ctx, f)
}

func (s *service) UpdateEquipment(ctx context.Context, id, ownerID string, req UpdateEquipmentRequest) (*Equipment, error) {
	eq, err := s.repo.GetEquipmentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if eq.OwnerID != ownerID {
		return nil, errors.New(errors.CodeUnauthorized, "You do not own this equipment")
	}
	if req.Condition != "" && !validConditions[req.Condition] {
		return nil, errors.New(errors.CodeValidation, "Condition must be 'new' or 'used'")
	}

	if req.Title != "" {
		eq.Title = req.Title
	}
	if req.Description != "" {
		eq.Description = req.Description
	}
	if req.Condition != "" {
		eq.Condition = req.Condition
	}
	if req.Region != "" {
		eq.Region = req.Region
	}
	if req.ImageURL != "" {
		eq.ImageURL = req.ImageURL
	}

	if err := s.repo.UpdateEquipment(ctx, eq); err != nil {
		return nil, err
	}
	return eq, nil
}

func (s *service) DeleteEquipment(ctx context.Context, id, ownerID string) error {
	eq, err := s.repo.GetEquipmentByID(ctx, id)
	if err != nil {
		return err
	}
	if eq.OwnerID != ownerID {
		return errors.New(errors.CodeUnauthorized, "You do not own this equipment")
	}
	return s.repo.DeleteEquipment(ctx, id)
}

// === Contracts (EquipmentProvider) ===

func (s *service) GetEquipmentBasic(ctx context.Context, equipmentID string) (*contracts.EquipmentBasic, error) {
	eq, err := s.repo.GetEquipmentByID(ctx, equipmentID)
	if err != nil {
		return nil, err
	}
	return &contracts.EquipmentBasic{
		ID:      eq.ID,
		Title:   eq.Title,
		OwnerID: eq.OwnerID,
	}, nil
}
