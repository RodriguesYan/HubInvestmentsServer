package usecase

import (
	"HubInvestments/internal/market_data/domain/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMarketDataRepository implements the IMarketDataRepository interface for testing
type MockMarketDataRepository struct {
	mock.Mock
}

func (m *MockMarketDataRepository) GetMarketData(symbols []string) ([]model.MarketDataModel, error) {
	args := m.Called(symbols)
	return args.Get(0).([]model.MarketDataModel), args.Error(1)
}

func TestNewGetMarketDataUseCase(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}

	// Act
	usecase := NewGetMarketDataUseCase(mockRepo)

	// Assert
	assert.NotNil(t, usecase)
	assert.IsType(t, &GetMarketDataUsecase{}, usecase)
}

func TestGetMarketDataUsecase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}
	symbols := []string{"AAPL", "GOOGL", "MSFT"}

	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},
		{Symbol: "GOOGL", Name: "Alphabet Inc.", LastQuote: 2650.75, Category: 1},
		{Symbol: "MSFT", Name: "Microsoft Corporation", LastQuote: 285.25, Category: 1},
	}

	// Mock the repository call
	mockRepo.On("GetMarketData", symbols).Return(expectedData, nil)

	usecase := NewGetMarketDataUseCase(mockRepo)

	// Act
	result, err := usecase.Execute(symbols)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(expectedData), len(result))
	assert.Equal(t, expectedData, result)

	// Verify that the repository method was called with correct parameters
	mockRepo.AssertExpectations(t)
	mockRepo.AssertCalled(t, "GetMarketData", symbols)
}

func TestGetMarketDataUsecase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}
	symbols := []string{"AAPL", "GOOGL"}
	repositoryError := errors.New("database connection failed")

	// Mock the repository to return an error
	mockRepo.On("GetMarketData", symbols).Return([]model.MarketDataModel(nil), repositoryError)

	usecase := NewGetMarketDataUseCase(mockRepo)

	// Act
	result, err := usecase.Execute(symbols)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repositoryError, err)

	// Verify that the repository method was called
	mockRepo.AssertExpectations(t)
}

func TestGetMarketDataUsecase_Execute_EmptySymbols(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}
	symbols := []string{}
	expectedData := []model.MarketDataModel{}

	// Mock the repository to return empty data for empty symbols
	mockRepo.On("GetMarketData", symbols).Return(expectedData, nil)

	usecase := NewGetMarketDataUseCase(mockRepo)

	// Act
	result, err := usecase.Execute(symbols)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, expectedData, result)

	// Verify that the repository method was called
	mockRepo.AssertExpectations(t)
}

func TestGetMarketDataUsecase_Execute_SingleSymbol(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}
	symbols := []string{"AAPL"}

	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},
	}

	// Mock the repository call
	mockRepo.On("GetMarketData", symbols).Return(expectedData, nil)

	usecase := NewGetMarketDataUseCase(mockRepo)

	// Act
	result, err := usecase.Execute(symbols)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "AAPL", result[0].Symbol)
	assert.Equal(t, "Apple Inc.", result[0].Name)
	assert.Equal(t, float32(155.50), result[0].LastQuote)
	assert.Equal(t, 1, result[0].Category)

	// Verify that the repository method was called
	mockRepo.AssertExpectations(t)
}

func TestGetMarketDataUsecase_Execute_PartialDataReturned(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}
	symbols := []string{"AAPL", "GOOGL", "INVALID_SYMBOL"}

	// Repository returns data for only valid symbols
	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},
		{Symbol: "GOOGL", Name: "Alphabet Inc.", LastQuote: 2650.75, Category: 1},
	}

	// Mock the repository call
	mockRepo.On("GetMarketData", symbols).Return(expectedData, nil)

	usecase := NewGetMarketDataUseCase(mockRepo)

	// Act
	result, err := usecase.Execute(symbols)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result)) // Only 2 valid symbols returned
	assert.Equal(t, expectedData, result)

	// Verify that the repository method was called
	mockRepo.AssertExpectations(t)
}

func TestGetMarketDataUsecase_Execute_DifferentCategories(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}
	symbols := []string{"AAPL", "VOO", "BTC"}

	expectedData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 155.50, Category: 1},          // Stock
		{Symbol: "VOO", Name: "Vanguard S&P 500 ETF", LastQuote: 385.25, Category: 2}, // ETF
		{Symbol: "BTC", Name: "Bitcoin", LastQuote: 45000.00, Category: 3},            // Crypto
	}

	// Mock the repository call
	mockRepo.On("GetMarketData", symbols).Return(expectedData, nil)

	usecase := NewGetMarketDataUseCase(mockRepo)

	// Act
	result, err := usecase.Execute(symbols)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result))

	// Check individual categories
	assert.Equal(t, 1, result[0].Category) // Stock
	assert.Equal(t, 2, result[1].Category) // ETF
	assert.Equal(t, 3, result[2].Category) // Crypto

	// Verify that the repository method was called
	mockRepo.AssertExpectations(t)
}

func TestGetMarketDataUsecase_Execute_NilSymbols(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}
	var symbols []string = nil
	expectedData := []model.MarketDataModel{}

	// Mock the repository to handle nil symbols
	mockRepo.On("GetMarketData", symbols).Return(expectedData, nil)

	usecase := NewGetMarketDataUseCase(mockRepo)

	// Act
	result, err := usecase.Execute(symbols)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))

	// Verify that the repository method was called
	mockRepo.AssertExpectations(t)
}

func TestGetMarketDataUsecase_Execute_RepositoryReturnsNil(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}
	symbols := []string{"AAPL"}

	// Mock the repository to return nil without error
	mockRepo.On("GetMarketData", symbols).Return([]model.MarketDataModel(nil), nil)

	usecase := NewGetMarketDataUseCase(mockRepo)

	// Act
	result, err := usecase.Execute(symbols)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Verify that the repository method was called
	mockRepo.AssertExpectations(t)
}

func TestGetMarketDataUsecase_Execute_LargeSymbolList(t *testing.T) {
	// Arrange
	mockRepo := &MockMarketDataRepository{}

	// Create a large list of symbols
	symbols := make([]string, 100)
	expectedData := make([]model.MarketDataModel, 100)

	for i := 0; i < 100; i++ {
		symbol := "SYMBOL" + string(rune(i))
		symbols[i] = symbol
		expectedData[i] = model.MarketDataModel{
			Symbol:    symbol,
			Name:      "Test Company " + string(rune(i)),
			LastQuote: float32(100.0 + float64(i)),
			Category:  1,
		}
	}

	// Mock the repository call
	mockRepo.On("GetMarketData", symbols).Return(expectedData, nil)

	usecase := NewGetMarketDataUseCase(mockRepo)

	// Act
	result, err := usecase.Execute(symbols)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100, len(result))
	assert.Equal(t, expectedData, result)

	// Verify that the repository method was called
	mockRepo.AssertExpectations(t)
}
