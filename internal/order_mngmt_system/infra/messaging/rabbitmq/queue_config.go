package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"HubInvestments/shared/infra/messaging"
)

// QueueNames defines all queue names used in the order management system
type QueueNames struct {
	// Primary processing queues
	OrdersSubmit     string
	OrdersProcessing string
	OrdersSettlement string

	// Management and monitoring queues
	OrdersStatus string
	OrdersDLQ    string
	OrdersRetry  string

	// Exchange names
	OrdersExchange string
	DLQExchange    string
}

// DefaultQueueNames returns the default queue naming convention
func DefaultQueueNames() QueueNames {
	return QueueNames{
		// Primary queues
		OrdersSubmit:     "orders.submit",
		OrdersProcessing: "orders.processing",
		OrdersSettlement: "orders.settlement",

		// Management queues
		OrdersStatus: "orders.status",
		OrdersDLQ:    "orders.dlq",
		OrdersRetry:  "orders.retry",

		// Exchanges
		OrdersExchange: "orders.exchange",
		DLQExchange:    "orders.dlq.exchange",
	}
}

// QueueConfig defines configuration for a specific queue
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Arguments  map[string]interface{}
}

// RetryConfig defines retry timing configuration
type RetryConfig struct {
	// Retry intervals in minutes: 5min → 15min → 1hr → 6hr
	RetryIntervals []time.Duration
	MaxRetries     int
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		RetryIntervals: []time.Duration{
			5 * time.Minute,   // First retry after 5 minutes
			15 * time.Minute,  // Second retry after 15 minutes
			60 * time.Minute,  // Third retry after 1 hour
			360 * time.Minute, // Fourth retry after 6 hours
		},
		MaxRetries: 4,
	}
}

// OrderQueueManager manages all order-related queues
type OrderQueueManager struct {
	messageHandler messaging.MessageHandler
	queueNames     QueueNames
	retryConfig    RetryConfig
}

func NewOrderQueueManager(messageHandler messaging.MessageHandler) *OrderQueueManager {
	return &OrderQueueManager{
		messageHandler: messageHandler,
		queueNames:     DefaultQueueNames(),
		retryConfig:    DefaultRetryConfig(),
	}
}

func NewOrderQueueManagerWithConfig(
	messageHandler messaging.MessageHandler,
	queueNames QueueNames,
	retryConfig RetryConfig,
) *OrderQueueManager {
	return &OrderQueueManager{
		messageHandler: messageHandler,
		queueNames:     queueNames,
		retryConfig:    retryConfig,
	}
}

func (qm *OrderQueueManager) SetupAllQueues(ctx context.Context) error {
	if err := qm.setupPrimaryQueues(ctx); err != nil {
		return fmt.Errorf("failed to setup primary queues: %w", err)
	}

	// Setup management queues (DLQ, retry, status)
	if err := qm.setupManagementQueues(ctx); err != nil {
		return fmt.Errorf("failed to setup management queues: %w", err)
	}

	return nil
}

// setupPrimaryQueues configures the main order processing queues
func (qm *OrderQueueManager) setupPrimaryQueues(ctx context.Context) error {
	primaryQueues := []QueueConfig{
		{
			Name:       qm.queueNames.OrdersSubmit,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: map[string]interface{}{
				// Route failed messages to DLQ after max retries
				"x-dead-letter-exchange":    qm.queueNames.DLQExchange,
				"x-dead-letter-routing-key": qm.queueNames.OrdersDLQ,
				// Message TTL: 24 hours for order submission
				"x-message-ttl": int64(24 * time.Hour / time.Millisecond),
			},
		},
		{
			Name:       qm.queueNames.OrdersProcessing,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: map[string]interface{}{
				"x-dead-letter-exchange":    qm.queueNames.DLQExchange,
				"x-dead-letter-routing-key": qm.queueNames.OrdersDLQ,
				// Message TTL: 2 hours for order processing
				"x-message-ttl": int64(2 * time.Hour / time.Millisecond),
			},
		},
		{
			Name:       qm.queueNames.OrdersSettlement,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: map[string]interface{}{
				"x-dead-letter-exchange":    qm.queueNames.DLQExchange,
				"x-dead-letter-routing-key": qm.queueNames.OrdersDLQ,
				// Message TTL: 4 hours for settlement
				"x-message-ttl": int64(4 * time.Hour / time.Millisecond),
			},
		},
		{
			Name:       qm.queueNames.OrdersStatus,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: map[string]interface{}{
				// Status updates have shorter TTL
				"x-message-ttl": int64(1 * time.Hour / time.Millisecond),
			},
		},
	}

	for _, queueConfig := range primaryQueues {
		options := messaging.QueueOptions{
			Durable:    queueConfig.Durable,
			AutoDelete: queueConfig.AutoDelete,
			Exclusive:  queueConfig.Exclusive,
			NoWait:     queueConfig.NoWait,
			Arguments:  queueConfig.Arguments,
		}

		if err := qm.messageHandler.DeclareQueue(queueConfig.Name, options); err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueConfig.Name, err)
		}
	}

	return nil
}

