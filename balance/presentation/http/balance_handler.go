package http

import (
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetBalance handles balance retrieval for authenticated users
//
// Endpoint: GET /getBalance
// Authentication: Bearer token required in Authorization header
// Content-Type: application/json
//
// Request Headers:
// Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
//
// Success Response (200 OK):
//
//	{
//	  "availableBalance": 15000.50
//	}
//
// Error Responses:
// 401 Unauthorized - Missing or invalid token:
//
//	{
//	  "error": "Missing authorization header"
//	}
//
// 500 Internal Server Error - Failed to retrieve balance:
// "Failed to get balance: database connection error"
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
