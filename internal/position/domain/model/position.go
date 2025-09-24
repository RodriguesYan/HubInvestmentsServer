package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Position struct {
	ID               uuid.UUID      `json:"id"`
	UserID           uuid.UUID      `json:"userId"`
	Symbol           string         `json:"symbol"`
	Quantity         float64        `json:"quantity"`
	AveragePrice     float64        `json:"averagePrice"`
	TotalInvestment  float64        `json:"totalInvestment"`
	CurrentPrice     float64        `json:"currentPrice,omitempty"`
	MarketValue      float64        `json:"marketValue,omitempty"`
	UnrealizedPnL    float64        `json:"unrealizedPnL,omitempty"`
	UnrealizedPnLPct float64        `json:"unrealizedPnLPct,omitempty"`
	PositionType     PositionType   `json:"positionType"`
	Status           PositionStatus `json:"status"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	LastTradeAt      *time.Time     `json:"lastTradeAt,omitempty"`

	// Domain events (not serialized to JSON)
	events []DomainEvent `json:"-"`
}

func NewPosition(userID uuid.UUID, symbol string, quantity float64, price float64, positionType PositionType) (*Position, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}

	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	if price <= 0 {
		return nil, errors.New("price must be greater than zero")
	}

	if !positionType.IsValid() {
		return nil, errors.New("invalid position type")
	}

	now := time.Now()

	position := &Position{
		ID:              uuid.New(),
		UserID:          userID,
		Symbol:          symbol,
		Quantity:        quantity,
		AveragePrice:    price,
		TotalInvestment: quantity * price,
		PositionType:    positionType,
		Status:          PositionStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
		LastTradeAt:     &now,
		events:          make([]DomainEvent, 0),
	}

	// Emit position created event
	createdEvent := NewPositionCreatedEvent(
		position.ID.String(),
		position.UserID.String(),
		position.Symbol,
		position.Quantity,
		position.AveragePrice,
		position.TotalInvestment,
		position.PositionType,
		"DIRECT_CREATION", // Could be "ORDER_EXECUTION" when called from order processing
		nil,               // No source order ID for direct creation
	)
	position.addEvent(createdEvent)

	return position, nil
}

// UpdateQuantity updates the position quantity and average price based on a new transaction
// For BUY orders: adds to quantity and recalculates average price
// For SELL orders: reduces quantity
func (p *Position) UpdateQuantity(tradeQuantity float64, tradePrice float64, isBuyOrder bool) error {
	return p.UpdateQuantityWithOrderID(tradeQuantity, tradePrice, isBuyOrder, nil)
}

// UpdateQuantityWithOrderID updates position quantity with optional order ID for event tracking
func (p *Position) UpdateQuantityWithOrderID(tradeQuantity float64, tradePrice float64, isBuyOrder bool, sourceOrderID *string) error {
	if tradeQuantity <= 0 {
		return errors.New("trade quantity must be greater than zero")
	}

	if tradePrice <= 0 {
		return errors.New("trade price must be greater than zero")
	}

	if !p.Status.CanBeUpdated() {
		return fmt.Errorf("cannot update position with status: %s", p.Status)
	}

	// Capture previous state for event
	prevQuantity := p.Quantity
	prevAveragePrice := p.AveragePrice
	prevStatus := p.Status

	now := time.Now()

	if isBuyOrder {
		// BUY order: increase quantity and recalculate average price
		newAveragePrice, err := p.CalculateNewAveragePrice(tradeQuantity, tradePrice)
		if err != nil {
			return fmt.Errorf("failed to calculate new average price: %w", err)
		}

		p.Quantity += tradeQuantity
		p.AveragePrice = newAveragePrice
		p.TotalInvestment = p.Quantity * p.AveragePrice

	} else {
		// SELL order: decrease quantity
		if !p.CanSell(tradeQuantity) {
			return fmt.Errorf("insufficient quantity to sell: have %.6f, trying to sell %.6f", p.Quantity, tradeQuantity)
		}

		p.Quantity -= tradeQuantity
		p.TotalInvestment = p.Quantity * p.AveragePrice

		// Update status based on remaining quantity
		if p.Quantity == 0 {
			p.Status = PositionStatusClosed
		} else if p.Status == PositionStatusActive {
			// Once we start selling from an active position, it becomes partial
			p.Status = PositionStatusPartial
		}
	}

	p.UpdatedAt = now
	p.LastTradeAt = &now

	// Emit position updated event
	transactionType := "BUY"
	if !isBuyOrder {
		transactionType = "SELL"
	}

	updatedEvent := NewPositionUpdatedEvent(
		p.ID.String(),
		p.UserID.String(),
		p.Symbol,
		prevQuantity, prevAveragePrice, prevStatus,
		p.Quantity, p.AveragePrice, p.Status,
		tradeQuantity, tradePrice, transactionType,
		p.TotalInvestment,
		sourceOrderID,
	)
	p.addEvent(updatedEvent)

	// If position was closed, emit closed event
	if p.Status == PositionStatusClosed {
		holdingPeriod := now.Sub(p.CreatedAt)
		realizedValue := tradeQuantity * tradePrice
		realizedPnL := realizedValue - (prevAveragePrice * tradeQuantity)
		var realizedPnLPct float64
		if prevAveragePrice > 0 {
			realizedPnLPct = (realizedPnL / (prevAveragePrice * tradeQuantity)) * 100
		}

		closedEvent := NewPositionClosedEvent(
			p.ID.String(),
			p.UserID.String(),
			p.Symbol,
			tradeQuantity,     // final quantity sold
			tradePrice,        // final sell price
			realizedValue,     // total realized value from this final sale
			p.TotalInvestment, // this should be 0 now
			realizedPnL,
			realizedPnLPct,
			holdingPeriod,
			p.CreatedAt,
			now,
			sourceOrderID,
		)
		p.addEvent(closedEvent)
	}

	return nil
}

func (p *Position) CalculateNewAveragePrice(newQuantity float64, newPrice float64) (float64, error) {
	if newQuantity <= 0 {
		return 0, errors.New("new quantity must be greater than zero")
	}

	if newPrice <= 0 {
		return 0, errors.New("new price must be greater than zero")
	}

	// New Average Price = (Existing Investment + New Investment) / (Existing Quantity + New Quantity)
	existingInvestment := p.Quantity * p.AveragePrice
	newInvestment := newQuantity * newPrice
	totalInvestment := existingInvestment + newInvestment
	totalQuantity := p.Quantity + newQuantity

	if totalQuantity == 0 {
		return 0, errors.New("total quantity cannot be zero")
	}

	return totalInvestment / totalQuantity, nil
}

func (p *Position) CanSell(sellQuantity float64) bool {
	if sellQuantity <= 0 {
		return false
	}

	if !p.Status.CanBeUpdated() {
		return false
	}

	return p.Quantity >= sellQuantity
}

// UpdateCurrentPrice updates the current market price and recalculates market value and PnL
func (p *Position) UpdateCurrentPrice(currentPrice float64) error {
	return p.UpdateCurrentPriceWithSource(currentPrice, "UNKNOWN", time.Now())
}

func (p *Position) UpdateCurrentPriceWithSource(currentPrice float64, dataSource string, marketTimestamp time.Time) error {
	if currentPrice <= 0 {
		return errors.New("current price must be greater than zero")
	}

	// Capture previous state for event
	prevPrice := p.CurrentPrice
	prevMarketValue := p.MarketValue
	prevUnrealizedPnL := p.UnrealizedPnL
	prevUnrealizedPnLPct := p.UnrealizedPnLPct

	p.CurrentPrice = currentPrice
	p.MarketValue = p.Quantity * currentPrice
	p.UnrealizedPnL = p.MarketValue - p.TotalInvestment

	if p.TotalInvestment > 0 {
		p.UnrealizedPnLPct = (p.UnrealizedPnL / p.TotalInvestment) * 100
	} else {
		p.UnrealizedPnLPct = 0
	}

	// Only emit event if the price actually changed
	if prevPrice != currentPrice {
		priceUpdatedEvent := NewPositionPriceUpdatedEvent(
			p.ID.String(),
			p.UserID.String(),
			p.Symbol,
			prevPrice, currentPrice,
			prevMarketValue, p.MarketValue,
			prevUnrealizedPnL, p.UnrealizedPnL,
			prevUnrealizedPnLPct, p.UnrealizedPnLPct,
			dataSource,
			marketTimestamp,
		)
		p.addEvent(priceUpdatedEvent)
	}

	return nil
}

// Validate performs comprehensive validation of the position
func (p *Position) Validate() error {
	if p.ID == uuid.Nil {
		return errors.New("position ID cannot be empty")
	}

	if p.UserID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	if p.Symbol == "" {
		return errors.New("symbol cannot be empty")
	}

	if p.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	if p.AveragePrice < 0 {
		return errors.New("average price cannot be negative")
	}

	if p.TotalInvestment < 0 {
		return errors.New("total investment cannot be negative")
	}

	if !p.PositionType.IsValid() {
		return errors.New("invalid position type")
	}

	if !p.Status.IsValid() {
		return errors.New("invalid position status")
	}

	if p.CreatedAt.IsZero() {
		return errors.New("created at cannot be zero")
	}

	if p.UpdatedAt.IsZero() {
		return errors.New("updated at cannot be zero")
	}

	if p.Status == PositionStatusClosed && p.Quantity > 0 {
		return errors.New("closed position cannot have quantity greater than zero")
	}

	if p.Status == PositionStatusActive && p.Quantity == 0 {
		return errors.New("active position must have quantity greater than zero")
	}

	return nil
}

func (p *Position) CanBeClosed() bool {
	return p.Status.CanBeUpdated() && p.Quantity > 0
}

// IsEmpty returns true if the position has zero quantity
func (p *Position) IsEmpty() bool {
	return p.Quantity == 0
}

func (p *Position) GetRealizedValue(currentPrice float64) float64 {
	if currentPrice <= 0 {
		return 0
	}
	return p.Quantity * currentPrice
}

// String returns a string representation of the position
func (p *Position) String() string {
	return fmt.Sprintf("Position{ID: %s, UserID: %s, Symbol: %s, Quantity: %.6f, AvgPrice: %.2f, Status: %s}",
		p.ID, p.UserID, p.Symbol, p.Quantity, p.AveragePrice, p.Status)
}

// Domain Event Management Methods

func (p *Position) addEvent(event DomainEvent) {
	if p.events == nil {
		p.events = make([]DomainEvent, 0)
	}
	p.events = append(p.events, event)
}

func (p *Position) GetEvents() []DomainEvent {
	if p.events == nil {
		return make([]DomainEvent, 0)
	}
	// Return a copy to prevent external modification
	eventsCopy := make([]DomainEvent, len(p.events))
	copy(eventsCopy, p.events)
	return eventsCopy
}

func (p *Position) ClearEvents() {
	p.events = make([]DomainEvent, 0)
}

func (p *Position) HasEvents() bool {
	return len(p.events) > 0
}

func (p *Position) EmitValidationFailedEvent(validationErrors []string, validationContext string) {
	validationEvent := NewPositionValidationFailedEvent(
		p.ID.String(),
		p.UserID.String(),
		p.Symbol,
		validationErrors,
		validationContext,
		time.Now(),
	)
	p.addEvent(validationEvent)
}
