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
- [x] **Step 2.3**: Position Database Schema and Persistence Implementation using postgres yanrodrigues schema (COMPLETED)
  - [x] **CLEANUP**: Removed legacy position table compatibility - using only new Position domain model
  - [x] Create database migration for `positions_v2` table with Position domain model schema
  - [x] Design table structure:
    ```sql
    yanrodrigues.positions_v2 (
      id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      user_id UUID NOT NULL,
      symbol VARCHAR(20) NOT NULL,
      quantity DECIMAL(20,8) NOT NULL,
      average_price DECIMAL(20,8) NOT NULL,
      total_investment DECIMAL(20,8) NOT NULL,
      current_price DECIMAL(20,8) DEFAULT 0,
      market_value DECIMAL(20,8) DEFAULT 0,
      unrealized_pnl DECIMAL(20,8) DEFAULT 0,
      unrealized_pnl_pct DECIMAL(10,4) DEFAULT 0,
      position_type VARCHAR(10) NOT NULL CHECK (position_type IN ('LONG', 'SHORT')),
      status VARCHAR(20) NOT NULL CHECK (status IN ('ACTIVE', 'PARTIAL', 'CLOSED')),
      created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
      last_trade_at TIMESTAMP WITH TIME ZONE,
      CONSTRAINT unique_user_symbol UNIQUE (user_id, symbol)
    )
    ```
  - [x] Add proper indexes for performance:
    ```sql
    CREATE INDEX idx_positions_v2_user_id ON yanrodrigues.positions_v2(user_id);
    CREATE INDEX idx_positions_v2_symbol ON yanrodrigues.positions_v2(symbol);
    CREATE INDEX idx_positions_v2_status ON yanrodrigues.positions_v2(status);
    CREATE INDEX idx_positions_v2_user_symbol ON yanrodrigues.positions_v2(user_id, symbol);
    CREATE INDEX idx_positions_v2_created_at ON yanrodrigues.positions_v2(created_at);
    CREATE INDEX idx_positions_v2_updated_at ON yanrodrigues.positions_v2(updated_at);
    CREATE INDEX idx_positions_v2_last_trade_at ON yanrodrigues.positions_v2(last_trade_at);
    ```
  - [x] Implement database repository methods in `position_repository.go`:
    - [x] `FindByID()`, `FindByUserID()`, `FindByUserIDAndSymbol()`
    - [x] `FindActivePositions()`, `Save()`, `Update()`, `Delete()`
    - [x] `ExistsForUser()`, `CountPositionsForUser()`, `GetTotalInvestmentForUser()`
  - [x] Create DTOs and mappers for Position domain model:
    - [x] `PositionDTO` struct for database mapping with validation
    - [x] `PositionMapper` for domain model ↔ DTO conversion with error handling
    - [x] Handle value object serialization (PositionType, PositionStatus)
    - [x] Query DTOs for flexible database filtering and pagination
  - [x] Add database integration tests:
    - [x] Test all repository methods with real database transactions
    - [x] Test constraint violations (duplicate positions)
    - [x] Test concurrent updates and race conditions
    - [x] Verify proper UUID generation and foreign key constraints
    - [x] Comprehensive test coverage with benchmarks
  - [x] Migration strategy from legacy `positions` table:
    - [x] Data migration script to populate `positions_v2` from existing data
    - [x] Migration comparison view for data integrity verification
    - [x] Rollback strategy with data preservation
    - [x] Database triggers for data consistency validation
  - [x] Advanced Database Features:
    - [x] PostgreSQL triggers for automatic timestamp updates
    - [x] Data consistency validation triggers
    - [x] Comprehensive constraint checking
    - [x] Schema-level yanrodrigues namespace isolation
    - [x] UUID extension integration
    - [x] Performance optimization with composite indexes

### **Step 3**: Event Publishing Integration
- [x] **Step 3.1**: Order Execution Event Enhancement (COMPLETED)
  - [x] Update `ProcessOrderUseCase.executeOrder()` to publish `OrderExecutedEvent`
  - [x] Include all position-relevant data: symbol, quantity, execution_price, order_side
  - [x] Add market data context: market_price_at_exec, market_data_timestamp  
  - [x] Ensure event publication is part of database transaction
  - [x] Remove legacy `NewOrderExecutedEvent` constructor to enforce position-relevant data
  - [x] Update all tests to use enhanced `NewOrderExecutedEventWithDetails` constructor

