package messaging

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQMessageHandler implements MessageHandler interface for RabbitMQ
type RabbitMQMessageHandler struct {
	config     MessageHandlerConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	mutex      sync.RWMutex
	closed     bool
}

// NewRabbitMQMessageHandler creates a new RabbitMQ message handler
func NewRabbitMQMessageHandler(config MessageHandlerConfig) (MessageHandler, error) {
	handler := &RabbitMQMessageHandler{
		config: config,
	}

	if err := handler.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return handler, nil
}

// NewRabbitMQMessageHandlerWithDefaults creates a handler with default configuration
func NewRabbitMQMessageHandlerWithDefaults() (MessageHandler, error) {
	return NewRabbitMQMessageHandler(DefaultMessageHandlerConfig())
}

func (r *RabbitMQMessageHandler) connect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Close existing connection if any
	if r.connection != nil && !r.connection.IsClosed() {
		r.connection.Close()
	}

	// Establish new connection
	conn, err := amqp.DialConfig(r.config.URL, amqp.Config{
		Heartbeat: time.Duration(r.config.HeartbeatInterval) * time.Second,
		Dial:      amqp.DefaultDial(time.Duration(r.config.ConnectionTimeout) * time.Second),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS if prefetch count is configured
	if r.config.PrefetchCount > 0 {
		if err := ch.Qos(r.config.PrefetchCount, 0, false); err != nil {
			ch.Close()
			conn.Close()
			return fmt.Errorf("failed to set QoS: %w", err)
		}
	}

	// Enable publisher confirmation if configured
	if r.config.EnableConfirmation {
		if err := ch.Confirm(false); err != nil {
			ch.Close()
			conn.Close()
			return fmt.Errorf("failed to enable publisher confirmation: %w", err)
		}
	}

	r.connection = conn
	r.channel = ch
	r.closed = false

	// Set up connection close notification
	go r.handleConnectionClose()

	return nil
}

// checks if the channel is open and recreates it if needed
// Must be called with mutex already held
func (r *RabbitMQMessageHandler) ensureChannelHealthy() error {
	// Check if channel exists and is open
	if r.channel != nil {
		// Try a simple operation to check if channel is actually working
		// If this fails, the channel is closed or broken
		select {
		case <-r.channel.NotifyClose(make(chan *amqp.Error, 1)):
			// Channel is closed
			log.Printf("Channel is closed, recreating...")
		default:
			// Channel appears to be open
			return nil
		}
	}

	// Channel is nil or closed, need to recreate
	if r.connection == nil || r.connection.IsClosed() {
		// Connection is also broken, need full reconnect
		log.Printf("Connection is closed, attempting full reconnection...")
		return r.connectInternal()
	}

	// Connection is good, just need new channel
	log.Printf("Creating new channel...")
	ch, err := r.connection.Channel()
	if err != nil {
		log.Printf("Failed to create channel: %v, attempting full reconnection...", err)
		return r.connectInternal()
	}

	// Set QoS if prefetch count is configured
	if r.config.PrefetchCount > 0 {
		if err := ch.Qos(r.config.PrefetchCount, 0, false); err != nil {
			ch.Close()
			return fmt.Errorf("failed to set QoS: %w", err)
		}
	}

	// Enable publisher confirmation if configured
	if r.config.EnableConfirmation {
		if err := ch.Confirm(false); err != nil {
			ch.Close()
			return fmt.Errorf("failed to enable publisher confirmation: %w", err)
		}
	}

	r.channel = ch
	log.Printf("Channel recreated successfully")
	return nil
}

// connectInternal is the internal connection method without locking
// Must be called with mutex already held
func (r *RabbitMQMessageHandler) connectInternal() error {
	// Close existing connection if any
	if r.connection != nil && !r.connection.IsClosed() {
		r.connection.Close()
	}

	// Establish new connection
	conn, err := amqp.DialConfig(r.config.URL, amqp.Config{
		Heartbeat: time.Duration(r.config.HeartbeatInterval) * time.Second,
		Dial:      amqp.DefaultDial(time.Duration(r.config.ConnectionTimeout) * time.Second),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS if prefetch count is configured
	if r.config.PrefetchCount > 0 {
		if err := ch.Qos(r.config.PrefetchCount, 0, false); err != nil {
			ch.Close()
			conn.Close()
			return fmt.Errorf("failed to set QoS: %w", err)
		}
	}

	// Enable publisher confirmation if configured
	if r.config.EnableConfirmation {
		if err := ch.Confirm(false); err != nil {
			ch.Close()
			conn.Close()
			return fmt.Errorf("failed to enable publisher confirmation: %w", err)
		}
	}

	r.connection = conn
	r.channel = ch
	r.closed = false

	return nil
}

// handleConnectionClose handles connection close events and attempts reconnection
func (r *RabbitMQMessageHandler) handleConnectionClose() {
	closeNotify := r.connection.NotifyClose(make(chan *amqp.Error))

	for closeErr := range closeNotify {
		if r.closed {
			return // Intentional close
		}

		log.Printf("RabbitMQ connection closed: %v. Attempting to reconnect...", closeErr)

		// Attempt reconnection with exponential backoff
		for attempt := 1; attempt <= r.config.MaxRetries; attempt++ {
			time.Sleep(time.Duration(r.config.RetryDelay*attempt) * time.Second)

			if err := r.connect(); err != nil {
				log.Printf("Reconnection attempt %d failed: %v", attempt, err)
				continue
			}

			log.Printf("Successfully reconnected to RabbitMQ after %d attempts", attempt)
			return
		}

		log.Printf("Failed to reconnect to RabbitMQ after %d attempts", r.config.MaxRetries)
	}
}

// Publish sends a message to the specified queue
func (r *RabbitMQMessageHandler) Publish(ctx context.Context, queueName string, message []byte) error {
	options := PublishOptions{
		QueueName:  queueName,
		Message:    message,
		Persistent: true,
	}
	return r.PublishWithOptions(ctx, options)
}

// PublishWithOptions sends a message with additional options
func (r *RabbitMQMessageHandler) PublishWithOptions(ctx context.Context, options PublishOptions) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.closed {
		return ErrConnectionClosed
	}

	// Ensure channel is healthy before publishing
	if err := r.ensureChannelHealthy(); err != nil {
		return fmt.Errorf("failed to ensure channel health: %w", err)
	}

	// Prepare message properties
	headers := make(amqp.Table)
	for k, v := range options.Headers {
		headers[k] = v
	}

	publishing := amqp.Publishing{
		ContentType:   "application/json",
		Body:          options.Message,
		Priority:      options.Priority,
		MessageId:     options.MessageID,
		CorrelationId: options.CorrelationID,
		ReplyTo:       options.ReplyTo,
		Headers:       headers,
		Timestamp:     time.Now(),
	}

	// Set persistence
	if options.Persistent {
		publishing.DeliveryMode = amqp.Persistent
	}

	// Set TTL
	if options.TTL > 0 {
		publishing.Expiration = fmt.Sprintf("%d", options.TTL)
	}

	// Publish message
	err := r.channel.PublishWithContext(
		ctx,
		"",                // exchange
		options.QueueName, // routing key
		false,             // mandatory
		false,             // immediate
		publishing,
	)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrPublishFailed, err)
	}

	return nil
}

