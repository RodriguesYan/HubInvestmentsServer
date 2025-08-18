package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// MockMarketDataClient is a mock for IMarketDataClient
type MockMarketDataClient struct {
	mock.Mock
}

func (m *MockMarketDataClient) ValidateSymbol(ctx context.Context, symbol string) (bool, error) {
	args := m.Called(ctx, symbol)
	return args.Bool(0), args.Error(1)
}

func (m *MockMarketDataClient) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	args := m.Called(ctx, symbol)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMarketDataClient) IsMarketOpen(ctx context.Context, symbol string) (bool, error) {
	args := m.Called(ctx, symbol)
	return args.Bool(0), args.Error(1)
}

func (m *MockMarketDataClient) GetAssetDetails(ctx context.Context, symbol string) (*AssetDetails, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AssetDetails), args.Error(1)
}

func (m *MockMarketDataClient) GetTradingHours(ctx context.Context, symbol string) (*TradingHours, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TradingHours), args.Error(1)
}

// MockPositionClient is a mock for IPositionClient
type MockPositionClient struct {
	mock.Mock
}

func (m *MockPositionClient) GetAvailableQuantity(userID, symbol string) (float64, error) {
	args := m.Called(userID, symbol)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockPositionClient) HasSufficientBalance(userID string, requiredAmount float64) (bool, error) {
	args := m.Called(userID, requiredAmount)
	return args.Bool(0), args.Error(1)
}

func TestNewOrderValidationService(t *testing.T) {
	config := OrderValidationConfig{
		MaxOrderValue:         100,
		MaxQuantityPerOrder:   10,
		PriceTolerancePercent: 5,
		MinOrderValue:         1,
	}
	service := NewOrderValidationService(config)
	assert.NotNil(t, service)
	s, ok := service.(*orderValidationService)
	assert.True(t, ok)
	assert.Equal(t, config.MaxOrderValue, s.maxOrderValue)
}

func TestNewOrderValidationServiceWithDefaults(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	assert.NotNil(t, service)
	s, ok := service.(*orderValidationService)
	assert.True(t, ok)
	assert.Equal(t, 1000000.0, s.maxOrderValue)
}

func TestOrderValidationService_ValidateOrder(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	result, err := service.ValidateOrder(context.Background(), order)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
}

func TestOrderValidationService_ValidateOrder_InvalidDomain(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	order := domain.NewOrderFromRepository("id", "", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)

	result, err := service.ValidateOrder(context.Background(), order)
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Errors)
}

