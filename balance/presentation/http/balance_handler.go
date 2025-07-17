package http

import (
	di "HubInvestments/pck"
	"encoding/json"
	"fmt"
	"net/http"
)

type TokenVerifier func(string, http.ResponseWriter) (string, error)

func GetBalance(w http.ResponseWriter, r *http.Request, verifyToken TokenVerifier, container di.Container) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")

	userId, err := verifyToken(tokenString, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

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
