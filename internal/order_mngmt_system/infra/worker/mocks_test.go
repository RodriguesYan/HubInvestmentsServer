package worker

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"HubInvestments/internal/order_mngmt_system/application/usecase"
	"HubInvestments/internal/order_mngmt_system/infra/messaging/rabbitmq"
	"HubInvestments/shared/infra/messaging"
)

// MockProcessOrderUseCase implements IProcessOrderUseCase for testing
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

// MockOrderConsumer implements OrderConsumer for testing
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

// MockMessageHandler implements MessageHandler for testing
type MockMessageHandler struct {
	mock.Mock
	consumers map[string]messaging.MessageConsumer
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
	// Store the consumer for simulation
	if m.consumers == nil {
		m.consumers = make(map[string]messaging.MessageConsumer)
	}
	m.consumers[queue] = consumer
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

// MockMessageConsumer implements MessageConsumer for testing
type MockMessageConsumer struct {
	mock.Mock
}

func (m *MockMessageConsumer) HandleMessage(ctx context.Context, message *messaging.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// SimulateMessage sends a message to the mock handler for testing
func (m *MockMessageHandler) SimulateMessage(queueName string, message *messaging.Message) {
	if consumer, exists := m.consumers[queueName]; exists {
		go consumer.HandleMessage(context.Background(), message)
	}
}

// Helper function to create a default successful ProcessOrderResult
func createSuccessfulProcessOrderResult(orderID string) *usecase.ProcessOrderResult {
	return &usecase.ProcessOrderResult{
		OrderID:        orderID,
		FinalStatus:    "EXECUTED",
		ExecutionPrice: func() *float64 { p := 150.00; return &p }(),
		ExecutionTime:  func() *time.Time { t := time.Now(); return &t }(),
		ProcessingTime: 100 * time.Millisecond,
		WorkerID:       "test-worker",
		ProcessingID:   "test-proc-123",
	}
}

// Helper function to create a default QueueInfo
func createDefaultQueueInfo(queueName string) *messaging.QueueInfo {
	return &messaging.QueueInfo{
		Name:      queueName,
		Messages:  0,
		Consumers: 1,
		Durable:   true,
	}
}

// NewMockMessageHandler creates a new MockMessageHandler with default behavior
func NewMockMessageHandler() *MockMessageHandler {
	handler := &MockMessageHandler{
		consumers: make(map[string]messaging.MessageConsumer),
	}

	// Set up default behaviors
	handler.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler.On("PublishWithOptions", mock.Anything, mock.Anything).Return(nil)
	handler.On("Consume", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler.On("DeclareQueue", mock.Anything, mock.Anything).Return(nil)
	handler.On("DeleteQueue", mock.Anything).Return(nil)
	handler.On("PurgeQueue", mock.Anything).Return(nil)
	handler.On("QueueInfo", mock.Anything).Return(createDefaultQueueInfo("test-queue"), nil)
	handler.On("HealthCheck", mock.Anything).Return(nil)
	handler.On("Close").Return(nil)

	return handler
}

// NewMockProcessOrderUseCase creates a new MockProcessOrderUseCase with default behavior
func NewMockProcessOrderUseCase() *MockProcessOrderUseCase {
	useCase := &MockProcessOrderUseCase{}

	// Set up default successful behavior
	useCase.On("Execute", mock.Anything, mock.Anything).Return(
		func(ctx context.Context, cmd *usecase.ProcessOrderCommand) *usecase.ProcessOrderResult {
			return createSuccessfulProcessOrderResult(cmd.OrderID)
		},
		nil,
	)

	return useCase
}
