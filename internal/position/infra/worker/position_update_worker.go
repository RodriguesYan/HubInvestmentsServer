package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"HubInvestments/internal/position/application/command"
	positionUsecase "HubInvestments/internal/position/application/usecase"
	domain "HubInvestments/internal/position/domain/model"
	positionRepository "HubInvestments/internal/position/domain/repository"
	"HubInvestments/internal/position/infra/messaging"
	sharedMessaging "HubInvestments/shared/infra/messaging"

	"github.com/google/uuid"
)

type PositionUpdateWorker struct {
	id                 string
	createPositionUC   positionUsecase.ICreatePositionUseCase
	updatePositionUC   positionUsecase.IUpdatePositionUseCase
	closePositionUC    positionUsecase.IClosePositionUseCase
	positionRepository positionRepository.IPositionRepository
	positionConsumer   *PositionConsumer
	messageHandler     sharedMessaging.MessageHandler
	queueManager       *messaging.PositionQueueManager
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
	isRunning          bool
	mu                 sync.RWMutex
	config             *PositionWorkerConfig
	metrics            *PositionWorkerMetrics
	healthStatus       HealthStatus
	lastHeartbeat      time.Time
	processedCount     int64
	errorCount         int64
	retryCount         int64
}

type PositionWorkerConfig struct {
	WorkerID                   string
	MaxConcurrentUpdates       int
	ProcessingTimeout          time.Duration
	HeartbeatInterval          time.Duration
	MaxRetries                 int
	RetryBackoffBase           time.Duration
	HealthCheckInterval        time.Duration
	ShutdownTimeout            time.Duration
	EnableMetrics              bool
	LogLevel                   string
	PositionConsistencyTimeout time.Duration // Time to wait for position consistency
}

type PositionWorkerMetrics struct {
	PositionsProcessed    int64
	PositionsCreated      int64
	PositionsUpdated      int64
	PositionsClosed       int64
	PositionsFailed       int64
	PositionsRetried      int64
	AverageProcessingTime time.Duration
	LastProcessingTime    time.Duration
	StartTime             time.Time
	LastActivityTime      time.Time
	mu                    sync.RWMutex
}

type PositionWorkerMetricsSnapshot struct {
	PositionsProcessed    int64
	PositionsCreated      int64
	PositionsUpdated      int64
	PositionsClosed       int64
	PositionsFailed       int64
	PositionsRetried      int64
	AverageProcessingTime time.Duration
	LastProcessingTime    time.Duration
	StartTime             time.Time
	LastActivityTime      time.Time
}

type HealthStatus int

const (
	HealthStatusUnknown HealthStatus = iota
	HealthStatusHealthy
	HealthStatusDegraded
	HealthStatusUnhealthy
	HealthStatusStopped
)

