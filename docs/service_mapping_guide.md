# Microservices Service Mapping Guide

## Overview

This document provides detailed mapping from the current monolithic structure to the proposed microservices architecture, including implementation guidelines and migration steps.

## Service Mapping Matrix

| Current Module | Target Microservice | Migration Priority | Complexity | Dependencies |
|---|---|---|---|---|
| `internal/auth/` | User Management Service | P1 - High | Low | None |
| `internal/login/` | User Management Service | P1 - High | Low | Auth |
| `internal/market_data/` | Market Data Service | P1 - High | Medium | None |
| `internal/realtime_quotes/` | Market Data Service | P1 - High | Medium | Market Data, WebSocket |
| `internal/watchlist/` | Watchlist Service | P2 - Medium | Low | Market Data, Auth |
| `internal/balance/` | Account Management Service | P2 - Medium | Medium | Auth |
| `internal/position/` | Position & Portfolio Service | P3 - High | High | Market Data, Auth |
| `internal/portfolio_summary/` | Position & Portfolio Service | P3 - High | Medium | Position, Balance |
| `internal/order_mngmt_system/` | Order Management Service | P4 - Critical | Very High | All Services |

## Detailed Service Specifications

### 1. User Management Service

**Service Name:** `hub-user-service`  
**Repository:** `hub-investments-user-service`  
**Port Assignment:** HTTP: 8081, gRPC: 50051

#### Current Modules Integration:
```
internal/auth/           → User Management Service
├── auth_service.go      → core/auth_service.go
├── token/
│   └── token_service.go → core/token_service.go
└── presentation/grpc/   → grpc/auth_server.go

internal/login/          → User Management Service  
├── application/usecase/ → usecase/login_usecase.go
├── domain/model/        → domain/user.go
├── domain/valueobject/  → domain/email.go, password.go
└── infra/persistense/   → repository/user_repository.go
```

#### Service Structure:
```
hub-user-service/
├── cmd/server/main.go
├── internal/
│   ├── core/                    # Business logic
│   │   ├── auth_service.go
│   │   ├── token_service.go
│   │   └── user_service.go
│   ├── domain/                  # Domain models
│   │   ├── user.go
│   │   ├── email.go
│   │   └── password.go
│   ├── usecase/                 # Application layer
│   │   ├── login_usecase.go
│   │   ├── register_usecase.go
│   │   └── token_usecase.go
│   ├── repository/              # Data layer
│   │   └── user_repository.go
│   ├── grpc/                    # gRPC interface
│   │   └── auth_server.go
│   └── http/                    # HTTP interface
│       └── auth_handler.go
├── config/
│   └── config.yaml
└── migrations/
    └── 001_create_users.sql
```

#### API Contracts:
```protobuf
service AuthService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc Register(RegisterRequest) returns (RegisterResponse);
}
```

### 2. Market Data Service

**Service Name:** `hub-market-service`  
**Repository:** `hub-investments-market-service`  
**Port Assignment:** HTTP: 8082, gRPC: 50052, WebSocket: 8092

#### Current Modules Integration:
```
internal/market_data/           → Market Data Service
├── application/usecase/        → usecase/market_data_usecase.go
├── domain/model/              → domain/market_data.go
├── infra/cache/               → cache/market_cache.go
├── infra/persistence/         → repository/market_repository.go
└── presentation/grpc/         → grpc/market_server.go

internal/realtime_quotes/       → Market Data Service
├── application/service/        → streaming/price_service.go
├── domain/service/            → streaming/asset_service.go
└── infra/websocket/           → websocket/quotes_handler.go
```

#### Service Structure:
```
hub-market-service/
├── cmd/server/main.go
├── internal/
│   ├── core/                    # Business logic
│   │   ├── market_service.go
│   │   └── price_service.go
│   ├── domain/                  # Domain models
│   │   ├── market_data.go
│   │   ├── asset.go
│   │   └── price.go
│   ├── usecase/                 # Application layer
│   │   ├── get_market_data.go
│   │   └── stream_prices.go
│   ├── repository/              # Data layer
│   │   └── market_repository.go
│   ├── cache/                   # Caching layer
│   │   └── market_cache.go
│   ├── streaming/               # Real-time streaming
│   │   ├── price_oscillation.go
│   │   └── websocket_manager.go
│   ├── grpc/                    # gRPC interface
│   │   └── market_server.go
│   ├── http/                    # HTTP interface
│   │   └── market_handler.go
│   └── websocket/               # WebSocket interface
│       └── quotes_handler.go
├── external/                    # External integrations
│   └── price_feed_client.go
└── config/
    └── config.yaml
```

