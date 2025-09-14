# Hub Investments Platform - Implementation Plan

## Implementation Roadmap Based on PRD

### ✅ Phase 1: Core Infrastructure (COMPLETED)
- [x] Basic authentication system with JWT tokens
- [x] Project structure with proper DDD implementation
- [x] Position service with clean architecture
- [x] Repository pattern implementation
- [x] Database schema for positions and instruments
- **Result**: Solid foundation with clean architecture and working authentication

### ⏳ Phase 2: Portfolio Summary Implementation (IN PROGRESS)
- [x] **Step 1**: Create Portfolio Domain Model
  - [x] Create `portfolio/domain/model/` directory structure
  - [x] Implement `portfolio_summary_model.go` with PortfolioSummaryModel struct
  - [x] Add imports for balance and position domain models
  - [x] Include calculated fields (TotalPortfolioValue, LastUpdated)
- [x] **Step 2**: Create Balance Use Case
  - [x] Create `balance/application/usecase/get_balance_usecase.go`
  - [x] Implement GetBalanceUseCase struct and Execute method
  - [x] Add proper error handling and validation
- [x] **Step 3**: Create Portfolio Use Case
  - [x] Create `portfolio/application/usecase/` directory
  - [x] Implement `get_portfolio_summary_usecase.go`
  - [x] Add dependency injection for Position and Balance use cases
  - [x] Implement orchestration logic for combining data
  - [x] Add business logic for calculating total portfolio value
  - [x] Include proper error handling and validation
- [x] **Step 4**: Create Portfolio Handler
  - [x] Create `portfolio/presentation/http/` directory
  - [x] Implement `portfolio_handler.go` with GetPortfolioSummary function
  - [x] Add authentication verification
  - [x] Implement proper HTTP error handling
  - [x] Add JSON serialization and response formatting
- [x] **Step 5**: Update Dependency Injection Container
  - [x] Add GetBalanceUseCase method to Container interface
  - [x] Add GetPortfolioSummaryUseCase method to Container interface
  - [x] Update containerImpl struct with new dependencies
  - [x] Modify NewContainer function to initialize new use cases
  - [x] Update TestContainer for testing support
- [x] **Step 6**: Add Portfolio Route to Main
  - [x] Import portfolio handler in main.go
  - [x] Add `/getPortfolioSummary` endpoint
  - [x] Wire up authentication and container dependencies
- [x] **Step 7**: Create Unit Tests
  - [x] Create `portfolio/presentation/http/portfolio_handler_test.go`
  - [x] Implement mock dependencies for testing
  - [x] Add test cases for success and error scenarios
  - [x] Test authentication and authorization flows
- [ ] **Step 8**: Documentation and Validation
  - [ ] Update API documentation with new endpoint
  - [x] Add example request/response in comments
  - [ ] Validate endpoint with real data
  - [ ] Performance testing with concurrent requests
- **Priority**: High - Core portfolio functionality combining position and balance data
- **Dependencies**: Balance Use Case implementation, existing Position Use Case
- **Result**: Single endpoint providing complete portfolio overview

### ⏳ Phase 5: Market Data Service Implementation
- [x] **Step 1**: Core Market Data Architecture (COMPLETED)
  - [x] Create market data domain models and repository interfaces
  - [x] Implement market data use case with clean architecture
  - [x] Create HTTP handler for REST API endpoints
  - [x] Add comprehensive unit tests (100% coverage for usecase, 94.7% for repository, 65% for handler)
  - [x] Integration with dependency injection container
- [x] **Step 2**: Redis Cache Aside Pattern Implementation (COMPLETED)
  - [x] **Step 2.1**: Cache Infrastructure Foundation (COMPLETED)
    - [x] Add Redis dependency: `go get github.com/redis/go-redis/v9`
    - [x] Create `shared/infra/cache/` directory structure
    - [x] Create `CacheHandler` interface with Get/Set/Delete operations
    - [x] Implement `RedisCacheHandler` basic structure with Redis client integration
    - [x] Create comprehensive README.md documentation with usage patterns, examples, and future extensibility
    - [x] Document cache key strategies and TTL recommendations
    - [x] Add testing patterns and mock implementation examples
    - [x] Complete Redis Set() and Delete() method implementations
    - [x] Fix Redis client reuse (Redis client properly injected via DI container)
    - [x] Add Redis connection to dependency injection container
    - [ ] Create Redis Docker container configuration for development
  - [x] **Step 2.2**: Cache Service Layer (Repository Layer) (COMPLETED)
    - [x] Create `market_data/infra/cache/` directory
    - [x] Implement `market_data_cache_repository.go` as a decorator/wrapper around existing repository
    - [x] Cache key strategy: `market_data:{symbol}` for individual symbols
    - [x] TTL strategy: 5 minutes for market data
    - [x] Implement cache aside pattern:
      ```
      1. Check Redis cache first
      2. If cache hit: return cached data
      3. If cache miss: fetch from database
      4. Store result in cache with TTL
      5. Return data
      ```
  - [x] **Step 2.3**: Cache Configuration and Management (COMPLETED)
    - [x] Add cache configuration with environment-friendly defaults
    - [x] Implement cache invalidation strategy for admin operations
    - [x] Add cache warming functionality for popular symbols
    - [x] Create cache health check via Redis ping
  - [x] **Step 2.4**: Error Handling and Fallback (COMPLETED)
    - [x] Implement graceful degradation when Redis is unavailable
    - [x] Ensure original functionality works even if cache layer fails
    - [x] Add comprehensive logging for cache hits/misses and errors
    - [x] Add proper error handling for cache operations
  - [x] **Step 2.5**: Testing and Validation (COMPLETED)
    - [x] Create unit tests for cache repository with mocked dependencies
    - [x] Add integration tests with Redis functionality
    - [x] Performance validation of cache effectiveness
    - [x] Test cache hit/miss scenarios and TTL expiration
  - [x] **Step 2.6**: Admin Cache Management (BONUS - COMPLETED)
    - [x] Create admin endpoints for cache invalidation
    - [x] Add cache warming endpoints for operational control
    - [x] Implement JWT authentication for admin cache operations
    - [x] Add proper HTTP status codes and JSON responses
