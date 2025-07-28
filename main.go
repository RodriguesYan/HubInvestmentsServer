// @title HubInvestments API
// @version 1.0
// @description HubInvestments is a comprehensive financial investment platform API that provides portfolio management, market data access, and user authentication.
// @contact.name HubInvestments Development Team
// @contact.email support@hubinvestments.com
// @host 192.168.0.3:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	_ "HubInvestments/docs"
	"HubInvestments/internal/auth"
	"HubInvestments/internal/auth/token"
	balanceHandler "HubInvestments/internal/balance/presentation/http"
	"HubInvestments/internal/login"
	grpcHandler "HubInvestments/internal/market_data/presentation/grpc"
	adminHandler "HubInvestments/internal/market_data/presentation/http"
	marketDataHandler "HubInvestments/internal/market_data/presentation/http"
	portfolioSummaryHandler "HubInvestments/internal/portfolio_summary/presentation/http"
	positionHandler "HubInvestments/internal/position/presentation/http"
	watchlistHandler "HubInvestments/internal/watchlist/presentation/http"
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	"log"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

// const portNum string = "localhost:8080"
const portNum string = "192.168.0.3:8080" //My home IP
// const portNum string = "192.168.0.48:8080" //Camila's home IP

const grpcPortNum string = "192.168.0.6:50051" // gRPC server port

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

	// Start gRPC server in background
	grpcHandler.StartGRPCServerAsync(container, grpcPortNum)
	log.Printf("gRPC server will start on %s", grpcPortNum)

	// API Routes
	http.HandleFunc("/login", login.Login)
	http.HandleFunc("/getAucAggregation", positionHandler.GetAucAggregationWithAuth(verifyToken, container))
	http.HandleFunc("/getBalance", balanceHandler.GetBalanceWithAuth(verifyToken, container))
	http.HandleFunc("/getPortfolioSummary", portfolioSummaryHandler.GetPortfolioSummaryWithAuth(verifyToken, container))
	http.HandleFunc("/getMarketData", marketDataHandler.GetMarketDataWithAuth(verifyToken, container))
	http.HandleFunc("/getWatchlist", watchlistHandler.GetWatchlistWithAuth(verifyToken, container))

	// Admin Routes for Cache Management
	http.HandleFunc("/admin/market-data/cache/invalidate", adminHandler.AdminInvalidateCacheWithAuth(verifyToken, container))
	http.HandleFunc("/admin/market-data/cache/warm", adminHandler.AdminWarmCacheWithAuth(verifyToken, container))

	// Swagger documentation route
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Printf("HTTP server starting on %s", portNum)
	log.Printf("Admin cache endpoints available:")
	log.Printf("  DELETE http://%s/admin/market-data/cache/invalidate?symbols=AAPL,GOOGL", portNum)
	log.Printf("  POST   http://%s/admin/market-data/cache/warm?symbols=AAPL,GOOGL", portNum)
	log.Printf("Swagger documentation available at: http://%s/swagger/index.html", portNum)

	err = http.ListenAndServe(portNum, nil)
	if err != nil {
		log.Fatal(err)
	}
}
