package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

func TestGetOrderStatusUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order := domain.NewOrderFromRepository(
				orderID, // Use the requested order ID
				"user123",
				"AAPL",
				domain.OrderSideBuy,
				domain.OrderTypeLimit,
				100.0,
				&price,
				domain.OrderStatusPending,
				time.Now(),
				time.Now(),
				nil,
				nil,
				nil,
				nil,
			)
			return order, nil
		},
	}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			return 151.00, nil
		},
	}

	useCase := NewGetOrderStatusUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	orderID := "order123"
	userID := "user123"

	// Act
	result, err := useCase.Execute(ctx, orderID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.OrderID != orderID {
		t.Errorf("Expected OrderID %s, got %s", orderID, result.OrderID)
	}

	if result.Status != "PENDING" {
		t.Errorf("Expected status PENDING, got %s", result.Status)
	}

	if result.CurrentMarketPrice == nil {
		t.Error("Expected CurrentMarketPrice to be set")
	}
}

func TestGetOrderStatusUseCase_Execute_OrderNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			return nil, errors.New("order not found")
		},
	}
	mockMarketData := &MockMarketDataClient{}

	useCase := NewGetOrderStatusUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	orderID := "nonexistent"
	userID := "user123"

	// Act
	result, err := useCase.Execute(ctx, orderID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for nonexistent order")
	}

	if result != nil {
		t.Error("Expected nil result for nonexistent order")
	}
}

func TestGetOrderStatusUseCase_Execute_UnauthorizedAccess(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			return order, nil
		},
	}
	mockMarketData := &MockMarketDataClient{}

	useCase := NewGetOrderStatusUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	orderID := "order123"
	userID := "different_user" // Different user trying to access order

	// Act
	result, err := useCase.Execute(ctx, orderID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for unauthorized access")
	}

	if result != nil {
		t.Error("Expected nil result for unauthorized access")
	}

	if !contains(err.Error(), "not authorized") {
		t.Errorf("Expected authorization error, got %v", err)
	}
}

func TestGetOrderStatusUseCase_Execute_MarketDataError(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			return order, nil
		},
	}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			return 0, errors.New("market data service unavailable")
		},
	}

	useCase := NewGetOrderStatusUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	orderID := "order123"
	userID := "user123"

	// Act
	result, err := useCase.Execute(ctx, orderID, userID)

	// Assert
	// Should still return result even if market data fails
	if err != nil {
		t.Fatalf("Expected no error even with market data failure, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result even with market data failure")
	}

	// Current market price should be nil when market data fails
	if result.CurrentMarketPrice != nil {
		t.Error("Expected CurrentMarketPrice to be nil when market data fails")
	}
}

func TestGetOrderStatusUseCase_Execute_ExecutedOrder(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			// Simulate executed order
			order.MarkAsExecuted(149.50)
			return order, nil
		},
	}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			return 151.00, nil
		},
	}

	useCase := NewGetOrderStatusUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	orderID := "order123"
	userID := "user123"

	// Act
	result, err := useCase.Execute(ctx, orderID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Status != "EXECUTED" {
		t.Errorf("Expected status EXECUTED, got %s", result.Status)
	}

	if result.ExecutionPrice == nil {
		t.Error("Expected ExecutionPrice to be set for executed order")
	}

	if *result.ExecutionPrice != 149.50 {
		t.Errorf("Expected ExecutionPrice 149.50, got %f", *result.ExecutionPrice)
	}

	if result.ExecutedAt == nil {
		t.Error("Expected ExecutedAt to be set for executed order")
	}
}

func TestGetOrderStatusUseCase_Execute_EmptyOrderID(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{}

	useCase := NewGetOrderStatusUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	orderID := ""
	userID := "user123"

	// Act
	result, err := useCase.Execute(ctx, orderID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty order ID")
	}

	if result != nil {
		t.Error("Expected nil result for empty order ID")
	}
}

func TestGetOrderStatusUseCase_Execute_EmptyUserID(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{}

	useCase := NewGetOrderStatusUseCase(mockRepo, mockMarketData)

	ctx := context.Background()
	orderID := "order123"
	userID := ""

	// Act
	result, err := useCase.Execute(ctx, orderID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty user ID")
	}

	if result != nil {
		t.Error("Expected nil result for empty user ID")
	}
}