- [x] **Step 3**: gRPC Service Implementation (COMPLETED)
  - [x] **Step 3.1**: Protocol Buffers Setup (COMPLETED)
    - [x] Add gRPC dependencies: `go get google.golang.org/grpc` and `go get google.golang.org/protobuf`
    - [x] Install protoc compiler and Go plugins
    - [x] Create `market_data/presentation/grpc/` directory structure
    - [x] Define `market_data.proto` file with service definitions:
      ```proto
      service MarketDataService {
        rpc GetMarketData(GetMarketDataRequest) returns (GetMarketDataResponse);
        rpc StreamMarketData(StreamMarketDataRequest) returns (stream MarketDataUpdate);
      }
      ```
  - [x] **Step 3.2**: gRPC Server Implementation (COMPLETED)
    - [x] Generate Go code from proto files: `protoc --go_out=. --go-grpc_out=. market_data.proto`
    - [x] Create `market_data_grpc_server.go` implementing the generated service interface
    - [x] Implement unary RPC for batch market data requests
    - [ ] Implement streaming RPC for real-time market data updates (placeholder implemented)
    - [x] Add proper error handling and gRPC status codes
  - [x] **Step 3.3**: gRPC Service Integration (COMPLETED)
    - [x] Update dependency injection container to include gRPC server
    - [x] Configure gRPC server startup in main.go (separate port from HTTP)
    - [x] Add gRPC interceptors for authentication, logging, and metrics
    - [x] Implement graceful shutdown for gRPC server
  - [x] **Step 3.4**: gRPC Client Library (COMPLETED)
    - [x] Create `market_data/client/` directory for gRPC client
    - [x] Implement `market_data_grpc_client.go` with connection pooling
    - [x] Add client-side caching and connection management
    - [x] Create client interface for easy mocking in tests
  - [x] **Step 3.5**: Authentication & Security (COMPLETED)
    - [x] Implement JWT authentication interceptors for gRPC
    - [x] Add proper authentication handling in gRPC metadata
    - [x] Ensure consistent authentication between HTTP and gRPC
    - [x] Add authentication error handling with proper gRPC status codes
  - [x] **Step 3.6**: Testing gRPC Implementation (COMPLETED)
    - [x] Create unit tests for gRPC server handlers
    - [x] Add integration tests for gRPC client-server communication
    - [x] Test authentication flow for gRPC requests
    - [x] Add comprehensive test coverage for gRPC functionality
- [x] **Step 4**: Architecture Integration and Optimization
  - [x] **Step 4.1**: Dual Protocol Support
    - [x] HTTP REST API for external clients (web, mobile apps)
    - [x] gRPC for internal service-to-service communication (watchlist service)
    - [x] Shared business logic between HTTP and gRPC handlers
    - [x] Consistent error handling and response formats
- [ ] **Implementation Architecture Overview**:
  ```
  External Clients (Web/Mobile)
           ↓ HTTP/REST
  ┌─────────────────────────┐
  │  HTTP Handler Layer     │
  └─────────────────────────┘
           ↓
  Internal Services (Watchlist)
           ↓ gRPC
  ┌─────────────────────────┐
  │   gRPC Server Layer     │
  └─────────────────────────┘
           ↓
  ┌─────────────────────────┐
  │   Use Case Layer        │ ← Shared business logic
  └─────────────────────────┘
           ↓
  ┌─────────────────────────┐
  │  Cache Repository       │ ← Redis Cache Aside
  │  (Decorator Pattern)    │
  └─────────────────────────┘
           ↓
  ┌─────────────────────────┐
  │ Database Repository     │ ← PostgreSQL
  └─────────────────────────┘
  ```
- **Priority**: High - Core business functionality with performance optimization
- **Dependencies**: Existing market data implementation, Redis infrastructure, gRPC tooling
- **Performance Targets**: 
  - < 50ms response time with cache hits
  - < 200ms response time with cache misses
  - Support 10,000+ concurrent gRPC connections
  - 95%+ cache hit ratio for popular symbols

