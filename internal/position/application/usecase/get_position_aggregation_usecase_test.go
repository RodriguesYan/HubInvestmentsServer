package usecase

import (
	domain "HubInvestments/internal/position/domain/model"
	service "HubInvestments/internal/position/domain/service"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetPositionAggregationUseCase_Success(t *testing.T) {
	userId := "some id"

	assets := []domain.AssetModel{
		{Symbol: "AAPL", Quantity: 5, AveragePrice: 10, LastPrice: 11, Category: 1},
		{Symbol: "AAPL", Quantity: 7, AveragePrice: 11, LastPrice: 11, Category: 1},
	}

	repo := MockPositionRepositoryLegacy{
		model: assets,
	}

	usecase, err := NewGetPositionAggregationUseCase(repo).Execute(userId)

	assert.NoError(t, err)

	assert.Equal(t, 1, len(usecase.PositionAggregation))
	assert.Equal(t, 1, usecase.PositionAggregation[0].Category)
	assert.Equal(t, float32(5.0), usecase.PositionAggregation[0].Assets[0].Quantity)
	assert.Equal(t, float32(10.0), usecase.PositionAggregation[0].Assets[0].AveragePrice)
	assert.Equal(t, float32(11.0), usecase.PositionAggregation[0].Assets[0].LastPrice)
	assert.Equal(t, float32(127.0), usecase.PositionAggregation[0].TotalInvested)
	assert.Equal(t, float32(132.0), usecase.PositionAggregation[0].CurrentTotal)
	assert.Equal(t, float32(5), usecase.PositionAggregation[0].Pnl)
	assert.Equal(t, float32(3.937008), usecase.PositionAggregation[0].PnlPercentage)
	assert.Equal(t, float32(127.0), usecase.TotalInvested)
	assert.Equal(t, float32(132.0), usecase.CurrentTotal)

	//print the result of usecase in a formatted way
	fmt.Printf("%+v\n", usecase)

}

func Test_GetPositionAggregationUseCase_FailRepo(t *testing.T) {
	userId := "some id"

	assets := []domain.AssetModel{}

	repo := MockPositionRepositoryLegacy{
		model: assets,
		err:   errors.New("Failed to get balance"),
	}

	_, err := NewGetPositionAggregationUseCase(repo).Execute(userId)

	assert.Error(t, err)
}

func Test_GetPositionAggregationUseCase_WithDependencyInjection(t *testing.T) {
	userId := "some id"

	assets := []domain.AssetModel{
		{Symbol: "AAPL", Quantity: 5, AveragePrice: 10, LastPrice: 11, Category: 1},
		{Symbol: "GOOGL", Quantity: 2, AveragePrice: 20, LastPrice: 22, Category: 1},
	}

	repo := MockPositionRepositoryLegacy{
		model: assets,
	}

	// Use the real domain service
	aggregationService := service.NewPositionAggregationService()

	// Create use case with dependency injection
	useCaseWithDI := NewGetPositionAggregationUseCaseWithService(repo, aggregationService)
	result, err := useCaseWithDI.Execute(userId)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.PositionAggregation))
	assert.Equal(t, 1, result.PositionAggregation[0].Category)
	assert.Equal(t, float32(90.0), result.PositionAggregation[0].TotalInvested) // (5*10) + (2*20) = 50 + 40 = 90
	assert.Equal(t, float32(99.0), result.PositionAggregation[0].CurrentTotal)  // (5*11) + (2*22) = 55 + 44 = 99
	assert.Equal(t, float32(9.0), result.PositionAggregation[0].Pnl)            // 99 - 90 = 9
	assert.Equal(t, float32(90.0), result.TotalInvested)
	assert.Equal(t, float32(99.0), result.CurrentTotal)
	assert.Len(t, result.PositionAggregation[0].Assets, 2)
}
