# WebSocket Architecture for Microservices

## Overview

This document explains how WebSocket communication works between the frontend, API Gateway, and microservices in our architecture.

## Architecture Decision

**✅ CHOSEN APPROACH: WebSocket Proxy Pattern**

The API Gateway acts as a **WebSocket-to-gRPC proxy**, translating between WebSocket (frontend) and gRPC Streaming (microservices).

```
┌──────────┐         ┌─────────────┐         ┌──────────────────┐
│ Frontend │◄───────►│ API Gateway │◄───────►│ Market Data      │
│          │ WebSocket│             │ gRPC    │ Service          │
│ (React)  │  JSON   │ (Go)        │ Stream  │ (Go)             │
└──────────┘         └─────────────┘         └──────────────────┘
                            │                         │
                            │                         ▼
                            │                  ┌─────────────┐
                            │                  │ Redis       │
                            │                  │ Pub/Sub     │
                            │                  └─────────────┘
                            ▼
                     ┌─────────────┐
                     │ Auth        │
                     │ Rate Limit  │
                     │ Metrics     │
                     └─────────────┘
```

## Why This Approach?

### ✅ Benefits

1. **Centralized Security**: API Gateway handles authentication, authorization, rate limiting
2. **Protocol Translation**: Frontend uses WebSocket (simple), microservices use gRPC (efficient)
3. **Load Balancing**: API Gateway can distribute connections across multiple microservice instances
4. **Observability**: Centralized metrics, tracing, and logging
5. **Backward Compatibility**: Can change microservice implementation without affecting frontend
6. **Service Isolation**: Microservices don't need to handle WebSocket complexity

### ❌ Alternative Approach (NOT Chosen)

**Direct WebSocket from Microservice:**
```
Frontend → API Gateway (HTTP Proxy) → Microservice (WebSocket Server)
```

**Why NOT:**
- ❌ Breaks API Gateway pattern (no centralized auth/rate limiting)
- ❌ Requires exposing microservice ports
- ❌ Harder to implement load balancing
- ❌ Loses API Gateway benefits (metrics, tracing, auth)
- ❌ Microservices must handle WebSocket complexity

## Implementation Details

### 1. Frontend → API Gateway (WebSocket)

**Protocol**: WebSocket with JSON messages

**Example Messages:**

```json
// Subscribe to quotes
{
  "type": "subscribe",
  "symbols": ["AAPL", "GOOGL", "MSFT"]
}

// Unsubscribe from quotes
{
  "type": "unsubscribe",
  "symbols": ["AAPL"]
}

// Receive quote update
{
  "type": "quote",
  "symbol": "AAPL",
  "price": 150.25,
  "timestamp": "2025-10-27T10:30:00Z"
}
```

**API Gateway WebSocket Handler:**

```go
// hub-api-gateway/internal/websocket/market_data_handler.go
package websocket

import (
    "context"
    "encoding/json"
    "log"
    
    "github.com/gorilla/websocket"
    pb "github.com/RodriguesYan/hub-proto-contracts/monolith"
)

type MarketDataWebSocketHandler struct {
    marketDataClient pb.MarketDataServiceClient
    upgrader         websocket.Upgrader
}

func (h *MarketDataWebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
    // 1. Upgrade HTTP to WebSocket
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v", err)
        return
    }
    defer conn.Close()
    
    // 2. Authenticate user from JWT token
    token := r.URL.Query().Get("token")
    userID, err := h.authenticateToken(token)
    if err != nil {
        conn.WriteJSON(map[string]string{"error": "unauthorized"})
        return
    }
    
    // 3. Open gRPC stream to Market Data Service
    ctx := context.Background()
    stream, err := h.marketDataClient.StreamQuotes(ctx)
    if err != nil {
        log.Printf("Failed to open gRPC stream: %v", err)
        return
    }
    defer stream.CloseSend()
    
    // 4. Forward WebSocket messages → gRPC stream
    go h.forwardWebSocketToGRPC(conn, stream)
    
    // 5. Forward gRPC stream → WebSocket messages
    h.forwardGRPCToWebSocket(stream, conn)
}

func (h *MarketDataWebSocketHandler) forwardWebSocketToGRPC(
    conn *websocket.Conn,
    stream pb.MarketDataService_StreamQuotesClient,
) {
    for {
        var msg struct {
            Type    string   `json:"type"`
            Symbols []string `json:"symbols"`
        }
        
        if err := conn.ReadJSON(&msg); err != nil {
            log.Printf("WebSocket read error: %v", err)
            return
        }
        
        // Translate WebSocket message to gRPC request
        grpcReq := &pb.StreamQuotesRequest{
            Action:  msg.Type, // "subscribe" or "unsubscribe"
            Symbols: msg.Symbols,
        }
        
        if err := stream.Send(grpcReq); err != nil {
            log.Printf("gRPC send error: %v", err)
            return
        }
    }
}

func (h *MarketDataWebSocketHandler) forwardGRPCToWebSocket(
    stream pb.MarketDataService_StreamQuotesClient,
    conn *websocket.Conn,
) {
    for {
        quote, err := stream.Recv()
        if err != nil {
            log.Printf("gRPC receive error: %v", err)
            return
        }
        
        // Translate gRPC response to WebSocket message
        wsMsg := map[string]interface{}{
            "type":      "quote",
            "symbol":    quote.Symbol,
            "price":     quote.Price,
            "timestamp": quote.Timestamp,
        }
        
        if err := conn.WriteJSON(wsMsg); err != nil {
            log.Printf("WebSocket write error: %v", err)
            return
        }
    }
}
```

