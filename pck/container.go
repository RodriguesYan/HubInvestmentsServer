package di

import (
	"fmt"
	"os"
	"strconv"
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
	orderService "HubInvestments/internal/order_mngmt_system/domain/service"
	orderMktClient "HubInvestments/internal/order_mngmt_system/infra/external"
	orderIdempotency "HubInvestments/internal/order_mngmt_system/infra/idempotency"
	orderRabbitMQ "HubInvestments/internal/order_mngmt_system/infra/messaging/rabbitmq"
	orderPersistence "HubInvestments/internal/order_mngmt_system/infra/persistence"
	orderWorker "HubInvestments/internal/order_mngmt_system/infra/worker"
	portfolioUsecase "HubInvestments/internal/portfolio_summary/application/usecase"
	posUsecase "HubInvestments/internal/position/application/usecase"
	positionPersistence "HubInvestments/internal/position/infra/persistence"
	quotesService "HubInvestments/internal/realtime_quotes/application/service"
	quotesAssetService "HubInvestments/internal/realtime_quotes/domain/service"
	quotesWebSocket "HubInvestments/internal/realtime_quotes/infra/websocket"
	quotesHttp "HubInvestments/internal/realtime_quotes/presentation/http"
	watchlistUsecase "HubInvestments/internal/watchlist/application/usecase"
	watchPersistence "HubInvestments/internal/watchlist/infra/persistence"
	"HubInvestments/shared/infra/cache"
	"HubInvestments/shared/infra/database"
	"HubInvestments/shared/infra/messaging"
	"HubInvestments/shared/infra/websocket"

	"github.com/redis/go-redis/v9"
)

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

	// Order Management System - Infrastructure
	GetOrderProducer() *orderRabbitMQ.OrderProducer
	GetOrderWorkerManager() *orderWorker.WorkerManager

	// Cache management methods for admin operations
	InvalidateMarketDataCache(symbols []string) error
	WarmMarketDataCache(symbols []string) error

	// Messaging infrastructure
	GetMessageHandler() messaging.MessageHandler

	// WebSocket infrastructure
	GetWebSocketManager() websocket.WebSocketManager

	// Realtime Quotes System
	GetAssetDataService() *quotesAssetService.AssetDataService
	GetPriceOscillationService() *quotesService.PriceOscillationService
	GetRealtimeQuotesWebSocketHandler() *quotesWebSocket.RealtimeQuotesWebSocketHandler
	GetQuotesHandler() *quotesHttp.QuotesHandler

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

	// Messaging infrastructure
	MessageHandler messaging.MessageHandler

	// WebSocket infrastructure
	WebSocketManager websocket.WebSocketManager

	// Order Management System - Market Data Integration
	OrderMarketDataClient orderMktClient.IMarketDataClient

	// Order Management System - Repository
	OrderRepository orderRepository.IOrderRepository

	// Order Management System - Use Cases
	SubmitOrderUseCase    orderUsecase.ISubmitOrderUseCase
	GetOrderStatusUseCase orderUsecase.IGetOrderStatusUseCase
	CancelOrderUseCase    orderUsecase.ICancelOrderUseCase
	ProcessOrderUseCase   orderUsecase.IProcessOrderUseCase

	// Order Management System - Infrastructure
	OrderProducer      *orderRabbitMQ.OrderProducer
	OrderWorkerManager *orderWorker.WorkerManager
	IdempotencyService orderService.IIdempotencyService

	// Realtime Quotes System
	AssetDataService               *quotesAssetService.AssetDataService
	PriceOscillationService        *quotesService.PriceOscillationService
	RealtimeQuotesWebSocketHandler *quotesWebSocket.RealtimeQuotesWebSocketHandler
	QuotesHandler                  *quotesHttp.QuotesHandler
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

func (c *containerImpl) GetMessageHandler() messaging.MessageHandler {
	return c.MessageHandler
}

