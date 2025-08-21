package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/shared/infra/messaging"
)

// OrderMessageHandler defines the interface for handling consumed order messages
type OrderMessageHandler interface {
	HandleOrderMessage(ctx context.Context, message *OrderMessage) error
	HandleStatusUpdate(ctx context.Context, statusUpdate *OrderStatusUpdate) error
}

// OrderConsumer handles consuming order messages from RabbitMQ queues
type OrderConsumer struct {
	queueManager   *OrderQueueManager
	messageHandler messaging.MessageHandler
	orderHandler   OrderMessageHandler
	activeQueues   map[string]bool
	consumersMutex sync.RWMutex
	shutdownChan   chan struct{}
	shutdownOnce   sync.Once
	isRunning      bool
	runningMutex   sync.RWMutex
}

// ConsumerConfig holds configuration for the order consumer
type ConsumerConfig struct {
	ConcurrentWorkers int           // Number of concurrent message processors per queue
	PrefetchCount     int           // Number of messages to prefetch
	RequeueOnError    bool          // Whether to requeue messages on processing errors
	RetryDelay        time.Duration // Delay before retrying failed messages
	MaxRetries        int           // Maximum number of retry attempts
}

// DefaultConsumerConfig returns sensible defaults for consumer configuration
func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		ConcurrentWorkers: 5,
		PrefetchCount:     10,
		RequeueOnError:    true,
		RetryDelay:        5 * time.Second,
		MaxRetries:        3,
	}
}

// NewOrderConsumer creates a new order consumer
func NewOrderConsumer(messageHandler messaging.MessageHandler, orderHandler OrderMessageHandler) *OrderConsumer {
	return &OrderConsumer{
		queueManager:   NewOrderQueueManager(messageHandler),
		messageHandler: messageHandler,
		orderHandler:   orderHandler,
		activeQueues:   make(map[string]bool),
		shutdownChan:   make(chan struct{}),
	}
}

// NewOrderConsumerWithQueueManager creates a consumer with existing queue manager
func NewOrderConsumerWithQueueManager(queueManager *OrderQueueManager, orderHandler OrderMessageHandler) *OrderConsumer {
	return &OrderConsumer{
		queueManager:   queueManager,
		messageHandler: queueManager.messageHandler,
		orderHandler:   orderHandler,
		activeQueues:   make(map[string]bool),
		shutdownChan:   make(chan struct{}),
	}
}

// StartConsumers starts consuming messages from all order-related queues
func (oc *OrderConsumer) StartConsumers(ctx context.Context, config *ConsumerConfig) error {
	oc.runningMutex.Lock()
	defer oc.runningMutex.Unlock()

	if oc.isRunning {
		return fmt.Errorf("consumers are already running")
	}

	if config == nil {
		config = DefaultConsumerConfig()
	}

	// Ensure all queues are set up before starting consumers
	if err := oc.queueManager.SetupAllQueues(ctx); err != nil {
		return fmt.Errorf("failed to setup queues: %w", err)
	}

	queueNames := oc.queueManager.GetQueueNames()

	// Start consumer for processing queue (main order processing)
	if err := oc.startQueueConsumer(ctx, queueNames.OrdersProcessing, config, oc.handleOrderProcessingMessage); err != nil {
		return fmt.Errorf("failed to start processing queue consumer: %w", err)
	}

	// Start consumer for submission queue (order validation and preparation)
	if err := oc.startQueueConsumer(ctx, queueNames.OrdersSubmit, config, oc.handleOrderSubmissionMessage); err != nil {
		return fmt.Errorf("failed to start submission queue consumer: %w", err)
	}

	// Start consumer for retry queue (failed order retries)
	if err := oc.startQueueConsumer(ctx, queueNames.OrdersRetry, config, oc.handleOrderRetryMessage); err != nil {
		return fmt.Errorf("failed to start retry queue consumer: %w", err)
	}

	// Start consumer for status updates queue
	if err := oc.startQueueConsumer(ctx, queueNames.OrdersStatus, config, oc.handleStatusUpdateMessage); err != nil {
		return fmt.Errorf("failed to start status queue consumer: %w", err)
	}

	oc.isRunning = true
	return nil
}

