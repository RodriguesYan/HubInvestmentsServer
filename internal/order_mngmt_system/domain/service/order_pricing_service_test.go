package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// MockPricingDataClient is a mock implementation of IPricingDataClient for testing
type MockPricingDataClient struct {
	mock.Mock
}

func (m *MockPricingDataClient) GetCurrentMarketPrice(symbol string) (*MarketPrice, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MarketPrice), args.Error(1)
}

func (m *MockPricingDataClient) GetOrderBookData(symbol string) (*OrderBookData, error) {
	args := m.Called(symbol)
	return args.Get(0).(*OrderBookData), args.Error(1)
}

func (m *MockPricingDataClient) GetHistoricalPrices(symbol string, period time.Duration) ([]HistoricalPrice, error) {
	args := m.Called(symbol, period)
	return args.Get(0).([]HistoricalPrice), args.Error(1)
}

func (m *MockPricingDataClient) GetMarketDepth(symbol string) (*MarketDepth, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MarketDepth), args.Error(1)
}

func (m *MockPricingDataClient) IsMarketOpen(symbol string) (bool, error) {
	args := m.Called(symbol)
	return args.Bool(0), args.Error(1)
}

func (m *MockPricingDataClient) GetTradingFees(orderType domain.OrderType, orderValue float64) (*TradingFees, error) {
	args := m.Called(orderType, orderValue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TradingFees), args.Error(1)
}

func (m *MockPricingDataClient) GetPriceImpactEstimate(symbol string, orderSide domain.OrderSide, quantity float64) (*PriceImpact, error) {
	args := m.Called(symbol, orderSide, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PriceImpact), args.Error(1)
}

func TestNewOrderPricingService(t *testing.T) {
	config := OrderPricingConfig{
		MaxSlippagePercent:    5.0,
		MinLiquidityThreshold: 20000.0,
		SpreadWarningPercent:  2.0,
		ImpactWarningPercent:  1.0,
		FeeCalculationMethod:  FeeCalculationFixed,
	}

	service := NewOrderPricingService(config)
	assert.NotNil(t, service)

	s, ok := service.(*orderPricingService)
	assert.True(t, ok)
	assert.Equal(t, config.MaxSlippagePercent, s.maxSlippagePercent)
	assert.Equal(t, config.MinLiquidityThreshold, s.minLiquidityThreshold)
	assert.Equal(t, config.SpreadWarningPercent, s.spreadWarningPercent)
	assert.Equal(t, config.ImpactWarningPercent, s.impactWarningPercent)
	assert.Equal(t, config.FeeCalculationMethod, s.feeCalculationMethod)
}

func TestNewOrderPricingServiceWithDefaults(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	assert.NotNil(t, service)

	s, ok := service.(*orderPricingService)
	assert.True(t, ok)
	assert.Equal(t, 2.0, s.maxSlippagePercent)
	assert.Equal(t, 10000.0, s.minLiquidityThreshold)
	assert.Equal(t, 1.0, s.spreadWarningPercent)
	assert.Equal(t, 0.5, s.impactWarningPercent)
	assert.Equal(t, FeeCalculationTiered, s.feeCalculationMethod)
}

func TestOrderPricingService_CalculateOptimalPrice(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	marketPrice := &MarketPrice{Symbol: "PETR4", BidPrice: 100, AskPrice: 101, LastPrice: 100.5, Spread: 1, SpreadPercent: 1}
	marketDepth := &MarketDepth{LiquidityScore: 0.7}

	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)
	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(marketDepth, nil)

	result, err := service.CalculateOptimalPrice(order, mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "PETR4", result.Symbol)
	assert.True(t, result.RecommendedPrice > 0)
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_CalculateOptimalPrice_MarketPriceError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(nil, fmt.Errorf("network error"))

	_, err := service.CalculateOptimalPrice(order, mockClient)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get market price")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_CreateExecutionPlan(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	marketPrice := &MarketPrice{Symbol: "PETR4", BidPrice: 100, AskPrice: 101, LastPrice: 100.5, Spread: 1, SpreadPercent: 1}
	marketDepth := &MarketDepth{LiquidityScore: 0.7}
	tradingFees := &TradingFees{TotalFees: 5.0}
	priceImpact := &PriceImpact{EstimatedImpact: 0.1}

	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(marketDepth, nil)
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)
	mockClient.On("GetTradingFees", order.OrderType(), order.CalculateOrderValue()).Return(tradingFees, nil)
	mockClient.On("GetPriceImpactEstimate", order.Symbol(), order.OrderSide(), order.Quantity()).Return(priceImpact, nil)

	plan, err := service.CreateExecutionPlan(order, mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, order.ID(), plan.OrderID)
	assert.NotEmpty(t, plan.ExecutionInstructions)
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_ValidateOrderPrice(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	price := 101.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketPrice := &MarketPrice{Symbol: "PETR4", BidPrice: 100, AskPrice: 100.5, LastPrice: 101, Spread: 0.5, SpreadPercent: 0.5}
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)

	err := service.ValidateOrderPrice(order, mockClient)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_ValidateOrderPrice_MarketOrder(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	err := service.ValidateOrderPrice(order, mockClient)
	assert.NoError(t, err)
	mockClient.AssertNotCalled(t, "GetCurrentMarketPrice")
}

