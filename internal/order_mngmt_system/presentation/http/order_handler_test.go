package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"HubInvestments/internal/auth"
	balUsecase "HubInvestments/internal/balance/application/usecase"
	doLoginUsecase "HubInvestments/internal/login/application/usecase"
	mktUsecase "HubInvestments/internal/market_data/application/usecase"
	"HubInvestments/internal/order_mngmt_system/application/command"
	orderUsecase "HubInvestments/internal/order_mngmt_system/application/usecase"
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	orderMktClient "HubInvestments/internal/order_mngmt_system/infra/external"
	orderRabbitMQ "HubInvestments/internal/order_mngmt_system/infra/messaging/rabbitmq"
	orderWorker "HubInvestments/internal/order_mngmt_system/infra/worker"
	portfolioUsecase "HubInvestments/internal/portfolio_summary/application/usecase"
	posUsecase "HubInvestments/internal/position/application/usecase"
	watchlistUsecase "HubInvestments/internal/watchlist/application/usecase"
	"HubInvestments/shared/infra/messaging"
	"HubInvestments/shared/infra/websocket"
)

// MockContainer implements the Container interface for testing
type MockContainer struct {
	submitOrderUseCase    MockSubmitOrderUseCase
	getOrderStatusUseCase MockGetOrderStatusUseCase
	cancelOrderUseCase    MockCancelOrderUseCase
}

func (m *MockContainer) DoLoginUsecase() doLoginUsecase.IDoLoginUsecase { return nil }
func (m *MockContainer) GetAuthService() auth.IAuthService              { return nil }
func (m *MockContainer) GetPositionAggregationUseCase() *posUsecase.GetPositionAggregationUseCase {
	return nil
}
func (m *MockContainer) GetBalanceUseCase() *balUsecase.GetBalanceUseCase { return nil }
func (m *MockContainer) GetPortfolioSummaryUsecase() portfolioUsecase.PortfolioSummaryUsecase {
	return nil
}
func (m *MockContainer) GetMarketDataUsecase() mktUsecase.IGetMarketDataUsecase     { return nil }
func (m *MockContainer) GetWatchlistUsecase() watchlistUsecase.IGetWatchlistUsecase { return nil }
func (m *MockContainer) GetOrderMarketDataClient() orderMktClient.IMarketDataClient { return nil }
func (m *MockContainer) InvalidateMarketDataCache(symbols []string) error           { return nil }
func (m *MockContainer) WarmMarketDataCache(symbols []string) error                 { return nil }
func (m *MockContainer) GetMessageHandler() messaging.MessageHandler                { return nil }
func (m *MockContainer) Close() error                                               { return nil }

func (m *MockContainer) GetSubmitOrderUseCase() orderUsecase.ISubmitOrderUseCase {
	return &m.submitOrderUseCase
}

func (m *MockContainer) GetGetOrderStatusUseCase() orderUsecase.IGetOrderStatusUseCase {
	return &m.getOrderStatusUseCase
}

func (m *MockContainer) GetCancelOrderUseCase() orderUsecase.ICancelOrderUseCase {
	return &m.cancelOrderUseCase
}

func (m *MockContainer) GetProcessOrderUseCase() orderUsecase.IProcessOrderUseCase {
	return nil
}

func (m *MockContainer) GetOrderProducer() *orderRabbitMQ.OrderProducer {
	return nil
}

func (m *MockContainer) GetOrderWorkerManager() *orderWorker.WorkerManager {
	return nil
}

func (m *MockContainer) GetWebSocketManager() websocket.WebSocketManager {
	return nil
}

// MockSubmitOrderUseCase implements ISubmitOrderUseCase for testing
type MockSubmitOrderUseCase struct {
	ExecuteFunc func(ctx context.Context, cmd *command.SubmitOrderCommand) (*command.SubmitOrderResult, error)
}

func (m *MockSubmitOrderUseCase) Execute(ctx context.Context, cmd *command.SubmitOrderCommand) (*command.SubmitOrderResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, cmd)
	}
	return &command.SubmitOrderResult{
		OrderID: "test-order-id",
		Status:  "PENDING",
		Message: "Order submitted successfully",
	}, nil
}

// MockGetOrderStatusUseCase implements IGetOrderStatusUseCase for testing
type MockGetOrderStatusUseCase struct {
	ExecuteFunc func(ctx context.Context, orderID, userID string) (*orderUsecase.OrderStatusResult, error)
}

func (m *MockGetOrderStatusUseCase) Execute(ctx context.Context, orderID, userID string) (*orderUsecase.OrderStatusResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, orderID, userID)
	}
	return &orderUsecase.OrderStatusResult{
		OrderID:           "test-order-id",
		UserID:            userID,
		Symbol:            "AAPL",
		OrderType:         "LIMIT",
		OrderSide:         "BUY",
		Quantity:          100,
		Status:            "PENDING",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		StatusDescription: "Order is pending",
		CanCancel:         true,
	}, nil
}

