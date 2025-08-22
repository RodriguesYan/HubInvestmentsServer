# HubInvestments System Architecture Diagrams

This document contains the Mermaid diagram code for the HubInvestments system architecture diagrams.

## Diagram 0: Simple Order Processing Flow

**Clean, linear flow showing the core order processing logic:**

```mermaid
graph LR
    %% Client Request
    Client[Asset Homebroker Page<br/>Web Client]:::client
    
    %% Order Submission
    Client -->|POST /api/orders/sendOrder| OrderAPI[Order Manager API<br/>HTTP Handler]:::api
    
    %% Validation
    OrderAPI --> Validation{Validate Order<br/>Is Valid?}:::decision
    Validation -->|Invalid| ErrorResponse[Return Error<br/>400 Bad Request]:::error
    Validation -->|Valid| ProcessOrder[Process Order<br/>Generate UUID<br/>Save to DB]:::process
    
    %% Success Response
    ProcessOrder --> SuccessResponse[Return 200 OK<br/>Order ID]:::success
    SuccessResponse --> Client
    
    %% Async Processing
    ProcessOrder --> RabbitMQ[Send to RabbitMQ<br/>orders.processing]:::queue
    RabbitMQ --> OrderConsumer[Order Consumer<br/>Background Worker]:::worker
    OrderConsumer --> SomeClearing[Send to<br/>Some Clearing]:::external
    SomeClearing --> Database[(Update Orders Table<br/>PostgreSQL)]:::database
    
    %% CSS Classes
    classDef client fill:#E3F2FD,stroke:#1976D2,stroke-width:2px,color:#000
    classDef api fill:#FFF3E0,stroke:#F57C00,stroke-width:2px,color:#000
    classDef decision fill:#FFF9C4,stroke:#F9A825,stroke-width:2px,color:#000
    classDef error fill:#FFEBEE,stroke:#D32F2F,stroke-width:2px,color:#000
    classDef success fill:#E8F5E8,stroke:#388E3C,stroke-width:2px,color:#000
    classDef process fill:#F3E5F5,stroke:#7B1FA2,stroke-width:2px,color:#000
    classDef queue fill:#E0F2F1,stroke:#00796B,stroke-width:2px,color:#000
    classDef worker fill:#FFF8E1,stroke:#FBC02D,stroke-width:2px,color:#000
    classDef external fill:#FCE4EC,stroke:#C2185B,stroke-width:2px,color:#000
    classDef database fill:#E1F5FE,stroke:#0277BD,stroke-width:2px,color:#000
```

## Diagram 0.1: Simple Microservices Overview

**Clean, organized view of the microservices architecture:**

