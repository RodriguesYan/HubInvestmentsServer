package auth_test

import (
	"HubInvestments/auth"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTokenService struct {
	mock.Mock
	shouldReturnError bool
}

func (m *MockTokenService) ValidateToken(tokenString string) (map[string]interface{}, error) {
	if m.shouldReturnError {
		return nil, assert.AnError
	}

	return map[string]interface{}{"userId": "user123"}, nil
}

func (s *MockTokenService) CreateAndSignToken(userName string, userId string) (string, error) {
	return "token", nil
}

func TestVerifyToken_Success(t *testing.T) {
	tokenService := MockTokenService{}

	authService := auth.NewAuthService(&tokenService)
	rr := httptest.NewRecorder()
	userId, err := authService.VerifyToken("Bearer token", rr)

	assert.NoError(t, err)
	assert.Equal(t, "user123", userId)
}

func TestVerifyToken_Error(t *testing.T) {
	tokenService := MockTokenService{shouldReturnError: true}

	authService := auth.NewAuthService(&tokenService)
	rr := httptest.NewRecorder()
	userId, err := authService.VerifyToken("Bearer token", rr)

	assert.Error(t, err)
	assert.Empty(t, userId)
}
