# Order Management System - Idempotency Flow

## Overview

The Order Management System now includes a comprehensive idempotency mechanism to prevent duplicate order submissions. This ensures that even if a client accidentally submits the same order multiple times, only one order will be processed.

## Architecture Components

### 1. Idempotency Service
- **Location**: `internal/order_mngmt_system/domain/service/idempotency_service.go`
- **Purpose**: Manages idempotency keys and their lifecycle
- **Key Generation**: Uses SHA-256 hash of order parameters (user, symbol, type, side, quantity, price)

### 2. Redis Repository
- **Location**: `internal/order_mngmt_system/infra/idempotency/redis_idempotency_repository.go`
- **Purpose**: Stores idempotency keys in Redis with TTL (24 hours)
- **Storage Format**: JSON serialized `IdempotencyKey` objects

### 3. Integration Points
- **Submit Order Use Case**: Integrated idempotency checks before order processing
- **HTTP Handler**: Transparent to clients - returns consistent responses
- **Error Handling**: Proper cleanup on failures

## Flow Description

### 1. Client Request
```
POST /api/orders/sendOrder
{
  "user_id": "user123",
  "symbol": "AAPL",
  "order_type": "LIMIT",
  "side": "BUY",
  "quantity": 100,
  "price": 150.00
}
```

### 2. Idempotency Key Generation
The system generates a deterministic key based on:
- User ID
- Symbol
- Order Type
- Side (BUY/SELL)
- Quantity
- Price (for limit orders)

**Example Key**: `order_idempotency:sha256(user123|AAPL|LIMIT|BUY|100|150.00)`

### 3. Idempotency Check Flow

#### First Request (New)
1. **Check**: Key doesn't exist in Redis
2. **Store**: Create key with status `PENDING`
3. **Process**: Continue with order validation and processing
4. **Complete**: Update key status to `COMPLETED` on success
5. **Response**: Return order ID (200 OK)

#### Duplicate Request (Existing)
1. **Check**: Key exists in Redis
2. **Status Check**:
   - `PENDING`: Return "processing" response
   - `COMPLETED`: Return existing order ID
   - `FAILED`: Allow retry (key expired or cleaned up)
   - `EXPIRED`: Allow retry
3. **Response**: Consistent response without duplicate processing

### 4. Error Handling
- **Validation Failure**: Key marked as `FAILED`
- **Processing Error**: Key marked as `FAILED`
- **Timeout**: Key expires after 24 hours
- **Cleanup**: Automatic cleanup of expired keys

## Key Features

### 🔒 **Duplicate Prevention**
- Prevents duplicate orders even with network retries
- Deterministic key generation ensures consistency
- 24-hour protection window

### ⚡ **Performance Optimized**
- Redis-based storage for fast lookups
- Minimal impact on order processing flow
- Efficient key generation using SHA-256

### 🛡️ **Robust Error Handling**
- Graceful degradation if Redis is unavailable
- Proper cleanup on failures
- Automatic expiration of old keys

### 📊 **Monitoring Ready**
- Health checks for Redis connectivity
- Metrics for idempotency hit rates
- Structured logging for debugging

## Status Transitions

```
PENDING → COMPLETED (successful processing)
PENDING → FAILED (validation/processing error)
COMPLETED → EXPIRED (after TTL)
FAILED → EXPIRED (after TTL)
```

## Configuration

### Redis Configuration
```go
type IdempotencyConfig struct {
    TTL:           24 * time.Hour,
    CleanupInterval: 1 * time.Hour,
    KeyPrefix:     "order_idempotency:",
}
```

### Integration
The idempotency service is automatically integrated into:
- `SubmitOrderUseCase`
- Dependency injection container
- HTTP handlers (transparent)

## Benefits

1. **Client Safety**: Prevents accidental duplicate orders
2. **Network Resilience**: Handles network retries gracefully  
3. **User Experience**: Consistent responses for duplicate requests
4. **System Integrity**: Maintains data consistency
5. **Operational Safety**: Reduces support overhead from duplicate orders

## Monitoring

### Key Metrics
- Idempotency hit rate (duplicate requests detected)
- Redis health and connectivity
- Key expiration and cleanup rates
- Processing time impact

### Alerts
- Redis connectivity issues
- High duplicate request rates (potential client issues)
- Idempotency service failures

## Testing

The idempotency system includes comprehensive tests:
- Unit tests for service logic
- Integration tests with Redis
- End-to-end tests for duplicate detection
- Error scenario testing

See: `internal/order_mngmt_system/domain/service/idempotency_service_test.go`
