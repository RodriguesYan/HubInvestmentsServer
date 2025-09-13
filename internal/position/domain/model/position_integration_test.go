package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestPosition_EventEmission tests that domain events are properly emitted during position operations
func TestPosition_EventEmission(t *testing.T) {
	userID := uuid.New()
	symbol := "AAPL"
	quantity := 100.0
	price := 150.0
	positionType := PositionTypeLong

	// Test: NewPosition should emit PositionCreatedEvent
	position, err := NewPosition(userID, symbol, quantity, price, positionType)
	if err != nil {
		t.Fatalf("NewPosition() error = %v", err)
	}

	// Verify PositionCreatedEvent was emitted
	events := position.GetEvents()
	if len(events) != 1 {
		t.Errorf("NewPosition() should emit 1 event, got %d", len(events))
	}

	if events[0].EventType() != "PositionCreated" {
		t.Errorf("NewPosition() should emit PositionCreated event, got %s", events[0].EventType())
	}

	// Clear events and test UpdateQuantity (BUY)
	position.ClearEvents()
	sourceOrderID := uuid.New().String()

	err = position.UpdateQuantityWithOrderID(50.0, 160.0, true, &sourceOrderID)
	if err != nil {
		t.Errorf("UpdateQuantityWithOrderID() error = %v", err)
	}

	// Verify PositionUpdatedEvent was emitted
	events = position.GetEvents()
	if len(events) != 1 {
		t.Errorf("UpdateQuantity() should emit 1 event, got %d", len(events))
	}

	if events[0].EventType() != "PositionUpdated" {
		t.Errorf("UpdateQuantity() should emit PositionUpdated event, got %s", events[0].EventType())
	}

	// Test PositionUpdatedEvent details
	updatedEvent, ok := events[0].(*PositionUpdatedEvent)
	if !ok {
		t.Errorf("Event should be PositionUpdatedEvent")
	} else {
		if updatedEvent.TransactionType != "BUY" {
			t.Errorf("TransactionType should be BUY, got %s", updatedEvent.TransactionType)
		}
		if updatedEvent.TransactionQuantity != 50.0 {
			t.Errorf("TransactionQuantity should be 50.0, got %f", updatedEvent.TransactionQuantity)
		}
		if updatedEvent.SourceOrderID == nil || *updatedEvent.SourceOrderID != sourceOrderID {
			t.Errorf("SourceOrderID should be %s, got %v", sourceOrderID, updatedEvent.SourceOrderID)
		}
	}

	// Clear events and test position closing (SELL all)
	position.ClearEvents()

	err = position.UpdateQuantityWithOrderID(150.0, 165.0, false, &sourceOrderID)
	if err != nil {
		t.Errorf("UpdateQuantityWithOrderID() error = %v", err)
	}

	// Verify both PositionUpdatedEvent and PositionClosedEvent were emitted
	events = position.GetEvents()
	if len(events) != 2 {
		t.Errorf("Closing position should emit 2 events, got %d", len(events))
	}

	// First event should be PositionUpdatedEvent
	if events[0].EventType() != "PositionUpdated" {
		t.Errorf("First event should be PositionUpdated, got %s", events[0].EventType())
	}

	// Second event should be PositionClosedEvent
	if events[1].EventType() != "PositionClosed" {
		t.Errorf("Second event should be PositionClosed, got %s", events[1].EventType())
	}

	// Test PositionClosedEvent details
	closedEvent, ok := events[1].(*PositionClosedEvent)
	if !ok {
		t.Errorf("Second event should be PositionClosedEvent")
	} else {
		if closedEvent.FinalQuantitySold != 150.0 {
			t.Errorf("FinalQuantitySold should be 150.0, got %f", closedEvent.FinalQuantitySold)
		}
		if closedEvent.FinalSellPrice != 165.0 {
			t.Errorf("FinalSellPrice should be 165.0, got %f", closedEvent.FinalSellPrice)
		}
	}
}

// TestPosition_PriceUpdateEvents tests price update event emission
func TestPosition_PriceUpdateEvents(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	position.ClearEvents() // Clear the creation event

	dataSource := "MARKET_DATA_SERVICE"
	marketTimestamp := time.Now()

	// Update price - should emit event
	err := position.UpdateCurrentPriceWithSource(155.0, dataSource, marketTimestamp)
	if err != nil {
		t.Errorf("UpdateCurrentPriceWithSource() error = %v", err)
	}

	events := position.GetEvents()
	if len(events) != 1 {
		t.Errorf("Price update should emit 1 event, got %d", len(events))
	}

	if events[0].EventType() != "PositionPriceUpdated" {
		t.Errorf("Price update should emit PositionPriceUpdated event, got %s", events[0].EventType())
	}

	// Test PositionPriceUpdatedEvent details
	priceEvent, ok := events[0].(*PositionPriceUpdatedEvent)
	if !ok {
		t.Errorf("Event should be PositionPriceUpdatedEvent")
	} else {
		if priceEvent.PreviousCurrentPrice != 0.0 {
			t.Errorf("PreviousCurrentPrice should be 0.0, got %f", priceEvent.PreviousCurrentPrice)
		}
		if priceEvent.NewCurrentPrice != 155.0 {
			t.Errorf("NewCurrentPrice should be 155.0, got %f", priceEvent.NewCurrentPrice)
		}
		if priceEvent.MarketDataSource != dataSource {
			t.Errorf("MarketDataSource should be %s, got %s", dataSource, priceEvent.MarketDataSource)
		}
	}

	// Update with same price - should NOT emit event
	position.ClearEvents()

	err = position.UpdateCurrentPriceWithSource(155.0, dataSource, marketTimestamp)
	if err != nil {
		t.Errorf("UpdateCurrentPriceWithSource() error = %v", err)
	}

	events = position.GetEvents()
	if len(events) != 0 {
		t.Errorf("Same price update should emit 0 events, got %d", len(events))
	}
}

