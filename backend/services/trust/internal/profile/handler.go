package profile

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

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url"`
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Retrieve profile information for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} User
// @Failure 401 {object} errors.Error
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
// @Description Update name and avatar of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Updated profile"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.Error
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
		AvatarURL: req.AvatarURL,
	}

	if err := h.service.UpdateProfile(c.Context(), user); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	return c.SendStatus(http.StatusOK)
}

func (h *Handler) RegisterRoutes(app fiber.Router) {
	users := app.Group("/users")
	users.Get("/me", h.GetProfile)
	users.Put("/me", h.UpdateProfile)
}
