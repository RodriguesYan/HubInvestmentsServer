package grpc

import (
	"fmt"
	"net"

	balanceGrpc "HubInvestments/internal/balance/presentation/grpc"
	marketDataGrpc "HubInvestments/internal/market_data/presentation/grpc"
	"HubInvestments/internal/market_data/presentation/grpc/interceptors"
	orderGrpc "HubInvestments/internal/order_mngmt_system/presentation/grpc"
	portfolioGrpc "HubInvestments/internal/portfolio_summary/presentation/grpc"
	positionGrpc "HubInvestments/internal/position/presentation/grpc"
	di "HubInvestments/pck"
	"HubInvestments/shared/grpc/proto"

	"google.golang.org/grpc"
)

func NewGRPCServer(container di.Container, port string) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s: %w", port, err)
	}

	authInterceptor := interceptors.NewAuthInterceptor()

	server := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor),
		grpc.StreamInterceptor(authInterceptor.StreamInterceptor),
	)

	// Register Auth Service (existing)
	authServer := NewAuthServiceServer(container)
	proto.RegisterAuthServiceServer(server, authServer)

	// Register new feature-based handlers
	portfolioHandler := portfolioGrpc.NewPortfolioGRPCHandler(container)
	balanceHandler := balanceGrpc.NewBalanceGRPCHandler(container)
	marketDataHandler := marketDataGrpc.NewMarketDataGRPCHandler(container)
	orderHandler := orderGrpc.NewOrderGRPCHandler(container)
	positionHandler := positionGrpc.NewPositionGRPCHandler(container)

	proto.RegisterPortfolioServiceServer(server, portfolioHandler)
	proto.RegisterBalanceServiceServer(server, balanceHandler)
	proto.RegisterMarketDataServiceServer(server, marketDataHandler)
	proto.RegisterOrderServiceServer(server, orderHandler)
	proto.RegisterPositionServiceServer(server, positionHandler)

	return server, lis, nil
}