### **Step 4**: RabbitMQ Position Update Infrastructure
- [x] **Step 4.1**: Position Queue Configuration (COMPLETED)
  - [x] Created `PositionQueueManager` following `OrderQueueManager` pattern
  - [x] Defined queues: `positions.updates`, `positions.updates.dlq`, `positions.retry`
  - [x] Configured queue durability, TTL, and DLQ routing with position-specific settings
  - [x] Implemented queue management methods: PublishToPositionUpdatesQueue, PublishToRetryQueue, PublishToDLQ
  - [x] Added faster retry intervals (2min → 10min → 30min → 2hr) for position consistency
  - [x] Added comprehensive unit tests with MockMessageHandler
  - [x] Queue length limits (100k messages) to prevent memory issues during high volume
- [x] **Step 4.2**: Position Update Worker Implementation (COMPLETED)
  - [x] Created `position/infra/worker/position_update_worker.go` with comprehensive implementation
  - [x] Created `position/infra/worker/position_consumer.go` for RabbitMQ message consumption
  - [x] Implemented RabbitMQ consumer for `positions.updates` and `positions.retry` queues
  - [x] Added worker lifecycle management: Start/Stop, heartbeat monitoring, health checks
  - [x] Added comprehensive error handling and retry logic with exponential backoff
  - [x] Integrated with existing position use cases (Create, Update, Close) 
  - [x] Added position-specific business logic: handle buy/sell orders, position splitting/merging
  - [x] Added semaphore pattern for concurrency control (20 concurrent position updates)
  - [x] Added comprehensive metrics tracking and health monitoring
  - [x] Created extensive unit tests (6 test cases, all passing ✅)
  - [x] Create repository method to fetch position by symbol to avoid calling FindByUserID in handleSellOrder (position_update_worker.go) and make a loop in all position

### ⏳ Phase 9: Advanced Architecture & Microservices (SIMPLIFIED APPROACH)
- [x] **Implement gRPC for inter-service communication** ✅ **COMPLETED**
  
  **🎯 SIMPLE APPROACH - CREATE ONLY 4 FILES:**
  
- [x] **Step 1**: Create single consolidated proto file (`shared/grpc/services.proto`) (COMPLETED)
  - [x] Define all service interfaces in one file: AuthService, OrderService, PositionService  
  - [x] Include common message types (APIResponse, UserInfo, ErrorDetails)
  - [x] Keep it simple with essential methods only (Login, SubmitOrder, GetPositions)
  - [x] **Result**: One proto file instead of 4+ separate files
    
  - [x] **Step 2**: Create unified gRPC server (`shared/grpc/server.go`) (COMPLETED)
    - [x] Single server that hosts all gRPC services on one port (e.g. :50051)
    - [x] Implement basic versions of each service (Auth, Order, Position)
    - [x] Use existing use cases from container (no new business logic)
    - [x] Include simple JWT authentication middleware
    - [x] **Result**: One server file instead of multiple service files
    
  - [x] **Step 3**: Create gRPC client helper (`shared/grpc/client.go`) ✅ **COMPLETED**
    - [x] Simple client with methods like CallAuth(), CallOrder(), CallPosition()
    - [x] Handle connection management and authentication
    - [x] Provide easy-to-use interface for inter-service calls
    - [x] **Result**: One client file instead of complex client management system
    
  - [x] **Step 4**: Integration into main.go (minimal changes) ✅ **COMPLETED**
    - [x] Start gRPC server alongside HTTP server
    - [x] Add 5-10 lines of code maximum
    - [x] Graceful shutdown integration
    - [x] **Result**: Simple integration without complex service registry
    
  **🚀 EXAMPLE USE CASE:**
  ```go
  // Portfolio service calls Position service via gRPC
  client := grpc.NewClient()
  positions, err := client.CallPosition("GetPositions", userID)
  
  // Order service calls Auth service for validation  
  valid, err := client.CallAuth("ValidateToken", token)
  ```
  
  **📦 TOTAL FILES CREATED: 4 files only**
  - `shared/grpc/services.proto` (consolidated protobuf)
  - `shared/grpc/server.go` (unified gRPC server)  
  - `shared/grpc/client.go` (simple client helper)
  - Generated protobuf Go files (auto-generated)
  
  **✅ ACHIEVES SAME GOALS:**
  - Inter-service communication via gRPC
  - Authentication between services
  - Performance benefits of gRPC over HTTP
  - Foundation for microservices architecture
  - **But with 90% fewer files and complexity!**
  

- **Priority**: High - gRPC inter-service communication (simplified), other features optional

### Make assets trading hour configurable in market_data_client.go(persist in DB, for now)

### ⏳ Phase 10: Microservices Architecture Migration (MAJOR)

- [x] Design microservices decomposition strategy (COMPLETED)

