package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"HubInvestments/internal/order_mngmt_system/application/usecase"
	"HubInvestments/internal/order_mngmt_system/infra/messaging/rabbitmq"
	"HubInvestments/shared/infra/messaging"
)

// OrderWorker handles asynchronous order processing
type OrderWorker struct {
	id             string
	processOrderUC usecase.IProcessOrderUseCase
	consumer       *rabbitmq.OrderConsumer
	messageHandler messaging.MessageHandler
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	isRunning      bool
	mu             sync.RWMutex
	config         *WorkerConfig
	metrics        *WorkerMetrics
	healthStatus   HealthStatus
	lastHeartbeat  time.Time
	processedCount int64
	errorCount     int64
	retryCount     int64
}

type WorkerConfig struct {
	WorkerID            string
	MaxConcurrentOrders int
	ProcessingTimeout   time.Duration
	HeartbeatInterval   time.Duration
	MaxRetries          int
	RetryBackoffBase    time.Duration
	HealthCheckInterval time.Duration
	ShutdownTimeout     time.Duration
	EnableMetrics       bool
	LogLevel            string
}

type WorkerMetrics struct {
	OrdersProcessed       int64
	OrdersSuccessful      int64
	OrdersFailed          int64
	OrdersRetried         int64
	AverageProcessingTime time.Duration
	LastProcessingTime    time.Duration
	StartTime             time.Time
	LastActivityTime      time.Time
	mu                    sync.RWMutex
}

// WorkerMetricsSnapshot represents a point-in-time snapshot of worker metrics without mutex
type WorkerMetricsSnapshot struct {
	OrdersProcessed       int64
	OrdersSuccessful      int64
	OrdersFailed          int64
	OrdersRetried         int64
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

func NewOrderWorker(
	workerID string,
	processOrderUC usecase.IProcessOrderUseCase,
	consumer *rabbitmq.OrderConsumer,
	messageHandler messaging.MessageHandler,
	config *WorkerConfig,
) *OrderWorker {
	if config == nil {
		config = DefaultWorkerConfig(workerID)
	}

	ctx, cancel := context.WithCancel(context.Background())

	worker := &OrderWorker{
		id:             workerID,
		processOrderUC: processOrderUC,
		consumer:       consumer,
		messageHandler: messageHandler,
		ctx:            ctx,
		cancel:         cancel,
		config:         config,
		metrics:        NewWorkerMetrics(),
		healthStatus:   HealthStatusUnknown,
		lastHeartbeat:  time.Now(),
	}

	// Create consumer if not provided
	if consumer == nil {
		// Create a message handler that will be passed to the consumer
		orderMessageHandler := &OrderMessageHandler{
			worker:    worker,
			semaphore: make(chan struct{}, config.MaxConcurrentOrders),
		}
		worker.consumer = rabbitmq.NewOrderConsumer(messageHandler, orderMessageHandler)
	}

	return worker
}

func DefaultWorkerConfig(workerID string) *WorkerConfig {
	return &WorkerConfig{
		WorkerID:            workerID,
		MaxConcurrentOrders: 10,
		ProcessingTimeout:   30 * time.Second,
		HeartbeatInterval:   10 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    5 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		ShutdownTimeout:     60 * time.Second,
		EnableMetrics:       true,
		LogLevel:            "INFO",
	}
}

func NewWorkerMetrics() *WorkerMetrics {
	return &WorkerMetrics{
		StartTime:        time.Now(),
		LastActivityTime: time.Now(),
	}
}

// Start begins the worker processing loop
func (w *OrderWorker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		return fmt.Errorf("worker %s is already running", w.id)
	}

	log.Printf("Starting order worker %s with config: max_concurrent=%d, timeout=%v",
		w.id, w.config.MaxConcurrentOrders, w.config.ProcessingTimeout)

	w.isRunning = true
	w.healthStatus = HealthStatusHealthy
	w.lastHeartbeat = time.Now()

	// Start heartbeat goroutine
	w.wg.Add(1)
	go w.heartbeatLoop()

	// Start health check goroutine
	w.wg.Add(1)
	go w.healthCheckLoop()

	// Start message processing goroutine only if consumer is available
	if w.consumer != nil {
		w.wg.Add(1)
		go w.processMessages()
	}

	log.Printf("Order worker %s started successfully", w.id)
	return nil
}