#### API Contracts:
```protobuf
service MarketDataService {
  rpc GetMarketData(GetMarketDataRequest) returns (GetMarketDataResponse);
  rpc ValidateSymbol(ValidateSymbolRequest) returns (ValidateSymbolResponse);
  rpc GetCurrentPrice(GetPriceRequest) returns (GetPriceResponse);
  rpc GetTradingHours(TradingHoursRequest) returns (TradingHoursResponse);
  rpc StreamPrices(StreamPricesRequest) returns (stream PriceUpdate);
}
```

### 3. Account Management Service

**Service Name:** `hub-account-service`  
**Repository:** `hub-investments-account-service`  
**Port Assignment:** HTTP: 8085, gRPC: 50055

#### Current Modules Integration:
```
internal/balance/               → Account Management Service
├── application/usecase/        → usecase/balance_usecase.go
├── domain/model/              → domain/balance.go, account.go
├── domain/repository/         → domain/account_repository.go
└── infra/persistence/         → repository/account_repository.go
```

#### Service Structure:
```
hub-account-service/
├── cmd/server/main.go
├── internal/
│   ├── core/                    # Business logic
│   │   ├── account_service.go
│   │   └── transaction_service.go
│   ├── domain/                  # Domain models
│   │   ├── account.go
│   │   ├── balance.go
│   │   └── transaction.go
│   ├── usecase/                 # Application layer
│   │   ├── get_balance.go
│   │   ├── transfer_funds.go
│   │   └── reserve_funds.go
│   ├── repository/              # Data layer
│   │   └── account_repository.go
│   ├── grpc/                    # gRPC interface
│   │   └── account_server.go
│   └── http/                    # HTTP interface
│       └── account_handler.go
└── external/                    # External integrations
    └── banking_client.go
```

#### API Contracts:
```protobuf
service AccountService {
  rpc GetBalance(GetBalanceRequest) returns (GetBalanceResponse);
  rpc CheckBalance(CheckBalanceRequest) returns (CheckBalanceResponse);
  rpc ReserveFunds(ReserveFundsRequest) returns (ReserveFundsResponse);
  rpc ReleaseFunds(ReleaseFundsRequest) returns (ReleaseFundsResponse);
  rpc TransferFunds(TransferFundsRequest) returns (TransferFundsResponse);
}
```

### 4. Position & Portfolio Service

**Service Name:** `hub-portfolio-service`  
**Repository:** `hub-investments-portfolio-service`  
**Port Assignment:** HTTP: 8084, gRPC: 50054

#### Current Modules Integration:
```
internal/position/              → Position & Portfolio Service
├── application/usecase/        → usecase/position_usecase.go
├── domain/model/              → domain/position.go
├── domain/repository/         → domain/position_repository.go
├── infra/persistence/         → repository/position_repository.go
└── infra/worker/              → worker/position_worker.go

internal/portfolio_summary/     → Position & Portfolio Service
├── application/usecase/        → usecase/portfolio_usecase.go
└── domain/model/              → domain/portfolio.go
```

#### Service Structure:
```
hub-portfolio-service/
├── cmd/server/main.go
├── internal/
│   ├── core/                    # Business logic
│   │   ├── position_service.go
│   │   └── portfolio_service.go
│   ├── domain/                  # Domain models
│   │   ├── position.go
│   │   ├── portfolio.go
│   │   └── aggregation.go
│   ├── usecase/                 # Application layer
│   │   ├── get_positions.go
│   │   ├── update_position.go
│   │   └── portfolio_summary.go
│   ├── repository/              # Data layer
│   │   └── position_repository.go
│   ├── worker/                  # Event processing
│   │   └── position_worker.go
│   ├── grpc/                    # gRPC interface
│   │   └── portfolio_server.go
│   └── http/                    # HTTP interface
│       └── portfolio_handler.go
└── events/                      # Event handling
    └── order_event_handler.go
```

