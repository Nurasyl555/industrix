package auth

import (
	"context"
	"time"

	"github.com/industrix/pkg/errors"
	"github.com/industrix/pkg/jwt"
)

// Repository interface
type Repository interface {
	UserExists(ctx context.Context, email, phone string) (bool, error)
	CreateUser(ctx context.Context, email, phone, password string) error
	SaveOTP(ctx context.Context, phone, code string, ttl time.Duration) error
	ValidateOTP(ctx context.Context, phone, code string) (bool, error)
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUserVerification(ctx context.Context, userID string, verified bool) error
	CheckPassword(hash, password string) bool
}

type User struct {
	ID           string
	Email        string
	Phone        string
	PasswordHash string
	Role         string
	Verified     bool
	CompanyID    string
}

type Service interface {
	Register(ctx context.Context, email, phone, password string) error
	VerifyOTP(ctx context.Context, phone, code string) (*jwt.TokenPair, error)
	Login(ctx context.Context, email, password string) (*jwt.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
}

type service struct {
	repo      Repository
	jwtClient jwt.Client
}

func NewService(repo Repository, jwtClient jwt.Client) Service {
	return &service{repo: repo, jwtClient: jwtClient}
}

func (s *service) Register(ctx context.Context, email, phone, password string) error {
	exists, err := s.repo.UserExists(ctx, email, phone)
	if err != nil {
		return err
	}
	if exists {
		return errors.New(errors.CodeConflict, "User already exists")
	}

	// Create user (unverified)
	// In real app, hash password here
	hashedPassword := password // Placeholder
	if err := s.repo.CreateUser(ctx, email, phone, hashedPassword); err != nil {
		return err
	}

	// Generate and save OTP
	otp := "123456" // Placeholder
	if err := s.repo.SaveOTP(ctx, phone, otp, 5*time.Minute); err != nil {
		return err
	}

	// Send OTP logic (omitted)
	return nil
}

func (s *service) VerifyOTP(ctx context.Context, phone, code string) (*jwt.TokenPair, error) {
	valid, err := s.repo.ValidateOTP(ctx, phone, code)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, errors.New(errors.CodeValidation, "Invalid OTP")
	}

	user, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	if !user.Verified {
		if err := s.repo.UpdateUserVerification(ctx, user.ID, true); err != nil {
			return nil, err
		}
	}

	return s.jwtClient.IssuePair(user.ID, user.CompanyID, user.Role, true, 15*time.Minute, 24*time.Hour)
}

func (s *service) Login(ctx context.Context, email, password string) (*jwt.TokenPair, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New(errors.CodeUnauthorized, "Invalid credentials")
	}

	if !s.repo.CheckPassword(user.PasswordHash, password) {
		return nil, errors.New(errors.CodeUnauthorized, "Invalid credentials")
	}

	return s.jwtClient.IssuePair(user.ID, user.CompanyID, user.Role, user.Verified, 15*time.Minute, 24*time.Hour)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	claims, err := s.jwtClient.ParseClaims(refreshToken)
	if err != nil {
		return nil, err
	}

	// Ideally check revocation list here

	return s.jwtClient.IssuePair(claims.UserID, claims.CompanyID, claims.Role, claims.Verified, 15*time.Minute, 24*time.Hour)
}
