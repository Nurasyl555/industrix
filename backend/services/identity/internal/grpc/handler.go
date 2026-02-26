package grpc

import (
	"context"

	"github.com/industrix/services/identity/internal/auth"
	"github.com/industrix/services/identity/internal/profile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the IdentityService gRPC server
type Server struct {
	identityv1.UnimplementedIdentityServiceServer
	authRepo    *auth.Repository
	profileRepo *profile.Repository
}

// NewServer creates a new gRPC server
func NewServer(authRepo *auth.Repository, profileRepo *profile.Repository) *Server {
	return &Server{
		authRepo:    authRepo,
		profileRepo: profileRepo,
	}
}

// GetUser implements identity.v1.IdentityService.GetUser
func (s *Server) GetUser(ctx context.Context, req *identityv1.GetUserRequest) (*identityv1.GetUserResponse, error) {
	user, err := s.authRepo.GetUserByID(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &identityv1.GetUserResponse{
		User: &identityv1.User{
			Id:        user.ID,
			Email:     user.Email,
			Phone:     user.Phone,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			Verified:  user.Verified,
			CompanyId: user.CompanyID,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		},
	}, nil
}

// GetCompany implements identity.v1.IdentityService.GetCompany
func (s *Server) GetCompany(ctx context.Context, req *identityv1.GetCompanyRequest) (*identityv1.GetCompanyResponse, error) {
	company, err := s.profileRepo.GetCompanyByID(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "company not found")
	}

	return &identityv1.GetCompanyResponse{
		Company: &identityv1.Company{
			Id:               company.ID,
			Name:             company.Name,
			Bin:              company.BIN,
			Address:          company.Address,
			Phone:            company.Phone,
			Email:            company.Email,
			Website:          company.Website,
			Verified:         company.Verified,
			SubscriptionPlan: company.SubscriptionPlan,
			CreatedAt:        company.CreatedAt.Unix(),
			UpdatedAt:        company.UpdatedAt.Unix(),
		},
	}, nil
}

// VerifyToken implements identity.v1.IdentityService.VerifyToken
func (s *Server) VerifyToken(ctx context.Context, req *identityv1.VerifyTokenRequest) (*identityv1.VerifyTokenResponse, error) {
	// This would use the JWT client to verify the token
	// For now, return a placeholder
	return &identityv1.VerifyTokenResponse{
		Valid: false,
	}, nil
}

// GetUserBatch implements identity.v1.IdentityService.GetUserBatch
func (s *Server) GetUserBatch(ctx context.Context, req *identityv1.GetUserBatchRequest) (*identityv1.GetUserBatchResponse, error) {
	users := make([]*identityv1.User, 0, len(req.GetIds()))

	for _, id := range req.GetIds() {
		user, err := s.authRepo.GetUserByID(ctx, id)
		if err != nil {
			continue // Skip invalid IDs
		}

		users = append(users, &identityv1.User{
			Id:        user.ID,
			Email:     user.Email,
			Phone:     user.Phone,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			Verified:  user.Verified,
			CompanyId: user.CompanyID,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		})
	}

	return &identityv1.GetUserBatchResponse{
		Users: users,
	}, nil
}

// RegisterGRPC registers the gRPC service
func RegisterGRPC(server *grpc.Server, svc *Server) {
	identityv1.RegisterIdentityServiceServer(server, svc)
}