#### API Contracts:
```protobuf
service PositionService {
  rpc GetPositions(GetPositionsRequest) returns (GetPositionsResponse);
  rpc GetPositionAggregation(GetAggregationRequest) returns (GetAggregationResponse);
  rpc UpdatePosition(UpdatePositionRequest) returns (UpdatePositionResponse);
  rpc GetPortfolioSummary(PortfolioRequest) returns (PortfolioResponse);
}
```

### 5. Watchlist Service

**Service Name:** `hub-watchlist-service`  
**Repository:** `hub-investments-watchlist-service`  
**Port Assignment:** HTTP: 8086, gRPC: 50056

#### Current Modules Integration:
```
internal/watchlist/             → Watchlist Service
├── application/usecase/        → usecase/watchlist_usecase.go
├── domain/repository/         → domain/watchlist_repository.go
└── infra/persistence/         → repository/watchlist_repository.go
```

#### Service Structure:
```
hub-watchlist-service/
├── cmd/server/main.go
├── internal/
│   ├── core/                    # Business logic
│   │   ├── watchlist_service.go
│   │   └── alert_service.go
│   ├── domain/                  # Domain models
│   │   ├── watchlist.go
│   │   └── price_alert.go
│   ├── usecase/                 # Application layer
│   │   ├── manage_watchlist.go
│   │   └── price_alerts.go
│   ├── repository/              # Data layer
│   │   └── watchlist_repository.go
│   ├── grpc/                    # gRPC interface
│   │   └── watchlist_server.go
│   └── http/                    # HTTP interface
│       └── watchlist_handler.go
└── notifications/               # Alert system
    └── alert_processor.go
```

### 6. Order Management Service

**Service Name:** `hub-order-service`  
**Repository:** `hub-investments-order-service`  
**Port Assignment:** HTTP: 8083, gRPC: 50053

#### Current Modules Integration:
```
internal/order_mngmt_system/    → Order Management Service
├── application/
│   ├── usecase/               → usecase/order_usecase.go
│   └── command/               → command/order_commands.go
├── domain/
│   ├── model/                 → domain/order.go
│   ├── repository/            → domain/order_repository.go
│   └── service/               → service/order_services.go
├── infra/
│   ├── persistence/           → repository/order_repository.go
│   ├── external/              → client/market_client.go
│   ├── messaging/             → messaging/order_events.go
│   └── worker/                → worker/order_worker.go
└── presentation/              → grpc/order_server.go, http/order_handler.go
```

#### Service Structure:
```
hub-order-service/
├── cmd/server/main.go
├── internal/
│   ├── core/                    # Business logic
│   │   ├── order_service.go
│   │   ├── validation_service.go
│   │   ├── pricing_service.go
│   │   └── risk_service.go
│   ├── domain/                  # Domain models
│   │   ├── order.go
│   │   ├── order_status.go
│   │   ├── order_type.go
│   │   └── order_events.go
│   ├── usecase/                 # Application layer
│   │   ├── submit_order.go
│   │   ├── process_order.go
│   │   ├── cancel_order.go
│   │   └── get_order_status.go
│   ├── command/                 # Command objects
│   │   ├── submit_command.go
│   │   └── cancel_command.go
│   ├── repository/              # Data layer
│   │   └── order_repository.go
│   ├── client/                  # External clients
│   │   ├── market_client.go
│   │   ├── account_client.go
│   │   └── broker_client.go
│   ├── messaging/               # Event handling
│   │   ├── order_producer.go
│   │   ├── order_consumer.go
│   │   └── saga_orchestrator.go
│   ├── worker/                  # Background processing
│   │   ├── order_worker.go
│   │   └── worker_manager.go
│   ├── grpc/                    # gRPC interface
│   │   └── order_server.go
│   └── http/                    # HTTP interface
│       └── order_handler.go
└── saga/                        # Distributed transaction
    └── order_saga.go
```

## Migration Implementation Steps

### Phase 1: Infrastructure Setup (Month 1)

