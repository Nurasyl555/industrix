package auth

import (
	"context"
	"time"

	"github.com/industrix/pkg/errors"
	"github.com/industrix/pkg/jwt"
	"github.com/industrix/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// Service handles business logic for authentication
type Service struct {
	repo      *Repository
	jwtClient *jwt.Client
	log       *logger.Logger

	// Config
	otpTTL   time.Duration
	resetTTL time.Duration
}

// ServiceResult contains results from auth operations
type ServiceResult struct {
	UserID       string
	Status       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

// NewService creates a new auth service
func NewService(repo *Repository, jwtClient *jwt.Client) *Service {
	return &Service{
		repo:      repo,
		jwtClient: jwtClient,
		log:       logger.New("auth-service"),
		otpTTL:    5 * time.Minute,
		resetTTL:  15 * time.Minute,
	}
}

// Register handles user registration
// 1. Validate input
// 2. Check if user exists
// 3. Hash password
// 4. Create user record
// 5. Generate and store OTP
// 6. Send OTP (simulated for now)
func (s *Service) Register(ctx context.Context, email, phone, password, role string) (*ServiceResult, *errors.Error) {
	// Validate role
	if role == "" {
		role = "BUYER"
	}
	if role != "BUYER" && role != "SELLER" && role != "SERVICE_COMPANY" {
		return nil, errors.Validation("invalid role")
	}

	// Check if user already exists
	exists, err := s.repo.UserExists(ctx, email, phone)
	if err != nil {
		return nil, errors.Internal("failed to check user existence")
	}
	if exists {
		return nil, errors.Conflict("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Internal("failed to hash password")
	}

	// Create user
	userID, err := s.repo.CreateUser(ctx, email, phone, string(hashedPassword), role)
	if err != nil {
		return nil, errors.Internal("failed to create user")
	}

	// Generate OTP
	otp := generateOTP()
	err = s.repo.StoreOTP(ctx, phone, otp, s.otpTTL)
	if err != nil {
		return nil, errors.Internal("failed to store OTP")
	}

	// TODO: Send OTP via SMS (Kcell/Beeline)
	s.log.Info().Str("phone", phone).Str("otp", otp).Msg("OTP generated (SMS sending not implemented)")

	return &ServiceResult{
		UserID: userID,
		Status: "PENDING_OTP",
	}, nil
}

// VerifyOTP verifies OTP and activates account
// 1. Validate OTP from Redis
// 2. Activate user account
// 3. Issue JWT tokens
func (s *Service) VerifyOTP(ctx context.Context, phone, otp string) (*ServiceResult, *errors.Error) {
	// Validate OTP
	storedOTP, err := s.repo.GetOTP(ctx, phone)
	if err != nil {
		return nil, errors.Unauthorized("invalid or expired OTP")
	}
	if storedOTP != otp {
		return nil, errors.Unauthorized("invalid OTP")
	}

	// Delete OTP after successful verification
	err = s.repo.DeleteOTP(ctx, phone)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to delete OTP")
	}

	// Activate user
	err = s.repo.ActivateUser(ctx, phone)
	if err != nil {
		return nil, errors.Internal("failed to activate user")
	}

	// Get user
	user, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, errors.Internal("failed to get user")
	}

	// Issue JWT tokens
	tokens, err := s.jwtClient.IssuePair(ctx, user.ID, user.CompanyID, user.Role, user.Verified, nil)
	if err != nil {
		return nil, errors.Internal("failed to issue tokens")
	}

	return &ServiceResult{
		UserID:       user.ID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Unix(),
	}, nil
}

// Login handles user login
// 1. Find user by email
// 2. Compare password with bcrypt
// 3. Issue JWT tokens
// 4. Store refresh token in Redis
func (s *Service) Login(ctx context.Context, email, password, deviceID string) (*ServiceResult, *errors.Error) {
	// Find user
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.Unauthorized("invalid credentials")
	}

	// Check if user is active
	if !user.Active {
		return nil, errors.Unauthorized("account not active")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.Unauthorized("invalid credentials")
	}

	// Issue JWT tokens
	tokens, err := s.jwtClient.IssuePair(ctx, user.ID, user.CompanyID, user.Role, user.Verified, nil)
	if err != nil {
		return nil, errors.Internal("failed to issue tokens")
	}

	// Store refresh token in Redis with device fingerprint
	if deviceID != "" {
		err = s.repo.StoreRefreshToken(ctx, user.ID, deviceID, tokens.RefreshToken, 7*24*time.Hour)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to store refresh token")
		}
	}

	return &ServiceResult{
		UserID:       user.ID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Unix(),
	}, nil
}

