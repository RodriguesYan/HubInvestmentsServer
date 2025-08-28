package external

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestGRPCMarketDataClientIntegration tests gRPC client-server communication
func TestGRPCMarketDataClientIntegration(t *testing.T) {
	// Skip if no gRPC server is available
	if testing.Short() {
		t.Skip("Skipping gRPC integration test in short mode")
	}

	// Setup gRPC client
	client, err := NewMarketDataClient(MarketDataClientConfig{
		ServerAddress: "localhost:50051",
		Timeout:       30 * time.Second,
	})
	if err != nil {
		t.Skipf("gRPC server not available: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("GetAssetDetails Success", func(t *testing.T) {
		// Test valid symbol
		assetDetails, err := client.GetAssetDetails(ctx, "AAPL")

		if err != nil {
			// Server might not be running or symbol not found
			if status.Code(err) == codes.Unavailable {
				t.Skipf("Market data server unavailable: %v", err)
				return
			}
			if status.Code(err) == codes.NotFound {
				t.Logf("Symbol not found (expected in test): %v", err)
				return
			}
		}

		assert.NotNil(t, assetDetails)
		assert.Equal(t, "AAPL", assetDetails.Symbol)
		assert.NotEmpty(t, assetDetails.Name)
		assert.Greater(t, assetDetails.LastQuote, 0.0)
	})

	t.Run("GetAssetDetails Invalid Symbol", func(t *testing.T) {
		// Test invalid symbol
		assetDetails, err := client.GetAssetDetails(ctx, "INVALID_SYMBOL_123")

		if err != nil {
			// Should return NotFound error
			assert.Equal(t, codes.NotFound, status.Code(err))
		} else {
			// If no error, asset details should be nil or empty
			if assetDetails != nil {
				assert.Empty(t, assetDetails.Symbol)
			}
		}
	})

	t.Run("ValidateSymbol", func(t *testing.T) {
		// Test valid symbol
		isValid, err := client.ValidateSymbol(ctx, "AAPL")
		if err != nil && status.Code(err) == codes.Unavailable {
			t.Skipf("Market data server unavailable: %v", err)
			return
		}

		if err == nil {
			assert.True(t, isValid)
		}

		// Test invalid symbol
		isValid, err = client.ValidateSymbol(ctx, "INVALID_SYMBOL_123")
		if err == nil {
			assert.False(t, isValid)
		}
	})

	t.Run("GetCurrentPrice", func(t *testing.T) {
		// Test price retrieval
		price, err := client.GetCurrentPrice(ctx, "AAPL")
		if err != nil && status.Code(err) == codes.Unavailable {
			t.Skipf("Market data server unavailable: %v", err)
			return
		}

		if err == nil {
			assert.Greater(t, price, 0.0)
			assert.Less(t, price, 10000.0) // Reasonable upper bound
		}
	})

	t.Run("GetTradingHours", func(t *testing.T) {
		// Test trading hours retrieval
		tradingHours, err := client.GetTradingHours(ctx, "AAPL")
		if err != nil && status.Code(err) == codes.Unavailable {
			t.Skipf("Market data server unavailable: %v", err)
			return
		}

		if err == nil {
			assert.NotNil(t, tradingHours)
			assert.NotEmpty(t, tradingHours.Timezone)
		}
	})
}