// Stop gracefully shuts down the worker
func (w *OrderWorker) Stop() error {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return fmt.Errorf("worker %s is not running", w.id)
	}
	w.isRunning = false
	w.healthStatus = HealthStatusStopped
	w.mu.Unlock()

	log.Printf("Stopping order worker %s...", w.id)

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
		log.Printf("Order worker %s stopped gracefully", w.id)
		return nil
	case <-time.After(w.config.ShutdownTimeout):
		log.Printf("Order worker %s shutdown timeout exceeded", w.id)
		return fmt.Errorf("shutdown timeout exceeded for worker %s", w.id)
	}
}

func (w *OrderWorker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.isRunning
}

func (w *OrderWorker) GetHealthStatus() HealthStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.healthStatus
}

// Returns a copy without mutex to avoid copying lock value
func (w *OrderWorker) GetMetrics() WorkerMetricsSnapshot {
	w.metrics.mu.RLock()
	defer w.metrics.mu.RUnlock()

	return WorkerMetricsSnapshot{
		OrdersProcessed:       w.metrics.OrdersProcessed,
		OrdersSuccessful:      w.metrics.OrdersSuccessful,
		OrdersFailed:          w.metrics.OrdersFailed,
		OrdersRetried:         w.metrics.OrdersRetried,
		AverageProcessingTime: w.metrics.AverageProcessingTime,
		LastProcessingTime:    w.metrics.LastProcessingTime,
		StartTime:             w.metrics.StartTime,
		LastActivityTime:      w.metrics.LastActivityTime,
	}
}

func (w *OrderWorker) GetWorkerInfo() WorkerInfo {
	w.mu.RLock()
	defer w.mu.RUnlock()

	metrics := w.GetMetrics()

	return WorkerInfo{
		ID:             w.id,
		IsRunning:      w.isRunning,
		HealthStatus:   w.healthStatus.String(),
		LastHeartbeat:  w.lastHeartbeat,
		ProcessedCount: w.processedCount,
		ErrorCount:     w.errorCount,
		RetryCount:     w.retryCount,
		Uptime:         time.Since(metrics.StartTime),
		Config:         *w.config,
		Metrics:        metrics, // Already a copy without mutex
	}
}

// WorkerInfo contains comprehensive information about a worker
type WorkerInfo struct {
	ID             string
	IsRunning      bool
	HealthStatus   string
	LastHeartbeat  time.Time
	ProcessedCount int64
	ErrorCount     int64
	RetryCount     int64
	Uptime         time.Duration
	Config         WorkerConfig
	Metrics        WorkerMetricsSnapshot
}

// processMessages is the main message processing loop
func (w *OrderWorker) processMessages() {
	defer w.wg.Done()

	log.Printf("Worker %s: Starting message processing loop", w.id)

	// Start consuming messages with context and config
	config := &rabbitmq.ConsumerConfig{
		ConcurrentWorkers: w.config.MaxConcurrentOrders,
		PrefetchCount:     w.config.MaxConcurrentOrders * 2,
		RequeueOnError:    true,
		RetryDelay:        w.config.RetryBackoffBase,
		MaxRetries:        w.config.MaxRetries,
	}

	err := w.consumer.StartConsumers(w.ctx, config)
	if err != nil {
		log.Printf("Worker %s: Failed to start consumers: %v", w.id, err)
		w.updateHealthStatus(HealthStatusUnhealthy)
		return
	}

	// Wait for context cancellation
	<-w.ctx.Done()

	log.Printf("Worker %s: Stopping message processing", w.id)

	// Stop consumers
	err = w.consumer.StopConsumers(w.ctx)
	if err != nil {
		log.Printf("Worker %s: Error stopping consumers: %v", w.id, err)
	}
}

// OrderMessageHandler implements the message handling interface
type OrderMessageHandler struct {
	worker    *OrderWorker
	semaphore chan struct{}
}

