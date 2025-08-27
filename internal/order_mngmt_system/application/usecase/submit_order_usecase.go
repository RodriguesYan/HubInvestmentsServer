package usecase

import (
	"context"
	"fmt"
	"time"

	"HubInvestments/internal/order_mngmt_system/application/command"
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/internal/order_mngmt_system/domain/repository"
	"HubInvestments/internal/order_mngmt_system/domain/service"
	"HubInvestments/internal/order_mngmt_system/infra/external"
)

type ISubmitOrderUseCase interface {
	Execute(ctx context.Context, cmd *command.SubmitOrderCommand) (*command.SubmitOrderResult, error)
}

type SubmitOrderUseCase struct {
	orderRepository    repository.IOrderRepository
	marketDataClient   external.IMarketDataClient
	idempotencyService service.IIdempotencyService
}

type SubmitOrderUseCaseConfig struct {
	ValidationTimeout     time.Duration
	MarketDataTimeout     time.Duration
	EnableRiskValidation  bool
	EnablePriceValidation bool
}

func NewSubmitOrderUseCase(
	orderRepository repository.IOrderRepository,
	marketDataClient external.IMarketDataClient,
	idempotencyService service.IIdempotencyService,
) ISubmitOrderUseCase {
	return &SubmitOrderUseCase{
		orderRepository:    orderRepository,
		marketDataClient:   marketDataClient,
		idempotencyService: idempotencyService,
	}
}

func (uc *SubmitOrderUseCase) Execute(ctx context.Context, cmd *command.SubmitOrderCommand) (*command.SubmitOrderResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	idempotencyKey := uc.idempotencyService.GenerateKey(
		cmd.UserID, cmd.Symbol, cmd.OrderType, cmd.OrderSide, cmd.Quantity, cmd.Price)

	// Check if this order has already been processed
	idempotencyResult, err := uc.idempotencyService.CheckIdempotency(ctx, idempotencyKey, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("idempotency check failed: %w", err)
	}

	// If already processed, return existing result
	if idempotencyResult.IsProcessed {
		switch idempotencyResult.Status {
		case service.IdempotencyStatusCompleted:
			return &command.SubmitOrderResult{
				OrderID:                 idempotencyResult.OrderID,
				Status:                  "PENDING",
				Message:                 "Order already submitted (idempotent request)",
				MarketPriceAtSubmission: nil,
				EstimatedExecutionPrice: nil,
			}, nil
		case service.IdempotencyStatusFailed:
			return nil, fmt.Errorf("previous order submission failed: %s", idempotencyResult.Error)
		case service.IdempotencyStatusPending:
			return nil, fmt.Errorf("order submission is already in progress")
		}
	}

	// Store idempotency key as pending
	if err := uc.idempotencyService.StoreIdempotencyKey(ctx, idempotencyKey, cmd.UserID, 24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store idempotency key: %w", err)
	}

	// Process the order with idempotency protection
	result, err := uc.processOrderSubmission(ctx, cmd)
	if err != nil {
		// Mark idempotency as failed
		_ = uc.idempotencyService.FailIdempotency(ctx, idempotencyKey, cmd.UserID, err.Error())
		return nil, err
	}

	// Mark idempotency as completed
	if err := uc.idempotencyService.CompleteIdempotency(ctx, idempotencyKey, cmd.UserID, result.OrderID, result.Message); err != nil {
		// Log error but don't fail the request since order was successfully created
		fmt.Printf("Warning: Failed to complete idempotency: %v\n", err)
	}

	return result, nil
}

