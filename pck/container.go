package di

import (
	"time"

	balUsecase "HubInvestments/internal/balance/application/usecase"
	balancePersistence "HubInvestments/internal/balance/infra/persistence"
	mktUsecase "HubInvestments/internal/market_data/application/usecase"
	mktCache "HubInvestments/internal/market_data/infra/cache"
	mktPersistence "HubInvestments/internal/market_data/infra/persistence"
	portfolioUsecase "HubInvestments/internal/portfolio_summary/application/usecase"
	posService "HubInvestments/internal/position/application/service"
	posUsecase "HubInvestments/internal/position/application/usecase"
	positionPersistence "HubInvestments/internal/position/infra/persistence"
	watchlistUsecase "HubInvestments/internal/watchlist/application/usecase"
	watchPersistence "HubInvestments/internal/watchlist/infra/persistence"
	"HubInvestments/shared/infra/cache"
	"HubInvestments/shared/infra/database"

	"github.com/redis/go-redis/v9"
)

type Container interface {
	GetAucService() *posService.AucService
	GetPositionAggregationUseCase() *posUsecase.GetPositionAggregationUseCase
	GetBalanceUseCase() *balUsecase.GetBalanceUseCase
	GetPortfolioSummaryUsecase() portfolioUsecase.PortfolioSummaryUsecase
	GetMarketDataUsecase() mktUsecase.IGetMarketDataUsecase
	GetWatchlistUsecase() watchlistUsecase.IGetWatchlistUsecase

	// Cache management methods for admin operations
	InvalidateMarketDataCache(symbols []string) error
	WarmMarketDataCache(symbols []string) error
}

type containerImpl struct {
	AucService                 *posService.AucService
	PositionAggregationUseCase *posUsecase.GetPositionAggregationUseCase
	BalanceUsecase             *balUsecase.GetBalanceUseCase
	PortfolioSummaryUsecase    portfolioUsecase.PortfolioSummaryUsecase
	MarketDataUsecase          mktUsecase.IGetMarketDataUsecase
	MarketDataCacheManager     mktCache.CacheManager
	WatchlistUsecase           watchlistUsecase.IGetWatchlistUsecase
}

func (c *containerImpl) GetAucService() *posService.AucService {
	return c.AucService
}

func (c *containerImpl) GetPositionAggregationUseCase() *posUsecase.GetPositionAggregationUseCase {
	return c.PositionAggregationUseCase
}

func (c *containerImpl) GetBalanceUseCase() *balUsecase.GetBalanceUseCase {
	return c.BalanceUsecase
}

func (c *containerImpl) GetPortfolioSummaryUsecase() portfolioUsecase.PortfolioSummaryUsecase {
	return c.PortfolioSummaryUsecase
}

func (c *containerImpl) GetMarketDataUsecase() mktUsecase.IGetMarketDataUsecase {
	return c.MarketDataUsecase
}

// Cache management methods implementation
func (c *containerImpl) InvalidateMarketDataCache(symbols []string) error {
	return c.MarketDataCacheManager.InvalidateCache(symbols)
}

func (c *containerImpl) WarmMarketDataCache(symbols []string) error {
	return c.MarketDataCacheManager.WarmCache(symbols)
}

func (c *containerImpl) GetWatchlistUsecase() watchlistUsecase.IGetWatchlistUsecase {
	return c.WatchlistUsecase
}

func NewContainer() (Container, error) {
	// Create database connection using the abstraction layer
	db, err := database.CreateDatabaseConnection()
	if err != nil {
		return nil, err
	}

	// Create repositories using the database abstraction
	positionRepo := positionPersistence.NewPositionRepository(db)
	aucService := posService.NewAucService(positionRepo)
	positionAggregationUseCase := posUsecase.NewGetPositionAggregationUseCase(positionRepo)

	balanceRepo := balancePersistence.NewBalanceRepository(db)
	balanceUsecase := balUsecase.NewGetBalanceUseCase(balanceRepo)
	portfolioSummaryUseCase := portfolioUsecase.NewGetPortfolioSummaryUsecase(*positionAggregationUseCase, *balanceUsecase)

	//====== Market data begin============
	marketDataDbRepo := mktPersistence.NewMarketDataRepository(db)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	cacheHandler := cache.NewRedisCacheHandler(redisClient)

	// Step 3: Wrap database repository with cache-aside pattern
	marketDataRepo := mktCache.NewMarketDataCacheRepository(
		marketDataDbRepo, // Database repository
		cacheHandler,     // Your cache handler
		5*time.Minute,    // TTL for cached data
	)

	// Step 4: Create use case with cached repository
	marketDataUsecase := mktUsecase.NewGetMarketDataUseCase(marketDataRepo)

	// Step 5: Extract cache manager for admin operations
	cacheManager := mktCache.GetCacheManager(marketDataRepo)
	//====== Market data end============

	watchRepo := watchPersistence.NewWatchlistRepository(db)
	watchlistUsecase := watchlistUsecase.NewGetWatchlistUsecase(watchRepo, marketDataUsecase)

	return &containerImpl{
		AucService:                 aucService,
		PositionAggregationUseCase: positionAggregationUseCase,
		BalanceUsecase:             balanceUsecase,
		PortfolioSummaryUsecase:    portfolioSummaryUseCase,
		MarketDataUsecase:          marketDataUsecase,
		MarketDataCacheManager:     cacheManager,
		WatchlistUsecase:           watchlistUsecase,
	}, nil
}
