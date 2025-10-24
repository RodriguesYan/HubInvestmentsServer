package grpc

import (
	"context"

	orderCommand "HubInvestments/internal/order_mngmt_system/application/command"
	orderUsecase "HubInvestments/internal/order_mngmt_system/application/usecase"
	di "HubInvestments/pck"

	commonpb "github.com/RodriguesYan/hub-proto-contracts/common"
	monolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGRPCHandler struct {
	monolithpb.UnimplementedOrderServiceServer
	container di.Container
}

func NewOrderGRPCHandler(container di.Container) *OrderGRPCHandler {
	return &OrderGRPCHandler{
		container: container,
	}
}

// SubmitOrder submits a new trading order
func (h *OrderGRPCHandler) SubmitOrder(ctx context.Context, req *monolithpb.SubmitOrderRequest) (*monolithpb.SubmitOrderResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}
	if req.OrderType == "" {
		return nil, status.Error(codes.InvalidArgument, "order_type is required")
	}
	if req.OrderSide == "" {
		return nil, status.Error(codes.InvalidArgument, "order_side is required")
	}
	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be greater than 0")
	}

	// Create command and call existing use case
	cmd := &orderCommand.SubmitOrderCommand{
		UserID:    req.UserId,
		Symbol:    req.Symbol,
		OrderType: req.OrderType,
		OrderSide: req.OrderSide,
		Quantity:  req.Quantity,
		Price:     req.Price,
	}

	result, err := h.container.GetSubmitOrderUseCase().Execute(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to submit order: %v", err)
	}

	// Map result to proto response
	response := &monolithpb.SubmitOrderResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   true,
			Message:   result.Message,
			Code:      202,
			Timestamp: 0,
		},
		OrderId: result.OrderID,
		Status:  result.Status,
	}

	if result.EstimatedExecutionPrice != nil {
		response.EstimatedPrice = result.EstimatedExecutionPrice
	}
	if result.MarketPriceAtSubmission != nil {
		response.MarketPrice = result.MarketPriceAtSubmission
	}

	return response, nil
}

// GetOrderDetails retrieves detailed information about a specific order
func (h *OrderGRPCHandler) GetOrderDetails(ctx context.Context, req *monolithpb.GetOrderDetailsRequest) (*monolithpb.GetOrderDetailsResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Call existing use case
	result, err := h.container.GetGetOrderStatusUseCase().Execute(ctx, req.OrderId, req.UserId)
	if err != nil {
		if contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get order details: %v", err)
	}

	// Map result to proto response
	orderDetails := &monolithpb.OrderDetails{
		OrderId:                 result.OrderID,
		UserId:                  result.UserID,
		Symbol:                  result.Symbol,
		OrderType:               result.OrderType,
		OrderSide:               result.OrderSide,
		Quantity:                result.Quantity,
		Price:                   result.Price,
		Status:                  result.Status,
		CreatedAt:               result.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:               result.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExecutionPrice:          result.ExecutionPrice,
		MarketPriceAtSubmission: result.MarketPriceAtSubmission,
	}

	if result.ExecutedAt != nil {
		executedAt := result.ExecutedAt.Format("2006-01-02T15:04:05Z07:00")
		orderDetails.ExecutedAt = &executedAt
	}

	if result.MarketDataTimestamp != nil {
		timestamp := result.MarketDataTimestamp.Format("2006-01-02T15:04:05Z07:00")
		orderDetails.MarketDataTimestamp = &timestamp
	}

	if result.EstimatedValue != nil {
		orderDetails.EstimatedValue = *result.EstimatedValue
	}

	return &monolithpb.GetOrderDetailsResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   true,
			Message:   "Order details retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		Order: orderDetails,
	}, nil
}

