package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketManager manages WebSocket connections and provides the interface for other modules
type WebSocketManager interface {
	// CreateConnection creates a new WebSocket connection from an HTTP request
	CreateConnection(w http.ResponseWriter, r *http.Request) (Websocket, error)

	// GetActiveConnections returns the number of active connections
	GetActiveConnections() int

	// BroadcastMessage sends a message to all active connections
	BroadcastMessage(messageType int, data []byte) error

	// RegisterConnection registers a connection for management
	RegisterConnection(conn Websocket) string

	// UnregisterConnection removes a connection from management
	UnregisterConnection(connectionID string)

	// GetConnection retrieves a specific connection by ID
	GetConnection(connectionID string) (Websocket, bool)

	// Close gracefully closes all connections
	Close() error

	// HealthCheck returns the health status of the WebSocket manager
	HealthCheck() error

	// Advanced connection management methods
	GetConnectionPool() *ConnectionPool
	GetConnectionMetrics() ConnectionMetrics
	GetHealthStatus() HealthStatus
}

// WebSocketManagerConfig holds configuration for the WebSocket manager
type WebSocketManagerConfig struct {
	// CheckOrigin defines the origin checker function
	CheckOrigin func(r *http.Request) bool

	// ReadBufferSize sets the read buffer size
	ReadBufferSize int

	// WriteBufferSize sets the write buffer size
	WriteBufferSize int

	// HandshakeTimeout sets the handshake timeout
	HandshakeTimeout time.Duration

	// EnableCompression enables per-message compression
	EnableCompression bool

	// MaxConnections limits the maximum number of concurrent connections
	MaxConnections int

	// PingInterval sets the interval for ping messages
	PingInterval time.Duration

	// PongTimeout sets the timeout for pong responses
	PongTimeout time.Duration
}

// DefaultWebSocketManagerConfig returns a default configuration
func DefaultWebSocketManagerConfig() WebSocketManagerConfig {
	return WebSocketManagerConfig{
		CheckOrigin: func(r *http.Request) bool {
			// In production, implement proper origin checking
			return true
		},
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: true,
		MaxConnections:    10000,
		PingInterval:      30 * time.Second,
		PongTimeout:       60 * time.Second,
	}
}

// GorillaWebSocketManager implements WebSocketManager using Gorilla WebSocket
type GorillaWebSocketManager struct {
	upgrader       websocket.Upgrader
	connections    map[string]*managedConnection
	mutex          sync.RWMutex
	config         WebSocketManagerConfig
	ctx            context.Context
	cancel         context.CancelFunc
	connectionPool *ConnectionPool
}

// managedConnection wraps a WebSocket connection with metadata
type managedConnection struct {
	conn       Websocket
	id         string
	createdAt  time.Time
	lastPing   time.Time
	isActive   bool
	clientInfo map[string]interface{}
}

// NewGorillaWebSocketManager creates a new WebSocket manager using Gorilla WebSocket
func NewGorillaWebSocketManager(config WebSocketManagerConfig) WebSocketManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &GorillaWebSocketManager{
		upgrader: websocket.Upgrader{
			CheckOrigin:       config.CheckOrigin,
			ReadBufferSize:    config.ReadBufferSize,
			WriteBufferSize:   config.WriteBufferSize,
			HandshakeTimeout:  config.HandshakeTimeout,
			EnableCompression: config.EnableCompression,
		},
		connections: make(map[string]*managedConnection),
		config:      config,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize advanced connection management
	poolConfig := DefaultConnectionPoolConfig()
	poolConfig.MaxConnections = config.MaxConnections
	poolConfig.HealthCheckInterval = 30 * time.Second
	manager.connectionPool = NewConnectionPool(poolConfig)

	// Start background maintenance routines
	go manager.startPingRoutine()
	go manager.startCleanupRoutine()

	return manager
}

// CreateConnection upgrades an HTTP connection to WebSocket
func (m *GorillaWebSocketManager) CreateConnection(w http.ResponseWriter, r *http.Request) (Websocket, error) {
	// Check connection limit
	if m.GetActiveConnections() >= m.config.MaxConnections {
		return nil, fmt.Errorf("maximum connections limit reached: %d", m.config.MaxConnections)
	}

	// Upgrade the HTTP connection to WebSocket
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade connection: %w", err)
	}

	// Wrap with our interface
	websocketConn := NewGorillaWebsocket(conn)

	// Register the connection
	connectionID := m.RegisterConnection(websocketConn)

	// Set up connection-specific handlers
	m.setupConnectionHandlers(conn, connectionID)

	return websocketConn, nil
}

// RegisterConnection registers a connection for management
func (m *GorillaWebSocketManager) RegisterConnection(conn Websocket) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Generate unique connection ID
	connectionID := fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(), len(m.connections))

	m.connections[connectionID] = &managedConnection{
		conn:       conn,
		id:         connectionID,
		createdAt:  time.Now(),
		lastPing:   time.Now(),
		isActive:   true,
		clientInfo: make(map[string]interface{}),
	}

	return connectionID
}