func TestOrderValidationService_ValidateOrderWithContext(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	marketDataClient := new(MockMarketDataClient)
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketDataClient.On("ValidateSymbol", mock.Anything, "PETR4").Return(true, nil)
	marketDataClient.On("GetAssetDetails", mock.Anything, "PETR4").Return(&AssetDetails{IsActive: true, IsTradeable: true}, nil)
	marketDataClient.On("IsMarketOpen", mock.Anything, "PETR4").Return(true, nil)
	marketDataClient.On("GetCurrentPrice", mock.Anything, "PETR4").Return(10.0, nil)
	marketDataClient.On("GetTradingHours", mock.Anything, "PETR4").Return(&TradingHours{IsOpen: true}, nil)
	positionClient.On("HasSufficientBalance", "user1", 100.0).Return(true, nil)

	result, err := service.ValidateOrderWithContext(context.Background(), order, marketDataClient, positionClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
}

func TestOrderValidationService_ValidateSymbol(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	marketDataClient := new(MockMarketDataClient)

	marketDataClient.On("ValidateSymbol", mock.Anything, "PETR4").Return(true, nil)
	marketDataClient.On("GetAssetDetails", mock.Anything, "PETR4").Return(&AssetDetails{IsActive: true, IsTradeable: true}, nil)

	result, err := service.ValidateSymbol(context.Background(), "PETR4", marketDataClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
}

func TestOrderValidationService_ValidateSymbol_Invalid(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	marketDataClient := new(MockMarketDataClient)

	marketDataClient.On("ValidateSymbol", mock.Anything, "INVALID").Return(false, nil)

	result, err := service.ValidateSymbol(context.Background(), "INVALID", marketDataClient)
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
}

func TestOrderValidationService_ValidateQuantity(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	result, err := service.ValidateQuantity(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
}

func TestOrderValidationService_ValidateQuantity_TooLarge(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 20000, &price)

	result, err := service.ValidateQuantity(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
}

func TestOrderValidationService_ValidatePrice(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	marketDataClient := new(MockMarketDataClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketDataClient.On("GetCurrentPrice", mock.Anything, "PETR4").Return(10.0, nil)

	result, err := service.ValidatePrice(context.Background(), order, marketDataClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
}

func TestOrderValidationService_ValidateTradingHours(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	marketDataClient := new(MockMarketDataClient)

	marketDataClient.On("IsMarketOpen", mock.Anything, "PETR4").Return(true, nil)
	marketDataClient.On("GetTradingHours", mock.Anything, "PETR4").Return(&TradingHours{IsOpen: true}, nil)

	result, err := service.ValidateTradingHours(context.Background(), "PETR4", marketDataClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
}

func TestOrderValidationService_ValidateOrderSide(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	positionClient.On("HasSufficientBalance", "user1", 100.0).Return(true, nil)

	result, err := service.ValidateOrderSide(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
}

func TestOrderValidationService_ValidateRiskLimits(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	result, err := service.ValidateRiskLimits(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
}

func TestOrderValidationService_ValidateOrderWithContext_SymbolError(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	marketDataClient := new(MockMarketDataClient)
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketDataClient.On("ValidateSymbol", mock.Anything, "PETR4").Return(false, errors.New("some error"))
	marketDataClient.On("IsMarketOpen", mock.Anything, "PETR4").Return(true, nil)
	marketDataClient.On("GetCurrentPrice", mock.Anything, "PETR4").Return(10.0, nil)
	marketDataClient.On("GetTradingHours", mock.Anything, "PETR4").Return(&TradingHours{IsOpen: true}, nil)
	positionClient.On("HasSufficientBalance", "user1", 100.0).Return(true, nil)

	result, err := service.ValidateOrderWithContext(context.Background(), order, marketDataClient, positionClient)
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
}

func TestOrderValidationService_ValidateQuantity_SellOrder_Insufficient(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideSell, domain.OrderTypeLimit, 10, &price)

	positionClient.On("GetAvailableQuantity", "user1", "PETR4").Return(5.0, nil)

	result, err := service.ValidateQuantity(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
	assert.Contains(t, result.Errors, "insufficient position: cannot sell more than available quantity")
}

func TestOrderValidationService_ValidateOrderSide_Sell(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideSell, domain.OrderTypeMarket, 10, nil)

	result, err := service.ValidateOrderSide(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Contains(t, result.Warnings, "Sell order - ensure you want to reduce your position")
}

func TestOrderValidationService_ValidateRiskLimits_TooHigh(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	price := 2000000.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)

	result, err := service.ValidateRiskLimits(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
}

func TestOrderValidationService_ValidateRiskLimits_TooLow(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	price := 0.5
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)

	result, err := service.ValidateRiskLimits(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
}

func TestOrderValidationService_ValidateOrderTypeRules_MarketWithPrice(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	price := 10.0
	order := domain.NewOrderFromRepository("id", "user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, &price, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
	result := &ValidationResult{IsValid: true, Errors: make([]string, 0), Warnings: make([]string, 0)}
	service.(*orderValidationService).validateOrderTypeRules(order, result)
	assert.True(t, result.IsValid)
	assert.Contains(t, result.Warnings, "Market orders should not have a price specified")
}

func TestOrderValidationService_ValidateOrderWithContext_TradingHoursWarning(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	marketDataClient := new(MockMarketDataClient)
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketDataClient.On("ValidateSymbol", mock.Anything, "PETR4").Return(true, nil)
	marketDataClient.On("GetAssetDetails", mock.Anything, "PETR4").Return(&AssetDetails{IsActive: true, IsTradeable: true}, nil)
	marketDataClient.On("IsMarketOpen", mock.Anything, "PETR4").Return(false, nil)
	marketDataClient.On("GetCurrentPrice", mock.Anything, "PETR4").Return(10.0, nil)
	marketDataClient.On("GetTradingHours", mock.Anything, "PETR4").Return(&TradingHours{IsOpen: false}, nil)
	positionClient.On("HasSufficientBalance", "user1", 100.0).Return(true, nil)

	result, err := service.ValidateOrderWithContext(context.Background(), order, marketDataClient, positionClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Contains(t, result.Warnings, "Market is currently closed for symbol 'PETR4'")
}

func TestOrderValidationService_ValidateOrderWithContext_PriceWarning(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	marketDataClient := new(MockMarketDataClient)
	positionClient := new(MockPositionClient)
	price := 12.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketDataClient.On("ValidateSymbol", mock.Anything, "PETR4").Return(true, nil)
	marketDataClient.On("GetAssetDetails", mock.Anything, "PETR4").Return(&AssetDetails{IsActive: true, IsTradeable: true}, nil)
	marketDataClient.On("IsMarketOpen", mock.Anything, "PETR4").Return(true, nil)
	marketDataClient.On("GetCurrentPrice", mock.Anything, "PETR4").Return(10.0, nil)
	marketDataClient.On("GetTradingHours", mock.Anything, "PETR4").Return(&TradingHours{IsOpen: true}, nil)
	positionClient.On("HasSufficientBalance", "user1", 120.0).Return(true, nil)

	result, err := service.ValidateOrderWithContext(context.Background(), order, marketDataClient, positionClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Contains(t, result.Warnings, "Order price 12.00 differs from market price 10.00 by 20.0% (tolerance: 10.0%)")
}

func TestOrderValidationService_ValidateSymbol_AssetDetailsError(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	marketDataClient := new(MockMarketDataClient)

	marketDataClient.On("ValidateSymbol", mock.Anything, "PETR4").Return(true, nil)
	marketDataClient.On("GetAssetDetails", mock.Anything, "PETR4").Return(nil, errors.New("some error"))

	result, err := service.ValidateSymbol(context.Background(), "PETR4", marketDataClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Contains(t, result.Warnings, "Could not retrieve asset details: some error")
}

func TestOrderValidationService_ValidateQuantity_SellLargePercentage(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideSell, domain.OrderTypeLimit, 9, &price)

	positionClient.On("GetAvailableQuantity", "user1", "PETR4").Return(10.0, nil)

	result, err := service.ValidateQuantity(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Contains(t, result.Warnings, "Selling more than 80% of available position")
}

func TestOrderValidationService_ValidateOrderSide_InsufficientBalance(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	positionClient := new(MockPositionClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	positionClient.On("HasSufficientBalance", "user1", 100.0).Return(false, nil)

	result, err := service.ValidateOrderSide(context.Background(), order, positionClient)
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
}

func TestOrderValidationService_validateQuantityLimits(t *testing.T) {
	service := NewOrderValidationServiceWithDefaults()
	order := domain.NewOrderFromRepository("id", "user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 0, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
	result := &ValidationResult{IsValid: true, Errors: make([]string, 0), Warnings: make([]string, 0)}
	service.(*orderValidationService).validateQuantityLimits(order, result)
	assert.False(t, result.IsValid)
}
