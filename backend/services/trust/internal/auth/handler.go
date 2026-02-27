package auth

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/pkg/errors"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with pending verification
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} map[string]string
// @Failure 400 {object} errors.Error
// @Failure 409 {object} errors.Error
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
// @Description Validates OTP sent to phone and issues initial tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "OTP details"
// @Success 200 {object} jwt.TokenPair
// @Failure 400 {object} errors.Error
// @Failure 401 {object} errors.Error
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
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} jwt.TokenPair
// @Failure 400 {object} errors.Error
// @Failure 401 {object} errors.Error
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
// @Description Issue new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} jwt.TokenPair
// @Failure 400 {object} errors.Error
// @Failure 401 {object} errors.Error
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

func (h *Handler) RegisterRoutes(app fiber.Router) {
	auth := app.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/verify-otp", h.VerifyOTP)
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
}