### ⏳ Phase 6: Order Management System
**Directory Structure:**
```
internal/order_mngmt_system/
├── domain/
│   ├── model/
│   │   ├── order.go                    # Order aggregate root
│   │   ├── order_status.go            # Order status value object  
│   │   ├── order_type.go              # Order type value object
│   │   ├── order_side.go              # Order side value object (BUY/SELL)
│   │   └── order_events.go            # Domain events
│   ├── repository/
│   │   └── order_repository.go        # Order repository interface
│   └── service/
│       ├── order_validation_service.go # Business validation logic
│       ├── risk_management_service.go  # Risk management logic
│       └── order_pricing_service.go    # Pricing and execution logic
├── application/
│   ├── usecase/
│   │   ├── submit_order_usecase.go     # Submit order use case
│   │   ├── get_order_status_usecase.go # Get order status use case
│   │   ├── cancel_order_usecase.go     # Cancel order use case
│   │   └── process_order_usecase.go    # Process order use case (worker)
│   └── command/
│       ├── submit_order_command.go     # Command objects
│       └── cancel_order_command.go
├── infra/
│   ├── persistence/
│   │   ├── order_repository.go         # Database implementation
│   │   └── dto/
│   │       ├── order_dto.go           # Data transfer objects
│   │       └── mapper.go              # DTO-Domain mapping
│   ├── external/
│   │   └── market_data_client.go      # Market data gRPC client wrapper
│   ├── messaging/
│   │   └── rabbitmq/
│   │       ├── order_producer.go      # RabbitMQ producer
│   │       ├── order_consumer.go      # RabbitMQ consumer
│   │       ├── connection_manager.go  # RabbitMQ connection management
│   │       └── queue_config.go        # Queue configuration
│   └── worker/
│       ├── order_worker.go            # Order processing worker
│       └── worker_manager.go          # Worker lifecycle management
└── presentation/
    ├── http/
    │   ├── order_handler.go           # HTTP endpoints
    │   └── order_handler_test.go
    └── grpc/                          # Future gRPC endpoints
        └── order_grpc_server.go
```

- [x] **Step 1**: Core Order Domain Model <!-- requeset id from chat to take all context generated for this -> cc575b3a-52e3-4cf8-bd7a-48fd210c84dc -->
  - [x] Create `order_mngmt_system/domain/model/` directory structure
  - [x] Implement `order.go` with Order aggregate root (UUID, UserID, Symbol, Quantity, Price, Status, Timestamps)
  - [x] Create `order_status.go` value object (PENDING, PROCESSING, EXECUTED, FAILED, CANCELLED)
  - [x] Create `order_type.go` value object (MARKET, LIMIT, STOP_LOSS, etc.)
  - [x] Implement `order_events.go` for domain events (OrderSubmitted, OrderExecuted, OrderFailed)
  - [x] Add business logic methods (Validate, CanCancel, MarkAsExecuted, etc.)
  - [x] Add business logic for selling orders
- [x] **Step 2**: Order Repository Interface
  - [x] Create `order_mngmt_system/domain/repository/order_repository.go`
  - [x] Define interface methods: Save, FindByID, FindByUserID, UpdateStatus
  - [x] Add query methods for order history and filtering
- [x] **Step 3**: Domain Services
  - [x] Create `order_validation_service.go` for business validation rules
  - [x] Implement `risk_management_service.go` for risk checks (balance, limits, etc.)
  - [x] Add order pricing and execution logic services
- [x] **Step 4**: Market Data Integration via gRPC Client
  - [x] **Step 4.1**: Market Data Client Infrastructure
    - [x] Create `infra/external/market_data_client.go` 
    - [x] Implement wrapper around existing market data gRPC client
    - [x] Create interface `IMarketDataClient` in domain layer for dependency inversion
    - [x] Add methods: `GetAssetDetails(symbol)`, `ValidateSymbol(symbol)`, `GetCurrentPrice(symbol)`
    - [x] Include error handling for gRPC communication failures
  - [x] **Step 4.2**: Market Data Integration in Use Cases
    - [x] Update `submit_order_usecase.go` to validate symbol exists via market data service
    - [x] Add price validation against current market price (for limit orders)
    - [x] Check trading hours and asset availability
    - [x] Update `process_order_usecase.go` to fetch current market price during execution
    - [x] Create command objects for submit and cancel order operations
    - [x] Create `get_order_status_usecase.go` and `cancel_order_usecase.go`
  - [x] **Step 4.3**: Order Domain Service Enhancement
    - [x] Update `order_validation_service.go` to use market data client
    - [x] Add symbol existence validation
    - [x] Implement price range validation (market price ± tolerance)
    - [x] Add trading session validation (market open/closed)
  - [x] **Step 4.4**: Dependency Injection Integration
    - [x] Add market data gRPC client to dependency injection container
    - [x] Configure gRPC client connection in `NewContainer()` function
    - [x] Inject market data client into order use cases and domain services
    - [x] Add proper client lifecycle management (connection, reconnection, shutdown)
- [ ] **Step 5**: Application Use Cases
  - [ ] Create `application/usecase/submit_order_usecase.go`
    - [ ] Generate UUID for new orders
    - [ ] Validate order through domain services (including market data validation)
    - [ ] Save order with PENDING status to database
    - [ ] Publish order to RabbitMQ for async processing
    - [ ] Return order ID immediately (202 Accepted response)
  - [ ] Create `get_order_status_usecase.go` for status tracking
  - [ ] Create `cancel_order_usecase.go` for order cancellation
  - [ ] Create `process_order_usecase.go` for worker order processing
    - [ ] Fetch current market data via gRPC client
    - [ ] Execute order with real-time price information
    - [ ] Update order with execution price and timestamp