// Consume starts consuming messages from a queue
// Creates a dedicated channel for this consumer to avoid conflicts
func (r *RabbitMQMessageHandler) Consume(ctx context.Context, queueName string, handler MessageConsumer) error {
	r.mutex.RLock()
	if r.closed {
		r.mutex.RUnlock()
		return ErrConnectionClosed
	}

	// Create dedicated channel for this consumer
	conn := r.connection
	r.mutex.RUnlock()

	if conn == nil || conn.IsClosed() {
		return ErrConnectionClosed
	}

	// Create a new dedicated channel for this consumer
	log.Printf("Creating dedicated channel for consumer on queue: %s", queueName)
	consumerChannel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create consumer channel for %s: %w", queueName, err)
	}

	// Set QoS for this consumer channel
	if r.config.PrefetchCount > 0 {
		if err := consumerChannel.Qos(r.config.PrefetchCount, 0, false); err != nil {
			consumerChannel.Close()
			return fmt.Errorf("failed to set QoS on consumer channel: %w", err)
		}
	}

	// NOTE: Queue should already be declared by queue setup manager
	// Don't redeclare here to avoid TTL configuration conflicts

	// Start consuming on dedicated channel
	msgs, err := consumerChannel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		consumerChannel.Close()
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Process messages in dedicated goroutine with dedicated channel
	go func() {
		defer consumerChannel.Close() // Close channel when goroutine exits
		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-msgs:
				if !ok {
					return
				}

				// Convert AMQP delivery to our Message type
				message := &Message{
					Body:          delivery.Body,
					MessageID:     delivery.MessageId,
					CorrelationID: delivery.CorrelationId,
					ReplyTo:       delivery.ReplyTo,
					Priority:      delivery.Priority,
					Timestamp:     delivery.Timestamp.Unix(),
					Headers:       make(map[string]interface{}),
					deliveryTag:   delivery.DeliveryTag,
					ack: func() error {
						return delivery.Ack(false)
					},
					nack: func(requeue bool) error {
						return delivery.Nack(false, requeue)
					},
				}

				// Convert headers
				for k, v := range delivery.Headers {
					message.Headers[k] = v
				}

				// Handle message
				if err := handler.HandleMessage(ctx, message); err != nil {
					log.Printf("Error handling message on queue %s: %v", queueName, err)
					message.Nack(true) // Requeue on error
				} else {
					message.Ack()
				}
			}
		}
	}()

	return nil
}

