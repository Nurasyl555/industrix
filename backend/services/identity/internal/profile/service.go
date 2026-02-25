package profile

import (
	"context"
	"time"

	"github.com/industrix/pkg/errors"
	"github.com/industrix/pkg/kafka"
	"github.com/industrix/pkg/logger"
	"github.com/industrix/pkg/minio"
)

// Service handles business logic for profile
type Service struct {
	repo        *Repository
	kafkaClient *kafka.Producer
	minioClient *minio.Client
	log         *logger.Logger
}

// Profile represents a user profile
type Profile struct {
	UserID            string          `json:"user_id"`
	Email             string          `json:"email"`
	Phone             string          `json:"phone"`
	FirstName         string          `json:"first_name"`
	LastName          string          `json:"last_name"`
	AvatarURL         string          `json:"avatar_url"`
	CompanyID         string          `json:"company_id"`
	CompanyName       string          `json:"company_name"`
	CompanyVerified   bool            `json:"company_verified"`
	Role              string          `json:"role"`
	Verified          bool            `json:"verified"`
	NotificationPrefs map[string]bool `json:"notification_preferences"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// PublicProfile represents a public profile
type PublicProfile struct {
	UserID          string  `json:"user_id"`
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	AvatarURL       string  `json:"avatar_url"`
	CompanyName     string  `json:"company_name"`
	CompanyVerified bool    `json:"company_verified"`
	Rating          float64 `json:"rating"`
	ReviewsCount    int     `json:"reviews_count"`
}

// UploadURL represents an upload URL response
type UploadURL struct {
	URL       string
	Fields    map[string]string
	ExpiresAt time.Time
}

// NewService creates a new profile service
func NewService(repo *Repository, kafkaClient *kafka.Producer, minioClient *minio.Client) *Service {
	return &Service{
		repo:        repo,
		kafkaClient: kafkaClient,
		minioClient: minioClient,
		log:         logger.New("profile-service"),
	}
}

// GetProfile returns a user's profile
func (s *Service) GetProfile(ctx context.Context, userID string) (*Profile, *errors.Error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, errors.Internal("failed to get profile")
	}

	return profile, nil
}

// UpdateProfile updates a user's profile
func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*Profile, *errors.Error) {
	// Update profile in database
	profile, err := s.repo.UpdateProfile(ctx, userID, req)
	if err != nil {
		return nil, errors.Internal("failed to update profile")
	}

	// Emit profile updated event to Kafka
	if s.kafkaClient != nil {
		event := map[string]interface{}{
			"user_id":    userID,
			"event_type": "user.profile.updated",
			"timestamp":  time.Now().Unix(),
			"data": map[string]interface{}{
				"first_name": req.FirstName,
				"last_name":  req.LastName,
				"phone":      req.Phone,
			},
		}
		s.kafkaClient.Publish(ctx, "user.events", event)
	}

	return profile, nil
}

// UpdateAvatar updates a user's avatar
func (s *Service) UpdateAvatar(ctx context.Context, userID, avatarURL string) (*Profile, *errors.Error) {
	profile, err := s.repo.UpdateAvatar(ctx, userID, avatarURL)
	if err != nil {
		return nil, errors.Internal("failed to update avatar")
	}

	// Emit avatar updated event
	if s.kafkaClient != nil {
		event := map[string]interface{}{
			"user_id":    userID,
			"event_type": "user.avatar.updated",
			"timestamp":  time.Now().Unix(),
			"data": map[string]interface{}{
				"avatar_url": avatarURL,
			},
		}
		s.kafkaClient.Publish(ctx, "user.events", event)
	}

	return profile, nil
}

// GetAvatarUploadURL generates a presigned URL for avatar upload
func (s *Service) GetAvatarUploadURL(ctx context.Context, userID, contentType, fileName string) (*UploadURL, *errors.Error) {
	if s.minioClient == nil {
		return nil, errors.Internal("media service not configured")
	}

	objectName := "avatars/" + userID + "/" + fileName

	uploadURL, err := s.minioClient.PresignPutURL(ctx, objectName, 15*time.Minute)
	if err != nil {
		return nil, errors.Internal("failed to generate upload URL")
	}

	return &UploadURL{
		URL:       uploadURL,
		Fields:    nil,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil
}

// GetPublicProfile returns a public profile
func (s *Service) GetPublicProfile(ctx context.Context, userID string) (*PublicProfile, *errors.Error) {
	profile, err := s.repo.GetPublicProfile(ctx, userID)
	if err != nil {
		return nil, errors.NotFound("user not found")
	}

	return profile, nil
}

// UpdateNotificationPreferences updates notification preferences
func (s *Service) UpdateNotificationPreferences(ctx context.Context, userID string, prefs map[string]bool) *errors.Error {
	err := s.repo.UpdateNotificationPreferences(ctx, userID, prefs)
	if err != nil {
		return errors.Internal("failed to update notification preferences")
	}

	return nil
}
