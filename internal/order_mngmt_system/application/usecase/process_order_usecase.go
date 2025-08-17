package usecase

import (
	"context"
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/internal/order_mngmt_system/domain/repository"
	"HubInvestments/internal/order_mngmt_system/infra/external"
)

type IProcessOrderUseCase interface {
	Execute(ctx context.Context, orderID string) error
}

type ProcessOrderUseCase struct {
	orderRepository  repository.IOrderRepository
	marketDataClient external.IMarketDataClient
	// Domain services will be added when interfaces are properly defined
}

type ProcessOrderUseCaseConfig struct {
	MaxRetryAttempts      int
	RetryDelay            time.Duration
	ExecutionTimeout      time.Duration
	PriceTolerancePercent float64
}

func NewProcessOrderUseCase(
	orderRepository repository.IOrderRepository,
	marketDataClient external.IMarketDataClient,
) IProcessOrderUseCase {
	return &ProcessOrderUseCase{
		orderRepository:  orderRepository,
		marketDataClient: marketDataClient,
	}
}

// Execute processes an order asynchronously with real-time market data
func (uc *ProcessOrderUseCase) Execute(ctx context.Context, orderID string) error {
	order, err := uc.orderRepository.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("failed to find order %s: %w", orderID, err)
	}

	if order == nil {
		return fmt.Errorf("order %s not found", orderID)
	}

	if err := uc.validateOrderForProcessing(order); err != nil {
		return uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Order validation failed: %v", err))
	}

	if err := uc.markOrderAsProcessing(ctx, order); err != nil {
		return fmt.Errorf("failed to mark order as processing: %w", err)
	}

	marketData, err := uc.getRealTimeMarketData(ctx, order.Symbol())
	if err != nil {
		return uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Failed to get market data: %v", err))
	}

	if err := uc.validateMarketConditions(ctx, order, marketData); err != nil {
		return uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Market conditions validation failed: %v", err))
	}

	executionPrice, err := uc.calculateExecutionPrice(ctx, order, marketData)
	if err != nil {
		return uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Failed to calculate execution price: %v", err))
	}

	if err := uc.performFinalRiskChecks(ctx, order, marketData, executionPrice); err != nil {
		return uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Final risk checks failed: %v", err))
	}

	if err := uc.executeOrder(ctx, order, executionPrice, marketData.Timestamp); err != nil {
		return uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Order execution failed: %v", err))
	}

	if err := uc.markOrderAsExecuted(ctx, order, executionPrice, marketData.Timestamp); err != nil {
		return fmt.Errorf("failed to mark order as executed: %w", err)
	}

	return nil
}

type OrderExecutionContext struct {
	CurrentPrice   float64
	AssetDetails   *external.AssetDetails
	TradingHours   *external.TradingHours
	Timestamp      time.Time
	ExecutionPrice float64
}

func (uc *ProcessOrderUseCase) validateOrderForProcessing(order *domain.Order) error {
	if !order.CanExecute() {
		return fmt.Errorf("order cannot be executed in current status: %s", order.Status())
	}

	if order.Quantity() <= 0 {
		return fmt.Errorf("invalid order quantity: %f", order.Quantity())
	}

	return nil
}

func (uc *ProcessOrderUseCase) markOrderAsProcessing(ctx context.Context, order *domain.Order) error {
	if err := order.MarkAsProcessing(); err != nil {
		return fmt.Errorf("failed to mark order as processing: %w", err)
	}

	if err := uc.orderRepository.UpdateStatus(order.ID(), order.Status()); err != nil {
		return fmt.Errorf("failed to update order status in database: %w", err)
	}

	return nil
}

func (uc *ProcessOrderUseCase) getRealTimeMarketData(ctx context.Context, symbol string) (*OrderExecutionContext, error) {
	currentPrice, err := uc.marketDataClient.GetCurrentPrice(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get current price: %w", err)
	}

	assetDetails, err := uc.marketDataClient.GetAssetDetails(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset details: %w", err)
	}

	tradingHours, err := uc.marketDataClient.GetTradingHours(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get trading hours: %w", err)
	}

	return &OrderExecutionContext{
		CurrentPrice: currentPrice,
		AssetDetails: assetDetails,
		TradingHours: tradingHours,
		Timestamp:    time.Now(),
	}, nil
}

func (uc *ProcessOrderUseCase) validateMarketConditions(ctx context.Context, order *domain.Order, marketData *OrderExecutionContext) error {
	if !marketData.TradingHours.IsOpen {
		return fmt.Errorf("market is closed for symbol %s", order.Symbol())
	}

	if !marketData.AssetDetails.IsTradeable {
		return fmt.Errorf("asset %s is not tradeable", order.Symbol())
	}

	if order.Quantity() < marketData.AssetDetails.MinOrderSize {
		return fmt.Errorf("order quantity %f is below minimum %f", order.Quantity(), marketData.AssetDetails.MinOrderSize)
	}

	if order.Quantity() > marketData.AssetDetails.MaxOrderSize {
		return fmt.Errorf("order quantity %f exceeds maximum %f", order.Quantity(), marketData.AssetDetails.MaxOrderSize)
	}

	return nil
}

