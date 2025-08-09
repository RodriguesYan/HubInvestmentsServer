package service

import (
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// IPricingDataClient defines the interface for pricing-related data operations (dependency inversion)
type IPricingDataClient interface {
	GetCurrentMarketPrice(symbol string) (*MarketPrice, error)
	GetOrderBookData(symbol string) (*OrderBookData, error)
	GetHistoricalPrices(symbol string, period time.Duration) ([]HistoricalPrice, error)
	GetMarketDepth(symbol string) (*MarketDepth, error)
	IsMarketOpen(symbol string) (bool, error)
	GetTradingFees(orderType domain.OrderType, orderValue float64) (*TradingFees, error)
	GetPriceImpactEstimate(symbol string, orderSide domain.OrderSide, quantity float64) (*PriceImpact, error)
}

// MarketPrice represents current market pricing information
type MarketPrice struct {
	Symbol        string
	BidPrice      float64
	AskPrice      float64
	LastPrice     float64
	Volume        int64
	Spread        float64
	SpreadPercent float64
	Timestamp     time.Time
}

// OrderBookData represents order book information
type OrderBookData struct {
	Symbol    string
	Bids      []PriceLevel
	Asks      []PriceLevel
	Timestamp time.Time
}

// PriceLevel represents a price level in the order book
type PriceLevel struct {
	Price    float64
	Quantity float64
	Orders   int
}

// HistoricalPrice represents historical price data
type HistoricalPrice struct {
	Symbol    string
	Price     float64
	Volume    int64
	Timestamp time.Time
}

// MarketDepth represents market depth information
type MarketDepth struct {
	Symbol         string
	BidDepth       float64
	AskDepth       float64
	ImbalanceRatio float64
	LiquidityScore float64
	LastUpdated    time.Time
}

// TradingFees represents fee structure for trading
type TradingFees struct {
	CommissionFee float64
	RegulatoryFee float64
	ExchangeFee   float64
	TotalFees     float64
	FeePercent    float64
}

// PriceImpact represents estimated price impact of an order
type PriceImpact struct {
	Symbol              string
	EstimatedImpact     float64
	EstimatedFillPrice  float64
	LiquidityRisk       LiquidityRisk
	RecommendedSlippage float64
	Timestamp           time.Time
}

// LiquidityRisk represents liquidity risk levels
type LiquidityRisk int32

const (
	LiquidityRiskLow LiquidityRisk = iota
	LiquidityRiskMedium
	LiquidityRiskHigh
	LiquidityRiskVeryHigh
)

// ExecutionPlan represents the plan for order execution
type ExecutionPlan struct {
	OrderID               string
	RecommendedStrategy   ExecutionStrategy
	EstimatedFillPrice    float64
	EstimatedFees         *TradingFees
	PriceImpact           *PriceImpact
	TimeInForce           TimeInForce
	SlippageTolerance     float64
	PartialFillAllowed    bool
	ExecutionInstructions []string
	RiskWarnings          []string
	CreatedAt             time.Time
}

// ExecutionStrategy represents different execution strategies
type ExecutionStrategy int32

const (
	ExecutionStrategyMarket ExecutionStrategy = iota
	ExecutionStrategyLimit
	ExecutionStrategyTWAP // Time Weighted Average Price
	ExecutionStrategyVWAP // Volume Weighted Average Price
	ExecutionStrategyIceberg
	ExecutionStrategyHidden
)

// TimeInForce represents order time in force options
type TimeInForce int32

const (
	TimeInForceDay TimeInForce = iota
	TimeInForceGTC             // Good Till Cancelled
	TimeInForceIOC             // Immediate Or Cancel
	TimeInForceFOK             // Fill Or Kill
)

// PricingResult represents the result of pricing calculations
type PricingResult struct {
	Symbol             string
	RecommendedPrice   float64
	PriceRange         PriceRange
	EstimatedExecution *ExecutionEstimate
	MarketConditions   *MarketConditions
	Recommendations    []string
	Warnings           []string
	CalculatedAt       time.Time
}

// PriceRange represents price range recommendations
type PriceRange struct {
	MinPrice      float64
	MaxPrice      float64
	OptimalPrice  float64
	CurrentSpread float64
}

// ExecutionEstimate represents execution estimates
type ExecutionEstimate struct {
	FillProbability   float64
	EstimatedFillTime time.Duration
	EstimatedSlippage float64
	PartialFillRisk   float64
}

// MarketConditions represents current market conditions
type MarketConditions struct {
	Volatility      float64
	LiquidityLevel  LiquidityLevel
	TradingVolume   int64
	MarketTrend     MarketTrend
	SpreadCondition SpreadCondition
}

// LiquidityLevel represents market liquidity levels
type LiquidityLevel int32

const (
	LiquidityLevelLow LiquidityLevel = iota
	LiquidityLevelNormal
	LiquidityLevelHigh
	LiquidityLevelVeryHigh
)

// MarketTrend represents market trend direction
type MarketTrend int32

const (
	MarketTrendNeutral MarketTrend = iota
	MarketTrendBullish
	MarketTrendBearish
	MarketTrendVolatile
)

// SpreadCondition represents bid-ask spread conditions
type SpreadCondition int32

const (
	SpreadConditionTight SpreadCondition = iota
	SpreadConditionNormal
	SpreadConditionWide
	SpreadConditionVeryWide
)

// OrderPricingService handles pricing calculations and execution logic
type OrderPricingService interface {
	// CalculateOptimalPrice calculates optimal pricing for an order
	CalculateOptimalPrice(order *domain.Order, pricingClient IPricingDataClient) (*PricingResult, error)

	// CreateExecutionPlan creates execution plan for an order
	CreateExecutionPlan(order *domain.Order, pricingClient IPricingDataClient) (*ExecutionPlan, error)

	// ValidateOrderPrice validates if order price is reasonable
	ValidateOrderPrice(order *domain.Order, pricingClient IPricingDataClient) error

	// EstimateFillPrice estimates the likely fill price for an order
	EstimateFillPrice(order *domain.Order, pricingClient IPricingDataClient) (float64, error)

	// CalculateTradingCosts calculates total trading costs including fees
	CalculateTradingCosts(order *domain.Order, pricingClient IPricingDataClient) (*TradingFees, error)

	// AssessPriceImpact assesses market impact of an order
	AssessPriceImpact(order *domain.Order, pricingClient IPricingDataClient) (*PriceImpact, error)

	// RecommendExecutionStrategy recommends best execution strategy
	RecommendExecutionStrategy(order *domain.Order, pricingClient IPricingDataClient) (ExecutionStrategy, error)

	// ValidateMarketConditions validates if market conditions are suitable for execution
	ValidateMarketConditions(order *domain.Order, pricingClient IPricingDataClient) (*MarketConditions, error)

	// CalculateSlippageTolerance calculates appropriate slippage tolerance
	CalculateSlippageTolerance(order *domain.Order, pricingClient IPricingDataClient) (float64, error)
}

type orderPricingService struct {
	// Configuration for pricing calculations
	maxSlippagePercent    float64
	minLiquidityThreshold float64
	spreadWarningPercent  float64
	impactWarningPercent  float64
	feeCalculationMethod  FeeCalculationMethod
}

// FeeCalculationMethod represents different fee calculation methods
type FeeCalculationMethod int32

const (
	FeeCalculationFixed FeeCalculationMethod = iota
	FeeCalculationTiered
	FeeCalculationPercentage
)

// OrderPricingConfig holds configuration for order pricing
type OrderPricingConfig struct {
	MaxSlippagePercent    float64              // Maximum allowed slippage percentage
	MinLiquidityThreshold float64              // Minimum liquidity threshold
	SpreadWarningPercent  float64              // Spread percentage for warnings
	ImpactWarningPercent  float64              // Price impact percentage for warnings
	FeeCalculationMethod  FeeCalculationMethod // Method for calculating fees
}

// NewOrderPricingService creates a new instance of OrderPricingService
func NewOrderPricingService(config OrderPricingConfig) OrderPricingService {
	return &orderPricingService{
		maxSlippagePercent:    config.MaxSlippagePercent,
		minLiquidityThreshold: config.MinLiquidityThreshold,
		spreadWarningPercent:  config.SpreadWarningPercent,
		impactWarningPercent:  config.ImpactWarningPercent,
		feeCalculationMethod:  config.FeeCalculationMethod,
	}
}

// NewOrderPricingServiceWithDefaults creates a service with default configuration
func NewOrderPricingServiceWithDefaults() OrderPricingService {
	return NewOrderPricingService(OrderPricingConfig{
		MaxSlippagePercent:    2.0,                  // 2% max slippage
		MinLiquidityThreshold: 10000.0,              // $10K minimum liquidity
		SpreadWarningPercent:  1.0,                  // 1% spread warning
		ImpactWarningPercent:  0.5,                  // 0.5% impact warning
		FeeCalculationMethod:  FeeCalculationTiered, // Tiered fee structure
	})
}

// CalculateOptimalPrice calculates optimal pricing for an order
func (s *orderPricingService) CalculateOptimalPrice(order *domain.Order, pricingClient IPricingDataClient) (*PricingResult, error) {
	result := &PricingResult{
		Symbol:          order.Symbol(),
		Recommendations: make([]string, 0),
		Warnings:        make([]string, 0),
		CalculatedAt:    time.Now(),
	}

	// Get current market price
	marketPrice, err := pricingClient.GetCurrentMarketPrice(order.Symbol())
	if err != nil {
		return result, fmt.Errorf("failed to get market price: %w", err)
	}

	// Calculate optimal price based on order type and side
	optimalPrice, err := s.calculateOptimalPriceForOrder(order, marketPrice)
	if err != nil {
		return result, fmt.Errorf("failed to calculate optimal price: %w", err)
	}

	result.RecommendedPrice = optimalPrice

	// Calculate price range
	priceRange, err := s.calculatePriceRange(order, marketPrice)
	if err != nil {
		return result, fmt.Errorf("failed to calculate price range: %w", err)
	}

	result.PriceRange = *priceRange

	// Get execution estimate
	executionEstimate, err := s.calculateExecutionEstimate(order, marketPrice, pricingClient)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not calculate execution estimate: %s", err.Error()))
	} else {
		result.EstimatedExecution = executionEstimate
	}

	// Assess market conditions
	marketConditions, err := s.ValidateMarketConditions(order, pricingClient)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not assess market conditions: %s", err.Error()))
	} else {
		result.MarketConditions = marketConditions
	}

	// Generate recommendations
	s.generatePricingRecommendations(order, result, marketPrice)

	return result, nil
}

