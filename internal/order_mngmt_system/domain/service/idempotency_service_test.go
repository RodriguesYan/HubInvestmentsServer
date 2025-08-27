package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// MockIdempotencyRepository provides a mock implementation for testing
type MockIdempotencyRepository struct {
	keys map[string]*IdempotencyKey
}

// NewMockIdempotencyRepository creates a new mock repository for testing
func NewMockIdempotencyRepository() IIdempotencyRepository {
	return &MockIdempotencyRepository{
		keys: make(map[string]*IdempotencyKey),
	}
}

func (m *MockIdempotencyRepository) Store(ctx context.Context, key *IdempotencyKey) error {
	if key == nil {
		return fmt.Errorf("idempotency key cannot be nil")
	}

	mockKey := fmt.Sprintf("%s:%s", key.UserID, key.Key)
	m.keys[mockKey] = key
	return nil
}

func (m *MockIdempotencyRepository) Get(ctx context.Context, key, userID string) (*IdempotencyKey, error) {
	mockKey := fmt.Sprintf("%s:%s", userID, key)
	idempotencyKey, exists := m.keys[mockKey]
	if !exists {
		return nil, fmt.Errorf("idempotency key not found")
	}

	// Check expiration
	if time.Now().After(idempotencyKey.ExpiresAt) {
		delete(m.keys, mockKey)
		return nil, fmt.Errorf("idempotency key expired")
	}

	return idempotencyKey, nil
}

func (m *MockIdempotencyRepository) Update(ctx context.Context, key *IdempotencyKey) error {
	if key == nil {
		return fmt.Errorf("idempotency key cannot be nil")
	}

	mockKey := fmt.Sprintf("%s:%s", key.UserID, key.Key)
	if _, exists := m.keys[mockKey]; !exists {
		return fmt.Errorf("idempotency key not found for update")
	}

	m.keys[mockKey] = key
	return nil
}

func (m *MockIdempotencyRepository) Delete(ctx context.Context, key, userID string) error {
	mockKey := fmt.Sprintf("%s:%s", userID, key)
	delete(m.keys, mockKey)
	return nil
}

func (m *MockIdempotencyRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for key, idempotencyKey := range m.keys {
		if now.After(idempotencyKey.ExpiresAt) {
			delete(m.keys, key)
		}
	}
	return nil
}

func TestIdempotencyService_GenerateKey(t *testing.T) {
	repo := NewMockIdempotencyRepository()
	service := NewIdempotencyService(repo)

	userID := "user123"
	symbol := "AAPL"
	orderType := "LIMIT"
	orderSide := "BUY"
	quantity := 100.0
	price := 150.50

	// Generate key twice with same parameters
	key1 := service.GenerateKey(userID, symbol, orderType, orderSide, quantity, &price)
	key2 := service.GenerateKey(userID, symbol, orderType, orderSide, quantity, &price)

	// Keys should be identical for same parameters
	if key1 != key2 {
		t.Errorf("Expected identical keys for same parameters, got %s and %s", key1, key2)
	}

	// Different parameters should generate different keys
	differentPrice := 151.00
	key3 := service.GenerateKey(userID, symbol, orderType, orderSide, quantity, &differentPrice)

	if key1 == key3 {
		t.Errorf("Expected different keys for different parameters, but got identical keys")
	}
}

func TestIdempotencyService_CheckIdempotency_NewOperation(t *testing.T) {
	repo := NewMockIdempotencyRepository()
	service := NewIdempotencyService(repo)

	ctx := context.Background()
	key := "test_key"
	userID := "user123"

	result, err := service.CheckIdempotency(ctx, key, userID)
	if err != nil {
		t.Fatalf("Expected no error for new operation, got %v", err)
	}

	if result.IsProcessed {
		t.Error("Expected IsProcessed to be false for new operation")
	}

	if result.Status != IdempotencyStatusPending {
		t.Errorf("Expected status PENDING for new operation, got %s", result.Status)
	}
}

