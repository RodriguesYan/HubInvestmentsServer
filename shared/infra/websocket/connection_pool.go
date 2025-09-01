package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ConnectionPool struct {
	connections   map[string]*PooledConnection
	mutex         sync.RWMutex
	config        ConnectionPoolConfig
	metrics       *ConnectionMetrics
	ctx           context.Context
	cancel        context.CancelFunc
	scaler        *ConnectionScaler
	healthMonitor *HealthMonitor
}

type ConnectionPoolConfig struct {
	MaxConnections       int                  `json:"max_connections"`
	MinConnections       int                  `json:"min_connections"`
	IdleTimeout          time.Duration        `json:"idle_timeout"`
	MaxIdleTime          time.Duration        `json:"max_idle_time"`
	ScaleUpThreshold     float64              `json:"scale_up_threshold"`   // CPU/Memory threshold to scale up
	ScaleDownThreshold   float64              `json:"scale_down_threshold"` // CPU/Memory threshold to scale down
	HealthCheckInterval  time.Duration        `json:"health_check_interval"`
	ReconnectAttempts    int                  `json:"reconnect_attempts"`
	ReconnectDelay       time.Duration        `json:"reconnect_delay"`
	CircuitBreakerConfig CircuitBreakerConfig `json:"circuit_breaker"`
}

type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	HalfOpenMaxCalls int           `json:"half_open_max_calls"`
}

func DefaultConnectionPoolConfig() ConnectionPoolConfig {
	return ConnectionPoolConfig{
		MaxConnections:      10000,
		MinConnections:      100,
		IdleTimeout:         30 * time.Minute,
		MaxIdleTime:         1 * time.Hour,
		ScaleUpThreshold:    0.8, // 80% CPU/Memory usage
		ScaleDownThreshold:  0.3, // 30% CPU/Memory usage
		HealthCheckInterval: 30 * time.Second,
		ReconnectAttempts:   3,
		ReconnectDelay:      5 * time.Second,
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  30 * time.Second,
			HalfOpenMaxCalls: 3,
		},
	}
}

type PooledConnection struct {
	Websocket
	ID             string
	CreatedAt      time.Time
	LastActivity   time.Time
	IsActive       bool
	FailureCount   int
	ClientInfo     ClientInfo
	CircuitBreaker *CircuitBreaker
	mutex          sync.RWMutex
}

type ClientInfo struct {
	UserAgent    string                 `json:"user_agent"`
	IPAddress    string                 `json:"ip_address"`
	UserID       string                 `json:"user_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	Capabilities []string               `json:"capabilities,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type ConnectionMetrics struct {
	TotalConnections  int64     `json:"total_connections"`
	ActiveConnections int64     `json:"active_connections"`
	IdleConnections   int64     `json:"idle_connections"`
	FailedConnections int64     `json:"failed_connections"`
	ReconnectAttempts int64     `json:"reconnect_attempts"`
	MessagesSent      int64     `json:"messages_sent"`
	MessagesReceived  int64     `json:"messages_received"`
	BytesSent         int64     `json:"bytes_sent"`
	BytesReceived     int64     `json:"bytes_received"`
	AverageLatency    float64   `json:"average_latency_ms"`
	LastUpdated       time.Time `json:"last_updated"`
	mutex             sync.RWMutex
}

func NewConnectionPool(config ConnectionPoolConfig) *ConnectionPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPool{
		connections: make(map[string]*PooledConnection),
		config:      config,
		metrics:     &ConnectionMetrics{LastUpdated: time.Now()},
		ctx:         ctx,
		cancel:      cancel,
	}

	pool.scaler = NewConnectionScaler(pool, config)
	pool.healthMonitor = NewHealthMonitor(pool, config)

	go pool.startMaintenanceRoutine()
	go pool.startMetricsCollection()

	return pool
}

func (p *ConnectionPool) AddConnection(conn Websocket, clientInfo ClientInfo) *PooledConnection {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	connectionID := fmt.Sprintf("pool_conn_%d_%d", time.Now().UnixNano(), len(p.connections))

	pooledConn := &PooledConnection{
		Websocket:      conn,
		ID:             connectionID,
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
		IsActive:       true,
		FailureCount:   0,
		ClientInfo:     clientInfo,
		CircuitBreaker: NewCircuitBreaker(p.config.CircuitBreakerConfig),
	}

	p.connections[connectionID] = pooledConn

	// Update metrics
	p.metrics.mutex.Lock()
	p.metrics.TotalConnections++
	p.metrics.ActiveConnections++
	p.metrics.LastUpdated = time.Now()
	p.metrics.mutex.Unlock()

	return pooledConn
}

func (p *ConnectionPool) RemoveConnection(connectionID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if pooledConn, exists := p.connections[connectionID]; exists {
		pooledConn.IsActive = false
		if err := pooledConn.Close(); err != nil {
			return fmt.Errorf("failed to close connection %s: %w", connectionID, err)
		}
		delete(p.connections, connectionID)

		// Update metrics
		p.metrics.mutex.Lock()
		p.metrics.ActiveConnections--
		p.metrics.LastUpdated = time.Now()
		p.metrics.mutex.Unlock()
	}

	return nil
}