#### 1.1 Kubernetes Cluster Setup
```bash
# Create namespaces
kubectl create namespace hub-services
kubectl create namespace hub-infrastructure
kubectl create namespace hub-monitoring

# Deploy infrastructure
kubectl apply -f k8s/infrastructure/
├── rabbitmq-cluster.yaml
├── redis-cluster.yaml
├── postgres-databases.yaml
└── monitoring-stack.yaml
```

#### 1.2 Service Mesh Installation
```bash
# Install Istio
curl -L https://istio.io/downloadIstio | sh -
istioctl install --set values.defaultRevision=default
kubectl label namespace hub-services istio-injection=enabled
```

#### 1.3 CI/CD Pipeline Setup
```yaml
# .github/workflows/service-deploy.yml
name: Deploy Microservice
on:
  push:
    paths:
      - 'services/*/cmd/**'
      - 'services/*/internal/**'
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Build and Deploy
        run: |
          docker build -t $SERVICE_NAME:$GITHUB_SHA .
          kubectl set image deployment/$SERVICE_NAME $SERVICE_NAME=$SERVICE_NAME:$GITHUB_SHA
```

### Phase 2: Service Extraction (Months 2-4)

#### 2.1 Extract User Management Service
```bash
# Create new repository
mkdir hub-user-service
cd hub-user-service

# Copy and refactor code
cp -r ../HubInvestmentsServer/internal/auth/ ./internal/core/
cp -r ../HubInvestmentsServer/internal/login/ ./internal/

# Create service-specific files
cat > cmd/server/main.go << EOF
package main

import (
    "log"
    "net"
    
    "hub-user-service/internal/grpc"
    "hub-user-service/internal/http"
    "google.golang.org/grpc"
)

func main() {
    // Start gRPC server
    lis, _ := net.Listen("tcp", ":50051")
    s := grpc.NewServer()
    authServer := grpc.NewAuthServer()
    RegisterAuthServiceServer(s, authServer)
    
    go func() {
        log.Fatal(s.Serve(lis))
    }()
    
    // Start HTTP server
    log.Fatal(http.StartServer(":8081"))
}
EOF

# Create Dockerfile
cat > Dockerfile << EOF
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /server
EXPOSE 8081 50051
CMD ["./server"]
EOF
```

#### 2.2 Database Migration Strategy
```sql
-- Create service-specific databases
CREATE DATABASE hub_users_db;
CREATE DATABASE hub_market_db; 
CREATE DATABASE hub_orders_db;
CREATE DATABASE hub_portfolio_db;
CREATE DATABASE hub_accounts_db;
CREATE DATABASE hub_watchlist_db;

-- Migration script for users data
INSERT INTO hub_users_db.users 
SELECT id, email, password, created_at, updated_at 
FROM hub_investments_db.users;
```

### Phase 3: Event-Driven Architecture (Months 3-5)

#### 3.1 Event Schema Definition
```protobuf
// events/order_events.proto
syntax = "proto3";

message OrderExecutedEvent {
  string event_id = 1;
  string order_id = 2;
  string user_id = 3;
  string symbol = 4;
  double quantity = 5;
  double execution_price = 6;
  string order_side = 7; // BUY/SELL
  int64 executed_at = 8;
  string source_service = 9;
}

message PositionUpdatedEvent {
  string event_id = 1;
  string position_id = 2;
  string user_id = 3;
  string symbol = 4;
  double new_quantity = 5;
  double new_average_price = 6;
  double total_investment = 7;
  string transaction_type = 8;
  int64 updated_at = 9;
}
```

#### 3.2 Event Publisher Implementation
```go
// messaging/event_publisher.go
type EventPublisher struct {
    channel *amqp.Channel
}

func (p *EventPublisher) PublishOrderExecutedEvent(event *OrderExecutedEvent) error {
    body, err := proto.Marshal(event)
    if err != nil {
        return err
    }
    
    return p.channel.Publish(
        "order.events",    // exchange
        "order.executed",  // routing key
        false,            // mandatory
        false,            // immediate
        amqp.Publishing{
            ContentType: "application/x-protobuf",
            Body:        body,
            Headers: amqp.Table{
                "event_type": "OrderExecutedEvent",
                "version":    "1.0",
            },
        },
    )
}
```

