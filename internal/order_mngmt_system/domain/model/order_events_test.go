package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewOrderEvent(t *testing.T) {
	// Arrange
	eventType := "TestEvent"
	orderID := "order-123"
	userID := "user-456"

	// Act
	event := NewOrderEvent(eventType, orderID, userID)

	// Assert
	if event.eventType != eventType {
		t.Errorf("Expected event type %s, got %s", eventType, event.eventType)
	}

	if event.orderID != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.orderID)
	}

	if event.userID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.userID)
	}

	if event.aggregateID != orderID {
		t.Errorf("Expected aggregate ID %s, got %s", orderID, event.aggregateID)
	}

	// Check that eventID is a valid UUID
	if _, err := uuid.Parse(event.eventID); err != nil {
		t.Errorf("Expected valid UUID for event ID, got %s", event.eventID)
	}

	// Check that occurredAt is recent (within last minute)
	timeDiff := time.Since(event.occurredAt)
	if timeDiff > time.Minute {
		t.Errorf("Expected recent timestamp, got %v", event.occurredAt)
	}
}

func TestOrderEvent_Methods(t *testing.T) {
	// Arrange
	eventType := "TestEvent"
	orderID := "order-123"
	userID := "user-456"
	event := NewOrderEvent(eventType, orderID, userID)

	// Act & Assert
	if event.EventID() != event.eventID {
		t.Errorf("EventID() method returned unexpected value")
	}

	if event.EventType() != eventType {
		t.Errorf("EventType() method returned unexpected value")
	}

	if event.AggregateID() != orderID {
		t.Errorf("AggregateID() method returned unexpected value")
	}

	if event.OrderID() != orderID {
		t.Errorf("OrderID() method returned unexpected value")
	}

	if event.UserID() != userID {
		t.Errorf("UserID() method returned unexpected value")
	}

	// Check OccurredAt returns the same timestamp
	if !event.OccurredAt().Equal(event.occurredAt) {
		t.Errorf("OccurredAt() method returned unexpected value")
	}
}

func TestNewOrderSubmittedEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	symbol := "AAPL"
	orderSide := OrderSideBuy
	orderType := OrderTypeLimit
	quantity := 100.0
	price := 150.50

	// Act
	event := NewOrderSubmittedEvent(orderID, userID, symbol, orderSide, orderType, quantity, &price)

	// Assert
	if event.EventType() != "OrderSubmitted" {
		t.Errorf("Expected event type 'OrderSubmitted', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if event.Symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, event.Symbol)
	}

	if event.OrderSide != orderSide {
		t.Errorf("Expected order side %v, got %v", orderSide, event.OrderSide)
	}

	if event.OrderType != orderType {
		t.Errorf("Expected order type %v, got %v", orderType, event.OrderType)
	}

	if event.Quantity != quantity {
		t.Errorf("Expected quantity %f, got %f", quantity, event.Quantity)
	}

	if event.Price == nil || *event.Price != price {
		t.Errorf("Expected price %f, got %v", price, event.Price)
	}
}

func TestNewOrderSubmittedEvent_NilPrice(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	symbol := "AAPL"
	orderSide := OrderSideBuy
	orderType := OrderTypeMarket
	quantity := 100.0

	// Act
	event := NewOrderSubmittedEvent(orderID, userID, symbol, orderSide, orderType, quantity, nil)

	// Assert
	if event.Price != nil {
		t.Errorf("Expected nil price for market order, got %v", event.Price)
	}
}

func TestNewOrderProcessingStartedEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	marketPrice := 150.75
	marketTimestamp := time.Now().Add(-5 * time.Minute)

	// Act
	event := NewOrderProcessingStartedEvent(orderID, userID, &marketPrice, &marketTimestamp)

	// Assert
	if event.EventType() != "OrderProcessingStarted" {
		t.Errorf("Expected event type 'OrderProcessingStarted', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if event.MarketPrice == nil || *event.MarketPrice != marketPrice {
		t.Errorf("Expected market price %f, got %v", marketPrice, event.MarketPrice)
	}

	if event.MarketTimestamp == nil || !event.MarketTimestamp.Equal(marketTimestamp) {
		t.Errorf("Expected market timestamp %v, got %v", marketTimestamp, event.MarketTimestamp)
	}
}

