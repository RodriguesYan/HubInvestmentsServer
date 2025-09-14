package usecase

import (
	domain "HubInvestments/internal/position/domain/model"
	service "HubInvestments/internal/position/domain/service"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_GetPositionAggregationUseCase_Success(t *testing.T) {
	userUUID := uuid.New()
	userId := userUUID.String()

	// Create positions using new domain model
	position1, _ := domain.NewPosition(userUUID, "AAPL", 5.0, 10.0, domain.PositionTypeLong)
	position1.CurrentPrice = 11.0
	position1.MarketValue = position1.Quantity * position1.CurrentPrice

	position2, _ := domain.NewPosition(userUUID, "AAPL", 7.0, 11.0, domain.PositionTypeLong)
	position2.CurrentPrice = 11.0
	position2.MarketValue = position2.Quantity * position2.CurrentPrice

	repo := NewMockPositionRepositoryForNew()
	repo.AddPosition(position1)
	repo.AddPosition(position2)

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
	userUUID := uuid.New()
	userId := userUUID.String()

	repo := NewMockPositionRepositoryForNew()
	repo.shouldFailFind = true

	_, err := NewGetPositionAggregationUseCase(repo).Execute(userId)

	assert.Error(t, err)
}

func Test_GetPositionAggregationUseCase_WithDependencyInjection(t *testing.T) {
	userUUID := uuid.New()
	userId := userUUID.String()

	// Create positions using new domain model
	position1, _ := domain.NewPosition(userUUID, "AAPL", 5.0, 10.0, domain.PositionTypeLong)
	position1.CurrentPrice = 11.0
	position1.MarketValue = position1.Quantity * position1.CurrentPrice

	position2, _ := domain.NewPosition(userUUID, "GOOGL", 2.0, 20.0, domain.PositionTypeLong)
	position2.CurrentPrice = 22.0
	position2.MarketValue = position2.Quantity * position2.CurrentPrice

	repo := NewMockPositionRepositoryForNew()
	repo.AddPosition(position1)
	repo.AddPosition(position2)

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

func Test_GetPositionAggregationUseCase_InvalidUUID(t *testing.T) {
	invalidUserId := "invalid-uuid"

	repo := NewMockPositionRepositoryForNew()

	_, err := NewGetPositionAggregationUseCase(repo).Execute(invalidUserId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID")
}

func Test_GetPositionAggregationUseCase_EmptyPositions(t *testing.T) {
	userUUID := uuid.New()
	userId := userUUID.String()

	repo := NewMockPositionRepositoryForNew()

	result, err := NewGetPositionAggregationUseCase(repo).Execute(userId)

	assert.NoError(t, err)
	assert.Equal(t, float32(0.0), result.TotalInvested)
	assert.Equal(t, float32(0.0), result.CurrentTotal)
	assert.Len(t, result.PositionAggregation, 0)
}
