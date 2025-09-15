package domain

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

type OrderEvent struct {
	eventID     string
	eventType   string
	aggregateID string
	occurredAt  time.Time
	orderID     string
	userID      string
}

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

func (e OrderEvent) EventID() string {
	return e.eventID
}

func (e OrderEvent) EventType() string {
	return e.eventType
}

func (e OrderEvent) AggregateID() string {
	return e.aggregateID
}

func (e OrderEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e OrderEvent) OrderID() string {
	return e.orderID
}

func (e OrderEvent) UserID() string {
	return e.userID
}

type OrderSubmittedEvent struct {
	OrderEvent
	Symbol    string
	OrderSide OrderSide
	OrderType OrderType
	Quantity  float64
	Price     *float64
}

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

type OrderProcessingStartedEvent struct {
	OrderEvent
	MarketPrice     *float64
	MarketTimestamp *time.Time
}

func NewOrderProcessingStartedEvent(orderID, userID string, marketPrice *float64, marketTimestamp *time.Time) *OrderProcessingStartedEvent {
	return &OrderProcessingStartedEvent{
		OrderEvent:      NewOrderEvent("OrderProcessingStarted", orderID, userID),
		MarketPrice:     marketPrice,
		MarketTimestamp: marketTimestamp,
	}
}

type OrderExecutedEvent struct {
	OrderEvent
	Symbol              string
	OrderSide           OrderSide
	OrderType           OrderType
	Quantity            float64
	ExecutionPrice      float64
	ExecutedAt          time.Time
	TotalValue          float64
	MarketPriceAtExec   *float64
	MarketDataTimestamp *time.Time
}

// NewOrderExecutedEvent creates a new OrderExecutedEvent with position-relevant data
func NewOrderExecutedEvent(
	orderID, userID, symbol string,
	orderSide OrderSide,
	orderType OrderType,
	quantity, executionPrice, totalValue float64,
	executedAt time.Time,
	marketPriceAtExec *float64,
	marketDataTimestamp *time.Time,
) *OrderExecutedEvent {
	return &OrderExecutedEvent{
		OrderEvent:          NewOrderEvent("OrderExecuted", orderID, userID),
		Symbol:              symbol,
		OrderSide:           orderSide,
		OrderType:           orderType,
		Quantity:            quantity,
		ExecutionPrice:      executionPrice,
		ExecutedAt:          executedAt,
		TotalValue:          totalValue,
		MarketPriceAtExec:   marketPriceAtExec,
		MarketDataTimestamp: marketDataTimestamp,
	}
}

func (e *OrderExecutedEvent) IsPositionRelevant() bool {
	return true // All executed orders affect positions
}

func (e *OrderExecutedEvent) IsBuyOrder() bool {
	return e.OrderSide.IsBuy()
}

func (e *OrderExecutedEvent) IsSellOrder() bool {
	return e.OrderSide.IsSell()
}

type OrderFailedEvent struct {
	OrderEvent
	FailureReason string
	FailedAt      time.Time
}

func NewOrderFailedEvent(orderID, userID, failureReason string, failedAt time.Time) *OrderFailedEvent {
	return &OrderFailedEvent{
		OrderEvent:    NewOrderEvent("OrderFailed", orderID, userID),
		FailureReason: failureReason,
		FailedAt:      failedAt,
	}
}

type OrderCancelledEvent struct {
	OrderEvent
	CancelledAt  time.Time
	CancelReason string
	CancelledBy  string // user ID who cancelled (could be admin)
}

func NewOrderCancelledEvent(orderID, userID, cancelReason, cancelledBy string, cancelledAt time.Time) *OrderCancelledEvent {
	return &OrderCancelledEvent{
		OrderEvent:   NewOrderEvent("OrderCancelled", orderID, userID),
		CancelledAt:  cancelledAt,
		CancelReason: cancelReason,
		CancelledBy:  cancelledBy,
	}
}

type MarketDataReceivedEvent struct {
	OrderEvent
	Symbol              string
	MarketPrice         float64
	MarketDataTimestamp time.Time
	DataSource          string
}

func NewMarketDataReceivedEvent(orderID, userID, symbol, dataSource string, marketPrice float64, marketDataTimestamp time.Time) *MarketDataReceivedEvent {
	return &MarketDataReceivedEvent{
		OrderEvent:          NewOrderEvent("MarketDataReceived", orderID, userID),
		Symbol:              symbol,
		MarketPrice:         marketPrice,
		MarketDataTimestamp: marketDataTimestamp,
		DataSource:          dataSource,
	}
}

type OrderStatusChangedEvent struct {
	OrderEvent
	FromStatus OrderStatus
	ToStatus   OrderStatus
	ChangedAt  time.Time
	Reason     string
}

func NewOrderStatusChangedEvent(orderID, userID string, fromStatus, toStatus OrderStatus, reason string, changedAt time.Time) *OrderStatusChangedEvent {
	return &OrderStatusChangedEvent{
		OrderEvent: NewOrderEvent("OrderStatusChanged", orderID, userID),
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		ChangedAt:  changedAt,
		Reason:     reason,
	}
}

type RiskCheckPerformedEvent struct {
	OrderEvent
	RiskCheckType   string
	RiskCheckResult string // PASSED, FAILED, WARNING
	RiskScore       float64
	CheckedAt       time.Time
}

func NewRiskCheckPerformedEvent(orderID, userID, riskCheckType, result string, riskScore float64, checkedAt time.Time) *RiskCheckPerformedEvent {
	return &RiskCheckPerformedEvent{
		OrderEvent:      NewOrderEvent("RiskCheckPerformed", orderID, userID),
		RiskCheckType:   riskCheckType,
		RiskCheckResult: result,
		RiskScore:       riskScore,
		CheckedAt:       checkedAt,
	}
}

type OrderValidationFailedEvent struct {
	OrderEvent
	ValidationErrors []string
	ValidatedAt      time.Time
}

func NewOrderValidationFailedEvent(orderID, userID string, validationErrors []string, validatedAt time.Time) *OrderValidationFailedEvent {
	return &OrderValidationFailedEvent{
		OrderEvent:       NewOrderEvent("OrderValidationFailed", orderID, userID),
		ValidationErrors: validationErrors,
		ValidatedAt:      validatedAt,
	}
}

type PositionValidationFailedEvent struct {
	OrderEvent
	Symbol            string
	RequestedQuantity float64
	AvailableQuantity float64
	ValidationError   string
}

func NewPositionValidationFailedEvent(orderID, userID, symbol string, requestedQuantity, availableQuantity float64, validationError string) *PositionValidationFailedEvent {
	return &PositionValidationFailedEvent{
		OrderEvent:        NewOrderEvent("PositionValidationFailed", orderID, userID),
		Symbol:            symbol,
		RequestedQuantity: requestedQuantity,
		AvailableQuantity: availableQuantity,
		ValidationError:   validationError,
	}
}
