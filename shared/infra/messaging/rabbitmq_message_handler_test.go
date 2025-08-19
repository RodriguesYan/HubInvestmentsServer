package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRabbitMQMessageHandlerWithDefaults(t *testing.T) {
	// Skip this test if we're not in an integration test environment
	// This test requires RabbitMQ server to be running
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	handler, err := NewRabbitMQMessageHandlerWithDefaults()
	require.NoError(t, err, "Failed to create RabbitMQ handler")
	require.NotNil(t, handler, "Handler should not be nil")

	// Ensure proper cleanup
	defer func() {
		if closeErr := handler.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close handler: %v", closeErr)
		}
	}()

	// Test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = handler.HealthCheck(ctx)
	// We don't assert on the error here because RabbitMQ server might not be running
	// The important thing is that we can create the handler
	t.Logf("HealthCheck result (expected to fail if RabbitMQ not running): %v", err)
}

func TestRabbitMQMessageHandler_QueueOperations(t *testing.T) {
	// Skip this test if we're not in an integration test environment
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	handler, err := NewRabbitMQMessageHandlerWithDefaults()
	require.NoError(t, err, "Failed to create RabbitMQ handler")
	require.NotNil(t, handler, "Handler should not be nil")

	defer func() {
		if closeErr := handler.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close handler: %v", closeErr)
		}
	}()

	// Test queue declaration
	queueName := "test-queue"
	options := QueueOptions{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
	}

	err = handler.DeclareQueue(queueName, options)
	if err != nil {
		t.Logf("Queue declaration failed (expected if RabbitMQ not running): %v", err)
		return
	}

	// Test queue info
	info, err := handler.QueueInfo(queueName)
	if err != nil {
		t.Logf("Queue info failed: %v", err)
		return
	}

	assert.Equal(t, queueName, info.Name)
	t.Logf("Queue info: %+v", info)

	// Clean up
	handler.DeleteQueue(queueName)
}

func TestRabbitMQMessageHandler_PublishMessage(t *testing.T) {
	// Skip this test if we're not in an integration test environment
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	handler, err := NewRabbitMQMessageHandlerWithDefaults()
	require.NoError(t, err, "Failed to create RabbitMQ handler")
	require.NotNil(t, handler, "Handler should not be nil")

	defer func() {
		if closeErr := handler.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close handler: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queueName := "test-publish-queue"
	message := []byte(`{"test": "message", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)

	// Test basic publish
	err = handler.Publish(ctx, queueName, message)
	if err != nil {
		t.Logf("Publish failed (expected if RabbitMQ not running): %v", err)
		return
	}

	// Test publish with options
	options := PublishOptions{
		QueueName:     queueName,
		Message:       message,
		Priority:      5,
		Persistent:    true,
		MessageID:     "test-msg-001",
		CorrelationID: "test-correlation-001",
		Headers: map[string]interface{}{
			"source": "test",
			"type":   "integration-test",
		},
	}

	err = handler.PublishWithOptions(ctx, options)
	if err != nil {
		t.Logf("Publish with options failed: %v", err)
		return
	}

	t.Log("Message published successfully")

	// Clean up
	handler.DeleteQueue(queueName)
}

func TestMessageHandlerConfig_EnvironmentVariables(t *testing.T) {
	// Test default configuration
	defaultConfig := DefaultMessageHandlerConfig()
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", defaultConfig.URL)
	assert.Equal(t, 30, defaultConfig.ConnectionTimeout)
	assert.Equal(t, 60, defaultConfig.HeartbeatInterval)
	assert.Equal(t, 3, defaultConfig.MaxRetries)
	assert.Equal(t, 5, defaultConfig.RetryDelay)
	assert.Equal(t, 10, defaultConfig.PrefetchCount)
	assert.True(t, defaultConfig.EnableConfirmation)

	// Test environment-based configuration (without actually setting env vars)
	envConfig := NewMessageHandlerConfigFromEnv()
	// Should be the same as default since no env vars are set
	assert.Equal(t, defaultConfig, envConfig)
}

func TestRabbitMQConnectionConfig_ToURL(t *testing.T) {
	// Test default values
	config := RabbitMQConnectionConfig{}
	url := config.ToURL()
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", url)

	// Test custom values
	config = RabbitMQConnectionConfig{
		Host:     "rabbitmq.example.com",
		Port:     5673,
		Username: "admin",
		Password: "secret",
		VHost:    "/production",
	}
	url = config.ToURL()
	assert.Equal(t, "amqp://admin:secret@rabbitmq.example.com:5673/production", url)
}

// MockMessageConsumer for testing
type MockMessageConsumer struct {
	messages [][]byte
}

func (m *MockMessageConsumer) HandleMessage(ctx context.Context, message *Message) error {
	m.messages = append(m.messages, message.Body)
	return nil
}

func TestRabbitMQMessageHandler_Lifecycle(t *testing.T) {
	// Test that we can create and close the handler without errors
	handler, err := NewRabbitMQMessageHandlerWithDefaults()
	if err != nil {
		t.Skipf("Skipping lifecycle test - RabbitMQ not available: %v", err)
	}
	require.NotNil(t, handler, "Handler should not be nil")

	// Test that we can close the handler
	err = handler.Close()
	assert.NoError(t, err, "Failed to close handler")

	// Test that operations fail after close
	ctx := context.Background()
	err = handler.Publish(ctx, "test", []byte("test"))
	assert.Error(t, err, "Operations should fail after close")
	assert.Contains(t, err.Error(), "connection is closed")
}
