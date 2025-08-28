package usecase

import (
	"context"
	"errors"
	"testing"

	"HubInvestments/internal/order_mngmt_system/application/command"
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

func TestCancelOrderUseCase_Execute_Success(t *testing.T) {
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

	useCase := NewCancelOrderUseCase(mockRepo)

	ctx := context.Background()
	cmd := &command.CancelOrderCommand{
		OrderID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:  "user123",
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

	if result.Status != "CANCELLED" {
		t.Errorf("Expected status CANCELLED, got %s", result.Status)
	}

	if result.Message == "" {
		t.Error("Expected Message to be set")
	}
}

func TestCancelOrderUseCase_Execute_OrderNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			return nil, errors.New("order not found")
		},
	}

	useCase := NewCancelOrderUseCase(mockRepo)

	ctx := context.Background()
	cmd := &command.CancelOrderCommand{
		OrderID: "550e8400-e29b-41d4-a716-446655440001",
		UserID:  "user123",
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

func TestCancelOrderUseCase_Execute_UnauthorizedAccess(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			return order, nil
		},
	}

	useCase := NewCancelOrderUseCase(mockRepo)

	ctx := context.Background()
	cmd := &command.CancelOrderCommand{
		OrderID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:  "different_user", // Different user trying to cancel order
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

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

func TestCancelOrderUseCase_Execute_OrderAlreadyExecuted(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			// Mark order as executed
			order.MarkAsExecuted(149.50)
			return order, nil
		},
	}

	useCase := NewCancelOrderUseCase(mockRepo)

	ctx := context.Background()
	cmd := &command.CancelOrderCommand{
		OrderID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:  "user123",
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

	if !contains(err.Error(), "cannot be cancelled") {
		t.Errorf("Expected cannot be cancelled error, got %v", err)
	}
}

func TestCancelOrderUseCase_Execute_OrderAlreadyCancelled(t *testing.T) {
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

	useCase := NewCancelOrderUseCase(mockRepo)

	ctx := context.Background()
	cmd := &command.CancelOrderCommand{
		OrderID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:  "user123",
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for already cancelled order")
	}

	if result != nil {
		t.Error("Expected nil result for already cancelled order")
	}

	if !contains(err.Error(), "cannot be cancelled") {
		t.Errorf("Expected cannot be cancelled error, got %v", err)
	}
}

func TestCancelOrderUseCase_Execute_RepositorySaveError(t *testing.T) {
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

	useCase := NewCancelOrderUseCase(mockRepo)

	ctx := context.Background()
	cmd := &command.CancelOrderCommand{
		OrderID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:  "user123",
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

	if !contains(err.Error(), "failed to save cancelled order") {
		t.Errorf("Expected save error, got %v", err)
	}
}

func TestCancelOrderUseCase_Execute_EmptyOrderID(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	useCase := NewCancelOrderUseCase(mockRepo)

	ctx := context.Background()
	cmd := &command.CancelOrderCommand{
		OrderID: "",
		UserID:  "user123",
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

func TestCancelOrderUseCase_Execute_EmptyUserID(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	useCase := NewCancelOrderUseCase(mockRepo)

	ctx := context.Background()
	cmd := &command.CancelOrderCommand{
		OrderID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:  "",
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty user ID")
	}

	if result != nil {
		t.Error("Expected nil result for empty user ID")
	}
}

func TestCancelOrderUseCase_Execute_ProcessingOrder(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		FindByIDFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			price := 150.00
			order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
			// Mark order as processing
			order.MarkAsProcessing()
			return order, nil
		},
	}

	useCase := NewCancelOrderUseCase(mockRepo)

	ctx := context.Background()
	cmd := &command.CancelOrderCommand{
		OrderID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:  "user123",
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for processing order")
	}

	if result != nil {
		t.Error("Expected nil result for processing order")
	}

	if !contains(err.Error(), "cannot be cancelled") {
		t.Errorf("Expected cannot be cancelled error, got %v", err)
	}
}
