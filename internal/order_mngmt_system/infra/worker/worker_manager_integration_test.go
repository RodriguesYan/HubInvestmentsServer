package worker

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"HubInvestments/shared/infra/messaging"
)

func TestWorkerManager_Start_Success(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := NewMockMessageHandler()

	config := &WorkerManagerConfig{
		MinWorkers:                2,
		MaxWorkers:                5,
		DefaultWorkers:            2,
		WorkerConfig:              DefaultWorkerConfig("test"),
		HealthCheckInterval:       1 * time.Second,
		MetricsCollectionInterval: 500 * time.Millisecond,
		AutoScalingEnabled:        false, // Disable for simpler testing
		ScaleUpThreshold:          0.8,
		ScaleDownThreshold:        0.2,
		ScaleUpCooldown:           1 * time.Second,
		ScaleDownCooldown:         2 * time.Second,
		ShutdownTimeout:           5 * time.Second,
		EnableDetailedMetrics:     true,
	}

	manager := NewWorkerManager(mockUseCase, mockHandler, config)

	// Act
	err := manager.Start()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error starting worker manager, got %v", err)
	}

	// Verify manager is running
	if !manager.IsRunning() {
		t.Error("Expected worker manager to be running")
	}

	// Verify initial workers are started
	workerCount := manager.GetWorkerCount()
	if workerCount != 2 {
		t.Errorf("Expected 2 active workers, got %d", workerCount)
	}

	// Cleanup
	manager.Stop()
}

func TestWorkerManager_Stop_Success(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := NewMockMessageHandler()

	config := DefaultWorkerManagerConfig()
	config.AutoScalingEnabled = false // Disable for simpler testing
	manager := NewWorkerManager(mockUseCase, mockHandler, config)
	manager.Start()

	// Verify manager is running
	if !manager.IsRunning() {
		t.Error("Expected worker manager to be running")
	}

	// Act
	err := manager.Stop()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error stopping worker manager, got %v", err)
	}

	// Verify manager is stopped
	if manager.IsRunning() {
		t.Error("Expected worker manager to be stopped")
	}

	// Verify all workers are stopped
	workerCount := manager.GetWorkerCount()
	if workerCount != 0 {
		t.Errorf("Expected 0 active workers after stop, got %d", workerCount)
	}
}

func TestWorkerManager_ScaleUp(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := NewMockMessageHandler()

	config := &WorkerManagerConfig{
		MinWorkers:                1,
		MaxWorkers:                3,
		DefaultWorkers:            1,
		WorkerConfig:              DefaultWorkerConfig("test"),
		HealthCheckInterval:       100 * time.Millisecond,
		MetricsCollectionInterval: 50 * time.Millisecond,
		AutoScalingEnabled:        false, // Manual scaling for testing
		ScaleUpThreshold:          0.8,
		ScaleDownThreshold:        0.2,
		ScaleUpCooldown:           100 * time.Millisecond,
		ScaleDownCooldown:         200 * time.Millisecond,
		ShutdownTimeout:           5 * time.Second,
		EnableDetailedMetrics:     true,
	}

	manager := NewWorkerManager(mockUseCase, mockHandler, config)
	manager.Start()
	defer manager.Stop()

	// Wait for initial setup
	time.Sleep(50 * time.Millisecond)

	initialCount := manager.GetWorkerCount()
	if initialCount != 1 {
		t.Errorf("Expected 1 initial worker, got %d", initialCount)
	}

	// Act - Manual scale up
	err := manager.ScaleUp(1)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error scaling up, got %v", err)
	}

	// Wait for scaling to complete
	time.Sleep(100 * time.Millisecond)

	finalCount := manager.GetWorkerCount()
	if finalCount <= initialCount {
		t.Errorf("Expected workers to scale up from %d, got %d", initialCount, finalCount)
	}

	if finalCount > 3 {
		t.Errorf("Expected workers not to exceed max of 3, got %d", finalCount)
	}
}

func TestWorkerManager_ScaleDown(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := NewMockMessageHandler()

	config := &WorkerManagerConfig{
		MinWorkers:                1,
		MaxWorkers:                5,
		DefaultWorkers:            3,
		WorkerConfig:              DefaultWorkerConfig("test"),
		HealthCheckInterval:       100 * time.Millisecond,
		MetricsCollectionInterval: 50 * time.Millisecond,
		AutoScalingEnabled:        false, // Manual scaling for testing
		ScaleUpThreshold:          0.8,
		ScaleDownThreshold:        0.2,
		ScaleUpCooldown:           100 * time.Millisecond,
		ScaleDownCooldown:         200 * time.Millisecond,
		ShutdownTimeout:           5 * time.Second,
		EnableDetailedMetrics:     true,
	}

	manager := NewWorkerManager(mockUseCase, mockHandler, config)
	manager.Start()
	defer manager.Stop()

	// Wait for initial setup
	time.Sleep(50 * time.Millisecond)

	initialCount := manager.GetWorkerCount()
	if initialCount != 3 {
		t.Errorf("Expected 3 initial workers, got %d", initialCount)
	}

	// Act - Manual scale down
	err := manager.ScaleDown(1)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error scaling down, got %v", err)
	}

	// Wait for scaling to complete
	time.Sleep(100 * time.Millisecond)

	finalCount := manager.GetWorkerCount()
	if finalCount >= initialCount {
		t.Errorf("Expected workers to scale down from %d, got %d", initialCount, finalCount)
	}

	if finalCount < 1 {
		t.Errorf("Expected workers not to go below min of 1, got %d", finalCount)
	}
}

