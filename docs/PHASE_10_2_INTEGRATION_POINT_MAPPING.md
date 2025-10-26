# PHASE 10.2: Market Data Service Migration - Step 1.5: Integration Point Mapping

## 1. Overview

This document provides a comprehensive mapping of all integration points where the Market Data Service is consumed by other services and frontend clients. Understanding these dependencies is critical for a successful microservice extraction, as each integration point will need to be updated to communicate with the new standalone `hub-market-data-service`.

### 1.1 Key Findings Summary

- **Total Integration Points**: 6 major consumers
- **Internal Services**: 4 (Order Management, Watchlist, Portfolio/Position, Realtime Quotes)
- **External Clients**: 2 (Frontend HTTP/REST, Frontend WebSocket)
- **Communication Protocols**: gRPC (internal), HTTP REST (external), WebSocket (external)
- **Migration Strategy**: Strangler Fig Pattern (gradual migration, no breaking changes)

---

## 2. Internal Service Integration Points

These are services within the monolith that currently call the Market Data Service via direct Go function calls or internal gRPC clients. After migration, they will need to call the external `hub-market-data-service` via gRPC.

### 2.1 Order Management Service

**Location**: `HubInvestmentsServer/internal/order_mngmt_system/`

**Purpose**: Validates symbols, checks trading hours, and fetches current prices for order execution.

#### 2.1.1 Integration Details

**Files Involved**:
- `domain/service/order_validation_service.go` - Core validation logic
- `infra/external/market_data_client.go` - gRPC client wrapper
- `application/usecase/submit_order_usecase.go` - Order submission flow
- `application/usecase/process_order_usecase.go` - Order execution flow

**Interface Used**: `IMarketDataClient` (defined in `order_validation_service.go`)

```go
type IMarketDataClient interface {
    ValidateSymbol(ctx context.Context, symbol string) (bool, error)
    GetCurrentPrice(ctx context.Context, symbol string) (float64, error)
    IsMarketOpen(ctx context.Context, symbol string) (bool, error)
    GetAssetDetails(ctx context.Context, symbol string) (*AssetDetails, error)
    GetTradingHours(ctx context.Context, symbol string) (*TradingHours, error)
}
```

**gRPC Methods Called**:
- `GetMarketData(symbols []string)` - Fetches market data for symbol validation and pricing

**Call Patterns**:
1. **Symbol Validation** (Submit Order):
   - Called by: `order_validation_service.ValidateSymbol()`
   - Frequency: Every order submission
   - Timeout: 5 seconds
   - Error Handling: Returns validation error, order rejected

2. **Price Fetching** (Order Execution):
   - Called by: `process_order_usecase.Execute()`
   - Frequency: Every order execution
   - Timeout: 5 seconds
   - Error Handling: Retries, falls back to order price

3. **Trading Hours Check** (Optional):
   - Called by: `order_validation_service.ValidateOrderWithContext()`
   - Frequency: Every order submission (if enabled)
   - Timeout: 5 seconds
   - Error Handling: Warning only, does not block order

**Current Implementation** (from `infra/external/market_data_client.go`):
```go
// MarketDataGRPCClientAdapter wraps the market data gRPC client
type MarketDataGRPCClientAdapter struct {
    grpcClient marketDataClient.IMarketDataGRPCClient
}

func (a *MarketDataGRPCClientAdapter) ValidateSymbol(ctx context.Context, symbol string) (bool, error) {
    marketData, err := a.grpcClient.GetMarketData(ctx, []string{symbol})
    if err != nil {
        return false, fmt.Errorf("failed to validate symbol via market data service: %w", err)
    }
    return len(marketData) > 0, nil
}

func (a *MarketDataGRPCClientAdapter) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
    marketData, err := a.grpcClient.GetMarketData(ctx, []string{symbol})
    if err != nil {
        return 0, fmt.Errorf("failed to get current price: %w", err)
    }
    if len(marketData) == 0 {
        return 0, fmt.Errorf("no market data found for symbol %s", symbol)
    }
    return float64(marketData[0].LastQuote), nil
}
```

**gRPC Client Configuration**:
- Server Address: `localhost:50060` (from `MARKET_DATA_GRPC_SERVER` env var)
- Timeout: 30 seconds (default)
- Connection: Insecure (no TLS)

