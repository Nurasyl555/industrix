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
	CompareEquipment(ctx context.Context, ids []string) ([]*Equipment, error)
	ListEquipment(ctx context.Context, f ListEquipmentFilter) ([]*Equipment, int64, error)
	UpdateEquipment(ctx context.Context, id, ownerID string, req UpdateEquipmentRequest) (*Equipment, error)
	DeleteEquipment(ctx context.Context, id, ownerID string) error

	// Contracts
	contracts.EquipmentProvider
}

type service struct {
	repo   *Repository
	events contracts.EventPublisher
}

// NewService creates a new catalog service
func NewService(repo *Repository, events contracts.EventPublisher) Service {
	return &service{repo: repo, events: events}
}

// equipmentEvent is the payload published on equipment.* topics. It carries
// enough for search indexing without consumers calling back into catalog.
type equipmentEvent struct {
	ID          string `json:"id"`
	OwnerID     string `json:"owner_id"`
	CategoryID  string `json:"category_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Condition   string `json:"condition"`
	Region      string `json:"region"`
	ImageURL    string `json:"image_url"`
}

func toEquipmentEvent(eq *Equipment) equipmentEvent {
	return equipmentEvent{
		ID:          eq.ID,
		OwnerID:     eq.OwnerID,
		CategoryID:  eq.CategoryID,
		Title:       eq.Title,
		Description: eq.Description,
		Condition:   eq.Condition,
		Region:      eq.Region,
		ImageURL:    eq.ImageURL,
	}
}

// emit publishes a domain event if a publisher is wired. Guarded so the service
// works with a nil publisher in tests.
func (s *service) emit(ctx context.Context, topic, key string, payload any) {
	if s.events != nil {
		s.events.Publish(ctx, topic, key, payload)
	}
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
	s.emit(ctx, contracts.TopicEquipmentCreated, eq.ID, toEquipmentEvent(eq))
	return eq, nil
}

func (s *service) GetEquipment(ctx context.Context, id string) (*Equipment, error) {
	return s.repo.GetEquipmentByID(ctx, id)
}

// CompareEquipment returns full details for 2–10 equipment ids for a
// side-by-side comparison view.
func (s *service) CompareEquipment(ctx context.Context, ids []string) ([]*Equipment, error) {
	if len(ids) < 2 {
		return nil, errors.New(errors.CodeValidation, "Provide at least 2 ids to compare")
	}
	if len(ids) > 10 {
		ids = ids[:10]
	}
	return s.repo.ListEquipmentByIDs(ctx, ids)
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
	s.emit(ctx, contracts.TopicEquipmentUpdated, eq.ID, toEquipmentEvent(eq))
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
	if err := s.repo.DeleteEquipment(ctx, id); err != nil {
		return err
	}
	s.emit(ctx, contracts.TopicEquipmentDeleted, id, equipmentEvent{ID: id, OwnerID: ownerID})
	return nil
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
