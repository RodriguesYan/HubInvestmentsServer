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
	doLoginHandler "HubInvestments/internal/login/presentation/http"
	adminHandler "HubInvestments/internal/market_data/presentation/http"
	marketDataHandler "HubInvestments/internal/market_data/presentation/http"
	orderHandler "HubInvestments/internal/order_mngmt_system/presentation/http"
	portfolioSummaryHandler "HubInvestments/internal/portfolio_summary/presentation/http"
	positionHandler "HubInvestments/internal/position/presentation/http"
	_ "HubInvestments/internal/realtime_quotes/presentation/http"
	watchlistHandler "HubInvestments/internal/watchlist/presentation/http"
	di "HubInvestments/pck"
	"HubInvestments/shared/config"
	grpcServer "HubInvestments/shared/grpc"
	"HubInvestments/shared/middleware"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Load configuration once at startup
	cfg := config.Load()

	tokenService := token.NewTokenService()
	aucService := auth.NewAuthService(tokenService)

	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		return aucService.VerifyToken(token, w)
	})

	container, err := di.NewContainer()
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv, lis, err := grpcServer.NewGRPCServer(container, cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	// API Routes
	// http.HandleFunc("/login", login.Login)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		doLoginHandler.DoLogin(w, r, container)
	})
	http.HandleFunc("/getAucAggregation", positionHandler.GetAucAggregationWithAuth(verifyToken, container))
	http.HandleFunc("/getBalance", balanceHandler.GetBalanceWithAuth(verifyToken, container))
	http.HandleFunc("/getPortfolioSummary", portfolioSummaryHandler.GetPortfolioSummaryWithAuth(verifyToken, container))
	http.HandleFunc("/getMarketData", marketDataHandler.GetMarketDataWithAuth(verifyToken, container))
	http.HandleFunc("/getWatchlist", watchlistHandler.GetWatchlistWithAuth(verifyToken, container))

	// Order Management Routes
	http.HandleFunc("/orders", orderHandler.SubmitOrderWithAuth(verifyToken, container))
	http.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/status") {
			orderHandler.GetOrderStatusWithAuth(verifyToken, container)(w, r)
		} else if strings.HasSuffix(path, "/cancel") {
			orderHandler.CancelOrderWithAuth(verifyToken, container)(w, r)
		} else {
			orderHandler.GetOrderDetailsWithAuth(verifyToken, container)(w, r)
		}
	})
	http.HandleFunc("/orders/history", orderHandler.GetOrderHistoryWithAuth(verifyToken, container))

	// Admin Routes for Cache Management
	http.HandleFunc("/admin/market-data/cache/invalidate", adminHandler.AdminInvalidateCacheWithAuth(verifyToken, container))
	http.HandleFunc("/admin/market-data/cache/warm", adminHandler.AdminWarmCacheWithAuth(verifyToken, container))

	// Realtime Quotes Routes
	http.HandleFunc("/quotes", func(w http.ResponseWriter, r *http.Request) {
		quotesHandler := container.GetQuotesHandler()
		quotesHandler.GetAllQuotes(w, r)
	})
	http.HandleFunc("/quotes/stocks", func(w http.ResponseWriter, r *http.Request) {
		quotesHandler := container.GetQuotesHandler()
		quotesHandler.GetStocks(w, r)
	})
	http.HandleFunc("/quotes/etfs", func(w http.ResponseWriter, r *http.Request) {
		quotesHandler := container.GetQuotesHandler()
		quotesHandler.GetETFs(w, r)
	})

	// WebSocket Routes for Realtime Quotes
	http.HandleFunc("/ws/quotes", func(w http.ResponseWriter, r *http.Request) {
		wsHandler := container.GetRealtimeQuotesWebSocketHandler()
		wsHandler.HandleConnection(w, r)
	})

	// Swagger documentation route
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	go func() {
		log.Printf("gRPC server starting on %s", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	httpSrv := &http.Server{Addr: cfg.HTTPPort}
	go func() {
		log.Printf("HTTP server starting on %s", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcSrv.GracefulStop()
	httpSrv.Shutdown(ctx)
}
