package di

import (
	"time"

	balUsecase "HubInvestments/balance/application/usecase"
	balancePersistence "HubInvestments/balance/infra/persistence"
	mktUsecase "HubInvestments/market_data/application/usecase"
	mktCache "HubInvestments/market_data/infra/cache"
	mktPersistence "HubInvestments/market_data/infra/persistence"
	portfolioUsecase "HubInvestments/portfolio_summary/application/usecase"
	posService "HubInvestments/position/application/service"
	posUsecase "HubInvestments/position/application/usecase"
	positionPersistence "HubInvestments/position/infra/persistence"
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
}

type containerImpl struct {
	AucService                 *posService.AucService
	PositionAggregationUseCase *posUsecase.GetPositionAggregationUseCase
	BalanceUsecase             *balUsecase.GetBalanceUseCase
	PortfolioSummaryUsecase    portfolioUsecase.PortfolioSummaryUsecase
	MarketDataUsecase          mktUsecase.IGetMarketDataUsecase
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

	// Create market data repository with cache-aside pattern
	// Step 1: Create base database repository
	marketDataDbRepo := mktPersistence.NewMarketDataRepository(db)

	// Step 2: Create Redis client and cache handler
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	cacheHandler := cache.NewRedisCacheHandler(redisClient)

	// Step 3: Wrap database repository with cache-aside pattern
	marketDataRepo := mktCache.NewMarketDataCacheRepository(
		marketDataDbRepo,    // Database repository
		cacheHandler,        // Your cache handler
		5*time.Minute,      // TTL for cached data
	)

	// Step 4: Create use case with cached repository
	marketDataUsecase := mktUsecase.NewGetMarketDataUseCase(marketDataRepo)

	return &containerImpl{
		AucService:                 aucService,
		PositionAggregationUseCase: positionAggregationUseCase,
		BalanceUsecase:             balanceUsecase,
		PortfolioSummaryUsecase:    portfolioSummaryUseCase,
		MarketDataUsecase:          marketDataUsecase,
	}, nil
}
