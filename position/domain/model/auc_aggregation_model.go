package domain

type AssetsModel struct {
	Symbol       string  `json:"symbol" db:"symbol"`
	Quantity     float32 `json:"quantity" db:"quantity"`
	AveragePrice float32 `json:"averagePrice" db:"average_price"`
	LastPrice    float32 `json:"currentPrice" db:"current_price"`
	Category     int     `json:"category" db:"category"`
}

type PositionAggregationModel struct {
	Category      int           `json:"category" db:"category"`
	TotalInvested float32       `json:"totalInvested" db:"total_invested"`
	CurrentTotal  float32       `json:"currentTotal" db:"current_total"`
	Pnl           float32       `json:"pnl" db:"pnl"`
	PnlPercentage float32       `json:"pnlPercentage" db:"pnl_percentage"`
	Assets        []AssetsModel `json:"assets"`
}

type AucAggregationModel struct {
	TotalBalance        float32                    `json:"totalBalance" db:"total_balance"`
	PositionAggregation []PositionAggregationModel `json:"positionAggregation"`
}
