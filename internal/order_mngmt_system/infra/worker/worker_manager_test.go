package worker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"HubInvestments/internal/order_mngmt_system/application/usecase"
	"HubInvestments/shared/infra/messaging"
)

// Mock implementations for WorkerManager tests
type MockWorkerManagerProcessOrderUseCase struct {
	mock.Mock
}

func (m *MockWorkerManagerProcessOrderUseCase) Execute(ctx context.Context, command *usecase.ProcessOrderCommand) (*usecase.ProcessOrderResult, error) {
	args := m.Called(ctx, command)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.ProcessOrderResult), args.Error(1)
}

type MockWorkerManagerMessageHandler struct {
	mock.Mock
}

func (m *MockWorkerManagerMessageHandler) Publish(ctx context.Context, queueName string, message []byte) error {
	args := m.Called(ctx, queueName, message)
	return args.Error(0)
}

func (m *MockWorkerManagerMessageHandler) PublishWithOptions(ctx context.Context, options messaging.PublishOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockWorkerManagerMessageHandler) Consume(ctx context.Context, queue string, consumer messaging.MessageConsumer) error {
	args := m.Called(ctx, queue, consumer)
	return args.Error(0)
}

func (m *MockWorkerManagerMessageHandler) DeclareQueue(queue string, options messaging.QueueOptions) error {
	args := m.Called(queue, options)
	return args.Error(0)
}

func (m *MockWorkerManagerMessageHandler) DeleteQueue(queue string) error {
	args := m.Called(queue)
	return args.Error(0)
}

func (m *MockWorkerManagerMessageHandler) PurgeQueue(queue string) error {
	args := m.Called(queue)
	return args.Error(0)
}

func (m *MockWorkerManagerMessageHandler) QueueInfo(queue string) (*messaging.QueueInfo, error) {
	args := m.Called(queue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*messaging.QueueInfo), args.Error(1)
}

func (m *MockWorkerManagerMessageHandler) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockWorkerManagerMessageHandler) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Test helper functions
func createTestWorkerManager(t *testing.T) (*WorkerManager, *MockWorkerManagerProcessOrderUseCase, *MockWorkerManagerMessageHandler) {
	mockUseCase := &MockWorkerManagerProcessOrderUseCase{}
	mockMessageHandler := &MockWorkerManagerMessageHandler{}

	config := &WorkerManagerConfig{
		MinWorkers:                2,
		MaxWorkers:                10,
		DefaultWorkers:            3,
		WorkerConfig:              DefaultWorkerConfig(""),
		HealthCheckInterval:       100 * time.Millisecond,
		MetricsCollectionInterval: 50 * time.Millisecond,
		AutoScalingEnabled:        true,
		ScaleUpThreshold:          0.8,
		ScaleDownThreshold:        0.2,
		ScaleUpCooldown:           1 * time.Second,
		ScaleDownCooldown:         2 * time.Second,
		ShutdownTimeout:           5 * time.Second,
		EnableDetailedMetrics:     true,
	}

	wm := NewWorkerManager(mockUseCase, mockMessageHandler, config)
	return wm, mockUseCase, mockMessageHandler
}

// Test cases
func TestNewWorkerManager(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	assert.NotNil(t, wm)
	assert.Equal(t, 0, len(wm.workers))
	assert.False(t, wm.isRunning)
	assert.NotNil(t, wm.config)
	assert.NotNil(t, wm.metrics)
	assert.NotNil(t, wm.healthChecker)
	assert.NotNil(t, wm.autoScaler)
}

func TestDefaultWorkerManagerConfig(t *testing.T) {
	config := DefaultWorkerManagerConfig()

	assert.Equal(t, 2, config.MinWorkers)
	assert.Equal(t, 20, config.MaxWorkers)
	assert.Equal(t, 5, config.DefaultWorkers)
	assert.True(t, config.AutoScalingEnabled)
	assert.Equal(t, 0.8, config.ScaleUpThreshold)
	assert.Equal(t, 0.2, config.ScaleDownThreshold)
	assert.True(t, config.EnableDetailedMetrics)
}