**📋 Overview:** Transform the monolithic Hub Investments application into 6 microservices following the comprehensive decomposition strategy using the **Strangler Fig Pattern** for incremental, low-risk migration. This is a major architectural evolution requiring careful planning and execution.

**📖 Reference Documents:**
- [Microservices Decomposition Strategy](docs/microservices_decomposition_strategy.md)
- [Service Mapping Guide](docs/service_mapping_guide.md)
- [Architecture Diagrams](docs/microservices_architecture.png)
- [Event Flow Diagrams](docs/microservices_event_flow.png)

**🎯 Migration Strategy: Strangler Fig Pattern**
The Strangler Fig Pattern allows us to gradually replace monolithic functionality by:
1. Building new microservice alongside monolith (both run in parallel)
2. Routing specific requests to the new microservice via API Gateway
3. Maintaining monolith functionality until microservice is proven stable
4. Gradually increasing traffic to microservice (0% → 10% → 50% → 100%)
5. Decommissioning monolith module only after full validation

**Key Principles:**
- ✅ **Zero Downtime**: Monolith continues working throughout migration
- ✅ **Incremental Rollout**: Gradual traffic shifting with feature toggles
- ✅ **Easy Rollback**: Can revert to monolith instantly if issues arise
- ✅ **Shared Authentication**: JWT tokens work across monolith and microservices
- ✅ **Data Consistency**: Maintain single source of truth during transition

---

## **🚀 PHASE 10.1: User Management Service Migration (First Microservice)**

**Target Module:** `internal/auth/` + `internal/login/`  
**Service Name:** `hub-user-service`  
**Estimated Duration:** 6-8 weeks  
**Risk Level:** Low (minimal dependencies, clear boundaries)

### **Pre-Migration Analysis (Week 1)**

- [x] **Step 1.1: Deep Code Analysis** ✅ **COMPLETED**
  - [x] Audit `internal/auth/auth_service.go` - understand current JWT token creation and validation
  - [x] Audit `internal/auth/token/token_service.go` - analyze token signing, expiration, and secret management
  - [x] Audit `internal/login/application/usecase/do_login_usecase.go` - understand login flow and dependencies
  - [x] Audit `internal/login/domain/model/user_model.go` - document User domain model structure
  - [x] Audit `internal/login/domain/valueobject/` - document Email and Password value objects
  - [x] Audit `internal/login/infra/persistense/login_repository.go` - understand database queries and schema
  - [x] **Deliverable**: Complete code inventory document with dependency map
  - [x] **Document Created**: `docs/PHASE_10_1_CODE_INVENTORY.md` (115KB, comprehensive analysis)

