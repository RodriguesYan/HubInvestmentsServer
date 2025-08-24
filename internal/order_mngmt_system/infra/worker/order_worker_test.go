package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"HubInvestments/internal/order_mngmt_system/application/usecase"
	"HubInvestments/internal/order_mngmt_system/infra/messaging/rabbitmq"
	"HubInvestments/shared/infra/messaging"
)

// Mock implementations
type MockProcessOrderUseCase struct {
	mock.Mock
}

func (m *MockProcessOrderUseCase) Execute(ctx context.Context, command *usecase.ProcessOrderCommand) (*usecase.ProcessOrderResult, error) {
	args := m.Called(ctx, command)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.ProcessOrderResult), args.Error(1)
}

type MockOrderConsumer struct {
	mock.Mock
}

func (m *MockOrderConsumer) StartConsumers(handler rabbitmq.OrderMessageHandler) error {
	args := m.Called(handler)
	return args.Error(0)
}

func (m *MockOrderConsumer) StopConsumers() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOrderConsumer) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOrderConsumer) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOrderConsumer) GetQueueManager() *rabbitmq.OrderQueueManager {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*rabbitmq.OrderQueueManager)
}

func (m *MockOrderConsumer) GetActiveConsumers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

type MockMessageHandler struct {
	mock.Mock
}

func (m *MockMessageHandler) Publish(ctx context.Context, queueName string, message []byte) error {
	args := m.Called(ctx, queueName, message)
	return args.Error(0)
}

func (m *MockMessageHandler) PublishWithOptions(ctx context.Context, options messaging.PublishOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockMessageHandler) Consume(ctx context.Context, queue string, consumer messaging.MessageConsumer) error {
	args := m.Called(ctx, queue, consumer)
	return args.Error(0)
}

func (m *MockMessageHandler) DeclareQueue(queue string, options messaging.QueueOptions) error {
	args := m.Called(queue, options)
	return args.Error(0)
}

func (m *MockMessageHandler) DeleteQueue(queue string) error {
	args := m.Called(queue)
	return args.Error(0)
}

func (m *MockMessageHandler) PurgeQueue(queue string) error {
	args := m.Called(queue)
	return args.Error(0)
}

func (m *MockMessageHandler) QueueInfo(queue string) (*messaging.QueueInfo, error) {
	args := m.Called(queue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*messaging.QueueInfo), args.Error(1)
}

func (m *MockMessageHandler) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMessageHandler) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockQueueManager struct {
	mock.Mock
}

func (m *MockQueueManager) PublishToRetryQueue(message *rabbitmq.OrderMessage, delay time.Duration) error {
	args := m.Called(message, delay)
	return args.Error(0)
}

// Test helper functions
func createTestWorker(t *testing.T) (*OrderWorker, *MockProcessOrderUseCase, *MockOrderConsumer, *MockMessageHandler) {
	mockUseCase := &MockProcessOrderUseCase{}
	mockConsumer := &MockOrderConsumer{}
	mockMessageHandler := &MockMessageHandler{}

	config := &WorkerConfig{
		WorkerID:            "test-worker",
		MaxConcurrentOrders: 5,
		ProcessingTimeout:   10 * time.Second,
		HeartbeatInterval:   1 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    1 * time.Second,
		HealthCheckInterval: 2 * time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
		LogLevel:            "INFO",
	}

	// Create worker with nil consumer but provide mock message handler to avoid RabbitMQ connection
	worker := NewOrderWorker("test-worker", mockUseCase, nil, mockMessageHandler, config)

	return worker, mockUseCase, mockConsumer, mockMessageHandler
}

func createTestOrderMessage() *rabbitmq.OrderMessage {
	return &rabbitmq.OrderMessage{
		OrderID:   "test-order-123",
		UserID:    "user-456",
		Symbol:    "AAPL",
		OrderType: "MARKET",
		OrderSide: "BUY",
		Quantity:  100.0,
		Price:     nil,
		MessageMetadata: rabbitmq.OrderMessageMetadata{
			MessageID:    "msg-123",
			Timestamp:    time.Now(),
			RetryAttempt: 0,
			Priority:     1,
		},
	}
}

// Test cases
func TestNewOrderWorker(t *testing.T) {
	worker, _, _, _ := createTestWorker(t)

	assert.Equal(t, "test-worker", worker.id)
	assert.Equal(t, HealthStatusUnknown, worker.healthStatus)
	assert.False(t, worker.isRunning)
	assert.NotNil(t, worker.config)
	assert.NotNil(t, worker.metrics)
}

