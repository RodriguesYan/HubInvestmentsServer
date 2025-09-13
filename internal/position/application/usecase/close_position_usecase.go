package usecase

import (
	"context"
	"fmt"
	"time"

	"HubInvestments/internal/position/application/command"
	domain "HubInvestments/internal/position/domain/model"
	"HubInvestments/internal/position/domain/repository"
)

type IClosePositionUseCase interface {
	Execute(ctx context.Context, cmd *command.ClosePositionCommand) (*command.ClosePositionResult, error)
}

type ClosePositionUseCase struct {
	positionRepository repository.IPositionRepository
}

type ClosePositionUseCaseConfig struct {
	ValidationTimeout        time.Duration // Timeout for validation operations
	EnableBusinessValidation bool          // Whether to perform additional business validation
	RequireCloseReason       bool          // Whether close reason is mandatory
	RequireOrderTracking     bool          // Whether source order ID is required
}

func NewClosePositionUseCase(
	positionRepository repository.IPositionRepository,
) IClosePositionUseCase {
	return &ClosePositionUseCase{
		positionRepository: positionRepository,
	}
}

func (uc *ClosePositionUseCase) Execute(ctx context.Context, cmd *command.ClosePositionCommand) (*command.ClosePositionResult, error) {
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

	if !position.CanBeClosed() {
		return nil, fmt.Errorf("position cannot be closed: status is %s or quantity is zero", position.Status)
	}

	originalQuantity := position.Quantity
	originalAveragePrice := position.AveragePrice
	originalTotalInvestment := position.TotalInvestment
	positionOpenedAt := position.CreatedAt

	if err := uc.validateBusinessRules(ctx, position, cmd); err != nil {
		return nil, fmt.Errorf("business validation failed: %w", err)
	}

	var sourceOrderIDPtr *string
	if sourceOrderID, err := cmd.ToSourceOrderID(); err != nil {
		return nil, fmt.Errorf("invalid source order ID: %w", err)
	} else if sourceOrderID != nil {
		sourceOrderIDStr := sourceOrderID.String()
		sourceOrderIDPtr = &sourceOrderIDStr
	}

	position.ClearEvents()

	err = position.UpdateQuantityWithOrderID(originalQuantity, cmd.ClosePrice, false, sourceOrderIDPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to close position: %w", err)
	}

	if position.Status != domain.PositionStatusClosed {
		return nil, fmt.Errorf("position was not properly closed: final status is %s", position.Status)
	}

	if err := position.Validate(); err != nil {
		return nil, fmt.Errorf("position validation failed after closure: %w", err)
	}

	if err := uc.positionRepository.Update(ctx, position); err != nil {
		return nil, fmt.Errorf("failed to save closed position: %w", err)
	}

	totalRealizedValue := originalQuantity * cmd.ClosePrice
	realizedPnL := totalRealizedValue - originalTotalInvestment
	realizedPnLPct := (realizedPnL / originalTotalInvestment) * 100
	holdingPeriod := time.Since(positionOpenedAt)
	holdingPeriodDays := holdingPeriod.Hours() / 24

	eventsPublished := len(position.GetEvents())

	baseResult := &command.UpdatePositionResult{
		PositionID:         position.ID.String(),
		NewQuantity:        0,                    // Position is closed
		NewAveragePrice:    originalAveragePrice, // Preserved for historical reference
		NewTotalInvestment: 0,                    // No investment remains
		Status:             string(position.Status),
		TransactionType:    "SELL",
		RealizedPnL:        &realizedPnL,
		RealizedPnLPct:     &realizedPnLPct,
		EventsPublished:    eventsPublished,
		Message:            uc.buildCloseMessage(originalQuantity, cmd.ClosePrice, cmd.CloseReason, holdingPeriodDays, realizedPnL),
	}

	result := &command.ClosePositionResult{
		UpdatePositionResult: baseResult,
		HoldingPeriodDays:    holdingPeriodDays,
		TotalRealizedValue:   totalRealizedValue,
		FinalSellPrice:       cmd.ClosePrice,
		PositionClosedAt:     position.UpdatedAt.Format(time.RFC3339),
	}

	return result, nil
}

