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
	orderUsecase "HubInvestments/internal/order_mngmt_system/application/usecase"
	orderRepository "HubInvestments/internal/order_mngmt_system/domain/repository"
	orderService "HubInvestments/internal/order_mngmt_system/domain/service"
	orderMktClient "HubInvestments/internal/order_mngmt_system/infra/external"
	orderIdempotency "HubInvestments/internal/order_mngmt_system/infra/idempotency"
	orderMessaging "HubInvestments/internal/order_mngmt_system/infra/messaging"
	orderRabbitMQ "HubInvestments/internal/order_mngmt_system/infra/messaging/rabbitmq"
	orderPersistence "HubInvestments/internal/order_mngmt_system/infra/persistence"
	orderWorker "HubInvestments/internal/order_mngmt_system/infra/worker"
	portfolioUsecase "HubInvestments/internal/portfolio_summary/application/usecase"
	posUsecase "HubInvestments/internal/position/application/usecase"
	positionPersistence "HubInvestments/internal/position/infra/persistence"
	positionWorker "HubInvestments/internal/position/infra/worker"
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
	GetCreatePositionUseCase() posUsecase.ICreatePositionUseCase
	GetUpdatePositionUseCase() posUsecase.IUpdatePositionUseCase
	GetClosePositionUseCase() posUsecase.IClosePositionUseCase
	GetBalanceUseCase() *balUsecase.GetBalanceUseCase
	GetPortfolioSummaryUsecase() portfolioUsecase.PortfolioSummaryUsecase
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

	// Position Management System - Infrastructure
	GetPositionWorkerManager() *positionWorker.PositionUpdateWorker

	// Messaging infrastructure
	GetMessageHandler() messaging.MessageHandler

	// WebSocket infrastructure
	GetWebSocketManager() websocket.WebSocketManager

	// Lifecycle management
	Close() error
}

type containerImpl struct {
	AuthService                auth.IAuthService
	PositionAggregationUseCase *posUsecase.GetPositionAggregationUseCase
	CreatePositionUseCase      posUsecase.ICreatePositionUseCase
	UpdatePositionUseCase      posUsecase.IUpdatePositionUseCase
	ClosePositionUseCase       posUsecase.IClosePositionUseCase
	BalanceUsecase             *balUsecase.GetBalanceUseCase
	PortfolioSummaryUsecase    portfolioUsecase.PortfolioSummaryUsecase
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
	OrderProducer       *orderRabbitMQ.OrderProducer
	OrderEventPublisher orderMessaging.IEventPublisher
	OrderWorkerManager  *orderWorker.WorkerManager
	IdempotencyService  orderService.IIdempotencyService

	// Position Management System - Infrastructure
	PositionWorkerManager *positionWorker.PositionUpdateWorker
}

func (c *containerImpl) GetAuthService() auth.IAuthService {
	return c.AuthService
}

func (c *containerImpl) GetPositionAggregationUseCase() *posUsecase.GetPositionAggregationUseCase {
	return c.PositionAggregationUseCase
}

func (c *containerImpl) GetCreatePositionUseCase() posUsecase.ICreatePositionUseCase {
	return c.CreatePositionUseCase
}

func (c *containerImpl) GetUpdatePositionUseCase() posUsecase.IUpdatePositionUseCase {
	return c.UpdatePositionUseCase
}

func (c *containerImpl) GetClosePositionUseCase() posUsecase.IClosePositionUseCase {
	return c.ClosePositionUseCase
}

func (c *containerImpl) GetBalanceUseCase() *balUsecase.GetBalanceUseCase {
	return c.BalanceUsecase
}

func (c *containerImpl) GetPortfolioSummaryUsecase() portfolioUsecase.PortfolioSummaryUsecase {
	return c.PortfolioSummaryUsecase
}

