package grpc

import (
	"context"

	identityv1 "github.com/industrix/gen/go/identity/v1"
	"github.com/industrix/pkg/jwt"
	"github.com/industrix/services/identity/internal/auth"
	"github.com/industrix/services/identity/internal/profile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	identityv1.UnimplementedIdentityServiceServer
	authRepo    *auth.Repository
	profileRepo *profile.Repository
	jwtClient   *jwt.Client
}

func NewServer(authRepo *auth.Repository, profileRepo *profile.Repository, jwtClient *jwt.Client) *Server {
	return &Server{
		authRepo:    authRepo,
		profileRepo: profileRepo,
		jwtClient:   jwtClient,
	}
}

func (s *Server) GetUser(ctx context.Context, req *identityv1.GetUserRequest) (*identityv1.GetUserResponse, error) {
	user, err := s.authRepo.GetUserByID(ctx, req.GetId())
	if err != nil { return nil, status.Error(codes.NotFound, "user not found") }
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

func (s *Server) GetCompany(ctx context.Context, req *identityv1.GetCompanyRequest) (*identityv1.GetCompanyResponse, error) {
	company, err := s.profileRepo.GetCompanyByID(ctx, req.GetId())
	if err != nil { return nil, status.Error(codes.NotFound, "company not found") }
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

func (s *Server) VerifyToken(ctx context.Context, req *identityv1.VerifyTokenRequest) (*identityv1.VerifyTokenResponse, error) {
	claims, err := s.jwtClient.ParseClaims(ctx, req.GetToken())
	if err != nil { return &identityv1.VerifyTokenResponse{Valid: false}, nil }
	return &identityv1.VerifyTokenResponse{
		Valid: true,
		Claims: &identityv1.Claims{
			UserId:    claims.UserID,
			CompanyId: claims.CompanyID,
			Role:      claims.Role,
			Verified:  claims.Verified,
			Scope:     claims.Scope,
			Exp:       claims.ExpiresAt.Unix(),
			Iat:       claims.IssuedAt.Unix(),
		},
	}, nil
}

func (s *Server) GetUserBatch(ctx context.Context, req *identityv1.GetUserBatchRequest) (*identityv1.GetUserBatchResponse, error) {
	users := make([]*identityv1.User, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		user, err := s.authRepo.GetUserByID(ctx, id)
		if err != nil { continue }
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
	return &identityv1.GetUserBatchResponse{Users: users}, nil
}

func RegisterGRPC(server *grpc.Server, svc *Server) {
	identityv1.RegisterIdentityServiceServer(server, svc)
}