// startQueueConsumer starts a consumer for a specific queue
func (oc *OrderConsumer) startQueueConsumer(ctx context.Context, queueName string, config *ConsumerConfig, handler func(context.Context, *messaging.Message) error) error {
	// Create a consumer that wraps our handler
	consumer := &orderMessageConsumer{
		handler: handler,
		config:  config,
	}

	err := oc.messageHandler.Consume(ctx, queueName, consumer)
	if err != nil {
		return fmt.Errorf("failed to start consumer for queue %s: %w", queueName, err)
	}

	oc.consumersMutex.Lock()
	oc.activeQueues[queueName] = true
	oc.consumersMutex.Unlock()

	return nil
}

// orderMessageConsumer implements the MessageConsumer interface
type orderMessageConsumer struct {
	handler func(context.Context, *messaging.Message) error
	config  *ConsumerConfig
}

func (omc *orderMessageConsumer) HandleMessage(ctx context.Context, message *messaging.Message) error {
	return omc.handler(ctx, message)
}

// handleOrderProcessingMessage processes messages from the processing queue
func (oc *OrderConsumer) handleOrderProcessingMessage(ctx context.Context, message *messaging.Message) error {
	orderMessage, err := oc.deserializeOrderMessage(message.Body)
	if err != nil {
		return fmt.Errorf("failed to deserialize order message: %w", err)
	}

	if err := oc.validateOrderMessage(orderMessage); err != nil {
		return fmt.Errorf("order message validation failed: %w", err)
	}

	// Process the order message through the handler
	if err := oc.orderHandler.HandleOrderMessage(ctx, orderMessage); err != nil {
		// Check if this is a retryable error
		if oc.shouldRetryMessage(orderMessage, err) {
			return fmt.Errorf("order processing failed (retryable): %w", err)
		}

		// Send to DLQ for non-retryable errors
		if dlqErr := oc.sendToDLQ(ctx, message, err); dlqErr != nil {
			return fmt.Errorf("order processing failed and DLQ send failed: %w (original: %v)", dlqErr, err)
		}

		// Acknowledge the message since we've handled it (sent to DLQ)
		return nil
	}

	return nil
}

// handleOrderSubmissionMessage processes messages from the submission queue
func (oc *OrderConsumer) handleOrderSubmissionMessage(ctx context.Context, message *messaging.Message) error {
	orderMessage, err := oc.deserializeOrderMessage(message.Body)
	if err != nil {
		return fmt.Errorf("failed to deserialize submission message: %w", err)
	}

	if err := oc.validateOrderMessage(orderMessage); err != nil {
		return fmt.Errorf("submission message validation failed: %w", err)
	}

	// For submission messages, we typically validate and then forward to processing
	if err := oc.orderHandler.HandleOrderMessage(ctx, orderMessage); err != nil {
		return fmt.Errorf("order submission handling failed: %w", err)
	}

	return nil
}

// handleOrderRetryMessage processes messages from the retry queue
func (oc *OrderConsumer) handleOrderRetryMessage(ctx context.Context, message *messaging.Message) error {
	orderMessage, err := oc.deserializeOrderMessage(message.Body)
	if err != nil {
		return fmt.Errorf("failed to deserialize retry message: %w", err)
	}

	if err := oc.validateOrderMessage(orderMessage); err != nil {
		return fmt.Errorf("retry message validation failed: %w", err)
	}

	// Check if we've exceeded maximum retry attempts
	if oc.hasExceededMaxRetries(orderMessage) {
		if dlqErr := oc.sendToDLQ(ctx, message, fmt.Errorf("exceeded maximum retry attempts")); dlqErr != nil {
			return fmt.Errorf("max retries exceeded and DLQ send failed: %w", dlqErr)
		}
		return nil
	}

	// Process the retry message
	if err := oc.orderHandler.HandleOrderMessage(ctx, orderMessage); err != nil {
		return fmt.Errorf("order retry processing failed: %w", err)
	}

	return nil
}

