package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/jwt"
	"github.com/industrix/backend/pkg/logger"
)

type AuthMiddleware struct {
	jwtClient jwt.Client
	log       *logger.Logger
}

func NewAuth(jwtClient jwt.Client) *AuthMiddleware {
	return &AuthMiddleware{
		jwtClient: jwtClient,
		log:       logger.New("auth-middleware"),
	}
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

		claims, err := m.jwtClient.ParseClaims(token)
		if err != nil {
			m.log.Error().Err(err).Msg("Token verification failed")
			return c.Status(fiber.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "Invalid token"))
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("company_id", claims.CompanyID)
		c.Locals("role", claims.Role)
		c.Locals("verified", claims.Verified)

		return c.Next()
	}
}

// RequireAdmin gates a route to admin users. Must run AFTER ValidateJWT so the
// role is already in Locals.
func (m *AuthMiddleware) RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(errors.New(errors.CodeUnauthorized, "Admin access required"))
		}
		return c.Next()
	}
}
