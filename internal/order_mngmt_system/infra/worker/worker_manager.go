package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"HubInvestments/internal/order_mngmt_system/application/usecase"
	"HubInvestments/shared/infra/messaging"
)

// WorkerManager manages multiple order workers for scaling and monitoring
type WorkerManager struct {
	workers        map[string]*OrderWorker
	processOrderUC usecase.IProcessOrderUseCase
	messageHandler messaging.MessageHandler
	config         *WorkerManagerConfig
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	mu             sync.RWMutex
	isRunning      bool
	metrics        *WorkerManagerMetrics
	healthChecker  *HealthChecker
	autoScaler     *AutoScaler
}

// WorkerManagerConfig contains configuration for the worker manager
type WorkerManagerConfig struct {
	MinWorkers                int
	MaxWorkers                int
	DefaultWorkers            int
	WorkerConfig              *WorkerConfig
	HealthCheckInterval       time.Duration
	MetricsCollectionInterval time.Duration
	AutoScalingEnabled        bool
	ScaleUpThreshold          float64 // Queue depth ratio to trigger scale up
	ScaleDownThreshold        float64 // Queue depth ratio to trigger scale down
	ScaleUpCooldown           time.Duration
	ScaleDownCooldown         time.Duration
	ShutdownTimeout           time.Duration
	EnableDetailedMetrics     bool
}

// WorkerManagerMetrics tracks overall worker manager performance
type WorkerManagerMetrics struct {
	ActiveWorkers         int
	TotalOrdersProcessed  int64
	TotalOrdersSuccessful int64
	TotalOrdersFailed     int64
	TotalOrdersRetried    int64
	AverageProcessingTime time.Duration
	QueueDepth            int64
	WorkerUtilization     float64
	LastScaleEvent        time.Time
	ScaleUpEvents         int64
	ScaleDownEvents       int64
	StartTime             time.Time
	LastMetricsUpdate     time.Time
	mu                    sync.RWMutex
}

// HealthChecker monitors worker health and performs recovery actions
type HealthChecker struct {
	manager             *WorkerManager
	unhealthyWorkers    map[string]time.Time
	recoveryAttempts    map[string]int
	maxRecoveryAttempts int
	recoveryDelay       time.Duration
	mu                  sync.RWMutex
}

// AutoScaler handles automatic scaling of workers based on load
type AutoScaler struct {
	manager             *WorkerManager
	lastScaleUp         time.Time
	lastScaleDown       time.Time
	scaleUpInProgress   bool
	scaleDownInProgress bool
	mu                  sync.RWMutex
}