### 2. API Gateway → Market Data Service (gRPC Streaming)

**Protocol**: gRPC Bidirectional Streaming

**Proto Definition:**

```protobuf
// hub-proto-contracts/monolith/market_data_service.proto
service MarketDataService {
    // Existing unary RPCs
    rpc GetMarketData(GetMarketDataRequest) returns (GetMarketDataResponse);
    rpc GetBatchMarketData(GetBatchMarketDataRequest) returns (GetBatchMarketDataResponse);
    
    // NEW: Bidirectional streaming for real-time quotes
    rpc StreamQuotes(stream StreamQuotesRequest) returns (stream StreamQuotesResponse);
}

message StreamQuotesRequest {
    string action = 1;           // "subscribe" or "unsubscribe"
    repeated string symbols = 2; // ["AAPL", "GOOGL", "MSFT"]
}

message StreamQuotesResponse {
    string symbol = 1;
    double price = 2;
    double change = 3;
    double change_percent = 4;
    int64 timestamp = 5;
    string type = 6; // "quote" or "error"
}
```

**Market Data Service gRPC Server:**

```go
// hub-market-data-service/internal/presentation/grpc/market_data_grpc_server.go
func (s *MarketDataGRPCServer) StreamQuotes(stream pb.MarketDataService_StreamQuotesServer) error {
    ctx := stream.Context()
    
    // Create Redis Pub/Sub subscriber
    pubsub := s.redisClient.Subscribe(ctx)
    defer pubsub.Close()
    
    subscribedSymbols := make(map[string]bool)
    
    // Handle incoming subscription requests
    go func() {
        for {
            req, err := stream.Recv()
            if err != nil {
                log.Printf("Stream receive error: %v", err)
                return
            }
            
            switch req.Action {
            case "subscribe":
                for _, symbol := range req.Symbols {
                    if !subscribedSymbols[symbol] {
                        pubsub.Subscribe(ctx, fmt.Sprintf("quote:%s", symbol))
                        subscribedSymbols[symbol] = true
                    }
                }
            case "unsubscribe":
                for _, symbol := range req.Symbols {
                    if subscribedSymbols[symbol] {
                        pubsub.Unsubscribe(ctx, fmt.Sprintf("quote:%s", symbol))
                        delete(subscribedSymbols, symbol)
                    }
                }
            }
        }
    }()
    
    // Forward Redis Pub/Sub messages to gRPC stream
    ch := pubsub.Channel()
    for msg := range ch {
        var quote Quote
        json.Unmarshal([]byte(msg.Payload), &quote)
        
        resp := &pb.StreamQuotesResponse{
            Symbol:        quote.Symbol,
            Price:         quote.Price,
            Change:        quote.Change,
            ChangePercent: quote.ChangePercent,
            Timestamp:     quote.Timestamp,
            Type:          "quote",
        }
        
        if err := stream.Send(resp); err != nil {
            log.Printf("Stream send error: %v", err)
            return err
        }
    }
    
    return nil
}
```

### 3. Market Data Service → Redis Pub/Sub

**Protocol**: Redis Pub/Sub

**Quote Publisher (Background Worker):**

```go
// hub-market-data-service/internal/infrastructure/worker/quote_publisher.go
type QuotePublisher struct {
    redisClient *redis.Client
}

func (p *QuotePublisher) PublishQuote(ctx context.Context, quote Quote) error {
    data, err := json.Marshal(quote)
    if err != nil {
        return err
    }
    
    channel := fmt.Sprintf("quote:%s", quote.Symbol)
    return p.redisClient.Publish(ctx, channel, data).Err()
}

// Background worker that fetches quotes from external API and publishes
func (p *QuotePublisher) Start(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            quotes := p.fetchQuotesFromExternalAPI()
            for _, quote := range quotes {
                p.PublishQuote(ctx, quote)
            }
        case <-ctx.Done():
            return
        }
    }
}
```

## Connection Lifecycle

### 1. Connection Establishment