- [x] **Step 1.2: Database Schema Analysis** ✅ **COMPLETED**
  - [x] Review existing migration: `shared/infra/migration/sql/000001_create_users_table.up.sql`
  - [x] Verify users table schema matches current implementation
  - [x] Analyzed actual database schema (13 columns vs 6 in migration)
  - [x] Identified schema discrepancies (email VARCHAR(50) in DB vs VARCHAR(255) in migration)
  - [x] **Decision**: ✅ Use migration file AS-IS (well-designed, properly constrained)
  - [x] Checked foreign key relationships: 5 tables reference users (orders, watchlist, balance, positions, aucAggregation)
  - [x] Analyzed indexes and constraints (migration has them, actual DB doesn't)
  - [x] Verified code compatibility: 100% compatible with both schemas
  - [x] **Deliverable**: Comprehensive database schema documentation
  - [x] **Document Created**: `docs/PHASE_10_1_DATABASE_SCHEMA_ANALYSIS.md` (comprehensive analysis with migration strategy)

- [x] **Step 1.3: Integration Point Mapping** ✅ **COMPLETED**
  - [x] Identified all `AuthService.VerifyToken()` calls: 3 direct + 1 WebSocket
  - [x] Identified all `AuthService.CreateToken()` calls: 2 direct calls
  - [x] Mapped all HTTP handlers using `TokenVerifier`: 12 protected endpoints
  - [x] Documented current authentication flow end-to-end (login + protected endpoints)
  - [x] Listed all services depending on authentication (balance, portfolio, orders, admin, etc.)
  - [x] Created dependency graph showing container → auth service → endpoints
  - [x] Analyzed migration impact: Only 3 files need changes (main.go, container.go, adapter)
  - [x] **Key Finding**: 12 protected endpoints require ZERO changes (interface unchanged)
  - [x] **Deliverable**: Comprehensive integration point documentation with flow diagrams
  - [x] **Document Created**: `docs/PHASE_10_1_INTEGRATION_POINTS.md` (detailed analysis)

- [ ] **Step 1.4: JWT Token Compatibility Analysis**
  - [ ] Document current JWT token structure (claims: username, userId, exp)
  - [ ] Document current JWT signing algorithm (HS256)
  - [ ] Document JWT secret management (from config.JWTSecret)
  - [ ] Identify token expiration policy (currently 10 minutes)
  - [ ] Plan for shared JWT secret between monolith and microservice
  - [ ] **Deliverable**: JWT token specification document ensuring backward compatibility

- [ ] **Step 1.5: Test Inventory**
  - [ ] Catalog existing tests to be reused:
    - `internal/auth/auth_service_test.go` (273 lines, comprehensive)
    - `internal/auth/token/token_service_test.go`
    - `internal/login/application/usecase/do_login_usecase_test.go`
    - `internal/login/domain/model/user_model_test.go`
    - `internal/login/domain/valueobject/email_test.go`
    - `internal/login/domain/valueobject/password_test.go`
    - `internal/login/infra/persistense/login_repository_test.go`
  - [ ] Plan to copy tests as-is to microservice (only update import paths)
  - [ ] **Deliverable**: Test migration plan

### **Microservice Development (Weeks 2-3)**

- [ ] **Step 2.1: Repository and Project Setup**
  - [ ] Create new Git repository: `hub-user-service`
  - [ ] Initialize Go module: `go mod init hub-user-service`
  - [ ] Set up project structure following clean architecture:
    ```
    hub-user-service/
    ├── cmd/
    │   └── server/
    │       └── main.go                    # Service entry point
    ├── internal/
    │   ├── core/
    │   │   ├── auth_service.go           # Copied from monolith (AS-IS)
    │   │   └── token_service.go          # Copied from monolith (AS-IS)
    │   ├── domain/
    │   │   ├── model/
    │   │   │   └── user.go               # Copied from monolith (AS-IS)
    │   │   ├── valueobject/
    │   │   │   ├── email.go              # Copied from monolith (AS-IS)
    │   │   │   └── password.go           # Copied from monolith (AS-IS)
    │   │   └── repository/
    │   │       └── user_repository.go    # Repository interface (AS-IS)
    │   ├── usecase/
    │   │   └── login_usecase.go          # Copied from monolith (AS-IS)
    │   ├── repository/
    │   │   └── postgres_user_repository.go # Copied from monolith (AS-IS)
    │   ├── grpc/
    │   │   ├── proto/
    │   │   │   └── auth_service.proto    # gRPC service definition
    │   │   └── auth_server.go            # gRPC server implementation
    │   └── http/
    │       └── auth_handler.go           # HTTP REST endpoints (optional)
    ├── config/
    │   ├── config.go                     # Configuration management
    │   └── config.yaml                   # Service configuration
    ├── migrations/
    │   └── 000001_create_users_table.up.sql   # Copied from monolith
    │   └── 000001_create_users_table.down.sql # Copied from monolith
    ├── Dockerfile                         # Container image definition
    ├── docker-compose.yml                # Local development setup
    ├── Makefile                          # Build and deployment commands
    └── README.md                         # Service documentation
    ```
  - [ ] Configure linting and code quality tools
  - [ ] **Deliverable**: Initialized repository with proper structure

- [ ] **Step 2.2: Copy Core Authentication Logic (AS-IS)**
  - [ ] Copy `internal/auth/auth_service.go` → `internal/core/auth_service.go`
  - [ ] Copy `internal/auth/token/token_service.go` → `internal/core/token_service.go`
  - [ ] **ONLY** update import paths (no business logic changes)
  - [ ] Update to use microservice config package
  - [ ] Ensure JWT secret is loaded from environment variable
  - [ ] **Deliverable**: Authentication core copied with minimal changes

- [ ] **Step 2.3: Copy Domain Layer (AS-IS)**
  - [ ] Copy `internal/login/domain/model/user_model.go` → `internal/domain/model/user.go`
  - [ ] Copy `internal/login/domain/valueobject/email.go` → `internal/domain/valueobject/email.go`
  - [ ] Copy `internal/login/domain/valueobject/password.go` → `internal/domain/valueobject/password.go`
  - [ ] Copy repository interface → `internal/domain/repository/user_repository.go`
  - [ ] **ONLY** update import paths (no business logic changes)
  - [ ] **Deliverable**: Domain layer copied as-is

- [ ] **Step 2.4: Copy Use Cases (AS-IS)**
  - [ ] Copy `internal/login/application/usecase/do_login_usecase.go` → `internal/usecase/login_usecase.go`
  - [ ] **ONLY** update import paths (no business logic changes)
  - [ ] **NO new use cases** (validate_token and refresh_token not needed for MVP)
  - [ ] **Deliverable**: Use case layer copied as-is

- [ ] **Step 2.5: Copy Repository Layer (AS-IS)**
  - [ ] Copy `internal/login/infra/persistense/login_repository.go` → `internal/repository/postgres_user_repository.go`
  - [ ] **ONLY** update import paths and database connection initialization
  - [ ] Update database connection to use microservice config
  - [ ] **NO changes to queries or business logic**
  - [ ] **Deliverable**: Repository implementation copied as-is

- [ ] **Step 2.6: Copy Database Migration Files**
  - [ ] Copy `shared/infra/migration/sql/000001_create_users_table.up.sql` → `migrations/000001_create_users_table.up.sql`
  - [ ] Copy `shared/infra/migration/sql/000001_create_users_table.down.sql` → `migrations/000001_create_users_table.down.sql`
  - [ ] **NO changes to migration files** (use as-is)
  - [ ] **Deliverable**: Migration files ready for microservice

- [ ] **Step 2.7: Implement gRPC Service Interface**
  - [ ] Copy existing proto from monolith: `shared/grpc/proto/auth_service.proto`
  - [ ] Use existing service definition (Login, ValidateToken)
  - [ ] Generate Go code from proto: `protoc --go_out=. --go-grpc_out=. auth_service.proto`
  - [ ] Implement `auth_server.go` with gRPC methods:
    - `Login()` - calls existing `login_usecase.Execute()`
    - `ValidateToken()` - calls existing `auth_service.VerifyToken()`
  - [ ] **NO new business logic** - just wire existing code to gRPC
  - [ ] **Deliverable**: gRPC server wrapping existing logic

- [ ] **Step 2.8: Configuration Management**
  - [ ] Create `config/config.go` for configuration loading
  - [ ] Support environment variables for all settings:
    - `JWT_SECRET` - JWT signing secret (MUST match monolith)
    - `DATABASE_URL` - PostgreSQL connection string
    - `GRPC_PORT` - gRPC server port (default: 50051)
    - `HTTP_PORT` - HTTP server port (default: 8081)
  - [ ] Create `config.yaml` for local development defaults
  - [ ] **Deliverable**: Configuration management

- [ ] **Step 2.9: Database Connection Strategy**
  - [ ] **Decision: Shared Database (Recommended for Phase 1)**
    - [ ] Microservice connects to **same PostgreSQL database** as monolith
    - [ ] Uses **same `users` table**
    - [ ] **NO data migration required**
    - [ ] Lower risk, faster deployment
    - [ ] Plan for eventual database separation in Phase 2
  - [ ] Configure database connection in microservice
  - [ ] Reuse existing connection pooling patterns from monolith
  - [ ] **Deliverable**: Database connection configuration

### **Testing and Validation (Week 4)**

- [ ] **Step 3.1: Copy Existing Unit Tests (AS-IS)**
  - [ ] Copy `internal/auth/auth_service_test.go` → `internal/core/auth_service_test.go`
  - [ ] Copy `internal/auth/token/token_service_test.go` → `internal/core/token_service_test.go`
  - [ ] Copy `internal/login/application/usecase/do_login_usecase_test.go` → `internal/usecase/login_usecase_test.go`
  - [ ] Copy `internal/login/domain/model/user_model_test.go` → `internal/domain/model/user_test.go`
  - [ ] Copy `internal/login/domain/valueobject/email_test.go` → `internal/domain/valueobject/email_test.go`
  - [ ] Copy `internal/login/domain/valueobject/password_test.go` → `internal/domain/valueobject/password_test.go`
  - [ ] Copy `internal/login/infra/persistense/login_repository_test.go` → `internal/repository/postgres_user_repository_test.go`
  - [ ] **ONLY** update import paths (no test logic changes)
  - [ ] Run all tests: `go test ./...`
  - [ ] Verify all tests pass
  - [ ] **Deliverable**: All existing tests passing in microservice

- [ ] **Step 3.2: gRPC Integration Testing**
  - [ ] Write integration tests for gRPC endpoints:
    - Test `Login()` RPC method
    - Test `ValidateToken()` RPC method
  - [ ] Test error scenarios (invalid credentials, expired tokens)
  - [ ] **Deliverable**: gRPC integration test suite

- [ ] **Step 3.3: JWT Token Compatibility Testing**
  - [ ] Generate JWT token in microservice
  - [ ] Validate token in monolith using existing `TokenService`
  - [ ] Generate JWT token in monolith
  - [ ] Validate token in microservice
  - [ ] Verify claims structure matches exactly (username, userId, exp)
  - [ ] Test token expiration behavior
  - [ ] **Deliverable**: Proof that tokens are 100% compatible

- [ ] **Step 3.4: Performance Testing (Optional)**
  - [ ] Load test gRPC endpoints (1000+ req/sec)
  - [ ] Measure login latency (target: <100ms p95)
  - [ ] Measure token validation latency (target: <10ms p95)
  - [ ] **Deliverable**: Performance benchmark report

### **Deployment Infrastructure (Week 5)**

- [ ] **Step 4.1: Containerization**
  - [ ] Create optimized `Dockerfile`:
    ```dockerfile
    # Multi-stage build for smaller image
    FROM golang:1.21-alpine AS builder
    WORKDIR /app
    COPY go.mod go.sum ./
    RUN go mod download
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux go build -o /hub-user-service cmd/server/main.go
    
    FROM alpine:latest
    RUN apk --no-cache add ca-certificates
    WORKDIR /root/
    COPY --from=builder /hub-user-service .
    EXPOSE 50051 8081
    CMD ["./hub-user-service"]
    ```
  - [ ] Create `docker-compose.yml` for local development
  - [ ] Test container build and startup
  - [ ] **Deliverable**: Docker image

- [ ] **Step 4.2: Local Development Environment**
  - [ ] Set up Docker Compose with:
    - User service container
    - PostgreSQL database (shared with monolith)
  - [ ] Create `Makefile` with common commands:
    - `make build` - Build service
    - `make test` - Run tests
    - `make run` - Run locally
    - `make docker-build` - Build Docker image
    - `make docker-run` - Run in Docker
  - [ ] Document setup instructions in README
  - [ ] **Deliverable**: Local development setup

- [ ] **Step 4.3: Basic Observability**
  - [ ] Add structured logging (simple text format is fine)
  - [ ] Add health check endpoint: `/health`
  - [ ] **Deliverable**: Basic observability

### **Integration with Monolith (Week 6)**

- [ ] **Step 5.1: Deploy Microservice Alongside Monolith**
  - [ ] Deploy `hub-user-service` to development environment
  - [ ] Configure to connect to same database as monolith
  - [ ] Ensure JWT secret matches monolith configuration
  - [ ] Verify service is healthy and responsive
  - [ ] Test gRPC connectivity from external client
  - [ ] **Deliverable**: Running microservice in dev environment

- [ ] **Step 5.2: Update Monolith to Use Microservice**
  - [ ] **Strategy**: Direct cutover (no feature toggle)
  - [ ] Update `main.go` to use gRPC client instead of local auth:
    ```go
    // OLD CODE (remove):
    // tokenService := token.NewTokenService()
    // aucService := auth.NewAuthService(tokenService)
    
    // NEW CODE (add):
    authClient, err := grpc.NewAuthClient("localhost:50051")
    if err != nil {
        log.Fatal("Failed to connect to auth service:", err)
    }
    aucService := auth.NewGRPCAuthServiceAdapter(authClient)
    ```
  - [ ] Create adapter that implements `auth.IAuthService` interface:
    ```go
    // internal/auth/grpc_auth_adapter.go
    type GRPCAuthServiceAdapter struct {
        grpcClient *grpc.AuthClient
    }
    
    func NewGRPCAuthServiceAdapter(client *grpc.AuthClient) auth.IAuthService {
        return &GRPCAuthServiceAdapter{grpcClient: client}
    }
    
    func (a *GRPCAuthServiceAdapter) VerifyToken(tokenString string, w http.ResponseWriter) (string, error) {
        resp, err := a.grpcClient.ValidateToken(context.Background(), tokenString)
        if err != nil {
            return "", err
        }
        return resp.UserInfo.UserId, nil
    }
    
    func (a *GRPCAuthServiceAdapter) CreateToken(userName string, userId string) (string, error) {
        resp, err := a.grpcClient.Login(context.Background(), &pb.LoginRequest{
            Email: userName,
            Password: "", // Token creation doesn't need password
        })
        if err != nil {
            return "", err
        }
        return resp.Token, nil
    }
    ```
  - [ ] Test monolith with microservice
  - [ ] **Deliverable**: Monolith using microservice for authentication

### **Validation and Monitoring (Weeks 7-8)**

- [ ] **Step 6.1: Functional Validation**
  - [ ] Test all authentication flows:
    - User login via `/login` endpoint
    - Token validation in all protected endpoints
    - Token creation for new sessions
  - [ ] Verify JWT tokens work correctly
  - [ ] Test error scenarios (invalid credentials, expired tokens)
  - [ ] **Deliverable**: All authentication working via microservice

- [ ] **Step 6.2: Performance Validation**
  - [ ] Monitor latency for 1 week:
    - Login requests
    - Token validation
  - [ ] Compare with baseline (monolith performance)
  - [ ] Ensure no degradation
  - [ ] **Deliverable**: Performance validation report

- [ ] **Step 6.3: Stability Validation**
  - [ ] Run microservice for 2 weeks in production
  - [ ] Monitor for errors and crashes
  - [ ] Verify database connection stability
  - [ ] Monitor memory and CPU usage
  - [ ] **Deliverable**: Stability validation report

### **Decommissioning Monolith Auth Module (Week 8+)**

- [ ] **Step 7.1: Validation Period**
  - [ ] Run microservice for 30 days minimum
  - [ ] Zero critical incidents
  - [ ] Performance meets or exceeds monolith
  - [ ] All stakeholders approve migration
  - [ ] **Deliverable**: Approval to decommission monolith module

- [ ] **Step 7.2: Remove Monolith Auth Code**
  - [ ] Remove `internal/auth/auth_service.go`
  - [ ] Remove `internal/auth/token/token_service.go`
  - [ ] Remove `internal/login/` module entirely
  - [ ] Update dependency injection container (remove auth/login dependencies)
  - [ ] Keep gRPC adapter in monolith
  - [ ] **Deliverable**: Cleaned up monolith codebase

- [ ] **Step 7.3: Documentation**
  - [ ] Document microservice architecture
  - [ ] Document gRPC API contracts
  - [ ] Document deployment procedures
  - [ ] Document how to run microservice locally
  - [ ] **Deliverable**: Complete documentation package

---

## **🎯 Success Criteria for Phase 10.1**

### **Technical Metrics**
- [ ] **Service Independence**: User service deploys independently of monolith
- [ ] **Zero Downtime**: No service interruptions during migration
- [ ] **Performance**: <100ms p95 latency for authentication
- [ ] **Reliability**: 99.9% uptime for user service
- [ ] **Compatibility**: 100% JWT token compatibility with monolith

### **Business Metrics**
- [ ] **No User Impact**: Users don't notice any change
- [ ] **No Functional Regression**: All auth features work exactly as before
- [ ] **Improved Observability**: Better metrics and logging than monolith
- [ ] **Team Confidence**: Team comfortable with microservices approach

### **Risk Mitigation**
- [ ] **Rollback Plan**: Can revert to monolith by reverting code changes
- [ ] **Monitoring**: Basic logging catches issues
- [ ] **Testing**: Reused tests from monolith prevent regressions

---

## **📋 FUTURE PHASES (After User Service Success)**

### **Phase 10.2: Market Data Service** (Weeks 9-16)
- Follow same strangler fig pattern
- Extract `internal/market_data/` and `internal/realtime_quotes/`
- More complex due to WebSocket connections and caching
- Estimated: 8 weeks

### **Phase 10.3: Watchlist Service** (Weeks 17-22)
- Simpler service with fewer dependencies
- Extract `internal/watchlist/`
- Estimated: 6 weeks

### **Phase 10.4: Account Management Service** (Weeks 23-30)
- Critical financial service requiring extra validation
- Extract `internal/balance/`
- Estimated: 8 weeks

### **Phase 10.5: Position & Portfolio Service** (Weeks 31-40)
- Complex aggregation logic
- Extract `internal/position/` and `internal/portfolio_summary/`
- Estimated: 10 weeks

### **Phase 10.6: Order Management Service** (Weeks 41-54)
- Most complex service with many dependencies
- Extract `internal/order_mngmt_system/`
- Requires saga pattern for distributed transactions
- Estimated: 14 weeks

---

## **🔧 Tools and Technologies**

### **Development**
- **Language**: Go 1.21+
- **gRPC Framework**: google.golang.org/grpc
- **Database**: PostgreSQL (shared initially, separate later)
- **Caching**: Redis
- **Messaging**: RabbitMQ

### **Deployment**
- **Containerization**: Docker
- **Orchestration**: Docker Compose (dev), Kubernetes (prod later)
- **CI/CD**: GitHub Actions
- **Service Mesh**: Istio (future phase)

### **Observability**
- **Metrics**: Prometheus + Grafana
- **Logging**: Structured JSON logs
- **Tracing**: OpenTelemetry + Jaeger
- **Alerting**: Prometheus AlertManager

### **Testing**
- **Unit Tests**: Go testing package
- **Integration Tests**: Testcontainers
- **Load Testing**: k6 or Apache Bench
- **Contract Testing**: Pact (future)

---

## **⚠️ Risk Management**

### **Technical Risks**
| Risk | Mitigation | Rollback Plan |
|------|-----------|---------------|
| JWT incompatibility | Extensive compatibility testing | Revert code changes |
| Performance degradation | Performance testing | Revert code changes |
| Database connection issues | Connection pooling + monitoring | Revert code changes |
| gRPC connectivity failures | Error handling + retries | Revert code changes |
| Service unavailability | Health checks + restart | Revert code changes |

### **Operational Risks**
| Risk | Mitigation | Rollback Plan |
|------|-----------|---------------|
| Team unfamiliarity | Documentation | Keep monolith code temporarily |
| Deployment complexity | Simple Docker deployment | Rollback deployment |
| Monitoring gaps | Basic logging | Increase logging |
| Incident response | Document procedures | Revert to monolith |

---

## **📊 Timeline Summary**

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| Pre-Migration Analysis | 1 week | Code audit + dependency map |
| Microservice Development | 2 weeks | Working user service |
| Testing & Validation | 1 week | Test suite + compatibility proof |
| Deployment Infrastructure | 1 week | Docker + basic observability |
| Integration with Monolith | 1 week | Direct cutover to microservice |
| Validation & Monitoring | 2 weeks | Stability validation |
| Decommissioning | Ongoing | Clean up monolith code |
| **TOTAL** | **8 weeks** | **Production microservice** |

---

**🎉 After Phase 10.1 completion, we will have:**
1. ✅ First microservice successfully extracted and running in production
2. ✅ Proven strangler fig pattern that works for our architecture
3. ✅ Team confidence and experience with microservices
4. ✅ Template and playbook for extracting remaining services
5. ✅ Foundation for scaling to 6 microservices architecture

### **📊 Microservices Success Metrics**
- **Technical KPIs:**
  - [ ] 100% service independence (deploy without affecting others)
  - [ ] <200ms p95 API response times maintained
  - [ ] 99.9% uptime across all services
  - [ ] <5s end-to-end event processing latency
  - [ ] 30% infrastructure cost reduction through optimization

- **Business KPIs:**
  - [ ] 50% faster feature delivery (independent service teams)
  - [ ] <30 minutes mean time to recovery for incidents
  - [ ] 10x scalability capacity for peak trading hours
  - [ ] 100% audit trail coverage for compliance
  - [ ] Zero performance degradation during migration

- **Operational KPIs:**
  - [ ] Daily deployments per service capability
  - [ ] <2 hours lead time from commit to production
  - [ ] <5% change failure rate across services
  - [ ] 100% test automation coverage
  - [ ] Full observability across all services

### **⚠️ Risk Mitigation Checklist**
- [ ] **Data Consistency:** Saga patterns tested with failure scenarios
- [ ] **Service Dependencies:** Circuit breakers implemented and tested
- [ ] **Security Vulnerabilities:** Regular security scans and penetration testing
- [ ] **Performance Degradation:** Comprehensive load testing before migration
- [ ] **Operational Complexity:** Team training and comprehensive documentation
- [ ] **Deployment Failures:** Automated rollback procedures tested

**📋 Total Estimated Timeline:** 12 months for complete microservices migration
**👥 Team Requirements:** 3-4 senior engineers + 1 DevOps engineer + 1 architect
**💰 Infrastructure Investment:** Kubernetes cluster + monitoring + CI/CD tooling

---

### ⏳ Phase 11: Authentication & Login Improvements
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

    ### ⏳ Phase 11: Orders last mile
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

### ⏳ Phase 12: Database Infrastructure & DevOps Setup
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

### ⏳ Phase 13: Security & Production Readiness
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

### ⏳ Phase 14: API Documentation & Testing
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
- [ ] Add service discovery and registration
- [ ] Implement circuit breaker patterns
- [ ] Add distributed tracing and monitoring
- [ ] Create independent service deployment capabilities
- [ ] Implement horizontal scaling considerations

### Technical Debits
- [ ] **Token Verification Duplication**: Handlers are repeating token verification logic - need to segregate into middleware to avoid code duplication (MockAuth and VerifyToken)
- [ ] **SQL Injection Vulnerability Assessment**: Need to audit all existing database queries and repositories for potential SQL injection vulnerabilities
- [ ] **mTls pinning**
- [ ] **oAuth 2.0?**
- [ ] **User Management Gap**: Currently missing user registration/creation functionality - only login exists
- [ ] **Input Validation Inconsistency**: Need standardized input validation across all endpoints and use cases
- [ ] **Security Headers Missing**: HTTP responses lack security headers (CSRF, XSS protection, etc.)
- [ ] **Password Security**: Need to implement proper password complexity requirements and secure hashing