func TestOrderPricingService_ValidateOrderPrice_NoPrice(t *testing.T) {
	_, err := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, nil)
	assert.Error(t, err)
	assert.Equal(t, "limit orders must have a price", err.Error())
}

func TestOrderPricingService_ValidateOrderPrice_PriceTooHigh(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	price := 120.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketPrice := &MarketPrice{Symbol: "PETR4", LastPrice: 100.0}
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)

	err := service.ValidateOrderPrice(order, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is too high")
}

func TestOrderPricingService_ValidateOrderPrice_PriceTooLow(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	price := 80.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketPrice := &MarketPrice{Symbol: "PETR4", LastPrice: 100.0}
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)

	err := service.ValidateOrderPrice(order, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is too low")
}

func TestOrderPricingService_ValidateOrderPrice_WideSpread(t *testing.T) {
	s := &orderPricingService{spreadWarningPercent: 1.0}
	mockClient := new(MockPricingDataClient)
	price := 101.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketPrice := &MarketPrice{Symbol: "PETR4", LastPrice: 100.0, SpreadPercent: 1.5}
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)

	err := s.ValidateOrderPrice(order, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wide spread detected")
}

func TestOrderPricingService_EstimateFillPrice(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	marketPrice := &MarketPrice{Symbol: "PETR4", AskPrice: 101}
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)
	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(&MarketDepth{LiquidityScore: 0.7}, nil)

	price, err := service.EstimateFillPrice(order, mockClient)
	assert.NoError(t, err)
	assert.True(t, price > 101) // Price should be ask + slippage
}

func TestOrderPricingService_CalculateTradingCosts(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	price := 100.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	tradingFees := &TradingFees{TotalFees: 5.0}
	mockClient.On("GetTradingFees", order.OrderType(), order.CalculateOrderValue()).Return(tradingFees, nil)

	fees, err := service.CalculateTradingCosts(order, mockClient)
	assert.NoError(t, err)
	assert.NotNil(t, fees)
	assert.Equal(t, 5.0, fees.TotalFees)
}

func TestOrderPricingService_AssessPriceImpact(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	priceImpact := &PriceImpact{EstimatedImpact: 0.1}
	mockClient.On("GetPriceImpactEstimate", order.Symbol(), order.OrderSide(), order.Quantity()).Return(priceImpact, nil)

	impact, err := service.AssessPriceImpact(order, mockClient)
	assert.NoError(t, err)
	assert.NotNil(t, impact)
	assert.Equal(t, LiquidityRiskLow, impact.LiquidityRisk)
}

func TestOrderPricingService_RecommendExecutionStrategy(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	price := 10.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10000, &price)

	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(&MarketDepth{LiquidityScore: 0.2}, nil)
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(&MarketPrice{Volume: 500000}, nil)

	strategy, err := service.RecommendExecutionStrategy(order, mockClient)
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStrategyTWAP, strategy)
}

func TestOrderPricingService_ValidateMarketConditions_MarketClosed(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	mockClient.On("IsMarketOpen", "PETR4").Return(false, nil)

	_, err := service.ValidateMarketConditions(order, mockClient)
	assert.Error(t, err)
	assert.Equal(t, "market is closed for symbol PETR4", err.Error())
}

func TestOrderPricingService_CalculateSlippageTolerance(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(&MarketDepth{LiquidityScore: 0.2}, nil)
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(&MarketPrice{SpreadPercent: 0.8}, nil)

	slippage, err := service.CalculateSlippageTolerance(order, mockClient)
	assert.NoError(t, err)
	assert.True(t, slippage > 0.1)
}

func TestOrderPricingService_CalculateSlippageTolerance_Error(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	mockClient.On("IsMarketOpen", "PETR4").Return(false, fmt.Errorf("some error"))

	slippage, err := service.CalculateSlippageTolerance(order, mockClient)
	assert.NoError(t, err)
	s, _ := service.(*orderPricingService)
	assert.Equal(t, s.maxSlippagePercent*0.5, slippage)
}

func Test_orderPricingService_assessMarketTrend(t *testing.T) {
	s := &orderPricingService{}
	tests := []struct {
		name          string
		marketDepth   *MarketDepth
		marketPrice   *MarketPrice
		expectedTrend MarketTrend
	}{
		{"Bullish", &MarketDepth{ImbalanceRatio: 0.7}, &MarketPrice{}, MarketTrendBullish},
		{"Bearish", &MarketDepth{ImbalanceRatio: 0.3}, &MarketPrice{}, MarketTrendBearish},
		{"Volatile", &MarketDepth{ImbalanceRatio: 0.5}, &MarketPrice{SpreadPercent: 1.5}, MarketTrendVolatile},
		{"Neutral", &MarketDepth{ImbalanceRatio: 0.5}, &MarketPrice{SpreadPercent: 0.5}, MarketTrendNeutral},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend := s.assessMarketTrend(tt.marketDepth, tt.marketPrice)
			assert.Equal(t, tt.expectedTrend, trend)
		})
	}
}

