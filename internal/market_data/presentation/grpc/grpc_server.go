package grpc

import (
	"log"
	"net"

	"HubInvestments/internal/market_data/presentation/grpc/interceptors"
	"HubInvestments/internal/market_data/presentation/grpc/proto"
	di "HubInvestments/pck"

	"google.golang.org/grpc"
)

// StartGRPCServer starts the gRPC server for market data service with authentication
func StartGRPCServer(container di.Container, port string) error {
	// Create listener on the specified port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	// Create authentication interceptor
	authInterceptor := interceptors.NewAuthInterceptor()

	// Create gRPC server with authentication interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor),
		grpc.StreamInterceptor(authInterceptor.StreamInterceptor),
	)

	// Create and register the market data service
	marketDataServer := NewMarketDataGRPCServer(container)
	proto.RegisterMarketDataServiceServer(grpcServer, marketDataServer)

	log.Printf("gRPC server starting on %s", port)

	// Start serving (this blocks)
	return grpcServer.Serve(lis)
}

// StartGRPCServerAsync starts the gRPC server in a separate goroutine
// This allows the main application to continue running the HTTP server
func StartGRPCServerAsync(container di.Container, port string) {
	go func() {
		if err := StartGRPCServer(container, port); err != nil {
			log.Printf("gRPC server failed: %v", err)
		}
	}()
}
