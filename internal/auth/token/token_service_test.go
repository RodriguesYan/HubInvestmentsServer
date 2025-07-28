package token

import (
	"testing"
	"time"

	"HubInvestments/shared/config"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestNewTokenService(t *testing.T) {
	service := NewTokenService()
	assert.NotNil(t, service)
}

func TestTokenService_CreateAndSignToken_Success(t *testing.T) {
	service := NewTokenService()

	token, err := service.CreateAndSignToken("testuser", "user123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token can be parsed and contains expected claims
	cfg := config.Get()
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Check claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, "testuser", claims["username"])
	assert.Equal(t, "user123", claims["userId"])
	assert.NotNil(t, claims["exp"])
}

func TestTokenService_CreateAndSignToken_WithEmptyValues(t *testing.T) {
	service := NewTokenService()

	token, err := service.CreateAndSignToken("", "")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should still be valid even with empty values
	cfg := config.Get()
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}

func TestTokenService_ValidateToken_Success(t *testing.T) {
	service := NewTokenService()

	// Create a valid token
	token, err := service.CreateAndSignToken("testuser", "user123")
	assert.NoError(t, err)

	// Validate the token
	claims, err := service.ValidateToken("Bearer " + token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "testuser", claims["username"])
	assert.Equal(t, "user123", claims["userId"])
}

func TestTokenService_ValidateToken_InvalidFormat(t *testing.T) {
	service := NewTokenService()

	// Test with invalid Bearer format
	claims, err := service.ValidateToken("InvalidFormat")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ValidateToken_ExpiredToken(t *testing.T) {
	service := NewTokenService()

	// Create an expired token manually
	cfg := config.Get()
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": "testuser",
			"userId":   "user123",
			"exp":      time.Now().Add(-time.Minute * 5).Unix(), // 5 minutes ago
		})

	tokenString, err := expiredToken.SignedString([]byte(cfg.JWTSecret))
	assert.NoError(t, err)

	// Try to validate the expired token
	claims, err := service.ValidateToken("Bearer " + tokenString)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ValidateToken_WrongKey(t *testing.T) {
	service := NewTokenService()

	// Create a token with a different secret key
	wrongToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": "testuser",
			"userId":   "user123",
			"exp":      time.Now().Add(time.Minute * 10).Unix(),
		})

	tokenString, err := wrongToken.SignedString([]byte("wrong-secret-key"))
	assert.NoError(t, err)

	// Try to validate the token with the correct service (which uses the correct key)
	claims, err := service.ValidateToken("Bearer " + tokenString)

	assert.Error(t, err)
	assert.Nil(t, claims)
}
