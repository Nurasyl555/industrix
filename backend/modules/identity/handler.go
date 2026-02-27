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
	auth.Post("/register", h.Register)
	auth.Post("/verify-otp", h.VerifyOTP)
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
}

// RegisterProtectedRoutes registers routes that require authentication
func (h *Handler) RegisterProtectedRoutes(router fiber.Router) {
	users := router.Group("/users")
	users.Get("/me", h.GetProfile)
	users.Put("/me", h.UpdateProfile)
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201
// @Failure 400 {object} errors.Error
// @Router /auth/register [post]
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	if err := h.service.Register(c.Context(), req.Email, req.Phone, req.Password); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	return c.SendStatus(http.StatusCreated)
}

// VerifyOTP godoc
// @Summary Verify phone OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "OTP details"
// @Success 200 {object} jwt.TokenPair
// @Router /auth/verify-otp [post]
func (h *Handler) VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	tokens, err := h.service.VerifyOTP(c.Context(), req.Phone, req.Code)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(err)
	}

	return c.JSON(tokens)
}

// Login godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} jwt.TokenPair
// @Router /auth/login [post]
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	tokens, err := h.service.Login(c.Context(), req.Email, req.Password)
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