func TestWorkerManager_GetHealthStatus_AllHealthy(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := NewMockMessageHandler()

	config := DefaultWorkerManagerConfig()
	config.AutoScalingEnabled = false // Disable for simpler testing
	manager := NewWorkerManager(mockUseCase, mockHandler, config)
	manager.Start()
	defer manager.Stop()

	// Wait for workers to start
	time.Sleep(100 * time.Millisecond)

	// Act
	health := manager.GetHealthStatus()

	// Assert
	if health.Status != "healthy" {
		t.Errorf("Expected worker manager to be healthy, got status: %s", health.Status)
	}

	if health.ActiveWorkers == 0 {
		t.Error("Expected worker health information")
	}

	// All workers should be healthy
	if health.HealthyWorkers != health.ActiveWorkers {
		t.Errorf("Expected all %d workers to be healthy, got %d healthy workers",
			health.ActiveWorkers, health.HealthyWorkers)
	}
}

func TestWorkerManager_GetHealthStatus_MessageHandlerUnhealthy(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := &MockMessageHandler{
		consumers: make(map[string]messaging.MessageConsumer),
	}
	// Override the default healthy behavior to return an error
	mockHandler.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockHandler.On("PublishWithOptions", mock.Anything, mock.Anything).Return(nil)
	mockHandler.On("Consume", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockHandler.On("DeclareQueue", mock.Anything, mock.Anything).Return(nil)
	mockHandler.On("DeleteQueue", mock.Anything).Return(nil)
	mockHandler.On("PurgeQueue", mock.Anything).Return(nil)
	mockHandler.On("QueueInfo", mock.Anything).Return(createDefaultQueueInfo("test-queue"), nil)
	mockHandler.On("HealthCheck", mock.Anything).Return(errors.New("message handler connection failed"))
	mockHandler.On("Close").Return(nil)

	config := DefaultWorkerManagerConfig()
	config.AutoScalingEnabled = false // Disable for simpler testing
	manager := NewWorkerManager(mockUseCase, mockHandler, config)
	manager.Start()
	defer manager.Stop()

	// Wait for workers to start and health checks to run
	time.Sleep(200 * time.Millisecond)

	// Act
	health := manager.GetHealthStatus()

	// Assert
	if health.Status == "healthy" {
		t.Error("Expected worker manager to be unhealthy due to message handler")
	}

	if health.UnhealthyWorkers == 0 {
		t.Error("Expected some workers to be unhealthy due to message handler error")
	}
}

func TestWorkerManager_GetMetrics(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := NewMockMessageHandler()

	config := DefaultWorkerManagerConfig()
	config.AutoScalingEnabled = false // Disable for simpler testing
	manager := NewWorkerManager(mockUseCase, mockHandler, config)
	manager.Start()
	defer manager.Stop()

	// Wait for workers to start
	time.Sleep(100 * time.Millisecond)

	// Act
	metrics := manager.GetMetrics()

	// Assert
	if metrics.ActiveWorkers == 0 {
		t.Error("Expected active workers in metrics")
	}

	if metrics.TotalOrdersProcessed < 0 {
		t.Error("Expected non-negative total processed count")
	}

	if metrics.TotalOrdersFailed < 0 {
		t.Error("Expected non-negative total error count")
	}

	if metrics.AverageProcessingTime < 0 {
		t.Error("Expected non-negative average processing time")
	}
}

func TestWorkerManager_ScaleUp_AboveMaximum(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := NewMockMessageHandler()

	config := &WorkerManagerConfig{
		MinWorkers:                1,
		MaxWorkers:                2,
		DefaultWorkers:            2,
		WorkerConfig:              DefaultWorkerConfig("test"),
		HealthCheckInterval:       1 * time.Second,
		MetricsCollectionInterval: 500 * time.Millisecond,
		AutoScalingEnabled:        false,
		ScaleUpThreshold:          0.8,
		ScaleDownThreshold:        0.2,
		ScaleUpCooldown:           1 * time.Second,
		ScaleDownCooldown:         2 * time.Second,
		ShutdownTimeout:           5 * time.Second,
		EnableDetailedMetrics:     true,
	}

	manager := NewWorkerManager(mockUseCase, mockHandler, config)
	manager.Start()
	defer manager.Stop()

	// Wait for initial setup
	time.Sleep(50 * time.Millisecond)

	// Act - Try to scale above maximum
	err := manager.ScaleUp(2) // This should exceed the max

	// Assert
	if err == nil {
		t.Fatal("Expected error when trying to scale above maximum")
	}

	if !contains(err.Error(), "maximum") {
		t.Errorf("Expected maximum worker error, got %v", err)
	}
}

func TestWorkerManager_ScaleDown_BelowMinimum(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := NewMockMessageHandler()

	config := &WorkerManagerConfig{
		MinWorkers:                2,
		MaxWorkers:                5,
		DefaultWorkers:            2,
		WorkerConfig:              DefaultWorkerConfig("test"),
		HealthCheckInterval:       1 * time.Second,
		MetricsCollectionInterval: 500 * time.Millisecond,
		AutoScalingEnabled:        false,
		ScaleUpThreshold:          0.8,
		ScaleDownThreshold:        0.2,
		ScaleUpCooldown:           1 * time.Second,
		ScaleDownCooldown:         2 * time.Second,
		ShutdownTimeout:           5 * time.Second,
		EnableDetailedMetrics:     true,
	}

	manager := NewWorkerManager(mockUseCase, mockHandler, config)
	manager.Start()
	defer manager.Stop()

	// Wait for initial setup
	time.Sleep(50 * time.Millisecond)

	// Act - Try to scale below minimum
	err := manager.ScaleDown(3) // This should go below the min

	// Assert
	if err == nil {
		t.Fatal("Expected error when trying to scale below minimum")
	}

	if !contains(err.Error(), "minimum") {
		t.Errorf("Expected minimum worker error, got %v", err)
	}
}

// contains function is already defined in order_worker.go
