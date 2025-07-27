package model

import (
	balanceDomain "HubInvestments/internal/balance/domain/model"
	positionDomain "HubInvestments/internal/position/domain/model"
)

// PortfolioSummaryModel represents the complete portfolio summary
// @Description Complete portfolio summary including balance and positions
type PortfolioSummaryModel struct {
	Balance             balanceDomain.BalanceModel         `json:"Balance"`
	TotalPortfolio      float32                            `json:"TotalPortfolio" example:"17000.0"`
	LastUpdatedDate     string                             `json:"LastUpdatedDate" example:""`
	PositionAggregation positionDomain.AucAggregationModel `json:"PositionAggregation"`
}