func TestNewOrderProcessingStartedEvent_NilValues(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"

	// Act
	event := NewOrderProcessingStartedEvent(orderID, userID, nil, nil)

	// Assert
	if event.MarketPrice != nil {
		t.Errorf("Expected nil market price, got %v", event.MarketPrice)
	}

	if event.MarketTimestamp != nil {
		t.Errorf("Expected nil market timestamp, got %v", event.MarketTimestamp)
	}
}

func TestNewOrderExecutedEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	executionPrice := 150.25
	totalValue := 15025.0
	executedAt := time.Now()

	// Act
	event := NewOrderExecutedEvent(orderID, userID, executionPrice, totalValue, executedAt)

	// Assert
	if event.EventType() != "OrderExecuted" {
		t.Errorf("Expected event type 'OrderExecuted', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if event.ExecutionPrice != executionPrice {
		t.Errorf("Expected execution price %f, got %f", executionPrice, event.ExecutionPrice)
	}

	if event.TotalValue != totalValue {
		t.Errorf("Expected total value %f, got %f", totalValue, event.TotalValue)
	}

	if !event.ExecutedAt.Equal(executedAt) {
		t.Errorf("Expected executed at %v, got %v", executedAt, event.ExecutedAt)
	}
}

func TestNewOrderFailedEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	failureReason := "Insufficient funds"
	failedAt := time.Now()

	// Act
	event := NewOrderFailedEvent(orderID, userID, failureReason, failedAt)

	// Assert
	if event.EventType() != "OrderFailed" {
		t.Errorf("Expected event type 'OrderFailed', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if event.FailureReason != failureReason {
		t.Errorf("Expected failure reason %s, got %s", failureReason, event.FailureReason)
	}

	if !event.FailedAt.Equal(failedAt) {
		t.Errorf("Expected failed at %v, got %v", failedAt, event.FailedAt)
	}
}

func TestNewOrderCancelledEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	cancelReason := "User requested cancellation"
	cancelledBy := "user-456"
	cancelledAt := time.Now()

	// Act
	event := NewOrderCancelledEvent(orderID, userID, cancelReason, cancelledBy, cancelledAt)

	// Assert
	if event.EventType() != "OrderCancelled" {
		t.Errorf("Expected event type 'OrderCancelled', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if event.CancelReason != cancelReason {
		t.Errorf("Expected cancel reason %s, got %s", cancelReason, event.CancelReason)
	}

	if event.CancelledBy != cancelledBy {
		t.Errorf("Expected cancelled by %s, got %s", cancelledBy, event.CancelledBy)
	}

	if !event.CancelledAt.Equal(cancelledAt) {
		t.Errorf("Expected cancelled at %v, got %v", cancelledAt, event.CancelledAt)
	}
}

func TestNewOrderCancelledEvent_AdminCancellation(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	cancelReason := "Risk management"
	cancelledBy := "admin-789"
	cancelledAt := time.Now()

	// Act
	event := NewOrderCancelledEvent(orderID, userID, cancelReason, cancelledBy, cancelledAt)

	// Assert
	if event.CancelledBy != cancelledBy {
		t.Errorf("Expected cancelled by admin %s, got %s", cancelledBy, event.CancelledBy)
	}

	// Verify that admin can cancel user's order
	if event.UserID() == event.CancelledBy {
		t.Errorf("Test setup error: admin should be different from user")
	}
}

func TestNewMarketDataReceivedEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	symbol := "AAPL"
	dataSource := "market_data_service"
	marketPrice := 150.75
	marketDataTimestamp := time.Now().Add(-1 * time.Minute)

	// Act
	event := NewMarketDataReceivedEvent(orderID, userID, symbol, dataSource, marketPrice, marketDataTimestamp)

	// Assert
	if event.EventType() != "MarketDataReceived" {
		t.Errorf("Expected event type 'MarketDataReceived', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if event.Symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, event.Symbol)
	}

	if event.DataSource != dataSource {
		t.Errorf("Expected data source %s, got %s", dataSource, event.DataSource)
	}

	if event.MarketPrice != marketPrice {
		t.Errorf("Expected market price %f, got %f", marketPrice, event.MarketPrice)
	}

	if !event.MarketDataTimestamp.Equal(marketDataTimestamp) {
		t.Errorf("Expected market data timestamp %v, got %v", marketDataTimestamp, event.MarketDataTimestamp)
	}
}

func TestNewOrderStatusChangedEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	fromStatus := OrderStatusPending
	toStatus := OrderStatusProcessing
	reason := "Order validation completed"
	changedAt := time.Now()

	// Act
	event := NewOrderStatusChangedEvent(orderID, userID, fromStatus, toStatus, reason, changedAt)

	// Assert
	if event.EventType() != "OrderStatusChanged" {
		t.Errorf("Expected event type 'OrderStatusChanged', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if event.FromStatus != fromStatus {
		t.Errorf("Expected from status %v, got %v", fromStatus, event.FromStatus)
	}

	if event.ToStatus != toStatus {
		t.Errorf("Expected to status %v, got %v", toStatus, event.ToStatus)
	}

	if event.Reason != reason {
		t.Errorf("Expected reason %s, got %s", reason, event.Reason)
	}

	if !event.ChangedAt.Equal(changedAt) {
		t.Errorf("Expected changed at %v, got %v", changedAt, event.ChangedAt)
	}
}

func TestNewRiskCheckPerformedEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	riskCheckType := "BALANCE_CHECK"
	result := "PASSED"
	riskScore := 0.25
	checkedAt := time.Now()

	// Act
	event := NewRiskCheckPerformedEvent(orderID, userID, riskCheckType, result, riskScore, checkedAt)

	// Assert
	if event.EventType() != "RiskCheckPerformed" {
		t.Errorf("Expected event type 'RiskCheckPerformed', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if event.RiskCheckType != riskCheckType {
		t.Errorf("Expected risk check type %s, got %s", riskCheckType, event.RiskCheckType)
	}

	if event.RiskCheckResult != result {
		t.Errorf("Expected risk check result %s, got %s", result, event.RiskCheckResult)
	}

	if event.RiskScore != riskScore {
		t.Errorf("Expected risk score %f, got %f", riskScore, event.RiskScore)
	}

	if !event.CheckedAt.Equal(checkedAt) {
		t.Errorf("Expected checked at %v, got %v", checkedAt, event.CheckedAt)
	}
}

func TestNewRiskCheckPerformedEvent_FailedCheck(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	riskCheckType := "POSITION_LIMIT_CHECK"
	result := "FAILED"
	riskScore := 0.95
	checkedAt := time.Now()

	// Act
	event := NewRiskCheckPerformedEvent(orderID, userID, riskCheckType, result, riskScore, checkedAt)

	// Assert
	if event.RiskCheckResult != result {
		t.Errorf("Expected failed risk check result %s, got %s", result, event.RiskCheckResult)
	}

	if event.RiskScore != riskScore {
		t.Errorf("Expected high risk score %f, got %f", riskScore, event.RiskScore)
	}
}

func TestNewOrderValidationFailedEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	validationErrors := []string{
		"Invalid symbol",
		"Price must be positive",
		"Quantity below minimum",
	}
	validatedAt := time.Now()

	// Act
	event := NewOrderValidationFailedEvent(orderID, userID, validationErrors, validatedAt)

	// Assert
	if event.EventType() != "OrderValidationFailed" {
		t.Errorf("Expected event type 'OrderValidationFailed', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if len(event.ValidationErrors) != len(validationErrors) {
		t.Errorf("Expected %d validation errors, got %d", len(validationErrors), len(event.ValidationErrors))
	}

	for i, expectedError := range validationErrors {
		if i >= len(event.ValidationErrors) || event.ValidationErrors[i] != expectedError {
			t.Errorf("Expected validation error %s at index %d, got %s", expectedError, i, event.ValidationErrors[i])
		}
	}

	if !event.ValidatedAt.Equal(validatedAt) {
		t.Errorf("Expected validated at %v, got %v", validatedAt, event.ValidatedAt)
	}
}

func TestNewOrderValidationFailedEvent_EmptyErrors(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	validationErrors := []string{}
	validatedAt := time.Now()

	// Act
	event := NewOrderValidationFailedEvent(orderID, userID, validationErrors, validatedAt)

	// Assert
	if len(event.ValidationErrors) != 0 {
		t.Errorf("Expected empty validation errors, got %v", event.ValidationErrors)
	}
}