#### 2.1.2 Migration Impact

**Changes Required**:
1. Update `MARKET_DATA_GRPC_SERVER` environment variable to point to new microservice (e.g., `hub-market-data-service:50051`)
2. No code changes required (already using gRPC client)
3. Update Dependency Injection container to use external gRPC client instead of internal one

**Testing Requirements**:
- Verify symbol validation works with new service
- Verify price fetching during order execution
- Load test: 1000 orders/minute (expected peak)
- Latency requirement: <100ms p95 for `GetMarketData` call

**Rollback Strategy**:
- Keep `MARKET_DATA_GRPC_SERVER` configurable
- Can switch back to monolith by updating environment variable

---

### 2.2 Watchlist Service

**Location**: `HubInvestmentsServer/internal/watchlist/`

**Purpose**: Enriches watchlist symbols with full market data (name, price, category).

#### 2.2.1 Integration Details

**Files Involved**:
- `application/usecase/get_watchlist_usecase.go` - Fetches watchlist and enriches with market data

**Interface Used**: `IGetMarketDataUsecase` (direct use case call, not gRPC)

```go
type GetWatchlistUsecase struct {
    repo           repository.IWatchlistRepository
    mktDataUsecase usecase.IGetMarketDataUsecase // Direct use case dependency
}

func (w *GetWatchlistUsecase) Execute(userId string) ([]model.MarketDataModel, error) {
    watchlistSymbols, err := w.repo.GetWatchlist(userId)
    if err != nil {
        return nil, err
    }

    // Calls Market Data Use Case directly
    mtkDataUsecase, err := w.mktDataUsecase.Execute(watchlistSymbols)
    if err != nil {
        return nil, err
    }

    return mtkDataUsecase, nil
}
```

**Call Patterns**:
1. **Watchlist Enrichment**:
   - Called by: `GetWatchlistUsecase.Execute()`
   - Frequency: Every watchlist fetch (user-initiated)
   - Timeout: N/A (direct function call)
   - Error Handling: Returns error to user

**Current Implementation**:
- **Type**: Direct Go function call (in-process)
- **No gRPC**: Uses `IGetMarketDataUsecase` interface directly

#### 2.2.2 Migration Impact

**Changes Required**:
1. **Replace direct use case call with gRPC client**:
   - Change `mktDataUsecase usecase.IGetMarketDataUsecase` to `marketDataClient marketDataClient.IMarketDataGRPCClient`
   - Update `Execute()` method to call `marketDataClient.GetMarketData(ctx, watchlistSymbols)`
2. **Update Dependency Injection**:
   - Inject gRPC client instead of use case in `pck/container.go`
3. **Add context propagation**:
   - Update `Execute(userId string)` to `Execute(ctx context.Context, userId string)`

**Code Changes** (estimated):
```go
// BEFORE (current)
type GetWatchlistUsecase struct {
    repo           repository.IWatchlistRepository
    mktDataUsecase usecase.IGetMarketDataUsecase
}

func (w *GetWatchlistUsecase) Execute(userId string) ([]model.MarketDataModel, error) {
    watchlistSymbols, err := w.repo.GetWatchlist(userId)
    if err != nil {
        return nil, err
    }
    return w.mktDataUsecase.Execute(watchlistSymbols)
}

// AFTER (migrated)
type GetWatchlistUsecase struct {
    repo             repository.IWatchlistRepository
    marketDataClient marketDataClient.IMarketDataGRPCClient
}

func (w *GetWatchlistUsecase) Execute(ctx context.Context, userId string) ([]model.MarketDataModel, error) {
    watchlistSymbols, err := w.repo.GetWatchlist(userId)
    if err != nil {
        return nil, err
    }
    
    // Call external Market Data Service via gRPC
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    return w.marketDataClient.GetMarketData(ctx, watchlistSymbols)
}
```

**Testing Requirements**:
- Verify watchlist enrichment works with new service
- Test with empty watchlist
- Test with 50+ symbols (max expected)
- Latency requirement: <200ms p95 for full watchlist fetch

**Rollback Strategy**:
- Temporarily keep both implementations (use case + gRPC client)
- Feature flag to switch between them
- Remove use case dependency after validation

---

### 2.3 Portfolio/Position Service

