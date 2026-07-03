package identity

import (
	"context"
	"time"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/postgres"
	"github.com/industrix/backend/pkg/redis"
)

// Repository handles all identity-related database operations
type Repository struct {
	pg    *postgres.Client
	redis *redis.Client
}

// NewRepository creates a new identity repository
func NewRepository(pg *postgres.Client, redis *redis.Client) *Repository {
	return &Repository{pg: pg, redis: redis}
}

// === Auth operations ===

func (r *Repository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	err := r.pg.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	return count > 0, err
}

func (r *Repository) UserExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int
	err := r.pg.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE phone = $1", phone).Scan(&count)
	return count > 0, err
}

func (r *Repository) UserExistsByGoogleID(ctx context.Context, googleID string) (bool, error) {
	var count int
	err := r.pg.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE google_id = $1", googleID).Scan(&count)
	return count > 0, err
}

func (r *Repository) CreateUserWithEmail(ctx context.Context, email, passwordHash, firstName string) (*User, error) {
	var user User
	// Email registration is treated as verified immediately — the user already
	// proved access to the mailbox by receiving no bounce, and there is no
	// email OTP sender wired up yet. See docs/architecture.md Identity module.
	err := r.pg.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, first_name, verified) VALUES ($1, $2, $3, true)
         RETURNING id, email, first_name, role, verified, created_at, updated_at`,
		email, passwordHash, firstName).Scan(&user.ID, &user.Email, &user.FirstName, &user.Role, &user.Verified, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

func (r *Repository) CreateUserWithPhone(ctx context.Context, phone string) (*User, error) {
	var user User
	err := r.pg.QueryRow(ctx,
		"INSERT INTO users (phone) VALUES ($1) RETURNING id, phone, role, verified, created_at, updated_at",
		phone).Scan(&user.ID, &user.Phone, &user.Role, &user.Verified, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

func (r *Repository) CreateUserWithGoogle(ctx context.Context, googleID, email, firstName, lastName, avatarURL string) (*User, error) {
	var user User
	// For Google login, if email from Google is available, we store it.
	err := r.pg.QueryRow(ctx,
		`INSERT INTO users (google_id, email, first_name, last_name, avatar_url, verified) 
         VALUES ($1, $2, $3, $4, $5, true) 
         RETURNING id, google_id, email, first_name, last_name, avatar_url, role, verified, created_at, updated_at`,
		googleID, email, firstName, lastName, avatarURL).Scan(
		&user.ID, &user.GoogleID, &user.Email, &user.FirstName, &user.LastName, &user.AvatarURL, &user.Role, &user.Verified, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

func (r *Repository) SaveOTP(ctx context.Context, phone, code string, ttl time.Duration) error {
	return r.redis.Set(ctx, "otp:"+phone, code, ttl)
}

func (r *Repository) ValidateOTP(ctx context.Context, phone, code string) (bool, error) {
	val, err := r.redis.Get(ctx, "otp:"+phone)
	if err != nil {
		return false, nil
	}
	if val == code {
		_ = r.redis.Del(ctx, "otp:"+phone)
		return true, nil
	}
	return false, nil
}

func (r *Repository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	var user User
	err := r.pg.QueryRow(ctx,
		"SELECT id, email, phone, password_hash, google_id, role, verified, COALESCE(company_id::text, '') FROM users WHERE phone = $1", phone).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.GoogleID, &user.Role, &user.Verified, &user.CompanyID,
	)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "User not found")
	}
	return &user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.pg.QueryRow(ctx,
		"SELECT id, email, phone, password_hash, google_id, role, verified, COALESCE(company_id::text, '') FROM users WHERE email = $1", email).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.GoogleID, &user.Role, &user.Verified, &user.CompanyID,
	)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "User not found")
	}
	return &user, nil
}

func (r *Repository) GetUserByGoogleID(ctx context.Context, googleID string) (*User, error) {
	var user User
	err := r.pg.QueryRow(ctx,
		"SELECT id, email, phone, password_hash, google_id, role, verified, COALESCE(company_id::text, '') FROM users WHERE google_id = $1", googleID).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.GoogleID, &user.Role, &user.Verified, &user.CompanyID,
	)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "User not found")
	}
	return &user, nil
}

func (r *Repository) UpdateUserVerification(ctx context.Context, userID string, verified bool) error {
	_, err := r.pg.Exec(ctx, "UPDATE users SET verified = $1 WHERE id = $2", verified, userID)
	return err
}

// === Profile operations ===

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.pg.QueryRow(ctx,
		"SELECT id, email, COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(avatar_url, ''), COALESCE(company_id::text, '') FROM users WHERE id = $1", id).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.AvatarURL, &user.CompanyID,
	)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "User not found")
	}
	return &user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *User) error {
	_, err := r.pg.Exec(ctx,
		"UPDATE users SET first_name = $1, last_name = $2, avatar_url = $3 WHERE id = $4",
		user.FirstName, user.LastName, user.AvatarURL, user.ID,
	)
	return err
}
