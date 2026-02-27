package profile

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/pkg/errors"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMyProfile returns the authenticated user's profile
// @Summary Get current user profile
// @Description Get profile of the authenticated user
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Profile
// @Failure 401 {object} map[string]interface{}
// @Router /users/me [get]
func (h *Handler) GetMyProfile(c *fiber.Ctx) error {
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
// @Summary Update current user profile
// @Description Update profile of the authenticated user
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Update Profile Request"
// @Success 200 {object} Profile
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me [put]
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
// @Summary Update user avatar
// @Description Update avatar URL of the authenticated user
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateAvatarRequest true "Update Avatar Request"
// @Success 200 {object} Profile
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me/avatar [put]
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
// @Summary Get public profile
// @Description Get public profile of any user by ID
// @Tags profile
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} PublicProfile
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /users/{id}/public [get]
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
// @Summary Update notification preferences
// @Description Update notification preferences of the authenticated user
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateNotificationPreferencesRequest true "Update Notification Preferences Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me/notification-preferences [put]
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

// RequestAvatarUploadURL requests a presigned URL for avatar upload
// @Summary Get avatar upload URL
// @Description Get a presigned URL for uploading an avatar to MinIO
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GetUploadURLRequest true "Get Upload URL Request"
// @Success 200 {object} UploadURLResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me/avatar/upload-url [post]
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
		ExpiresAt: uploadURL.ExpiresAt.Unix(),
	})
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	users.Get("/me", h.GetMyProfile)
	users.Put("/me", h.UpdateMyProfile)
	users.Put("/me/avatar", h.UpdateAvatar)
	users.Post("/me/avatar/upload-url", h.RequestAvatarUploadURL)
	users.Get("/:id/public", h.GetPublicProfile)
	users.Put("/me/notification-preferences", h.UpdateNotificationPreferences)
}
