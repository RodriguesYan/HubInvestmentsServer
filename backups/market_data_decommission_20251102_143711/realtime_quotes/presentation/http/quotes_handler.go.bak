package http

import (
	"HubInvestments/internal/realtime_quotes/domain/service"
	"encoding/json"
	"net/http"
)

type QuotesHandler struct {
	assetDataService *service.AssetDataService
}

func NewQuotesHandler(assetDataService *service.AssetDataService) *QuotesHandler {
	return &QuotesHandler{
		assetDataService: assetDataService,
	}
}

func (h *QuotesHandler) GetAllQuotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	quotes := h.assetDataService.GetAllAssets()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(quotes); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *QuotesHandler) GetStocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stocks := h.assetDataService.GetStocks()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stocks); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *QuotesHandler) GetETFs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	etfs := h.assetDataService.GetETFs()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(etfs); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