// CreateExecutionPlan creates execution plan for an order
func (s *orderPricingService) CreateExecutionPlan(order *domain.Order, pricingClient IPricingDataClient) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		OrderID:               order.ID(),
		ExecutionInstructions: make([]string, 0),
		RiskWarnings:          make([]string, 0),
		CreatedAt:             time.Now(),
	}

	// Recommend execution strategy
	strategy, err := s.RecommendExecutionStrategy(order, pricingClient)
	if err != nil {
		return plan, fmt.Errorf("failed to recommend execution strategy: %w", err)
	}

	plan.RecommendedStrategy = strategy

	// Estimate fill price
	fillPrice, err := s.EstimateFillPrice(order, pricingClient)
	if err != nil {
		return plan, fmt.Errorf("failed to estimate fill price: %w", err)
	}

	plan.EstimatedFillPrice = fillPrice

	// Calculate trading fees
	fees, err := s.CalculateTradingCosts(order, pricingClient)
	if err != nil {
		plan.RiskWarnings = append(plan.RiskWarnings, fmt.Sprintf("Could not calculate fees: %s", err.Error()))
	} else {
		plan.EstimatedFees = fees
	}

	// Assess price impact
	priceImpact, err := s.AssessPriceImpact(order, pricingClient)
	if err != nil {
		plan.RiskWarnings = append(plan.RiskWarnings, fmt.Sprintf("Could not assess price impact: %s", err.Error()))
	} else {
		plan.PriceImpact = priceImpact
	}

	// Calculate slippage tolerance
	slippage, err := s.CalculateSlippageTolerance(order, pricingClient)
	if err != nil {
		plan.RiskWarnings = append(plan.RiskWarnings, fmt.Sprintf("Could not calculate slippage: %s", err.Error()))
	} else {
		plan.SlippageTolerance = slippage
	}

	// Set time in force based on order type
	plan.TimeInForce = s.determineTimeInForce(order)

	// Set partial fill allowance
	plan.PartialFillAllowed = s.shouldAllowPartialFills(order, plan)

	// Generate execution instructions
	s.generateExecutionInstructions(order, plan)

	return plan, nil
}

