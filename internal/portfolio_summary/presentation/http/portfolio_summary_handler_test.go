package http

import (
	balDomain "HubInvestments/internal/balance/domain/model"
	"HubInvestments/internal/portfolio_summary/application/usecase"
	"HubInvestments/internal/portfolio_summary/domain/model"
	posDomain "HubInvestments/internal/position/domain/model"
	di "HubInvestments/pck"
	"HubInvestments/shared/middleware"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockPortfolioSummaryUsecase implements the PortfolioSummaryUsecase interface for testing
type MockPortfolioSummaryUsecase struct {
	result model.PortfolioSummaryModel
	err    error
}

func (m *MockPortfolioSummaryUsecase) Execute(userId string) (model.PortfolioSummaryModel, error) {
	if m.err != nil {
		return model.PortfolioSummaryModel{}, m.err
	}
	return m.result, nil
}

// Helper function to create a successful token verifier
func createSuccessfulTokenVerifier(expectedUserId string) middleware.TokenVerifier {
	return func(token string, w http.ResponseWriter) (string, error) {
		return expectedUserId, nil
	}
}

// Helper function to create a failing token verifier
func createFailingTokenVerifier(errorMsg string) middleware.TokenVerifier {
	return func(token string, w http.ResponseWriter) (string, error) {
		return "", errors.New(errorMsg)
	}
}

// Helper function to create test container with mocked portfolio usecase
func createTestContainer(portfolioUsecase usecase.PortfolioSummaryUsecase) di.Container {
	return di.NewTestContainer().WithPortfolioSummaryUsecase(portfolioUsecase)
}

func TestGetPortfolioSummary_Success(t *testing.T) {
	// Arrange
	expectedUserId := "user123"
	expectedResult := model.PortfolioSummaryModel{
		Balance:         balDomain.BalanceModel{AvailableBalance: 5000.0},
		TotalPortfolio:  17000.0,
		LastUpdatedDate: "",
		PositionAggregation: posDomain.AucAggregationModel{
			TotalInvested: 11500.0,
			CurrentTotal:  12000.0,
			PositionAggregation: []posDomain.PositionAggregationModel{
				{
					Category:      1,
					TotalInvested: 6500.0,
					CurrentTotal:  6750.0,
					Pnl:           250.0,
					PnlPercentage: 3.85,
					Assets: []posDomain.AssetsModel{
						{Symbol: "AAPL", Category: 1, AveragePrice: 150.0, LastPrice: 155.0, Quantity: 10.0},
						{Symbol: "GOOGL", Category: 1, AveragePrice: 2500.0, LastPrice: 2600.0, Quantity: 2.0},
					},
				},
			},
		},
	}

	mockUsecase := &MockPortfolioSummaryUsecase{result: expectedResult, err: nil}
	testContainer := createTestContainer(mockUsecase)

	// Create request
	req, err := http.NewRequest("GET", "/getPortfolioSummary", nil)
	assert.NoError(t, err)

	// Act - Test the direct handler (without middleware authentication)
	rr := httptest.NewRecorder()
	GetPortfolioSummary(rr, req, expectedUserId, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.PortfolioSummaryModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response matches expected result
	assert.Equal(t, expectedResult.Balance.AvailableBalance, response.Balance.AvailableBalance)
	assert.Equal(t, expectedResult.TotalPortfolio, response.TotalPortfolio)
	assert.Equal(t, expectedResult.PositionAggregation.TotalInvested, response.PositionAggregation.TotalInvested)
	assert.Equal(t, expectedResult.PositionAggregation.CurrentTotal, response.PositionAggregation.CurrentTotal)
}

func TestGetPortfolioSummaryWithAuth_Success(t *testing.T) {
	// Arrange
	expectedUserId := "user123"
	expectedResult := model.PortfolioSummaryModel{
		Balance:         balDomain.BalanceModel{AvailableBalance: 5000.0},
		TotalPortfolio:  17000.0,
		LastUpdatedDate: "",
		PositionAggregation: posDomain.AucAggregationModel{
			TotalInvested: 11500.0,
			CurrentTotal:  12000.0,
			PositionAggregation: []posDomain.PositionAggregationModel{
				{
					Category:      1,
					TotalInvested: 6500.0,
					CurrentTotal:  6750.0,
					Pnl:           250.0,
					PnlPercentage: 3.85,
					Assets: []posDomain.AssetsModel{
						{Symbol: "AAPL", Category: 1, AveragePrice: 150.0, LastPrice: 155.0, Quantity: 10.0},
						{Symbol: "GOOGL", Category: 1, AveragePrice: 2500.0, LastPrice: 2600.0, Quantity: 2.0},
					},
				},
			},
		},
	}

	mockUsecase := &MockPortfolioSummaryUsecase{result: expectedResult, err: nil}
	testContainer := createTestContainer(mockUsecase)
	verifyToken := createSuccessfulTokenVerifier(expectedUserId)

	// Create request
	req, err := http.NewRequest("GET", "/getPortfolioSummary", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer valid-token")

	// Act - Test the middleware-wrapped handler
	rr := httptest.NewRecorder()
	handler := GetPortfolioSummaryWithAuth(verifyToken, testContainer)
	handler(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response model.PortfolioSummaryModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response matches expected result
	assert.Equal(t, expectedResult.Balance.AvailableBalance, response.Balance.AvailableBalance)
	assert.Equal(t, expectedResult.TotalPortfolio, response.TotalPortfolio)
	assert.Equal(t, expectedResult.PositionAggregation.TotalInvested, response.PositionAggregation.TotalInvested)
	assert.Equal(t, expectedResult.PositionAggregation.CurrentTotal, response.PositionAggregation.CurrentTotal)
}

func TestGetPortfolioSummaryWithAuth_AuthenticationFailure(t *testing.T) {
	// Arrange
	mockUsecase := &MockPortfolioSummaryUsecase{result: model.PortfolioSummaryModel{}, err: nil}
	testContainer := createTestContainer(mockUsecase)
	verifyToken := createFailingTokenVerifier("invalid token")

	// Create request
	req, err := http.NewRequest("GET", "/getPortfolioSummary", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid-token")

	// Act - Test the middleware-wrapped handler
	rr := httptest.NewRecorder()
	handler := GetPortfolioSummaryWithAuth(verifyToken, testContainer)
	handler(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid token")
}

func TestGetPortfolioSummary_UseCaseError(t *testing.T) {
	// Arrange
	expectedUserId := "user123"
	mockUsecase := &MockPortfolioSummaryUsecase{
		result: model.PortfolioSummaryModel{},
		err:    errors.New("database connection failed"),
	}
	testContainer := createTestContainer(mockUsecase)

	// Create request
	req, err := http.NewRequest("GET", "/getPortfolioSummary", nil)
	assert.NoError(t, err)

	// Act - Test the direct handler
	rr := httptest.NewRecorder()
	GetPortfolioSummary(rr, req, expectedUserId, testContainer)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to get portfolio summary")
	assert.Contains(t, rr.Body.String(), "database connection failed")
}

func TestGetPortfolioSummary_EmptyPortfolio(t *testing.T) {
	// Arrange
	expectedUserId := "user123"
	expectedResult := model.PortfolioSummaryModel{
		Balance:         balDomain.BalanceModel{AvailableBalance: 1000.0},
		TotalPortfolio:  1000.0,
		LastUpdatedDate: "",
		PositionAggregation: posDomain.AucAggregationModel{
			TotalInvested:       0.0,
			CurrentTotal:        0.0,
			PositionAggregation: []posDomain.PositionAggregationModel{},
		},
	}

	mockUsecase := &MockPortfolioSummaryUsecase{result: expectedResult, err: nil}
	testContainer := createTestContainer(mockUsecase)

	// Create request
	req, err := http.NewRequest("GET", "/getPortfolioSummary", nil)
	assert.NoError(t, err)

	// Act - Test the direct handler
	rr := httptest.NewRecorder()
	GetPortfolioSummary(rr, req, expectedUserId, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.PortfolioSummaryModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify balance and totals
	assert.Equal(t, expectedResult.Balance.AvailableBalance, response.Balance.AvailableBalance)
	assert.Equal(t, expectedResult.TotalPortfolio, response.TotalPortfolio)
	assert.Equal(t, float32(0), response.PositionAggregation.TotalInvested)
	assert.Equal(t, float32(0), response.PositionAggregation.CurrentTotal)
}

func TestGetPortfolioSummaryWithAuth_MissingAuthorizationHeader(t *testing.T) {
	// Arrange
	mockUsecase := &MockPortfolioSummaryUsecase{result: model.PortfolioSummaryModel{}, err: nil}
	testContainer := createTestContainer(mockUsecase)
	verifyToken := createFailingTokenVerifier("missing authorization header")

	// Create request without Authorization header
	req, err := http.NewRequest("GET", "/getPortfolioSummary", nil)
	assert.NoError(t, err)

	// Act - Test the middleware-wrapped handler
	rr := httptest.NewRecorder()
	handler := GetPortfolioSummaryWithAuth(verifyToken, testContainer)
	handler(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing authorization header")
}

func TestGetPortfolioSummary_JSONResponseStructure(t *testing.T) {
	// Arrange
	expectedUserId := "user123"
	expectedResult := model.PortfolioSummaryModel{
		Balance:         balDomain.BalanceModel{AvailableBalance: 1500.0},
		TotalPortfolio:  3050.0,
		LastUpdatedDate: "",
		PositionAggregation: posDomain.AucAggregationModel{
			TotalInvested: 1500.0,
			CurrentTotal:  1550.0,
			PositionAggregation: []posDomain.PositionAggregationModel{
				{
					Category:      1,
					TotalInvested: 1500.0,
					CurrentTotal:  1550.0,
					Pnl:           50.0,
					PnlPercentage: 3.33,
					Assets: []posDomain.AssetsModel{
						{Symbol: "AAPL", Category: 1, AveragePrice: 150.0, LastPrice: 155.0, Quantity: 10.0},
					},
				},
			},
		},
	}

	mockUsecase := &MockPortfolioSummaryUsecase{result: expectedResult, err: nil}
	testContainer := createTestContainer(mockUsecase)

	// Create request
	req, err := http.NewRequest("GET", "/getPortfolioSummary", nil)
	assert.NoError(t, err)

	// Act - Test the direct handler
	rr := httptest.NewRecorder()
	GetPortfolioSummary(rr, req, expectedUserId, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse JSON response
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)

	// Verify all expected fields are present
	assert.Contains(t, jsonResponse, "Balance")
	assert.Contains(t, jsonResponse, "TotalPortfolio")
	assert.Contains(t, jsonResponse, "LastUpdatedDate")
	assert.Contains(t, jsonResponse, "PositionAggregation")

	// Verify Balance structure
	balance, ok := jsonResponse["Balance"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, balance, "availableBalance")

	// Verify PositionAggregation structure
	positionAgg, ok := jsonResponse["PositionAggregation"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, positionAgg, "totalInvested")
	assert.Contains(t, positionAgg, "currentTotal")
	assert.Contains(t, positionAgg, "positionAggregation")

	// Verify numeric values are correct type
	totalPortfolio, ok := jsonResponse["TotalPortfolio"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(3050.0), totalPortfolio)
}
