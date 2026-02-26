package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/industrix/pkg/postgres"
	"github.com/industrix/pkg/redis"
)

// User represents a user in the system
type User struct {
	ID        string
	Email     string
	Phone     string
	Password  string
	Role      string
	Verified  bool
	CompanyID string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Repository handles database operations for auth
type Repository struct {
	pg    *postgres.Client
	redis *redis.Client
}

// NewRepository creates a new auth repository
func NewRepository(pg *postgres.Client, redis *redis.Client) *Repository {
	return &Repository{
		pg:    pg,
		redis: redis,
	}
}

// UserExists checks if a user exists with the given email or phone
func (r *Repository) UserExists(ctx context.Context, email, phone string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE email = $1 OR phone = $2
			AND active = true
		)
	`
	var exists bool
	err := r.pg.QueryRow(ctx, query, email, phone).Scan(&exists)
	return exists, err
}

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, email, phone, password, role string) (string, error) {
	userID := generateUUID()

	query := `
		INSERT INTO users (id, email, phone, password_hash, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, false, NOW(), NOW())
		RETURNING id
	`

	err := r.pg.QueryRow(ctx, query, userID, email, phone, password, role).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, phone, password_hash, role, verified, company_id, active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.pg.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.Role,
		&user.Verified,
		&user.CompanyID,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByPhone retrieves a user by phone
func (r *Repository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	query := `
		SELECT id, email, phone, password_hash, role, verified, company_id, active, created_at, updated_at
		FROM users
		WHERE phone = $1
	`

	var user User
	err := r.pg.QueryRow(ctx, query, phone).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.Role,
		&user.Verified,
		&user.CompanyID,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ActivateUser activates a user account
func (r *Repository) ActivateUser(ctx context.Context, phone string) error {
	query := `
		UPDATE users
		SET active = true, verified = true, updated_at = NOW()
		WHERE phone = $1
	`

	_, err := r.pg.Exec(ctx, query, phone)
	return err
}

// UpdatePassword updates user password
func (r *Repository) UpdatePassword(ctx context.Context, phoneOrEmail, newHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE phone = $1 OR email = $1
	`

	_, err := r.pg.Exec(ctx, query, phoneOrEmail, newHash)
	return err
}

// OTP methods using Redis

// StoreOTP stores an OTP in Redis
func (r *Repository) StoreOTP(ctx context.Context, key, otp string, ttl time.Duration) error {
	return r.redis.Set(ctx, "otp:"+key, otp, ttl)
}

// GetOTP retrieves an OTP from Redis
func (r *Repository) GetOTP(ctx context.Context, key string) (string, error) {
	return r.redis.Get(ctx, "otp:"+key)
}

// DeleteOTP deletes an OTP from Redis
func (r *Repository) DeleteOTP(ctx context.Context, key string) error {
	return r.redis.Del(ctx, "otp:"+key)
}

// Refresh token methods using Redis

// StoreRefreshToken stores a refresh token in Redis
func (r *Repository) StoreRefreshToken(ctx context.Context, userID, deviceID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s:%s", userID, deviceID)
	return r.redis.Set(ctx, key, token, ttl)
}

// GetRefreshToken retrieves a refresh token from Redis
func (r *Repository) GetRefreshToken(ctx context.Context, userID, deviceID string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s:%s", userID, deviceID)
	return r.redis.Get(ctx, key)
}

// IsRefreshTokenRevoked checks if a refresh token has been revoked
func (r *Repository) IsRefreshTokenRevoked(ctx context.Context, token string) (bool, error) {
	// Check if token is in the revoked set
	return r.redis.Exists(ctx, "revoked_token:"+token)
}

// RevokeRefreshToken revokes a refresh token
func (r *Repository) RevokeRefreshToken(ctx context.Context, token string) error {
	// Store in revoked set with same TTL as original token
	return r.redis.Set(ctx, "revoked_token:"+token, "1", 7*24*time.Hour)
}

// RevokeAllUserTokens revokes all tokens for a user
func (r *Repository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	// Get all device IDs for this user and revoke each
	pattern := fmt.Sprintf("refresh_token:%s:*", userID)
	keys, err := r.redis.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		token, err := r.redis.Get(ctx, key)
		if err != nil {
			continue
		}
		r.redis.Set(ctx, "revoked_token:"+token, "1", 7*24*time.Hour)
		r.redis.Del(ctx, key)
	}

	return nil
}

// Helper functions

// generateOTP generates a 6-digit OTP
func generateOTP() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	code := int32(0)
	for _, b := range bytes {
		code = code*256 + int32(b)
	}
	otp := code % 1000000
	return fmt.Sprintf("%06d", otp)
}

// generateUUID generates a UUID
func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return hex.EncodeToString(bytes)
}