func (h HealthStatus) String() string {
	switch h {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

type PositionUpdateMessage struct {
	OrderID             string                        `json:"order_id"`
	UserID              string                        `json:"user_id"`
	Symbol              string                        `json:"symbol"`
	OrderSide           string                        `json:"order_side"` // BUY/SELL
	OrderType           string                        `json:"order_type"`
	Quantity            float64                       `json:"quantity"`
	ExecutionPrice      float64                       `json:"execution_price"`
	TotalValue          float64                       `json:"total_value"`
	ExecutedAt          time.Time                     `json:"executed_at"`
	MarketPriceAtExec   *float64                      `json:"market_price_at_exec,omitempty"`
	MarketDataTimestamp *time.Time                    `json:"market_data_timestamp,omitempty"`
	MessageMetadata     PositionUpdateMessageMetadata `json:"message_metadata"`
}

type PositionUpdateMessageMetadata struct {
	MessageID       string    `json:"message_id"`
	CorrelationID   string    `json:"correlation_id"`
	Timestamp       time.Time `json:"timestamp"`
	RetryAttempt    int       `json:"retry_attempt"`
	Priority        uint8     `json:"priority"`
	Source          string    `json:"source"`
	MessageType     string    `json:"message_type"`
	ProcessingStage string    `json:"processing_stage"`
}

func NewPositionUpdateWorker(
	workerID string,
	createPositionUC positionUsecase.ICreatePositionUseCase,
	updatePositionUC positionUsecase.IUpdatePositionUseCase,
	closePositionUC positionUsecase.IClosePositionUseCase,
	positionRepo positionRepository.IPositionRepository,
	messageHandler sharedMessaging.MessageHandler,
	config *PositionWorkerConfig,
) *PositionUpdateWorker {
	if config == nil {
		config = DefaultPositionWorkerConfig(workerID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	queueManager := messaging.NewPositionQueueManager(messageHandler)

	worker := &PositionUpdateWorker{
		id:                 workerID,
		createPositionUC:   createPositionUC,
		updatePositionUC:   updatePositionUC,
		closePositionUC:    closePositionUC,
		positionRepository: positionRepo,
		messageHandler:     messageHandler,
		queueManager:       queueManager,
		ctx:                ctx,
		cancel:             cancel,
		config:             config,
		metrics:            NewPositionWorkerMetrics(),
		healthStatus:       HealthStatusUnknown,
		lastHeartbeat:      time.Now(),
	}

	// Create position message handler with concurrency control
	positionMessageHandler := &PositionMessageHandlerImpl{
		worker:    worker,
		semaphore: make(chan struct{}, config.MaxConcurrentUpdates),
	}
	worker.positionConsumer = NewPositionConsumer(messageHandler, queueManager, positionMessageHandler)

	return worker
}

func DefaultPositionWorkerConfig(workerID string) *PositionWorkerConfig {
	return &PositionWorkerConfig{
		WorkerID:                   workerID,
		MaxConcurrentUpdates:       20,               // Higher than orders since positions are lighter operations
		ProcessingTimeout:          15 * time.Second, // Shorter than order processing
		HeartbeatInterval:          10 * time.Second,
		MaxRetries:                 4,               // Same as position queue config
		RetryBackoffBase:           2 * time.Second, // Faster backoff for position consistency
		HealthCheckInterval:        30 * time.Second,
		ShutdownTimeout:            60 * time.Second,
		EnableMetrics:              true,
		LogLevel:                   "INFO",
		PositionConsistencyTimeout: 5 * time.Second,
	}
}

func NewPositionWorkerMetrics() *PositionWorkerMetrics {
	return &PositionWorkerMetrics{
		StartTime:        time.Now(),
		LastActivityTime: time.Now(),
	}
}

// begins the worker processing loop
func (w *PositionUpdateWorker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		return fmt.Errorf("position worker %s is already running", w.id)
	}

	log.Printf("Starting position update worker %s with config: max_concurrent=%d, timeout=%v",
		w.id, w.config.MaxConcurrentUpdates, w.config.ProcessingTimeout)

	if err := w.queueManager.SetupAllQueues(w.ctx); err != nil {
		return fmt.Errorf("failed to setup position queues: %w", err)
	}

	w.isRunning = true
	w.healthStatus = HealthStatusHealthy
	w.lastHeartbeat = time.Now()

	// Start heartbeat goroutine
	w.wg.Add(1)
	go w.heartbeatLoop()

	// Start health check goroutine
	w.wg.Add(1)
	go w.healthCheckLoop()

	// Start message processing goroutine
	w.wg.Add(1)
	go w.processMessages()

	log.Printf("Position update worker %s started successfully", w.id)
	return nil
}

func (w *PositionUpdateWorker) Stop() error {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return fmt.Errorf("position worker %s is not running", w.id)
	}
	w.isRunning = false
	w.healthStatus = HealthStatusStopped
	w.mu.Unlock()

	log.Printf("Stopping position update worker %s...", w.id)

	// Cancel context to signal all goroutines to stop
	w.cancel()

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("Position update worker %s stopped gracefully", w.id)
		return nil
	case <-time.After(w.config.ShutdownTimeout):
		log.Printf("Position update worker %s shutdown timeout exceeded", w.id)
		return fmt.Errorf("shutdown timeout exceeded for position worker %s", w.id)
	}
}

func (w *PositionUpdateWorker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.isRunning
}

func (w *PositionUpdateWorker) GetHealthStatus() HealthStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.healthStatus
}