// processOrderSubmission handles the actual order processing logic
func (uc *SubmitOrderUseCase) processOrderSubmission(ctx context.Context, cmd *command.SubmitOrderCommand) (*command.SubmitOrderResult, error) {
	if err := uc.validateSymbolWithMarketData(ctx, cmd.Symbol); err != nil {
		return nil, fmt.Errorf("symbol validation failed: %w", err)
	}

	marketData, err := uc.getMarketDataForOrder(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	if err := uc.validateTradingHours(ctx, cmd.Symbol); err != nil {
		return nil, fmt.Errorf("trading hours validation failed: %w", err)
	}

	if err := uc.validateOrderPrice(cmd, marketData.CurrentPrice); err != nil {
		return nil, fmt.Errorf("price validation failed: %w", err)
	}

	orderSide, err := cmd.ToOrderSide()
	if err != nil {
		return nil, fmt.Errorf("invalid order side: %w", err)
	}

	orderType, err := cmd.ToOrderType()
	if err != nil {
		return nil, fmt.Errorf("invalid order type: %w", err)
	}

	order, err := domain.NewOrder(cmd.UserID, cmd.Symbol, orderSide, orderType, cmd.Quantity, cmd.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	order.SetMarketDataContext(marketData.CurrentPrice, marketData.Timestamp)

	if err := uc.performBusinessValidation(ctx, order, marketData); err != nil {
		return nil, fmt.Errorf("business validation failed: %w", err)
	}

	if err := uc.orderRepository.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	estimatedPrice := uc.calculateEstimatedExecutionPrice(order, marketData.CurrentPrice)

	result := &command.SubmitOrderResult{
		OrderID:                 order.ID(),
		Status:                  string(order.Status()),
		MarketPriceAtSubmission: &marketData.CurrentPrice,
		EstimatedExecutionPrice: estimatedPrice,
		Message:                 fmt.Sprintf("Order submitted successfully. %s", cmd.GetDescription()),
	}

	return result, nil
}

type MarketDataContext struct {
	CurrentPrice float64
	AssetDetails *external.AssetDetails
	TradingHours *external.TradingHours
	Timestamp    time.Time
}

func (uc *SubmitOrderUseCase) validateSymbolWithMarketData(ctx context.Context, symbol string) error {
	isValid, err := uc.marketDataClient.ValidateSymbol(ctx, symbol)
	if err != nil {
		return fmt.Errorf("market data service error: %w", err)
	}

	if !isValid {
		return fmt.Errorf("symbol %s is not valid or not tradeable", symbol)
	}

	return nil
}

func (uc *SubmitOrderUseCase) getMarketDataForOrder(ctx context.Context, cmd *command.SubmitOrderCommand) (*MarketDataContext, error) {
	currentPrice, err := uc.marketDataClient.GetCurrentPrice(ctx, cmd.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get current price: %w", err)
	}

	assetDetails, err := uc.marketDataClient.GetAssetDetails(ctx, cmd.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset details: %w", err)
	}

	tradingHours, err := uc.marketDataClient.GetTradingHours(ctx, cmd.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get trading hours: %w", err)
	}

	return &MarketDataContext{
		CurrentPrice: currentPrice,
		AssetDetails: assetDetails,
		TradingHours: tradingHours,
		Timestamp:    time.Now(),
	}, nil
}

func (uc *SubmitOrderUseCase) validateTradingHours(ctx context.Context, symbol string) error {
	isOpen, err := uc.marketDataClient.IsMarketOpen(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to check market hours: %w", err)
	}

	if !isOpen {
		return fmt.Errorf("market is closed for symbol %s", symbol)
	}

	return nil
}

func (uc *SubmitOrderUseCase) validateOrderPrice(cmd *command.SubmitOrderCommand, currentPrice float64) error {
	if cmd.IsMarketOrder() {
		return nil
	}

	if cmd.Price == nil {
		return fmt.Errorf("limit orders must have a price")
	}

	orderPrice := *cmd.Price

	// Define acceptable price deviation (e.g., 10% from current market price)
	maxDeviation := 0.10 // 10%
	minPrice := currentPrice * (1 - maxDeviation)
	maxPrice := currentPrice * (1 + maxDeviation)

	if orderPrice < minPrice || orderPrice > maxPrice {
		return fmt.Errorf("order price $%.2f is outside acceptable range ($%.2f - $%.2f) based on current market price $%.2f",
			orderPrice, minPrice, maxPrice, currentPrice)
	}

	if cmd.IsBuyOrder() {
		// For buy limit orders, price shouldn't be too far above market price
		if orderPrice > currentPrice*1.05 { // 5% above market
			return fmt.Errorf("buy limit price $%.2f is significantly above market price $%.2f", orderPrice, currentPrice)
		}
	}

	if cmd.IsSellOrder() {
		// For sell limit orders, price shouldn't be too far below market price
		if orderPrice < currentPrice*0.95 { // 5% below market
			return fmt.Errorf("sell limit price $%.2f is significantly below market price $%.2f", orderPrice, currentPrice)
		}
	}

	return nil
}

func (uc *SubmitOrderUseCase) performBusinessValidation(ctx context.Context, order *domain.Order, marketData *MarketDataContext) error {
	// integrate with the domain services in the future

	if err := order.Validate(); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	if !order.CanExecute() {
		return fmt.Errorf("order cannot be executed in current status: %s", order.Status())
	}

	// For example: position validation for sell orders, risk limits, etc.
	return nil
}

func (uc *SubmitOrderUseCase) calculateEstimatedExecutionPrice(order *domain.Order, currentPrice float64) *float64 {
	if order.OrderType() == domain.OrderTypeMarket {
		return &currentPrice
	}

	if order.Price() != nil {
		// For limit orders, estimated price is the limit price
		limitPrice := *order.Price()
		return &limitPrice
	}

	// Fallback to current price
	return &currentPrice
}
