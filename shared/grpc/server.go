package grpc

import (
	"context"
	"fmt"
	"net"

	balanceGrpc "HubInvestments/internal/balance/presentation/grpc"
	marketDataGrpc "HubInvestments/internal/market_data/presentation/grpc"
	orderGrpc "HubInvestments/internal/order_mngmt_system/presentation/grpc"
	portfolioGrpc "HubInvestments/internal/portfolio_summary/presentation/grpc"
	positionGrpc "HubInvestments/internal/position/presentation/grpc"
	di "HubInvestments/pck"

	authpb "github.com/RodriguesYan/hub-proto-contracts/auth"
	monolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewGRPCServer(container di.Container, port string) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s: %w", port, err)
	}

	// Use a simple context interceptor that trusts the API Gateway
	// The gateway has already validated authentication and forwards user context
	contextInterceptor := newGatewayContextInterceptor()

	server := grpc.NewServer(
		grpc.UnaryInterceptor(contextInterceptor.unaryInterceptor),
		grpc.StreamInterceptor(contextInterceptor.streamInterceptor),
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

// gatewayContextInterceptor extracts user context from API Gateway metadata
// This interceptor trusts that the API Gateway has already validated authentication
type gatewayContextInterceptor struct{}

func newGatewayContextInterceptor() *gatewayContextInterceptor {
	return &gatewayContextInterceptor{}
}

// unaryInterceptor extracts user ID from gateway metadata and adds it to context
func (i *gatewayContextInterceptor) unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// Get user ID from gateway metadata (x-user-id)
		if userIDs := md.Get("x-user-id"); len(userIDs) > 0 {
			ctx = context.WithValue(ctx, "userId", userIDs[0])
		}

		// Get user email from gateway metadata (x-user-email)
		if emails := md.Get("x-user-email"); len(emails) > 0 {
			ctx = context.WithValue(ctx, "userEmail", emails[0])
		}
	}

	// Continue with the request
	return handler(ctx, req)
}

// streamInterceptor extracts user ID from gateway metadata for streaming calls
func (i *gatewayContextInterceptor) streamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// Extract metadata from stream context
	ctx := stream.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// Get user ID from gateway metadata
		if userIDs := md.Get("x-user-id"); len(userIDs) > 0 {
			ctx = context.WithValue(ctx, "userId", userIDs[0])
		}

		// Get user email from gateway metadata
		if emails := md.Get("x-user-email"); len(emails) > 0 {
			ctx = context.WithValue(ctx, "userEmail", emails[0])
		}
	}

	// Create wrapped stream with updated context
	wrappedStream := &wrappedServerStream{
		ServerStream: stream,
		ctx:          ctx,
	}

	// Continue with the request
	return handler(srv, wrappedStream)
}

// wrappedServerStream wraps grpc.ServerStream to override context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