**Location**: `HubInvestmentsServer/internal/position/`

**Purpose**: Fetches current market prices to update position valuations and calculate unrealized P&L.

#### 2.3.1 Integration Details

**Files Involved**:
- `application/usecase/get_position_aggregation_usecase.go` - Fetches positions and enriches with current prices
- `infra/worker/position_update_worker.go` - Background worker for position price updates (not currently using market data, but should)

**Interface Used**: `IMarketDataGRPCClient` (already using gRPC client!)

```go
type GetPositionAggregationUseCase struct {
    repo               repository.PositionRepository
    aggregationService service.PositionAggregationService
    marketDataClient   marketDataClient.IMarketDataGRPCClient // Already using gRPC!
}

func NewGetPositionAggregationUseCase(repo repository.PositionRepository) *GetPositionAggregationUseCase {
    // Creates market data client for fetching current prices
    mdClient, err := marketDataClient.NewMarketDataGRPCClient(marketDataClient.MarketDataGRPCClientConfig{
        ServerAddress: "localhost:50060", // Hardcoded address
        Timeout:       0,                 // Use default
    })
    if err != nil {
        log.Printf("Warning: Failed to create market data client: %v. Positions will show 0 for current prices.", err)
        mdClient = nil
    }

    return &GetPositionAggregationUseCase{
        repo:               repo,
        aggregationService: service.NewPositionAggregationService(),
        marketDataClient:   mdClient,
    }
}
```

**gRPC Methods Called**:
- `GetMarketData(symbols []string)` - Fetches current prices for all position symbols

**Call Patterns**:
1. **Position Price Update**:
   - Called by: `GetPositionAggregationUseCase.Execute()`
   - Frequency: Every portfolio/position fetch (user-initiated)
   - Timeout: 5 seconds
   - Error Handling: Graceful degradation (falls back to stored `CurrentPrice` if market data unavailable)

**Current Implementation** (from `get_position_aggregation_usecase.go`):
```go
// fetchMarketPrices fetches current market prices for all position symbols
func (uc *GetPositionAggregationUseCase) fetchMarketPrices(positions []*domain.Position) map[string]float64 {
    if uc.marketDataClient == nil || len(positions) == 0 {
        return make(map[string]float64)
    }

    // Collect unique symbols
    symbolSet := make(map[string]bool)
    for _, pos := range positions {
        symbolSet[pos.Symbol] = true
    }

    symbols := make([]string, 0, len(symbolSet))
    for symbol := range symbolSet {
        symbols = append(symbols, symbol)
    }

    // Fetch market data for all symbols
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    marketDataList, err := uc.marketDataClient.GetMarketData(ctx, symbols)
    if err != nil {
        log.Printf("Warning: Failed to fetch market data for positions: %v", err)
        return make(map[string]float64)
    }

    // Build price map
    priceMap := make(map[string]float64)
    for _, md := range marketDataList {
        priceMap[md.Symbol] = float64(md.LastQuote)
    }

    return priceMap
}
```

#### 2.3.2 Migration Impact

**Changes Required**:
1. **Update hardcoded server address**:
   - Change `"localhost:50060"` to read from environment variable or config
   - Use `config.MarketDataGRPCServer` or similar
2. **Dependency Injection**:
   - Inject `marketDataClient` instead of creating it in the constructor
   - This allows for easier testing and configuration

**Code Changes** (estimated):
```go
// BEFORE (current)
func NewGetPositionAggregationUseCase(repo repository.PositionRepository) *GetPositionAggregationUseCase {
    mdClient, err := marketDataClient.NewMarketDataGRPCClient(marketDataClient.MarketDataGRPCClientConfig{
        ServerAddress: "localhost:50060", // Hardcoded
        Timeout:       0,
    })
    // ...
}

// AFTER (migrated)
func NewGetPositionAggregationUseCase(
    repo repository.PositionRepository,
    marketDataClient marketDataClient.IMarketDataGRPCClient, // Injected
) *GetPositionAggregationUseCase {
    return &GetPositionAggregationUseCase{
        repo:               repo,
        aggregationService: service.NewPositionAggregationService(),
        marketDataClient:   marketDataClient, // Use injected client
    }
}
```

