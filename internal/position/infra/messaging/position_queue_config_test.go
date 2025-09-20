package messaging

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"HubInvestments/shared/infra/messaging"
)

// MockMessageHandler for testing
type MockMessageHandler struct {
	queues            map[string]messaging.QueueOptions
	publishedMessages []messaging.PublishOptions
	declareQueueCalls []string
	shouldFailDeclare bool
	shouldFailPublish bool
}

func NewMockMessageHandler() *MockMessageHandler {
	return &MockMessageHandler{
		queues:            make(map[string]messaging.QueueOptions),
		publishedMessages: make([]messaging.PublishOptions, 0),
		declareQueueCalls: make([]string, 0),
	}
}

func (m *MockMessageHandler) Publish(ctx context.Context, queueName string, message []byte) error {
	if m.shouldFailPublish {
		return fmt.Errorf("mock publish failure")
	}
	options := messaging.PublishOptions{
		QueueName: queueName,
		Message:   message,
	}
	m.publishedMessages = append(m.publishedMessages, options)
	return nil
}

func (m *MockMessageHandler) PublishWithOptions(ctx context.Context, options messaging.PublishOptions) error {
	if m.shouldFailPublish {
		return fmt.Errorf("mock publish failure")
	}
	m.publishedMessages = append(m.publishedMessages, options)
	return nil
}

func (m *MockMessageHandler) Consume(ctx context.Context, queueName string, handler messaging.MessageConsumer) error {
	return nil
}

func (m *MockMessageHandler) DeclareQueue(name string, options messaging.QueueOptions) error {
	if m.shouldFailDeclare {
		return fmt.Errorf("mock declare queue failure")
	}
	m.queues[name] = options
	m.declareQueueCalls = append(m.declareQueueCalls, name)
	return nil
}

func (m *MockMessageHandler) DeleteQueue(queueName string) error {
	delete(m.queues, queueName)
	return nil
}

func (m *MockMessageHandler) PurgeQueue(queueName string) error {
	return nil
}

func (m *MockMessageHandler) QueueInfo(queueName string) (*messaging.QueueInfo, error) {
	return &messaging.QueueInfo{
		Name:     queueName,
		Messages: 0,
	}, nil
}

func (m *MockMessageHandler) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockMessageHandler) Close() error {
	return nil
}

func TestDefaultPositionQueueNames(t *testing.T) {
	names := DefaultPositionQueueNames()

	expectedNames := map[string]string{
		"PositionUpdates":   "positions.updates",
		"PositionsDLQ":      "positions.updates.dlq",
		"PositionsRetry":    "positions.retry",
		"PositionsExchange": "positions.exchange",
		"DLQExchange":       "positions.dlq.exchange",
	}

	if names.PositionUpdates != expectedNames["PositionUpdates"] {
		t.Errorf("Expected PositionUpdates %s, got %s", expectedNames["PositionUpdates"], names.PositionUpdates)
	}

	if names.PositionsDLQ != expectedNames["PositionsDLQ"] {
		t.Errorf("Expected PositionsDLQ %s, got %s", expectedNames["PositionsDLQ"], names.PositionsDLQ)
	}

	if names.PositionsRetry != expectedNames["PositionsRetry"] {
		t.Errorf("Expected PositionsRetry %s, got %s", expectedNames["PositionsRetry"], names.PositionsRetry)
	}

	if names.PositionsExchange != expectedNames["PositionsExchange"] {
		t.Errorf("Expected PositionsExchange %s, got %s", expectedNames["PositionsExchange"], names.PositionsExchange)
	}

	if names.DLQExchange != expectedNames["DLQExchange"] {
		t.Errorf("Expected DLQExchange %s, got %s", expectedNames["DLQExchange"], names.DLQExchange)
	}
}

func TestDefaultPositionRetryConfig(t *testing.T) {
	config := DefaultPositionRetryConfig()

	expectedRetries := 4
	if config.MaxRetries != expectedRetries {
		t.Errorf("Expected MaxRetries %d, got %d", expectedRetries, config.MaxRetries)
	}

	expectedIntervals := []time.Duration{
		2 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
		120 * time.Minute,
	}

	if len(config.RetryIntervals) != len(expectedIntervals) {
		t.Errorf("Expected %d retry intervals, got %d", len(expectedIntervals), len(config.RetryIntervals))
	}

	for i, expected := range expectedIntervals {
		if config.RetryIntervals[i] != expected {
			t.Errorf("Expected retry interval[%d] %v, got %v", i, expected, config.RetryIntervals[i])
		}
	}
}

