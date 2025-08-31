package service

import (
	"HubInvestments/internal/realtime_quotes/domain/model"
)

type AssetDataService struct {
	assets map[string]*model.AssetQuote
}

func NewAssetDataService() *AssetDataService {
	service := &AssetDataService{
		assets: make(map[string]*model.AssetQuote),
	}
	service.initializeAssets()
	return service
}

func (s *AssetDataService) initializeAssets() {
	// 10 Well-known stocks with realistic base prices
	stocks := []struct {
		symbol    string
		name      string
		basePrice float64
		volume    int64
		marketCap int64
	}{
		{"AAPL", "Apple Inc.", 175.50, 50000000, 2800000000000},
		{"MSFT", "Microsoft Corporation", 420.25, 25000000, 3100000000000},
		{"GOOGL", "Alphabet Inc.", 140.80, 20000000, 1800000000000},
		{"AMZN", "Amazon.com Inc.", 155.30, 35000000, 1600000000000},
		{"TSLA", "Tesla Inc.", 248.75, 80000000, 790000000000},
		{"NVDA", "NVIDIA Corporation", 875.20, 45000000, 2200000000000},
		{"META", "Meta Platforms Inc.", 485.60, 15000000, 1200000000000},
		{"NFLX", "Netflix Inc.", 485.90, 8000000, 210000000000},
		{"JPM", "JPMorgan Chase & Co.", 185.40, 12000000, 540000000000},
		{"V", "Visa Inc.", 275.80, 6000000, 580000000000},
	}

	// 10 Well-known ETFs with realistic base prices
	etfs := []struct {
		symbol    string
		name      string
		basePrice float64
		volume    int64
	}{
		{"SPY", "SPDR S&P 500 ETF Trust", 485.20, 40000000},
		{"QQQ", "Invesco QQQ Trust", 395.75, 35000000},
		{"VTI", "Vanguard Total Stock Market ETF", 245.30, 25000000},
		{"IWM", "iShares Russell 2000 ETF", 195.85, 20000000},
		{"EFA", "iShares MSCI EAFE ETF", 78.90, 15000000},
		{"GLD", "SPDR Gold Shares", 185.45, 10000000},
		{"TLT", "iShares 20+ Year Treasury Bond ETF", 92.30, 8000000},
		{"VNQ", "Vanguard Real Estate ETF", 85.75, 5000000},
		{"XLF", "Financial Select Sector SPDR Fund", 38.20, 18000000},
		{"XLK", "Technology Select Sector SPDR Fund", 195.60, 12000000},
	}

	for _, stock := range stocks {
		quote := model.NewAssetQuote(
			stock.symbol,
			stock.name,
			model.AssetTypeStock,
			stock.basePrice,
			stock.volume,
			stock.marketCap,
		)
		s.assets[stock.symbol] = quote
	}

	for _, etf := range etfs {
		quote := model.NewAssetQuote(
			etf.symbol,
			etf.name,
			model.AssetTypeETF,
			etf.basePrice,
			etf.volume,
			0, // ETFs don't have market cap
		)
		s.assets[etf.symbol] = quote
	}
}

func (s *AssetDataService) GetAllAssets() map[string]*model.AssetQuote {
	result := make(map[string]*model.AssetQuote)
	for symbol, quote := range s.assets {
		result[symbol] = quote
	}
	return result
}

func (s *AssetDataService) GetAssetBySymbol(symbol string) (*model.AssetQuote, bool) {
	quote, exists := s.assets[symbol]
	return quote, exists
}

func (s *AssetDataService) GetStocks() []*model.AssetQuote {
	var stocks []*model.AssetQuote
	for _, quote := range s.assets {
		if quote.Type == model.AssetTypeStock {
			stocks = append(stocks, quote)
		}
	}
	return stocks
}

func (s *AssetDataService) GetETFs() []*model.AssetQuote {
	var etfs []*model.AssetQuote
	for _, quote := range s.assets {
		if quote.Type == model.AssetTypeETF {
			etfs = append(etfs, quote)
		}
	}
	return etfs
}
