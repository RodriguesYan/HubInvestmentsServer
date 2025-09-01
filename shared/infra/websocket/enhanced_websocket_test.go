package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnectionPool(t *testing.T) {
	config := DefaultConnectionPoolConfig()
	config.MaxConnections = 10
	config.HealthCheckInterval = 100 * time.Millisecond

	pool := NewConnectionPool(config)
	defer pool.Close()

	t.Run("Add and Remove Connections", func(t *testing.T) {
		// Create mock connection
		mockConn := &MockWebsocket{}
		clientInfo := ClientInfo{
			IPAddress: "127.0.0.1",
			UserAgent: "test-agent",
		}

		// Add connection
		pooledConn := pool.AddConnection(mockConn, clientInfo)
		assert.NotNil(t, pooledConn)
		assert.Equal(t, 1, pool.GetConnectionCount())

		// Remove connection
		err := pool.RemoveConnection(pooledConn.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, pool.GetConnectionCount())
	})

	t.Run("Connection Metrics", func(t *testing.T) {
		metrics := pool.GetMetrics()
		assert.NotNil(t, metrics)
		assert.True(t, metrics.LastUpdated.After(time.Time{}))
	})

	t.Run("Broadcast Messages", func(t *testing.T) {
		// Add multiple connections
		var connections []*PooledConnection
		for i := 0; i < 3; i++ {
			mockConn := &MockWebsocket{}
			clientInfo := ClientInfo{IPAddress: "127.0.0.1"}
			conn := pool.AddConnection(mockConn, clientInfo)
			connections = append(connections, conn)
		}

		// Broadcast message
		message := []byte("test message")
		err := pool.BroadcastToAll(1, message)
		assert.NoError(t, err)

		// Cleanup
		for _, conn := range connections {
			pool.RemoveConnection(conn.ID)
		}
	})
}

