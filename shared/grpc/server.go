package grpc

import (
	"fmt"
	"log"
	"net"

	"HubInvestments/internal/market_data/presentation/grpc/interceptors"
	di "HubInvestments/pck"
	"HubInvestments/shared/grpc/proto"

	"google.golang.org/grpc"
)

// ====================================
// SERVER STARTUP FUNCTIONS
// ====================================

// StartGRPCServer starts the gRPC server with all services
func StartGRPCServer(container di.Container, port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", port, err)
	}

	authInterceptor := interceptors.NewAuthInterceptor()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor),
		grpc.StreamInterceptor(authInterceptor.StreamInterceptor),
	)

	// Create separate service servers
	authServer := NewAuthServiceServer(container)
	orderServer := NewOrderServiceServer(container)
	positionServer := NewPositionServiceServer(container)

	// Register all services with their respective servers
	proto.RegisterAuthServiceServer(grpcServer, authServer)
	proto.RegisterOrderServiceServer(grpcServer, orderServer)
	proto.RegisterPositionServiceServer(grpcServer, positionServer)

	log.Printf("gRPC server starting on %s with all services (Auth, Order, Position)", port)
	log.Printf("Available services:")
	log.Printf("  - AuthService: /hub_investments.AuthService/*")
	log.Printf("  - OrderService: /hub_investments.OrderService/*")
	log.Printf("  - PositionService: /hub_investments.PositionService/*")

	return grpcServer.Serve(lis)
}

// StartGRPCServerAsync starts the gRPC server in a separate goroutine
func StartGRPCServerAsync(container di.Container, port string) {
	go func() {
		if err := StartGRPCServer(container, port); err != nil {
			log.Printf("gRPC server failed: %v", err)
		}
	}()
}
