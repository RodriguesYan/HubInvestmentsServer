package command

import (
	domain "HubInvestments/internal/position/domain/model"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type CreatePositionCommand struct {
	UserID        string  `json:"user_id" validate:"required"`
	Symbol        string  `json:"symbol" validate:"required"`
	Quantity      float64 `json:"quantity" validate:"required,gt=0"`
	Price         float64 `json:"price" validate:"required,gt=0"`
	PositionType  string  `json:"position_type" validate:"required,oneof=LONG SHORT"`
	SourceOrderID *string `json:"source_order_id,omitempty"`
	CreatedFrom   string  `json:"created_from,omitempty"` // e.g., "ORDER_EXECUTION", "MANUAL_ENTRY"
}

type CreatePositionResult struct {
	PositionID      string  `json:"position_id"`
	Status          string  `json:"status"`
	TotalInvestment float64 `json:"total_investment"`
	Message         string  `json:"message"`
}

func (cmd *CreatePositionCommand) Validate() error {
	if cmd.UserID == "" {
		return errors.New("user ID is required")
	}

	if _, err := uuid.Parse(cmd.UserID); err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	if cmd.Symbol == "" {
		return errors.New("symbol is required")
	}

	if cmd.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	if cmd.Price <= 0 {
		return errors.New("price must be positive")
	}

	_, err := domain.NewPositionType(cmd.PositionType)
	if err != nil {
		return fmt.Errorf("invalid position type: %w", err)
	}

	if cmd.SourceOrderID != nil && *cmd.SourceOrderID != "" {
		if _, err := uuid.Parse(*cmd.SourceOrderID); err != nil {
			return fmt.Errorf("invalid source order ID format: %w", err)
		}
	}

	if cmd.CreatedFrom == "" {
		if cmd.SourceOrderID != nil {
			cmd.CreatedFrom = "ORDER_EXECUTION"
		} else {
			cmd.CreatedFrom = "MANUAL_ENTRY"
		}
	}

	return nil
}

func (cmd *CreatePositionCommand) ToPositionType() (domain.PositionType, error) {
	return domain.NewPositionType(cmd.PositionType)
}

func (cmd *CreatePositionCommand) ToUserID() (uuid.UUID, error) {
	return uuid.Parse(cmd.UserID)
}

func (cmd *CreatePositionCommand) ToSourceOrderID() (*uuid.UUID, error) {
	if cmd.SourceOrderID == nil || *cmd.SourceOrderID == "" {
		return nil, nil
	}

	orderID, err := uuid.Parse(*cmd.SourceOrderID)
	if err != nil {
		return nil, err
	}

	return &orderID, nil
}

func (cmd *CreatePositionCommand) GetDescription() string {
	return fmt.Sprintf("Create %s position for %.2f shares of %s at $%.2f",
		cmd.PositionType, cmd.Quantity, cmd.Symbol, cmd.Price)
}

func (cmd *CreatePositionCommand) CalculateTotalInvestment() float64 {
	return cmd.Quantity * cmd.Price
}