func TestDefaultWorkerConfig(t *testing.T) {
	config := DefaultWorkerConfig("test-worker")

	assert.Equal(t, "test-worker", config.WorkerID)
	assert.Equal(t, 10, config.MaxConcurrentOrders)
	assert.Equal(t, 30*time.Second, config.ProcessingTimeout)
	assert.Equal(t, 10*time.Second, config.HeartbeatInterval)
	assert.Equal(t, 3, config.MaxRetries)
	assert.True(t, config.EnableMetrics)
}

func TestWorkerStart(t *testing.T) {
	worker, _, mockConsumer, mockMessageHandler := createTestWorker(t)

	// Mock successful startup
	mockConsumer.On("StartConsumers", mock.AnythingOfType("*worker.OrderMessageHandler")).Return(nil)
	mockMessageHandler.On("HealthCheck").Return(nil)
	mockConsumer.On("IsRunning").Return(true)
	mockConsumer.On("HealthCheck").Return(nil)

	err := worker.Start()
	assert.NoError(t, err)
	assert.True(t, worker.IsRunning())
	assert.Equal(t, HealthStatusHealthy, worker.GetHealthStatus())

	// Clean up
	worker.Stop()
}

func TestWorkerStartAlreadyRunning(t *testing.T) {
	worker, _, mockConsumer, mockMessageHandler := createTestWorker(t)

	// Mock successful startup
	mockConsumer.On("StartConsumers", mock.AnythingOfType("*worker.OrderMessageHandler")).Return(nil)
	mockMessageHandler.On("HealthCheck").Return(nil)
	mockConsumer.On("IsRunning").Return(true)
	mockConsumer.On("HealthCheck").Return(nil)

	// Start worker first time
	err := worker.Start()
	assert.NoError(t, err)

	// Try to start again
	err = worker.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Clean up
	worker.Stop()
}

func TestWorkerStop(t *testing.T) {
	worker, _, mockConsumer, mockMessageHandler := createTestWorker(t)

	// Mock successful startup and shutdown
	mockConsumer.On("StartConsumers", mock.AnythingOfType("*worker.OrderMessageHandler")).Return(nil)
	mockConsumer.On("StopConsumers").Return(nil)
	mockMessageHandler.On("HealthCheck").Return(nil)
	mockConsumer.On("IsRunning").Return(true)
	mockConsumer.On("HealthCheck").Return(nil)

	// Start worker
	err := worker.Start()
	assert.NoError(t, err)

	// Stop worker
	err = worker.Stop()
	assert.NoError(t, err)
	assert.False(t, worker.IsRunning())
	assert.Equal(t, HealthStatusStopped, worker.GetHealthStatus())
}

