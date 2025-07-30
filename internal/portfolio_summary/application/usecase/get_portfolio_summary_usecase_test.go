package usecase

import (
	balUsecase "HubInvestments/internal/balance/application/usecase"
	balDomain "HubInvestments/internal/balance/domain/model"
	posUsecase "HubInvestments/internal/position/application/usecase"
	posModel "HubInvestments/internal/position/domain/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock repository for positions
type MockPositionRepository struct {
	positions []posModel.AssetModel
	err       error
}

func (m *MockPositionRepository) GetPositionsByUserId(userId string) ([]posModel.AssetModel, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.positions, nil
}

// Mock repository for balance
type MockBalanceRepository struct {
	balance balDomain.BalanceModel
	err     error
}

func (m *MockBalanceRepository) GetBalance(userId string) (balDomain.BalanceModel, error) {
	if m.err != nil {
		return balDomain.BalanceModel{}, m.err
	}
	return m.balance, nil
}

func TestGetPortfolioSummary_Success(t *testing.T) {
	// Arrange - Create mock data
	mockBalance := balDomain.BalanceModel{AvailableBalance: 5000.0}
	mockPositions := []posModel.AssetModel{
		{Symbol: "AAPL", Category: 1, AveragePrice: 150.0, LastPrice: 155.0, Quantity: 10.0},
		{Symbol: "GOOGL", Category: 1, AveragePrice: 2500.0, LastPrice: 2600.0, Quantity: 2.0},
	}

	// Create mock repositories
	mockBalanceRepo := &MockBalanceRepository{balance: mockBalance, err: nil}
	mockPositionRepo := &MockPositionRepository{positions: mockPositions, err: nil}

	// Create real use cases with mocked repositories
	balanceUsecase := balUsecase.NewGetBalanceUseCase(mockBalanceRepo)
	positionUsecase := posUsecase.NewGetPositionAggregationUseCase(mockPositionRepo)

	// Create the actual GetPortfolioSummaryUsecase we want to test
	portfolioUsecase := NewGetPortfolioSummaryUsecase(*positionUsecase, *balanceUsecase)

	// Act - Execute the method we're testing
	result, err := portfolioUsecase.Execute("testUser")

	// Assert - Verify the results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check balance is preserved
	if result.Balance.AvailableBalance != mockBalance.AvailableBalance {
		t.Errorf("Expected balance %f, got %f", mockBalance.AvailableBalance, result.Balance.AvailableBalance)
	}

	// Check position aggregation is calculated correctly
	expectedPositionTotal := float32(155.0*10.0 + 2600.0*2.0) // AAPL: 1550 + GOOGL: 5200 = 6750
	if result.PositionAggregation.CurrentTotal != expectedPositionTotal {
		t.Errorf("Expected position total %f, got %f", expectedPositionTotal, result.PositionAggregation.CurrentTotal)
	}

	// Check total portfolio calculation
	expectedTotalPortfolio := mockBalance.AvailableBalance + expectedPositionTotal
	if result.TotalPortfolio != expectedTotalPortfolio {
		t.Errorf("Expected total portfolio %f, got %f", expectedTotalPortfolio, result.TotalPortfolio)
	}

	// Verify position aggregation has the expected structure
	if len(result.PositionAggregation.PositionAggregation) != 1 {
		t.Errorf("Expected 1 position aggregation category, got %d", len(result.PositionAggregation.PositionAggregation))
	}

	// Check the aggregated category (both AAPL and GOOGL are category 1)
	categoryAgg := result.PositionAggregation.PositionAggregation[0]
	if categoryAgg.Category != 1 {
		t.Errorf("Expected category 1, got %d", categoryAgg.Category)
	}

	if len(categoryAgg.Assets) != 2 {
		t.Errorf("Expected 2 assets in category 1, got %d", len(categoryAgg.Assets))
	}
}

func TestGetPortfolioSummary_BalanceFailure(t *testing.T) {
	// Arrange - Create mock data
	mockBalance := balDomain.BalanceModel{}
	mockPositions := []posModel.AssetModel{
		{Symbol: "AAPL", Category: 1, AveragePrice: 150.0, LastPrice: 155.0, Quantity: 10.0},
		{Symbol: "GOOGL", Category: 1, AveragePrice: 2500.0, LastPrice: 2600.0, Quantity: 2.0},
	}

	// Create mock repositories
	mockBalanceRepo := &MockBalanceRepository{balance: mockBalance, err: errors.New("Failed to get balance")}
	mockPositionRepo := &MockPositionRepository{positions: mockPositions, err: nil}

	// Create real use cases with mocked repositories
	balanceUsecase := balUsecase.NewGetBalanceUseCase(mockBalanceRepo)
	positionUsecase := posUsecase.NewGetPositionAggregationUseCase(mockPositionRepo)

	// Create the actual GetPortfolioSummaryUsecase we want to test
	portfolioUsecase := NewGetPortfolioSummaryUsecase(*positionUsecase, *balanceUsecase)
	_, err := portfolioUsecase.Execute("testUser")

	assert.Error(t, err)
	assert.Equal(t, "Failed to get balance", err.Error())
}

func TestGetPortfolioSummary_PositionFailure(t *testing.T) {
	// Arrange - Create mock data
	mockBalance := balDomain.BalanceModel{}
	mockPositions := []posModel.AssetModel{}

	// Create mock repositories
	mockBalanceRepo := &MockBalanceRepository{balance: mockBalance, err: nil}
	mockPositionRepo := &MockPositionRepository{positions: mockPositions, err: errors.New("Failed to get position")}

	// Create real use cases with mocked repositories
	balanceUsecase := balUsecase.NewGetBalanceUseCase(mockBalanceRepo)
	positionUsecase := posUsecase.NewGetPositionAggregationUseCase(mockPositionRepo)

	// Create the actual GetPortfolioSummaryUsecase we want to test
	portfolioUsecase := NewGetPortfolioSummaryUsecase(*positionUsecase, *balanceUsecase)
	_, err := portfolioUsecase.Execute("testUser")

	assert.Error(t, err)
	assert.Equal(t, "Failed to get position", err.Error())
}