// ValidateOrderPrice validates if order price is reasonable
func (s *orderPricingService) ValidateOrderPrice(order *domain.Order, pricingClient IPricingDataClient) error {
	// Skip validation for market orders (no price specified)
	if order.OrderType() == domain.OrderTypeMarket {
		return nil
	}

	if order.Price() == nil {
		return fmt.Errorf("limit order must have a price specified")
	}

	marketPrice, err := pricingClient.GetCurrentMarketPrice(order.Symbol())
	if err != nil {
		return fmt.Errorf("failed to get market price for validation: %w", err)
	}

	orderPrice := *order.Price()

	// Validate price is within reasonable range
	if err := s.validatePriceWithinRange(order, orderPrice, marketPrice); err != nil {
		return err
	}

	// Check spread conditions
	if err := s.validateSpreadConditions(orderPrice, marketPrice); err != nil {
		return err
	}

	return nil
}

// EstimateFillPrice estimates the likely fill price for an order
func (s *orderPricingService) EstimateFillPrice(order *domain.Order, pricingClient IPricingDataClient) (float64, error) {
	marketPrice, err := pricingClient.GetCurrentMarketPrice(order.Symbol())
	if err != nil {
		return 0, fmt.Errorf("failed to get market price: %w", err)
	}

	switch order.OrderType() {
	case domain.OrderTypeMarket:
		return s.estimateMarketOrderFillPrice(order, marketPrice, pricingClient)
	case domain.OrderTypeLimit:
		return s.estimateLimitOrderFillPrice(order, marketPrice)
	default:
		// For other order types, use current market price as estimate
		if order.IsBuyOrder() {
			return marketPrice.AskPrice, nil
		}
		return marketPrice.BidPrice, nil
	}
}

