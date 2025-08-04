package service

import (
	domain "HubInvestments/internal/position/domain/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPositionAggregationService_AggregateAssetsByCategory(t *testing.T) {
	service := NewPositionAggregationService()

	t.Run("Single category with multiple assets", func(t *testing.T) {
		assets := []domain.AssetModel{
			{Symbol: "AAPL", Quantity: 5, AveragePrice: 10, LastPrice: 11, Category: 1},
			{Symbol: "AAPL", Quantity: 7, AveragePrice: 11, LastPrice: 11, Category: 1},
		}

		result := service.AggregateAssetsByCategory(assets)

		assert.Len(t, result, 1)
		assert.Equal(t, 1, result[0].Category)
		assert.Equal(t, float32(127.0), result[0].TotalInvested)    // (5*10) + (7*11) = 50 + 77 = 127
		assert.Equal(t, float32(132.0), result[0].CurrentTotal)     // (5*11) + (7*11) = 55 + 77 = 132
		assert.Equal(t, float32(5.0), result[0].Pnl)                // 132 - 127 = 5
		assert.Equal(t, float32(3.937008), result[0].PnlPercentage) // (5/127)*100 = 3.937008
		assert.Len(t, result[0].Assets, 2)
	})

	t.Run("Multiple categories", func(t *testing.T) {
		assets := []domain.AssetModel{
			{Symbol: "AAPL", Quantity: 10, AveragePrice: 150, LastPrice: 155, Category: 1},
			{Symbol: "GOOGL", Quantity: 2, AveragePrice: 2500, LastPrice: 2600, Category: 1},
			{Symbol: "VTI", Quantity: 50, AveragePrice: 200, LastPrice: 210, Category: 2},
		}

		result := service.AggregateAssetsByCategory(assets)

		assert.Len(t, result, 2)

		// Check first category (stocks)
		assert.Equal(t, 1, result[0].Category)
		assert.Equal(t, float32(6500.0), result[0].TotalInvested) // (10*150) + (2*2500) = 1500 + 5000 = 6500
		assert.Equal(t, float32(6750.0), result[0].CurrentTotal)  // (10*155) + (2*2600) = 1550 + 5200 = 6750
		assert.Equal(t, float32(250.0), result[0].Pnl)            // 6750 - 6500 = 250
		assert.Len(t, result[0].Assets, 2)

		// Check second category (ETFs)
		assert.Equal(t, 2, result[1].Category)
		assert.Equal(t, float32(10000.0), result[1].TotalInvested) // 50*200 = 10000
		assert.Equal(t, float32(10500.0), result[1].CurrentTotal)  // 50*210 = 10500
		assert.Equal(t, float32(500.0), result[1].Pnl)             // 10500 - 10000 = 500
		assert.Len(t, result[1].Assets, 1)
	})

	t.Run("Empty assets", func(t *testing.T) {
		assets := []domain.AssetModel{}

		result := service.AggregateAssetsByCategory(assets)

		assert.Len(t, result, 0)
	})

	t.Run("Single asset", func(t *testing.T) {
		assets := []domain.AssetModel{
			{Symbol: "AAPL", Quantity: 10, AveragePrice: 150, LastPrice: 155, Category: 1},
		}

		result := service.AggregateAssetsByCategory(assets)

		assert.Len(t, result, 1)
		assert.Equal(t, 1, result[0].Category)
		assert.Equal(t, float32(1500.0), result[0].TotalInvested)
		assert.Equal(t, float32(1550.0), result[0].CurrentTotal)
		assert.Equal(t, float32(50.0), result[0].Pnl)
		assert.Len(t, result[0].Assets, 1)
	})
}

func TestPositionAggregationService_CalculateTotals(t *testing.T) {
	service := NewPositionAggregationService()

	t.Run("Multiple assets", func(t *testing.T) {
		assets := []domain.AssetModel{
			{Symbol: "AAPL", Quantity: 5, AveragePrice: 10, LastPrice: 11, Category: 1},
			{Symbol: "AAPL", Quantity: 7, AveragePrice: 11, LastPrice: 11, Category: 1},
		}

		totalInvested, currentTotal := service.CalculateTotals(assets)

		assert.Equal(t, float32(127.0), totalInvested) // (5*10) + (7*11) = 50 + 77 = 127
		assert.Equal(t, float32(132.0), currentTotal)  // (5*11) + (7*11) = 55 + 77 = 132
	})

	t.Run("Empty assets", func(t *testing.T) {
		assets := []domain.AssetModel{}

		totalInvested, currentTotal := service.CalculateTotals(assets)

		assert.Equal(t, float32(0), totalInvested)
		assert.Equal(t, float32(0), currentTotal)
	})

	t.Run("Single asset", func(t *testing.T) {
		assets := []domain.AssetModel{
			{Symbol: "AAPL", Quantity: 10, AveragePrice: 150, LastPrice: 155, Category: 1},
		}

		totalInvested, currentTotal := service.CalculateTotals(assets)

		assert.Equal(t, float32(1500.0), totalInvested)
		assert.Equal(t, float32(1550.0), currentTotal)
	})
}

func TestPositionAggregationService_ZeroInvestment(t *testing.T) {
	service := NewPositionAggregationService()

	t.Run("Asset with zero investment", func(t *testing.T) {
		assets := []domain.AssetModel{
			{Symbol: "FREE", Quantity: 10, AveragePrice: 0, LastPrice: 5, Category: 1},
		}

		result := service.AggregateAssetsByCategory(assets)

		assert.Len(t, result, 1)
		assert.Equal(t, float32(0.0), result[0].TotalInvested)
		assert.Equal(t, float32(50.0), result[0].CurrentTotal)
		assert.Equal(t, float32(50.0), result[0].Pnl)
		assert.Equal(t, float32(0.0), result[0].PnlPercentage) // Should be 0 when investment is 0
	})
}
