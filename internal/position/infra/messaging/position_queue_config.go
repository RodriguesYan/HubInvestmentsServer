package messaging

import (
	"context"
	"fmt"
	"time"

	"HubInvestments/shared/infra/messaging"
)

type PositionQueueNames struct {
	// Primary processing queue
	PositionUpdates string

	// Management and monitoring queues
	PositionsDLQ   string
	PositionsRetry string

	// Exchange names
	PositionsExchange string
	DLQExchange       string
}

func DefaultPositionQueueNames() PositionQueueNames {
	return PositionQueueNames{
		PositionUpdates: "positions.updates",

		PositionsDLQ:   "positions.updates.dlq",
		PositionsRetry: "positions.retry",

		PositionsExchange: "positions.exchange",
		DLQExchange:       "positions.dlq.exchange",
	}
}

type PositionQueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Arguments  map[string]interface{}
}

type PositionRetryConfig struct {
	// Retry intervals: 2min → 10min → 30min → 2hr for faster position consistency
	RetryIntervals []time.Duration
	MaxRetries     int
}

func DefaultPositionRetryConfig() PositionRetryConfig {
	return PositionRetryConfig{
		RetryIntervals: []time.Duration{
			2 * time.Minute,   // First retry after 2 minutes
			10 * time.Minute,  // Second retry after 10 minutes
			30 * time.Minute,  // Third retry after 30 minutes
			120 * time.Minute, // Fourth retry after 2 hours
		},
		MaxRetries: 4,
	}
}

type PositionQueueManager struct {
	messageHandler messaging.MessageHandler
	queueNames     PositionQueueNames
	retryConfig    PositionRetryConfig
}

func NewPositionQueueManager(messageHandler messaging.MessageHandler) *PositionQueueManager {
	return &PositionQueueManager{
		messageHandler: messageHandler,
		queueNames:     DefaultPositionQueueNames(),
		retryConfig:    DefaultPositionRetryConfig(),
	}
}

func NewPositionQueueManagerWithConfig(
	messageHandler messaging.MessageHandler,
	queueNames PositionQueueNames,
	retryConfig PositionRetryConfig,
) *PositionQueueManager {
	return &PositionQueueManager{
		messageHandler: messageHandler,
		queueNames:     queueNames,
		retryConfig:    retryConfig,
	}
}

func (pqm *PositionQueueManager) SetupAllQueues(ctx context.Context) error {
	if err := pqm.setupPrimaryQueues(); err != nil {
		return fmt.Errorf("failed to setup primary position queues: %w", err)
	}

	if err := pqm.setupManagementQueues(); err != nil {
		return fmt.Errorf("failed to setup position management queues: %w", err)
	}

	return nil
}

func (pqm *PositionQueueManager) setupPrimaryQueues() error {
	// Primary position updates queue
	updatesConfig := PositionQueueConfig{
		Name:       pqm.queueNames.PositionUpdates,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Arguments: map[string]interface{}{
			// Route failed messages to DLQ after max retries
			"x-dead-letter-exchange":    pqm.queueNames.DLQExchange,
			"x-dead-letter-routing-key": pqm.queueNames.PositionsDLQ,
			// Message TTL: 6 hours for position updates (positions need timely processing)
			"x-message-ttl": int64(6 * time.Hour / time.Millisecond),
			// Queue length limit to prevent memory issues during high volume
			"x-max-length": 100000,
		},
	}

	options := messaging.QueueOptions{
		Durable:    updatesConfig.Durable,
		AutoDelete: updatesConfig.AutoDelete,
		Exclusive:  updatesConfig.Exclusive,
		NoWait:     updatesConfig.NoWait,
		Arguments:  updatesConfig.Arguments,
	}

	if err := pqm.messageHandler.DeclareQueue(updatesConfig.Name, options); err != nil {
		return fmt.Errorf("failed to declare position updates queue %s: %w", updatesConfig.Name, err)
	}

	return nil
}

