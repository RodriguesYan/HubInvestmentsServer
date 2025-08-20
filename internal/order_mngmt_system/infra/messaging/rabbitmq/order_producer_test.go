package rabbitmq

import (
	"context"
	"encoding/json"
	"testing"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/shared/infra/messaging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewOrderProducer(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)

	assert.NotNil(t, producer)
	assert.NotNil(t, producer.queueManager)
	assert.Equal(t, mockHandler, producer.messageHandler)
}

func TestNewOrderProducerWithQueueManager(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	producer := NewOrderProducerWithQueueManager(queueManager)

	assert.NotNil(t, producer)
	assert.Equal(t, queueManager, producer.queueManager)
	assert.Equal(t, mockHandler, producer.messageHandler)
}

func TestPublishOrderForProcessing_Success(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	// Create test order
	order, err := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)
	assert.NoError(t, err)

	// Mock successful publish
	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		// Verify queue name and basic properties
		if options.QueueName != "orders.processing" {
			return false
		}
		if !options.Persistent {
			return false
		}
		if options.Priority != 8 { // Market orders should have high priority
			return false
		}

		// Verify message content
		var orderMessage OrderMessage
		err := json.Unmarshal(options.Message, &orderMessage)
		if err != nil {
			return false
		}

		return orderMessage.OrderID == order.ID() &&
			orderMessage.UserID == "user123" &&
			orderMessage.Symbol == "AAPL" &&
			orderMessage.OrderSide == "BUY" &&
			orderMessage.OrderType == "MARKET" &&
			orderMessage.Quantity == 100.0 &&
			orderMessage.MessageMetadata.MessageType == "order_processing"
	})).Return(nil)

	err = producer.PublishOrderForProcessing(ctx, order)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestPublishOrderForProcessing_NilOrder(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	err := producer.PublishOrderForProcessing(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order cannot be nil")
}

func TestPublishOrderForSubmission_Success(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	// Create test order
	price := 150.0
	order, err := domain.NewOrder("user123", "AAPL", domain.OrderSideSell, domain.OrderTypeLimit, 50.0, &price)
	assert.NoError(t, err)

	// Mock successful publish
	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		if options.QueueName != "orders.submit" {
			return false
		}

		var orderMessage OrderMessage
		err := json.Unmarshal(options.Message, &orderMessage)
		if err != nil {
			return false
		}

		return orderMessage.OrderID == order.ID() &&
			orderMessage.OrderSide == "SELL" &&
			orderMessage.OrderType == "LIMIT" &&
			*orderMessage.Price == 150.0 &&
			orderMessage.MessageMetadata.MessageType == "order_submission"
	})).Return(nil)

	err = producer.PublishOrderForSubmission(ctx, order)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestPublishOrderForRetry_Success(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	order, err := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)
	assert.NoError(t, err)

	retryAttempt := 2
	reason := "market_data_unavailable"

	// Mock successful publish
	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		if options.QueueName != "orders.retry" {
			return false
		}

		var orderMessage OrderMessage
		err := json.Unmarshal(options.Message, &orderMessage)
		if err != nil {
			return false
		}

		return orderMessage.OrderID == order.ID() &&
			orderMessage.MessageMetadata.RetryAttempt == retryAttempt &&
			orderMessage.MessageMetadata.MessageType == "order_retry"
	})).Return(nil)

	err = producer.PublishOrderForRetry(ctx, order, retryAttempt, reason)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestPublishOrderForRetry_InvalidRetryAttempt(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	order, err := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)
	assert.NoError(t, err)

	err = producer.PublishOrderForRetry(ctx, order, -1, "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retry attempt must be non-negative")
}