// DeclareQueue creates a queue if it doesn't exist
func (r *RabbitMQMessageHandler) DeclareQueue(queueName string, options QueueOptions) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.closed {
		return ErrConnectionClosed
	}

	// Ensure channel is healthy before declaring queue
	if err := r.ensureChannelHealthy(); err != nil {
		return fmt.Errorf("failed to ensure channel health: %w", err)
	}

	return r.declareQueueInternal(queueName, options)
}

// declareQueueInternal is the internal implementation without locking
func (r *RabbitMQMessageHandler) declareQueueInternal(queueName string, options QueueOptions) error {
	if r.closed || r.channel == nil {
		return ErrConnectionClosed
	}

	args := make(amqp.Table)
	for k, v := range options.Arguments {
		args[k] = v
	}

	_, err := r.channel.QueueDeclare(
		queueName,
		options.Durable,
		options.AutoDelete,
		options.Exclusive,
		options.NoWait,
		args,
	)

	return err
}

// DeleteQueue removes a queue
func (r *RabbitMQMessageHandler) DeleteQueue(queueName string) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.closed || r.channel == nil {
		return ErrConnectionClosed
	}

	_, err := r.channel.QueueDelete(queueName, false, false, false)
	return err
}

// PurgeQueue removes all messages from a queue
func (r *RabbitMQMessageHandler) PurgeQueue(queueName string) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.closed || r.channel == nil {
		return ErrConnectionClosed
	}

	_, err := r.channel.QueuePurge(queueName, false)
	return err
}

// QueueInfo returns information about a queue
func (r *RabbitMQMessageHandler) QueueInfo(queueName string) (*QueueInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.closed || r.channel == nil {
		return nil, ErrConnectionClosed
	}

	queue, err := r.channel.QueueInspect(queueName)
	if err != nil {
		return nil, err
	}

	return &QueueInfo{
		Name:      queue.Name,
		Messages:  queue.Messages,
		Consumers: queue.Consumers,
	}, nil
}

// HealthCheck verifies the connection is healthy
func (r *RabbitMQMessageHandler) HealthCheck(ctx context.Context) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.closed {
		return ErrConnectionClosed
	}

	if r.connection == nil || r.connection.IsClosed() {
		return ErrConnectionClosed
	}

	if r.channel == nil || r.channel.IsClosed() {
		return ErrChannelClosed
	}

	// Try to declare a temporary queue to test the connection
	tempQueueName := fmt.Sprintf("health-check-%d", time.Now().UnixNano())
	_, err := r.channel.QueueDeclare(
		tempQueueName,
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Clean up the temporary queue
	r.channel.QueueDelete(tempQueueName, false, false, false)

	return nil
}

// Close closes the connection and cleans up resources
func (r *RabbitMQMessageHandler) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.closed = true

	var errors []error

	if r.channel != nil && !r.channel.IsClosed() {
		if err := r.channel.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if r.connection != nil && !r.connection.IsClosed() {
		if err := r.connection.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if len(errors) > 0 {
		return errors[0] // Return first error
	}

	return nil
}
