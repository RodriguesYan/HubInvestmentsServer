package model

import (
	"time"
)

type AssetType string

const (
	AssetTypeStock AssetType = "STOCK"
	AssetTypeETF   AssetType = "ETF"
)

type AssetQuote struct {
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	Type          AssetType `json:"type"`
	CurrentPrice  float64   `json:"current_price"`
	BasePrice     float64   `json:"base_price"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
	LastUpdated   time.Time `json:"last_updated"`
	Volume        int64     `json:"volume"`
	MarketCap     int64     `json:"market_cap,omitempty"`
}

func NewAssetQuote(symbol, name string, assetType AssetType, basePrice float64, volume, marketCap int64) *AssetQuote {
	return &AssetQuote{
		Symbol:        symbol,
		Name:          name,
		Type:          assetType,
		CurrentPrice:  basePrice,
		BasePrice:     basePrice,
		Change:        0.0,
		ChangePercent: 0.0,
		LastUpdated:   time.Now(),
		Volume:        volume,
		MarketCap:     marketCap,
	}
}

func (q *AssetQuote) UpdatePrice(newPrice float64) {
	q.Change = newPrice - q.BasePrice
	q.ChangePercent = (q.Change / q.BasePrice) * 100
	q.CurrentPrice = newPrice
	q.LastUpdated = time.Now()
}

func (q *AssetQuote) IsPositiveChange() bool {
	return q.Change >= 0
}
