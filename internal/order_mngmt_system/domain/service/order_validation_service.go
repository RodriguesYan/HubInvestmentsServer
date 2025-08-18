package service

import (
	"context"
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// IMarketDataClient defines the interface for market data operations (dependency inversion)
type IMarketDataClient interface {
	ValidateSymbol(ctx context.Context, symbol string) (bool, error)
	GetCurrentPrice(ctx context.Context, symbol string) (float64, error)
	IsMarketOpen(ctx context.Context, symbol string) (bool, error)
	GetAssetDetails(ctx context.Context, symbol string) (*AssetDetails, error)
	GetTradingHours(ctx context.Context, symbol string) (*TradingHours, error)
}

// IPositionClient defines the interface for position operations (dependency inversion)
type IPositionClient interface {
	GetAvailableQuantity(userID, symbol string) (float64, error)
	HasSufficientBalance(userID string, requiredAmount float64) (bool, error)
}

// AssetDetails represents asset information from market data service
type AssetDetails struct {
	Symbol       string
	Name         string
	Category     int32
	LastQuote    float64
	IsActive     bool
	IsTradeable  bool
	MinOrderSize float64
	MaxOrderSize float64
	PriceStep    float64
	LastUpdated  time.Time
}

// TradingHours represents trading session information
type TradingHours struct {
	Symbol          string
	MarketOpen      time.Time
	MarketClose     time.Time
	IsOpen          bool
	NextOpenTime    time.Time
	NextCloseTime   time.Time
	Timezone        string
	ExtendedHours   bool
	PreMarketOpen   time.Time
	PostMarketClose time.Time
}

// ValidationContext provides context for order validation
type ValidationContext struct {
	Order             *domain.Order
	MarketData        *AssetDetails
	AvailableQuantity *float64
	AvailableBalance  *float64
	ValidationTime    time.Time
}

// ValidationResult contains the result of order validation
type ValidationResult struct {
	IsValid           bool
	Errors            []string
	Warnings          []string
	ValidationContext *ValidationContext
}

// OrderValidationService handles business validation rules for orders
type OrderValidationService interface {
	// ValidateOrder performs comprehensive order validation
	ValidateOrder(ctx context.Context, order *domain.Order) (*ValidationResult, error)

	// ValidateOrderWithContext performs validation with external data
	ValidateOrderWithContext(ctx context.Context, order *domain.Order, marketDataClient IMarketDataClient, positionClient IPositionClient) (*ValidationResult, error)

	// ValidateSymbol validates if a symbol is tradeable
	ValidateSymbol(ctx context.Context, symbol string, marketDataClient IMarketDataClient) (*ValidationResult, error)

	// ValidateQuantity validates order quantity
	ValidateQuantity(ctx context.Context, order *domain.Order, positionClient IPositionClient) (*ValidationResult, error)

	// ValidatePrice validates order price against market conditions
	ValidatePrice(ctx context.Context, order *domain.Order, marketDataClient IMarketDataClient) (*ValidationResult, error)

	// ValidateTradingHours validates if trading is allowed at current time
	ValidateTradingHours(ctx context.Context, symbol string, marketDataClient IMarketDataClient) (*ValidationResult, error)

	// ValidateOrderSide validates order side specific rules
	ValidateOrderSide(ctx context.Context, order *domain.Order, positionClient IPositionClient) (*ValidationResult, error)

	// ValidateRiskLimits validates order against risk management rules
	ValidateRiskLimits(ctx context.Context, order *domain.Order, positionClient IPositionClient) (*ValidationResult, error)
}

type orderValidationService struct {
	// Configuration for validation rules
	maxOrderValue         float64
	maxQuantityPerOrder   float64
	priceTolerancePercent float64
	minOrderValue         float64
}

// OrderValidationConfig holds configuration for order validation
type OrderValidationConfig struct {
	MaxOrderValue         float64 // Maximum allowed order value
	MaxQuantityPerOrder   float64 // Maximum quantity per order
	PriceTolerancePercent float64 // Price tolerance percentage for limit orders
	MinOrderValue         float64 // Minimum order value
}

// NewOrderValidationService creates a new instance of OrderValidationService
func NewOrderValidationService(config OrderValidationConfig) OrderValidationService {
	return &orderValidationService{
		maxOrderValue:         config.MaxOrderValue,
		maxQuantityPerOrder:   config.MaxQuantityPerOrder,
		priceTolerancePercent: config.PriceTolerancePercent,
		minOrderValue:         config.MinOrderValue,
	}
}

// NewOrderValidationServiceWithDefaults creates a service with default configuration
func NewOrderValidationServiceWithDefaults() OrderValidationService {
	return NewOrderValidationService(OrderValidationConfig{
		MaxOrderValue:         1000000.0, // $1M max order value
		MaxQuantityPerOrder:   10000.0,   // 10K shares max
		PriceTolerancePercent: 10.0,      // 10% price tolerance
		MinOrderValue:         1.0,       // $1 minimum order
	})
}

// ValidateOrder performs comprehensive order validation
func (s *orderValidationService) ValidateOrder(ctx context.Context, order *domain.Order) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		ValidationContext: &ValidationContext{
			Order:          order,
			ValidationTime: time.Now(),
		},
	}

	// Perform basic domain validation
	if err := order.Validate(); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Domain validation failed: %s", err.Error()))
	}

	// Validate order value limits
	s.validateOrderValueLimits(order, result)

	// Validate quantity limits
	s.validateQuantityLimits(order, result)

	// Validate order type specific rules
	s.validateOrderTypeRules(order, result)

	return result, nil
}