- [ ] **Step 6**: RabbitMQ Infrastructure Implementation
  - [x] **Step 6.1**: RabbitMQ Connection Management
    - [x] Add RabbitMQ dependency: `go get github.com/rabbitmq/amqp091-go` (updated to non-deprecated package)
    - [x] Create `shared/infra/messaging/` directory structure
    - [x] Implement messaging interface and RabbitMQ adapter with connection pooling and reconnection logic
    - [x] Add RabbitMQ configuration with environment-friendly defaults
    - [x] Create health check functionality for RabbitMQ connection
  - [x] **Step 6.2**: Queue Configuration and Setup
    - [x] Define queue structure:
      ```
      Primary Queues:
      - orders.submit (order submission)
      - orders.processing (order execution)
      - orders.settlement (order settlement)
      
      Management Queues:
      - orders.dlq (dead letter queue)
      - orders.retry (retry with TTL)
      - orders.status (status updates)
      ```
    - [x] Implement queue declaration and binding logic
    - [x] Configure message persistence and TTL settings
    - [x] Set up Dead Letter Queue (DLQ) with retry timing (5min, 15min, 1hr, 6hr)
  - [x] **Step 6.3**: Order Producer Implementation
    - [x] Create `infra/messaging/rabbitmq/order_producer.go`
    - [x] Implement `PublishOrderForProcessing(order)` method
    - [x] Add message serialization and routing logic
    - [x] Include error handling and fallback mechanisms
    - [x] Add message confirmation and delivery guarantees
  - [x] **Step 6.4**: Order Consumer Implementation
    - [x] Create `infra/messaging/rabbitmq/order_consumer.go`
    - [x] Implement message consumption with acknowledgment
    - [x] Add message deserialization and validation
    - [x] Include graceful shutdown and reconnection handling
- [x] **Step 7**: Order Worker for Asynchronous Processing
  - [x] Create `infra/worker/order_worker.go`
  - [x] Implement worker lifecycle management (Start, Stop, Health Check)
  - [x] Add order processing logic:
    - [x] Consume messages from RabbitMQ
    - [x] Execute order processing use case (with market data integration)
    - [x] Update order status in database
    - [x] Publish status updates
    - [x] Handle processing errors and retries
  - [x] Create `worker_manager.go` for worker scaling and monitoring
  - [x] Add worker metrics and performance monitoring
- [x] **Step 8**: Database Implementation
  - [x] Create `infra/persistence/order_repository.go`
  - [x] Implement database schema for orders table:
    ```sql
    orders (
      id UUID PRIMARY KEY,
      user_id UUID NOT NULL,
      symbol VARCHAR NOT NULL,
      order_type VARCHAR NOT NULL,
      quantity DECIMAL NOT NULL,
      price DECIMAL,
      status VARCHAR NOT NULL,
      created_at TIMESTAMP,
      updated_at TIMESTAMP,
      executed_at TIMESTAMP,
      execution_price DECIMAL,
      market_price_at_submission DECIMAL,
      market_data_timestamp TIMESTAMP
    )
    ```
  - [x] Add proper indexes for performance (user_id, status, created_at, symbol)
  - [x] Implement repository methods following existing patterns
- [x] **Step 9**: HTTP Presentation Layer
  - [x] Create `presentation/http/order_handler.go`
  - [x] Implement REST endpoints:
    - [x] `POST /orders` - Submit new order (returns 202 + order ID)
    - [x] `GET /orders/{id}` - Get order details (with market data context)
    - [x] `GET /orders/{id}/status` - Get order status
    - [x] `PUT /orders/{id}/cancel` - Cancel pending order
    - [x] `GET /orders/history` - Get user order history
  - [x] Add proper authentication and authorization
  - [x] Implement request validation and error handling
  - [x] Add comprehensive unit tests for handlers
- [x] **Step 10**: Dependency Injection Integration
  - [x] Update `pck/container.go` with order management dependencies:
    - [x] Add `GetSubmitOrderUseCase()` method
    - [x] Add `GetOrderWorkerManager()` method
    - [x] Add `GetOrderProducer()` method
    - [x] Add `GetOrderStatusUseCase()` method
    - [x] Add `GetMarketDataClient()` method for gRPC client
  - [x] Update `NewContainer()` function to initialize:
    - [x] Market data gRPC client with proper configuration
    - [x] RabbitMQ connection and producer
    - [x] Order repository with database connection
    - [x] Order use cases with market data client dependency
    - [x] Worker manager for background processing
    - [x] Idempotency service with Redis repository
  - [ ] Update `TestContainer` for testing support with mock market data client
- [x] **Step 10.1**: Idempotency System Implementation
  - [x] Create `IdempotencyService` domain service for preventing duplicate order submissions
  - [x] Implement Redis-based `IdempotencyRepository` for storing idempotency keys
  - [x] Add idempotency key generation based on order parameters (user, symbol, type, side, quantity, price)
  - [x] Integrate idempotency service into `SubmitOrderUseCase`
  - [x] Add idempotency status tracking (PENDING, COMPLETED, FAILED, EXPIRED)
  - [x] Implement automatic cleanup of expired idempotency keys
  - [x] Add mock idempotency repository for testing
  - [x] Configure 24-hour TTL for idempotency keys
