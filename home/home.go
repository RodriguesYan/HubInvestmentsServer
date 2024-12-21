package home

import (
	"fmt"
	"net/http"

	"HubInvestments/auth"
)

func GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")

	err := auth.VerifyToken(tokenString, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	//Fazer query pra trazer saldo do usuario
	fmt.Fprint(w, 1000)
}