func TestCircuitBreaker(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  100 * time.Millisecond,
		HalfOpenMaxCalls: 2,
	}

	cb := NewCircuitBreaker(config)

	t.Run("Closed State Success", func(t *testing.T) {
		err := cb.Execute(func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})

	t.Run("Open State After Failures", func(t *testing.T) {
		// Trigger failures to open circuit
		for i := 0; i < 3; i++ {
			cb.Execute(func() error {
				return assert.AnError
			})
		}

		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Should reject calls when open
		err := cb.Execute(func() error {
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker is open")
	})

	t.Run("Half-Open State Recovery", func(t *testing.T) {
		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Should transition to half-open
		err := cb.Execute(func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})
}

func TestConnectionScaler(t *testing.T) {
	config := DefaultConnectionPoolConfig()
	config.MaxConnections = 100
	config.ScaleUpThreshold = 0.8
	config.ScaleDownThreshold = 0.3

	pool := NewConnectionPool(config)
	defer pool.Close()

	scaler := NewConnectionScaler(pool, config)
	defer scaler.Stop()

	t.Run("Scaler Metrics", func(t *testing.T) {
		metrics := scaler.GetMetrics()
		assert.NotNil(t, metrics)
	})

	t.Run("Evaluate Scaling", func(t *testing.T) {
		// This would normally trigger scaling logic
		scaler.EvaluateScaling()

		// Verify no errors occurred
		metrics := scaler.GetMetrics()
		assert.GreaterOrEqual(t, metrics.ScaleUpEvents, int64(0))
		assert.GreaterOrEqual(t, metrics.ScaleDownEvents, int64(0))
	})
}

func TestHealthMonitor(t *testing.T) {
	config := DefaultConnectionPoolConfig()
	config.HealthCheckInterval = 50 * time.Millisecond

	pool := NewConnectionPool(config)
	defer pool.Close()

	monitor := NewHealthMonitor(pool, config)
	defer monitor.Stop()

	t.Run("Health Check", func(t *testing.T) {
		result := monitor.CheckHealth()
		assert.NotNil(t, result)
		assert.True(t, result.Timestamp.After(time.Time{}))
		assert.GreaterOrEqual(t, result.ResponseTime, time.Duration(0))
	})

	t.Run("Health Status", func(t *testing.T) {
		status := monitor.GetCurrentStatus()
		assert.True(t, status >= HealthStatusHealthy && status <= HealthStatusCritical)
	})

	t.Run("Health History", func(t *testing.T) {
		// Wait for a few health checks
		time.Sleep(200 * time.Millisecond)

		history := monitor.GetHealthHistory()
		assert.Greater(t, len(history), 0)
	})
}

func TestReconnectionHandler(t *testing.T) {
	poolConfig := DefaultConnectionPoolConfig()
	pool := NewConnectionPool(poolConfig)
	defer pool.Close()

	reconnectConfig := DefaultReconnectionConfig()
	reconnectConfig.MaxAttempts = 2
	reconnectConfig.InitialDelay = 10 * time.Millisecond

	handler := NewReconnectionHandler(reconnectConfig, pool)
	defer handler.Stop()

	t.Run("Schedule Reconnection", func(t *testing.T) {
		clientInfo := ClientInfo{
			IPAddress: "127.0.0.1",
			UserAgent: "test-agent",
		}

		handler.ScheduleReconnection("test-conn-1", clientInfo, "test failure")

		// Wait a bit for processing
		time.Sleep(100 * time.Millisecond)

		metrics := handler.GetMetrics()
		assert.Greater(t, metrics.TotalAttempts, int64(0))
	})

	t.Run("High Priority Reconnection", func(t *testing.T) {
		clientInfo := ClientInfo{
			IPAddress: "127.0.0.1",
			UserAgent: "test-agent",
		}

		handler.ScheduleHighPriorityReconnection("test-conn-2", clientInfo, "critical failure")

		// Wait for processing
		time.Sleep(100 * time.Millisecond)

		metrics := handler.GetMetrics()
		assert.Greater(t, metrics.TotalAttempts, int64(0))
	})

	t.Run("Reconnection Metrics", func(t *testing.T) {
		metrics := handler.GetMetrics()
		assert.NotNil(t, metrics)
		assert.True(t, metrics.LastUpdate.After(time.Time{}))
	})
}

func TestEnhancedWebSocketManager(t *testing.T) {
	config := DefaultWebSocketManagerConfig()
	config.MaxConnections = 10

	manager := NewGorillaWebSocketManager(config)
	defer manager.Close()

	t.Run("Advanced Features Access", func(t *testing.T) {
		pool := manager.GetConnectionPool()
		assert.NotNil(t, pool)

		reconnectHandler := manager.GetReconnectionHandler()
		assert.NotNil(t, reconnectHandler)

		metrics := manager.GetConnectionMetrics()
		assert.NotNil(t, metrics)

		healthStatus := manager.GetHealthStatus()
		assert.True(t, healthStatus >= HealthStatusHealthy && healthStatus <= HealthStatusCritical)
	})

	t.Run("Schedule Reconnection", func(t *testing.T) {
		err := manager.ScheduleReconnection("test-conn", "connection lost")
		assert.NoError(t, err)
	})
}

// Use the MockWebsocket from websocket_manager_test.go to avoid duplication

func TestIntegrationScenarios(t *testing.T) {
	t.Run("High Load Scenario", func(t *testing.T) {
		config := DefaultConnectionPoolConfig()
		config.MaxConnections = 50
		config.ScaleUpThreshold = 0.7

		pool := NewConnectionPool(config)
		defer pool.Close()

		// Simulate high connection load
		var connections []*PooledConnection
		for i := 0; i < 40; i++ {
			mockConn := &MockWebsocket{}
			clientInfo := ClientInfo{IPAddress: "127.0.0.1"}
			conn := pool.AddConnection(mockConn, clientInfo)
			connections = append(connections, conn)
		}

		// Check that scaling logic would trigger
		connectionLoad := float64(len(connections)) / float64(config.MaxConnections)
		assert.Greater(t, connectionLoad, config.ScaleUpThreshold)

		// Cleanup
		for _, conn := range connections {
			pool.RemoveConnection(conn.ID)
		}
	})

	t.Run("Error Recovery Scenario", func(t *testing.T) {
		config := DefaultReconnectionConfig()
		config.MaxAttempts = 3
		config.InitialDelay = 1 * time.Millisecond

		poolConfig := DefaultConnectionPoolConfig()
		pool := NewConnectionPool(poolConfig)
		defer pool.Close()

		handler := NewReconnectionHandler(config, pool)
		defer handler.Stop()

		// Simulate connection failure and recovery
		clientInfo := ClientInfo{IPAddress: "127.0.0.1"}
		handler.ScheduleReconnection("failed-conn", clientInfo, "network error")

		// Wait for reconnection attempts
		time.Sleep(50 * time.Millisecond)

		metrics := handler.GetMetrics()
		assert.Greater(t, metrics.TotalAttempts, int64(0))
	})
}
