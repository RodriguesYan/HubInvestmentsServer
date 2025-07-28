package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	HTTPPort    string
	GRPCPort    string
	JWTSecret   string
	RedisHost   string
	RedisPort   string
	DatabaseURL string
}

var (
	instance *Config
	once     sync.Once
)

// Load loads configuration from environment variables or config file
// This function is thread-safe and will only load configuration once
func Load() *Config {
	once.Do(func() {
		// Try to load from config.env file
		err := godotenv.Load("config.env")
		if err != nil {
			log.Printf("Warning: Could not load config.env file: %v", err)
			log.Println("Using environment variables or default values...")
		}

		instance = &Config{
			HTTPPort:    getEnvWithDefault("HTTP_PORT", "localhost:8080"),
			GRPCPort:    getEnvWithDefault("GRPC_PORT", "localhost:50051"),
			JWTSecret:   getEnvWithDefault("MY_JWT_SECRET", "default-secret-key-change-in-production"),
			RedisHost:   getEnvWithDefault("REDIS_HOST", "localhost"),
			RedisPort:   getEnvWithDefault("REDIS_PORT", "6379"),
			DatabaseURL: getEnvWithDefault("DATABASE_URL", ""),
		}

		// Validate required configuration
		if instance.JWTSecret == "default-secret-key-change-in-production" {
			log.Println("Warning: Using default JWT secret. Please set MY_JWT_SECRET environment variable for production.")
		}
	})

	return instance
}

// Get returns the current configuration instance
// If configuration hasn't been loaded yet, it will load it first
func Get() *Config {
	if instance == nil {
		return Load()
	}
	return instance
}

// getEnvWithDefault gets an environment variable with a fallback default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsProduction checks if the application is running in production mode
func (c *Config) IsProduction() bool {
	return os.Getenv("ENVIRONMENT") == "production"
}

// GetRedisAddress returns the complete Redis address
func (c *Config) GetRedisAddress() string {
	return c.RedisHost + ":" + c.RedisPort
}
