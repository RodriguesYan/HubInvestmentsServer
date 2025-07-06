package usecase

import (
	domain "HubInvestments/position/domain/model"
	"HubInvestments/position/domain/repository"
	"sort"
)

type GetPositionAggregationUseCase struct {
	repo repository.PositionRepository
}

func NewGetPositionAggregationUseCase(repo repository.PositionRepository) *GetPositionAggregationUseCase {
	return &GetPositionAggregationUseCase{repo: repo}
}

func (uc *GetPositionAggregationUseCase) Execute(userId string) (domain.AucAggregationModel, error) {
	assets, err := uc.repo.GetPositionsByUserId(userId)
	if err != nil {
		return domain.AucAggregationModel{}, err
	}

	positionAggregations, totalInvested, currentTotal := uc.aggregateAssetsByCategory(assets)

	aucAggregation := domain.AucAggregationModel{
		TotalInvested:       totalInvested,
		CurrentTotal:        currentTotal,
		PositionAggregation: positionAggregations,
	}

	return aucAggregation, nil
}

func (uc *GetPositionAggregationUseCase) aggregateAssetsByCategory(assets []domain.AssetsModel) (aggregation []domain.PositionAggregationModel, totalInvested float32, currentTotal float32) {
	var positionAggregations []domain.PositionAggregationModel
	var invested float32 = 0
	var current float32 = 0

	for _, element := range assets {
		index := sort.Search(len(positionAggregations), func(i int) bool {
			return positionAggregations[i].Category >= element.Category
		})

		if index < len(positionAggregations) && positionAggregations[index].Category == element.Category {
			uc.updateExistingAggregation(&positionAggregations[index], element)
			invested += positionAggregations[index].TotalInvested
			currentTotal += positionAggregations[index].CurrentTotal
		} else {
			newAggregation := uc.createNewAggregation(element)
			positionAggregations = append(positionAggregations, domain.PositionAggregationModel{})
			copy(positionAggregations[index+1:], positionAggregations[index:])
			positionAggregations[index] = newAggregation
			invested += positionAggregations[index].TotalInvested
			currentTotal += positionAggregations[index].CurrentTotal
		}
	}

	return positionAggregations, invested, current
}

func (uc *GetPositionAggregationUseCase) updateExistingAggregation(aggregation *domain.PositionAggregationModel, asset domain.AssetsModel) {
	aggregation.Assets = append(aggregation.Assets, asset)

	assetInvestment := asset.AveragePrice * asset.Quantity
	assetCurrentValue := asset.LastPrice * asset.Quantity
	assetPnl := assetCurrentValue - assetInvestment

	aggregation.TotalInvested += assetInvestment
	aggregation.CurrentTotal += assetCurrentValue
	aggregation.Pnl += assetPnl

	if aggregation.TotalInvested > 0 {
		aggregation.PnlPercentage = (aggregation.Pnl / aggregation.TotalInvested) * 100
	}
}

func (uc *GetPositionAggregationUseCase) createNewAggregation(asset domain.AssetsModel) domain.PositionAggregationModel {
	assetInvestment := asset.AveragePrice * asset.Quantity
	assetCurrentValue := asset.LastPrice * asset.Quantity
	assetPnl := assetCurrentValue - assetInvestment

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
		Assets:        []domain.AssetsModel{asset},
	}
}