**Testing Requirements**:
- Verify position valuations are correct with new service
- Test with 20+ positions (max expected per user)
- Test graceful degradation when market data service is unavailable
- Latency requirement: <150ms p95 for position aggregation

**Rollback Strategy**:
- Update `MARKET_DATA_GRPC_SERVER` environment variable to point back to monolith
- No code changes needed for rollback (already using gRPC)

---

### 2.4 Realtime Quotes Service (Internal WebSocket)

**Location**: `HubInvestmentsServer/internal/realtime_quotes/`

**Purpose**: Provides real-time price updates via WebSocket. Currently simulates price oscillations, but should fetch initial prices from Market Data Service.

#### 2.4.1 Integration Details

**Files Involved**:
- `application/service/price_oscillation_service.go` - Simulates price changes
- `application/service/asset_data_service.go` - Fetches initial asset data (likely uses Market Data)
- `infra/websocket/realtime_quotes_websocket_handler.go` - WebSocket handler

**Interface Used**: Likely `IGetMarketDataUsecase` (direct use case call)

**Call Patterns**:
1. **Initial Price Fetch** (on WebSocket connection):
   - Called by: `PriceOscillationService` or `AssetDataService`
   - Frequency: Once per WebSocket connection per symbol
   - Timeout: N/A (direct function call)
   - Error Handling: Returns error to client, connection fails

2. **Price Oscillation** (ongoing):
   - Currently: Simulated (no external calls)
   - Future: Should periodically fetch real prices from external provider

#### 2.4.2 Migration Impact

**Changes Required**:
1. **Update initial price fetching**:
   - Replace direct use case call with gRPC client
   - Similar to Watchlist Service migration
2. **Consider future real-time data integration**:
   - If Market Data Service will integrate with external real-time data providers (e.g., Alpha Vantage, IEX Cloud), the Realtime Quotes Service should consume that data via WebSocket or streaming gRPC

**Testing Requirements**:
- Verify initial quotes are correct on WebSocket connection
- Test with 10,000 concurrent WebSocket connections
- Latency requirement: <50ms for initial quote fetch

**Rollback Strategy**:
- Feature flag to switch between direct use case and gRPC client
- Keep both implementations temporarily

---

## 3. External Client Integration Points

These are frontend clients (web, mobile) that directly call the Market Data Service via HTTP REST or WebSocket.

### 3.1 Frontend HTTP REST API

**Location**: `HubInvestmentsServer/main.go` (HTTP routes)

**Purpose**: Provides market data to frontend for instrument search, quotes display, and watchlist enrichment.

#### 3.1.1 Integration Details

**HTTP Endpoints**:

1. **`GET /getMarketData?symbols=AAPL,MSFT`**
   - **Handler**: `marketDataHandler.GetMarketDataWithAuth`
   - **File**: `internal/market_data/presentation/http/market_data_handler.go`
   - **Authentication**: Required (JWT)
   - **Request**: Query parameter `symbols` (comma-separated)
   - **Response**:
     ```json
     [
       {
         "symbol": "AAPL",
         "name": "Apple Inc.",
         "last_quote": 150.25,
         "category": 1
       }
     ]
     ```
   - **Frequency**: User-initiated (search, watchlist, order form)
   - **Expected Load**: 100 requests/minute (peak)

2. **`POST /admin/market-data/cache/invalidate`** (Admin Only)
   - **Handler**: `adminHandler.AdminInvalidateCacheWithAuth`
   - **File**: `internal/market_data/presentation/http/market_data_handler.go`
   - **Authentication**: Required (JWT, admin role)
   - **Request**:
     ```json
     {
       "symbols": ["AAPL", "MSFT"]
     }
     ```
   - **Response**:
     ```json
     {
       "message": "Cache invalidated for 2 symbols",
       "symbols": ["AAPL", "MSFT"]
     }
     ```
   - **Frequency**: Rare (manual admin action)

3. **`POST /admin/market-data/cache/warm`** (Admin Only)
   - **Handler**: `adminHandler.AdminWarmCacheWithAuth`
   - **File**: `internal/market_data/presentation/http/market_data_handler.go`
   - **Authentication**: Required (JWT, admin role)
   - **Request**:
     ```json
     {
       "symbols": ["AAPL", "MSFT", "TSLA"]
     }
     ```
   - **Response**:
     ```json
     {
       "message": "Cache warmed for 3 symbols",
       "symbols": ["AAPL", "MSFT", "TSLA"]
     }
     ```
   - **Frequency**: Rare (manual admin action or startup)

