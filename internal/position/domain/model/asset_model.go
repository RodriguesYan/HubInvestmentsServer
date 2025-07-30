package domain

// AssetModel represents an individual asset in a position
// @Description Individual asset information in a user's portfolio
type AssetModel struct {
	Symbol       string  `json:"symbol" example:"AAPL"`
	Quantity     float32 `json:"quantity" example:"10.0"`
	AveragePrice float32 `json:"averagePrice" example:"150.0"`
	LastPrice    float32 `json:"currentPrice" example:"155.0"`
	Category     int     `json:"category" example:"1"`
}

// CalculateInvestment returns the total amount invested in this asset
func (a AssetModel) CalculateInvestment() float32 {
	return a.AveragePrice * a.Quantity
}

// CalculateCurrentValue returns the current value of this asset
func (a AssetModel) CalculateCurrentValue() float32 {
	return a.LastPrice * a.Quantity
}

// CalculatePnL returns the profit/loss for this asset
func (a AssetModel) CalculatePnL() float32 {
	return a.CalculateCurrentValue() - a.CalculateInvestment()
}

// CalculatePnLPercentage returns the profit/loss percentage for this asset
func (a AssetModel) CalculatePnLPercentage() float32 {
	investment := a.CalculateInvestment()
	if investment == 0 {
		return 0
	}
	return (a.CalculatePnL() / investment) * 100
}