// UnregisterConnection removes a connection from management
func (m *GorillaWebSocketManager) UnregisterConnection(connectionID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if managedConn, exists := m.connections[connectionID]; exists {
		managedConn.isActive = false
		managedConn.conn.Close()
		delete(m.connections, connectionID)
	}
}

// GetConnection retrieves a specific connection by ID
func (m *GorillaWebSocketManager) GetConnection(connectionID string) (Websocket, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if managedConn, exists := m.connections[connectionID]; exists && managedConn.isActive {
		return managedConn.conn, true
	}

	return nil, false
}

// GetActiveConnections returns the number of active connections
func (m *GorillaWebSocketManager) GetActiveConnections() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	activeCount := 0
	for _, conn := range m.connections {
		if conn.isActive {
			activeCount++
		}
	}

	return activeCount
}

// BroadcastMessage sends a message to all active connections
func (m *GorillaWebSocketManager) BroadcastMessage(messageType int, data []byte) error {
	m.mutex.RLock()
	connections := make([]*managedConnection, 0, len(m.connections))
	for _, conn := range m.connections {
		if conn.isActive {
			connections = append(connections, conn)
		}
	}
	m.mutex.RUnlock()

	var errors []error
	for _, managedConn := range connections {
		if err := managedConn.conn.WriteMessage(messageType, data); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to connection %s: %w", managedConn.id, err))
			// Mark connection as inactive and schedule for cleanup
			m.UnregisterConnection(managedConn.id)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed for %d connections: %v", len(errors), errors[0])
	}

	return nil
}

// setupConnectionHandlers sets up ping/pong handlers for a connection
func (m *GorillaWebSocketManager) setupConnectionHandlers(conn *websocket.Conn, connectionID string) {
	// Set up pong handler
	conn.SetPongHandler(func(string) error {
		m.mutex.Lock()
		if managedConn, exists := m.connections[connectionID]; exists {
			managedConn.lastPing = time.Now()
		}
		m.mutex.Unlock()
		return nil
	})

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(m.config.PongTimeout))
}

// startPingRoutine starts the ping routine for all connections
func (m *GorillaWebSocketManager) startPingRoutine() {
	ticker := time.NewTicker(m.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.sendPingToAllConnections()
		}
	}
}

// sendPingToAllConnections sends ping messages to all active connections
func (m *GorillaWebSocketManager) sendPingToAllConnections() {
	m.mutex.RLock()
	connections := make([]*managedConnection, 0, len(m.connections))
	for _, conn := range m.connections {
		if conn.isActive {
			connections = append(connections, conn)
		}
	}
	m.mutex.RUnlock()

	for _, managedConn := range connections {
		if err := managedConn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			// Connection is likely dead, mark for cleanup
			m.UnregisterConnection(managedConn.id)
		}
	}
}

// startCleanupRoutine starts the cleanup routine for stale connections
func (m *GorillaWebSocketManager) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupStaleConnections()
		}
	}
}

// cleanupStaleConnections removes connections that haven't responded to pings
func (m *GorillaWebSocketManager) cleanupStaleConnections() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	staleThreshold := time.Now().Add(-m.config.PongTimeout)
	var staleConnections []string

	for id, conn := range m.connections {
		if conn.lastPing.Before(staleThreshold) {
			staleConnections = append(staleConnections, id)
		}
	}

	// Remove stale connections
	for _, id := range staleConnections {
		if managedConn, exists := m.connections[id]; exists {
			managedConn.conn.Close()
			delete(m.connections, id)
		}
	}
}

// HealthCheck returns the health status of the WebSocket manager
func (m *GorillaWebSocketManager) HealthCheck() error {
	activeConnections := m.GetActiveConnections()

	if activeConnections > m.config.MaxConnections {
		return fmt.Errorf("too many active connections: %d > %d", activeConnections, m.config.MaxConnections)
	}

	return nil
}

// GetConnectionPool returns the connection pool
func (m *GorillaWebSocketManager) GetConnectionPool() *ConnectionPool {
	return m.connectionPool
}

// GetConnectionMetrics returns connection metrics
func (m *GorillaWebSocketManager) GetConnectionMetrics() ConnectionMetrics {
	return m.connectionPool.GetMetrics()
}

// GetHealthStatus returns the current health status
func (m *GorillaWebSocketManager) GetHealthStatus() HealthStatus {
	return m.connectionPool.healthMonitor.GetCurrentStatus()
}

// Close gracefully closes all connections and stops background routines
func (m *GorillaWebSocketManager) Close() error {
	// Cancel background routines
	m.cancel()

	// Close advanced components
	if m.connectionPool != nil {
		m.connectionPool.Close()
	}

	// Close all connections
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var errors []error
	for id, managedConn := range m.connections {
		if err := managedConn.conn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection %s: %w", id, err))
		}
	}

	// Clear connections map
	m.connections = make(map[string]*managedConnection)

	if len(errors) > 0 {
		return fmt.Errorf("failed to close some connections: %v", errors[0])
	}

	return nil
}
