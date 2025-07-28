package http

import (
	di "HubInvestments/pck"
	"HubInvestments/shared/middleware"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetPortfolioSummary handles portfolio summary retrieval for authenticated users
// @Summary Get Portfolio Summary
// @Description Retrieve complete portfolio summary including balance, total portfolio value, and position aggregation
// @Tags Portfolio
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.PortfolioSummaryResponse "Portfolio summary retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /getPortfolioSummary [get]
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