func TestNewPositionQueueManager(t *testing.T) {
	mockHandler := NewMockMessageHandler()
	manager := NewPositionQueueManager(mockHandler)

	// MessageHandler is properly set (interface, can't compare directly)

	expectedNames := DefaultPositionQueueNames()
	if manager.queueNames != expectedNames {
		t.Errorf("Expected default queue names to be set")
	}

	expectedRetryConfig := DefaultPositionRetryConfig()
	if manager.retryConfig.MaxRetries != expectedRetryConfig.MaxRetries {
		t.Errorf("Expected default retry config to be set")
	}
}

func TestNewPositionQueueManagerWithConfig(t *testing.T) {
	mockHandler := NewMockMessageHandler()
	customNames := PositionQueueNames{
		PositionUpdates:   "custom.positions.updates",
		PositionsDLQ:      "custom.positions.dlq",
		PositionsRetry:    "custom.positions.retry",
		PositionsExchange: "custom.positions.exchange",
		DLQExchange:       "custom.dlq.exchange",
	}
	customRetryConfig := PositionRetryConfig{
		RetryIntervals: []time.Duration{1 * time.Minute},
		MaxRetries:     1,
	}

	manager := NewPositionQueueManagerWithConfig(mockHandler, customNames, customRetryConfig)

	if manager.queueNames.PositionUpdates != customNames.PositionUpdates {
		t.Errorf("Expected custom queue names to be set")
	}

	if manager.retryConfig.MaxRetries != customRetryConfig.MaxRetries {
		t.Errorf("Expected custom retry config to be set")
	}
}

func TestPositionQueueManager_SetupAllQueues_Success(t *testing.T) {
	mockHandler := NewMockMessageHandler()
	manager := NewPositionQueueManager(mockHandler)
	ctx := context.Background()

	err := manager.SetupAllQueues(ctx)
	if err != nil {
		t.Errorf("Expected successful queue setup, got error: %v", err)
	}

	// Verify all expected queues were declared
	expectedQueues := []string{
		"positions.updates",
		"positions.updates.dlq",
		"positions.retry",
	}

	for _, expectedQueue := range expectedQueues {
		if _, exists := mockHandler.queues[expectedQueue]; !exists {
			t.Errorf("Expected queue %s to be declared", expectedQueue)
		}
	}

	// Verify position updates queue configuration
	updatesOptions := mockHandler.queues["positions.updates"]
	if !updatesOptions.Durable {
		t.Errorf("Expected positions.updates queue to be durable")
	}

	if updatesOptions.AutoDelete {
		t.Errorf("Expected positions.updates queue to not auto-delete")
	}

	// Verify DLQ routing is configured
	dlxExchange, exists := updatesOptions.Arguments["x-dead-letter-exchange"]
	if !exists {
		t.Errorf("Expected positions.updates to have DLX configured")
	}

	if dlxExchange != "positions.dlq.exchange" {
		t.Errorf("Expected DLX to be positions.dlq.exchange, got %v", dlxExchange)
	}

	// Verify message TTL is configured
	ttl, exists := updatesOptions.Arguments["x-message-ttl"]
	if !exists {
		t.Errorf("Expected positions.updates to have TTL configured")
	}

	expectedTTL := int64(6 * time.Hour / time.Millisecond)
	if ttl != expectedTTL {
		t.Errorf("Expected TTL %d, got %v", expectedTTL, ttl)
	}

	// Verify queue length limit
	maxLength, exists := updatesOptions.Arguments["x-max-length"]
	if !exists {
		t.Errorf("Expected positions.updates to have max length configured")
	}

	if maxLength != 100000 {
		t.Errorf("Expected max length 100000, got %v", maxLength)
	}
}

func TestPositionQueueManager_SetupAllQueues_PrimaryQueueFailure(t *testing.T) {
	mockHandler := NewMockMessageHandler()
	mockHandler.shouldFailDeclare = true
	manager := NewPositionQueueManager(mockHandler)
	ctx := context.Background()

	err := manager.SetupAllQueues(ctx)
	if err == nil {
		t.Errorf("Expected error when primary queue setup fails")
	}

	if !strings.Contains(err.Error(), "failed to setup primary position queues") {
		t.Errorf("Expected primary queue failure error, got: %v", err)
	}
}

func TestPositionQueueManager_PublishToPositionUpdatesQueue_Success(t *testing.T) {
	mockHandler := NewMockMessageHandler()
	manager := NewPositionQueueManager(mockHandler)
	ctx := context.Background()

	message := []byte(`{"order_id":"123","user_id":"456","symbol":"AAPL","quantity":100}`)
	messageID := "test-message-123"

	err := manager.PublishToPositionUpdatesQueue(ctx, message, messageID)
	if err != nil {
		t.Errorf("Expected successful publish, got error: %v", err)
	}

	if len(mockHandler.publishedMessages) != 1 {
		t.Errorf("Expected 1 published message, got %d", len(mockHandler.publishedMessages))
	}

	publishedMsg := mockHandler.publishedMessages[0]
	if publishedMsg.QueueName != "positions.updates" {
		t.Errorf("Expected queue name positions.updates, got %s", publishedMsg.QueueName)
	}

	if publishedMsg.MessageID != messageID {
		t.Errorf("Expected message ID %s, got %s", messageID, publishedMsg.MessageID)
	}

	if publishedMsg.Priority != 7 {
		t.Errorf("Expected priority 7 for position updates, got %d", publishedMsg.Priority)
	}

	if !publishedMsg.Persistent {
		t.Errorf("Expected message to be persistent")
	}

	// Verify headers
	if msgType, exists := publishedMsg.Headers["message_type"]; !exists || msgType != "position_update" {
		t.Errorf("Expected message_type header to be position_update, got %v", msgType)
	}
}

