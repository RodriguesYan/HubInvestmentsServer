package command

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type UpdatePositionCommand struct {
	PositionID    string  `json:"position_id" validate:"required"`
	UserID        string  `json:"user_id" validate:"required"`
	TradeQuantity float64 `json:"trade_quantity" validate:"required,gt=0"`
	TradePrice    float64 `json:"trade_price" validate:"required,gt=0"`
	IsBuyOrder    bool    `json:"is_buy_order"`
	SourceOrderID *string `json:"source_order_id,omitempty"`
	ExecutionTime *string `json:"execution_time,omitempty"` // ISO 8601 format
}

type UpdatePositionResult struct {
	PositionID         string   `json:"position_id"`
	NewQuantity        float64  `json:"new_quantity"`
	NewAveragePrice    float64  `json:"new_average_price"`
	NewTotalInvestment float64  `json:"new_total_investment"`
	Status             string   `json:"status"`
	TransactionType    string   `json:"transaction_type"`
	RealizedPnL        *float64 `json:"realized_pnl,omitempty"`     // Only for SELL orders
	RealizedPnLPct     *float64 `json:"realized_pnl_pct,omitempty"` // Only for SELL orders
	Message            string   `json:"message"`
	EventsPublished    int      `json:"events_published"`
}

type ClosePositionResult struct {
	*UpdatePositionResult
	HoldingPeriodDays  float64 `json:"holding_period_days"`
	TotalRealizedValue float64 `json:"total_realized_value"`
	FinalSellPrice     float64 `json:"final_sell_price"`
	PositionClosedAt   string  `json:"position_closed_at"` // ISO 8601 format
}

func (cmd *UpdatePositionCommand) Validate() error {
	if cmd.PositionID == "" {
		return errors.New("position ID is required")
	}

	if _, err := uuid.Parse(cmd.PositionID); err != nil {
		return fmt.Errorf("invalid position ID format: %w", err)
	}

	if cmd.UserID == "" {
		return errors.New("user ID is required")
	}

	if _, err := uuid.Parse(cmd.UserID); err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	if cmd.TradeQuantity <= 0 {
		return errors.New("trade quantity must be positive")
	}

	if cmd.TradePrice <= 0 {
		return errors.New("trade price must be positive")
	}

	if cmd.SourceOrderID != nil && *cmd.SourceOrderID != "" {
		if _, err := uuid.Parse(*cmd.SourceOrderID); err != nil {
			return fmt.Errorf("invalid source order ID format: %w", err)
		}
	}

	return nil
}

func (cmd *UpdatePositionCommand) ToPositionID() (uuid.UUID, error) {
	return uuid.Parse(cmd.PositionID)
}

func (cmd *UpdatePositionCommand) ToUserID() (uuid.UUID, error) {
	return uuid.Parse(cmd.UserID)
}

func (cmd *UpdatePositionCommand) ToSourceOrderID() (*uuid.UUID, error) {
	if cmd.SourceOrderID == nil || *cmd.SourceOrderID == "" {
		return nil, nil
	}

	orderID, err := uuid.Parse(*cmd.SourceOrderID)
	if err != nil {
		return nil, err
	}

	return &orderID, nil
}

func (cmd *UpdatePositionCommand) GetTransactionType() string {
	if cmd.IsBuyOrder {
		return "BUY"
	}
	return "SELL"
}

func (cmd *UpdatePositionCommand) GetDescription() string {
	transactionType := cmd.GetTransactionType()
	return fmt.Sprintf("%s %.2f shares at $%.2f for position %s",
		transactionType, cmd.TradeQuantity, cmd.TradePrice, cmd.PositionID)
}

func (cmd *UpdatePositionCommand) CalculateTradeValue() float64 {
	return cmd.TradeQuantity * cmd.TradePrice
}

func (cmd *UpdatePositionCommand) IsSellOrder() bool {
	return !cmd.IsBuyOrder
}
