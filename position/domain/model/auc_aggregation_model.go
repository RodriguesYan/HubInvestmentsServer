package domain

type AssetsModel struct {
	Symbol       string  `json:"symbol"`
	Quantity     float32 `json:"quantity"`
	AveragePrice float32 `json:"averagePrice"`
	LastPrice    float32 `json:"currentPrice"`
	Category     int     `json:"category"`
}

type PositionAggregationModel struct {
	Category      int           `json:"category"`
	TotalInvested float32       `json:"totalInvested"`
	CurrentTotal  float32       `json:"currentTotal"`
	Pnl           float32       `json:"pnl"`
	PnlPercentage float32       `json:"pnlPercentage"`
	Assets        []AssetsModel `json:"assets"`
}

type AucAggregationModel struct {
	TotalInvested       float32                    `json:"totalInvested"`
	CurrentTotal        float32                    `json:"currentTotal"`
	PositionAggregation []PositionAggregationModel `json:"positionAggregation"`
}
