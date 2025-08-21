package rabbitmq

import (
	"context"
	"time"

	"HubInvestments/shared/infra/messaging"

	"github.com/stretchr/testify/mock"
)

// SharedMockMessageHandler is a shared mock implementation of messaging.MessageHandler
type SharedMockMessageHandler struct {
	mock.Mock
}

func (m *SharedMockMessageHandler) Publish(ctx context.Context, queueName string, message []byte) error {
	args := m.Called(ctx, queueName, message)
	return args.Error(0)
}

func (m *SharedMockMessageHandler) PublishWithOptions(ctx context.Context, options messaging.PublishOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *SharedMockMessageHandler) Consume(ctx context.Context, queueName string, handler messaging.MessageConsumer) error {
	args := m.Called(ctx, queueName, handler)
	return args.Error(0)
}

func (m *SharedMockMessageHandler) DeclareQueue(queueName string, options messaging.QueueOptions) error {
	args := m.Called(queueName, options)
	return args.Error(0)
}

func (m *SharedMockMessageHandler) DeleteQueue(queueName string) error {
	args := m.Called(queueName)
	return args.Error(0)
}

func (m *SharedMockMessageHandler) PurgeQueue(queueName string) error {
	args := m.Called(queueName)
	return args.Error(0)
}

func (m *SharedMockMessageHandler) QueueInfo(queueName string) (*messaging.QueueInfo, error) {
	args := m.Called(queueName)
	return args.Get(0).(*messaging.QueueInfo), args.Error(1)
}

func (m *SharedMockMessageHandler) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *SharedMockMessageHandler) Close() error {
	args := m.Called()
	return args.Error(0)
}

// SharedMockMessageConsumer is a shared mock implementation of messaging.MessageConsumer
type SharedMockMessageConsumer struct {
	mock.Mock
}

func (m *SharedMockMessageConsumer) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test helper functions

func CreateTestOrderMessage() *OrderMessage {
	return &OrderMessage{
		OrderID:   "test-order-123",
		UserID:    "user-456",
		Symbol:    "AAPL",
		OrderSide: "BUY",
		OrderType: "MARKET",
		Quantity:  100,
		Status:    "PENDING",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		MessageMetadata: OrderMessageMetadata{
			MessageID:       "msg-123",
			CorrelationID:   "test-order-123",
			Timestamp:       time.Now(),
			RetryAttempt:    0,
			Priority:        5,
			Source:          "test",
			MessageType:     "order_processing",
			ProcessingStage: "processing",
		},
	}
}

func CreateTestStatusUpdate() *OrderStatusUpdate {
	return &OrderStatusUpdate{
		OrderID:        "test-order-123",
		UserID:         "user-456",
		PreviousStatus: "PENDING",
		CurrentStatus:  "EXECUTED",
		UpdatedAt:      time.Now(),
		MessageMetadata: OrderMessageMetadata{
			MessageID:     "status-msg-123",
			CorrelationID: "test-order-123",
			Timestamp:     time.Now(),
			MessageType:   "status_update",
		},
	}
}

func CreateTestMessage(data []byte) *messaging.Message {
	return &messaging.Message{
		MessageID: "msg-123",
		Body:      data,
		Headers:   make(map[string]interface{}),
	}
}