```
Frontend                API Gateway              Market Data Service       Redis
   |                         |                            |                  |
   |--- WebSocket Connect -->|                            |                  |
   |                         |--- Authenticate Token ---->|                  |
   |                         |<-- Auth Success -----------|                  |
   |                         |                            |                  |
   |                         |--- Open gRPC Stream ------>|                  |
   |                         |<-- Stream Ready -----------|                  |
   |<-- Connection Ready ----|                            |                  |
```

### 2. Subscribe to Symbols

```
Frontend                API Gateway              Market Data Service       Redis
   |                         |                            |                  |
   |--- Subscribe AAPL ----->|                            |                  |
   |                         |--- gRPC Subscribe -------->|                  |
   |                         |                            |--- SUBSCRIBE ---->|
   |                         |                            |                  quote:AAPL
   |                         |<-- Subscribed -------------|                  |
   |<-- Subscribed ----------|                            |                  |
```

### 3. Receive Quote Updates

```
Frontend                API Gateway              Market Data Service       Redis
   |                         |                            |                  |
   |                         |                            |<-- PUBLISH ------| quote:AAPL
   |                         |                            |    {"price":150} |
   |                         |<-- gRPC Stream Msg --------|                  |
   |<-- WebSocket Msg -------|                            |                  |
   |   {"price": 150}        |                            |                  |
```

### 4. Unsubscribe

```
Frontend                API Gateway              Market Data Service       Redis
   |                         |                            |                  |
   |--- Unsubscribe AAPL --->|                            |                  |
   |                         |--- gRPC Unsubscribe ------>|                  |
   |                         |                            |--- UNSUBSCRIBE -->|
   |                         |                            |                  quote:AAPL
   |                         |<-- Unsubscribed -----------|                  |
   |<-- Unsubscribed --------|                            |                  |
```

### 5. Connection Teardown

```
Frontend                API Gateway              Market Data Service       Redis
   |                         |                            |                  |
   |--- Close WebSocket ---->|                            |                  |
   |                         |--- Close gRPC Stream ----->|                  |
   |                         |                            |--- UNSUBSCRIBE -->|
   |                         |                            |                  ALL
   |                         |<-- Stream Closed ----------|                  |
   |<-- Connection Closed ---|                            |                  |
```

## Error Handling

### 1. WebSocket Connection Lost

```go
// API Gateway: Detect WebSocket disconnection
func (h *MarketDataWebSocketHandler) forwardGRPCToWebSocket(...) {
    for {
        quote, err := stream.Recv()
        if err != nil {
            return
        }
        
        if err := conn.WriteJSON(wsMsg); err != nil {
            // WebSocket connection lost
            stream.CloseSend() // Close gRPC stream
            return
        }
    }
}
```

### 2. gRPC Stream Failure

```go
// API Gateway: Detect gRPC stream failure
func (h *MarketDataWebSocketHandler) HandleConnection(...) {
    stream, err := h.marketDataClient.StreamQuotes(ctx)
    if err != nil {
        // gRPC stream failed to open
        conn.WriteJSON(map[string]string{
            "error": "service_unavailable",
        })
        return
    }
    
    // Monitor stream health
    go func() {
        <-stream.Context().Done()
        // gRPC stream closed
        conn.WriteJSON(map[string]string{
            "error": "connection_lost",
        })
        conn.Close()
    }()
}
```

### 3. Redis Pub/Sub Failure

```go
// Market Data Service: Detect Redis failure
func (s *MarketDataGRPCServer) StreamQuotes(stream pb.MarketDataService_StreamQuotesServer) error {
    pubsub := s.redisClient.Subscribe(ctx)
    defer pubsub.Close()
    
    ch := pubsub.Channel()
    for {
        select {
        case msg := <-ch:
            // Forward message
            stream.Send(...)
        case <-time.After(30 * time.Second):
            // No messages for 30s, send heartbeat
            stream.Send(&pb.StreamQuotesResponse{
                Type: "heartbeat",
            })
        case <-ctx.Done():
            // Context cancelled
            return ctx.Err()
        }
    }
}
```

## Performance Considerations

### 1. Connection Pooling

- **API Gateway**: Maintain gRPC connection pool to Market Data Service
- **Market Data Service**: Maintain Redis connection pool

### 2. Load Balancing

- **Multiple API Gateway instances**: Load balancer distributes WebSocket connections
- **Multiple Market Data Service instances**: gRPC load balancing (round-robin)

### 3. Horizontal Scaling