func (c *containerImpl) DoLoginUsecase() doLoginUsecase.IDoLoginUsecase {
	return c.LoginUsecase
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

func (c *containerImpl) GetPositionWorkerManager() *positionWorker.PositionUpdateWorker {
	return c.PositionWorkerManager
}

// Close gracefully shuts down all resources managed by the container
func (c *containerImpl) Close() error {
	var errors []error

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
	// Read database configuration from environment variables
	dbConfig := database.ConnectionConfig{
		Driver:   "postgres",
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Database: os.Getenv("DB_NAME"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  getEnvWithDefault("DB_SSLMODE", "disable"),
	}

	// Use default config if environment variables are not set
	if dbConfig.Host == "" {
		dbConfig = database.DefaultConfig()
	}

	db, err := database.CreateDatabaseConnectionWithConfig(dbConfig)
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

	// Position Management Use Cases
	createPositionUseCase := posUsecase.NewCreatePositionUseCase(positionRepo)
	updatePositionUseCase := posUsecase.NewUpdatePositionUseCase(positionRepo)
	closePositionUseCase := posUsecase.NewClosePositionUseCase(positionRepo)

	balanceRepo := balancePersistence.NewBalanceRepository(db)
	balanceUsecase := balUsecase.NewGetBalanceUseCase(balanceRepo)
	portfolioSummaryUseCase := portfolioUsecase.NewGetPortfolioSummaryUsecase(*positionAggregationUseCase, *balanceUsecase)

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

	// Create Redis client for idempotency
	redisHost := getEnvWithDefault("REDIS_HOST", "localhost")
	redisPort := getEnvWithDefault("REDIS_PORT", "6379")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB, _ := strconv.Atoi(getEnvWithDefault("REDIS_DB", "0"))

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	cacheHandler := cache.NewRedisCacheHandler(redisClient)

	// Create idempotency service with Redis repository
	idempotencyRepo := orderIdempotency.NewRedisIdempotencyRepository(cacheHandler)
	idempotencyService := orderService.NewIdempotencyService(idempotencyRepo)

	// Create event publisher for order domain events
	var orderEventPublisher orderMessaging.IEventPublisher
	if messageHandler != nil {
		orderEventPublisher = orderMessaging.NewEventPublisher(messageHandler, "orders.events")
	}

	// Create order management use cases with dependencies
	// Note: SubmitOrderUseCase will be created after OrderProducer is available
	getOrderStatusUseCase := orderUsecase.NewGetOrderStatusUseCase(orderRepo, orderMarketDataClient)
	cancelOrderUseCase := orderUsecase.NewCancelOrderUseCase(orderRepo)
	processOrderUseCase := orderUsecase.NewProcessOrderUseCase(orderRepo, orderMarketDataClient, orderEventPublisher)
	//====== Order Management System Use Cases end============

	//====== Order Management Infrastructure begin============
	var orderProducer *orderRabbitMQ.OrderProducer
	var orderWorkerManager *orderWorker.WorkerManager
	var submitOrderUseCase orderUsecase.ISubmitOrderUseCase

	// Only create producer and worker manager if messaging is available
	if messageHandler != nil {
		orderProducer = orderRabbitMQ.NewOrderProducer(messageHandler)

		// Create SubmitOrderUseCase with OrderProducer dependency
		submitOrderUseCase = orderUsecase.NewSubmitOrderUseCase(orderRepo, orderMarketDataClient, idempotencyService, orderProducer)

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
	} else {
		// Create SubmitOrderUseCase without OrderProducer when messaging is not available
		submitOrderUseCase = orderUsecase.NewSubmitOrderUseCase(orderRepo, orderMarketDataClient, idempotencyService, nil)
	}
	//====== Order Management Infrastructure end============

	//====== Position Management Infrastructure begin============
	var positionWorkerManager *positionWorker.PositionUpdateWorker

	// Only create position worker manager if messaging is available
	if messageHandler != nil {
		// Create position worker with default configuration
		workerConfig := positionWorker.DefaultPositionWorkerConfig("position-worker-1")
		positionWorkerManager = positionWorker.NewPositionUpdateWorker(
			"position-worker-1",
			createPositionUseCase,
			updatePositionUseCase,
			closePositionUseCase,
			positionRepo,
			messageHandler,
			workerConfig,
		)

		// Start position worker in background
		go func() {
			if err := positionWorkerManager.Start(); err != nil {
				fmt.Printf("Warning: Failed to start position worker manager: %v\n", err)
			}
		}()
	}
	//====== Position Management Infrastructure end============

	watchRepo := watchPersistence.NewWatchlistRepository(db)
	watchlistUsecase := watchlistUsecase.NewGetWatchlistUsecase(watchRepo, orderMarketDataClient)

	return &containerImpl{
		PositionAggregationUseCase: positionAggregationUseCase,
		CreatePositionUseCase:      createPositionUseCase,
		UpdatePositionUseCase:      updatePositionUseCase,
		ClosePositionUseCase:       closePositionUseCase,
		BalanceUsecase:             balanceUsecase,
		PortfolioSummaryUsecase:    portfolioSummaryUseCase,
		WatchlistUsecase:           watchlistUsecase,
		LoginUsecase:               loginUsecase,
		AuthService:                authService,
		MessageHandler:             messageHandler,
		WebSocketManager:           webSocketManager,
		OrderMarketDataClient:      orderMarketDataClient,
		OrderRepository:            orderRepo,
		SubmitOrderUseCase:         submitOrderUseCase,
		GetOrderStatusUseCase:      getOrderStatusUseCase,
		CancelOrderUseCase:         cancelOrderUseCase,
		ProcessOrderUseCase:        processOrderUseCase,
		OrderProducer:              orderProducer,
		OrderEventPublisher:        orderEventPublisher,
		OrderWorkerManager:         orderWorkerManager,
		IdempotencyService:         idempotencyService,
		PositionWorkerManager:      positionWorkerManager,
	}, nil
}

// getEnvWithDefault gets an environment variable with a fallback default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
