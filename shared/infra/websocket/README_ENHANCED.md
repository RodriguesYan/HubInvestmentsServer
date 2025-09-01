# Enhanced WebSocket Infrastructure

## Overview

The enhanced WebSocket infrastructure provides enterprise-grade connection management, automatic scaling, error handling, and reconnection capabilities for real-time applications.

## Key Features

### 1. Advanced Connection Management
- **Connection Pooling**: Efficient management of WebSocket connections with metadata
- **Connection Limits**: Configurable maximum connections with graceful handling
- **Connection Metrics**: Real-time statistics on active, idle, and failed connections
- **Client Information**: Detailed tracking of client metadata (IP, User-Agent, User ID, etc.)

### 2. Automatic Scaling
- **Resource Monitoring**: CPU and memory usage tracking
- **Dynamic Scaling**: Automatic scale up/down based on configurable thresholds
- **Load Balancing**: Intelligent distribution of connections
- **Performance Optimization**: Automatic tuning for high/low load scenarios

### 3. Circuit Breaker Pattern
- **Failure Protection**: Prevents cascade failures with configurable thresholds
- **Recovery Logic**: Automatic recovery with half-open state testing
- **Graceful Degradation**: Maintains service availability during failures

### 4. Health Monitoring
- **Real-time Health Checks**: Continuous monitoring of system health
- **Health Status Levels**: Healthy, Degraded, Unhealthy, Critical states
- **Alert System**: Automated alerts for health status changes
- **Health History**: Historical tracking of system health

### 5. Reconnection Management
- **Automatic Reconnection**: Intelligent reconnection with multiple strategies
- **Backoff Strategies**: Linear, Exponential, and Fixed delay strategies
- **Jitter Support**: Prevents thundering herd problems
- **Priority Queuing**: High-priority reconnections for critical connections

## Architecture Components

```
┌─────────────────────────────────────────────────────────────────┐
│                    Enhanced WebSocket Manager                    │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ Connection Pool │  │ Health Monitor  │  │ Circuit Breaker │  │
│  │                 │  │                 │  │                 │  │
│  │ • Pooling       │  │ • Health Checks │  │ • Failure Guard │  │
│  │ • Metrics       │  │ • Alerting      │  │ • Recovery      │  │
│  │ • Lifecycle     │  │ • History       │  │ • State Mgmt    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐                      │
│  │ Auto Scaler     │  │ Reconnection    │                      │
│  │                 │  │ Handler         │                      │
│  │ • Resource Mon  │  │ • Auto Reconnect│                      │
│  │ • Scale Up/Down │  │ • Backoff Logic │                      │
│  │ • Optimization  │  │ • Priority Queue│                      │
│  └─────────────────┘  └─────────────────┘                      │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

### Connection Pool Configuration
```go
config := websocket.ConnectionPoolConfig{
    MaxConnections:      10000,           // Maximum concurrent connections
    MinConnections:      100,             // Minimum connections to maintain
    IdleTimeout:         30 * time.Minute, // Idle connection timeout
    MaxIdleTime:         1 * time.Hour,   // Maximum connection lifetime
    ScaleUpThreshold:    0.8,             // 80% usage triggers scale up
    ScaleDownThreshold:  0.3,             // 30% usage triggers scale down
    HealthCheckInterval: 30 * time.Second, // Health check frequency
    ReconnectAttempts:   3,               // Maximum reconnection attempts
    ReconnectDelay:      5 * time.Second, // Initial reconnection delay
}
```

### Circuit Breaker Configuration
```go
circuitConfig := websocket.CircuitBreakerConfig{
    FailureThreshold: 5,                  // Failures before opening circuit
    RecoveryTimeout:  30 * time.Second,   // Time before attempting recovery
    HalfOpenMaxCalls: 3,                  // Max calls in half-open state
}
```

### Reconnection Configuration
```go
reconnectConfig := websocket.ReconnectionConfig{
    MaxAttempts:         5,                              // Maximum reconnection attempts
    InitialDelay:        1 * time.Second,               // Initial delay
    MaxDelay:            30 * time.Second,              // Maximum delay
    Strategy:            ReconnectionStrategyExponential, // Backoff strategy
    Jitter:              true,                          // Add random jitter
    BackoffMultiplier:   2.0,                          // Exponential multiplier
    TimeoutPerAttempt:   10 * time.Second,             // Timeout per attempt
    EnableHealthCheck:   true,                         // Enable health monitoring
    HealthCheckInterval: 30 * time.Second,             // Health check frequency
}
```

## Usage Examples

### Basic Enhanced WebSocket Manager
```go
// Create enhanced WebSocket manager
config := websocket.DefaultWebSocketManagerConfig()
config.MaxConnections = 10000