```
┌──────────┐         ┌─────────────┐         ┌──────────────────┐
│ Frontend │◄───────►│ API Gateway │◄───────►│ Market Data      │
│          │         │ Instance 1  │         │ Service Inst 1   │
└──────────┘         └─────────────┘         └──────────────────┘
                            │                         │
┌──────────┐         ┌─────────────┐                 │
│ Frontend │◄───────►│ API Gateway │                 │
│          │         │ Instance 2  │                 │
└──────────┘         └─────────────┘                 │
                            │                         │
                            │                  ┌──────────────────┐
                            │                  │ Market Data      │
                            └─────────────────►│ Service Inst 2   │
                                               └──────────────────┘
                                                       │
                                                ┌─────────────┐
                                                │ Redis       │
                                                │ Pub/Sub     │
                                                │ (Shared)    │
                                                └─────────────┘
```

**Key Points:**
- Redis Pub/Sub is **shared** across all Market Data Service instances
- Each API Gateway instance can connect to any Market Data Service instance
- Load balancer distributes WebSocket connections across API Gateway instances

## Security

### 1. Authentication

```go
// API Gateway: Authenticate WebSocket connection
func (h *MarketDataWebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
    // Extract JWT token from query param or header
    token := r.URL.Query().Get("token")
    if token == "" {
        token = r.Header.Get("Authorization")
    }
    
    // Validate JWT token
    claims, err := h.jwtService.ValidateToken(token)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Add user context to gRPC metadata
    ctx := metadata.AppendToOutgoingContext(
        context.Background(),
        "user_id", claims.UserID,
        "authorization", token,
    )
    
    stream, err := h.marketDataClient.StreamQuotes(ctx)
    // ...
}
```

### 2. Rate Limiting

```go
// API Gateway: Rate limit WebSocket connections per user
type RateLimiter struct {
    connections map[string]int // userID → connection count
    mu          sync.RWMutex
}

func (rl *RateLimiter) AllowConnection(userID string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    if rl.connections[userID] >= 5 { // Max 5 connections per user
        return false
    }
    
    rl.connections[userID]++
    return true
}
```

### 3. Authorization

```go
// Market Data Service: Authorize symbol access
func (s *MarketDataGRPCServer) StreamQuotes(stream pb.MarketDataService_StreamQuotesServer) error {
    // Extract user from gRPC metadata
    md, _ := metadata.FromIncomingContext(stream.Context())
    userID := md.Get("user_id")[0]
    
    // Check user subscription tier
    tier := s.getUserSubscriptionTier(userID)
    
    for {
        req, _ := stream.Recv()
        
        // Authorize symbol access based on tier
        for _, symbol := range req.Symbols {
            if !s.isSymbolAllowedForTier(symbol, tier) {
                stream.Send(&pb.StreamQuotesResponse{
                    Type:    "error",
                    Message: fmt.Sprintf("Symbol %s not allowed for your subscription", symbol),
                })
                continue
            }
            
            // Subscribe to allowed symbol
            pubsub.Subscribe(ctx, fmt.Sprintf("quote:%s", symbol))
        }
    }
}
```

## Monitoring and Observability

### 1. Metrics

```go
// API Gateway metrics
var (
    websocketConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "websocket_connections_total",
        Help: "Total number of active WebSocket connections",
    })
    
    websocketMessagesReceived = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "websocket_messages_received_total",
        Help: "Total number of WebSocket messages received",
    })
    
    grpcStreamDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name:    "grpc_stream_duration_seconds",
        Help:    "Duration of gRPC stream connections",
        Buckets: prometheus.ExponentialBuckets(1, 2, 10),
    })
)

// Market Data Service metrics
var (
    redisSubscriptions = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "redis_subscriptions_total",
        Help: "Total number of active Redis subscriptions",
    })
    
    quotesPublished = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "quotes_published_total",
        Help: "Total number of quotes published",
    })
)
```

### 2. Logging

```go
// Structured logging with context
log.WithFields(log.Fields{
    "user_id":    userID,
    "symbols":    symbols,
    "connection": connectionID,
    "action":     "subscribe",
}).Info("User subscribed to symbols")
```

### 3. Tracing

```go
// OpenTelemetry tracing
ctx, span := tracer.Start(ctx, "StreamQuotes")
defer span.End()

span.SetAttributes(
    attribute.String("user_id", userID),
    attribute.StringSlice("symbols", symbols),
)
```

## Summary

| Component | Protocol | Responsibility |
|-----------|----------|----------------|
| **Frontend** | WebSocket (JSON) | User interface, display quotes |
| **API Gateway** | WebSocket ↔ gRPC | Protocol translation, auth, rate limiting |
| **Market Data Service** | gRPC Stream ↔ Redis Pub/Sub | Business logic, data fetching |
| **Redis** | Pub/Sub | Message broker for real-time quotes |

**Key Takeaways:**
1. ✅ API Gateway handles WebSocket complexity
2. ✅ Microservices only expose gRPC (simpler, more efficient)
3. ✅ Centralized security and observability
4. ✅ Easy to scale horizontally
5. ✅ Protocol translation allows frontend and backend to evolve independently

