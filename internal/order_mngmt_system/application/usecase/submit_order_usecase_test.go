package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"HubInvestments/internal/order_mngmt_system/application/command"
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/internal/order_mngmt_system/domain/service"
	"HubInvestments/internal/order_mngmt_system/infra/external"
)

// MockOrderRepository implements IOrderRepository for testing
type MockOrderRepository struct {
	SaveFunc     func(ctx context.Context, order *domain.Order) error
	FindByIDFunc func(ctx context.Context, orderID string) (*domain.Order, error)
}

func (m *MockOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, order)
	}
	return nil
}

func (m *MockOrderRepository) FindByID(ctx context.Context, orderID string) (*domain.Order, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, orderID)
	}
	return nil, errors.New("order not found")
}

func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	return nil, nil
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	return nil
}

func (m *MockOrderRepository) UpdateExecutionDetails(ctx context.Context, orderID string, executionPrice float64, executedAt time.Time) error {
	return nil
}

func (m *MockOrderRepository) FindByUserIDAndStatus(ctx context.Context, userID string, status domain.OrderStatus) ([]*domain.Order, error) {
	return nil, nil
}

func (m *MockOrderRepository) FindByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	return nil, nil
}

func (m *MockOrderRepository) FindOrderHistory(ctx context.Context, userID string, limit int, offset int) ([]*domain.Order, error) {
	return nil, nil
}

func (m *MockOrderRepository) FindOrdersBySymbol(ctx context.Context, symbol string) ([]*domain.Order, error) {
	return nil, nil
}

func (m *MockOrderRepository) FindOrdersByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*domain.Order, error) {
	return nil, nil
}

func (m *MockOrderRepository) CountOrdersByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockOrderRepository) Delete(ctx context.Context, orderID string) error {
	return nil
}

// MockMarketDataClient implements IMarketDataClient for testing
type MockMarketDataClient struct {
	ValidateSymbolFunc  func(ctx context.Context, symbol string) (bool, error)
	GetCurrentPriceFunc func(ctx context.Context, symbol string) (float64, error)
	GetAssetDetailsFunc func(ctx context.Context, symbol string) (*external.AssetDetails, error)
	GetTradingHoursFunc func(ctx context.Context, symbol string) (*external.TradingHours, error)
	IsMarketOpenFunc    func(ctx context.Context, symbol string) (bool, error)
}

func (m *MockMarketDataClient) ValidateSymbol(ctx context.Context, symbol string) (bool, error) {
	if m.ValidateSymbolFunc != nil {
		return m.ValidateSymbolFunc(ctx, symbol)
	}
	return true, nil
}

func (m *MockMarketDataClient) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	if m.GetCurrentPriceFunc != nil {
		return m.GetCurrentPriceFunc(ctx, symbol)
	}
	return 150.50, nil
}

func (m *MockMarketDataClient) GetAssetDetails(ctx context.Context, symbol string) (*external.AssetDetails, error) {
	if m.GetAssetDetailsFunc != nil {
		return m.GetAssetDetailsFunc(ctx, symbol)
	}
	return &external.AssetDetails{
		Symbol:       symbol,
		Name:         "Test Stock",
		Category:     external.AssetCategoryStock,
		LastQuote:    150.50,
		IsActive:     true,
		IsTradeable:  true,
		MaxOrderSize: 10000.0,
		PriceStep:    0.01,
		LastUpdated:  time.Now(),
	}, nil
}

func (m *MockMarketDataClient) GetTradingHours(ctx context.Context, symbol string) (*external.TradingHours, error) {
	if m.GetTradingHoursFunc != nil {
		return m.GetTradingHoursFunc(ctx, symbol)
	}
	return &external.TradingHours{
		Symbol:        symbol,
		MarketOpen:    time.Now().Add(-2 * time.Hour),
		MarketClose:   time.Now().Add(6 * time.Hour),
		IsOpen:        true,
		NextOpenTime:  time.Now().Add(18 * time.Hour),
		NextCloseTime: time.Now().Add(24 * time.Hour),
		Timezone:      "EST",
		ExtendedHours: false,
	}, nil
}

func (m *MockMarketDataClient) IsMarketOpen(ctx context.Context, symbol string) (bool, error) {
	if m.IsMarketOpenFunc != nil {
		return m.IsMarketOpenFunc(ctx, symbol)
	}
	return true, nil
}

func (m *MockMarketDataClient) Close() error {
	return nil
}