manager := websocket.NewGorillaWebSocketManager(config)
defer manager.Close()

// Access advanced features
pool := manager.GetConnectionPool()
reconnectHandler := manager.GetReconnectionHandler()
metrics := manager.GetConnectionMetrics()
healthStatus := manager.GetHealthStatus()
```

### Connection Pool Usage
```go
// Create connection pool
poolConfig := websocket.DefaultConnectionPoolConfig()
pool := websocket.NewConnectionPool(poolConfig)
defer pool.Close()

// Add connection with client info
clientInfo := websocket.ClientInfo{
    IPAddress: "192.168.1.100",
    UserAgent: "Mozilla/5.0...",
    UserID:    "user123",
    SessionID: "session456",
}

conn := pool.AddConnection(websocketConn, clientInfo)

// Broadcast to all connections
message := []byte("Hello, WebSocket!")
err := pool.BroadcastToAll(websocket.TextMessage, message)

// Get metrics
metrics := pool.GetMetrics()
log.Printf("Active connections: %d", metrics.ActiveConnections)
```

### Reconnection Handler Usage
```go
// Create reconnection handler
reconnectConfig := websocket.DefaultReconnectionConfig()
handler := websocket.NewReconnectionHandler(reconnectConfig, pool)
defer handler.Stop()

// Schedule reconnection
clientInfo := websocket.ClientInfo{
    IPAddress: "192.168.1.100",
    UserID:    "user123",
}

handler.ScheduleReconnection("conn-123", clientInfo, "connection lost")

// Schedule high-priority reconnection
handler.ScheduleHighPriorityReconnection("critical-conn", clientInfo, "critical failure")

// Get reconnection metrics
metrics := handler.GetMetrics()
log.Printf("Total attempts: %d, Success rate: %.2f%%", 
    metrics.TotalAttempts, 
    float64(metrics.SuccessfulReconnects)/float64(metrics.TotalAttempts)*100)
```

### Health Monitoring
```go
// Create health monitor
healthConfig := websocket.DefaultConnectionPoolConfig()
monitor := websocket.NewHealthMonitor(pool, healthConfig)
defer monitor.Stop()

// Perform health check
result := monitor.CheckHealth()
log.Printf("Health status: %s, Active connections: %d, Error rate: %.2f%%",
    result.Status.String(), result.ActiveConnections, result.ErrorRate)

// Get health history
history := monitor.GetHealthHistory()
for _, check := range history {
    log.Printf("Time: %s, Status: %s, Response time: %v",
        check.Timestamp.Format(time.RFC3339),
        check.Status.String(),
        check.ResponseTime)
}
```

### Circuit Breaker Usage
```go
// Create circuit breaker
config := websocket.CircuitBreakerConfig{
    FailureThreshold: 5,
    RecoveryTimeout:  30 * time.Second,
    HalfOpenMaxCalls: 3,
}

cb := websocket.NewCircuitBreaker(config)

// Execute with circuit breaker protection
err := cb.Execute(func() error {
    // Your WebSocket operation here
    return conn.WriteMessage(messageType, data)
})

if err != nil {
    log.Printf("Circuit breaker error: %v", err)
}

