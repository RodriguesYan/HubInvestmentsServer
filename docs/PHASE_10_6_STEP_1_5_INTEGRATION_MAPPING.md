# Phase 10.6 - Step 1.5: Integration Point Mapping

**Service:** Order Management Service  
**Date:** November 4, 2025  
**Status:** ✅ COMPLETED

---

## Table of Contents
1. [External Service Dependencies](#external-service-dependencies)
2. [API Contracts](#api-contracts)
3. [RabbitMQ Integration](#rabbitmq-integration)
4. [Database Integration](#database-integration)
5. [Migration Strategy](#migration-strategy)

---

## 1. External Service Dependencies

### 1.1 Market Data Service (Critical Dependency)

**Purpose:** Real-time market data, symbol validation, price information

**Current Implementation:** `internal/order_mngmt_system/infra/external/market_data_client.go`

**Interface Contract:**
```go
type IMarketDataClient interface {
    // Symbol validation
    ValidateSymbol(ctx context.Context, symbol string) (bool, error)
    
    // Price information
    GetCurrentPrice(ctx context.Context, symbol string) (float64, error)
    GetAssetDetails(ctx context.Context, symbol string) (*AssetDetails, error)
    GetBatchMarketData(ctx context.Context, symbols []string) ([]MarketDataResponse, error)
    
    // Market status
    IsMarketOpen(ctx context.Context, symbol string) (bool, error)
    GetTradingHours(ctx context.Context, symbol string) (*TradingHours, error)
    
    // Lifecycle
    Close() error
}
```

**gRPC Methods Used:**
- `GetMarketData(symbol)` → Current price, company name, category
- `GetBatchMarketData(symbols[])` → Batch price retrieval

**Current Address:** `localhost:50054` (hub-market-data-service)

**Migration Impact:**
- ✅ Already using gRPC client
- ✅ Interface-based design (easy to test)
- ⚠️ Need to update address to use service discovery
- ⚠️ Need to implement circuit breaker pattern

---

### 1.2 User/Account Service (Critical Dependency)

**Purpose:** User authentication, balance verification, account validation

**Current Implementation:** Implicit via middleware and balance checks

**Required Operations:**
- **Balance Check:** Verify user has sufficient funds
- **Account Validation:** Verify user account is active and authorized for trading
- **User Context:** Extract user ID from JWT token

**Current Flow:**
```
HTTP Request → Auth Middleware → Extract UserID → Use Case
```

**Migration Impact:**
- ⚠️ Need explicit Account Service gRPC client
- ⚠️ Need to implement balance check before order submission
- ⚠️ Need to implement account status validation

**Proposed Interface:**
```go
type IAccountServiceClient interface {
    // Balance operations
    GetBalance(ctx context.Context, userID string) (*Balance, error)
    CheckSufficientFunds(ctx context.Context, userID string, requiredAmount float64) (bool, error)
    
    // Account validation
    ValidateAccount(ctx context.Context, userID string) (*AccountStatus, error)
    IsAccountActive(ctx context.Context, userID string) (bool, error)
    
    // Trading permissions
    CanTrade(ctx context.Context, userID string, symbol string) (bool, error)
}
```

---

### 1.3 Position Service (Event-Driven Dependency)

**Purpose:** Update user positions after order execution

**Current Implementation:** Event-driven via RabbitMQ

**Integration Type:** Asynchronous (Event Publishing)

**Event Published:**
- **Event:** `OrderExecutedEvent`
- **Queue:** `positions.updates`
- **Message Format:**
```json
{
  "order_id": "uuid",
  "user_id": "user_uuid",
  "symbol": "AAPL",
  "order_side": "BUY",
  "order_type": "MARKET",
  "quantity": 10.0,
  "execution_price": 150.25,
  "executed_at": "2025-11-04T10:30:00Z",
  "total_value": 1502.50,
  "market_price_at_exec": 150.25,
  "market_data_timestamp": "2025-11-04T10:30:00Z",
  "message_metadata": {
    "message_id": "position_update_order_uuid_timestamp",
    "correlation_id": "order_uuid",
    "timestamp": "2025-11-04T10:30:00Z",
    "retry_attempt": 0,
    "priority": 1,
    "source": "order_execution",
    "message_type": "position_update",
    "processing_stage": "initial"
  }
}
```

**Migration Impact:**
- ✅ Already decoupled via events
- ✅ No direct dependency
- ⚠️ Need to ensure RabbitMQ connection configuration is shared

---

### 1.4 Balance Service (Critical Dependency)

**Purpose:** Deduct/credit user balance on order execution

**Current Implementation:** Direct database access to `balances` table

**Required Operations:**
- **Deduct Balance:** On BUY order execution
- **Credit Balance:** On SELL order execution
- **Reserve Balance:** On LIMIT order submission (future)
- **Release Balance:** On order cancellation (future)

**Migration Impact:**
- ⚠️ **HIGH PRIORITY:** Need to create Balance Service gRPC client
- ⚠️ Need to implement transactional consistency (Saga pattern)
- ⚠️ Current direct DB access must be replaced

**Proposed Interface:**
```go
type IBalanceServiceClient interface {
    // Balance operations
    GetBalance(ctx context.Context, userID string) (*Balance, error)
    DeductBalance(ctx context.Context, userID string, amount float64, orderID string) error
    CreditBalance(ctx context.Context, userID string, amount float64, orderID string) error
    
    // Reservation (for LIMIT orders)
    ReserveBalance(ctx context.Context, userID string, amount float64, orderID string) error
    ReleaseBalance(ctx context.Context, userID string, orderID string) error
    
    // Validation
    CheckSufficientFunds(ctx context.Context, userID string, requiredAmount float64) (bool, error)
}
```

---

## 2. API Contracts

### 2.1 gRPC Service Definition

**Proto File:** `hub-proto-contracts/monolith/order_service.proto`

**Service Methods:**

#### 2.1.1 SubmitOrder
```protobuf
rpc SubmitOrder(SubmitOrderRequest) returns (SubmitOrderResponse);

message SubmitOrderRequest {
  string user_id = 1;
  string symbol = 2;
  string order_type = 3;  // MARKET, LIMIT, STOP_LOSS, STOP_LIMIT
  string order_side = 4;  // BUY, SELL
  double quantity = 5;
  optional double price = 6;
}

message SubmitOrderResponse {
  APIResponse api_response = 1;
  string order_id = 2;
  string status = 3;
  optional double estimated_price = 4;
  optional double estimated_value = 5;
  optional double market_price = 6;
  string submitted_at = 7;
}
```

**HTTP Equivalent:**
- **Endpoint:** `POST /api/v1/orders`
- **Auth:** Bearer token (JWT)
- **Status Code:** `202 Accepted`

---

#### 2.1.2 GetOrderDetails
```protobuf
rpc GetOrderDetails(GetOrderDetailsRequest) returns (GetOrderDetailsResponse);

message GetOrderDetailsRequest {
  string order_id = 1;
  string user_id = 2;
}

message GetOrderDetailsResponse {
  APIResponse api_response = 1;
  OrderDetails order = 2;
}

message OrderDetails {
  string order_id = 1;
  string user_id = 2;
  string symbol = 3;
  string order_type = 4;
  string order_side = 5;
  double quantity = 6;
  optional double price = 7;
  string status = 8;
  string created_at = 9;
  string updated_at = 10;
  optional string executed_at = 11;
  optional double execution_price = 12;
  optional double market_price_at_submission = 13;
  optional string market_data_timestamp = 14;
  double estimated_value = 15;
  optional double execution_value = 16;
}
```

**HTTP Equivalent:**
- **Endpoint:** `GET /api/v1/orders/{order_id}`
- **Auth:** Bearer token (JWT)
- **Status Code:** `200 OK`

---

#### 2.1.3 GetOrderStatus
```protobuf
rpc GetOrderStatus(GetOrderStatusRequest) returns (GetOrderStatusResponse);

message GetOrderStatusRequest {
  string order_id = 1;
  string user_id = 2;
}

message GetOrderStatusResponse {
  APIResponse api_response = 1;
  string order_id = 2;
  string status = 3;
  string status_message = 4;
  string updated_at = 5;
}
```

**HTTP Equivalent:**
- **Endpoint:** `GET /api/v1/orders/{order_id}/status`
- **Auth:** Bearer token (JWT)
- **Status Code:** `200 OK`

---

#### 2.1.4 CancelOrder
```protobuf
rpc CancelOrder(CancelOrderRequest) returns (CancelOrderResponse);

message CancelOrderRequest {
  string order_id = 1;
  string user_id = 2;
}

message CancelOrderResponse {
  APIResponse api_response = 1;
  string order_id = 2;
  string status = 3;
  string cancelled_at = 4;
}
```

**HTTP Equivalent:**
- **Endpoint:** `PUT /api/v1/orders/{order_id}/cancel`
- **Auth:** Bearer token (JWT)
- **Status Code:** `200 OK`

---

#### 2.1.5 GetOrderHistory
```protobuf
rpc GetOrderHistory(GetOrderHistoryRequest) returns (GetOrderHistoryResponse);

message GetOrderHistoryRequest {
  string user_id = 1;
  optional int32 limit = 2;
  optional int32 offset = 3;
  optional string status = 4;
}

message GetOrderHistoryResponse {
  APIResponse api_response = 1;
  repeated OrderDetails orders = 2;
  int32 total_count = 3;
  int32 limit = 4;
  int32 offset = 5;
}
```

**HTTP Equivalent:**
- **Endpoint:** `GET /api/v1/orders/history?page=1&limit=20`
- **Auth:** Bearer token (JWT)
- **Status Code:** `200 OK`

---

### 2.2 REST API Endpoints (via API Gateway)

All REST endpoints will be proxied through the API Gateway, which will:
1. Authenticate the user (JWT validation)
2. Extract user context (user_id, email)
3. Forward to Order Service via gRPC
4. Translate gRPC response to JSON

**API Gateway Routes:**
```yaml
- path: /api/v1/orders
  method: POST
  service: hub-order-service
  grpc_method: SubmitOrder
  auth_required: true

- path: /api/v1/orders/{order_id}
  method: GET
  service: hub-order-service
  grpc_method: GetOrderDetails
  auth_required: true

- path: /api/v1/orders/{order_id}/status
  method: GET
  service: hub-order-service
  grpc_method: GetOrderStatus
  auth_required: true

- path: /api/v1/orders/{order_id}/cancel
  method: PUT
  service: hub-order-service
  grpc_method: CancelOrder
  auth_required: true

- path: /api/v1/orders/history
  method: GET
  service: hub-order-service
  grpc_method: GetOrderHistory
  auth_required: true
```

---

## 3. RabbitMQ Integration

### 3.1 Queue Configuration

**Exchange:** `orders.exchange` (topic exchange)

**Queues:**

#### 3.1.1 Primary Processing Queues
| Queue Name | Purpose | Durability | DLQ | Priority |
|------------|---------|------------|-----|----------|
| `orders.submit` | Order submission | Durable | Yes | Normal (5) |
| `orders.processing` | Order execution | Durable | Yes | Normal (5) |
| `orders.settlement` | Order settlement | Durable | Yes | Normal (5) |

#### 3.1.2 Management Queues
| Queue Name | Purpose | Durability | DLQ | Priority |
|------------|---------|------------|-----|----------|
| `orders.status` | Status updates | Durable | No | High (8) |
| `orders.dlq` | Dead letter queue | Durable | No | N/A |
| `orders.retry` | Retry queue (TTL-based) | Durable | No | N/A |

---

### 3.2 Message Formats

#### 3.2.1 Order Submission Message
**Queue:** `orders.submit`

```json
{
  "order_id": "uuid",
  "user_id": "user_uuid",
  "symbol": "AAPL",
  "order_side": "BUY",
  "order_type": "MARKET",
  "quantity": 10.0,
  "price": null,
  "status": "PENDING",
  "created_at": "2025-11-04T10:00:00Z",
  "updated_at": "2025-11-04T10:00:00Z",
  "message_metadata": {
    "message_id": "msg_uuid",
    "correlation_id": "order_uuid",
    "timestamp": "2025-11-04T10:00:00Z",
    "retry_attempt": 0,
    "priority": 5,
    "source": "order_submission",
    "message_type": "order_submission",
    "processing_stage": "initial"
  }
}
```

---

#### 3.2.2 Order Processing Message
**Queue:** `orders.processing`

```json
{
  "order_id": "uuid",
  "user_id": "user_uuid",
  "symbol": "AAPL",
  "order_side": "BUY",
  "order_type": "MARKET",
  "quantity": 10.0,
  "price": null,
  "status": "PROCESSING",
  "created_at": "2025-11-04T10:00:00Z",
  "updated_at": "2025-11-04T10:00:05Z",
  "market_price_at_submission": 150.25,
  "market_data_timestamp": "2025-11-04T10:00:05Z",
  "message_metadata": {
    "message_id": "msg_uuid",
    "correlation_id": "order_uuid",
    "timestamp": "2025-11-04T10:00:05Z",
    "retry_attempt": 0,
    "priority": 5,
    "source": "order_processing",
    "message_type": "order_processing",
    "processing_stage": "execution"
  }
}
```

---

#### 3.2.3 Position Update Event (Published)
**Queue:** `positions.updates`

```json
{
  "order_id": "uuid",
  "user_id": "user_uuid",
  "symbol": "AAPL",
  "order_side": "BUY",
  "order_type": "MARKET",
  "quantity": 10.0,
  "execution_price": 150.25,
  "executed_at": "2025-11-04T10:00:10Z",
  "total_value": 1502.50,
  "market_price_at_exec": 150.25,
  "market_data_timestamp": "2025-11-04T10:00:10Z",
  "message_metadata": {
    "message_id": "position_update_order_uuid_timestamp",
    "correlation_id": "order_uuid",
    "timestamp": "2025-11-04T10:00:10Z",
    "retry_attempt": 0,
    "priority": 1,
    "source": "order_execution",
    "message_type": "position_update",
    "processing_stage": "initial"
  }
}
```

---

### 3.3 Event Publishing

**Events Published by Order Service:**

| Event Type | Queue | Consumer | Purpose |
|------------|-------|----------|---------|
| `OrderExecutedEvent` | `positions.updates` | Position Service | Update user positions |
| `OrderFailedEvent` | `orders.events` | Notification Service | Alert user of failure |
| `OrderCancelledEvent` | `orders.events` | Notification Service | Alert user of cancellation |

**Event Schema:**
```go
type EventMessage struct {
    EventID       string                 `json:"event_id"`
    EventType     string                 `json:"event_type"`
    AggregateID   string                 `json:"aggregate_id"`
    OccurredAt    time.Time              `json:"occurred_at"`
    EventData     map[string]interface{} `json:"event_data"`
    MessageID     string                 `json:"message_id"`
    CorrelationID string                 `json:"correlation_id"`
    Timestamp     time.Time              `json:"timestamp"`
    Source        string                 `json:"source"`
}
```

---

### 3.4 Retry Strategy

**Retry Intervals:**
1. **First Retry:** 5 minutes
2. **Second Retry:** 15 minutes
3. **Third Retry:** 1 hour
4. **Fourth Retry:** 6 hours
5. **Max Retries:** 4

**Retry Queue Configuration:**
- **Queue:** `orders.retry`
- **TTL:** Dynamic (based on retry attempt)
- **DLX:** `orders.processing` (dead letter exchange routes back to processing)
- **Max Retries:** 4 (after which message goes to DLQ)

---

## 4. Database Integration

### 4.1 Orders Table Schema

**Table:** `orders`

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    order_type VARCHAR(20) NOT NULL,
    order_side VARCHAR(10) NOT NULL,
    quantity DECIMAL(18, 8) NOT NULL,
    price DECIMAL(18, 8),
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMP,
    execution_price DECIMAL(18, 8),
    market_price_at_submission DECIMAL(18, 8),
    market_data_timestamp TIMESTAMP,
    failure_reason TEXT,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_symbol ON orders(symbol);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
```

---

### 4.2 Idempotency Table Schema

**Table:** `order_idempotency`

```sql
CREATE TABLE order_idempotency (
    idempotency_key VARCHAR(255) PRIMARY KEY,
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    response_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE INDEX idx_idempotency_user_id ON order_idempotency(user_id);
CREATE INDEX idx_idempotency_expires_at ON order_idempotency(expires_at);
```

---

### 4.3 Database Migration Strategy

**Approach:** Separate database for Order Service

**Steps:**
1. Create new database: `hub_order_service_db`
2. Run migrations to create `orders` and `order_idempotency` tables
3. Copy existing order data from monolith (if needed)
4. Update connection strings in Order Service

**Migration Script Location:** `hub-order-service/migrations/`

---

## 5. Migration Strategy

### 5.1 Strangler Fig Pattern

**Phase 1: Dual Write (Transition Period)**
- Monolith continues to handle orders
- New Order Service runs in parallel
- API Gateway routes to monolith
- Order Service consumes events for testing

**Phase 2: Gradual Cutover**
- API Gateway routes 10% of traffic to Order Service
- Monitor metrics and errors
- Gradually increase to 50%, 90%, 100%

**Phase 3: Full Migration**
- API Gateway routes 100% to Order Service
- Monolith order code is decommissioned
- Database migration complete

---

### 5.2 Integration Points Summary

| Integration | Type | Current | Target | Priority | Risk |
|-------------|------|---------|--------|----------|------|
| Market Data | gRPC | ✅ Implemented | Update address | HIGH | LOW |
| Account/Balance | Direct DB | ❌ Not implemented | gRPC client | HIGH | HIGH |
| Position Service | RabbitMQ | ✅ Implemented | No change | MEDIUM | LOW |
| API Gateway | HTTP → gRPC | ❌ Monolith | New routes | HIGH | MEDIUM |
| Database | Direct | ✅ Implemented | Separate DB | HIGH | MEDIUM |
| RabbitMQ | Events | ✅ Implemented | No change | MEDIUM | LOW |

---

### 5.3 Critical Dependencies for Migration

**Must Be Completed Before Order Service Migration:**
1. ✅ Market Data Service (DONE)
2. ❌ Account/Balance Service (REQUIRED)
3. ❌ User Service gRPC endpoints (REQUIRED)
4. ✅ API Gateway gRPC proxy (DONE)
5. ✅ RabbitMQ shared infrastructure (DONE)

---

## 6. Next Steps

### Step 2.1: Repository and Project Setup
- [ ] Create `hub-order-service` repository
- [ ] Initialize Go module
- [ ] Setup project structure (DDD layers)
- [ ] Create `Dockerfile` and `docker-compose.yml`
- [ ] Setup CI/CD pipeline

### Step 2.2: Copy Core Order Logic (AS-IS)
- [ ] Copy domain models
- [ ] Copy use cases
- [ ] Copy repositories
- [ ] Copy domain services
- [ ] Copy workers

### Step 2.3: Implement gRPC Service
- [ ] Create gRPC server
- [ ] Implement all 5 RPC methods
- [ ] Add authentication interceptor
- [ ] Add logging and metrics

---

## Conclusion

The Order Management Service has **5 critical external dependencies**:
1. **Market Data Service** (gRPC) - ✅ Ready
2. **Account/Balance Service** (gRPC) - ❌ Must be created
3. **Position Service** (RabbitMQ) - ✅ Ready
4. **User Service** (JWT validation) - ✅ Ready
5. **API Gateway** (HTTP → gRPC) - ✅ Ready

**Next Action:** Proceed to **Step 2.1: Repository and Project Setup** for the Order Service microservice.

---

**Document Version:** 1.0  
**Last Updated:** November 4, 2025  
**Author:** AI Assistant  
**Status:** ✅ COMPLETED

