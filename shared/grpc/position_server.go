package grpc

import (
	"context"
	"fmt"
	"time"

	di "HubInvestments/pck"
	"HubInvestments/shared/grpc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PositionServiceServer implements the PositionService gRPC interface
type PositionServiceServer struct {
	proto.UnimplementedPositionServiceServer
	container di.Container
}

// NewPositionServiceServer creates a new PositionServiceServer
func NewPositionServiceServer(container di.Container) *PositionServiceServer {
	return &PositionServiceServer{
		container: container,
	}
}

// GetPositions retrieves all positions for a user
func (s *PositionServiceServer) GetPositions(ctx context.Context, req *proto.GetPositionsRequest) (*proto.GetPositionsResponse, error) {
	userID, ok := ctx.Value("userId").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	positionAggregationUseCase := s.container.GetPositionAggregationUseCase()
	aggregation, err := positionAggregationUseCase.Execute(userID)
	if err != nil {
		return &proto.GetPositionsResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "Failed to retrieve positions: " + err.Error(),
				Code:      int32(codes.Internal),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	var protoPositions []*proto.Position
	for _, posAggregation := range aggregation.PositionAggregation {
		for _, asset := range posAggregation.Assets {
			protoPos := &proto.Position{
				PositionId:       "pos-" + asset.Symbol,
				UserId:           userID,
				Symbol:           asset.Symbol,
				Quantity:         float64(asset.Quantity),
				AveragePrice:     float64(asset.AveragePrice),
				TotalInvestment:  float64(asset.CalculateInvestment()),
				CurrentPrice:     float64(asset.LastPrice),
				MarketValue:      float64(asset.CalculateCurrentValue()),
				UnrealizedPnl:    float64(asset.CalculatePnL()),
				UnrealizedPnlPct: float64(asset.CalculatePnLPercentage()),
				PositionType:     "LONG",
				Status:           "ACTIVE",
				CreatedAt:        time.Now().Format(time.RFC3339),
				UpdatedAt:        time.Now().Format(time.RFC3339),
			}
			protoPositions = append(protoPositions, protoPos)
		}
	}

	return &proto.GetPositionsResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   fmt.Sprintf("Retrieved %d positions", len(protoPositions)),
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		Positions: protoPositions,
	}, nil
}

// GetPositionAggregation retrieves aggregated position data for a user
func (s *PositionServiceServer) GetPositionAggregation(ctx context.Context, req *proto.GetPositionAggregationRequest) (*proto.GetPositionAggregationResponse, error) {
	userID, ok := ctx.Value("userId").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	positionAggregationUseCase := s.container.GetPositionAggregationUseCase()
	aggregation, err := positionAggregationUseCase.Execute(userID)
	if err != nil {
		return &proto.GetPositionAggregationResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "Failed to retrieve position aggregation: " + err.Error(),
				Code:      int32(codes.Internal),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	var protoCategories []*proto.CategoryAggregation
	var protoPositions []*proto.Position
	var totalPositions int32 = 0

	for _, posAggregation := range aggregation.PositionAggregation {
		protoCategories = append(protoCategories, &proto.CategoryAggregation{
			CategoryId:         int32(posAggregation.Category),
			CategoryName:       fmt.Sprintf("Category %d", posAggregation.Category),
			TotalInvested:      float64(posAggregation.TotalInvested),
			TotalCurrentValue:  float64(posAggregation.CurrentTotal),
			TotalUnrealizedPnl: float64(posAggregation.Pnl),
			UnrealizedPnlPct:   float64(posAggregation.PnlPercentage),
			PositionCount:      int32(len(posAggregation.Assets)),
			WeightPct:          0,
		})

		for _, asset := range posAggregation.Assets {
			protoPos := &proto.Position{
				PositionId:       "pos-" + asset.Symbol,
				UserId:           userID,
				Symbol:           asset.Symbol,
				Quantity:         float64(asset.Quantity),
				AveragePrice:     float64(asset.AveragePrice),
				TotalInvestment:  float64(asset.CalculateInvestment()),
				CurrentPrice:     float64(asset.LastPrice),
				MarketValue:      float64(asset.CalculateCurrentValue()),
				UnrealizedPnl:    float64(asset.CalculatePnL()),
				UnrealizedPnlPct: float64(asset.CalculatePnLPercentage()),
				PositionType:     "LONG",
				Status:           "ACTIVE",
				CreatedAt:        time.Now().Format(time.RFC3339),
				UpdatedAt:        time.Now().Format(time.RFC3339),
			}
			protoPositions = append(protoPositions, protoPos)
			totalPositions++
		}
	}

	protoAggregation := &proto.PositionAggregation{
		TotalInvested:         float64(aggregation.TotalInvested),
		TotalCurrentValue:     float64(aggregation.CurrentTotal),
		TotalUnrealizedPnl:    float64(aggregation.CurrentTotal - aggregation.TotalInvested),
		TotalUnrealizedPnlPct: 0,
		TotalPositions:        totalPositions,
		ActivePositions:       totalPositions,
		Categories:            protoCategories,
		Positions:             protoPositions,
	}

	return &proto.GetPositionAggregationResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "Position aggregation retrieved successfully",
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		Aggregation: protoAggregation,
	}, nil
}

// CreatePosition creates a new position (for internal use)
func (s *PositionServiceServer) CreatePosition(ctx context.Context, req *proto.CreatePositionRequest) (*proto.CreatePositionResponse, error) {
	return &proto.CreatePositionResponse{
		ApiResponse: &proto.APIResponse{
			Success:   false,
			Message:   "CreatePosition is not implemented yet",
			Code:      int32(codes.Unimplemented),
			Timestamp: time.Now().Unix(),
		},
	}, nil
}

// UpdatePosition updates an existing position (for internal use)
func (s *PositionServiceServer) UpdatePosition(ctx context.Context, req *proto.UpdatePositionRequest) (*proto.UpdatePositionResponse, error) {
	return &proto.UpdatePositionResponse{
		ApiResponse: &proto.APIResponse{
			Success:   false,
			Message:   "UpdatePosition is not implemented yet",
			Code:      int32(codes.Unimplemented),
			Timestamp: time.Now().Unix(),
		},
	}, nil
}
