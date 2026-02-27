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

type Service struct {
	repo        *Repository
	kafkaClient *kafka.Producer
	minioClient *minio.Client
	log         *logger.Logger
}

type CompanyStatus string
const (
	StatusPending     CompanyStatus = "pending"
	StatusUnderReview CompanyStatus = "under_review"
	StatusVerified    CompanyStatus = "verified"
	StatusRejected    CompanyStatus = "rejected"
)

func NewService(repo *Repository, kafkaClient *kafka.Producer, minioClient *minio.Client) *Service {
	return &Service{repo: repo, kafkaClient: kafkaClient, minioClient: minioClient, log: logger.New("company-service")}
}

func (s *Service) CreateCompany(ctx context.Context, userID string, req CreateCompanyRequest) (*Company, *errors.Error) {
	if req.BIN != "" && !isValidBIN(req.BIN) { return nil, errors.Validation("invalid BIN format") }
	existing, err := s.repo.GetCompanyByUserID(ctx, userID)
	if err == nil && existing != nil { return nil, errors.Conflict("user already has a company") }
	company, err := s.repo.CreateCompany(ctx, userID, req)
	if err != nil { return nil, errors.Internal("failed to create company") }
	return company, nil
}

func (s *Service) UpdateCompany(ctx context.Context, userID string, req UpdateCompanyRequest) (*Company, *errors.Error) {
	company, err := s.repo.GetCompanyByUserID(ctx, userID)
	if err != nil { return nil, errors.NotFound("company not found") }
	if company.Status == StatusVerified && (req.Name != "" || req.BIN != "" || req.Address != "") {
		s.repo.UpdateStatus(ctx, company.ID, StatusPending)
	}
	company, err = s.repo.UpdateCompany(ctx, userID, req)
	if err != nil { return nil, errors.Internal("failed to update company") }
	return company, nil
}

func (s *Service) UploadDocument(ctx context.Context, userID string, req UploadDocumentRequest) (*VerificationDocument, *errors.Error) {
	company, err := s.repo.GetCompanyByUserID(ctx, userID)
	if err != nil { return nil, errors.NotFound("company not found") }
	var uploadURL string
	if req.ContentType != "" && req.FileName != "" {
		objectName := "companies/" + company.ID + "/verification/" + req.FileName
		url, _ := s.minioClient.PresignPutURL(ctx, objectName, 15*time.Minute)
		uploadURL = url
	}
	doc, err := s.repo.CreateDocument(ctx, company.ID, req)
	if err != nil { return nil, errors.Internal("failed to create document") }
	doc.UploadURL = uploadURL
	s.repo.UpdateStatus(ctx, company.ID, StatusUnderReview)
	return doc, nil
}

func (s *Service) GetVerificationStatus(ctx context.Context, userID string) (*VerificationStatus, *errors.Error) {
	company, err := s.repo.GetCompanyByUserID(ctx, userID)
	if err != nil { return nil, errors.NotFound("company not found") }
	docs, _ := s.repo.GetDocuments(ctx, company.ID)
	return &VerificationStatus{Status: string(company.Status), ReviewerNote: company.ReviewerNote, Documents: docs, UpdatedAt: company.UpdatedAt}, nil
}

func (s *Service) UpdateVerificationStatus(ctx context.Context, companyID string, status CompanyStatus, reviewerNote string) *errors.Error {
	err := s.repo.UpdateVerificationStatus(ctx, companyID, status, reviewerNote)
	if err != nil { return errors.Internal("failed to update verification status") }
	if s.kafkaClient != nil {
		eventType := "company.verified"
		if status == StatusRejected { eventType = "company.rejected" }
		s.kafkaClient.Publish(ctx, "company.events", companyID, map[string]interface{}{"company_id": companyID, "event_type": eventType})
	}
	return nil
}

func (s *Service) HandleReviewCreated(ctx context.Context, event map[string]interface{}) *errors.Error {
	targetUserID, _ := event["target_user_id"].(string)
	s.repo.UpdateCompanyReputation(ctx, targetUserID)
	return nil
}

var binRegex = regexp.MustCompile(`^\d{12}$`)
func isValidBIN(bin string) bool { return binRegex.MatchString(bin) }
