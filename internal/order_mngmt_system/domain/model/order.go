package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Order represents the aggregate root for order management
// @Description Order entity representing a trading order
type Order struct {
	id                      string
	userID                  string
	symbol                  string
	orderType               OrderType
	quantity                float64
	price                   *float64 // nil for market orders
	status                  OrderStatus
	createdAt               time.Time
	updatedAt               time.Time
	executedAt              *time.Time
	executionPrice          *float64
	marketPriceAtSubmission *float64
	marketDataTimestamp     *time.Time
}

// NewOrder creates a new order with generated UUID and PENDING status
func NewOrder(userID, symbol string, orderType OrderType, quantity float64, price *float64) (*Order, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	if orderType == OrderTypeLimit && (price == nil || quantity <= 0) {
		return nil, errors.New("limit orders must have a price and quantity")
	}
	if orderType == OrderTypeMarket && price != nil {
		return nil, errors.New("market orders cannot have a price")
	}

	now := time.Now()
	return &Order{
		id:        uuid.New().String(),
		userID:    userID,
		symbol:    symbol,
		orderType: orderType,
		quantity:  quantity,
		price:     price,
		status:    OrderStatusPending,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// NewOrderFromRepository creates an order from repository data (for reconstruction)
func NewOrderFromRepository(id, userID, symbol string, orderType OrderType, quantity float64, price *float64,
	status OrderStatus, createdAt, updatedAt time.Time, executedAt *time.Time,
	executionPrice, marketPriceAtSubmission *float64, marketDataTimestamp *time.Time) *Order {
	return &Order{
		id:                      id,
		userID:                  userID,
		symbol:                  symbol,
		orderType:               orderType,
		quantity:                quantity,
		price:                   price,
		status:                  status,
		createdAt:               createdAt,
		updatedAt:               updatedAt,
		executedAt:              executedAt,
		executionPrice:          executionPrice,
		marketPriceAtSubmission: marketPriceAtSubmission,
		marketDataTimestamp:     marketDataTimestamp,
	}
}

// Getters
func (o *Order) ID() string                        { return o.id }
func (o *Order) UserID() string                    { return o.userID }
func (o *Order) Symbol() string                    { return o.symbol }
func (o *Order) OrderType() OrderType              { return o.orderType }
func (o *Order) Quantity() float64                 { return o.quantity }
func (o *Order) Price() *float64                   { return o.price }
func (o *Order) Status() OrderStatus               { return o.status }
func (o *Order) CreatedAt() time.Time              { return o.createdAt }
func (o *Order) UpdatedAt() time.Time              { return o.updatedAt }
func (o *Order) ExecutedAt() *time.Time            { return o.executedAt }
func (o *Order) ExecutionPrice() *float64          { return o.executionPrice }
func (o *Order) MarketPriceAtSubmission() *float64 { return o.marketPriceAtSubmission }
func (o *Order) MarketDataTimestamp() *time.Time   { return o.marketDataTimestamp }

// Business Logic Methods

// CanCancel checks if the order can be cancelled
func (o *Order) CanCancel() bool {
	return o.status == OrderStatusPending || o.status == OrderStatusProcessing
}

// CanExecute checks if the order can be executed
func (o *Order) CanExecute() bool {
	return o.status == OrderStatusPending || o.status == OrderStatusProcessing
}

// IsPending checks if the order is in pending status
func (o *Order) IsPending() bool {
	return o.status == OrderStatusPending
}

// IsExecuted checks if the order has been executed
func (o *Order) IsExecuted() bool {
	return o.status == OrderStatusExecuted
}

// IsFailed checks if the order has failed
func (o *Order) IsFailed() bool {
	return o.status == OrderStatusFailed
}

// IsCancelled checks if the order has been cancelled
func (o *Order) IsCancelled() bool {
	return o.status == OrderStatusCancelled
}

// SetMarketDataContext sets market data information for the order
func (o *Order) SetMarketDataContext(marketPrice float64, timestamp time.Time) {
	o.marketPriceAtSubmission = &marketPrice
	o.marketDataTimestamp = &timestamp
	o.updatedAt = time.Now()
}

// MarkAsProcessing changes the order status to processing
func (o *Order) MarkAsProcessing() error {
	if !o.CanExecute() {
		return errors.New("order cannot be processed in current status")
	}
	o.status = OrderStatusProcessing
	o.updatedAt = time.Now()
	return nil
}

// MarkAsExecuted marks the order as executed with execution details
func (o *Order) MarkAsExecuted(executionPrice float64) error {
	if !o.CanExecute() {
		return errors.New("order cannot be executed in current status")
	}
	now := time.Now()
	o.status = OrderStatusExecuted
	o.executionPrice = &executionPrice
	o.executedAt = &now
	o.updatedAt = now
	return nil
}

// MarkAsFailed marks the order as failed
func (o *Order) MarkAsFailed() error {
	if o.status == OrderStatusExecuted {
		return errors.New("cannot fail an already executed order")
	}
	o.status = OrderStatusFailed
	o.updatedAt = time.Now()
	return nil
}

// MarkAsCancelled marks the order as cancelled
func (o *Order) MarkAsCancelled() error {
	if !o.CanCancel() {
		return errors.New("order cannot be cancelled in current status")
	}
	o.status = OrderStatusCancelled
	o.updatedAt = time.Now()
	return nil
}

// CalculateOrderValue calculates the total value of the order
func (o *Order) CalculateOrderValue() float64 {
	if o.price != nil {
		return *o.price * o.quantity
	}
	return 0 // Market orders don't have a predetermined value
}

// CalculateExecutionValue calculates the actual execution value
func (o *Order) CalculateExecutionValue() float64 {
	if o.executionPrice != nil {
		return *o.executionPrice * o.quantity
	}
	return 0
}

// GetPriceForExecution returns the price to use for execution
func (o *Order) GetPriceForExecution(currentMarketPrice float64) float64 {
	switch o.orderType {
	case OrderTypeMarket:
		return currentMarketPrice
	case OrderTypeLimit:
		if o.price != nil {
			return *o.price
		}
		return currentMarketPrice // fallback
	default:
		return currentMarketPrice
	}
}

// ValidateForExecution performs validation before order execution
func (o *Order) ValidateForExecution(currentMarketPrice float64) error {
	if !o.CanExecute() {
		return errors.New("order cannot be executed in current status")
	}

	// For limit orders, check if the limit price is reasonable
	if o.orderType == OrderTypeLimit && o.price != nil {
		// Allow some tolerance for price movements
		tolerance := 0.1 // 10% tolerance
		if *o.price > currentMarketPrice*(1+tolerance) {
			return errors.New("limit price too far above market price")
		}
		if *o.price < currentMarketPrice*(1-tolerance) {
			return errors.New("limit price too far below market price")
		}
	}

	return nil
}

// Validate performs comprehensive order validation
func (o *Order) Validate() error {
	if o.userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if o.symbol == "" {
		return errors.New("symbol cannot be empty")
	}
	if o.quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if o.orderType == OrderTypeLimit && o.price == nil {
		return errors.New("limit orders must have a price")
	}
	if o.orderType == OrderTypeLimit && o.price != nil && *o.price <= 0 {
		return errors.New("limit price must be positive")
	}
	return nil
}