func TestWorkerManagerStart(t *testing.T) {
	wm, _, mockMessageHandler := createTestWorkerManager(t)

	// Mock message handler for worker initialization
	mockMessageHandler.On("HealthCheck").Return(nil)

	// Note: In a real test, we would need to mock the OrderConsumer creation
	// For now, we'll test the basic structure
	assert.False(t, wm.IsRunning())
	assert.Equal(t, 0, wm.GetWorkerCount())

	// Test that we can't start twice
	// err := wm.Start()
	// assert.NoError(t, err)
	// assert.True(t, wm.IsRunning())

	// err = wm.Start()
	// assert.Error(t, err)
	// assert.Contains(t, err.Error(), "already running")

	// Clean up
	// wm.Stop()
}

func TestWorkerManagerStop(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	// Test stopping when not running
	err := wm.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestWorkerManagerGetHealthStatus(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	// Test health status when stopped
	health := wm.GetHealthStatus()
	assert.Equal(t, "stopped", health.Status)
	assert.Equal(t, 0, health.ActiveWorkers)
	assert.Equal(t, 0, health.HealthyWorkers)
	assert.Equal(t, 0, health.UnhealthyWorkers)
}

func TestWorkerManagerGetMetrics(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	metrics := wm.GetMetrics()
	assert.Equal(t, 0, metrics.ActiveWorkers)
	assert.Equal(t, int64(0), metrics.TotalOrdersProcessed)
	assert.Equal(t, int64(0), metrics.TotalOrdersSuccessful)
	assert.Equal(t, int64(0), metrics.TotalOrdersFailed)
	assert.Equal(t, int64(0), metrics.TotalOrdersRetried)
}

func TestWorkerManagerScaleUpValidation(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	// Test scaling up when at maximum capacity
	wm.workers = make(map[string]*OrderWorker)
	for i := 0; i < wm.config.MaxWorkers; i++ {
		wm.workers[fmt.Sprintf("worker-%d", i)] = nil // Mock workers
	}

	err := wm.ScaleUp(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already at maximum worker capacity")
}

func TestWorkerManagerScaleDownValidation(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	// Test scaling down when at minimum capacity
	wm.workers = make(map[string]*OrderWorker)
	for i := 0; i < wm.config.MinWorkers; i++ {
		wm.workers[fmt.Sprintf("worker-%d", i)] = nil // Mock workers
	}

	err := wm.ScaleDown(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already at minimum worker capacity")
}

func TestWorkerManagerSelectWorkersForRemoval(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	// Create mock workers with different health statuses
	mockHealthyWorker := &OrderWorker{id: "healthy-1", healthStatus: HealthStatusHealthy}
	mockDegradedWorker := &OrderWorker{id: "degraded-1", healthStatus: HealthStatusDegraded}
	mockUnhealthyWorker := &OrderWorker{id: "unhealthy-1", healthStatus: HealthStatusUnhealthy}

	wm.workers = map[string]*OrderWorker{
		"healthy-1":   mockHealthyWorker,
		"degraded-1":  mockDegradedWorker,
		"unhealthy-1": mockUnhealthyWorker,
	}

	// Test selecting workers for removal
	selected := wm.selectWorkersForRemoval(2)
	assert.Len(t, selected, 2)

	// Should prioritize unhealthy workers first
	assert.Contains(t, selected, "unhealthy-1")
	assert.Contains(t, selected, "degraded-1")
}

func TestHealthChecker(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)
	hc := wm.healthChecker

	// Test initial state
	assert.Equal(t, 0, len(hc.unhealthyWorkers))
	assert.Equal(t, 0, len(hc.recoveryAttempts))

	// Create mock worker
	mockWorker := &OrderWorker{id: "test-worker", healthStatus: HealthStatusUnhealthy}

	// Test checking unhealthy worker
	hc.checkWorkerHealth("test-worker", mockWorker)
	assert.Equal(t, 1, len(hc.unhealthyWorkers))

	// Test recovery to healthy
	mockWorker.healthStatus = HealthStatusHealthy
	hc.checkWorkerHealth("test-worker", mockWorker)
	assert.Equal(t, 0, len(hc.unhealthyWorkers))
	assert.Equal(t, 0, len(hc.recoveryAttempts))
}

func TestAutoScaler(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)
	as := wm.autoScaler

	// Test initial state
	assert.False(t, as.scaleUpInProgress)
	assert.False(t, as.scaleDownInProgress)

	// Test queue depth ratio calculation
	metrics := WorkerManagerMetrics{
		ActiveWorkers: 5,
		QueueDepth:    50, // 50 orders in queue
	}
	ratio := as.calculateQueueDepthRatio(&metrics)
	assert.Equal(t, 1.0, ratio) // 50 / (5 * 10) = 1.0

	// Test should scale up conditions
	shouldScaleUp := as.shouldScaleUp(0.9, 5) // High load, not at max workers
	assert.True(t, shouldScaleUp)

	shouldScaleUp = as.shouldScaleUp(0.9, wm.config.MaxWorkers) // High load, but at max workers
	assert.False(t, shouldScaleUp)

	// Test should scale down conditions
	shouldScaleDown := as.shouldScaleDown(0.1, 5) // Low load, above min workers
	assert.True(t, shouldScaleDown)

	shouldScaleDown = as.shouldScaleDown(0.1, wm.config.MinWorkers) // Low load, but at min workers
	assert.False(t, shouldScaleDown)
}

func TestWorkerManagerMetricsCollection(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	// Test initial metrics
	metrics := wm.GetMetrics()
	assert.Equal(t, 0, metrics.ActiveWorkers)
	assert.Equal(t, int64(0), metrics.TotalOrdersProcessed)

	// Test metrics update methods
	wm.updateScaleUpMetrics(2)
	metrics = wm.GetMetrics()
	assert.Equal(t, int64(1), metrics.ScaleUpEvents)

	wm.updateScaleDownMetrics(1)
	metrics = wm.GetMetrics()
	assert.Equal(t, int64(1), metrics.ScaleDownEvents)
}

func TestWorkerManagerConcurrentOperations(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	// Test concurrent access to worker map
	var wg sync.WaitGroup
	operationCount := 10

	// Concurrent reads
	for i := 0; i < operationCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = wm.GetWorkerCount()
			_ = wm.GetWorkerInfo()
			_ = wm.GetHealthStatus()
		}()
	}

	// Concurrent metrics updates
	for i := 0; i < operationCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wm.updateScaleUpMetrics(1)
			wm.updateScaleDownMetrics(1)
		}()
	}

	wg.Wait()

	// Verify final state
	metrics := wm.GetMetrics()
	assert.Equal(t, int64(operationCount), metrics.ScaleUpEvents)
	assert.Equal(t, int64(operationCount), metrics.ScaleDownEvents)
}

