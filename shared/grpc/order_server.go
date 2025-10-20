package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"HubInvestments/internal/order_mngmt_system/application/command"
	di "HubInvestments/pck"
	monolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OrderServiceServer implements the OrderService gRPC interface
type OrderServiceServer struct {
	monolithpb.UnimplementedOrderServiceServer
	container di.Container
}

// NewOrderServiceServer creates a new OrderServiceServer
func NewOrderServiceServer(container di.Container) *OrderServiceServer {
	return &OrderServiceServer{
		container: container,
	}
}

// SubmitOrder submits a new trading order
func (s *OrderServiceServer) SubmitOrder(ctx context.Context, req *monolithpb.SubmitOrderRequest) (*monolithpb.SubmitOrderResponse, error) {
	userID, ok := ctx.Value("userId").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.Symbol == "" || req.Quantity <= 0 {
		return &monolithpb.SubmitOrderResponse{
			ApiResponse: &monolithpb.APIResponse{
				Success:   false,
				Message:   "Symbol and positive quantity are required",
				Code:      int32(codes.InvalidArgument),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	cmd := &command.SubmitOrderCommand{
		UserID:    userID,
		Symbol:    strings.ToUpper(req.Symbol),
		OrderType: req.OrderType,
		OrderSide: req.OrderSide,
		Quantity:  req.Quantity,
	}

	if req.Price != nil {
		cmd.Price = req.Price
	}

	submitOrderUseCase := s.container.GetSubmitOrderUseCase()
	result, err := submitOrderUseCase.Execute(ctx, cmd)
	if err != nil {
		return &monolithpb.SubmitOrderResponse{
			ApiResponse: &monolithpb.APIResponse{
				Success:   false,
				Message:   "Order submission failed: " + err.Error(),
				Code:      int32(codes.Internal),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	response := &monolithpb.SubmitOrderResponse{
		ApiResponse: &monolithpb.APIResponse{
			Success:   true,
			Message:   result.Message,
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		OrderId:     result.OrderID,
		Status:      result.Status,
		SubmittedAt: time.Now().Format(time.RFC3339),
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
func (s *OrderServiceServer) GetOrderDetails(ctx context.Context, req *monolithpb.GetOrderDetailsRequest) (*monolithpb.GetOrderDetailsResponse, error) {
	userID, ok := ctx.Value("userId").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.OrderId == "" {
		return &monolithpb.GetOrderDetailsResponse{
			ApiResponse: &monolithpb.APIResponse{
				Success:   false,
				Message:   "Order ID is required",
				Code:      int32(codes.InvalidArgument),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	if req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	getOrderStatusUseCase := s.container.GetGetOrderStatusUseCase()
	orderStatus, err := getOrderStatusUseCase.Execute(ctx, req.OrderId, userID)
	if err != nil {
		return &monolithpb.GetOrderDetailsResponse{
			ApiResponse: &monolithpb.APIResponse{
				Success:   false,
				Message:   "Failed to retrieve order details: " + err.Error(),
				Code:      int32(codes.NotFound),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	orderDetails := &monolithpb.OrderDetails{
		OrderId:        req.OrderId,
		UserId:         userID,
		Status:         orderStatus.Status,
		CreatedAt:      orderStatus.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      orderStatus.UpdatedAt.Format(time.RFC3339),
		EstimatedValue: 0,
	}

	return &monolithpb.GetOrderDetailsResponse{
		ApiResponse: &monolithpb.APIResponse{
			Success:   true,
			Message:   "Order details retrieved successfully",
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		Order: orderDetails,
	}, nil
}

// GetOrderStatus retrieves the status of a specific order
func (s *OrderServiceServer) GetOrderStatus(ctx context.Context, req *monolithpb.GetOrderStatusRequest) (*monolithpb.GetOrderStatusResponse, error) {
	userID, ok := ctx.Value("userId").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.OrderId == "" {
		return &monolithpb.GetOrderStatusResponse{
			ApiResponse: &monolithpb.APIResponse{
				Success:   false,
				Message:   "Order ID is required",
				Code:      int32(codes.InvalidArgument),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	if req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	getOrderStatusUseCase := s.container.GetGetOrderStatusUseCase()
	orderStatus, err := getOrderStatusUseCase.Execute(ctx, req.OrderId, userID)
	if err != nil {
		return &monolithpb.GetOrderStatusResponse{
			ApiResponse: &monolithpb.APIResponse{
				Success:   false,
				Message:   "Failed to retrieve order status: " + err.Error(),
				Code:      int32(codes.NotFound),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	return &monolithpb.GetOrderStatusResponse{
		ApiResponse: &monolithpb.APIResponse{
			Success:   true,
			Message:   "Order status retrieved successfully",
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		OrderId:       req.OrderId,
		Status:        orderStatus.Status,
		StatusMessage: fmt.Sprintf("Order is currently %s", orderStatus.Status),
		UpdatedAt:     orderStatus.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// CancelOrder cancels a pending order
func (s *OrderServiceServer) CancelOrder(ctx context.Context, req *monolithpb.CancelOrderRequest) (*monolithpb.CancelOrderResponse, error) {
	userID, ok := ctx.Value("userId").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.OrderId == "" {
		return &monolithpb.CancelOrderResponse{
			ApiResponse: &monolithpb.APIResponse{
				Success:   false,
				Message:   "Order ID is required",
				Code:      int32(codes.InvalidArgument),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	if req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	cmd := &command.CancelOrderCommand{
		OrderID: req.OrderId,
		UserID:  userID,
		Reason:  "User cancellation request via gRPC",
	}

	cancelOrderUseCase := s.container.GetCancelOrderUseCase()
	_, err := cancelOrderUseCase.Execute(ctx, cmd)
	if err != nil {
		return &monolithpb.CancelOrderResponse{
			ApiResponse: &monolithpb.APIResponse{
				Success:   false,
				Message:   "Failed to cancel order: " + err.Error(),
				Code:      int32(codes.Internal),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	return &monolithpb.CancelOrderResponse{
		ApiResponse: &monolithpb.APIResponse{
			Success:   true,
			Message:   "Order cancelled successfully",
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		OrderId:     req.OrderId,
		Status:      "CANCELLED",
		CancelledAt: time.Now().Format(time.RFC3339),
	}, nil
}