func Test_orderPricingService_generateExecutionInstructions(t *testing.T) {
	s := &orderPricingService{}
	plan := &ExecutionPlan{EstimatedFillPrice: 100.0}

	tests := []struct {
		name           string
		strategy       ExecutionStrategy
		expectedPhrase string
	}{
		{"Market", ExecutionStrategyMarket, "Execute as market order"},
		{"Limit", ExecutionStrategyLimit, "Place limit order at 100.00"},
		{"TWAP", ExecutionStrategyTWAP, "Execute using Time Weighted Average Price"},
		{"VWAP", ExecutionStrategyVWAP, "Execute using Volume Weighted Average Price"},
		{"Iceberg", ExecutionStrategyIceberg, "Use iceberg strategy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan.RecommendedStrategy = tt.strategy
			plan.ExecutionInstructions = []string{}
			s.generateExecutionInstructions(nil, plan)
			assert.Contains(t, plan.ExecutionInstructions[0], tt.expectedPhrase)
		})
	}
}

func Test_orderPricingService_addPriceLevelRecommendations(t *testing.T) {
	s := &orderPricingService{}
	result := &PricingResult{Warnings: []string{}}
	marketPrice := &MarketPrice{BidPrice: 99, AskPrice: 101}

	t.Run("Buy above ask", func(t *testing.T) {
		price := 102.0
		order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
		s.addPriceLevelRecommendations(order, marketPrice, result)
		assert.Contains(t, result.Warnings[0], "Buy limit price above market ask")
	})

	result.Warnings = []string{}
	t.Run("Sell below bid", func(t *testing.T) {
		price := 98.0
		order, _ := domain.NewOrder("u1", "s1", domain.OrderSideSell, domain.OrderTypeLimit, 1, &price)
		s.addPriceLevelRecommendations(order, marketPrice, result)
		assert.Contains(t, result.Warnings[0], "Sell limit price below market bid")
	})
}

func TestOrderPricingService_EstimateFillPrice_LimitOrder(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	price := 105.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	marketPrice := &MarketPrice{Symbol: "PETR4", AskPrice: 101}
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)

	fillPrice, err := service.EstimateFillPrice(order, mockClient)
	assert.NoError(t, err)
	assert.Equal(t, 105.0, fillPrice)
}

func TestOrderPricingService_AssessPriceImpact_MediumHigh(t *testing.T) {
	service := &orderPricingService{impactWarningPercent: 0.5}
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	// High impact
	priceImpactHigh := &PriceImpact{EstimatedImpact: 0.6}
	mockClient.On("GetPriceImpactEstimate", order.Symbol(), order.OrderSide(), order.Quantity()).Return(priceImpactHigh, nil).Once()
	impact, err := service.AssessPriceImpact(order, mockClient)
	assert.NoError(t, err)
	assert.Equal(t, LiquidityRiskHigh, impact.LiquidityRisk)

	// Medium impact
	priceImpactMedium := &PriceImpact{EstimatedImpact: 0.3}
	mockClient.On("GetPriceImpactEstimate", order.Symbol(), order.OrderSide(), order.Quantity()).Return(priceImpactMedium, nil).Once()
	impact, err = service.AssessPriceImpact(order, mockClient)
	assert.NoError(t, err)
	assert.Equal(t, LiquidityRiskMedium, impact.LiquidityRisk)
}

func Test_orderPricingService_getDefaultStrategy(t *testing.T) {
	s := &orderPricingService{}
	marketOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeMarket, 1, nil)
	assert.Equal(t, ExecutionStrategyMarket, s.getDefaultStrategy(marketOrder))

	price := 1.0
	limitOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	assert.Equal(t, ExecutionStrategyLimit, s.getDefaultStrategy(limitOrder))
}

func Test_orderPricingService_selectStrategyBasedOnConditions(t *testing.T) {
	s := &orderPricingService{}
	price := 500.0
	mediumOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 100, &price)
	marketConditions := &MarketConditions{}
	assert.Equal(t, ExecutionStrategyIceberg, s.selectStrategyBasedOnConditions(mediumOrder, marketConditions))

	wideSpreadConditions := &MarketConditions{SpreadCondition: SpreadConditionWide}
	price = 400.0
	mediumOrder, _ = domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 100, &price)
	assert.Equal(t, ExecutionStrategyLimit, s.selectStrategyBasedOnConditions(mediumOrder, wideSpreadConditions))
}

func Test_orderPricingService_selectLargeOrderStrategy(t *testing.T) {
	s := &orderPricingService{}
	highVolume := &MarketConditions{TradingVolume: 2000000}
	assert.Equal(t, ExecutionStrategyVWAP, s.selectLargeOrderStrategy(highVolume))

	lowVolume := &MarketConditions{TradingVolume: 500000}
	assert.Equal(t, ExecutionStrategyTWAP, s.selectLargeOrderStrategy(lowVolume))
}

