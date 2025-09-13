package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents the base interface for all domain events
// This reuses the same interface from order events for consistency
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

type PositionEvent struct {
	eventID     string
	eventType   string
	aggregateID string
	occurredAt  time.Time
	positionID  string
	userID      string
	symbol      string
}

func NewPositionEvent(eventType, positionID, userID, symbol string) PositionEvent {
	return PositionEvent{
		eventID:     uuid.New().String(),
		eventType:   eventType,
		aggregateID: positionID,
		occurredAt:  time.Now(),
		positionID:  positionID,
		userID:      userID,
		symbol:      symbol,
	}
}

func (e PositionEvent) EventID() string {
	return e.eventID
}

func (e PositionEvent) EventType() string {
	return e.eventType
}

func (e PositionEvent) AggregateID() string {
	return e.aggregateID
}

func (e PositionEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e PositionEvent) PositionID() string {
	return e.positionID
}

func (e PositionEvent) UserID() string {
	return e.userID
}

func (e PositionEvent) Symbol() string {
	return e.symbol
}

type PositionCreatedEvent struct {
	PositionEvent
	InitialQuantity float64
	InitialPrice    float64
	TotalInvestment float64
	PositionType    PositionType
	CreatedFrom     string  // e.g., "ORDER_EXECUTION", "MANUAL_ENTRY"
	SourceOrderID   *string // optional: ID of the order that created this position
}

func NewPositionCreatedEvent(positionID, userID, symbol string, quantity, price, totalInvestment float64,
	positionType PositionType, createdFrom string, sourceOrderID *string) *PositionCreatedEvent {
	return &PositionCreatedEvent{
		PositionEvent:   NewPositionEvent("PositionCreated", positionID, userID, symbol),
		InitialQuantity: quantity,
		InitialPrice:    price,
		TotalInvestment: totalInvestment,
		PositionType:    positionType,
		CreatedFrom:     createdFrom,
		SourceOrderID:   sourceOrderID,
	}
}

type PositionUpdatedEvent struct {
	PositionEvent
	// Before state
	PreviousQuantity     float64
	PreviousAveragePrice float64
	PreviousStatus       PositionStatus
	// After state
	NewQuantity     float64
	NewAveragePrice float64
	NewStatus       PositionStatus
	// Transaction details
	TransactionQuantity float64
	TransactionPrice    float64
	TransactionType     string // "BUY" or "SELL"
	SourceOrderID       *string
	// Calculated fields
	TotalInvestment float64
}

func NewPositionUpdatedEvent(positionID, userID, symbol string,
	prevQuantity, prevAvgPrice float64, prevStatus PositionStatus,
	newQuantity, newAvgPrice float64, newStatus PositionStatus,
	transactionQuantity, transactionPrice float64, transactionType string,
	totalInvestment float64, sourceOrderID *string) *PositionUpdatedEvent {
	return &PositionUpdatedEvent{
		PositionEvent:        NewPositionEvent("PositionUpdated", positionID, userID, symbol),
		PreviousQuantity:     prevQuantity,
		PreviousAveragePrice: prevAvgPrice,
		PreviousStatus:       prevStatus,
		NewQuantity:          newQuantity,
		NewAveragePrice:      newAvgPrice,
		NewStatus:            newStatus,
		TransactionQuantity:  transactionQuantity,
		TransactionPrice:     transactionPrice,
		TransactionType:      transactionType,
		SourceOrderID:        sourceOrderID,
		TotalInvestment:      totalInvestment,
	}
}

type PositionClosedEvent struct {
	PositionEvent
	FinalQuantitySold  float64
	FinalSellPrice     float64
	TotalRealizedValue float64
	TotalInvestment    float64
	RealizedPnL        float64
	RealizedPnLPct     float64
	HoldingPeriod      time.Duration // how long the position was held
	SourceOrderID      *string
	PositionOpenedAt   time.Time
	PositionClosedAt   time.Time
}

func NewPositionClosedEvent(positionID, userID, symbol string,
	finalQuantity, sellPrice, totalRealizedValue, totalInvestment float64,
	realizedPnL, realizedPnLPct float64, holdingPeriod time.Duration,
	positionOpenedAt, closedAt time.Time, sourceOrderID *string) *PositionClosedEvent {
	return &PositionClosedEvent{
		PositionEvent:      NewPositionEvent("PositionClosed", positionID, userID, symbol),
		FinalQuantitySold:  finalQuantity,
		FinalSellPrice:     sellPrice,
		TotalRealizedValue: totalRealizedValue,
		TotalInvestment:    totalInvestment,
		RealizedPnL:        realizedPnL,
		RealizedPnLPct:     realizedPnLPct,
		HoldingPeriod:      holdingPeriod,
		SourceOrderID:      sourceOrderID,
		PositionOpenedAt:   positionOpenedAt,
		PositionClosedAt:   closedAt,
	}
}

type PositionPriceUpdatedEvent struct {
	PositionEvent
	PreviousCurrentPrice     float64
	NewCurrentPrice          float64
	PreviousMarketValue      float64
	NewMarketValue           float64
	PreviousUnrealizedPnL    float64
	NewUnrealizedPnL         float64
	PreviousUnrealizedPnLPct float64
	NewUnrealizedPnLPct      float64
	MarketDataSource         string
	MarketDataTimestamp      time.Time
}

func NewPositionPriceUpdatedEvent(positionID, userID, symbol string,
	prevPrice, newPrice, prevMarketValue, newMarketValue float64,
	prevPnL, newPnL, prevPnLPct, newPnLPct float64,
	dataSource string, marketTimestamp time.Time) *PositionPriceUpdatedEvent {
	return &PositionPriceUpdatedEvent{
		PositionEvent:            NewPositionEvent("PositionPriceUpdated", positionID, userID, symbol),
		PreviousCurrentPrice:     prevPrice,
		NewCurrentPrice:          newPrice,
		PreviousMarketValue:      prevMarketValue,
		NewMarketValue:           newMarketValue,
		PreviousUnrealizedPnL:    prevPnL,
		NewUnrealizedPnL:         newPnL,
		PreviousUnrealizedPnLPct: prevPnLPct,
		NewUnrealizedPnLPct:      newPnLPct,
		MarketDataSource:         dataSource,
		MarketDataTimestamp:      marketTimestamp,
	}
}

type PositionValidationFailedEvent struct {
	PositionEvent
	ValidationErrors  []string
	ValidatedAt       time.Time
	ValidationContext string // e.g., "ORDER_EXECUTION", "POSITION_UPDATE"
}

func NewPositionValidationFailedEvent(positionID, userID, symbol string,
	validationErrors []string, validationContext string, validatedAt time.Time) *PositionValidationFailedEvent {
	return &PositionValidationFailedEvent{
		PositionEvent:     NewPositionEvent("PositionValidationFailed", positionID, userID, symbol),
		ValidationErrors:  validationErrors,
		ValidatedAt:       validatedAt,
		ValidationContext: validationContext,
	}
}
