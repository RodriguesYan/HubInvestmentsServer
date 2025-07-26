# Market Data gRPC Implementation

This directory contains the gRPC implementation for the Market Data service, following Domain-Driven Design (DDD) principles.

## Architecture Overview

The gRPC implementation follows the same DDD layered architecture as the HTTP REST implementation:

```
External Services (Watchlist, etc.)
         ↓ gRPC + JWT Auth
┌─────────────────────────┐
│   gRPC Server/Client    │ ← PRESENTATION LAYER
│   (Protocol Specific)  │   (Authentication Interceptor)
└─────────────────────────┘
         ↓
┌─────────────────────────┐
│    Use Case Layer       │ ← APPLICATION LAYER (SHARED)
│   (Business Logic)     │
└─────────────────────────┘
         ↓
┌─────────────────────────┐
│  Repository Layer       │ ← INFRASTRUCTURE LAYER
│  (Database Access)      │
└─────────────────────────┘
```

## Directory Structure

```
market_data/presentation/grpc/
├── README.md                           # This file
├── market_data.proto                   # Protocol Buffers definition
├── grpc_server.go                      # gRPC server startup functions
├── market_data_grpc_server.go          # gRPC server implementation
├── interceptors/                       # Authentication layer
│   └── auth_interceptor.go             # JWT authentication interceptor
├── proto/                              # Generated protobuf files
│   ├── market_data.pb.go               # Generated message types
│   └── market_data_grpc.pb.go          # Generated gRPC service code
└── client/                             # gRPC client implementation
    ├── market_data_grpc_client.go      # gRPC client
    └── market_data_grpc_client_test.go # Client unit tests
```

## 🔒 Authentication

### **JWT Authentication Required**
The gRPC service now includes **JWT authentication** using interceptors, matching the HTTP endpoints' security:

- **HTTP**: Uses middleware for JWT authentication
- **gRPC**: Uses interceptors for JWT authentication
- **Same tokens**: Both protocols accept the same JWT tokens

### **How Authentication Works**
1. **Interceptor**: gRPC requests pass through authentication interceptor
2. **Header Extraction**: JWT token extracted from `authorization` metadata
3. **Token Validation**: Same token service validates JWT as HTTP endpoints
4. **Context Injection**: User ID injected into request context
5. **Handler Execution**: Authenticated request proceeds to business logic

### **Authentication in Postman**
When making gRPC requests in Postman, you need to include JWT token in metadata:

```
Key: authorization
Value: Bearer YOUR_JWT_TOKEN_HERE
```

## Key Design Principles

### 1. **Shared Business Logic**
Both HTTP and gRPC handlers use the **same use cases** from the application layer:

```go
// Both HTTP and gRPC use the same use case
useCase := container.GetMarketDataUsecase()
result, err := useCase.Execute(symbols)
```

### 2. **Shared Authentication**
Both protocols use the **same authentication logic**:

```go
// HTTP: Middleware
func GetMarketDataWithAuth(verifyToken middleware.TokenVerifier, container di.Container)

// gRPC: Interceptor
func (interceptor *AuthInterceptor) UnaryInterceptor(...)
```

### 3. **Protocol Independence**
The repository layer is completely unaware of whether it's serving HTTP or gRPC requests. This ensures:
- Single source of truth for business logic
- Easier maintenance and testing
- Protocol-agnostic data access

### 4. **Clean Architecture**
- **Presentation Layer**: HTTP and gRPC handlers (protocol-specific)
- **Application Layer**: Use cases (shared business logic)
- **Domain Layer**: Models and repository interfaces
- **Infrastructure Layer**: Database repositories and external services

## Files Overview

### `market_data.proto`
Protocol Buffers definition file that defines:
- `MarketDataService` with `GetMarketData` RPC method
- Request/Response message types
- Streaming support (for future real-time data)

### `grpc_server.go`
gRPC server startup functions that:
- Configure and start the gRPC server with authentication
- Set up authentication interceptors
- Handle server lifecycle (start/stop)
- Support both synchronous and asynchronous startup

### `market_data_grpc_server.go`
gRPC server implementation that:
- Implements the `MarketDataService` interface
- Uses the same dependency injection container as HTTP handlers
- Shares business logic through use cases
- Provides proper error handling with gRPC status codes

