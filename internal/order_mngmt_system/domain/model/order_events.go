package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents the base interface for all domain events
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

// OrderEvent represents the base struct for all order-related domain events
type OrderEvent struct {
	eventID     string
	eventType   string
	aggregateID string
	occurredAt  time.Time
	orderID     string
	userID      string
}

// NewOrderEvent creates a new base order event
func NewOrderEvent(eventType, orderID, userID string) OrderEvent {
	return OrderEvent{
		eventID:     uuid.New().String(),
		eventType:   eventType,
		aggregateID: orderID,
		occurredAt:  time.Now(),
		orderID:     orderID,
		userID:      userID,
	}
}

// EventID returns the unique event identifier
func (e OrderEvent) EventID() string {
	return e.eventID
}

// EventType returns the type of the event
func (e OrderEvent) EventType() string {
	return e.eventType
}

// AggregateID returns the ID of the aggregate that generated this event
func (e OrderEvent) AggregateID() string {
	return e.aggregateID
}

// OccurredAt returns when the event occurred
func (e OrderEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// OrderID returns the order ID
func (e OrderEvent) OrderID() string {
	return e.orderID
}

// UserID returns the user ID who owns the order
func (e OrderEvent) UserID() string {
	return e.userID
}

// OrderSubmittedEvent represents an event when an order is submitted
type OrderSubmittedEvent struct {
	OrderEvent
	Symbol    string
	OrderSide OrderSide
	OrderType OrderType
	Quantity  float64
	Price     *float64
}

// NewOrderSubmittedEvent creates a new OrderSubmittedEvent
func NewOrderSubmittedEvent(orderID, userID, symbol string, orderSide OrderSide, orderType OrderType, quantity float64, price *float64) *OrderSubmittedEvent {
	return &OrderSubmittedEvent{
		OrderEvent: NewOrderEvent("OrderSubmitted", orderID, userID),
		Symbol:     symbol,
		OrderSide:  orderSide,
		OrderType:  orderType,
		Quantity:   quantity,
		Price:      price,
	}
}

// OrderProcessingStartedEvent represents an event when order processing starts
type OrderProcessingStartedEvent struct {
	OrderEvent
	MarketPrice     *float64
	MarketTimestamp *time.Time
}

// NewOrderProcessingStartedEvent creates a new OrderProcessingStartedEvent
func NewOrderProcessingStartedEvent(orderID, userID string, marketPrice *float64, marketTimestamp *time.Time) *OrderProcessingStartedEvent {
	return &OrderProcessingStartedEvent{
		OrderEvent:      NewOrderEvent("OrderProcessingStarted", orderID, userID),
		MarketPrice:     marketPrice,
		MarketTimestamp: marketTimestamp,
	}
}

// OrderExecutedEvent represents an event when an order is successfully executed
type OrderExecutedEvent struct {
	OrderEvent
	ExecutionPrice float64
	ExecutedAt     time.Time
	TotalValue     float64
}

// NewOrderExecutedEvent creates a new OrderExecutedEvent
func NewOrderExecutedEvent(orderID, userID string, executionPrice, totalValue float64, executedAt time.Time) *OrderExecutedEvent {
	return &OrderExecutedEvent{
		OrderEvent:     NewOrderEvent("OrderExecuted", orderID, userID),
		ExecutionPrice: executionPrice,
		ExecutedAt:     executedAt,
		TotalValue:     totalValue,
	}
}

// OrderFailedEvent represents an event when an order fails to execute
type OrderFailedEvent struct {
	OrderEvent
	FailureReason string
	FailedAt      time.Time
}

// NewOrderFailedEvent creates a new OrderFailedEvent
func NewOrderFailedEvent(orderID, userID, failureReason string, failedAt time.Time) *OrderFailedEvent {
	return &OrderFailedEvent{
		OrderEvent:    NewOrderEvent("OrderFailed", orderID, userID),
		FailureReason: failureReason,
		FailedAt:      failedAt,
	}
}

// OrderCancelledEvent represents an event when an order is cancelled
type OrderCancelledEvent struct {
	OrderEvent
	CancelledAt  time.Time
	CancelReason string
	CancelledBy  string // user ID who cancelled (could be admin)
}

// NewOrderCancelledEvent creates a new OrderCancelledEvent
func NewOrderCancelledEvent(orderID, userID, cancelReason, cancelledBy string, cancelledAt time.Time) *OrderCancelledEvent {
	return &OrderCancelledEvent{
		OrderEvent:   NewOrderEvent("OrderCancelled", orderID, userID),
		CancelledAt:  cancelledAt,
		CancelReason: cancelReason,
		CancelledBy:  cancelledBy,
	}
}

// MarketDataReceivedEvent represents an event when market data is received for an order
type MarketDataReceivedEvent struct {
	OrderEvent
	Symbol              string
	MarketPrice         float64
	MarketDataTimestamp time.Time
	DataSource          string
}

// NewMarketDataReceivedEvent creates a new MarketDataReceivedEvent
func NewMarketDataReceivedEvent(orderID, userID, symbol, dataSource string, marketPrice float64, marketDataTimestamp time.Time) *MarketDataReceivedEvent {
	return &MarketDataReceivedEvent{
		OrderEvent:          NewOrderEvent("MarketDataReceived", orderID, userID),
		Symbol:              symbol,
		MarketPrice:         marketPrice,
		MarketDataTimestamp: marketDataTimestamp,
		DataSource:          dataSource,
	}
}

// OrderStatusChangedEvent represents an event when an order status changes
type OrderStatusChangedEvent struct {
	OrderEvent
	FromStatus OrderStatus
	ToStatus   OrderStatus
	ChangedAt  time.Time
	Reason     string
}

// NewOrderStatusChangedEvent creates a new OrderStatusChangedEvent
func NewOrderStatusChangedEvent(orderID, userID string, fromStatus, toStatus OrderStatus, reason string, changedAt time.Time) *OrderStatusChangedEvent {
	return &OrderStatusChangedEvent{
		OrderEvent: NewOrderEvent("OrderStatusChanged", orderID, userID),
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		ChangedAt:  changedAt,
		Reason:     reason,
	}
}

// RiskCheckPerformedEvent represents an event when risk checks are performed
type RiskCheckPerformedEvent struct {
	OrderEvent
	RiskCheckType   string
	RiskCheckResult string // PASSED, FAILED, WARNING
	RiskScore       float64
	CheckedAt       time.Time
}

// NewRiskCheckPerformedEvent creates a new RiskCheckPerformedEvent
func NewRiskCheckPerformedEvent(orderID, userID, riskCheckType, result string, riskScore float64, checkedAt time.Time) *RiskCheckPerformedEvent {
	return &RiskCheckPerformedEvent{
		OrderEvent:      NewOrderEvent("RiskCheckPerformed", orderID, userID),
		RiskCheckType:   riskCheckType,
		RiskCheckResult: result,
		RiskScore:       riskScore,
		CheckedAt:       checkedAt,
	}
}

// OrderValidationFailedEvent represents an event when order validation fails
type OrderValidationFailedEvent struct {
	OrderEvent
	ValidationErrors []string
	ValidatedAt      time.Time
}

// NewOrderValidationFailedEvent creates a new OrderValidationFailedEvent
func NewOrderValidationFailedEvent(orderID, userID string, validationErrors []string, validatedAt time.Time) *OrderValidationFailedEvent {
	return &OrderValidationFailedEvent{
		OrderEvent:       NewOrderEvent("OrderValidationFailed", orderID, userID),
		ValidationErrors: validationErrors,
		ValidatedAt:      validatedAt,
	}
}

// PositionValidationFailedEvent represents an event when position validation fails for sell orders
type PositionValidationFailedEvent struct {
	OrderEvent
	Symbol            string
	RequestedQuantity float64
	AvailableQuantity float64
	ValidationError   string
}

// NewPositionValidationFailedEvent creates a new PositionValidationFailedEvent
func NewPositionValidationFailedEvent(orderID, userID, symbol string, requestedQuantity, availableQuantity float64, validationError string) *PositionValidationFailedEvent {
	return &PositionValidationFailedEvent{
		OrderEvent:        NewOrderEvent("PositionValidationFailed", orderID, userID),
		Symbol:            symbol,
		RequestedQuantity: requestedQuantity,
		AvailableQuantity: availableQuantity,
		ValidationError:   validationError,
	}
}