// setupManagementQueues configures DLQ and retry queues
func (qm *OrderQueueManager) setupManagementQueues(ctx context.Context) error {
	// Dead Letter Queue - stores messages that failed all retries
	dlqConfig := QueueConfig{
		Name:       qm.queueNames.OrdersDLQ,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Arguments: map[string]interface{}{
			// DLQ messages persist for 7 days for manual investigation
			"x-message-ttl": int64(7 * 24 * time.Hour / time.Millisecond),
		},
	}

	dlqOptions := messaging.QueueOptions{
		Durable:    dlqConfig.Durable,
		AutoDelete: dlqConfig.AutoDelete,
		Exclusive:  dlqConfig.Exclusive,
		NoWait:     dlqConfig.NoWait,
		Arguments:  dlqConfig.Arguments,
	}

	if err := qm.messageHandler.DeclareQueue(dlqConfig.Name, dlqOptions); err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Retry Queue with TTL-based retry mechanism
	retryConfig := QueueConfig{
		Name:       qm.queueNames.OrdersRetry,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Arguments: map[string]interface{}{
			// Messages in retry queue will be routed back to processing after TTL
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": qm.queueNames.OrdersProcessing,
			// Default retry TTL (will be overridden per message)
			"x-message-ttl": int64(qm.retryConfig.RetryIntervals[0] / time.Millisecond),
		},
	}

	retryOptions := messaging.QueueOptions{
		Durable:    retryConfig.Durable,
		AutoDelete: retryConfig.AutoDelete,
		Exclusive:  retryConfig.Exclusive,
		NoWait:     retryConfig.NoWait,
		Arguments:  retryConfig.Arguments,
	}

	if err := qm.messageHandler.DeclareQueue(retryConfig.Name, retryOptions); err != nil {
		return fmt.Errorf("failed to declare retry queue: %w", err)
	}

	return nil
}

// GetQueueNames returns the configured queue names
func (qm *OrderQueueManager) GetQueueNames() QueueNames {
	return qm.queueNames
}

// GetRetryConfig returns the retry configuration
func (qm *OrderQueueManager) GetRetryConfig() RetryConfig {
	return qm.retryConfig
}

func (qm *OrderQueueManager) PublishToSubmitQueue(ctx context.Context, orderMessage []byte, messageID string) error {
	options := messaging.PublishOptions{
		QueueName:     qm.queueNames.OrdersSubmit,
		Message:       orderMessage,
		Persistent:    true,
		Priority:      5, // Normal priority
		MessageID:     messageID,
		CorrelationID: messageID,
		Headers: map[string]interface{}{
			"message_type": "order_submission",
			"timestamp":    time.Now().Unix(),
		},
	}

	return qm.messageHandler.PublishWithOptions(ctx, options)
}

func (qm *OrderQueueManager) PublishToProcessingQueue(ctx context.Context, orderMessage []byte, messageID string, priority uint8) error {
	options := messaging.PublishOptions{
		QueueName:     qm.queueNames.OrdersProcessing,
		Message:       orderMessage,
		Persistent:    true,
		Priority:      priority,
		MessageID:     messageID,
		CorrelationID: messageID,
		Headers: map[string]interface{}{
			"message_type": "order_processing",
			"timestamp":    time.Now().Unix(),
		},
	}

	return qm.messageHandler.PublishWithOptions(ctx, options)
}

