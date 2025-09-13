package usecase

import (
	"context"
	"fmt"
	"time"

	"HubInvestments/internal/position/application/command"
	domain "HubInvestments/internal/position/domain/model"
	"HubInvestments/internal/position/domain/repository"
)

type ICreatePositionUseCase interface {
	Execute(ctx context.Context, cmd *command.CreatePositionCommand) (*command.CreatePositionResult, error)
}

type CreatePositionUseCase struct {
	positionRepository repository.IPositionRepository
}

type CreatePositionUseCaseConfig struct {
	AllowDuplicatePositions  bool          // Whether to allow multiple positions for same symbol
	ValidationTimeout        time.Duration // Timeout for validation operations
	EnableBusinessValidation bool          // Whether to perform additional business validation
}

func NewCreatePositionUseCase(
	positionRepository repository.IPositionRepository,
) ICreatePositionUseCase {
	return &CreatePositionUseCase{
		positionRepository: positionRepository,
	}
}

func (uc *CreatePositionUseCase) Execute(ctx context.Context, cmd *command.CreatePositionCommand) (*command.CreatePositionResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	userID, err := cmd.ToUserID()
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	positionType, err := cmd.ToPositionType()
	if err != nil {
		return nil, fmt.Errorf("invalid position type: %w", err)
	}

	exists, err := uc.positionRepository.ExistsForUser(ctx, userID, cmd.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing position: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("position already exists for user %s and symbol %s", cmd.UserID, cmd.Symbol)
	}

	position, err := domain.NewPosition(userID, cmd.Symbol, cmd.Quantity, cmd.Price, positionType)
	if err != nil {
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	if err := position.Validate(); err != nil {
		return nil, fmt.Errorf("position validation failed: %w", err)
	}

	if err := uc.positionRepository.Save(ctx, position); err != nil {
		return nil, fmt.Errorf("failed to save position: %w", err)
	}

	result := &command.CreatePositionResult{
		PositionID:      position.ID.String(),
		Status:          string(position.Status),
		TotalInvestment: position.TotalInvestment,
		Message:         fmt.Sprintf("Position created successfully for %.2f shares of %s", cmd.Quantity, cmd.Symbol),
	}

	return result, nil
}

func (uc *CreatePositionUseCase) validateBusinessRules(ctx context.Context, cmd *command.CreatePositionCommand) error {
	userID, err := cmd.ToUserID()
	if err != nil {
		return err
	}

	// Prevent users from accumulating too many positions
	positionCount, err := uc.positionRepository.CountPositionsForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to count user positions: %w", err)
	}

	const maxPositionsPerUser = 50 // Example limit
	if positionCount >= maxPositionsPerUser {
		return fmt.Errorf("user has reached maximum position limit of %d", maxPositionsPerUser)
	}

	return nil
}

func (uc *CreatePositionUseCase) calculatePositionMetrics(position *domain.Position) map[string]interface{} {
	metrics := make(map[string]interface{})

	metrics["total_investment"] = position.TotalInvestment
	metrics["average_price"] = position.AveragePrice
	metrics["quantity"] = position.Quantity
	metrics["position_type"] = string(position.PositionType)
	metrics["status"] = string(position.Status)
	metrics["created_at"] = position.CreatedAt

	return metrics
}
