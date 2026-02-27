package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/jwt"
	"github.com/industrix/backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the identity service interface
type Service interface {
	// Auth
	Register(ctx context.Context, email, phone, password string) error
	VerifyOTP(ctx context.Context, phone, code string) (*jwt.TokenPair, error)
	Login(ctx context.Context, email, password string) (*jwt.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)

	// Profile
	GetProfile(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, user *User) error

	// Contracts
	contracts.UserProvider
}

type service struct {
	repo      *Repository
	jwtClient jwt.Client
	log       *logger.Logger
}

// NewService creates a new identity service
func NewService(repo *Repository, jwtClient jwt.Client) Service {
	return &service{
		repo:      repo,
		jwtClient: jwtClient,
		log:       logger.New("identity-service"),
	}
}

// === Auth ===

func (s *service) Register(ctx context.Context, email, phone, password string) error {
	exists, err := s.repo.UserExists(ctx, email, phone)
	if err != nil {
		return err
	}
	if exists {
		return errors.New(errors.CodeConflict, "User already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New(errors.CodeInternal, "Failed to hash password")
	}

	if err := s.repo.CreateUser(ctx, email, phone, string(hashedPassword)); err != nil {
		return err
	}

	otp := "123456" // In production, generate random code
	if err := s.repo.SaveOTP(ctx, phone, otp, 5*time.Minute); err != nil {
		return err
	}

	if err := s.sendSMS(ctx, phone, fmt.Sprintf("Your verification code is: %s", otp)); err != nil {
		s.log.Error().Err(err).Msg("Failed to send OTP SMS")
	}

	return nil
}

func (s *service) sendSMS(ctx context.Context, phone, message string) error {
	s.log.Info().
		Str("phone", phone).
		Str("message", message).
		Msg("Sending SMS (Mock)")
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New(errors.CodeUnauthorized, "Invalid credentials")
	}

	return s.jwtClient.IssuePair(user.ID, user.CompanyID, user.Role, user.Verified, 15*time.Minute, 24*time.Hour)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	claims, err := s.jwtClient.ParseClaims(refreshToken)
	if err != nil {
		return nil, err
	}

	return s.jwtClient.IssuePair(claims.UserID, claims.CompanyID, claims.Role, claims.Verified, 15*time.Minute, 24*time.Hour)
}

// === Profile ===

func (s *service) GetProfile(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *service) UpdateProfile(ctx context.Context, user *User) error {
	if user.ID == "" {
		return errors.New(errors.CodeValidation, "User ID is required")
	}
	return s.repo.UpdateUser(ctx, user)
}

// === Contracts (UserProvider) ===

func (s *service) GetUserBasic(ctx context.Context, userID string) (*contracts.UserBasic, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &contracts.UserBasic{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
	}, nil
}
