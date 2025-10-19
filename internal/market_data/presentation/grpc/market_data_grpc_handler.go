package grpc

import (
	"context"

	di "HubInvestments/pck"
	"HubInvestments/shared/grpc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MarketDataGRPCHandler struct {
	proto.UnimplementedMarketDataServiceServer
	container di.Container
}

func NewMarketDataGRPCHandler(container di.Container) *MarketDataGRPCHandler {
	return &MarketDataGRPCHandler{
		container: container,
	}
}

// GetMarketData retrieves market data for a specific symbol
func (h *MarketDataGRPCHandler) GetMarketData(ctx context.Context, req *proto.GetMarketDataRequest) (*proto.GetMarketDataResponse, error) {
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	// Call existing use case
	marketDataList, err := h.container.GetMarketDataUsecase().Execute([]string{req.Symbol})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get market data: %v", err)
	}

	if len(marketDataList) == 0 {
		return nil, status.Error(codes.NotFound, "market data not found for symbol")
	}

	marketData := marketDataList[0]

	// Map domain model to proto response
	return &proto.GetMarketDataResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "Market data retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		MarketData: &proto.MarketData{
			Symbol:        marketData.Symbol,
			CompanyName:   marketData.Name,
			CurrentPrice:  float64(marketData.LastQuote),
			PreviousClose: 0,
			OpenPrice:     0,
			HighPrice:     0,
			LowPrice:      0,
			Change:        0,
			ChangePercent: 0,
			Volume:        0,
			LastUpdated:   "",
		},
	}, nil
}

// GetAssetDetails retrieves detailed asset information
func (h *MarketDataGRPCHandler) GetAssetDetails(ctx context.Context, req *proto.GetAssetDetailsRequest) (*proto.GetAssetDetailsResponse, error) {
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	// Call existing use case
	marketDataList, err := h.container.GetMarketDataUsecase().Execute([]string{req.Symbol})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get asset details: %v", err)
	}

	if len(marketDataList) == 0 {
		return nil, status.Error(codes.NotFound, "asset not found")
	}

	marketData := marketDataList[0]

	// Map domain model to proto response
	return &proto.GetAssetDetailsResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "Asset details retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		Asset: &proto.AssetDetails{
			Symbol:           marketData.Symbol,
			CompanyName:      marketData.Name,
			Sector:           "",
			Industry:         "",
			Description:      "",
			MarketCap:        0,
			PeRatio:          0,
			DividendYield:    0,
			FiftyTwoWeekHigh: 0,
			FiftyTwoWeekLow:  0,
			Currency:         "USD",
			Exchange:         "",
		},
	}, nil
}

// GetBatchMarketData retrieves market data for multiple symbols
func (h *MarketDataGRPCHandler) GetBatchMarketData(ctx context.Context, req *proto.GetBatchMarketDataRequest) (*proto.GetBatchMarketDataResponse, error) {
	if len(req.Symbols) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one symbol is required")
	}

	// Call existing use case
	marketDataList, err := h.container.GetMarketDataUsecase().Execute(req.Symbols)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get batch market data: %v", err)
	}

	// Map domain models to proto response
	protoMarketData := make([]*proto.MarketData, len(marketDataList))
	for i, md := range marketDataList {
		protoMarketData[i] = &proto.MarketData{
			Symbol:        md.Symbol,
			CompanyName:   md.Name,
			CurrentPrice:  float64(md.LastQuote),
			PreviousClose: 0,
			OpenPrice:     0,
			HighPrice:     0,
			LowPrice:      0,
			Change:        0,
			ChangePercent: 0,
			Volume:        0,
			LastUpdated:   "",
		}
	}

	return &proto.GetBatchMarketDataResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "Batch market data retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		MarketData: protoMarketData,
	}, nil
}