func Test_orderPricingService_setSellOrderPriceRange(t *testing.T) {
	s := &orderPricingService{}
	marketPrice := &MarketPrice{BidPrice: 100, AskPrice: 102, Spread: 2}
	priceRange := &PriceRange{}
	s.setSellOrderPriceRange(marketPrice, 4, priceRange)
	assert.Equal(t, 96.0, priceRange.MinPrice)
	assert.Equal(t, 102.0, priceRange.MaxPrice)
}

func Test_orderPricingService_setFillProbabilityAndTime(t *testing.T) {
	s := &orderPricingService{}
	estimate := &ExecutionEstimate{}
	price := 1.0
	limitOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	marketPrice := &MarketPrice{BidPrice: 0.9, AskPrice: 1.1, Spread: 0.2}
	s.setLimitOrderEstimate(limitOrder, marketPrice, estimate)
	assert.True(t, estimate.FillProbability > 0)

	_, err := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeStopLimit, 1, &price)
	assert.NoError(t, err)
	s.setDefaultOrderEstimate(estimate)
	assert.Equal(t, 0.7, estimate.FillProbability)
}

func Test_orderPricingService_estimateLimitOrderFillPrice(t *testing.T) {
	s := &orderPricingService{}
	price := 10.0
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	fillPrice, err := s.estimateLimitOrderFillPrice(order, nil)
	assert.NoError(t, err)
	assert.Equal(t, 10.0, fillPrice)
}

func Test_orderPricingService_adjustFeesBasedOnMethod(t *testing.T) {
	s := &orderPricingService{feeCalculationMethod: FeeCalculationPercentage}
	fees := &TradingFees{FeePercent: 0.1}
	price := 100.0
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)
	s.adjustFeesBasedOnMethod(fees, order)
	assert.Equal(t, 1.0, fees.TotalFees)
}

func Test_orderPricingService_calculateLimitOrderFillProbability(t *testing.T) {
	s := &orderPricingService{}
	marketPrice := &MarketPrice{BidPrice: 100, AskPrice: 102, Spread: 2}

	// Buy order
	buyPrice := 101.0
	buyOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &buyPrice)
	prob := s.calculateLimitOrderFillProbability(buyOrder, marketPrice)
	assert.True(t, prob > 0.3 && prob < 0.8)

	// Sell order
	sellPrice := 101.0
	sellOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideSell, domain.OrderTypeLimit, 1, &sellPrice)
	prob = s.calculateLimitOrderFillProbability(sellOrder, marketPrice)
	assert.True(t, prob > 0.3 && prob < 0.8)
}

func Test_orderPricingService_determineTimeInForce(t *testing.T) {
	s := &orderPricingService{}
	marketOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeMarket, 1, nil)
	assert.Equal(t, TimeInForceIOC, s.determineTimeInForce(marketOrder))

	price := 1.0
	limitOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	assert.Equal(t, TimeInForceDay, s.determineTimeInForce(limitOrder))

	stopOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeStopLimit, 1, &price)
	assert.Equal(t, TimeInForceGTC, s.determineTimeInForce(stopOrder))
}

func Test_orderPricingService_addSpreadBasedRecommendations(t *testing.T) {
	s := &orderPricingService{spreadWarningPercent: 1.0}
	result := &PricingResult{Warnings: []string{}, Recommendations: []string{}}
	marketPrice := &MarketPrice{SpreadPercent: 1.5}
	s.addSpreadBasedRecommendations(marketPrice, result)
	assert.NotEmpty(t, result.Warnings)
	assert.NotEmpty(t, result.Recommendations)
}

func TestOrderPricingService_CreateExecutionPlan_ErrorHandling(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	// Error in RecommendExecutionStrategy
	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil).Once()
	mockClient.On("GetMarketDepth", "PETR4").Return(nil, fmt.Errorf("depth error")).Once()
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(nil, fmt.Errorf("price error")).Once()
	_, err := service.CreateExecutionPlan(order, mockClient)
	assert.Error(t, err)
}

func TestOrderPricingService_CalculateOptimalPrice_SellOrder(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideSell, domain.OrderTypeMarket, 10, nil)

	marketPrice := &MarketPrice{Symbol: "PETR4", BidPrice: 100, AskPrice: 101, LastPrice: 100.5, Spread: 1, SpreadPercent: 1}
	marketDepth := &MarketDepth{LiquidityScore: 0.7}

	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)
	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(marketDepth, nil)

	result, err := service.CalculateOptimalPrice(order, mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, marketPrice.BidPrice, result.RecommendedPrice)
}