#### 3.1.2 Migration Impact

**Changes Required**:
1. **API Gateway Routing**:
   - Add routes in `hub-api-gateway` to forward `/getMarketData` and `/admin/market-data/*` to `hub-market-data-service`
   - Example (in `hub-api-gateway/internal/proxy/proxy_handler.go`):
     ```go
     case strings.HasPrefix(path, "/getMarketData"):
         targetService = "hub-market-data-service"
         targetAddress = "hub-market-data-service:50051"
         grpcMethod = "/market_data.MarketDataService/GetMarketData"
     ```

2. **No Frontend Changes**:
   - Frontend continues to call `/getMarketData` via API Gateway
   - API Gateway transparently routes to new microservice

3. **Authentication**:
   - API Gateway validates JWT and forwards `x-user-id` to microservice
   - Microservice trusts API Gateway (no JWT validation)

**Testing Requirements**:
- Verify `/getMarketData` returns correct data via API Gateway
- Test with 1-50 symbols in a single request
- Verify admin endpoints work (cache invalidation, warming)
- Load test: 200 requests/minute via API Gateway
- Latency requirement: <150ms p95 end-to-end (including API Gateway overhead)

**Rollback Strategy**:
- Update API Gateway routing to point back to monolith
- No frontend changes needed

---

### 3.2 Frontend WebSocket (Real-time Quotes)

**Location**: `HubInvestmentsServer/main.go` (WebSocket route)

**Purpose**: Provides real-time streaming quotes to frontend via WebSocket.

#### 3.2.1 Integration Details

**WebSocket Endpoint**:

1. **`WS /ws/quotes?symbols=AAPL,MSFT&token=<JWT>`**
   - **Handler**: `container.GetRealtimeQuotesWebSocketHandler()`
   - **File**: `internal/realtime_quotes/infra/websocket/realtime_quotes_websocket_handler.go`
   - **Authentication**: Required (JWT in query parameter or header)
   - **Protocol**: WebSocket (RFC 6455)
   - **Message Format**: JSON Patch (RFC 6902)
   - **Initial Message** (on connection):
     ```json
     {
       "type": "quotes_patch",
       "operations": [
         {
           "op": "add",
           "path": "/quotes/AAPL",
           "value": {
             "symbol": "AAPL",
             "current_price": 150.25,
             "change": 1.50,
             "change_percent": 1.01,
             "last_updated": "2024-07-20T10:30:00Z"
           }
         }
       ]
     }
     ```
   - **Update Messages** (every 4 seconds):
     ```json
     {
       "type": "quotes_patch",
       "operations": [
         {
           "op": "replace",
           "path": "/quotes/AAPL/current_price",
           "value": 150.75
         },
         {
           "op": "replace",
           "path": "/quotes/AAPL/change",
           "value": 2.00
         }
       ]
     }
     ```
   - **Frequency**: Continuous (4-second updates)
   - **Expected Load**: 1,000 concurrent connections (peak)

#### 3.2.2 Migration Impact

**Changes Required**:
1. **Direct Connection to Microservice**:
   - **Decision**: WebSocket connections will **bypass the API Gateway** and connect directly to `hub-market-data-service`
   - **Rationale**: Reduces latency, simplifies architecture, allows dedicated WebSocket load balancer
   - **Frontend Update**: Change WebSocket URL from `ws://api-gateway:8081/ws/quotes` to `ws://market-data-service:8082/ws/quotes` (or via dedicated load balancer)

2. **Authentication**:
   - Microservice will validate JWT directly (using `hub-user-service` for token verification)
   - No reliance on API Gateway for WebSocket authentication

3. **Load Balancing**:
   - Deploy multiple instances of `hub-market-data-service`
   - Use Nginx or AWS ALB for WebSocket load balancing
   - Sticky sessions recommended (but not required due to Redis Pub/Sub)

**Testing Requirements**:
- Verify WebSocket connection works with new microservice
- Test with 10,000 concurrent connections
- Verify JSON Patch updates are correct
- Test reconnection logic
- Latency requirement: <25ms p95 for quote updates