func TestPositionQueueManager_PublishToRetryQueue_Success(t *testing.T) {
	mockHandler := NewMockMessageHandler()
	manager := NewPositionQueueManager(mockHandler)
	ctx := context.Background()

	message := []byte(`{"order_id":"123","error":"temporary failure"}`)
	messageID := "retry-message-123"
	retryAttempt := 2

	err := manager.PublishToRetryQueue(ctx, message, messageID, retryAttempt)
	if err != nil {
		t.Errorf("Expected successful publish to retry queue, got error: %v", err)
	}

	if len(mockHandler.publishedMessages) != 1 {
		t.Errorf("Expected 1 published message, got %d", len(mockHandler.publishedMessages))
	}

	publishedMsg := mockHandler.publishedMessages[0]
	if publishedMsg.QueueName != "positions.retry" {
		t.Errorf("Expected queue name positions.retry, got %s", publishedMsg.QueueName)
	}

	// Verify retry-specific headers
	if attempt, exists := publishedMsg.Headers["retry_attempt"]; !exists || attempt != retryAttempt {
		t.Errorf("Expected retry_attempt header to be %d, got %v", retryAttempt, attempt)
	}

	if originalQueue, exists := publishedMsg.Headers["original_queue"]; !exists || originalQueue != "positions.updates" {
		t.Errorf("Expected original_queue header to be positions.updates, got %v", originalQueue)
	}

	// Verify TTL is set to second retry interval (10 minutes)
	expectedTTL := int64(10 * time.Minute / time.Millisecond)
	if publishedMsg.TTL != expectedTTL {
		t.Errorf("Expected TTL %d ms for retry attempt 2, got %d ms", expectedTTL, publishedMsg.TTL)
	}
}

func TestPositionQueueManager_PublishToDLQ_Success(t *testing.T) {
	mockHandler := NewMockMessageHandler()
	manager := NewPositionQueueManager(mockHandler)
	ctx := context.Background()

	message := []byte(`{"order_id":"123","failed_permanently":true}`)
	messageID := "dlq-message-123"
	failureReason := "exceeded max retries"

	err := manager.PublishToDLQ(ctx, message, messageID, failureReason)
	if err != nil {
		t.Errorf("Expected successful publish to DLQ, got error: %v", err)
	}

	if len(mockHandler.publishedMessages) != 1 {
		t.Errorf("Expected 1 published message, got %d", len(mockHandler.publishedMessages))
	}

	publishedMsg := mockHandler.publishedMessages[0]
	if publishedMsg.QueueName != "positions.updates.dlq" {
		t.Errorf("Expected queue name positions.updates.dlq, got %s", publishedMsg.QueueName)
	}

	if publishedMsg.Priority != 1 {
		t.Errorf("Expected priority 1 for DLQ messages, got %d", publishedMsg.Priority)
	}

	// Verify DLQ-specific headers
	if reason, exists := publishedMsg.Headers["failure_reason"]; !exists || reason != failureReason {
		t.Errorf("Expected failure_reason header to be %s, got %v", failureReason, reason)
	}

	if msgType, exists := publishedMsg.Headers["message_type"]; !exists || msgType != "position_dlq" {
		t.Errorf("Expected message_type header to be position_dlq, got %v", msgType)
	}
}

func TestPositionQueueManager_GetQueueNames(t *testing.T) {
	mockHandler := NewMockMessageHandler()
	manager := NewPositionQueueManager(mockHandler)

	names := manager.GetQueueNames()
	expected := DefaultPositionQueueNames()

	if names.PositionUpdates != expected.PositionUpdates {
		t.Errorf("Expected PositionUpdates %s, got %s", expected.PositionUpdates, names.PositionUpdates)
	}
}

func TestPositionQueueManager_GetRetryConfig(t *testing.T) {
	mockHandler := NewMockMessageHandler()
	manager := NewPositionQueueManager(mockHandler)

	config := manager.GetRetryConfig()
	expected := DefaultPositionRetryConfig()

	if config.MaxRetries != expected.MaxRetries {
		t.Errorf("Expected MaxRetries %d, got %d", expected.MaxRetries, config.MaxRetries)
	}
}
