package di

import (
	"HubInvestments/internal/auth"
	balUsecase "HubInvestments/internal/balance/application/usecase"
	doLoginUsecase "HubInvestments/internal/login/application/usecase"
	mktUsecase "HubInvestments/internal/market_data/application/usecase"
	portfolioUsecase "HubInvestments/internal/portfolio_summary/application/usecase"
	posUsecase "HubInvestments/internal/position/application/usecase"
	watchlistUsecase "HubInvestments/internal/watchlist/application/usecase"
)

// TestContainer is a simple mock container for testing
// It implements the Container interface with configurable services
type TestContainer struct {
	authService                auth.IAuthService
	positionAggregationUseCase *posUsecase.GetPositionAggregationUseCase
	getBalanceUsecase          *balUsecase.GetBalanceUseCase
	getPortfolioSummary        portfolioUsecase.PortfolioSummaryUsecase
	getMarketDataUsecase       mktUsecase.IGetMarketDataUsecase
	getWatchlistUsecase        watchlistUsecase.IGetWatchlistUsecase
	loginUsecase               doLoginUsecase.IDoLoginUsecase
}

// NewTestContainer creates a new test container with optional services
func NewTestContainer() *TestContainer {
	return &TestContainer{}
}

// WithLoginUsecase sets the LoginUsecase for testing
func (c *TestContainer) WithLoginUsecase(usecase doLoginUsecase.IDoLoginUsecase) *TestContainer {
	c.loginUsecase = usecase
	return c
}

// WithAuthService sets the AuthService for testing
func (c *TestContainer) WithAuthService(service auth.IAuthService) *TestContainer {
	c.authService = service
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

// GetAuthService returns the configured AuthService or nil
func (c *TestContainer) GetAuthService() auth.IAuthService {
	return c.authService
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

func (c *TestContainer) DoLoginUsecase() doLoginUsecase.IDoLoginUsecase {
	return c.loginUsecase
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
