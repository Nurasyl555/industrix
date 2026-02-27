package repository

import (
	"context"
	"time"

	"github.com/industrix/pkg/errors"
	"github.com/industrix/pkg/postgres"
	"github.com/industrix/pkg/redis"
	"github.com/industrix/services/trust/internal/auth"
	"github.com/industrix/services/trust/internal/company"
	"github.com/industrix/services/trust/internal/profile"
	"github.com/industrix/services/trust/internal/review"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	pg    *postgres.Client
	redis *redis.Client
}

func NewRepository(pg *postgres.Client, r *redis.Client) *Repository {
	return &Repository{pg: pg, redis: r}
}

// === Auth Repository Implementation ===

func (r *Repository) UserExists(ctx context.Context, email, phone string) (bool, error) {
	var count int
	err := r.pg.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1 OR phone = $2", email, phone).Scan(&count)
	return count > 0, err
}

func (r *Repository) CreateUser(ctx context.Context, email, phone, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New(errors.CodeInternal, "Failed to hash password")
	}

	_, err = r.pg.Exec(ctx, "INSERT INTO users (email, phone, password_hash) VALUES ($1, $2, $3)", email, phone, string(hashedPassword))
	return err
}

func (r *Repository) SaveOTP(ctx context.Context, phone, code string, ttl time.Duration) error {
	return r.redis.Set(ctx, "otp:"+phone, code, ttl)
}

func (r *Repository) ValidateOTP(ctx context.Context, phone, code string) (bool, error) {
	val, err := r.redis.Get(ctx, "otp:"+phone)
	if err != nil {
		return false, nil // Not found or error
	}
	if val == code {
		_ = r.redis.Del(ctx, "otp:"+phone)
		return true, nil
	}
	return false, nil
}

func (r *Repository) GetUserByPhone(ctx context.Context, phone string) (*auth.User, error) {
	var user auth.User
	err := r.pg.QueryRow(ctx, "SELECT id, email, phone, password_hash, role, verified, company_id FROM users WHERE phone = $1", phone).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Role, &user.Verified, &user.CompanyID,
	)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "User not found")
	}
	return &user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	var user auth.User
	err := r.pg.QueryRow(ctx, "SELECT id, email, phone, password_hash, role, verified, company_id FROM users WHERE email = $1", email).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Role, &user.Verified, &user.CompanyID,
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

func (r *Repository) CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// === Profile Repository Implementation ===

func (r *Repository) GetUserByID(ctx context.Context, id string) (*profile.User, error) {
	var user profile.User
	err := r.pg.QueryRow(ctx, "SELECT id, email, first_name, last_name, avatar_url, company_id FROM users WHERE id = $1", id).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.AvatarURL, &user.CompanyID,
	)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "User not found")
	}
	return &user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *profile.User) error {
	_, err := r.pg.Exec(ctx, "UPDATE users SET first_name = $1, last_name = $2, avatar_url = $3 WHERE id = $4",
		user.FirstName, user.LastName, user.AvatarURL, user.ID)
	return err
}

// === Company Repository Implementation ===

func (r *Repository) CreateCompany(ctx context.Context, company *company.Company) error {
	err := r.pg.QueryRow(ctx,
		"INSERT INTO companies (name, bin, address, phone, email, website, owner_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at",
		company.Name, company.BIN, company.Address, company.Phone, company.Email, company.Website, company.OwnerID).
		Scan(&company.ID, &company.CreatedAt)
	return err
}

func (r *Repository) GetCompanyByID(ctx context.Context, id string) (*company.Company, error) {
	var c company.Company
	err := r.pg.QueryRow(ctx, "SELECT id, name, bin, address, phone, email, website, verified, created_at, owner_id FROM companies WHERE id = $1", id).Scan(
		&c.ID, &c.Name, &c.BIN, &c.Address, &c.Phone, &c.Email, &c.Website, &c.Verified, &c.CreatedAt, &c.OwnerID,
	)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Company not found")
	}
	return &c, nil
}

func (r *Repository) GetCompanyByBIN(ctx context.Context, bin string) (*company.Company, error) {
	var c company.Company
	err := r.pg.QueryRow(ctx, "SELECT id, name, bin, address, phone, email, website, verified, created_at, owner_id FROM companies WHERE bin = $1", bin).Scan(
		&c.ID, &c.Name, &c.BIN, &c.Address, &c.Phone, &c.Email, &c.Website, &c.Verified, &c.CreatedAt, &c.OwnerID,
	)
	if err != nil {
		return nil, nil // Return nil if not found, let service handle
	}
	return &c, nil
}

func (r *Repository) UpdateCompany(ctx context.Context, company *company.Company) error {
	_, err := r.pg.Exec(ctx, "UPDATE companies SET name = $1, address = $2, phone = $3, email = $4, website = $5 WHERE id = $6",
		company.Name, company.Address, company.Phone, company.Email, company.Website, company.ID)
	return err
}

// === Review Repository Implementation ===

func (r *Repository) CreateReview(ctx context.Context, rev *review.Review) error {
	err := r.pg.QueryRow(ctx,
		"INSERT INTO reviews (author_id, target_entity_id, rating, comment, transaction_id) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at",
		rev.AuthorID, rev.TargetEntityID, rev.Rating, rev.Comment, rev.TransactionID).
		Scan(&rev.ID, &rev.CreatedAt)
	return err
}

func (r *Repository) GetReviewsByEntity(ctx context.Context, entityID string, page, limit int) ([]*review.Review, int64, error) {
	offset := (page - 1) * limit
	rows, err := r.pg.Query(ctx, "SELECT id, author_id, target_entity_id, rating, comment, created_at FROM reviews WHERE target_entity_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", entityID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []*review.Review
	for rows.Next() {
		var rev review.Review
		if err := rows.Scan(&rev.ID, &rev.AuthorID, &rev.TargetEntityID, &rev.Rating, &rev.Comment, &rev.CreatedAt); err != nil {
			continue
		}
		reviews = append(reviews, &rev)
	}

	var total int64
	_ = r.pg.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE target_entity_id = $1", entityID).Scan(&total)

	return reviews, total, nil
}

func (r *Repository) GetReputationScore(ctx context.Context, entityID string) (*review.ReputationScore, error) {
	var s review.ReputationScore
	err := r.pg.QueryRow(ctx, "SELECT entity_id, average_rating, review_count, tier FROM reputation_scores WHERE entity_id = $1", entityID).Scan(
		&s.EntityID, &s.AverageRating, &s.ReviewCount, &s.Tier,
	)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Reputation not found")
	}
	return &s, nil
}

func (r *Repository) UpdateReputationScore(ctx context.Context, score *review.ReputationScore) error {
	_, err := r.pg.Exec(ctx,
		`INSERT INTO reputation_scores (entity_id, average_rating, review_count, tier, last_updated)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (entity_id) DO UPDATE SET
		 average_rating = EXCLUDED.average_rating,
		 review_count = EXCLUDED.review_count,
		 tier = EXCLUDED.tier,
		 last_updated = NOW()`,
		score.EntityID, score.AverageRating, score.ReviewCount, score.Tier)
	return err
}