func Test_orderPricingService_calculateOptimalPriceForOrder_Sell(t *testing.T) {
	s := &orderPricingService{}
	marketPrice := &MarketPrice{BidPrice: 100, AskPrice: 102, Spread: 2, LastPrice: 101}
	price := 101.0
	limitOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideSell, domain.OrderTypeLimit, 1, &price)
	optimalPrice, err := s.calculateOptimalPriceForOrder(limitOrder, marketPrice)
	assert.NoError(t, err)
	assert.True(t, optimalPrice < 102 && optimalPrice > 100)

	stopOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideSell, domain.OrderTypeStopLimit, 1, &price)
	optimalPrice, err = s.calculateOptimalPriceForOrder(stopOrder, marketPrice)
	assert.NoError(t, err)
	assert.Equal(t, 101.0, optimalPrice)
}

func Test_orderPricingService_calculatePartialFillRisk(t *testing.T) {
	s := &orderPricingService{}
	price := 100.0
	largeOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1000, &price)
	risk := s.calculatePartialFillRisk(largeOrder, nil)
	assert.Equal(t, 0.6, risk)

	mediumOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 500, &price)
	risk = s.calculatePartialFillRisk(mediumOrder, nil)
	assert.Equal(t, 0.4, risk)

	smallOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 100, &price)
	risk = s.calculatePartialFillRisk(smallOrder, nil)
	assert.Equal(t, 0.2, risk)
}

func Test_orderPricingService_shouldAllowPartialFills(t *testing.T) {
	s := &orderPricingService{}
	price := 100.0
	largeOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 600, &price)
	assert.True(t, s.shouldAllowPartialFills(largeOrder, &ExecutionPlan{}))

	smallOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 100, &price)
	plan := &ExecutionPlan{RecommendedStrategy: ExecutionStrategyTWAP}
	assert.True(t, s.shouldAllowPartialFills(smallOrder, plan))

	plan.RecommendedStrategy = ExecutionStrategyMarket
	assert.False(t, s.shouldAllowPartialFills(smallOrder, plan))
}

func Test_orderPricingService_addOrderSizeRecommendations(t *testing.T) {
	s := &orderPricingService{}
	result := &PricingResult{Recommendations: []string{}}
	price := 100.0
	largeOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1001, &price)
	s.addOrderSizeRecommendations(largeOrder, result)
	assert.NotEmpty(t, result.Recommendations)
}

func Test_orderPricingService_selectStrategyBasedOnConditions_Market(t *testing.T) {
	s := &orderPricingService{}
	marketOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeMarket, 1, nil)
	marketConditions := &MarketConditions{}
	assert.Equal(t, ExecutionStrategyMarket, s.selectStrategyBasedOnConditions(marketOrder, marketConditions))
}

func Test_orderPricingService_CalculateSlippageTolerance_Conditions(t *testing.T) {
	s := &orderPricingService{maxSlippagePercent: 5.0}
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeMarket, 1, nil)

	mockClient.On("IsMarketOpen", "s1").Return(true, nil)
	mockClient.On("GetMarketDepth", "s1").Return(&MarketDepth{LiquidityScore: 0.1}, nil)       // Low liquidity
	mockClient.On("GetCurrentMarketPrice", "s1").Return(&MarketPrice{SpreadPercent: 2.0}, nil) // Very wide spread

	slippage, err := s.CalculateSlippageTolerance(order, mockClient)
	assert.NoError(t, err)
	assert.True(t, slippage > 0.5)
}

func Test_orderPricingService_estimateMarketOrderFillPrice_Sell(t *testing.T) {
	s := &orderPricingService{maxSlippagePercent: 1.0}
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideSell, domain.OrderTypeMarket, 1, nil)
	marketPrice := &MarketPrice{BidPrice: 100}
	mockClient.On("IsMarketOpen", "s1").Return(true, nil)
	mockClient.On("GetMarketDepth", "s1").Return(&MarketDepth{LiquidityScore: 0.7}, nil)
	mockClient.On("GetCurrentMarketPrice", "s1").Return(marketPrice, nil)

	price, err := s.estimateMarketOrderFillPrice(order, marketPrice, mockClient)
	assert.NoError(t, err)
	assert.True(t, price < 100)
}

func Test_orderPricingService_estimateLimitOrderFillPrice_NoPrice(t *testing.T) {
	s := &orderPricingService{}
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeMarket, 1, nil) // Market order has no price
	marketPrice := &MarketPrice{LastPrice: 100}
	price, err := s.estimateLimitOrderFillPrice(order, marketPrice)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, price)
}

func Test_orderPricingService_adjustFeesBasedOnMethod_Tiered(t *testing.T) {
	s := &orderPricingService{feeCalculationMethod: FeeCalculationTiered}
	fees := &TradingFees{CommissionFee: 10.0}
	price := 100.0
	largeOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1000, &price)
	s.adjustFeesBasedOnMethod(fees, largeOrder)
	assert.Equal(t, 8.0, fees.CommissionFee)

	fees.CommissionFee = 10.0
	mediumOrder, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 500, &price)
	s.adjustFeesBasedOnMethod(fees, mediumOrder)
	assert.Equal(t, 9.0, fees.CommissionFee)
}

