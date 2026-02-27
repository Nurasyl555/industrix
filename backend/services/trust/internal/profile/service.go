package profile

import (
	"context"

	"github.com/industrix/pkg/errors"
)

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url"`
	CompanyID string `json:"company_id"`
}

type Repository interface {
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
}

type Service interface {
	GetProfile(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, user *User) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetProfile(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *service) UpdateProfile(ctx context.Context, user *User) error {
	if user.ID == "" {
		return errors.New(errors.CodeValidation, "User ID is required")
	}
	return s.repo.UpdateUser(ctx, user)
}
