package auth_test

import (
	"HubInvestments/internal/auth"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) ValidateToken(tokenString string) (map[string]interface{}, error) {
	args := m.Called(tokenString)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockTokenService) CreateAndSignToken(userName string, userId string) (string, error) {
	args := m.Called(userName, userId)
	return args.String(0), args.Error(1)
}

func TestNewAuthService(t *testing.T) {
	tokenService := &MockTokenService{}
	authService := auth.NewAuthService(tokenService)

	assert.NotNil(t, authService)
	assert.Implements(t, (*auth.IAuthService)(nil), authService)
}

func TestVerifyToken_Success(t *testing.T) {
	// Arrange
	tokenService := &MockTokenService{}
	tokenService.On("ValidateToken", "valid-token").Return(
		map[string]interface{}{"userId": "user123", "userName": "testuser"}, nil)

	authService := auth.NewAuthService(tokenService)
	rr := httptest.NewRecorder()

	// Act
	userId, err := authService.VerifyToken("valid-token", rr)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "user123", userId)
	assert.Equal(t, http.StatusOK, rr.Code)
	tokenService.AssertExpectations(t)
}

func TestVerifyToken_EmptyToken(t *testing.T) {
	// Arrange
	tokenService := &MockTokenService{}
	authService := auth.NewAuthService(tokenService)
	rr := httptest.NewRecorder()

	// Act
	userId, err := authService.VerifyToken("", rr)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "missing authorization header", err.Error())
	assert.Empty(t, userId)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Missing authorization header")

	// Ensure ValidateToken was not called
	tokenService.AssertNotCalled(t, "ValidateToken")
}

func TestVerifyToken_InvalidToken(t *testing.T) {
	// Arrange
	tokenService := &MockTokenService{}
	expectedError := errors.New("invalid token signature")
	tokenService.On("ValidateToken", "invalid-token").Return(
		map[string]interface{}(nil), expectedError)

	authService := auth.NewAuthService(tokenService)
	rr := httptest.NewRecorder()

	// Act
	userId, err := authService.VerifyToken("invalid-token", rr)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, userId)
	tokenService.AssertExpectations(t)
}

func TestVerifyToken_ExpiredToken(t *testing.T) {
	// Arrange
	tokenService := &MockTokenService{}
	expectedError := errors.New("token has expired")
	tokenService.On("ValidateToken", "expired-token").Return(
		map[string]interface{}(nil), expectedError)

	authService := auth.NewAuthService(tokenService)
	rr := httptest.NewRecorder()

	// Act
	userId, err := authService.VerifyToken("expired-token", rr)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, userId)
	tokenService.AssertExpectations(t)
}

func TestVerifyToken_MalformedClaims(t *testing.T) {
	// Test case where token is valid but userId claim is missing or wrong type
	testCases := []struct {
		name           string
		claims         map[string]interface{}
		expectedUserId string
	}{
		{
			name:           "missing userId claim",
			claims:         map[string]interface{}{"userName": "testuser"},
			expectedUserId: "",
		},
		{
			name:           "userId claim is not string",
			claims:         map[string]interface{}{"userId": 123, "userName": "testuser"},
			expectedUserId: "",
		},
		{
			name:           "nil userId claim",
			claims:         map[string]interface{}{"userId": nil, "userName": "testuser"},
			expectedUserId: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			tokenService := &MockTokenService{}
			tokenService.On("ValidateToken", "token-with-malformed-claims").Return(
				tc.claims, nil)

			authService := auth.NewAuthService(tokenService)
			rr := httptest.NewRecorder()

			// Act
			userId, err := authService.VerifyToken("token-with-malformed-claims", rr)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUserId, userId)
			tokenService.AssertExpectations(t)
		})
	}
}

func TestCreateToken_Success(t *testing.T) {
	// Arrange
	tokenService := &MockTokenService{}
	expectedToken := "generated-jwt-token"
	tokenService.On("CreateAndSignToken", "testuser", "user123").Return(
		expectedToken, nil)

	authService := auth.NewAuthService(tokenService)

	// Act
	token, err := authService.CreateToken("testuser", "user123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	tokenService.AssertExpectations(t)
}

func TestCreateToken_Error(t *testing.T) {
	// Arrange
	tokenService := &MockTokenService{}
	expectedError := errors.New("failed to sign token")
	tokenService.On("CreateAndSignToken", "testuser", "user123").Return(
		"", expectedError)

	authService := auth.NewAuthService(tokenService)

	// Act
	token, err := authService.CreateToken("testuser", "user123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, token)
	tokenService.AssertExpectations(t)
}

func TestCreateToken_EmptyParameters(t *testing.T) {
	testCases := []struct {
		name     string
		userName string
		userId   string
	}{
		{"empty userName", "", "user123"},
		{"empty userId", "testuser", ""},
		{"both empty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			tokenService := &MockTokenService{}
			expectedToken := "generated-token-even-with-empty-params"
			tokenService.On("CreateAndSignToken", tc.userName, tc.userId).Return(
				expectedToken, nil)

			authService := auth.NewAuthService(tokenService)

			// Act
			token, err := authService.CreateToken(tc.userName, tc.userId)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedToken, token)
			tokenService.AssertExpectations(t)
		})
	}
}

func TestVerifyToken_NilResponseWriter(t *testing.T) {
	// Test edge case where response writer is nil
	tokenService := &MockTokenService{}
	authService := auth.NewAuthService(tokenService)

	// This test documents that the current implementation will panic with nil response writer
	// In a production system, this should be fixed to handle nil gracefully
	assert.Panics(t, func() {
		authService.VerifyToken("", nil)
	}, "VerifyToken should panic with nil response writer in current implementation")
}

func TestAuth_IntegrationScenarios(t *testing.T) {
	t.Run("complete auth flow simulation", func(t *testing.T) {
		// Arrange
		tokenService := &MockTokenService{}
		authService := auth.NewAuthService(tokenService)

		// Mock successful token creation
		tokenService.On("CreateAndSignToken", "john.doe", "user456").Return(
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", nil)

		// Mock successful token validation
		tokenService.On("ValidateToken", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...").Return(
			map[string]interface{}{
				"userId":   "user456",
				"userName": "john.doe",
				"exp":      1234567890,
			}, nil)

		// Act & Assert - Create token
		token, err := authService.CreateToken("john.doe", "user456")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Act & Assert - Verify token
		rr := httptest.NewRecorder()
		userId, err := authService.VerifyToken(token, rr)
		assert.NoError(t, err)
		assert.Equal(t, "user456", userId)

		tokenService.AssertExpectations(t)
	})
}
