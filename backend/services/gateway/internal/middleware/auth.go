package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	trustpb "github.com/industrix/gen/go/trust/v1"
	"github.com/industrix/pkg/errors"
	"github.com/industrix/pkg/logger"
)

type AuthMiddleware struct {
	trustClient trustpb.TrustServiceClient
}

func NewAuth(client trustpb.TrustServiceClient) *AuthMiddleware {
	return &AuthMiddleware{trustClient: client}
}

func (m *AuthMiddleware) ValidateJWT() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "Missing Authorization header"))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "Invalid Authorization header format"))
		}

		token := parts[1]

		resp, err := m.trustClient.VerifyToken(context.Background(), &trustpb.VerifyTokenRequest{Token: token})
		if err != nil {
			logger.New("gateway").Error().Err(err).Msg("Token verification failed")
			return c.Status(fiber.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "Invalid token"))
		}

		c.Locals("user_id", resp.UserId)
		c.Locals("company_id", resp.CompanyId)
		c.Locals("role", resp.Role)
		c.Locals("verified", resp.Verified)

		return c.Next()
	}
}
