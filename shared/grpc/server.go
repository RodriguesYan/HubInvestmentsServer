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

	authpb "github.com/RodriguesYan/hub-proto-contracts/auth"
	monolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"
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
	authpb.RegisterAuthServiceServer(server, authServer)

	// Register new feature-based handlers
	portfolioHandler := portfolioGrpc.NewPortfolioGRPCHandler(container)
	balanceHandler := balanceGrpc.NewBalanceGRPCHandler(container)
	marketDataHandler := marketDataGrpc.NewMarketDataGRPCHandler(container)
	orderHandler := orderGrpc.NewOrderGRPCHandler(container)
	positionHandler := positionGrpc.NewPositionGRPCHandler(container)

	monolithpb.RegisterPortfolioServiceServer(server, portfolioHandler)
	monolithpb.RegisterBalanceServiceServer(server, balanceHandler)
	monolithpb.RegisterMarketDataServiceServer(server, marketDataHandler)
	monolithpb.RegisterOrderServiceServer(server, orderHandler)
	monolithpb.RegisterPositionServiceServer(server, positionHandler)

	return server, lis, nil
}
