package usecase

import (
	domain "HubInvestments/position/domain/model"
	"HubInvestments/position/domain/repository"
	"sort"
)

type GetPositionAggregationUseCase struct {
	repo repository.AucRepository
}

func NewGetPositionAggregationUseCase(repo repository.AucRepository) *GetPositionAggregationUseCase {
	return &GetPositionAggregationUseCase{repo: repo}
}

func (uc *GetPositionAggregationUseCase) Execute(userId string) (domain.AucAggregationModel, error) {
	assets, err := uc.repo.GetPositionAggregation(userId)
	if err != nil {
		return domain.AucAggregationModel{}, err
	}

	positionAggregations := uc.aggregateAssetsByCategory(assets)

	var totalBalance float32 = 0

	aucAggregation := domain.AucAggregationModel{
		TotalBalance:        totalBalance,
		PositionAggregation: positionAggregations,
	}

	return aucAggregation, nil
}

func (uc *GetPositionAggregationUseCase) aggregateAssetsByCategory(assets []domain.AssetsModel) []domain.PositionAggregationModel {
	var positionAggregations []domain.PositionAggregationModel

	for _, element := range assets {
		index := sort.Search(len(positionAggregations), func(i int) bool {
			return positionAggregations[i].Category >= element.Category
		})

		if index < len(positionAggregations) && positionAggregations[index].Category == element.Category {
			uc.updateExistingAggregation(&positionAggregations[index], element)
		} else {
			newAggregation := uc.createNewAggregation(element)
			positionAggregations = append(positionAggregations, domain.PositionAggregationModel{})
			copy(positionAggregations[index+1:], positionAggregations[index:])
			positionAggregations[index] = newAggregation
		}
	}

	return positionAggregations
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