// handleStatusUpdateMessage processes status update messages
func (oc *OrderConsumer) handleStatusUpdateMessage(ctx context.Context, message *messaging.Message) error {
	statusUpdate, err := oc.deserializeStatusUpdate(message.Body)
	if err != nil {
		return fmt.Errorf("failed to deserialize status update: %w", err)
	}

	if err := oc.validateStatusUpdate(statusUpdate); err != nil {
		return fmt.Errorf("status update validation failed: %w", err)
	}

	if err := oc.orderHandler.HandleStatusUpdate(ctx, statusUpdate); err != nil {
		return fmt.Errorf("status update handling failed: %w", err)
	}

	return nil
}

// deserializeOrderMessage converts JSON bytes to OrderMessage
func (oc *OrderConsumer) deserializeOrderMessage(data []byte) (*OrderMessage, error) {
	var orderMessage OrderMessage
	if err := json.Unmarshal(data, &orderMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order message: %w", err)
	}
	return &orderMessage, nil
}

// deserializeStatusUpdate converts JSON bytes to OrderStatusUpdate
func (oc *OrderConsumer) deserializeStatusUpdate(data []byte) (*OrderStatusUpdate, error) {
	var statusUpdate OrderStatusUpdate
	if err := json.Unmarshal(data, &statusUpdate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status update: %w", err)
	}
	return &statusUpdate, nil
}

// validateOrderMessage performs validation on the deserialized order message
func (oc *OrderConsumer) validateOrderMessage(message *OrderMessage) error {
	if message == nil {
		return fmt.Errorf("order message cannot be nil")
	}

	if message.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}

	if message.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if message.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if message.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Validate order side
	if _, err := domain.ParseOrderSide(message.OrderSide); err != nil {
		return fmt.Errorf("invalid order side: %w", err)
	}

	// Validate order type
	if _, err := domain.ParseOrderType(message.OrderType); err != nil {
		return fmt.Errorf("invalid order type: %w", err)
	}

	// Validate order status
	if _, err := domain.ParseOrderStatus(message.Status); err != nil {
		return fmt.Errorf("invalid order status: %w", err)
	}

	// Validate message metadata
	if message.MessageMetadata.MessageID == "" {
		return fmt.Errorf("message ID is required")
	}

	if message.MessageMetadata.CorrelationID == "" {
		return fmt.Errorf("correlation ID is required")
	}

	return nil
}

// validateStatusUpdate performs validation on the deserialized status update
func (oc *OrderConsumer) validateStatusUpdate(update *OrderStatusUpdate) error {
	if update == nil {
		return fmt.Errorf("status update cannot be nil")
	}

	if update.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}

	if update.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if update.CurrentStatus == "" {
		return fmt.Errorf("current status is required")
	}

	// Validate status values
	if _, err := domain.ParseOrderStatus(update.CurrentStatus); err != nil {
		return fmt.Errorf("invalid current status: %w", err)
	}

	if update.PreviousStatus != "" {
		if _, err := domain.ParseOrderStatus(update.PreviousStatus); err != nil {
			return fmt.Errorf("invalid previous status: %w", err)
		}
	}

	return nil
}

// shouldRetryMessage determines if a message should be retried based on the error
func (oc *OrderConsumer) shouldRetryMessage(message *OrderMessage, err error) bool {
	// Don't retry validation errors or business logic errors
	if isValidationError(err) || isBusinessLogicError(err) {
		return false
	}

	// Retry network errors, temporary failures, etc.
	return isRetryableError(err)
}

// hasExceededMaxRetries checks if the message has exceeded maximum retry attempts
func (oc *OrderConsumer) hasExceededMaxRetries(message *OrderMessage) bool {
	maxRetries := 3 // Could be configurable
	return message.MessageMetadata.RetryAttempt >= maxRetries
}

