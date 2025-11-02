package http

import (
	"HubInvestments/internal/market_data/application/usecase"
	"HubInvestments/internal/market_data/domain/model"
	di "HubInvestments/pck"
	"HubInvestments/shared/middleware"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMarketDataUsecase implements the IGetMarketDataUsecase interface for testing
type MockMarketDataUsecase struct {
	mock.Mock
}

func (m *MockMarketDataUsecase) Execute(symbols []string) ([]model.MarketDataModel, error) {
	args := m.Called(symbols)
	return args.Get(0).([]model.MarketDataModel), args.Error(1)
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

// Helper function to create test container with mocked market data usecase
func createTestContainer(marketDataUsecase usecase.IGetMarketDataUsecase) di.Container {
	return di.NewTestContainer().WithMarketDataUsecase(marketDataUsecase)
}

func TestGetMarketData_Success(t *testing.T) {
	// Arrange
	symbols := []string{"AAPL", "GOOGL", "MSFT"}
	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},
		{Symbol: "GOOGL", Name: "Alphabet Inc.", LastQuote: 2650.75, Category: 1},
		{Symbol: "MSFT", Name: "Microsoft Corporation", LastQuote: 285.25, Category: 1},
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request with query parameters
	req, err := http.NewRequest("GET", "/getMarketData?symbols=AAPL,GOOGL,MSFT", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act - Test the direct handler (without middleware authentication)
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, len(expectedData), len(response))
	assert.Equal(t, expectedData, response)

	// Verify that the usecase was called with correct parameters
	mockUsecase.AssertExpectations(t)
	mockUsecase.AssertCalled(t, "Execute", symbols)
}

func TestGetMarketData_SingleSymbol(t *testing.T) {
	// Arrange
	symbols := []string{"AAPL"}
	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request with single symbol
	req, err := http.NewRequest("GET", "/getMarketData?symbols=AAPL", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(response))
	assert.Equal(t, "AAPL", response[0].Symbol)
	assert.Equal(t, "Apple Inc.", response[0].Name)
	assert.Equal(t, float32(155.50), response[0].LastQuote)
	assert.Equal(t, 1, response[0].Category)

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketData_EmptySymbols(t *testing.T) {
	// Arrange
	symbols := []string{""}
	expectedData := []model.MarketDataModel{}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request with empty symbols parameter
	req, err := http.NewRequest("GET", "/getMarketData?symbols=", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(response))

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketData_NoSymbolsParameter(t *testing.T) {
	// Arrange - No symbols parameter at all
	symbols := []string{""}
	expectedData := []model.MarketDataModel{}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request without symbols parameter
	req, err := http.NewRequest("GET", "/getMarketData", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(response))

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketData_UsecaseError(t *testing.T) {
	// Arrange
	symbols := []string{"AAPL", "GOOGL"}
	usecaseError := errors.New("database connection failed")

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return([]model.MarketDataModel(nil), usecaseError)

	testContainer := createTestContainer(mockUsecase)

	// Create request
	req, err := http.NewRequest("GET", "/getMarketData?symbols=AAPL,GOOGL", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to get market data: "+usecaseError.Error())

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketData_JSONMarshalError(t *testing.T) {
	// Arrange - Create data that will cause JSON marshal to fail
	symbols := []string{"AAPL"}
	unmarshallableData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: float32(math.NaN()), Category: 1}, // NaN causes JSON marshal to fail
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(unmarshallableData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request
	req, err := http.NewRequest("GET", "/getMarketData?symbols=AAPL", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "json: unsupported value")

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketData_MultipleSymbolsWithSpaces(t *testing.T) {
	// Arrange - Test URL encoding and spaces
	symbols := []string{"AAPL", "GOOGL", "TSLA"}
	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},
		{Symbol: "GOOGL", Name: "Alphabet Inc.", LastQuote: 2650.75, Category: 1},
		{Symbol: "TSLA", Name: "Tesla Inc.", LastQuote: 800.25, Category: 1},
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request with URL encoded comma-separated symbols
	req, err := http.NewRequest("GET", "/getMarketData", nil)
	assert.NoError(t, err)

	// Manually set query parameters
	q := req.URL.Query()
	q.Add("symbols", "AAPL,GOOGL,TSLA")
	req.URL.RawQuery = q.Encode()

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(response))
	assert.Equal(t, expectedData, response)

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketData_DifferentCategories(t *testing.T) {
	// Arrange
	symbols := []string{"AAPL", "VOO", "BTC"}
	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},          // Stock
		{Symbol: "VOO", Name: "Vanguard S&P 500 ETF", LastQuote: 385.25, Category: 2}, // ETF
		{Symbol: "BTC", Name: "Bitcoin", LastQuote: 45000.00, Category: 3},            // Crypto
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request
	req, err := http.NewRequest("GET", "/getMarketData?symbols=AAPL,VOO,BTC", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(response))

	// Check different categories
	assert.Equal(t, 1, response[0].Category) // Stock
	assert.Equal(t, 2, response[1].Category) // ETF
	assert.Equal(t, 3, response[2].Category) // Crypto

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketDataWithAuth_Success(t *testing.T) {
	// Arrange
	expectedUserId := "user123"
	symbols := []string{"AAPL", "GOOGL"}
	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},
		{Symbol: "GOOGL", Name: "Alphabet Inc.", LastQuote: 2650.75, Category: 1},
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)
	verifyToken := createSuccessfulTokenVerifier(expectedUserId)

	// Create request
	req, err := http.NewRequest("GET", "/getMarketData?symbols=AAPL,GOOGL", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()

	// Act - Test the middleware-wrapped handler
	handler := GetMarketDataWithAuth(verifyToken, testContainer)
	handler(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, len(expectedData), len(response))
	assert.Equal(t, expectedData, response)

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketDataWithAuth_AuthenticationFailure(t *testing.T) {
	// Arrange
	mockUsecase := &MockMarketDataUsecase{}
	testContainer := createTestContainer(mockUsecase)
	verifyToken := createFailingTokenVerifier("invalid token")

	// Create request
	req, err := http.NewRequest("GET", "/getMarketData?symbols=AAPL", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rr := httptest.NewRecorder()

	// Act - Test the middleware-wrapped handler
	handler := GetMarketDataWithAuth(verifyToken, testContainer)
	handler(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid token")

	// Usecase should not be called when authentication fails
	mockUsecase.AssertNotCalled(t, "Execute")
}

func TestGetMarketData_SpecialCharactersInSymbols(t *testing.T) {
	// Arrange - Test symbols with special characters
	symbols := []string{"BRK.B", "SPY", "QQQ"}
	expectedData := []model.MarketDataModel{
		{Symbol: "BRK.B", Name: "Berkshire Hathaway Inc.", LastQuote: 275.50, Category: 1},
		{Symbol: "SPY", Name: "SPDR S&P 500 ETF Trust", LastQuote: 420.75, Category: 2},
		{Symbol: "QQQ", Name: "Invesco QQQ Trust", LastQuote: 350.25, Category: 2},
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request with URL encoded symbols
	symbolsParam := url.QueryEscape("BRK.B,SPY,QQQ")
	req, err := http.NewRequest("GET", "/getMarketData?symbols="+symbolsParam, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(response))
	assert.Equal(t, "BRK.B", response[0].Symbol)

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketData_LargeSymbolList(t *testing.T) {
	// Arrange - Test with many symbols
	symbols := make([]string, 50)
	expectedData := make([]model.MarketDataModel, 50)
	symbolsStr := make([]string, 50)

	for i := 0; i < 50; i++ {
		symbol := "SYM" + strings.Repeat("0", 2-len(strconv.Itoa(i))) + strconv.Itoa(i)
		symbols[i] = symbol
		symbolsStr[i] = symbol
		expectedData[i] = model.MarketDataModel{
			Symbol:    symbol,
			Name:      "Company " + symbol,
			LastQuote: float32(100.0 + float64(i)),
			Category:  1,
		}
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request with many symbols
	symbolsParam := strings.Join(symbolsStr, ",")
	req, err := http.NewRequest("GET", "/getMarketData?symbols="+symbolsParam, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 50, len(response))
	assert.Equal(t, expectedData, response)

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketData_HTTPMethods(t *testing.T) {
	// Arrange
	symbols := []string{"AAPL"}
	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil).Times(3) // Will be called for each method

	testContainer := createTestContainer(mockUsecase)

	// Test different HTTP methods (handler doesn't restrict them)
	methods := []string{"GET", "POST", "PUT"}

	for _, method := range methods {
		t.Run("method_"+method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/getMarketData?symbols=AAPL", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			// Act
			GetMarketData(rr, req, testContainer)

			// Assert
			assert.Equal(t, http.StatusOK, rr.Code)

			var response []model.MarketDataModel
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(response))
		})
	}

	mockUsecase.AssertExpectations(t)
}

func TestGetMarketData_ResponseContentType(t *testing.T) {
	// Arrange
	symbols := []string{"AAPL"}
	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},
	}

	mockUsecase := &MockMarketDataUsecase{}
	mockUsecase.On("Execute", symbols).Return(expectedData, nil)

	testContainer := createTestContainer(mockUsecase)

	// Create request
	req, err := http.NewRequest("GET", "/getMarketData?symbols=AAPL", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	GetMarketData(rr, req, testContainer)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check that response is valid JSON format
	var response []model.MarketDataModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check response structure
	assert.Equal(t, 1, len(response))
	assert.Contains(t, rr.Body.String(), "\"Symbol\"")
	assert.Contains(t, rr.Body.String(), "\"Name\"")
	assert.Contains(t, rr.Body.String(), "\"LastQuote\"")
	assert.Contains(t, rr.Body.String(), "\"Category\"")

	mockUsecase.AssertExpectations(t)
}
