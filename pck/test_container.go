package di

import (
	balService "HubInvestments/balance/application/service"
	balUsecase "HubInvestments/balance/application/usecase"
	portfolioUsecase "HubInvestments/portfolio_summary/application/usecase"
	posService "HubInvestments/position/application/service"
	posUsecase "HubInvestments/position/application/usecase"
)

// TestContainer is a simple mock container for testing
// It implements the Container interface with configurable services
type TestContainer struct {
	aucService                 *posService.AucService
	balanceService             *balService.BalanceService
	positionAggregationUseCase *posUsecase.GetPositionAggregationUseCase
	getBalanceUsecase          *balUsecase.GetBalanceUseCase
	getPortfolioSummary        portfolioUsecase.PortfolioSummaryUsecase
}

// NewTestContainer creates a new test container with optional services
func NewTestContainer() *TestContainer {
	return &TestContainer{}
}

// WithAucService sets the AucService for testing
func (c *TestContainer) WithAucService(service *posService.AucService) *TestContainer {
	c.aucService = service
	return c
}

// WithBalanceService sets the BalanceService for testing
func (c *TestContainer) WithBalanceService(service *balService.BalanceService) *TestContainer {
	c.balanceService = service
	return c
}

// WithPositionAggregationUseCase sets the PositionAggregationUseCase for testing
func (c *TestContainer) WithPositionAggregationUseCase(usecase *posUsecase.GetPositionAggregationUseCase) *TestContainer {
	c.positionAggregationUseCase = usecase
	return c
}

// WithBalanceUseCase sets the BalanceUseCase for testing
func (c *TestContainer) WithBalanceUseCase(usecase *balUsecase.GetBalanceUseCase) *TestContainer {
	c.getBalanceUsecase = usecase
	return c
}

// WithPortfolioSummaryUsecase sets the PortfolioSummaryUsecase for testing
func (c *TestContainer) WithPortfolioSummaryUsecase(usecase portfolioUsecase.PortfolioSummaryUsecase) *TestContainer {
	c.getPortfolioSummary = usecase
	return c
}

// GetAucService returns the configured AucService or nil
func (c *TestContainer) GetAucService() *posService.AucService {
	return c.aucService
}

// GetBalanceService returns the configured BalanceService or nil
func (c *TestContainer) GetBalanceService() *balService.BalanceService {
	return c.balanceService
}

// GetPositionAggregationUseCase returns the configured PositionAggregationUseCase or nil
func (c *TestContainer) GetPositionAggregationUseCase() *posUsecase.GetPositionAggregationUseCase {
	return c.positionAggregationUseCase
}

func (c *TestContainer) GetBalanceUseCase() *balUsecase.GetBalanceUseCase {
	return c.getBalanceUsecase
}

func (c *TestContainer) GetPortfolioSummaryUsecase() portfolioUsecase.PortfolioSummaryUsecase {
	return c.getPortfolioSummary
}

// Add new methods here as you add them to the Container interface
// Example:
// func (c *TestContainer) GetNewService() *NewService {
//     return c.newService
// }