func (c *containerImpl) GetWebSocketManager() websocket.WebSocketManager {
	return c.WebSocketManager
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

func (c *containerImpl) GetOrderProducer() *orderRabbitMQ.OrderProducer {
	return c.OrderProducer
}

func (c *containerImpl) GetOrderWorkerManager() *orderWorker.WorkerManager {
	return c.OrderWorkerManager
}

func (c *containerImpl) GetAssetDataService() *quotesAssetService.AssetDataService {
	return c.AssetDataService
}

func (c *containerImpl) GetPriceOscillationService() *quotesService.PriceOscillationService {
	return c.PriceOscillationService
}

func (c *containerImpl) GetRealtimeQuotesWebSocketHandler() *quotesWebSocket.RealtimeQuotesWebSocketHandler {
	return c.RealtimeQuotesWebSocketHandler
}

func (c *containerImpl) GetQuotesHandler() *quotesHttp.QuotesHandler {
	return c.QuotesHandler
}

// Close gracefully shuts down all resources managed by the container
func (c *containerImpl) Close() error {
	var errors []error

	// Stop price oscillation service first
	if c.PriceOscillationService != nil {
		c.PriceOscillationService.Stop()
	}

	// Stop worker manager first to gracefully shutdown workers
	if c.OrderWorkerManager != nil {
		if err := c.OrderWorkerManager.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop order worker manager: %w", err))
		}
	}

	// Close order producer
	if c.OrderProducer != nil {
		if err := c.OrderProducer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close order producer: %w", err))
		}
	}

	// Close message handler connection
	if c.MessageHandler != nil {
		if err := c.MessageHandler.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close message handler: %w", err))
		}
	}

	// Close WebSocket manager
	if c.WebSocketManager != nil {
		if err := c.WebSocketManager.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close websocket manager: %w", err))
		}
	}

	// Close order market data client gRPC connection
	if c.OrderMarketDataClient != nil {
		if err := c.OrderMarketDataClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close order market data client: %w", err))
		}
	}

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

	//====== Messaging Infrastructure begin============
	// Create RabbitMQ message handler with environment-based configuration
	messageHandlerConfig := messaging.NewMessageHandlerConfigFromEnv()
	messageHandler, err := messaging.NewRabbitMQMessageHandler(messageHandlerConfig)
	if err != nil {
		// Log the error but don't fail container creation - messaging is optional for development
		fmt.Printf("Warning: Failed to create RabbitMQ message handler: %v\n", err)
		fmt.Println("Continuing without messaging infrastructure. Some features may not work.")
		messageHandler = nil
	}
	//====== Messaging Infrastructure end============

	//====== WebSocket Infrastructure begin============
	// Create WebSocket manager with environment-based configuration
	webSocketConfig := websocket.DefaultWebSocketManagerConfig()

	// Override defaults with environment variables if needed
	if maxConnStr := os.Getenv("WEBSOCKET_MAX_CONNECTIONS"); maxConnStr != "" {
		if maxConn, err := strconv.Atoi(maxConnStr); err == nil {
			webSocketConfig.MaxConnections = maxConn
		}
	}

	webSocketManager := websocket.NewGorillaWebSocketManager(webSocketConfig)
	//====== WebSocket Infrastructure end============

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
	// Create order repository with database connection
	orderRepo := orderPersistence.NewOrderRepository(db)

	// Create idempotency service with Redis repository
	idempotencyRepo := orderIdempotency.NewRedisIdempotencyRepository(cacheHandler)
	idempotencyService := orderService.NewIdempotencyService(idempotencyRepo)

	// Create order management use cases with dependencies
	submitOrderUseCase := orderUsecase.NewSubmitOrderUseCase(orderRepo, orderMarketDataClient, idempotencyService)
	getOrderStatusUseCase := orderUsecase.NewGetOrderStatusUseCase(orderRepo, orderMarketDataClient)
	cancelOrderUseCase := orderUsecase.NewCancelOrderUseCase(orderRepo)
	processOrderUseCase := orderUsecase.NewProcessOrderUseCase(orderRepo, orderMarketDataClient)
	//====== Order Management System Use Cases end============

	//====== Order Management Infrastructure begin============
	var orderProducer *orderRabbitMQ.OrderProducer
	var orderWorkerManager *orderWorker.WorkerManager

	// Only create producer and worker manager if messaging is available
	if messageHandler != nil {
		orderProducer = orderRabbitMQ.NewOrderProducer(messageHandler)

		// Create worker manager with default configuration
		workerManagerConfig := orderWorker.DefaultWorkerManagerConfig()
		orderWorkerManager = orderWorker.NewWorkerManager(
			processOrderUseCase,
			messageHandler,
			workerManagerConfig,
		)

		// Start worker manager in background
		go func() {
			if err := orderWorkerManager.Start(); err != nil {
				fmt.Printf("Warning: Failed to start order worker manager: %v\n", err)
			}
		}()
	}
	//====== Order Management Infrastructure end============

	//====== Realtime Quotes System begin============
	// Create asset data service with predefined stocks and ETFs
	assetDataService := quotesAssetService.NewAssetDataService()

	// Create price oscillation service
	priceOscillationService := quotesService.NewPriceOscillationService(assetDataService)

	// Start price oscillation in background
	priceOscillationService.Start()

	// Create WebSocket handler for realtime quotes with same auth service as HTTP handlers
	realtimeQuotesWebSocketHandler := quotesWebSocket.NewRealtimeQuotesWebSocketHandler(
		webSocketManager,
		priceOscillationService,
		authService,
	)

	// Create HTTP handler for quotes API
	quotesHandler := quotesHttp.NewQuotesHandler(assetDataService)
	//====== Realtime Quotes System end============

	watchRepo := watchPersistence.NewWatchlistRepository(db)
	watchlistUsecase := watchlistUsecase.NewGetWatchlistUsecase(watchRepo, marketDataUsecase)

	return &containerImpl{
		PositionAggregationUseCase:     positionAggregationUseCase,
		BalanceUsecase:                 balanceUsecase,
		PortfolioSummaryUsecase:        portfolioSummaryUseCase,
		MarketDataUsecase:              marketDataUsecase,
		MarketDataCacheManager:         cacheManager,
		WatchlistUsecase:               watchlistUsecase,
		LoginUsecase:                   loginUsecase,
		AuthService:                    authService,
		MessageHandler:                 messageHandler,
		WebSocketManager:               webSocketManager,
		OrderMarketDataClient:          orderMarketDataClient,
		OrderRepository:                orderRepo,
		SubmitOrderUseCase:             submitOrderUseCase,
		GetOrderStatusUseCase:          getOrderStatusUseCase,
		CancelOrderUseCase:             cancelOrderUseCase,
		ProcessOrderUseCase:            processOrderUseCase,
		OrderProducer:                  orderProducer,
		OrderWorkerManager:             orderWorkerManager,
		IdempotencyService:             idempotencyService,
		AssetDataService:               assetDataService,
		PriceOscillationService:        priceOscillationService,
		RealtimeQuotesWebSocketHandler: realtimeQuotesWebSocketHandler,
		QuotesHandler:                  quotesHandler,
	}, nil
}
