package usecase

import (
	"HubInvestments/market_data/application/usecase"
	"HubInvestments/market_data/domain/model"
	"HubInvestments/watchlist/domain/repository"
)

type IGetWatchlistUsecase interface {
	Execute(userId string) ([]model.MarketDataModel, error)
}

type GetWatchlistUsecase struct {
	repo           repository.IWatchlistRepository
	mktDataUsecase usecase.IGetMarketDataUsecase
}

func NewGetWatchlistUsecase(repo repository.IWatchlistRepository, mktDataUsecase usecase.IGetMarketDataUsecase) IGetWatchlistUsecase {
	return &GetWatchlistUsecase{repo: repo, mktDataUsecase: mktDataUsecase}
}

func (w *GetWatchlistUsecase) Execute(userId string) ([]model.MarketDataModel, error) {
	watchlistSymbols, err := w.repo.GetWatchlist(userId)

	if err != nil {
		return nil, err
	}

	mtkDataUsecase, err := w.mktDataUsecase.Execute(watchlistSymbols)

	if err != nil {
		return nil, err
	}

	return mtkDataUsecase, nil
}
