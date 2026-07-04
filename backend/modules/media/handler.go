package media

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/industrix/backend/pkg/errors"
)

// presigner is the subset of the minio client the handler needs (kept small
// so the module is easy to test/mock).
type presigner interface {
	PresignPutURL(ctx context.Context, bucket, object string, expiry time.Duration) (string, error)
}

// Handler serves media upload endpoints.
type Handler struct {
	presign    presigner
	publicBase string // e.g. http://localhost:9000/equipment-media
}

func NewHandler(presign presigner, publicBase string) *Handler {
	return &Handler{presign: presign, publicBase: publicBase}
}

// RegisterRoutes mounts the protected media routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Post("/media/upload-url", h.UploadURL)
}

// allowedExt maps a content type to a file extension. Restricting types keeps
// the bucket to images only.
var allowedExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type uploadURLRequest struct {
	ContentType string `json:"content_type"`
}

type uploadURLResponse struct {
	UploadURL string `json:"upload_url"` // presigned PUT — browser uploads the bytes here
	PublicURL string `json:"public_url"` // where the image will be readable afterwards
}

// UploadURL godoc
// @Summary Get a presigned URL to upload an equipment image directly to storage
// @Tags media
// @Security BearerAuth
// @Param request body uploadURLRequest true "Image content type"
// @Success 200 {object} uploadURLResponse
// @Router /media/upload-url [post]
func (h *Handler) UploadURL(c *fiber.Ctx) error {
	var req uploadURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}
	ext, ok := allowedExt[req.ContentType]
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Only JPEG, PNG or WebP images are allowed"))
	}

	object := uuid.New().String() + ext
	url, err := h.presign.PresignPutURL(c.Context(), bucketName, object, 5*time.Minute)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Failed to create upload URL"))
	}

	return c.JSON(uploadURLResponse{
		UploadURL: url,
		PublicURL: h.publicBase + "/" + object,
	})
}