func (uc *ClosePositionUseCase) buildCloseMessage(quantity, price float64, closeReason string,
	holdingDays, realizedPnL float64) string {

	pnlStatus := "loss"
	if realizedPnL > 0 {
		pnlStatus = "profit"
	} else if realizedPnL == 0 {
		pnlStatus = "break-even"
	}

	return fmt.Sprintf("Position closed: sold %.2f shares at $%.2f after %.1f days (%s) - %s of $%.2f",
		quantity, price, holdingDays, closeReason, pnlStatus, realizedPnL)
}

func (uc *ClosePositionUseCase) validateBusinessRules(ctx context.Context, position *domain.Position,
	cmd *command.ClosePositionCommand) error {

	// Prevent premature stop-loss triggers due to market noise
	if cmd.CloseReason == "STOP_LOSS" {
		minHoldingPeriod := 1 * time.Hour // Example: must hold for at least 1 hour before stop loss
		if time.Since(position.CreatedAt) < minHoldingPeriod {
			return fmt.Errorf("position must be held for at least %v before stop loss can be triggered", minHoldingPeriod)
		}
	}

	// Sanity check: prevent closing at obviously wrong prices
	if position.AveragePrice > 0 {
		maxDeviationBelow := 0.90 // Allow up to 90% loss
		if cmd.ClosePrice < position.AveragePrice*maxDeviationBelow {
			return fmt.Errorf("close price $%.2f is unreasonably low (%.1f%% below average price $%.2f)",
				cmd.ClosePrice, (1-cmd.ClosePrice/position.AveragePrice)*100, position.AveragePrice)
		}
	}

	// Prevent closure of dust positions that cost more in fees
	const minClosureValue = 0.01 // $0.01 minimum
	totalValue := position.Quantity * cmd.ClosePrice
	if totalValue < minClosureValue {
		return fmt.Errorf("total closure value $%.4f is below minimum of $%.2f", totalValue, minClosureValue)
	}

	return nil
}

func (uc *ClosePositionUseCase) calculateClosureMetrics(position *domain.Position, closePrice float64) map[string]interface{} {
	metrics := make(map[string]interface{})

	metrics["position_id"] = position.ID.String()
	metrics["symbol"] = position.Symbol
	metrics["user_id"] = position.UserID.String()

	originalQuantity := position.Quantity
	metrics["quantity_closed"] = originalQuantity
	metrics["average_price"] = position.AveragePrice
	metrics["close_price"] = closePrice

	totalInvestment := position.TotalInvestment
	totalRealizedValue := originalQuantity * closePrice
	realizedPnL := totalRealizedValue - totalInvestment

	metrics["total_investment"] = totalInvestment
	metrics["total_realized_value"] = totalRealizedValue
	metrics["realized_pnl"] = realizedPnL

	if totalInvestment > 0 {
		metrics["realized_pnl_pct"] = (realizedPnL / totalInvestment) * 100
	}

	holdingPeriod := time.Since(position.CreatedAt)
	metrics["holding_period_hours"] = holdingPeriod.Hours()
	metrics["holding_period_days"] = holdingPeriod.Hours() / 24
	metrics["position_opened_at"] = position.CreatedAt
	metrics["position_closed_at"] = time.Now()

	metrics["position_type"] = string(position.PositionType)
	metrics["final_status"] = string(position.Status)

	if position.AveragePrice > 0 {
		metrics["price_return_pct"] = ((closePrice - position.AveragePrice) / position.AveragePrice) * 100
	}

	return metrics
}

func (uc *ClosePositionUseCase) generateClosureReport(position *domain.Position, cmd *command.ClosePositionCommand) string {
	metrics := uc.calculateClosureMetrics(position, cmd.ClosePrice)

	report := fmt.Sprintf(`
Position Closure Report
======================
Position ID: %s
Symbol: %s
Quantity Closed: %.6f shares
Average Price: $%.2f
Close Price: $%.2f
Total Investment: $%.2f
Total Realized: $%.2f
Realized P&L: $%.2f (%.2f%%)
Holding Period: %.1f days
Close Reason: %s
`,
		position.ID.String(),
		position.Symbol,
		metrics["quantity_closed"],
		metrics["average_price"],
		metrics["close_price"],
		metrics["total_investment"],
		metrics["total_realized_value"],
		metrics["realized_pnl"],
		metrics["realized_pnl_pct"],
		metrics["holding_period_days"],
		cmd.CloseReason,
	)

	return report
}