func (qm *OrderQueueManager) PublishToRetryQueue(ctx context.Context, orderMessage []byte, messageID string, retryAttempt int) error {
	// Calculate TTL based on retry attempt
	var ttl time.Duration
	if retryAttempt < len(qm.retryConfig.RetryIntervals) {
		ttl = qm.retryConfig.RetryIntervals[retryAttempt]
	} else {
		// Use last interval for any attempts beyond configured intervals
		ttl = qm.retryConfig.RetryIntervals[len(qm.retryConfig.RetryIntervals)-1]
	}

	options := messaging.PublishOptions{
		QueueName:     qm.queueNames.OrdersRetry,
		Message:       orderMessage,
		Persistent:    true,
		TTL:           int64(ttl / time.Millisecond),
		MessageID:     messageID,
		CorrelationID: messageID,
		Headers: map[string]interface{}{
			"message_type":   "order_retry",
			"retry_attempt":  retryAttempt,
			"retry_delay_ms": int64(ttl / time.Millisecond),
			"timestamp":      time.Now().Unix(),
		},
	}

	return qm.messageHandler.PublishWithOptions(ctx, options)
}

func (qm *OrderQueueManager) PublishStatusUpdate(ctx context.Context, statusMessage []byte, orderID string) error {
	options := messaging.PublishOptions{
		QueueName:     qm.queueNames.OrdersStatus,
		Message:       statusMessage,
		Persistent:    true,
		Priority:      8, // High priority for status updates
		MessageID:     fmt.Sprintf("status_%s_%d", orderID, time.Now().UnixNano()),
		CorrelationID: orderID,
		Headers: map[string]interface{}{
			"message_type": "status_update",
			"order_id":     orderID,
			"timestamp":    time.Now().Unix(),
		},
	}

	return qm.messageHandler.PublishWithOptions(ctx, options)
}

// GetQueueInfo returns information about all order management queues
func (qm *OrderQueueManager) GetQueueInfo(ctx context.Context) (map[string]*messaging.QueueInfo, error) {
	queueNames := []string{
		qm.queueNames.OrdersSubmit,
		qm.queueNames.OrdersProcessing,
		qm.queueNames.OrdersSettlement,
		qm.queueNames.OrdersStatus,
		qm.queueNames.OrdersDLQ,
		qm.queueNames.OrdersRetry,
	}

	queueInfoMap := make(map[string]*messaging.QueueInfo)

	for _, queueName := range queueNames {
		info, err := qm.messageHandler.QueueInfo(queueName)
		if err != nil {
			return nil, fmt.Errorf("failed to get info for queue %s: %w", queueName, err)
		}
		queueInfoMap[queueName] = info
	}

	return queueInfoMap, nil
}

// PurgeAllQueues removes all messages from all order management queues
// WARNING: This should only be used in development/testing environments
func (qm *OrderQueueManager) PurgeAllQueues(ctx context.Context) error {
	queueNames := []string{
		qm.queueNames.OrdersSubmit,
		qm.queueNames.OrdersProcessing,
		qm.queueNames.OrdersSettlement,
		qm.queueNames.OrdersStatus,
		qm.queueNames.OrdersDLQ,
		qm.queueNames.OrdersRetry,
	}

	for _, queueName := range queueNames {
		if err := qm.messageHandler.PurgeQueue(queueName); err != nil {
			return fmt.Errorf("failed to purge queue %s: %w", queueName, err)
		}
	}

	return nil
}

// HealthCheck verifies all queues are accessible and healthy
func (qm *OrderQueueManager) HealthCheck(ctx context.Context) error {
	// First check the underlying message handler
	if err := qm.messageHandler.HealthCheck(ctx); err != nil {
		return fmt.Errorf("message handler health check failed: %w", err)
	}

	// Check that all queues are accessible
	queueNames := []string{
		qm.queueNames.OrdersSubmit,
		qm.queueNames.OrdersProcessing,
		qm.queueNames.OrdersSettlement,
		qm.queueNames.OrdersStatus,
		qm.queueNames.OrdersDLQ,
		qm.queueNames.OrdersRetry,
	}

	for _, queueName := range queueNames {
		_, err := qm.messageHandler.QueueInfo(queueName)
		if err != nil {
			return fmt.Errorf("queue %s is not accessible: %w", queueName, err)
		}
	}

	return nil
}