// NewWorkerManager creates a new worker manager instance
func NewWorkerManager(
	processOrderUC usecase.IProcessOrderUseCase,
	messageHandler messaging.MessageHandler,
	config *WorkerManagerConfig,
) *WorkerManager {
	if config == nil {
		config = DefaultWorkerManagerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	wm := &WorkerManager{
		workers:        make(map[string]*OrderWorker),
		processOrderUC: processOrderUC,
		messageHandler: messageHandler,
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
		metrics:        NewWorkerManagerMetrics(),
	}

	wm.healthChecker = NewHealthChecker(wm)
	wm.autoScaler = NewAutoScaler(wm)

	return wm
}

// DefaultWorkerManagerConfig returns default configuration for worker manager
func DefaultWorkerManagerConfig() *WorkerManagerConfig {
	return &WorkerManagerConfig{
		MinWorkers:                2,
		MaxWorkers:                20,
		DefaultWorkers:            5,
		WorkerConfig:              DefaultWorkerConfig(""),
		HealthCheckInterval:       30 * time.Second,
		MetricsCollectionInterval: 10 * time.Second,
		AutoScalingEnabled:        true,
		ScaleUpThreshold:          0.8, // Scale up when queue depth > 80% of capacity
		ScaleDownThreshold:        0.2, // Scale down when queue depth < 20% of capacity
		ScaleUpCooldown:           2 * time.Minute,
		ScaleDownCooldown:         5 * time.Minute,
		ShutdownTimeout:           60 * time.Second,
		EnableDetailedMetrics:     true,
	}
}

// NewWorkerManagerMetrics creates new worker manager metrics instance
func NewWorkerManagerMetrics() *WorkerManagerMetrics {
	return &WorkerManagerMetrics{
		StartTime:         time.Now(),
		LastMetricsUpdate: time.Now(),
	}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(manager *WorkerManager) *HealthChecker {
	return &HealthChecker{
		manager:             manager,
		unhealthyWorkers:    make(map[string]time.Time),
		recoveryAttempts:    make(map[string]int),
		maxRecoveryAttempts: 3,
		recoveryDelay:       30 * time.Second,
	}
}

// NewAutoScaler creates a new auto scaler
func NewAutoScaler(manager *WorkerManager) *AutoScaler {
	return &AutoScaler{
		manager: manager,
	}
}

// Start initializes and starts the worker manager
func (wm *WorkerManager) Start() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.isRunning {
		return fmt.Errorf("worker manager is already running")
	}

	log.Printf("Starting worker manager with config: min=%d, max=%d, default=%d",
		wm.config.MinWorkers, wm.config.MaxWorkers, wm.config.DefaultWorkers)

	// Start initial workers
	for i := 0; i < wm.config.DefaultWorkers; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		if err := wm.startWorker(workerID); err != nil {
			log.Printf("Failed to start initial worker %s: %v", workerID, err)
			// Continue starting other workers
		}
	}

	wm.isRunning = true

	// Start management goroutines
	wm.wg.Add(1)
	go wm.healthCheckLoop()

	wm.wg.Add(1)
	go wm.metricsCollectionLoop()

	if wm.config.AutoScalingEnabled {
		wm.wg.Add(1)
		go wm.autoScalingLoop()
	}

	log.Printf("Worker manager started successfully with %d workers", len(wm.workers))
	return nil
}

// Stop gracefully shuts down all workers and the manager
func (wm *WorkerManager) Stop() error {
	wm.mu.Lock()
	if !wm.isRunning {
		wm.mu.Unlock()
		return fmt.Errorf("worker manager is not running")
	}
	wm.isRunning = false
	wm.mu.Unlock()

	log.Printf("Stopping worker manager with %d workers...", len(wm.workers))

	// Cancel context to signal all goroutines to stop
	wm.cancel()

	// Stop all workers
	var workerWg sync.WaitGroup
	for workerID, worker := range wm.workers {
		workerWg.Add(1)
		go func(id string, w *OrderWorker) {
			defer workerWg.Done()
			if err := w.Stop(); err != nil {
				log.Printf("Error stopping worker %s: %v", id, err)
			}
		}(workerID, worker)
	}

	// Wait for all workers to stop with timeout
	done := make(chan struct{})
	go func() {
		workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("All workers stopped successfully")
	case <-time.After(wm.config.ShutdownTimeout):
		log.Printf("Worker shutdown timeout exceeded")
	}

	// Wait for management goroutines to finish
	managerDone := make(chan struct{})
	go func() {
		wm.wg.Wait()
		close(managerDone)
	}()

	select {
	case <-managerDone:
		log.Printf("Worker manager stopped gracefully")
		return nil
	case <-time.After(wm.config.ShutdownTimeout):
		log.Printf("Worker manager shutdown timeout exceeded")
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// IsRunning returns whether the worker manager is currently running
func (wm *WorkerManager) IsRunning() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.isRunning
}

// GetWorkerCount returns the current number of active workers
func (wm *WorkerManager) GetWorkerCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.workers)
}

// GetWorkerInfo returns information about all workers
func (wm *WorkerManager) GetWorkerInfo() map[string]WorkerInfo {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	info := make(map[string]WorkerInfo)
	for workerID, worker := range wm.workers {
		info[workerID] = worker.GetWorkerInfo()
	}
	return info
}