// GetOrderStatus retrieves the status of a specific order
func (h *OrderGRPCHandler) GetOrderStatus(ctx context.Context, req *monolithpb.GetOrderStatusRequest) (*monolithpb.GetOrderStatusResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Call existing use case
	result, err := h.container.GetGetOrderStatusUseCase().Execute(ctx, req.OrderId, req.UserId)
	if err != nil {
		if contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get order status: %v", err)
	}

	// Map result to proto response
	return &monolithpb.GetOrderStatusResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   true,
			Message:   "Order status retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		OrderId:       result.OrderID,
		Status:        result.Status,
		StatusMessage: result.StatusDescription,
		UpdatedAt:     result.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CancelOrder cancels a pending order
func (h *OrderGRPCHandler) CancelOrder(ctx context.Context, req *monolithpb.CancelOrderRequest) (*monolithpb.CancelOrderResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Create command and call existing use case
	cmd := &orderCommand.CancelOrderCommand{
		OrderID: req.OrderId,
		UserID:  req.UserId,
		Reason:  "User requested cancellation via gRPC",
	}

	result, err := h.container.GetCancelOrderUseCase().Execute(ctx, cmd)
	if err != nil {
		if contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		if contains(err.Error(), "cannot be cancelled") {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to cancel order: %v", err)
	}

	// Map result to proto response
	return &monolithpb.CancelOrderResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   true,
			Message:   "Order cancelled successfully",
			Code:      200,
			Timestamp: 0,
		},
		OrderId:     result.OrderID,
		Status:      result.Status,
		CancelledAt: result.Timestamp,
	}, nil
}

func (h *OrderGRPCHandler) GetOrderHistory(ctx context.Context, req *monolithpb.GetOrderHistoryRequest) (*monolithpb.GetOrderHistoryResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Set defaults for pagination
	limit := int32(50)
	if req.Limit != nil && *req.Limit > 0 {
		limit = *req.Limit
		if limit > 100 {
			limit = 100
		}
	}

	offset := int32(0)
	if req.Offset != nil && *req.Offset >= 0 {
		offset = *req.Offset
	}

	options := &orderUsecase.OrderHistoryOptions{
		Limit:     int(limit),
		Offset:    int(offset),
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	result, err := h.container.GetGetOrderStatusUseCase().GetOrderHistory(ctx, req.UserId, options)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get order history: %v", err)
	}

	// Map orders to proto response
	orderDetails := make([]*monolithpb.OrderDetails, 0, len(result.Orders))
	for _, orderResult := range result.Orders {
		details := &monolithpb.OrderDetails{
			OrderId:   orderResult.OrderID,
			UserId:    orderResult.UserID,
			Symbol:    orderResult.Symbol,
			OrderType: orderResult.OrderType,
			OrderSide: orderResult.OrderSide,
			Quantity:  orderResult.Quantity,
			Status:    orderResult.Status,
			CreatedAt: orderResult.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: orderResult.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Add optional fields
		if orderResult.Price != nil {
			details.Price = orderResult.Price
		}
		if orderResult.ExecutionPrice != nil {
			details.ExecutionPrice = orderResult.ExecutionPrice
		}
		if orderResult.MarketPriceAtSubmission != nil {
			details.MarketPriceAtSubmission = orderResult.MarketPriceAtSubmission
		}
		if orderResult.ExecutedAt != nil {
			executedAt := orderResult.ExecutedAt.Format("2006-01-02T15:04:05Z07:00")
			details.ExecutedAt = &executedAt
		}
		if orderResult.MarketDataTimestamp != nil {
			timestamp := orderResult.MarketDataTimestamp.Format("2006-01-02T15:04:05Z07:00")
			details.MarketDataTimestamp = &timestamp
		}
		if orderResult.EstimatedValue != nil {
			details.EstimatedValue = *orderResult.EstimatedValue
		}

		orderDetails = append(orderDetails, details)
	}

	return &monolithpb.GetOrderHistoryResponse{
		ApiResponse: &commonpb.APIResponse{
			Success:   true,
			Message:   "Order history retrieved successfully",
			Code:      200,
			Timestamp: 0,
		},
		Orders:     orderDetails,
		TotalCount: int32(result.TotalCount),
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
