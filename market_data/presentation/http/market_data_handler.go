package http

import (
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func GetMarketData(w http.ResponseWriter, r *http.Request, container di.Container) {
	symbols := r.URL.Query().Get("symbols")
	arraySymbols := strings.Split(symbols, `,`)

	marketDataList, err := container.GetMarketDataUsecase().Execute(arraySymbols)

	if err != nil {
		http.Error(w, "Failed to get market data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(marketDataList)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(result))
}

func GetMarketDataWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, _ string) {
		GetMarketData(w, r, container)
	})
}
