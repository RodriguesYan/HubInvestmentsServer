package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewPositionEvent(t *testing.T) {
	eventType := "TestEvent"
	positionID := uuid.New().String()
	userID := uuid.New().String()
	symbol := "AAPL"

	event := NewPositionEvent(eventType, positionID, userID, symbol)

	if event.EventType() != eventType {
		t.Errorf("EventType() = %v, want %v", event.EventType(), eventType)
	}

	if event.AggregateID() != positionID {
		t.Errorf("AggregateID() = %v, want %v", event.AggregateID(), positionID)
	}

	if event.PositionID() != positionID {
		t.Errorf("PositionID() = %v, want %v", event.PositionID(), positionID)
	}

	if event.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", event.UserID(), userID)
	}

	if event.Symbol() != symbol {
		t.Errorf("Symbol() = %v, want %v", event.Symbol(), symbol)
	}

	// Check that EventID is generated (UUID format)
	if event.EventID() == "" {
		t.Error("EventID() should not be empty")
	}

	// Check that OccurredAt is set
	if event.OccurredAt().IsZero() {
		t.Error("OccurredAt() should not be zero")
	}

	// Verify that OccurredAt is recent (within 1 second)
	timeDiff := time.Since(event.OccurredAt())
	if timeDiff > time.Second {
		t.Errorf("OccurredAt() should be recent, but was %v ago", timeDiff)
	}
}

func TestNewPositionCreatedEvent(t *testing.T) {
	positionID := uuid.New().String()
	userID := uuid.New().String()
	symbol := "AAPL"
	quantity := 100.0
	price := 150.0
	totalInvestment := 15000.0
	positionType := PositionTypeLong
	createdFrom := "ORDER_EXECUTION"
	sourceOrderID := uuid.New().String()

	event := NewPositionCreatedEvent(
		positionID, userID, symbol, quantity, price, totalInvestment,
		positionType, createdFrom, &sourceOrderID,
	)

	// Test base event properties
	if event.EventType() != "PositionCreated" {
		t.Errorf("EventType() = %v, want %v", event.EventType(), "PositionCreated")
	}

	if event.PositionID() != positionID {
		t.Errorf("PositionID() = %v, want %v", event.PositionID(), positionID)
	}

	if event.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", event.UserID(), userID)
	}

	if event.Symbol() != symbol {
		t.Errorf("Symbol() = %v, want %v", event.Symbol(), symbol)
	}

	// Test specific event properties
	if event.InitialQuantity != quantity {
		t.Errorf("InitialQuantity = %v, want %v", event.InitialQuantity, quantity)
	}

	if event.InitialPrice != price {
		t.Errorf("InitialPrice = %v, want %v", event.InitialPrice, price)
	}

	if event.TotalInvestment != totalInvestment {
		t.Errorf("TotalInvestment = %v, want %v", event.TotalInvestment, totalInvestment)
	}

	if event.PositionType != positionType {
		t.Errorf("PositionType = %v, want %v", event.PositionType, positionType)
	}

	if event.CreatedFrom != createdFrom {
		t.Errorf("CreatedFrom = %v, want %v", event.CreatedFrom, createdFrom)
	}

	if event.SourceOrderID == nil || *event.SourceOrderID != sourceOrderID {
		t.Errorf("SourceOrderID = %v, want %v", event.SourceOrderID, &sourceOrderID)
	}
}

func TestNewPositionCreatedEvent_WithNilSourceOrderID(t *testing.T) {
	positionID := uuid.New().String()
	userID := uuid.New().String()
	symbol := "AAPL"

	event := NewPositionCreatedEvent(
		positionID, userID, symbol, 100.0, 150.0, 15000.0,
		PositionTypeLong, "MANUAL_ENTRY", nil,
	)

	if event.SourceOrderID != nil {
		t.Errorf("SourceOrderID should be nil, got %v", event.SourceOrderID)
	}
}

