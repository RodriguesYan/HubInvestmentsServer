package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"HubInvestments/shared/infra/messaging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockOrderMessageHandler is a mock implementation of OrderMessageHandler
type MockOrderMessageHandler struct {
	mock.Mock
}

func (m *MockOrderMessageHandler) HandleOrderMessage(ctx context.Context, message *OrderMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockOrderMessageHandler) HandleStatusUpdate(ctx context.Context, statusUpdate *OrderStatusUpdate) error {
	args := m.Called(ctx, statusUpdate)
	return args.Error(0)
}

// Unit Tests

func TestNewOrderConsumer(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)

	assert.NotNil(t, consumer)
	assert.NotNil(t, consumer.queueManager)
	assert.Equal(t, mockHandler, consumer.messageHandler)
	assert.Equal(t, mockOrderHandler, consumer.orderHandler)
	assert.NotNil(t, consumer.activeQueues)
	assert.NotNil(t, consumer.shutdownChan)
	assert.False(t, consumer.isRunning)
}

func TestNewOrderConsumerWithQueueManager(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)

	consumer := NewOrderConsumerWithQueueManager(queueManager, mockOrderHandler)

	assert.NotNil(t, consumer)
	assert.Equal(t, queueManager, consumer.queueManager)
	assert.Equal(t, mockHandler, consumer.messageHandler)
	assert.Equal(t, mockOrderHandler, consumer.orderHandler)
}

func TestDefaultConsumerConfig(t *testing.T) {
	config := DefaultConsumerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 5, config.ConcurrentWorkers)
	assert.Equal(t, 10, config.PrefetchCount)
	assert.True(t, config.RequeueOnError)
	assert.Equal(t, 5*time.Second, config.RetryDelay)
	assert.Equal(t, 3, config.MaxRetries)
}

func TestStartConsumers_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	config := DefaultConsumerConfig()

	ctx := context.Background()

	// Mock queue setup
	mockHandler.On("DeclareQueue", mock.AnythingOfType("string"), mock.AnythingOfType("messaging.QueueOptions")).Return(nil).Times(6)

	// Mock consumer creation for each queue
	mockHandler.On("Consume", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*rabbitmq.orderMessageConsumer")).Return(nil).Times(4)

	err := consumer.StartConsumers(ctx, config)

	assert.NoError(t, err)
	assert.True(t, consumer.IsRunning())
	mockHandler.AssertExpectations(t)
}

func TestStartConsumers_AlreadyRunning(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	consumer.isRunning = true

	ctx := context.Background()
	config := DefaultConsumerConfig()

	err := consumer.StartConsumers(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestStartConsumers_QueueSetupFails(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	config := DefaultConsumerConfig()

	ctx := context.Background()

	// Mock queue setup failure
	mockHandler.On("DeclareQueue", mock.AnythingOfType("string"), mock.AnythingOfType("messaging.QueueOptions")).Return(errors.New("queue setup failed")).Once()

	err := consumer.StartConsumers(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to setup queues")
	assert.False(t, consumer.IsRunning())
	mockHandler.AssertExpectations(t)
}

func TestStopConsumers_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	consumer.isRunning = true
	consumer.activeQueues["test-queue"] = true

	ctx := context.Background()

	err := consumer.StopConsumers(ctx)

	assert.NoError(t, err)
	assert.False(t, consumer.IsRunning())
	assert.Empty(t, consumer.activeQueues)
}

func TestStopConsumers_NotRunning(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	consumer.isRunning = false

	ctx := context.Background()

	err := consumer.StopConsumers(ctx)

	assert.NoError(t, err)
}

func TestDeserializeOrderMessage_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()

	data, err := json.Marshal(orderMessage)
	require.NoError(t, err)

	result, err := consumer.deserializeOrderMessage(data)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, orderMessage.OrderID, result.OrderID)
	assert.Equal(t, orderMessage.UserID, result.UserID)
	assert.Equal(t, orderMessage.Symbol, result.Symbol)
}

