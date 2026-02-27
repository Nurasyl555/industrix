package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/industrix/pkg/errors"
	"github.com/industrix/pkg/jwt"
	"github.com/industrix/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// Repository interface
type Repository interface {
	UserExists(ctx context.Context, email, phone string) (bool, error)
	CreateUser(ctx context.Context, email, phone, passwordHash string) error
	SaveOTP(ctx context.Context, phone, code string, ttl time.Duration) error
	ValidateOTP(ctx context.Context, phone, code string) (bool, error)
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUserVerification(ctx context.Context, userID string, verified bool) error
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New(errors.CodeInternal, "Failed to hash password")
	}

	if err := s.repo.CreateUser(ctx, email, phone, string(hashedPassword)); err != nil {
		return err
	}

	// Generate and save OTP
	otp := "123456" // In production, generate random code
	if err := s.repo.SaveOTP(ctx, phone, otp, 5*time.Minute); err != nil {
		return err
	}

	// Send OTP via SMS
	if err := s.sendSMS(ctx, phone, fmt.Sprintf("Your verification code is: %s", otp)); err != nil {
		logger.New("auth-service").Error().Err(err).Msg("Failed to send OTP SMS")
	}

	return nil
}

func (s *service) sendSMS(ctx context.Context, phone, message string) error {
	logger.New("auth-service").Info().
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

	// Ideally check revocation list here

	return s.jwtClient.IssuePair(claims.UserID, claims.CompanyID, claims.Role, claims.Verified, 15*time.Minute, 24*time.Hour)
}