**Rollback Strategy**:
- Update frontend WebSocket URL to point back to monolith
- Requires frontend deployment (not instant)

---

## 4. Integration Dependency Map

### 4.1 Visual Dependency Graph

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Market Data Service                             │
│                  (hub-market-data-service)                          │
│                                                                     │
│  - GetMarketData(symbols) -> MarketDataModel[]                     │
│  - StreamMarketData(symbols) -> stream MarketDataUpdate            │
│  - InvalidateCache(symbols)                                         │
│  - WarmCache(symbols)                                               │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ (consumed by)
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Order Mgmt    │   │  Watchlist    │   │ Portfolio/Pos │
│   Service     │   │   Service     │   │   Service     │
│               │   │               │   │               │
│ - Validate    │   │ - Enrich      │   │ - Update      │
│   Symbol      │   │   Watchlist   │   │   Prices      │
│ - Get Price   │   │   with Quotes │   │ - Calculate   │
│ - Check Hours │   │               │   │   P&L         │
└───────────────┘   └───────────────┘   └───────────────┘

        │                   │                   │
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Realtime      │   │  Frontend     │   │  Frontend     │
│ Quotes (WS)   │   │  (HTTP REST)  │   │  (WebSocket)  │
│               │   │               │   │               │
│ - Initial     │   │ - Search      │   │ - Real-time   │
│   Prices      │   │ - Watchlist   │   │   Quotes      │
│ - Oscillation │   │ - Order Form  │   │ - Live        │
│   Base        │   │               │   │   Updates     │
└───────────────┘   └───────────────┘   └───────────────┘
```

### 4.2 gRPC Method Usage Matrix

| Consumer Service         | gRPC Method       | Frequency        | Timeout | Error Handling       |
|--------------------------|-------------------|------------------|---------|----------------------|
| Order Management         | GetMarketData     | Per order        | 5s      | Reject order         |
| Watchlist                | GetMarketData     | Per fetch        | 5s      | Return error         |
| Portfolio/Position       | GetMarketData     | Per fetch        | 5s      | Graceful degradation |
| Realtime Quotes (init)   | GetMarketData     | Per WS connect   | 5s      | Fail connection      |
| Frontend (via Gateway)   | GetMarketData     | User-initiated   | 10s     | Return error         |
| Admin (cache invalidate) | InvalidateCache   | Rare (manual)    | 10s     | Return error         |
| Admin (cache warm)       | WarmCache         | Rare (manual)    | 30s     | Return error         |

### 4.3 HTTP REST Endpoint Usage Matrix

| HTTP Endpoint                         | Consumer      | Method | Auth Required | Frequency        |
|---------------------------------------|---------------|--------|---------------|------------------|
| `/getMarketData?symbols=...`          | Frontend      | GET    | Yes (JWT)     | User-initiated   |
| `/admin/market-data/cache/invalidate` | Admin UI      | POST   | Yes (Admin)   | Rare (manual)    |
| `/admin/market-data/cache/warm`       | Admin UI      | POST   | Yes (Admin)   | Rare (manual)    |

### 4.4 WebSocket Endpoint Usage Matrix

| WebSocket Endpoint                    | Consumer      | Protocol | Auth Required | Frequency        |
|---------------------------------------|---------------|----------|---------------|------------------|
| `/ws/quotes?symbols=...&token=...`    | Frontend      | WS       | Yes (JWT)     | Continuous       |

---

## 5. Migration Strategy and Sequencing

### 5.1 Strangler Fig Pattern

The migration will follow the **Strangler Fig Pattern**, where the new microservice is deployed alongside the monolith, and traffic is gradually shifted from the monolith to the microservice.

**Phases**:
1. **Phase 1: Deploy Microservice** (Week 10)
   - Deploy `hub-market-data-service` with same functionality as monolith
   - No traffic routed to it yet
   - Run in parallel for validation

2. **Phase 2: Migrate Internal Services** (Week 11-12)
   - Update Order Management, Watchlist, Portfolio/Position to use new microservice
   - Update environment variables to point to new service
   - Monitor for errors, rollback if needed

3. **Phase 3: Migrate Frontend HTTP** (Week 13)
   - Update API Gateway to route `/getMarketData` to new microservice
   - Monitor latency and error rates
   - Rollback API Gateway routing if issues

4. **Phase 4: Migrate Frontend WebSocket** (Week 14)
   - Update frontend to connect directly to new microservice
   - Requires frontend deployment
   - Monitor WebSocket connection stability

5. **Phase 5: Decommission Monolith Code** (Week 15-16)
   - After 2 weeks of stable operation, remove market data code from monolith
   - Archive for reference

### 5.2 Feature Flags

**Recommended Feature Flags**:
- `USE_MARKET_DATA_MICROSERVICE` - Global flag to enable/disable microservice usage
- `MARKET_DATA_SERVICE_ADDRESS` - Configurable service address (env var)
- `MARKET_DATA_FALLBACK_ENABLED` - Enable fallback to monolith if microservice fails

### 5.3 Monitoring and Observability

**Key Metrics to Track**:
1. **Latency**:
   - p50, p95, p99 for `GetMarketData` gRPC calls
   - End-to-end latency for HTTP REST endpoints
   - WebSocket message delivery latency

2. **Error Rates**:
   - gRPC error rate by status code
   - HTTP 5xx error rate
   - WebSocket connection failures

3. **Throughput**:
   - Requests per second (RPS) for gRPC and HTTP
   - Concurrent WebSocket connections

4. **Cache Performance**:
   - Cache hit rate (target: >95%)
   - Cache miss latency
   - Redis memory usage

5. **Service Health**:
   - Service uptime
   - Database connection pool usage
   - Redis connection pool usage

**Alerting Thresholds**:
- Latency p95 > 200ms
- Error rate > 1%
- Cache hit rate < 90%
- Service downtime > 30 seconds

---

## 6. Testing Strategy

### 6.1 Unit Tests

**Scope**: Test individual components in isolation.

**Test Cases**:
- gRPC client connection and retry logic
- Market data repository queries
- Cache-aside pattern logic
- WebSocket connection management

**Tools**: Go `testing` package, `testify/mock`

### 6.2 Integration Tests

**Scope**: Test interactions between services.

**Test Cases**:
1. **Order Management → Market Data**:
   - Submit order with valid symbol
   - Submit order with invalid symbol
   - Submit order when market data service is down

2. **Watchlist → Market Data**:
   - Fetch watchlist with 10 symbols
   - Fetch watchlist with 0 symbols
   - Fetch watchlist when market data service is down

3. **Portfolio → Market Data**:
   - Fetch portfolio with 20 positions
   - Fetch portfolio when market data service is down (graceful degradation)

4. **Frontend → Market Data (via Gateway)**:
   - Fetch market data via API Gateway
   - Verify JWT authentication
   - Verify error handling

**Tools**: Docker Compose for test environment, Go `testing` package

### 6.3 End-to-End Tests

**Scope**: Test complete user flows.

**Test Cases**:
1. **Order Submission Flow**:
   - User logs in → searches for symbol → submits order → order validated via market data → order executed

2. **Watchlist Flow**:
   - User logs in → fetches watchlist → watchlist enriched with market data

3. **Portfolio Flow**:
   - User logs in → fetches portfolio → positions updated with current prices from market data

4. **Real-time Quotes Flow**:
   - User logs in → connects to WebSocket → receives initial quotes → receives updates every 4 seconds

**Tools**: Postman/Newman, Selenium/Cypress for frontend, custom Go scripts

### 6.4 Load Tests

**Scope**: Verify performance under expected and peak load.

**Test Scenarios**:
1. **gRPC Load Test**:
   - 1000 `GetMarketData` requests/second
   - 10,000 concurrent WebSocket connections
   - Verify p95 latency < 200ms

2. **HTTP Load Test**:
   - 200 `/getMarketData` requests/minute via API Gateway
   - Verify p95 latency < 150ms

3. **Cache Performance Test**:
   - Measure cache hit rate with realistic traffic
   - Target: >95% hit rate

**Tools**: `k6`, `vegeta`, `wrk`, `ghz` (for gRPC)

### 6.5 Chaos Engineering

**Scope**: Test system resilience.

**Test Scenarios**:
1. **Market Data Service Failure**:
   - Kill market data service pod
   - Verify Order Management rejects orders gracefully
   - Verify Portfolio Service falls back to stored prices

2. **Database Failure**:
   - Kill PostgreSQL pod
   - Verify cache continues to serve requests
   - Verify graceful degradation

3. **Redis Failure**:
   - Kill Redis pod
   - Verify service falls back to database
   - Verify performance degradation is acceptable

**Tools**: Chaos Mesh, Gremlin, custom scripts

---

## 7. Rollback Plan

### 7.1 Rollback Triggers

**Automatic Rollback**:
- Error rate > 5% for 5 minutes
- Latency p95 > 500ms for 5 minutes
- Service downtime > 2 minutes

**Manual Rollback**:
- Critical bug discovered
- Data integrity issues
- Performance degradation affecting user experience

### 7.2 Rollback Procedure

**Step 1: Revert Environment Variables**
```bash
# For Order Management, Watchlist, Portfolio
export MARKET_DATA_GRPC_SERVER=hub-monolith:50060