// GetMetrics returns current worker manager metrics
func (wm *WorkerManager) GetMetrics() WorkerManagerMetrics {
	wm.metrics.mu.RLock()
	defer wm.metrics.mu.RUnlock()

	// Create a copy without the mutex to avoid copying lock value
	return WorkerManagerMetrics{
		ActiveWorkers:         wm.metrics.ActiveWorkers,
		TotalOrdersProcessed:  wm.metrics.TotalOrdersProcessed,
		TotalOrdersSuccessful: wm.metrics.TotalOrdersSuccessful,
		TotalOrdersFailed:     wm.metrics.TotalOrdersFailed,
		TotalOrdersRetried:    wm.metrics.TotalOrdersRetried,
		AverageProcessingTime: wm.metrics.AverageProcessingTime,
		QueueDepth:            wm.metrics.QueueDepth,
		WorkerUtilization:     wm.metrics.WorkerUtilization,
		LastScaleEvent:        wm.metrics.LastScaleEvent,
		ScaleUpEvents:         wm.metrics.ScaleUpEvents,
		ScaleDownEvents:       wm.metrics.ScaleDownEvents,
		StartTime:             wm.metrics.StartTime,
		LastMetricsUpdate:     wm.metrics.LastMetricsUpdate,
	}
}

// GetHealthStatus returns overall health status of the worker manager
func (wm *WorkerManager) GetHealthStatus() ManagerHealthStatus {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if !wm.isRunning {
		return ManagerHealthStatus{
			Status:           "stopped",
			ActiveWorkers:    0,
			HealthyWorkers:   0,
			UnhealthyWorkers: 0,
		}
	}

	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0

	for _, worker := range wm.workers {
		switch worker.GetHealthStatus() {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		}
	}

	status := "healthy"
	if unhealthyCount > 0 || len(wm.workers) < wm.config.MinWorkers {
		status = "unhealthy"
	} else if degradedCount > 0 || len(wm.workers) < wm.config.DefaultWorkers {
		status = "degraded"
	}

	return ManagerHealthStatus{
		Status:           status,
		ActiveWorkers:    len(wm.workers),
		HealthyWorkers:   healthyCount,
		DegradedWorkers:  degradedCount,
		UnhealthyWorkers: unhealthyCount,
		MinWorkers:       wm.config.MinWorkers,
		MaxWorkers:       wm.config.MaxWorkers,
	}
}

// ManagerHealthStatus represents the health status of the worker manager
type ManagerHealthStatus struct {
	Status           string
	ActiveWorkers    int
	HealthyWorkers   int
	DegradedWorkers  int
	UnhealthyWorkers int
	MinWorkers       int
	MaxWorkers       int
}

// ScaleUp adds new workers up to the maximum limit
func (wm *WorkerManager) ScaleUp(count int) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	currentCount := len(wm.workers)
	maxNewWorkers := wm.config.MaxWorkers - currentCount

	if maxNewWorkers <= 0 {
		return fmt.Errorf("already at maximum worker capacity (%d)", wm.config.MaxWorkers)
	}

	if count > maxNewWorkers {
		count = maxNewWorkers
	}

	log.Printf("Scaling up by %d workers (current: %d, max: %d)", count, currentCount, wm.config.MaxWorkers)

	var errors []error
	successCount := 0

	for i := 0; i < count; i++ {
		workerID := fmt.Sprintf("worker-%d", currentCount+i+1)
		if err := wm.startWorker(workerID); err != nil {
			errors = append(errors, fmt.Errorf("failed to start worker %s: %w", workerID, err))
		} else {
			successCount++
		}
	}

	wm.updateScaleUpMetrics(successCount)

	if len(errors) > 0 {
		return fmt.Errorf("scale up partially failed: started %d/%d workers, errors: %v", successCount, count, errors)
	}

	log.Printf("Successfully scaled up by %d workers", successCount)
	return nil
}