func (uc *ProcessOrderUseCase) calculateExecutionPrice(ctx context.Context, order *domain.Order, marketData *OrderExecutionContext) (float64, error) {
	switch order.OrderType() {
	case domain.OrderTypeMarket:
		// Market orders execute at current market price
		return marketData.CurrentPrice, nil

	case domain.OrderTypeLimit:
		// Limit orders execute at limit price if conditions are met
		return uc.calculateLimitOrderExecutionPrice(order, marketData)

	case domain.OrderTypeStopLoss:
		// Stop loss orders become market orders when triggered
		return uc.calculateStopLossExecutionPrice(order, marketData)

	case domain.OrderTypeStopLimit:
		// Stop limit orders become limit orders when triggered
		return uc.calculateStopLimitExecutionPrice(order, marketData)

	default:
		return 0, fmt.Errorf("unsupported order type: %s", order.OrderType())
	}
}

func (uc *ProcessOrderUseCase) calculateLimitOrderExecutionPrice(order *domain.Order, marketData *OrderExecutionContext) (float64, error) {
	if order.Price() == nil {
		return 0, fmt.Errorf("limit order must have a price")
	}

	limitPrice := *order.Price()
	currentPrice := marketData.CurrentPrice

	if order.OrderSide() == domain.OrderSideBuy {
		// Buy limit: execute if current price <= limit price
		if currentPrice <= limitPrice {
			return currentPrice, nil // Execute at better price
		}
		return 0, fmt.Errorf("buy limit order cannot be executed: current price %f > limit price %f", currentPrice, limitPrice)
	}

	// Sell limit: execute if current price >= limit price
	if currentPrice >= limitPrice {
		return currentPrice, nil // Execute at better price
	}
	return 0, fmt.Errorf("sell limit order cannot be executed: current price %f < limit price %f", currentPrice, limitPrice)
}

func (uc *ProcessOrderUseCase) calculateStopLossExecutionPrice(order *domain.Order, marketData *OrderExecutionContext) (float64, error) {
	if order.Price() == nil {
		return 0, fmt.Errorf("stop loss order must have a stop price")
	}

	stopPrice := *order.Price()
	currentPrice := marketData.CurrentPrice

	if order.OrderSide() == domain.OrderSideBuy {
		// Buy stop: triggered when price rises above stop price
		if currentPrice >= stopPrice {
			return currentPrice, nil
		}
		return 0, fmt.Errorf("buy stop order not triggered: current price %f < stop price %f", currentPrice, stopPrice)
	}

	// Sell stop: triggered when price falls below stop price
	if currentPrice <= stopPrice {
		return currentPrice, nil
	}
	return 0, fmt.Errorf("sell stop order not triggered: current price %f > stop price %f", currentPrice, stopPrice)
}

func (uc *ProcessOrderUseCase) calculateStopLimitExecutionPrice(order *domain.Order, marketData *OrderExecutionContext) (float64, error) {
	// In the future, we need to separate stop and limit prices
	return uc.calculateLimitOrderExecutionPrice(order, marketData)
}

func (uc *ProcessOrderUseCase) performFinalRiskChecks(ctx context.Context, order *domain.Order, marketData *OrderExecutionContext, executionPrice float64) error {
	if order.MarketPriceAtSubmission() == nil {
		return nil
	}

	// Check for significant price movement since submission
	submissionPrice := *order.MarketPriceAtSubmission()
	priceChange := abs(executionPrice-submissionPrice) / submissionPrice

	// If price moved more than 5% since submission, require additional validation
	if priceChange > 0.05 {
		return fmt.Errorf("significant price movement detected: %.2f%% change from submission price", priceChange*100)
	}

	return nil
}

func (uc *ProcessOrderUseCase) executeOrder(ctx context.Context, order *domain.Order, executionPrice float64, executionTime time.Time) error {
	// In a real implementation, this would integrate with:
	// 1. Broker APIs for actual trade execution
	// 2. Settlement systems
	// 3. Position management systems
	// 4. Accounting systems

	// For now, we'll simulate the execution
	if err := order.MarkAsExecuted(executionPrice); err != nil {
		return fmt.Errorf("failed to mark order as executed: %w", err)
	}

	return nil
}

func (uc *ProcessOrderUseCase) markOrderAsExecuted(ctx context.Context, order *domain.Order, executionPrice float64, executionTime time.Time) error {
	// In a complete implementation, you would use a more comprehensive update method
	if err := uc.orderRepository.UpdateStatus(order.ID(), order.Status()); err != nil {
		return fmt.Errorf("failed to update order execution in database: %w", err)
	}

	return nil
}

func (uc *ProcessOrderUseCase) markOrderAsFailed(ctx context.Context, order *domain.Order, errorMessage string) error {
	if err := order.MarkAsFailed(); err != nil {
		return fmt.Errorf("failed to mark order as failed: %w", err)
	}

	if err := uc.orderRepository.UpdateStatus(order.ID(), order.Status()); err != nil {
		return fmt.Errorf("failed to update failed order status in database: %w", err)
	}

	return fmt.Errorf("order processing failed: %s", errorMessage)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
