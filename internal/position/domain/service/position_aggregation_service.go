package service

import (
	domain "HubInvestments/internal/position/domain/model"
	"sort"
)

// PositionAggregationService handles the business logic for aggregating positions by category
type PositionAggregationService interface {
	AggregateAssetsByCategory(assets []domain.AssetModel) []domain.PositionAggregationModel
	CalculateTotals(assets []domain.AssetModel) (totalInvested, currentTotal float32)
}

type positionAggregationService struct{}

// NewPositionAggregationService creates a new instance of PositionAggregationService
func NewPositionAggregationService() PositionAggregationService {
	return &positionAggregationService{}
}

// AggregateAssetsByCategory groups assets by category and calculates aggregated values
func (s *positionAggregationService) AggregateAssetsByCategory(assets []domain.AssetModel) []domain.PositionAggregationModel {
	var positionAggregations []domain.PositionAggregationModel

	for _, element := range assets {
		// sort.Search returns the index where element.Category should be inserted
		// to maintain sorted order. We need to check two things:
		// 1. If the index is within bounds (index < len)
		// 2. If the category at that index matches our element's category
		//
		// If both conditions are true, we found an existing aggregation for this category
		// If either condition is false, we need to create a new aggregation
		index := sort.Search(len(positionAggregations), func(i int) bool {
			return positionAggregations[i].Category >= element.Category
		})

		if index < len(positionAggregations) && positionAggregations[index].Category == element.Category {
			s.updateExistingAggregation(&positionAggregations[index], element)
		} else {
			newAggregation := s.createNewAggregation(element)
			positionAggregations = append(positionAggregations, domain.PositionAggregationModel{})
			copy(positionAggregations[index+1:], positionAggregations[index:])
			positionAggregations[index] = newAggregation
		}
	}

	return positionAggregations
}

// CalculateTotals calculates the total invested and current total values across all assets
func (s *positionAggregationService) CalculateTotals(assets []domain.AssetModel) (totalInvested, currentTotal float32) {
	var invested float32 = 0
	var current float32 = 0

	for _, element := range assets {
		// Calculate individual asset values using domain methods
		assetInvestment := element.CalculateInvestment()
		assetCurrentValue := element.CalculateCurrentValue()

		// Add to running totals
		invested += assetInvestment
		current += assetCurrentValue
	}

	return invested, current
}

// updateExistingAggregation updates an existing category aggregation with a new asset
func (s *positionAggregationService) updateExistingAggregation(aggregation *domain.PositionAggregationModel, asset domain.AssetModel) {
	aggregation.Assets = append(aggregation.Assets, asset)

	assetInvestment := asset.CalculateInvestment()
	assetCurrentValue := asset.CalculateCurrentValue()
	assetPnl := asset.CalculatePnL()

	aggregation.TotalInvested += assetInvestment
	aggregation.CurrentTotal += assetCurrentValue
	aggregation.Pnl += assetPnl

	if aggregation.TotalInvested > 0 {
		aggregation.PnlPercentage = (aggregation.Pnl / aggregation.TotalInvested) * 100
	}
}

// createNewAggregation creates a new category aggregation for an asset
func (s *positionAggregationService) createNewAggregation(asset domain.AssetModel) domain.PositionAggregationModel {
	assetInvestment := asset.CalculateInvestment()
	assetCurrentValue := asset.CalculateCurrentValue()
	assetPnl := asset.CalculatePnL()

	var pnlPercentage float32 = 0
	if assetInvestment > 0 {
		pnlPercentage = (assetPnl / assetInvestment) * 100
	}

	return domain.PositionAggregationModel{
		Category:      asset.Category,
		TotalInvested: assetInvestment,
		CurrentTotal:  assetCurrentValue,
		Pnl:           assetPnl,
		PnlPercentage: pnlPercentage,
		Assets:        []domain.AssetModel{asset},
	}
}