// CalculateTradingCosts calculates total trading costs including fees
func (s *orderPricingService) CalculateTradingCosts(order *domain.Order, pricingClient IPricingDataClient) (*TradingFees, error) {
	orderValue := order.CalculateOrderValue()

	fees, err := pricingClient.GetTradingFees(order.OrderType(), orderValue)
	if err != nil {
		return nil, fmt.Errorf("failed to get trading fees: %w", err)
	}

	// Apply fee calculation method adjustments if needed
	s.adjustFeesBasedOnMethod(fees, order)

	return fees, nil
}

// AssessPriceImpact assesses market impact of an order
func (s *orderPricingService) AssessPriceImpact(order *domain.Order, pricingClient IPricingDataClient) (*PriceImpact, error) {
	priceImpact, err := pricingClient.GetPriceImpactEstimate(order.Symbol(), order.OrderSide(), order.Quantity())
	if err != nil {
		return nil, fmt.Errorf("failed to get price impact estimate: %w", err)
	}

	// Validate impact levels
	if priceImpact.EstimatedImpact > s.impactWarningPercent {
		priceImpact.LiquidityRisk = LiquidityRiskHigh
		return priceImpact, nil
	}

	if priceImpact.EstimatedImpact > s.impactWarningPercent*0.5 {
		priceImpact.LiquidityRisk = LiquidityRiskMedium
		return priceImpact, nil
	}

	priceImpact.LiquidityRisk = LiquidityRiskLow
	return priceImpact, nil
}

// RecommendExecutionStrategy recommends best execution strategy
func (s *orderPricingService) RecommendExecutionStrategy(order *domain.Order, pricingClient IPricingDataClient) (ExecutionStrategy, error) {
	// Get market conditions
	marketConditions, err := s.ValidateMarketConditions(order, pricingClient)
	if err != nil {
		return s.getDefaultStrategy(order), nil
	}

	return s.selectStrategyBasedOnConditions(order, marketConditions), nil
}

// getDefaultStrategy returns default strategy when market conditions unavailable
func (s *orderPricingService) getDefaultStrategy(order *domain.Order) ExecutionStrategy {
	if order.OrderType() == domain.OrderTypeMarket {
		return ExecutionStrategyMarket
	}
	return ExecutionStrategyLimit
}

// selectStrategyBasedOnConditions selects strategy based on order size and market conditions
func (s *orderPricingService) selectStrategyBasedOnConditions(order *domain.Order, marketConditions *MarketConditions) ExecutionStrategy {
	orderValue := order.CalculateOrderValue()

	// Large orders in low liquidity - use TWAP or VWAP
	if orderValue >= 100000 && marketConditions.LiquidityLevel <= LiquidityLevelNormal {
		return s.selectLargeOrderStrategy(marketConditions)
	}

	// Medium orders - consider iceberg strategy
	if orderValue >= 50000 {
		return ExecutionStrategyIceberg
	}

	// Wide spreads - prefer limit orders
	if marketConditions.SpreadCondition >= SpreadConditionWide {
		return ExecutionStrategyLimit
	}

	// Market orders
	if order.OrderType() == domain.OrderTypeMarket {
		return ExecutionStrategyMarket
	}

	return ExecutionStrategyLimit
}

// selectLargeOrderStrategy selects strategy for large orders based on volume
func (s *orderPricingService) selectLargeOrderStrategy(marketConditions *MarketConditions) ExecutionStrategy {
	if marketConditions.TradingVolume > 1000000 {
		return ExecutionStrategyVWAP
	}
	return ExecutionStrategyTWAP
}

// ValidateMarketConditions validates if market conditions are suitable for execution
func (s *orderPricingService) ValidateMarketConditions(order *domain.Order, pricingClient IPricingDataClient) (*MarketConditions, error) {
	conditions := &MarketConditions{}

	// Check if market is open
	isOpen, err := pricingClient.IsMarketOpen(order.Symbol())
	if err != nil {
		return conditions, fmt.Errorf("failed to check market status: %w", err)
	}

	if !isOpen {
		return conditions, fmt.Errorf("market is closed for symbol %s", order.Symbol())
	}

	// Get market depth
	marketDepth, err := pricingClient.GetMarketDepth(order.Symbol())
	if err != nil {
		return conditions, fmt.Errorf("failed to get market depth: %w", err)
	}

	// Assess liquidity level
	conditions.LiquidityLevel = s.assessLiquidityLevel(marketDepth)

	// Get market price for spread analysis
	marketPrice, err := pricingClient.GetCurrentMarketPrice(order.Symbol())
	if err != nil {
		return conditions, fmt.Errorf("failed to get market price: %w", err)
	}

	// Assess spread condition
	conditions.SpreadCondition = s.assessSpreadCondition(marketPrice)
	conditions.TradingVolume = marketPrice.Volume

	// Set volatility (would typically come from market data)
	conditions.Volatility = marketPrice.SpreadPercent // Simplified volatility measure

	// Determine market trend (simplified)
	conditions.MarketTrend = s.assessMarketTrend(marketDepth, marketPrice)

	return conditions, nil
}