func TestIdempotencyService_StoreAndCheckIdempotency(t *testing.T) {
	repo := NewMockIdempotencyRepository()
	service := NewIdempotencyService(repo)

	ctx := context.Background()
	key := "test_key"
	userID := "user123"
	ttl := 1 * time.Hour

	// Store idempotency key
	err := service.StoreIdempotencyKey(ctx, key, userID, ttl)
	if err != nil {
		t.Fatalf("Failed to store idempotency key: %v", err)
	}

	// Check idempotency - should now be processed
	result, err := service.CheckIdempotency(ctx, key, userID)
	if err != nil {
		t.Fatalf("Failed to check idempotency: %v", err)
	}

	if !result.IsProcessed {
		t.Error("Expected IsProcessed to be true after storing key")
	}

	if result.Status != IdempotencyStatusPending {
		t.Errorf("Expected status PENDING after storing, got %s", result.Status)
	}
}

func TestIdempotencyService_CompleteIdempotency(t *testing.T) {
	repo := NewMockIdempotencyRepository()
	service := NewIdempotencyService(repo)

	ctx := context.Background()
	key := "test_key"
	userID := "user123"
	orderID := "order123"
	result := "Order submitted successfully"

	// Store and complete idempotency
	_ = service.StoreIdempotencyKey(ctx, key, userID, 1*time.Hour)
	err := service.CompleteIdempotency(ctx, key, userID, orderID, result)
	if err != nil {
		t.Fatalf("Failed to complete idempotency: %v", err)
	}

	// Check result
	idempotencyResult, err := service.CheckIdempotency(ctx, key, userID)
	if err != nil {
		t.Fatalf("Failed to check completed idempotency: %v", err)
	}

	if idempotencyResult.Status != IdempotencyStatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s", idempotencyResult.Status)
	}

	if idempotencyResult.OrderID != orderID {
		t.Errorf("Expected OrderID %s, got %s", orderID, idempotencyResult.OrderID)
	}

	if idempotencyResult.Result != result {
		t.Errorf("Expected result %s, got %s", result, idempotencyResult.Result)
	}
}

func TestIdempotencyService_FailIdempotency(t *testing.T) {
	repo := NewMockIdempotencyRepository()
	service := NewIdempotencyService(repo)

	ctx := context.Background()
	key := "test_key"
	userID := "user123"
	errorMsg := "Order validation failed"

	// Store and fail idempotency
	_ = service.StoreIdempotencyKey(ctx, key, userID, 1*time.Hour)
	err := service.FailIdempotency(ctx, key, userID, errorMsg)
	if err != nil {
		t.Fatalf("Failed to fail idempotency: %v", err)
	}

	// Check result
	idempotencyResult, err := service.CheckIdempotency(ctx, key, userID)
	if err != nil {
		t.Fatalf("Failed to check failed idempotency: %v", err)
	}

	if idempotencyResult.Status != IdempotencyStatusFailed {
		t.Errorf("Expected status FAILED, got %s", idempotencyResult.Status)
	}

	if idempotencyResult.Result != errorMsg {
		t.Errorf("Expected result %s, got %s", errorMsg, idempotencyResult.Result)
	}
}

func TestIdempotencyService_OrderKeyGeneration(t *testing.T) {
	repo := NewMockIdempotencyRepository()
	service := NewIdempotencyService(repo)

	// Create a test order
	price := 150.50
	order, err := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
	if err != nil {
		t.Fatalf("Failed to create test order: %v", err)
	}

	// Generate key for order
	key := service.GenerateKey(order.UserID(), order.Symbol(), order.OrderType().String(), order.OrderSide().String(), order.Quantity(), order.Price())

	if key == "" {
		t.Error("Expected non-empty key for order")
	}

	// Same order should generate same key
	key2 := service.GenerateKey(order.UserID(), order.Symbol(), order.OrderType().String(), order.OrderSide().String(), order.Quantity(), order.Price())

	if key != key2 {
		t.Errorf("Expected same key for identical order parameters, got %s and %s", key, key2)
	}
}