func TestNewPositionValidationFailedEvent(t *testing.T) {
	// Arrange
	orderID := "order-123"
	userID := "user-456"
	symbol := "AAPL"
	requestedQuantity := 100.0
	availableQuantity := 50.0
	validationError := "Insufficient position: requested 100, available 50"

	// Act
	event := NewPositionValidationFailedEvent(orderID, userID, symbol, requestedQuantity, availableQuantity, validationError)

	// Assert
	if event.EventType() != "PositionValidationFailed" {
		t.Errorf("Expected event type 'PositionValidationFailed', got %s", event.EventType())
	}

	if event.OrderID() != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, event.OrderID())
	}

	if event.UserID() != userID {
		t.Errorf("Expected user ID %s, got %s", userID, event.UserID())
	}

	if event.Symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, event.Symbol)
	}

	if event.RequestedQuantity != requestedQuantity {
		t.Errorf("Expected requested quantity %f, got %f", requestedQuantity, event.RequestedQuantity)
	}

	if event.AvailableQuantity != availableQuantity {
		t.Errorf("Expected available quantity %f, got %f", availableQuantity, event.AvailableQuantity)
	}

	if event.ValidationError != validationError {
		t.Errorf("Expected validation error %s, got %s", validationError, event.ValidationError)
	}
}

func TestDomainEventInterface(t *testing.T) {
	// Test that all event types implement DomainEvent interface
	orderID := "order-123"
	userID := "user-456"

	events := []DomainEvent{
		NewOrderSubmittedEvent(orderID, userID, "AAPL", OrderSideBuy, OrderTypeLimit, 100.0, nil),
		NewOrderProcessingStartedEvent(orderID, userID, nil, nil),
		NewOrderExecutedEventWithDetails(orderID, userID, "AAPL", OrderSideBuy, OrderTypeLimit, 100.0, 150.0, 15000.0, time.Now(), nil, nil),
		NewOrderFailedEvent(orderID, userID, "Test failure", time.Now()),
		NewOrderCancelledEvent(orderID, userID, "Test cancel", userID, time.Now()),
		NewMarketDataReceivedEvent(orderID, userID, "AAPL", "test", 150.0, time.Now()),
		NewOrderStatusChangedEvent(orderID, userID, OrderStatusPending, OrderStatusProcessing, "test", time.Now()),
		NewRiskCheckPerformedEvent(orderID, userID, "test", "PASSED", 0.1, time.Now()),
		NewOrderValidationFailedEvent(orderID, userID, []string{"test"}, time.Now()),
		NewPositionValidationFailedEvent(orderID, userID, "AAPL", 100.0, 50.0, "test"),
	}

	for i, event := range events {
		// Test that all required methods are implemented
		if event.EventID() == "" {
			t.Errorf("Event %d: EventID() returned empty string", i)
		}

		if event.EventType() == "" {
			t.Errorf("Event %d: EventType() returned empty string", i)
		}

		if event.AggregateID() == "" {
			t.Errorf("Event %d: AggregateID() returned empty string", i)
		}

		if event.OccurredAt().IsZero() {
			t.Errorf("Event %d: OccurredAt() returned zero time", i)
		}

		// Test that EventID is a valid UUID
		if _, err := uuid.Parse(event.EventID()); err != nil {
			t.Errorf("Event %d: EventID() is not a valid UUID: %s", i, event.EventID())
		}
	}
}

func TestEventTimestamps(t *testing.T) {
	// Ensure events are created with proper timestamps
	beforeCreation := time.Now()
	time.Sleep(1 * time.Millisecond) // Small delay to ensure timestamp difference

	event := NewOrderSubmittedEvent("order-123", "user-456", "AAPL", OrderSideBuy, OrderTypeLimit, 100.0, nil)

	time.Sleep(1 * time.Millisecond) // Small delay to ensure timestamp difference
	afterCreation := time.Now()

	if event.OccurredAt().Before(beforeCreation) {
		t.Errorf("Event timestamp %v is before creation started %v", event.OccurredAt(), beforeCreation)
	}

	if event.OccurredAt().After(afterCreation) {
		t.Errorf("Event timestamp %v is after creation finished %v", event.OccurredAt(), afterCreation)
	}
}

func TestEventUniqueIDs(t *testing.T) {
	// Ensure each event gets a unique ID
	const numEvents = 100
	eventIDs := make(map[string]bool)

	for i := 0; i < numEvents; i++ {
		event := NewOrderSubmittedEvent("order-123", "user-456", "AAPL", OrderSideBuy, OrderTypeLimit, 100.0, nil)

		if eventIDs[event.EventID()] {
			t.Errorf("Duplicate event ID found: %s", event.EventID())
		}

		eventIDs[event.EventID()] = true
	}

	if len(eventIDs) != numEvents {
		t.Errorf("Expected %d unique event IDs, got %d", numEvents, len(eventIDs))
	}
}
