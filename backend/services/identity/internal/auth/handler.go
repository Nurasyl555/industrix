package auth

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/pkg/errors"
)

// Handler handles HTTP requests for auth
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRequest represents register request
type RegisterRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"` // BUYER, SELLER, SERVICE_COMPANY
}

// RegisterResponse represents register response
type RegisterResponse struct {
	UserID string `json:"user_id"`
	Status string `json:"status"` // PENDING_OTP
}

// Register handles user registration
// POST /auth/register
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	// Validate required fields
	if req.Email == "" && req.Phone == "" {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "email or phone is required"))
	}
	if req.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "password is required"))
	}

	result, err := h.service.Register(c.Context(), req.Email, req.Phone, req.Password, req.Role)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.Status(http.StatusCreated).JSON(RegisterResponse{
		UserID: result.UserID,
		Status: result.Status,
	})
}

// VerifyOTPRequest represents verify OTP request
type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

// VerifyOTPResponse represents verify OTP response
type VerifyOTPResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	UserID       string `json:"user_id"`
}

// VerifyOTP handles OTP verification
// POST /auth/verify-otp
func (h *Handler) VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	if req.Phone == "" || req.OTP == "" {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "phone and OTP are required"))
	}

	result, err := h.service.VerifyOTP(c.Context(), req.Phone, req.OTP)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(VerifyOTPResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
		UserID:       result.UserID,
	})
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"` // Optional: for refresh token tracking
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	UserID       string `json:"user_id"`
}

// Login handles user login
// POST /auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "email and password are required"))
	}

	result, err := h.service.Login(c.Context(), req.Email, req.Password, req.DeviceID)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
		UserID:       result.UserID,
	})
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponse represents token refresh response
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// Refresh handles token refresh
// POST /auth/refresh
func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	if req.RefreshToken == "" {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "refresh token is required"))
	}

	result, err := h.service.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
	})
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	AllSessions  bool   `json:"all_sessions"` // Invalidate all user sessions
}

// Logout handles user logout
// POST /auth/logout
func (h *Handler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	err := h.service.Logout(c.Context(), req.RefreshToken, req.AllSessions)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(fiber.Map{"message": "logged out successfully"})
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// ForgotPassword handles forgot password
// POST /auth/forgot-password
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	if req.Email == "" && req.Phone == "" {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "email or phone is required"))
	}

	err := h.service.ForgotPassword(c.Context(), req.Email, req.Phone)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(fiber.Map{"message": "reset token sent"})
}

// ResetPasswordRequest represents reset password request
type ResetPasswordRequest struct {
	Phone       string `json:"phone"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

// ResetPassword handles password reset
// POST /auth/reset-password
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	if req.Phone == "" || req.OTP == "" || req.NewPassword == "" {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "phone, OTP, and new password are required"))
	}

	err := h.service.ResetPassword(c.Context(), req.Phone, req.OTP, req.NewPassword)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(fiber.Map{"message": "password reset successfully"})
}

// RegisterRoutes registers auth routes
func (h *Handler) RegisterRoutes(router fiber.Router) {
	auth := router.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/verify-otp", h.VerifyOTP)
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)
	auth.Post("/forgot-password", h.ForgotPassword)
	auth.Post("/reset-password", h.ResetPassword)
}
