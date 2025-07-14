package usecase

import (
	balUsecase "HubInvestments/balance/application/usecase"
	balDomain "HubInvestments/balance/domain/model"
	"HubInvestments/portfolio_summary/domain/model"
	posUsecase "HubInvestments/position/application/usecase"
	domain "HubInvestments/position/domain/model"
	"fmt"
)

// PortfolioSummaryUsecase interface defines the contract for portfolio summary operations
type PortfolioSummaryUsecase interface {
	Execute(userId string) (model.PortfolioSummaryModel, error)
}

type GetPortfolioSummaryUsecase struct {
	balance  balUsecase.GetBalanceUseCase
	position posUsecase.GetPositionAggregationUseCase
}

func NewGetPortfolioSummaryUsecase(position posUsecase.GetPositionAggregationUseCase, balance balUsecase.GetBalanceUseCase) PortfolioSummaryUsecase {
	return &GetPortfolioSummaryUsecase{position: position, balance: balance}
}

func (uc *GetPortfolioSummaryUsecase) Execute(userId string) (model.PortfolioSummaryModel, error) {
	balanceResult, err := uc.balance.Execute(userId)

	if err != nil {
		return model.PortfolioSummaryModel{}, err
	}

	positionResult, err := uc.position.Execute(userId)

	if err != nil {
		return model.PortfolioSummaryModel{}, err
	}

	fmt.Println(positionResult.CurrentTotal)

	totalPortfolio := getTotalPortfolio(balanceResult, positionResult)

	return model.PortfolioSummaryModel{
		Balance:             balanceResult,
		TotalPortfolio:      totalPortfolio,
		PositionAggregation: positionResult,
	}, err
}

func getTotalPortfolio(balance balDomain.BalanceModel, aggregation domain.AucAggregationModel) float32 {
	return balance.AvailableBalance + aggregation.CurrentTotal
}
