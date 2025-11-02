package usecase

import (
	"context"
	"fmt"

	"HubInvestments/internal/order_mngmt_system/infra/external"
	repository "HubInvestments/internal/watchlist/domain/repository"
)

// MarketDataModel represents market data for watchlist items
type MarketDataModel struct {
	Symbol      string
	CompanyName string
	LastQuote   float64
	Category    string
}

type IGetWatchlistUsecase interface {
	Execute(userId string) ([]MarketDataModel, error)
}

type GetWatchlistUsecase struct {
	repo             repository.IWatchlistRepository
	marketDataClient external.IMarketDataClient
}

func NewGetWatchlistUsecase(repo repository.IWatchlistRepository, marketDataClient external.IMarketDataClient) IGetWatchlistUsecase {
	return &GetWatchlistUsecase{
		repo:             repo,
		marketDataClient: marketDataClient,
	}
}

func (w *GetWatchlistUsecase) Execute(userId string) ([]MarketDataModel, error) {
	watchlistSymbols, err := w.repo.GetWatchlist(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist: %w", err)
	}

	ctx := context.Background()
	marketDataList, err := w.marketDataClient.GetBatchMarketData(ctx, watchlistSymbols)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	result := make([]MarketDataModel, len(marketDataList))
	for i, md := range marketDataList {
		result[i] = MarketDataModel{
			Symbol:      md.Symbol,
			CompanyName: md.CompanyName,
			LastQuote:   md.LastQuote,
			Category:    md.Category,
		}
	}

	return result, nil
}
