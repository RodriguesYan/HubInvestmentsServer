package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	orderUsecase "HubInvestments/internal/order_mngmt_system/application/usecase"
	"HubInvestments/internal/order_mngmt_system/infra/messaging/rabbitmq"
	"HubInvestments/shared/infra/messaging"
)

func TestOrderWorker_Start_Success(t *testing.T) {
	// Arrange
	mockUseCase := NewMockProcessOrderUseCase()
	mockHandler := NewMockMessageHandler()

	config := &WorkerConfig{
		WorkerID:            "test-worker-1",
		MaxConcurrentOrders: 2,
		ProcessingTimeout:   30 * time.Second,
		HeartbeatInterval:   1 * time.Second,
		HealthCheckInterval: 1 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	worker := NewOrderWorker("test-worker-1", mockUseCase, nil, mockHandler, config)

	// Act
	err := worker.Start()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error starting worker, got %v", err)
	}

	// Verify worker is running
	if !worker.IsRunning() {
		t.Error("Expected worker to be running")
	}

	// Cleanup
	worker.Stop()
}

func TestOrderWorker_ProcessMessage_Success(t *testing.T) {
	// Arrange
	processedOrderID := ""
	mockUseCase := &MockProcessOrderUseCase{}
	mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(
		func(ctx context.Context, cmd *orderUsecase.ProcessOrderCommand) *orderUsecase.ProcessOrderResult {
			processedOrderID = cmd.OrderID
			return createSuccessfulProcessOrderResult(cmd.OrderID)
		},
		nil,
	)
	mockHandler := NewMockMessageHandler()

	config := &WorkerConfig{
		WorkerID:            "test-worker-1",
		MaxConcurrentOrders: 2,
		ProcessingTimeout:   30 * time.Second,
		HeartbeatInterval:   1 * time.Second,
		HealthCheckInterval: 1 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	worker := NewOrderWorker("test-worker-1", mockUseCase, nil, mockHandler, config)
	worker.Start()
	defer worker.Stop()

	// Create test message using the correct OrderMessage structure
	orderMessage := rabbitmq.OrderMessage{
		OrderID:   "order123",
		UserID:    "user123",
		Symbol:    "AAPL",
		Status:    "PENDING",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		MessageMetadata: rabbitmq.OrderMessageMetadata{
			MessageID:    "msg123",
			Timestamp:    time.Now(),
			RetryAttempt: 1,
			MessageType:  "ORDER_PROCESSING",
		},
	}

	messageBody, _ := json.Marshal(orderMessage)
	message := &messaging.Message{
		Body:      messageBody,
		MessageID: "msg123",
		Timestamp: time.Now().Unix(),
	}

	// Act
	mockHandler.SimulateMessage("orders.processing", message)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Assert
	if processedOrderID != "order123" {
		t.Errorf("Expected processed order ID 'order123', got '%s'", processedOrderID)
	}

	// Check metrics
	metrics := worker.GetMetrics()
	if metrics.OrdersProcessed != 1 {
		t.Errorf("Expected OrdersProcessed 1, got %d", metrics.OrdersProcessed)
	}

	if metrics.OrdersSuccessful != 1 {
		t.Errorf("Expected OrdersSuccessful 1, got %d", metrics.OrdersSuccessful)
	}
}

func TestOrderWorker_ProcessMessage_UseCaseError(t *testing.T) {
	// Arrange
	mockUseCase := &MockProcessOrderUseCase{}
	mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(
		(*orderUsecase.ProcessOrderResult)(nil),
		errors.New("order processing failed"),
	)
	mockHandler := NewMockMessageHandler()

	config := &WorkerConfig{
		WorkerID:            "test-worker-1",
		MaxConcurrentOrders: 2,
		ProcessingTimeout:   30 * time.Second,
		HeartbeatInterval:   1 * time.Second,
		HealthCheckInterval: 1 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	worker := NewOrderWorker("test-worker-1", mockUseCase, nil, mockHandler, config)
	worker.Start()
	defer worker.Stop()

	// Create test message
	orderMessage := rabbitmq.OrderMessage{
		OrderID:   "order123",
		UserID:    "user123",
		Symbol:    "AAPL",
		Status:    "PENDING",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		MessageMetadata: rabbitmq.OrderMessageMetadata{
			MessageID:    "msg123",
			Timestamp:    time.Now(),
			RetryAttempt: 1,
			MessageType:  "ORDER_PROCESSING",
		},
	}

	messageBody, _ := json.Marshal(orderMessage)
	message := &messaging.Message{
		Body:      messageBody,
		MessageID: "msg123",
		Timestamp: time.Now().Unix(),
	}

	// Act
	mockHandler.SimulateMessage("orders.processing", message)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Assert
	metrics := worker.GetMetrics()
	if metrics.OrdersProcessed != 1 {
		t.Errorf("Expected OrdersProcessed 1, got %d", metrics.OrdersProcessed)
	}

	if metrics.OrdersFailed != 1 {
		t.Errorf("Expected OrdersFailed 1, got %d", metrics.OrdersFailed)
	}

	if metrics.OrdersSuccessful != 0 {
		t.Errorf("Expected OrdersSuccessful 0, got %d", metrics.OrdersSuccessful)
	}
}

func TestOrderWorker_ProcessMessage_InvalidJSON(t *testing.T) {
	// Arrange
	mockUseCase := &MockProcessOrderUseCase{}
	mockHandler := NewMockMessageHandler()

	config := &WorkerConfig{
		WorkerID:            "test-worker-1",
		MaxConcurrentOrders: 2,
		ProcessingTimeout:   30 * time.Second,
		HeartbeatInterval:   1 * time.Second,
		HealthCheckInterval: 1 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	worker := NewOrderWorker("test-worker-1", mockUseCase, nil, mockHandler, config)
	worker.Start()
	defer worker.Stop()

	// Create invalid message
	message := &messaging.Message{
		Body:      []byte("invalid json"),
		MessageID: "msg123",
		Timestamp: time.Now().Unix(),
	}

	// Act
	mockHandler.SimulateMessage("orders.processing", message)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Assert
	metrics := worker.GetMetrics()
	if metrics.OrdersProcessed != 1 {
		t.Errorf("Expected OrdersProcessed 1, got %d", metrics.OrdersProcessed)
	}

	if metrics.OrdersFailed != 1 {
		t.Errorf("Expected OrdersFailed 1, got %d", metrics.OrdersFailed)
	}
}

func TestOrderWorker_Stop_Success(t *testing.T) {
	// Arrange
	mockUseCase := &MockProcessOrderUseCase{}
	mockHandler := NewMockMessageHandler()

	config := &WorkerConfig{
		WorkerID:            "test-worker-1",
		MaxConcurrentOrders: 2,
		ProcessingTimeout:   30 * time.Second,
		HeartbeatInterval:   1 * time.Second,
		HealthCheckInterval: 1 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	worker := NewOrderWorker("test-worker-1", mockUseCase, nil, mockHandler, config)
	worker.Start()

	// Verify worker is running
	if !worker.IsRunning() {
		t.Error("Expected worker to be running")
	}

	// Act
	err := worker.Stop()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error stopping worker, got %v", err)
	}

	// Verify worker is stopped
	if worker.IsRunning() {
		t.Error("Expected worker to be stopped")
	}
}

func TestOrderWorker_GetHealthStatus_Success(t *testing.T) {
	// Arrange
	mockUseCase := &MockProcessOrderUseCase{}
	mockHandler := NewMockMessageHandler()

	config := &WorkerConfig{
		WorkerID:            "test-worker-1",
		MaxConcurrentOrders: 2,
		ProcessingTimeout:   30 * time.Second,
		HeartbeatInterval:   1 * time.Second,
		HealthCheckInterval: 1 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	worker := NewOrderWorker("test-worker-1", mockUseCase, nil, mockHandler, config)
	worker.Start()
	defer worker.Stop()

	// Act
	health := worker.GetHealthStatus()

	// Assert
	if health != HealthStatusHealthy {
		t.Errorf("Expected worker to be healthy, got %v", health)
	}
}

func TestOrderWorker_GetHealthStatus_MessageHandlerError(t *testing.T) {
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
	mockHandler.On("HealthCheck", mock.Anything).Return(errors.New("message handler unhealthy"))
	mockHandler.On("Close").Return(nil)

	config := &WorkerConfig{
		WorkerID:            "test-worker-1",
		MaxConcurrentOrders: 2,
		ProcessingTimeout:   30 * time.Second,
		HeartbeatInterval:   1 * time.Second,
		HealthCheckInterval: 1 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	worker := NewOrderWorker("test-worker-1", mockUseCase, nil, mockHandler, config)
	worker.Start()
	defer worker.Stop()

	// Wait for health check to run
	time.Sleep(200 * time.Millisecond)

	// Act
	health := worker.GetHealthStatus()

	// Assert - The worker should detect the unhealthy message handler
	if health == HealthStatusHealthy {
		t.Error("Expected worker to be unhealthy due to message handler error")
	}
}

func TestOrderWorker_ConcurrentProcessing(t *testing.T) {
	// Arrange
	processedOrders := make(map[string]bool)
	processedOrdersMutex := make(chan struct{}, 1)
	processedOrdersMutex <- struct{}{} // Initialize mutex

	mockUseCase := &MockProcessOrderUseCase{}
	mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(
		func(ctx context.Context, cmd *orderUsecase.ProcessOrderCommand) *orderUsecase.ProcessOrderResult {
			// Simulate processing time
			time.Sleep(50 * time.Millisecond)

			<-processedOrdersMutex
			processedOrders[cmd.OrderID] = true
			processedOrdersMutex <- struct{}{}

			return createSuccessfulProcessOrderResult(cmd.OrderID)
		},
		nil,
	)
	mockHandler := NewMockMessageHandler()

	config := &WorkerConfig{
		WorkerID:            "test-worker-1",
		MaxConcurrentOrders: 3, // Allow 3 concurrent processes
		ProcessingTimeout:   30 * time.Second,
		HeartbeatInterval:   1 * time.Second,
		HealthCheckInterval: 1 * time.Second,
		MaxRetries:          3,
		RetryBackoffBase:    time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	worker := NewOrderWorker("test-worker-1", mockUseCase, nil, mockHandler, config)
	worker.Start()
	defer worker.Stop()

	// Act - Send multiple messages concurrently
	orderIDs := []string{"order1", "order2", "order3", "order4", "order5"}
	for _, orderID := range orderIDs {
		orderMessage := rabbitmq.OrderMessage{
			OrderID:   orderID,
			UserID:    "user123",
			Symbol:    "AAPL",
			Status:    "PENDING",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			MessageMetadata: rabbitmq.OrderMessageMetadata{
				MessageID:    "msg" + orderID,
				Timestamp:    time.Now(),
				RetryAttempt: 1,
				MessageType:  "ORDER_PROCESSING",
			},
		}

		messageBody, _ := json.Marshal(orderMessage)
		message := &messaging.Message{
			Body:      messageBody,
			MessageID: "msg" + orderID,
			Timestamp: time.Now().Unix(),
		}

		mockHandler.SimulateMessage("orders.processing", message)
	}

	// Wait for all processing to complete
	time.Sleep(500 * time.Millisecond)

	// Assert
	<-processedOrdersMutex
	processedCount := len(processedOrders)
	processedOrdersMutex <- struct{}{}

	if processedCount != len(orderIDs) {
		t.Errorf("Expected %d processed orders, got %d", len(orderIDs), processedCount)
	}

	// Verify all orders were processed
	for _, orderID := range orderIDs {
		<-processedOrdersMutex
		if !processedOrders[orderID] {
			t.Errorf("Expected order %s to be processed", orderID)
		}
		processedOrdersMutex <- struct{}{}
	}

	// Check metrics
	metrics := worker.GetMetrics()
	if metrics.OrdersProcessed != int64(len(orderIDs)) {
		t.Errorf("Expected OrdersProcessed %d, got %d", len(orderIDs), metrics.OrdersProcessed)
	}
}
