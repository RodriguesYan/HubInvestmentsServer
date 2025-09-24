package grpc

import (
	"fmt"
	"net"

	"HubInvestments/internal/market_data/presentation/grpc/interceptors"
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

	authServer := NewAuthServiceServer(container)
	orderServer := NewOrderServiceServer(container)
	positionServer := NewPositionServiceServer(container)

	proto.RegisterAuthServiceServer(server, authServer)
	proto.RegisterOrderServiceServer(server, orderServer)
	proto.RegisterPositionServiceServer(server, positionServer)

	return server, lis, nil
}