```mermaid
graph TB
    %% Client Layer
    WebApp[Web Application<br/>React/Angular]:::client
    MobileApp[Mobile App<br/>iOS/Android]:::client
    
    %% API Gateway
    Gateway[API Gateway<br/>Kong/Nginx<br/>Port 443]:::gateway
    
    %% Core Services Row 1
    AuthSvc[🔐 Auth Service<br/>Port 8001]:::service
    LoginSvc[👤 Login Service<br/>Port 8002]:::service
    BalanceSvc[💰 Balance Service<br/>Port 8003]:::service
    
    %% Core Services Row 2
    PositionSvc[📊 Position Service<br/>Port 8004]:::service
    PortfolioSvc[📈 Portfolio Service<br/>Port 8005]:::service
    MarketSvc[📉 Market Data Service<br/>Port 8006]:::service
    
    %% Core Services Row 3
    WatchlistSvc[⭐ Watchlist Service<br/>Port 8007]:::service
    OrderSvc[🛒 Order Service<br/>Port 8008]:::service
    
    %% Infrastructure
    MessageQueue[RabbitMQ<br/>Message Broker]:::infra
    Cache[Redis<br/>Cache Cluster]:::infra
    Database[(PostgreSQL<br/>Database Cluster)]:::database
    
    %% External
    Broker[External Broker<br/>Order Execution]:::external
    MarketData[Market Data Provider<br/>Real-time Feeds]:::external
    
    %% Client Connections
    WebApp --> Gateway
    MobileApp --> Gateway
    
    %% Gateway to Services
    Gateway --> AuthSvc
    Gateway --> LoginSvc
    Gateway --> BalanceSvc
    Gateway --> PositionSvc
    Gateway --> PortfolioSvc
    Gateway --> MarketSvc
    Gateway --> WatchlistSvc
    Gateway --> OrderSvc
    
    %% Service Dependencies
    PortfolioSvc -.-> PositionSvc
    PortfolioSvc -.-> BalanceSvc
    OrderSvc -.-> MarketSvc
    WatchlistSvc -.-> MarketSvc
    
    %% Infrastructure Connections
    OrderSvc --> MessageQueue
    MarketSvc --> Cache
    AuthSvc --> Cache
    
    %% Database Connections
    AuthSvc --> Database
    LoginSvc --> Database
    BalanceSvc --> Database
    PositionSvc --> Database
    PortfolioSvc --> Database
    MarketSvc --> Database
    WatchlistSvc --> Database
    OrderSvc --> Database
    
    %% External Connections
    OrderSvc --> Broker
    MarketSvc --> MarketData
    
    %% CSS Classes
    classDef client fill:#E3F2FD,stroke:#1976D2,stroke-width:2px,color:#000
    classDef gateway fill:#FFF3E0,stroke:#FF9800,stroke-width:3px,color:#000
    classDef service fill:#4A90E2,stroke:#2E5C8A,stroke-width:2px,color:#fff
    classDef infra fill:#9C27B0,stroke:#7B1FA2,stroke-width:2px,color:#fff
    classDef database fill:#4CAF50,stroke:#388E3C,stroke-width:2px,color:#fff
    classDef external fill:#FF5722,stroke:#D84315,stroke-width:2px,color:#fff
```

## Diagram 1: Current Implementation Status (Color-Coded)

**Legend:**
- 🔵 Blue: Fully implemented and working components
- 🔴 Red: Not implemented yet (planned in TODO.md)
- 🟠 Orange: Partially implemented (domain logic exists but missing infrastructure)

