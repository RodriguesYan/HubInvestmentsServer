package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"HubInvestments/internal/position/infra/messaging"
	sharedMessaging "HubInvestments/shared/infra/messaging"
)

type PositionMessageHandler interface {
	HandlePositionUpdateMessage(ctx context.Context, message *PositionUpdateMessage) error
}

type PositionConsumer struct {
	queueManager    *messaging.PositionQueueManager
	messageHandler  sharedMessaging.MessageHandler
	positionHandler PositionMessageHandler
	activeQueues    map[string]bool
	consumersMutex  sync.RWMutex
	shutdownChan    chan struct{}
	shutdownOnce    sync.Once
	isRunning       bool
	runningMutex    sync.RWMutex
}

type PositionConsumerConfig struct {
	ConcurrentWorkers int           // Number of concurrent message processors per queue
	PrefetchCount     int           // Number of messages to prefetch
	RequeueOnError    bool          // Whether to requeue messages on processing errors
	RetryDelay        time.Duration // Delay before retrying failed messages
	MaxRetries        int           // Maximum number of retry attempts
}

func DefaultPositionConsumerConfig() *PositionConsumerConfig {
	return &PositionConsumerConfig{
		ConcurrentWorkers: 10, // Higher concurrency for position updates
		PrefetchCount:     20, // Higher prefetch for better throughput
		RequeueOnError:    true,
		RetryDelay:        2 * time.Second, // Faster retry for position consistency
		MaxRetries:        4,               // Same as position queue config
	}
}

func NewPositionConsumer(
	messageHandler sharedMessaging.MessageHandler,
	queueManager *messaging.PositionQueueManager,
	positionHandler PositionMessageHandler,
) *PositionConsumer {
	return &PositionConsumer{
		queueManager:    queueManager,
		messageHandler:  messageHandler,
		positionHandler: positionHandler,
		activeQueues:    make(map[string]bool),
		shutdownChan:    make(chan struct{}),
	}
}

func (pc *PositionConsumer) StartConsumers(ctx context.Context, config *PositionConsumerConfig) error {
	pc.runningMutex.Lock()
	defer pc.runningMutex.Unlock()

	if pc.isRunning {
		return fmt.Errorf("position consumers are already running")
	}

	if config == nil {
		config = DefaultPositionConsumerConfig()
	}

	queueNames := pc.queueManager.GetQueueNames()

	// Start consumer for position updates queue (main position processing)
	if err := pc.startQueueConsumer(ctx, queueNames.PositionUpdates, config, pc.handlePositionUpdateMessage); err != nil {
		return fmt.Errorf("failed to start position updates queue consumer: %w", err)
	}

	// Start consumer for retry queue (failed position update retries)
	if err := pc.startQueueConsumer(ctx, queueNames.PositionsRetry, config, pc.handlePositionRetryMessage); err != nil {
		return fmt.Errorf("failed to start position retry queue consumer: %w", err)
	}

	pc.isRunning = true
	return nil
}

// StopConsumers gracefully shuts down all consumers
func (pc *PositionConsumer) StopConsumers(ctx context.Context) error {
	pc.runningMutex.Lock()
	defer pc.runningMutex.Unlock()

	if !pc.isRunning {
		return fmt.Errorf("position consumers are not running")
	}

	pc.shutdownOnce.Do(func() {
		close(pc.shutdownChan)
	})

	// Mark all queues as inactive
	pc.consumersMutex.Lock()
	for queueName := range pc.activeQueues {
		pc.activeQueues[queueName] = false
	}
	pc.consumersMutex.Unlock()

	pc.isRunning = false
	return nil
}

func (pc *PositionConsumer) IsRunning() bool {
	pc.runningMutex.RLock()
	defer pc.runningMutex.RUnlock()
	return pc.isRunning
}

func (pc *PositionConsumer) GetActiveQueues() []string {
	pc.consumersMutex.RLock()
	defer pc.consumersMutex.RUnlock()

	var activeQueues []string
	for queueName, isActive := range pc.activeQueues {
		if isActive {
			activeQueues = append(activeQueues, queueName)
		}
	}
	return activeQueues
}

func (pc *PositionConsumer) startQueueConsumer(
	ctx context.Context,
	queueName string,
	config *PositionConsumerConfig,
	messageProcessor func(context.Context, []byte, map[string]interface{}) error,
) error {
	pc.consumersMutex.Lock()
	pc.activeQueues[queueName] = true
	pc.consumersMutex.Unlock()

	// Create a message consumer for this specific queue
	consumer := &PositionMessageConsumer{
		queueName:        queueName,
		messageProcessor: messageProcessor,
		config:           config,
		shutdownChan:     pc.shutdownChan,
	}

	// Start consuming from the queue
	if err := pc.messageHandler.Consume(ctx, queueName, consumer); err != nil {
		pc.consumersMutex.Lock()
		pc.activeQueues[queueName] = false
		pc.consumersMutex.Unlock()
		return fmt.Errorf("failed to start consumer for queue %s: %w", queueName, err)
	}

	return nil
}

type PositionMessageConsumer struct {
	queueName        string
	messageProcessor func(context.Context, []byte, map[string]interface{}) error
	config           *PositionConsumerConfig
	shutdownChan     chan struct{}
}

func (pmc *PositionMessageConsumer) HandleMessage(ctx context.Context, message *sharedMessaging.Message) error {
	// Check if shutdown was requested
	select {
	case <-pmc.shutdownChan:
		return fmt.Errorf("consumer shutting down")
	default:
	}

	// Process message with timeout
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := pmc.messageProcessor(processCtx, message.Body, message.Headers)
	if err != nil {
		// Log error but don't return it to avoid automatic requeuing
		// Our custom retry logic will handle retries
		fmt.Printf("Error processing position message from queue %s: %v\n", pmc.queueName, err)
	}

	return err
}

func (pc *PositionConsumer) handlePositionUpdateMessage(ctx context.Context, messageBody []byte, headers map[string]interface{}) error {
	var message PositionUpdateMessage
	if err := json.Unmarshal(messageBody, &message); err != nil {
		return fmt.Errorf("failed to unmarshal position update message: %w", err)
	}

	// Add correlation info from headers if available
	if correlationID, ok := headers["correlation_id"].(string); ok && correlationID != "" {
		message.MessageMetadata.CorrelationID = correlationID
	}
	if messageID, ok := headers["message_id"].(string); ok && messageID != "" {
		message.MessageMetadata.MessageID = messageID
	}

	return pc.positionHandler.HandlePositionUpdateMessage(ctx, &message)
}

func (pc *PositionConsumer) handlePositionRetryMessage(ctx context.Context, messageBody []byte, headers map[string]interface{}) error {
	var message PositionUpdateMessage
	if err := json.Unmarshal(messageBody, &message); err != nil {
		return fmt.Errorf("failed to unmarshal position retry message: %w", err)
	}

	// Mark as retry message
	message.MessageMetadata.ProcessingStage = "retry"

	// Add retry info from headers if available
	if retryAttempt, ok := headers["retry_attempt"].(int); ok {
		message.MessageMetadata.RetryAttempt = retryAttempt
	}

	return pc.positionHandler.HandlePositionUpdateMessage(ctx, &message)
}
