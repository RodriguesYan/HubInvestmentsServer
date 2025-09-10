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
	connectionStates        map[websocket.Websocket]*ConnectionState
	mu                      sync.RWMutex
}

// JSON Patch operation structure according to RFC 6902
type PatchOperation struct {
	Op    string      `json:"op"`    // "add" or "replace"
	Path  string      `json:"path"`  // JSON pointer path
	Value interface{} `json:"value"` // The value to add/replace
}

type QuotePatchMessage struct {
	Type       string           `json:"type"`
	Operations []PatchOperation `json:"operations"`
}

// Connection state to track what quotes each connection has received
type ConnectionState struct {
	conn          websocket.Websocket
	lastQuotes    map[string]*model.AssetQuote
	isInitialized bool
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
		connectionStates:        make(map[websocket.Websocket]*ConnectionState),
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
	h.connectionStates[conn] = &ConnectionState{
		conn:          conn,
		lastQuotes:    make(map[string]*model.AssetQuote),
		isInitialized: false,
	}
	h.mu.Unlock()

	healthStatus := h.wsManager.GetHealthStatus()
	log.Printf("New authenticated realtime quotes WebSocket connection for user %s. Active: %d, Health: %s",
		userId, len(h.connectionStates), healthStatus.String())

	// Send initial quotes with add operations for new connection
	h.sendInitialQuotes(conn)

	go h.handleConnection(conn)
}

func (h *RealtimeQuotesWebSocketHandler) handleConnection(conn websocket.Websocket) {
	defer func() {
		h.mu.Lock()
		delete(h.connectionStates, conn)
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

// sendInitialQuotes sends add operations for all available quotes when connection is established
func (h *RealtimeQuotesWebSocketHandler) sendInitialQuotes(conn websocket.Websocket) {
	// Get all available quotes from the service
	allQuotes := h.priceOscillationService.GetAllQuotes()

	operations := make([]PatchOperation, 0, len(allQuotes))

	for symbol, quote := range allQuotes {
		operations = append(operations, PatchOperation{
			Op:    "add",
			Path:  "/quotes/" + symbol,
			Value: quote,
		})
	}

	message := QuotePatchMessage{
		Type:       "quotes_patch",
		Operations: operations,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling initial quotes patch message: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending initial quotes to WebSocket connection: %v", err)
		return
	}

	// Update connection state
	h.mu.Lock()
	if state, exists := h.connectionStates[conn]; exists {
		for symbol, quote := range allQuotes {
			state.lastQuotes[symbol] = h.copyQuote(quote)
		}
		state.isInitialized = true
	}
	h.mu.Unlock()

	log.Printf("Sent initial quotes with %d add operations to new connection", len(operations))
}

// broadcastQuotes sends replace operations for changed quotes to existing connections
func (h *RealtimeQuotesWebSocketHandler) broadcastQuotes(quotes map[string]*model.AssetQuote) {
	h.mu.RLock()
	connectionStates := make([]*ConnectionState, 0, len(h.connectionStates))
	for _, state := range h.connectionStates {
		connectionStates = append(connectionStates, state)
	}
	h.mu.RUnlock()

	for _, state := range connectionStates {
		if !state.isInitialized {
			continue
		}

		operations := h.generateReplaceOperations(state, quotes)
		if len(operations) == 0 {
			continue
		}

		message := QuotePatchMessage{
			Type:       "quotes_patch",
			Operations: operations,
		}

		data, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling quotes patch message: %v", err)
			continue
		}

		if err := state.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error broadcasting patch to WebSocket connection: %v", err)
			continue
		}

		// Update connection's last known state
		h.updateConnectionState(state, quotes)
	}
}

func (h *RealtimeQuotesWebSocketHandler) GetActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connectionStates)
}

// generateReplaceOperations generates replace operations for quotes that have changed
func (h *RealtimeQuotesWebSocketHandler) generateReplaceOperations(state *ConnectionState, newQuotes map[string]*model.AssetQuote) []PatchOperation {
	operations := make([]PatchOperation, 0)

	for symbol, newQuote := range newQuotes {
		lastQuote, exists := state.lastQuotes[symbol]

		// If quote doesn't exist in connection state, add it as a new quote
		if !exists {
			operations = append(operations, PatchOperation{
				Op:    "add",
				Path:  "/quotes/" + symbol,
				Value: newQuote,
			})
			continue
		}

		// Check for changes in individual fields and create replace operations
		if !h.quotesEqual(lastQuote, newQuote) {
			// Check specific fields that commonly change
			if lastQuote.CurrentPrice != newQuote.CurrentPrice {
				operations = append(operations, PatchOperation{
					Op:    "replace",
					Path:  "/quotes/" + symbol + "/current_price",
					Value: newQuote.CurrentPrice,
				})
			}

			if lastQuote.Change != newQuote.Change {
				operations = append(operations, PatchOperation{
					Op:    "replace",
					Path:  "/quotes/" + symbol + "/change",
					Value: newQuote.Change,
				})
			}

			if lastQuote.ChangePercent != newQuote.ChangePercent {
				operations = append(operations, PatchOperation{
					Op:    "replace",
					Path:  "/quotes/" + symbol + "/change_percent",
					Value: newQuote.ChangePercent,
				})
			}

			if lastQuote.LastUpdated != newQuote.LastUpdated {
				operations = append(operations, PatchOperation{
					Op:    "replace",
					Path:  "/quotes/" + symbol + "/last_updated",
					Value: newQuote.LastUpdated,
				})
			}
		}
	}

	return operations
}

// updateConnectionState updates the connection's last known state with new quotes
func (h *RealtimeQuotesWebSocketHandler) updateConnectionState(state *ConnectionState, newQuotes map[string]*model.AssetQuote) {
	for symbol, quote := range newQuotes {
		state.lastQuotes[symbol] = h.copyQuote(quote)
	}
}

func (h *RealtimeQuotesWebSocketHandler) copyQuote(quote *model.AssetQuote) *model.AssetQuote {
	if quote == nil {
		return nil
	}

	copied := *quote
	return &copied
}

// quotesEqual compares two AssetQuote objects for equality
func (h *RealtimeQuotesWebSocketHandler) quotesEqual(quote1, quote2 *model.AssetQuote) bool {
	if quote1 == nil && quote2 == nil {
		return true
	}
	if quote1 == nil || quote2 == nil {
		return false
	}

	return quote1.CurrentPrice == quote2.CurrentPrice &&
		quote1.Change == quote2.Change &&
		quote1.ChangePercent == quote2.ChangePercent &&
		quote1.LastUpdated == quote2.LastUpdated
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
