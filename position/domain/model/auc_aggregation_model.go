package domain

// AssetsModel represents an individual asset in a position
// @Description Individual asset information in a user's portfolio
type AssetsModel struct {
	Symbol       string  `json:"symbol" example:"AAPL"`
	Quantity     float32 `json:"quantity" example:"10.0"`
	AveragePrice float32 `json:"averagePrice" example:"150.0"`
	LastPrice    float32 `json:"currentPrice" example:"155.0"`
	Category     int     `json:"category" example:"1"`
}

// PositionAggregationModel represents aggregated position data by category
// @Description Position aggregation grouped by asset category
type PositionAggregationModel struct {
	Category      int           `json:"category" example:"1"`
	TotalInvested float32       `json:"totalInvested" example:"6500.0"`
	CurrentTotal  float32       `json:"currentTotal" example:"6750.0"`
	Pnl           float32       `json:"pnl" example:"250.0"`
	PnlPercentage float32       `json:"pnlPercentage" example:"3.85"`
	Assets        []AssetsModel `json:"assets"`
}

// AucAggregationModel represents the complete position aggregation
// @Description Complete position aggregation response
type AucAggregationModel struct {
	TotalInvested       float32                    `json:"totalInvested" example:"11500.0"`
	CurrentTotal        float32                    `json:"currentTotal" example:"12000.0"`
	PositionAggregation []PositionAggregationModel `json:"positionAggregation"`
}