func TestDeserializeOrderMessage_InvalidJSON(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)

	result, err := consumer.deserializeOrderMessage([]byte("invalid json"))

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestDeserializeStatusUpdate_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	statusUpdate := CreateTestStatusUpdate()

	data, err := json.Marshal(statusUpdate)
	require.NoError(t, err)

	result, err := consumer.deserializeStatusUpdate(data)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, statusUpdate.OrderID, result.OrderID)
	assert.Equal(t, statusUpdate.CurrentStatus, result.CurrentStatus)
}

func TestValidateOrderMessage_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()

	err := consumer.validateOrderMessage(orderMessage)

	assert.NoError(t, err)
}

func TestValidateOrderMessage_NilMessage(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)

	err := consumer.validateOrderMessage(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestValidateOrderMessage_MissingOrderID(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.OrderID = ""

	err := consumer.validateOrderMessage(orderMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order ID is required")
}

func TestValidateOrderMessage_MissingUserID(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.UserID = ""

	err := consumer.validateOrderMessage(orderMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestValidateOrderMessage_MissingSymbol(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.Symbol = ""

	err := consumer.validateOrderMessage(orderMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symbol is required")
}

func TestValidateOrderMessage_InvalidQuantity(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.Quantity = 0

	err := consumer.validateOrderMessage(orderMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

func TestValidateOrderMessage_InvalidOrderSide(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.OrderSide = "INVALID"

	err := consumer.validateOrderMessage(orderMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order side")
}

func TestValidateOrderMessage_InvalidOrderType(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.OrderType = "INVALID"

	err := consumer.validateOrderMessage(orderMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order type")
}

func TestValidateOrderMessage_InvalidOrderStatus(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.Status = "INVALID"

	err := consumer.validateOrderMessage(orderMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order status")
}

func TestValidateOrderMessage_MissingMessageID(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.MessageMetadata.MessageID = ""

	err := consumer.validateOrderMessage(orderMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message ID is required")
}

func TestValidateOrderMessage_MissingCorrelationID(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.MessageMetadata.CorrelationID = ""

	err := consumer.validateOrderMessage(orderMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "correlation ID is required")
}

func TestValidateStatusUpdate_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	statusUpdate := CreateTestStatusUpdate()

	err := consumer.validateStatusUpdate(statusUpdate)

	assert.NoError(t, err)
}

func TestValidateStatusUpdate_NilUpdate(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)

	err := consumer.validateStatusUpdate(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestValidateStatusUpdate_MissingOrderID(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	statusUpdate := CreateTestStatusUpdate()
	statusUpdate.OrderID = ""

	err := consumer.validateStatusUpdate(statusUpdate)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order ID is required")
}

func TestValidateStatusUpdate_InvalidCurrentStatus(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	statusUpdate := CreateTestStatusUpdate()
	statusUpdate.CurrentStatus = "INVALID"

	err := consumer.validateStatusUpdate(statusUpdate)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid current status")
}

func TestHandleOrderProcessingMessage_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()

	data, err := json.Marshal(orderMessage)
	require.NoError(t, err)

	message := CreateTestMessage(data)
	ctx := context.Background()

	// Mock successful handling
	mockOrderHandler.On("HandleOrderMessage", ctx, mock.AnythingOfType("*rabbitmq.OrderMessage")).Return(nil)

	err = consumer.handleOrderProcessingMessage(ctx, message)

	assert.NoError(t, err)
	mockOrderHandler.AssertExpectations(t)
}

func TestHandleOrderProcessingMessage_DeserializationError(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	message := CreateTestMessage([]byte("invalid json"))
	ctx := context.Background()

	err := consumer.handleOrderProcessingMessage(ctx, message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to deserialize")
}

func TestHandleOrderProcessingMessage_ValidationError(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	orderMessage := CreateTestOrderMessage()
	orderMessage.OrderID = "" // Invalid

	data, err := json.Marshal(orderMessage)
	require.NoError(t, err)

	message := CreateTestMessage(data)
	ctx := context.Background()

	err = consumer.handleOrderProcessingMessage(ctx, message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestHandleStatusUpdateMessage_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	statusUpdate := CreateTestStatusUpdate()

	data, err := json.Marshal(statusUpdate)
	require.NoError(t, err)

	message := CreateTestMessage(data)
	ctx := context.Background()

	// Mock successful handling
	mockOrderHandler.On("HandleStatusUpdate", ctx, mock.AnythingOfType("*rabbitmq.OrderStatusUpdate")).Return(nil)

	err = consumer.handleStatusUpdateMessage(ctx, message)

	assert.NoError(t, err)
	mockOrderHandler.AssertExpectations(t)
}

func TestHealthCheck_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	consumer.isRunning = true
	consumer.activeQueues["test-queue"] = true

	ctx := context.Background()

	// Mock queue manager health check
	mockHandler.On("HealthCheck", ctx).Return(nil)

	// Mock QueueInfo calls for all queues (the health check verifies queue existence)
	mockQueueInfo := &messaging.QueueInfo{Name: "test-queue", Messages: 0, Consumers: 1}
	mockHandler.On("QueueInfo", mock.AnythingOfType("string")).Return(mockQueueInfo, nil).Maybe()

	err := consumer.HealthCheck(ctx)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestHealthCheck_NotRunning(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	consumer.isRunning = false

	ctx := context.Background()

	err := consumer.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestHealthCheck_NoActiveConsumers(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	consumer.isRunning = true
	// No consumers added

	ctx := context.Background()

	// Mock queue manager health check
	mockHandler.On("HealthCheck", ctx).Return(nil)

	// Mock QueueInfo calls for all queues (the health check verifies queue existence)
	mockQueueInfo := &messaging.QueueInfo{Name: "test-queue", Messages: 0, Consumers: 1}
	mockHandler.On("QueueInfo", mock.AnythingOfType("string")).Return(mockQueueInfo, nil).Maybe()

	err := consumer.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active consumers")
	mockHandler.AssertExpectations(t)
}

func TestGetActiveConsumers(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)
	consumer.activeQueues["queue1"] = true
	consumer.activeQueues["queue2"] = true

	active := consumer.GetActiveConsumers()

	assert.Len(t, active, 2)
	assert.True(t, active["queue1"])
	assert.True(t, active["queue2"])
}

func TestOrderConsumer_GetQueueManager(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	mockOrderHandler := &MockOrderMessageHandler{}

	consumer := NewOrderConsumer(mockHandler, mockOrderHandler)

	queueManager := consumer.GetQueueManager()

	assert.NotNil(t, queueManager)
	assert.Equal(t, consumer.queueManager, queueManager)
}

// Error classification tests

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"validation error", errors.New("validation failed"), true},
		{"invalid error", errors.New("invalid input"), true},
		{"required error", errors.New("field is required"), true},
		{"malformed error", errors.New("malformed data"), true},
		{"network error", errors.New("connection failed"), false},
		{"other error", errors.New("something else"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBusinessLogicError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"insufficient balance", errors.New("insufficient balance"), true},
		{"market closed", errors.New("market closed"), true},
		{"position limit", errors.New("position limit exceeded"), true},
		{"duplicate order", errors.New("duplicate order"), true},
		{"validation error", errors.New("validation failed"), false},
		{"network error", errors.New("connection failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBusinessLogicError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"connection error", errors.New("connection failed"), true},
		{"timeout error", errors.New("request timeout"), true},
		{"temporary error", errors.New("temporary failure"), true},
		{"unavailable error", errors.New("service unavailable"), true},
		{"service down", errors.New("service down"), true},
		{"validation error", errors.New("validation failed"), false},
		{"business logic error", errors.New("insufficient balance"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Integration test (skipped if RabbitMQ not available)
func TestOrderConsumer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Try to create a real RabbitMQ connection
	config := messaging.MessageHandlerConfig{
		URL:                "amqp://localhost:5672",
		MaxRetries:         3,
		RetryDelay:         1,
		ConnectionTimeout:  5,
		HeartbeatInterval:  30,
		PrefetchCount:      10,
		EnableConfirmation: true,
	}

	messageHandler, err := messaging.NewRabbitMQMessageHandler(config)
	if err != nil {
		t.Skipf("Skipping integration test: RabbitMQ not available: %v", err)
		return
	}
	defer messageHandler.Close()

	// Test with real RabbitMQ connection
	mockOrderHandler := &MockOrderMessageHandler{}
	consumer := NewOrderConsumer(messageHandler, mockOrderHandler)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test health check with real connection
	err = consumer.HealthCheck(ctx)
	// Should fail because consumer is not running
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}