func TestNewPositionUpdatedEvent(t *testing.T) {
	positionID := uuid.New().String()
	userID := uuid.New().String()
	symbol := "AAPL"
	sourceOrderID := uuid.New().String()

	// Previous state
	prevQuantity := 100.0
	prevAvgPrice := 150.0
	prevStatus := PositionStatusActive

	// New state
	newQuantity := 150.0
	newAvgPrice := 153.33
	newStatus := PositionStatusActive

	// Transaction details
	transactionQuantity := 50.0
	transactionPrice := 160.0
	transactionType := "BUY"
	totalInvestment := 23000.0

	event := NewPositionUpdatedEvent(
		positionID, userID, symbol,
		prevQuantity, prevAvgPrice, prevStatus,
		newQuantity, newAvgPrice, newStatus,
		transactionQuantity, transactionPrice, transactionType,
		totalInvestment, &sourceOrderID,
	)

	// Test base event properties
	if event.EventType() != "PositionUpdated" {
		t.Errorf("EventType() = %v, want %v", event.EventType(), "PositionUpdated")
	}

	// Test previous state
	if event.PreviousQuantity != prevQuantity {
		t.Errorf("PreviousQuantity = %v, want %v", event.PreviousQuantity, prevQuantity)
	}

	if event.PreviousAveragePrice != prevAvgPrice {
		t.Errorf("PreviousAveragePrice = %v, want %v", event.PreviousAveragePrice, prevAvgPrice)
	}

	if event.PreviousStatus != prevStatus {
		t.Errorf("PreviousStatus = %v, want %v", event.PreviousStatus, prevStatus)
	}

	// Test new state
	if event.NewQuantity != newQuantity {
		t.Errorf("NewQuantity = %v, want %v", event.NewQuantity, newQuantity)
	}

	if event.NewAveragePrice != newAvgPrice {
		t.Errorf("NewAveragePrice = %v, want %v", event.NewAveragePrice, newAvgPrice)
	}

	if event.NewStatus != newStatus {
		t.Errorf("NewStatus = %v, want %v", event.NewStatus, newStatus)
	}

	// Test transaction details
	if event.TransactionQuantity != transactionQuantity {
		t.Errorf("TransactionQuantity = %v, want %v", event.TransactionQuantity, transactionQuantity)
	}

	if event.TransactionPrice != transactionPrice {
		t.Errorf("TransactionPrice = %v, want %v", event.TransactionPrice, transactionPrice)
	}

	if event.TransactionType != transactionType {
		t.Errorf("TransactionType = %v, want %v", event.TransactionType, transactionType)
	}

	if event.TotalInvestment != totalInvestment {
		t.Errorf("TotalInvestment = %v, want %v", event.TotalInvestment, totalInvestment)
	}

	if event.SourceOrderID == nil || *event.SourceOrderID != sourceOrderID {
		t.Errorf("SourceOrderID = %v, want %v", event.SourceOrderID, &sourceOrderID)
	}
}

func TestNewPositionClosedEvent(t *testing.T) {
	positionID := uuid.New().String()
	userID := uuid.New().String()
	symbol := "AAPL"
	sourceOrderID := uuid.New().String()

	finalQuantity := 100.0
	sellPrice := 165.0
	totalRealizedValue := 16500.0
	totalInvestment := 15000.0
	realizedPnL := 1500.0
	realizedPnLPct := 10.0
	holdingPeriod := 30 * 24 * time.Hour             // 30 days
	positionOpenedAt := time.Now().AddDate(0, -1, 0) // 1 month ago
	closedAt := time.Now()

	event := NewPositionClosedEvent(
		positionID, userID, symbol,
		finalQuantity, sellPrice, totalRealizedValue, totalInvestment,
		realizedPnL, realizedPnLPct, holdingPeriod,
		positionOpenedAt, closedAt, &sourceOrderID,
	)

	// Test base event properties
	if event.EventType() != "PositionClosed" {
		t.Errorf("EventType() = %v, want %v", event.EventType(), "PositionClosed")
	}

	// Test specific event properties
	if event.FinalQuantitySold != finalQuantity {
		t.Errorf("FinalQuantitySold = %v, want %v", event.FinalQuantitySold, finalQuantity)
	}

	if event.FinalSellPrice != sellPrice {
		t.Errorf("FinalSellPrice = %v, want %v", event.FinalSellPrice, sellPrice)
	}

	if event.TotalRealizedValue != totalRealizedValue {
		t.Errorf("TotalRealizedValue = %v, want %v", event.TotalRealizedValue, totalRealizedValue)
	}

	if event.TotalInvestment != totalInvestment {
		t.Errorf("TotalInvestment = %v, want %v", event.TotalInvestment, totalInvestment)
	}

	if event.RealizedPnL != realizedPnL {
		t.Errorf("RealizedPnL = %v, want %v", event.RealizedPnL, realizedPnL)
	}

	if event.RealizedPnLPct != realizedPnLPct {
		t.Errorf("RealizedPnLPct = %v, want %v", event.RealizedPnLPct, realizedPnLPct)
	}

	if event.HoldingPeriod != holdingPeriod {
		t.Errorf("HoldingPeriod = %v, want %v", event.HoldingPeriod, holdingPeriod)
	}

	if !event.PositionOpenedAt.Equal(positionOpenedAt) {
		t.Errorf("PositionOpenedAt = %v, want %v", event.PositionOpenedAt, positionOpenedAt)
	}

	if !event.PositionClosedAt.Equal(closedAt) {
		t.Errorf("PositionClosedAt = %v, want %v", event.PositionClosedAt, closedAt)
	}

	if event.SourceOrderID == nil || *event.SourceOrderID != sourceOrderID {
		t.Errorf("SourceOrderID = %v, want %v", event.SourceOrderID, &sourceOrderID)
	}
}