```mermaid
graph TB
    %% External Clients - IMPLEMENTED
    WebClient[Web Client]:::implemented
    MobileClient[Mobile Client]:::implemented
    AdminClient[Admin Client]:::implemented
    
    %% Load Balancer & Gateway - NOT IMPLEMENTED
    LoadBalancer[Load Balancer/Gateway<br/>Nginx]:::notImplemented
    
    %% Main Application Server - IMPLEMENTED
    MainApp[HubInvestments Server<br/>main.go<br/>Port: 8080]:::implemented
    
    %% Authentication & Middleware - IMPLEMENTED
    AuthMiddleware[Authentication Middleware<br/>JWT Token Verification]:::implemented
    AuthService[Auth Service<br/>Token Generation & Validation]:::implemented
    
    %% HTTP Handlers Layer
    LoginHandler[Login Handler<br/>/login]:::implemented
    PositionHandler[Position Handler<br/>/getAucAggregation]:::implemented
    BalanceHandler[Balance Handler<br/>/getBalance]:::implemented
    PortfolioHandler[Portfolio Handler<br/>/getPortfolioSummary]:::implemented
    MarketDataHandler[Market Data Handler<br/>/getMarketData]:::implemented
    WatchlistHandler[Watchlist Handler<br/>/getWatchlist]:::implemented
    AdminHandler[Admin Handler<br/>/admin/market-data/cache/*]:::implemented
    OrderHandler[Order Handler<br/>/orders/*<br/>NOT IMPLEMENTED]:::notImplemented
    
    %% gRPC Server - IMPLEMENTED
    GRPCServer[gRPC Server<br/>Port: 50051<br/>Market Data Service]:::implemented
    GRPCClient[gRPC Client<br/>Order Management Integration]:::implemented
    
    %% Use Case Layer (Application Services)
    LoginUseCase[Login Use Case<br/>Authentication Logic]:::implemented
    PositionUseCase[Position Aggregation Use Case<br/>Asset Position Logic]:::implemented
    BalanceUseCase[Balance Use Case<br/>Account Balance Logic]:::implemented
    PortfolioUseCase[Portfolio Summary Use Case<br/>Combined Portfolio Logic]:::implemented
    MarketDataUseCase[Market Data Use Case<br/>Market Information Logic]:::implemented
    WatchlistUseCase[Watchlist Use Case<br/>User Watchlist Logic]:::implemented
    
    %% Order Management Use Cases - DOMAIN ONLY
    SubmitOrderUseCase[Submit Order Use Case<br/>Order Submission Logic<br/>DOMAIN ONLY]:::partiallyImplemented
    ProcessOrderUseCase[Process Order Use Case<br/>Async Order Processing<br/>DOMAIN ONLY]:::partiallyImplemented
    GetOrderStatusUseCase[Get Order Status Use Case<br/>Order Status Tracking<br/>DOMAIN ONLY]:::partiallyImplemented
    CancelOrderUseCase[Cancel Order Use Case<br/>Order Cancellation<br/>DOMAIN ONLY]:::partiallyImplemented
    
    %% Domain Services
    PositionAggregationService[Position Aggregation Service<br/>Business Logic]:::implemented
    OrderValidationService[Order Validation Service<br/>Business Rules]:::implemented
    RiskManagementService[Risk Management Service<br/>Risk Assessment]:::implemented
    OrderPricingService[Order Pricing Service<br/>Pricing Logic]:::implemented
    
    %% Repository Layer (Data Access)
    LoginRepo[Login Repository<br/>User Authentication Data]:::implemented
    PositionRepo[Position Repository<br/>Asset Position Data]:::implemented
    BalanceRepo[Balance Repository<br/>Account Balance Data]:::implemented
    MarketDataRepo[Market Data Repository<br/>Market Information Data]:::implemented
    WatchlistRepo[Watchlist Repository<br/>User Watchlist Data]:::implemented
    OrderRepo[Order Repository<br/>MOCK IMPLEMENTATION ONLY]:::notImplemented
    
    %% Cache Layer - IMPLEMENTED
    CacheManager[Cache Manager<br/>Redis Cache Operations]:::implemented
    MarketDataCache[Market Data Cache Repository<br/>Cache-Aside Pattern]:::implemented
    RedisCache[Redis Cache<br/>In-Memory Storage<br/>TTL: 5 minutes]:::implemented
    
    %% Message Queue Infrastructure
    RabbitMQProducer[RabbitMQ Producer<br/>Order Message Publishing]:::implemented
    RabbitMQConsumer[RabbitMQ Consumer<br/>Order Message Processing]:::implemented
    RabbitMQ[RabbitMQ Message Broker<br/>Queues:<br/>- orders.submit<br/>- orders.processing<br/>- orders.settlement<br/>- orders.status<br/>- orders.dlq<br/>- orders.retry]:::implemented
    OrderWorker[Order Worker<br/>Background Processing<br/>NOT IMPLEMENTED]:::notImplemented
    
    %% Database Layer
    PostgreSQL[(PostgreSQL Database<br/>Primary Data Store)]:::implemented
    UsersTable[(users table<br/>Authentication Data)]:::implemented
    BalancesTable[(balances table<br/>Account Balances)]:::implemented
    MarketDataTable[(market_data table<br/>Market Information)]:::implemented
    PositionsTable[(positions table<br/>Asset Positions)]:::implemented
    WatchlistTable[(watchlist tables<br/>User Watchlists)]:::implemented
    OrdersTable[(orders table<br/>NOT IMPLEMENTED)]:::notImplemented
    
    %% External Services - NOT IMPLEMENTED
    ExternalBroker[External Broker API<br/>Order Execution<br/>NOT IMPLEMENTED]:::notImplemented
    MarketDataProvider[Market Data Provider<br/>Real-time Data<br/>NOT IMPLEMENTED]:::notImplemented
    WebSocketServer[WebSocket Server<br/>Real-time Updates<br/>NOT IMPLEMENTED]:::notImplemented
    
    %% Dependency Injection Container - IMPLEMENTED
    DIContainer[Dependency Injection Container<br/>Service Orchestration<br/>Lifecycle Management]:::implemented
    
    %% Swagger Documentation - IMPLEMENTED
    SwaggerUI[Swagger UI<br/>/swagger/index.html<br/>API Documentation]:::implemented
    
    %% User Registration - NOT IMPLEMENTED
    UserRegistration[User Registration<br/>Account Creation<br/>NOT IMPLEMENTED]:::notImplemented
    
    %% Security Features - NOT IMPLEMENTED
    SecurityHeaders[Security Headers<br/>CSRF, XSS Protection<br/>NOT IMPLEMENTED]:::notImplemented
    RateLimiting[Rate Limiting<br/>DDoS Protection<br/>NOT IMPLEMENTED]:::notImplemented
    
    %% Monitoring & Logging - NOT IMPLEMENTED
    Monitoring[Monitoring & Metrics<br/>Prometheus/Grafana<br/>NOT IMPLEMENTED]:::notImplemented
    StructuredLogging[Structured Logging<br/>JSON Logs<br/>NOT IMPLEMENTED]:::notImplemented
    
    %% Client Connections
    WebClient --> LoadBalancer
    MobileClient --> LoadBalancer
    AdminClient --> LoadBalancer
    LoadBalancer --> MainApp
    
    %% Main Application Flow
    MainApp --> AuthMiddleware
    AuthMiddleware --> AuthService
    
    %% HTTP Handler Routing
    MainApp --> LoginHandler
    MainApp --> PositionHandler
    MainApp --> BalanceHandler
    MainApp --> PortfolioHandler
    MainApp --> MarketDataHandler
    MainApp --> WatchlistHandler
    MainApp --> AdminHandler
    MainApp --> OrderHandler
    MainApp --> SwaggerUI
    
    %% gRPC Integration
    MainApp --> GRPCServer
    GRPCClient --> GRPCServer
    
    %% Use Case Dependencies
    LoginHandler --> LoginUseCase
    PositionHandler --> PositionUseCase
    BalanceHandler --> BalanceUseCase
    PortfolioHandler --> PortfolioUseCase
    MarketDataHandler --> MarketDataUseCase
    WatchlistHandler --> WatchlistUseCase
    AdminHandler --> CacheManager
    OrderHandler --> SubmitOrderUseCase
    OrderHandler --> GetOrderStatusUseCase
    OrderHandler --> CancelOrderUseCase
    
    %% gRPC Use Case Integration
    GRPCServer --> MarketDataUseCase
    SubmitOrderUseCase --> GRPCClient
    ProcessOrderUseCase --> GRPCClient
    
    %% Domain Service Dependencies
    PositionUseCase --> PositionAggregationService
    SubmitOrderUseCase --> OrderValidationService
    SubmitOrderUseCase --> RiskManagementService
    ProcessOrderUseCase --> OrderPricingService
    
    %% Repository Dependencies
    LoginUseCase --> LoginRepo
    PositionUseCase --> PositionRepo
    BalanceUseCase --> BalanceRepo
    MarketDataUseCase --> MarketDataCache
    WatchlistUseCase --> WatchlistRepo
    SubmitOrderUseCase --> OrderRepo
    ProcessOrderUseCase --> OrderRepo
    GetOrderStatusUseCase --> OrderRepo
    CancelOrderUseCase --> OrderRepo
    
    %% Cache Layer Integration
    MarketDataCache --> MarketDataRepo
    MarketDataCache --> RedisCache
    CacheManager --> RedisCache
    
    %% Message Queue Integration
    SubmitOrderUseCase --> RabbitMQProducer
    RabbitMQProducer --> RabbitMQ
    RabbitMQ --> RabbitMQConsumer
    RabbitMQConsumer --> OrderWorker
    OrderWorker --> ProcessOrderUseCase
    
    %% Database Connections
    LoginRepo --> PostgreSQL
    PositionRepo --> PostgreSQL
    BalanceRepo --> PostgreSQL
    MarketDataRepo --> PostgreSQL
    WatchlistRepo --> PostgreSQL
    OrderRepo --> PostgreSQL
    
    %% Database Tables
    PostgreSQL --> UsersTable
    PostgreSQL --> BalancesTable
    PostgreSQL --> MarketDataTable
    PostgreSQL --> PositionsTable
    PostgreSQL --> WatchlistTable
    PostgreSQL --> OrdersTable
    
    %% External Service Integration
    ProcessOrderUseCase --> ExternalBroker
    MarketDataUseCase --> MarketDataProvider
    MainApp --> WebSocketServer
    
    %% Security Integration
    MainApp --> SecurityHeaders
    MainApp --> RateLimiting
    
    %% Monitoring Integration
    MainApp --> Monitoring
    MainApp --> StructuredLogging
    
    %% User Registration Integration
    MainApp --> UserRegistration
    
    %% Dependency Injection
    DIContainer --> LoginUseCase
    DIContainer --> PositionUseCase
    DIContainer --> BalanceUseCase
    DIContainer --> PortfolioUseCase
    DIContainer --> MarketDataUseCase
    DIContainer --> WatchlistUseCase
    DIContainer --> SubmitOrderUseCase
    DIContainer --> ProcessOrderUseCase
    DIContainer --> GetOrderStatusUseCase
    DIContainer --> CancelOrderUseCase
    DIContainer --> AuthService
    DIContainer --> CacheManager
    DIContainer --> RabbitMQProducer
    DIContainer --> GRPCClient
    
    %% Portfolio Use Case Composition
    PortfolioUseCase --> PositionUseCase
    PortfolioUseCase --> BalanceUseCase
    
    %% CSS Classes for Color Coding
    classDef implemented fill:#4A90E2,stroke:#2E5C8A,stroke-width:2px,color:#fff
    classDef notImplemented fill:#E74C3C,stroke:#C0392B,stroke-width:2px,color:#fff
    classDef partiallyImplemented fill:#F39C12,stroke:#D68910,stroke-width:2px,color:#fff
```

