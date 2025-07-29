package http

import (
	"HubInvestments/internal/login/domain/model"
	"HubInvestments/internal/login/domain/valueobject"
	di "HubInvestments/pck"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDoLoginUsecase mocks the login usecase
type MockDoLoginUsecase struct {
	mock.Mock
}

func (m *MockDoLoginUsecase) Execute(email string, password string) (*model.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// MockAuthService mocks the auth service
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) VerifyToken(tokenString string, w http.ResponseWriter) (string, error) {
	args := m.Called(tokenString, w)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) CreateToken(userName string, userId string) (string, error) {
	args := m.Called(userName, userId)
	return args.String(0), args.Error(1)
}

// Helper function to create a test user
func createTestUser() *model.User {
	email := valueobject.NewEmailFromRepository("test@example.com")
	password := valueobject.NewPasswordFromRepository("password123")
	return &model.User{
		ID:       "user123",
		Email:    email,
		Password: password,
	}
}

func TestDoLogin_Success(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockDoLoginUsecase)
	mockAuthService := new(MockAuthService)

	testUser := createTestUser()

	mockLoginUsecase.On("Execute", "test@example.com", "password123").Return(testUser, nil)
	mockAuthService.On("CreateToken", "test@example.com", "user123").Return("mock-token-123", nil)

	container := di.NewTestContainer().
		WithLoginUsecase(mockLoginUsecase).
		WithAuthService(mockAuthService)

	loginRequest := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "mock-token-123", response["token"])

	mockLoginUsecase.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestDoLogin_InvalidRequestBody(t *testing.T) {
	// Arrange
	container := di.NewTestContainer()

	// Invalid JSON
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid request body")
}

func TestDoLogin_EmptyRequestBody(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockDoLoginUsecase)

	// Empty request body should result in empty email and password being passed to usecase
	mockLoginUsecase.On("Execute", "", "").Return(nil, errors.New("email and password are required"))

	container := di.NewTestContainer().WithLoginUsecase(mockLoginUsecase)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid credentials")

	mockLoginUsecase.AssertExpectations(t)
}

func TestDoLogin_InvalidCredentials(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockDoLoginUsecase)

	mockLoginUsecase.On("Execute", "test@example.com", "wrongpassword").Return(nil, errors.New("invalid password"))

	container := di.NewTestContainer().WithLoginUsecase(mockLoginUsecase)

	loginRequest := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid credentials")

	mockLoginUsecase.AssertExpectations(t)
}

func TestDoLogin_UserNotFound(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockDoLoginUsecase)

	mockLoginUsecase.On("Execute", "nonexistent@example.com", "password123").Return(nil, errors.New("user not found"))

	container := di.NewTestContainer().WithLoginUsecase(mockLoginUsecase)

	loginRequest := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid credentials")

	mockLoginUsecase.AssertExpectations(t)
}

func TestDoLogin_TokenGenerationFailure(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockDoLoginUsecase)
	mockAuthService := new(MockAuthService)

	testUser := createTestUser()

	mockLoginUsecase.On("Execute", "test@example.com", "password123").Return(testUser, nil)
	mockAuthService.On("CreateToken", "test@example.com", "user123").Return("", errors.New("token service error"))

	container := di.NewTestContainer().
		WithLoginUsecase(mockLoginUsecase).
		WithAuthService(mockAuthService)

	loginRequest := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to generate token")

	mockLoginUsecase.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestDoLogin_MissingEmailField(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockDoLoginUsecase)

	// Call with empty email should result in authentication failure
	mockLoginUsecase.On("Execute", "", "password123").Return(nil, errors.New("email is required"))

	container := di.NewTestContainer().WithLoginUsecase(mockLoginUsecase)

	loginRequest := map[string]string{
		"password": "password123",
		// email field is missing
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid credentials")

	mockLoginUsecase.AssertExpectations(t)
}

func TestDoLogin_MissingPasswordField(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockDoLoginUsecase)

	// Call with empty password should result in authentication failure
	mockLoginUsecase.On("Execute", "test@example.com", "").Return(nil, errors.New("password is required"))

	container := di.NewTestContainer().WithLoginUsecase(mockLoginUsecase)

	loginRequest := map[string]string{
		"email": "test@example.com",
		// password field is missing
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid credentials")

	mockLoginUsecase.AssertExpectations(t)
}

func TestDoLogin_ContentTypeHeader(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockDoLoginUsecase)
	mockAuthService := new(MockAuthService)

	testUser := createTestUser()

	mockLoginUsecase.On("Execute", "test@example.com", "password123").Return(testUser, nil)
	mockAuthService.On("CreateToken", "test@example.com", "user123").Return("mock-token-123", nil)

	container := di.NewTestContainer().
		WithLoginUsecase(mockLoginUsecase).
		WithAuthService(mockAuthService)

	loginRequest := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	mockLoginUsecase.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestDoLogin_ResponseFormat(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockDoLoginUsecase)
	mockAuthService := new(MockAuthService)

	testUser := createTestUser()

	mockLoginUsecase.On("Execute", "test@example.com", "password123").Return(testUser, nil)
	mockAuthService.On("CreateToken", "test@example.com", "user123").Return("mock-token-123", nil)

	container := di.NewTestContainer().
		WithLoginUsecase(mockLoginUsecase).
		WithAuthService(mockAuthService)

	loginRequest := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	DoLogin(rr, req, container)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify response is valid JSON
	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "token")
	assert.Equal(t, "mock-token-123", response["token"])

	// Verify only token field is present
	assert.Len(t, response, 1)

	mockLoginUsecase.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}
