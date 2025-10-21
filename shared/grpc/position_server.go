package grpc

import (
	"context"
	"fmt"
	"time"

	di "HubInvestments/pck"

	commonpb "github.com/RodriguesYan/hub-proto-contracts/common"
	monolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PositionServiceServer implements the PositionService gRPC interface
type PositionServiceServer struct {
	monolithpb.UnimplementedPositionServiceServer
	container di.Container
}

// NewPositionServiceServer creates a new PositionServiceServer
func NewPositionServiceServer(container di.Container) *PositionServiceServer {
	return &PositionServiceServer{
		container: container,
	}
}

// GetPositions retrieves all positions for a user
func (s *PositionServiceServer) GetPositions(ctx context.Context, req *monolithpb.GetPositionsRequest) (*monolithpb.GetPositionsResponse, error) {
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
		return &monolithpb.GetPositionsResponse{
			ApiResponse: &commonpb.APIResponse{
				Success:   false,
				Message:   "Failed to retrieve positions: " + err.Error(),
				Code:      int32(codes.Internal),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	var protoPositions []*monolithpb.Position
	for _, posAggregation := range aggregation.PositionAggregation {
		for _, asset := range posAggregation.Assets {
			protoPos := &monolithpb.Position{
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

	return &monolithpb.GetPositionsResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   true,
			Message:   fmt.Sprintf("Retrieved %d positions", len(protoPositions)),
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		Positions: protoPositions,
	}, nil
}

// GetPositionAggregation retrieves aggregated position data for a user
func (s *PositionServiceServer) GetPositionAggregation(ctx context.Context, req *monolithpb.GetPositionAggregationRequest) (*monolithpb.GetPositionAggregationResponse, error) {
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
		return &monolithpb.GetPositionAggregationResponse{
			ApiResponse: &commonpb.APIResponse{
				Success:   false,
				Message:   "Failed to retrieve position aggregation: " + err.Error(),
				Code:      int32(codes.Internal),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	var protoCategories []*monolithpb.CategoryAggregation
	var protoPositions []*monolithpb.Position
	var totalPositions int32 = 0

	for _, posAggregation := range aggregation.PositionAggregation {
		protoCategories = append(protoCategories, &monolithpb.CategoryAggregation{
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
			protoPos := &monolithpb.Position{
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

	protoAggregation := &monolithpb.PositionAggregation{
		TotalInvested:         float64(aggregation.TotalInvested),
		TotalCurrentValue:     float64(aggregation.CurrentTotal),
		TotalUnrealizedPnl:    float64(aggregation.CurrentTotal - aggregation.TotalInvested),
		TotalUnrealizedPnlPct: 0,
		TotalPositions:        totalPositions,
		ActivePositions:       totalPositions,
		Categories:            protoCategories,
		Positions:             protoPositions,
	}

	return &monolithpb.GetPositionAggregationResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   true,
			Message:   "Position aggregation retrieved successfully",
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		Aggregation: protoAggregation,
	}, nil
}

// CreatePosition creates a new position (for internal use)
func (s *PositionServiceServer) CreatePosition(ctx context.Context, req *monolithpb.CreatePositionRequest) (*monolithpb.CreatePositionResponse, error) {
	return &monolithpb.CreatePositionResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   false,
			Message:   "CreatePosition is not implemented yet",
			Code:      int32(codes.Unimplemented),
			Timestamp: time.Now().Unix(),
		},
	}, nil
}

// UpdatePosition updates an existing position (for internal use)
func (s *PositionServiceServer) UpdatePosition(ctx context.Context, req *monolithpb.UpdatePositionRequest) (*monolithpb.UpdatePositionResponse, error) {
	return &monolithpb.UpdatePositionResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   false,
			Message:   "UpdatePosition is not implemented yet",
			Code:      int32(codes.Unimplemented),
			Timestamp: time.Now().Unix(),
		},
	}, nil
}