// CalculateSlippageTolerance calculates appropriate slippage tolerance
func (s *orderPricingService) CalculateSlippageTolerance(order *domain.Order, pricingClient IPricingDataClient) (float64, error) {
	marketConditions, err := s.ValidateMarketConditions(order, pricingClient)
	if err != nil {
		// Return default slippage if conditions unavailable
		return s.maxSlippagePercent * 0.5, nil
	}

	baseSlippage := 0.1 // 0.1% base slippage

	// Adjust based on market conditions
	switch marketConditions.LiquidityLevel {
	case LiquidityLevelLow:
		baseSlippage *= 3.0
	case LiquidityLevelNormal:
		baseSlippage *= 1.5
	case LiquidityLevelHigh:
		baseSlippage *= 0.8
	case LiquidityLevelVeryHigh:
		baseSlippage *= 0.5
	}

	// Adjust based on spread conditions
	switch marketConditions.SpreadCondition {
	case SpreadConditionVeryWide:
		baseSlippage *= 2.5
	case SpreadConditionWide:
		baseSlippage *= 1.8
	case SpreadConditionNormal:
		baseSlippage *= 1.0
	case SpreadConditionTight:
		baseSlippage *= 0.7
	}

	// Adjust based on order size
	orderValue := order.CalculateOrderValue()
	if orderValue >= 100000 {
		baseSlippage *= 1.5
	}

	if orderValue >= 50000 && orderValue < 100000 {
		baseSlippage *= 1.2
	}

	// Cap at maximum allowed slippage
	if baseSlippage > s.maxSlippagePercent {
		baseSlippage = s.maxSlippagePercent
	}

	return baseSlippage, nil
}

// Helper methods

func (s *orderPricingService) calculateOptimalPriceForOrder(order *domain.Order, marketPrice *MarketPrice) (float64, error) {
	switch order.OrderType() {
	case domain.OrderTypeMarket:
		// Market orders use current market price
		if order.IsBuyOrder() {
			return marketPrice.AskPrice, nil
		}
		return marketPrice.BidPrice, nil

	case domain.OrderTypeLimit:
		// For limit orders, provide optimal price recommendations
		if order.IsBuyOrder() {
			// For buy orders, optimal price is slightly above bid but below ask
			return marketPrice.BidPrice + (marketPrice.Spread * 0.3), nil
		}
		// For sell orders, optimal price is slightly below ask but above bid
		return marketPrice.AskPrice - (marketPrice.Spread * 0.3), nil

	default:
		// For other order types, use last price
		return marketPrice.LastPrice, nil
	}
}

func (s *orderPricingService) calculatePriceRange(order *domain.Order, marketPrice *MarketPrice) (*PriceRange, error) {
	optimalPrice, err := s.calculateOptimalPriceForOrder(order, marketPrice)
	if err != nil {
		return nil, err
	}

	priceRange := &PriceRange{
		OptimalPrice:  optimalPrice,
		CurrentSpread: marketPrice.Spread,
	}

	s.setPriceRangeBasedOnOrderSide(order, marketPrice, priceRange)

	return priceRange, nil
}

// setPriceRangeBasedOnOrderSide sets min/max prices based on order side
func (s *orderPricingService) setPriceRangeBasedOnOrderSide(order *domain.Order, marketPrice *MarketPrice, priceRange *PriceRange) {
	spreadBuffer := marketPrice.Spread * 2 // 2x spread as buffer

	if order.IsBuyOrder() {
		s.setBuyOrderPriceRange(marketPrice, spreadBuffer, priceRange)
		return
	}

	s.setSellOrderPriceRange(marketPrice, spreadBuffer, priceRange)
}

// setBuyOrderPriceRange sets price range for buy orders
func (s *orderPricingService) setBuyOrderPriceRange(marketPrice *MarketPrice, spreadBuffer float64, priceRange *PriceRange) {
	priceRange.MinPrice = marketPrice.BidPrice
	priceRange.MaxPrice = marketPrice.AskPrice + spreadBuffer
}

// setSellOrderPriceRange sets price range for sell orders
func (s *orderPricingService) setSellOrderPriceRange(marketPrice *MarketPrice, spreadBuffer float64, priceRange *PriceRange) {
	priceRange.MinPrice = marketPrice.BidPrice - spreadBuffer
	priceRange.MaxPrice = marketPrice.AskPrice
}

