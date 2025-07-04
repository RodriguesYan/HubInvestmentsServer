package usecase

import (
	balUsecase "HubInvestments/balance/application/usecase"
	posUsecase "HubInvestments/position/application/usecase"
)

type GetPortfolioSummaryUsecase struct {
	balance  balUsecase.GetBalanceUseCase
	position posUsecase.GetPositionAggregationUseCase
}

func NewGetPortfolioSummaryUsecase(position posUsecase.GetPositionAggregationUseCase, balance balUsecase.GetBalanceUseCase) *GetPortfolioSummaryUsecase {
	return &GetPortfolioSummaryUsecase{position: position, balance: balance}
}

//TODO: to terminando esse usecase. Ele vai chamar o use case de position e balance pra criar um objeto agregado
//TODO: depois preciso criar um handler dele e disponibilizar a rota
