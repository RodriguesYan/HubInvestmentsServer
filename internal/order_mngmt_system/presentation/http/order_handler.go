package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"HubInvestments/internal/order_mngmt_system/application/command"
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	di "HubInvestments/pck"
	"HubInvestments/shared/middleware"
)

// SubmitOrderRequest represents the request body for submitting an order
type SubmitOrderRequest struct {
	Symbol    string   `json:"symbol" validate:"required"`
	OrderType string   `json:"order_type" validate:"required,oneof=MARKET LIMIT STOP_LOSS STOP_LIMIT"`
	OrderSide string   `json:"order_side" validate:"required,oneof=BUY SELL"`
	Quantity  float64  `json:"quantity" validate:"required,gt=0"`
	Price     *float64 `json:"price,omitempty"`
}

// SubmitOrderResponse represents the response after submitting an order
type SubmitOrderResponse struct {
	OrderID        string  `json:"order_id"`
	Status         string  `json:"status"`
	Message        string  `json:"message"`
	EstimatedPrice float64 `json:"estimated_price,omitempty"`
	EstimatedValue float64 `json:"estimated_value,omitempty"`
	MarketPrice    float64 `json:"market_price,omitempty"`
	SubmittedAt    string  `json:"submitted_at"`
}

// OrderDetailsResponse represents detailed order information
type OrderDetailsResponse struct {
	OrderID                 string   `json:"order_id"`
	UserID                  string   `json:"user_id"`
	Symbol                  string   `json:"symbol"`
	OrderType               string   `json:"order_type"`
	OrderSide               string   `json:"order_side"`
	Quantity                float64  `json:"quantity"`
	Price                   *float64 `json:"price,omitempty"`
	Status                  string   `json:"status"`
	CreatedAt               string   `json:"created_at"`
	UpdatedAt               string   `json:"updated_at"`
	ExecutedAt              *string  `json:"executed_at,omitempty"`
	ExecutionPrice          *float64 `json:"execution_price,omitempty"`
	MarketPriceAtSubmission *float64 `json:"market_price_at_submission,omitempty"`
	MarketDataTimestamp     *string  `json:"market_data_timestamp,omitempty"`
	EstimatedValue          float64  `json:"estimated_value"`
	ExecutionValue          float64  `json:"execution_value,omitempty"`
}

// OrderStatusResponse represents order status information
type OrderStatusResponse struct {
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	UpdatedAt string `json:"updated_at"`
	CanCancel bool   `json:"can_cancel"`
}

// OrderHistoryResponse represents a list of orders
type OrderHistoryResponse struct {
	Orders []OrderDetailsResponse `json:"orders"`
	Total  int                    `json:"total"`
	Page   int                    `json:"page"`
	Limit  int                    `json:"limit"`
}

