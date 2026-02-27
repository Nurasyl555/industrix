package profile

// UpdateProfileRequest represents update profile request
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// UpdateAvatarRequest represents update avatar request
type UpdateAvatarRequest struct {
	AvatarURL string `json:"avatar_url"`
}

// GetUploadURLRequest represents get upload URL request
type GetUploadURLRequest struct {
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name"`
}

// UploadURLResponse represents upload URL response
type UploadURLResponse struct {
	UploadURL string            `json:"upload_url"`
	Fields    map[string]string `json:"fields"`
	ExpiresAt int64             `json:"expires_at"`
}

// UpdateNotificationPreferencesRequest represents update notification preferences request
type UpdateNotificationPreferencesRequest struct {
	Preferences map[string]bool `json:"preferences"`
}