func TestNewPositionPriceUpdatedEvent(t *testing.T) {
	positionID := uuid.New().String()
	userID := uuid.New().String()
	symbol := "AAPL"
	dataSource := "MARKET_DATA_SERVICE"
	marketTimestamp := time.Now()

	prevPrice := 150.0
	newPrice := 155.0
	prevMarketValue := 15000.0
	newMarketValue := 15500.0
	prevPnL := 0.0
	newPnL := 500.0
	prevPnLPct := 0.0
	newPnLPct := 3.33

	event := NewPositionPriceUpdatedEvent(
		positionID, userID, symbol,
		prevPrice, newPrice, prevMarketValue, newMarketValue,
		prevPnL, newPnL, prevPnLPct, newPnLPct,
		dataSource, marketTimestamp,
	)

	// Test base event properties
	if event.EventType() != "PositionPriceUpdated" {
		t.Errorf("EventType() = %v, want %v", event.EventType(), "PositionPriceUpdated")
	}

	// Test specific event properties
	if event.PreviousCurrentPrice != prevPrice {
		t.Errorf("PreviousCurrentPrice = %v, want %v", event.PreviousCurrentPrice, prevPrice)
	}

	if event.NewCurrentPrice != newPrice {
		t.Errorf("NewCurrentPrice = %v, want %v", event.NewCurrentPrice, newPrice)
	}

	if event.PreviousMarketValue != prevMarketValue {
		t.Errorf("PreviousMarketValue = %v, want %v", event.PreviousMarketValue, prevMarketValue)
	}

	if event.NewMarketValue != newMarketValue {
		t.Errorf("NewMarketValue = %v, want %v", event.NewMarketValue, newMarketValue)
	}

	if event.PreviousUnrealizedPnL != prevPnL {
		t.Errorf("PreviousUnrealizedPnL = %v, want %v", event.PreviousUnrealizedPnL, prevPnL)
	}

	if event.NewUnrealizedPnL != newPnL {
		t.Errorf("NewUnrealizedPnL = %v, want %v", event.NewUnrealizedPnL, newPnL)
	}

	if event.PreviousUnrealizedPnLPct != prevPnLPct {
		t.Errorf("PreviousUnrealizedPnLPct = %v, want %v", event.PreviousUnrealizedPnLPct, prevPnLPct)
	}

	if event.NewUnrealizedPnLPct != newPnLPct {
		t.Errorf("NewUnrealizedPnLPct = %v, want %v", event.NewUnrealizedPnLPct, newPnLPct)
	}

	if event.MarketDataSource != dataSource {
		t.Errorf("MarketDataSource = %v, want %v", event.MarketDataSource, dataSource)
	}

	if !event.MarketDataTimestamp.Equal(marketTimestamp) {
		t.Errorf("MarketDataTimestamp = %v, want %v", event.MarketDataTimestamp, marketTimestamp)
	}
}