- [x] **Step 11**: Integration and Testing
  - [x] Create comprehensive unit tests:
    - [x] Domain model tests (order validation, status transitions)
    - [x] Use case tests with mocked dependencies (including market data client)
    - [x] Handler tests with authentication flows
    - [x] Worker tests with RabbitMQ message processing
    - [x] Market data integration tests with mock gRPC responses
  - [x] Add integration tests:
    - [x] End-to-end order submission flow with market data validation
    - [x] RabbitMQ producer-consumer integration
    - [x] Database transaction and rollback scenarios
    - [x] Worker error handling and DLQ functionality
    - [x] gRPC client-server communication for market data
  - [ ] Performance testing:
    - [ ] Concurrent order submission (100+ orders/second)
    - [ ] Worker throughput and queue processing
    - [ ] Database connection pooling under load
    - [ ] RabbitMQ message throughput and latency
    - [ ] Market data gRPC call performance and caching
    ### ✅ Phase 7: Real-time Data & WebSocket Infrastructure
- [x] Implement WebSocket infrastructure for real-time asset quotations
- [x] Design and implement market data streaming architecture
- [x] Create connection management and scaling for WebSocket
- [x] Implement error handling and reconnection logic
- [x] Reuse same auth system in realtime_qoutes_websocket_handler.go
- [ ] Support 10,000+ concurrent WebSocket connections
- [x] Implement json patch updates for quotes to avoid sending whole data objects
- [x] REquest only assets that are needed (use ws message?)
- [x] fix WS early disconnections (connection lasting less than one minute)
- [ ] test multiple connections
- **Priority**: Medium - Real-time features

### ⏳ Phase 8: Position Update System - Order Execution to Position Persistence

**Current State Analysis:**
- ✅ Orders are successfully marked as `EXECUTED` in `orders` table
- ✅ Domain events system exists with `OrderExecutedEvent` 
- ✅ RabbitMQ messaging infrastructure is in place
- ❌ Position updates are missing when orders execute
- ❌ No position creation for new instruments
- ❌ No average price calculation for multiple purchases

**Business Requirements:**
1. **BUY Orders**: Create new position or update existing position (quantity + average price recalculation)
2. **SELL Orders**: Reduce position quantity or close position entirely  
3. **Data Consistency**: Ensure atomic operations between order execution and position updates
4. **Error Handling**: Handle partial updates, rollbacks, and retry scenarios
5. **Performance**: Support high-frequency trading without blocking order execution

---

## 🏗️ **APPROACH 1: Event-Driven Architecture with RabbitMQ** (RECOMMENDED)

**Architecture Diagram**: 📊 [Position Update Flow Diagram](docs/position_update_flow.png)

**Architecture Overview:**
```
Order Execution (Worker)
         ↓ [MarkAsExecuted()]
Domain Event: OrderExecutedEvent
         ↓ [Event Publisher]
RabbitMQ: positions.updates queue
         ↓ [Position Update Worker]
Position Service: Update/Create Position
         ↓ [Position Repository]
PostgreSQL: positions table
```

**Advantages:**
- ✅ Decoupled architecture - order execution doesn't block on position updates
- ✅ Reliable message delivery with RabbitMQ persistence and DLQ
- ✅ Horizontal scaling - multiple position workers can process updates
- ✅ Event sourcing capabilities for audit trails
- ✅ Easy to extend with other event consumers (notifications, analytics)

**Disadvantages:**
- ❌ Eventual consistency - small delay between order execution and position update
- ❌ More complex error handling across distributed components
- ❌ Requires message queue monitoring and management

### **Step 1**: Position Domain Model Enhancement
- [x] **Step 1.1**: Enhance Position Domain Model (COMPLETED)
  - [x] Create `position/domain/model/position.go` with complete Position entity
  - [x] Add methods: `UpdateQuantity()`, `CalculateNewAveragePrice()`, `CanSell(quantity)`
  - [x] Implement position validation rules and business logic
  - [x] Add position value objects: `PositionType`, `PositionStatus`
  - [x] Create comprehensive unit tests with 100% test coverage
  - [x] Implement business logic for BUY/SELL order position updates
  - [x] Add position status management (ACTIVE, PARTIAL, CLOSED)
- [x] **Step 1.2**: Position Domain Events (COMPLETED)
  - [x] Create `position/domain/model/position_events.go` 
  - [x] Implement events: `PositionCreatedEvent`, `PositionUpdatedEvent`, `PositionClosedEvent`
  - [x] Add `PositionPriceUpdatedEvent` for market price changes
  - [x] Add `PositionValidationFailedEvent` for error tracking
  - [x] Add position change tracking and audit capabilities
  - [x] Integrate domain events into Position aggregate methods
  - [x] Create comprehensive unit and integration tests (100% coverage)
  - [x] Implement event management methods (GetEvents, ClearEvents, HasEvents)
  - [x] Follow existing domain event patterns from order management system

### **Step 2**: Position Use Cases Implementation
- [x] **Step 2.1**: Position Management Use Cases (COMPLETED)
  - [x] Create `position/application/usecase/update_position_usecase.go`
  - [x] Implement `create_position_usecase.go` for new instruments
  - [x] Add `close_position_usecase.go` for complete SELL orders
  - [x] Create command objects: `UpdatePositionCommand`, `CreatePositionCommand`, `ClosePositionCommand`
  - [x] Implement comprehensive input validation and error handling
  - [x] Add business logic for position lifecycle management
  - [x] Create comprehensive unit tests with mock repository
  - [x] Update position repository interface for new domain model
  - [x] Add support for source order ID tracking and audit trails
  - [x] Implement realized P&L calculations for sell transactions
  - [x] Add position closure metrics and reporting