func (h *OrderMessageHandler) HandleOrderMessage(ctx context.Context, message *rabbitmq.OrderMessage) error {
	// Semaphore pattern: limit concurrent order processing to prevent resource exhaustion
	// The semaphore channel acts as a counting semaphore with MaxConcurrentOrders capacity
	// - Sending to channel acquires a "slot" (blocks if all slots are taken)
	// - Receiving from channel releases a "slot" (allows another goroutine to proceed)
	// This ensures we never process more than MaxConcurrentOrders simultaneously
	select {
	case h.semaphore <- struct{}{}:
		defer func() { <-h.semaphore }() // Release semaphore slot when processing completes
	case <-ctx.Done():
		return ctx.Err() // Context cancelled, abort processing
	}

	return h.worker.processOrderMessage(ctx, message)
}

func (h *OrderMessageHandler) HandleStatusUpdate(ctx context.Context, update *rabbitmq.OrderStatusUpdate) error {
	log.Printf("Worker %s: Received status update for order %s: %s -> %s",
		h.worker.id, update.OrderID, update.PreviousStatus, update.CurrentStatus)
	return nil
}

func (w *OrderWorker) processOrderMessage(ctx context.Context, message *rabbitmq.OrderMessage) error {
	startTime := time.Now()

	// Create processing context with timeout
	processCtx, cancel := context.WithTimeout(ctx, w.config.ProcessingTimeout)
	defer cancel()

	log.Printf("Worker %s: Processing order %s (symbol: %s, quantity: %.2f)",
		w.id, message.OrderID, message.Symbol, message.Quantity)

	w.updateLastActivity()
	w.incrementProcessedCount()

	command := &usecase.ProcessOrderCommand{
		OrderID: message.OrderID,
		Context: usecase.ProcessingContext{
			WorkerID:     w.id,
			ProcessingID: fmt.Sprintf("%s-%d", w.id, time.Now().UnixNano()),
			StartTime:    startTime,
			MaxRetries:   w.config.MaxRetries,
			RetryAttempt: message.MessageMetadata.RetryAttempt,
		},
	}

	result, err := w.processOrderUC.Execute(processCtx, command)

	processingTime := time.Since(startTime)
	w.updateProcessingTime(processingTime)

	if err != nil {
		w.incrementErrorCount()
		log.Printf("Worker %s: Failed to process order %s: %v", w.id, message.OrderID, err)

		if w.shouldRetryOrder(message, err) {
			return w.scheduleRetry(message, err)
		}

		return fmt.Errorf("order processing failed: %w", err)
	}

	log.Printf("Worker %s: Successfully processed order %s in %v (status: %s)",
		w.id, message.OrderID, processingTime, result.FinalStatus)

	return nil
}

func (w *OrderWorker) shouldRetryOrder(message *rabbitmq.OrderMessage, err error) bool {
	if message.MessageMetadata.RetryAttempt >= w.config.MaxRetries {
		return false
	}

	return w.isRetryableError(err)
}

// isRetryableError checks if an error is retryable
func (w *OrderWorker) isRetryableError(err error) bool {
	// Network errors, temporary failures, etc. are retryable
	// Business logic errors, validation errors are not retryable
	errorStr := err.Error()

	retryableErrors := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"service unavailable",
		"market data unavailable",
	}

	for _, retryableError := range retryableErrors {
		if contains(errorStr, retryableError) {
			return true
		}
	}

	return false
}

func (w *OrderWorker) scheduleRetry(message *rabbitmq.OrderMessage, err error) error {
	retryAttempt := message.MessageMetadata.RetryAttempt

	log.Printf("Worker %s: Scheduling retry for order %s (attempt %d/%d)",
		w.id, message.OrderID, retryAttempt+1, w.config.MaxRetries)

	message.MessageMetadata.RetryAttempt++

	messageBytes, marshalErr := json.Marshal(message)
	if marshalErr != nil {
		return fmt.Errorf("failed to serialize message for retry: %w", marshalErr)
	}
	queueManager := w.consumer.GetQueueManager()
	return queueManager.PublishToRetryQueue(w.ctx, messageBytes, message.MessageMetadata.MessageID, retryAttempt)
}

