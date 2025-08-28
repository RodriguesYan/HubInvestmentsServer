package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

func TestProcessOrderUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			return order, nil
		},
		SaveFunc: func(ctx context.Context, order *domain.Order) error {
			return nil
		},
	}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			return 149.50, nil
		},
	}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "order123",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.OrderID != cmd.OrderID {
		t.Errorf("Expected OrderID %s, got %s", cmd.OrderID, result.OrderID)
	}

	if result.FinalStatus != "EXECUTED" {
		t.Errorf("Expected FinalStatus EXECUTED, got %s", result.FinalStatus)
	}

	if result.ExecutionPrice == nil {
		t.Error("Expected ExecutionPrice to be set")
	}

	if *result.ExecutionPrice != 149.50 {
		t.Errorf("Expected ExecutionPrice 149.50, got %f", *result.ExecutionPrice)
	}
}

func TestProcessOrderUseCase_Execute_OrderNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			return nil, errors.New("order not found")
		},
	}
	mockMarketData := &MockMarketDataClient{}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "nonexistent",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for nonexistent order")
	}

	if result != nil {
		t.Error("Expected nil result for nonexistent order")
	}

	if !contains(err.Error(), "order not found") {
		t.Errorf("Expected order not found error, got %v", err)
	}
}

func TestProcessOrderUseCase_Execute_OrderAlreadyExecuted(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			// Mark order as already executed
			order.MarkAsExecuted(149.50)
			return order, nil
		},
	}
	mockMarketData := &MockMarketDataClient{}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "order123",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for already executed order")
	}

	if result != nil {
		t.Error("Expected nil result for already executed order")
	}

	if !contains(err.Error(), "cannot be processed") {
		t.Errorf("Expected cannot be processed error, got %v", err)
	}
}

func TestProcessOrderUseCase_Execute_OrderCancelled(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			// Mark order as cancelled
			order.MarkAsCancelled()
			return order, nil
		},
	}
	mockMarketData := &MockMarketDataClient{}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "order123",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for cancelled order")
	}

	if result != nil {
		t.Error("Expected nil result for cancelled order")
	}

	if !contains(err.Error(), "cannot be processed") {
		t.Errorf("Expected cannot be processed error, got %v", err)
	}
}

func TestProcessOrderUseCase_Execute_MarketDataError(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			return order, nil
		},
		SaveFunc: func(ctx context.Context, order *domain.Order) error {
			return nil
		},
	}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			return 0, errors.New("market data service unavailable")
		},
	}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "order123",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for market data failure")
	}

	if result != nil {
		t.Error("Expected nil result for market data failure")
	}

	if !contains(err.Error(), "failed to get current market price") {
		t.Errorf("Expected market data error, got %v", err)
	}
}

func TestProcessOrderUseCase_Execute_RepositorySaveError(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			return order, nil
		},
		SaveFunc: func(ctx context.Context, order *domain.Order) error {
			return errors.New("database connection failed")
		},
	}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			return 149.50, nil
		},
	}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "order123",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository save failure")
	}

	if result != nil {
		t.Error("Expected nil result for repository save failure")
	}

	if !contains(err.Error(), "failed to save processed order") {
		t.Errorf("Expected save error, got %v", err)
	}
}

func TestProcessOrderUseCase_Execute_MarketOrder(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			// Market order (no price)
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)
			return order, nil
		},
		SaveFunc: func(ctx context.Context, order *domain.Order) error {
			return nil
		},
	}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			return 149.50, nil
		},
	}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "order123",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for market order, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result for market order")
	}

	if result.FinalStatus != "EXECUTED" {
		t.Errorf("Expected FinalStatus EXECUTED, got %s", result.FinalStatus)
	}

	// Market order should execute at current market price
	if result.ExecutionPrice == nil {
		t.Error("Expected ExecutionPrice to be set for market order")
	}

	if *result.ExecutionPrice != 149.50 {
		t.Errorf("Expected ExecutionPrice 149.50, got %f", *result.ExecutionPrice)
	}
}

func TestProcessOrderUseCase_Execute_LimitOrderPriceNotMet(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			// Limit order with price 145.00
			price := 145.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			return order, nil
		},
		SaveFunc: func(ctx context.Context, order *domain.Order) error {
			return nil
		},
	}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			// Market price is higher than limit price
			return 149.50, nil
		},
	}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "order123",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for limit order price not met")
	}

	if result != nil {
		t.Error("Expected nil result for limit order price not met")
	}

	if !contains(err.Error(), "limit price not met") {
		t.Errorf("Expected limit price not met error, got %v", err)
	}
}

func TestProcessOrderUseCase_Execute_SellLimitOrderPriceNotMet(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			// Sell limit order with price 155.00
			price := 155.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideSell, domain.OrderTypeLimit, 100.0, &price)
			return order, nil
		},
		SaveFunc: func(ctx context.Context, order *domain.Order) error {
			return nil
		},
	}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			// Market price is lower than sell limit price
			return 149.50, nil
		},
	}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "order123",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for sell limit order price not met")
	}

	if result != nil {
		t.Error("Expected nil result for sell limit order price not met")
	}

	if !contains(err.Error(), "limit price not met") {
		t.Errorf("Expected limit price not met error, got %v", err)
	}
}

func TestProcessOrderUseCase_Execute_EmptyOrderID(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{}

	useCase := NewProcessOrderUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	cmd := &ProcessOrderCommand{
		OrderID: "",
		Context: ProcessingContext{
			WorkerID:     "worker-1",
			ProcessingID: "proc-123",
			StartTime:    time.Now(),
			MaxRetries:   3,
			RetryAttempt: 1,
		},
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty order ID")
	}

	if result != nil {
		t.Error("Expected nil result for empty order ID")
	}
}
