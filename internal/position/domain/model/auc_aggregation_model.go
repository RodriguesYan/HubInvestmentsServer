package domain

// PositionAggregationModel represents aggregated position data by category
// @Description Position aggregation grouped by asset category
type PositionAggregationModel struct {
	Category      int          `json:"category" example:"1"`
	TotalInvested float32      `json:"totalInvested" example:"6500.0"`
	CurrentTotal  float32      `json:"currentTotal" example:"6750.0"`
	Pnl           float32      `json:"pnl" example:"250.0"`
	PnlPercentage float32      `json:"pnlPercentage" example:"3.85"`
	Assets        []AssetModel `json:"assets"`
}

// AucAggregationModel represents the complete position aggregation
// @Description Complete position aggregation response
type AucAggregationModel struct {
	TotalInvested       float32                    `json:"totalInvested" example:"11500.0"`
	CurrentTotal        float32                    `json:"currentTotal" example:"12000.0"`
	PositionAggregation []PositionAggregationModel `json:"positionAggregation"`
}
