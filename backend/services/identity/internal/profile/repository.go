package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/industrix/pkg/postgres"
)

// Repository handles database operations for profile
type Repository struct {
	pg *postgres.Client
}

// NewRepository creates a new profile repository
func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

// GetProfile retrieves a user's full profile
func (r *Repository) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	query := `
		SELECT 
			u.id, u.email, u.phone, u.first_name, u.last_name, u.avatar_url,
			u.company_id, u.role, u.verified, u.active, u.created_at, u.updated_at,
			c.name as company_name, c.verified as company_verified
		FROM users u
		LEFT JOIN companies c ON u.company_id = c.id
		WHERE u.id = $1
	`

	var profile Profile
	var firstName, lastName, avatarURL, companyID, companyName, phone *string

	err := r.pg.QueryRow(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.Email,
		&phone,
		&firstName,
		&lastName,
		&avatarURL,
		&companyID,
		&profile.Role,
		&profile.Verified,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&companyName,
		&profile.CompanyVerified,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	if firstName != nil {
		profile.FirstName = *firstName
	}
	if lastName != nil {
		profile.LastName = *lastName
	}
	if avatarURL != nil {
		profile.AvatarURL = *avatarURL
	}
	if companyID != nil {
		profile.CompanyID = *companyID
	}
	if companyName != nil {
		profile.CompanyName = *companyName
	}
	if phone != nil {
		profile.Phone = *phone
	}

	// Get notification preferences
	prefs, err := r.getNotificationPreferences(ctx, userID)
	if err == nil {
		profile.NotificationPrefs = prefs
	}

	return &profile, nil
}

// UpdateProfile updates a user's profile
func (r *Repository) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*Profile, error) {
	query := `
		UPDATE users
		SET first_name = COALESCE(NULLIF($2, ''), first_name),
			last_name = COALESCE(NULLIF($3, ''), last_name),
			phone = COALESCE(NULLIF($4, ''), phone),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, phone, first_name, last_name, avatar_url, company_id, role, verified, created_at, updated_at
	`

	var profile Profile
	err := r.pg.QueryRow(ctx, query, userID, req.FirstName, req.LastName, req.Phone).Scan(
		&profile.UserID,
		&profile.Email,
		&profile.Phone,
		&profile.FirstName,
		&profile.LastName,
		&profile.AvatarURL,
		&profile.CompanyID,
		&profile.Role,
		&profile.Verified,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &profile, nil
}

// UpdateAvatar updates a user's avatar URL
func (r *Repository) UpdateAvatar(ctx context.Context, userID, avatarURL string) (*Profile, error) {
	query := `
		UPDATE users
		SET avatar_url = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, phone, first_name, last_name, avatar_url, company_id, role, verified, created_at, updated_at
	`

	var profile Profile
	err := r.pg.QueryRow(ctx, query, userID, avatarURL).Scan(
		&profile.UserID,
		&profile.Email,
		&profile.Phone,
		&profile.FirstName,
		&profile.LastName,
		&profile.AvatarURL,
		&profile.CompanyID,
		&profile.Role,
		&profile.Verified,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}

	return &profile, nil
}

// GetPublicProfile retrieves a public profile
func (r *Repository) GetPublicProfile(ctx context.Context, userID string) (*PublicProfile, error) {
	query := `
		SELECT 
			u.id, u.first_name, u.last_name, u.avatar_url,
			c.name as company_name, c.verified as company_verified,
			COALESCE(AVG(r.rating), 0) as rating, COUNT(r.id) as reviews_count
		FROM users u
		LEFT JOIN companies c ON u.company_id = c.id
		LEFT JOIN reviews r ON r.target_user_id = u.id
		WHERE u.id = $1 AND u.active = true
		GROUP BY u.id, c.name, c.verified
	`

	var profile PublicProfile
	err := r.pg.QueryRow(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.FirstName,
		&profile.LastName,
		&profile.AvatarURL,
		&profile.CompanyName,
		&profile.CompanyVerified,
		&profile.Rating,
		&profile.ReviewsCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get public profile: %w", err)
	}

	return &profile, nil
}

// GetNotificationPreferences retrieves notification preferences
func (r *Repository) GetNotificationPreferences(ctx context.Context, userID string) (map[string]bool, error) {
	return r.getNotificationPreferences(ctx, userID)
}

func (r *Repository) getNotificationPreferences(ctx context.Context, userID string) (map[string]bool, error) {
	query := `
		SELECT notification_preferences
		FROM users
		WHERE id = $1
	`

	var prefsJSON []byte
	err := r.pg.QueryRow(ctx, query, userID).Scan(&prefsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification preferences: %w", err)
	}

	var prefs map[string]bool
	err = json.Unmarshal(prefsJSON, &prefs)
	if err != nil {
		// Return default preferences if parsing fails
		return map[string]bool{
			"push":  true,
			"email": true,
			"sms":   false,
		}, nil
	}

	return prefs, nil
}

// UpdateNotificationPreferences updates notification preferences
func (r *Repository) UpdateNotificationPreferences(ctx context.Context, userID string, prefs map[string]bool) error {
	prefsJSON, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `
		UPDATE users
		SET notification_preferences = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err = r.pg.Exec(ctx, query, userID, prefsJSON)
	if err != nil {
		return fmt.Errorf("failed to update notification preferences: %w", err)
	}

	return nil
}

// GetCompanyByID retrieves a company by ID
func (r *Repository) GetCompanyByID(ctx context.Context, companyID string) (*Company, error) {
	query := `
		SELECT id, name, bin, address, phone, email, website, verified, subscription_plan, created_at, updated_at
		FROM companies
		WHERE id = $1
	`

	var company Company
	err := r.pg.QueryRow(ctx, query, companyID).Scan(
		&company.ID,
		&company.Name,
		&company.BIN,
		&company.Address,
		&company.Phone,
		&company.Email,
		&company.Website,
		&company.Verified,
		&company.SubscriptionPlan,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return &company, nil
}

// Company represents a company
type Company struct {
	ID               string
	Name             string
	BIN              string
	Address          string
	Phone            string
	Email            string
	Website          string
	Verified         bool
	SubscriptionPlan string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