func (p *ConnectionPool) GetConnection(connectionID string) (*PooledConnection, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if conn, exists := p.connections[connectionID]; exists && conn.IsActive {
		conn.LastActivity = time.Now()
		return conn, true
	}

	return nil, false
}

func (p *ConnectionPool) GetActiveConnections() []*PooledConnection {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var activeConnections []*PooledConnection
	for _, conn := range p.connections {
		if conn.IsActive {
			activeConnections = append(activeConnections, conn)
		}
	}

	return activeConnections
}

func (p *ConnectionPool) GetConnectionCount() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	count := 0
	for _, conn := range p.connections {
		if conn.IsActive {
			count++
		}
	}

	return count
}

func (p *ConnectionPool) BroadcastToAll(messageType int, data []byte) error {
	activeConnections := p.GetActiveConnections()

	var errors []error
	successCount := 0

	for _, conn := range activeConnections {
		if err := p.sendMessageWithCircuitBreaker(conn, messageType, data); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to connection %s: %w", conn.ID, err))
		} else {
			successCount++
		}
	}

	// Update metrics
	p.metrics.mutex.Lock()
	p.metrics.MessagesSent += int64(successCount)
	p.metrics.BytesSent += int64(len(data) * successCount)
	p.metrics.LastUpdated = time.Now()
	p.metrics.mutex.Unlock()

	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed for %d/%d connections: %v", len(errors), len(activeConnections), errors[0])
	}

	return nil
}

func (p *ConnectionPool) sendMessageWithCircuitBreaker(conn *PooledConnection, messageType int, data []byte) error {
	return conn.CircuitBreaker.Execute(func() error {
		return conn.WriteMessage(messageType, data)
	})
}

func (p *ConnectionPool) startMaintenanceRoutine() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performMaintenance()
		}
	}
}

func (p *ConnectionPool) performMaintenance() {
	p.cleanupIdleConnections()
	p.healthMonitor.CheckHealth()
	p.scaler.EvaluateScaling()
}

func (p *ConnectionPool) cleanupIdleConnections() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	now := time.Now()
	var toRemove []string

	for id, conn := range p.connections {
		if !conn.IsActive {
			continue
		}

		// Check if connection is idle
		if now.Sub(conn.LastActivity) > p.config.IdleTimeout {
			toRemove = append(toRemove, id)
			continue
		}

		// Check if connection is too old
		if now.Sub(conn.CreatedAt) > p.config.MaxIdleTime {
			toRemove = append(toRemove, id)
			continue
		}

		// Check if connection has too many failures
		if conn.FailureCount > p.config.ReconnectAttempts {
			toRemove = append(toRemove, id)
		}
	}

	// Remove stale connections
	for _, id := range toRemove {
		if conn, exists := p.connections[id]; exists {
			conn.IsActive = false
			conn.Close()
			delete(p.connections, id)

			// Update metrics
			p.metrics.mutex.Lock()
			p.metrics.ActiveConnections--
			p.metrics.FailedConnections++
			p.metrics.LastUpdated = time.Now()
			p.metrics.mutex.Unlock()
		}
	}
}

func (p *ConnectionPool) startMetricsCollection() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.updateMetrics()
		}
	}
}

func (p *ConnectionPool) updateMetrics() {
	p.mutex.RLock()
	activeCount := 0
	idleCount := 0
	now := time.Now()

	for _, conn := range p.connections {
		if conn.IsActive {
			activeCount++
			if now.Sub(conn.LastActivity) > p.config.IdleTimeout {
				idleCount++
			}
		}
	}
	p.mutex.RUnlock()

	p.metrics.mutex.Lock()
	p.metrics.ActiveConnections = int64(activeCount)
	p.metrics.IdleConnections = int64(idleCount)
	p.metrics.LastUpdated = now
	p.metrics.mutex.Unlock()
}

func (p *ConnectionPool) GetMetrics() ConnectionMetrics {
	p.metrics.mutex.RLock()
	defer p.metrics.mutex.RUnlock()

	// Return a copy to avoid mutex copying
	return ConnectionMetrics{
		TotalConnections:  p.metrics.TotalConnections,
		ActiveConnections: p.metrics.ActiveConnections,
		IdleConnections:   p.metrics.IdleConnections,
		FailedConnections: p.metrics.FailedConnections,
		ReconnectAttempts: p.metrics.ReconnectAttempts,
		MessagesSent:      p.metrics.MessagesSent,
		MessagesReceived:  p.metrics.MessagesReceived,
		BytesSent:         p.metrics.BytesSent,
		BytesReceived:     p.metrics.BytesReceived,
		AverageLatency:    p.metrics.AverageLatency,
		LastUpdated:       p.metrics.LastUpdated,
	}
}

func (p *ConnectionPool) Close() error {
	p.cancel()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	var errors []error
	for id, conn := range p.connections {
		if err := conn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection %s: %w", id, err))
		}
	}

	p.connections = make(map[string]*PooledConnection)

	if len(errors) > 0 {
		return fmt.Errorf("failed to close some connections: %v", errors[0])
	}

	return nil
}