func (m *MockGetOrderStatusUseCase) GetOrderHistory(ctx context.Context, userID string, options *orderUsecase.OrderHistoryOptions) (*orderUsecase.OrderHistoryResult, error) {
	return &orderUsecase.OrderHistoryResult{}, nil
}

// MockCancelOrderUseCase implements ICancelOrderUseCase for testing
type MockCancelOrderUseCase struct {
	ExecuteFunc func(ctx context.Context, cmd *command.CancelOrderCommand) (*command.CancelOrderResult, error)
}

func (m *MockCancelOrderUseCase) Execute(ctx context.Context, cmd *command.CancelOrderCommand) (*command.CancelOrderResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, cmd)
	}
	return &command.CancelOrderResult{
		OrderID:   cmd.OrderID,
		Status:    "CANCELLED",
		Message:   "Order cancelled successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// Mock token verifier for testing
func mockTokenVerifier(token string, w http.ResponseWriter) (string, error) {
	if token == "Bearer valid-token" {
		return "test-user-id", nil
	}
	return "", fmt.Errorf("invalid token")
}

func TestSubmitOrder_Success(t *testing.T) {
	container := &MockContainer{}

	requestBody := SubmitOrderRequest{
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100,
		Price:     func() *float64 { p := 150.50; return &p }(),
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler := SubmitOrderWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	var response SubmitOrderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.OrderID != "test-order-id" {
		t.Errorf("Expected OrderID 'test-order-id', got '%s'", response.OrderID)
	}

	if response.Status != "PENDING" {
		t.Errorf("Expected Status 'PENDING', got '%s'", response.Status)
	}
}

func TestSubmitOrder_InvalidJSON(t *testing.T) {
	container := &MockContainer{}

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader("invalid json"))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler := SubmitOrderWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSubmitOrder_ValidationError(t *testing.T) {
	container := &MockContainer{}

	requestBody := SubmitOrderRequest{
		Symbol:    "", // Invalid: empty symbol
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100,
		Price:     func() *float64 { p := 150.50; return &p }(),
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler := SubmitOrderWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSubmitOrder_UseCaseError(t *testing.T) {
	container := &MockContainer{
		submitOrderUseCase: MockSubmitOrderUseCase{
			ExecuteFunc: func(ctx context.Context, cmd *command.SubmitOrderCommand) (*command.SubmitOrderResult, error) {
				return nil, fmt.Errorf("use case error")
			},
		},
	}

	requestBody := SubmitOrderRequest{
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100,
		Price:     func() *float64 { p := 150.50; return &p }(),
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler := SubmitOrderWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGetOrderDetails_Success(t *testing.T) {
	container := &MockContainer{}

	req := httptest.NewRequest(http.MethodGet, "/orders/test-order-id", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()

	handler := GetOrderDetailsWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response OrderDetailsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.OrderID != "test-order-id" {
		t.Errorf("Expected OrderID 'test-order-id', got '%s'", response.OrderID)
	}
}

func TestGetOrderDetails_InvalidPath(t *testing.T) {
	container := &MockContainer{}

	req := httptest.NewRequest(http.MethodGet, "/orders/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()

	handler := GetOrderDetailsWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetOrderDetails_NotFound(t *testing.T) {
	container := &MockContainer{
		getOrderStatusUseCase: MockGetOrderStatusUseCase{
			ExecuteFunc: func(ctx context.Context, orderID, userID string) (*orderUsecase.OrderStatusResult, error) {
				return nil, fmt.Errorf("order not found")
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/orders/nonexistent-order", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()

	handler := GetOrderDetailsWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetOrderStatus_Success(t *testing.T) {
	container := &MockContainer{}

	req := httptest.NewRequest(http.MethodGet, "/orders/test-order-id/status", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()

	handler := GetOrderStatusWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response OrderStatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.OrderID != "test-order-id" {
		t.Errorf("Expected OrderID 'test-order-id', got '%s'", response.OrderID)
	}

	if response.Status != "PENDING" {
		t.Errorf("Expected Status 'PENDING', got '%s'", response.Status)
	}
}

func TestCancelOrder_Success(t *testing.T) {
	container := &MockContainer{}

	req := httptest.NewRequest(http.MethodPut, "/orders/test-order-id/cancel", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()

	handler := CancelOrderWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response CancelOrderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.OrderID != "test-order-id" {
		t.Errorf("Expected OrderID 'test-order-id', got '%s'", response.OrderID)
	}

	if response.Status != "CANCELLED" {
		t.Errorf("Expected Status 'CANCELLED', got '%s'", response.Status)
	}
}

func TestCancelOrder_CannotCancel(t *testing.T) {
	container := &MockContainer{
		cancelOrderUseCase: MockCancelOrderUseCase{
			ExecuteFunc: func(ctx context.Context, cmd *command.CancelOrderCommand) (*command.CancelOrderResult, error) {
				return nil, fmt.Errorf("order cannot be cancelled")
			},
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/orders/test-order-id/cancel", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()

	handler := CancelOrderWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetOrderHistory_Success(t *testing.T) {
	container := &MockContainer{}

	req := httptest.NewRequest(http.MethodGet, "/orders/history?page=1&limit=20", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()

	handler := GetOrderHistoryWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response OrderHistoryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Page != 1 {
		t.Errorf("Expected Page 1, got %d", response.Page)
	}

	if response.Limit != 20 {
		t.Errorf("Expected Limit 20, got %d", response.Limit)
	}
}

func TestValidateSubmitOrderRequest_Success(t *testing.T) {
	req := &SubmitOrderRequest{
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100,
		Price:     func() *float64 { p := 150.50; return &p }(),
	}

	err := validateSubmitOrderRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidateSubmitOrderRequest_EmptySymbol(t *testing.T) {
	req := &SubmitOrderRequest{
		Symbol:    "",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100,
		Price:     func() *float64 { p := 150.50; return &p }(),
	}

	err := validateSubmitOrderRequest(req)
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
}

func TestValidateSubmitOrderRequest_InvalidOrderType(t *testing.T) {
	req := &SubmitOrderRequest{
		Symbol:    "AAPL",
		OrderType: "INVALID",
		OrderSide: "BUY",
		Quantity:  100,
		Price:     func() *float64 { p := 150.50; return &p }(),
	}

	err := validateSubmitOrderRequest(req)
	if err == nil {
		t.Error("Expected error for invalid order type")
	}
}

func TestValidateSubmitOrderRequest_LimitOrderWithoutPrice(t *testing.T) {
	req := &SubmitOrderRequest{
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100,
		Price:     nil,
	}

	err := validateSubmitOrderRequest(req)
	if err == nil {
		t.Error("Expected error for LIMIT order without price")
	}
}

func TestValidateSubmitOrderRequest_MarketOrderWithPrice(t *testing.T) {
	req := &SubmitOrderRequest{
		Symbol:    "AAPL",
		OrderType: "MARKET",
		OrderSide: "BUY",
		Quantity:  100,
		Price:     func() *float64 { p := 150.50; return &p }(),
	}

	err := validateSubmitOrderRequest(req)
	if err == nil {
		t.Error("Expected error for MARKET order with price")
	}
}

func TestExtractOrderIDFromPath_Success(t *testing.T) {
	orderID, err := extractOrderIDFromPath("/orders/test-order-id")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if orderID != "test-order-id" {
		t.Errorf("Expected 'test-order-id', got '%s'", orderID)
	}
}

func TestExtractOrderIDFromPath_InvalidPath(t *testing.T) {
	_, err := extractOrderIDFromPath("/orders/")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestExtractOrderIDFromPath_EmptyOrderID(t *testing.T) {
	_, err := extractOrderIDFromPath("/orders//")
	if err == nil {
		t.Error("Expected error for empty order ID")
	}
}

func TestConvertToOrderDetailsResponse(t *testing.T) {
	now := time.Now()
	executedAt := now.Add(time.Hour)
	price := 150.50
	executionPrice := 150.75
	marketPrice := 150.25

	order := domain.NewOrderFromDatabase(
		"test-order-id",
		"test-user-id",
		"AAPL",
		domain.OrderSideBuy,
		domain.OrderTypeLimit,
		100,
		&price,
		domain.OrderStatusExecuted,
		now,
		now,
		&executedAt,
		&executionPrice,
		&marketPrice,
		&now,
	)

	response := convertToOrderDetailsResponse(order)

	if response.OrderID != "test-order-id" {
		t.Errorf("Expected OrderID 'test-order-id', got '%s'", response.OrderID)
	}

	if response.Symbol != "AAPL" {
		t.Errorf("Expected Symbol 'AAPL', got '%s'", response.Symbol)
	}

	if response.ExecutionPrice == nil || *response.ExecutionPrice != 150.75 {
		t.Errorf("Expected ExecutionPrice 150.75, got %v", response.ExecutionPrice)
	}
}

// Test authentication middleware integration
func TestAuthenticationMiddleware_ValidToken(t *testing.T) {
	container := &MockContainer{}

	req := httptest.NewRequest(http.MethodGet, "/orders/test-order-id/status", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()

	handler := GetOrderStatusWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthenticationMiddleware_InvalidToken(t *testing.T) {
	container := &MockContainer{}

	req := httptest.NewRequest(http.MethodGet, "/orders/test-order-id/status", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()

	handler := GetOrderStatusWithAuth(mockTokenVerifier, container)
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
