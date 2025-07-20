package response

import (
	balanceDomain "HubInvestments/balance/domain/model"
	portfolioDomain "HubInvestments/portfolio_summary/domain/model"
	positionDomain "HubInvestments/position/domain/model"
)

// BalanceResponse represents the balance response using domain model
type BalanceResponse = balanceDomain.BalanceModel

// PositionAggregationResponse represents the position aggregation response using domain model
type PositionAggregationResponse = positionDomain.AucAggregationModel

// PortfolioSummaryResponse represents the portfolio summary response using domain model
type PortfolioSummaryResponse = portfolioDomain.PortfolioSummaryModel

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Missing authorization header"`
}
