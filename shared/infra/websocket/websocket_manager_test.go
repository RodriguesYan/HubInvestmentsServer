package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestGorillaWebSocketManager(t *testing.T) {
	config := DefaultWebSocketManagerConfig()
	config.MaxConnections = 5 // Lower limit for testing

	manager := NewGorillaWebSocketManager(config)
	defer manager.Close()

	t.Run("Manager Creation", func(t *testing.T) {
		assert.NotNil(t, manager)
		assert.Equal(t, 0, manager.GetActiveConnections())
		assert.NoError(t, manager.HealthCheck())
	})

	t.Run("Connection Registration", func(t *testing.T) {
		// Create a mock WebSocket connection
		mockConn := &MockWebsocket{}

		connectionID := manager.RegisterConnection(mockConn)
		assert.NotEmpty(t, connectionID)
		assert.Equal(t, 1, manager.GetActiveConnections())

		// Verify connection can be retrieved
		retrievedConn, exists := manager.GetConnection(connectionID)
		assert.True(t, exists)
		assert.Equal(t, mockConn, retrievedConn)

		// Unregister connection
		manager.UnregisterConnection(connectionID)
		assert.Equal(t, 0, manager.GetActiveConnections())

		// Verify connection no longer exists
		_, exists = manager.GetConnection(connectionID)
		assert.False(t, exists)
	})

	t.Run("Broadcast Message", func(t *testing.T) {
		// Register multiple mock connections
		mockConn1 := &MockWebsocket{}
		mockConn2 := &MockWebsocket{}

		id1 := manager.RegisterConnection(mockConn1)
		id2 := manager.RegisterConnection(mockConn2)

		// Broadcast message
		testMessage := []byte("test broadcast message")
		err := manager.BroadcastMessage(websocket.TextMessage, testMessage)
		assert.NoError(t, err)

		// Verify both connections received the message
		assert.Equal(t, 1, mockConn1.WriteMessageCallCount)
		assert.Equal(t, websocket.TextMessage, mockConn1.LastMessageType)
		assert.Equal(t, testMessage, mockConn1.LastMessage)

		assert.Equal(t, 1, mockConn2.WriteMessageCallCount)
		assert.Equal(t, websocket.TextMessage, mockConn2.LastMessageType)
		assert.Equal(t, testMessage, mockConn2.LastMessage)

		// Cleanup
		manager.UnregisterConnection(id1)
		manager.UnregisterConnection(id2)
	})

	t.Run("Connection Limit", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := manager.CreateConnection(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
		}))
		defer server.Close()

		// Convert HTTP URL to WebSocket URL
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Create connections up to the limit
		var connections []*websocket.Conn
		for i := 0; i < config.MaxConnections; i++ {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Logf("Failed to create connection %d: %v", i, err)
				break
			}
			connections = append(connections, conn)
		}

		// Verify we have connections
		assert.Greater(t, len(connections), 0)

		// Clean up connections
		for _, conn := range connections {
			conn.Close()
		}

		// Wait a bit for cleanup
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Health Check", func(t *testing.T) {
		// Create a fresh manager for this test to avoid interference
		freshConfig := DefaultWebSocketManagerConfig()
		freshConfig.MaxConnections = 5
		freshManager := NewGorillaWebSocketManager(freshConfig)
		defer freshManager.Close()

		// Manager should be healthy initially
		assert.NoError(t, freshManager.HealthCheck())

		// Register connections within limit
		mockConn := &MockWebsocket{}
		id := freshManager.RegisterConnection(mockConn)
		assert.NoError(t, freshManager.HealthCheck())

		// Cleanup
		freshManager.UnregisterConnection(id)
		assert.NoError(t, freshManager.HealthCheck())
	})
}

func TestWebSocketHandler(t *testing.T) {
	config := DefaultWebSocketManagerConfig()
	manager := NewGorillaWebSocketManager(config)
	defer manager.Close()

	handler := NewWebSocketHandler(manager)

	t.Run("Handler Creation", func(t *testing.T) {
		assert.NotNil(t, handler)
	})

	t.Run("Connection Stats", func(t *testing.T) {
		stats := handler.GetConnectionStats()
		assert.Contains(t, stats, "active_connections")
		assert.Contains(t, stats, "health_status")
		assert.Contains(t, stats, "timestamp")

		assert.Equal(t, 0, stats["active_connections"])
		assert.Equal(t, "healthy", stats["health_status"])
	})

	t.Run("Broadcast", func(t *testing.T) {
		// Register a mock connection
		mockConn := &MockWebsocket{}
		manager.RegisterConnection(mockConn)

		// Broadcast message through handler
		testMessage := []byte("handler broadcast test")
		err := handler.BroadcastToAll(websocket.TextMessage, testMessage)
		assert.NoError(t, err)

		// Verify message was sent
		assert.Equal(t, 1, mockConn.WriteMessageCallCount)
		assert.Equal(t, testMessage, mockConn.LastMessage)
	})
}

// MockWebsocket implements the Websocket interface for testing
type MockWebsocket struct {
	WriteMessageCallCount int
	LastMessageType       int
	LastMessage           []byte
	ReadMessageResponse   []byte
	ReadMessageError      error
	CloseError            error
}

func (m *MockWebsocket) ReadMessage() (messageType int, p []byte, err error) {
	if m.ReadMessageError != nil {
		return 0, nil, m.ReadMessageError
	}
	return websocket.TextMessage, m.ReadMessageResponse, nil
}

func (m *MockWebsocket) WriteMessage(messageType int, data []byte) error {
	m.WriteMessageCallCount++
	m.LastMessageType = messageType
	m.LastMessage = make([]byte, len(data))
	copy(m.LastMessage, data)
	return nil
}

func (m *MockWebsocket) Close() error {
	return m.CloseError
}
