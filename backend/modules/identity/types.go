package identity

import "time"

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Email        *string   `json:"email"`
	Phone        *string   `json:"phone"`
	PasswordHash *string   `json:"-"`
	GoogleID     *string   `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Role         string    `json:"role"`
	Verified     bool      `json:"verified"`
	AvatarURL    string    `json:"avatar_url"`
	CompanyID    string    `json:"company_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// EmailRegisterRequest represents email registration request
type EmailRegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
}

// EmailLoginRequest represents email login request
type EmailLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PhoneLoginRequest represents request for OTP to login/register via phone
type PhoneLoginRequest struct {
	Phone string `json:"phone"`
}

// VerifyOTPRequest represents OTP verification request
type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