// ValidateOrderWithContext performs validation with external data
func (s *orderValidationService) ValidateOrderWithContext(ctx context.Context, order *domain.Order, marketDataClient IMarketDataClient, positionClient IPositionClient) (*ValidationResult, error) {
	// Start with basic validation
	result, err := s.ValidateOrder(ctx, order)
	if err != nil {
		return result, err
	}

	// Validate symbol and get market data
	if err := s.validateSymbolStep(ctx, order, marketDataClient, result); err != nil {
		return result, err
	}

	// Validate trading hours
	s.validateTradingHoursStep(ctx, order, marketDataClient, result)

	// Validate price if applicable
	if order.Price() != nil {
		s.validatePriceStep(ctx, order, marketDataClient, result)
	}

	// Validate order side specific rules (especially for sell orders)
	if err := s.validateOrderSideStep(ctx, order, positionClient, result); err != nil {
		return result, err
	}

	// Validate risk limits
	s.validateRiskLimitsStep(ctx, order, positionClient, result)

	return result, nil
}

// validateSymbolStep handles symbol validation with error handling
func (s *orderValidationService) validateSymbolStep(ctx context.Context, order *domain.Order, marketDataClient IMarketDataClient, result *ValidationResult) error {
	symbolResult, err := s.ValidateSymbol(ctx, order.Symbol(), marketDataClient)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Symbol validation failed: %s", err.Error()))
		result.IsValid = false
		return nil // Don't return error, just mark as invalid
	}

	s.mergeValidationResults(result, symbolResult)
	return nil
}

// validateTradingHoursStep handles trading hours validation with warning handling
func (s *orderValidationService) validateTradingHoursStep(ctx context.Context, order *domain.Order, marketDataClient IMarketDataClient, result *ValidationResult) {
	tradingResult, err := s.ValidateTradingHours(ctx, order.Symbol(), marketDataClient)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Trading hours validation warning: %s", err.Error()))
		return
	}

	s.mergeValidationResults(result, tradingResult)
}

// validatePriceStep handles price validation with warning handling
func (s *orderValidationService) validatePriceStep(ctx context.Context, order *domain.Order, marketDataClient IMarketDataClient, result *ValidationResult) {
	priceResult, err := s.ValidatePrice(ctx, order, marketDataClient)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Price validation warning: %s", err.Error()))
		return
	}

	s.mergeValidationResults(result, priceResult)
}

