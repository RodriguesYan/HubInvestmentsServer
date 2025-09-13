package usecase

import (
	"context"
	"fmt"
	"time"

	"HubInvestments/internal/position/application/command"
	domain "HubInvestments/internal/position/domain/model"
	"HubInvestments/internal/position/domain/repository"
)

type IUpdatePositionUseCase interface {
	Execute(ctx context.Context, cmd *command.UpdatePositionCommand) (*command.UpdatePositionResult, error)
}

type UpdatePositionUseCase struct {
	positionRepository repository.IPositionRepository
}

type UpdatePositionUseCaseConfig struct {
	ValidationTimeout        time.Duration // Timeout for validation operations
	EnableBusinessValidation bool          // Whether to perform additional business validation
	AllowPartialSales        bool          // Whether to allow partial position sales
	RequireOrderTracking     bool          // Whether source order ID is required
}

func NewUpdatePositionUseCase(
	positionRepository repository.IPositionRepository,
) IUpdatePositionUseCase {
	return &UpdatePositionUseCase{
		positionRepository: positionRepository,
	}
}

func (uc *UpdatePositionUseCase) Execute(ctx context.Context, cmd *command.UpdatePositionCommand) (*command.UpdatePositionResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	positionID, err := cmd.ToPositionID()
	if err != nil {
		return nil, fmt.Errorf("invalid position ID: %w", err)
	}

	userID, err := cmd.ToUserID()
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	position, err := uc.positionRepository.FindByID(ctx, positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find position: %w", err)
	}

	if position == nil {
		return nil, fmt.Errorf("position not found: %s", cmd.PositionID)
	}

	if position.UserID != userID {
		return nil, fmt.Errorf("position does not belong to user %s", cmd.UserID)
	}

	if !position.Status.CanBeUpdated() {
		return nil, fmt.Errorf("position cannot be updated: status is %s", position.Status)
	}

	previousQuantity := position.Quantity
	previousAveragePrice := position.AveragePrice

	if cmd.IsSellOrder() {
		if !position.CanSell(cmd.TradeQuantity) {
			return nil, fmt.Errorf("insufficient quantity to sell: available %.6f, requested %.6f",
				position.Quantity, cmd.TradeQuantity)
		}
	}

	var sourceOrderIDPtr *string
	if sourceOrderID, err := cmd.ToSourceOrderID(); err != nil {
		return nil, fmt.Errorf("invalid source order ID: %w", err)
	} else if sourceOrderID != nil {
		sourceOrderIDStr := sourceOrderID.String()
		sourceOrderIDPtr = &sourceOrderIDStr
	}

	eventsBeforeUpdate := len(position.GetEvents())
	position.ClearEvents()

	err = position.UpdateQuantityWithOrderID(cmd.TradeQuantity, cmd.TradePrice, cmd.IsBuyOrder, sourceOrderIDPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	if err := position.Validate(); err != nil {
		return nil, fmt.Errorf("position validation failed after update: %w", err)
	}

	if err := uc.positionRepository.Update(ctx, position); err != nil {
		return nil, fmt.Errorf("failed to save updated position: %w", err)
	}

	var realizedPnL *float64
	var realizedPnLPct *float64
	if cmd.IsSellOrder() {
		pnl := (cmd.TradePrice - previousAveragePrice) * cmd.TradeQuantity
		realizedPnL = &pnl

		if previousAveragePrice > 0 {
			pnlPct := (pnl / (previousAveragePrice * cmd.TradeQuantity)) * 100
			realizedPnLPct = &pnlPct
		}
	}

	eventsAfterUpdate := len(position.GetEvents())
	eventsPublished := eventsAfterUpdate - eventsBeforeUpdate + eventsBeforeUpdate

	result := &command.UpdatePositionResult{
		PositionID:         position.ID.String(),
		NewQuantity:        position.Quantity,
		NewAveragePrice:    position.AveragePrice,
		NewTotalInvestment: position.TotalInvestment,
		Status:             string(position.Status),
		TransactionType:    cmd.GetTransactionType(),
		RealizedPnL:        realizedPnL,
		RealizedPnLPct:     realizedPnLPct,
		EventsPublished:    eventsPublished,
		Message:            uc.buildUpdateMessage(cmd, previousQuantity, position.Quantity),
	}

	if position.Status == domain.PositionStatusClosed {
		return uc.buildClosePositionResult(result, position, cmd), nil
	}

	return result, nil
}

func (uc *UpdatePositionUseCase) buildUpdateMessage(cmd *command.UpdatePositionCommand,
	prevQuantity, newQuantity float64) string {

	if cmd.IsBuyOrder {
		return fmt.Sprintf("Position updated: bought %.2f shares at $%.2f, new quantity: %.2f",
			cmd.TradeQuantity, cmd.TradePrice, newQuantity)
	}

	if newQuantity == 0 {
		return fmt.Sprintf("Position closed: sold %.2f shares at $%.2f",
			cmd.TradeQuantity, cmd.TradePrice)
	}

	return fmt.Sprintf("Position updated: sold %.2f shares at $%.2f, remaining quantity: %.2f",
		cmd.TradeQuantity, cmd.TradePrice, newQuantity)
}

func (uc *UpdatePositionUseCase) buildClosePositionResult(baseResult *command.UpdatePositionResult,
	position *domain.Position, cmd *command.UpdatePositionCommand) *command.UpdatePositionResult {

	holdingPeriod := time.Since(position.CreatedAt)
	holdingPeriodDays := holdingPeriod.Hours() / 24

	// TODO: Consider using separate ClosePositionResult type
	baseResult.Message = fmt.Sprintf("Position fully closed after %.1f days. %s",
		holdingPeriodDays, baseResult.Message)

	return baseResult
}

func (uc *UpdatePositionUseCase) validateBusinessRules(ctx context.Context, position *domain.Position,
	cmd *command.UpdatePositionCommand) error {

	// Enforce minimum trade value to prevent micro-transactions
	const minTradeValue = 1.0 // $1 minimum trade
	tradeValue := cmd.CalculateTradeValue()
	if tradeValue < minTradeValue {
		return fmt.Errorf("trade value $%.2f is below minimum of $%.2f", tradeValue, minTradeValue)
	}

	// Prevent obvious fat-finger errors on sell orders
	if cmd.IsSellOrder() {
		maxAllowedDeviation := 0.50 // 50%
		priceDeviation := (cmd.TradePrice - position.AveragePrice) / position.AveragePrice

		if priceDeviation < -maxAllowedDeviation {
			return fmt.Errorf("sell price $%.2f is more than %.0f%% below average price $%.2f",
				cmd.TradePrice, maxAllowedDeviation*100, position.AveragePrice)
		}
	}

	return nil
}

func (uc *UpdatePositionUseCase) calculatePositionMetrics(position *domain.Position) map[string]interface{} {
	metrics := make(map[string]interface{})

	metrics["quantity"] = position.Quantity
	metrics["average_price"] = position.AveragePrice
	metrics["total_investment"] = position.TotalInvestment
	metrics["status"] = string(position.Status)
	metrics["updated_at"] = position.UpdatedAt

	if position.LastTradeAt != nil {
		metrics["last_trade_at"] = *position.LastTradeAt
	}

	// Calculate unrealized P&L if current price is available
	if position.CurrentPrice > 0 {
		metrics["current_price"] = position.CurrentPrice
		metrics["market_value"] = position.MarketValue
		metrics["unrealized_pnl"] = position.UnrealizedPnL
		metrics["unrealized_pnl_pct"] = position.UnrealizedPnLPct
	}

	return metrics
}
