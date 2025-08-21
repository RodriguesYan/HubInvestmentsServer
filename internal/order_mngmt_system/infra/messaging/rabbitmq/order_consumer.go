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

type OrderMessageHandler interface {
	HandleOrderMessage(ctx context.Context, message *OrderMessage) error
	HandleStatusUpdate(ctx context.Context, statusUpdate *OrderStatusUpdate) error
}

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

type ConsumerConfig struct {
	ConcurrentWorkers int           // Number of concurrent message processors per queue
	PrefetchCount     int           // Number of messages to prefetch
	RequeueOnError    bool          // Whether to requeue messages on processing errors
	RetryDelay        time.Duration // Delay before retrying failed messages
	MaxRetries        int           // Maximum number of retry attempts
}

func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		ConcurrentWorkers: 5,
		PrefetchCount:     10,
		RequeueOnError:    true,
		RetryDelay:        5 * time.Second,
		MaxRetries:        3,
	}
}

func NewOrderConsumer(messageHandler messaging.MessageHandler, orderHandler OrderMessageHandler) *OrderConsumer {
	return &OrderConsumer{
		queueManager:   NewOrderQueueManager(messageHandler),
		messageHandler: messageHandler,
		orderHandler:   orderHandler,
		activeQueues:   make(map[string]bool),
		shutdownChan:   make(chan struct{}),
	}
}

func NewOrderConsumerWithQueueManager(queueManager *OrderQueueManager, orderHandler OrderMessageHandler) *OrderConsumer {
	return &OrderConsumer{
		queueManager:   queueManager,
		messageHandler: queueManager.messageHandler,
		orderHandler:   orderHandler,
		activeQueues:   make(map[string]bool),
		shutdownChan:   make(chan struct{}),
	}
}

// Starts consuming messages from all order-related queues
func (oc *OrderConsumer) StartConsumers(ctx context.Context, config *ConsumerConfig) error {
	oc.runningMutex.Lock()
	defer oc.runningMutex.Unlock()

	if oc.isRunning {
		return fmt.Errorf("consumers are already running")
	}

	if config == nil {
		config = DefaultConsumerConfig()
	}

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

func (oc *OrderConsumer) startQueueConsumer(ctx context.Context, queueName string, config *ConsumerConfig, handler func(context.Context, *messaging.Message) error) error {
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

// Implements the MessageConsumer interface
type orderMessageConsumer struct {
	handler func(context.Context, *messaging.Message) error
	config  *ConsumerConfig
}

func (omc *orderMessageConsumer) HandleMessage(ctx context.Context, message *messaging.Message) error {
	return omc.handler(ctx, message)
}

func (oc *OrderConsumer) handleOrderProcessingMessage(ctx context.Context, message *messaging.Message) error {
	orderMessage, err := oc.deserializeOrderMessage(message.Body)
	if err != nil {
		return fmt.Errorf("failed to deserialize order message: %w", err)
	}

	if err := oc.validateOrderMessage(orderMessage); err != nil {
		return fmt.Errorf("order message validation failed: %w", err)
	}

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

func (oc *OrderConsumer) deserializeOrderMessage(data []byte) (*OrderMessage, error) {
	var orderMessage OrderMessage
	if err := json.Unmarshal(data, &orderMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order message: %w", err)
	}
	return &orderMessage, nil
}

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

	if _, err := domain.ParseOrderSide(message.OrderSide); err != nil {
		return fmt.Errorf("invalid order side: %w", err)
	}

	if _, err := domain.ParseOrderType(message.OrderType); err != nil {
		return fmt.Errorf("invalid order type: %w", err)
	}

	if _, err := domain.ParseOrderStatus(message.Status); err != nil {
		return fmt.Errorf("invalid order status: %w", err)
	}

	if message.MessageMetadata.MessageID == "" {
		return fmt.Errorf("message ID is required")
	}

	if message.MessageMetadata.CorrelationID == "" {
		return fmt.Errorf("correlation ID is required")
	}

	return nil
}

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

	if _, err := domain.ParseOrderStatus(update.CurrentStatus); err != nil {
		return fmt.Errorf("invalid current status: %w", err)
	}

	if update.PreviousStatus == "" {
		return nil
	}

	if _, err := domain.ParseOrderStatus(update.PreviousStatus); err != nil {
		return fmt.Errorf("invalid previous status: %w", err)
	}

	return nil
}

func (oc *OrderConsumer) shouldRetryMessage(message *OrderMessage, err error) bool {
	// Don't retry validation errors or business logic errors
	if isValidationError(err) || isBusinessLogicError(err) {
		return false
	}

	// Retry network errors, temporary failures, etc.
	return isRetryableError(err)
}

func (oc *OrderConsumer) hasExceededMaxRetries(message *OrderMessage) bool {
	maxRetries := 3 // Could be configurable
	return message.MessageMetadata.RetryAttempt >= maxRetries
}

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

func (oc *OrderConsumer) IsRunning() bool {
	oc.runningMutex.RLock()
	defer oc.runningMutex.RUnlock()
	return oc.isRunning
}

func (oc *OrderConsumer) HealthCheck(ctx context.Context) error {
	if !oc.IsRunning() {
		return fmt.Errorf("consumer is not running")
	}

	if err := oc.queueManager.HealthCheck(ctx); err != nil {
		return fmt.Errorf("queue manager health check failed: %w", err)
	}

	oc.consumersMutex.RLock()
	consumerCount := len(oc.activeQueues)
	oc.consumersMutex.RUnlock()

	if consumerCount == 0 {
		return fmt.Errorf("no active consumers")
	}

	return nil
}

func (oc *OrderConsumer) GetQueueManager() *OrderQueueManager {
	return oc.queueManager
}

func (oc *OrderConsumer) GetActiveConsumers() map[string]bool {
	oc.consumersMutex.RLock()
	defer oc.consumersMutex.RUnlock()

	active := make(map[string]bool)
	for queueName := range oc.activeQueues {
		active[queueName] = true
	}
	return active
}

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