## Diagram 2: Future Microservices Architecture

**Legend:**
- 🔵 Blue: Services that can be extracted from current implementation
- 🟠 Orange: Services that need completion before extraction
- 🟣 Purple: Infrastructure components
- 🟢 Green: Database services
- 🔴 Red: External services
- 📊 Gray: Monitoring and observability
- 🛡️ Red: Security services

```mermaid
graph TB
    %% External Clients
    WebClient[Web Client]:::client
    MobileClient[Mobile Client]:::client
    AdminClient[Admin Client]:::client
    
    %% API Gateway & Load Balancer
    APIGateway[API Gateway<br/>Kong/Nginx<br/>Port: 443/80<br/>Authentication<br/>Rate Limiting<br/>Load Balancing<br/>SSL Termination]:::gateway
    
    %% Service Discovery & Configuration
    ServiceDiscovery[Service Discovery<br/>Consul/Eureka<br/>Service Registration<br/>Health Checks<br/>Configuration Management]:::infrastructure
    
    %% Authentication Microservice
    AuthService[🔐 Authentication Service<br/>Port: 8001<br/>JWT Token Management<br/>User Authentication<br/>Token Validation<br/>Session Management]:::implemented
    
    %% Login Microservice  
    LoginService[👤 Login Service<br/>Port: 8002<br/>User Login Logic<br/>Password Validation<br/>Account Verification<br/>Login Audit]:::implemented
    
    %% Balance Microservice
    BalanceService[💰 Balance Service<br/>Port: 8003<br/>Account Balance Management<br/>Balance Calculations<br/>Transaction History<br/>Balance Validation]:::implemented
    
    %% Position Microservice
    PositionService[📊 Position Service<br/>Port: 8004<br/>Asset Position Tracking<br/>Position Aggregation<br/>Portfolio Calculations<br/>Asset Management]:::implemented
    
    %% Portfolio Summary Microservice
    PortfolioService[📈 Portfolio Service<br/>Port: 8005<br/>Portfolio Aggregation<br/>Cross-Service Orchestration<br/>Portfolio Analytics<br/>Performance Metrics]:::implemented
    
    %% Market Data Microservice
    MarketDataService[📉 Market Data Service<br/>Port: 8006 HTTP<br/>Port: 50051 gRPC<br/>Real-time Market Data<br/>Price Information<br/>Market Analytics<br/>Data Caching]:::implemented
    
    %% Watchlist Microservice
    WatchlistService[⭐ Watchlist Service<br/>Port: 8007<br/>User Watchlists<br/>Symbol Tracking<br/>Watchlist Management<br/>Notifications]:::implemented
    
    %% Order Management Microservice
    OrderService[🛒 Order Management Service<br/>Port: 8008<br/>Order Processing<br/>Order Validation<br/>Risk Management<br/>Order Execution<br/>Order History]:::partiallyImplemented
    
    %% Message Queue Cluster
    MessageBroker[📨 Message Broker Cluster<br/>RabbitMQ/Apache Kafka<br/>Order Processing Queues<br/>Event Streaming<br/>Dead Letter Queues<br/>Message Persistence]:::infrastructure
    
    %% Cache Cluster
    CacheCluster[⚡ Cache Cluster<br/>Redis Cluster<br/>Distributed Caching<br/>Session Storage<br/>Market Data Cache<br/>High Availability]:::infrastructure
    
    %% Database Cluster
    DatabaseCluster[🗄️ Database Cluster<br/>PostgreSQL Cluster<br/>Master-Slave Replication<br/>Connection Pooling<br/>Backup & Recovery<br/>Data Partitioning]:::infrastructure
    
    %% Individual Service Databases
    AuthDB[(Auth Database<br/>Users & Sessions)]:::database
    LoginDB[(Login Database<br/>Authentication Data)]:::database
    BalanceDB[(Balance Database<br/>Account Balances)]:::database
    PositionDB[(Position Database<br/>Asset Positions)]:::database
    MarketDataDB[(Market Data Database<br/>Price & Market Info)]:::database
    WatchlistDB[(Watchlist Database<br/>User Watchlists)]:::database
    OrderDB[(Order Database<br/>Orders & Transactions)]:::database
    
    %% External Services
    ExternalBroker[🏦 External Broker API<br/>Order Execution<br/>Settlement<br/>Clearing]:::external
    MarketDataProvider[📊 Market Data Provider<br/>Real-time Feeds<br/>Historical Data<br/>Market Events]:::external
    
    %% Monitoring & Observability
    MonitoringStack[📊 Monitoring Stack<br/>Prometheus + Grafana<br/>Metrics Collection<br/>Alerting<br/>Dashboards<br/>Performance Monitoring]:::monitoring
    
    LoggingStack[📝 Logging Stack<br/>ELK Stack<br/>Centralized Logging<br/>Log Aggregation<br/>Search & Analytics<br/>Audit Trails]:::monitoring
    
    TracingSystem[🔍 Distributed Tracing<br/>Jaeger/Zipkin<br/>Request Tracing<br/>Performance Analysis<br/>Dependency Mapping<br/>Bottleneck Detection]:::monitoring
    
    %% Security Services
    SecurityService[🛡️ Security Service<br/>WAF Protection<br/>DDoS Mitigation<br/>Security Headers<br/>Threat Detection]:::security
    
    %% Client Connections
    WebClient --> APIGateway
    MobileClient --> APIGateway
    AdminClient --> APIGateway
    
    %% API Gateway Routing
    APIGateway --> AuthService
    APIGateway --> LoginService
    APIGateway --> BalanceService
    APIGateway --> PositionService
    APIGateway --> PortfolioService
    APIGateway --> MarketDataService
    APIGateway --> WatchlistService
    APIGateway --> OrderService
    
    %% Service Discovery Integration
    ServiceDiscovery --> AuthService
    ServiceDiscovery --> LoginService
    ServiceDiscovery --> BalanceService
    ServiceDiscovery --> PositionService
    ServiceDiscovery --> PortfolioService
    ServiceDiscovery --> MarketDataService
    ServiceDiscovery --> WatchlistService
    ServiceDiscovery --> OrderService
    
    %% Inter-Service Communication gRPC/HTTP
    PortfolioService -.->|gRPC| PositionService
    PortfolioService -.->|gRPC| BalanceService
    WatchlistService -.->|gRPC| MarketDataService
    OrderService -.->|gRPC| MarketDataService
    OrderService -.->|gRPC| BalanceService
    OrderService -.->|gRPC| PositionService
    
    %% Authentication Flow
    APIGateway -.->|Token Validation| AuthService
    LoginService -.->|Token Generation| AuthService
    
    %% Message Queue Integration
    OrderService --> MessageBroker
    MessageBroker --> OrderService
    
    %% Cache Integration
    MarketDataService --> CacheCluster
    AuthService --> CacheCluster
    PortfolioService --> CacheCluster
    
    %% Database Connections
    AuthService --> DatabaseCluster
    LoginService --> DatabaseCluster
    BalanceService --> DatabaseCluster
    PositionService --> DatabaseCluster
    MarketDataService --> DatabaseCluster
    WatchlistService --> DatabaseCluster
    OrderService --> DatabaseCluster
    
    %% Individual Database Mapping
    DatabaseCluster --> AuthDB
    DatabaseCluster --> LoginDB
    DatabaseCluster --> BalanceDB
    DatabaseCluster --> PositionDB
    DatabaseCluster --> MarketDataDB
    DatabaseCluster --> WatchlistDB
    DatabaseCluster --> OrderDB
    
    %% External Service Integration
    OrderService --> ExternalBroker
    MarketDataService --> MarketDataProvider
    
    %% Security Integration
    APIGateway --> SecurityService
    
    %% Monitoring Integration
    AuthService --> MonitoringStack
    LoginService --> MonitoringStack
    BalanceService --> MonitoringStack
    PositionService --> MonitoringStack
    PortfolioService --> MonitoringStack
    MarketDataService --> MonitoringStack
    WatchlistService --> MonitoringStack
    OrderService --> MonitoringStack
    
    %% Logging Integration
    AuthService --> LoggingStack
    LoginService --> LoggingStack
    BalanceService --> LoggingStack
    PositionService --> LoggingStack
    PortfolioService --> LoggingStack
    MarketDataService --> LoggingStack
    WatchlistService --> LoggingStack
    OrderService --> LoggingStack
    
    %% Tracing Integration
    AuthService --> TracingSystem
    LoginService --> TracingSystem
    BalanceService --> TracingSystem
    PositionService --> TracingSystem
    PortfolioService --> TracingSystem
    MarketDataService --> TracingSystem
    WatchlistService --> TracingSystem
    OrderService --> TracingSystem
    
    %% CSS Classes for Microservices Architecture
    classDef client fill:#E8F4FD,stroke:#1E88E5,stroke-width:2px,color:#000
    classDef gateway fill:#FFF3E0,stroke:#FF9800,stroke-width:3px,color:#000
    classDef implemented fill:#4A90E2,stroke:#2E5C8A,stroke-width:2px,color:#fff
    classDef partiallyImplemented fill:#F39C12,stroke:#D68910,stroke-width:2px,color:#fff
    classDef infrastructure fill:#9C27B0,stroke:#7B1FA2,stroke-width:2px,color:#fff
    classDef database fill:#4CAF50,stroke:#388E3C,stroke-width:2px,color:#fff
    classDef external fill:#FF5722,stroke:#D84315,stroke-width:2px,color:#fff
    classDef monitoring fill:#607D8B,stroke:#455A64,stroke-width:2px,color:#fff
    classDef security fill:#F44336,stroke:#C62828,stroke-width:2px,color:#fff
```