func (s *orderPricingService) calculateExecutionEstimate(order *domain.Order, marketPrice *MarketPrice, pricingClient IPricingDataClient) (*ExecutionEstimate, error) {
	estimate := &ExecutionEstimate{}

	// Calculate fill probability and time based on order type
	s.setFillProbabilityAndTime(order, marketPrice, estimate)

	// Calculate estimated slippage
	s.setEstimatedSlippage(order, pricingClient, estimate)

	// Assess partial fill risk
	estimate.PartialFillRisk = s.calculatePartialFillRisk(order, marketPrice)

	return estimate, nil
}

// setFillProbabilityAndTime sets fill probability and estimated time based on order type
func (s *orderPricingService) setFillProbabilityAndTime(order *domain.Order, marketPrice *MarketPrice, estimate *ExecutionEstimate) {
	switch order.OrderType() {
	case domain.OrderTypeMarket:
		s.setMarketOrderEstimate(estimate)
	case domain.OrderTypeLimit:
		s.setLimitOrderEstimate(order, marketPrice, estimate)
	default:
		s.setDefaultOrderEstimate(estimate)
	}
}

// setMarketOrderEstimate sets estimates for market orders
func (s *orderPricingService) setMarketOrderEstimate(estimate *ExecutionEstimate) {
	estimate.FillProbability = 0.95 // High probability for market orders
	estimate.EstimatedFillTime = time.Second * 5
}

// setLimitOrderEstimate sets estimates for limit orders
func (s *orderPricingService) setLimitOrderEstimate(order *domain.Order, marketPrice *MarketPrice, estimate *ExecutionEstimate) {
	estimate.FillProbability = s.calculateLimitOrderFillProbability(order, marketPrice)
	estimate.EstimatedFillTime = s.calculateEstimatedFillTime(order, marketPrice)
}

// setDefaultOrderEstimate sets estimates for other order types
func (s *orderPricingService) setDefaultOrderEstimate(estimate *ExecutionEstimate) {
	estimate.FillProbability = 0.7
	estimate.EstimatedFillTime = time.Minute * 5
}

// setEstimatedSlippage calculates and sets estimated slippage
func (s *orderPricingService) setEstimatedSlippage(order *domain.Order, pricingClient IPricingDataClient, estimate *ExecutionEstimate) {
	slippage, err := s.CalculateSlippageTolerance(order, pricingClient)
	if err == nil {
		estimate.EstimatedSlippage = slippage
	}
}

func (s *orderPricingService) validatePriceWithinRange(order *domain.Order, orderPrice float64, marketPrice *MarketPrice) error {
	maxDeviation := marketPrice.LastPrice * 0.1 // 10% max deviation

	if orderPrice > marketPrice.LastPrice+maxDeviation {
		return fmt.Errorf("order price %.2f is too high (market: %.2f, max: %.2f)",
			orderPrice, marketPrice.LastPrice, marketPrice.LastPrice+maxDeviation)
	}

	if orderPrice < marketPrice.LastPrice-maxDeviation {
		return fmt.Errorf("order price %.2f is too low (market: %.2f, min: %.2f)",
			orderPrice, marketPrice.LastPrice, marketPrice.LastPrice-maxDeviation)
	}

	return nil
}

func (s *orderPricingService) validateSpreadConditions(orderPrice float64, marketPrice *MarketPrice) error {
	if marketPrice.SpreadPercent > s.spreadWarningPercent {
		return fmt.Errorf("wide spread detected (%.2f%%), consider market conditions before execution", marketPrice.SpreadPercent)
	}

	return nil
}

func (s *orderPricingService) estimateMarketOrderFillPrice(order *domain.Order, marketPrice *MarketPrice, pricingClient IPricingDataClient) (float64, error) {
	// For market orders, estimate fill price considering potential slippage
	basePrice := marketPrice.LastPrice
	if order.IsBuyOrder() {
		basePrice = marketPrice.AskPrice
	} else {
		basePrice = marketPrice.BidPrice
	}

	// Add estimated slippage
	slippage, err := s.CalculateSlippageTolerance(order, pricingClient)
	if err != nil {
		slippage = 0.1 // Default 0.1% slippage
	}

	slippageAmount := basePrice * (slippage / 100.0)
	if order.IsBuyOrder() {
		return basePrice + slippageAmount, nil
	}
	return basePrice - slippageAmount, nil
}

func (s *orderPricingService) estimateLimitOrderFillPrice(order *domain.Order, marketPrice *MarketPrice) (float64, error) {
	// For limit orders, fill price is the order price (if filled)
	if order.Price() != nil {
		return *order.Price(), nil
	}

	// Fallback to market price
	return marketPrice.LastPrice, nil
}

