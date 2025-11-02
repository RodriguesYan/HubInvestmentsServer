package service

import (
	"HubInvestments/internal/realtime_quotes/domain/model"
	"testing"
)

func TestNewAssetDataService(t *testing.T) {
	service := NewAssetDataService()

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	assets := service.GetAllAssets()
	if len(assets) != 20 {
		t.Errorf("Expected 20 assets (10 stocks + 10 ETFs), got %d", len(assets))
	}
}

func TestGetStocks(t *testing.T) {
	service := NewAssetDataService()
	stocks := service.GetStocks()

	if len(stocks) != 10 {
		t.Errorf("Expected 10 stocks, got %d", len(stocks))
	}

	for _, stock := range stocks {
		if stock.Type != model.AssetTypeStock {
			t.Errorf("Expected stock type, got %s for symbol %s", stock.Type, stock.Symbol)
		}
	}
}

func TestGetETFs(t *testing.T) {
	service := NewAssetDataService()
	etfs := service.GetETFs()

	if len(etfs) != 10 {
		t.Errorf("Expected 10 ETFs, got %d", len(etfs))
	}

	for _, etf := range etfs {
		if etf.Type != model.AssetTypeETF {
			t.Errorf("Expected ETF type, got %s for symbol %s", etf.Type, etf.Symbol)
		}
	}
}

func TestGetAssetBySymbol(t *testing.T) {
	service := NewAssetDataService()

	// Test existing symbol
	asset, exists := service.GetAssetBySymbol("AAPL")
	if !exists {
		t.Error("Expected AAPL to exist")
	}
	if asset.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", asset.Symbol)
	}

	// Test non-existing symbol
	_, exists = service.GetAssetBySymbol("NONEXISTENT")
	if exists {
		t.Error("Expected NONEXISTENT to not exist")
	}
}