- [x] **Step 2.2**: Business Logic Implementation (COMPLETED)
  - [x] Average price calculation for BUY orders:
    ```
    NewAvgPrice = (ExistingQty * ExistingAvgPrice + NewQty * ExecutionPrice) / (ExistingQty + NewQty)
    ```
  - [x] Position quantity validation for SELL orders
  - [x] Handle fractional shares and position splitting
  - [x] Advanced position splitting utilities (by quantity and percentage)
  - [x] Financial precision handling with rounding and validation
  - [x] Enhanced business validation for trade operations
  - [x] Position merging capabilities for consolidation
  - [x] Comprehensive unit tests with 100% coverage
- [ ] **Step 2.3**: Position Database Schema and Persistence Implementation using postgres yanrodrigues schema
  - [ ] Create database migration for `positions_v2` table with Position domain model schema
  - [ ] Design table structure:
    ```sql
    positions_v2 (
      id UUID PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id),
      symbol VARCHAR NOT NULL,
      quantity DECIMAL(20,8) NOT NULL,
      average_price DECIMAL(20,8) NOT NULL,
      total_investment DECIMAL(20,8) NOT NULL,
      current_price DECIMAL(20,8) DEFAULT 0,
      market_value DECIMAL(20,8) DEFAULT 0,
      unrealized_pnl DECIMAL(20,8) DEFAULT 0,
      unrealized_pnl_pct DECIMAL(10,4) DEFAULT 0,
      position_type VARCHAR(10) NOT NULL CHECK (position_type IN ('LONG', 'SHORT')),
      status VARCHAR(20) NOT NULL CHECK (status IN ('ACTIVE', 'PARTIAL', 'CLOSED')),
      created_at TIMESTAMP DEFAULT NOW(),
      updated_at TIMESTAMP DEFAULT NOW(),
      last_trade_at TIMESTAMP,
      UNIQUE(user_id, symbol) -- Prevent duplicate positions per user/symbol
    )
    ```
  - [ ] Add proper indexes for performance:
    ```sql
    CREATE INDEX idx_positions_v2_user_id ON positions_v2(user_id);
    CREATE INDEX idx_positions_v2_symbol ON positions_v2(symbol);
    CREATE INDEX idx_positions_v2_status ON positions_v2(status);
    CREATE INDEX idx_positions_v2_user_symbol ON positions_v2(user_id, symbol);
    ```
  - [ ] Implement database repository methods in `position_repository.go`:
    - [ ] `FindByID()`, `FindByUserID()`, `FindByUserIDAndSymbol()`
    - [ ] `FindActivePositions()`, `Save()`, `Update()`, `Delete()`
    - [ ] `ExistsForUser()`, `CountPositionsForUser()`, `GetTotalInvestmentForUser()`
  - [ ] Create DTOs and mappers for Position domain model:
    - [ ] `PositionDTO` struct for database mapping
    - [ ] `PositionMapper` for domain model ↔ DTO conversion
    - [ ] Handle value object serialization (PositionType, PositionStatus)
  - [ ] Add database integration tests:
    - [ ] Test all repository methods with real database transactions
    - [ ] Test constraint violations (duplicate positions)
    - [ ] Test concurrent updates and race conditions
    - [ ] Verify proper UUID generation and foreign key constraints
  - [ ] Migration strategy from legacy `positions` table:
    - [ ] Data migration script to populate `positions_v2` from existing data
    - [ ] Backward compatibility considerations for existing aggregation queries
    - [ ] Gradual migration plan to avoid downtime

### **Step 3**: Event Publishing Integration
- [ ] **Step 3.1**: Order Execution Event Enhancement
  - [ ] Update `ProcessOrderUseCase.executeOrder()` to publish `OrderExecutedEvent`
  - [ ] Include all position-relevant data: symbol, quantity, execution_price, order_side
  - [ ] Add instrument_id resolution from symbol lookup
  - [ ] Ensure event publication is part of database transaction

### **Step 4**: RabbitMQ Position Update Infrastructure
- [ ] **Step 4.1**: Position Queue Configuration
  - [ ] Extend `OrderQueueManager` or create `PositionQueueManager`
  - [ ] Define queues: `positions.updates`, `positions.updates.dlq`, `positions.retry`
  - [ ] Configure queue durability, TTL, and DLQ routing
- [ ] **Step 4.2**: Position Update Worker Implementation
  - [ ] Create `position/infra/worker/position_update_worker.go`
  - [ ] Implement RabbitMQ consumer for `positions.updates` queue
  - [ ] Add worker lifecycle management and error handling



### Make assets trading hour configurable in market_data_client.go(persist in DB, for now)

### ⏳ Phase 9: Authentication & Login Improvements
- [ ] **Step 1**: User Registration and Account Creation
  - [ ] Create user registration domain model and validation
  - [ ] Implement `create_user_usecase.go` with proper business logic
  - [ ] Add email validation and uniqueness checks
  - [ ] Implement secure password hashing (bcrypt/argon2)
  - [ ] Create user registration HTTP endpoint
  - [ ] Add comprehensive validation for user input
  - [ ] Implement email verification workflow (optional)
  - [ ] Add user creation audit logging
- [ ] **Step 2**: Login Module Improvements
- [ ] Apply DDD pattern to login module
- [ ] Refactor login methods into smaller, more maintainable functions
- [ ] Implement comprehensive unit tests for login functionality
- [ ] Add password complexity requirements validation
- [ ] Implement rate limiting for login attempts
- [ ] Add session management and token refresh mechanisms
- [ ] Implement secure password handling improvements
- **Priority**: High - Security and maintainability improvements

    ### ⏳ Phase 10: Orders last mile
