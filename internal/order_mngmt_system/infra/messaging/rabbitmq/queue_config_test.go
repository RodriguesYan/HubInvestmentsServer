package rabbitmq

import (
	"context"
	"testing"
	"time"

	"HubInvestments/shared/infra/messaging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDefaultQueueNames(t *testing.T) {
	queueNames := DefaultQueueNames()

	assert.Equal(t, "orders.submit", queueNames.OrdersSubmit)
	assert.Equal(t, "orders.processing", queueNames.OrdersProcessing)
	assert.Equal(t, "orders.settlement", queueNames.OrdersSettlement)
	assert.Equal(t, "orders.status", queueNames.OrdersStatus)
	assert.Equal(t, "orders.dlq", queueNames.OrdersDLQ)
	assert.Equal(t, "orders.retry", queueNames.OrdersRetry)
	assert.Equal(t, "orders.exchange", queueNames.OrdersExchange)
	assert.Equal(t, "orders.dlq.exchange", queueNames.DLQExchange)
}

func TestDefaultRetryConfig(t *testing.T) {
	retryConfig := DefaultRetryConfig()

	expectedIntervals := []time.Duration{
		5 * time.Minute,
		15 * time.Minute,
		60 * time.Minute,
		360 * time.Minute,
	}

	assert.Equal(t, expectedIntervals, retryConfig.RetryIntervals)
	assert.Equal(t, 4, retryConfig.MaxRetries)
}

func TestNewOrderQueueManager(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)

	assert.NotNil(t, queueManager)
	assert.Equal(t, mockHandler, queueManager.messageHandler)
	assert.Equal(t, DefaultQueueNames(), queueManager.queueNames)
	assert.Equal(t, DefaultRetryConfig(), queueManager.retryConfig)
}

func TestNewOrderQueueManagerWithConfig(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	customQueueNames := QueueNames{
		OrdersSubmit:     "custom.submit",
		OrdersProcessing: "custom.processing",
		OrdersSettlement: "custom.settlement",
		OrdersStatus:     "custom.status",
		OrdersDLQ:        "custom.dlq",
		OrdersRetry:      "custom.retry",
		OrdersExchange:   "custom.exchange",
		DLQExchange:      "custom.dlq.exchange",
	}
	customRetryConfig := RetryConfig{
		RetryIntervals: []time.Duration{1 * time.Minute, 5 * time.Minute},
		MaxRetries:     2,
	}

	queueManager := NewOrderQueueManagerWithConfig(mockHandler, customQueueNames, customRetryConfig)

	assert.NotNil(t, queueManager)
	assert.Equal(t, mockHandler, queueManager.messageHandler)
	assert.Equal(t, customQueueNames, queueManager.queueNames)
	assert.Equal(t, customRetryConfig, queueManager.retryConfig)
}

func TestSetupAllQueues_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	// Mock successful queue declarations
	mockHandler.On("DeclareQueue", mock.AnythingOfType("string"), mock.AnythingOfType("messaging.QueueOptions")).Return(nil)

	err := queueManager.SetupAllQueues(ctx)

	assert.NoError(t, err)

	// Verify all queues were declared (4 primary + 2 management = 6 total)
	mockHandler.AssertNumberOfCalls(t, "DeclareQueue", 6)
}

func TestSetupPrimaryQueues_VerifyConfiguration(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	// Capture queue declarations to verify configuration
	var capturedCalls []mock.Call
	mockHandler.On("DeclareQueue", mock.AnythingOfType("string"), mock.AnythingOfType("messaging.QueueOptions")).
		Return(nil).
		Run(func(args mock.Arguments) {
			capturedCalls = append(capturedCalls, mock.Call{
				Method:    "DeclareQueue",
				Arguments: args,
			})
		})

	err := queueManager.setupPrimaryQueues(ctx)
	assert.NoError(t, err)

	// Verify submit queue configuration
	submitCall := capturedCalls[0]
	assert.Equal(t, "orders.submit", submitCall.Arguments[0])

	submitOptions := submitCall.Arguments[1].(messaging.QueueOptions)
	assert.True(t, submitOptions.Durable)
	assert.False(t, submitOptions.AutoDelete)
	assert.False(t, submitOptions.Exclusive)
	assert.Contains(t, submitOptions.Arguments, "x-dead-letter-exchange")
	assert.Contains(t, submitOptions.Arguments, "x-message-ttl")
}

