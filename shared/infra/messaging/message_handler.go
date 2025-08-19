package messaging

import (
	"context"
	"errors"
)

var (
	ErrConnectionClosed = errors.New("messaging connection is closed")
	ErrChannelClosed    = errors.New("messaging channel is closed")
	ErrPublishFailed    = errors.New("failed to publish message")
	ErrQueueNotFound    = errors.New("queue not found")
)

// MessageHandler defines the interface for message queue operations
// This interface abstracts the underlying messaging implementation (RabbitMQ, etc.)
type MessageHandler interface {
	// Publish sends a message to the specified queue
	Publish(ctx context.Context, queueName string, message []byte) error

	// PublishWithOptions sends a message with additional options
	PublishWithOptions(ctx context.Context, options PublishOptions) error

	// Consume starts consuming messages from a queue
	Consume(ctx context.Context, queueName string, handler MessageConsumer) error

	// DeclareQueue creates a queue if it doesn't exist
	DeclareQueue(queueName string, options QueueOptions) error

	// DeleteQueue removes a queue
	DeleteQueue(queueName string) error

	// PurgeQueue removes all messages from a queue
	PurgeQueue(queueName string) error

	// QueueInfo returns information about a queue
	QueueInfo(queueName string) (*QueueInfo, error)

	// HealthCheck verifies the connection is healthy
	HealthCheck(ctx context.Context) error

	// Close closes the connection and cleans up resources
	Close() error
}

// PublishOptions provides options for publishing messages
type PublishOptions struct {
	QueueName     string
	Message       []byte
	Priority      uint8
	Persistent    bool
	TTL           int64 // Time to live in milliseconds
	MessageID     string
	CorrelationID string
	ReplyTo       string
	Headers       map[string]interface{}
}

// QueueOptions provides options for queue declaration
type QueueOptions struct {
	Durable    bool // Queue survives server restart
	AutoDelete bool // Queue is deleted when last consumer disconnects
	Exclusive  bool // Queue is used by only one connection
	NoWait     bool // Don't wait for server confirmation
	Arguments  map[string]interface{}
}

// QueueInfo contains information about a queue
type QueueInfo struct {
	Name       string
	Messages   int
	Consumers  int
	Durable    bool
	AutoDelete bool
	Exclusive  bool
}

// MessageConsumer defines the interface for message consumption
type MessageConsumer interface {
	// HandleMessage processes a received message
	// Return error to reject the message, nil to acknowledge
	HandleMessage(ctx context.Context, message *Message) error
}

// Message represents a received message
type Message struct {
	Body          []byte
	MessageID     string
	CorrelationID string
	ReplyTo       string
	Priority      uint8
	Timestamp     int64
	Headers       map[string]interface{}

	// Internal fields for acknowledgment
	deliveryTag uint64
	ack         func() error
	nack        func(requeue bool) error
}

// Ack acknowledges the message
func (m *Message) Ack() error {
	if m.ack != nil {
		return m.ack()
	}
	return nil
}

// Nack negatively acknowledges the message
func (m *Message) Nack(requeue bool) error {
	if m.nack != nil {
		return m.nack(requeue)
	}
	return nil
}

// MessageHandlerConfig holds configuration for message handlers
type MessageHandlerConfig struct {
	URL                string
	ConnectionTimeout  int // seconds
	HeartbeatInterval  int // seconds
	MaxRetries         int
	RetryDelay         int  // seconds
	PrefetchCount      int  // number of messages to prefetch
	EnableConfirmation bool // enable publisher confirmation
}

// DefaultMessageHandlerConfig returns a configuration with sensible defaults
func DefaultMessageHandlerConfig() MessageHandlerConfig {
	return MessageHandlerConfig{
		URL:                "amqp://guest:guest@localhost:5672/",
		ConnectionTimeout:  30,
		HeartbeatInterval:  60,
		MaxRetries:         3,
		RetryDelay:         5,
		PrefetchCount:      10,
		EnableConfirmation: true,
	}
}
