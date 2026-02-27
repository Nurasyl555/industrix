package grpc

import (
	"context"

	trustpb "github.com/industrix/gen/go/backend/proto/trust/v1"
	"github.com/industrix/pkg/jwt"
	"github.com/industrix/services/trust/internal/auth"
	"github.com/industrix/services/trust/internal/company"
	"github.com/industrix/services/trust/internal/profile"
	"github.com/industrix/services/trust/internal/review"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	trustpb.UnimplementedTrustServiceServer
	authRepo    auth.Repository
	profileRepo profile.Repository
	companyRepo company.Repository
	reviewRepo  review.Repository
	jwtClient   jwt.Client
}

func NewServer(
	authRepo auth.Repository,
	profileRepo profile.Repository,
	companyRepo company.Repository,
	reviewRepo review.Repository,
	jwtClient jwt.Client,
) *Server {
	return &Server{
		authRepo:    authRepo,
		profileRepo: profileRepo,
		companyRepo: companyRepo,
		reviewRepo:  reviewRepo,
		jwtClient:   jwtClient,
	}
}

func (s *Server) GetUser(ctx context.Context, req *trustpb.GetUserRequest) (*trustpb.User, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	user, err := s.profileRepo.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User not found: %v", err)
	}

	return &trustpb.User{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CompanyId: user.CompanyID,
		// Role and Verified would come from Auth repo join or enhanced profile model
		AvatarUrl: user.AvatarURL,
	}, nil
}

func (s *Server) GetUserBatch(ctx context.Context, req *trustpb.GetUserBatchRequest) (*trustpb.GetUserBatchResponse, error) {
	// Implement batch retrieval (loop for now, optimize with IN query later)
	var users []*trustpb.User
	for _, id := range req.Ids {
		u, err := s.GetUser(ctx, &trustpb.GetUserRequest{Id: id})
		if err == nil {
			users = append(users, u)
		}
	}
	return &trustpb.GetUserBatchResponse{Users: users}, nil
}

func (s *Server) GetCompany(ctx context.Context, req *trustpb.GetCompanyRequest) (*trustpb.Company, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Company ID is required")
	}

	comp, err := s.companyRepo.GetCompanyByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Company not found: %v", err)
	}

	return &trustpb.Company{
		Id:       comp.ID,
		Name:     comp.Name,
		Bin:      comp.BIN,
		Verified: comp.Verified,
		Address:  comp.Address,
		Phone:    comp.Phone,
		Email:    comp.Email,
		Website:  comp.Website,
	}, nil
}

func (s *Server) VerifyToken(ctx context.Context, req *trustpb.VerifyTokenRequest) (*trustpb.Claims, error) {
	claims, err := s.jwtClient.ParseClaims(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token: %v", err)
	}

	return &trustpb.Claims{
		UserId:    claims.UserID,
		CompanyId: claims.CompanyID,
		Role:      claims.Role,
		Verified:  claims.Verified,
		Exp:       claims.ExpiresAt.Unix(),
	}, nil
}

func (s *Server) GetVerificationStatus(ctx context.Context, req *trustpb.GetVerificationStatusRequest) (*trustpb.VerificationStatus, error) {
	// Mock implementation for now, or fetch from company repo
	comp, err := s.companyRepo.GetCompanyByID(ctx, req.CompanyId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Company not found: %v", err)
	}

	statusStr := "pending"
	if comp.Verified {
		statusStr = "verified"
	}

	return &trustpb.VerificationStatus{
		Status: statusStr,
	}, nil
}

func (s *Server) SubmitDocument(ctx context.Context, req *trustpb.SubmitDocumentRequest) (*trustpb.SubmitDocumentResponse, error) {
	// Placeholder for document submission logic
	// In real impl, validate inputs, create record in company_documents
	return &trustpb.SubmitDocumentResponse{
		DocumentId: "doc-123", // generated ID
		Status:     "submitted",
	}, nil
}

func (s *Server) GetReputation(ctx context.Context, req *trustpb.GetReputationRequest) (*trustpb.ReputationScore, error) {
	score, err := s.reviewRepo.GetReputationScore(ctx, req.EntityId)
	if err != nil {
		// Return empty score if not found
		return &trustpb.ReputationScore{
			EntityId: req.EntityId,
			Score:    0,
			Tier:     "none",
		}, nil
	}

	return &trustpb.ReputationScore{
		EntityId:    score.EntityID,
		Score:       float32(score.AverageRating),
		ReviewCount: int32(score.ReviewCount),
		Tier:        score.Tier,
	}, nil
}
