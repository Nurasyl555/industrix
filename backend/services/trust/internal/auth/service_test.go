package auth

import (
	"context"
	"testing"
	"time"

	"github.com/industrix/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) UserExists(ctx context.Context, email, phone string) (bool, error) {
	args := m.Called(ctx, email, phone)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) CreateUser(ctx context.Context, email, phone, passwordHash string) error {
	args := m.Called(ctx, email, phone, passwordHash)
	return args.Error(0)
}

func (m *MockRepository) SaveOTP(ctx context.Context, phone, code string, ttl time.Duration) error {
	args := m.Called(ctx, phone, code, ttl)
	return args.Error(0)
}

func (m *MockRepository) ValidateOTP(ctx context.Context, phone, code string) (bool, error) {
	args := m.Called(ctx, phone, code)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) UpdateUserVerification(ctx context.Context, userID string, verified bool) error {
	args := m.Called(ctx, userID, verified)
	return args.Error(0)
}

func (m *MockRepository) CheckPassword(hash, password string) bool {
	args := m.Called(hash, password)
	return args.Bool(0)
}

// Mock JWT Client
type MockJWTClient struct {
	mock.Mock
}

func (m *MockJWTClient) IssuePair(userID, companyID, role string, verified bool, accessExpiry, refreshExpiry time.Duration) (*jwt.TokenPair, error) {
	args := m.Called(userID, companyID, role, verified, accessExpiry, refreshExpiry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.TokenPair), args.Error(1)
}

func (m *MockJWTClient) ParseClaims(tokenString string) (*jwt.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

func TestRegister_Success(t *testing.T) {
	repo := new(MockRepository)
	jwtClient := new(MockJWTClient)
	svc := NewService(repo, jwtClient)

	ctx := context.Background()
	email := "test@example.com"
	phone := "+1234567890"
	password := "password123"

	repo.On("UserExists", ctx, email, phone).Return(false, nil)
	repo.On("CreateUser", ctx, email, phone, mock.AnythingOfType("string")).Return(nil)
	repo.On("SaveOTP", ctx, phone, mock.AnythingOfType("string"), 5*time.Minute).Return(nil)

	err := svc.Register(ctx, email, phone, password)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestVerifyOTP_Success(t *testing.T) {
	repo := new(MockRepository)
	jwtClient := new(MockJWTClient)
	svc := NewService(repo, jwtClient)

	ctx := context.Background()
	phone := "+1234567890"
	code := "123456"
	user := &User{ID: "user-123", Role: "user", Verified: false}

	repo.On("ValidateOTP", ctx, phone, code).Return(true, nil)
	repo.On("GetUserByPhone", ctx, phone).Return(user, nil)
	repo.On("UpdateUserVerification", ctx, user.ID, true).Return(nil)

	expectedTokens := &jwt.TokenPair{AccessToken: "access", RefreshToken: "refresh"}
	jwtClient.On("IssuePair", user.ID, "", "user", true, 15*time.Minute, 24*time.Hour).Return(expectedTokens, nil)

	tokens, err := svc.VerifyOTP(ctx, phone, code)

	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, tokens)
	repo.AssertExpectations(t)
	jwtClient.AssertExpectations(t)
}
