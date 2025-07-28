package http

import (
	di "HubInvestments/pck"
	"HubInvestments/shared/middleware"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetWatchlist(w http.ResponseWriter, r *http.Request, container di.Container, usedId string) {
	watchlist, err := container.GetWatchlistUsecase().Execute(usedId)

	if err != nil {
		http.Error(w, "Failed to get watchlist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(watchlist)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(result))
}

func GetWatchlistWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userId string) {
		GetWatchlist(w, r, container, userId)
	})
}
