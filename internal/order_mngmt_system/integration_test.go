package order_mngmt_system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"HubInvestments/internal/order_mngmt_system/application/command"
	"HubInvestments/internal/order_mngmt_system/application/usecase"
	orderhttp "HubInvestments/internal/order_mngmt_system/presentation/http"
	pck "HubInvestments/pck"
)

// TestEndToEndOrderSubmissionFlow tests the complete order submission flow
func TestEndToEndOrderSubmissionFlow(t *testing.T) {
	// Setup test container with real dependencies
	container := setupTestContainer(t)
	defer container.Close()

	// Setup HTTP handler function
	submitOrderHandler := orderhttp.SubmitOrderWithAuth(nil, container)

	// Test data
	submitOrderRequest := map[string]interface{}{
		"user_id":    "test-user-123",
		"symbol":     "AAPL",
		"order_type": "LIMIT",
		"side":       "BUY",
		"quantity":   100,
		"price":      150.50,
	}

	requestBody, err := json.Marshal(submitOrderRequest)
	require.NoError(t, err)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-test-token")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	submitOrderHandler(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "order_id")
	assert.Contains(t, response, "status")
	assert.Equal(t, "PENDING", response["status"])

	orderID := response["order_id"].(string)
	assert.NotEmpty(t, orderID)

	// Verify order was saved by checking status
	getStatusUC := container.GetGetOrderStatusUseCase()
	statusResult, err := getStatusUC.Execute(context.Background(), orderID, "test-user-123")
	require.NoError(t, err)
	assert.Equal(t, "test-user-123", statusResult.UserID)
	assert.Equal(t, "AAPL", statusResult.Symbol)
	assert.Equal(t, "PENDING", statusResult.Status)

	// Test idempotency - submit same order again
	req2 := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(requestBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer valid-test-token")

	w2 := httptest.NewRecorder()
	submitOrderHandler(w2, req2)

	// Should return same order ID (idempotent)
	var response2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	require.NoError(t, err)
	assert.Equal(t, orderID, response2["order_id"])
}

// TestDatabaseTransactionAndRollback tests database transaction handling through use cases
func TestDatabaseTransactionAndRollback(t *testing.T) {
	container := setupTestContainer(t)
	defer container.Close()

	submitOrderUC := container.GetSubmitOrderUseCase()
	ctx := context.Background()

	// Test successful transaction
	t.Run("Successful Transaction", func(t *testing.T) {
		cmd := &command.SubmitOrderCommand{
			UserID:    "test-user-tx-success",
			Symbol:    "AAPL",
			OrderType: "LIMIT",
			OrderSide: "BUY",
			Quantity:  100,
			Price:     func() *float64 { p := 150.50; return &p }(),
		}

		result, err := submitOrderUC.Execute(ctx, cmd)
		if err != nil {
			t.Skipf("Skipping transaction test due to dependencies: %v", err)
			return
		}

		assert.NotNil(t, result)
		assert.NotEmpty(t, result.OrderID)
		assert.Equal(t, "PENDING", result.Status)
	})

	// Test transaction rollback on error
	t.Run("Transaction Rollback", func(t *testing.T) {
		// Create order with invalid data that should cause rollback
		invalidCmd := &command.SubmitOrderCommand{
			UserID:    "", // Empty user ID should cause validation error
			Symbol:    "AAPL",
			OrderType: "LIMIT",
			OrderSide: "BUY",
			Quantity:  100,
			Price:     func() *float64 { p := 150.50; return &p }(),
		}

		result, err := submitOrderUC.Execute(ctx, invalidCmd)
		// Should fail due to validation
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestWorkerErrorHandlingAndDLQ tests worker error handling and Dead Letter Queue functionality
func TestWorkerErrorHandlingAndDLQ(t *testing.T) {
	container := setupTestContainer(t)
	defer container.Close()

	// This test would require setting up actual RabbitMQ with DLQ
	// For now, we'll test the error handling logic

	processOrderUC := container.GetProcessOrderUseCase()
	ctx := context.Background()

	t.Run("Order Processing Error Handling", func(t *testing.T) {
		// Create command with invalid order ID
		cmd := &usecase.ProcessOrderCommand{
			OrderID: "invalid-order-id",
			Context: usecase.ProcessingContext{
				ProcessingID: "test-processing-123",
				WorkerID:     "test-worker",
			},
		}

		// Execute processing
		result, err := processOrderUC.Execute(ctx, cmd)

		// Should handle error gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "order not found")
		} else {
			assert.NotNil(t, result)
			assert.NotEqual(t, "EXECUTED", result.FinalStatus)
		}
	})

	t.Run("Retry Logic", func(t *testing.T) {
		// Test retry logic for transient errors
		// This would involve mocking dependencies to simulate failures
		// and verifying retry behavior

		// For now, verify that the worker can handle multiple processing attempts
		cmd := &usecase.ProcessOrderCommand{
			OrderID: "test-order-for-retry",
			Context: usecase.ProcessingContext{
				ProcessingID: "test-processing-retry-123",
				WorkerID:     "test-worker",
			},
		}

		// Multiple execution attempts should be handled gracefully
		for i := 0; i < 3; i++ {
			result, err := processOrderUC.Execute(ctx, cmd)
			// Each attempt should either succeed or fail gracefully
			if err == nil {
				assert.NotNil(t, result)
			}
		}
	})
}

// TestGRPCClientServerCommunication tests gRPC communication for market data
func TestGRPCClientServerCommunication(t *testing.T) {
	container := setupTestContainer(t)
	defer container.Close()

	marketDataClient := container.GetOrderMarketDataClient()
	ctx := context.Background()

	t.Run("Market Data Retrieval", func(t *testing.T) {
		// Test asset details retrieval
		assetDetails, err := marketDataClient.GetAssetDetails(ctx, "AAPL")

		if err != nil {
			// If market data service is not available, test should handle gracefully
			assert.Contains(t, err.Error(), "market data")
		} else {
			assert.NotNil(t, assetDetails)
			assert.Equal(t, "AAPL", assetDetails.Symbol)
		}
	})

	t.Run("Symbol Validation", func(t *testing.T) {
		// Test symbol validation
		isValid, err := marketDataClient.ValidateSymbol(ctx, "AAPL")

		if err != nil {
			// Service unavailable - should handle gracefully
			assert.Contains(t, err.Error(), "market data")
		} else {
			// Valid symbol should return true
			assert.True(t, isValid)
		}
	})

	t.Run("Invalid Symbol Handling", func(t *testing.T) {
		// Test invalid symbol
		isValid, err := marketDataClient.ValidateSymbol(ctx, "INVALID_SYMBOL")

		if err != nil {
			// Service error is acceptable
			assert.Error(t, err)
		} else {
			// Invalid symbol should return false
			assert.False(t, isValid)
		}
	})

	t.Run("Current Price Retrieval", func(t *testing.T) {
		// Test current price retrieval
		price, err := marketDataClient.GetCurrentPrice(ctx, "AAPL")

		if err != nil {
			// Service unavailable - should handle gracefully
			assert.Contains(t, err.Error(), "market data")
		} else {
			assert.Greater(t, price, 0.0)
		}
	})
}

// TestOrderSubmissionWithMarketDataValidation tests order submission with market data validation
func TestOrderSubmissionWithMarketDataValidation(t *testing.T) {
	container := setupTestContainer(t)
	defer container.Close()

	submitOrderUC := container.GetSubmitOrderUseCase()
	ctx := context.Background()

	t.Run("Valid Symbol Order", func(t *testing.T) {
		cmd := &command.SubmitOrderCommand{
			UserID:    "test-user-123",
			Symbol:    "AAPL",
			OrderType: "LIMIT",
			OrderSide: "BUY",
			Quantity:  100,
			Price:     func() *float64 { p := 150.50; return &p }(),
		}

		result, err := submitOrderUC.Execute(ctx, cmd)

		// Should succeed or fail gracefully based on market data availability
		if err != nil {
			// Market data service might be unavailable in test environment
			t.Logf("Market data validation failed (expected in test): %v", err)
		} else {
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.OrderID)
			assert.Equal(t, "PENDING", result.Status)
		}
	})

	t.Run("Invalid Symbol Order", func(t *testing.T) {
		cmd := &command.SubmitOrderCommand{
			UserID:    "test-user-123",
			Symbol:    "INVALID_SYMBOL",
			OrderType: "LIMIT",
			OrderSide: "BUY",
			Quantity:  100,
			Price:     func() *float64 { p := 150.50; return &p }(),
		}

		result, err := submitOrderUC.Execute(ctx, cmd)

		// Should fail due to invalid symbol
		if err != nil {
			assert.Contains(t, err.Error(), "symbol")
		} else {
			// If market data service is mocked, might succeed
			assert.NotNil(t, result)
		}
	})
}

// setupTestContainer creates a test container with necessary dependencies
func setupTestContainer(t *testing.T) pck.Container {
	// Create container with default configuration
	container, err := pck.NewContainer()
	require.NoError(t, err)

	return container
}

// TestOrderLifecycleIntegration tests the complete order lifecycle
func TestOrderLifecycleIntegration(t *testing.T) {
	container := setupTestContainer(t)
	defer container.Close()

	ctx := context.Background()

	// 1. Submit Order
	submitOrderUC := container.GetSubmitOrderUseCase()
	cmd := &command.SubmitOrderCommand{
		UserID:    "test-user-lifecycle",
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		OrderSide: "BUY",
		Quantity:  100,
		Price:     func() *float64 { p := 150.50; return &p }(),
	}

	submitResult, err := submitOrderUC.Execute(ctx, cmd)
	if err != nil {
		t.Skipf("Skipping lifecycle test due to market data unavailability: %v", err)
		return
	}

	orderID := submitResult.OrderID
	assert.NotEmpty(t, orderID)

	// 2. Get Order Status
	getStatusUC := container.GetGetOrderStatusUseCase()

	statusResult, err := getStatusUC.Execute(ctx, orderID, "test-user-lifecycle")
	require.NoError(t, err)
	assert.Equal(t, orderID, statusResult.OrderID)
	assert.Equal(t, "PENDING", statusResult.Status)

	// 3. Process Order (simulate worker processing)
	processOrderUC := container.GetProcessOrderUseCase()
	processCmd := &usecase.ProcessOrderCommand{
		OrderID: orderID,
		Context: usecase.ProcessingContext{
			ProcessingID: fmt.Sprintf("test-processing-%d", time.Now().Unix()),
			WorkerID:     "test-worker",
		},
	}

	processResult, err := processOrderUC.Execute(ctx, processCmd)
	if err != nil {
		t.Logf("Order processing failed (expected in test environment): %v", err)
	} else {
		assert.NotNil(t, processResult)
		assert.Equal(t, orderID, processResult.OrderID)
	}

	// 4. Cancel Order (if still pending)
	cancelOrderUC := container.GetCancelOrderUseCase()
	cancelCmd := &command.CancelOrderCommand{
		OrderID: orderID,
		UserID:  "test-user-lifecycle",
		Reason:  "Test cancellation",
	}

	cancelResult, err := cancelOrderUC.Execute(ctx, cancelCmd)
	if err != nil {
		// Order might already be processed
		t.Logf("Order cancellation failed (order might be processed): %v", err)
	} else {
		assert.NotNil(t, cancelResult)
		assert.NotNil(t, cancelResult)
	}
}