// ScaleDown removes workers down to the minimum limit
func (wm *WorkerManager) ScaleDown(count int) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	currentCount := len(wm.workers)
	maxRemovable := currentCount - wm.config.MinWorkers

	if maxRemovable <= 0 {
		return fmt.Errorf("already at minimum worker capacity (%d)", wm.config.MinWorkers)
	}

	if count > maxRemovable {
		count = maxRemovable
	}

	log.Printf("Scaling down by %d workers (current: %d, min: %d)", count, currentCount, wm.config.MinWorkers)

	// Select workers to remove (prefer unhealthy ones first)
	workersToRemove := wm.selectWorkersForRemoval(count)

	var errors []error
	successCount := 0

	for _, workerID := range workersToRemove {
		if err := wm.stopWorker(workerID); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop worker %s: %w", workerID, err))
		} else {
			successCount++
		}
	}

	wm.updateScaleDownMetrics(successCount)

	if len(errors) > 0 {
		return fmt.Errorf("scale down partially failed: stopped %d/%d workers, errors: %v", successCount, count, errors)
	}

	log.Printf("Successfully scaled down by %d workers", successCount)
	return nil
}

// startWorker creates and starts a new worker
func (wm *WorkerManager) startWorker(workerID string) error {
	// Create worker config
	workerConfig := *wm.config.WorkerConfig
	workerConfig.WorkerID = workerID

	// Create and start worker (consumer will be created inside the worker)
	worker := NewOrderWorker(workerID, wm.processOrderUC, nil, wm.messageHandler, &workerConfig)

	if err := worker.Start(); err != nil {
		return fmt.Errorf("failed to start worker %s: %w", workerID, err)
	}

	wm.workers[workerID] = worker
	log.Printf("Started worker %s", workerID)
	return nil
}

// stopWorker stops and removes a worker
func (wm *WorkerManager) stopWorker(workerID string) error {
	worker, exists := wm.workers[workerID]
	if !exists {
		return fmt.Errorf("worker %s not found", workerID)
	}

	if err := worker.Stop(); err != nil {
		return fmt.Errorf("failed to stop worker %s: %w", workerID, err)
	}

	delete(wm.workers, workerID)
	log.Printf("Stopped worker %s", workerID)
	return nil
}

// selectWorkersForRemoval selects workers to remove during scale down
func (wm *WorkerManager) selectWorkersForRemoval(count int) []string {
	var unhealthy, degraded, healthy []string

	for workerID, worker := range wm.workers {
		switch worker.GetHealthStatus() {
		case HealthStatusUnhealthy:
			unhealthy = append(unhealthy, workerID)
		case HealthStatusDegraded:
			degraded = append(degraded, workerID)
		case HealthStatusHealthy:
			healthy = append(healthy, workerID)
		}
	}

	var selected []string

	// Remove unhealthy workers first
	for i := 0; i < len(unhealthy) && len(selected) < count; i++ {
		selected = append(selected, unhealthy[i])
	}

	// Then degraded workers
	for i := 0; i < len(degraded) && len(selected) < count; i++ {
		selected = append(selected, degraded[i])
	}

	// Finally healthy workers if needed
	for i := 0; i < len(healthy) && len(selected) < count; i++ {
		selected = append(selected, healthy[i])
	}

	return selected
}

// healthCheckLoop performs periodic health checks on all workers
func (wm *WorkerManager) healthCheckLoop() {
	defer wm.wg.Done()

	ticker := time.NewTicker(wm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wm.performHealthChecks()
		case <-wm.ctx.Done():
			return
		}
	}
}

// performHealthChecks checks health of all workers and takes recovery actions
func (wm *WorkerManager) performHealthChecks() {
	wm.mu.RLock()
	workers := make(map[string]*OrderWorker)
	for id, worker := range wm.workers {
		workers[id] = worker
	}
	wm.mu.RUnlock()

	for workerID, worker := range workers {
		wm.healthChecker.checkWorkerHealth(workerID, worker)
	}
}

