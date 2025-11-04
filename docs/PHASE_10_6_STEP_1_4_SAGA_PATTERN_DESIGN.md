# Phase 10.6 - Step 1.4: Saga Pattern Design for Order Management

**Service:** Order Management Service  
**Date:** November 4, 2025  
**Status:** ✅ COMPLETED

---

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Current Order Processing Flow](#current-order-processing-flow)
3. [Saga Pattern Overview](#saga-pattern-overview)
4. [Order Processing Saga Design](#order-processing-saga-design)
5. [Compensating Transactions](#compensating-transactions)
6. [Implementation Strategy](#implementation-strategy)
7. [Error Handling and Recovery](#error-handling-and-recovery)
8. [Monitoring and Observability](#monitoring-and-observability)

---

## 1. Executive Summary

### Problem Statement

The Order Management System involves **distributed transactions** across multiple microservices:
1. **Market Data Service** - Price validation and symbol verification
2. **Account/Balance Service** - Balance checks and deductions
3. **Order Service** - Order state management
4. **Position Service** - Position updates after execution

**Current Risk:** If any service fails mid-transaction, the system can end up in an **inconsistent state**:
- Order marked as EXECUTED, but balance not deducted
- Balance deducted, but position not updated
- Position updated, but order still shows PENDING

### Solution: Saga Pattern

The **Saga Pattern** ensures **eventual consistency** in distributed transactions by:
1. Breaking the transaction into a sequence of **local transactions**
2. Each local transaction updates one service and publishes an event
3. If a step fails, **compensating transactions** undo the previous steps
4. The system eventually reaches a consistent state (success or rollback)

### Saga Type: **Orchestration-Based Saga**

We will use **orchestration** (centralized coordinator) instead of choreography (event-driven) because:
- ✅ **Easier to understand** - Single place to see the entire workflow
- ✅ **Easier to debug** - Centralized logging and monitoring
- ✅ **Easier to modify** - Add/remove steps without changing multiple services
- ✅ **Better error handling** - Coordinator can implement sophisticated retry logic
- ❌ **Single point of failure** - Mitigated by making coordinator stateless and idempotent

---

## 2. Current Order Processing Flow

### 2.1 Order Submission Flow (Submit Order Use Case)

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ POST /api/v1/orders
       ▼
┌─────────────────────────────────────────────────────────┐
│              API Gateway (Authentication)                │
└──────────────────────────┬──────────────────────────────┘
                           │ gRPC: SubmitOrder
                           ▼
┌─────────────────────────────────────────────────────────┐
│              Order Service: SubmitOrderUseCase           │
│  1. Validate command                                     │
│  2. Check idempotency (Redis)                            │
│  3. Validate symbol (Market Data Service - gRPC)         │
│  4. Get current price (Market Data Service - gRPC)       │
│  5. Validate trading hours (Market Data Service - gRPC)  │
│  6. Validate order price                                 │
│  7. Create Order domain object (status = PENDING)        │
│  8. Save order to database                               │
│  9. Publish to RabbitMQ (orders.submit queue)            │
│ 10. Mark idempotency as completed                        │
│ 11. Return OrderID to client                             │
└──────────────────────────┬──────────────────────────────┘
                           │ 202 Accepted
                           ▼
┌─────────────┐
│   Client    │ (Receives OrderID)
└─────────────┘
```

**Key Characteristics:**
- **Synchronous validation** (Market Data Service calls)
- **Asynchronous processing** (RabbitMQ)
- **Idempotency** (Redis-based deduplication)
- **No balance check** (❌ Missing - must be added)

---

### 2.2 Order Processing Flow (Process Order Use Case)

```
┌─────────────────────────────────────────────────────────┐
│              RabbitMQ: orders.submit queue               │
└──────────────────────────┬──────────────────────────────┘
                           │ Consume message
                           ▼
┌─────────────────────────────────────────────────────────┐
│              Order Worker (Async Processing)             │
│  1. Receive order message from queue                     │
│  2. Call ProcessOrderUseCase                             │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│           Order Service: ProcessOrderUseCase             │
│  1. Find order by ID (database)                          │
│  2. Validate order can be processed                      │
│  3. Mark order as PROCESSING (database)                  │
│  4. Get real-time market data (Market Data Service)      │
│  5. Validate market conditions                           │
│  6. Calculate execution price                            │
│  7. Perform final risk checks                            │
│  8. Execute order (mark as EXECUTED in domain)           │
│  9. Update order status in database (EXECUTED)           │
│ 10. Publish OrderExecutedEvent to RabbitMQ               │
└──────────────────────────┬──────────────────────────────┘
                           │ OrderExecutedEvent
                           ▼
┌─────────────────────────────────────────────────────────┐
│         RabbitMQ: positions.updates queue                │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│              Position Service (Event Consumer)           │
│  1. Consume OrderExecutedEvent                           │
│  2. Update user position (create or update)              │
│  3. Calculate realized P&L                               │
│  4. Save position to database                            │
└─────────────────────────────────────────────────────────┘
```

**Key Characteristics:**
- **Asynchronous processing** (Worker-based)
- **Real-time market data** (gRPC call during execution)
- **Event-driven position updates** (RabbitMQ)
- **No balance deduction** (❌ Missing - must be added)
- **No rollback mechanism** (❌ Missing - saga pattern needed)

---

### 2.3 Current Issues (Why We Need Saga)

| Issue | Description | Impact |
|-------|-------------|--------|
| **No Balance Check** | Order submission doesn't verify sufficient funds | Orders can be submitted without available balance |
| **No Balance Deduction** | Balance is never deducted after execution | User can execute unlimited orders |
| **No Atomicity** | Order execution and position update are separate | Inconsistent state if position update fails |
| **No Rollback** | If position update fails, order stays EXECUTED | Manual intervention required |
| **No Compensation** | No way to undo partial transactions | Data inconsistency |

---

## 3. Saga Pattern Overview

### 3.1 What is a Saga?

A **saga** is a sequence of **local transactions** where:
1. Each transaction updates a single service
2. Each transaction publishes an event/message
3. If a transaction fails, **compensating transactions** undo previous steps
4. The saga either completes successfully or rolls back completely

### 3.2 Saga Types

#### Choreography-Based Saga (Event-Driven)
- Services listen to events and react
- No central coordinator
- **Pros:** Loose coupling, high scalability
- **Cons:** Hard to understand, debug, and modify

#### Orchestration-Based Saga (Coordinator)
- Central coordinator manages the workflow
- Coordinator calls services and handles failures
- **Pros:** Easy to understand, debug, and modify
- **Cons:** Single point of failure (mitigated by stateless design)

### 3.3 Our Choice: Orchestration-Based Saga

We choose **orchestration** because:
1. **Complexity:** Order processing involves 4+ services
2. **Debugging:** Centralized logging and monitoring
3. **Flexibility:** Easy to add/remove steps
4. **Error Handling:** Sophisticated retry and compensation logic

---

## 4. Order Processing Saga Design

### 4.1 Saga Coordinator

**Component:** `OrderProcessingSagaCoordinator`

**Responsibilities:**
1. Orchestrate the order processing workflow
2. Call each service in sequence
3. Handle failures and trigger compensations
4. Maintain saga state (in database or memory)
5. Publish events for observability

**Location:** `internal/order_mngmt_system/application/saga/`

---

### 4.2 Saga Steps (Happy Path)

```
┌─────────────────────────────────────────────────────────────┐
│                   Order Processing Saga                      │
└─────────────────────────────────────────────────────────────┘

Step 1: Validate Order
├─ Service: Order Service (internal)
├─ Action: Validate order can be processed
├─ Compensation: None (read-only)
└─ Output: Order validated

Step 2: Check Market Data
├─ Service: Market Data Service (gRPC)
├─ Action: Get real-time price and validate market conditions
├─ Compensation: None (read-only)
└─ Output: Market data retrieved

Step 3: Reserve Balance
├─ Service: Account/Balance Service (gRPC)
├─ Action: Reserve funds for order (balance - order value)
├─ Compensation: Release reserved balance
└─ Output: Balance reserved (reservation ID)

Step 4: Mark Order as Processing
├─ Service: Order Service (internal)
├─ Action: Update order status to PROCESSING
├─ Compensation: Revert order status to PENDING
└─ Output: Order marked as PROCESSING

Step 5: Execute Order (Simulate Trade)
├─ Service: Order Service (internal)
├─ Action: Calculate execution price, mark as EXECUTED
├─ Compensation: Revert order status to FAILED
└─ Output: Order executed (execution price)

Step 6: Deduct Balance
├─ Service: Account/Balance Service (gRPC)
├─ Action: Deduct order value from user balance
├─ Compensation: Credit balance back
└─ Output: Balance deducted

Step 7: Update Position
├─ Service: Position Service (RabbitMQ event)
├─ Action: Publish OrderExecutedEvent to positions.updates queue
├─ Compensation: Publish OrderExecutionFailedEvent (position service reverts)
└─ Output: Position update event published

Step 8: Finalize Order
├─ Service: Order Service (internal)
├─ Action: Persist execution details, mark saga as complete
├─ Compensation: None (final step)
└─ Output: Saga completed successfully
```

---

### 4.3 Saga State Machine

```
┌──────────┐
│  STARTED │ (Saga initiated)
└────┬─────┘
     │
     ▼
┌──────────────────┐
│ VALIDATING_ORDER │ (Step 1)
└────┬─────────────┘
     │
     ▼
┌───────────────────┐
│ CHECKING_MARKET   │ (Step 2)
└────┬──────────────┘
     │
     ▼
┌───────────────────┐
│ RESERVING_BALANCE │ (Step 3)
└────┬──────────────┘
     │
     ▼
┌───────────────────┐
│ PROCESSING_ORDER  │ (Step 4)
└────┬──────────────┘
     │
     ▼
┌───────────────────┐
│ EXECUTING_ORDER   │ (Step 5)
└────┬──────────────┘
     │
     ▼
┌───────────────────┐
│ DEDUCTING_BALANCE │ (Step 6)
└────┬──────────────┘
     │
     ▼
┌───────────────────┐
│ UPDATING_POSITION │ (Step 7)
└────┬──────────────┘
     │
     ▼
┌───────────────────┐
│ FINALIZING_ORDER  │ (Step 8)
└────┬──────────────┘
     │
     ▼
┌──────────┐
│ COMPLETED│ (Success)
└──────────┘

     (If any step fails)
     │
     ▼
┌──────────────┐
│ COMPENSATING │ (Rolling back)
└────┬─────────┘
     │
     ▼
┌──────────┐
│  FAILED  │ (Rollback complete)
└──────────┘
```

---

### 4.4 Saga Execution Flow (Sequence Diagram)

```
Client          API Gateway    Order Service    Market Data    Balance Service    Position Service
  │                 │                │                │                │                  │
  │ POST /orders    │                │                │                │                  │
  ├────────────────>│                │                │                │                  │
  │                 │ SubmitOrder    │                │                │                  │
  │                 ├───────────────>│                │                │                  │
  │                 │                │ Validate       │                │                  │
  │                 │                │ Symbol         │                │                  │
  │                 │                ├───────────────>│                │                  │
  │                 │                │<───────────────┤                │                  │
  │                 │                │ (Price data)   │                │                  │
  │                 │                │                │                │                  │
  │                 │                │ Save Order     │                │                  │
  │                 │                │ (PENDING)      │                │                  │
  │                 │                │                │                │                  │
  │                 │<───────────────┤                │                │                  │
  │<────────────────┤ 202 Accepted   │                │                │                  │
  │ (OrderID)       │                │                │                │                  │
  │                 │                │                │                │                  │
  │                 │                │ [Async: Saga Coordinator]       │                  │
  │                 │                │                │                │                  │
  │                 │                │ Step 1: Validate Order          │                  │
  │                 │                │ ✓              │                │                  │
  │                 │                │                │                │                  │
  │                 │                │ Step 2: Get Market Data         │                  │
  │                 │                ├───────────────>│                │                  │
  │                 │                │<───────────────┤                │                  │
  │                 │                │ ✓              │                │                  │
  │                 │                │                │                │                  │
  │                 │                │ Step 3: Reserve Balance         │                  │
  │                 │                ├───────────────────────────────>│                  │
  │                 │                │<───────────────────────────────┤                  │
  │                 │                │ ✓ (Reservation ID)              │                  │
  │                 │                │                │                │                  │
  │                 │                │ Step 4: Mark as PROCESSING      │                  │
  │                 │                │ ✓              │                │                  │
  │                 │                │                │                │                  │
  │                 │                │ Step 5: Execute Order           │                  │
  │                 │                │ ✓              │                │                  │
  │                 │                │                │                │                  │
  │                 │                │ Step 6: Deduct Balance          │                  │
  │                 │                ├───────────────────────────────>│                  │
  │                 │                │<───────────────────────────────┤                  │
  │                 │                │ ✓              │                │                  │
  │                 │                │                │                │                  │
  │                 │                │ Step 7: Publish OrderExecutedEvent               │
  │                 │                ├────────────────────────────────────────────────>│
  │                 │                │                │                │                  │
  │                 │                │ Step 8: Finalize Order          │                  │
  │                 │                │ ✓ (COMPLETED)  │                │                  │
```

---

## 5. Compensating Transactions

### 5.1 Compensation Strategy

**Principle:** For every **forward transaction**, define a **compensating transaction** that undoes its effect.

**Compensation Order:** **Reverse order** of execution (LIFO - Last In, First Out)

---

### 5.2 Compensation Matrix

| Step | Forward Transaction | Compensating Transaction | Idempotent? |
|------|---------------------|--------------------------|-------------|
| **1. Validate Order** | Validate order state | None (read-only) | ✅ Yes |
| **2. Check Market Data** | Get market data | None (read-only) | ✅ Yes |
| **3. Reserve Balance** | `ReserveBalance(userID, amount)` | `ReleaseBalance(userID, reservationID)` | ✅ Yes |
| **4. Mark as Processing** | Update order status to PROCESSING | Update order status to PENDING | ✅ Yes |
| **5. Execute Order** | Mark order as EXECUTED | Mark order as FAILED | ✅ Yes |
| **6. Deduct Balance** | `DeductBalance(userID, amount, orderID)` | `CreditBalance(userID, amount, orderID)` | ✅ Yes |
| **7. Update Position** | Publish `OrderExecutedEvent` | Publish `OrderExecutionFailedEvent` | ✅ Yes |
| **8. Finalize Order** | Mark saga as COMPLETED | None (final step) | ✅ Yes |

---

### 5.3 Compensation Scenarios

#### Scenario 1: Market Data Service Unavailable (Step 2 Fails)

```
Step 1: Validate Order ✓
Step 2: Check Market Data ✗ (Service unavailable)

Compensation:
- None needed (no state changes yet)

Final State:
- Order status: PENDING
- Balance: Unchanged
- Position: Unchanged
```

---

#### Scenario 2: Insufficient Balance (Step 3 Fails)

```
Step 1: Validate Order ✓
Step 2: Check Market Data ✓
Step 3: Reserve Balance ✗ (Insufficient funds)

Compensation:
- None needed (reservation failed, no state change)

Final State:
- Order status: PENDING
- Balance: Unchanged
- Position: Unchanged
- Error: "Insufficient balance"
```

---

#### Scenario 3: Balance Service Fails During Deduction (Step 6 Fails)

```
Step 1: Validate Order ✓
Step 2: Check Market Data ✓
Step 3: Reserve Balance ✓ (Reservation ID: res_123)
Step 4: Mark as Processing ✓
Step 5: Execute Order ✓ (Execution price: $150.25)
Step 6: Deduct Balance ✗ (Service timeout)

Compensation (Reverse Order):
- Step 5: Revert order status to FAILED
- Step 4: Revert order status to PENDING
- Step 3: Release reserved balance (res_123)

Final State:
- Order status: FAILED
- Balance: Unchanged (reservation released)
- Position: Unchanged
- Error: "Failed to deduct balance: service timeout"
```

---

#### Scenario 4: Position Service Fails (Step 7 Fails)

```
Step 1: Validate Order ✓
Step 2: Check Market Data ✓
Step 3: Reserve Balance ✓ (Reservation ID: res_123)
Step 4: Mark as Processing ✓
Step 5: Execute Order ✓ (Execution price: $150.25)
Step 6: Deduct Balance ✓ (Balance: $10,000 → $8,497.50)
Step 7: Update Position ✗ (RabbitMQ publish failed)

Compensation (Reverse Order):
- Step 6: Credit balance back (+$1,502.50)
- Step 5: Revert order status to FAILED
- Step 4: Revert order status to PENDING
- Step 3: Release reserved balance (res_123)

Final State:
- Order status: FAILED
- Balance: $10,000 (credited back)
- Position: Unchanged
- Error: "Failed to update position: message queue unavailable"
```

---

### 5.4 Idempotency Requirements

**All saga steps and compensations MUST be idempotent** to handle:
- Network retries
- Duplicate messages
- Partial failures

**Implementation:**
- Use **idempotency keys** (order ID + step ID)
- Store idempotency state in **Redis** or **database**
- Check idempotency before executing each step

**Example:**
```go
func (s *SagaCoordinator) ReserveBalance(ctx context.Context, orderID string, userID string, amount float64) error {
    idempotencyKey := fmt.Sprintf("saga:%s:step:reserve_balance", orderID)
    
    // Check if already executed
    if s.idempotencyService.IsCompleted(ctx, idempotencyKey) {
        return nil // Already done, skip
    }
    
    // Execute
    reservationID, err := s.balanceClient.ReserveBalance(ctx, userID, amount, orderID)
    if err != nil {
        return err
    }
    
    // Mark as completed
    s.idempotencyService.MarkCompleted(ctx, idempotencyKey, reservationID)
    return nil
}
```

---

## 6. Implementation Strategy

### 6.1 Saga Coordinator Interface

```go
// ISagaCoordinator defines the interface for saga orchestration
type ISagaCoordinator interface {
    // ExecuteOrderProcessingSaga orchestrates the entire order processing workflow
    ExecuteOrderProcessingSaga(ctx context.Context, orderID string) (*SagaResult, error)
    
    // GetSagaStatus retrieves the current status of a saga
    GetSagaStatus(ctx context.Context, sagaID string) (*SagaStatus, error)
    
    // CompensateSaga triggers compensation for a failed saga
    CompensateSaga(ctx context.Context, sagaID string) error
}

// SagaResult contains the result of saga execution
type SagaResult struct {
    SagaID         string
    OrderID        string
    Status         SagaStatus
    CompletedSteps []string
    FailedStep     string
    ErrorMessage   string
    StartTime      time.Time
    EndTime        time.Time
    Duration       time.Duration
}

// SagaStatus represents the current state of a saga
type SagaStatus string

const (
    SagaStatusStarted      SagaStatus = "STARTED"
    SagaStatusInProgress   SagaStatus = "IN_PROGRESS"
    SagaStatusCompleted    SagaStatus = "COMPLETED"
    SagaStatusCompensating SagaStatus = "COMPENSATING"
    SagaStatusFailed       SagaStatus = "FAILED"
)
```

---

### 6.2 Saga Coordinator Implementation

```go
type OrderProcessingSagaCoordinator struct {
    orderRepository    repository.IOrderRepository
    marketDataClient   external.IMarketDataClient
    balanceClient      external.IBalanceServiceClient // NEW
    eventPublisher     messaging.IEventPublisher
    idempotencyService service.IIdempotencyService
    sagaRepository     repository.ISagaRepository // NEW
}

func (c *OrderProcessingSagaCoordinator) ExecuteOrderProcessingSaga(
    ctx context.Context,
    orderID string,
) (*SagaResult, error) {
    sagaID := generateSagaID(orderID)
    
    saga := &Saga{
        ID:        sagaID,
        OrderID:   orderID,
        Status:    SagaStatusStarted,
        StartTime: time.Now(),
    }
    
    // Save saga state
    if err := c.sagaRepository.Save(ctx, saga); err != nil {
        return nil, fmt.Errorf("failed to save saga: %w", err)
    }
    
    // Execute steps in sequence
    steps := []SagaStep{
        {Name: "ValidateOrder", Execute: c.validateOrder, Compensate: nil},
        {Name: "CheckMarketData", Execute: c.checkMarketData, Compensate: nil},
        {Name: "ReserveBalance", Execute: c.reserveBalance, Compensate: c.releaseBalance},
        {Name: "MarkAsProcessing", Execute: c.markAsProcessing, Compensate: c.revertToPending},
        {Name: "ExecuteOrder", Execute: c.executeOrder, Compensate: c.markAsFailed},
        {Name: "DeductBalance", Execute: c.deductBalance, Compensate: c.creditBalance},
        {Name: "UpdatePosition", Execute: c.updatePosition, Compensate: c.revertPosition},
        {Name: "FinalizeOrder", Execute: c.finalizeOrder, Compensate: nil},
    }
    
    // Execute each step
    for i, step := range steps {
        saga.CurrentStep = step.Name
        saga.Status = SagaStatusInProgress
        c.sagaRepository.Update(ctx, saga)
        
        if err := step.Execute(ctx, orderID); err != nil {
            // Step failed, trigger compensation
            saga.Status = SagaStatusCompensating
            saga.FailedStep = step.Name
            saga.ErrorMessage = err.Error()
            c.sagaRepository.Update(ctx, saga)
            
            // Compensate in reverse order
            if err := c.compensate(ctx, orderID, steps[:i]); err != nil {
                saga.Status = SagaStatusFailed
                c.sagaRepository.Update(ctx, saga)
                return nil, fmt.Errorf("compensation failed: %w", err)
            }
            
            saga.Status = SagaStatusFailed
            saga.EndTime = time.Now()
            c.sagaRepository.Update(ctx, saga)
            
            return &SagaResult{
                SagaID:       sagaID,
                OrderID:      orderID,
                Status:       SagaStatusFailed,
                FailedStep:   step.Name,
                ErrorMessage: err.Error(),
                StartTime:    saga.StartTime,
                EndTime:      saga.EndTime,
                Duration:     saga.EndTime.Sub(saga.StartTime),
            }, nil
        }
        
        saga.CompletedSteps = append(saga.CompletedSteps, step.Name)
    }
    
    // All steps completed successfully
    saga.Status = SagaStatusCompleted
    saga.EndTime = time.Now()
    c.sagaRepository.Update(ctx, saga)
    
    return &SagaResult{
        SagaID:         sagaID,
        OrderID:        orderID,
        Status:         SagaStatusCompleted,
        CompletedSteps: saga.CompletedSteps,
        StartTime:      saga.StartTime,
        EndTime:        saga.EndTime,
        Duration:       saga.EndTime.Sub(saga.StartTime),
    }, nil
}

func (c *OrderProcessingSagaCoordinator) compensate(
    ctx context.Context,
    orderID string,
    steps []SagaStep,
) error {
    // Compensate in reverse order (LIFO)
    for i := len(steps) - 1; i >= 0; i-- {
        step := steps[i]
        if step.Compensate != nil {
            if err := step.Compensate(ctx, orderID); err != nil {
                // Log error but continue compensating
                log.Printf("Compensation failed for step %s: %v", step.Name, err)
                // In production, you might want to retry or alert
            }
        }
    }
    return nil
}
```

---

### 6.3 Saga Persistence

**Table:** `order_sagas`

```sql
CREATE TABLE order_sagas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_step VARCHAR(100),
    completed_steps JSONB DEFAULT '[]',
    failed_step VARCHAR(100),
    error_message TEXT,
    start_time TIMESTAMP NOT NULL DEFAULT NOW(),
    end_time TIMESTAMP,
    CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE INDEX idx_order_sagas_order_id ON order_sagas(order_id);
CREATE INDEX idx_order_sagas_status ON order_sagas(status);
CREATE INDEX idx_order_sagas_start_time ON order_sagas(start_time DESC);
```

---

## 7. Error Handling and Recovery

### 7.1 Retry Strategy

**Transient Failures:** Retry with exponential backoff
- Network timeouts
- Service temporarily unavailable
- Database connection errors

**Permanent Failures:** Trigger compensation immediately
- Insufficient balance
- Invalid order state
- Business rule violations

**Retry Configuration:**
```go
type RetryConfig struct {
    MaxAttempts    int           // 3
    InitialDelay   time.Duration // 1 second
    MaxDelay       time.Duration // 30 seconds
    BackoffFactor  float64       // 2.0 (exponential)
}
```

---

### 7.2 Timeout Strategy

Each saga step has a **timeout**:
- **Market Data calls:** 5 seconds
- **Balance Service calls:** 10 seconds
- **Database operations:** 5 seconds
- **RabbitMQ publish:** 3 seconds

**Total Saga Timeout:** 2 minutes (after which saga is marked as FAILED)

---

### 7.3 Dead Letter Queue (DLQ)

Failed sagas (after all retries) are sent to a **Dead Letter Queue** for:
- Manual review
- Alerting
- Reprocessing

---

## 8. Monitoring and Observability

### 8.1 Metrics

| Metric | Description | Type |
|--------|-------------|------|
| `saga_executions_total` | Total saga executions | Counter |
| `saga_executions_success` | Successful saga executions | Counter |
| `saga_executions_failed` | Failed saga executions | Counter |
| `saga_compensations_total` | Total compensations triggered | Counter |
| `saga_duration_seconds` | Saga execution duration | Histogram |
| `saga_step_duration_seconds` | Individual step duration | Histogram |

---

### 8.2 Distributed Tracing

Use **OpenTelemetry** or **Jaeger** to trace saga execution across services:
- Trace ID: Saga ID
- Span per saga step
- Tag each span with step name and status

---

### 8.3 Logging

**Log Level:** INFO for success, ERROR for failures

**Log Format:**
```json
{
  "saga_id": "saga_abc123",
  "order_id": "order_xyz789",
  "step": "ReserveBalance",
  "status": "SUCCESS",
  "duration_ms": 45,
  "timestamp": "2025-11-04T10:30:00Z"
}
```

---

## 9. Next Steps

### Implementation Phases

**Phase 1: Create Balance Service** (4-6 weeks)
- [ ] Design balance database schema
- [ ] Implement balance operations (Reserve, Deduct, Credit, Release)
- [ ] Create gRPC server
- [ ] Deploy to development

**Phase 2: Implement Saga Coordinator** (2-3 weeks)
- [ ] Create saga coordinator interface
- [ ] Implement saga steps
- [ ] Implement compensating transactions
- [ ] Add saga persistence
- [ ] Add idempotency checks

**Phase 3: Integration** (2-3 weeks)
- [ ] Replace ProcessOrderUseCase with SagaCoordinator
- [ ] Update order workers to use saga
- [ ] Add monitoring and observability
- [ ] Write integration tests

**Phase 4: Testing** (2-3 weeks)
- [ ] Unit tests for each saga step
- [ ] Integration tests for happy path
- [ ] Chaos testing for failure scenarios
- [ ] Load testing

**Phase 5: Deployment** (1-2 weeks)
- [ ] Deploy to staging
- [ ] Run validation tests
- [ ] Deploy to production (gradual rollout)
- [ ] Monitor for issues

---

## 10. Conclusion

The **Saga Pattern** is **essential** for the Order Management Service to ensure:
- ✅ **Data Consistency** across distributed services
- ✅ **Fault Tolerance** with automatic compensation
- ✅ **Observability** with centralized monitoring
- ✅ **Maintainability** with clear workflow definition

**Critical Dependency:** Account/Balance Service must be created before implementing the saga.

**Estimated Effort:** 10-15 weeks (including Balance Service creation)

---

**Document Version:** 1.0  
**Last Updated:** November 4, 2025  
**Author:** AI Assistant  
**Status:** ✅ COMPLETED