func (s *orderPricingService) adjustFeesBasedOnMethod(fees *TradingFees, order *domain.Order) {
	switch s.feeCalculationMethod {
	case FeeCalculationTiered:
		// Apply tiered fee structure adjustments
		orderValue := order.CalculateOrderValue()
		if orderValue >= 100000 {
			fees.CommissionFee *= 0.8 // 20% discount for large orders
		}

		if orderValue >= 50000 && orderValue < 100000 {
			fees.CommissionFee *= 0.9 // 10% discount for medium orders
		}
	case FeeCalculationPercentage:
		// Ensure percentage-based fees are applied correctly
		orderValue := order.CalculateOrderValue()
		fees.TotalFees = orderValue * (fees.FeePercent / 100.0)
	}
}

func (s *orderPricingService) assessLiquidityLevel(marketDepth *MarketDepth) LiquidityLevel {
	if marketDepth.LiquidityScore >= 0.8 {
		return LiquidityLevelVeryHigh
	}

	if marketDepth.LiquidityScore >= 0.6 {
		return LiquidityLevelHigh
	}

	if marketDepth.LiquidityScore >= 0.4 {
		return LiquidityLevelNormal
	}

	return LiquidityLevelLow
}

func (s *orderPricingService) assessSpreadCondition(marketPrice *MarketPrice) SpreadCondition {
	if marketPrice.SpreadPercent <= 0.1 {
		return SpreadConditionTight
	}

	if marketPrice.SpreadPercent <= 0.5 {
		return SpreadConditionNormal
	}

	if marketPrice.SpreadPercent <= 1.0 {
		return SpreadConditionWide
	}

	return SpreadConditionVeryWide
}

func (s *orderPricingService) assessMarketTrend(marketDepth *MarketDepth, marketPrice *MarketPrice) MarketTrend {
	// Simplified trend assessment based on order book imbalance
	if marketDepth.ImbalanceRatio > 0.6 {
		return MarketTrendBullish
	}

	if marketDepth.ImbalanceRatio < 0.4 {
		return MarketTrendBearish
	}

	if marketPrice.SpreadPercent > 1.0 {
		return MarketTrendVolatile
	}

	return MarketTrendNeutral
}

func (s *orderPricingService) calculateLimitOrderFillProbability(order *domain.Order, marketPrice *MarketPrice) float64 {
	if order.Price() == nil {
		return 0.5 // Default probability
	}

	orderPrice := *order.Price()

	if order.IsBuyOrder() {
		return s.calculateBuyOrderFillProbability(orderPrice, marketPrice)
	}

	return s.calculateSellOrderFillProbability(orderPrice, marketPrice)
}

// calculateBuyOrderFillProbability calculates fill probability for buy orders
func (s *orderPricingService) calculateBuyOrderFillProbability(orderPrice float64, marketPrice *MarketPrice) float64 {
	// Buy order fill probability increases as price approaches ask
	if orderPrice >= marketPrice.AskPrice {
		return 0.9
	}

	if orderPrice >= marketPrice.BidPrice {
		ratio := (orderPrice - marketPrice.BidPrice) / marketPrice.Spread
		return 0.3 + (ratio * 0.5) // 30-80% probability
	}

	return 0.1
}

// calculateSellOrderFillProbability calculates fill probability for sell orders
func (s *orderPricingService) calculateSellOrderFillProbability(orderPrice float64, marketPrice *MarketPrice) float64 {
	// Sell order fill probability increases as price approaches bid
	if orderPrice <= marketPrice.BidPrice {
		return 0.9
	}

	if orderPrice <= marketPrice.AskPrice {
		ratio := (marketPrice.AskPrice - orderPrice) / marketPrice.Spread
		return 0.3 + (ratio * 0.5) // 30-80% probability
	}

	return 0.1
}

func (s *orderPricingService) calculateEstimatedFillTime(order *domain.Order, marketPrice *MarketPrice) time.Duration {
	fillProbability := s.calculateLimitOrderFillProbability(order, marketPrice)

	// Higher fill probability = shorter estimated time
	switch {
	case fillProbability >= 0.8:
		return time.Minute * 2
	case fillProbability >= 0.6:
		return time.Minute * 10
	case fillProbability >= 0.4:
		return time.Hour * 1
	case fillProbability >= 0.2:
		return time.Hour * 4
	default:
		return time.Hour * 24
	}
}

func (s *orderPricingService) calculatePartialFillRisk(order *domain.Order, marketPrice *MarketPrice) float64 {
	orderValue := order.CalculateOrderValue()

	// Large orders have higher partial fill risk
	if orderValue >= 100000 {
		return 0.6
	}

	if orderValue >= 50000 {
		return 0.4
	}

	if orderValue >= 10000 {
		return 0.2
	}

	return 0.1
}

