package command

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type ClosePositionCommand struct {
	PositionID    string  `json:"position_id" validate:"required"`
	UserID        string  `json:"user_id" validate:"required"`
	ClosePrice    float64 `json:"close_price" validate:"required,gt=0"`
	SourceOrderID *string `json:"source_order_id,omitempty"`
	CloseReason   string  `json:"close_reason,omitempty"`   // e.g., "ORDER_EXECUTION", "MANUAL_CLOSE", "STOP_LOSS"
	ExecutionTime *string `json:"execution_time,omitempty"` // ISO 8601 format
}

func (cmd *ClosePositionCommand) Validate() error {
	if cmd.PositionID == "" {
		return errors.New("position ID is required")
	}

	if _, err := uuid.Parse(cmd.PositionID); err != nil {
		return fmt.Errorf("invalid position ID format: %w", err)
	}

	if cmd.UserID == "" {
		return errors.New("user ID is required")
	}

	if _, err := parseUserIDToUUID(cmd.UserID); err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	if cmd.ClosePrice <= 0 {
		return errors.New("close price must be positive")
	}

	if cmd.SourceOrderID != nil && *cmd.SourceOrderID != "" {
		if _, err := uuid.Parse(*cmd.SourceOrderID); err != nil {
			return fmt.Errorf("invalid source order ID format: %w", err)
		}
	}

	if cmd.CloseReason == "" {
		if cmd.SourceOrderID != nil {
			cmd.CloseReason = "ORDER_EXECUTION"
		} else {
			cmd.CloseReason = "MANUAL_CLOSE"
		}
	}

	return nil
}

func (cmd *ClosePositionCommand) ToPositionID() (uuid.UUID, error) {
	return uuid.Parse(cmd.PositionID)
}

func (cmd *ClosePositionCommand) ToUserID() (uuid.UUID, error) {
	return parseUserIDToUUID(cmd.UserID)
}

func (cmd *ClosePositionCommand) ToSourceOrderID() (*uuid.UUID, error) {
	if cmd.SourceOrderID == nil || *cmd.SourceOrderID == "" {
		return nil, nil
	}

	orderID, err := uuid.Parse(*cmd.SourceOrderID)
	if err != nil {
		return nil, err
	}

	return &orderID, nil
}

func (cmd *ClosePositionCommand) GetDescription() string {
	return fmt.Sprintf("Close position %s at $%.2f (%s)",
		cmd.PositionID, cmd.ClosePrice, cmd.CloseReason)
}
