package http

import (
	di "HubInvestments/pck"
	"encoding/json"
	"fmt"
	"net/http"
)

type TokenVerifier func(string, http.ResponseWriter) (string, error)

func GetAucAggregation(w http.ResponseWriter, r *http.Request, verifyToken TokenVerifier, container di.Container) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")

	// Authentication
	userId, err := verifyToken(tokenString, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

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