func TestPublishToSubmitQueue(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	orderMessage := []byte(`{"order_id":"123","symbol":"AAPL"}`)
	messageID := "msg-123"

	// Mock successful publish
	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		return options.QueueName == "orders.submit" &&
			string(options.Message) == string(orderMessage) &&
			options.MessageID == messageID &&
			options.Persistent == true &&
			options.Priority == 5
	})).Return(nil)

	err := queueManager.PublishToSubmitQueue(ctx, orderMessage, messageID)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestPublishToProcessingQueue(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	orderMessage := []byte(`{"order_id":"123","symbol":"AAPL"}`)
	messageID := "msg-123"
	priority := uint8(8)

	// Mock successful publish
	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		return options.QueueName == "orders.processing" &&
			string(options.Message) == string(orderMessage) &&
			options.MessageID == messageID &&
			options.Priority == priority
	})).Return(nil)

	err := queueManager.PublishToProcessingQueue(ctx, orderMessage, messageID, priority)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestPublishToRetryQueue_WithRetryAttempt(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	orderMessage := []byte(`{"order_id":"123","symbol":"AAPL"}`)
	messageID := "msg-123"
	retryAttempt := 1 // Second retry (15 minutes)

	expectedTTL := int64(15 * time.Minute / time.Millisecond)

	// Mock successful publish
	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		return options.QueueName == "orders.retry" &&
			string(options.Message) == string(orderMessage) &&
			options.MessageID == messageID &&
			options.TTL == expectedTTL &&
			options.Headers["retry_attempt"] == retryAttempt
	})).Return(nil)

	err := queueManager.PublishToRetryQueue(ctx, orderMessage, messageID, retryAttempt)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestPublishToRetryQueue_ExceedsConfiguredRetries(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	orderMessage := []byte(`{"order_id":"123","symbol":"AAPL"}`)
	messageID := "msg-123"
	retryAttempt := 10 // Exceeds configured retry intervals

	// Should use the last configured interval (6 hours)
	expectedTTL := int64(360 * time.Minute / time.Millisecond)

	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		return options.TTL == expectedTTL
	})).Return(nil)

	err := queueManager.PublishToRetryQueue(ctx, orderMessage, messageID, retryAttempt)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestPublishStatusUpdate(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	statusMessage := []byte(`{"order_id":"123","status":"EXECUTED"}`)
	orderID := "123"

	// Mock successful publish
	mockHandler.On("PublishWithOptions", ctx, mock.MatchedBy(func(options messaging.PublishOptions) bool {
		return options.QueueName == "orders.status" &&
			string(options.Message) == string(statusMessage) &&
			options.CorrelationID == orderID &&
			options.Priority == 8 &&
			options.Headers["order_id"] == orderID
	})).Return(nil)

	err := queueManager.PublishStatusUpdate(ctx, statusMessage, orderID)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestGetQueueInfo(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	// Mock queue info responses
	mockQueueInfo := &messaging.QueueInfo{
		Name:      "orders.submit",
		Messages:  5,
		Consumers: 1,
	}

	mockHandler.On("QueueInfo", mock.AnythingOfType("string")).Return(mockQueueInfo, nil)

	queueInfoMap, err := queueManager.GetQueueInfo(ctx)

	assert.NoError(t, err)
	assert.Len(t, queueInfoMap, 6) // 6 queues total
	assert.Contains(t, queueInfoMap, "orders.submit")
	assert.Contains(t, queueInfoMap, "orders.processing")
	assert.Contains(t, queueInfoMap, "orders.settlement")
	assert.Contains(t, queueInfoMap, "orders.status")
	assert.Contains(t, queueInfoMap, "orders.dlq")
	assert.Contains(t, queueInfoMap, "orders.retry")

	mockHandler.AssertExpectations(t)
}

func TestPurgeAllQueues(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	// Mock successful purge for all queues
	mockHandler.On("PurgeQueue", mock.AnythingOfType("string")).Return(nil)

	err := queueManager.PurgeAllQueues(ctx)

	assert.NoError(t, err)

	// Verify all 6 queues were purged
	mockHandler.AssertNumberOfCalls(t, "PurgeQueue", 6)
	mockHandler.AssertExpectations(t)
}

func TestQueueManager_HealthCheck_Success(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	// Mock successful health check and queue info
	mockHandler.On("HealthCheck", ctx).Return(nil)
	mockHandler.On("QueueInfo", mock.AnythingOfType("string")).Return(&messaging.QueueInfo{}, nil)

	err := queueManager.HealthCheck(ctx)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestHealthCheck_MessageHandlerFails(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	// Mock failed health check
	mockHandler.On("HealthCheck", ctx).Return(messaging.ErrConnectionClosed)

	err := queueManager.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message handler health check failed")
	mockHandler.AssertExpectations(t)
}

func TestHealthCheck_QueueNotAccessible(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)
	ctx := context.Background()

	// Mock successful handler health check but failed queue info
	mockHandler.On("HealthCheck", ctx).Return(nil)
	mockHandler.On("QueueInfo", "orders.submit").Return((*messaging.QueueInfo)(nil), messaging.ErrQueueNotFound)

	err := queueManager.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue orders.submit is not accessible")
	mockHandler.AssertExpectations(t)
}

func TestGetQueueNames(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)

	queueNames := queueManager.GetQueueNames()

	assert.Equal(t, DefaultQueueNames(), queueNames)
}

func TestGetRetryConfig(t *testing.T) {
	mockHandler := &SharedMockMessageHandler{}
	queueManager := NewOrderQueueManager(mockHandler)

	retryConfig := queueManager.GetRetryConfig()

	assert.Equal(t, DefaultRetryConfig(), retryConfig)
}

// Integration test with real RabbitMQ (skipped if RabbitMQ not available)
func TestOrderQueueManager_Integration(t *testing.T) {
	// Try to create a real RabbitMQ connection
	realHandler, err := messaging.NewRabbitMQMessageHandlerWithDefaults()
	if err != nil {
		t.Skipf("Skipping integration test: RabbitMQ not available: %v", err)
		return
	}
	defer realHandler.Close()

	queueManager := NewOrderQueueManager(realHandler)
	ctx := context.Background()

	// Test setup
	err = queueManager.SetupAllQueues(ctx)
	assert.NoError(t, err)

	// Test health check
	err = queueManager.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test queue info
	queueInfoMap, err := queueManager.GetQueueInfo(ctx)
	assert.NoError(t, err)
	assert.Len(t, queueInfoMap, 6)

	// Test publishing
	testMessage := []byte(`{"test":"message"}`)
	err = queueManager.PublishToSubmitQueue(ctx, testMessage, "test-msg-id")
	assert.NoError(t, err)

	// Clean up - purge test messages
	err = queueManager.PurgeAllQueues(ctx)
	assert.NoError(t, err)
}