// sendToDLQ sends a message to the Dead Letter Queue
func (oc *OrderConsumer) sendToDLQ(ctx context.Context, message *messaging.Message, processingError error) error {
	dlqMessage := map[string]interface{}{
		"original_message":    string(message.Body),
		"processing_error":    processingError.Error(),
		"failed_at":           time.Now(),
		"original_message_id": message.MessageID,
		"retry_attempts":      message.Headers["retry_attempts"],
	}

	dlqBytes, err := json.Marshal(dlqMessage)
	if err != nil {
		return fmt.Errorf("failed to serialize DLQ message: %w", err)
	}

	queueNames := oc.queueManager.GetQueueNames()
	return oc.messageHandler.PublishWithOptions(ctx, messaging.PublishOptions{
		QueueName:  queueNames.OrdersDLQ,
		Message:    dlqBytes,
		Persistent: true,
		Headers: map[string]interface{}{
			"original_message_id": message.MessageID,
			"error_type":          "processing_failed",
		},
	})
}

// StopConsumers gracefully stops all consumers
func (oc *OrderConsumer) StopConsumers(ctx context.Context) error {
	oc.runningMutex.Lock()
	defer oc.runningMutex.Unlock()

	if !oc.isRunning {
		return nil
	}

	oc.shutdownOnce.Do(func() {
		close(oc.shutdownChan)

		oc.consumersMutex.Lock()
		oc.activeQueues = make(map[string]bool)
		oc.consumersMutex.Unlock()
	})

	oc.isRunning = false
	return nil
}

// IsRunning returns whether the consumer is currently running
func (oc *OrderConsumer) IsRunning() bool {
	oc.runningMutex.RLock()
	defer oc.runningMutex.RUnlock()
	return oc.isRunning
}

// HealthCheck verifies the consumer and underlying infrastructure is healthy
func (oc *OrderConsumer) HealthCheck(ctx context.Context) error {
	if !oc.IsRunning() {
		return fmt.Errorf("consumer is not running")
	}

	// Check queue manager health
	if err := oc.queueManager.HealthCheck(ctx); err != nil {
		return fmt.Errorf("queue manager health check failed: %w", err)
	}

	// Check individual consumers
	oc.consumersMutex.RLock()
	consumerCount := len(oc.activeQueues)
	oc.consumersMutex.RUnlock()

	if consumerCount == 0 {
		return fmt.Errorf("no active consumers")
	}

	return nil
}

// GetQueueManager returns the underlying queue manager
func (oc *OrderConsumer) GetQueueManager() *OrderQueueManager {
	return oc.queueManager
}

// GetActiveConsumers returns information about active consumers
func (oc *OrderConsumer) GetActiveConsumers() map[string]bool {
	oc.consumersMutex.RLock()
	defer oc.consumersMutex.RUnlock()

	active := make(map[string]bool)
	for queueName := range oc.activeQueues {
		active[queueName] = true
	}
	return active
}

// Helper functions for error classification

func isValidationError(err error) bool {
	// Check if error is related to validation
	errStr := err.Error()
	return contains(errStr, "validation") ||
		contains(errStr, "invalid") ||
		contains(errStr, "required") ||
		contains(errStr, "malformed")
}

func isBusinessLogicError(err error) bool {
	// Check if error is related to business logic
	errStr := err.Error()
	return contains(errStr, "insufficient balance") ||
		contains(errStr, "market closed") ||
		contains(errStr, "position limit") ||
		contains(errStr, "duplicate order")
}

func isRetryableError(err error) bool {
	// Check if error is retryable (network, temporary failures, etc.)
	errStr := err.Error()
	return contains(errStr, "connection") ||
		contains(errStr, "timeout") ||
		contains(errStr, "temporary") ||
		contains(errStr, "unavailable") ||
		contains(errStr, "service down")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