func (s *orderPricingService) determineTimeInForce(order *domain.Order) TimeInForce {
	switch order.OrderType() {
	case domain.OrderTypeMarket:
		return TimeInForceIOC // Market orders should fill immediately or cancel
	case domain.OrderTypeLimit:
		return TimeInForceDay // Limit orders good for day by default
	default:
		return TimeInForceGTC // Other orders good till cancelled
	}
}

func (s *orderPricingService) shouldAllowPartialFills(order *domain.Order, plan *ExecutionPlan) bool {
	orderValue := order.CalculateOrderValue()

	// Allow partial fills for large orders
	if orderValue >= 50000 {
		return true
	}

	// Allow partial fills for specific strategies
	if s.isPartialFillStrategy(plan.RecommendedStrategy) {
		return true
	}

	return false
}

// isPartialFillStrategy checks if strategy supports partial fills
func (s *orderPricingService) isPartialFillStrategy(strategy ExecutionStrategy) bool {
	return strategy == ExecutionStrategyTWAP ||
		strategy == ExecutionStrategyVWAP ||
		strategy == ExecutionStrategyIceberg
}

func (s *orderPricingService) generateExecutionInstructions(order *domain.Order, plan *ExecutionPlan) {
	switch plan.RecommendedStrategy {
	case ExecutionStrategyMarket:
		plan.ExecutionInstructions = append(plan.ExecutionInstructions,
			"Execute as market order for immediate fill",
			"Monitor for price impact during execution",
			"Consider order size relative to average volume")

	case ExecutionStrategyLimit:
		plan.ExecutionInstructions = append(plan.ExecutionInstructions,
			fmt.Sprintf("Place limit order at %.2f", plan.EstimatedFillPrice),
			"Monitor market conditions for price improvement",
			"Consider adjusting price if market moves significantly")

	case ExecutionStrategyTWAP:
		plan.ExecutionInstructions = append(plan.ExecutionInstructions,
			"Execute using Time Weighted Average Price strategy",
			"Split order into smaller chunks over time",
			"Monitor market impact and adjust timing")

	case ExecutionStrategyVWAP:
		plan.ExecutionInstructions = append(plan.ExecutionInstructions,
			"Execute using Volume Weighted Average Price strategy",
			"Align execution with historical volume patterns",
			"Increase pace during high volume periods")

	case ExecutionStrategyIceberg:
		plan.ExecutionInstructions = append(plan.ExecutionInstructions,
			"Use iceberg strategy to hide order size",
			"Display small portions of total order",
			"Refresh displayed quantity as portions fill")
	}
}

func (s *orderPricingService) generatePricingRecommendations(order *domain.Order, result *PricingResult, marketPrice *MarketPrice) {
	// Spread-based recommendations
	s.addSpreadBasedRecommendations(marketPrice, result)

	// Order size recommendations
	s.addOrderSizeRecommendations(order, result)

	// Price level recommendations
	s.addPriceLevelRecommendations(order, marketPrice, result)
}

// addSpreadBasedRecommendations adds recommendations based on spread conditions
func (s *orderPricingService) addSpreadBasedRecommendations(marketPrice *MarketPrice, result *PricingResult) {
	if marketPrice.SpreadPercent <= s.spreadWarningPercent {
		return
	}

	result.Warnings = append(result.Warnings,
		fmt.Sprintf("Wide spread detected (%.2f%%) - consider market conditions", marketPrice.SpreadPercent))
	result.Recommendations = append(result.Recommendations,
		"Consider using limit orders to avoid excessive costs",
		"Monitor market depth before execution")
}

// addOrderSizeRecommendations adds recommendations based on order size
func (s *orderPricingService) addOrderSizeRecommendations(order *domain.Order, result *PricingResult) {
	orderValue := order.CalculateOrderValue()

	if orderValue < 100000 {
		return
	}

	result.Recommendations = append(result.Recommendations,
		"Consider breaking large order into smaller portions",
		"Use advanced execution strategies (TWAP/VWAP)",
		"Monitor market impact during execution")
}

// addPriceLevelRecommendations adds recommendations based on price levels
func (s *orderPricingService) addPriceLevelRecommendations(order *domain.Order, marketPrice *MarketPrice, result *PricingResult) {
	if order.OrderType() != domain.OrderTypeLimit || order.Price() == nil {
		return
	}

	orderPrice := *order.Price()

	if order.IsBuyOrder() && orderPrice > marketPrice.AskPrice {
		result.Warnings = append(result.Warnings,
			"Buy limit price above market ask - will execute immediately")
		return
	}

	if order.IsSellOrder() && orderPrice < marketPrice.BidPrice {
		result.Warnings = append(result.Warnings,
			"Sell limit price below market bid - will execute immediately")
	}
}
