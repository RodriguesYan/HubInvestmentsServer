package grpc

import (
	"context"

	di "HubInvestments/pck"
	"HubInvestments/shared/grpc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PortfolioGRPCHandler struct {
	proto.UnimplementedPortfolioServiceServer
	container di.Container
}

func NewPortfolioGRPCHandler(container di.Container) *PortfolioGRPCHandler {
	return &PortfolioGRPCHandler{
		container: container,
	}
}

// GetPortfolioSummary retrieves complete portfolio summary for a user
func (h *PortfolioGRPCHandler) GetPortfolioSummary(ctx context.Context, req *proto.GetPortfolioSummaryRequest) (*proto.GetPortfolioSummaryResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Call existing use case (same as HTTP handler)
	portfolioSummary, err := h.container.GetPortfolioSummaryUsecase().Execute(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get portfolio summary: %v", err)
	}

	// Map domain model to proto response
	return &proto.GetPortfolioSummaryResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "Portfolio summary retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		Portfolio: &proto.PortfolioSummary{
			TotalBalance:              float64(portfolioSummary.Balance.AvailableBalance),
			TotalInvested:             float64(portfolioSummary.PositionAggregation.TotalInvested),
			TotalCurrentValue:         float64(portfolioSummary.PositionAggregation.CurrentTotal),
			TotalProfitLoss:           float64(portfolioSummary.PositionAggregation.CurrentTotal - portfolioSummary.PositionAggregation.TotalInvested),
			TotalProfitLossPercentage: calculateProfitLossPercentage(float64(portfolioSummary.PositionAggregation.CurrentTotal), float64(portfolioSummary.PositionAggregation.TotalInvested)),
			Positions:                 mapPositionsToProto(req.UserId, portfolioSummary),
			LastUpdated:               portfolioSummary.LastUpdatedDate,
		},
	}, nil
}

// Helper function to map positions
func mapPositionsToProto(userId string, portfolioSummary interface{}) []*proto.PortfolioPosition {
	// This is just a placeholder - the actual mapping should be done by a mapper
	// For now, return empty to avoid business logic in handler
	return []*proto.PortfolioPosition{}
}

// Helper function for percentage calculation
func calculateProfitLossPercentage(currentValue, totalInvested float64) float64 {
	if totalInvested == 0 {
		return 0
	}
	return ((currentValue - totalInvested) / totalInvested) * 100
}