func (w *PositionUpdateWorker) GetMetrics() PositionWorkerMetricsSnapshot {
	w.metrics.mu.RLock()
	defer w.metrics.mu.RUnlock()

	return PositionWorkerMetricsSnapshot{
		PositionsProcessed:    w.metrics.PositionsProcessed,
		PositionsCreated:      w.metrics.PositionsCreated,
		PositionsUpdated:      w.metrics.PositionsUpdated,
		PositionsClosed:       w.metrics.PositionsClosed,
		PositionsFailed:       w.metrics.PositionsFailed,
		PositionsRetried:      w.metrics.PositionsRetried,
		AverageProcessingTime: w.metrics.AverageProcessingTime,
		LastProcessingTime:    w.metrics.LastProcessingTime,
		StartTime:             w.metrics.StartTime,
		LastActivityTime:      w.metrics.LastActivityTime,
	}
}

func (w *PositionUpdateWorker) GetID() string {
	return w.id
}

func (w *PositionUpdateWorker) updateHealthStatus(status HealthStatus) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.healthStatus = status
}

func (w *PositionUpdateWorker) updateLastActivity() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.metrics.LastActivityTime = time.Now()
}

func (w *PositionUpdateWorker) incrementProcessedCount() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.processedCount++
	w.metrics.PositionsProcessed++
}

func (w *PositionUpdateWorker) incrementErrorCount() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.errorCount++
	w.metrics.PositionsFailed++
}

func (w *PositionUpdateWorker) incrementCreatedCount() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.metrics.PositionsCreated++
}

func (w *PositionUpdateWorker) incrementUpdatedCount() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.metrics.PositionsUpdated++
}

func (w *PositionUpdateWorker) incrementClosedCount() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.metrics.PositionsClosed++
}

func (w *PositionUpdateWorker) incrementRetryCount() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.retryCount++
	w.metrics.PositionsRetried++
}

func (w *PositionUpdateWorker) updateProcessingTime(duration time.Duration) {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()

	w.metrics.LastProcessingTime = duration

	// Calculate rolling average (simple moving average with weight to recent values)
	if w.metrics.AverageProcessingTime == 0 {
		w.metrics.AverageProcessingTime = duration
	} else {
		// 90% weight to previous average, 10% to new value for stability
		w.metrics.AverageProcessingTime = time.Duration(
			0.9*float64(w.metrics.AverageProcessingTime) + 0.1*float64(duration),
		)
	}
}

func (w *PositionUpdateWorker) heartbeatLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("Position worker %s: Heartbeat loop stopped", w.id)
			return
		case <-ticker.C:
			w.mu.Lock()
			w.lastHeartbeat = time.Now()
			w.mu.Unlock()
		}
	}
}

func (w *PositionUpdateWorker) healthCheckLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("Position worker %s: Health check loop stopped", w.id)
			return
		case <-ticker.C:
			w.performHealthCheck()
		}
	}
}

func (w *PositionUpdateWorker) performHealthCheck() {
	w.mu.RLock()
	lastHeartbeat := w.lastHeartbeat
	errorCount := w.errorCount
	processedCount := w.processedCount
	w.mu.RUnlock()

	// Check if worker is responsive
	timeSinceLastHeartbeat := time.Since(lastHeartbeat)
	if timeSinceLastHeartbeat > w.config.HeartbeatInterval*3 {
		w.updateHealthStatus(HealthStatusUnhealthy)
		log.Printf("Position worker %s: Unhealthy - no heartbeat for %v", w.id, timeSinceLastHeartbeat)
		return
	}

	// Check error rate (if more than 50% of recent operations failed)
	if processedCount > 10 && errorCount > processedCount/2 {
		w.updateHealthStatus(HealthStatusDegraded)
		log.Printf("Position worker %s: Degraded - high error rate: %d errors out of %d processed",
			w.id, errorCount, processedCount)
		return
	}

	// Worker is healthy
	w.updateHealthStatus(HealthStatusHealthy)
}