// Check circuit breaker state
state := cb.GetState()
log.Printf("Circuit breaker state: %v", state)
```

## Metrics and Monitoring

### Connection Metrics
- **TotalConnections**: Total connections created
- **ActiveConnections**: Currently active connections
- **IdleConnections**: Idle connections
- **FailedConnections**: Failed connections
- **MessagesSent**: Total messages sent
- **MessagesReceived**: Total messages received
- **BytesSent**: Total bytes sent
- **BytesReceived**: Total bytes received
- **AverageLatency**: Average message latency

### Scaler Metrics
- **ScaleUpEvents**: Number of scale-up events
- **ScaleDownEvents**: Number of scale-down events
- **CPUUsage**: Current CPU usage percentage
- **MemoryUsage**: Current memory usage percentage
- **ConnectionLoad**: Connection load percentage

### Reconnection Metrics
- **TotalAttempts**: Total reconnection attempts
- **SuccessfulReconnects**: Successful reconnections
- **FailedReconnects**: Failed reconnections
- **AverageReconnectTime**: Average reconnection time
- **ActiveReconnections**: Currently active reconnections
- **QueueSize**: Reconnection queue size

## Health Status Levels

1. **Healthy**: All systems operating normally
2. **Degraded**: Performance issues detected, monitoring required
3. **Unhealthy**: High error rate or resource exhaustion detected
4. **Critical**: System failure detected, immediate attention required

## Best Practices

### 1. Configuration Tuning
- Set appropriate connection limits based on server capacity
- Configure health check intervals based on your SLA requirements
- Tune circuit breaker thresholds for your failure tolerance
- Adjust reconnection strategies based on network conditions

### 2. Monitoring and Alerting
- Monitor connection metrics continuously
- Set up alerts for health status changes
- Track reconnection success rates
- Monitor resource usage trends

### 3. Error Handling
- Implement graceful degradation for circuit breaker open states
- Log all reconnection attempts with context
- Handle connection limit exceeded scenarios gracefully
- Provide meaningful error messages to clients

### 4. Performance Optimization
- Use connection pooling for better resource utilization
- Enable compression for large message payloads
- Implement message batching where appropriate
- Monitor and optimize memory usage

### 5. Security Considerations
- Validate client information and implement rate limiting
- Use secure WebSocket connections (WSS) in production
- Implement proper authentication and authorization
- Monitor for suspicious connection patterns

## Integration with Realtime Quotes

The enhanced WebSocket infrastructure is fully integrated with the realtime quotes system:

```go
// Enhanced quotes handler with advanced features
handler := &RealtimeQuotesWebSocketHandler{
    wsManager:               enhancedWebSocketManager,
    priceOscillationService: priceService,
}

// Connection handling with error recovery
func (h *RealtimeQuotesWebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
    conn, err := h.wsManager.CreateConnection(w, r)
    if err != nil {
        // Enhanced error handling with capacity checks
        if err.Error() == "maximum connections limit reached" {
            http.Error(w, "Server at capacity, please try again later", http.StatusServiceUnavailable)
        } else {
            http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
        }
        return
    }
    
    // Log with health status
    healthStatus := h.wsManager.GetHealthStatus()
    log.Printf("New connection. Active: %d, Health: %s", 
        h.wsManager.GetActiveConnections(), healthStatus.String())
    
    go h.handleConnection(conn)
}
```

## Testing

The enhanced WebSocket infrastructure includes comprehensive tests:

```bash
# Run all WebSocket tests
go test ./shared/infra/websocket -v

# Run specific test suites
go test ./shared/infra/websocket -v -run TestConnectionPool
go test ./shared/infra/websocket -v -run TestCircuitBreaker
go test ./shared/infra/websocket -v -run TestReconnectionHandler
go test ./shared/infra/websocket -v -run TestHealthMonitor
```

## Performance Characteristics

### Scalability
- **Concurrent Connections**: Supports 10,000+ concurrent WebSocket connections
- **Message Throughput**: Optimized for high-frequency message broadcasting
- **Memory Efficiency**: Connection pooling reduces memory overhead
- **CPU Optimization**: Automatic scaling based on resource usage

### Reliability
- **Failure Recovery**: Automatic reconnection with intelligent backoff
- **Circuit Protection**: Prevents cascade failures
- **Health Monitoring**: Proactive issue detection
- **Graceful Degradation**: Maintains service availability during issues

### Monitoring
- **Real-time Metrics**: Live connection and performance statistics
- **Historical Data**: Health check history and trend analysis
- **Alerting**: Automated alerts for critical issues
- **Debugging**: Comprehensive logging for troubleshooting

This enhanced WebSocket infrastructure provides a robust foundation for real-time applications with enterprise-grade reliability, scalability, and monitoring capabilities.