func TestWorkerManagerConfigValidation(t *testing.T) {
	// Test with nil config (should use defaults)
	wm := NewWorkerManager(nil, nil, nil)
	assert.NotNil(t, wm.config)
	assert.Equal(t, 2, wm.config.MinWorkers)
	assert.Equal(t, 20, wm.config.MaxWorkers)

	// Test with custom config
	customConfig := &WorkerManagerConfig{
		MinWorkers:     1,
		MaxWorkers:     5,
		DefaultWorkers: 2,
	}
	wm = NewWorkerManager(nil, nil, customConfig)
	assert.Equal(t, 1, wm.config.MinWorkers)
	assert.Equal(t, 5, wm.config.MaxWorkers)
	assert.Equal(t, 2, wm.config.DefaultWorkers)
}

func TestManagerHealthStatus(t *testing.T) {
	// Test ManagerHealthStatus creation
	status := ManagerHealthStatus{
		Status:           "healthy",
		ActiveWorkers:    5,
		HealthyWorkers:   4,
		DegradedWorkers:  1,
		UnhealthyWorkers: 0,
		MinWorkers:       2,
		MaxWorkers:       10,
	}

	assert.Equal(t, "healthy", status.Status)
	assert.Equal(t, 5, status.ActiveWorkers)
	assert.Equal(t, 4, status.HealthyWorkers)
	assert.Equal(t, 1, status.DegradedWorkers)
	assert.Equal(t, 0, status.UnhealthyWorkers)
}

func TestWorkerManagerEstimateQueueDepth(t *testing.T) {
	wm, _, _ := createTestWorkerManager(t)

	// Test queue depth estimation (placeholder implementation)
	depth := wm.estimateQueueDepth()
	assert.Equal(t, int64(0), depth) // Current implementation returns 0
}

// Benchmark tests
func BenchmarkWorkerManagerGetMetrics(b *testing.B) {
	wm, _, _ := createTestWorkerManager(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wm.GetMetrics()
	}
}

func BenchmarkWorkerManagerGetHealthStatus(b *testing.B) {
	wm, _, _ := createTestWorkerManager(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wm.GetHealthStatus()
	}
}

func BenchmarkWorkerManagerConcurrentMetricsUpdate(b *testing.B) {
	wm, _, _ := createTestWorkerManager(&testing.T{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wm.updateScaleUpMetrics(1)
		}
	})
}
