package domain

import "fmt"

// OrderStatus represents the status of an order
// @Description Order status enumeration
type OrderStatus string

const (
	// OrderStatusPending represents a newly created order waiting for processing
	OrderStatusPending OrderStatus = "PENDING"

	// OrderStatusProcessing represents an order currently being processed
	OrderStatusProcessing OrderStatus = "PROCESSING"

	// OrderStatusExecuted represents a successfully executed order
	OrderStatusExecuted OrderStatus = "EXECUTED"

	// OrderStatusFailed represents an order that failed to execute
	OrderStatusFailed OrderStatus = "FAILED"

	// OrderStatusCancelled represents a cancelled order
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// AllOrderStatuses returns all valid order statuses
func AllOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusPending,
		OrderStatusProcessing,
		OrderStatusExecuted,
		OrderStatusFailed,
		OrderStatusCancelled,
	}
}

// IsValid checks if the order status is valid
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusProcessing, OrderStatusExecuted, OrderStatusFailed, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation of the order status
func (s OrderStatus) String() string {
	return string(s)
}

// IsTerminal checks if the order status is terminal (no further state changes allowed)
func (s OrderStatus) IsTerminal() bool {
	return s == OrderStatusExecuted || s == OrderStatusFailed || s == OrderStatusCancelled
}

// IsActive checks if the order is in an active state (can be processed or cancelled)
func (s OrderStatus) IsActive() bool {
	return s == OrderStatusPending || s == OrderStatusProcessing
}

// CanTransitionTo checks if transition to the target status is allowed
func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	// No transition from terminal states
	if s.IsTerminal() {
		return false
	}

	switch s {
	case OrderStatusPending:
		return target == OrderStatusProcessing || target == OrderStatusCancelled || target == OrderStatusFailed
	case OrderStatusProcessing:
		return target == OrderStatusExecuted || target == OrderStatusFailed || target == OrderStatusCancelled
	default:
		return false
	}
}

// ParseOrderStatus parses a string into an OrderStatus
func ParseOrderStatus(s string) (OrderStatus, error) {
	status := OrderStatus(s)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid order status: %s", s)
	}
	return status, nil
}

// GetOrderStatusDescription returns a human-readable description of the status
func (s OrderStatus) GetDescription() string {
	switch s {
	case OrderStatusPending:
		return "Order submitted and waiting for processing"
	case OrderStatusProcessing:
		return "Order is currently being processed"
	case OrderStatusExecuted:
		return "Order has been successfully executed"
	case OrderStatusFailed:
		return "Order execution failed"
	case OrderStatusCancelled:
		return "Order has been cancelled"
	default:
		return "Unknown status"
	}
}