// processMessages is the main message processing loop
func (w *PositionUpdateWorker) processMessages() {
	defer w.wg.Done()

	log.Printf("Position worker %s: Starting message processing loop", w.id)

	// Start consuming messages with context and config
	config := &PositionConsumerConfig{
		ConcurrentWorkers: w.config.MaxConcurrentUpdates,
		PrefetchCount:     w.config.MaxConcurrentUpdates * 2,
		RequeueOnError:    true,
		RetryDelay:        w.config.RetryBackoffBase,
		MaxRetries:        w.config.MaxRetries,
	}

	err := w.positionConsumer.StartConsumers(w.ctx, config)
	if err != nil {
		log.Printf("Position worker %s: Failed to start consumers: %v", w.id, err)
		w.updateHealthStatus(HealthStatusUnhealthy)
		return
	}

	// Wait for context cancellation
	<-w.ctx.Done()

	log.Printf("Position worker %s: Stopping message processing", w.id)

	// Stop consumers
	err = w.positionConsumer.StopConsumers(w.ctx)
	if err != nil {
		log.Printf("Position worker %s: Error stopping consumers: %v", w.id, err)
	}
}

type PositionMessageHandlerImpl struct {
	worker    *PositionUpdateWorker
	semaphore chan struct{}
}

func (h *PositionMessageHandlerImpl) HandlePositionUpdateMessage(ctx context.Context, message *PositionUpdateMessage) error {
	// Semaphore pattern: limit concurrent position processing
	select {
	case h.semaphore <- struct{}{}:
		defer func() { <-h.semaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}

	return h.worker.processPositionUpdateMessage(ctx, message)
}

func (w *PositionUpdateWorker) processPositionUpdateMessage(ctx context.Context, message *PositionUpdateMessage) error {
	startTime := time.Now()

	processCtx, cancel := context.WithTimeout(ctx, w.config.ProcessingTimeout)
	defer cancel()

	log.Printf("Position worker %s: Processing position update for order %s (user: %s, symbol: %s, side: %s, quantity: %.2f)",
		w.id, message.OrderID, message.UserID, message.Symbol, message.OrderSide, message.Quantity)

	w.updateLastActivity()
	w.incrementProcessedCount()

	var err error
	var operationType string

	// Determine the operation type based on order side and existing positions
	switch message.OrderSide {
	case "BUY":
		operationType, err = w.handleBuyOrder(processCtx, message)
	case "SELL":
		operationType, err = w.handleSellOrder(processCtx, message)
	default:
		err = fmt.Errorf("invalid order side: %s", message.OrderSide)
	}

	processingTime := time.Since(startTime)
	w.updateProcessingTime(processingTime)

	if err != nil {
		w.incrementErrorCount()
		log.Printf("Position worker %s: Failed to process position update for order %s: %v",
			w.id, message.OrderID, err)

		if w.shouldRetryMessage(message, err) {
			return w.scheduleRetry(message, err)
		}

		return fmt.Errorf("position update processing failed: %w", err)
	}

	log.Printf("Position worker %s: Successfully processed %s for order %s in %v (symbol: %s)",
		w.id, operationType, message.OrderID, processingTime, message.Symbol)

	return nil
}

func (w *PositionUpdateWorker) handleBuyOrder(ctx context.Context, message *PositionUpdateMessage) (string, error) {
	// Check if position already exists for this user and symbol
	userID, err := uuid.Parse(message.UserID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// First, check if a position already exists
	exists, err := w.positionRepository.ExistsForUser(ctx, userID, message.Symbol)
	if err != nil {
		return "", fmt.Errorf("failed to check existing position: %w", err)
	}

	sourceOrderID := message.OrderID

	if !exists {
		// Create new position for buy order
		createCmd := &command.CreatePositionCommand{
			UserID:        message.UserID,
			Symbol:        message.Symbol,
			Quantity:      message.Quantity,
			Price:         message.ExecutionPrice,
			PositionType:  "LONG", // Buy orders create long positions
			SourceOrderID: &sourceOrderID,
			CreatedFrom:   "ORDER_EXECUTION",
		}

		_, err := w.createPositionUC.Execute(ctx, createCmd)
		if err != nil {
			return "", fmt.Errorf("failed to create position: %w", err)
		}

		w.incrementCreatedCount()
		return "position_create", nil
	} else {
		// Update existing position for buy order
		//TODO: create repo method to fetch only one position instead of all positions
		positions, err := w.positionRepository.FindByUserID(ctx, userID)
		if err != nil {
			return "", fmt.Errorf("failed to find existing positions: %w", err)
		}

		// Find the position for this symbol
		var targetPosition *domain.Position
		for _, pos := range positions {
			if pos.Symbol == message.Symbol && pos.Status == domain.PositionStatusActive {
				targetPosition = pos
				break
			}
		}

		if targetPosition == nil {
			return "", fmt.Errorf("position not found for user %s and symbol %s", message.UserID, message.Symbol)
		}

		updateCmd := &command.UpdatePositionCommand{
			PositionID:    targetPosition.ID.String(),
			UserID:        message.UserID,
			TradeQuantity: message.Quantity,
			TradePrice:    message.ExecutionPrice,
			IsBuyOrder:    true,
			SourceOrderID: &sourceOrderID,
		}

		_, err = w.updatePositionUC.Execute(ctx, updateCmd)
		if err != nil {
			return "", fmt.Errorf("failed to update position: %w", err)
		}

		w.incrementUpdatedCount()
		return "position_update", nil
	}
}

func (w *PositionUpdateWorker) handleSellOrder(ctx context.Context, message *PositionUpdateMessage) (string, error) {
	// For sell orders, find the existing position and update it
	userID, err := uuid.Parse(message.UserID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// Find positions for this user
	positions, err := w.positionRepository.FindByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to find existing positions: %w", err)
	}

	// Find the position for this symbol
	var targetPosition *domain.Position
	for _, pos := range positions {
		if pos.Symbol == message.Symbol && pos.Status == domain.PositionStatusActive {
			targetPosition = pos
			break
		}
	}

	if targetPosition == nil {
		return "", fmt.Errorf("no active position found for user %s and symbol %s", message.UserID, message.Symbol)
	}

	sourceOrderID := message.OrderID

	// Check if this sell will close the position entirely
	if message.Quantity >= targetPosition.Quantity {
		closeCmd := &command.ClosePositionCommand{
			PositionID:    targetPosition.ID.String(),
			UserID:        message.UserID,
			ClosePrice:    message.ExecutionPrice,
			SourceOrderID: &sourceOrderID,
			CloseReason:   "ORDER_EXECUTION",
		}

		_, err = w.closePositionUC.Execute(ctx, closeCmd)
		if err != nil {
			return "", fmt.Errorf("failed to close position: %w", err)
		}

		w.incrementClosedCount()
		return "position_close", nil
	} else {
		// Partial sell - update the position
		updateCmd := &command.UpdatePositionCommand{
			PositionID:    targetPosition.ID.String(),
			UserID:        message.UserID,
			TradeQuantity: message.Quantity,
			TradePrice:    message.ExecutionPrice,
			IsBuyOrder:    false,
			SourceOrderID: &sourceOrderID,
		}

		_, err = w.updatePositionUC.Execute(ctx, updateCmd)
		if err != nil {
			return "", fmt.Errorf("failed to update position for sell order: %w", err)
		}

		w.incrementUpdatedCount()
		return "position_update", nil
	}
}

func (w *PositionUpdateWorker) shouldRetryMessage(message *PositionUpdateMessage, err error) bool {
	if message.MessageMetadata.RetryAttempt >= w.config.MaxRetries {
		return false
	}

	return w.isRetryableError(err)
}

func (w *PositionUpdateWorker) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := strings.ToLower(err.Error())

	// Network/connection errors are retryable
	retryablePatterns := []string{
		"connection",
		"timeout",
		"temporary",
		"network",
		"unavailable",
		"deadline exceeded",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errorStr, pattern) {
			return true
		}
	}

	return false
}

func (w *PositionUpdateWorker) scheduleRetry(message *PositionUpdateMessage, err error) error {
	w.incrementRetryCount()

	// Calculate retry delay based on attempt number
	retryDelay := w.config.RetryBackoffBase * time.Duration(message.MessageMetadata.RetryAttempt+1)

	log.Printf("Position worker %s: Scheduling retry %d/%d for order %s after %v (error: %v)",
		w.id, message.MessageMetadata.RetryAttempt+1, w.config.MaxRetries,
		message.OrderID, retryDelay, err)

	// Update message metadata for retry
	message.MessageMetadata.RetryAttempt++
	message.MessageMetadata.Timestamp = time.Now()

	// Serialize message
	messageBytes, marshalErr := json.Marshal(message)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal retry message: %w", marshalErr)
	}

	// Send to retry queue
	return w.queueManager.PublishToRetryQueue(
		w.ctx,
		messageBytes,
		message.MessageMetadata.MessageID,
		message.MessageMetadata.RetryAttempt,
	)
}
