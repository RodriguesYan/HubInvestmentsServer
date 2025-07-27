package grpc

import (
	"context"
	"fmt"

	"HubInvestments/internal/market_data/presentation/grpc/proto"
	di "HubInvestments/pck"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MarketDataGRPCServer implements the gRPC server for market data service
type MarketDataGRPCServer struct {
	proto.UnimplementedMarketDataServiceServer
	container di.Container
}

// NewMarketDataGRPCServer creates a new gRPC server for market data service
func NewMarketDataGRPCServer(container di.Container) *MarketDataGRPCServer {
	return &MarketDataGRPCServer{
		container: container,
	}
}

// GetMarketData implements the gRPC method for retrieving market data
func (s *MarketDataGRPCServer) GetMarketData(ctx context.Context, req *proto.GetMarketDataRequest) (*proto.GetMarketDataResponse, error) {
	// Validate request
	if len(req.Symbols) == 0 {
		return nil, status.Error(codes.InvalidArgument, "symbols list cannot be empty")
	}

	// Use the same use case as HTTP handler - shared business logic
	useCase := s.container.GetMarketDataUsecase()

	// Execute the use case
	marketDataList, err := useCase.Execute(req.Symbols)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get market data: %v", err))
	}

	// Convert domain models to gRPC response
	var protoMarketData []*proto.MarketData
	for _, data := range marketDataList {
		protoMarketData = append(protoMarketData, &proto.MarketData{
			Symbol:    data.Symbol,
			Name:      data.Name,
			LastQuote: data.LastQuote,
			Category:  int32(data.Category),
		})
	}

	return &proto.GetMarketDataResponse{
		MarketData: protoMarketData,
	}, nil
}

// StreamMarketData implements the streaming gRPC method (placeholder for future implementation)
func (s *MarketDataGRPCServer) StreamMarketData(req *proto.StreamMarketDataRequest, stream proto.MarketDataService_StreamMarketDataServer) error {
	// TODO: Implement streaming market data
	// This is a placeholder for future real-time streaming functionality
	return status.Error(codes.Unimplemented, "streaming market data not implemented yet")
}
