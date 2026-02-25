package company

import (
	"context"
	"regexp"
	"time"

	"github.com/industrix/pkg/errors"
	"github.com/industrix/pkg/kafka"
	"github.com/industrix/pkg/logger"
	"github.com/industrix/pkg/minio"
)

// Service handles business logic for company
type Service struct {
	repo        *Repository
	kafkaClient *kafka.Producer
	minioClient *minio.Client
	log         *logger.Logger
}

// CompanyStatus represents company verification status
type CompanyStatus string

const (
	StatusPending     CompanyStatus = "pending"
	StatusUnderReview CompanyStatus = "under_review"
	StatusVerified    CompanyStatus = "verified"
	StatusRejected    CompanyStatus = "rejected"
)

// NewService creates a new company service
func NewService(repo *Repository, kafkaClient *kafka.Producer, minioClient *minio.Client) *Service {
	return &Service{
		repo:        repo,
		kafkaClient: kafkaClient,
		minioClient: minioClient,
		log:         logger.New("company-service"),
	}
}

// CreateCompany creates a new company
func (s *Service) CreateCompany(ctx context.Context, userID string, req CreateCompanyRequest) (*Company, *errors.Error) {
	// Validate BIN format (12-digit Kazakhstan BIN)
	if req.BIN != "" {
		if !isValidBIN(req.BIN) {
			return nil, errors.Validation("invalid BIN format: must be 12 digits")
		}
	}

	// Check if user already has a company
	existing, err := s.repo.GetCompanyByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, errors.Conflict("user already has a company")
	}

	// Create company
	company, err := s.repo.CreateCompany(ctx, userID, req)
	if err != nil {
		return nil, errors.Internal("failed to create company")
	}

	return company, nil
}

// UpdateCompany updates company info
func (s *Service) UpdateCompany(ctx context.Context, userID string, req UpdateCompanyRequest) (*Company, *errors.Error) {
	company, err := s.repo.GetCompanyByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NotFound("company not found")
	}

	// If already verified, changes to legal fields require re-verification
	if company.Status == StatusVerified {
		if req.Name != "" || req.BIN != "" || req.Address != "" {
			// Set status back to pending for re-verification
			err = s.repo.UpdateStatus(ctx, company.ID, StatusPending)
			if err != nil {
				return nil, errors.Internal("failed to update company status")
			}
		}
	}

	company, err = s.repo.UpdateCompany(ctx, userID, req)
	if err != nil {
		return nil, errors.Internal("failed to update company")
	}

	return company, nil
}

// UploadDocument uploads verification documents
func (s *Service) UploadDocument(ctx context.Context, userID string, req UploadDocumentRequest) (*VerificationDocument, *errors.Error) {
	company, err := s.repo.GetCompanyByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NotFound("company not found")
	}

	// Generate presigned URL if file data provided
	var uploadURL string
	if req.ContentType != "" && req.FileName != "" {
		objectName := "companies/" + company.ID + "/verification/" + req.FileName
		url, err := s.minioClient.PresignPutURL(ctx, objectName, 15*time.Minute)
		if err != nil {
			return nil, errors.Internal("failed to generate upload URL")
		}
		uploadURL = url
	}

	// Create document record
	doc, err := s.repo.CreateDocument(ctx, company.ID, req)
	if err != nil {
		return nil, errors.Internal("failed to create document")
	}

	doc.UploadURL = uploadURL

	// Update company status to under_review
	err = s.repo.UpdateStatus(ctx, company.ID, StatusUnderReview)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to update company status")
	}

	return doc, nil
}

// GetVerificationStatus returns verification status
func (s *Service) GetVerificationStatus(ctx context.Context, userID string) (*VerificationStatus, *errors.Error) {
	company, err := s.repo.GetCompanyByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NotFound("company not found")
	}

	docs, err := s.repo.GetDocuments(ctx, company.ID)
	if err != nil {
		return nil, errors.Internal("failed to get documents")
	}

	return &VerificationStatus{
		Status:       string(company.Status),
		ReviewerNote: company.ReviewerNote,
		Documents:    docs,
		UpdatedAt:    company.UpdatedAt,
	}, nil
}

// UpdateVerificationStatus updates company verification status (called by admin)
func (s *Service) UpdateVerificationStatus(ctx context.Context, companyID string, status CompanyStatus, reviewerNote string) *errors.Error {
	err := s.repo.UpdateVerificationStatus(ctx, companyID, status, reviewerNote)
	if err != nil {
		return errors.Internal("failed to update verification status")
	}

	// Emit Kafka events
	if s.kafkaClient != nil {
		if status == StatusVerified {
			s.kafkaClient.Publish(ctx, "company.events", map[string]interface{}{
				"company_id": companyID,
				"event_type": "company.verified",
				"timestamp":  time.Now().Unix(),
			})
		} else if status == StatusRejected {
			s.kafkaClient.Publish(ctx, "company.events", map[string]interface{}{
				"company_id": companyID,
				"event_type": "company.rejected",
				"reason":     reviewerNote,
				"timestamp":  time.Now().Unix(),
			})
		}
	}

	return nil
}

// HandleReviewCreated handles review.created Kafka event
func (s *Service) HandleReviewCreated(ctx context.Context, event map[string]interface{}) *errors.Error {
	targetUserID, ok := event["target_user_id"].(string)
	if !ok {
		return nil
	}

	// Recalculate company reputation score
	err := s.repo.UpdateCompanyReputation(ctx, targetUserID)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to update company reputation")
	}

	return nil
}

// BIN validation regex (12-digit Kazakhstan BIN)
var binRegex = regexp.MustCompile(`^\d{12}$`)

func isValidBIN(bin string) bool {
	return binRegex.MatchString(bin)
}
