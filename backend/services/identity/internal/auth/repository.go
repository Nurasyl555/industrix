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

type User struct {
	ID        string
	Email     string
	Phone     string
	Password  string
	FirstName string
	LastName  string
	Role      string
	Verified  bool
	CompanyID string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository struct {
	pg    *postgres.Client
	redis *redis.Client
}

func NewRepository(pg *postgres.Client, redis *redis.Client) *Repository {
	return &Repository{pg: pg, redis: redis}
}

func (r *Repository) UserExists(ctx context.Context, email, phone string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR phone = $2 AND active = true)`
	var exists bool
	err := r.pg.QueryRow(ctx, query, email, phone).Scan(&exists)
	return exists, err
}

func (r *Repository) CreateUser(ctx context.Context, email, phone, password, role string) (string, error) {
	userID := generateUUID()
	query := `INSERT INTO users (id, email, phone, password_hash, role, active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, false, NOW(), NOW()) RETURNING id`
	err := r.pg.QueryRow(ctx, query, userID, email, phone, password, role).Scan(&userID)
	if err != nil { return "", fmt.Errorf("failed to create user: %w", err) }
	return userID, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, phone, password_hash, first_name, last_name, role, verified, company_id, active, created_at, updated_at FROM users WHERE email = $1`
	var user User
	err := r.pg.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName, &user.LastName, &user.Role, &user.Verified, &user.CompanyID, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil { return nil, err }
	return &user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, email, phone, password_hash, first_name, last_name, role, verified, company_id, active, created_at, updated_at FROM users WHERE id = $1`
	var user User
	err := r.pg.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName, &user.LastName, &user.Role, &user.Verified, &user.CompanyID, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil { return nil, err }
	return &user, nil
}

func (r *Repository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	query := `SELECT id, email, phone, password_hash, first_name, last_name, role, verified, company_id, active, created_at, updated_at FROM users WHERE phone = $1`
	var user User
	err := r.pg.QueryRow(ctx, query, phone).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName, &user.LastName, &user.Role, &user.Verified, &user.CompanyID, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil { return nil, err }
	return &user, nil
}

func (r *Repository) ActivateUser(ctx context.Context, phone string) error {
	query := `UPDATE users SET active = true, verified = true, updated_at = NOW() WHERE phone = $1`
	_, err := r.pg.Exec(ctx, query, phone)
	return err
}

func (r *Repository) UpdatePassword(ctx context.Context, phoneOrEmail, newHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE phone = $1 OR email = $1`
	_, err := r.pg.Exec(ctx, query, phoneOrEmail, newHash)
	return err
}

func (r *Repository) StoreOTP(ctx context.Context, key, otp string, ttl time.Duration) error {
	return r.redis.Set(ctx, "otp:"+key, otp, ttl)
}
func (r *Repository) GetOTP(ctx context.Context, key string) (string, error) {
	return r.redis.Get(ctx, "otp:"+key)
}
func (r *Repository) DeleteOTP(ctx context.Context, key string) error {
	return r.redis.Del(ctx, "otp:"+key)
}

func (r *Repository) StoreRefreshToken(ctx context.Context, userID, deviceID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s:%s", userID, deviceID)
	return r.redis.Set(ctx, key, token, ttl)
}
func (r *Repository) GetRefreshToken(ctx context.Context, userID, deviceID string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s:%s", userID, deviceID)
	return r.redis.Get(ctx, key)
}
func (r *Repository) IsRefreshTokenRevoked(ctx context.Context, token string) (bool, error) {
	return r.redis.Exists(ctx, "revoked_token:"+token)
}
func (r *Repository) RevokeRefreshToken(ctx context.Context, token string) error {
	return r.redis.Set(ctx, "revoked_token:"+token, "1", 7*24*time.Hour)
}
func (r *Repository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("refresh_token:%s:*", userID)
	keys, err := r.redis.Keys(ctx, pattern)
	if err != nil { return err }
	for _, key := range keys {
		token, _ := r.redis.Get(ctx, key)
		r.redis.Set(ctx, "revoked_token:"+token, "1", 7*24*time.Hour)
		r.redis.Del(ctx, key)
	}
	return nil
}

func generateOTP() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	code := int32(0)
	for _, b := range bytes { code = code*256 + int32(b) }
	return fmt.Sprintf("%06d", code%1000000)
}
func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return hex.EncodeToString(bytes)
}