func TestPublishOrderStatusUpdate_Success(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	order, err := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)
	assert.NoError(t, err)

	// Mark order as executed
	executionPrice := 155.0
	err = order.MarkAsExecuted(executionPrice)
	assert.NoError(t, err)

	previousStatus := "PROCESSING"

	// Mock successful publish
	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		if options.QueueName != "orders.status" {
			return false
		}
		if options.Priority != 8 { // Status updates should have high priority
			return false
		}

		var statusUpdate OrderStatusUpdate
		err := json.Unmarshal(options.Message, &statusUpdate)
		if err != nil {
			return false
		}

		return statusUpdate.OrderID == order.ID() &&
			statusUpdate.UserID == "user123" &&
			statusUpdate.PreviousStatus == previousStatus &&
			statusUpdate.CurrentStatus == "EXECUTED" &&
			*statusUpdate.ExecutionPrice == executionPrice &&
			statusUpdate.MessageMetadata.MessageType == "status_update"
	})).Return(nil)

	err = producer.PublishOrderStatusUpdate(ctx, order, previousStatus)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestPublishBatchOrders_Success(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	// Create test orders
	order1, err := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)
	assert.NoError(t, err)

	price := 200.0
	order2, err := domain.NewOrder("user456", "GOOGL", domain.OrderSideSell, domain.OrderTypeLimit, 50.0, &price)
	assert.NoError(t, err)

	orders := []*domain.Order{order1, order2}

	// Mock successful publishes for both orders
	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		var orderMessage OrderMessage
		err := json.Unmarshal(options.Message, &orderMessage)
		if err != nil {
			return false
		}

		// Check if it's one of our orders
		return (orderMessage.OrderID == order1.ID() || orderMessage.OrderID == order2.ID()) &&
			orderMessage.MessageMetadata.MessageType == "batch_order_processing"
	})).Return(nil).Times(2)

	err = producer.PublishBatchOrders(ctx, orders)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestPublishBatchOrders_EmptyList(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	err := producer.PublishBatchOrders(ctx, []*domain.Order{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "orders list cannot be empty")
}

func TestPublishBatchOrders_NilOrderInList(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	order1, err := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)
	assert.NoError(t, err)

	orders := []*domain.Order{order1, nil} // Second order is nil

	err = producer.PublishBatchOrders(ctx, orders)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order at index 1 cannot be nil")
}

func TestCalculateMessagePriority(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)

	tests := []struct {
		name             string
		orderType        domain.OrderType
		orderSide        domain.OrderSide
		quantity         float64
		price            *float64
		expectedPriority uint8
	}{
		{
			name:             "Market order - high priority",
			orderType:        domain.OrderTypeMarket,
			orderSide:        domain.OrderSideBuy,
			quantity:         100.0,
			price:            nil,
			expectedPriority: 8,
		},
		{
			name:             "Stop loss order - high priority",
			orderType:        domain.OrderTypeStopLoss,
			orderSide:        domain.OrderSideSell,
			quantity:         50.0,
			price:            func() *float64 { p := 150.0; return &p }(),
			expectedPriority: 7,
		},
		{
			name:             "Large limit order - medium-high priority",
			orderType:        domain.OrderTypeLimit,
			orderSide:        domain.OrderSideBuy,
			quantity:         1000.0,                                      // Large quantity
			price:            func() *float64 { p := 150.0; return &p }(), // $150K value
			expectedPriority: 7,
		},
		{
			name:             "Medium limit order - medium priority",
			orderType:        domain.OrderTypeLimit,
			orderSide:        domain.OrderSideBuy,
			quantity:         100.0,
			price:            func() *float64 { p := 150.0; return &p }(), // $15K value
			expectedPriority: 6,
		},
		{
			name:             "Small limit order - normal priority",
			orderType:        domain.OrderTypeLimit,
			orderSide:        domain.OrderSideBuy,
			quantity:         10.0,
			price:            func() *float64 { p := 100.0; return &p }(), // $1K value
			expectedPriority: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := domain.NewOrder("user123", "AAPL", tt.orderSide, tt.orderType, tt.quantity, tt.price)
			assert.NoError(t, err)

			priority := producer.calculateMessagePriority(order)
			assert.Equal(t, tt.expectedPriority, priority)
		})
	}
}

