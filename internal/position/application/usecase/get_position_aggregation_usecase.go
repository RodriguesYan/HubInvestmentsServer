package usecase

import (
	domain "HubInvestments/internal/position/domain/model"
	repository "HubInvestments/internal/position/domain/repository"
	service "HubInvestments/internal/position/domain/service"
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

type GetPositionAggregationUseCase struct {
	repo               repository.PositionRepository
	aggregationService service.PositionAggregationService
}

func NewGetPositionAggregationUseCase(repo repository.PositionRepository) *GetPositionAggregationUseCase {
	return &GetPositionAggregationUseCase{
		repo:               repo,
		aggregationService: service.NewPositionAggregationService(),
	}
}

// NewGetPositionAggregationUseCaseWithService allows dependency injection of the aggregation service for testing
func NewGetPositionAggregationUseCaseWithService(repo repository.PositionRepository, aggregationService service.PositionAggregationService) *GetPositionAggregationUseCase {
	return &GetPositionAggregationUseCase{
		repo:               repo,
		aggregationService: aggregationService,
	}
}

func (uc *GetPositionAggregationUseCase) Execute(userId string) (domain.AucAggregationModel, error) {
	userUUID, err := parseUserIDToUUID(userId)
	if err != nil {
		return domain.AucAggregationModel{}, fmt.Errorf("invalid user ID format '%s': %w", userId, err)
	}

	positions, err := uc.repo.FindByUserID(context.Background(), userUUID)
	if err != nil {
		return domain.AucAggregationModel{}, err
	}

	// Convert positions to AssetModel for existing aggregation service
	assets := make([]domain.AssetModel, len(positions))
	for i, position := range positions {
		assets[i] = domain.AssetModel{
			Symbol:       position.Symbol,
			Quantity:     float32(position.Quantity),
			AveragePrice: float32(position.AveragePrice),
			LastPrice:    float32(position.CurrentPrice),
			Category:     1,
		}
	}

	positionAggregations := uc.aggregationService.AggregateAssetsByCategory(assets)
	totalInvested, currentTotal := uc.aggregationService.CalculateTotals(assets)

	return domain.AucAggregationModel{
		TotalInvested:       totalInvested,
		CurrentTotal:        currentTotal,
		PositionAggregation: positionAggregations,
	}, nil
}

// converts user ID string to UUID with flexible parsing
// Supports both UUID format strings and integer strings (for backward compatibility)
func parseUserIDToUUID(userId string) (uuid.UUID, error) {
	// First, try parsing as a direct UUID
	if userUUID, err := uuid.Parse(userId); err == nil {
		return userUUID, nil
	}

	// If UUID parsing fails, try parsing as integer and convert to deterministic UUID
	// This provides backward compatibility with integer-based user IDs
	if userInt, err := strconv.Atoi(userId); err == nil {
		// Create a deterministic UUID based on the integer ID
		// Using a namespace UUID to ensure consistency
		namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8") // Standard namespace for names
		return uuid.NewSHA1(namespace, []byte(fmt.Sprintf("user_%d", userInt))), nil
	}

	// If both parsing attempts fail, return error
	return uuid.Nil, fmt.Errorf("user ID '%s' cannot be parsed as UUID or integer", userId)
}