// metricsCollectionLoop collects and updates metrics periodically
func (wm *WorkerManager) metricsCollectionLoop() {
	defer wm.wg.Done()

	ticker := time.NewTicker(wm.config.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wm.collectMetrics()
		case <-wm.ctx.Done():
			return
		}
	}
}

// collectMetrics aggregates metrics from all workers
func (wm *WorkerManager) collectMetrics() {
	wm.mu.RLock()
	workers := make(map[string]*OrderWorker)
	for id, worker := range wm.workers {
		workers[id] = worker
	}
	wm.mu.RUnlock()

	wm.metrics.mu.Lock()
	defer wm.metrics.mu.Unlock()

	wm.metrics.ActiveWorkers = len(workers)
	wm.metrics.LastMetricsUpdate = time.Now()

	var totalProcessed, totalSuccessful, totalFailed, totalRetried int64
	var totalProcessingTime time.Duration
	var activeWorkers int

	for _, worker := range workers {
		metrics := worker.GetMetrics()
		totalProcessed += metrics.OrdersProcessed
		totalSuccessful += metrics.OrdersSuccessful
		totalFailed += metrics.OrdersFailed
		totalRetried += metrics.OrdersRetried

		if metrics.OrdersProcessed > 0 {
			totalProcessingTime += metrics.AverageProcessingTime
			activeWorkers++
		}
	}

	wm.metrics.TotalOrdersProcessed = totalProcessed
	wm.metrics.TotalOrdersSuccessful = totalSuccessful
	wm.metrics.TotalOrdersFailed = totalFailed
	wm.metrics.TotalOrdersRetried = totalRetried

	if activeWorkers > 0 {
		wm.metrics.AverageProcessingTime = totalProcessingTime / time.Duration(activeWorkers)
		wm.metrics.WorkerUtilization = float64(activeWorkers) / float64(len(workers))
	}

	// Get queue depth (simplified - in real implementation would query RabbitMQ)
	wm.metrics.QueueDepth = wm.estimateQueueDepth()
}

// estimateQueueDepth estimates current queue depth
func (wm *WorkerManager) estimateQueueDepth() int64 {
	// In a real implementation, this would query RabbitMQ for actual queue depths
	// For now, we'll use a simple estimation based on processing rates
	return 0 // Placeholder
}

// autoScalingLoop handles automatic scaling based on load
func (wm *WorkerManager) autoScalingLoop() {
	defer wm.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wm.autoScaler.evaluateScaling()
		case <-wm.ctx.Done():
			return
		}
	}
}

// Metric update methods
func (wm *WorkerManager) updateScaleUpMetrics(count int) {
	wm.metrics.mu.Lock()
	defer wm.metrics.mu.Unlock()
	wm.metrics.ScaleUpEvents++
	wm.metrics.LastScaleEvent = time.Now()
}

func (wm *WorkerManager) updateScaleDownMetrics(count int) {
	wm.metrics.mu.Lock()
	defer wm.metrics.mu.Unlock()
	wm.metrics.ScaleDownEvents++
	wm.metrics.LastScaleEvent = time.Now()
}

// Health checker methods
func (hc *HealthChecker) checkWorkerHealth(workerID string, worker *OrderWorker) {
	healthStatus := worker.GetHealthStatus()

	hc.mu.Lock()
	defer hc.mu.Unlock()

	switch healthStatus {
	case HealthStatusUnhealthy:
		if _, exists := hc.unhealthyWorkers[workerID]; !exists {
			hc.unhealthyWorkers[workerID] = time.Now()
			log.Printf("Worker %s marked as unhealthy", workerID)
		}
		hc.attemptRecovery(workerID, worker)

	case HealthStatusHealthy:
		if _, exists := hc.unhealthyWorkers[workerID]; exists {
			delete(hc.unhealthyWorkers, workerID)
			delete(hc.recoveryAttempts, workerID)
			log.Printf("Worker %s recovered to healthy status", workerID)
		}
	}
}