func Test_orderPricingService_assessLiquidityLevel(t *testing.T) {
	s := &orderPricingService{}
	assert.Equal(t, LiquidityLevelVeryHigh, s.assessLiquidityLevel(&MarketDepth{LiquidityScore: 0.9}))
	assert.Equal(t, LiquidityLevelHigh, s.assessLiquidityLevel(&MarketDepth{LiquidityScore: 0.7}))
	assert.Equal(t, LiquidityLevelNormal, s.assessLiquidityLevel(&MarketDepth{LiquidityScore: 0.5}))
	assert.Equal(t, LiquidityLevelLow, s.assessLiquidityLevel(&MarketDepth{LiquidityScore: 0.3}))
}

func Test_orderPricingService_assessSpreadCondition(t *testing.T) {
	s := &orderPricingService{}
	assert.Equal(t, SpreadConditionTight, s.assessSpreadCondition(&MarketPrice{SpreadPercent: 0.05}))
	assert.Equal(t, SpreadConditionNormal, s.assessSpreadCondition(&MarketPrice{SpreadPercent: 0.3}))
	assert.Equal(t, SpreadConditionWide, s.assessSpreadCondition(&MarketPrice{SpreadPercent: 0.8}))
	assert.Equal(t, SpreadConditionVeryWide, s.assessSpreadCondition(&MarketPrice{SpreadPercent: 1.2}))
}

func Test_orderPricingService_calculateBuyOrderFillProbability(t *testing.T) {
	s := &orderPricingService{}
	marketPrice := &MarketPrice{BidPrice: 100, AskPrice: 102, Spread: 2}
	assert.Equal(t, 0.9, s.calculateBuyOrderFillProbability(103, marketPrice))
	assert.True(t, s.calculateBuyOrderFillProbability(101, marketPrice) > 0.3)
	assert.Equal(t, 0.1, s.calculateBuyOrderFillProbability(99, marketPrice))
}

func Test_orderPricingService_calculateSellOrderFillProbability(t *testing.T) {
	s := &orderPricingService{}
	marketPrice := &MarketPrice{BidPrice: 100, AskPrice: 102, Spread: 2}
	assert.Equal(t, 0.9, s.calculateSellOrderFillProbability(99, marketPrice))
	assert.True(t, s.calculateSellOrderFillProbability(101, marketPrice) > 0.3)
	assert.Equal(t, 0.1, s.calculateSellOrderFillProbability(103, marketPrice))
}

func Test_orderPricingService_calculateEstimatedFillTime(t *testing.T) {
	s := &orderPricingService{}
	marketPrice := &MarketPrice{BidPrice: 100, AskPrice: 102, Spread: 2}
	price := 102.0
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	assert.Equal(t, time.Minute*2, s.calculateEstimatedFillTime(order, marketPrice))

	price = 101.5
	order, _ = domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	assert.Equal(t, time.Minute*10, s.calculateEstimatedFillTime(order, marketPrice))

	price = 100.5
	order, _ = domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	assert.Equal(t, time.Hour, s.calculateEstimatedFillTime(order, marketPrice))

	price = 100.2
	order, _ = domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	assert.Equal(t, time.Hour*4, s.calculateEstimatedFillTime(order, marketPrice))

	price = 99.0
	order, _ = domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	assert.Equal(t, time.Hour*24, s.calculateEstimatedFillTime(order, marketPrice))
}

func Test_orderPricingService_calculateLimitOrderFillProbability_NoPrice(t *testing.T) {
	s := &orderPricingService{}
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeMarket, 1, nil)
	prob := s.calculateLimitOrderFillProbability(order, nil)
	assert.Equal(t, 0.5, prob)
}

func Test_orderPricingService_CalculateSlippageTolerance_Capped(t *testing.T) {
	s := &orderPricingService{maxSlippagePercent: 1.0}
	mockClient := new(MockPricingDataClient)
	price := 100.0
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1000, &price)

	mockClient.On("IsMarketOpen", "s1").Return(true, nil)
	mockClient.On("GetMarketDepth", "s1").Return(&MarketDepth{LiquidityScore: 0.1}, nil)       // Low liquidity
	mockClient.On("GetCurrentMarketPrice", "s1").Return(&MarketPrice{SpreadPercent: 2.0}, nil) // Very wide spread

	slippage, err := s.CalculateSlippageTolerance(order, mockClient)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, slippage)
}

func Test_orderPricingService_setFillProbabilityAndTime_MarketOrder(t *testing.T) {
	s := &orderPricingService{}
	estimate := &ExecutionEstimate{}
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeMarket, 1, nil)
	s.setFillProbabilityAndTime(order, nil, estimate)
	assert.Equal(t, 0.95, estimate.FillProbability)
	assert.Equal(t, time.Second*5, estimate.EstimatedFillTime)
}

func Test_orderPricingService_setFillProbabilityAndTime_LimitOrder(t *testing.T) {
	s := &orderPricingService{}
	estimate := &ExecutionEstimate{}
	price := 1.0
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideBuy, domain.OrderTypeLimit, 1, &price)
	marketPrice := &MarketPrice{BidPrice: 0.9, AskPrice: 1.1, Spread: 0.2}
	s.setFillProbabilityAndTime(order, marketPrice, estimate)
	assert.True(t, estimate.FillProbability > 0)
}

