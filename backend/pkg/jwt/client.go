package jwt

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/industrix/pkg/logger"
)

// Config holds JWT configuration
type Config struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		AccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-access-secret-key"),
		RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "industrix",
	}
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID    string   `json:"user_id"`
	CompanyID string   `json:"company_id"`
	Role      string   `json:"role"`
	Verified  bool     `json:"verified"`
	Scope     []string `json:"scope"`
	jwt.RegisteredClaims
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Client handles JWT token operations
type Client struct {
	config *Config
	log    *logger.Logger
}

// NewClient creates a new JWT client
func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &Client{
		config: cfg,
		log:    logger.New("jwt-client"),
	}
}

// IssuePair generates a new access/refresh token pair
func (c *Client) IssuePair(ctx context.Context, userID, companyID, role string, verified bool, scope []string) (*TokenPair, error) {
	now := time.Now()

	// Access token
	accessClaims := Claims{
		UserID:    userID,
		CompanyID: companyID,
		Role:      role,
		Verified:  verified,
		Scope:     scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(c.config.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    c.config.Issuer,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(c.config.AccessSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := Claims{
		UserID:    userID,
		CompanyID: companyID,
		Role:      role,
		Verified:  verified,
		Scope:     scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(c.config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    c.config.Issuer,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(c.config.RefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(c.config.AccessExpiry),
	}, nil
}

// IssueAccessToken generates just an access token
func (c *Client) IssueAccessToken(ctx context.Context, userID, companyID, role string, verified bool, scope []string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:    userID,
		CompanyID: companyID,
		Role:      role,
		Verified:  verified,
		Scope:     scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(c.config.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    c.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.config.AccessSecret))
}

// ParseClaims parses and validates an access token
func (c *Client) ParseClaims(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.config.AccessSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ParseRefreshClaims parses and validates a refresh token
func (c *Client) ParseRefreshClaims(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.config.RefreshSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid refresh token")
}

// RefreshPair refreshes an access token using a refresh token
func (c *Client) RefreshPair(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := c.ParseRefreshClaims(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	return c.IssuePair(ctx, claims.UserID, claims.CompanyID, claims.Role, claims.Verified, claims.Scope)
}

// ValidateToken validates a token without extracting claims
func (c *Client) ValidateToken(ctx context.Context, tokenString string) bool {
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.config.AccessSecret), nil
	})
	return err == nil
}