func (hc *HealthChecker) attemptRecovery(workerID string, worker *OrderWorker) {
	attempts := hc.recoveryAttempts[workerID]
	if attempts >= hc.maxRecoveryAttempts {
		log.Printf("Worker %s exceeded max recovery attempts (%d), scheduling for replacement",
			workerID, hc.maxRecoveryAttempts)
		// In a real implementation, we might schedule worker replacement here
		return
	}

	lastAttempt, exists := hc.unhealthyWorkers[workerID]
	if exists && time.Since(lastAttempt) < hc.recoveryDelay {
		return // Too soon for another recovery attempt
	}

	log.Printf("Attempting recovery for worker %s (attempt %d/%d)",
		workerID, attempts+1, hc.maxRecoveryAttempts)

	// Simple recovery: restart the worker
	// In a more sophisticated implementation, we might try different recovery strategies
	hc.recoveryAttempts[workerID] = attempts + 1
	hc.unhealthyWorkers[workerID] = time.Now()
}

// Auto scaler methods
func (as *AutoScaler) evaluateScaling() {
	as.mu.Lock()
	defer as.mu.Unlock()

	metrics := as.manager.GetMetrics()
	currentWorkers := as.manager.GetWorkerCount()

	// Calculate load metrics
	queueDepthRatio := as.calculateQueueDepthRatio(&metrics)

	// Check if we should scale up
	if as.shouldScaleUp(queueDepthRatio, currentWorkers) {
		as.performScaleUp()
	} else if as.shouldScaleDown(queueDepthRatio, currentWorkers) {
		as.performScaleDown()
	}
}

func (as *AutoScaler) calculateQueueDepthRatio(metrics *WorkerManagerMetrics) float64 {
	if metrics.ActiveWorkers == 0 {
		return 1.0 // High ratio to trigger scale up
	}

	// Simplified calculation - in real implementation would use actual queue metrics
	capacity := float64(metrics.ActiveWorkers * 10) // Assume each worker can handle 10 orders
	load := float64(metrics.QueueDepth)

	if capacity == 0 {
		return 1.0
	}

	return load / capacity
}

func (as *AutoScaler) shouldScaleUp(queueDepthRatio float64, currentWorkers int) bool {
	if as.scaleUpInProgress || currentWorkers >= as.manager.config.MaxWorkers {
		return false
	}

	if time.Since(as.lastScaleUp) < as.manager.config.ScaleUpCooldown {
		return false
	}

	return queueDepthRatio > as.manager.config.ScaleUpThreshold
}

func (as *AutoScaler) shouldScaleDown(queueDepthRatio float64, currentWorkers int) bool {
	if as.scaleDownInProgress || currentWorkers <= as.manager.config.MinWorkers {
		return false
	}

	if time.Since(as.lastScaleDown) < as.manager.config.ScaleDownCooldown {
		return false
	}

	return queueDepthRatio < as.manager.config.ScaleDownThreshold
}

func (as *AutoScaler) performScaleUp() {
	as.scaleUpInProgress = true
	defer func() { as.scaleUpInProgress = false }()

	scaleCount := 1 // Conservative scaling
	log.Printf("Auto-scaling up by %d workers", scaleCount)

	if err := as.manager.ScaleUp(scaleCount); err != nil {
		log.Printf("Auto scale up failed: %v", err)
	} else {
		as.lastScaleUp = time.Now()
	}
}

func (as *AutoScaler) performScaleDown() {
	as.scaleDownInProgress = true
	defer func() { as.scaleDownInProgress = false }()

	scaleCount := 1 // Conservative scaling
	log.Printf("Auto-scaling down by %d workers", scaleCount)

	if err := as.manager.ScaleDown(scaleCount); err != nil {
		log.Printf("Auto scale down failed: %v", err)
	} else {
		as.lastScaleDown = time.Now()
	}
}
