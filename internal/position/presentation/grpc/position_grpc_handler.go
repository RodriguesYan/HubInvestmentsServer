package grpc

import (
	"context"

	di "HubInvestments/pck"
	"HubInvestments/shared/grpc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PositionGRPCHandler struct {
	proto.UnimplementedPositionServiceServer
	container di.Container
}

func NewPositionGRPCHandler(container di.Container) *PositionGRPCHandler {
	return &PositionGRPCHandler{
		container: container,
	}
}

// GetPositions retrieves all positions for a user
func (h *PositionGRPCHandler) GetPositions(ctx context.Context, req *proto.GetPositionsRequest) (*proto.GetPositionsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Call existing use case
	portfolioSummary, err := h.container.GetPortfolioSummaryUsecase().Execute(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get positions: %v", err)
	}

	// Map domain model to proto response
	var positions []*proto.Position
	for _, category := range portfolioSummary.PositionAggregation.PositionAggregation {
		for _, asset := range category.Assets {
			currentValue := float64(asset.Quantity * asset.LastPrice)
			totalInvested := float64(asset.Quantity * asset.AveragePrice)
			profitLoss := currentValue - totalInvested
			profitLossPercentage := calculateProfitLossPercentage(currentValue, totalInvested)

			positions = append(positions, &proto.Position{
				PositionId:       "",
				UserId:           req.UserId,
				Symbol:           asset.Symbol,
				Quantity:         float64(asset.Quantity),
				AveragePrice:     float64(asset.AveragePrice),
				TotalInvestment:  totalInvested,
				CurrentPrice:     float64(asset.LastPrice),
				MarketValue:      currentValue,
				UnrealizedPnl:    profitLoss,
				UnrealizedPnlPct: profitLossPercentage,
				PositionType:     "LONG",
				Status:           "ACTIVE",
				CreatedAt:        portfolioSummary.LastUpdatedDate,
				UpdatedAt:        portfolioSummary.LastUpdatedDate,
			})
		}
	}

	return &proto.GetPositionsResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "Positions retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		Positions: positions,
	}, nil
}

// GetPositionAggregation retrieves aggregated position data for a user
func (h *PositionGRPCHandler) GetPositionAggregation(ctx context.Context, req *proto.GetPositionAggregationRequest) (*proto.GetPositionAggregationResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Call existing use case
	aggregation, err := h.container.GetPositionAggregationUseCase().Execute(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get position aggregation: %v", err)
	}

	// Map domain model to proto response
	return &proto.GetPositionAggregationResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "Position aggregation retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		Aggregation: &proto.PositionAggregation{
			TotalInvested:         float64(aggregation.TotalInvested),
			TotalCurrentValue:     float64(aggregation.CurrentTotal),
			TotalUnrealizedPnl:    float64(aggregation.CurrentTotal - aggregation.TotalInvested),
			TotalUnrealizedPnlPct: calculateProfitLossPercentage(float64(aggregation.CurrentTotal), float64(aggregation.TotalInvested)),
			TotalPositions:        int32(len(aggregation.PositionAggregation)),
		},
	}, nil
}

// CreatePosition creates a new position (for internal use)
func (h *PositionGRPCHandler) CreatePosition(ctx context.Context, req *proto.CreatePositionRequest) (*proto.CreatePositionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "CreatePosition is for internal use only")
}

// UpdatePosition updates an existing position (for internal use)
func (h *PositionGRPCHandler) UpdatePosition(ctx context.Context, req *proto.UpdatePositionRequest) (*proto.UpdatePositionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdatePosition is for internal use only")
}

// Helper function for percentage calculation
func calculateProfitLossPercentage(currentValue, totalInvested float64) float64 {
	if totalInvested == 0 {
		return 0
	}
	return ((currentValue - totalInvested) / totalInvested) * 100
}