func TestWorkerStopNotRunning(t *testing.T) {
	worker, _, _, _ := createTestWorker(t)

	err := worker.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestWorkerProcessOrderMessage(t *testing.T) {
	worker, mockUseCase, _, _ := createTestWorker(t)

	// Create test message and expected result
	message := createTestOrderMessage()
	expectedResult := &usecase.ProcessOrderResult{
		OrderID:        message.OrderID,
		FinalStatus:    "EXECUTED",
		ExecutionPrice: func() *float64 { p := 150.0; return &p }(),
		ProcessingTime: 100 * time.Millisecond,
		WorkerID:       worker.id,
	}

	// Mock successful processing
	mockUseCase.On("Execute", mock.Anything, mock.AnythingOfType("*usecase.ProcessOrderCommand")).Return(expectedResult, nil)

	// Process message
	err := worker.processOrderMessage(context.Background(), message)
	assert.NoError(t, err)

	// Verify metrics were updated
	metrics := worker.GetMetrics()
	assert.Equal(t, int64(1), metrics.OrdersProcessed)

	mockUseCase.AssertExpectations(t)
}

func TestWorkerProcessOrderMessageFailure(t *testing.T) {
	worker, mockUseCase, mockConsumer, _ := createTestWorker(t)

	// Create test message
	message := createTestOrderMessage()
	mockQueueManager := &MockQueueManager{}

	// Mock processing failure
	mockUseCase.On("Execute", mock.Anything, mock.AnythingOfType("*usecase.ProcessOrderCommand")).Return(nil, errors.New("processing failed"))
	mockConsumer.On("GetQueueManager").Return(mockQueueManager)
	mockQueueManager.On("PublishToRetryQueue", message, mock.AnythingOfType("time.Duration")).Return(nil)

	// Process message
	err := worker.processOrderMessage(context.Background(), message)
	assert.Error(t, err)

	// Verify metrics were updated
	metrics := worker.GetMetrics()
	assert.Equal(t, int64(1), metrics.OrdersProcessed)
	assert.Equal(t, int64(1), metrics.OrdersFailed)

	mockUseCase.AssertExpectations(t)
}

func TestWorkerShouldRetryOrder(t *testing.T) {
	worker, _, _, _ := createTestWorker(t)

	message := createTestOrderMessage()

	// Test retryable error
	retryableErr := errors.New("connection refused")
	assert.True(t, worker.shouldRetryOrder(message, retryableErr))

	// Test non-retryable error
	nonRetryableErr := errors.New("validation failed")
	assert.False(t, worker.shouldRetryOrder(message, nonRetryableErr))

	// Test max retries exceeded
	message.MessageMetadata.RetryAttempt = 5
	assert.False(t, worker.shouldRetryOrder(message, retryableErr))
}

func TestWorkerIsRetryableError(t *testing.T) {
	worker, _, _, _ := createTestWorker(t)

	// Test retryable errors
	retryableErrors := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"service unavailable",
		"market data unavailable",
	}

	for _, errMsg := range retryableErrors {
		err := errors.New(errMsg)
		assert.True(t, worker.isRetryableError(err), "Error should be retryable: %s", errMsg)
	}

	// Test non-retryable errors
	nonRetryableErrors := []string{
		"validation failed",
		"invalid order",
		"insufficient funds",
		"order not found",
	}

	for _, errMsg := range nonRetryableErrors {
		err := errors.New(errMsg)
		assert.False(t, worker.isRetryableError(err), "Error should not be retryable: %s", errMsg)
	}
}

func TestWorkerCalculateRetryDelay(t *testing.T) {
	worker, _, _, _ := createTestWorker(t)

	// Test exponential backoff
	delay0 := worker.calculateRetryDelay(0)
	delay1 := worker.calculateRetryDelay(1)
	delay2 := worker.calculateRetryDelay(2)

	assert.Equal(t, 1*time.Second, delay0)
	assert.Equal(t, 2*time.Second, delay1)
	assert.Equal(t, 4*time.Second, delay2)

	// Test maximum delay cap (2^10 = 1024 seconds = ~17 minutes)
	delay10 := worker.calculateRetryDelay(10)
	expectedDelay10 := 1 * time.Second * time.Duration(1<<uint(10)) // 1024 seconds
	maxDelay := 1 * time.Hour
	if expectedDelay10 > maxDelay {
		expectedDelay10 = maxDelay
	}
	assert.Equal(t, expectedDelay10, delay10)
}

func TestWorkerGetWorkerInfo(t *testing.T) {
	worker, _, _, _ := createTestWorker(t)

	info := worker.GetWorkerInfo()

	assert.Equal(t, "test-worker", info.ID)
	assert.False(t, info.IsRunning)
	assert.Equal(t, "unknown", info.HealthStatus)
	assert.Equal(t, int64(0), info.ProcessedCount)
	assert.Equal(t, int64(0), info.ErrorCount)
	assert.Equal(t, int64(0), info.RetryCount)
}

func TestWorkerMetricsUpdate(t *testing.T) {
	worker, _, _, _ := createTestWorker(t)

	// Test metric updates
	worker.incrementProcessedCount()
	worker.incrementErrorCount()
	worker.incrementRetryCount()
	worker.updateProcessingTime(100 * time.Millisecond)

	metrics := worker.GetMetrics()
	assert.Equal(t, int64(1), metrics.OrdersProcessed)
	assert.Equal(t, int64(1), metrics.OrdersFailed)
	assert.Equal(t, int64(1), metrics.OrdersRetried)
	assert.Equal(t, 100*time.Millisecond, metrics.LastProcessingTime)
	assert.Equal(t, 100*time.Millisecond, metrics.AverageProcessingTime)

	// Test average processing time calculation
	worker.incrementProcessedCount()
	worker.updateProcessingTime(200 * time.Millisecond)

	metrics = worker.GetMetrics()
	assert.Equal(t, int64(2), metrics.OrdersProcessed)
	assert.Equal(t, 150*time.Millisecond, metrics.AverageProcessingTime) // (100 + 200) / 2
}