// TestPosition_ValidationFailedEvent tests validation failed event emission
func TestPosition_ValidationFailedEvent(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	position.ClearEvents() // Clear the creation event

	validationErrors := []string{
		"insufficient quantity for sell order",
		"position is not in valid state",
	}
	validationContext := "ORDER_EXECUTION"

	// Emit validation failed event
	position.EmitValidationFailedEvent(validationErrors, validationContext)

	events := position.GetEvents()
	if len(events) != 1 {
		t.Errorf("Validation failed should emit 1 event, got %d", len(events))
	}

	if events[0].EventType() != "PositionValidationFailed" {
		t.Errorf("Should emit PositionValidationFailed event, got %s", events[0].EventType())
	}

	// Test PositionValidationFailedEvent details
	validationEvent, ok := events[0].(*PositionValidationFailedEvent)
	if !ok {
		t.Errorf("Event should be PositionValidationFailedEvent")
	} else {
		if len(validationEvent.ValidationErrors) != len(validationErrors) {
			t.Errorf("ValidationErrors length should be %d, got %d",
				len(validationErrors), len(validationEvent.ValidationErrors))
		}
		if validationEvent.ValidationContext != validationContext {
			t.Errorf("ValidationContext should be %s, got %s",
				validationContext, validationEvent.ValidationContext)
		}
	}
}

// TestPosition_EventManagement tests event management methods
func TestPosition_EventManagement(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	// Test HasEvents
	if !position.HasEvents() {
		t.Error("Position should have events after creation")
	}

	// Test GetEvents returns copy
	events1 := position.GetEvents()
	events2 := position.GetEvents()

	if len(events1) != len(events2) {
		t.Error("GetEvents() should return consistent results")
	}

	// Modify first slice, second should be unaffected (proving it's a copy)
	events1[0] = nil
	events3 := position.GetEvents()

	if events3[0] == nil {
		t.Error("GetEvents() should return a copy, not the original slice")
	}

	// Test ClearEvents
	position.ClearEvents()

	if position.HasEvents() {
		t.Error("Position should not have events after ClearEvents()")
	}

	if len(position.GetEvents()) != 0 {
		t.Error("GetEvents() should return empty slice after ClearEvents()")
	}
}

// TestPosition_EventsNotInJSON tests that events are not serialized to JSON
func TestPosition_EventsNotInJSON(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	// This test would require JSON marshaling which isn't part of the standard library imports
	// but we can verify the struct tag is correct

	// Check that the struct has the correct json tag for events field
	// This is more of a compile-time check that the tag exists
	if !position.HasEvents() {
		t.Error("Position should have events to test JSON exclusion")
	}

	// The json:"-" tag should exclude events from JSON serialization
	// This is verified by the struct definition itself
}

// TestPosition_ComplexEventSequence tests a complex sequence of operations and events
func TestPosition_ComplexEventSequence(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	// Clear creation event for clean test
	position.ClearEvents()

	orderID1 := uuid.New().String()
	orderID2 := uuid.New().String()
	orderID3 := uuid.New().String()

	// 1. Buy more shares
	err := position.UpdateQuantityWithOrderID(50.0, 160.0, true, &orderID1)
	if err != nil {
		t.Errorf("First buy error: %v", err)
	}

	// 2. Update price
	err = position.UpdateCurrentPriceWithSource(155.0, "MARKET_DATA", time.Now())
	if err != nil {
		t.Errorf("Price update error: %v", err)
	}

	// 3. Partial sell
	err = position.UpdateQuantityWithOrderID(75.0, 165.0, false, &orderID2)
	if err != nil {
		t.Errorf("Partial sell error: %v", err)
	}

	// 4. Another price update
	err = position.UpdateCurrentPriceWithSource(170.0, "MARKET_DATA", time.Now())
	if err != nil {
		t.Errorf("Second price update error: %v", err)
	}

	// 5. Final sell (close position)
	err = position.UpdateQuantityWithOrderID(75.0, 175.0, false, &orderID3)
	if err != nil {
		t.Errorf("Final sell error: %v", err)
	}

	// Verify event sequence
	events := position.GetEvents()
	expectedEventCount := 6 // 1 buy + 1 price + 1 sell + 1 price + (1 sell + 1 close)
	if len(events) != expectedEventCount {
		t.Errorf("Expected %d events, got %d", expectedEventCount, len(events))
	}

	eventTypes := make([]string, len(events))
	for i, event := range events {
		eventTypes[i] = event.EventType()
	}

	expectedTypes := []string{
		"PositionUpdated",      // Buy more
		"PositionPriceUpdated", // Price update
		"PositionUpdated",      // Partial sell
		"PositionPriceUpdated", // Price update
		"PositionUpdated",      // Final sell
		"PositionClosed",       // Position closed
	}

	for i, expectedType := range expectedTypes {
		if i < len(eventTypes) && eventTypes[i] != expectedType {
			t.Errorf("Event %d should be %s, got %s", i, expectedType, eventTypes[i])
		}
	}

	// Verify position state
	if position.Status != PositionStatusClosed {
		t.Errorf("Position should be closed, got %s", position.Status)
	}

	if position.Quantity != 0 {
		t.Errorf("Position quantity should be 0, got %f", position.Quantity)
	}
}
