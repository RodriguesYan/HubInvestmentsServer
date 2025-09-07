package websocket

import (
	"HubInvestments/internal/auth"
	"HubInvestments/internal/realtime_quotes/application/service"
	"HubInvestments/internal/realtime_quotes/domain/model"
	"HubInvestments/shared/infra/websocket"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type RealtimeQuotesWebSocketHandler struct {
	wsManager               websocket.WebSocketManager
	priceOscillationService *service.PriceOscillationService
	authService             auth.IAuthService
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
	authService auth.IAuthService,
) *RealtimeQuotesWebSocketHandler {
	handler := &RealtimeQuotesWebSocketHandler{
		wsManager:               wsManager,
		priceOscillationService: priceOscillationService,
		authService:             authService,
		activeConnections:       make(map[websocket.Websocket]bool),
	}

	handler.startPriceSubscription()
	return handler
}

func (h *RealtimeQuotesWebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Authenticate BEFORE WebSocket upgrade (same pattern as HTTP handlers)
	tokenString := h.extractToken(r)
	log.Printf("WebSocket Debug: Extracted token: '%s'", h.safeTokenLog(tokenString))

	if tokenString == "" {
		log.Printf("WebSocket Debug: No token provided - rejecting connection")
		http.Error(w, "Unauthorized - missing token", http.StatusUnauthorized)
		return
	}

	userId, err := h.authService.VerifyToken(tokenString, w)
	if err != nil {
		log.Printf("WebSocket Debug: Authentication failed with error: %v", err)
		// authService.VerifyToken already wrote the HTTP error response
		return
	}

	log.Printf("WebSocket Debug: Authentication successful for user: %s", userId)

	// Only upgrade to WebSocket if authentication succeeded
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
	log.Printf("New authenticated realtime quotes WebSocket connection for user %s. Active: %d, Health: %s",
		userId, len(h.activeConnections), healthStatus.String())

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

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			} else {
				log.Printf("WebSocket connection closed normally: %v", err)
			}
			// Clean up connection and exit - client will reconnect if needed
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

// extractToken extracts JWT token from Authorization header or query parameter
func (h *RealtimeQuotesWebSocketHandler) extractToken(r *http.Request) string {
	// Try Authorization header first (preferred method)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		return authHeader
	}

	// Fallback to query parameter for WebSocket clients that can't set headers
	token := r.URL.Query().Get("token")
	if token != "" {
		return "Bearer " + token
	}

	return ""
}

// safeTokenLog safely logs token for debugging (shows first part only)
func (h *RealtimeQuotesWebSocketHandler) safeTokenLog(token string) string {
	if token == "" {
		return "<empty>"
	}
	if len(token) <= 50 {
		return token
	}
	return token[:50] + "..."
}
