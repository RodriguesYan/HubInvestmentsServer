package http

import (
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAucAggregation handles position aggregation retrieval for authenticated users
func GetAucAggregation(w http.ResponseWriter, r *http.Request, userId string, container di.Container) {
	// Execute use case
	aucAggregation, err := container.GetPositionAggregationUseCase().Execute(userId)
	if err != nil {
		http.Error(w, "Failed to get position aggregation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Serialize response
	result, err := json.Marshal(aucAggregation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(result))
}

// GetAucAggregationWithAuth returns a handler wrapped with authentication middleware
func GetAucAggregationWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userId string) {
		GetAucAggregation(w, r, userId, container)
	})
}