// Refresh handles token refresh
// 1. Validate refresh token
// 2. Check if token exists in Redis
// 3. Rotate tokens (invalidate old, issue new)
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*ServiceResult, *errors.Error) {
	// Parse refresh token
	claims, err := s.jwtClient.ParseRefreshClaims(ctx, refreshToken)
	if err != nil {
		return nil, errors.Unauthorized("invalid refresh token")
	}

	// Check if token is revoked
	revoked, err := s.repo.IsRefreshTokenRevoked(ctx, refreshToken)
	if err != nil {
		return nil, errors.Internal("failed to check token revocation")
	}
	if revoked {
		return nil, errors.Unauthorized("refresh token has been revoked")
	}

	// Revoke old refresh token (rotation)
	err = s.repo.RevokeRefreshToken(ctx, refreshToken)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to revoke old refresh token")
	}

	// Issue new tokens
	tokens, err := s.jwtClient.IssuePair(ctx, claims.UserID, claims.CompanyID, claims.Role, claims.Verified, claims.Scope)
	if err != nil {
		return nil, errors.Internal("failed to issue tokens")
	}

	return &ServiceResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Unix(),
	}, nil
}

// Logout handles user logout
// 1. Invalidate refresh token in Redis
// 2. If AllSessions, invalidate all user tokens
func (s *Service) Logout(ctx context.Context, refreshToken string, allSessions bool) *errors.Error {
	if allSessions && refreshToken != "" {
		// Get user from token
		claims, err := s.jwtClient.ParseRefreshClaims(ctx, refreshToken)
		if err == nil {
			err = s.repo.RevokeAllUserTokens(ctx, claims.UserID)
			if err != nil {
				s.log.Error().Err(err).Msg("failed to revoke all user tokens")
			}
		}
	} else if refreshToken != "" {
		err := s.repo.RevokeRefreshToken(ctx, refreshToken)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to revoke refresh token")
		}
	}

	return nil
}

// ForgotPassword handles forgot password request
// 1. Find user by email/phone
// 2. Generate reset token
// 3. Store in Redis with 15min TTL
// 4. Send reset link/OTP
func (s *Service) ForgotPassword(ctx context.Context, email, phone string) *errors.Error {
	if email == "" && phone == "" {
		return errors.Validation("email or phone is required")
	}

	// Check if user exists (but don't reveal to caller)
	var exists bool
	var err error

	if email != "" {
		_, err = s.repo.GetUserByEmail(ctx, email)
	} else {
		_, err = s.repo.GetUserByPhone(ctx, phone)
	}

	if err != nil {
		// Don't reveal if user exists
		return nil
	}
	exists = true
	if !exists {
		// Don't reveal if user exists
		return nil
	}

	// Generate OTP for password reset
	otp := generateOTP()
	var key string
	if phone != "" {
		key = "password_reset:" + phone
	} else {
		key = "password_reset:" + email
	}
	err = s.repo.StoreOTP(ctx, key, otp, s.resetTTL)
	if err != nil {
		return errors.Internal("failed to store reset token")
	}

	// TODO: Send OTP via SMS/Email
	s.log.Info().Str("key", key).Str("otp", otp).Msg("Password reset OTP generated")

	return nil
}

// ResetPassword handles password reset
// 1. Validate reset token
// 2. Hash new password
// 3. Update user password
// 4. Invalidate all existing sessions
func (s *Service) ResetPassword(ctx context.Context, phone, otp, newPassword string) *errors.Error {
	if phone == "" {
		return errors.Validation("phone is required")
	}

	// Validate OTP
	key := "password_reset:" + phone
	storedOTP, err := s.repo.GetOTP(ctx, key)
	if err != nil {
		return errors.Unauthorized("invalid or expired reset token")
	}
	if storedOTP != otp {
		return errors.Unauthorized("invalid OTP")
	}

	// Delete OTP after successful validation
	err = s.repo.DeleteOTP(ctx, key)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to delete reset OTP")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Internal("failed to hash password")
	}

	// Update password
	err = s.repo.UpdatePassword(ctx, phone, string(hashedPassword))
	if err != nil {
		return errors.Internal("failed to update password")
	}

	// Invalidate all user sessions
	err = s.repo.RevokeAllUserTokens(ctx, phone)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to revoke user tokens")
	}

	return nil
}
