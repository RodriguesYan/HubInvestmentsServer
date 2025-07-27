package di

import (
	balUsecase "HubInvestments/balance/application/usecase"
	mktUsecase "HubInvestments/market_data/application/usecase"
	portfolioUsecase "HubInvestments/portfolio_summary/application/usecase"
	posService "HubInvestments/position/application/service"
	posUsecase "HubInvestments/position/application/usecase"
	watchlistUsecase "HubInvestments/watchlist/application/usecase"
)

// TestContainer is a simple mock container for testing
// It implements the Container interface with configurable services
type TestContainer struct {
	aucService                 *posService.AucService
	positionAggregationUseCase *posUsecase.GetPositionAggregationUseCase
	getBalanceUsecase          *balUsecase.GetBalanceUseCase
	getPortfolioSummary        portfolioUsecase.PortfolioSummaryUsecase
	getMarketDataUsecase       mktUsecase.IGetMarketDataUsecase
	getWatchlistUsecase        watchlistUsecase.IGetWatchlistUsecase
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

// WithMarketDataUsecase sets the MarketDataUsecase for testing
func (c *TestContainer) WithMarketDataUsecase(usecase mktUsecase.IGetMarketDataUsecase) *TestContainer {
	c.getMarketDataUsecase = usecase
	return c
}

func (c *TestContainer) WithWatchlistUsecase(usecase watchlistUsecase.IGetWatchlistUsecase) *TestContainer {
	c.getWatchlistUsecase = usecase
	return c
}

// GetAucService returns the configured AucService or nil
func (c *TestContainer) GetAucService() *posService.AucService {
	return c.aucService
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

func (c *TestContainer) GetMarketDataUsecase() mktUsecase.IGetMarketDataUsecase {
	return c.getMarketDataUsecase
}

func (c *TestContainer) GetWatchlistUsecase() watchlistUsecase.IGetWatchlistUsecase {
	// No-op implementation for testing
	return c.getWatchlistUsecase
}

// Cache management methods for testing (no-op implementations)
func (c *TestContainer) InvalidateMarketDataCache(symbols []string) error {
	// No-op implementation for testing
	return nil
}

func (c *TestContainer) WarmMarketDataCache(symbols []string) error {
	// No-op implementation for testing
	return nil
}

// Add new methods here as you add them to the Container interface
// Example:
// func (c *TestContainer) GetNewService() *NewService {
//     return c.newService
// }
