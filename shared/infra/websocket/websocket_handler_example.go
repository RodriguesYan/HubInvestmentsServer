package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketHandler demonstrates how to use the WebSocket interface in HTTP handlers
type WebSocketHandler struct {
	manager WebSocketManager
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(manager WebSocketManager) *WebSocketHandler {
	return &WebSocketHandler{
		manager: manager,
	}
}

// HandleWebSocketConnection handles WebSocket upgrade and connection management
func (h *WebSocketHandler) HandleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	// Create WebSocket connection using the manager
	conn, err := h.manager.CreateConnection(w, r)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	// Register connection with manager (this is done automatically in CreateConnection)
	log.Printf("New WebSocket connection established. Active connections: %d", h.manager.GetActiveConnections())

	// Handle the connection in a goroutine
	go h.handleConnection(conn)
}

// handleConnection handles messages from a WebSocket connection
func (h *WebSocketHandler) handleConnection(conn Websocket) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	for {
		// Read message from client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process the message
		response, err := h.processMessage(messageType, message)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			continue
		}

		// Send response back to client
		if err := conn.WriteMessage(messageType, response); err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}
}

// processMessage processes incoming WebSocket messages
func (h *WebSocketHandler) processMessage(messageType int, message []byte) ([]byte, error) {
	// Example message processing - in real implementation, this would be more sophisticated
	switch messageType {
	case websocket.TextMessage:
		return h.processTextMessage(message)
	case websocket.BinaryMessage:
		return h.processBinaryMessage(message)
	default:
		return nil, fmt.Errorf("unsupported message type: %d", messageType)
	}
}

// processTextMessage processes text messages
func (h *WebSocketHandler) processTextMessage(message []byte) ([]byte, error) {
	// Parse JSON message
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		return nil, fmt.Errorf("invalid JSON message: %w", err)
	}

	// Example: Echo message with timestamp
	response := map[string]interface{}{
		"type":      "response",
		"original":  msg,
		"timestamp": time.Now().Unix(),
		"server":    "websocket-handler",
	}

	return json.Marshal(response)
}

// processBinaryMessage processes binary messages
func (h *WebSocketHandler) processBinaryMessage(message []byte) ([]byte, error) {
	// Example: Return message length as binary response
	length := len(message)
	response := make([]byte, 4)
	response[0] = byte(length >> 24)
	response[1] = byte(length >> 16)
	response[2] = byte(length >> 8)
	response[3] = byte(length)

	return response, nil
}

// BroadcastToAll broadcasts a message to all connected clients
func (h *WebSocketHandler) BroadcastToAll(messageType int, data []byte) error {
	return h.manager.BroadcastMessage(messageType, data)
}

// GetConnectionStats returns connection statistics
func (h *WebSocketHandler) GetConnectionStats() map[string]interface{} {
	return map[string]interface{}{
		"active_connections": h.manager.GetActiveConnections(),
		"health_status":      h.getHealthStatus(),
		"timestamp":          time.Now().Unix(),
	}
}

// getHealthStatus returns the health status of the WebSocket system
func (h *WebSocketHandler) getHealthStatus() string {
	if err := h.manager.HealthCheck(); err != nil {
		return fmt.Sprintf("unhealthy: %v", err)
	}
	return "healthy"
}

// Example usage in HTTP routes:
//
// func SetupWebSocketRoutes(mux *http.ServeMux, container di.Container) {
//     wsManager := container.GetWebSocketManager()
//     wsHandler := websocket.NewWebSocketHandler(wsManager)
//
//     mux.HandleFunc("/ws", wsHandler.HandleWebSocketConnection)
//     mux.HandleFunc("/ws/stats", func(w http.ResponseWriter, r *http.Request) {
//         stats := wsHandler.GetConnectionStats()
//         w.Header().Set("Content-Type", "application/json")
//         json.NewEncoder(w).Encode(stats)
//     })
// }
