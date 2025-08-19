package di

import (
	"fmt"
	"os"
	"time"

	"HubInvestments/internal/auth"
	"HubInvestments/internal/auth/token"
	balUsecase "HubInvestments/internal/balance/application/usecase"
	balancePersistence "HubInvestments/internal/balance/infra/persistence"
	doLoginUsecase "HubInvestments/internal/login/application/usecase"
	loginPersistence "HubInvestments/internal/login/infra/persistense"
	mktUsecase "HubInvestments/internal/market_data/application/usecase"
	mktCache "HubInvestments/internal/market_data/infra/cache"
	mktPersistence "HubInvestments/internal/market_data/infra/persistence"
	orderUsecase "HubInvestments/internal/order_mngmt_system/application/usecase"
	orderRepository "HubInvestments/internal/order_mngmt_system/domain/repository"
	orderMktClient "HubInvestments/internal/order_mngmt_system/infra/external"
	portfolioUsecase "HubInvestments/internal/portfolio_summary/application/usecase"
	posUsecase "HubInvestments/internal/position/application/usecase"
	positionPersistence "HubInvestments/internal/position/infra/persistence"
	watchlistUsecase "HubInvestments/internal/watchlist/application/usecase"
	watchPersistence "HubInvestments/internal/watchlist/infra/persistence"
	"HubInvestments/shared/infra/cache"
	"HubInvestments/shared/infra/database"

	"github.com/redis/go-redis/v9"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// MockOrderRepository is a temporary mock implementation for development
// TODO: Replace with actual database implementation
type MockOrderRepository struct{}

func (m *MockOrderRepository) Save(order *domain.Order) error {
	// Mock implementation - in real implementation this would save to database
	return nil
}

func (m *MockOrderRepository) FindByID(id string) (*domain.Order, error) {
	// Mock implementation - in real implementation this would query database
	return nil, fmt.Errorf("order not found: %s", id)
}

func (m *MockOrderRepository) FindByUserID(userID string) ([]*domain.Order, error) {
	// Mock implementation - in real implementation this would query database
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) UpdateStatus(id string, status domain.OrderStatus) error {
	// Mock implementation - in real implementation this would update database
	return nil
}

func (m *MockOrderRepository) FindOrderHistory(userID string, limit int, offset int) ([]*domain.Order, error) {
	// Mock implementation - in real implementation this would query database with pagination
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) FindOrdersByStatus(status domain.OrderStatus) ([]*domain.Order, error) {
	// Mock implementation - in real implementation this would query database
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) FindOrdersBySymbol(symbol string) ([]*domain.Order, error) {
	// Mock implementation - in real implementation this would query database
	return []*domain.Order{}, nil
}

// Additional methods required by IOrderRepository interface
func (m *MockOrderRepository) UpdateOrderWithExecution(orderID string, executionPrice float64, executedAt time.Time) error {
	return nil
}

func (m *MockOrderRepository) FindByUserIDAndStatus(userID string, status domain.OrderStatus) ([]*domain.Order, error) {
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) FindByUserIDAndSymbol(userID string, symbol string) ([]*domain.Order, error) {
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) FindByUserIDWithPagination(userID string, limit, offset int) ([]*domain.Order, error) {
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) FindByUserIDAndDateRange(userID string, startDate, endDate time.Time) ([]*domain.Order, error) {
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) FindActiveOrdersByUserID(userID string) ([]*domain.Order, error) {
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) FindActiveOrdersByUserIDAndSymbol(userID string, symbol string) ([]*domain.Order, error) {
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) CountOrdersByUserID(userID string) (int, error) {
	return 0, nil
}

func (m *MockOrderRepository) CountOrdersByUserIDAndStatus(userID string, status domain.OrderStatus) (int, error) {
	return 0, nil
}

func (m *MockOrderRepository) FindOrdersForProcessing(limit int) ([]*domain.Order, error) {
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) FindExpiredOrders(expiredBefore time.Time) ([]*domain.Order, error) {
	return []*domain.Order{}, nil
}

func (m *MockOrderRepository) Delete(orderID string) error {
	return nil
}

func (m *MockOrderRepository) ExistsOrderByID(orderID string) (bool, error) {
	return false, nil
}

func (m *MockOrderRepository) FindOrdersNeedingCancellation(beforeTime time.Time) ([]*domain.Order, error) {
	return []*domain.Order{}, nil
}

type Container interface {
	DoLoginUsecase() doLoginUsecase.IDoLoginUsecase
	GetAuthService() auth.IAuthService
	GetPositionAggregationUseCase() *posUsecase.GetPositionAggregationUseCase
	GetBalanceUseCase() *balUsecase.GetBalanceUseCase
	GetPortfolioSummaryUsecase() portfolioUsecase.PortfolioSummaryUsecase
	GetMarketDataUsecase() mktUsecase.IGetMarketDataUsecase
	GetWatchlistUsecase() watchlistUsecase.IGetWatchlistUsecase

	// Order Management System - Market Data Integration
	GetOrderMarketDataClient() orderMktClient.IMarketDataClient

	// Order Management System - Use Cases
	GetSubmitOrderUseCase() orderUsecase.ISubmitOrderUseCase
	GetGetOrderStatusUseCase() orderUsecase.IGetOrderStatusUseCase
	GetCancelOrderUseCase() orderUsecase.ICancelOrderUseCase
	GetProcessOrderUseCase() orderUsecase.IProcessOrderUseCase

	// Cache management methods for admin operations
	InvalidateMarketDataCache(symbols []string) error
	WarmMarketDataCache(symbols []string) error

	// Lifecycle management
	Close() error
}

type containerImpl struct {
	AuthService                auth.IAuthService
	PositionAggregationUseCase *posUsecase.GetPositionAggregationUseCase
	BalanceUsecase             *balUsecase.GetBalanceUseCase
	PortfolioSummaryUsecase    portfolioUsecase.PortfolioSummaryUsecase
	MarketDataUsecase          mktUsecase.IGetMarketDataUsecase
	MarketDataCacheManager     mktCache.CacheManager
	WatchlistUsecase           watchlistUsecase.IGetWatchlistUsecase
	LoginUsecase               doLoginUsecase.IDoLoginUsecase

	// Order Management System - Market Data Integration
	OrderMarketDataClient orderMktClient.IMarketDataClient

	// Order Management System - Repository
	OrderRepository orderRepository.IOrderRepository

	// Order Management System - Use Cases
	SubmitOrderUseCase    orderUsecase.ISubmitOrderUseCase
	GetOrderStatusUseCase orderUsecase.IGetOrderStatusUseCase
	CancelOrderUseCase    orderUsecase.ICancelOrderUseCase
	ProcessOrderUseCase   orderUsecase.IProcessOrderUseCase
}

func (c *containerImpl) GetAuthService() auth.IAuthService {
	return c.AuthService
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

func (c *containerImpl) DoLoginUsecase() doLoginUsecase.IDoLoginUsecase {
	return c.LoginUsecase
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

func (c *containerImpl) GetOrderMarketDataClient() orderMktClient.IMarketDataClient {
	return c.OrderMarketDataClient
}

func (c *containerImpl) GetSubmitOrderUseCase() orderUsecase.ISubmitOrderUseCase {
	return c.SubmitOrderUseCase
}

func (c *containerImpl) GetGetOrderStatusUseCase() orderUsecase.IGetOrderStatusUseCase {
	return c.GetOrderStatusUseCase
}

func (c *containerImpl) GetCancelOrderUseCase() orderUsecase.ICancelOrderUseCase {
	return c.CancelOrderUseCase
}

func (c *containerImpl) GetProcessOrderUseCase() orderUsecase.IProcessOrderUseCase {
	return c.ProcessOrderUseCase
}

// Close gracefully shuts down all resources managed by the container
func (c *containerImpl) Close() error {
	var errors []error

	// Close order market data client gRPC connection
	if c.OrderMarketDataClient != nil {
		if err := c.OrderMarketDataClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close order market data client: %w", err))
		}
	}

	// If there are any errors, return the first one
	// In a production system, you might want to return all errors or use a multi-error type
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func NewContainer() (Container, error) {
	// Create database connection using the abstraction layer
	db, err := database.CreateDatabaseConnection()
	if err != nil {
		return nil, err
	}

	loginRepo := loginPersistence.NewLoginRepository(db)
	loginUsecase := doLoginUsecase.NewDoLoginUsecase(loginRepo)
	tokenService := token.NewTokenService()
	authService := auth.NewAuthService(tokenService)

	// Create repositories using the database abstraction
	positionRepo := positionPersistence.NewPositionRepository(db)
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

	//====== Order Management Market Data Client begin============
	// Create market data client for order management system with environment-based configuration
	marketDataServerAddr := os.Getenv("MARKET_DATA_GRPC_SERVER")
	if marketDataServerAddr == "" {
		marketDataServerAddr = "localhost:50051" // Default for development
	}

	orderMarketDataClientConfig := orderMktClient.MarketDataClientConfig{
		ServerAddress: marketDataServerAddr,
		Timeout:       30 * time.Second,
	}

	orderMarketDataClient, err := orderMktClient.NewMarketDataClient(orderMarketDataClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create order market data client: %w", err)
	}
	//====== Order Management Market Data Client end============

	//====== Order Management System Use Cases begin============
	// Create mock order repository (TODO: Replace with actual database implementation)
	orderRepo := &MockOrderRepository{}

	// Create order management use cases with dependencies
	submitOrderUseCase := orderUsecase.NewSubmitOrderUseCase(orderRepo, orderMarketDataClient)
	getOrderStatusUseCase := orderUsecase.NewGetOrderStatusUseCase(orderRepo, orderMarketDataClient)
	cancelOrderUseCase := orderUsecase.NewCancelOrderUseCase(orderRepo)
	processOrderUseCase := orderUsecase.NewProcessOrderUseCase(orderRepo, orderMarketDataClient)
	//====== Order Management System Use Cases end============

	watchRepo := watchPersistence.NewWatchlistRepository(db)
	watchlistUsecase := watchlistUsecase.NewGetWatchlistUsecase(watchRepo, marketDataUsecase)

	return &containerImpl{
		PositionAggregationUseCase: positionAggregationUseCase,
		BalanceUsecase:             balanceUsecase,
		PortfolioSummaryUsecase:    portfolioSummaryUseCase,
		MarketDataUsecase:          marketDataUsecase,
		MarketDataCacheManager:     cacheManager,
		WatchlistUsecase:           watchlistUsecase,
		LoginUsecase:               loginUsecase,
		AuthService:                authService,
		OrderMarketDataClient:      orderMarketDataClient,
		OrderRepository:            orderRepo,
		SubmitOrderUseCase:         submitOrderUseCase,
		GetOrderStatusUseCase:      getOrderStatusUseCase,
		CancelOrderUseCase:         cancelOrderUseCase,
		ProcessOrderUseCase:        processOrderUseCase,
	}, nil
}
