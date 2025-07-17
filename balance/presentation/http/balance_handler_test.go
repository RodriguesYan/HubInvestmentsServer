package http

import (
	"HubInvestments/balance/application/usecase"
	domain "HubInvestments/balance/domain/model"
	di "HubInvestments/pck"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockBalanceRepository struct {
	result domain.BalanceModel
	err    error
}

func (m *MockBalanceRepository) GetBalance(userId string) (domain.BalanceModel, error) {
	return m.result, m.err
}

// Helper function to create a successful token verifier
func createSuccessfulTokenVerifier(expectedUserId string) TokenVerifier {
	return func(token string, w http.ResponseWriter) (string, error) {
		return expectedUserId, nil
	}
}

// Helper function to create a failing token verifier
func createFailingTokenVerifier(errorMsg string) TokenVerifier {
	return func(token string, w http.ResponseWriter) (string, error) {
		return "", errors.New(errorMsg)
	}
}

// Helper function to create test container with mocked portfolio usecase
func createTestContainer(balanceUsecase *usecase.GetBalanceUseCase) di.Container {
	return di.NewTestContainer().WithBalanceUseCase(balanceUsecase)
}

func TestGetBalance_Success(t *testing.T) {
	req, err := http.NewRequest("GET", "/getBalance", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer valid-token")

	expectedUserId := "user123"
	rr := httptest.NewRecorder()

	expectedBalance := domain.BalanceModel{
		AvailableBalance: 10000,
	}

	mockRepo := &MockBalanceRepository{result: expectedBalance, err: nil}
	balanceUsecase := usecase.NewGetBalanceUseCase(mockRepo)
	testContainer := createTestContainer(balanceUsecase)
	verifyToken := createSuccessfulTokenVerifier(expectedUserId)

	GetBalance(rr, req, verifyToken, testContainer)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var actualBalance domain.BalanceModel
	err = json.Unmarshal(rr.Body.Bytes(), &actualBalance)
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, actualBalance)
}

func TestGetBalance_AuthenticationFailure(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/getBalance", nil)
	verifyToken := createFailingTokenVerifier("invalid token")
	mockRepo := &MockBalanceRepository{result: domain.BalanceModel{}, err: nil}
	balanceUsecase := usecase.NewGetBalanceUseCase(mockRepo)
	testContainer := createTestContainer(balanceUsecase)

	GetBalance(rr, req, verifyToken, testContainer)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid token")
}

func TestGetBalance_UseCaseError(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/getBalance", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()
	useCaseError := errors.New("database connection failed")

	mockRepo := &MockBalanceRepository{err: useCaseError}
	balanceUsecase := usecase.NewGetBalanceUseCase(mockRepo)
	testContainer := createTestContainer(balanceUsecase)
	verifyToken := createSuccessfulTokenVerifier("user123")

	// Act
	GetBalance(rr, req, verifyToken, testContainer)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to get balance: "+useCaseError.Error())
}

func TestGetBalance_JSONMarshalError(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/getBalance", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()

	// Create a balance model with NaN, which causes json.Marshal to fail
	unmarshallableBalance := domain.BalanceModel{
		AvailableBalance: float32(math.NaN()),
	}

	mockRepo := &MockBalanceRepository{result: unmarshallableBalance}
	balanceUsecase := usecase.NewGetBalanceUseCase(mockRepo)
	testContainer := createTestContainer(balanceUsecase)
	verifyToken := createSuccessfulTokenVerifier("user123")

	// Act
	GetBalance(rr, req, verifyToken, testContainer)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "json: unsupported value")
}