func TestWorkerHealthStatusTransitions(t *testing.T) {
	worker, _, _, _ := createTestWorker(t)

	// Test initial status
	assert.Equal(t, HealthStatusUnknown, worker.GetHealthStatus())

	// Test status updates
	worker.updateHealthStatus(HealthStatusHealthy)
	assert.Equal(t, HealthStatusHealthy, worker.GetHealthStatus())

	worker.updateHealthStatus(HealthStatusDegraded)
	assert.Equal(t, HealthStatusDegraded, worker.GetHealthStatus())

	worker.updateHealthStatus(HealthStatusUnhealthy)
	assert.Equal(t, HealthStatusUnhealthy, worker.GetHealthStatus())

	worker.updateHealthStatus(HealthStatusStopped)
	assert.Equal(t, HealthStatusStopped, worker.GetHealthStatus())
}

func TestHealthStatusString(t *testing.T) {
	assert.Equal(t, "healthy", HealthStatusHealthy.String())
	assert.Equal(t, "degraded", HealthStatusDegraded.String())
	assert.Equal(t, "unhealthy", HealthStatusUnhealthy.String())
	assert.Equal(t, "stopped", HealthStatusStopped.String())
	assert.Equal(t, "unknown", HealthStatusUnknown.String())
}

func TestOrderMessageHandlerMethods(t *testing.T) {
	worker, mockUseCase, _, _ := createTestWorker(t)
	handler := &OrderMessageHandler{
		worker:    worker,
		semaphore: make(chan struct{}, 5),
	}

	message := createTestOrderMessage()
	ctx := context.Background()

	// Mock successful processing
	expectedResult := &usecase.ProcessOrderResult{
		OrderID:     message.OrderID,
		FinalStatus: "EXECUTED",
		WorkerID:    worker.id,
	}
	mockUseCase.On("Execute", mock.Anything, mock.AnythingOfType("*usecase.ProcessOrderCommand")).Return(expectedResult, nil)

	// Test HandleOrderMessage
	err := handler.HandleOrderMessage(ctx, message)
	assert.NoError(t, err)

	// Test HandleStatusUpdate
	statusUpdate := &rabbitmq.OrderStatusUpdate{
		OrderID:        message.OrderID,
		CurrentStatus:  "EXECUTED",
		PreviousStatus: "PROCESSING",
	}
	err = handler.HandleStatusUpdate(ctx, statusUpdate)
	assert.NoError(t, err)

	mockUseCase.AssertExpectations(t)
}

func TestWorkerConcurrentProcessing(t *testing.T) {
	worker, mockUseCase, _, _ := createTestWorker(t)

	// Mock processing that takes some time
	mockUseCase.On("Execute", mock.Anything, mock.AnythingOfType("*usecase.ProcessOrderCommand")).Return(
		&usecase.ProcessOrderResult{
			OrderID:     "test-order",
			FinalStatus: "EXECUTED",
			WorkerID:    worker.id,
		},
		nil,
	).Maybe() // Allow multiple calls without strict matching

	// Process multiple messages concurrently
	var wg sync.WaitGroup
	messageCount := 10

	for i := 0; i < messageCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			message := createTestOrderMessage()
			message.OrderID = fmt.Sprintf("order-%d", id)
			err := worker.processOrderMessage(context.Background(), message)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all messages were processed
	metrics := worker.GetMetrics()
	assert.Equal(t, int64(messageCount), metrics.OrdersProcessed)

	mockUseCase.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkWorkerProcessOrderMessage(b *testing.B) {
	worker, mockUseCase, _, _ := createTestWorker(&testing.T{})

	message := createTestOrderMessage()
	expectedResult := &usecase.ProcessOrderResult{
		OrderID:     message.OrderID,
		FinalStatus: "EXECUTED",
		WorkerID:    worker.id,
	}

	mockUseCase.On("Execute", mock.Anything, mock.AnythingOfType("*usecase.ProcessOrderCommand")).Return(expectedResult, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker.processOrderMessage(context.Background(), message)
	}
}

func BenchmarkWorkerMetricsUpdate(b *testing.B) {
	worker, _, _, _ := createTestWorker(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker.incrementProcessedCount()
		worker.updateProcessingTime(100 * time.Millisecond)
	}
}