// validateOrderSideStep handles order side validation with error handling
func (s *orderValidationService) validateOrderSideStep(ctx context.Context, order *domain.Order, positionClient IPositionClient, result *ValidationResult) error {
	sideResult, err := s.ValidateOrderSide(ctx, order, positionClient)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Order side validation failed: %s", err.Error()))
		result.IsValid = false
		return nil // Don't return error, just mark as invalid
	}

	s.mergeValidationResults(result, sideResult)
	return nil
}

// validateRiskLimitsStep handles risk limits validation with warning handling
func (s *orderValidationService) validateRiskLimitsStep(ctx context.Context, order *domain.Order, positionClient IPositionClient, result *ValidationResult) {
	riskResult, err := s.ValidateRiskLimits(ctx, order, positionClient)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Risk validation warning: %s", err.Error()))
		return
	}

	s.mergeValidationResults(result, riskResult)
}

// ValidateSymbol validates if a symbol is tradeable
func (s *orderValidationService) ValidateSymbol(ctx context.Context, symbol string, marketDataClient IMarketDataClient) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		ValidationContext: &ValidationContext{
			ValidationTime: time.Now(),
		},
	}

	// Check if symbol exists and is valid
	if err := s.validateSymbolExistence(ctx, symbol, marketDataClient, result); err != nil {
		return result, err
	}

	// Return early if symbol is not valid
	if !result.IsValid {
		return result, nil
	}

	// Get asset details for additional validation
	s.validateAssetDetails(ctx, symbol, marketDataClient, result)

	return result, nil
}

// validateSymbolExistence checks if symbol exists and is valid
func (s *orderValidationService) validateSymbolExistence(ctx context.Context, symbol string, marketDataClient IMarketDataClient, result *ValidationResult) error {
	isValid, err := marketDataClient.ValidateSymbol(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to validate symbol: %w", err)
	}

	if !isValid {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Symbol '%s' is not valid or not tradeable", symbol))
	}

	return nil
}

// validateAssetDetails gets asset details and validates them
func (s *orderValidationService) validateAssetDetails(ctx context.Context, symbol string, marketDataClient IMarketDataClient, result *ValidationResult) {
	assetDetails, err := marketDataClient.GetAssetDetails(ctx, symbol)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not retrieve asset details: %s", err.Error()))
		return
	}

	result.ValidationContext.MarketData = assetDetails

	if !assetDetails.IsActive {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Symbol '%s' is not active for trading", symbol))
	}

	if !assetDetails.IsTradeable {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Symbol '%s' is not tradeable", symbol))
	}

	// Add informational warnings about asset details
	if assetDetails.MinOrderSize > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Minimum order size for %s is %.2f", symbol, assetDetails.MinOrderSize))
	}

	if assetDetails.MaxOrderSize > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Maximum order size for %s is %.2f", symbol, assetDetails.MaxOrderSize))
	}

	if assetDetails.PriceStep > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Price increment for %s is %.4f", symbol, assetDetails.PriceStep))
	}
}

// ValidateQuantity validates order quantity
func (s *orderValidationService) ValidateQuantity(ctx context.Context, order *domain.Order, positionClient IPositionClient) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		ValidationContext: &ValidationContext{
			Order:          order,
			ValidationTime: time.Now(),
		},
	}

	// Check quantity limits
	if order.Quantity() > s.maxQuantityPerOrder {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Order quantity %.2f exceeds maximum allowed %.2f", order.Quantity(), s.maxQuantityPerOrder))
	}

	// For sell orders, validate against available position
	if order.IsSellOrder() {
		return s.validateSellOrderQuantity(ctx, order, positionClient, result)
	}

	return result, nil
}

