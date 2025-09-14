package usecase

import (
	domain "HubInvestments/internal/position/domain/model"
	repository "HubInvestments/internal/position/domain/repository"
	service "HubInvestments/internal/position/domain/service"
	"context"

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
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return domain.AucAggregationModel{}, err
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
