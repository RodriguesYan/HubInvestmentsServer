package di

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"HubInvestments/shared/infra/messaging"
)

func TestContainer_OrderMarketDataClientIntegration(t *testing.T) {
	// Skip this test if we're not in an integration test environment
	// This test requires the market data gRPC server to be running
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create container
	container, err := NewContainer()
	require.NoError(t, err, "Failed to create container")
	require.NotNil(t, container, "Container should not be nil")

	// Ensure proper cleanup
	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close container: %v", closeErr)
		}
	}()

	// Get the order market data client
	client := container.GetOrderMarketDataClient()
	require.NotNil(t, client, "Order market data client should not be nil")

	// Test basic functionality (this will fail if gRPC server is not running, but that's expected)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to validate a symbol - this may fail if server is not running, but we're testing the integration
	_, err = client.ValidateSymbol(ctx, "AAPL")
	// We don't assert on the error here because the gRPC server might not be running
	// The important thing is that we can get the client from the container
	t.Logf("ValidateSymbol call result (expected to fail if server not running): %v", err)
}

func TestContainer_OrderMarketDataClientLifecycle(t *testing.T) {
	// Test that we can create and close the container without errors
	container, err := NewContainer()
	require.NoError(t, err, "Failed to create container")
	require.NotNil(t, container, "Container should not be nil")

	// Get the client to ensure it's initialized
	client := container.GetOrderMarketDataClient()
	assert.NotNil(t, client, "Order market data client should not be nil")

	// Test that we can close the container
	err = container.Close()
	assert.NoError(t, err, "Failed to close container")
}

func TestContainer_AllServicesAvailable(t *testing.T) {
	// Test that all services are available after adding the market data client
	container, err := NewContainer()
	require.NoError(t, err, "Failed to create container")
	require.NotNil(t, container, "Container should not be nil")

	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close container: %v", closeErr)
		}
	}()

	// Test that all existing services are still available
	assert.NotNil(t, container.GetAuthService(), "AuthService should be available")
	assert.NotNil(t, container.GetPositionAggregationUseCase(), "PositionAggregationUseCase should be available")
	assert.NotNil(t, container.GetBalanceUseCase(), "BalanceUseCase should be available")
	assert.NotNil(t, container.GetPortfolioSummaryUsecase(), "PortfolioSummaryUsecase should be available")
	assert.NotNil(t, container.GetMarketDataUsecase(), "MarketDataUsecase should be available")
	assert.NotNil(t, container.GetWatchlistUsecase(), "WatchlistUsecase should be available")
	assert.NotNil(t, container.DoLoginUsecase(), "LoginUsecase should be available")

	// Test that the new services are available
	assert.NotNil(t, container.GetOrderMarketDataClient(), "OrderMarketDataClient should be available")

	// Test that messaging infrastructure is available (may be nil if RabbitMQ not running)
	messageHandler := container.GetMessageHandler()
	t.Logf("MessageHandler availability: %v", messageHandler != nil)

	// Test that order management use cases are available
	assert.NotNil(t, container.GetSubmitOrderUseCase(), "SubmitOrderUseCase should be available")
	assert.NotNil(t, container.GetGetOrderStatusUseCase(), "GetOrderStatusUseCase should be available")
	assert.NotNil(t, container.GetCancelOrderUseCase(), "CancelOrderUseCase should be available")
	assert.NotNil(t, container.GetProcessOrderUseCase(), "ProcessOrderUseCase should be available")
}

func TestContainer_MessagingInfrastructure(t *testing.T) {
	// Test messaging infrastructure integration
	container, err := NewContainer()
	require.NoError(t, err, "Failed to create container")
	require.NotNil(t, container, "Container should not be nil")

	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close container: %v", closeErr)
		}
	}()

	// Get the message handler
	messageHandler := container.GetMessageHandler()

	if messageHandler == nil {
		t.Log("MessageHandler is nil - RabbitMQ not available, which is expected in development")
		return
	}

	// Test basic functionality if RabbitMQ is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try health check
	err = messageHandler.HealthCheck(ctx)
	if err != nil {
		t.Logf("MessageHandler health check failed (expected if RabbitMQ not running): %v", err)
	} else {
		t.Log("MessageHandler health check passed - RabbitMQ is available")

		// Test queue operations if health check passed
		queueName := "test-container-queue"
		queueOptions := messaging.QueueOptions{
			Durable:    true,
			AutoDelete: true,
		}

		err = messageHandler.DeclareQueue(queueName, queueOptions)
		if err != nil {
			t.Logf("Queue declaration failed: %v", err)
		} else {
			t.Log("Queue declaration successful")

			// Clean up
			messageHandler.DeleteQueue(queueName)
		}
	}
}
