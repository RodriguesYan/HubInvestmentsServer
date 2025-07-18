package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestNewTokenService(t *testing.T) {
	tokenService := NewTokenService()

	assert.NotNil(t, tokenService)
	assert.IsType(t, &TokenService{}, tokenService)
}

func TestTokenService_CreateAndSignToken_Success(t *testing.T) {
	service := &TokenService{}
	userName := "testuser"
	userId := "user123"

	token, err := service.CreateAndSignToken(userName, userId)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token can be parsed and contains expected claims
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userName, claims["username"])
	assert.Equal(t, userId, claims["userId"])
	assert.NotNil(t, claims["exp"])
}

func TestTokenService_CreateAndSignToken_WithEmptyValues(t *testing.T) {
	service := &TokenService{}

	// Test with empty values - should still create valid tokens
	token, err := service.CreateAndSignToken("", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestTokenService_ValidateToken_Success(t *testing.T) {
	service := &TokenService{}
	userName := "testuser"
	userId := "user123"

	// Create a valid token
	token, err := service.CreateAndSignToken(userName, userId)
	assert.NoError(t, err)

	// Validate the token
	claims, err := service.ValidateToken("Bearer " + token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userName, claims["username"])
	assert.Equal(t, userId, claims["userId"])
	assert.NotNil(t, claims["exp"])
}

func TestTokenService_ValidateToken_InvalidFormat(t *testing.T) {
	service := &TokenService{}

	claims, err := service.ValidateToken("Bearer invalid-jwt-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ValidateToken_ExpiredToken(t *testing.T) {
	service := &TokenService{}

	// Create a token that's already expired
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": "testuser",
			"userId":   "user123",
			"exp":      time.Now().Add(-time.Minute * 5).Unix(), // 5 minutes ago
		})

	tokenString, err := expiredToken.SignedString(secretKey)
	assert.NoError(t, err)

	// Try to validate the expired token
	claims, err := service.ValidateToken("Bearer " + tokenString)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ValidateToken_WrongKey(t *testing.T) {
	service := &TokenService{}
	wrongKey := []byte("wrong-secret-key")

	// Create a token signed with wrong key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": "testuser",
			"userId":   "user123",
			"exp":      time.Now().Add(time.Minute * 10).Unix(),
		})

	tokenString, err := token.SignedString(wrongKey)
	assert.NoError(t, err)

	// Try to validate the token signed with wrong key
	claims, err := service.ValidateToken("Bearer " + tokenString)

	assert.Error(t, err)
	assert.Nil(t, claims)
}