// validateSellOrderQuantity validates quantity for sell orders against available position
func (s *orderValidationService) validateSellOrderQuantity(ctx context.Context, order *domain.Order, positionClient IPositionClient, result *ValidationResult) (*ValidationResult, error) {
	availableQty, err := positionClient.GetAvailableQuantity(order.UserID(), order.Symbol())
	if err != nil {
		return result, fmt.Errorf("failed to get available quantity: %w", err)
	}

	result.ValidationContext.AvailableQuantity = &availableQty

	if err := order.ValidatePositionForSellOrder(availableQty); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// Warning if selling large percentage of position
	if availableQty > 0 && order.Quantity()/availableQty > 0.8 {
		result.Warnings = append(result.Warnings, "Selling more than 80% of available position")
	}

	return result, nil
}

// ValidatePrice validates order price against market conditions
func (s *orderValidationService) ValidatePrice(ctx context.Context, order *domain.Order, marketDataClient IMarketDataClient) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		ValidationContext: &ValidationContext{
			Order:          order,
			ValidationTime: time.Now(),
		},
	}

	if order.Price() == nil {
		return result, nil // No price to validate for market orders
	}

	// Get current market price
	currentPrice, err := marketDataClient.GetCurrentPrice(ctx, order.Symbol())
	if err != nil {
		return result, fmt.Errorf("failed to get current price: %w", err)
	}

	// Validate order against current market price
	if err := order.ValidateForExecution(currentPrice); err != nil {
		result.Warnings = append(result.Warnings, err.Error())
	}

	// Enhanced price range validation with market price ± tolerance
	tolerance := s.priceTolerancePercent / 100.0
	priceDiff := abs((*order.Price() - currentPrice) / currentPrice)

	upperLimit := currentPrice * (1 + tolerance)
	lowerLimit := currentPrice * (1 - tolerance)
	orderPrice := *order.Price()

	if priceDiff > tolerance {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Order price %.2f differs from market price %.2f by %.1f%% (tolerance: %.1f%%)",
			orderPrice, currentPrice, priceDiff*100, s.priceTolerancePercent))

		if orderPrice > upperLimit {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Order price %.2f exceeds upper limit %.2f", orderPrice, upperLimit))
		}

		if orderPrice < lowerLimit {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Order price %.2f below lower limit %.2f", orderPrice, lowerLimit))
		}
	}

	// Additional validation for extremely high price deviations
	extremeTolerancePercent := 50.0 // 50% extreme tolerance threshold
	extremeTolerance := extremeTolerancePercent / 100.0

	if priceDiff > extremeTolerance {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Order price %.2f deviates extremely from market price %.2f by %.1f%% (max allowed: %.1f%%)",
			orderPrice, currentPrice, priceDiff*100, extremeTolerancePercent))
	}

	return result, nil
}

// ValidateTradingHours validates if trading is allowed at current time
func (s *orderValidationService) ValidateTradingHours(ctx context.Context, symbol string, marketDataClient IMarketDataClient) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		ValidationContext: &ValidationContext{
			ValidationTime: time.Now(),
		},
	}

	isOpen, err := marketDataClient.IsMarketOpen(ctx, symbol)
	if err != nil {
		return result, fmt.Errorf("failed to check market hours: %w", err)
	}

	if !isOpen {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Market is currently closed for symbol '%s'", symbol))
	}

	// Get detailed trading hours for additional context
	tradingHours, err := marketDataClient.GetTradingHours(ctx, symbol)
	if err != nil {
		result.Warnings = append(result.Warnings, "Could not retrieve detailed trading hours")
	} else {
		if !tradingHours.IsOpen {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Market closed. Next open: %s", tradingHours.NextOpenTime.Format("2006-01-02 15:04:05 MST")))
		}
	}

	return result, nil
}

// ValidateOrderSide validates order side specific rules
func (s *orderValidationService) ValidateOrderSide(ctx context.Context, order *domain.Order, positionClient IPositionClient) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		ValidationContext: &ValidationContext{
			Order:          order,
			ValidationTime: time.Now(),
		},
	}

	if order.IsBuyOrder() {
		return s.validateBuyOrderSide(ctx, order, positionClient, result)
	}

	if order.IsSellOrder() {
		return s.validateSellOrderSide(ctx, result)
	}

	return result, nil
}

