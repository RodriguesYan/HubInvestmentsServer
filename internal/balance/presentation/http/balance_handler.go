package http

import (
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetBalance handles balance retrieval for authenticated users
// @Summary Get User Balance
// @Description Retrieve the available balance for the authenticated user
// @Tags Balance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BalanceResponse "Balance retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /getBalance [get]
func GetBalance(w http.ResponseWriter, r *http.Request, userId string, container di.Container) {
	balance, err := container.GetBalanceUseCase().Execute(userId)

	if err != nil {
		http.Error(w, "Failed to get balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(result))
}

// GetBalanceWithAuth returns a handler wrapped with authentication middleware
func GetBalanceWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userId string) {
		GetBalance(w, r, userId, container)
	})
}