// Exponential backoff with maximum delay cap at 1 hour
func (w *OrderWorker) calculateRetryDelay(retryCount int) time.Duration {
	delay := w.config.RetryBackoffBase * time.Duration(1<<uint(retryCount))

	maxDelay := 1 * time.Hour
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

func (w *OrderWorker) heartbeatLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.sendHeartbeat()
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *OrderWorker) sendHeartbeat() {
	w.mu.Lock()
	w.lastHeartbeat = time.Now()
	w.mu.Unlock()

	log.Printf("Worker %s: Heartbeat sent", w.id)
}

func (w *OrderWorker) healthCheckLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.performHealthCheck()
		case <-w.ctx.Done():
			return
		}
	}
}

// Evaluates worker health based on consumer status and error rates
func (w *OrderWorker) performHealthCheck() {
	if w.consumer != nil {
		if !w.consumer.IsRunning() {
			w.updateHealthStatus(HealthStatusUnhealthy)
			return
		}

		if err := w.consumer.HealthCheck(context.Background()); err != nil {
			log.Printf("Worker %s: Consumer health check failed: %v", w.id, err)
			w.updateHealthStatus(HealthStatusDegraded)
			return
		}
	}

	if w.messageHandler != nil {
		if err := w.messageHandler.HealthCheck(context.Background()); err != nil {
			log.Printf("Worker %s: Message handler health check failed: %v", w.id, err)
			w.updateHealthStatus(HealthStatusDegraded)
			return
		}
	}

	// Mark as degraded if error rate exceeds 10% threshold
	metrics := w.GetMetrics()
	if metrics.OrdersProcessed > 0 {
		errorRate := float64(w.errorCount) / float64(metrics.OrdersProcessed)
		if errorRate > 0.1 {
			w.updateHealthStatus(HealthStatusDegraded)
			return
		}
	}

	w.updateHealthStatus(HealthStatusHealthy)
}

func (w *OrderWorker) updateHealthStatus(status HealthStatus) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.healthStatus != status {
		log.Printf("Worker %s: Health status changed from %s to %s",
			w.id, w.healthStatus.String(), status.String())
		w.healthStatus = status
	}
}

func (w *OrderWorker) incrementProcessedCount() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.metrics.OrdersProcessed++
	w.processedCount++
}

func (w *OrderWorker) incrementErrorCount() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.metrics.OrdersFailed++
	w.errorCount++
}

func (w *OrderWorker) incrementRetryCount() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.metrics.OrdersRetried++
	w.retryCount++
}

func (w *OrderWorker) updateProcessingTime(duration time.Duration) {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()

	w.metrics.LastProcessingTime = duration

	// Calculate rolling average processing time
	if w.metrics.OrdersProcessed > 0 {
		totalTime := w.metrics.AverageProcessingTime * time.Duration(w.metrics.OrdersProcessed-1)
		w.metrics.AverageProcessingTime = (totalTime + duration) / time.Duration(w.metrics.OrdersProcessed)
	} else {
		w.metrics.AverageProcessingTime = duration
	}
}

func (w *OrderWorker) updateLastActivity() {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()
	w.metrics.LastActivityTime = time.Now()
}

// contains checks if a substring exists anywhere within a string
// Used for error classification to determine if errors are retryable
// This implementation checks multiple positions to find the substring:
// 1. Exact match (s == substr)
// 2. Prefix match (starts with substr)
// 3. Suffix match (ends with substr)
// 4. Substring match (contains substr anywhere in the middle)
func contains(s, substr string) bool {
	// Early return if substring is longer than string
	if len(s) < len(substr) {
		return false
	}

	// Check exact match
	if s == substr {
		return true
	}

	// Check prefix match
	if len(s) >= len(substr) && s[:len(substr)] == substr {
		return true
	}

	// Check suffix match
	if len(s) >= len(substr) && s[len(s)-len(substr):] == substr {
		return true
	}

	// Check substring match in the middle
	if len(s) > len(substr) {
		for i := 1; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
	}

	return false
}
