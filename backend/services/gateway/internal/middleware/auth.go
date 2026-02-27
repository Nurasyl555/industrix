package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/industrix/services/gateway/internal/config"
)

const (
	ContextKeyUserID = "userID"
	ContextKeyRole   = "role"
	ContextKeyEmail  = "email"
	ContextKeyToken  = "token"
)

type AuthMiddleware struct {
	jwtSecret   string
	identityURL string
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:   getJWTSecret(),
		identityURL: cfg.Identity.URL,
	}
}

func getJWTSecret() string {
	secret := "your-256-bit-secret-key-here"
	return secret
}

// ValidateJWT validates the JWT token from Authorization header
func (m *AuthMiddleware) ValidateJWT() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Validate JWT signature
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "invalid signing method")
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token claims",
			})
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid user id in token",
			})
		}

		role, _ := claims["role"].(string)
		email, _ := claims["email"].(string)

		// Call Identity gRPC to check if token is revoked
		if m.isTokenRevoked(c.Context(), tokenString) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "token has been revoked",
			})
		}

		// Inject user info into request context
		c.Locals(ContextKeyUserID, userID)
		c.Locals(ContextKeyRole, role)
		c.Locals(ContextKeyEmail, email)
		c.Locals(ContextKeyToken, tokenString)

		return c.Next()
	}
}

// isTokenRevoked checks with Identity service if token is revoked
func (m *AuthMiddleware) isTokenRevoked(ctx context.Context, token string) bool {
	// TODO: Implement gRPC call to Identity service
	// conn, err := grpc.DialContext(ctx, m.identityURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	return false // Fail open, let service handle auth
	// }
	// defer conn.Close()

	// client := identity.NewIdentityClient(conn)
	// resp, err := client.VerifyToken(ctx, &identity.VerifyTokenRequest{Token: token})
	// if err != nil || !resp.Valid {
	// 	return true
	// }

	return false
}

// OptionalAuth validates JWT if present, but doesn't require it
func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}

		tokenString := parts[1]

		// Validate JWT signature
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "invalid signing method")
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Next()
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Next()
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			return c.Next()
		}

		role, _ := claims["role"].(string)
		email, _ := claims["email"].(string)

		// Inject user info into request context
		c.Locals(ContextKeyUserID, userID)
		c.Locals(ContextKeyRole, role)
		c.Locals(ContextKeyEmail, email)
		c.Locals(ContextKeyToken, tokenString)

		return c.Next()
	}
}

// RequireRole checks if user has the required role
func (m *AuthMiddleware) RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals(ContextKeyRole).(string)

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions",
		})
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) string {
	if userID, ok := c.Locals(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

// GetRole extracts role from context
func GetRole(c *fiber.Ctx) string {
	if role, ok := c.Locals(ContextKeyRole).(string); ok {
		return role
	}
	return ""
}

// GetEmail extracts email from context
func GetEmail(c *fiber.Ctx) string {
	if email, ok := c.Locals(ContextKeyEmail).(string); ok {
		return email
	}
	return ""
}