func TestOrderPricingService_CalculateOptimalPrice_ExecutionEstimateError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	marketPrice := &MarketPrice{Symbol: "PETR4", BidPrice: 100, AskPrice: 101, LastPrice: 100.5, Spread: 1, SpreadPercent: 1}

	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil).Once()
	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil).Twice()
	mockClient.On("GetMarketDepth", "PETR4").Return(nil, fmt.Errorf("depth error")).Twice()

	result, err := service.CalculateOptimalPrice(order, mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Warnings, "Could not assess market conditions: failed to get market depth: depth error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_CalculateOptimalPrice_MarketConditionsError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	marketPrice := &MarketPrice{Symbol: "PETR4", BidPrice: 100, AskPrice: 101, LastPrice: 100.5, Spread: 1, SpreadPercent: 1}

	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil).Once()
	mockClient.On("IsMarketOpen", "PETR4").Return(false, fmt.Errorf("market closed error")).Twice()

	result, err := service.CalculateOptimalPrice(order, mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Warnings, "Could not assess market conditions: failed to check market status: market closed error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_CreateExecutionPlan_FeeError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	marketPrice := &MarketPrice{Symbol: "PETR4", BidPrice: 100, AskPrice: 101, LastPrice: 100.5, Spread: 1, SpreadPercent: 1}
	marketDepth := &MarketDepth{LiquidityScore: 0.7}
	priceImpact := &PriceImpact{EstimatedImpact: 0.1}

	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(marketDepth, nil)
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)
	mockClient.On("GetTradingFees", order.OrderType(), order.CalculateOrderValue()).Return(nil, fmt.Errorf("fee error"))
	mockClient.On("GetPriceImpactEstimate", order.Symbol(), order.OrderSide(), order.Quantity()).Return(priceImpact, nil)

	plan, err := service.CreateExecutionPlan(order, mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Contains(t, plan.RiskWarnings, "Could not calculate fees: failed to get trading fees: fee error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_CreateExecutionPlan_PriceImpactError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	marketPrice := &MarketPrice{Symbol: "PETR4", BidPrice: 100, AskPrice: 101, LastPrice: 100.5, Spread: 1, SpreadPercent: 1}
	marketDepth := &MarketDepth{LiquidityScore: 0.7}
	tradingFees := &TradingFees{TotalFees: 5.0}

	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(marketDepth, nil)
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)
	mockClient.On("GetTradingFees", order.OrderType(), order.CalculateOrderValue()).Return(tradingFees, nil)
	mockClient.On("GetPriceImpactEstimate", order.Symbol(), order.OrderSide(), order.Quantity()).Return(nil, fmt.Errorf("impact error"))

	plan, err := service.CreateExecutionPlan(order, mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Contains(t, plan.RiskWarnings, "Could not assess price impact: failed to get price impact estimate: impact error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_ValidateOrderPrice_GetPriceError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	price := 101.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(nil, fmt.Errorf("network error"))

	err := service.ValidateOrderPrice(order, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get market price for validation: network error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_EstimateFillPrice_GetPriceError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(nil, fmt.Errorf("network error"))

	_, err := service.EstimateFillPrice(order, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get market price: network error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_CalculateTradingCosts_Error(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	price := 100.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

	mockClient.On("GetTradingFees", order.OrderType(), order.CalculateOrderValue()).Return(nil, fmt.Errorf("fee error"))

	_, err := service.CalculateTradingCosts(order, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get trading fees: fee error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_AssessPriceImpact_Error(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	mockClient.On("GetPriceImpactEstimate", order.Symbol(), order.OrderSide(), order.Quantity()).Return(nil, fmt.Errorf("impact error"))

	_, err := service.AssessPriceImpact(order, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get price impact estimate: impact error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_RecommendExecutionStrategy_Default(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	mockClient.On("IsMarketOpen", "PETR4").Return(false, fmt.Errorf("market closed error"))

	strategy, err := service.RecommendExecutionStrategy(order, mockClient)
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStrategyMarket, strategy)
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_ValidateMarketConditions_GetDepthError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(nil, fmt.Errorf("depth error"))

	_, err := service.ValidateMarketConditions(order, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get market depth: depth error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_ValidateMarketConditions_GetPriceError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(&MarketDepth{}, nil)
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(nil, fmt.Errorf("price error"))

	_, err := service.ValidateMarketConditions(order, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get market price: price error")
	mockClient.AssertExpectations(t)
}

func TestOrderPricingService_CalculateOptimalPrice_calculateOptimalPriceForOrderError(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order := domain.NewOrderFromRepository("id", "user1", "PETR4", domain.OrderSideBuy, "99", 10, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)

	marketPrice := &MarketPrice{Symbol: "PETR4", BidPrice: 100, AskPrice: 101, LastPrice: 100.5, Spread: 1, SpreadPercent: 1}

	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)
	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(&MarketDepth{}, nil)

	_, err := service.CalculateOptimalPrice(order, mockClient)

	assert.NoError(t, err)
}

func TestOrderPricingService_AssessPriceImpact_HighRisk(t *testing.T) {
	service := &orderPricingService{impactWarningPercent: 0.5}
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	priceImpact := &PriceImpact{EstimatedImpact: 0.6}
	mockClient.On("GetPriceImpactEstimate", order.Symbol(), order.OrderSide(), order.Quantity()).Return(priceImpact, nil)

	impact, err := service.AssessPriceImpact(order, mockClient)
	assert.NoError(t, err)
	assert.Equal(t, LiquidityRiskHigh, impact.LiquidityRisk)
}

func TestOrderPricingService_AssessPriceImpact_MediumRisk(t *testing.T) {
	service := &orderPricingService{impactWarningPercent: 0.5}
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	priceImpact := &PriceImpact{EstimatedImpact: 0.3}
	mockClient.On("GetPriceImpactEstimate", order.Symbol(), order.OrderSide(), order.Quantity()).Return(priceImpact, nil)

	impact, err := service.AssessPriceImpact(order, mockClient)
	assert.NoError(t, err)
	assert.Equal(t, LiquidityRiskMedium, impact.LiquidityRisk)
}

func TestOrderPricingService_CalculateSlippageTolerance_MediumOrder(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	price := 500.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 100, &price)

	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(&MarketDepth{LiquidityScore: 0.5}, nil)
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(&MarketPrice{SpreadPercent: 0.5}, nil)

	slippage, err := service.CalculateSlippageTolerance(order, mockClient)
	assert.NoError(t, err)
	assert.True(t, slippage > 0.1)
}

func TestOrderPricingService_setPriceRangeBasedOnOrderSide_Sell(t *testing.T) {
	s := &orderPricingService{}
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideSell, domain.OrderTypeMarket, 10, nil)
	marketPrice := &MarketPrice{BidPrice: 100, AskPrice: 102, Spread: 2}
	priceRange := &PriceRange{}
	s.setPriceRangeBasedOnOrderSide(order, marketPrice, priceRange)
	assert.Equal(t, 96.0, priceRange.MinPrice)
	assert.Equal(t, 102.0, priceRange.MaxPrice)
}

func TestOrderPricingService_setFillProbabilityAndTime_Default(t *testing.T) {
	s := &orderPricingService{}
	order := domain.NewOrderFromRepository("id", "user1", "PETR4", domain.OrderSideBuy, "99", 10, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
	estimate := &ExecutionEstimate{}
	s.setFillProbabilityAndTime(order, nil, estimate)
	assert.Equal(t, 0.7, estimate.FillProbability)
}

func TestOrderPricingService_validatePriceWithinRange_TooLow(t *testing.T) {
	s := &orderPricingService{}
	price := 89.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)
	marketPrice := &MarketPrice{LastPrice: 100.0}
	err := s.validatePriceWithinRange(order, *order.Price(), marketPrice)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is too low")
}

func TestOrderPricingService_estimateMarketOrderFillPrice_Sell(t *testing.T) {
	service := NewOrderPricingServiceWithDefaults()
	mockClient := new(MockPricingDataClient)
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideSell, domain.OrderTypeMarket, 10, nil)
	marketPrice := &MarketPrice{BidPrice: 100, AskPrice: 101, LastPrice: 100.5}

	mockClient.On("IsMarketOpen", "PETR4").Return(true, nil)
	mockClient.On("GetMarketDepth", "PETR4").Return(&MarketDepth{LiquidityScore: 0.7}, nil)
	mockClient.On("GetCurrentMarketPrice", "PETR4").Return(marketPrice, nil)

	price, err := service.EstimateFillPrice(order, mockClient)
	assert.NoError(t, err)
	assert.True(t, price < 100)
}

func TestOrderPricingService_adjustFeesBasedOnMethod_MediumOrder(t *testing.T) {
	s := &orderPricingService{feeCalculationMethod: FeeCalculationTiered}
	fees := &TradingFees{CommissionFee: 10.0}
	price := 600.0
	order, _ := domain.NewOrder("user1", "PETR4", domain.OrderSideBuy, domain.OrderTypeLimit, 100, &price)
	s.adjustFeesBasedOnMethod(fees, order)
	assert.Equal(t, 9.0, fees.CommissionFee)
}

func TestOrderPricingService_addPriceLevelRecommendations_Sell(t *testing.T) {
	s := &orderPricingService{}
	result := &PricingResult{Warnings: []string{}}
	marketPrice := &MarketPrice{BidPrice: 99, AskPrice: 101}
	price := 98.0
	order, _ := domain.NewOrder("u1", "s1", domain.OrderSideSell, domain.OrderTypeLimit, 1, &price)
	s.addPriceLevelRecommendations(order, marketPrice, result)
	assert.Contains(t, result.Warnings[0], "Sell limit price below market bid")
}