#### 3.3 Saga Pattern Implementation
```go
// saga/order_execution_saga.go
type OrderExecutionSaga struct {
    orderID     string
    steps       []SagaStep
    currentStep int
    completed   bool
    compensated bool
}

type SagaStep struct {
    Name        string
    Execute     func(ctx context.Context) error
    Compensate  func(ctx context.Context) error
    Completed   bool
}

func (s *OrderExecutionSaga) Execute(ctx context.Context) error {
    for i, step := range s.steps {
        s.currentStep = i
        if err := step.Execute(ctx); err != nil {
            // Compensate previous steps
            return s.compensate(ctx)
        }
        step.Completed = true
    }
    s.completed = true
    return nil
}

func (s *OrderExecutionSaga) compensate(ctx context.Context) error {
    for i := s.currentStep - 1; i >= 0; i-- {
        if s.steps[i].Completed {
            if err := s.steps[i].Compensate(ctx); err != nil {
                log.Printf("Compensation failed for step %s: %v", s.steps[i].Name, err)
            }
        }
    }
    s.compensated = true
    return errors.New("saga execution failed and compensated")
}
```

### Phase 4: Monitoring and Observability (Month 4-6)

#### 4.1 Prometheus Metrics
```go
// monitoring/metrics.go
var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"service", "method", "endpoint", "status"},
    )
    
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2.5, 5, 10},
        },
        []string{"service", "method", "endpoint"},
    )
    
    grpcRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total", 
            Help: "Total number of gRPC requests",
        },
        []string{"service", "method", "status"},
    )
)
```

#### 4.2 Distributed Tracing
```go
// tracing/tracer.go
func InitTracing(serviceName string) {
    exporter, err := jaeger.New(
        jaeger.WithCollectorEndpoint(
            jaeger.WithEndpoint("http://jaeger-collector:14268/api/traces"),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
            semconv.ServiceVersionKey.String("1.0.0"),
        )),
    )
    
    otel.SetTracerProvider(tp)
}
```

### Phase 5: Service Communication Testing (Month 5-7)

#### 5.1 Contract Testing
```go
// contracts/auth_service_test.go
func TestAuthServiceContract(t *testing.T) {
    // Set up gRPC server
    server := grpc.NewServer()
    authService := &AuthServiceImpl{}
    RegisterAuthServiceServer(server, authService)
    
    // Test contract compliance
    client := NewAuthServiceClient(conn)
    
    tests := []struct {
        name     string
        request  *LoginRequest
        expected *LoginResponse
    }{
        {
            name: "Valid login",
            request: &LoginRequest{
                Email:    "test@example.com",
                Password: "password123",
            },
            expected: &LoginResponse{
                Success: true,
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := client.Login(context.Background(), tt.request)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected.Success, resp.Success)
        })
    }
}
```

## Success Criteria

### Technical Metrics
- [ ] **Service Isolation:** Each service has independent deployment and scaling
- [ ] **Data Ownership:** Each service owns its database schema
- [ ] **API Compliance:** All gRPC contracts implemented and tested
- [ ] **Event Flow:** Asynchronous events working with <5s latency
- [ ] **Observability:** 100% service coverage for metrics, logs, and traces

### Business Metrics
- [ ] **Performance:** No degradation in API response times (<200ms p95)
- [ ] **Reliability:** 99.9% uptime maintained during migration
- [ ] **Scalability:** Services can scale independently under load
- [ ] **Development Velocity:** Team can deploy services independently
- [ ] **Operational Excellence:** Incidents resolved <30 minutes MTTR

## Risk Mitigation

### Technical Risks
1. **Data Consistency Issues**
   - Mitigation: Implement robust saga patterns with compensation
   - Testing: Chaos engineering to simulate failures

2. **Network Partitions**
   - Mitigation: Circuit breakers with fallback mechanisms
   - Testing: Network fault injection during testing

3. **Service Dependencies**
   - Mitigation: Async communication where possible
   - Testing: Service virtualization for dependencies

### Operational Risks
1. **Deployment Complexity** 
   - Mitigation: Blue-green deployments with automated rollback
   - Testing: Canary releases with monitoring

2. **Monitoring Gaps**
   - Mitigation: Comprehensive observability from day one
   - Testing: Synthetic monitoring and alerting validation

This service mapping guide provides the foundation for a successful microservices migration while maintaining system reliability and performance.
