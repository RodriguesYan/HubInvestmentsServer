package websocket

import (
	"HubInvestments/internal/realtime_quotes/application/service"
	"HubInvestments/internal/realtime_quotes/domain/model"
	"HubInvestments/shared/infra/websocket"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type RealtimeQuotesWebSocketHandler struct {
	wsManager               websocket.WebSocketManager
	priceOscillationService *service.PriceOscillationService
	activeConnections       map[websocket.Websocket]bool
	mu                      sync.RWMutex
}

type QuoteMessage struct {
	Type   string                       `json:"type"`
	Quotes map[string]*model.AssetQuote `json:"quotes"`
}

func NewRealtimeQuotesWebSocketHandler(
	wsManager websocket.WebSocketManager,
	priceOscillationService *service.PriceOscillationService,
) *RealtimeQuotesWebSocketHandler {
	handler := &RealtimeQuotesWebSocketHandler{
		wsManager:               wsManager,
		priceOscillationService: priceOscillationService,
		activeConnections:       make(map[websocket.Websocket]bool),
	}

	handler.startPriceSubscription()
	return handler
}

func (h *RealtimeQuotesWebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := h.wsManager.CreateConnection(w, r)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)

		// Check if it's a connection limit error and provide better error message
		if err.Error() == "maximum connections limit reached" {
			http.Error(w, "Server at capacity, please try again later", http.StatusServiceUnavailable)
		} else {
			http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		}
		return
	}

	h.mu.Lock()
	h.activeConnections[conn] = true
	h.mu.Unlock()

	healthStatus := h.wsManager.GetHealthStatus()
	log.Printf("New realtime quotes WebSocket connection. Active: %d, Health: %s",
		len(h.activeConnections), healthStatus.String())

	go h.handleConnection(conn)
}

func (h *RealtimeQuotesWebSocketHandler) handleConnection(conn websocket.Websocket) {
	defer func() {
		h.mu.Lock()
		delete(h.activeConnections, conn)
		h.mu.Unlock()

		if err := conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}()

	connectionID := fmt.Sprintf("quotes_conn_%d", time.Now().UnixNano())

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
				// Schedule reconnection for unexpected errors
				if reconnectErr := h.wsManager.ScheduleReconnection(connectionID, "unexpected_close"); reconnectErr != nil {
					log.Printf("Failed to schedule reconnection: %v", reconnectErr)
				}
			} else {
				log.Printf("WebSocket connection closed normally: %v", err)
			}
			break
		}
		// For realtime quotes, we only broadcast data, no need to process incoming messages
	}
}

func (h *RealtimeQuotesWebSocketHandler) startPriceSubscription() {
	priceUpdates := h.priceOscillationService.Subscribe()

	go func() {
		for quotes := range priceUpdates {
			h.broadcastQuotes(quotes)
		}
	}()
}

func (h *RealtimeQuotesWebSocketHandler) broadcastQuotes(quotes map[string]*model.AssetQuote) {
	message := QuoteMessage{
		Type:   "price_update",
		Quotes: quotes,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling quote message: %v", err)
		return
	}

	h.mu.RLock()
	connections := make([]websocket.Websocket, 0, len(h.activeConnections))
	for conn := range h.activeConnections {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()

	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error broadcasting to WebSocket connection: %v", err)
			// Connection will be cleaned up by handleConnection goroutine
		}
	}
}

func (h *RealtimeQuotesWebSocketHandler) GetActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.activeConnections)
}