### `interceptors/auth_interceptor.go`
Authentication interceptor that:
- Extracts JWT tokens from gRPC metadata
- Validates tokens using the same token service as HTTP
- Injects user ID into request context
- Returns proper gRPC authentication errors

### `client/market_data_grpc_client.go`
gRPC client implementation that:
- Provides a clean interface for calling the market data service
- Handles connection management and timeouts
- Converts between protobuf messages and domain models
- Includes proper error handling

## Usage Examples

### Server Usage

```go
// In main.go or server setup - now with authentication
func startGRPCServer(container di.Container) {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create authentication interceptor
    authInterceptor := interceptors.NewAuthInterceptor()
    
    // Create gRPC server with interceptors
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor),
        grpc.StreamInterceptor(authInterceptor.StreamInterceptor),
    )
    
    marketDataServer := grpc.NewMarketDataGRPCServer(container)
    proto.RegisterMarketDataServiceServer(grpcServer, marketDataServer)
    
    log.Fatal(grpcServer.Serve(lis))
}
```

### Client Usage with Authentication

```go
// Create client with default configuration
client, err := client.NewMarketDataGRPCClientWithDefaults()
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Create context with JWT token
ctx := context.Background()
ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+jwtToken)

// Make authenticated gRPC call
symbols := []string{"AAPL", "GOOGL", "MSFT"}
marketData, err := client.GetMarketData(ctx, symbols)
if err != nil {
    log.Printf("Error: %v", err)
    return
}

// Use the market data
for _, data := range marketData {
    fmt.Printf("Symbol: %s, Price: %.2f\n", data.Symbol, data.LastQuote)
}
```

### Client with Custom Configuration

```go
config := client.MarketDataGRPCClientConfig{
    ServerAddress: "market-data-service:50051",
    Timeout:       10 * time.Second,
}

client, err := client.NewMarketDataGRPCClient(config)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## 🔐 Authentication Flow

### **1. Get JWT Token (same as HTTP)**
```bash
curl -X POST http://192.168.0.6:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "your_username", "password": "your_password"}'
```

### **2. Use Token in gRPC Request**
In Postman metadata:
```
authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### **3. Authentication Errors**
- `UNAUTHENTICATED` (16): Missing or invalid token
- `PERMISSION_DENIED` (7): Valid token but insufficient permissions (future)

## Benefits of This Implementation

### 1. **Dual Protocol Support**
- **HTTP REST**: For external clients (web apps, mobile apps)
- **gRPC**: For internal service-to-service communication (higher performance)
- **Same Authentication**: Both use identical JWT tokens

### 2. **Shared Business Logic**
- No code duplication between protocols
- Single source of truth for business rules
- Changes to business logic affect both protocols automatically

### 3. **Security Consistency**
- Same JWT tokens work for both HTTP and gRPC
- Consistent authentication errors and handling
- Same user session across protocols

### 4. **Performance Optimization**
- gRPC uses HTTP/2 and binary serialization (faster than JSON)
- Connection pooling and multiplexing
- Efficient for high-throughput internal communication

### 5. **Type Safety**
- Protocol Buffers provide strong typing
- Generated code ensures API contract compliance
- Compile-time error detection

### 6. **Backward Compatibility**
- Existing HTTP endpoints continue to work unchanged
- gradual migration to gRPC for internal services
- Supports both protocols simultaneously

## Testing

The implementation includes comprehensive unit tests:

```bash
# Run gRPC client tests
go test ./market_data/presentation/grpc/client/

# Run all market data tests
go test ./market_data/...
```

## Generating Proto Files

When you modify `market_data.proto`, regenerate the Go code:

```bash
cd market_data/presentation/grpc
export PATH=$PATH:$(go env GOPATH)/bin
protoc --go_out=. --go-grpc_out=. market_data.proto
```

## Integration with Existing System

This gRPC implementation integrates seamlessly with the existing DDD architecture:

1. **Same Container**: Uses the existing dependency injection container
2. **Same Use Cases**: Reuses existing business logic
3. **Same Domain Models**: No additional model conversion needed
4. **Same Repository**: Database access remains unchanged
5. **Same Authentication**: Uses identical JWT token service

## Future Enhancements

1. **Streaming Support**: Real-time market data streaming
2. **Advanced Authentication**: User roles and permissions
3. **Load Balancing**: Service discovery and load balancing
4. **Metrics**: Prometheus metrics for gRPC endpoints
5. **Circuit Breaker**: Resilience patterns for gRPC calls 