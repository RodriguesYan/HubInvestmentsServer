package main

import (
	"HubInvestments/auth"
	"HubInvestments/auth/token"
	balanceHandler "HubInvestments/balance/presentation/http"
	"HubInvestments/login"
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	portfolioSummaryHandler "HubInvestments/portfolio_summary/presentation/http"
	positionHandler "HubInvestments/position/presentation/http"
	"log"
	"net/http"
)

// const portNum string = "localhost:8080"
const portNum string = "192.168.0.6:8080" //My home IP
// const portNum string = "192.168.0.48:8080" //Camila's home IP

func main() {
	tokenService := token.NewTokenService()
	aucService := auth.NewAuthService(tokenService)

	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		return aucService.VerifyToken(token, w)
	})

	container, err := di.NewContainer()

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/login", login.Login)
	http.HandleFunc("/getAucAggregation", positionHandler.GetAucAggregationWithAuth(verifyToken, container))
	http.HandleFunc("/getBalance", balanceHandler.GetBalanceWithAuth(verifyToken, container))
	http.HandleFunc("/getPortfolioSummary", portfolioSummaryHandler.GetPortfolioSummaryWithAuth(verifyToken, container))

	err = http.ListenAndServe(portNum, nil)
	if err != nil {
		log.Fatal(err)
	}
}
