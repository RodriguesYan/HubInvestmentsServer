package command

import (
	"errors"
	"fmt"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// SubmitOrderCommand represents a command to submit a new order
// @Description Command object for order submission with validation
type SubmitOrderCommand struct {
	UserID    string   `json:"user_id" validate:"required"`
	Symbol    string   `json:"symbol" validate:"required"`
	OrderSide string   `json:"order_side" validate:"required,oneof=BUY SELL"`
	OrderType string   `json:"order_type" validate:"required,oneof=MARKET LIMIT STOP_LOSS STOP_LIMIT"`
	Quantity  float64  `json:"quantity" validate:"required,gt=0"`
	Price     *float64 `json:"price,omitempty"` // Optional for market orders
}

// SubmitOrderResult represents the result of a successful order submission
type SubmitOrderResult struct {
	OrderID                 string   `json:"order_id"`
	Status                  string   `json:"status"`
	MarketPriceAtSubmission *float64 `json:"market_price_at_submission,omitempty"`
	EstimatedExecutionPrice *float64 `json:"estimated_execution_price,omitempty"`
	Message                 string   `json:"message"`
}

// Validate validates the submit order command
func (cmd *SubmitOrderCommand) Validate() error {
	if cmd.UserID == "" {
		return errors.New("user ID is required")
	}

	if cmd.Symbol == "" {
		return errors.New("symbol is required")
	}

	if cmd.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	// Validate order side
	_, err := domain.ParseOrderSide(cmd.OrderSide)
	if err != nil {
		return fmt.Errorf("invalid order side: %w", err)
	}

	// Validate order type
	orderType, err := domain.ParseOrderType(cmd.OrderType)
	if err != nil {
		return fmt.Errorf("invalid order type: %w", err)
	}

	// Validate price requirements based on order type
	if orderType.RequiresPrice() && cmd.Price == nil {
		return fmt.Errorf("%s orders require a price", cmd.OrderType)
	}

	if orderType == domain.OrderTypeMarket && cmd.Price != nil {
		return errors.New("market orders cannot have a price")
	}

	if cmd.Price != nil && *cmd.Price <= 0 {
		return errors.New("price must be positive")
	}

	return nil
}

// ToOrderSide converts the string order side to domain OrderSide
func (cmd *SubmitOrderCommand) ToOrderSide() (domain.OrderSide, error) {
	return domain.ParseOrderSide(cmd.OrderSide)
}

// ToOrderType converts the string order type to domain OrderType
func (cmd *SubmitOrderCommand) ToOrderType() (domain.OrderType, error) {
	return domain.ParseOrderType(cmd.OrderType)
}

// GetDescription returns a human-readable description of the order
func (cmd *SubmitOrderCommand) GetDescription() string {
	priceStr := "market price"
	if cmd.Price != nil {
		priceStr = fmt.Sprintf("$%.2f", *cmd.Price)
	}

	return fmt.Sprintf("%s %s order for %.2f shares of %s at %s",
		cmd.OrderSide, cmd.OrderType, cmd.Quantity, cmd.Symbol, priceStr)
}

// IsMarketOrder checks if this is a market order
func (cmd *SubmitOrderCommand) IsMarketOrder() bool {
	return cmd.OrderType == "MARKET"
}

// IsBuyOrder checks if this is a buy order
func (cmd *SubmitOrderCommand) IsBuyOrder() bool {
	return cmd.OrderSide == "BUY"
}

// IsSellOrder checks if this is a sell order
func (cmd *SubmitOrderCommand) IsSellOrder() bool {
	return cmd.OrderSide == "SELL"
}