func TestCreateOrderMessage(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)

	price := 150.0
	order, err := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, &price)
	assert.NoError(t, err)

	messageType := "test_message"
	processingStage := "test_stage"

	orderMessage, err := producer.createOrderMessage(order, messageType, processingStage)

	assert.NoError(t, err)
	assert.Equal(t, order.ID(), orderMessage.OrderID)
	assert.Equal(t, "user123", orderMessage.UserID)
	assert.Equal(t, "AAPL", orderMessage.Symbol)
	assert.Equal(t, "BUY", orderMessage.OrderSide)
	assert.Equal(t, "LIMIT", orderMessage.OrderType)
	assert.Equal(t, 100.0, orderMessage.Quantity)
	assert.Equal(t, &price, orderMessage.Price)
	assert.Equal(t, "PENDING", orderMessage.Status)
	assert.Equal(t, messageType, orderMessage.MessageMetadata.MessageType)
	assert.Equal(t, processingStage, orderMessage.MessageMetadata.ProcessingStage)
	assert.Equal(t, order.ID(), orderMessage.MessageMetadata.CorrelationID)
	assert.Equal(t, "order_producer", orderMessage.MessageMetadata.Source)
	assert.Equal(t, 0, orderMessage.MessageMetadata.RetryAttempt)
}

func TestHealthCheck(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	// Mock successful health check
	mockHandler.On("HealthCheck", ctx).Return(nil)
	mockHandler.On("QueueInfo", mock.AnythingOfType("string")).Return(&messaging.QueueInfo{}, nil)

	err := producer.HealthCheck(ctx)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestClose(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)

	err := producer.Close()

	assert.NoError(t, err)
}

func TestGetQueueManager(t *testing.T) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)

	queueManager := producer.GetQueueManager()

	assert.NotNil(t, queueManager)
	assert.Equal(t, producer.queueManager, queueManager)
}

// Integration test with real RabbitMQ (skipped if RabbitMQ not available)
func TestOrderProducer_Integration(t *testing.T) {
	// Try to create a real RabbitMQ connection
	realHandler, err := messaging.NewRabbitMQMessageHandlerWithDefaults()
	if err != nil {
		t.Skipf("Skipping integration test: RabbitMQ not available: %v", err)
		return
	}
	defer realHandler.Close()

	producer := NewOrderProducer(realHandler)
	ctx := context.Background()

	// Setup queues
	err = producer.queueManager.SetupAllQueues(ctx)
	assert.NoError(t, err)

	// Create test order
	order, err := domain.NewOrder("integration_test_user", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)
	assert.NoError(t, err)

	// Test publishing to processing queue
	err = producer.PublishOrderForProcessing(ctx, order)
	assert.NoError(t, err)

	// Test publishing to submission queue
	err = producer.PublishOrderForSubmission(ctx, order)
	assert.NoError(t, err)

	// Test status update
	err = producer.PublishOrderStatusUpdate(ctx, order, "PENDING")
	assert.NoError(t, err)

	// Test health check
	err = producer.HealthCheck(ctx)
	assert.NoError(t, err)

	// Clean up - purge test messages
	err = producer.queueManager.PurgeAllQueues(ctx)
	assert.NoError(t, err)
}

// Benchmark tests for performance validation
func BenchmarkPublishOrderForProcessing(b *testing.B) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)
	ctx := context.Background()

	order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)

	// Mock successful publish
	mockHandler.On("PublishWithOptions", ctx, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		producer.PublishOrderForProcessing(ctx, order)
	}
}

func BenchmarkCreateOrderMessage(b *testing.B) {
	mockHandler := &MockMessageHandler{}
	producer := NewOrderProducer(mockHandler)

	order, _ := domain.NewOrder("user123", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		producer.createOrderMessage(order, "test", "test")
	}
}
