package domain

import "fmt"

// OrderType represents the type of an order
// @Description Order type enumeration
type OrderType string

const (
	// OrderTypeMarket represents a market order (execute immediately at current market price)
	OrderTypeMarket OrderType = "MARKET"

	// OrderTypeLimit represents a limit order (execute only at specified price or better)
	OrderTypeLimit OrderType = "LIMIT"

	// OrderTypeStopLoss represents a stop loss order (trigger when price reaches stop price)
	OrderTypeStopLoss OrderType = "STOP_LOSS"

	// OrderTypeStopLimit represents a stop limit order (becomes limit order when stop price is reached)
	OrderTypeStopLimit OrderType = "STOP_LIMIT"
)

// AllOrderTypes returns all valid order types
func AllOrderTypes() []OrderType {
	return []OrderType{
		OrderTypeMarket,
		OrderTypeLimit,
		OrderTypeStopLoss,
		OrderTypeStopLimit,
	}
}

// IsValid checks if the order type is valid
func (t OrderType) IsValid() bool {
	switch t {
	case OrderTypeMarket, OrderTypeLimit, OrderTypeStopLoss, OrderTypeStopLimit:
		return true
	default:
		return false
	}
}

// String returns the string representation of the order type
func (t OrderType) String() string {
	return string(t)
}

// RequiresPrice checks if the order type requires a price to be specified
func (t OrderType) RequiresPrice() bool {
	switch t {
	case OrderTypeLimit, OrderTypeStopLoss, OrderTypeStopLimit:
		return true
	case OrderTypeMarket:
		return false
	default:
		return false
	}
}

// IsImmediateExecution checks if the order type should be executed immediately
func (t OrderType) IsImmediateExecution() bool {
	return t == OrderTypeMarket
}

// IsConditional checks if the order type is conditional (depends on market conditions)
func (t OrderType) IsConditional() bool {
	return t == OrderTypeStopLoss || t == OrderTypeStopLimit
}

// ParseOrderType parses a string into an OrderType
func ParseOrderType(s string) (OrderType, error) {
	orderType := OrderType(s)
	if !orderType.IsValid() {
		return "", fmt.Errorf("invalid order type: %s", s)
	}
	return orderType, nil
}

// GetOrderTypeDescription returns a human-readable description of the order type
func (t OrderType) GetDescription() string {
	switch t {
	case OrderTypeMarket:
		return "Execute immediately at current market price"
	case OrderTypeLimit:
		return "Execute only at specified price or better"
	case OrderTypeStopLoss:
		return "Trigger when price reaches specified stop price"
	case OrderTypeStopLimit:
		return "Becomes limit order when stop price is reached"
	default:
		return "Unknown order type"
	}
}

// GetExecutionPriority returns the execution priority for the order type
// Lower numbers indicate higher priority
func (t OrderType) GetExecutionPriority() int {
	switch t {
	case OrderTypeMarket:
		return 1 // Highest priority - immediate execution
	case OrderTypeLimit:
		return 2 // Medium priority - price dependent
	case OrderTypeStopLoss:
		return 3 // Lower priority - conditional
	case OrderTypeStopLimit:
		return 4 // Lowest priority - conditional + price dependent
	default:
		return 5 // Unknown types get lowest priority
	}
}

// CanExecuteAtPrice checks if the order can be executed at the given market price
func (t OrderType) CanExecuteAtPrice(orderPrice, marketPrice *float64) bool {
	switch t {
	case OrderTypeMarket:
		return true // Market orders always execute
	case OrderTypeLimit:
		if orderPrice == nil {
			return false
		}
		// For simplicity, assuming buy orders (can be extended for sell orders)
		return marketPrice != nil && *orderPrice >= *marketPrice
	case OrderTypeStopLoss:
		if orderPrice == nil {
			return false
		}
		// Trigger when market price reaches or goes below stop price
		return marketPrice != nil && *marketPrice <= *orderPrice
	case OrderTypeStopLimit:
		// More complex logic - would need additional fields for stop price vs limit price
		return false // Not implemented in this simplified version
	default:
		return false
	}
}