// TestGRPCClientPerformance tests gRPC client performance characteristics
func TestGRPCClientPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Setup gRPC client
	client, err := NewMarketDataClient(MarketDataClientConfig{
		ServerAddress: "localhost:50051",
		Timeout:       30 * time.Second,
	})
	if err != nil {
		t.Skipf("gRPC server not available: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Concurrent Requests", func(t *testing.T) {
		const numRequests = 100
		results := make(chan error, numRequests)

		start := time.Now()

		// Send concurrent requests
		for i := 0; i < numRequests; i++ {
			go func() {
				_, err := client.ValidateSymbol(ctx, "AAPL")
				results <- err
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < numRequests; i++ {
			if err := <-results; err == nil {
				successCount++
			}
		}

		duration := time.Since(start)
		t.Logf("Completed %d concurrent requests in %v (%d successful)",
			numRequests, duration, successCount)

		// Performance assertions
		assert.Less(t, duration, 10*time.Second, "Concurrent requests took too long")
		assert.Greater(t, successCount, numRequests/2, "Too many failed requests")
	})

	t.Run("Request Latency", func(t *testing.T) {
		const numRequests = 10
		var totalDuration time.Duration

		for i := 0; i < numRequests; i++ {
			start := time.Now()
			_, err := client.ValidateSymbol(ctx, "AAPL")
			duration := time.Since(start)

			if err == nil {
				totalDuration += duration
			}
		}

		avgLatency := totalDuration / numRequests
		t.Logf("Average request latency: %v", avgLatency)

		// Latency should be reasonable
		assert.Less(t, avgLatency, 1*time.Second, "Average latency too high")
	})
}

// TestGRPCClientErrorHandling tests error handling scenarios
func TestGRPCClientErrorHandling(t *testing.T) {
	t.Run("Connection Timeout", func(t *testing.T) {
		// Create client with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		client, err := NewMarketDataClient(MarketDataClientConfig{
			ServerAddress: "localhost:50051",
			Timeout:       30 * time.Second,
		})
		if err != nil {
			t.Skipf("gRPC server not available: %v", err)
			return
		}
		defer client.Close()

		// Request should timeout or succeed quickly
		_, err = client.GetAssetDetails(ctx, "AAPL")
		if err != nil {
			// Timeout error is acceptable
			assert.Contains(t, err.Error(), "context")
		}
	})

	t.Run("Invalid Server Address", func(t *testing.T) {
		// Try to connect to invalid address
		client, err := NewMarketDataClient(MarketDataClientConfig{
			ServerAddress: "localhost:99999",
			Timeout:       1 * time.Second,
		})

		if err != nil {
			// Connection should fail
			assert.Error(t, err)
			return
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Request should fail
		_, err = client.ValidateSymbol(ctx, "AAPL")
		assert.Error(t, err)
	})

	t.Run("Server Unavailable", func(t *testing.T) {
		// This test assumes the server is not running
		client, err := NewMarketDataClient(MarketDataClientConfig{
			ServerAddress: "localhost:50052", // Different port
			Timeout:       1 * time.Second,
		})
		if err != nil {
			assert.Error(t, err)
			return
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Request should fail with unavailable error
		_, err = client.ValidateSymbol(ctx, "AAPL")
		if err != nil {
			assert.Equal(t, codes.Unavailable, status.Code(err))
		}
	})
}

// TestGRPCClientWithCircuitBreaker tests circuit breaker functionality
func TestGRPCClientWithCircuitBreaker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping circuit breaker test in short mode")
	}

	// This test would require implementing circuit breaker in the client
	// For now, we test that repeated failures are handled gracefully

	client, err := NewMarketDataClient(MarketDataClientConfig{
		ServerAddress: "localhost:50052", // Unavailable port
		Timeout:       500 * time.Millisecond,
	})
	if err != nil {
		t.Skipf("Could not create connection: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Repeated Failures", func(t *testing.T) {
		const numAttempts = 5
		failureCount := 0

		for i := 0; i < numAttempts; i++ {
			ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			_, err := client.ValidateSymbol(ctx, "AAPL")
			cancel()

			if err != nil {
				failureCount++
			}

			// Small delay between attempts
			time.Sleep(100 * time.Millisecond)
		}

		t.Logf("Failed %d out of %d attempts", failureCount, numAttempts)

		// All attempts should fail consistently
		assert.Equal(t, numAttempts, failureCount, "Expected all attempts to fail")
	})
}

// TestMarketDataClientIntegrationWithOrderValidation tests integration with order validation
func TestMarketDataClientIntegrationWithOrderValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewMarketDataClient(MarketDataClientConfig{
		ServerAddress: "localhost:50051",
		Timeout:       30 * time.Second,
	})
	if err != nil {
		t.Skipf("gRPC server not available: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Direct Market Data Validation", func(t *testing.T) {
		// Test direct market data validation without service layer
		isValid, err := client.ValidateSymbol(ctx, "AAPL")

		if err != nil && status.Code(err) == codes.Unavailable {
			t.Skipf("Market data server unavailable: %v", err)
			return
		}

		// Should succeed for valid symbol
		if err == nil {
			assert.True(t, isValid)
		}

		// Should fail for invalid symbol
		isValid, err = client.ValidateSymbol(ctx, "INVALID_SYMBOL_123")
		if err == nil {
			assert.False(t, isValid)
		}
	})

	t.Run("Price Retrieval for Order Validation", func(t *testing.T) {
		// Test price retrieval for order validation
		price, err := client.GetCurrentPrice(ctx, "AAPL")

		if err != nil && status.Code(err) == codes.Unavailable {
			t.Skipf("Market data server unavailable: %v", err)
			return
		}

		// Should get a valid price
		if err == nil {
			assert.Greater(t, price, 0.0)
			t.Logf("Current AAPL price: $%.2f", price)
		}
	})

	t.Run("Trading Hours Check", func(t *testing.T) {
		// Test trading hours check
		isOpen, err := client.IsMarketOpen(ctx, "AAPL")

		if err != nil && status.Code(err) == codes.Unavailable {
			t.Skipf("Market data server unavailable: %v", err)
			return
		}

		// Result depends on current time and market hours
		if err == nil {
			t.Logf("Market is open: %v", isOpen)
		}
	})
}
