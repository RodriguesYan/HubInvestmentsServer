package usecase

import (
	"context"
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/internal/order_mngmt_system/domain/repository"
	"HubInvestments/internal/order_mngmt_system/infra/external"
	"HubInvestments/internal/order_mngmt_system/infra/messaging"
)

type IProcessOrderUseCase interface {
	Execute(ctx context.Context, command *ProcessOrderCommand) (*ProcessOrderResult, error)
}

// ProcessOrderCommand contains the data needed to process an order
type ProcessOrderCommand struct {
	OrderID string
	Context ProcessingContext
}

// ProcessingContext provides additional context for order processing
type ProcessingContext struct {
	WorkerID     string
	ProcessingID string
	StartTime    time.Time
	MaxRetries   int
	RetryAttempt int
}

// ProcessOrderResult contains the result of order processing
type ProcessOrderResult struct {
	OrderID        string
	FinalStatus    string
	ExecutionPrice *float64
	ExecutionTime  *time.Time
	ProcessingTime time.Duration
	ErrorMessage   string
	WorkerID       string
	ProcessingID   string
}

type ProcessOrderUseCase struct {
	orderRepository  repository.IOrderRepository
	marketDataClient external.IMarketDataClient
	eventPublisher   messaging.IEventPublisher
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
	eventPublisher messaging.IEventPublisher,
) IProcessOrderUseCase {
	return &ProcessOrderUseCase{
		orderRepository:  orderRepository,
		marketDataClient: marketDataClient,
		eventPublisher:   eventPublisher,
	}
}

// Execute processes an order asynchronously with real-time market data
func (uc *ProcessOrderUseCase) Execute(ctx context.Context, command *ProcessOrderCommand) (*ProcessOrderResult, error) {
	startTime := time.Now()

	result := &ProcessOrderResult{
		OrderID:      command.OrderID,
		WorkerID:     command.Context.WorkerID,
		ProcessingID: command.Context.ProcessingID,
	}

	order, err := uc.orderRepository.FindByID(ctx, command.OrderID)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to find order %s: %v", command.OrderID, err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("failed to find order %s: %w", command.OrderID, err)
	}

	if order == nil {
		result.ErrorMessage = fmt.Sprintf("order %s not found", command.OrderID)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("order %s not found", command.OrderID)
	}

	if err := uc.validateOrderForProcessing(order); err != nil {
		uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Order validation failed: %v", err))
		result.FinalStatus = string(order.Status())
		result.ErrorMessage = fmt.Sprintf("Order validation failed: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("order validation failed: %w", err)
	}

	if err := uc.markOrderAsProcessing(ctx, order); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to mark order as processing: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("failed to mark order as processing: %w", err)
	}

	marketData, err := uc.getRealTimeMarketData(ctx, order.Symbol())
	if err != nil {
		uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Failed to get market data: %v", err))
		result.FinalStatus = string(order.Status())
		result.ErrorMessage = fmt.Sprintf("Failed to get market data: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("failed to get market data: %w", err)
	}

	if err := uc.validateMarketConditions(ctx, order, marketData); err != nil {
		uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Market conditions validation failed: %v", err))
		result.FinalStatus = string(order.Status())
		result.ErrorMessage = fmt.Sprintf("Market conditions validation failed: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("market conditions validation failed: %w", err)
	}

	executionPrice, err := uc.calculateExecutionPrice(ctx, order, marketData)
	if err != nil {
		uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Failed to calculate execution price: %v", err))
		result.FinalStatus = string(order.Status())
		result.ErrorMessage = fmt.Sprintf("Failed to calculate execution price: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("failed to calculate execution price: %w", err)
	}

	if err := uc.performFinalRiskChecks(ctx, order, marketData, executionPrice); err != nil {
		uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Final risk checks failed: %v", err))
		result.FinalStatus = string(order.Status())
		result.ErrorMessage = fmt.Sprintf("Final risk checks failed: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("final risk checks failed: %w", err)
	}

	if err := uc.executeOrder(ctx, order, executionPrice, marketData.Timestamp); err != nil {
		uc.markOrderAsFailed(ctx, order, fmt.Sprintf("Order execution failed: %v", err))
		result.FinalStatus = string(order.Status())
		result.ErrorMessage = fmt.Sprintf("Order execution failed: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("order execution failed: %w", err)
	}

	if err := uc.markOrderAsExecuted(ctx, order, executionPrice, marketData.Timestamp); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to mark order as executed: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("failed to mark order as executed: %w", err)
	}

	// Success case
	executionTime := marketData.Timestamp
	result.FinalStatus = string(order.Status())
	result.ExecutionPrice = &executionPrice
	result.ExecutionTime = &executionTime
	result.ProcessingTime = time.Since(startTime)

	return result, nil
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

	if err := uc.orderRepository.UpdateStatus(ctx, order.ID(), order.Status()); err != nil {
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
	if err := uc.orderRepository.UpdateStatus(ctx, order.ID(), order.Status()); err != nil {
		return fmt.Errorf("failed to update order execution in database: %w", err)
	}

	totalValue := executionPrice * order.Quantity()

	event := domain.NewOrderExecutedEvent(
		order.ID(),
		order.UserID(),
		order.Symbol(),
		order.OrderSide(),
		order.OrderType(),
		order.Quantity(),
		executionPrice,
		totalValue,
		executionTime,
		order.MarketPriceAtSubmission(),
		order.MarketDataTimestamp(),
	)

	if err := uc.eventPublisher.PublishOrderExecutedEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to publish order executed event: %w", err)
	}

	return nil
}

func (uc *ProcessOrderUseCase) markOrderAsFailed(ctx context.Context, order *domain.Order, errorMessage string) error {
	if err := order.MarkAsFailed(); err != nil {
		return fmt.Errorf("failed to mark order as failed: %w", err)
	}

	if err := uc.orderRepository.UpdateStatus(ctx, order.ID(), order.Status()); err != nil {
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
