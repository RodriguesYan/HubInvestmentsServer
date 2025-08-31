# WebSocket Infrastructure

This package provides a WebSocket abstraction layer using the adapter pattern, allowing modules to use WebSocket functionality through a clean interface without being tightly coupled to the Gorilla WebSocket implementation.

## Architecture

```
┌─────────────────────────┐
│   Business Modules      │ ← Use only the Websocket interface
│   (Order Management,    │
│    Market Data, etc.)   │
└─────────────────────────┘
           ↓
┌─────────────────────────┐
│  WebSocket Interface    │ ← Clean abstraction layer
│  (websocket.Websocket)  │
└─────────────────────────┘
           ↓
┌─────────────────────────┐
│ WebSocket Manager       │ ← Connection management & lifecycle
│ (WebSocketManager)      │
└─────────────────────────┘
           ↓
┌─────────────────────────┐
│ Gorilla WebSocket       │ ← Concrete implementation
│ (GorillaWebsocket)      │
└─────────────────────────┘
```

## Components

### 1. Core Interface (`websocket.Websocket`)

```go
type Websocket interface {
    ReadMessage() (messageType int, p []byte, err error)
    WriteMessage(messageType int, data []byte) error
    Close() error
}
```

### 2. WebSocket Manager (`WebSocketManager`)

Manages WebSocket connections, provides:
- Connection lifecycle management
- Broadcasting to multiple connections
- Health monitoring and connection limits
- Automatic ping/pong handling
- Connection cleanup and resource management

### 3. Gorilla WebSocket Implementation (`GorillaWebsocket`)

Concrete implementation using the Gorilla WebSocket library.

## Usage

### 1. Access via Dependency Injection Container

```go
// Get WebSocket manager from container
wsManager := container.GetWebSocketManager()

// Create WebSocket handler
wsHandler := websocket.NewWebSocketHandler(wsManager)
```

### 2. HTTP Handler Setup

```go
func SetupWebSocketRoutes(mux *http.ServeMux, container di.Container) {
    wsManager := container.GetWebSocketManager()
    wsHandler := websocket.NewWebSocketHandler(wsManager)
    
    // WebSocket endpoint
    mux.HandleFunc("/ws", wsHandler.HandleWebSocketConnection)
    
    // Connection statistics endpoint
    mux.HandleFunc("/ws/stats", func(w http.ResponseWriter, r *http.Request) {
        stats := wsHandler.GetConnectionStats()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(stats)
    })
}
```

### 3. Using WebSocket in Business Modules

```go
// Example: Real-time order updates
type OrderNotificationService struct {
    wsManager websocket.WebSocketManager
}

func (s *OrderNotificationService) NotifyOrderUpdate(orderID string, status string) error {
    message := map[string]interface{}{
        "type": "order_update",
        "order_id": orderID,
        "status": status,
        "timestamp": time.Now().Unix(),
    }
    
    data, _ := json.Marshal(message)
    return s.wsManager.BroadcastMessage(websocket.TextMessage, data)
}
```

### 4. Client-Side JavaScript Example

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = function(event) {
    console.log('Connected to WebSocket');
    
    // Send subscription message
    ws.send(JSON.stringify({
        type: 'subscribe',
        channels: ['orders', 'market_data']
    }));
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
    
    // Handle different message types
    switch(data.type) {
        case 'order_update':
            handleOrderUpdate(data);
            break;
        case 'market_data':
            handleMarketData(data);
            break;
    }
};

ws.onclose = function(event) {
    console.log('WebSocket connection closed');
};
```

## Configuration

### Environment Variables

- `WEBSOCKET_MAX_CONNECTIONS`: Maximum number of concurrent connections (default: 10000)

### Configuration Options

```go
type WebSocketManagerConfig struct {
    CheckOrigin       func(r *http.Request) bool
    ReadBufferSize    int
    WriteBufferSize   int
    HandshakeTimeout  time.Duration
    EnableCompression bool
    MaxConnections    int
    PingInterval      time.Duration
    PongTimeout       time.Duration
}
```

## Features

### Connection Management
- Automatic connection registration and cleanup
- Connection limits and health monitoring
- Ping/pong heartbeat mechanism
- Graceful shutdown handling

### Broadcasting
- Send messages to all connected clients
- Efficient message delivery
- Automatic cleanup of dead connections

### Monitoring
- Active connection count
- Health status checking
- Connection statistics

## Testing

The package includes comprehensive tests:

```bash
go test ./shared/infra/websocket/...
```

### Mock Implementation

For testing business modules that use WebSocket:

```go
type MockWebsocket struct {
    WriteMessageCallCount int
    LastMessageType       int
    LastMessage           []byte
    // ... other mock fields
}
```

## Integration with Other Systems

### Order Management System
- Real-time order status updates
- Order execution notifications
- Portfolio value changes

### Market Data Service
- Real-time price feeds
- Market alerts and notifications
- Trading session updates

### User Notifications
- System alerts
- Account notifications
- Trading confirmations

## Performance Considerations

- **Connection Limits**: Default 10,000 concurrent connections
- **Memory Usage**: Each connection uses ~8KB of memory
- **CPU Usage**: Ping/pong and cleanup routines run in background
- **Network**: Efficient message broadcasting with automatic cleanup

## Security Considerations

- **Origin Checking**: Implement proper origin validation in production
- **Authentication**: Integrate with existing JWT authentication
- **Rate Limiting**: Consider implementing per-connection rate limits
- **Message Validation**: Always validate incoming messages

## Future Enhancements

- [ ] Authentication integration with JWT tokens
- [ ] Per-connection rate limiting
- [ ] Message queuing for offline clients
- [ ] Connection clustering for horizontal scaling
- [ ] Metrics integration with Prometheus
- [ ] SSL/TLS support configuration
