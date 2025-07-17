package http

import (
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetPortfolioSummary handles portfolio summary retrieval for authenticated users
func GetPortfolioSummary(w http.ResponseWriter, r *http.Request, userId string, container di.Container) {
	aggregation, err := container.GetPortfolioSummaryUsecase().Execute(userId)

	if err != nil {
		http.Error(w, "Failed to get portfolio summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(aggregation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(result))
}

// GetPortfolioSummaryWithAuth returns a handler wrapped with authentication middleware
func GetPortfolioSummaryWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userId string) {
		GetPortfolioSummary(w, r, userId, container)
	})
}