// MockIdempotencyService implements IIdempotencyService for testing
type MockIdempotencyService struct {
	GenerateKeyFunc         func(userID, symbol, orderType, orderSide string, quantity float64, price *float64) string
	CheckIdempotencyFunc    func(ctx context.Context, key, userID string) (*service.IdempotencyResult, error)
	StoreIdempotencyKeyFunc func(ctx context.Context, key, userID string, ttl time.Duration) error
	CompleteIdempotencyFunc func(ctx context.Context, key, userID, orderID, result string) error
	FailIdempotencyFunc     func(ctx context.Context, key, userID, errorMsg string) error
	CleanupExpiredKeysFunc  func(ctx context.Context) error
}

func (m *MockIdempotencyService) GenerateKey(userID, symbol, orderType, orderSide string, quantity float64, price *float64) string {
	if m.GenerateKeyFunc != nil {
		return m.GenerateKeyFunc(userID, symbol, orderType, orderSide, quantity, price)
	}
	return "test_idempotency_key"
}

func (m *MockIdempotencyService) CheckIdempotency(ctx context.Context, key, userID string) (*service.IdempotencyResult, error) {
	if m.CheckIdempotencyFunc != nil {
		return m.CheckIdempotencyFunc(ctx, key, userID)
	}
	return &service.IdempotencyResult{
		IsProcessed: false,
		Status:      service.IdempotencyStatusPending,
	}, nil
}

func (m *MockIdempotencyService) StoreIdempotencyKey(ctx context.Context, key, userID string, ttl time.Duration) error {
	if m.StoreIdempotencyKeyFunc != nil {
		return m.StoreIdempotencyKeyFunc(ctx, key, userID, ttl)
	}
	return nil
}

func (m *MockIdempotencyService) CompleteIdempotency(ctx context.Context, key, userID, orderID, result string) error {
	if m.CompleteIdempotencyFunc != nil {
		return m.CompleteIdempotencyFunc(ctx, key, userID, orderID, result)
	}
	return nil
}

func (m *MockIdempotencyService) FailIdempotency(ctx context.Context, key, userID, errorMsg string) error {
	if m.FailIdempotencyFunc != nil {
		return m.FailIdempotencyFunc(ctx, key, userID, errorMsg)
	}
	return nil
}

func (m *MockIdempotencyService) CleanupExpiredKeys(ctx context.Context) error {
	if m.CleanupExpiredKeysFunc != nil {
		return m.CleanupExpiredKeysFunc(ctx)
	}
	return nil
}

func TestSubmitOrderUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{}
	mockIdempotency := &MockIdempotencyService{}

	useCase := NewSubmitOrderUseCase(mockRepo, mockMarketData, mockIdempotency, nil)

	ctx := context.Background()
	price := 150.00
	cmd := &command.SubmitOrderCommand{
		UserID:    "user123",
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100.0,
		Price:     &price,
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

	if result.OrderID == "" {
		t.Error("Expected OrderID to be set")
	}

	if result.Status != "PENDING" {
		t.Errorf("Expected status PENDING, got %s", result.Status)
	}

	if result.MarketPriceAtSubmission == nil {
		t.Error("Expected MarketPriceAtSubmission to be set")
	}
}

func TestSubmitOrderUseCase_Execute_InvalidCommand(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{}
	mockIdempotency := &MockIdempotencyService{}

	useCase := NewSubmitOrderUseCase(mockRepo, mockMarketData, mockIdempotency, nil)

	ctx := context.Background()
	cmd := &command.SubmitOrderCommand{
		// Missing required fields
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid command")
	}

	if result != nil {
		t.Error("Expected nil result for invalid command")
	}
}

func TestSubmitOrderUseCase_Execute_IdempotencyAlreadyProcessed(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{}
	mockIdempotency := &MockIdempotencyService{
		CheckIdempotencyFunc: func(ctx context.Context, key, userID string) (*service.IdempotencyResult, error) {
			return &service.IdempotencyResult{
				IsProcessed: true,
				Status:      service.IdempotencyStatusCompleted,
				OrderID:     "existing-order-123",
			}, nil
		},
	}

	useCase := NewSubmitOrderUseCase(mockRepo, mockMarketData, mockIdempotency, nil)

	ctx := context.Background()
	price := 150.00
	cmd := &command.SubmitOrderCommand{
		UserID:    "user123",
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100.0,
		Price:     &price,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for idempotent request, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result for idempotent request")
	}

	if result.OrderID != "existing-order-123" {
		t.Errorf("Expected existing OrderID, got %s", result.OrderID)
	}

	if result.Message != "Order already submitted (idempotent request)" {
		t.Errorf("Expected idempotent message, got %s", result.Message)
	}
}

