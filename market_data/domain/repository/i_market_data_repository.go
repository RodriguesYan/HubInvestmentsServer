package repository

import "HubInvestments/market_data/domain/model"

type IMarketDataRepository interface {
	GetMarketData(symbols []string) ([]model.MarketDataModel, error)
}