## How to Convert to Images

To convert these Mermaid diagrams to PNG/JPEG images, you can use:

1. **Online Tools:**
   - [Mermaid Live Editor](https://mermaid.live/) - Paste the code and export as PNG/SVG
   - [Mermaid Ink](https://mermaid.ink/) - URL-based image generation

2. **CLI Tools:**
   ```bash
   npm install -g @mermaid-js/mermaid-cli
   mmdc -i diagram.mmd -o diagram.png
   ```

3. **VS Code Extensions:**
   - Mermaid Preview extension
   - Export functionality available

4. **GitHub/GitLab:**
   - Both platforms render Mermaid diagrams natively in markdown files

## Architecture Migration Strategy

### Phase 1: Extract Independent Services
1. **Market Data Service** - Already has gRPC interface
2. **Authentication Service** - Stateless, easy to extract
3. **Balance Service** - Simple CRUD operations
4. **Position Service** - Independent business logic

### Phase 2: Extract Orchestration Services
1. **Portfolio Service** - Depends on Position and Balance services
2. **Watchlist Service** - Depends on Market Data service

### Phase 3: Complete and Extract Complex Services
1. **Order Management Service** - Complete implementation first
2. **Login Service** - Extract after user registration is implemented

### Phase 4: Add Infrastructure
1. **API Gateway** - Kong or Nginx
2. **Service Discovery** - Consul or Eureka
3. **Monitoring Stack** - Prometheus + Grafana
4. **Logging Stack** - ELK Stack
5. **Distributed Tracing** - Jaeger or Zipkin