// validateBuyOrderSide validates buy order specific rules
func (s *orderValidationService) validateBuyOrderSide(ctx context.Context, order *domain.Order, positionClient IPositionClient, result *ValidationResult) (*ValidationResult, error) {
	orderValue := order.CalculateOrderValue()
	if orderValue <= 0 {
		return result, nil
	}

	hasSufficientBalance, err := positionClient.HasSufficientBalance(order.UserID(), orderValue)
	if err != nil {
		return result, fmt.Errorf("failed to check balance: %w", err)
	}

	if !hasSufficientBalance {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Insufficient balance for order value %.2f", orderValue))
	}

	return result, nil
}

// validateSellOrderSide validates sell order specific rules
func (s *orderValidationService) validateSellOrderSide(ctx context.Context, result *ValidationResult) (*ValidationResult, error) {
	// For sell orders, position validation is handled in ValidateQuantity
	// Additional sell-specific validations can be added here
	result.Warnings = append(result.Warnings, "Sell order - ensure you want to reduce your position")
	return result, nil
}

// ValidateRiskLimits validates order against risk management rules
func (s *orderValidationService) ValidateRiskLimits(ctx context.Context, order *domain.Order, positionClient IPositionClient) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		ValidationContext: &ValidationContext{
			Order:          order,
			ValidationTime: time.Now(),
		},
	}

	// Check order value limits
	orderValue := order.CalculateOrderValue()

	if orderValue > s.maxOrderValue {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Order value %.2f exceeds maximum allowed %.2f", orderValue, s.maxOrderValue))
	}

	if orderValue > 0 && orderValue < s.minOrderValue {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Order value %.2f is below minimum required %.2f", orderValue, s.minOrderValue))
	}

	// Risk warning for large orders
	if orderValue > s.maxOrderValue*0.1 { // 10% of max order value
		result.Warnings = append(result.Warnings, fmt.Sprintf("Large order value: %.2f", orderValue))
	}

	return result, nil
}

// Helper methods

func (s *orderValidationService) validateOrderValueLimits(order *domain.Order, result *ValidationResult) {
	orderValue := order.CalculateOrderValue()

	if orderValue > s.maxOrderValue {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Order value %.2f exceeds maximum allowed %.2f", orderValue, s.maxOrderValue))
	}

	if orderValue > 0 && orderValue < s.minOrderValue {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Order value %.2f is below minimum required %.2f", orderValue, s.minOrderValue))
	}
}

func (s *orderValidationService) validateQuantityLimits(order *domain.Order, result *ValidationResult) {
	if order.Quantity() > s.maxQuantityPerOrder {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Order quantity %.2f exceeds maximum allowed %.2f", order.Quantity(), s.maxQuantityPerOrder))
	}

	if order.Quantity() <= 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Order quantity must be positive")
	}
}

func (s *orderValidationService) validateOrderTypeRules(order *domain.Order, result *ValidationResult) {
	switch order.OrderType() {
	case domain.OrderTypeLimit:
		if order.Price() == nil {
			result.IsValid = false
			result.Errors = append(result.Errors, "Limit orders must have a price")
		} else if *order.Price() <= 0 {
			result.IsValid = false
			result.Errors = append(result.Errors, "Limit order price must be positive")
		}
	case domain.OrderTypeMarket:
		if order.Price() != nil {
			result.Warnings = append(result.Warnings, "Market orders should not have a price specified")
		}
	}
}

func (s *orderValidationService) mergeValidationResults(target *ValidationResult, source *ValidationResult) {
	if !source.IsValid {
		target.IsValid = false
	}
	target.Errors = append(target.Errors, source.Errors...)
	target.Warnings = append(target.Warnings, source.Warnings...)

	// Merge validation context if source has market data
	if source.ValidationContext == nil {
		return
	}

	if source.ValidationContext.MarketData == nil {
		return
	}

	if target.ValidationContext == nil {
		target.ValidationContext = &ValidationContext{}
	}

	target.ValidationContext.MarketData = source.ValidationContext.MarketData

}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
