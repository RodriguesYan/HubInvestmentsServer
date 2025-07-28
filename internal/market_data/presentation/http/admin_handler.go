package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	di "HubInvestments/pck"
	"HubInvestments/shared/middleware"
)

// AdminInvalidateCache handles cache invalidation requests
// @Summary Invalidate Market Data Cache
// @Description Invalidate cached market data for specific symbols (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param symbols query string true "Comma-separated symbols to invalidate (e.g., AAPL,GOOGL)"
// @Success 200 {object} map[string]interface{} "Cache invalidated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/market-data/cache/invalidate [delete]
func AdminInvalidateCache(w http.ResponseWriter, r *http.Request, userId string, container di.Container) {
	// Get symbols from query parameter
	symbolsParam := r.URL.Query().Get("symbols")
	if symbolsParam == "" {
		http.Error(w, "symbols parameter is required", http.StatusBadRequest)
		return
	}

	// Parse and clean symbols
	symbols := strings.Split(symbolsParam, ",")
	cleanSymbols := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		cleaned := strings.TrimSpace(strings.ToUpper(symbol))
		if cleaned != "" {
			cleanSymbols = append(cleanSymbols, cleaned)
		}
	}

	if len(cleanSymbols) == 0 {
		http.Error(w, "no valid symbols provided", http.StatusBadRequest)
		return
	}

	// Invalidate cache
	err := container.InvalidateMarketDataCache(cleanSymbols)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to invalidate cache: %v", err), http.StatusInternalServerError)
		return
	}

	// Success response
	response := map[string]interface{}{
		"message": "Cache invalidated successfully",
		"symbols": cleanSymbols,
		"count":   len(cleanSymbols),
		"status":  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminInvalidateCacheWithAuth returns a handler wrapped with authentication middleware
func AdminInvalidateCacheWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userId string) {
		AdminInvalidateCache(w, r, userId, container)
	})
}

// AdminWarmCache handles cache warming requests
// @Summary Warm Market Data Cache
// @Description Pre-load market data into cache for better performance
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param symbols query string true "Comma-separated symbols to warm (e.g., AAPL,GOOGL)"
// @Success 200 {object} map[string]interface{} "Cache warmed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/market-data/cache/warm [post]
func AdminWarmCache(w http.ResponseWriter, r *http.Request, userId string, container di.Container) {
	// Get symbols from query parameter
	symbolsParam := r.URL.Query().Get("symbols")
	if symbolsParam == "" {
		http.Error(w, "symbols parameter is required", http.StatusBadRequest)
		return
	}

	// Parse and clean symbols
	symbols := strings.Split(symbolsParam, ",")
	cleanSymbols := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		cleaned := strings.TrimSpace(strings.ToUpper(symbol))
		if cleaned != "" {
			cleanSymbols = append(cleanSymbols, cleaned)
		}
	}

	if len(cleanSymbols) == 0 {
		http.Error(w, "no valid symbols provided", http.StatusBadRequest)
		return
	}

	// Warm cache
	err := container.WarmMarketDataCache(cleanSymbols)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to warm cache: %v", err), http.StatusInternalServerError)
		return
	}

	// Success response
	response := map[string]interface{}{
		"message": "Cache warmed successfully",
		"symbols": cleanSymbols,
		"count":   len(cleanSymbols),
		"status":  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminWarmCacheWithAuth returns a handler wrapped with authentication middleware
func AdminWarmCacheWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userId string) {
		AdminWarmCache(w, r, userId, container)
	})
}
