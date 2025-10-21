package grpc

import (
	"context"

	di "HubInvestments/pck"

	commonpb "github.com/RodriguesYan/hub-proto-contracts/common"
	monolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BalanceGRPCHandler struct {
	monolithpb.UnimplementedBalanceServiceServer
	container di.Container
}

func NewBalanceGRPCHandler(container di.Container) *BalanceGRPCHandler {
	return &BalanceGRPCHandler{
		container: container,
	}
}

// GetBalance retrieves user balance
func (h *BalanceGRPCHandler) GetBalance(ctx context.Context, req *monolithpb.GetBalanceRequest) (*monolithpb.GetBalanceResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Call existing use case (same as HTTP handler)
	balance, err := h.container.GetBalanceUseCase().Execute(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	// Map domain model to proto response
	return &monolithpb.GetBalanceResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   true,
			Message:   "Balance retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		Balance: &monolithpb.Balance{
			UserId:           req.UserId,
			AvailableBalance: float64(balance.AvailableBalance),
			TotalBalance:     float64(balance.AvailableBalance),
			ReservedBalance:  0,
			Currency:         "USD",
			LastUpdated:      "",
		},
	}, nil
}
