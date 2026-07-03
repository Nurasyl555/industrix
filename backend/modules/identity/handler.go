package identity

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
)

// Handler handles identity HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new identity handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterPublicRoutes registers routes that don't require authentication
func (h *Handler) RegisterPublicRoutes(router fiber.Router) {
	auth := router.Group("/auth")

	// Email flow
	auth.Post("/email/register", h.RegisterEmail)
	auth.Post("/email/login", h.LoginEmail)

	// Phone flow
	auth.Post("/phone/login", h.LoginPhone)
	auth.Post("/phone/verify", h.VerifyPhone)

	// Google OAuth flow
	auth.Get("/oauth/google", h.GoogleOAuthLogin)
	auth.Get("/oauth/google/callback", h.GoogleOAuthCallback)

	auth.Post("/refresh", h.Refresh)
}

// RegisterProtectedRoutes registers routes that require authentication
func (h *Handler) RegisterProtectedRoutes(router fiber.Router) {
	users := router.Group("/users")
	users.Get("/me", h.GetProfile)
	users.Put("/me", h.UpdateProfile)
}

// RegisterEmail godoc
// @Summary Register a new user via Email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body EmailRegisterRequest true "Registration details"
// @Success 200 {object} jwt.TokenPair
// @Failure 400 {object} errors.Error
// @Router /auth/email/register [post]
func (h *Handler) RegisterEmail(c *fiber.Ctx) error {
	var req EmailRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	tokens, err := h.service.RegisterEmail(c.Context(), req.Email, req.Password, req.FirstName)
	if err != nil {
		if domainErr, ok := err.(*errors.Error); ok {
			return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
		}
		return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Registration failed"))
	}

	return c.JSON(tokens)
}

// LoginEmail godoc
// @Summary Login user via Email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body EmailLoginRequest true "Login credentials"
// @Success 200 {object} jwt.TokenPair
// @Router /auth/email/login [post]
func (h *Handler) LoginEmail(c *fiber.Ctx) error {
	var req EmailLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	tokens, err := h.service.LoginEmail(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(err)
	}

	return c.JSON(tokens)
}

// LoginPhone godoc
// @Summary Request OTP for Phone login/registration
// @Tags auth
// @Accept json
// @Produce json
// @Param request body PhoneLoginRequest true "Phone number"
// @Success 200
// @Router /auth/phone/login [post]
func (h *Handler) LoginPhone(c *fiber.Ctx) error {
	var req PhoneLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	if err := h.service.RequestPhoneOTP(c.Context(), req.Phone); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	return c.SendStatus(http.StatusOK)
}

// VerifyPhone godoc
// @Summary Verify phone OTP and get tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "OTP details"
// @Success 200 {object} jwt.TokenPair
// @Router /auth/phone/verify [post]
func (h *Handler) VerifyPhone(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	tokens, err := h.service.VerifyPhoneOTP(c.Context(), req.Phone, req.Code)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(err)
	}

	return c.JSON(tokens)
}

// GoogleOAuthLogin godoc
// @Summary Redirect to Google OAuth consent screen
// @Tags auth
// @Router /auth/oauth/google [get]
func (h *Handler) GoogleOAuthLogin(c *fiber.Ctx) error {
	// TODO: Implement actual redirect using golang.org/x/oauth2
	return c.SendString("Redirecting to Google...")
}

// GoogleOAuthCallback godoc
// @Summary Handle Google OAuth callback and get tokens
// @Tags auth
// @Param code query string true "OAuth code"
// @Success 200 {object} jwt.TokenPair
// @Router /auth/oauth/google/callback [get]
func (h *Handler) GoogleOAuthCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Missing code"))
	}

	tokens, err := h.service.LoginGoogle(c.Context(), code)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(err)
	}

	return c.JSON(tokens)
}

// Refresh godoc
// @Summary Refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} jwt.TokenPair
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	tokens, err := h.service.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(err)
	}

	return c.JSON(tokens)
}

// GetProfile godoc
// @Summary Get current user profile
// @Tags users
// @Security BearerAuth
// @Success 200 {object} User
// @Router /users/me [get]
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	profile, err := h.service.GetProfile(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(err)
	}
	return c.JSON(profile)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Tags users
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Updated profile"
// @Success 200
// @Router /users/me [put]
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	user := &User{
		ID:        userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := h.service.UpdateProfile(c.Context(), user); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	return c.SendStatus(http.StatusOK)
}