- [ ] **Step 12**: Error Handling and Monitoring
  - [ ] Implement comprehensive error handling:
    - [ ] Order validation errors with user-friendly messages
    - [ ] Market data service unavailability fallback strategies
    - [ ] RabbitMQ connection failures with fallback
    - [ ] Database transaction errors with proper rollback
    - [ ] Worker processing errors with retry logic
    - [ ] gRPC client timeout and circuit breaker patterns
  - [ ] Add logging and monitoring:
    - [ ] Structured logging for order lifecycle events
    - [ ] Metrics for order processing times and success rates
    - [ ] RabbitMQ queue depth and processing lag monitoring
    - [ ] Market data gRPC call latency and error rate monitoring
    - [ ] Alert system for critical failures (DLQ accumulation, market data unavailable)
- [ ] **Step 13**: Production Readiness
  - [ ] Add order execution and settlement integration:
    - [ ] External broker API integration (mock for development)
    - [ ] Order execution confirmation handling with market data
    - [ ] Settlement and clearing logic
    - [ ] Transaction reconciliation with market data timestamps
  - [ ] Implement compliance and audit features:
    - [ ] Order audit trail with immutable logs and market data snapshots
    - [ ] Regulatory reporting capabilities
    - [ ] Risk management and position limits with real-time market data
    - [ ] Order book and trade history with market context
  - [ ] Add operational features:
    - [ ] Admin endpoints for queue management
    - [ ] Order reprocessing capabilities
    - [ ] Manual order intervention tools
    - [ ] System health and status dashboard (including market data service health)
- **Architecture Overview**:
  ```
  Client Request (POST /orders)
           ↓
  ┌─────────────────────────┐
  │   HTTP Handler Layer    │ ← Generate UUID, Return 202 + Order ID
  └─────────────────────────┘
           ↓
  ┌─────────────────────────┐
  │   Submit Order UseCase  │ ← Validate with Market Data, Save to DB
  └─────────────────────────┘
           ↓
  ┌─────────────────────────┐         ┌─────────────────────────┐
  │   Market Data Client    │ ←────── │   Market Data Service   │
  │     (gRPC Client)       │  gRPC   │      (gRPC Server)      │
  └─────────────────────────┘         └─────────────────────────┘
           ↓
  ┌─────────────────────────┐
  │   RabbitMQ Producer     │ ← orders.processing queue
  └─────────────────────────┘
           ↓
  ┌─────────────────────────┐
  │   Order Worker          │ ← Async processing with Market Data
  │   (RabbitMQ Consumer)   │
  └─────────────────────────┘
           ↓
  ┌─────────────────────────┐
  │ Process Order UseCase   │ ← Execute with Real-time Market Data
  └─────────────────────────┘
           ↓
  ┌─────────────────────────┐
  │ Database Repository     │ ← PostgreSQL persistence + Market Data
  └─────────────────────────┘
  ```
- **Priority**: High - Core trading functionality with async processing and market data integration
- **Dependencies**: RabbitMQ setup, Order database schema, Worker infrastructure, Market Data gRPC Service
- **Performance Targets**: 
  - < 50ms order submission response time (including market data validation)
  - < 2 seconds order execution processing
  - Support 1000+ orders/minute throughput
  - 99.9% order processing reliability
  - < 100ms market data gRPC call latency
  - DLQ retry intervals: 5min → 15min → 1hr → 6hr → manual intervention
- **Market Data Integration Points**:
  - Order submission: Symbol validation, price validation, trading hours check
  - Order processing: Real-time price fetching, execution price determination
  - Order monitoring: Market data context for order analysis and reporting

### ⏳ Phase 11: Database Infrastructure & DevOps Setup
- [ ] Create comprehensive database schema for all entities:
  - [ ] **Users table** with proper authentication fields:
    ```sql
    users (
      id UUID PRIMARY KEY,
      email VARCHAR UNIQUE NOT NULL,
      password_hash VARCHAR NOT NULL,
      first_name VARCHAR,
      last_name VARCHAR,
      is_active BOOLEAN DEFAULT true,
      email_verified BOOLEAN DEFAULT false,
      created_at TIMESTAMP DEFAULT NOW(),
      updated_at TIMESTAMP DEFAULT NOW(),
      last_login_at TIMESTAMP,
      failed_login_attempts INTEGER DEFAULT 0,
      locked_until TIMESTAMP
    )
    ```
  - [ ] Instruments table with asset details
  - [ ] Enhanced balances table structure
  - [ ] Watchlists and watchlist_items tables
- [ ] Implement Docker containerization for database services
- [ ] Create Makefile for database operations (drop, recreate, populate)
- [ ] Add database migration scripts and versioning
- [ ] Implement database seeding with realistic test data
- [ ] Set up Redis containerization for caching
- [ ] **Step 4.2**: Performance Optimization
    - [ ] Database connection pooling optimization for high throughput
    - [ ] Implement database query optimization (indexes, query patterns)
    - [ ] Add metrics and monitoring for all layers (HTTP, gRPC, Cache, DB)
    - [ ] Profile and optimize memory usage and garbage collection
- **Priority**: High - Foundation for all other features

