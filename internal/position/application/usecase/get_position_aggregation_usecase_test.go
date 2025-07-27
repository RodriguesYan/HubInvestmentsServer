package usecase

import (
	domain "HubInvestments/internal/position/domain/model"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockPositionRepository struct {
	model []domain.AssetsModel
	err   error
}

func (r MockPositionRepository) GetPositionsByUserId(userId string) ([]domain.AssetsModel, error) {
	if r.err != nil {
		return []domain.AssetsModel{}, r.err
	}

	return r.model, nil
}

func Test_GetPositionAggregationUseCase_Success(t *testing.T) {
	userId := "some id"

	assets := []domain.AssetsModel{
		{Symbol: "AAPL", Quantity: 5, AveragePrice: 10, LastPrice: 11, Category: 1},
		{Symbol: "AAPL", Quantity: 7, AveragePrice: 11, LastPrice: 11, Category: 1},
	}

	repo := MockPositionRepository{
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

	assets := []domain.AssetsModel{}

	repo := MockPositionRepository{
		model: assets,
		err:   errors.New("Failed to get balance"),
	}

	_, err := NewGetPositionAggregationUseCase(repo).Execute(userId)

	assert.Error(t, err)
}
