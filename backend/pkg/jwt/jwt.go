package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/industrix/pkg/errors"
)

type Claims struct {
	UserID    string `json:"user_id"`
	CompanyID string `json:"company_id,omitempty"`
	Role      string `json:"role"`
	Verified  bool   `json:"verified"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Client interface {
	IssuePair(userID, companyID, role string, verified bool, accessExpiry, refreshExpiry time.Duration) (*TokenPair, error)
	ParseClaims(tokenString string) (*Claims, error)
}

type client struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewClient(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) Client {
	return &client{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (c *client) IssuePair(userID, companyID, role string, verified bool, accessExpiry, refreshExpiry time.Duration) (*TokenPair, error) {
	now := time.Now()

	accessClaims := &Claims{
		UserID:    userID,
		CompanyID: companyID,
		Role:      role,
		Verified:  verified,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
			Issuer:    "industrix",
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims).SignedString(c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
			Issuer:    "industrix",
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims).SignedString(c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (c *client) ParseClaims(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New(errors.CodeUnauthorized, "token expired")
		}
		return nil, errors.New(errors.CodeUnauthorized, "invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New(errors.CodeUnauthorized, "invalid token claims")
}
