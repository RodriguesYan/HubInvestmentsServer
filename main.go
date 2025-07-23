// @title HubInvestments API
// @version 1.0
// @description HubInvestments is a comprehensive financial investment platform API that provides portfolio management, market data access, and user authentication.
// @contact.name HubInvestments Development Team
// @contact.email support@hubinvestments.com
// @host 192.168.0.6:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"HubInvestments/auth"
	"HubInvestments/auth/token"
	balanceHandler "HubInvestments/balance/presentation/http"
	_ "HubInvestments/docs"
	"HubInvestments/login"
	marketDataHandler "HubInvestments/market_data/presentation"
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	portfolioSummaryHandler "HubInvestments/portfolio_summary/presentation/http"
	positionHandler "HubInvestments/position/presentation/http"
	"log"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
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

	// API Routes
	http.HandleFunc("/login", login.Login)
	http.HandleFunc("/getAucAggregation", positionHandler.GetAucAggregationWithAuth(verifyToken, container))
	http.HandleFunc("/getBalance", balanceHandler.GetBalanceWithAuth(verifyToken, container))
	http.HandleFunc("/getPortfolioSummary", portfolioSummaryHandler.GetPortfolioSummaryWithAuth(verifyToken, container))
	http.HandleFunc("/getMarketData", marketDataHandler.GetMarketDataWithAuth(verifyToken, container))

	// Swagger documentation route
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Printf("Server starting on %s", portNum)
	log.Printf("Swagger documentation available at: http://%s/swagger/index.html", portNum)

	err = http.ListenAndServe(portNum, nil)
	if err != nil {
		log.Fatal(err)
	}
}