// configures DLQ and retry queues for positions
func (pqm *PositionQueueManager) setupManagementQueues() error {
	// Dead Letter Queue - stores position updates that failed all retries
	dlqConfig := PositionQueueConfig{
		Name:       pqm.queueNames.PositionsDLQ,
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

	if err := pqm.messageHandler.DeclareQueue(dlqConfig.Name, dlqOptions); err != nil {
		return fmt.Errorf("failed to declare position DLQ: %w", err)
	}

	// Retry Queue with TTL-based retry mechanism for position updates
	retryConfig := PositionQueueConfig{
		Name:       pqm.queueNames.PositionsRetry,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Arguments: map[string]interface{}{
			// Messages in retry queue will be routed back to position updates after TTL
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": pqm.queueNames.PositionUpdates,
			// Default retry TTL (will be overridden per message based on retry attempt)
			"x-message-ttl": int64(pqm.retryConfig.RetryIntervals[0] / time.Millisecond),
		},
	}

	retryOptions := messaging.QueueOptions{
		Durable:    retryConfig.Durable,
		AutoDelete: retryConfig.AutoDelete,
		Exclusive:  retryConfig.Exclusive,
		NoWait:     retryConfig.NoWait,
		Arguments:  retryConfig.Arguments,
	}

	if err := pqm.messageHandler.DeclareQueue(retryConfig.Name, retryOptions); err != nil {
		return fmt.Errorf("failed to declare position retry queue: %w", err)
	}

	return nil
}

func (pqm *PositionQueueManager) GetQueueNames() PositionQueueNames {
	return pqm.queueNames
}

func (pqm *PositionQueueManager) GetRetryConfig() PositionRetryConfig {
	return pqm.retryConfig
}

func (pqm *PositionQueueManager) PublishToPositionUpdatesQueue(ctx context.Context, positionMessage []byte, messageID string) error {
	options := messaging.PublishOptions{
		QueueName:     pqm.queueNames.PositionUpdates,
		Message:       positionMessage,
		Persistent:    true,
		Priority:      7, // High priority for position updates (higher than normal orders)
		MessageID:     messageID,
		CorrelationID: messageID,
		Headers: map[string]interface{}{
			"message_type": "position_update",
			"timestamp":    time.Now().Unix(),
		},
	}

	return pqm.messageHandler.PublishWithOptions(ctx, options)
}

func (pqm *PositionQueueManager) PublishToRetryQueue(ctx context.Context, positionMessage []byte, messageID string, retryAttempt int) error {
	// Calculate retry delay based on attempt number
	retryDelay := pqm.retryConfig.RetryIntervals[0] // Default
	if retryAttempt > 0 && retryAttempt <= len(pqm.retryConfig.RetryIntervals) {
		retryDelay = pqm.retryConfig.RetryIntervals[retryAttempt-1]
	}

	options := messaging.PublishOptions{
		QueueName:     pqm.queueNames.PositionsRetry,
		Message:       positionMessage,
		Persistent:    true,
		Priority:      5,
		MessageID:     messageID,
		CorrelationID: messageID,
		Headers: map[string]interface{}{
			"message_type":   "position_retry",
			"retry_attempt":  retryAttempt,
			"original_queue": pqm.queueNames.PositionUpdates,
			"retry_delay_ms": int64(retryDelay / time.Millisecond),
			"timestamp":      time.Now().Unix(),
		},
		// Override TTL for this specific message
		TTL: int64(retryDelay / time.Millisecond),
	}

	return pqm.messageHandler.PublishWithOptions(ctx, options)
}

func (pqm *PositionQueueManager) PublishToDLQ(ctx context.Context, positionMessage []byte, messageID string, failureReason string) error {
	options := messaging.PublishOptions{
		QueueName:     pqm.queueNames.PositionsDLQ,
		Message:       positionMessage,
		Persistent:    true,
		Priority:      1, // Low priority for DLQ messages
		MessageID:     messageID,
		CorrelationID: messageID,
		Headers: map[string]interface{}{
			"message_type":   "position_dlq",
			"failure_reason": failureReason,
			"original_queue": pqm.queueNames.PositionUpdates,
			"timestamp":      time.Now().Unix(),
			"dlq_timestamp":  time.Now().Unix(),
		},
	}

	return pqm.messageHandler.PublishWithOptions(ctx, options)
}
