package model

import (
	balanceDomain "HubInvestments/balance/domain/model"
	positionDomain "HubInvestments/position/domain/model"
)

type PortfolioSummaryModel struct {
	Balance             balanceDomain.BalanceModel
	PositionAggregation positionDomain.PositionAggregationModel
	TotalPortfolio      float32
	LastUpdatedDate     string
}