// CancelOrderResponse represents the response after cancelling an order
type CancelOrderResponse struct {
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	UpdatedAt string `json:"updated_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// extractOrderIDFromPath extracts order ID from URL path like "/orders/{id}"
func extractOrderIDFromPath(path string) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid path format")
	}

	orderID := parts[1]
	if orderID == "" {
		return "", fmt.Errorf("order ID cannot be empty")
	}

	return orderID, nil
}

// validateSubmitOrderRequest validates the submit order request
func validateSubmitOrderRequest(req *SubmitOrderRequest) error {
	if req.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if req.OrderType == "" {
		return fmt.Errorf("order_type is required")
	}

	if req.OrderSide == "" {
		return fmt.Errorf("order_side is required")
	}

	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	// Validate order type
	switch req.OrderType {
	case "MARKET", "LIMIT", "STOP_LOSS", "STOP_LIMIT":
		// Valid
	default:
		return fmt.Errorf("invalid order_type: %s", req.OrderType)
	}

	// Validate order side
	switch req.OrderSide {
	case "BUY", "SELL":
		// Valid
	default:
		return fmt.Errorf("invalid order_side: %s", req.OrderSide)
	}

	// Validate price for limit orders
	if req.OrderType == "LIMIT" && req.Price == nil {
		return fmt.Errorf("price is required for LIMIT orders")
	}

	if req.OrderType == "MARKET" && req.Price != nil {
		return fmt.Errorf("price should not be specified for MARKET orders")
	}

	return nil
}

// convertToOrderDetailsResponse converts domain order to response
func convertToOrderDetailsResponse(order *domain.Order) OrderDetailsResponse {
	response := OrderDetailsResponse{
		OrderID:        order.ID(),
		UserID:         order.UserID(),
		Symbol:         order.Symbol(),
		OrderType:      order.OrderType().String(),
		OrderSide:      order.OrderSide().String(),
		Quantity:       order.Quantity(),
		Price:          order.Price(),
		Status:         order.Status().String(),
		CreatedAt:      order.CreatedAt().Format(time.RFC3339),
		UpdatedAt:      order.UpdatedAt().Format(time.RFC3339),
		EstimatedValue: order.CalculateOrderValue(),
	}

	if order.ExecutedAt() != nil {
		executedAt := order.ExecutedAt().Format(time.RFC3339)
		response.ExecutedAt = &executedAt
		response.ExecutionValue = order.CalculateExecutionValue()
	}

	if order.ExecutionPrice() != nil {
		response.ExecutionPrice = order.ExecutionPrice()
	}

	if order.MarketPriceAtSubmission() != nil {
		response.MarketPriceAtSubmission = order.MarketPriceAtSubmission()
	}

	if order.MarketDataTimestamp() != nil {
		timestamp := order.MarketDataTimestamp().Format(time.RFC3339)
		response.MarketDataTimestamp = &timestamp
	}

	return response
}

// SubmitOrder handles order submission
// @Summary Submit New Order
// @Description Submit a new trading order for processing
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body SubmitOrderRequest true "Order details"
// @Success 202 {object} SubmitOrderResponse "Order submitted successfully"
// @Failure 400 {object} ErrorResponse "Bad request - Invalid order data"
// @Failure 401 {object} ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /orders [post]
func SubmitOrder(w http.ResponseWriter, r *http.Request, userID string, container di.Container) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse := ErrorResponse{
			Error:   "Invalid JSON",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	if err := validateSubmitOrderRequest(&req); err != nil {
		errorResponse := ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Convert request to command
	cmd := &command.SubmitOrderCommand{
		UserID:    userID,
		Symbol:    strings.ToUpper(req.Symbol),
		OrderType: req.OrderType,
		OrderSide: req.OrderSide,
		Quantity:  req.Quantity,
		Price:     req.Price,
	}

	ctx := context.Background()
	result, err := container.GetSubmitOrderUseCase().Execute(ctx, cmd)
	if err != nil {
		errorResponse := ErrorResponse{
			Error:   "Order Submission Failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	response := SubmitOrderResponse{
		OrderID:     result.OrderID,
		Status:      result.Status,
		Message:     result.Message,
		SubmittedAt: time.Now().Format(time.RFC3339),
	}

	if result.EstimatedExecutionPrice != nil {
		response.EstimatedPrice = *result.EstimatedExecutionPrice
	}

	if result.MarketPriceAtSubmission != nil {
		response.MarketPrice = *result.MarketPriceAtSubmission
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// GetOrderDetails handles order details retrieval
// @Summary Get Order Details
// @Description Retrieve detailed information about a specific order
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} OrderDetailsResponse "Order details retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - Invalid order ID"
// @Failure 401 {object} ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 404 {object} ErrorResponse "Order not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /orders/{id} [get]
func GetOrderDetails(w http.ResponseWriter, r *http.Request, userID string, container di.Container) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID, err := extractOrderIDFromPath(r.URL.Path)
	if err != nil {
		errorResponse := ErrorResponse{
			Error:   "Invalid Order ID",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	ctx := context.Background()
	result, err := container.GetGetOrderStatusUseCase().Execute(ctx, orderID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			errorResponse := ErrorResponse{
				Error:   "Order Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		errorResponse := ErrorResponse{
			Error:   "Failed to Get Order",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	response := OrderDetailsResponse{
		OrderID:                 result.OrderID,
		UserID:                  result.UserID,
		Symbol:                  result.Symbol,
		OrderType:               result.OrderType,
		OrderSide:               result.OrderSide,
		Quantity:                result.Quantity,
		Price:                   result.Price,
		Status:                  result.Status,
		CreatedAt:               result.CreatedAt.Format(time.RFC3339),
		UpdatedAt:               result.UpdatedAt.Format(time.RFC3339),
		ExecutionPrice:          result.ExecutionPrice,
		MarketPriceAtSubmission: result.MarketPriceAtSubmission,
	}

	if result.ExecutedAt != nil {
		executedAt := result.ExecutedAt.Format(time.RFC3339)
		response.ExecutedAt = &executedAt
	}

	if result.MarketDataTimestamp != nil {
		timestamp := result.MarketDataTimestamp.Format(time.RFC3339)
		response.MarketDataTimestamp = &timestamp
	}

	if result.EstimatedValue != nil {
		response.EstimatedValue = *result.EstimatedValue
	}
	json.NewEncoder(w).Encode(response)
}

// GetOrderStatus handles order status retrieval
// @Summary Get Order Status
// @Description Retrieve the current status of a specific order
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} OrderStatusResponse "Order status retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - Invalid order ID"
// @Failure 401 {object} ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 404 {object} ErrorResponse "Order not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /orders/{id}/status [get]
func GetOrderStatus(w http.ResponseWriter, r *http.Request, userID string, container di.Container) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract order ID from path like "/orders/{id}/status"
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[2] != "status" {
		errorResponse := ErrorResponse{
			Error:   "Invalid Path",
			Message: "Expected path format: /orders/{id}/status",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	orderID := parts[1]
	if orderID == "" {
		errorResponse := ErrorResponse{
			Error:   "Invalid Order ID",
			Message: "Order ID cannot be empty",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	ctx := context.Background()
	result, err := container.GetGetOrderStatusUseCase().Execute(ctx, orderID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			errorResponse := ErrorResponse{
				Error:   "Order Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		errorResponse := ErrorResponse{
			Error:   "Failed to Get Order Status",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	response := OrderStatusResponse{
		OrderID:   result.OrderID,
		Status:    result.Status,
		Message:   result.StatusDescription,
		UpdatedAt: result.UpdatedAt.Format(time.RFC3339),
		CanCancel: result.CanCancel,
	}

	json.NewEncoder(w).Encode(response)
}

// CancelOrder handles order cancellation
// @Summary Cancel Order
// @Description Cancel a pending order
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} CancelOrderResponse "Order cancelled successfully"
// @Failure 400 {object} ErrorResponse "Bad request - Invalid order ID or order cannot be cancelled"
// @Failure 401 {object} ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 404 {object} ErrorResponse "Order not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /orders/{id}/cancel [put]
func CancelOrder(w http.ResponseWriter, r *http.Request, userID string, container di.Container) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract order ID from path like "/orders/{id}/cancel"
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[2] != "cancel" {
		errorResponse := ErrorResponse{
			Error:   "Invalid Path",
			Message: "Expected path format: /orders/{id}/cancel",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	orderID := parts[1]
	if orderID == "" {
		errorResponse := ErrorResponse{
			Error:   "Invalid Order ID",
			Message: "Order ID cannot be empty",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	cmd := &command.CancelOrderCommand{
		OrderID: orderID,
		UserID:  userID,
		Reason:  "User requested cancellation",
	}

	ctx := context.Background()
	result, err := container.GetCancelOrderUseCase().Execute(ctx, cmd)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			errorResponse := ErrorResponse{
				Error:   "Order Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		if strings.Contains(err.Error(), "cannot be cancelled") {
			errorResponse := ErrorResponse{
				Error:   "Cannot Cancel Order",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		errorResponse := ErrorResponse{
			Error:   "Failed to Cancel Order",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	response := CancelOrderResponse{
		OrderID:   result.OrderID,
		Status:    result.Status,
		Message:   result.Message,
		UpdatedAt: result.Timestamp,
	}

	json.NewEncoder(w).Encode(response)
}

// GetOrderHistory handles order history retrieval
// @Summary Get Order History
// @Description Retrieve order history for the authenticated user
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of orders per page (default: 20, max: 100)"
// @Success 200 {object} OrderHistoryResponse "Order history retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - Invalid pagination parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /orders/history [get]
func GetOrderHistory(w http.ResponseWriter, r *http.Request, userID string, container di.Container) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 20

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// For now, return a simple response indicating the feature is not fully implemented
	// TODO: Implement proper GetOrderHistory method in the use case
	response := OrderHistoryResponse{
		Orders: []OrderDetailsResponse{},
		Total:  0,
		Page:   page,
		Limit:  limit,
	}

	json.NewEncoder(w).Encode(response)
}

// SubmitOrderWithAuth returns a handler wrapped with authentication middleware
func SubmitOrderWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userID string) {
		SubmitOrder(w, r, userID, container)
	})
}

// GetOrderDetailsWithAuth returns a handler wrapped with authentication middleware
func GetOrderDetailsWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userID string) {
		GetOrderDetails(w, r, userID, container)
	})
}

// GetOrderStatusWithAuth returns a handler wrapped with authentication middleware
func GetOrderStatusWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userID string) {
		GetOrderStatus(w, r, userID, container)
	})
}

// CancelOrderWithAuth returns a handler wrapped with authentication middleware
func CancelOrderWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userID string) {
		CancelOrder(w, r, userID, container)
	})
}

// GetOrderHistoryWithAuth returns a handler wrapped with authentication middleware
func GetOrderHistoryWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userID string) {
		GetOrderHistory(w, r, userID, container)
	})
}