# Restart affected services
kubectl rollout restart deployment/hub-order-service
kubectl rollout restart deployment/hub-watchlist-service
kubectl rollout restart deployment/hub-portfolio-service
```

**Step 2: Revert API Gateway Routing**
```bash
# Update API Gateway config to route to monolith
kubectl edit configmap hub-api-gateway-config

# Change:
# market_data_service: hub-market-data-service:50051
# To:
# market_data_service: hub-monolith:50060

# Restart API Gateway
kubectl rollout restart deployment/hub-api-gateway
```

**Step 3: Revert Frontend WebSocket URL** (if needed)
```bash
# Deploy frontend with old WebSocket URL
# This requires a frontend deployment, so it's not instant
```

**Step 4: Monitor**
- Verify error rates return to normal
- Verify latency returns to baseline
- Verify all services are healthy

**Step 5: Post-Mortem**
- Analyze root cause
- Document lessons learned
- Plan fixes before re-attempting migration

---

## 8. Success Criteria

### 8.1 Functional Success Criteria

- ✅ All integration points successfully migrated
- ✅ Order submission works with new service
- ✅ Watchlist enrichment works with new service
- ✅ Portfolio price updates work with new service
- ✅ Frontend HTTP REST API works via API Gateway
- ✅ Frontend WebSocket works with direct connection
- ✅ Admin cache management endpoints work

### 8.2 Performance Success Criteria

- ✅ gRPC latency p95 < 200ms
- ✅ HTTP REST latency p95 < 150ms (via Gateway)
- ✅ WebSocket latency p95 < 25ms
- ✅ Cache hit rate > 95%
- ✅ Error rate < 0.1%
- ✅ Service uptime > 99.9%

### 8.3 Operational Success Criteria

- ✅ Monitoring dashboards created
- ✅ Alerts configured
- ✅ Runbooks documented
- ✅ Rollback procedure tested
- ✅ On-call team trained

---

## 9. Conclusion of Step 1.5

The integration point mapping is complete! We've identified **6 major consumers** of the Market Data Service:

1. **Order Management Service** (gRPC, already using client)
2. **Watchlist Service** (direct use case, needs migration)
3. **Portfolio/Position Service** (gRPC, already using client)
4. **Realtime Quotes Service** (direct use case, needs migration)
5. **Frontend HTTP REST** (via API Gateway)
6. **Frontend WebSocket** (direct connection)

**Key Insights**:
- 2 services already use gRPC clients (Order Management, Portfolio) → minimal changes
- 2 services use direct use case calls (Watchlist, Realtime Quotes) → need refactoring
- Frontend integration is straightforward via API Gateway (HTTP) and direct connection (WebSocket)

**Migration Complexity**: 🟡 **MEDIUM**
- Most services already use gRPC (low effort)
- Watchlist and Realtime Quotes need refactoring (medium effort)
- Frontend changes are minimal (low effort)

**Estimated Effort**: **2-3 weeks** for full integration migration

---

## 10. Next Steps

With Step 1.5 complete, we've finished the **Pre-Migration Analysis** phase! 🎉

**Next Phase**: **Microservice Development (Week 10-12)**

**Step 2.1: Repository and Project Setup**
- Create `hub-market-data-service` repository
- Set up Go module structure
- Configure Docker and Docker Compose
- Set up CI/CD pipeline
- Initialize database and Redis

Let's proceed to Step 2.1! 🚀

