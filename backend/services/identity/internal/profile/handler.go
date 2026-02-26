package profile

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/pkg/errors"
)

// Handler handles HTTP requests for profile
type Handler struct {
	service *Service
}

// NewHandler creates a new profile handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMyProfile returns the authenticated user's profile
// GET /users/me
func (h *Handler) GetMyProfile(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "unauthorized"))
	}

	profile, err := h.service.GetProfile(c.Context(), userID.(string))
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(profile)
}

// UpdateMyProfile updates the authenticated user's profile
// PUT /users/me
func (h *Handler) UpdateMyProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "unauthorized"))
	}

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	profile, err := h.service.UpdateProfile(c.Context(), userID.(string), req)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(profile)
}

// UpdateAvatar updates the user's avatar
// PUT /users/me/avatar
func (h *Handler) UpdateAvatar(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "unauthorized"))
	}

	var req UpdateAvatarRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	profile, err := h.service.UpdateAvatar(c.Context(), userID.(string), req.AvatarURL)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(profile)
}

// GetPublicProfile returns a public profile
// GET /users/:id/public
func (h *Handler) GetPublicProfile(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "user id is required"))
	}

	profile, err := h.service.GetPublicProfile(c.Context(), userID)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(profile)
}

// UpdateNotificationPreferences updates notification preferences
// PUT /users/me/notification-preferences
func (h *Handler) UpdateNotificationPreferences(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "unauthorized"))
	}

	var req UpdateNotificationPreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	err := h.service.UpdateNotificationPreferences(c.Context(), userID.(string), req.Preferences)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(fiber.Map{"message": "notification preferences updated"})
}

// RequestAvatarUpload requests a presigned URL for avatar upload
// POST /users/me/avatar/upload-url
func (h *Handler) RequestAvatarUploadURL(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "unauthorized"))
	}

	var req GetUploadURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	uploadURL, err := h.service.GetAvatarUploadURL(c.Context(), userID.(string), req.ContentType, req.FileName)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(UploadURLResponse{
		UploadURL: uploadURL.URL,
		Fields:    uploadURL.Fields,
		ExpiresAt: uploadURL.ExpiresAt,
	})
}

// RegisterRoutes registers profile routes
func (h *Handler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	users.Get("/me", h.GetMyProfile)
	users.Put("/me", h.UpdateMyProfile)
	users.Put("/me/avatar", h.UpdateAvatar)
	users.Post("/me/avatar/upload-url", h.RequestAvatarUploadURL)
	users.Get("/:id/public", h.GetPublicProfile)
	users.Put("/me/notification-preferences", h.UpdateNotificationPreferences)
}
