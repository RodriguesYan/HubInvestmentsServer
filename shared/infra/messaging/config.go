package messaging

import (
	"fmt"
	"os"
	"strconv"
)

// NewMessageHandlerConfigFromEnv creates a configuration from environment variables
// with fallback to defaults
func NewMessageHandlerConfigFromEnv() MessageHandlerConfig {
	config := DefaultMessageHandlerConfig()

	// Override with environment variables if present
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		config.URL = url
	}

	if timeout := os.Getenv("RABBITMQ_CONNECTION_TIMEOUT"); timeout != "" {
		if val, err := strconv.Atoi(timeout); err == nil {
			config.ConnectionTimeout = val
		}
	}

	if heartbeat := os.Getenv("RABBITMQ_HEARTBEAT_INTERVAL"); heartbeat != "" {
		if val, err := strconv.Atoi(heartbeat); err == nil {
			config.HeartbeatInterval = val
		}
	}

	if retries := os.Getenv("RABBITMQ_MAX_RETRIES"); retries != "" {
		if val, err := strconv.Atoi(retries); err == nil {
			config.MaxRetries = val
		}
	}

	if delay := os.Getenv("RABBITMQ_RETRY_DELAY"); delay != "" {
		if val, err := strconv.Atoi(delay); err == nil {
			config.RetryDelay = val
		}
	}

	if prefetch := os.Getenv("RABBITMQ_PREFETCH_COUNT"); prefetch != "" {
		if val, err := strconv.Atoi(prefetch); err == nil {
			config.PrefetchCount = val
		}
	}

	if confirm := os.Getenv("RABBITMQ_ENABLE_CONFIRMATION"); confirm != "" {
		if val, err := strconv.ParseBool(confirm); err == nil {
			config.EnableConfirmation = val
		}
	}

	return config
}

// RabbitMQConnectionConfig provides a more specific configuration for RabbitMQ
type RabbitMQConnectionConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	VHost    string
}

// ToURL converts RabbitMQConnectionConfig to AMQP URL
func (c RabbitMQConnectionConfig) ToURL() string {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 5672
	}
	if c.Username == "" {
		c.Username = "guest"
	}
	if c.Password == "" {
		c.Password = "guest"
	}
	if c.VHost == "" {
		c.VHost = "/"
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		c.Username, c.Password, c.Host, c.Port, c.VHost)
}

// NewRabbitMQConnectionConfigFromEnv creates RabbitMQ connection config from environment
func NewRabbitMQConnectionConfigFromEnv() RabbitMQConnectionConfig {
	config := RabbitMQConnectionConfig{
		Host:     "localhost",
		Port:     5672,
		Username: "guest",
		Password: "guest",
		VHost:    "/",
	}

	if host := os.Getenv("RABBITMQ_HOST"); host != "" {
		config.Host = host
	}

	if port := os.Getenv("RABBITMQ_PORT"); port != "" {
		if val, err := strconv.Atoi(port); err == nil {
			config.Port = val
		}
	}

	if username := os.Getenv("RABBITMQ_USERNAME"); username != "" {
		config.Username = username
	}

	if password := os.Getenv("RABBITMQ_PASSWORD"); password != "" {
		config.Password = password
	}

	if vhost := os.Getenv("RABBITMQ_VHOST"); vhost != "" {
		config.VHost = vhost
	}

	return config
}
