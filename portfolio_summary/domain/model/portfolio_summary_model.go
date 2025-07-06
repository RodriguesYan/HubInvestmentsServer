package model

import (
	balanceDomain "HubInvestments/balance/domain/model"
	positionDomain "HubInvestments/position/domain/model"
)

type PortfolioSummaryModel struct {
	Balance             balanceDomain.BalanceModel
	TotalPortfolio      float32
	LastUpdatedDate     string
	PositionAggregation positionDomain.AucAggregationModel
}