func TestNewPositionValidationFailedEvent(t *testing.T) {
	positionID := uuid.New().String()
	userID := uuid.New().String()
	symbol := "AAPL"
	validationErrors := []string{
		"insufficient quantity for sell order",
		"position is already closed",
	}
	validationContext := "ORDER_EXECUTION"
	validatedAt := time.Now()

	event := NewPositionValidationFailedEvent(
		positionID, userID, symbol,
		validationErrors, validationContext, validatedAt,
	)

	// Test base event properties
	if event.EventType() != "PositionValidationFailed" {
		t.Errorf("EventType() = %v, want %v", event.EventType(), "PositionValidationFailed")
	}

	// Test specific event properties
	if len(event.ValidationErrors) != len(validationErrors) {
		t.Errorf("ValidationErrors length = %v, want %v", len(event.ValidationErrors), len(validationErrors))
	}

	for i, expectedError := range validationErrors {
		if i < len(event.ValidationErrors) && event.ValidationErrors[i] != expectedError {
			t.Errorf("ValidationErrors[%d] = %v, want %v", i, event.ValidationErrors[i], expectedError)
		}
	}

	if event.ValidationContext != validationContext {
		t.Errorf("ValidationContext = %v, want %v", event.ValidationContext, validationContext)
	}

	if !event.ValidatedAt.Equal(validatedAt) {
		t.Errorf("ValidatedAt = %v, want %v", event.ValidatedAt, validatedAt)
	}
}

func TestDomainEventInterface_Compliance(t *testing.T) {
	// Test that all position events implement DomainEvent interface
	positionID := uuid.New().String()
	userID := uuid.New().String()
	symbol := "AAPL"

	events := []DomainEvent{
		NewPositionCreatedEvent(positionID, userID, symbol, 100.0, 150.0, 15000.0, PositionTypeLong, "ORDER_EXECUTION", nil),
		NewPositionUpdatedEvent(positionID, userID, symbol, 100.0, 150.0, PositionStatusActive, 150.0, 153.33, PositionStatusActive, 50.0, 160.0, "BUY", 23000.0, nil),
		NewPositionClosedEvent(positionID, userID, symbol, 100.0, 165.0, 16500.0, 15000.0, 1500.0, 10.0, time.Hour, time.Now().Add(-time.Hour), time.Now(), nil),
		NewPositionPriceUpdatedEvent(positionID, userID, symbol, 150.0, 155.0, 15000.0, 15500.0, 0.0, 500.0, 0.0, 3.33, "MARKET_DATA_SERVICE", time.Now()),
		NewPositionValidationFailedEvent(positionID, userID, symbol, []string{"test error"}, "ORDER_EXECUTION", time.Now()),
	}

	for i, event := range events {
		t.Run(event.EventType(), func(t *testing.T) {
			// Test that all required methods exist and return non-empty values
			if event.EventID() == "" {
				t.Errorf("Event %d: EventID() should not be empty", i)
			}

			if event.EventType() == "" {
				t.Errorf("Event %d: EventType() should not be empty", i)
			}

			if event.AggregateID() == "" {
				t.Errorf("Event %d: AggregateID() should not be empty", i)
			}

			if event.OccurredAt().IsZero() {
				t.Errorf("Event %d: OccurredAt() should not be zero", i)
			}
		})
	}
}

func TestPositionEvent_UniqueEventIDs(t *testing.T) {
	positionID := uuid.New().String()
	userID := uuid.New().String()
	symbol := "AAPL"

	// Create multiple events and ensure they have unique IDs
	events := []*PositionCreatedEvent{
		NewPositionCreatedEvent(positionID, userID, symbol, 100.0, 150.0, 15000.0, PositionTypeLong, "ORDER_EXECUTION", nil),
		NewPositionCreatedEvent(positionID, userID, symbol, 200.0, 160.0, 32000.0, PositionTypeLong, "ORDER_EXECUTION", nil),
		NewPositionCreatedEvent(positionID, userID, symbol, 50.0, 140.0, 7000.0, PositionTypeLong, "ORDER_EXECUTION", nil),
	}

	eventIDs := make(map[string]bool)
	for i, event := range events {
		eventID := event.EventID()
		if eventIDs[eventID] {
			t.Errorf("Duplicate EventID found for event %d: %s", i, eventID)
		}
		eventIDs[eventID] = true
	}
}