func TestSubmitOrderUseCase_Execute_MarketDataValidationFailure(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{
		ValidateSymbolFunc: func(ctx context.Context, symbol string) (bool, error) {
			return false, nil
		},
	}
	mockIdempotency := &MockIdempotencyService{}

	useCase := NewSubmitOrderUseCase(mockRepo, mockMarketData, mockIdempotency, nil)

	ctx := context.Background()
	price := 150.00
	cmd := &command.SubmitOrderCommand{
		UserID:    "user123",
		Symbol:    "INVALID",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100.0,
		Price:     &price,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid symbol")
	}

	if result != nil {
		t.Error("Expected nil result for invalid symbol")
	}

	if !contains(err.Error(), "symbol validation failed") {
		t.Errorf("Expected symbol validation error, got %v", err)
	}
}

func TestSubmitOrderUseCase_Execute_MarketClosed(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{
		IsMarketOpenFunc: func(ctx context.Context, symbol string) (bool, error) {
			return false, nil
		},
	}
	mockIdempotency := &MockIdempotencyService{}

	useCase := NewSubmitOrderUseCase(mockRepo, mockMarketData, mockIdempotency, nil)

	ctx := context.Background()
	price := 150.00
	cmd := &command.SubmitOrderCommand{
		UserID:    "user123",
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100.0,
		Price:     &price,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for closed market")
	}

	if result != nil {
		t.Error("Expected nil result for closed market")
	}

	if !contains(err.Error(), "trading hours validation failed") {
		t.Errorf("Expected trading hours error, got %v", err)
	}
}

func TestSubmitOrderUseCase_Execute_PriceValidationFailure(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{
		GetCurrentPriceFunc: func(ctx context.Context, symbol string) (float64, error) {
			return 150.00, nil
		},
	}
	mockIdempotency := &MockIdempotencyService{}

	useCase := NewSubmitOrderUseCase(mockRepo, mockMarketData, mockIdempotency, nil)

	ctx := context.Background()
	// Price too far from market price (should fail validation)
	price := 200.00
	cmd := &command.SubmitOrderCommand{
		UserID:    "user123",
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100.0,
		Price:     &price,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid price")
	}

	if result != nil {
		t.Error("Expected nil result for invalid price")
	}

	if !contains(err.Error(), "price validation failed") {
		t.Errorf("Expected price validation error, got %v", err)
	}
}

func TestSubmitOrderUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{
		SaveFunc: func(ctx context.Context, order *domain.Order) error {
			return errors.New("database connection failed")
		},
	}
	mockMarketData := &MockMarketDataClient{}
	mockIdempotency := &MockIdempotencyService{}

	useCase := NewSubmitOrderUseCase(mockRepo, mockMarketData, mockIdempotency, nil)

	ctx := context.Background()
	price := 150.00
	cmd := &command.SubmitOrderCommand{
		UserID:    "user123",
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100.0,
		Price:     &price,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository failure")
	}

	if result != nil {
		t.Error("Expected nil result for repository failure")
	}

	if !contains(err.Error(), "failed to save order") {
		t.Errorf("Expected save order error, got %v", err)
	}
}

func TestSubmitOrderUseCase_Execute_MarketOrder(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{}
	mockIdempotency := &MockIdempotencyService{}

	useCase := NewSubmitOrderUseCase(mockRepo, mockMarketData, mockIdempotency, nil)

	ctx := context.Background()
	cmd := &command.SubmitOrderCommand{
		UserID:    "user123",
		Symbol:    "AAPL",
		OrderType: "MARKET",
		OrderSide: "BUY",
		Quantity:  100.0,
		// No price for market order
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

	if result.EstimatedExecutionPrice == nil {
		t.Error("Expected EstimatedExecutionPrice to be set for market order")
	}
}

func TestSubmitOrderUseCase_Execute_IdempotencyServiceError(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	mockMarketData := &MockMarketDataClient{}
	mockIdempotency := &MockIdempotencyService{
		CheckIdempotencyFunc: func(ctx context.Context, key, userID string) (*service.IdempotencyResult, error) {
			return nil, errors.New("idempotency service unavailable")
		},
	}

	useCase := NewSubmitOrderUseCase(mockRepo, mockMarketData, mockIdempotency, nil)

	ctx := context.Background()
	price := 150.00
	cmd := &command.SubmitOrderCommand{
		UserID:    "user123",
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100.0,
		Price:     &price,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	if err == nil {
		t.Fatal("Expected error for idempotency service failure")
	}

	if result != nil {
		t.Error("Expected nil result for idempotency service failure")
	}

	if !contains(err.Error(), "idempotency check failed") {
		t.Errorf("Expected idempotency check error, got %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
