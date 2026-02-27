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
