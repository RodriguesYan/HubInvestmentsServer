package command

import (
	"errors"
	"fmt"
)

// CancelOrderCommand represents a command to cancel an existing order
// @Description Command object for order cancellation with validation
type CancelOrderCommand struct {
	OrderID string `json:"order_id" validate:"required"`
	UserID  string `json:"user_id" validate:"required"`
	Reason  string `json:"reason,omitempty"` // Optional cancellation reason
}

// CancelOrderResult represents the result of a successful order cancellation
type CancelOrderResult struct {
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// CancellationReason represents predefined cancellation reasons
type CancellationReason string

const (
	CancellationReasonUserRequested     CancellationReason = "USER_REQUESTED"
	CancellationReasonMarketClosed      CancellationReason = "MARKET_CLOSED"
	CancellationReasonInsufficientFunds CancellationReason = "INSUFFICIENT_FUNDS"
	CancellationReasonRiskManagement    CancellationReason = "RISK_MANAGEMENT"
	CancellationReasonSystemError       CancellationReason = "SYSTEM_ERROR"
	CancellationReasonExpired           CancellationReason = "EXPIRED"
	CancellationReasonAdminAction       CancellationReason = "ADMIN_ACTION"
)

// Validate validates the cancel order command
func (cmd *CancelOrderCommand) Validate() error {
	if cmd.OrderID == "" {
		return errors.New("order ID is required")
	}

	if cmd.UserID == "" {
		return errors.New("user ID is required")
	}

	// Validate UUID format for order ID (basic check)
	if len(cmd.OrderID) < 36 {
		return errors.New("invalid order ID format")
	}

	return nil
}

// GetReason returns the cancellation reason or default
func (cmd *CancelOrderCommand) GetReason() CancellationReason {
	if cmd.Reason == "" {
		return CancellationReasonUserRequested
	}
	return CancellationReason(cmd.Reason)
}

// GetDescription returns a human-readable description of the cancellation
func (cmd *CancelOrderCommand) GetDescription() string {
	reason := cmd.GetReason()
	return fmt.Sprintf("Cancel order %s requested by user %s (reason: %s)",
		cmd.OrderID, cmd.UserID, reason)
}

// IsValidReason checks if the provided reason is valid
func (cmd *CancelOrderCommand) IsValidReason() bool {
	if cmd.Reason == "" {
		return true // Empty reason defaults to USER_REQUESTED
	}

	validReasons := []CancellationReason{
		CancellationReasonUserRequested,
		CancellationReasonMarketClosed,
		CancellationReasonInsufficientFunds,
		CancellationReasonRiskManagement,
		CancellationReasonSystemError,
		CancellationReasonExpired,
		CancellationReasonAdminAction,
	}

	reason := CancellationReason(cmd.Reason)
	for _, validReason := range validReasons {
		if reason == validReason {
			return true
		}
	}

	return false
}

// ValidateWithReason validates the command including the reason
func (cmd *CancelOrderCommand) ValidateWithReason() error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	if !cmd.IsValidReason() {
		return fmt.Errorf("invalid cancellation reason: %s", cmd.Reason)
	}

	return nil
}
