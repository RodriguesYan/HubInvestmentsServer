package usecase

import (
	balUsecase "HubInvestments/balance/application/usecase"
	balDomain "HubInvestments/balance/domain/model"
	"HubInvestments/portfolio_summary/domain/model"
	posUsecase "HubInvestments/position/application/usecase"
	domain "HubInvestments/position/domain/model"
)

type GetPortfolioSummaryUsecase struct {
	balance  balUsecase.GetBalanceUseCase
	position posUsecase.GetPositionAggregationUseCase
}

func NewGetPortfolioSummaryUsecase(position posUsecase.GetPositionAggregationUseCase, balance balUsecase.GetBalanceUseCase) *GetPortfolioSummaryUsecase {
	return &GetPortfolioSummaryUsecase{position: position, balance: balance}
}

//TODO: depois preciso criar um handler dele e disponibilizar a rota

func (uc *GetPortfolioSummaryUsecase) Execute(userId string) (model.PortfolioSummaryModel, error) {
	balanceResult, err := uc.balance.Execute(userId)

	if err != nil {
		return model.PortfolioSummaryModel{}, err
	}

	positionResult, err := uc.position.Execute(userId)

	if err != nil {
		return model.PortfolioSummaryModel{}, err
	}

	totalPortfolio := getTotalPortfolio(balanceResult, positionResult)

	return model.PortfolioSummaryModel{
		Balance:             balanceResult,
		PositionAggregation: positionResult.PositionAggregation,
		TotalPortfolio:      totalPortfolio,
	}, err
}

func getTotalPortfolio(balance balDomain.BalanceModel, aggregation domain.AucAggregationModel) float32 {
	return balance.AvailableBalance + aggregation.CurrentTotal
}