### ⏳ Phase 12: Security & Production Readiness
- [ ] **Step 1**: Database Security and SQL Injection Prevention
  - [ ] Audit all database queries for SQL injection vulnerabilities
  - [ ] Implement parameterized queries/prepared statements across all repositories
  - [ ] Add input sanitization and validation for all database operations
  - [ ] Create security testing for SQL injection attempts
  - [ ] Implement database query logging and monitoring
  - [ ] Add database access control and least privilege principles
  - [ ] Create secure database connection configurations
- [ ] **Step 2**: Network and Application Security
- [ ] Implement SSL/TLS encryption for all communications
- [ ] Set up Nginx load balancer with caching and security features
- [ ] Add WAF (Web Application Firewall) protection
- [ ] Implement DDoS protection and advanced rate limiting
- [ ] Add comprehensive audit logging for all transactions
- [ ] Implement database encryption at rest
- [ ] Add PII data protection and compliance measures
- [ ] Create security headers and protection policies
- **Priority**: High - Production security requirements

### ⏳ Phase 13: API Documentation & Testing
- [ ] Implement Swagger/OpenAPI documentation
- [ ] Create interactive API explorer
- [ ] Add automated API documentation generation
- [ ] Implement comprehensive unit test suite
- [ ] Add integration tests for service interactions
- [ ] Create end-to-end tests for complete workflows
- [ ] Add performance and load testing
- [ ] Implement security and penetration testing
- [ ] Performance testing with concurrent requests using tools like Apache Bench, wrk, or Go's testing framework
- [ ] Test concurrent portfolio summary requests to simulate multiple users (10, 50, 100+ concurrent users)
- [ ] Monitor response times under load and verify sub-200ms target performance
- [ ] Validate database connection pooling and resource usage under sustained load
- [ ] Test for memory leaks and connection issues during concurrent access
- [ ] Mutation tests
- **Priority**: Medium - Quality assurance and developer experience

### ⏳ Phase 14: Advanced Architecture & Microservices
- [ ] Implement gRPC for inter-service communication
- [ ] Design microservices decomposition strategy
- [ ] Add service discovery and registration
- [ ] Implement circuit breaker patterns
- [ ] Add distributed tracing and monitoring
- [ ] Create independent service deployment capabilities
- [ ] Implement horizontal scaling considerations
- [ ] **Step 3.7**: Service Discovery and Load Balancing (PENDING)
    - [ ] Implement service registration for gRPC endpoints
    - [ ] Add health checks for gRPC service
    - [ ] Configure load balancing for multiple gRPC instances
    - [ ] Add monitoring and metrics for gRPC performance
- **Priority**: Low - Advanced architecture (optional but recommended)

### ⏳ Phase 15: Performance & Monitoring
- [ ] Implement application and infrastructure monitoring (Prometheus para coletar metricas , grafana para exibir dash, jaeger tracing distribuido, openTelemetry coleta unificada de dados)
    https://www.youtube.com/watch?v=Wu0Ajkxh69Y
    https://github.com/ErickWendel/rinha-de-backend-2024-q1-nodejs
- [ ] Add performance metrics and alerting
- [ ] Create database performance optimization
- [ ] Implement caching strategies and optimization
- [ ] Add API response time monitoring (target < 200ms)
- [ ] Support 1000+ concurrent users
- [ ] Achieve 99.9% uptime target
- [ ] Implement real-time data within 100ms latency
- [ ] **Step 4.3**: Production Readiness
    - [ ] Add comprehensive logging with structured logs (JSON format)
    - [ ] Implement distributed tracing across all layers
    - [ ] Add Prometheus metrics for monitoring and alerting
    - [ ] Create health check endpoints for both HTTP and gRPC
    - [ ] Add graceful shutdown handling for all components
- **Priority**: Medium - Production performance requirements

### ⏳ Phase 16: CI/CD & DevOps Pipeline
- [ ] Set up automated CI/CD pipeline
- [ ] Implement automated testing in pipeline
- [ ] Add code quality checks and linting
- [ ] Create automated deployment processes
- [ ] Implement rollback capabilities
- [ ] Add environment management (dev, staging, prod)
- [ ] Create infrastructure as code (IaC)
- [ ] Add monitoring and alerting integration
- **Priority**: Medium - Development efficiency and reliability

### Additional Improvements to Consider:
- [ ] Add proper error handling with domain-specific errors
- [ ] Implement input validation in use cases
- [ ] Add logging and monitoring
- [ ] Consider adding domain events for complex workflows
- [ ] Add integration tests for the complete flow
- [ ] Mobile application development
- [ ] Advanced analytics and AI-powered insights
- [ ] Social trading features
- [ ] Cryptocurrency support
- [ ] International market expansion
- [ ] Advanced charting and technical analysis tools
- [ ] Send email? Or SMS? after sending and order

### Technical Debits
- [ ] **Token Verification Duplication**: Handlers are repeating token verification logic - need to segregate into middleware to avoid code duplication (MockAuth and VerifyToken)
- [ ] **SQL Injection Vulnerability Assessment**: Need to audit all existing database queries and repositories for potential SQL injection vulnerabilities
- [ ] **mTls pinning**
- [ ] **oAuth 2.0?**
- [ ] **User Management Gap**: Currently missing user registration/creation functionality - only login exists
- [ ] **Input Validation Inconsistency**: Need standardized input validation across all endpoints and use cases
- [ ] **Security Headers Missing**: HTTP responses lack security headers (CSRF, XSS protection, etc.)
- [ ] **Password Security**: Need to implement proper password complexity requirements and secure hashing